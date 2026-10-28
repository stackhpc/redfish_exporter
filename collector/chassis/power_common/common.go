package power_common

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem                    = "chassis"
	ChassisPowerSupplyLabelNames = []string{"resource", "chassis_id", "power_supply", "power_supply_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_state", fmt.Sprintf("powersupply state of chassis component,%s", common.StateHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_health", fmt.Sprintf("powersupply health of chassis component,%s", common.HealthHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_efficiency_percentage", "rated efficiency, as a percentage, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_last_power_output_watts", "average power output, measured in Watts, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_input_watts", "measured input power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_output_watts", "measured output power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_capacity_watts", "power_capacity_watts of powersupply on this chassis", ChassisPowerSupplyLabelNames)

	return metrics
}

func Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
}

func ParsePowerSupply(ch chan<- prometheus.Metric, chassisID string, powerSupply schemas.PowerSupply) {
	labelValues := []string{"power_supply", chassisID, powerSupply.Name, powerSupply.MemberID}
	if chassisPowerInfoPowerSupplyStateValue, ok := common.ParseCommonStatusState(powerSupply.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_state"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, labelValues...)
	}
	if chassisPowerInfoPowerSupplyHealthValue, ok := common.ParseCommonStatusHealth(powerSupply.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_health"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, labelValues...)
	}
	if powerSupply.EfficiencyPercent != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_power_efficiency_percentage"].Desc, prometheus.GaugeValue, float64(*powerSupply.EfficiencyPercent), labelValues...)
	}
	if powerSupply.LastPowerOutputWatts != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_last_power_output_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.LastPowerOutputWatts), labelValues...)
	}
	if powerSupply.PowerCapacityWatts != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_power_capacity_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerCapacityWatts), labelValues...)
	}
	if powerSupply.PowerInputWatts != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_power_input_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerInputWatts), labelValues...)
	}
	if powerSupply.PowerOutputWatts != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_powersupply_power_output_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerOutputWatts), labelValues...)
	}
}
