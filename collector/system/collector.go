package system

import (
	"bytes"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/system/memory"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

// SystemSubsystem is the system subsystem
var (
	SystemSubsystem                   = "system"
	SystemLabelNames                  = []string{"hostname", "resource", "system_id"}
	SystemBIOSLabelNames              = append(slices.Clone(SystemLabelNames), "bios_version")
	SystemProcessorLabelNames         = []string{"hostname", "resource", "processor", "processor_id"}
	SystemVolumeLabelNames            = []string{"hostname", "resource", "volume", "volume_id"}
	SystemDeviceLabelNames            = []string{"hostname", "resource", "device"}
	SystemDriveLabelNames             = []string{"hostname", "resource", "drive", "drive_id"}
	SystemStorageControllerLabelNames = []string{"hostname", "resource", "storage_controller", "storage_controller_id"}
	SystemPCIeDeviceLabelNames        = []string{"hostname", "resource", "pcie_device", "pcie_device_id", "pcie_device_partnumber", "pcie_device_type", "pcie_serial_number"}
	SystemNetworkInterfaceLabelNames  = []string{"hostname", "resource", "network_interface", "network_interface_id"}
	SystemEthernetInterfaceLabelNames = []string{"hostname", "resource", "ethernet_interface", "ethernet_interface_id", "ethernet_interface_speed"}
	SystemPCIeFunctionLabelNames      = []string{"hostname", "resource", "pcie_function_name", "pcie_function_id", "pci_function_deviceclass", "pci_function_type"}

	SystemLogServiceLabelNames = []string{"system_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	SystemLogEntryLabelNames   = []string{"system_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	systemMetrics = createSystemMetricMap()
)

// SystemCollector implements the prometheus.Collector.
type SystemCollector struct {
	Ctx     *common.CollectionContext
	metrics map[string]common.Metric
	prometheus.Collector
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
}

func createSystemMetricMap() map[string]common.Metric {
	systemMetrics := make(map[string]common.Metric)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "state", fmt.Sprintf("system state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "health_state", fmt.Sprintf("system health,%s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "power_state", "system power state", SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "bios_info", "system BIOS version, value is always 1", SystemBIOSLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_memory_state", fmt.Sprintf("system overall memory state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_memory_health_state", fmt.Sprintf("system overall memory health,%s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_memory_size", "system total memory size, GiB", SystemLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_processor_state", fmt.Sprintf("system overall processor state,%s", common.StateHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_processor_health_state", fmt.Sprintf("system overall processor health,%s", common.HealthHelp), SystemLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "total_processor_count", "system total processor count", SystemLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "processor_state", fmt.Sprintf("system processor state,%s", common.StateHelp), SystemProcessorLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "processor_health_state", fmt.Sprintf("system processor health state,%s", common.HealthHelp), SystemProcessorLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "processor_total_threads", "system processor total threads", SystemProcessorLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "processor_total_threads_enabled", "system processor total enabled threads", SystemProcessorLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "processor_total_cores", "system processor total cores", SystemProcessorLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "processor_total_cores_enabled", "system processor total enabled cores", SystemProcessorLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "simple_storage_device_state", fmt.Sprintf("system simple storage device state,%s", common.StateHelp), SystemDeviceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "simple_storage_device_health_state", fmt.Sprintf("system simple storage device health state,%s", common.HealthHelp), SystemDeviceLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_volume_state", fmt.Sprintf("system storage volume state,%s", common.StateHelp), SystemVolumeLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_volume_health_state", fmt.Sprintf("system storage volume health state,%s", common.HealthHelp), SystemVolumeLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_volume_capacity", "system storage volume capacity, Bytes", SystemVolumeLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_drive_state", fmt.Sprintf("system storage drive state,%s", common.StateHelp), SystemDriveLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_drive_health_state", fmt.Sprintf("system storage drive health state,%s", common.HealthHelp), SystemDriveLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_drive_capacity", "system storage drive capacity, Bytes", SystemDriveLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_controller_state", fmt.Sprintf("system storage controller state,%s", common.StateHelp), SystemStorageControllerLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "storage_controller_health_state", fmt.Sprintf("system storage controller health state,%s", common.HealthHelp), SystemStorageControllerLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "pcie_device_state", fmt.Sprintf("system pcie device state,%s", common.StateHelp), SystemPCIeDeviceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "pcie_device_health_state", fmt.Sprintf("system pcie device health state,%s", common.HealthHelp), SystemPCIeDeviceLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "pcie_function_state", fmt.Sprintf("system pcie function state,%s", common.StateHelp), SystemPCIeFunctionLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "pcie_function_health_state", fmt.Sprintf("system pcie device function state,%s", common.HealthHelp), SystemPCIeFunctionLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "network_interface_state", fmt.Sprintf("system network interface state,%s", common.StateHelp), SystemNetworkInterfaceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "network_interface_health_state", fmt.Sprintf("system network interface health state,%s", common.HealthHelp), SystemNetworkInterfaceLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_state", fmt.Sprintf("system ethernet interface state,%s", common.StateHelp), SystemEthernetInterfaceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_health_state", fmt.Sprintf("system ethernet interface health state,%s", common.HealthHelp), SystemEthernetInterfaceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_link_status", fmt.Sprintf("system ethernet interface link status,%s", common.LinkHelp), SystemEthernetInterfaceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_link_enabled", "system ethernet interface if the link is enabled", SystemEthernetInterfaceLabelNames)

	common.AddToMetricMap(systemMetrics, SystemSubsystem, "log_service_state", fmt.Sprintf("system log service state,%s", common.StateHelp), SystemLogServiceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "log_service_health_state", fmt.Sprintf("system log service health state,%s", common.HealthHelp), SystemLogServiceLabelNames)
	common.AddToMetricMap(systemMetrics, SystemSubsystem, "log_entry_severity_state", fmt.Sprintf("system log entry severity state,%s", common.SeverityHelp), SystemLogEntryLabelNames)

	return systemMetrics
}

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(ctx *common.CollectionContext, logger *slog.Logger) *SystemCollector {
	return &SystemCollector{
		metrics: systemMetrics,
		Ctx:     ctx,
		Log: logger.With(
			"collector", "SystemCollector",
		),
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: common.Namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
	}
}

// Describe implements prometheus.Collector.
func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.Desc
	}
	s.collectorScrapeStatus.Describe(ch)
}

// Collect implements prometheus.Collector.
func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := s.Log
	disabledMetrics := s.Ctx.DisabledMetrics
	collectLogs := s.Ctx.CollectLogs
	//get service
	service := s.Ctx.RedfishClient.Service

	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		collectorLogContext.Error("error getting systems from service", "operation", "service.Systems()", "err", err)
	} else {
		for _, system := range systems {
			systemLogContext := collectorLogContext.With("System", system.ID)
			systemLogContext.Info("collector scrape started")
			// overall system metrics

			SystemID := system.ID
			systemHostName := system.HostName
			systemPowerState := system.PowerState
			systemState := system.Status.State
			systemHealthState := system.Status.Health
			systemTotalProcessorsState := system.ProcessorSummary.Status.State
			systemTotalProcessorsHealthState := system.ProcessorSummary.Status.Health
			systemTotalMemoryState := system.MemorySummary.Status.State
			systemTotalMemoryHealthState := system.MemorySummary.Status.Health

			systemLabelValues := []string{systemHostName, "system", SystemID}
			if systemHealthStateValue, ok := common.ParseCommonStatusHealth(systemHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].Desc, prometheus.GaugeValue, systemHealthStateValue, systemLabelValues...)
			}
			if systemStateValue, ok := common.ParseCommonStatusState(systemState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].Desc, prometheus.GaugeValue, systemStateValue, systemLabelValues...)
			}
			if systemPowerStateValue, ok := common.ParseCommonPowerState(systemPowerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].Desc, prometheus.GaugeValue, systemPowerStateValue, systemLabelValues...)
			}
			if system.BiosVersion != "" {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_bios_info"].Desc, prometheus.GaugeValue, 1, append(systemLabelValues, system.BiosVersion)...)
			}
			if systemTotalProcessorsStateValue, ok := common.ParseCommonStatusState(systemTotalProcessorsState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].Desc, prometheus.GaugeValue, systemTotalProcessorsStateValue, systemLabelValues...)
				if system.ProcessorSummary.Count != nil {
					ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].Desc, prometheus.GaugeValue, float64(*system.ProcessorSummary.Count), systemLabelValues...)
				}
			}
			if systemTotalProcessorsHealthStateValue, ok := common.ParseCommonStatusHealth(systemTotalProcessorsHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].Desc, prometheus.GaugeValue, systemTotalProcessorsHealthStateValue, systemLabelValues...)
			}
			if systemTotalMemoryStateValue, ok := common.ParseCommonStatusState(systemTotalMemoryState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].Desc, prometheus.GaugeValue, systemTotalMemoryStateValue, systemLabelValues...)
				if system.MemorySummary.TotalSystemMemoryGiB != nil {
					ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].Desc, prometheus.GaugeValue, float64(*system.MemorySummary.TotalSystemMemoryGiB), systemLabelValues...)
				}
			}
			if systemTotalMemoryHealthStateValue, ok := common.ParseCommonStatusHealth(systemTotalMemoryHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].Desc, prometheus.GaugeValue, systemTotalMemoryHealthStateValue, systemLabelValues...)
			}

			// get system OdataID
			//systemOdataID := system.ODataID

			wg2 := &sync.WaitGroup{}  // processor
			wg3 := &sync.WaitGroup{}  // storage/volumes
			wg4 := &sync.WaitGroup{}  // storage/drives
			wg5 := &sync.WaitGroup{}  // pcie devices
			wg6 := &sync.WaitGroup{}  // network interfaces
			wg7 := &sync.WaitGroup{}  // ethernet interfaces
			wg8 := &sync.WaitGroup{}  // simple storage
			wg9 := &sync.WaitGroup{}  // pcie functions
			wg10 := &sync.WaitGroup{} // log services

			wg := &sync.WaitGroup{}

			if !slices.Contains(disabledMetrics, "memory") {
				collector := memory.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "processor") {
				processProcessors(system, systemLogContext, wg2, ch, systemHostName)
			}

			if !slices.Contains(disabledMetrics, "storage") {
				processStorage(system, systemLogContext, wg3, ch, systemHostName, wg4)
			}

			if !slices.Contains(disabledMetrics, "pcie_devices") {
				processPcieDevices(system, systemLogContext, wg5, ch, systemHostName)
			}

			if !slices.Contains(disabledMetrics, "network_interfaces") {
				processNetworkInterfaces(system, systemLogContext, wg6, ch, systemHostName)
			}

			if !slices.Contains(disabledMetrics, "ethernet_interfaces") {
				processEthernetInterfaces(system, systemLogContext, wg7, ch, systemHostName)
			}

			if !slices.Contains(disabledMetrics, "simple_storage") {
				processSimpleStorage(system, systemLogContext, wg8, ch, systemHostName)
			}

			if !slices.Contains(disabledMetrics, "pcie_functions") {
				processPcieFunctions(system, systemLogContext, wg9, ch, systemHostName)
			}

			// process log services
			if collectLogs {
				logServices, err := system.LogServices()
				if err != nil {
					systemLogContext.Error("error getting log services from system", "operation", "system.LogServices()", "err", err)
				} else if logServices == nil {
					systemLogContext.Info("no log services found", "operation", "system.LogServices()")
				} else {
					wg10.Add(len(logServices))

					for _, logService := range logServices {
						if err = common.ParseLogService(ch, systemMetrics, s.Ctx, systemLogContext, SystemSubsystem, SystemID, logService); err != nil {
							systemLogContext.Error("error getting log entries from log service", "operation", "system.LogServices()", "err", err)
						}
					}
				}
			}

			wg2.Wait()
			wg3.Wait()
			wg4.Wait()
			wg5.Wait()
			wg6.Wait()
			wg7.Wait()
			wg8.Wait()
			wg9.Wait()
			wg10.Wait()

			wg.Wait()

			systemLogContext.Info("collector scrape completed")
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	}
}

func processProcessors(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg2 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string) {
	processors, err := system.Processors()
	if err != nil {
		systemLogContext.Error("error getting processor data from system", "operation", "system.Processors()", "err", err)
	} else if processors == nil {
		systemLogContext.Info("no processor data found", "operation", "system.Processors()")
	} else {
		wg2.Add(len(processors))

		for _, processor := range processors {
			go parseProcessor(ch, systemHostName, processor, wg2)

		}
	}
}

func processStorage(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg3 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string, wg4 *sync.WaitGroup) {
	storages, err := system.Storage()
	if err != nil {
		systemLogContext.Error("error getting storage data from system", "operation", "system.Storage()", "err", err)
	} else if storages == nil {
		systemLogContext.Info("no storage data found", "operation", "system.Storage()")
	} else {
		processed := make(map[string]bool)
		for _, storage := range storages {
			if volumes, err := storage.Volumes(); err != nil {
				systemLogContext.Error("error getting storage data from system", "operation", "system.Volumes()", "err", err)
			} else {
				wg3.Add(len(volumes))

				for _, volume := range volumes {
					_, exists := processed[volume.Name]
					if exists {
						systemLogContext.With("operation",
							"system.Storage()").Info(fmt.Sprintf("Ignoring "+
							"duplicate storage volume: %s. Please check whether this "+
							"volume is returning duplicate data and report to the vendor.",
							volume.Name))
						wg3.Done()
						continue
					}
					go parseVolume(ch, systemHostName, volume, wg3)
					processed[volume.Name] = true
				}
			}

			drives, err := storage.Drives()
			if err != nil {
				systemLogContext.Error("error getting drive data from system", "operation", "system.Drives()", "err", err)
			} else if drives == nil {
				systemLogContext.Info("no drive data found", "operation", "system.Drives()", "storage", storage.ID)
			} else {
				wg4.Add(len(drives))
				for _, drive := range drives {
					go parseDrive(ch, systemHostName, drive, wg4)
				}
			}

			//					if storagecontrollers, err := storage.StorageControllers(); err != nil {
			//						log.Infof("Errors Getting storagecontrollers from system storage : %s", err)
			//					} else {
			//
			//						for _, controller := range storagecontrollers {
			//
			//							controllerODataIDslice := strings.Split(controller.ODataID, "/")
			//							controllerName := controllerODataIDslice[len(controllerODataIDslice)-1]
			//							controllerState := controller.Status.State
			//							controllerHealthState := controller.Status.Health
			//							controllerLabelValues := []string{ "storage_controller", controllerName, systemHostName)
			//							if controllerStateValue,ok := parseCommonStatusState(controllerState); ok {
			//								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_state"].desc, prometheus.GaugeValue, controllerStateValue, //controllerLabelValues...)
			//
			//							}
			//							if controllerHealthStateValue,ok := parseCommonStatusHealth(controllerHealthState); ok {
			//								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_health_state"].desc, prometheus.GaugeValue, controllerHealthStateValue, //controllerLabelValues...)
			//
			//							}
			//
			//						}
			//
			//					}

		}
	}
}

func processPcieDevices(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg5 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string) {
	pcieDevices, err := system.PCIeDevices()
	if err != nil {
		systemLogContext.Error("error getting PCI-E device data from system", "operation", "system.PCIeDevices()", "err", err)
	} else if pcieDevices == nil {
		systemLogContext.Info("no PCI-E device data found", "operation", "system.PCIeDevices()")
	} else {
		processed := make(map[string]bool)
		wg5.Add(len(pcieDevices))
		//Some devices are returning duplicated PCIeDevices. This is workaround for this. Example of such data can be found in sampleOut/system_duplicated_devices.json
		for _, pcieDevice := range pcieDevices {
			_, exists := processed[pcieDevice.ODataID]
			if exists {
				systemLogContext.Info(fmt.Sprintf("Ignoring duplicate pci device: %s", pcieDevice.ODataID), "operation", "system.PCIeDevices()")
				wg5.Done()
				continue
			}
			processed[pcieDevice.ODataID] = true
			go parsePcieDevice(ch, systemHostName, pcieDevice, wg5)
		}
	}
}

func processPcieFunctions(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg9 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string) {
	pcieFunctions, err := system.PCIeFunctions()
	if err != nil {
		systemLogContext.Error("error getting PCI-E device function data from system", "operation", "system.PCIEFunctions()", "err", err)
	} else if pcieFunctions == nil {
		systemLogContext.Info("no PCI-E device function data found", "operation", "system.PCIeFunctions()")
	} else {
		wg9.Add(len(pcieFunctions))
		for _, pcieFunction := range pcieFunctions {
			go parsePcieFunction(ch, systemHostName, pcieFunction, wg9)
		}
	}
}

func processSimpleStorage(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg8 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string) {
	simpleStorages, err := system.SimpleStorage()
	if err != nil {
		systemLogContext.Error("error getting simple storage data from system", "operation", "system.SimpleStorages()", "err", err)
	} else if simpleStorages == nil {
		systemLogContext.Info("no simple storage data found", "operation", "system.SimpleStorages()")
	} else {
		processed := make(map[string]bool)
		for _, simpleStorage := range simpleStorages {
			devices := simpleStorage.Devices
			wg8.Add(len(devices))
			for _, device := range devices {
				_, exists := processed[device.Name]
				if exists {
					systemLogContext.With("operation",
						"system.SimpleStorages()").Info(fmt.Sprintf("Ignoring "+
						"duplicate storage device: %s. Please check whether this "+
						"device is returning duplicate data and report to the vendor.",
						device.Name))
					wg8.Done()
					continue
				}
				go parseDevice(ch, systemHostName, device, wg8)
				processed[device.Name] = true
			}
		}
	}
}

func processEthernetInterfaces(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg7 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string) {
	ethernetInterfaces, err := system.EthernetInterfaces()
	if err != nil {
		systemLogContext.Error("error getting ethernet interface data from system", "operation", "system.EthernetInterfaces()", "err", err)
	} else if ethernetInterfaces == nil {
		systemLogContext.Info("no ethernet interface data found", "operation", "system.PCIeDevices()")
	} else {
		wg7.Add(len(ethernetInterfaces))
		for _, ethernetInterface := range ethernetInterfaces {
			go parseEthernetInterface(ch, systemHostName, ethernetInterface, wg7)
		}
	}
}

func processNetworkInterfaces(system *schemas.ComputerSystem, systemLogContext *slog.Logger, wg6 *sync.WaitGroup, ch chan<- prometheus.Metric, systemHostName string) {
	networkInterfaces, err := system.NetworkInterfaces()
	if err != nil {
		systemLogContext.Error("error getting network interface data from system", "operation", "system.NetworkInterfaces()", "err", err)
	} else if networkInterfaces == nil {
		systemLogContext.Info("no network interface data found", "operation", "system.NetworkInterfaces()")
	} else {
		wg6.Add(len(networkInterfaces))
		for _, networkInterface := range networkInterfaces {
			go parseNetworkInterface(ch, systemHostName, networkInterface, wg6)
		}
	}
}

func parseProcessor(ch chan<- prometheus.Metric, systemHostName string, processor *schemas.Processor, wg *sync.WaitGroup) {
	defer wg.Done()

	systemProcessorLabelValues := []string{systemHostName, "processor", processor.Name, processor.ID}

	if processorStateValue, ok := common.ParseCommonStatusState(processor.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_state"].Desc, prometheus.GaugeValue, processorStateValue, systemProcessorLabelValues...)
	}
	if processorHealthStateValue, ok := common.ParseCommonStatusHealth(processor.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_health_state"].Desc, prometheus.GaugeValue, processorHealthStateValue, systemProcessorLabelValues...)
	}
	if processor.TotalThreads != nil {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_threads"].Desc, prometheus.GaugeValue, float64(*processor.TotalThreads), systemProcessorLabelValues...)
	}
	if processor.TotalEnabledThreads != nil {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_threads_enabled"].Desc, prometheus.GaugeValue, float64(*processor.TotalThreads), systemProcessorLabelValues...)
	}
	if processor.TotalCores != nil {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_cores"].Desc, prometheus.GaugeValue, float64(*processor.TotalCores), systemProcessorLabelValues...)
	}
	if processor.TotalEnabledCores != nil {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_cores_enabled"].Desc, prometheus.GaugeValue, float64(*processor.TotalCores), systemProcessorLabelValues...)
	}
}
func parseVolume(ch chan<- prometheus.Metric, systemHostName string, volume *schemas.Volume, wg *sync.WaitGroup) {
	defer wg.Done()
	volumeName := volume.Name
	volumeID := volume.ID
	volumeState := volume.Status.State
	volumeHealthState := volume.Status.Health
	systemVolumeLabelValues := []string{systemHostName, "volume", volumeName, volumeID}
	if volumeStateValue, ok := common.ParseCommonStatusState(volumeState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_state"].Desc, prometheus.GaugeValue, volumeStateValue, systemVolumeLabelValues...)
	}
	if volumeHealthStateValue, ok := common.ParseCommonStatusHealth(volumeHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_health_state"].Desc, prometheus.GaugeValue, volumeHealthStateValue, systemVolumeLabelValues...)
	}
	if volume.CapacityBytes != nil {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_capacity"].Desc, prometheus.GaugeValue, float64(*volume.CapacityBytes), systemVolumeLabelValues...)
	}
}
func parseDevice(ch chan<- prometheus.Metric, systemHostName string, device schemas.Device, wg *sync.WaitGroup) {
	defer wg.Done()
	deviceName := device.Name
	deviceState := device.Status.State
	deviceHealthState := device.Status.Health
	systemDeviceLabelValues := []string{systemHostName, "device", deviceName}
	if deviceStateValue, ok := common.ParseCommonStatusState(deviceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_simple_storage_device_state"].Desc, prometheus.GaugeValue, deviceStateValue, systemDeviceLabelValues...)
	}
	if deviceHealthStateValue, ok := common.ParseCommonStatusHealth(deviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_simple_storage_device_health_state"].Desc, prometheus.GaugeValue, deviceHealthStateValue, systemDeviceLabelValues...)
	}
}
func parseDrive(ch chan<- prometheus.Metric, systemHostName string, drive *schemas.Drive, wg *sync.WaitGroup) {
	defer wg.Done()
	driveName := drive.Name
	driveID := drive.ID
	driveState := drive.Status.State
	driveHealthState := drive.Status.Health
	systemdriveLabelValues := []string{systemHostName, "drive", driveName, driveID}
	if driveStateValue, ok := common.ParseCommonStatusState(driveState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_state"].Desc, prometheus.GaugeValue, driveStateValue, systemdriveLabelValues...)
	}
	if driveHealthStateValue, ok := common.ParseCommonStatusHealth(driveHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_health_state"].Desc, prometheus.GaugeValue, driveHealthStateValue, systemdriveLabelValues...)
	}
	if drive.CapacityBytes != nil {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_capacity"].Desc, prometheus.GaugeValue, float64(*drive.CapacityBytes), systemdriveLabelValues...)
	}
}

func parsePcieDevice(ch chan<- prometheus.Metric, systemHostName string, pcieDevice *schemas.PCIeDevice, wg *sync.WaitGroup) {
	defer wg.Done()
	pcieDeviceName := pcieDevice.Name
	pcieDeviceID := pcieDevice.ID
	pcieDeviceState := pcieDevice.Status.State
	pcieDeviceHealthState := pcieDevice.Status.Health
	pcieDevicePartNumber := pcieDevice.PartNumber
	pcieDeviceType := fmt.Sprint(pcieDevice.DeviceType)
	pcieSerialNumber := pcieDevice.SerialNumber
	systemPCIeDeviceLabelValues := []string{systemHostName, "pcie_device", pcieDeviceName, pcieDeviceID, pcieDevicePartNumber, pcieDeviceType, pcieSerialNumber}

	if pcieStateVaule, ok := common.ParseCommonStatusState(pcieDeviceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_device_state"].Desc, prometheus.GaugeValue, pcieStateVaule, systemPCIeDeviceLabelValues...)
	}
	if pcieHealthStateVaule, ok := common.ParseCommonStatusHealth(pcieDeviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_device_health_state"].Desc, prometheus.GaugeValue, pcieHealthStateVaule, systemPCIeDeviceLabelValues...)
	}
}

func parseNetworkInterface(ch chan<- prometheus.Metric, systemHostName string, networkInterface *schemas.NetworkInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	networkInterfaceName := networkInterface.Name
	networkInterfaceID := networkInterface.ID
	networkInterfaceState := networkInterface.Status.State
	networkInterfaceHealthState := networkInterface.Status.Health
	systemNetworkInterfaceLabelValues := []string{systemHostName, "network_interface", networkInterfaceName, networkInterfaceID}

	if networknetworkInterfaceStateVaule, ok := common.ParseCommonStatusState(networkInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_network_interface_state"].Desc, prometheus.GaugeValue, networknetworkInterfaceStateVaule, systemNetworkInterfaceLabelValues...)
	}
	if networknetworkInterfaceHealthStateVaule, ok := common.ParseCommonStatusHealth(networkInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_network_interface_health_state"].Desc, prometheus.GaugeValue, networknetworkInterfaceHealthStateVaule, systemNetworkInterfaceLabelValues...)
	}
}

func parseEthernetInterface(ch chan<- prometheus.Metric, systemHostName string, ethernetInterface *schemas.EthernetInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	//ethernetInterfaceODataIDslice := strings.Split(ethernetInterface.ODataID, "/")
	//ethernetInterfaceName := ethernetInterfaceODataIDslice[len(ethernetInterfaceODataIDslice)-1]
	ethernetInterfaceName := ethernetInterface.Name
	ethernetInterfaceID := ethernetInterface.ID
	ethernetInterfaceLinkStatus := ethernetInterface.LinkStatus
	ethernetInterfaceEnabled := ethernetInterface.InterfaceEnabled
	ethernetInterfaceSpeed := fmt.Sprintf("%d Mbps", ethernetInterface.SpeedMbps)
	ethernetInterfaceState := ethernetInterface.Status.State
	ethernetInterfaceHealthState := ethernetInterface.Status.Health
	systemEthernetInterfaceLabelValues := []string{systemHostName, "ethernet_interface", ethernetInterfaceName, ethernetInterfaceID, ethernetInterfaceSpeed}
	if ethernetInterfaceStateValue, ok := common.ParseCommonStatusState(ethernetInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_state"].Desc, prometheus.GaugeValue, ethernetInterfaceStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceHealthStateValue, ok := common.ParseCommonStatusHealth(ethernetInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_health_state"].Desc, prometheus.GaugeValue, ethernetInterfaceHealthStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceLinkStatusValue, ok := parseLinkStatus(ethernetInterfaceLinkStatus); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_link_status"].Desc, prometheus.GaugeValue, ethernetInterfaceLinkStatusValue, systemEthernetInterfaceLabelValues...)
	}

	ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_link_enabled"].Desc, prometheus.GaugeValue, boolToFloat64(ethernetInterfaceEnabled), systemEthernetInterfaceLabelValues...)
}

func parseLinkStatus(status schemas.LinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("LinkUp")) {
		return float64(1), true
	} else if bytes.Equal([]byte(status), []byte("NoLink")) {
		return float64(2), true
	} else if bytes.Equal([]byte(status), []byte("LinkDown")) {
		return float64(3), true
	}
	return float64(0), false
}

func boolToFloat64(data bool) float64 {

	if data {
		return float64(1)
	}
	return float64(0)

}

func parsePcieFunction(ch chan<- prometheus.Metric, systemHostName string, pcieFunction *schemas.PCIeFunction, wg *sync.WaitGroup) {
	defer wg.Done()
	pcieFunctionName := pcieFunction.Name
	pcieFunctionID := fmt.Sprint(pcieFunction.ID)
	pciFunctionDeviceclass := fmt.Sprint(pcieFunction.DeviceClass)
	pciFunctionType := fmt.Sprint(pcieFunction.FunctionType)
	pciFunctionState := pcieFunction.Status.State
	pciFunctionHealthState := pcieFunction.Status.Health

	systemPCIeFunctionLabelLabelValues := []string{systemHostName, "pcie_function", pcieFunctionName, pcieFunctionID, pciFunctionDeviceclass, pciFunctionType}

	if pciFunctionStateValue, ok := common.ParseCommonStatusState(pciFunctionState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_state"].Desc, prometheus.GaugeValue, pciFunctionStateValue, systemPCIeFunctionLabelLabelValues...)
	}

	if pciFunctionHealthStateValue, ok := common.ParseCommonStatusHealth(pciFunctionHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_health_state"].Desc, prometheus.GaugeValue, pciFunctionHealthStateValue, systemPCIeFunctionLabelLabelValues...)
	}
}
