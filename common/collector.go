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

func ParseLogService(ch chan<- prometheus.Metric, metrics map[string]Metric, ctx *CollectionContext, logger *slog.Logger, subsystem, collectorID string, logService *schemas.LogService) (err error) {
	logServiceEnabled := fmt.Sprint(logService.ServiceEnabled)
	logServiceOverWritePolicy := string(logService.OverWritePolicy)

	logServiceLabelValues := []string{collectorID, logService.Name, logService.ID, logServiceEnabled, logServiceOverWritePolicy}

	if logServiceStateValue, ok := ParseCommonStatusState(logService.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metrics[fmt.Sprintf("%s_%s", subsystem, "log_service_state")].Desc, prometheus.GaugeValue, logServiceStateValue, logServiceLabelValues...)
	}
	if logServiceHealthStateValue, ok := ParseCommonStatusHealth(logService.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metrics[fmt.Sprintf("%s_%s", subsystem, "log_service_health_state")].Desc, prometheus.GaugeValue, logServiceHealthStateValue, logServiceLabelValues...)
	}
	var (
		logEntries []*schemas.LogEntry
	)
	logCount, ok := ctx.LogCount[logService.ID]
	if !ok {
		logCount, ok = ctx.LogCount["DEFAULT"]
		if !ok {
			logCount = -1
		}
	}

	logger.Info(fmt.Sprintf("logServiceID=%s, logcount=%d", logService.ID, logCount), "operation", "parseLogService")

	if logCount > 0 {
		logEntries, err = logService.FilteredEntries(schemas.WithTop(logCount))
	} else {
		logEntries, err = logService.Entries()
	}
	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}
	processed := make(map[string]bool)
	for _, logEntry := range logEntries {
		_, exists := processed[logEntry.MessageID]
		if exists {
			continue
		}
		wg.Go(func() {
			ParseLogEntry(ch, metrics[fmt.Sprintf("%s_%s", subsystem, "log_entry_severity_state")].Desc, collectorID, logService.Name, logService.ID, logEntry)
		})
		processed[logEntry.MessageID] = true
	}

	wg.Wait()

	return nil
}

func ParseLogEntry(ch chan<- prometheus.Metric, desc *prometheus.Desc, collectorID, logServiceName, logServiceID string, logEntry *schemas.LogEntry) {
	logEntryCode := string(logEntry.EntryCode)
	logEntryType := string(logEntry.EntryType)
	logEntrySensorType := string(logEntry.SensorType)

	sensorNumber := ""
	if logEntry.SensorNumber != nil {
		sensorNumber = fmt.Sprintf("%d", *logEntry.SensorNumber)
	}

	logEntryLabelValues := []string{collectorID, logServiceName, logServiceID, logEntry.Name, logEntry.ID, logEntryCode, logEntryType, logEntry.MessageID, sensorNumber, logEntrySensorType}

	if logEntrySeverityStateValue, ok := ParseCommonSeverityState(logEntry.Severity); ok {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, logEntrySeverityStateValue, logEntryLabelValues...)
	}
}

func BoolToFloat64(data bool) float64 {

	if data {
		return float64(1)
	}
	return float64(0)

}
