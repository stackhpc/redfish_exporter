package environment_metrics_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis/environment_metrics"
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

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := environment_metrics.NewCollector(&ctx, logger, chassis[0])
	expected := `
       	redfish_chassis_power_average_consumed_watts{chassis_id="Tower",resource="power_wattage",sensor_id="EnvironmentMetrics",sensor_name="Chassis Environment Metrics"} 974
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		// No mockups include voltage on chassis environment metrics
		// "redfish_chassis_power_voltage_volts",
		"redfish_chassis_power_average_consumed_watts",
	)
}
