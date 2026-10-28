package chassis_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/chassis"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
)

func TestPublicCXL_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-cxl")

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
		CollectLogs:   true,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := chassis.NewCollector(&ctx, logger)
	expected := `
		redfish_chassis_log_entry_severity_state{chassis_id="CXL-Chassis",log_entry="Log Entry 1",log_entry_code="",log_entry_id="1",log_entry_message_id="Event.1.0.TempAssert",log_entry_sensor_number="1",log_entry_sensor_type="",log_entry_type="Event",log_service="CXL Module Log Service",log_service_id="LocalLog1"} 3
       	redfish_chassis_log_service_health_state{chassis_id="CXL-Chassis",log_service="CXL Module Log Service",log_service_enabled="true",log_service_id="LocalLog1",log_service_overwrite_policy="WrapsWhenFull"} 1
       	redfish_chassis_log_service_state{chassis_id="CXL-Chassis",log_service="CXL Module Log Service",log_service_enabled="true",log_service_id="LocalLog1",log_service_overwrite_policy="WrapsWhenFull"} 1
       	redfish_chassis_model_info{chassis_id="CXL-Chassis",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="CXLType1",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="CXLType2",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="CXLType3",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="PooledCXLType1",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="PooledCXLType2",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="PooledCXLType3",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_state{chassis_id="CXL-Chassis",resource="chassis"} 1
       	redfish_chassis_state{chassis_id="CXLType1",resource="chassis"} 1
       	redfish_chassis_state{chassis_id="CXLType2",resource="chassis"} 1
       	redfish_chassis_state{chassis_id="CXLType3",resource="chassis"} 1
       	redfish_chassis_state{chassis_id="PooledCXLType1",resource="chassis"} 1
       	redfish_chassis_state{chassis_id="PooledCXLType2",resource="chassis"} 1
       	redfish_chassis_state{chassis_id="PooledCXLType3",resource="chassis"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_state",
		"redfish_chassis_model_info",
		// None of the Redfish mockups include physical security sensors
		//"redfish_chassis_physical_security_sensor_state",
		"redfish_chassis_log_service_state",
		"redfish_chassis_log_service_health_state",
		"redfish_chassis_log_entry_severity_state",
	)
}

func TestPublicCables_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-cables")

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
		CollectLogs:   true,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := chassis.NewCollector(&ctx, logger)
	expected := `
		redfish_chassis_health{chassis_id="bmc",resource="chassis"} 1
       	redfish_chassis_health{chassis_id="hsbp",resource="chassis"} 1
       	redfish_chassis_model_info{chassis_id="bmc",manufacturer="",model="",part_number="1234567-89",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="chassis",manufacturer="",model="",part_number="",resource="chassis",sku=""} 1
       	redfish_chassis_model_info{chassis_id="hsbp",manufacturer="Oogle",model="",part_number="9876543-21",resource="chassis",sku=""} 1
       	redfish_chassis_state{chassis_id="bmc",resource="chassis"} 3
       	redfish_chassis_state{chassis_id="chassis",resource="chassis"} 3
       	redfish_chassis_state{chassis_id="hsbp",resource="chassis"} 3
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_chassis_state",
		"redfish_chassis_health",
		"redfish_chassis_model_info",
	)
}
