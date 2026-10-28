package volumes

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

	SystemVolumeLabelNames = []string{"hostname", "resource", "volume", "volume_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "storage_volume_state", fmt.Sprintf("system storage volume state,%s", common.StateHelp), SystemVolumeLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "storage_volume_health_state", fmt.Sprintf("system storage volume health state,%s", common.HealthHelp), SystemVolumeLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "storage_volume_capacity", "system storage volume capacity, Bytes", SystemVolumeLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx                *common.CollectionContext
	Log                *slog.Logger
	storage            *schemas.Storage
	hostname           *string
	volumeDeduplicator *VolumeDeduplicator
}

// NewCollector returns a collector that collects storage volume data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, storage *schemas.Storage, hostname *string, volumeDeduplicator *VolumeDeduplicator) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", Subsystem,
			"storage", storage.ID,
		),
		storage:            storage,
		hostname:           hostname,
		volumeDeduplicator: volumeDeduplicator,
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
	volumes, err := c.storage.Volumes()
	if err != nil {
		c.Log.Error("error getting storage data from system", "operation", "system.Volumes()", "err", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, volume := range volumes {
		if c.volumeDeduplicator.SetVolumeIsSeen(volume) {
			wg.Go(func() { parseVolume(ch, *c.hostname, volume) })
		} else {
			c.Log.With("operation",
				"system.Storage()").Info(fmt.Sprintf("Ignoring "+
				"duplicate storage volume: %s. Please check whether this "+
				"volume is returning duplicate data and report to the vendor.",
				volume.Name))
		}
	}

	wg.Wait()
}

func parseVolume(ch chan<- prometheus.Metric, systemHostName string, volume *schemas.Volume) {
	systemVolumeLabelValues := []string{systemHostName, "volume", volume.Name, volume.ID}
	if volumeStateValue, ok := common.ParseCommonStatusState(volume.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_volume_state"].Desc, prometheus.GaugeValue, volumeStateValue, systemVolumeLabelValues...)
	}
	if volumeHealthStateValue, ok := common.ParseCommonStatusHealth(volume.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_volume_health_state"].Desc, prometheus.GaugeValue, volumeHealthStateValue, systemVolumeLabelValues...)
	}
	if volume.CapacityBytes != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_volume_capacity"].Desc, prometheus.GaugeValue, float64(*volume.CapacityBytes), systemVolumeLabelValues...)
	}
}

type VolumeDeduplicator struct {
	seenVolumes map[string]any
	mu          sync.Mutex
}

// Returns whether the volume was previously unseen
func (d *VolumeDeduplicator) SetVolumeIsSeen(volume *schemas.Volume) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, exists := d.seenVolumes[volume.ID]

	if !exists {
		d.seenVolumes[volume.ID] = struct{}{}
		return true
	}

	return false
}

func NewVolumeDeduplicator() VolumeDeduplicator {
	return VolumeDeduplicator{
		seenVolumes: make(map[string]any),
	}
}
