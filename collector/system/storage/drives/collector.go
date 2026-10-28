package drives

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

	SystemDriveLabelNames = []string{"hostname", "resource", "drive", "drive_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "storage_drive_state", fmt.Sprintf("system storage drive state,%s", common.StateHelp), SystemDriveLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "storage_drive_health_state", fmt.Sprintf("system storage drive health state,%s", common.HealthHelp), SystemDriveLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "storage_drive_capacity", "system storage drive capacity, Bytes", SystemDriveLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx      *common.CollectionContext
	Log      *slog.Logger
	storage  *schemas.Storage
	hostname *string
}

// NewCollector returns a collector that collects storage volume data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, storage *schemas.Storage, hostname *string) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", Subsystem,
			"storage", storage.ID,
		),
		storage:  storage,
		hostname: hostname,
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
	drives, err := c.storage.Drives()
	if err != nil {
		c.Log.Error("error getting drive data from system", "operation", "system.Drives()", "err", err)
		return
	} else if drives == nil {
		c.Log.Debug("no drive data found", "operation", "system.Drives()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, drive := range drives {
		wg.Go(func() { parseDrive(ch, *c.hostname, drive) })
	}
	wg.Wait()
}

func parseDrive(ch chan<- prometheus.Metric, systemHostName string, drive *schemas.Drive) {
	systemdriveLabelValues := []string{systemHostName, "drive", drive.Name, drive.ID}
	if driveStateValue, ok := common.ParseCommonStatusState(drive.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_drive_state"].Desc, prometheus.GaugeValue, driveStateValue, systemdriveLabelValues...)
	}
	if driveHealthStateValue, ok := common.ParseCommonStatusHealth(drive.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_drive_health_state"].Desc, prometheus.GaugeValue, driveHealthStateValue, systemdriveLabelValues...)
	}
	if drive.CapacityBytes != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_drive_capacity"].Desc, prometheus.GaugeValue, float64(*drive.CapacityBytes), systemdriveLabelValues...)
	}
}
