package thermal_common

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
)

var (
	Subsystem             = "chassis"
	FanLabelNames         = []string{"resource", "chassis_id", "fan", "fan_id"}
	TemperatureLabelNames = []string{"resource", "chassis_id", "sensor", "sensor_id"}

	MetricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "fan_health", fmt.Sprintf("fan health on this chassis component,%s", common.HealthHelp), FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_state", fmt.Sprintf("fan state on this chassis component,%s", common.StateHelp), FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm", "fan RPM", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_percentage", "fan RPM, as a percentage of the rated RPMs, on this chassis component", FanLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "fan_rpm_max", "highest possible fan RPM or percentage, on this chassis component", FanLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "temperature_celsius", "celsius of temperature on this chassis component", TemperatureLabelNames)

	return metrics
}

func Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range MetricMap {
		ch <- metric.Desc
	}
}
