package manager

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
)

// ManagerSubmanager is the manager subsystem
var (
	ManagerSubmanager         = "manager"
	ManagerLabelNames         = []string{"manager_id", "name", "model", "type"}
	ManagerFirmwareLabelNames = append(slices.Clone(ManagerLabelNames), "firmware_version")

	ManagerLogServiceLabelNames = []string{"manager_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	ManagerLogEntryLabelNames   = []string{"manager_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	managerMetrics = createManagerMetricMap()
)

// ManagerCollector implements the prometheus.Collector.
type ManagerCollector struct {
	Ctx                   *common.CollectionContext
	metrics               map[string]common.Metric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
}

func createManagerMetricMap() map[string]common.Metric {
	managerMetrics := make(map[string]common.Metric)
	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "state", fmt.Sprintf("manager state,%s", common.StateHelp), ManagerLabelNames)
	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "health_state", fmt.Sprintf("manager health,%s", common.HealthHelp), ManagerLabelNames)
	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "power_state", "manager power state", ManagerLabelNames)
	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "firmware_info", "manager firmware version, value is always 1", ManagerFirmwareLabelNames)

	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "log_service_state", fmt.Sprintf("manager log service state,%s", common.StateHelp), ManagerLogServiceLabelNames)
	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "log_service_health_state", fmt.Sprintf("manager log service health state,%s", common.HealthHelp), ManagerLogServiceLabelNames)
	common.AddToMetricMap(managerMetrics, ManagerSubmanager, "log_entry_severity_state", fmt.Sprintf("manager log entry severity state,%s", common.SeverityHelp), ManagerLogEntryLabelNames)

	return managerMetrics
}

// NewManagerCollector returns a collector that collecting memory statistics
func NewManagerCollector(ctx *common.CollectionContext, logger *slog.Logger) *ManagerCollector {
	return &ManagerCollector{
		metrics: managerMetrics,
		Ctx:     ctx,
		Log: logger.With(
			"collector", "ManagerCollector",
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

// Describe implemented prometheus.Collector
func (m *ManagerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.Desc
	}
	m.collectorScrapeStatus.Describe(ch)

}

// Collect implemented prometheus.Collector
func (m *ManagerCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := m.Log
	collectLogs := m.Ctx.CollectLogs
	//get service
	service := m.Ctx.RedfishClient.Service

	// get a list of managers from service
	if managers, err := service.Managers(); err != nil {
		collectorLogContext.Error("error getting managers from service", "operation", "service.Managers()", "err", err)
	} else {
		for _, manager := range managers {
			managerLogContext := collectorLogContext.With("Manager", manager.ID)
			managerLogContext.Info("collector scrape started")
			// overall manager metrics
			ManagerID := manager.ID
			managerName := manager.Name
			managerModel := manager.Model
			managerType := fmt.Sprint(manager.ManagerType)
			managerPowerState := manager.PowerState
			managerState := manager.Status.State
			managerHealthState := manager.Status.Health

			ManagerLabelValues := []string{ManagerID, managerName, managerModel, managerType}

			if managerHealthStateValue, ok := common.ParseCommonStatusHealth(managerHealthState); ok {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_health_state"].Desc, prometheus.GaugeValue, managerHealthStateValue, ManagerLabelValues...)
			}
			if managerStateValue, ok := common.ParseCommonStatusState(managerState); ok {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_state"].Desc, prometheus.GaugeValue, managerStateValue, ManagerLabelValues...)
			}
			if managerPowerStateValue, ok := common.ParseCommonPowerState(managerPowerState); ok {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_power_state"].Desc, prometheus.GaugeValue, managerPowerStateValue, ManagerLabelValues...)
			}
			if manager.FirmwareVersion != "" {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_firmware_info"].Desc, prometheus.GaugeValue, 1, append(ManagerLabelValues, manager.FirmwareVersion)...)
			}

			// process log services
			if collectLogs {
				logServices, err := manager.LogServices()
				if err != nil {
					managerLogContext.Error("error getting log services from manager", "operation", "manager.LogServices()", "err", err)
				} else if logServices == nil {
					managerLogContext.Info("no log services found", "operation", "manager.LogServices()")
				} else {
					wg := &sync.WaitGroup{}
					wg.Add(len(logServices))

					for _, logService := range logServices {
						if err = common.ParseLogService(ch, managerMetrics, m.Ctx, managerLogContext, ManagerSubmanager, ManagerID, logService); err != nil {
							managerLogContext.Error("error getting log entries from log service", "operation", "manager.LogServices()", "err", err)
						}
					}
				}
			}

			managerLogContext.Info("collector scrape completed")
		}
		m.collectorScrapeStatus.WithLabelValues("manager").Set(float64(1))

	}

}
