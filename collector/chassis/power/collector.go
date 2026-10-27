package power

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_common"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem                     = "chassis"
	ChassisPowerVoltageLabelNames = []string{"resource", "chassis_id", "power_voltage", "power_voltage_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "power_voltage_state", fmt.Sprintf("power voltage state of chassis component,%s", common.StateHelp), ChassisPowerVoltageLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_voltage_volts", "power voltage volts number of chassis component", ChassisPowerVoltageLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_average_consumed_watts", "power wattage watts number of chassis component", ChassisPowerVoltageLabelNames)

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
		wg.Go(func() { power_common.ParsePowerSupply(ch, c.chassis.ID, powerSupply) })
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
