package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

var mocksForAllTests = []string{"DSP2043_1.0.0/DSP2043-server", "DSP2043_2026.2/public-rackmount1"}

func TestScrape(t *testing.T) {
	for _, path := range mocksForAllTests {
		collector := CreateStaticBMC(t, path, false)

		ch := make(chan prometheus.Metric)

		go collector.Collect(ch)

		<-ch
	}
}
