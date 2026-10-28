package power_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/power"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestDSP2043_1_0_0(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_1.0.0/DSP2043-server")

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := power.NewCollector(&ctx, logger, chassis[0])
	expected := `
		redfish_chassis_power_average_consumed_watts{chassis_id="1U",power_voltage="Server Power Control",power_voltage_id="0",resource="power_wattage"} 319
       	redfish_chassis_power_powersupply_health{chassis_id="1U",power_supply="Power Supply Bay",power_supply_id="0",resource="power_supply"} 2
       	redfish_chassis_power_powersupply_last_power_output_watts{chassis_id="1U",power_supply="Power Supply Bay",power_supply_id="0",resource="power_supply"} 325
       	redfish_chassis_power_powersupply_power_capacity_watts{chassis_id="1U",power_supply="Power Supply Bay",power_supply_id="0",resource="power_supply"} 800
       	redfish_chassis_power_powersupply_state{chassis_id="1U",power_supply="Power Supply Bay",power_supply_id="0",resource="power_supply"} 1
       	redfish_chassis_power_voltage_state{chassis_id="1U",power_voltage="VRM1 Voltage",power_voltage_id="0",resource="power_voltage"} 1
       	redfish_chassis_power_voltage_state{chassis_id="1U",power_voltage="VRM2 Voltage",power_voltage_id="1",resource="power_voltage"} 1
       	redfish_chassis_power_voltage_volts{chassis_id="1U",power_voltage="VRM1 Voltage",power_voltage_id="0",resource="power_voltage"} 12
       	redfish_chassis_power_voltage_volts{chassis_id="1U",power_voltage="VRM2 Voltage",power_voltage_id="1",resource="power_voltage"} 5

	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_power_voltage_state",
		"redfish_chassis_power_voltage_volts",
		"redfish_chassis_power_average_consumed_watts",
		"redfish_chassis_power_powersupply_state",
		"redfish_chassis_power_powersupply_health",
		"redfish_chassis_power_powersupply_last_power_output_watts",
		"redfish_chassis_power_powersupply_power_capacity_watts",
	)
}
