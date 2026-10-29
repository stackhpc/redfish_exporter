package thermal_metrics

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_common"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

// Implements the prometheus.Collector.
type Collector struct {
	Ctx              *common.CollectionContext
	Log              *slog.Logger
	thermalSubsystem *schemas.ThermalSubsystem
	chassisID        *string
}

// NewCollector returns a collector that collects thermal subsystem fan data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassisID *string, thermalSubsystem *schemas.ThermalSubsystem) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.thermal_subsystem.thermal_metrics",
		),
		thermalSubsystem: thermalSubsystem,
		chassisID:        chassisID,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	thermal_common.Describe(ch)
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

// Given a ThermalSubsystem, collect thermal metric data
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	thermalMetrics, err := c.thermalSubsystem.ThermalMetrics()
	if err != nil {
		c.Log.Error("Error getting thermal metrics from thermal subsystem", "id", c.thermalSubsystem.ID, "err", err)
		return
	}

	wg := &sync.WaitGroup{}

	for _, reading := range thermalMetrics.TemperatureReadingsCelsius {
		wg.Go(func() { c.parseTemperatureReading(ch, reading) })
	}

	wg.Wait()
}

func (c *Collector) parseTemperatureReading(ch chan<- prometheus.Metric, reading schemas.SensorArrayExcerpt) {
	chassisTemperatureLabelvalues := []string{"temperature", *c.chassisID, reading.DeviceName, reading.DataSourceURI}
	if reading.Reading != nil {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_temperature_celsius"].Desc, prometheus.GaugeValue, float64(*reading.Reading), chassisTemperatureLabelvalues...)
	}
}
