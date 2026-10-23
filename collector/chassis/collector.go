package chassis

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/stackhpc/redfish_exporter/common"
)

// ChassisSubsystem is the chassis subsystem
var (
	ChassisSubsystem                  = "chassis"
	ChassisLabelNames                 = []string{"resource", "chassis_id"}
	ChassisModel                      = []string{"resource", "chassis_id", "manufacturer", "model", "part_number", "sku"}
	ChassisTemperatureLabelNames      = []string{"resource", "chassis_id", "sensor", "sensor_id"}
	ChassisFanLabelNames              = []string{"resource", "chassis_id", "fan", "fan_id", "fan_unit"}
	ChassisPowerVoltageLabelNames     = []string{"resource", "chassis_id", "power_voltage", "power_voltage_id"}
	ChassisPowerSupplyLabelNames      = []string{"resource", "chassis_id", "power_supply", "power_supply_id"}
	ChassisNetworkAdapterLabelNames   = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id"}
	ChassisNetworkPortLabelNames      = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id", "network_port", "network_port_id", "network_port_type", "network_port_speed", "network_port_connectiont_type", "network_physical_port_number"}
	ChassisPhysicalSecurityLabelNames = []string{"resource", "chassis_id", "intrusion_sensor_number", "intrusion_sensor_rearm"}

	ChassisLogServiceLabelNames = []string{"chassis_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	ChassisLogEntryLabelNames   = []string{"chassis_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	chassisMetrics = createChassisMetricMap()
)

// ChassisCollector implements the prometheus.Collector.
type ChassisCollector struct {
	Ctx                   *common.CollectionContext
	metrics               map[string]common.Metric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *slog.Logger
}

func createChassisMetricMap() map[string]common.Metric {
	chassisMetrics := make(map[string]common.Metric)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "health", fmt.Sprintf("health of chassis,%s", common.HealthHelp), ChassisLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "state", fmt.Sprintf("state of chassis,%s", common.StateHelp), ChassisLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "model_info", "organization responsible for producing the chassis, the name by which the manufacturer generally refers to the chassis, and a part number and sku assigned by the organization that is responsible for producing or manufacturing the chassis", ChassisModel)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "temperature_sensor_state", fmt.Sprintf("status state of temperature on this chassis component,%s", common.StateHelp), ChassisTemperatureLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "temperature_sensor_health", fmt.Sprintf("status health of temperature on this chassis component,%s", common.StateHelp), ChassisTemperatureLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "temperature_celsius", "celsius of temperature on this chassis component", ChassisTemperatureLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_health", fmt.Sprintf("fan health on this chassis component,%s", common.HealthHelp), ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_state", fmt.Sprintf("fan state on this chassis component,%s", common.StateHelp), ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm", "fan RPM or percentage on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_percentage", "fan RPM, as a percentage of the min-max RPMs possible, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_min", "lowest possible fan RPM or percentage, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_max", "highest possible fan RPM or percentage, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_lower_threshold_critical", "threshold below the normal range fan RPM or percentage, but not fatal, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_lower_threshold_non_critical", "threshold below the normal range fan RPM or percentage, but not critical, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_lower_threshold_fatal", "threshold below the normal range fan RPM or percentage, and is fatal, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_upper_threshold_critical", "threshold above the normal range fan RPM or percentage, but not fatal, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_upper_threshold_non_critical", "threshold above the normal range fan RPM or percentage, but not critical, on this chassis component", ChassisFanLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "fan_rpm_upper_threshold_fatal", "threshold above the normal range fan RPM or percentage, and is fatal, on this chassis component", ChassisFanLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_voltage_state", fmt.Sprintf("power voltage state of chassis component,%s", common.StateHelp), ChassisPowerVoltageLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_voltage_volts", "power voltage volts number of chassis component", ChassisPowerVoltageLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_average_consumed_watts", "power wattage watts number of chassis component", ChassisPowerVoltageLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_state", fmt.Sprintf("powersupply state of chassis component,%s", common.StateHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_health", fmt.Sprintf("powersupply health of chassis component,%s", common.HealthHelp), ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_power_efficiency_percentage", "rated efficiency, as a percentage, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_last_power_output_watts", "average power output, measured in Watts, of the associated power supply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_power_input_watts", "measured input power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_power_output_watts", "measured output power, in Watts, of powersupply on this chassis", ChassisPowerSupplyLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "power_powersupply_power_capacity_watts", "power_capacity_watts of powersupply on this chassis", ChassisPowerSupplyLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "network_adapter_state", fmt.Sprintf("chassis network adapter state,%s", common.StateHelp), ChassisNetworkAdapterLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "network_adapter_health_state", fmt.Sprintf("chassis network adapter health state,%s", common.HealthHelp), ChassisNetworkAdapterLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "network_port_state", fmt.Sprintf("chassis network port state,%s", common.StateHelp), ChassisNetworkPortLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "network_port_health_state", fmt.Sprintf("chassis network port health state,%s", common.HealthHelp), ChassisNetworkPortLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "network_port_link_state", fmt.Sprintf("chassis network port link state state,%s", common.PortLinkHelp), ChassisNetworkPortLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "physical_security_sensor_state", fmt.Sprintf("indicates the known state of the physical security sensor, such as if it is hardware intrusion detected,%s", common.IntrusionSensorHelp), ChassisPhysicalSecurityLabelNames)

	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "log_service_state", fmt.Sprintf("chassis log service state,%s", common.StateHelp), ChassisLogServiceLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "log_service_health_state", fmt.Sprintf("chassis log service health state,%s", common.HealthHelp), ChassisLogServiceLabelNames)
	common.AddToMetricMap(chassisMetrics, ChassisSubsystem, "log_entry_severity_state", fmt.Sprintf("chassis log entry severity state,%s", common.SeverityHelp), ChassisLogEntryLabelNames)

	return chassisMetrics
}

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(ctx *common.CollectionContext, logger *slog.Logger) *ChassisCollector {
	// get service from redfish client

	return &ChassisCollector{
		metrics: chassisMetrics,
		Ctx:     ctx,
		Log: logger.With(
			"collector", "ChassisCollector",
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

// Describe implemented prometheus.Collector
func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
	c.collectorScrapeStatus.Describe(ch)

}

// Collect implemented prometheus.Collector
func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := c.Log
	collectLogs := c.Ctx.CollectLogs
	service := c.Ctx.RedfishClient.Service

	// get a list of chassis from service
	if chassises, err := service.Chassis(); err != nil {
		collectorLogContext.Error("error getting chassis from service", "operation", "service.Chassis()", "err", err)
	} else {
		// process the chassises
		for _, chassis := range chassises {
			chassisLogContext := collectorLogContext.With("Chassis", chassis.ID)
			chassisLogContext.Info("collector scrape started")
			chassisID := chassis.ID
			chassisStatus := chassis.Status
			chassisStatusState := chassisStatus.State
			chassisStatusHealth := chassisStatus.Health
			ChassisLabelValues := []string{"chassis", chassisID}
			if chassisStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisStatusHealth); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].Desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
			}
			if chassisStatusStateValue, ok := common.ParseCommonStatusState(chassisStatusState); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].Desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
			}

			chassisManufacturer := chassis.Manufacturer
			chassisModel := chassis.Model
			chassisPartNumber := chassis.PartNumber
			chassisSKU := chassis.SKU
			ChassisModelLabelValues := []string{"chassis", chassisID, chassisManufacturer, chassisModel, chassisPartNumber, chassisSKU}
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_model_info"].Desc, prometheus.GaugeValue, 1, ChassisModelLabelValues...)

			chassisThermal, err := chassis.Thermal()
			if err != nil {
				chassisLogContext.Error("error getting thermal data from chassis", "operation", "chassis.Thermal()", "err", err)
			} else if chassisThermal == nil {
				chassisLogContext.Info("no thermal data found", "operation", "chassis.Thermal()")
			} else {
				// process temperature
				chassisTemperatures := chassisThermal.Temperatures
				wg := &sync.WaitGroup{}
				wg.Add(len(chassisTemperatures))

				for _, chassisTemperature := range chassisTemperatures {
					go parseChassisTemperature(ch, chassisID, chassisTemperature, wg)
				}

				// process fans

				chassisFans := chassisThermal.Fans
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(chassisFans))
				for _, chassisFan := range chassisFans {
					go parseChassisFan(ch, chassisID, chassisFan, wg2)
				}
			}

			chassisPowerInfo, err := chassis.Power()
			if err != nil {
				chassisLogContext.Error("error getting power data from chassis", "operation", "chassis.Power()", "err", err)
			} else if chassisPowerInfo == nil {
				chassisLogContext.Info("no power data found", "operation", "chassis.Power()")
			} else {
				// power voltages
				chassisPowerInfoVoltages := chassisPowerInfo.Voltages
				wg3 := &sync.WaitGroup{}
				wg3.Add(len(chassisPowerInfoVoltages))
				for _, chassisPowerInfoVoltage := range chassisPowerInfoVoltages {
					go parseChassisPowerInfoVoltage(ch, chassisID, chassisPowerInfoVoltage, wg3)
				}

				// power control
				chassisPowerInfoPowerControls := chassisPowerInfo.PowerControl
				wg4 := &sync.WaitGroup{}
				wg4.Add(len(chassisPowerInfoPowerControls))
				for _, chassisPowerInfoPowerControl := range chassisPowerInfoPowerControls {
					go parseChassisPowerInfoPowerControl(ch, chassisID, chassisPowerInfoPowerControl, wg4)
				}

				// powerSupply
				chassisPowerInfoPowerSupplies := chassisPowerInfo.PowerSupplies
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(chassisPowerInfoPowerSupplies))
				for _, chassisPowerInfoPowerSupply := range chassisPowerInfoPowerSupplies {
					go parseChassisPowerInfoPowerSupply(ch, chassisID, chassisPowerInfoPowerSupply, wg5)
				}
			}

			// process NetapAdapter

			networkAdapters, err := chassis.NetworkAdapters()
			if err != nil {
				chassisLogContext.Error("error getting network adapters data from chassis", "operation", "chassis.NetworkAdapters()", "err", err)
			} else if networkAdapters == nil {
				chassisLogContext.Info("no network adapters data found", "operation", "chassis.NetworkAdapters()")
			} else {
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(networkAdapters))

				for _, networkAdapter := range networkAdapters {
					if err = parseNetworkAdapter(ch, chassisID, networkAdapter, wg5); err != nil {
						chassisLogContext.Error("error getting network ports from network adapter", "operation", "chassis.NetworkAdapters()", "err", err)
					}
				}
			}

			physicalSecurity := chassis.PhysicalSecurity
			if physicalSecurity != (schemas.PhysicalSecurity{}) {
				physicalSecurityIntrusionSensor := physicalSecurity.IntrusionSensor
				physicalSecurityIntrusionSensorNumber := fmt.Sprint(physicalSecurity.IntrusionSensorNumber)
				physicalSecurityIntrusionSensorReArmMethod := string(physicalSecurity.IntrusionSensorReArm)

				if phySecIntrusionSensor, ok := parsePhySecIntrusionSensor(physicalSecurityIntrusionSensor); ok {
					ChassisPhysicalSecurityLabelValues := []string{"physical_security", chassisID, physicalSecurityIntrusionSensorNumber, physicalSecurityIntrusionSensorReArmMethod}
					ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_physical_security_sensor_state"].Desc, prometheus.GaugeValue, phySecIntrusionSensor, ChassisPhysicalSecurityLabelValues...)
				}
			}

			// process log services
			if collectLogs {
				logServices, err := chassis.LogServices()
				if err != nil {
					chassisLogContext.Error("error getting log services from chassis", "operation", "chassis.LogServices()", "err", err)
				} else if logServices == nil {
					chassisLogContext.Info("no log services found", "operation", "chassis.LogServices()")
				} else {
					wg6 := &sync.WaitGroup{}
					wg6.Add(len(logServices))

					for _, logService := range logServices {
						if err = common.ParseLogService(ch, chassisMetrics, c.Ctx, chassisLogContext, ChassisSubsystem, chassisID, logService, wg6); err != nil {
							chassisLogContext.Error("error getting log entries from log service", "operation", "chassis.LogServices()", "err", err)
						}
					}
				}
			}
			chassisLogContext.Info("collector scrape completed")
		}
	}

	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}

func parseChassisTemperature(ch chan<- prometheus.Metric, chassisID string, chassisTemperature schemas.Temperature, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisTemperatureSensorName := chassisTemperature.Name
	chassisTemperatureSensorID := chassisTemperature.MemberID
	chassisTemperatureStatus := chassisTemperature.Status
	chassisTemperatureLabelvalues := []string{"temperature", chassisID, chassisTemperatureSensorName, chassisTemperatureSensorID}

	chassisTemperatureStatusHealth := chassisTemperatureStatus.Health
	if chassisTemperatureStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisTemperatureStatusHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_sensor_health"].Desc, prometheus.GaugeValue, chassisTemperatureStatusHealthValue, chassisTemperatureLabelvalues...)
	}

	chassisTemperatureStatusState := chassisTemperatureStatus.State
	//			chassisTemperatureStatusLabelNames :=[]string{BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
	//		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_status_health"].Desc, prometheus.GaugeValue, parseCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureLabelvalues...)
	if chassisTemperatureStatusStateValue, ok := common.ParseCommonStatusState(chassisTemperatureStatusState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_sensor_state"].Desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
	}

	if chassisTemperature.ReadingCelsius != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_celsius"].Desc, prometheus.GaugeValue, float64(*chassisTemperature.ReadingCelsius), chassisTemperatureLabelvalues...)
	}
}

func parseChassisFan(ch chan<- prometheus.Metric, chassisID string, chassisFan schemas.ThermalFan, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisFanID := chassisFan.MemberID
	chassisFanLabelvalues := []string{"fan", chassisID, chassisFan.Name, chassisFanID, strings.ToLower(string(chassisFan.ReadingUnits))} // e.g. RPM -> rpm, Percentage -> percentage

	var chassisFanPercentage *float64
	switch chassisFan.ReadingUnits {
	case schemas.PercentReadingUnits:
		if chassisFan.Reading == nil {
			break
		}
		chassisFanPercentage = new(float64)
		*chassisFanPercentage = float64(*chassisFan.Reading)
	case schemas.RPMReadingUnits: // Calculate fan percentage from best value for min/max
		if chassisFan.Reading == nil {
			break
		}
		// Some vendors (e.g. PowerEdge C6420) report null RPMs for Min/Max, as well as Lower/UpperFatal,
		// but provide Lower/UpperCritical, so we use the most extreme value we have as the limits.
		var min *int = chassisFan.MinReadingRange
		var max *int = chassisFan.MaxReadingRange
		if chassisFan.LowerThresholdCritical != nil && (min == nil || *min > *chassisFan.LowerThresholdCritical) {
			min = chassisFan.LowerThresholdCritical
		}
		if chassisFan.LowerThresholdFatal != nil && (min == nil || *min > *chassisFan.LowerThresholdFatal) {
			min = chassisFan.LowerThresholdFatal
		}
		if chassisFan.LowerThresholdNonCritical != nil && (min == nil || *min > *chassisFan.LowerThresholdNonCritical) {
			min = chassisFan.LowerThresholdNonCritical
		}
		if chassisFan.UpperThresholdCritical != nil && (max == nil || *max < *chassisFan.UpperThresholdCritical) {
			max = chassisFan.UpperThresholdCritical
		}
		if chassisFan.UpperThresholdFatal != nil && (max == nil || *max < *chassisFan.UpperThresholdFatal) {
			max = chassisFan.UpperThresholdFatal
		}
		if chassisFan.UpperThresholdNonCritical != nil && (max == nil || *max < *chassisFan.UpperThresholdNonCritical) {
			max = chassisFan.UpperThresholdNonCritical
		}

		rpm_f := float64(*chassisFan.Reading)
		if min != nil && max != nil {
			min_f := float64(*min)
			max_f := float64(*max)
			chassisFanPercentage = new(float64)
			*chassisFanPercentage = (rpm_f - min_f) / (max_f - min_f) * 100.0
		}

		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm"].Desc, prometheus.GaugeValue, rpm_f, chassisFanLabelvalues...)
	}

	if chassisFanStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisFan.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_health"].Desc, prometheus.GaugeValue, chassisFanStatusHealthValue, chassisFanLabelvalues...)
	}

	if chassisFanStatusStateValue, ok := common.ParseCommonStatusState(chassisFan.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_state"].Desc, prometheus.GaugeValue, chassisFanStatusStateValue, chassisFanLabelvalues...)
	}
	if chassisFan.MinReadingRange != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_min"].Desc, prometheus.GaugeValue, float64(*chassisFan.MinReadingRange), chassisFanLabelvalues...)
	}
	if chassisFan.MaxReadingRange != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_max"].Desc, prometheus.GaugeValue, float64(*chassisFan.MaxReadingRange), chassisFanLabelvalues...)
	}
	if chassisFanPercentage != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_percentage"].Desc, prometheus.GaugeValue, *chassisFanPercentage, chassisFanLabelvalues...)
	}
	if chassisFan.LowerThresholdCritical != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_lower_threshold_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.LowerThresholdCritical), chassisFanLabelvalues...)
	}
	if chassisFan.LowerThresholdFatal != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_lower_threshold_fatal"].Desc, prometheus.GaugeValue, float64(*chassisFan.LowerThresholdFatal), chassisFanLabelvalues...)
	}
	if chassisFan.LowerThresholdNonCritical != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_lower_threshold_non_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.LowerThresholdNonCritical), chassisFanLabelvalues...)
	}
	if chassisFan.UpperThresholdCritical != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_upper_threshold_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.UpperThresholdCritical), chassisFanLabelvalues...)
	}
	if chassisFan.UpperThresholdFatal != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_upper_threshold_fatal"].Desc, prometheus.GaugeValue, float64(*chassisFan.UpperThresholdFatal), chassisFanLabelvalues...)
	}
	if chassisFan.UpperThresholdNonCritical != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm_upper_threshold_non_critical"].Desc, prometheus.GaugeValue, float64(*chassisFan.UpperThresholdNonCritical), chassisFanLabelvalues...)
	}
}

func parseChassisPowerInfoVoltage(ch chan<- prometheus.Metric, chassisID string, voltageInfo schemas.Voltage, wg *sync.WaitGroup) {
	defer wg.Done()
	voltageInfoID := voltageInfo.MemberID
	chassisPowerVoltageLabelvalues := []string{"power_voltage", chassisID, voltageInfo.Name, voltageInfoID}
	if chassisPowerInfoVoltageStateValue, ok := common.ParseCommonStatusState(voltageInfo.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_voltage_state"].Desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, chassisPowerVoltageLabelvalues...)
	}
	if voltageInfo.ReadingVolts != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_voltage_volts"].Desc, prometheus.GaugeValue, float64(*voltageInfo.ReadingVolts), chassisPowerVoltageLabelvalues...)
	}
}

func parseChassisPowerInfoPowerControl(ch chan<- prometheus.Metric, chassisID string, powerControl schemas.PowerControl, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisPowerVoltageLabelvalues := []string{"power_wattage", chassisID, powerControl.Name, powerControl.MemberID}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_average_consumed_watts"].Desc, prometheus.GaugeValue, float64(*powerControl.PowerMetrics.AverageConsumedWatts), chassisPowerVoltageLabelvalues...)
}

func parseChassisPowerInfoPowerSupply(ch chan<- prometheus.Metric, chassisID string, powerSupply schemas.PowerSupply, wg *sync.WaitGroup) {
	defer wg.Done()
	powerSupplyID := powerSupply.MemberID

	chassisPowerInfoPowerSupplyState := powerSupply.Status.State
	chassisPowerInfoPowerSupplyHealth := powerSupply.Status.Health
	chassisPowerSupplyLabelvalues := []string{"power_supply", chassisID, powerSupply.Name, powerSupplyID}
	if chassisPowerInfoPowerSupplyStateValue, ok := common.ParseCommonStatusState(chassisPowerInfoPowerSupplyState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_state"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, chassisPowerSupplyLabelvalues...)
	}
	if chassisPowerInfoPowerSupplyHealthValue, ok := common.ParseCommonStatusHealth(chassisPowerInfoPowerSupplyHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_health"].Desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, chassisPowerSupplyLabelvalues...)
	}
	if powerSupply.EfficiencyPercent != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_efficiency_percentage"].Desc, prometheus.GaugeValue, float64(*powerSupply.EfficiencyPercent), chassisPowerSupplyLabelvalues...)
	}
	if powerSupply.LastPowerOutputWatts != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_last_power_output_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.LastPowerOutputWatts), chassisPowerSupplyLabelvalues...)
	}
	if powerSupply.PowerCapacityWatts != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_capacity_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerCapacityWatts), chassisPowerSupplyLabelvalues...)
	}
	if powerSupply.PowerInputWatts != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_input_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerInputWatts), chassisPowerSupplyLabelvalues...)
	}
	if powerSupply.PowerOutputWatts != nil {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_output_watts"].Desc, prometheus.GaugeValue, float64(*powerSupply.PowerOutputWatts), chassisPowerSupplyLabelvalues...)
	}
}

func parseNetworkAdapter(ch chan<- prometheus.Metric, chassisID string, networkAdapter *schemas.NetworkAdapter, wg *sync.WaitGroup) error {
	defer wg.Done()
	networkAdapterName := networkAdapter.Name
	networkAdapterID := networkAdapter.ID
	networkAdapterState := networkAdapter.Status.State
	networkAdapterHealthState := networkAdapter.Status.Health
	chassisNetworkAdapterLabelValues := []string{"network_adapter", chassisID, networkAdapterName, networkAdapterID}
	if networkAdapterStateValue, ok := common.ParseCommonStatusState(networkAdapterState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_adapter_state"].Desc, prometheus.GaugeValue, networkAdapterStateValue, chassisNetworkAdapterLabelValues...)
	}
	if networkAdapterHealthStateValue, ok := common.ParseCommonStatusHealth(networkAdapterHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_adapter_health_state"].Desc, prometheus.GaugeValue, networkAdapterHealthStateValue, chassisNetworkAdapterLabelValues...)
	}

	if networkPorts, err := networkAdapter.NetworkPorts(); err != nil {
		return err
	} else {
		wg6 := &sync.WaitGroup{}
		wg6.Add(len(networkPorts))
		for _, networkPort := range networkPorts {
			go parseNetworkPort(ch, chassisID, networkPort, networkAdapterName, networkAdapterID, wg6)
		}
		wg6.Wait()
	}
	return nil
}

func parsePortLinkStatus(status schemas.NetworkPortLinkStatus) (float64, bool) {
	if bytes.Equal([]byte(status), []byte("Up")) {
		return float64(1), true
	}
	return float64(0), false
}

func parseNetworkPort(ch chan<- prometheus.Metric, chassisID string, networkPort *schemas.NetworkPort, networkAdapterName string, networkAdapterID string, wg *sync.WaitGroup) {
	defer wg.Done()
	networkPortName := networkPort.Name
	networkPortID := networkPort.ID
	networkPortState := networkPort.Status.State
	networkLinkStatus := networkPort.LinkStatus
	networkPortLinkType := networkPort.ActiveLinkTechnology
	networkPortLinkSpeed := fmt.Sprintf("%d Mbps", networkPort.CurrentLinkSpeedMbps)
	networkPortHealthState := networkPort.Status.Health
	networkPortConnectionType := networkPort.FCPortConnectionType
	networkPhysicalPortNumber := networkPort.PhysicalPortNumber
	chassisNetworkPortLabelValues := []string{"network_port", chassisID, networkAdapterName, networkAdapterID, networkPortName, networkPortID, string(networkPortLinkType), networkPortLinkSpeed, string(networkPortConnectionType), networkPhysicalPortNumber}

	if networkLinkStatusValue, ok := parsePortLinkStatus(networkLinkStatus); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_link_state"].Desc, prometheus.GaugeValue, networkLinkStatusValue, chassisNetworkPortLabelValues...)
	}

	if networkPortStateValue, ok := common.ParseCommonStatusState(networkPortState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_state"].Desc, prometheus.GaugeValue, networkPortStateValue, chassisNetworkPortLabelValues...)
	}
	if networkPortHealthStateValue, ok := common.ParseCommonStatusHealth(networkPortHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_health_state"].Desc, prometheus.GaugeValue, networkPortHealthStateValue, chassisNetworkPortLabelValues...)
	}
}

func parsePhySecIntrusionSensor(method schemas.IntrusionSensor) (float64, bool) {
	if bytes.Equal([]byte(method), []byte("Normal")) {
		return float64(1), true

	}
	if bytes.Equal([]byte(method), []byte("TamperingDetected")) {
		return float64(2), true
	}
	if bytes.Equal([]byte(method), []byte("HardwareIntrusion")) {
		return float64(3), true
	}

	return float64(0), false
}
