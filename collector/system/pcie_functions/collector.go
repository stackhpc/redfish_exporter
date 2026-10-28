package pcie_functions

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

	SystemPCIeFunctionLabelNames = []string{"hostname", "resource", "pcie_function_name", "pcie_function_id", "pci_function_deviceclass", "pci_function_type"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "pcie_function_state", fmt.Sprintf("system pcie function state,%s", common.StateHelp), SystemPCIeFunctionLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "pcie_function_health_state", fmt.Sprintf("system pcie device function state,%s", common.HealthHelp), SystemPCIeFunctionLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	Log     *slog.Logger
	system  *schemas.ComputerSystem
}

// NewCollector returns a collector that collects system memory data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, system *schemas.ComputerSystem) *Collector {
	return &Collector{
		Ctx:     ctx,
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
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	pcieFunctions, err := c.system.PCIeFunctions()
	if err != nil {
		c.Log.Error("error getting PCI-E device function data from system", "operation", "system.PCIEFunctions()", "err", err)
		return
	} else if pcieFunctions == nil {
		c.Log.Debug("no PCI-E device function data found", "operation", "system.PCIeFunctions()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, pcieFunction := range pcieFunctions {
		wg.Go(func() { c.parsePcieFunction(ch, pcieFunction) })
	}
	wg.Wait()
}

func (c *Collector) parsePcieFunction(ch chan<- prometheus.Metric, pcieFunction *schemas.PCIeFunction) {
	pciFunctionDeviceclass := fmt.Sprint(pcieFunction.DeviceClass)
	pciFunctionType := fmt.Sprint(pcieFunction.FunctionType)

	systemPCIeFunctionLabelLabelValues := []string{c.system.HostName, "pcie_function", pcieFunction.Name, pcieFunction.ID, pciFunctionDeviceclass, pciFunctionType}

	if pciFunctionStateValue, ok := common.ParseCommonStatusState(pcieFunction.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_pcie_function_state"].Desc, prometheus.GaugeValue, pciFunctionStateValue, systemPCIeFunctionLabelLabelValues...)
	}

	if pciFunctionHealthStateValue, ok := common.ParseCommonStatusHealth(pcieFunction.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_pcie_function_health_state"].Desc, prometheus.GaugeValue, pciFunctionHealthStateValue, systemPCIeFunctionLabelLabelValues...)
	}
}
