package memory

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "system"

	SystemMemoryLabelNames = []string{"system", "hostname", "resource", "memory", "memory_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "memory_state", fmt.Sprintf("system memory state,%s", common.StateHelp), SystemMemoryLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "memory_health_state", fmt.Sprintf("health of memory itself: %s", common.HealthHelp), SystemMemoryLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "memory_health_rollup", fmt.Sprintf("health of memory, considering dependents: %s", common.HealthHelp), SystemMemoryLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "memory_capacity", "system memory capacity, MiB", SystemMemoryLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx    *common.CollectionContext
	Log    *slog.Logger
	system *schemas.ComputerSystem
}

// NewCollector returns a collector that collects system memory data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, system *schemas.ComputerSystem) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "system.memory",
			"system", system.ID,
		),
		system: system,
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
	startTime := time.Now()
	memories, err := c.system.Memory()
	if err != nil {
		c.Log.Error("error getting memory data from system", "operation", "system.Memory()", "err", err)
		return
	} else if memories == nil {
		c.Log.Debug("no memory data found", "operation", "system.Memory()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, memory := range memories {
		wg.Go(func() { c.parseMemory(ch, memory) })
	}
	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

func (c *Collector) parseMemory(ch chan<- prometheus.Metric, memory *schemas.Memory) {
	systemMemoryLabelValues := []string{c.system.Name, c.system.HostName, "memory", memory.Name, memory.ID}
	if stateValue, ok := common.ParseCommonStatusState(memory.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_state"].Desc, prometheus.GaugeValue, stateValue, systemMemoryLabelValues...)
	}
	if healthStatusValue, ok := common.ParseCommonStatusHealth(memory.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, systemMemoryLabelValues...)
	}
	if healthRollup, ok := common.ParseCommonStatusHealth(memory.Status.HealthRollup); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_health_rollup"].Desc, prometheus.GaugeValue, healthRollup, systemMemoryLabelValues...)
	}

	if memory.CapacityMiB != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_capacity"].Desc, prometheus.GaugeValue, float64(*memory.CapacityMiB), systemMemoryLabelValues...)
	}
}
