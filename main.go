package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promslog "github.com/prometheus/common/promslog"
	promslogflag "github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/stackhpc/redfish_exporter/collector"
	"github.com/stackhpc/redfish_exporter/common"
)

var (
	Version       string
	BuildRevision string
	BuildBranch   string
	BuildTime     string
	BuildHost     string
	rootLoggerCtx *slog.Logger

	configFile = kingpin.Flag(
		"config.file",
		"Path to configuration file.",
	).String()
	webConfig = webflag.AddFlags(kingpin.CommandLine, ":9610")

	sc = &common.SafeConfig{
		C: &common.Config{},
	}
	reloadCh chan chan error
)

func init() {
	promslogConfig := promslog.Config{}
	promslogflag.AddFlags(kingpin.CommandLine, &promslogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	rootLoggerCtx = promslog.New(&promslogConfig).With("app", "redfish_exporter")
	slog.SetDefault(rootLoggerCtx)

	hostname, _ := os.Hostname()
	slog.Info(fmt.Sprintf("version %s, build revision %s, build branch %s, built at %s on host %s", Version, BuildRevision, BuildBranch, BuildTime, hostname))
}

func reloadHandler(configLoggerCtx *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			configLoggerCtx.Info("Triggered configuration reload from /-/reload HTTP endpoint")
			err := sc.ReloadConfig(*configFile)
			if err != nil {
				configLoggerCtx.Error("failed to reload config file", "err", err)
				http.Error(w, "failed to reload config file", http.StatusInternalServerError)
			}
			configLoggerCtx.Info("config file reloaded", "operation", "sc.ReloadConfig")

			w.WriteHeader(http.StatusOK)
			_, err = io.WriteString(w, "Configuration reloaded successfully!")
			if err != nil {
				configLoggerCtx.Warn("failed to send configuration reload status message")
			}
		} else {
			http.Error(w, "Only PUT and POST methods are allowed", http.StatusBadRequest)
		}
	}
}

func SetLogLevel() {
	var level slog.Level
	switch sc.AppLogLevel() {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	slog.SetLogLoggerLevel(level)
}

type ErrorLogger struct {
	logger *slog.Logger
}

func (l *ErrorLogger) Println(v ...any) {
	l.logger.Error("Error during gathering", v...)
}

// define new http handleer
func metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewPedanticRegistry()
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified", 400)
			return
		}
		targetLoggerCtx := rootLoggerCtx.With("target", target)
		targetLoggerCtx.Info("scraping target host")

		var (
			hostConfig *common.HostConfig
			err        error
			ok         bool
			group      []string
		)

		group, ok = r.URL.Query()["group"]

		if ok && len(group[0]) >= 1 {
			// Trying to get hostConfig from group.
			if hostConfig, err = sc.HostConfigForGroup(group[0]); err != nil {
				targetLoggerCtx.Error("error getting credentials", "err", err)
				return
			}
		}

		// Always falling back to single host config when group config failed.
		if hostConfig == nil {
			if hostConfig, err = sc.HostConfigForTarget(target); err != nil {
				targetLoggerCtx.Error("error getting credentials", "err", err)
				return
			}
		}

		// Support optionally overriding collectlogs setting using a query parameter
		collectLogs := hostConfig.Collectlogs
		collectLogsOverride := r.URL.Query().Get("collectlogs")
		if collectLogsOverride != "" {
			if collectLogsQuery, err := strconv.ParseBool(collectLogsOverride); err != nil {
				targetLoggerCtx.Error("error parsing collectlogs query parameter as a boolean", "err", err)
			} else {
				collectLogs = collectLogsQuery
			}
		}

		collectionCtx, err := common.NewCollectionContext(target, hostConfig, targetLoggerCtx, collectLogs)
		if err != nil {
			targetLoggerCtx.Error("error creating collection context", "err", err)
			return
		}
		collector := collector.NewRedfishCollector(collectionCtx, targetLoggerCtx)
		registry.MustRegister(collector)
		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}

		errorLogger := ErrorLogger{
			logger: targetLoggerCtx,
		}

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{ErrorLog: &errorLogger})
		h.ServeHTTP(w, r)

	}
}

func main() {
	configLoggerCtx := rootLoggerCtx.With("config", *configFile)
	configLoggerCtx.Info("starting app")
	// load config  first time
	if err := sc.ReloadConfig(*configFile); err != nil {
		configLoggerCtx.Error("error parsing config file", "err", err)
		panic(err)
	}

	configLoggerCtx.Info("config file loaded", "operation", "sc.ReloadConfig")

	SetLogLevel()

	// load config in background to watch for config changes
	hup := make(chan os.Signal, 1)
	reloadCh = make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile); err != nil {
					configLoggerCtx.Error("failed to reload config file", "err", err)
					break
				}
				configLoggerCtx.Info("config file reload", "operation", "sc.ReloadConfig")
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile); err != nil {
					configLoggerCtx.Error("failed to reload config file", "err", err)
					rc <- err
					break
				}
				configLoggerCtx.Info("config file reloaded", "operation", "sc.ReloadConfig")
				rc <- nil
			}
		}
	}()

	http.Handle("/redfish", metricsHandler())                // Regular metrics endpoint for local Redfish metrics.
	http.Handle("/-/reload", reloadHandler(configLoggerCtx)) // HTTP endpoint for triggering configuration reload
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// nolint
		w.Write([]byte(`<html>
            <head>
            <title>Redfish Exporter</title>
            </head>
						<body>
            <h1>redfish Exporter</h1>
            <form action="/redfish">
            <label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value="1.2.3.4"><br>
            <label>CollectLogs:</label> <input type="text" name="collectlogs" placeholder="collectlogs (optional)" value="bool"><br>
            <label>Group:</label> <input type="text" name="group" placeholder="group (optional)" value=""><br>
            <input type="submit" value="Submit">
						</form>
						<p><a href="/metrics">Local metrics</a></p>
            </body>
            </html>`))
	})

	slog.Info(fmt.Sprintf("app started. listening on %s", (*webConfig.WebListenAddresses)[0]))
	srv := &http.Server{Addr: (*webConfig.WebListenAddresses)[0]}
	err := web.ListenAndServe(srv, webConfig, rootLoggerCtx)
	if err != nil {
		slog.Error(err.Error())
	}
}
