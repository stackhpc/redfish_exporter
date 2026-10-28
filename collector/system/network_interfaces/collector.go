package network_interfaces

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

	SystemNetworkInterfaceLabelNames  = []string{"hostname", "resource", "network_interface", "network_interface_id"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "network_interface_state", fmt.Sprintf("system network interface state,%s", common.StateHelp), SystemNetworkInterfaceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "network_interface_health_state", fmt.Sprintf("system network interface health state,%s", common.HealthHelp), SystemNetworkInterfaceLabelNames)

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
	networkInterfaces, err := c.system.NetworkInterfaces()
	if err != nil {
		c.Log.Error("error getting network interface data from system", "operation", "system.NetworkInterfaces()", "err", err)
		return
	} else if networkInterfaces == nil {
		c.Log.Info("no network interface data found", "operation", "system.NetworkInterfaces()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, networkInterface := range networkInterfaces {
		wg.Go(func() {c.parseNetworkInterface(ch, networkInterface)})
	}
	wg.Wait()
}

func (c *Collector) parseNetworkInterface(ch chan<- prometheus.Metric, networkInterface *schemas.NetworkInterface) {
	systemNetworkInterfaceLabelValues := []string{c.system.HostName, "network_interface", networkInterface.Name, networkInterface.ID}

	if stateValue, ok := common.ParseCommonStatusState(networkInterface.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_network_interface_state"].Desc, prometheus.GaugeValue, stateValue, systemNetworkInterfaceLabelValues...)
	}
	if healthStatusValue, ok := common.ParseCommonStatusHealth(networkInterface.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_network_interface_health_state"].Desc, prometheus.GaugeValue, healthStatusValue, systemNetworkInterfaceLabelValues...)
	}
}
