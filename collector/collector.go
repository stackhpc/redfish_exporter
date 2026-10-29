package collector

import (
	"bytes"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stackhpc/redfish_exporter/collector/chassis"
	"github.com/stackhpc/redfish_exporter/collector/manager"
	"github.com/stackhpc/redfish_exporter/collector/system"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

// Metric name parts.
const (
	// Exporter namespace.
	namespace = "redfish"
	// Subsystem(s).
	exporter = "exporter"
	// Math constant for picoseconds to seconds.
	picoSeconds = 1e12
)

// Metric descriptors.
var (
	totalScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		nil, nil,
	)
	upDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"redfish up",
		nil,
		nil,
	)
)

// RedfishCollector collects redfish metrics. It implements prometheus.Collector.
type RedfishCollector struct {
	redfishClient *gofish.APIClient
	collectors    map[string]prometheus.Collector
	redfishUp     prometheus.Gauge
}

// NewRedfishCollector return RedfishCollector
func NewRedfishCollector(ctx *common.CollectionContext, logger *slog.Logger) *RedfishCollector {
	var collectors map[string]prometheus.Collector
	chassisCollector := chassis.NewCollector(ctx, logger)
	systemCollector := system.NewCollector(ctx, logger)
	managerCollector := manager.NewCollector(ctx, logger)

	collectors = map[string]prometheus.Collector{"chassis": chassisCollector, "system": systemCollector, "manager": managerCollector}

	return &RedfishCollector{
		redfishClient: ctx.RedfishClient,
		collectors:    collectors,
		redfishUp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "",
				Name:      "up",
				Help:      "redfish up",
			},
		),
	}
}

// Describe implements prometheus.Collector.
func (r *RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range r.collectors {
		collector.Describe(ch)
	}
	ch <- totalScrapeDurationDesc
	ch <- upDesc
}

// Collect implements prometheus.Collector.
func (r *RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeTime := time.Now()
	if r.redfishClient != nil {
		defer r.redfishClient.Logout()
		r.redfishUp.Set(1)
		wg := &sync.WaitGroup{}
		wg.Add(len(r.collectors))

		for _, collector := range r.collectors {
			go func(collector prometheus.Collector) {
				defer wg.Done()
				collector.Collect(ch)
			}(collector)
		}
		wg.Wait()
	} else {
		r.redfishUp.Set(0)
	}

	ch <- r.redfishUp
	ch <- prometheus.MustNewConstMetric(totalScrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
}

func parsePhySecReArmMethod(method schemas.IntrusionSensorReArm) (float64, bool) {
	if bytes.Equal([]byte(method), []byte("Manual")) {
		return float64(1), true
	}
	if bytes.Equal([]byte(method), []byte("Automatic")) {
		return float64(2), true
	}

	return float64(0), false
}
