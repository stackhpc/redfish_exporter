package ethernet_interfaces

import (
	"bytes"
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

	SystemEthernetInterfaceLabelNames = []string{"hostname", "resource", "ethernet_interface", "ethernet_interface_id", "ethernet_interface_speed"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "ethernet_interface_state", fmt.Sprintf("system ethernet interface state,%s", common.StateHelp), SystemEthernetInterfaceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "ethernet_interface_health_state", fmt.Sprintf("system ethernet interface health state,%s", common.HealthHelp), SystemEthernetInterfaceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "ethernet_interface_link_status", fmt.Sprintf("system ethernet interface link status,%s", common.LinkHelp), SystemEthernetInterfaceLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "ethernet_interface_link_enabled", "system ethernet interface if the link is enabled", SystemEthernetInterfaceLabelNames)

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
			"collector", "system.ethernet_interfaces",
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
	ethernetInterfaces, err := c.system.EthernetInterfaces()
	if err != nil {
		c.Log.Error("error getting ethernet interface data from system", "operation", "system.EthernetInterfaces()", "err", err)
		return
	} else if ethernetInterfaces == nil {
		c.Log.Debug("no ethernet interface data found", "operation", "system.PCIeDevices()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, ethernetInterface := range ethernetInterfaces {
		wg.Go(func() { c.parseEthernetInterface(ch, ethernetInterface) })
	}
	wg.Wait()
	c.Log.Debug("collection complete", "duration", time.Since(startTime))
}

func (c *Collector) parseEthernetInterface(ch chan<- prometheus.Metric, ethernetInterface *schemas.EthernetInterface) {
	ethernetInterfaceSpeed := fmt.Sprintf("%d Mbps", *ethernetInterface.SpeedMbps)
	systemEthernetInterfaceLabelValues := []string{c.system.HostName, "ethernet_interface", ethernetInterface.Name, ethernetInterface.ID, ethernetInterfaceSpeed}
	if ethernetInterfaceStateValue, ok := common.ParseCommonStatusState(ethernetInterface.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_ethernet_interface_state"].Desc, prometheus.GaugeValue, ethernetInterfaceStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceHealthStateValue, ok := common.ParseCommonStatusHealth(ethernetInterface.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_ethernet_interface_health_state"].Desc, prometheus.GaugeValue, ethernetInterfaceHealthStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceLinkStatusValue, ok := parseLinkStatus(ethernetInterface.LinkStatus); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["system_ethernet_interface_link_status"].Desc, prometheus.GaugeValue, ethernetInterfaceLinkStatusValue, systemEthernetInterfaceLabelValues...)
	}

	ch <- prometheus.MustNewConstMetric(metricMap["system_ethernet_interface_link_enabled"].Desc, prometheus.GaugeValue, common.BoolToFloat64(ethernetInterface.InterfaceEnabled), systemEthernetInterfaceLabelValues...)
}

func parseLinkStatus(status schemas.LinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("LinkUp")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("NoLink")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("LinkDown")) {
		return float64(3), true
	}
	return float64(0), false
}
