package thermal_common

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem             = "chassis"
	FanLabelNames         = []string{"resource", "chassis_id", "fan", "fan_id", "fan_unit"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

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

func ParseFan(ch chan<- prometheus.Metric, chassisID *string, chassisFan *schemas.ThermalFan) {
	chassisFanID := chassisFan.MemberID
	chassisFanLabelvalues := []string{"fan", *chassisID, chassisFan.Name, chassisFanID, strings.ToLower(string(chassisFan.ReadingUnits))} // e.g. RPM -> rpm, Percentage -> percentage

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

		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm"].Desc, prometheus.GaugeValue, rpm_f, chassisFanLabelvalues...)
	}

	if chassisFanStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisFan.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_health"].Desc, prometheus.GaugeValue, chassisFanStatusHealthValue, chassisFanLabelvalues...)
	}

	if chassisFanStatusStateValue, ok := common.ParseCommonStatusState(chassisFan.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_state"].Desc, prometheus.GaugeValue, chassisFanStatusStateValue, chassisFanLabelvalues...)
	}
	if chassisFan.MinReadingRange != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_min"].Desc, prometheus.GaugeValue, float64(*chassisFan.MinReadingRange), chassisFanLabelvalues...)
	}
	if chassisFan.MaxReadingRange != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_max"].Desc, prometheus.GaugeValue, float64(*chassisFan.MaxReadingRange), chassisFanLabelvalues...)
	}
	if chassisFanPercentage != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_fan_rpm_percentage"].Desc, prometheus.GaugeValue, *chassisFanPercentage, chassisFanLabelvalues...)
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
