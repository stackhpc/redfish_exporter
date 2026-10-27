package simple_storage

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

	SystemDeviceLabelNames = []string{"hostname", "resource", "device"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "simple_storage_device_state", fmt.Sprintf("system simple storage device state,%s", common.StateHelp), SystemDeviceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "simple_storage_device_health_state", fmt.Sprintf("system simple storage device health state,%s", common.HealthHelp), SystemDeviceLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	metrics map[string]common.Metric
	Log     *slog.Logger
	system  *schemas.ComputerSystem
}

// NewCollector returns a collector that collects system memory data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, system *schemas.ComputerSystem) *Collector {
	return &Collector{
		metrics: metricMap,
		Ctx:     ctx,
		Log: logger.With(
			"collector", Subsystem,
			"system", system.ID,
		),
		system: system,
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	simpleStorages, err := c.system.SimpleStorage()
	if err != nil {
		c.Log.Error("error getting simple storage data from system", "operation", "system.SimpleStorages()", "err", err)
		return
	} else if simpleStorages == nil {
		c.Log.Info("no simple storage data found", "operation", "system.SimpleStorages()")
		return
	}

	wg := &sync.WaitGroup{}
	processed := make(map[string]bool)
	for _, simpleStorage := range simpleStorages {
		devices := simpleStorage.Devices
		for _, device := range devices {
			_, exists := processed[device.Name]
			if exists {
				c.Log.With("operation",
					"system.SimpleStorages()").Info(fmt.Sprintf("Ignoring "+
					"duplicate storage device: %s. Please check whether this "+
					"device is returning duplicate data and report to the vendor.",
					device.Name))
				continue
			}
			wg.Go(func() { c.parseDevice(ch, device) })
			processed[device.Name] = true
		}
	}
	wg.Wait()
}

func (c *Collector) parseDevice(ch chan<- prometheus.Metric, device schemas.Device) {
	systemDeviceLabelValues := []string{c.system.HostName, "device", device.Name}
	if stateValue, ok := common.ParseCommonStatusState(device.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_simple_storage_device_state"].Desc, prometheus.GaugeValue, stateValue, systemDeviceLabelValues...)
	}
	if healthStatusValue, ok := common.ParseCommonStatusHealth(device.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_simple_storage_device_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, systemDeviceLabelValues...)
	}
}
