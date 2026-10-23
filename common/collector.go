package common

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish/schemas"
)

const (
	StateHelp           = "1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)"
	HealthHelp          = "1(OK),2(Warning),3(Critical)"
	SeverityHelp        = HealthHelp
	LinkHelp            = "1(LinkUp),2(NoLink),3(LinkDown)"
	PortLinkHelp        = "1(Up),0(Down)"
	IntrusionSensorHelp = "1(Normal),2(TamperingDetected),3(HardwareIntrusion)"
)

type Metric struct {
	Desc *prometheus.Desc
}

func AddToMetricMap(metricMap map[string]Metric, subsystem, name, help string, variableLabels []string) {
	metricKey := fmt.Sprintf("%s_%s", subsystem, name)
	metricMap[metricKey] = Metric{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, name),
			help,
			variableLabels,
			nil,
		),
	}
}

func ParseLogService(ch chan<- prometheus.Metric, metrics map[string]Metric, ctx *CollectionContext, logger *slog.Logger, subsystem, collectorID string, logService *schemas.LogService, wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	logServiceName := logService.Name
	logServiceID := logService.ID
	logServiceEnabled := fmt.Sprint(logService.ServiceEnabled)
	logServiceOverWritePolicy := string(logService.OverWritePolicy)
	logServiceState := logService.Status.State
	logServiceHealthState := logService.Status.Health

	logServiceLabelValues := []string{collectorID, logServiceName, logServiceID, logServiceEnabled, logServiceOverWritePolicy}

	if logServiceStateValue, ok := ParseCommonStatusState(logServiceState); ok {
		ch <- prometheus.MustNewConstMetric(metrics[fmt.Sprintf("%s_%s", subsystem, "log_service_state")].Desc, prometheus.GaugeValue, logServiceStateValue, logServiceLabelValues...)
	}
	if logServiceHealthStateValue, ok := ParseCommonStatusHealth(logServiceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(metrics[fmt.Sprintf("%s_%s", subsystem, "log_service_health_state")].Desc, prometheus.GaugeValue, logServiceHealthStateValue, logServiceLabelValues...)
	}
	var (
		logEntries []*schemas.LogEntry
	)
	logCount, ok := ctx.LogCount[logServiceID]
	if !ok {
		logCount, ok = ctx.LogCount["DEFAULT"]
		if !ok {
			logCount = -1
		}
	}

	logger.Info(fmt.Sprintf("logServiceID=%s, logcount=%d", logServiceID, logCount), "operation", "parseLogService")

	if logCount > 0 {
		//logEntries, err = logService.FilteredEntries(common.WithTop(logCount))
	} else {
		logEntries, err = logService.Entries()
	}
	if err != nil {
		return
	}
	wg2 := &sync.WaitGroup{}
	wg2.Add(len(logEntries))
	processed := make(map[string]bool)
	for _, logEntry := range logEntries {
		_, exists := processed[logEntry.MessageID]
		if exists {
			wg2.Done()
			continue
		}
		go ParseLogEntry(ch, metrics[fmt.Sprintf("%s_%s", subsystem, "log_entry_severity_state")].Desc, collectorID, logServiceName, logServiceID, logEntry, wg2)
		processed[logEntry.MessageID] = true
	}
	return
}

func ParseLogEntry(ch chan<- prometheus.Metric, desc *prometheus.Desc, collectorID, logServiceName, logServiceID string, logEntry *schemas.LogEntry, wg *sync.WaitGroup) {
	defer wg.Done()
	logEntryName := logEntry.Name
	logEntryID := logEntry.ID
	logEntryCode := string(logEntry.EntryCode)
	logEntryType := string(logEntry.EntryType)
	logEntryMessageID := logEntry.MessageID
	logEntrySensorNumber := fmt.Sprint(logEntry.SensorNumber)
	logEntrySensorType := string(logEntry.SensorType)
	logEntrySeverityState := logEntry.Severity

	logEntryLabelValues := []string{collectorID, logServiceName, logServiceID, logEntryName, logEntryID, logEntryCode, logEntryType, logEntryMessageID, logEntrySensorNumber, logEntrySensorType}

	if logEntrySeverityStateValue, ok := ParseCommonSeverityState(logEntrySeverityState); ok {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, logEntrySeverityStateValue, logEntryLabelValues...)
	}
}
