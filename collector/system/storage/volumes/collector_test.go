package volumes_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/storage/volumes"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicLocalStorage_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-localstorage", false)

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	storages, err := systems[0].Storage()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	volumeDeduplicator := volumes.NewVolumeDeduplicator()

	collector := volumes.NewCollector(&ctx, logger, storages[0], &systems[0].HostName, &volumeDeduplicator)
	expected := `
		redfish_system_storage_volume_capacity{hostname="web483",resource="volume",volume="Virtual Disk 1",volume_id="1"} 8.99527000064e+11
       	redfish_system_storage_volume_capacity{hostname="web483",resource="volume",volume="Virtual Disk 2",volume_id="2"} 1.073741824e+11
       	redfish_system_storage_volume_capacity{hostname="web483",resource="volume",volume="Virtual Disk 3",volume_id="3"} 1.073741824e+11
       	redfish_system_storage_volume_health_state{hostname="web483",resource="volume",volume="Virtual Disk 1",volume_id="1"} 1
       	redfish_system_storage_volume_health_state{hostname="web483",resource="volume",volume="Virtual Disk 2",volume_id="2"} 1
       	redfish_system_storage_volume_health_state{hostname="web483",resource="volume",volume="Virtual Disk 3",volume_id="3"} 1
       	redfish_system_storage_volume_state{hostname="web483",resource="volume",volume="Virtual Disk 1",volume_id="1"} 1
       	redfish_system_storage_volume_state{hostname="web483",resource="volume",volume="Virtual Disk 2",volume_id="2"} 1
       	redfish_system_storage_volume_state{hostname="web483",resource="volume",volume="Virtual Disk 3",volume_id="3"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_storage_volume_state",
		"redfish_system_storage_volume_health_state",
		"redfish_system_storage_volume_capacity",
	)
}
