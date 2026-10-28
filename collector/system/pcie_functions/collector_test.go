package pcie_functions_test

import (
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/collector/system/pcie_functions"
	"github.com/stackhpc/redfish_exporter/common"
	redfish_test "github.com/stackhpc/redfish_exporter/common/test"
	"github.com/stretchr/testify/require"
)

func TestPublicCXL_2026_2(t *testing.T) {
	redfishClient := redfish_test.CreateRedfishClient(t, "DSP2043_2026.2/public-cxl")

	systems, err := redfishClient.Service.Systems()
	require.NoError(t, err)

	ctx := common.CollectionContext{
		RedfishClient: redfishClient,
	}

	promslogConfig := promslog.Config{}
	logger := promslog.New(&promslogConfig)

	collector := pcie_functions.NewCollector(&ctx, logger, systems[0])
	expected := `
		redfish_system_pcie_function_health_state{hostname="",pci_function_deviceclass="MemoryController",pci_function_type="Physical",pcie_function_id="1",pcie_function_name="CXL Function",resource="pcie_function"} 1
       	redfish_system_pcie_function_health_state{hostname="",pci_function_deviceclass="ProcessingAccelerators",pci_function_type="Physical",pcie_function_id="1",pcie_function_name="CXL Function",resource="pcie_function"} 1
       	redfish_system_pcie_function_health_state{hostname="",pci_function_deviceclass="Processor",pci_function_type="Physical",pcie_function_id="1",pcie_function_name="CXL Function",resource="pcie_function"} 1
       	redfish_system_pcie_function_state{hostname="",pci_function_deviceclass="MemoryController",pci_function_type="Physical",pcie_function_id="1",pcie_function_name="CXL Function",resource="pcie_function"} 1
       	redfish_system_pcie_function_state{hostname="",pci_function_deviceclass="ProcessingAccelerators",pci_function_type="Physical",pcie_function_id="1",pcie_function_name="CXL Function",resource="pcie_function"} 1
       	redfish_system_pcie_function_state{hostname="",pci_function_deviceclass="Processor",pci_function_type="Physical",pcie_function_id="1",pcie_function_name="CXL Function",resource="pcie_function"} 1
	`
	redfish_test.CollectAndCompare(t, collector, expected,
		"redfish_system_pcie_function_state",
		"redfish_system_pcie_function_health_state",
	)
}
