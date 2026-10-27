package memory

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "system"

	SystemMemoryLabelNames = []string{"hostname", "resource", "memory", "memory_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "memory_state", fmt.Sprintf("system memory state,%s", common.StateHelp), SystemMemoryLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "memory_health_state", fmt.Sprintf("system memory health state,%s", common.HealthHelp), SystemMemoryLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "memory_capacity", "system memory capacity, MiB", SystemMemoryLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	metrics map[string]common.Metric
	Log     *slog.Logger
	system  *schemas.ComputerSystem
}

// NewCollector returns a collector that collects system memory data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, system *schemas.ComputerSystem) *Collector {
	return &Collector{
		metrics: metricMap,
		Ctx:     ctx,
		Log: logger.With(
			"collector", Subsystem,
			"system", system.ID,
		),
		system: system,
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	memories, err := c.system.Memory()
	if err != nil {
		c.Log.Error("error getting memory data from system", "operation", "system.Memory()", "err", err)
		return
	} else if memories == nil {
		c.Log.Info("no memory data found", "operation", "system.Memory()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, memory := range memories {
		wg.Go(func() { parseMemory(ch, c.system.HostName, memory) })
	}
	wg.Wait()
}

func parseMemory(ch chan<- prometheus.Metric, systemHostName string, memory *schemas.Memory) {
	systemMemoryLabelValues := []string{systemHostName, "memory", memory.Name, memory.ID}
	if stateValue, ok := common.ParseCommonStatusState(memory.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_state"].Desc, prometheus.GaugeValue, stateValue, systemMemoryLabelValues...)
	}
	if healthStatusValue, ok := common.ParseCommonStatusHealth(memory.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, systemMemoryLabelValues...)
	}
	if memory.CapacityMiB != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_memory_capacity"].Desc, prometheus.GaugeValue, float64(*memory.CapacityMiB), systemMemoryLabelValues...)
	}
}
