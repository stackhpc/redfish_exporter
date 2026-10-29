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
	Subsystem = "chassis"

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_min", "lowest possible fan RPM or percentage, on this chassis component", thermal_common.FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_lower_threshold_critical", "threshold below the normal range fan RPM or percentage, but not fatal, on this chassis component", thermal_common.FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_lower_threshold_non_critical", "threshold below the normal range fan RPM or percentage, but not critical, on this chassis component", thermal_common.FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_lower_threshold_fatal", "threshold below the normal range fan RPM or percentage, and is fatal, on this chassis component", thermal_common.FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_upper_threshold_critical", "threshold above the normal range fan RPM or percentage, but not fatal, on this chassis component", thermal_common.FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_upper_threshold_non_critical", "threshold above the normal range fan RPM or percentage, but not critical, on this chassis component", thermal_common.FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_upper_threshold_fatal", "threshold above the normal range fan RPM or percentage, and is fatal, on this chassis component", thermal_common.FanLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "temperature_sensor_state", fmt.Sprintf("status state of temperature on this chassis component,%s", common.StateHelp), thermal_common.TemperatureLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "temperature_sensor_health", fmt.Sprintf("status health of temperature on this chassis component,%s", common.StateHelp), thermal_common.TemperatureLabelNames)
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
	thermal_common.Describe(ch)
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
		wg.Go(func() { parseFan(ch, &c.chassis.ID, &chassisFan) })
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
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_temperature_celsius"].Desc, prometheus.GaugeValue, float64(*chassisTemperature.ReadingCelsius), chassisTemperatureLabelvalues...)
	}
}

func parseFan(ch chan<- prometheus.Metric, chassisID *string, chassisFan *schemas.ThermalFan) {
	chassisFanID := chassisFan.MemberID
	chassisFanLabelvalues := []string{"fan", *chassisID, chassisFan.Name, chassisFanID}

	var chassisFanPercentage *float64
	switch chassisFan.ReadingUnits {
	case schemas.PercentReadingUnits:
		if chassisFan.Reading == nil {
			break
		}
		chassisFanPercentage = new(float64)
		*chassisFanPercentage = float64(*chassisFan.Reading)
	case schemas.RPMReadingUnits: // Calculate fan percentage from best value for min/max
		if chassisFan.Reading == nil {
			break
		}
		// Some vendors (e.g. PowerEdge C6420) report null RPMs for Min/Max, as well as Lower/UpperFatal,
		// but provide Lower/UpperCritical, so we use the most extreme value we have as the limits.
		var min *int = chassisFan.MinReadingRange
		var max *int = chassisFan.MaxReadingRange
		if chassisFan.LowerThresholdCritical != nil && (min == nil || *min > *chassisFan.LowerThresholdCritical) {
			min = chassisFan.LowerThresholdCritical
		}
		if chassisFan.LowerThresholdFatal != nil && (min == nil || *min > *chassisFan.LowerThresholdFatal) {
			min = chassisFan.LowerThresholdFatal
		}
		if chassisFan.LowerThresholdNonCritical != nil && (min == nil || *min > *chassisFan.LowerThresholdNonCritical) {
			min = chassisFan.LowerThresholdNonCritical
		}
		if chassisFan.UpperThresholdCritical != nil && (max == nil || *max < *chassisFan.UpperThresholdCritical) {
			max = chassisFan.UpperThresholdCritical
		}
		if chassisFan.UpperThresholdFatal != nil && (max == nil || *max < *chassisFan.UpperThresholdFatal) {
			max = chassisFan.UpperThresholdFatal
		}
		if chassisFan.UpperThresholdNonCritical != nil && (max == nil || *max < *chassisFan.UpperThresholdNonCritical) {
			max = chassisFan.UpperThresholdNonCritical
		}

		rpm_f := float64(*chassisFan.Reading)
		if min != nil && max != nil {
			min_f := float64(*min)
			max_f := float64(*max)
			chassisFanPercentage = new(float64)
			*chassisFanPercentage = (rpm_f - min_f) / (max_f - min_f) * 100.0
		}

		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm"].Desc, prometheus.GaugeValue, rpm_f, chassisFanLabelvalues...)
	}

	if chassisFanStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisFan.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_health"].Desc, prometheus.GaugeValue, chassisFanStatusHealthValue, chassisFanLabelvalues...)
	}

	if chassisFanStatusStateValue, ok := common.ParseCommonStatusState(chassisFan.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_state"].Desc, prometheus.GaugeValue, chassisFanStatusStateValue, chassisFanLabelvalues...)
	}
	if chassisFan.MinReadingRange != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_min"].Desc, prometheus.GaugeValue, float64(*chassisFan.MinReadingRange), chassisFanLabelvalues...)
	}
	if chassisFan.MaxReadingRange != nil {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm_max"].Desc, prometheus.GaugeValue, float64(*chassisFan.MaxReadingRange), chassisFanLabelvalues...)
	}
	if chassisFanPercentage != nil {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm_percentage"].Desc, prometheus.GaugeValue, *chassisFanPercentage, chassisFanLabelvalues...)
	}
	if chassisFan.LowerThresholdCritical != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_lower_threshold_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.LowerThresholdCritical), chassisFanLabelvalues...)
	}
	if chassisFan.LowerThresholdFatal != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_lower_threshold_fatal"].Desc, prometheus.GaugeValue, float64(*chassisFan.LowerThresholdFatal), chassisFanLabelvalues...)
	}
	if chassisFan.LowerThresholdNonCritical != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_lower_threshold_non_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.LowerThresholdNonCritical), chassisFanLabelvalues...)
	}
	if chassisFan.UpperThresholdCritical != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_upper_threshold_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.UpperThresholdCritical), chassisFanLabelvalues...)
	}
	if chassisFan.UpperThresholdFatal != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_upper_threshold_fatal"].Desc, prometheus.GaugeValue, float64(*chassisFan.UpperThresholdFatal), chassisFanLabelvalues...)
	}
	if chassisFan.UpperThresholdNonCritical != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_upper_threshold_non_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.UpperThresholdNonCritical), chassisFanLabelvalues...)
	}
}
