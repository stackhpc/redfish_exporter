package network_adapters_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/network_adapters"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicTower_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-tower", false)

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := network_adapters.NewCollector(&ctx, logger, chassis[0])
	expected := `
		redfish_chassis_network_port_health_state{chassis_id="Tower",network_adapter="Network Adapter",network_adapter_id="NIC1",network_physical_port_number="NIC 1",network_port="NIC Port 1",network_port_connectiont_type="",network_port_id="1",network_port_speed="10.000000 Gbps",network_port_type="",resource="network_port"} 1
       	redfish_chassis_network_port_health_state{chassis_id="Tower",network_adapter="Network Adapter",network_adapter_id="NIC1",network_physical_port_number="NIC 1",network_port="NIC Port 2",network_port_connectiont_type="",network_port_id="2",network_port_speed="10.000000 Gbps",network_port_type="",resource="network_port"} 1
       	redfish_chassis_network_port_state{chassis_id="Tower",network_adapter="Network Adapter",network_adapter_id="NIC1",network_physical_port_number="NIC 1",network_port="NIC Port 1",network_port_connectiont_type="",network_port_id="1",network_port_speed="10.000000 Gbps",network_port_type="",resource="network_port"} 1
       	redfish_chassis_network_port_state{chassis_id="Tower",network_adapter="Network Adapter",network_adapter_id="NIC1",network_physical_port_number="NIC 1",network_port="NIC Port 2",network_port_connectiont_type="",network_port_id="2",network_port_speed="10.000000 Gbps",network_port_type="",resource="network_port"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_network_port_state",
		"redfish_chassis_network_port_health_state",
	)
}

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd", false)

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := network_adapters.NewCollector(&ctx, logger, chassis[0])
	expected := `
		redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="Contoso PS100 NIC",network_adapter_id="Slot3",network_physical_port_number="",network_port="25Gbps Ethernet Subport",network_port_connectiont_type="",network_port_id="CSP_P1",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="Contoso PS100 NIC",network_adapter_id="Slot3",network_physical_port_number="",network_port="25Gbps Ethernet Subport",network_port_connectiont_type="",network_port_id="CSP_P2",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="Contoso PS100 NIC",network_adapter_id="Slot3",network_physical_port_number="",network_port="25Gbps Ethernet Subport",network_port_connectiont_type="",network_port_id="CSP_P3",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="Contoso PS100 NIC",network_adapter_id="Slot3",network_physical_port_number="",network_port="25Gbps Ethernet Subport",network_port_connectiont_type="",network_port_id="CSP_P4",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="Contoso SC32G-FC HBA",network_adapter_id="Slot2",network_physical_port_number="",network_port="Port 2",network_port_connectiont_type="",network_port_id="SC2KP2",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="NetworkAdapter",network_adapter_id="Slot1",network_physical_port_number="",network_port="Port 2",network_port_connectiont_type="",network_port_id="SC2KP2",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="NetworkAdapter",network_adapter_id="Slot4",network_physical_port_number="",network_port="100Gbps Ethernet Support",network_port_connectiont_type="",network_port_id="AG_P1",network_port_speed="100.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="NetworkAdapter",network_adapter_id="Slot4",network_physical_port_number="",network_port="25Gbps Ethernet Support",network_port_connectiont_type="",network_port_id="PHY_P1",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="NetworkAdapter",network_adapter_id="Slot4",network_physical_port_number="",network_port="25Gbps Ethernet Support",network_port_connectiont_type="",network_port_id="PHY_P2",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="NetworkAdapter",network_adapter_id="Slot4",network_physical_port_number="",network_port="25Gbps Ethernet Support",network_port_connectiont_type="",network_port_id="PHY_P3",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
       	redfish_chassis_network_port_link_state{chassis_id="1U",network_adapter="NetworkAdapter",network_adapter_id="Slot4",network_physical_port_number="",network_port="25Gbps Ethernet Support",network_port_connectiont_type="",network_port_id="PHY_P4",network_port_speed="25.000000 Gbps",network_port_type="Ethernet",resource="network_port"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_network_port_link_state",
	)
}
