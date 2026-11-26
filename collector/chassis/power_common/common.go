package power_common

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
)

var (
	Subsystem                    = "chassis"
	ChassisPowerSupplyLabelNames = []string{"resource", "chassis_id", "power_supply", "power_supply_id"}
	ChassisPowerLabelNames       = []string{"resource", "chassis_id", "sensor_name", "sensor_id"}

	MetricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_state", fmt.Sprintf("powersupply state of chassis component,%s", common.StateHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_health", fmt.Sprintf("health of power supply itself: %s", common.HealthHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_health_rollup", fmt.Sprintf("health of power supply, considering dependents: %s", common.HealthHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_efficiency_percentage", "rated efficiency, as a percentage, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_last_power_output_watts", "average power output, measured in Watts, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_input_watts", "measured input power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_output_watts", "measured output power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_capacity_watts", "power_capacity_watts of powersupply on this chassis", ChassisPowerSupplyLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "power_voltage_volts", "power voltage volts number of chassis component", ChassisPowerLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_average_consumed_watts", "power wattage watts number of chassis component", ChassisPowerLabelNames)

	return metrics
}

func Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range MetricMap {
		ch <- metric.Desc
	}
}
