package collector

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	CommonStateHelp           = "1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)"
	CommonHealthHelp          = "1(OK),2(Warning),3(Critical)"
	CommonSeverityHelp        = CommonHealthHelp
	CommonLinkHelp            = "1(LinkUp),2(NoLink),3(LinkDown)"
	CommonPortLinkHelp        = "1(Up),0(Down)"
	CommonIntrusionSensorHelp = "1(Normal),2(TamperingDetected),3(HardwareIntrusion)"
)

type Metric struct {
	desc *prometheus.Desc
}

func addToMetricMap(metricMap map[string]Metric, subsystem, name, help string, variableLabels []string) {
	metricKey := fmt.Sprintf("%s_%s", subsystem, name)
	metricMap[metricKey] = Metric{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, name),
			help,
			variableLabels,
			nil,
		),
	}
}

func parseLogService(ch chan<- prometheus.Metric, metrics map[string]Metric, subsystem, collectorID string, logService *redfish.LogService, wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	logServiceName := logService.Name
	logServiceID := logService.ID
	logServiceEnabled := fmt.Sprintf("%t", logService.ServiceEnabled)
	logServiceOverWritePolicy := string(logService.OverWritePolicy)
	logServiceState := logService.Status.State
	logServiceHealthState := logService.Status.Health

	logServiceLabelValues := []string{collectorID, logServiceName, logServiceID, logServiceEnabled, logServiceOverWritePolicy}

	if logServiceStateValue, ok := parseCommonStatusState(logServiceState); ok {
		ch <- prometheus.MustNewConstMetric(metrics[fmt.Sprintf("%s_%s", subsystem, "log_service_state")].desc, prometheus.GaugeValue, logServiceStateValue, logServiceLabelValues...)
	}
	if logServiceHealthStateValue, ok := parseCommonStatusHealth(logServiceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(metrics[fmt.Sprintf("%s_%s", subsystem, "log_service_health_state")].desc, prometheus.GaugeValue, logServiceHealthStateValue, logServiceLabelValues...)
	}

	logEntries, err := logService.Entries()
	if err != nil {
		return
	}
	wg2 := &sync.WaitGroup{}
	wg2.Add(len(logEntries))
	for _, logEntry := range logEntries {
		go parseLogEntry(ch, metrics[fmt.Sprintf("%s_%s", subsystem, "log_entry_severity_state")].desc, collectorID, logServiceName, logServiceID, logEntry, wg2)
	}
	return
}

func parseLogEntry(ch chan<- prometheus.Metric, desc *prometheus.Desc, collectorID, logServiceName, logServiceID string, logEntry *redfish.LogEntry, wg *sync.WaitGroup) {
	defer wg.Done()
	logEntryName := logEntry.Name
	logEntryID := logEntry.ID
	logEntryCode := string(logEntry.EntryCode)
	logEntryType := string(logEntry.EntryType)
	logEntryMessageID := logEntry.MessageID
	logEntrySensorNumber := fmt.Sprintf("%d", logEntry.SensorNumber)
	logEntrySensorType := string(logEntry.SensorType)
	logEntrySeverityState := logEntry.Severity

	logEntryLabelValues := []string{collectorID, logServiceName, logServiceID, logEntryName, logEntryID, logEntryCode, logEntryType, logEntryMessageID, logEntrySensorNumber, logEntrySensorType}

	if logEntrySeverityStateValue, ok := parseCommonSeverityState(logEntrySeverityState); ok {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, logEntrySeverityStateValue, logEntryLabelValues...)
	}
}
