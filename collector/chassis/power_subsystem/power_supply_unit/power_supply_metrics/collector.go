package power_supply_metrics

import (
	"log/slog"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_common"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "chassis"
)

// Implements the prometheus.Collector.
type Collector struct {
	Ctx             *common.CollectionContext
	Log             *slog.Logger
	powerSupplyUnit *schemas.PowerSupplyUnit
	chassisID       *string
}

// NewCollector returns a collector that collects power supply unit metrics
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, powerSupplyUnit *schemas.PowerSupplyUnit, chassisID *string) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.power_subsystem.power_supply_unit.power_supply_metrics",
		),
		powerSupplyUnit: powerSupplyUnit,
		chassisID:       chassisID,
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
	powerSupplyUnitMetrics, err := c.powerSupplyUnit.Metrics()

	if err != nil {
		c.Log.Error("error getting power supply unit metrics from power supply unit", "operation", "powerSupplyUnit.Metrics()", "err", err)
		return
	} else if powerSupplyUnitMetrics == nil {
		c.Log.Debug("no power supply unit metrics found", "operation", "powerSupplyUnit.Metrics()")
		return
	}

	labelValues := []string{"power_supply", *c.chassisID, c.powerSupplyUnit.Name, c.powerSupplyUnit.ID}

	if powerSupplyUnitMetrics.InputPowerWatts.Reading != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_input_watts"].Desc, prometheus.GaugeValue, *powerSupplyUnitMetrics.InputPowerWatts.Reading, labelValues...)
	}

	if powerSupplyUnitMetrics.OutputPowerWatts.Reading != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_output_watts"].Desc, prometheus.GaugeValue, *powerSupplyUnitMetrics.OutputPowerWatts.Reading, labelValues...)
	}

	if powerSupplyUnitMetrics.PowerLoadPercent.Reading != nil {
		loadPercent := *powerSupplyUnitMetrics.PowerLoadPercent.Reading
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_load_percent"].Desc, prometheus.GaugeValue, loadPercent, labelValues...)

		// Find efficiency rating that is closest to our current load percent
		var currentEfficiency *float64
		lowestDistanceToMeasurement := 100.0
		for _, efficiency := range c.powerSupplyUnit.EfficiencyRatings {
			distanceToMeasurement := math.Abs(loadPercent - *efficiency.LoadPercent)
			if distanceToMeasurement < lowestDistanceToMeasurement {
				lowestDistanceToMeasurement = distanceToMeasurement
				currentEfficiency = efficiency.EfficiencyPercent
			}
		}

		if currentEfficiency != nil {
			ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_efficiency_percentage"].Desc, prometheus.GaugeValue, *currentEfficiency, labelValues...)
		}
	}

	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}
