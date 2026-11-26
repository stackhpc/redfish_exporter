package power

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_common"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "chassis"

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "power_voltage_state", fmt.Sprintf("power voltage state of chassis component,%s", common.StateHelp), power_common.ChassisPowerLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	Log     *slog.Logger
	chassis *schemas.Chassis
}

// NewCollector returns a collector that collects chassis power data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.power",
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
	power_common.Describe(ch)
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	powerInfo, err := c.chassis.Power()
	if err != nil {
		c.Log.Error("error getting power data from chassis", "operation", "chassis.Power()", "err", err)
		return
	} else if powerInfo == nil {
		c.Log.Debug("no power data found", "operation", "chassis.Power()")
		return
	}

	wg := &sync.WaitGroup{}

	for _, voltage := range powerInfo.Voltages {
		wg.Go(func() { parseVoltage(ch, c.chassis.ID, voltage) })
	}

	for _, powerControl := range powerInfo.PowerControl {
		wg.Go(func() { parsePowerControl(ch, c.chassis.ID, powerControl) })
	}

	for _, powerSupply := range powerInfo.PowerSupplies {
		wg.Go(func() { parsePowerSupply(ch, c.chassis.ID, powerSupply) })
	}

	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

func parseVoltage(ch chan<- prometheus.Metric, chassisID string, voltageInfo schemas.Voltage) {
	labelValues := []string{"power_voltage", chassisID, voltageInfo.Name, voltageInfo.MemberID}
	if chassisPowerInfoVoltageStateValue, ok := common.ParseCommonStatusState(voltageInfo.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_voltage_state"].Desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, labelValues...)
	}
	if voltageInfo.ReadingVolts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_voltage_volts"].Desc, prometheus.GaugeValue, float64(*voltageInfo.ReadingVolts), labelValues...)
	}
}

func parsePowerControl(ch chan<- prometheus.Metric, chassisID string, powerControl schemas.PowerControl) {
	labelValues := []string{"power_wattage", chassisID, powerControl.Name, powerControl.MemberID}
	if powerControl.PowerMetrics.AverageConsumedWatts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_average_consumed_watts"].Desc, prometheus.GaugeValue, float64(*powerControl.PowerMetrics.AverageConsumedWatts), labelValues...)
	}
}

func parsePowerSupply(ch chan<- prometheus.Metric, chassisID string, powerSupply schemas.PowerSupply) {
	labelValues := []string{"Power", chassisID, powerSupply.Name, powerSupply.MemberID}
	if chassisPowerInfoPowerSupplyStateValue, ok := common.ParseCommonStatusState(powerSupply.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_state"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, labelValues...)
	}
	if chassisPowerInfoPowerSupplyHealthValue, ok := common.ParseCommonStatusHealth(powerSupply.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_health"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, labelValues...)
	}
	if healthRollup, ok := common.ParseCommonStatusHealth(powerSupply.Status.HealthRollup); ok {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_health_rollup"].Desc, prometheus.GaugeValue, healthRollup, labelValues...)
	}

	if powerSupply.EfficiencyPercent != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_efficiency_percentage"].Desc, prometheus.GaugeValue, float64(*powerSupply.EfficiencyPercent), labelValues...)
	}
	if powerSupply.LastPowerOutputWatts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_last_power_output_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.LastPowerOutputWatts), labelValues...)
	}
	if powerSupply.PowerCapacityWatts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_capacity_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerCapacityWatts), labelValues...)
	}
	if powerSupply.PowerInputWatts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_input_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerInputWatts), labelValues...)
	}
	if powerSupply.PowerOutputWatts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_output_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerOutputWatts), labelValues...)
	}
}
