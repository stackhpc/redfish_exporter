package power

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem                     = "chassis"
	ChassisPowerVoltageLabelNames = []string{"resource", "chassis_id", "power_voltage", "power_voltage_id"}
	ChassisPowerSupplyLabelNames  = []string{"resource", "chassis_id", "power_supply", "power_supply_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "power_voltage_state", fmt.Sprintf("power voltage state of chassis component,%s", common.StateHelp), ChassisPowerVoltageLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_voltage_volts", "power voltage volts number of chassis component", ChassisPowerVoltageLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_average_consumed_watts", "power wattage watts number of chassis component", ChassisPowerVoltageLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_state", fmt.Sprintf("powersupply state of chassis component,%s", common.StateHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_health", fmt.Sprintf("powersupply health of chassis component,%s", common.HealthHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_efficiency_percentage", "rated efficiency, as a percentage, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_last_power_output_watts", "average power output, measured in Watts, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_input_watts", "measured input power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_output_watts", "measured output power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_powersupply_power_capacity_watts", "power_capacity_watts of powersupply on this chassis", ChassisPowerSupplyLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	metrics map[string]common.Metric
	Log     *slog.Logger
	chassis *schemas.Chassis
}

// NewCollector returns a collector that collects chassis power data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		metrics: metricMap,
		Ctx:     ctx,
		Log: logger.With(
			"collector", Subsystem,
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	wg := &sync.WaitGroup{}

	// chassis.Power() is deprecated since Redfish v1.15.0
	wg.Go(func() { c.collectDeprecatedPower(ch) })

	// chassis.PowerSubsystem() replaces it
	wg.Go(func() { c.collectPowerSubsystem(ch) })

	wg.Wait()
}

// chassis.Power() was deprecated in Redfish v1.15.0, and replaced with chassis.PowerSubsystem()
func (c *Collector) collectDeprecatedPower(ch chan<- prometheus.Metric) {
	powerInfo, err := c.chassis.Power()
	if err != nil {
		c.Log.Error("error getting power data from chassis", "operation", "chassis.Power()", "err", err)
		return
	} else if powerInfo == nil {
		c.Log.Info("no power data found", "operation", "chassis.Power()")
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
}

// PowerSubsystem exposes less information than Power, since much of it is relocated to Sensors or EnvironmentMetrics
func (c *Collector) collectPowerSubsystem(ch chan<- prometheus.Metric) {
	powerSubsystem, err := c.chassis.PowerSubsystem()

	if err != nil {
		c.Log.Error("error getting power subsystem from chassis", "operation", "chassis.PowerSubsystem()", "err", err)
		return
	} else if powerSubsystem == nil {
		c.Log.Info("no power subsystem found", "operation", "chassis.PowerSubsystem()")
		return
	}

	powerSupplies, err := powerSubsystem.PowerSupplies()
	if err != nil {
		c.Log.Error("error getting power supplies from power subsystem", "operation", "powerSubsystem.PowerSupplies()", "err", err)
		return
	} else if powerSupplies == nil {
		c.Log.Info("no power supplies found", "operation", "powerSubsystem.PowerSupplies()")
		return
	}

	wg := &sync.WaitGroup{}

	for _, powerSupply := range powerSupplies {
		wg.Go(func() { parsePowerSupply(ch, c.chassis.ID, *powerSupply) })
	}

	wg.Wait()
}

func parseVoltage(ch chan<- prometheus.Metric, chassisID string, voltageInfo schemas.Voltage) {
	labelValues := []string{"power_voltage", chassisID, voltageInfo.Name, voltageInfo.MemberID}
	if chassisPowerInfoVoltageStateValue, ok := common.ParseCommonStatusState(voltageInfo.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_voltage_state"].Desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, labelValues...)
	}
	if voltageInfo.ReadingVolts != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_voltage_volts"].Desc, prometheus.GaugeValue, float64(*voltageInfo.ReadingVolts), labelValues...)
	}
}

func parsePowerControl(ch chan<- prometheus.Metric, chassisID string, powerControl schemas.PowerControl) {
	labelValues := []string{"power_wattage", chassisID, powerControl.Name, powerControl.MemberID}
	ch <- prometheus.MustNewConstMetric(metricMap["chassis_power_average_consumed_watts"].Desc, prometheus.GaugeValue, float64(*powerControl.PowerMetrics.AverageConsumedWatts), labelValues...)
}

func parsePowerSupply(ch chan<- prometheus.Metric, chassisID string, powerSupply schemas.PowerSupply) {
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
