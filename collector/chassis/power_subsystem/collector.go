package power_subsystem

import (
	"log/slog"
	"sync"

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

// NewCollector returns a collector that collects chassis power data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		Ctx:     ctx,
		Log: logger.With(
			"collector", Subsystem,
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	powerSubsystem, err := c.chassis.PowerSubsystem()

	if err != nil {
		c.Log.Error("error getting power subsystem from chassis", "operation", "chassis.PowerSubsystem()", "err", err)
		return
	} else if powerSubsystem == nil {
		c.Log.Debug("no power subsystem found", "operation", "chassis.PowerSubsystem()")
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
		wg.Go(func() { power_common.ParsePowerSupply(ch, c.chassis.ID, *powerSupply) })
	}

	wg.Wait()
}
