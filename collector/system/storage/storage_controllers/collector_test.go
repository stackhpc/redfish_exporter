package storage_controllers_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/storage/storage_controllers"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicLocalStorage_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-localstorage")

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	storages, err := systems[0].Storage()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := storage_controllers.NewCollector(&ctx, logger, storages[0], &systems[0].HostName)
	expected := `
		redfish_system_storage_controller_health_state{hostname="web483",resource="storage_controller",storage_controller="Contoso Integrated RAID",storage_controller_id="/redfish/v1/Systems/437XR1138R2/Storage/1#/StorageControllers/0"} 1
       	redfish_system_storage_controller_state{hostname="web483",resource="storage_controller",storage_controller="Contoso Integrated RAID",storage_controller_id="/redfish/v1/Systems/437XR1138R2/Storage/1#/StorageControllers/0"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_storage_controller_state",
		"redfish_system_storage_controller_health_state",
	)
}
