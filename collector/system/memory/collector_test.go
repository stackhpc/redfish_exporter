package memory_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/memory"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicTower_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-tower")

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := memory.NewCollector(&ctx, logger, systems[0])
	expected := `
		redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM1",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM10",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM11",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM12",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM13",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM14",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM15",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM16",resource="memory"} 0
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM2",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM3",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM4",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM5",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM6",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM7",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM8",resource="memory"} 131072
       	redfish_system_memory_capacity{hostname="devrender2",memory="Regular Memory",memory_id="DIMM9",resource="memory"} 0
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM1",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM10",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM11",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM12",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM13",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM14",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM15",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM16",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM2",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM3",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM4",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM5",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM6",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM7",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM8",resource="memory"} 1
       	redfish_system_memory_health_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM9",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM1",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM10",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM11",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM12",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM13",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM14",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM15",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM16",resource="memory"} 7
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM2",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM3",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM4",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM5",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM6",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM7",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM8",resource="memory"} 1
       	redfish_system_memory_state{hostname="devrender2",memory="Regular Memory",memory_id="DIMM9",resource="memory"} 7
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_memory_state",
		"redfish_system_memory_health_state",
		"redfish_system_memory_capacity",
	)
}
