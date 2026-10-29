package pcie_devices

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "system"

	SystemPCIeDeviceLabelNames = []string{"hostname", "resource", "pcie_device", "pcie_device_id", "pcie_device_partnumber", "pcie_device_type", "pcie_serial_number"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "pcie_device_state", fmt.Sprintf("system pcie device state,%s", common.StateHelp), SystemPCIeDeviceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "pcie_device_health_state", fmt.Sprintf("system pcie device health state,%s", common.HealthHelp), SystemPCIeDeviceLabelNames)

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
			"collector", "system.pcie_devices",
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
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	startTime := time.Now()
	pcieDevices, err := c.system.PCIeDevices()
	if err != nil {
		c.Log.Error("error getting PCI-E device data from system", "operation", "system.PCIeDevices()", "err", err)
		return
	} else if pcieDevices == nil {
		c.Log.Debug("no PCI-E device data found", "operation", "system.PCIeDevices()")
		return
	}

	processed := make(map[string]bool)
	wg := &sync.WaitGroup{}
	//Some devices are returning duplicated PCIeDevices. This is workaround for this. Example of such data can be found in sampleOut/system_duplicated_devices.json
	for _, pcieDevice := range pcieDevices {
		_, exists := processed[pcieDevice.ODataID]
		if exists {
			c.Log.Info(fmt.Sprintf("Ignoring duplicate pci device: %s", pcieDevice.ODataID), "operation", "system.PCIeDevices()")
			continue
		}
		processed[pcieDevice.ODataID] = true
		wg.Go(func() { c.parsePcieDevice(ch, pcieDevice) })
	}
	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

func (c *Collector) parsePcieDevice(ch chan<- prometheus.Metric, pcieDevice *schemas.PCIeDevice) {
	pcieDeviceType := fmt.Sprint(pcieDevice.DeviceType)
	systemPCIeDeviceLabelValues := []string{c.system.HostName, "pcie_device", pcieDevice.Name, pcieDevice.ID, pcieDevice.PartNumber, pcieDeviceType, pcieDevice.SerialNumber}

	if stateValue, ok := common.ParseCommonStatusState(pcieDevice.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_pcie_device_state"].Desc, prometheus.GaugeValue, stateValue, systemPCIeDeviceLabelValues...)
	}
	if healthStateValue, ok := common.ParseCommonStatusHealth(pcieDevice.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_pcie_device_health_state"].Desc, prometheus.GaugeValue, healthStateValue, systemPCIeDeviceLabelValues...)
	}
}
