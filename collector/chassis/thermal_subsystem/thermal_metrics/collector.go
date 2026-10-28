package thermal_metrics

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "chassis"

	TemperatureLabelNames = []string{"resource", "chassis_id", "sensor", "sensor_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "temperature_celsius", "celsius of temperature on this chassis component", TemperatureLabelNames)

	return metrics
}

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
			"collector", Subsystem,
		),
		thermalSubsystem: thermalSubsystem,
		chassisID:        chassisID,
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
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_temperature_celsius"].Desc, prometheus.GaugeValue, float64(*reading.Reading), chassisTemperatureLabelvalues...)
	}
}
