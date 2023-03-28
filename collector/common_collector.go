package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
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
