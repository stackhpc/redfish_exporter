package storage

import (
	"log/slog"
	"slices"
	"sync"
	"time"

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

type Children struct {
	drives              *drives.Collector
	storage_controllers *storage_controllers.Collector
	volumes             *volumes.Collector
}

func (c *Children) collect(ch chan<- prometheus.Metric) {
	wg := &sync.WaitGroup{}
	wg.Go(func() { c.drives.Collect(ch) })
	wg.Go(func() { c.storage_controllers.Collect(ch) })
	wg.Go(func() { c.volumes.Collect(ch) })
	wg.Wait()
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx                *common.CollectionContext
	Log                *slog.Logger
	system             *schemas.ComputerSystem
	children           map[string]Children // Map of Storage ID to children
	collectionMu       sync.Mutex          // Ensure we don't perform volume collection concurrently or we'll trip the deduplicator.
	volumeDeduplicator volumes.VolumeDeduplicator
}

// NewCollector returns a collector that collects system memory data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, system *schemas.ComputerSystem) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "system.storage",
			"system", system.ID,
		),
		system:             system,
		children:           make(map[string]Children),
		volumeDeduplicator: volumes.NewVolumeDeduplicator(),
	}
}

// Add/remove children from new list of Storages. Collect new Storages.
func (c *Collector) updateChildren(ch chan<- prometheus.Metric, storages []*schemas.Storage) {
	updated_storages := make([]string, 0, len(storages))
	wg := &sync.WaitGroup{}
	for _, storage := range storages {
		updated_storages = append(updated_storages, storage.ID)
		_, exists := c.children[storage.ID]
		if !exists {
			child := Children{
				volumes:             volumes.NewCollector(c.Ctx, c.Log, storage, &c.system.HostName, &c.volumeDeduplicator),
				drives:              drives.NewCollector(c.Ctx, c.Log, storage, &c.system.HostName),
				storage_controllers: storage_controllers.NewCollector(c.Ctx, c.Log, storage, &c.system.HostName),
			}
			c.children[storage.ID] = child
			wg.Go(func() { child.collect(ch) })
		}
	}

	// Clear children for Storages that no longer exist
	for ID, _ := range c.children {
		if !slices.Contains(updated_storages, ID) {
			delete(c.children, ID)
		}
	}
	wg.Wait()
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
	startTime := time.Now()

	c.collectionMu.Lock()
	// Clear deduplication state from last run
	c.volumeDeduplicator.Clear()
	wg := &sync.WaitGroup{}
	// Collect from cached children
	for _, child := range c.children {
		wg.Go(func() { child.collect(ch) })
	}

	storages, err := c.system.Storage()
	if err != nil {
		c.Log.Error("error getting storage data from system", "operation", "system.Storage()", "err", err)
		wg.Wait()
		c.collectionMu.Unlock()
		return
	} else if storages == nil {
		c.Log.Debug("no storage data found", "operation", "system.Storage()")
		wg.Wait()
		c.collectionMu.Unlock()
		return
	}

	c.updateChildren(ch, storages)
	wg.Wait()
	c.collectionMu.Unlock()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}
