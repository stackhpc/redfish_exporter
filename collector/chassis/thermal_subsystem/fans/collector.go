package fans

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_common"
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
	Ctx              *common.CollectionContext
	metrics          map[string]common.Metric
	Log              *slog.Logger
	thermalSubsystem *schemas.ThermalSubsystem
	chassisID        *string
}

// NewCollector returns a collector that collects thermal subsystem fan data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassisID *string, thermalSubsystem *schemas.ThermalSubsystem) *Collector {
	return &Collector{
		metrics: metricMap,
		Ctx:     ctx,
		Log: logger.With(
			"collector", Subsystem,
		),
		thermalSubsystem: thermalSubsystem,
		chassisID: chassisID,
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}

// Given a ThermalSubsystem, collect data from its Fans
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	chassisFans, err := c.thermalSubsystem.Fans()
	if err != nil {
		c.Log.Error("Error getting fans from thermal subsystem", "id", c.thermalSubsystem.ID, "err", err)
		return
	}

	wg := sync.WaitGroup{}
	for _, chassisFan := range chassisFans {
		wg.Go(func() { thermal_common.ParseFan(ch, c.chassisID, chassisFan) })
	}

	wg.Wait()
}
