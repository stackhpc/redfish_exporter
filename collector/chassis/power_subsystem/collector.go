package power_subsystem

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_common"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_subsystem/power_supply_unit"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "chassis"
)

// Implements the prometheus.Collector.
type Collector struct {
	Ctx      *common.CollectionContext
	Log      *slog.Logger
	chassis  *schemas.Chassis
	children *power_supply_unit.Collector
}

// NewCollector returns a collector that collects chassis power data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.power_subsystem",
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	power_common.Describe(ch)
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()

	if c.children != nil {
		c.children.Collect(ch)
	} else {
		powerSubsystem, err := c.chassis.PowerSubsystem()

		if err != nil {
			c.Log.Error("error getting power subsystem from chassis", "operation", "chassis.PowerSubsystem()", "err", err)
			return
		} else if powerSubsystem == nil {
			c.Log.Debug("no power subsystem found", "operation", "chassis.PowerSubsystem()")
			return
		}

		c.children = power_supply_unit.NewCollector(c.Ctx, c.Log, powerSubsystem, &c.chassis.ID)
		c.children.Collect(ch)
	}

	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}
