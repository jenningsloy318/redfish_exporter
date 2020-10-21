package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	alog "github.com/apex/log"
	"github.com/jenningsloy318/redfish_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	Version       string
	BuildRevision string
	BuildBranch   string
	BuildTime     string
	BuildHost     string
	rootLoggerCtx *alog.Entry

	configFile = kingpin.Flag(
		"config.file",
		"Path to configuration file.",
	).String()
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default("127.0.0.1:9610").String()
	sc = &SafeConfig{
		Config: &Config{},
	}
	reloadCh chan chan error
)

func init() {
	rootLoggerCtx = alog.WithFields(alog.Fields{
		"app": "redfish_exporter",
	})

	hostname, _ := os.Hostname()
	rootLoggerCtx.Infof("version %s, build reversion %s, build branch %s, build at %s on host %s", Version, BuildRevision, BuildBranch, BuildTime, hostname)
}

// define new http handleer
func metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified", 400)
			return
		}
		targetLoggerCtx := rootLoggerCtx.WithField("target", target)
		targetLoggerCtx.Info("scraping target host")

		var hostConfig *HostConfig
		var err error

		if hostConfig, err = sc.HostConfigForTarget(target); err != nil {
			targetLoggerCtx.WithError(err).Error("error getting credentials")
			return
		}

		collector := collector.NewRedfishCollector(target, hostConfig.Username, hostConfig.Password, targetLoggerCtx)
		registry.MustRegister(collector)
		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)

	}
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	configLoggerCtx := rootLoggerCtx.WithField("config", *configFile)
	configLoggerCtx.Info("starting app")
	// load config  first time
	if err := sc.ReloadConfig(*configFile); err != nil {
		configLoggerCtx.WithError(err).Error("error parsing config file")
		panic(err)
	}

	configLoggerCtx.WithField("operation", "sc.ReloadConfig").Info("config file loaded")

	// load config in background to watch for config changes
	hup := make(chan os.Signal)
	reloadCh = make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile); err != nil {
					configLoggerCtx.WithError(err).Error("failed to reload config file")
					break
				}
				configLoggerCtx.WithField("operation", "sc.ReloadConfig").Info("config file reload")
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile); err != nil {
					configLoggerCtx.WithError(err).Error("failed to reload config file")
					rc <- err
					break
				}
				configLoggerCtx.WithField("operation", "sc.ReloadConfig").Info("config file reloaded")
				rc <- nil
			}
		}
	}()

	http.Handle("/redfish", metricsHandler()) // Regular metrics endpoint for local Redfish metrics.
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>Redfish Exporter</title>
            </head>
						<body>
            <h1>redfish Exporter</h1>
            <form action="/redfish">
            <label>Target:</label> <input type="text" name="target" placeholder="X.X.X.X" value="1.2.3.4"><br>
            <input type="submit" value="Submit">
						</form>
						<p><a href="/metrics">Local metrics</a></p>
            </body>
            </html>`))
	})

	rootLoggerCtx.Infof("app started. listening on %s", *listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
