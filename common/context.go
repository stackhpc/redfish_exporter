package common

import (
	"fmt"
	"log/slog"
	"time"

	gofish "github.com/stmcginnis/gofish"
)

// Metric name parts.
const (
	// Exporter namespace.
	Namespace = "redfish"
	// Subsystem(s).
	exporter = "exporter"
	// Math constant for picoseconds to seconds.
	picoSeconds = 1e12
)

type CollectionContext struct {
	RedfishClient   *gofish.APIClient
	CollectLogs     bool
	LogCount        map[string]int
	DisabledMetrics []string
}

func NewCollectionContext(target string, hostconfig *HostConfig, logger *slog.Logger, collectLogs bool) (*CollectionContext, error) {
	startTime := time.Now()
	client, err := newRedfishClient(target, hostconfig.Username, hostconfig.Password)
	if err != nil {
		logger.Error("error creating redfish client", "err", err)
		return nil, err
	}
	duration := time.Now().Sub(startTime)
	logger.Debug("Created new client", "duration", duration)

	logCount := hostconfig.Logcount
	disabledMetrics := hostconfig.DisabledMetrics

	// TODO.. query parameter could logcount_<logServiceName>=10

	return &CollectionContext{RedfishClient: client, CollectLogs: collectLogs, LogCount: logCount, DisabledMetrics: disabledMetrics}, nil
}

func newRedfishClient(host string, username string, password string) (*gofish.APIClient, error) {

	url := fmt.Sprintf("https://%s", host)

	config := gofish.ClientConfig{
		Endpoint:              url,
		Username:              username,
		Password:              password,
		Insecure:              true,
		MaxConcurrentRequests: 1024,
		ReuseConnections:      true,
	}
	redfishClient, err := gofish.Connect(config)
	if err != nil {
		return nil, err
	}
	return redfishClient, nil
}
