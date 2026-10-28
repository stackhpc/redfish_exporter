package pcie_devices_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/pcie_devices"
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

	collector := pcie_devices.NewCollector(&ctx, logger, systems[0])
	expected := `
		redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="Integrated NVMe Drive 1",pcie_device_id="IntegratedNVMe1",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="Integrated NVMe Drive 2",pcie_device_id="IntegratedNVMe2",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="Integrated SAS/SATA RAID Controller",pcie_device_id="IntegratedRAID",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="NVMe to PCIe Bridge",pcie_device_id="EmbeddedNVMeBridge",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="PCI Device in Slot 1",pcie_device_id="Slot1",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="PCI Device in Slot 2",pcie_device_id="Slot2",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="PCI Device in Slot 3",pcie_device_id="Slot3",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="PCI Device in Slot 4",pcie_device_id="Slot4",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SM",resource="pcie_device"} 1
       	redfish_system_pcie_device_health_state{hostname="devrender2",pcie_device="Simple Two-Port NIC",pcie_device_id="IntegratedNIC",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="Integrated NVMe Drive 1",pcie_device_id="IntegratedNVMe1",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="Integrated NVMe Drive 2",pcie_device_id="IntegratedNVMe2",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="Integrated SAS/SATA RAID Controller",pcie_device_id="IntegratedRAID",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="NVMe to PCIe Bridge",pcie_device_id="EmbeddedNVMeBridge",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="PCI Device in Slot 1",pcie_device_id="Slot1",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="PCI Device in Slot 2",pcie_device_id="Slot2",pcie_device_partnumber="232-4598D7",pcie_device_type="SingleFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="PCI Device in Slot 3",pcie_device_id="Slot3",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="PCI Device in Slot 4",pcie_device_id="Slot4",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SM",resource="pcie_device"} 1
       	redfish_system_pcie_device_state{hostname="devrender2",pcie_device="Simple Two-Port NIC",pcie_device_id="IntegratedNIC",pcie_device_partnumber="232-4598D7",pcie_device_type="MultiFunction",pcie_serial_number="2M220100SL",resource="pcie_device"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_pcie_device_state",
		"redfish_system_pcie_device_health_state",
	)
}
