package system

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/system/ethernet_interfaces"
	"github.com/stackhpc/redfish_exporter/collector/system/memory"
	"github.com/stackhpc/redfish_exporter/collector/system/network_interfaces"
	"github.com/stackhpc/redfish_exporter/collector/system/pcie_devices"
	"github.com/stackhpc/redfish_exporter/collector/system/pcie_functions"
	"github.com/stackhpc/redfish_exporter/collector/system/processors"
	"github.com/stackhpc/redfish_exporter/collector/system/simple_storage"
	"github.com/stackhpc/redfish_exporter/collector/system/storage"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem            = "system"
	SystemLabelNames     = []string{"hostname", "resource", "system_id"}
	SystemBIOSLabelNames = append(slices.Clone(SystemLabelNames), "bios_version")

	SystemLogServiceLabelNames = []string{"system_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	SystemLogEntryLabelNames   = []string{"system_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "state", fmt.Sprintf("system state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "health_state", fmt.Sprintf("health of system itself: %s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "health_rollup", fmt.Sprintf("health of system, considering dependents: %s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "power_state", "system power state", SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "bios_info", "system BIOS version, value is always 1", SystemBIOSLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "total_memory_state", fmt.Sprintf("system overall memory state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "total_memory_health_state", fmt.Sprintf("health of all memory itself: %s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "total_memory_health_rollup", fmt.Sprintf("health of all memory, considering dependents: %s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "total_memory_size", "system total memory size, GiB", SystemLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "total_processor_state", fmt.Sprintf("system overall processor state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "total_processor_health_state", fmt.Sprintf("health of processors themselves overall: %s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "total_processor_health_rollup", fmt.Sprintf("health of processors, considering dependents: %s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "total_processor_count", "system total processor count", SystemLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "log_service_state", fmt.Sprintf("system log service state,%s", common.StateHelp), SystemLogServiceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "log_service_health_state", fmt.Sprintf("system log service health state,%s", common.HealthHelp), SystemLogServiceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "log_entry_severity_state", fmt.Sprintf("system log entry severity state,%s", common.SeverityHelp), SystemLogEntryLabelNames)

	return metrics
}

type Children struct {
	ethernet_interfaces *ethernet_interfaces.Collector
	memory              *memory.Collector
	network_interfaces  *network_interfaces.Collector
	pcie_devices        *pcie_devices.Collector
	pcie_functions      *pcie_functions.Collector
	processors          *processors.Collector
	simple_storage      *simple_storage.Collector
	storage             *storage.Collector
}

func (c *Children) collect(ch chan<- prometheus.Metric, disabledMetrics []string) {
	wg := &sync.WaitGroup{}
	if !slices.Contains(disabledMetrics, "ethernet_interfaces") {
		wg.Go(func() { c.ethernet_interfaces.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "memory") {
		wg.Go(func() { c.memory.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "network_interfaces") {
		wg.Go(func() { c.network_interfaces.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "pcie_devices") {
		wg.Go(func() { c.pcie_devices.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "pcie_functions") {
		wg.Go(func() { c.pcie_functions.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "processor") {
		wg.Go(func() { c.processors.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "simple_storage") {
		wg.Go(func() { c.simple_storage.Collect(ch) })
	}
	if !slices.Contains(disabledMetrics, "storage") {
		wg.Go(func() { c.storage.Collect(ch) })
	}
	wg.Wait()
}

// Collector implements the prometheus.Collector.
type Collector struct {
	Ctx *common.CollectionContext
	prometheus.Collector
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
	children              map[string]Children // Map of System ID to children
}

// NewCollector returns a collector that collecting memory statistics
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "SystemCollector",
		),
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: common.Namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
		children: make(map[string]Children),
	}
}

// Add/remove children from new list of systems. Collect new systems.
func (c *Collector) updateChildren(ch chan<- prometheus.Metric, systems []*schemas.ComputerSystem) {
	updated_systems := make([]string, 0, len(systems))
	wg := &sync.WaitGroup{}
	for _, system := range systems {
		updated_systems = append(updated_systems, system.ID)
		_, exists := c.children[system.ID]
		if !exists {
			systemLogContext := c.Log.With("system", system.ID)
			child := Children{
				ethernet_interfaces: ethernet_interfaces.NewCollector(c.Ctx, systemLogContext, system),
				memory:              memory.NewCollector(c.Ctx, systemLogContext, system),
				network_interfaces:  network_interfaces.NewCollector(c.Ctx, systemLogContext, system),
				pcie_devices:        pcie_devices.NewCollector(c.Ctx, systemLogContext, system),
				pcie_functions:      pcie_functions.NewCollector(c.Ctx, systemLogContext, system),
				processors:          processors.NewCollector(c.Ctx, systemLogContext, system),
				simple_storage:      simple_storage.NewCollector(c.Ctx, systemLogContext, system),
				storage:             storage.NewCollector(c.Ctx, systemLogContext, system),
			}
			c.children[system.ID] = child
			wg.Go(func() { child.collect(ch, c.Ctx.DisabledMetrics) })
		}
	}

	// Clear children for Systems that no longer exist
	for ID, _ := range c.children {
		if !slices.Contains(updated_systems, ID) {
			delete(c.children, ID)
		}
	}
	wg.Wait()
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
	c.collectorScrapeStatus.Describe(ch)
	ethernet_interfaces.Describe(ch)
	memory.Describe(ch)
	network_interfaces.Describe(ch)
	pcie_devices.Describe(ch)
	pcie_functions.Describe(ch)
	processors.Describe(ch)
	simple_storage.Describe(ch)
	storage.Describe(ch)
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	service := c.Ctx.RedfishClient.Service

	wg := &sync.WaitGroup{}
	// Collect from cached children
	for _, child := range c.children {
		wg.Go(func() { child.collect(ch, c.Ctx.DisabledMetrics) })
	}

	systems, err := service.Systems()
	if err != nil {
		c.Log.Error("error getting systems from service", "operation", "service.Systems()", "err", err)
		wg.Wait()
		return
	}

	for _, system := range systems {
		wg.Go(func() { c.collectOverallMetrics(ch, system) })

		// process log services
		if c.Ctx.CollectLogs {
			wg.Go(func() { c.collectLogs(ch, system) })
		}
	}

	// Update cached children
	c.updateChildren(ch, systems)

	wg.Wait()

	c.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	c.Log.Info("collection complete", "duration", time.Now().Sub(startTime))
}

func (c *Collector) collectOverallMetrics(ch chan<- prometheus.Metric, system *schemas.ComputerSystem) {
	systemLabelValues := []string{system.HostName, "system", system.ID}
	if healthStatusValue, ok := common.ParseCommonStatusHealth(system.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, systemLabelValues...)
	}
	if healthRollup, ok := common.ParseCommonStatusHealth(system.Status.HealthRollup); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_health_rollup"].Desc, prometheus.GaugeValue, healthRollup, systemLabelValues...)
	}
	if stateValue, ok := common.ParseCommonStatusState(system.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_state"].Desc, prometheus.GaugeValue, stateValue, systemLabelValues...)
	}

	if powerStateValue, ok := common.ParseCommonPowerState(system.PowerState); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_power_state"].Desc, prometheus.GaugeValue, powerStateValue, systemLabelValues...)
	}
	if system.BiosVersion != "" {
		ch <- prometheus.MustNewConstMetric(metricMap["system_bios_info"].Desc, prometheus.GaugeValue, 1, append(systemLabelValues, system.BiosVersion)...)
	}

	if totalProcessorsStateValue, ok := common.ParseCommonStatusState(system.ProcessorSummary.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_total_processor_state"].Desc, prometheus.GaugeValue, totalProcessorsStateValue, systemLabelValues...)
		if system.ProcessorSummary.Count != nil {
			ch <- prometheus.MustNewConstMetric(metricMap["system_total_processor_count"].Desc, prometheus.GaugeValue, float64(*system.ProcessorSummary.Count), systemLabelValues...)
		}
	}
	if totalProcessorsHealthStatusValue, ok := common.ParseCommonStatusHealth(system.ProcessorSummary.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_total_processor_health_state"].Desc, prometheus.GaugeValue, totalProcessorsHealthStatusValue, systemLabelValues...)
	}
	if healthRollup, ok := common.ParseCommonStatusHealth(system.ProcessorSummary.Status.HealthRollup); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_total_processor_health_rollup"].Desc, prometheus.GaugeValue, healthRollup, systemLabelValues...)
	}

	if totalMemoryStateValue, ok := common.ParseCommonStatusState(system.MemorySummary.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_total_memory_state"].Desc, prometheus.GaugeValue, totalMemoryStateValue, systemLabelValues...)
		if system.MemorySummary.TotalSystemMemoryGiB != nil {
			ch <- prometheus.MustNewConstMetric(metricMap["system_total_memory_size"].Desc, prometheus.GaugeValue, float64(*system.MemorySummary.TotalSystemMemoryGiB), systemLabelValues...)
		}
	}
	if totalMemoryHealthStatusValue, ok := common.ParseCommonStatusHealth(system.MemorySummary.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_total_memory_health_state"].Desc, prometheus.GaugeValue, totalMemoryHealthStatusValue, systemLabelValues...)
	}
	if healthRollup, ok := common.ParseCommonStatusHealth(system.MemorySummary.Status.HealthRollup); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_total_memory_health_rollup"].Desc, prometheus.GaugeValue, healthRollup, systemLabelValues...)
	}
}

func (c *Collector) collectLogs(ch chan<- prometheus.Metric, system *schemas.ComputerSystem) {
	logServices, err := system.LogServices()
	if err != nil {
		c.Log.Error("error getting log services from system", "operation", "system.LogServices()", "err", err)
		return
	} else if logServices == nil {
		c.Log.Debug("no log services found", "operation", "system.LogServices()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, logService := range logServices {
		wg.Go(func() {
			if err = common.ParseLogService(ch, metricMap, c.Ctx, c.Log.With("log_service", logService.ID), Subsystem, system.ID, logService); err != nil {
				c.Log.Error("error getting log entries from log service", "operation", "chassis.LogServices()", "err", err)
			}
		})
	}

	wg.Wait()
}
