package drives_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/storage/drives"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicApplications_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-applications", false)

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	storages, err := systems[0].Storage()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := drives.NewCollector(&ctx, logger, storages[0], &systems[0].HostName)
	expected := `
		redfish_system_storage_drive_capacity{drive="Drive Sample",drive_id="32ADF365C6C1B7BD",hostname="web483",resource="drive"} 8.99527e+11
       	redfish_system_storage_drive_capacity{drive="Drive Sample",drive_id="35D38F11ACEF7BD3",hostname="web483",resource="drive"} 8.99527e+11
       	redfish_system_storage_drive_capacity{drive="Drive Sample",drive_id="3D58ECBC375FD9F2",hostname="web483",resource="drive"} 8.99527e+11
       	redfish_system_storage_drive_capacity{drive="Drive Sample",drive_id="3F5A8C54207B7233",hostname="web483",resource="drive"} 8.99527e+11
       	redfish_system_storage_drive_health_state{drive="Drive Sample",drive_id="32ADF365C6C1B7BD",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_health_state{drive="Drive Sample",drive_id="35D38F11ACEF7BD3",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_health_state{drive="Drive Sample",drive_id="3D58ECBC375FD9F2",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_health_state{drive="Drive Sample",drive_id="3F5A8C54207B7233",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_state{drive="Drive Sample",drive_id="32ADF365C6C1B7BD",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_state{drive="Drive Sample",drive_id="35D38F11ACEF7BD3",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_state{drive="Drive Sample",drive_id="3D58ECBC375FD9F2",hostname="web483",resource="drive"} 1
       	redfish_system_storage_drive_state{drive="Drive Sample",drive_id="3F5A8C54207B7233",hostname="web483",resource="drive"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_storage_drive_state",
		"redfish_system_storage_drive_health_state",
		"redfish_system_storage_drive_capacity",
	)
}
