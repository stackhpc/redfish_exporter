package network_interfaces_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/network_interfaces"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd", false)

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := network_interfaces.NewCollector(&ctx, logger, systems[0])
	expected := `
       	redfish_system_network_interface_state{hostname="web483",network_interface="E810 Network Interface",network_interface_id="SCFCP1F0",resource="network_interface"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_network_interface_state",
		// None of the Redfish mockups report health state for a NetworkInterface
		// "redfish_system_network_interface_health_state"
	)
}
