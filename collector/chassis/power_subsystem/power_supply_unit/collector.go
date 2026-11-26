package power_supply_unit

import (
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_common"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_subsystem/power_supply_unit/power_supply_metrics"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "chassis"
)

type Children struct {
	powerSupplyMetrics *power_supply_metrics.Collector
}

func (c *Children) collect(ch chan<- prometheus.Metric) {
	c.powerSupplyMetrics.Collect(ch)
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx            *common.CollectionContext
	Log            *slog.Logger
	powerSubsystem *schemas.PowerSubsystem
	// Map of power supply unit IDs to Children
	children  map[string]Children
	chassisID *string
}

// NewCollector returns a collector that collects data on all the power supply units belonging to a power subsystem
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, powerSubsystem *schemas.PowerSubsystem, chassisID *string) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.power_subsystem.power_supply_unit",
		),
		powerSubsystem: powerSubsystem,
		children:       make(map[string]Children),
		chassisID:      chassisID,
	}
}

// Add/remove children based on new list of PowerSupplyUnits. Collect new PowerSupplyUnits.
func (c *Collector) updateChildren(ch chan<- prometheus.Metric, powerSupplyUnits []*schemas.PowerSupplyUnit) {
	updated_powerSupplyUnits := make([]string, 0, len(powerSupplyUnits))
	wg := &sync.WaitGroup{}
	for _, powerSupplyUnit := range powerSupplyUnits {
		updated_powerSupplyUnits = append(updated_powerSupplyUnits, powerSupplyUnit.ID)
		_, exists := c.children[powerSupplyUnit.ID]
		if !exists {
			child := Children{
				powerSupplyMetrics: power_supply_metrics.NewCollector(c.Ctx, c.Log, powerSupplyUnit, c.chassisID),
			}
			c.children[powerSupplyUnit.ID] = child
			wg.Go(func() { child.collect(ch) })
		}
	}

	// Clear children of PowerSupplyUnits that no longer exist
	for ID, _ := range c.children {
		if !slices.Contains(updated_powerSupplyUnits, ID) {
			delete(c.children, ID)
		}
	}
	wg.Wait()
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

	wg := &sync.WaitGroup{}
	// Collect from cached children
	for _, child := range c.children {
		wg.Go(func() { child.collect(ch) })
	}

	powerSupplyUnits, err := c.powerSubsystem.PowerSupplies()
	if err != nil {
		c.Log.Error("error getting power supplies from power subsystem", "operation", "powerSubsystem.PowerSupplies()", "err", err)
		wg.Wait()
		return
	} else if powerSupplyUnits == nil {
		c.Log.Debug("no power supplies found in power subsystem", "operation", "powerSubsystem.PowerSupples()")
		wg.Wait()
		return
	}

	for _, powerSupplyUnit := range powerSupplyUnits {
		wg.Go(func() { c.collectPowerSupplyUnit(ch, powerSupplyUnit) })
	}

	c.updateChildren(ch, powerSupplyUnits)
	wg.Wait()

	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

// Collect metrics for the PowerSupplyUnit itself
func (c *Collector) collectPowerSupplyUnit(ch chan<- prometheus.Metric, powerSupplyUnit *schemas.PowerSupplyUnit) {
	labelValues := []string{"PowerSupply", *c.chassisID, powerSupplyUnit.Name, powerSupplyUnit.ID}
	if chassisPowerInfoPowerSupplyStateValue, ok := common.ParseCommonStatusState(powerSupplyUnit.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_state"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, labelValues...)
	}
	if chassisPowerInfoPowerSupplyHealthValue, ok := common.ParseCommonStatusHealth(powerSupplyUnit.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_health"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, labelValues...)
	}
	if healthRollup, ok := common.ParseCommonStatusHealth(powerSupplyUnit.Status.HealthRollup); ok {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_health_rollup"].Desc, prometheus.GaugeValue, healthRollup, labelValues...)
	}

	if powerSupplyUnit.PowerCapacityWatts != nil {
		ch <- prometheus.MustNewConstMetric(power_common.MetricMap["chassis_power_powersupply_power_capacity_watts"].Desc, prometheus.GaugeValue, float64(*powerSupplyUnit.PowerCapacityWatts), labelValues...)
	}
}
