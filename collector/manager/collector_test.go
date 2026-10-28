package manager_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/manager"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
)

func TestPublicACD_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-acd", false)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
		CollectLogs: true,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := manager.NewCollector(&ctx, logger)
	expected := `
		redfish_manager_firmware_info{firmware_version="1.45.455b66-rev4",manager_id="BMC",model="Joo Janta 200",name="Manager",type="BMC"} 1
       	redfish_manager_health_state{manager_id="BMC",model="Joo Janta 200",name="Manager",type="BMC"} 1
       	redfish_manager_log_entry_severity_state{log_entry="Log Entry 1",log_entry_code="",log_entry_id="1",log_entry_message_id="Event.1.0.TempWayTooHot",log_entry_sensor_number="0x13f1b133aed0",log_entry_sensor_type="",log_entry_type="Event",log_service="System Log Service",log_service_id="Log",manager_id="BMC"} 3
       	redfish_manager_log_service_health_state{log_service="System Log Service",log_service_enabled="true",log_service_id="Log",log_service_overwrite_policy="WrapsWhenFull",manager_id="BMC"} 1
       	redfish_manager_log_service_state{log_service="System Log Service",log_service_enabled="true",log_service_id="Log",log_service_overwrite_policy="WrapsWhenFull",manager_id="BMC"} 1
       	redfish_manager_power_state{manager_id="BMC",model="Joo Janta 200",name="Manager",type="BMC"} 1
       	redfish_manager_state{manager_id="BMC",model="Joo Janta 200",name="Manager",type="BMC"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_manager_state",
		"redfish_manager_health_state",
		"redfish_manager_power_state",
		"redfish_manager_firmware_info",
		"redfish_manager_log_service_state",
		"redfish_manager_log_service_health_state",
		"redfish_manager_log_entry_severity_state",
	)
}
