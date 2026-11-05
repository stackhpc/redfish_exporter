package thermal_subsystem

import (
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem/fans"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem/thermal_metrics"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

type Children struct {
	fans            *fans.Collector
	thermal_metrics *thermal_metrics.Collector
}

func (c *Children) collect(ch chan<- prometheus.Metric) {
	wg := &sync.WaitGroup{}
	wg.Go(func() { c.fans.Collect(ch) })
	wg.Go(func() { c.thermal_metrics.Collect(ch) })
	wg.Wait()
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx      *common.CollectionContext
	Log      *slog.Logger
	chassis  *schemas.Chassis
	children *Children
}

// NewCollector returns a collector that collects chassis thermal data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.thermal_subsystem",
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	fans.Describe(ch)
	thermal_metrics.Describe(ch)
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

// Given a chassis, collect data from its ThermalSubsystem
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	if c.children != nil {
		c.children.collect(ch)
	} else {
		thermalSubsystem, err := c.chassis.ThermalSubsystem()
		if err != nil {
			c.Log.Error("error getting thermal subsystem from chassis", "operation", "chassis.ThermalSubsystem()", "err", err)
			return
		}
		if thermalSubsystem == nil {
			c.Log.Debug("no thermal subsystem found", "operation", "chassis.ThermalSubsystem()")
			return
		}

		c.children = &Children{
			fans:            fans.NewCollector(c.Ctx, c.Log, &c.chassis.ID, thermalSubsystem),
			thermal_metrics: thermal_metrics.NewCollector(c.Ctx, c.Log, &c.chassis.ID, thermalSubsystem),
		}

		c.children.collect(ch)
	}

	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}
