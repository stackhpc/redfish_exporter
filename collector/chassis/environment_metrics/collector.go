package environment_metrics

import (
	"log/slog"

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

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	Log     *slog.Logger
	chassis *schemas.Chassis
}

// NewCollector returns a collector that collects chassis environment metrics
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.environment_metrics",
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
	environmentMetrics, err := c.chassis.EnvironmentMetrics()
	if err != nil {
		c.Log.Error("error getting environment metrics from chassis", "operation", "chassis.EnvironmentMetrics()", "err", err)
		return
	}
	if environmentMetrics == nil {
		return
	}

	if environmentMetrics.PowerWatts.Reading != nil {
		labelValues := []string{"power_wattage", c.chassis.ID, environmentMetrics.Name, environmentMetrics.ID}
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_average_consumed_watts"].Desc, prometheus.GaugeValue, *environmentMetrics.PowerWatts.Reading, labelValues...)
	}

	if environmentMetrics.Voltage.Reading != nil {
		labelValues := []string{"power_voltage", c.chassis.ID, environmentMetrics.Name, environmentMetrics.ID}
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_voltage_volts"].Desc, prometheus.GaugeValue, *environmentMetrics.Voltage.Reading, labelValues...)
	}
}
