package system

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

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

// SystemSubsystem is the system subsystem
var (
	SystemSubsystem                   = "system"
	SystemLabelNames                  = []string{"hostname", "resource", "system_id"}
	SystemBIOSLabelNames              = append(slices.Clone(SystemLabelNames), "bios_version")

	SystemLogServiceLabelNames = []string{"system_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	SystemLogEntryLabelNames   = []string{"system_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	systemMetrics = createSystemMetricMap()
)

// SystemCollector implements the prometheus.Collector.
type SystemCollector struct {
	Ctx     *common.CollectionContext
	metrics map[string]common.Metric
	prometheus.Collector
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
}

func createSystemMetricMap() map[string]common.Metric {
	systemMetrics := make(map[string]common.Metric)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "state", fmt.Sprintf("system state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "health_state", fmt.Sprintf("system health,%s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "power_state", "system power state", SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "bios_info", "system BIOS version, value is always 1", SystemBIOSLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_memory_state", fmt.Sprintf("system overall memory state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_memory_health_state", fmt.Sprintf("system overall memory health,%s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_memory_size", "system total memory size, GiB", SystemLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_processor_state", fmt.Sprintf("system overall processor state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_processor_health_state", fmt.Sprintf("system overall processor health,%s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_processor_count", "system total processor count", SystemLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "log_service_state", fmt.Sprintf("system log service state,%s", common.StateHelp), SystemLogServiceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "log_service_health_state", fmt.Sprintf("system log service health state,%s", common.HealthHelp), SystemLogServiceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "log_entry_severity_state", fmt.Sprintf("system log entry severity state,%s", common.SeverityHelp), SystemLogEntryLabelNames)

	return systemMetrics
}

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(ctx *common.CollectionContext, logger *slog.Logger) *SystemCollector {
	return &SystemCollector{
		metrics: systemMetrics,
		Ctx:     ctx,
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
	}
}

// Describe implements prometheus.Collector.
func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.Desc
	}
	s.collectorScrapeStatus.Describe(ch)
}

// Collect implements prometheus.Collector.
func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := s.Log
	disabledMetrics := s.Ctx.DisabledMetrics
	collectLogs := s.Ctx.CollectLogs
	//get service
	service := s.Ctx.RedfishClient.Service

	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		collectorLogContext.Error("error getting systems from service", "operation", "service.Systems()", "err", err)
	} else {
		for _, system := range systems {
			systemLogContext := collectorLogContext.With("System", system.ID)
			systemLogContext.Info("collector scrape started")
			// overall system metrics

			SystemID := system.ID
			systemHostName := system.HostName
			systemPowerState := system.PowerState
			systemState := system.Status.State
			systemHealthState := system.Status.Health
			systemTotalProcessorsState := system.ProcessorSummary.Status.State
			systemTotalProcessorsHealthState := system.ProcessorSummary.Status.Health
			systemTotalMemoryState := system.MemorySummary.Status.State
			systemTotalMemoryHealthState := system.MemorySummary.Status.Health

			systemLabelValues := []string{systemHostName, "system", SystemID}
			if systemHealthStateValue, ok := common.ParseCommonStatusHealth(systemHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].Desc, prometheus.GaugeValue, systemHealthStateValue, systemLabelValues...)
			}
			if systemStateValue, ok := common.ParseCommonStatusState(systemState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].Desc, prometheus.GaugeValue, systemStateValue, systemLabelValues...)
			}
			if systemPowerStateValue, ok := common.ParseCommonPowerState(systemPowerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].Desc, prometheus.GaugeValue, systemPowerStateValue, systemLabelValues...)
			}
			if system.BiosVersion != "" {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_bios_info"].Desc, prometheus.GaugeValue, 1, append(systemLabelValues, system.BiosVersion)...)
			}
			if systemTotalProcessorsStateValue, ok := common.ParseCommonStatusState(systemTotalProcessorsState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].Desc, prometheus.GaugeValue, systemTotalProcessorsStateValue, systemLabelValues...)
				if system.ProcessorSummary.Count != nil {
					ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].Desc, prometheus.GaugeValue, float64(*system.ProcessorSummary.Count), systemLabelValues...)
				}
			}
			if systemTotalProcessorsHealthStateValue, ok := common.ParseCommonStatusHealth(systemTotalProcessorsHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].Desc, prometheus.GaugeValue, systemTotalProcessorsHealthStateValue, systemLabelValues...)
			}
			if systemTotalMemoryStateValue, ok := common.ParseCommonStatusState(systemTotalMemoryState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].Desc, prometheus.GaugeValue, systemTotalMemoryStateValue, systemLabelValues...)
				if system.MemorySummary.TotalSystemMemoryGiB != nil {
					ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].Desc, prometheus.GaugeValue, float64(*system.MemorySummary.TotalSystemMemoryGiB), systemLabelValues...)
				}
			}
			if systemTotalMemoryHealthStateValue, ok := common.ParseCommonStatusHealth(systemTotalMemoryHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].Desc, prometheus.GaugeValue, systemTotalMemoryHealthStateValue, systemLabelValues...)
			}

			wg := &sync.WaitGroup{}

			if !slices.Contains(disabledMetrics, "memory") {
				collector := memory.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "processor") {
				collector := processors.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "storage") {
				collector := storage.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "pcie_devices") {
				collector := pcie_devices.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "network_interfaces") {
				collector := network_interfaces.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "ethernet_interfaces") {
				collector := ethernet_interfaces.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "simple_storage") {
				collector := simple_storage.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "pcie_functions") {
				collector := pcie_functions.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			// process log services
			if collectLogs {
				wg.Go(func() { s.collectLogs(ch, system) })
			}

			wg.Wait()

			systemLogContext.Info("collector scrape completed")
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	}
}

func (c* SystemCollector) collectLogs(ch chan<- prometheus.Metric, system *schemas.ComputerSystem) {
	logServices, err := system.LogServices()
	if err != nil {
		c.Log.Error("error getting log services from system", "operation", "system.LogServices()", "err", err)
		return
	} else if logServices == nil {
		c.Log.Info("no log services found", "operation", "system.LogServices()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, logService := range logServices {
		wg.Go(func() {
			if err = common.ParseLogService(ch, systemMetrics, c.Ctx, c.Log.With("log_service", logService.ID), SystemSubsystem, system.ID, logService); err != nil {
				c.Log.Error("error getting log entries from log service", "operation", "chassis.LogServices()", "err", err)
			}
		})
	}

	wg.Wait()
}
