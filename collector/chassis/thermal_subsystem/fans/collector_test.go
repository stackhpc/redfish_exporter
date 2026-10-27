package fans_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem/fans"
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

	thermalSubsystem, err := chassis[0].ThermalSubsystem()
	require.NoError(t, err)

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := fans.NewCollector(&ctx, logger, &chassis[0].ID, thermalSubsystem)
	expected := `
		redfish_chassis_fan_health{chassis_id="Tower",fan="Fan Bay 1",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_health{chassis_id="Tower",fan="Fan Bay 2",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_health{chassis_id="Tower",fan="Fan for CPU 1",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_health{chassis_id="Tower",fan="Fan for CPU 2",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_state{chassis_id="Tower",fan="Fan Bay 1",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_state{chassis_id="Tower",fan="Fan Bay 2",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_state{chassis_id="Tower",fan="Fan for CPU 1",fan_id="",fan_unit="",resource="fan"} 1
       	redfish_chassis_fan_state{chassis_id="Tower",fan="Fan for CPU 2",fan_id="",fan_unit="",resource="fan"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_fan_health",
		// Modern fans are incompatible with the ThermalFan type currently used by gofish.
		// "redfish_chassis_fan_rpm_lower_threshold_fatal",
		// "redfish_chassis_fan_rpm_max",
		// "redfish_chassis_fan_rpm_min",
		// "redfish_chassis_fan_rpm_percentage",
		// "redfish_chassis_fan_rpm",
		"redfish_chassis_fan_state",
	)
}
