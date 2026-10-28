package thermal_metrics_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_subsystem/thermal_metrics"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicTower_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-tower")

	chassis, err := redfishClient.Service.Chassis()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	thermalSubsystem, err := chassis[0].ThermalSubsystem()
	require.NoError(t, err)

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := thermal_metrics.NewCollector(&ctx, logger, &chassis[0].ID, thermalSubsystem)
	expected := `
		redfish_chassis_temperature_celsius{chassis_id="Tower",resource="temperature",sensor="Exhaust",sensor_id="/redfish/v1/Chassis/Tower/Sensors/ExhaustTemp"} 40.5
       	redfish_chassis_temperature_celsius{chassis_id="Tower",resource="temperature",sensor="Intake",sensor_id="/redfish/v1/Chassis/Tower/Sensors/IntakeTemp"} 24.8
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_temperature_celsius",
	)
}
