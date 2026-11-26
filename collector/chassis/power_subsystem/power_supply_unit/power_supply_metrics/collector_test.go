package power_supply_metrics_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power_subsystem/power_supply_unit/power_supply_metrics"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd")

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	powerSubsystem, err := chassis[0].PowerSubsystem()
	require.NoError(t, err)

	powerSupplyUnits, err := powerSubsystem.PowerSupplies()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := power_supply_metrics.NewCollector(&ctx, logger, powerSupplyUnits[0], &chassis[0].ID)
	expected := `
		redfish_chassis_power_powersupply_power_input_watts{chassis_id="1U",power_supply="Power Supply Bay 1",power_supply_id="Bay1",resource="PowerSupply"} 937.4
       	redfish_chassis_power_powersupply_power_output_watts{chassis_id="1U",power_supply="Power Supply Bay 1",power_supply_id="Bay1",resource="PowerSupply"} 937.4
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_power_powersupply_power_input_watts",
		"redfish_chassis_power_powersupply_power_output_watts",
		// None of the Redfish mockups include power load percent
		// "redfish_chassis_power_powersupply_power_load_percent",
		// "redfish_chassis_power_powersupply_power_efficiency_percent",
	)
}
