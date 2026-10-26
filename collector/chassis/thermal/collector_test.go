package thermal_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestDSP2043_1_0_0(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_1.0.0/DSP2043-server", false)

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := thermal.NewCollector(&ctx, logger, chassis[0])
	expected := `
		redfish_chassis_fan_health{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 1
       	redfish_chassis_fan_health{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 1
       	redfish_chassis_fan_rpm_lower_threshold_fatal{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 0
       	redfish_chassis_fan_rpm_lower_threshold_fatal{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 0
       	redfish_chassis_fan_rpm_max{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 5000
       	redfish_chassis_fan_rpm_max{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 5000
       	redfish_chassis_fan_rpm_min{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 0
       	redfish_chassis_fan_rpm_min{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 0
       	redfish_chassis_fan_rpm_percentage{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 41
       	redfish_chassis_fan_rpm_percentage{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 42
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 2050
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 2100
       	redfish_chassis_fan_state{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",fan_unit="rpm",resource="fan"} 1
       	redfish_chassis_fan_state{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",fan_unit="rpm",resource="fan"} 1
       	redfish_chassis_temperature_celsius{chassis_id="1U",resource="temperature",sensor="CPU1 Temp",sensor_id="0"} 41
       	redfish_chassis_temperature_celsius{chassis_id="1U",resource="temperature",sensor="Chassis Intake Temp",sensor_id="2"} 25
       	redfish_chassis_temperature_sensor_health{chassis_id="1U",resource="temperature",sensor="CPU1 Temp",sensor_id="0"} 1
       	redfish_chassis_temperature_sensor_health{chassis_id="1U",resource="temperature",sensor="Chassis Intake Temp",sensor_id="2"} 1
       	redfish_chassis_temperature_sensor_state{chassis_id="1U",resource="temperature",sensor="CPU1 Temp",sensor_id="0"} 1
       	redfish_chassis_temperature_sensor_state{chassis_id="1U",resource="temperature",sensor="CPU2 Temp",sensor_id="1"} 2
       	redfish_chassis_temperature_sensor_state{chassis_id="1U",resource="temperature",sensor="Chassis Intake Temp",sensor_id="2"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_fan_health",
		"redfish_chassis_fan_rpm_lower_threshold_fatal",
		"redfish_chassis_fan_rpm_max",
		"redfish_chassis_fan_rpm_min",
		"redfish_chassis_fan_rpm_percentage",
		"redfish_chassis_fan_rpm",
		"redfish_chassis_fan_state",
		"redfish_chassis_temperature_celsius",
		"redfish_chassis_temperature_sensor_health",
		"redfish_chassis_temperature_sensor_state",
	)
}
