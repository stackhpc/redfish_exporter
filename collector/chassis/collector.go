package chassis

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/stackhpc/redfish_exporter/collector/chassis/network_adapters"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_subsystem"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem"
	"github.com/stackhpc/redfish_exporter/common"
)

// ChassisSubsystem is the chassis subsystem
var (
	ChassisSubsystem                  = "chassis"
	ChassisLabelNames                 = []string{"resource", "chassis_id"}
	ChassisModel                      = []string{"resource", "chassis_id", "manufacturer", "model", "part_number", "sku"}
	ChassisPhysicalSecurityLabelNames = []string{"resource", "chassis_id", "intrusion_sensor_number", "intrusion_sensor_rearm"}

	ChassisLogServiceLabelNames = []string{"chassis_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	ChassisLogEntryLabelNames   = []string{"chassis_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)
	common.AddToMetricMap(metrics, ChassisSubsystem, "health", fmt.Sprintf("health of chassis,%s", common.HealthHelp), ChassisLabelNames)
	common.AddToMetricMap(metrics, ChassisSubsystem, "state", fmt.Sprintf("state of chassis,%s", common.StateHelp), ChassisLabelNames)
	common.AddToMetricMap(metrics, ChassisSubsystem, "model_info", "organization responsible for producing the chassis, the name by which the manufacturer generally refers to the chassis, and a part number and sku assigned by the organization that is responsible for producing or manufacturing the chassis", ChassisModel)

	common.AddToMetricMap(metrics, ChassisSubsystem, "physical_security_sensor_state", fmt.Sprintf("indicates the known state of the physical security sensor, such as if it is hardware intrusion detected,%s", common.IntrusionSensorHelp), ChassisPhysicalSecurityLabelNames)

	common.AddToMetricMap(metrics, ChassisSubsystem, "log_service_state", fmt.Sprintf("chassis log service state,%s", common.StateHelp), ChassisLogServiceLabelNames)
	common.AddToMetricMap(metrics, ChassisSubsystem, "log_service_health_state", fmt.Sprintf("chassis log service health state,%s", common.HealthHelp), ChassisLogServiceLabelNames)
	common.AddToMetricMap(metrics, ChassisSubsystem, "log_entry_severity_state", fmt.Sprintf("chassis log entry severity state,%s", common.SeverityHelp), ChassisLogEntryLabelNames)

	return metrics
}

// Collector implements the prometheus.Collector.
type Collector struct {
	Ctx                   *common.CollectionContext
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
	thermal               *thermal.Collector
}

// NewCollector returns a collector that collecting chassis statistics
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger) *Collector {
	// get service from redfish client

	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "ChassisCollector",
		),
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: common.Namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
	c.collectorScrapeStatus.Describe(ch)
	network_adapters.Describe(ch)
	power.Describe(ch)
	power_subsystem.Describe(ch)
	thermal.Describe(ch)
	thermal_subsystem.Describe(ch)
}

// Collect implemented prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	service := c.Ctx.RedfishClient.Service

	wg := &sync.WaitGroup{}

	// get a list of chassis from service
	chassises, err := service.Chassis()
	if err != nil {
		c.Log.Error("error getting chassis from service", "operation", "service.Chassis()", "err", err)
		return
	}

	// process the chassises
	for _, chassis := range chassises {
		chassisLogContext := c.Log.With("chassis", chassis.ID)

		wg.Go(func() { c.collectChassis(ch, chassis) })

		// Thermal() was deprecated in Redfish v1.15.0
		thermal := thermal.NewCollector(c.Ctx, chassisLogContext, chassis)
		wg.Go(func() { thermal.Collect(ch) })

		// It was replaced with ThermalSubsystem()
		thermal_subsystem := thermal_subsystem.NewCollector(c.Ctx, chassisLogContext, chassis)
		wg.Go(func() { thermal_subsystem.Collect(ch) })

		// Power() was deprecated in Redfish v1.15.0
		power := power.NewCollector(c.Ctx, chassisLogContext, chassis)
		wg.Go(func() { power.Collect(ch) })

		// It was replaced with PowerSubsystem()
		power_subsystem := power_subsystem.NewCollector(c.Ctx, chassisLogContext, chassis)
		wg.Go(func() { power_subsystem.Collect(ch) })

		network_adapters := network_adapters.NewCollector(c.Ctx, chassisLogContext, chassis)
		wg.Go(func() { network_adapters.Collect(ch) })

		// process log services
		if c.Ctx.CollectLogs {
			wg.Go(func() { c.collectLogs(ch, chassis) })
		}
	}

	wg.Wait()
	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
	c.Log.Info("collection complete", "duration", time.Now().Sub(startTime))
}

func (c *Collector) collectChassis(ch chan<- prometheus.Metric, chassis *schemas.Chassis) {
	ChassisLabelValues := []string{"chassis", chassis.ID}
	if chassisStatusHealthValue, ok := common.ParseCommonStatusHealth(chassis.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_health"].Desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
	}
	if chassisStatusStateValue, ok := common.ParseCommonStatusState(chassis.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_state"].Desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
	}

	ChassisModelLabelValues := []string{"chassis", chassis.ID, chassis.Manufacturer, chassis.Model, chassis.PartNumber, chassis.SKU}
	ch <- prometheus.MustNewConstMetric(metricMap["chassis_model_info"].Desc, prometheus.GaugeValue, 1, ChassisModelLabelValues...)

	physicalSecurity := chassis.PhysicalSecurity
	if physicalSecurity != (schemas.PhysicalSecurity{}) {
		physicalSecurityIntrusionSensor := physicalSecurity.IntrusionSensor
		physicalSecurityIntrusionSensorNumber := fmt.Sprint(physicalSecurity.IntrusionSensorNumber)
		physicalSecurityIntrusionSensorReArmMethod := string(physicalSecurity.IntrusionSensorReArm)

		if phySecIntrusionSensor, ok := parsePhySecIntrusionSensor(physicalSecurityIntrusionSensor); ok {
			ChassisPhysicalSecurityLabelValues := []string{"physical_security", chassis.ID, physicalSecurityIntrusionSensorNumber, physicalSecurityIntrusionSensorReArmMethod}
			ch <- prometheus.MustNewConstMetric(metricMap["chassis_physical_security_sensor_state"].Desc, prometheus.GaugeValue, phySecIntrusionSensor, ChassisPhysicalSecurityLabelValues...)
		}
	}
}

func (c *Collector) collectLogs(ch chan<- prometheus.Metric, chassis *schemas.Chassis) {
	logServices, err := chassis.LogServices()
	if err != nil {
		c.Log.Error("error getting log services from chassis", "operation", "chassis.LogServices()", "err", err)
		return
	} else if logServices == nil {
		c.Log.Debug("no log services found", "operation", "chassis.LogServices()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, logService := range logServices {
		wg.Go(func() {
			if err = common.ParseLogService(ch, metricMap, c.Ctx, c.Log.With("log_service", logService.ID), ChassisSubsystem, chassis.ID, logService); err != nil {
				c.Log.Error("error getting log entries from log service", "operation", "chassis.LogServices()", "err", err)
			}
		})
	}

	wg.Wait()
}

func parsePhySecIntrusionSensor(method schemas.IntrusionSensor) (float64, bool) {
	if bytes.Equal([]byte(method), []byte("Normal")) {
		return float64(1), true

	}
	if bytes.Equal([]byte(method), []byte("TamperingDetected")) {
		return float64(2), true
	}
	if bytes.Equal([]byte(method), []byte("HardwareIntrusion")) {
		return float64(3), true
	}

	return float64(0), false
}
