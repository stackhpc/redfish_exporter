package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestScrape(t *testing.T) {
	collector := CreateStaticBMC(t, "DSP2043-server", false)

	ch := make(chan prometheus.Metric)

	go collector.Collect(ch)

	metric := <- ch

	t.Log(metric)
}
