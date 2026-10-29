package fans

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis/thermal_common"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish/schemas"
)

// Implements the prometheus.Collector.
type Collector struct {
	Ctx              *common.CollectionContext
	Log              *slog.Logger
	thermalSubsystem *schemas.ThermalSubsystem
	chassisID        *string
}

// NewCollector returns a collector that collects thermal subsystem fan data
func NewCollector(ctx *common.CollectionContext, logger *slog.Logger, chassisID *string, thermalSubsystem *schemas.ThermalSubsystem) *Collector {
	return &Collector{
		Ctx: ctx,
		Log: logger.With(
			"collector", "chassis.thermal_subsystem.fans",
		),
		thermalSubsystem: thermalSubsystem,
		chassisID:        chassisID,
	}
}

// Describe implemented prometheus.Collector
func Describe(ch chan<- *prometheus.Desc) {
	thermal_common.Describe(ch)
}

// Describe implemented prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	Describe(ch)
}

// Given a ThermalSubsystem, collect data from its Fans
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	chassisFans, err := c.thermalSubsystem.Fans()
	if err != nil {
		c.Log.Error("Error getting fans from thermal subsystem", "id", c.thermalSubsystem.ID, "err", err)
		return
	}

	wg := sync.WaitGroup{}
	for _, chassisFan := range chassisFans {
		wg.Go(func() { parseFan(ch, c.chassisID, chassisFan) })
	}

	wg.Wait()
}

func parseFan(ch chan<- prometheus.Metric, chassisID *string, chassisFan *schemas.Fan) {
	chassisFanLabelvalues := []string{"fan", *chassisID, chassisFan.Name, chassisFan.ID}

	if chassisFan.SpeedPercent.SpeedRPM != nil {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm"].Desc, prometheus.GaugeValue, *chassisFan.SpeedPercent.SpeedRPM, chassisFanLabelvalues...)
	} else if chassisFan.SpeedPercent.Reading != nil && chassisFan.RatedSpeedRPM != nil { // Calculate from percentage and rating
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm"].Desc, prometheus.GaugeValue, (*chassisFan.SpeedPercent.Reading/100.0)*float64(*chassisFan.RatedSpeedRPM), chassisFanLabelvalues...)
	}

	if chassisFan.SpeedPercent.Reading != nil {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm_percentage"].Desc, prometheus.GaugeValue, *chassisFan.SpeedPercent.Reading, chassisFanLabelvalues...)
	} else if chassisFan.SpeedPercent.SpeedRPM != nil && chassisFan.RatedSpeedRPM != nil { // Calculate from RPM and rating
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm_percentage"].Desc, prometheus.GaugeValue, 100.0*(*chassisFan.SpeedPercent.SpeedRPM/float64(*chassisFan.RatedSpeedRPM)), chassisFanLabelvalues...)
	}

	if chassisFan.RatedSpeedRPM != nil {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_rpm_max"].Desc, prometheus.GaugeValue, float64(*chassisFan.RatedSpeedRPM), chassisFanLabelvalues...)
	}

	if chassisFanStatusHealthValue, ok := common.ParseCommonStatusHealth(chassisFan.Status.Health); ok {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_health"].Desc, prometheus.GaugeValue, chassisFanStatusHealthValue, chassisFanLabelvalues...)
	}

	if chassisFanStatusStateValue, ok := common.ParseCommonStatusState(chassisFan.Status.State); ok {
		ch <- prometheus.MustNewConstMetric(thermal_common.MetricMap["chassis_fan_state"].Desc, prometheus.GaugeValue, chassisFanStatusStateValue, chassisFanLabelvalues...)
	}
}
