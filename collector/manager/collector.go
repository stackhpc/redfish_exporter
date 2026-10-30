package manager

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

// ManagerSubmanager is the manager subsystem
var (
	ManagerSubmanager         = "manager"
	ManagerLabelNames         = []string{"manager_id", "name", "model", "type"}
	ManagerFirmwareLabelNames = append(slices.Clone(ManagerLabelNames), "firmware_version")

	ManagerLogServiceLabelNames = []string{"manager_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	ManagerLogEntryLabelNames   = []string{"manager_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	metricMap = createMetricMap()
)

// Collector implements the prometheus.Collector.
type Collector struct {
	Ctx                   *common.CollectionContext
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
}

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)
	common.AddToMetricMap(metrics, ManagerSubmanager, "state", fmt.Sprintf("manager state,%s", common.StateHelp), ManagerLabelNames)
	common.AddToMetricMap(metrics, ManagerSubmanager, "health_state", fmt.Sprintf("manager health,%s", common.HealthHelp), ManagerLabelNames)
	common.AddToMetricMap(metrics, ManagerSubmanager, "power_state", "manager power state", ManagerLabelNames)
	common.AddToMetricMap(metrics, ManagerSubmanager, "firmware_info", "manager firmware version, value is always 1", ManagerFirmwareLabelNames)

	common.AddToMetricMap(metrics, ManagerSubmanager, "log_service_state", fmt.Sprintf("manager log service state,%s", common.StateHelp), ManagerLogServiceLabelNames)
	common.AddToMetricMap(metrics, ManagerSubmanager, "log_service_health_state", fmt.Sprintf("manager log service health state,%s", common.HealthHelp), ManagerLogServiceLabelNames)
	common.AddToMetricMap(metrics, ManagerSubmanager, "log_entry_severity_state", fmt.Sprintf("manager log entry severity state,%s", common.SeverityHelp), ManagerLogEntryLabelNames)

	return metrics
}

// NewCollector returns a collector that collecting memory statistics
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger) *Collector {
	return &Collector{
		Ctx: ctx,
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
func (m *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range metricMap {
		ch <- metric.Desc
	}
	m.collectorScrapeStatus.Describe(ch)

}

// Collect implemented prometheus.Collector
func (m *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	//get service
	service := m.Ctx.RedfishClient.Service

	// get a list of managers from service
	managers, err := service.Managers()
	if err != nil {
		m.Log.Error("error getting managers from service", "operation", "service.Managers()", "err", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, manager := range managers {
		wg.Go(func() {
			m.collectManager(ch, manager)
		})
	}

	wg.Wait()
	m.collectorScrapeStatus.WithLabelValues("manager").Set(float64(1))
	m.Log.Info("collection complete", "duration", time.Now().Sub(startTime))
}

func (c *Collector) collectManager(ch chan<- prometheus.Metric, manager *schemas.Manager) {
	logContext := c.Log.With("manager", manager.ID)
	// overall manager metrics
	managerType := fmt.Sprint(manager.ManagerType)

	ManagerLabelValues := []string{manager.ID, manager.Name, manager.Model, managerType}

	if healthStatusValue, ok := common.ParseCommonStatusHealth(manager.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["manager_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, ManagerLabelValues...)
	}
	if stateValue, ok := common.ParseCommonStatusState(manager.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["manager_state"].Desc, prometheus.GaugeValue, stateValue, ManagerLabelValues...)
	}
	if powerStateValue, ok := common.ParseCommonPowerState(manager.PowerState); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["manager_power_state"].Desc, prometheus.GaugeValue, powerStateValue, ManagerLabelValues...)
	}
	if manager.FirmwareVersion != "" {
		ch <- prometheus.MustNewConstMetric(metricMap["manager_firmware_info"].Desc, prometheus.GaugeValue, 1, append(ManagerLabelValues, manager.FirmwareVersion)...)
	}

	// process log services
	if c.Ctx.CollectLogs {
		logServices, err := manager.LogServices()
		if err != nil {
			logContext.Error("error getting log services from manager", "operation", "manager.LogServices()", "err", err)
		} else if logServices == nil {
			logContext.Debug("no log services found", "operation", "manager.LogServices()")
		} else {
			wg := &sync.WaitGroup{}

			for _, logService := range logServices {
				wg.Go(func() {
					if err = common.ParseLogService(ch, metricMap, c.Ctx, logContext, ManagerSubmanager, manager.ID, logService); err != nil {
						logContext.Error("error getting log entries from log service", "operation", "manager.LogServices()", "err", err)
					}
				})
			}

			wg.Wait()
		}
	}
}
