package common

import (
	"fmt"
	"log/slog"

	gofish "github.com/stmcginnis/gofish"
)

type CollectionContext struct {
	RedfishClient   *gofish.APIClient
	CollectLogs     bool
	LogCount        map[string]int
	DisabledMetrics []string
}

func NewCollectionContext(target string, hostconfig *HostConfig, logger *slog.Logger, collectLogs bool) (*CollectionContext, error) {
	client, err := newRedfishClient(target, hostconfig.Username, hostconfig.Password)
	if err != nil {
		logger.Error("error creating redfish client", "err", err)
		return nil, err
	}

	logCount := hostconfig.Logcount
	disabledMetrics := hostconfig.DisabledMetrics

	// TODO.. query parameter could logcount_<logServiceName>=10

	return &CollectionContext{RedfishClient: client, CollectLogs: collectLogs, LogCount: logCount, DisabledMetrics: disabledMetrics}, nil
}

func newRedfishClient(host string, username string, password string) (*gofish.APIClient, error) {

	url := fmt.Sprintf("https://%s", host)

	config := gofish.ClientConfig{
		Endpoint: url,
		Username: username,
		Password: password,
		Insecure: true,
	}
	redfishClient, err := gofish.Connect(config)
	if err != nil {
		return nil, err
	}
	return redfishClient, nil
}
