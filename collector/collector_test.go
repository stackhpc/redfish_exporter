package collector

import (
	"testing"

	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
)

// Test that we can reuse a Collector.
func TestReuse_PublicRackmount1_2026_2(t *testing.T) {
	collector := CreateStaticBMC(t, "DSP2043_2026.2/public-rackmount1", false)

	expected := `
		redfish_chassis_fan_rpm{chassis_id="1U",fan="BaseBoard System Fan Backup",fan_id="1",resource="fan"} 2050
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="BaseBoard System Fan",fan_id="0",resource="fan"} 2100
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="Fan Bay 1",fan_id="Bay1",resource="fan"} 2200
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="Fan Bay 2",fan_id="Bay2",resource="fan"} 2400
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="Fan for CPU 1",fan_id="CPU1",resource="fan"} 1490
       	redfish_chassis_fan_rpm{chassis_id="1U",fan="Fan for CPU 2",fan_id="CPU2",resource="fan"} 1490
	`

	for i := 0; i < 5; i++ {
		redfish_test.CollectAndCompare(t, collector, expected,
			"redfish_chassis_fan_rpm",
		)
	}
}
