package thermal

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_common"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem             = "chassis"
	TemperatureLabelNames = []string{"resource", "chassis_id", "sensor", "sensor_id"}
	FanLabelNames         = []string{"resource", "chassis_id", "fan", "fan_id", "fan_unit"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "temperature_sensor_state", fmt.Sprintf("status state of temperature on this chassis component,%s", common.StateHelp), TemperatureLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "temperature_sensor_health", fmt.Sprintf("status health of temperature on this chassis component,%s", common.StateHelp), TemperatureLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "temperature_celsius", "celsius of temperature on this chassis component", TemperatureLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "fan_health", fmt.Sprintf("fan health on this chassis component,%s", common.HealthHelp), FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_state", fmt.Sprintf("fan state on this chassis component,%s", common.StateHelp), FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm", "fan RPM or percentage on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_percentage", "fan RPM, as a percentage of the min-max RPMs possible, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_min", "lowest possible fan RPM or percentage, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_max", "highest possible fan RPM or percentage, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_lower_threshold_critical", "threshold below the normal range fan RPM or percentage, but not fatal, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_lower_threshold_non_critical", "threshold below the normal range fan RPM or percentage, but not critical, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_lower_threshold_fatal", "threshold below the normal range fan RPM or percentage, and is fatal, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_upper_threshold_critical", "threshold above the normal range fan RPM or percentage, but not fatal, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_upper_threshold_non_critical", "threshold above the normal range fan RPM or percentage, but not critical, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_upper_threshold_fatal", "threshold above the normal range fan RPM or percentage, and is fatal, on this chassis component", FanLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	Log     *slog.Logger
	chassis *schemas.Chassis
}

// NewCollector returns a collector that collects chassis thermal data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.thermal",
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

// Given a chassis, collect data from its Thermal resource
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	thermal, err := c.chassis.Thermal()
	if err != nil {
		c.Log.Error("error getting thermal data from chassis", "operation", "chassis.Thermal()", "err", err)
		return
	}
	if thermal == nil {
		c.Log.Debug("no thermal data found", "operation", "chassis.Thermal()")
		return
	}

	// Process temperature sensors
	chassisTemperatures := thermal.Temperatures
	wg := &sync.WaitGroup{}
	for _, chassisTemperature := range chassisTemperatures {
		wg.Go(func() { parseChassisTemperature(ch, c.chassis.ID, chassisTemperature) })
	}

	// Process fans
	chassisFans := thermal.Fans
	for _, chassisFan := range chassisFans {
		wg.Go(func() { thermal_common.ParseFan(ch, &c.chassis.ID, &chassisFan) })
	}

	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

func parseChassisTemperature(ch chan<- prometheus.Metric, chassisID string, chassisTemperature schemas.Temperature) {
	chassisTemperatureSensorName := chassisTemperature.Name
	chassisTemperatureSensorID := chassisTemperature.MemberID
	chassisTemperatureStatus := chassisTemperature.Status
	chassisTemperatureLabelvalues := []string{"temperature", chassisID, chassisTemperatureSensorName, chassisTemperatureSensorID}

	chassisTemperatureStatusHealth := chassisTemperatureStatus.Health
	if chassisTemperatureStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisTemperatureStatusHealth); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_temperature_sensor_health"].Desc, prometheus.GaugeValue, chassisTemperatureStatusHealthValue, chassisTemperatureLabelvalues...)
	}

	chassisTemperatureStatusState := chassisTemperatureStatus.State
	//			chassisTemperatureStatusLabelNames :=[]string{BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
	//		ch <- prometheus.MustNewConstMetric(metricMap["chassis_temperature_status_health"].Desc, prometheus.GaugeValue, parseCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureLabelvalues...)
	if chassisTemperatureStatusStateValue, ok := common.ParseCommonStatusState(chassisTemperatureStatusState); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_temperature_sensor_state"].Desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
	}

	if chassisTemperature.ReadingCelsius != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_temperature_celsius"].Desc, prometheus.GaugeValue, float64(*chassisTemperature.ReadingCelsius), chassisTemperatureLabelvalues...)
	}
}
