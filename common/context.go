package common

import (
	"fmt"
	"net/http"
	"strconv"

	alog "github.com/apex/log"
	gofish "github.com/stmcginnis/gofish"
)

type CollectionContext struct {
	Request       *http.Request
	RedfishClient *gofish.APIClient
	CollectLogs   bool
	LogCount      int
}

func NewCollectionContext(r *http.Request, target string, hostconfig *HostConfig, logger *alog.Entry) (*CollectionContext, error) {
	client, err := newRedfishClient(target, hostconfig.Username, hostconfig.Password)
	if err != nil {
		logger.WithError(err).Error("error creating redfish client")
		return nil, err
	}

	// Support optionally overriding logCounts setting using a query parameter
	logCount := hostconfig.Logcount
	logCountOverride := r.URL.Query().Get("logcount")
	if logCountOverride != "" {
		if logCountQuery, err := strconv.Atoi(logCountOverride); err != nil {
			logger.WithError(err).Error("error parsing collectlogs query parameter as a boolean")
		} else {
			logCount = logCountQuery
		}
	}
	logger.WithField("operation", "NewCollectionContext()").Info(fmt.Sprintf("logcount=%d", logCount))

	// Support optionally overriding collectlogs setting using a query parameter
	collectLogs := hostconfig.Collectlogs
	collectLogsOverride := r.URL.Query().Get("collectlogs")
	if collectLogsOverride != "" {
		if collectLogsQuery, err := strconv.ParseBool(collectLogsOverride); err != nil {
			logger.WithError(err).Error("error parsing collectlogs query parameter as a boolean")
		} else {
			collectLogs = collectLogsQuery
		}
	}
	return &CollectionContext{Request: r, RedfishClient: client, CollectLogs: collectLogs, LogCount: logCount}, nil
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
