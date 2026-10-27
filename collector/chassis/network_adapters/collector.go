package network_adapters

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

var (
	Subsystem = "chassis"

	ChassisNetworkAdapterLabelNames = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id"}
	ChassisNetworkPortLabelNames    = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id", "network_port", "network_port_id", "network_port_type", "network_port_connectiont_type", "network_physical_port_number"}

	metricMap = createMetricMap()
)

func createMetricMap() map[string]common.Metric {
	metrics := make(map[string]common.Metric)

	common.AddToMetricMap(metrics, Subsystem, "network_adapter_state", fmt.Sprintf("chassis network adapter state,%s", common.StateHelp), ChassisNetworkAdapterLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "network_adapter_health_state", fmt.Sprintf("chassis network adapter health state,%s", common.HealthHelp), ChassisNetworkAdapterLabelNames)

	common.AddToMetricMap(metrics, Subsystem, "network_port_state", fmt.Sprintf("chassis network port state,%s", common.StateHelp), ChassisNetworkPortLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "network_port_health_state", fmt.Sprintf("chassis network port health state,%s", common.HealthHelp), ChassisNetworkPortLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "network_port_link_state", fmt.Sprintf("chassis network port link state state,%s", common.PortLinkHelp), ChassisNetworkPortLabelNames)
	common.AddToMetricMap(metrics, Subsystem, "network_port_link_speed_gbps", fmt.Sprintf("chassis network port link speed in gbps"), ChassisNetworkPortLabelNames)

	return metrics
}

// Implements the prometheus.Collector.
type Collector struct {
	Ctx     *common.CollectionContext
	metrics map[string]common.Metric
	Log     *slog.Logger
	chassis *schemas.Chassis
}

// NewCollector returns a collector that collects chassis network adapter data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassis *schemas.Chassis) *Collector {
	return &Collector{
		metrics: metricMap,
		Ctx:     ctx,
		Log: logger.With(
			"collector", Subsystem,
		),
		chassis: chassis,
	}
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	wg := &sync.WaitGroup{}

	networkAdapters, err := c.chassis.NetworkAdapters()
	if err != nil {
		c.Log.Error("error getting network adapters data from chassis", "operation", "chassis.NetworkAdapters()", "err", err)
		return
	} else if networkAdapters == nil {
		c.Log.Info("no network adapters data found", "operation", "chassis.NetworkAdapters()")
		return
	}

	for _, networkAdapter := range networkAdapters {
		c.Log.Debug("Found network adapter", "location", networkAdapter.ODataID)
		wg.Go(func() {
			if err = c.parseNetworkAdapter(ch, networkAdapter); err != nil {
				c.Log.Error("error getting network ports from network adapter", "operation", "chassis.NetworkAdapters()", "err", err)
			}
		})
	}

	wg.Wait()
}

func (c *Collector) parseNetworkAdapter(ch chan<- prometheus.Metric, networkAdapter *schemas.NetworkAdapter) error {
	chassisNetworkAdapterLabelValues := []string{"network_adapter", c.chassis.ID, networkAdapter.Name, networkAdapter.ID}
	if networkAdapterStateValue, ok := common.ParseCommonStatusState(networkAdapter.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_adapter_state"].Desc, prometheus.GaugeValue, networkAdapterStateValue, chassisNetworkAdapterLabelValues...)
	}
	if networkAdapterHealthStateValue, ok := common.ParseCommonStatusHealth(networkAdapter.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_adapter_health_state"].Desc, prometheus.GaugeValue, networkAdapterHealthStateValue, chassisNetworkAdapterLabelValues...)
	}

	// NetworkPorts has been deprecated in Redfish v1.5.0
	networkPorts, err := networkAdapter.NetworkPorts()
	if err != nil {
		return err
	}

	// In favour of Ports
	ports, err := networkAdapter.Ports()
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}

	for _, networkPort := range networkPorts {
		c.Log.Debug("Found network port", "location", networkPort.ODataID)
		wg.Go(func() { c.parseNetworkPort(ch, networkPort, networkAdapter.Name, networkAdapter.ID) })
	}

	for _, port := range ports {
		c.Log.Debug("Found port", "location", port.ODataID)
		wg.Go(func() { c.parsePort(ch, port, networkAdapter.Name, networkAdapter.ID) })
	}

	wg.Wait()
	return nil
}

func parseNetworkPortLinkStatus(status schemas.NetworkPortLinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("Up")) {
		return float64(1), true
	}
	return float64(0), false
}

func parsePortLinkStatus(status schemas.PortLinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("LinkUp")) {
		return float64(1), true
	}
	if bytes.Equal([]byte(status), []byte("LinkDown")) {
		return float64(0), true
	}
	return float64(0), false
}

func (c *Collector) parseNetworkPort(ch chan<- prometheus.Metric, networkPort *schemas.NetworkPort, networkAdapterName string, networkAdapterID string) {
	networkPortLinkType := networkPort.ActiveLinkTechnology
	networkPortConnectionType := networkPort.FCPortConnectionType
	chassisNetworkPortLabelValues := []string{"network_port", c.chassis.ID, networkAdapterName, networkAdapterID, networkPort.Name, networkPort.ID, string(networkPortLinkType), string(networkPortConnectionType), networkPort.PhysicalPortNumber}

	if networkPort.CurrentLinkSpeedMbps != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_link_speed_gbps"].Desc, prometheus.GaugeValue, float64(*networkPort.CurrentLinkSpeedMbps) / 1000.0, chassisNetworkPortLabelValues...)
	}

	if linkStatus, ok := parseNetworkPortLinkStatus(networkPort.LinkStatus); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_link_state"].Desc, prometheus.GaugeValue, linkStatus, chassisNetworkPortLabelValues...)
	}

	if portState, ok := common.ParseCommonStatusState(networkPort.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_state"].Desc, prometheus.GaugeValue, portState, chassisNetworkPortLabelValues...)
	}
	if healthStatus, ok := common.ParseCommonStatusHealth(networkPort.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_health_state"].Desc, prometheus.GaugeValue, healthStatus, chassisNetworkPortLabelValues...)
	}
}

func (c *Collector) parsePort(ch chan<- prometheus.Metric, port *schemas.Port, networkAdapterName string, networkAdapterID string) {
	networkPortLinkType := port.LinkNetworkTechnology
	networkPortConnectionType := port.FibreChannel.PortConnectionType
	chassisNetworkPortLabelValues := []string{"network_port", c.chassis.ID, networkAdapterName, networkAdapterID, port.Name, port.ID, string(networkPortLinkType), string(networkPortConnectionType), port.Location.PartLocation.ServiceLabel}

	if port.CurrentSpeedGbps != nil {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_link_speed_gbps"].Desc, prometheus.GaugeValue, *port.CurrentSpeedGbps, chassisNetworkPortLabelValues...)
	}

	if linkStatus, ok := parsePortLinkStatus(port.LinkStatus); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_link_state"].Desc, prometheus.GaugeValue, linkStatus, chassisNetworkPortLabelValues...)
	}

	if portState, ok := common.ParseCommonStatusState(port.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_state"].Desc, prometheus.GaugeValue, portState, chassisNetworkPortLabelValues...)
	}

	if healthStatus, ok := common.ParseCommonStatusHealth(port.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(metricMap["chassis_network_port_health_state"].Desc, prometheus.GaugeValue, healthStatus, chassisNetworkPortLabelValues...)
	}
}
