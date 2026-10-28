package system_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
)

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd")

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
		CollectLogs:   true,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := system.NewCollector(&ctx, logger)
	expected := `
		redfish_system_bios_info{bios_version="P79 v1.45 (12/06/2017)",hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_health_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_log_entry_severity_state{log_entry="Log Entry 1",log_entry_code="Upper Critical - going high",log_entry_id="1",log_entry_message_id="0x592A28",log_entry_sensor_number="1",log_entry_sensor_type="Temperature",log_entry_type="SEL",log_service="System Log Service",log_service_id="Log1",system_id="437XR1138R2"} 3
       	redfish_system_log_entry_severity_state{log_entry="Log Entry 2",log_entry_code="Upper Critical - going high",log_entry_id="2",log_entry_message_id="0x592E28",log_entry_sensor_number="2",log_entry_sensor_type="Temperature",log_entry_type="SEL",log_service="System Log Service",log_service_id="Log1",system_id="437XR1138R2"} 3
       	redfish_system_log_service_health_state{log_service="System Log Service",log_service_enabled="true",log_service_id="Log1",log_service_overwrite_policy="WrapsWhenFull",system_id="437XR1138R2"} 1
       	redfish_system_log_service_state{log_service="System Log Service",log_service_enabled="true",log_service_id="Log1",log_service_overwrite_policy="WrapsWhenFull",system_id="437XR1138R2"} 1
       	redfish_system_power_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_total_memory_health_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_total_memory_size{hostname="web483",resource="system",system_id="437XR1138R2"} 96
       	redfish_system_total_memory_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_total_processor_count{hostname="web483",resource="system",system_id="437XR1138R2"} 2
       	redfish_system_total_processor_health_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
       	redfish_system_total_processor_state{hostname="web483",resource="system",system_id="437XR1138R2"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_state",
		"redfish_system_health_state",
		"redfish_system_power_state",
		"redfish_system_bios_info",
		"redfish_system_total_memory_state",
		"redfish_system_total_memory_health_state",
		"redfish_system_total_memory_size",
		"redfish_system_total_processor_state",
		"redfish_system_total_processor_health_state",
		"redfish_system_total_processor_count",
		"redfish_system_log_service_state",
		"redfish_system_log_service_health_state",
		"redfish_system_log_entry_severity_state",
	)
}
