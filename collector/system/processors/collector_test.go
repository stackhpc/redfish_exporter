package processors_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/processors"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicTower_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-tower", false)

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := processors.NewCollector(&ctx, logger, systems[0])
	expected := `
		redfish_system_processor_health_state{hostname="devrender2",processor="GPU1",processor_id="GPU1",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="GPU2",processor_id="GPU2",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU1",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU2",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU3",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU4",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU5",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU6",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU7",resource="processor"} 1
       	redfish_system_processor_health_state{hostname="devrender2",processor="Processor",processor_id="CPU8",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="GPU1",processor_id="GPU1",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="GPU2",processor_id="GPU2",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU1",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU2",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU3",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU4",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU5",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU6",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU7",resource="processor"} 1
       	redfish_system_processor_state{hostname="devrender2",processor="Processor",processor_id="CPU8",resource="processor"} 1
       	redfish_system_processor_total_cores{hostname="devrender2",processor="GPU1",processor_id="GPU1",resource="processor"} 576
       	redfish_system_processor_total_cores{hostname="devrender2",processor="GPU2",processor_id="GPU2",resource="processor"} 576
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU1",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU2",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU3",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU4",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU5",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU6",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU7",resource="processor"} 16
       	redfish_system_processor_total_cores{hostname="devrender2",processor="Processor",processor_id="CPU8",resource="processor"} 16
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU1",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU2",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU3",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU4",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU5",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU6",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU7",resource="processor"} 32
       	redfish_system_processor_total_threads{hostname="devrender2",processor="Processor",processor_id="CPU8",resource="processor"} 32
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_processor_state",
		"redfish_system_processor_health_state",
		"redfish_system_processor_total_threads",
		"redfish_system_processor_total_cores",
	)
}
