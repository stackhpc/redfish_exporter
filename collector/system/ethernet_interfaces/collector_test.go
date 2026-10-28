package ethernet_interfaces_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/ethernet_interfaces"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd")

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := ethernet_interfaces.NewCollector(&ctx, logger, systems[0])
	expected := `
		redfish_system_ethernet_interface_health_state{ethernet_interface="",ethernet_interface_id="SC2KP1F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_health_state{ethernet_interface="",ethernet_interface_id="SC2KP2F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_link_enabled{ethernet_interface="",ethernet_interface_id="SC2KP1F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_link_enabled{ethernet_interface="",ethernet_interface_id="SC2KP2F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_link_status{ethernet_interface="",ethernet_interface_id="SC2KP1F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_link_status{ethernet_interface="",ethernet_interface_id="SC2KP2F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_state{ethernet_interface="",ethernet_interface_id="SC2KP1F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
       	redfish_system_ethernet_interface_state{ethernet_interface="",ethernet_interface_id="SC2KP2F0",ethernet_interface_speed="1000 Mbps",hostname="web483",resource="ethernet_interface"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_ethernet_interface_state",
		"redfish_system_ethernet_interface_health_state",
		"redfish_system_ethernet_interface_link_status",
		"redfish_system_ethernet_interface_link_enabled",
	)
}
