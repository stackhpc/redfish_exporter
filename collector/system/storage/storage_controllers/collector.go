package storage_controllers

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

	SystemStorageControllerLabelNames = []string{"hostname", "resource", "storage_controller", "storage_controller_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "storage_controller_state", fmt.Sprintf("system storage controller state,%s", common.StateHelp), SystemStorageControllerLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "storage_controller_health_state", fmt.Sprintf("system storage controller health state,%s", common.HealthHelp), SystemStorageControllerLabelNames)

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
		Ctx:     ctx,
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
	storageControllers, err := c.storage.Controllers()
	if err != nil {
		c.Log.Info("Errors Getting storagecontrollers from system storage", "err", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, controller := range storageControllers {
		wg.Go(func() { c.parseController(ch, controller)})
	}
	for _, controller := range c.storage.StorageControllers { // Deprecated in Redfish v1.13.0
		wg.Go(func() { c.parseController(ch, &controller)})
	}
	wg.Wait()
}

func (c *Collector) parseController(ch chan<- prometheus.Metric, controller *schemas.StorageController) {
	controllerLabelValues := []string{*c.hostname, "storage_controller", controller.Name, controller.ODataID}
	if stateValue, ok := common.ParseCommonStatusState(controller.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_controller_state"].Desc, prometheus.GaugeValue, stateValue, controllerLabelValues...)
	}
	if healthStatusValue, ok := common.ParseCommonStatusHealth(controller.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_storage_controller_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, controllerLabelValues...)
	}
}
