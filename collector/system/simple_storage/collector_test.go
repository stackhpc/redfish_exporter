package simple_storage_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/simple_storage"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicRackmount1_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-rackmount1")

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := simple_storage.NewCollector(&ctx, logger, systems[0])
	expected := `
		redfish_system_simple_storage_device_health_state{device="SATA Bay 1",hostname="web483",resource="device"} 1
       	redfish_system_simple_storage_device_health_state{device="SATA Bay 2",hostname="web483",resource="device"} 2
       	redfish_system_simple_storage_device_state{device="SATA Bay 1",hostname="web483",resource="device"} 1
       	redfish_system_simple_storage_device_state{device="SATA Bay 2",hostname="web483",resource="device"} 1
       	redfish_system_simple_storage_device_state{device="SATA Bay 3",hostname="web483",resource="device"} 7
       	redfish_system_simple_storage_device_state{device="SATA Bay 4",hostname="web483",resource="device"} 7
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_simple_storage_device_state",
		"redfish_system_simple_storage_device_health_state",
	)
}
