package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

var mocksForAllTests = []string{"DSP2043-server", "proposed-ocp-profile", "public-bladed", "public-rackmount"}

func TestScrape(t *testing.T) {
	for _, path := range mocksForAllTests {
		collector := CreateStaticBMC(t, path, false)

		ch := make(chan prometheus.Metric)

		go collector.Collect(ch)

		<-ch
	}
}
