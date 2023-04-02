package collector

import (
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish"
)

// ManagerSubmanager is the manager subsystem
var (
	ManagerSubmanager = "manager"
	ManagerLabelNames = []string{"manager_id", "name", "model", "type"}

	ManagerLogServiceLabelNames = []string{"manager_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	ManagerLogEntryLabelNames   = []string{"manager_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	managerMetrics = createManagerMetricMap()
)

// ManagerCollector implements the prometheus.Collector.
type ManagerCollector struct {
	redfishClient         *gofish.APIClient
	metrics               map[string]Metric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *log.Entry
}

func createManagerMetricMap() map[string]Metric {
	managerMetrics := make(map[string]Metric)
	addToMetricMap(managerMetrics, ManagerSubmanager, "state", fmt.Sprintf("manager state,%s", CommonStateHelp), ManagerLabelNames)
	addToMetricMap(managerMetrics, ManagerSubmanager, "health_state", fmt.Sprintf("manager health,%s", CommonHealthHelp), ManagerLabelNames)
	addToMetricMap(managerMetrics, ManagerSubmanager, "power_state", "manager power state", ManagerLabelNames)

	addToMetricMap(managerMetrics, ManagerSubmanager, "log_service_state", fmt.Sprintf("manager log service state,%s", CommonStateHelp), ManagerLogServiceLabelNames)
	addToMetricMap(managerMetrics, ManagerSubmanager, "log_service_health_state", fmt.Sprintf("manager log service health state,%s", CommonHealthHelp), ManagerLogServiceLabelNames)
	addToMetricMap(managerMetrics, ManagerSubmanager, "log_entry_severity_state", fmt.Sprintf("manager log entry severity state,%s", CommonSeverityHelp), ManagerLogEntryLabelNames)

	return managerMetrics
}

// NewManagerCollector returns a collector that collecting memory statistics
func NewManagerCollector(redfishClient *gofish.APIClient, logger *log.Entry) *ManagerCollector {
	return &ManagerCollector{
		redfishClient: redfishClient,
		metrics:       managerMetrics,
		Log: logger.WithFields(log.Fields{
			"collector": "ManagerCollector",
		}),
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
	}
}

// Describe implemented prometheus.Collector
func (m *ManagerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.desc
	}
	m.collectorScrapeStatus.Describe(ch)

}

// Collect implemented prometheus.Collector
func (m *ManagerCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := m.Log
	//get service
	service := m.redfishClient.Service

	// get a list of managers from service
	if managers, err := service.Managers(); err != nil {
		collectorLogContext.WithField("operation", "service.Managers()").WithError(err).Error("error getting managers from service")
	} else {
		for _, manager := range managers {
			managerLogContext := collectorLogContext.WithField("Manager", manager.ID)
			managerLogContext.Info("collector scrape started")
			// overall manager metrics
			ManagerID := manager.ID
			managerName := manager.Name
			managerModel := manager.Model
			managerType := fmt.Sprint(manager.ManagerType)
			managerPowerState := manager.PowerState
			managerState := manager.Status.State
			managerHealthState := manager.Status.Health

			ManagerLabelValues := []string{ManagerID, managerName, managerModel, managerType}

			if managerHealthStateValue, ok := parseCommonStatusHealth(managerHealthState); ok {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_health_state"].desc, prometheus.GaugeValue, managerHealthStateValue, ManagerLabelValues...)
			}
			if managerStateValue, ok := parseCommonStatusState(managerState); ok {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_state"].desc, prometheus.GaugeValue, managerStateValue, ManagerLabelValues...)
			}
			if managerPowerStateValue, ok := parseCommonPowerState(managerPowerState); ok {
				ch <- prometheus.MustNewConstMetric(m.metrics["manager_power_state"].desc, prometheus.GaugeValue, managerPowerStateValue, ManagerLabelValues...)
			}

			// process log services
			logServices, err := manager.LogServices()
			if err != nil {
				managerLogContext.WithField("operation", "manager.LogServices()").WithError(err).Error("error getting log services from manager")
			} else if logServices == nil {
				managerLogContext.WithField("operation", "manager.LogServices()").Info("no log services found")
			} else {
				wg := &sync.WaitGroup{}
				wg.Add(len(logServices))

				for _, logService := range logServices {
					if err = parseLogService(ch, managerMetrics, ManagerSubmanager, ManagerID, logService, wg); err != nil {
						managerLogContext.WithField("operation", "manager.LogServices()").WithError(err).Error("error getting log entries from log service")
					}
				}
			}

			managerLogContext.Info("collector scrape completed")
		}
		m.collectorScrapeStatus.WithLabelValues("manager").Set(float64(1))

	}

}
