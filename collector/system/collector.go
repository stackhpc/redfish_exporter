package system

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/system/ethernet_interfaces"
	"github.com/stackhpc/redfish_exporter/collector/system/memory"
	"github.com/stackhpc/redfish_exporter/collector/system/network_interfaces"
	"github.com/stackhpc/redfish_exporter/collector/system/pcie_devices"
	"github.com/stackhpc/redfish_exporter/collector/system/pcie_functions"
	"github.com/stackhpc/redfish_exporter/collector/system/processors"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

// SystemSubsystem is the system subsystem
var (
	SystemSubsystem                   = "system"
	SystemLabelNames                  = []string{"hostname", "resource", "system_id"}
	SystemBIOSLabelNames              = append(slices.Clone(SystemLabelNames), "bios_version")
	SystemVolumeLabelNames            = []string{"hostname", "resource", "volume", "volume_id"}
	SystemDeviceLabelNames            = []string{"hostname", "resource", "device"}
	SystemDriveLabelNames             = []string{"hostname", "resource", "drive", "drive_id"}
	SystemStorageControllerLabelNames = []string{"hostname", "resource", "storage_controller", "storage_controller_id"}

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

			wg3 := &sync.WaitGroup{}  // storage/volumes
			wg4 := &sync.WaitGroup{}  // storage/drives
			wg8 := &sync.WaitGroup{}  // simple storage

			wg := &sync.WaitGroup{}

			if !slices.Contains(disabledMetrics, "memory") {
				collector := memory.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "processor") {
				collector := processors.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "storage") {
				processStorage(system, systemLogContext, wg3, ch, systemHostName, wg4)
			}

			if !slices.Contains(disabledMetrics, "pcie_devices") {
				collector := pcie_devices.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "network_interfaces") {
				collector := network_interfaces.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "ethernet_interfaces") {
				collector := ethernet_interfaces.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			if !slices.Contains(disabledMetrics, "simple_storage") {
				processSimpleStorage(system, systemLogContext, wg8, ch, systemHostName)
			}

			if !slices.Contains(disabledMetrics, "pcie_functions") {
				collector := pcie_functions.NewCollector(s.Ctx, s.Log, system)
				wg.Go(func() { collector.Collect(ch) })
			}

			// process log services
			if collectLogs {
				wg.Go(func() { s.collectLogs(ch, system) })
			}

			wg3.Wait()
			wg4.Wait()
			wg8.Wait()

			wg.Wait()

			systemLogContext.Info("collector scrape completed")
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	}
}

func (c* SystemCollector) collectLogs(ch chan<- prometheus.Metric, system *schemas.ComputerSystem) {
	logServices, err := system.LogServices()
	if err != nil {
		c.Log.Error("error getting log services from system", "operation", "system.LogServices()", "err", err)
		return
	} else if logServices == nil {
		c.Log.Info("no log services found", "operation", "system.LogServices()")
		return
	}

	wg := &sync.WaitGroup{}
	for _, logService := range logServices {
		wg.Go(func() {
			if err = common.ParseLogService(ch, systemMetrics, c.Ctx, c.Log.With("log_service", logService.ID), SystemSubsystem, system.ID, logService); err != nil {
				c.Log.Error("error getting log entries from log service", "operation", "chassis.LogServices()", "err", err)
			}
		})
	}

	wg.Wait()
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
