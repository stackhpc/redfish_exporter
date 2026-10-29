package processors

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

	SystemProcessorLabelNames = []string{"hostname", "resource", "processor", "processor_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "processor_state", fmt.Sprintf("system processor state,%s", common.StateHelp), SystemProcessorLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "processor_health_state", fmt.Sprintf("system processor health state,%s", common.HealthHelp), SystemProcessorLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "processor_total_threads", "system processor total threads", SystemProcessorLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "processor_total_threads_enabled", "system processor total enabled threads", SystemProcessorLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "processor_total_cores", "system processor total cores", SystemProcessorLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "processor_total_cores_enabled", "system processor total enabled cores", SystemProcessorLabelNames)

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
			"collector", "system.processors",
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
	processors, err := c.system.Processors()
	if err != nil {
		c.Log.Error("error getting processor data from system", "operation", "system.Processors()", "err", err)
		return
	} else if processors == nil {
		c.Log.Debug("no processor data found", "operation", "system.Processors()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, processor := range processors {
		wg.Go(func() { parseProcessor(ch, c.system.HostName, processor) })
	}
	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

func parseProcessor(ch chan<- prometheus.Metric, systemHostName string, processor *schemas.Processor) {
	systemProcessorLabelValues := []string{systemHostName, "processor", processor.Name, processor.ID}

	if processorStateValue, ok := common.ParseCommonStatusState(processor.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_processor_state"].Desc, prometheus.GaugeValue, processorStateValue, systemProcessorLabelValues...)
	}
	if processorHealthStateValue, ok := common.ParseCommonStatusHealth(processor.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_processor_health_state"].Desc, prometheus.GaugeValue, processorHealthStateValue, systemProcessorLabelValues...)
	}
	if processor.TotalThreads != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_processor_total_threads"].Desc, prometheus.GaugeValue, float64(*processor.TotalThreads), systemProcessorLabelValues...)
	}
	if processor.TotalEnabledThreads != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_processor_total_threads_enabled"].Desc, prometheus.GaugeValue, float64(*processor.TotalThreads), systemProcessorLabelValues...)
	}
	if processor.TotalCores != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_processor_total_cores"].Desc, prometheus.GaugeValue, float64(*processor.TotalCores), systemProcessorLabelValues...)
	}
	if processor.TotalEnabledCores != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_processor_total_cores_enabled"].Desc, prometheus.GaugeValue, float64(*processor.TotalCores), systemProcessorLabelValues...)
	}
}
