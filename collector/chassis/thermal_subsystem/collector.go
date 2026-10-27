package thermal_subsystem

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem/fans"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem/thermal_metrics"
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
	metrics map[string]common.Metric
	Log     *slog.Logger
	chassis *schemas.Chassis
}

// NewCollector returns a collector that collects chassis thermal data
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

// Given a chassis, collect data from its ThermalSubsystem
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	thermalSubsystem, err := c.chassis.ThermalSubsystem()
	if err != nil {
		c.Log.Error("error getting thermal subsystem from chassis", "operation", "chassis.ThermalSubsystem()", "err", err)
		return
	}
	if thermalSubsystem == nil {
		c.Log.Debug("no thermal subsystem found", "operation", "chassis.ThermalSubsystem()")
		return
	}

	wg := &sync.WaitGroup{}

	thermalMetricsCollector := thermal_metrics.NewCollector(c.Ctx, c.Log, &c.chassis.ID, thermalSubsystem)
	wg.Go(func() { thermalMetricsCollector.Collect(ch) })

	fanCollector := fans.NewCollector(c.Ctx, c.Log, &c.chassis.ID, thermalSubsystem)
	wg.Go(func() { fanCollector.Collect(ch) })

	wg.Wait()
}
