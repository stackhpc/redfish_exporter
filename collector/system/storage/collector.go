package storage

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/system/storage/drives"
	"github.com/stackhpc/redfish_exporter/collector/system/storage/storage_controllers"
	"github.com/stackhpc/redfish_exporter/collector/system/storage/volumes"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "system"

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

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
			"collector", Subsystem,
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
	drives.Describe(ch)
	storage_controllers.Describe(ch)
	volumes.Describe(ch)
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	storages, err := c.system.Storage()
	if err != nil {
		c.Log.Error("error getting storage data from system", "operation", "system.Storage()", "err", err)
		return
	} else if storages == nil {
		c.Log.Debug("no storage data found", "operation", "system.Storage()")
		return
	}

	wg := &sync.WaitGroup{}
	volumeDeduplicator := volumes.NewVolumeDeduplicator()
	for _, storage := range storages {
		volumes := volumes.NewCollector(c.Ctx, c.Log, storage, &c.system.HostName, &volumeDeduplicator)
		wg.Go(func() { volumes.Collect(ch) })

		drives := drives.NewCollector(c.Ctx, c.Log, storage, &c.system.HostName)
		wg.Go(func() { drives.Collect(ch) })

		storageControllers := storage_controllers.NewCollector(c.Ctx, c.Log, storage, &c.system.HostName)
		wg.Go(func() { storageControllers.Collect(ch) })
	}
	wg.Wait()
}
