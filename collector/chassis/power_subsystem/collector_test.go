package power_subsystem_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_subsystem"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd")

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := power_subsystem.NewCollector(&ctx, logger, chassis[0])
	expected := `
		redfish_chassis_power_powersupply_health{chassis_id="1U",power_supply="Power Supply Bay 1",power_supply_id="",resource="power_supply"} 2
       	redfish_chassis_power_powersupply_power_capacity_watts{chassis_id="1U",power_supply="Power Supply Bay 1",power_supply_id="",resource="power_supply"} 400
       	redfish_chassis_power_powersupply_state{chassis_id="1U",power_supply="Power Supply Bay 1",power_supply_id="",resource="power_supply"} 1
       	redfish_chassis_power_powersupply_state{chassis_id="1U",power_supply="Power Supply Bay 2",power_supply_id="",resource="power_supply"} 7
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_power_powersupply_state",
		"redfish_chassis_power_powersupply_health",
		"redfish_chassis_power_powersupply_power_capacity_watts",
	)
}
