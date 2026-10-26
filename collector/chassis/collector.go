package chassis

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/stackhpc/redfish_exporter/collector/chassis/network_adapters"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal"
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

	chassisMetrics = createChassisMetricMap()
)

// ChassisCollector implements the prometheus.Collector.
type ChassisCollector struct {
	Ctx                   *common.CollectionContext
	metrics               map[string]common.Metric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
	thermal               *thermal.Collector
}

func createChassisMetricMap() map[string]common.Metric {
	chassisMetrics := make(map[string]common.Metric)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "health", fmt.Sprintf("health of chassis,%s", common.HealthHelp), ChassisLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "state", fmt.Sprintf("state of chassis,%s", common.StateHelp), ChassisLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "model_info", "organization responsible for producing the chassis, the name by which the manufacturer generally refers to the chassis, and a part number and sku assigned by the organization that is responsible for producing or manufacturing the chassis", ChassisModel)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "physical_security_sensor_state", fmt.Sprintf("indicates the known state of the physical security sensor, such as if it is hardware intrusion detected,%s", common.IntrusionSensorHelp), ChassisPhysicalSecurityLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "log_service_state", fmt.Sprintf("chassis log service state,%s", common.StateHelp), ChassisLogServiceLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "log_service_health_state", fmt.Sprintf("chassis log service health state,%s", common.HealthHelp), ChassisLogServiceLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "log_entry_severity_state", fmt.Sprintf("chassis log entry severity state,%s", common.SeverityHelp), ChassisLogEntryLabelNames)

	return chassisMetrics
}

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(ctx *common.CollectionContext, logger *slog.Logger) *ChassisCollector {
	// get service from redfish client

	return &ChassisCollector{
		metrics: chassisMetrics,
		Ctx:     ctx,
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
func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
	c.collectorScrapeStatus.Describe(ch)

}

// Collect implemented prometheus.Collector
func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := c.Log
	collectLogs := c.Ctx.CollectLogs
	service := c.Ctx.RedfishClient.Service

	wg := &sync.WaitGroup{}

	// get a list of chassis from service
	if chassises, err := service.Chassis(); err != nil {
		collectorLogContext.Error("error getting chassis from service", "operation", "service.Chassis()", "err", err)
	} else {
		// process the chassises
		for _, chassis := range chassises {
			chassisLogContext := collectorLogContext.With("Chassis", chassis.ID)
			chassisLogContext.Info("collector scrape started")
			chassisID := chassis.ID
			chassisStatus := chassis.Status
			chassisStatusState := chassisStatus.State
			chassisStatusHealth := chassisStatus.Health
			ChassisLabelValues := []string{"chassis", chassisID}
			if chassisStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisStatusHealth); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].Desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
			}
			if chassisStatusStateValue, ok := common.ParseCommonStatusState(chassisStatusState); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].Desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
			}

			chassisManufacturer := chassis.Manufacturer
			chassisModel := chassis.Model
			chassisPartNumber := chassis.PartNumber
			chassisSKU := chassis.SKU
			ChassisModelLabelValues := []string{"chassis", chassisID, chassisManufacturer, chassisModel, chassisPartNumber, chassisSKU}
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_model_info"].Desc, prometheus.GaugeValue, 1, ChassisModelLabelValues...)

			thermal := thermal.NewCollector(c.Ctx, c.Log, chassis)
			wg.Go(func() { thermal.Collect(ch) })

			power := power.NewCollector(c.Ctx, c.Log, chassis)
			wg.Go(func() { power.Collect(ch) })

			network_adapters := network_adapters.NewCollector(c.Ctx, c.Log, chassis)
			wg.Go(func() { network_adapters.Collect(ch) })

			physicalSecurity := chassis.PhysicalSecurity
			if physicalSecurity != (schemas.PhysicalSecurity{}) {
				physicalSecurityIntrusionSensor := physicalSecurity.IntrusionSensor
				physicalSecurityIntrusionSensorNumber := fmt.Sprint(physicalSecurity.IntrusionSensorNumber)
				physicalSecurityIntrusionSensorReArmMethod := string(physicalSecurity.IntrusionSensorReArm)

				if phySecIntrusionSensor, ok := parsePhySecIntrusionSensor(physicalSecurityIntrusionSensor); ok {
					ChassisPhysicalSecurityLabelValues := []string{"physical_security", chassisID, physicalSecurityIntrusionSensorNumber, physicalSecurityIntrusionSensorReArmMethod}
					ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_physical_security_sensor_state"].Desc, prometheus.GaugeValue, phySecIntrusionSensor, ChassisPhysicalSecurityLabelValues...)
				}
			}

			// process log services
			if collectLogs {
				logServices, err := chassis.LogServices()
				if err != nil {
					chassisLogContext.Error("error getting log services from chassis", "operation", "chassis.LogServices()", "err", err)
				} else if logServices == nil {
					chassisLogContext.Info("no log services found", "operation", "chassis.LogServices()")
				} else {
					wg6 := &sync.WaitGroup{}
					wg6.Add(len(logServices))

					for _, logService := range logServices {
						if err = common.ParseLogService(ch, chassisMetrics, c.Ctx, chassisLogContext, ChassisSubsystem, chassisID, logService, wg6); err != nil {
							chassisLogContext.Error("error getting log entries from log service", "operation", "chassis.LogServices()", "err", err)
						}
					}
				}
			}
			chassisLogContext.Info("collector scrape completed")
		}
	}

	wg.Wait()
	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
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
