package collector


import (
	//redfish "github.com/stmcginnis/gofish/school/redfish"
	gofish "github.com/stmcginnis/gofish/school"
	"github.com/prometheus/client_golang/prometheus"

)

// A MemoryCollector implements the prometheus.Collector.
type MemoryCollector struct {
	redfishClient *gofish.ApiClient
	metrics                   map[string]memoryMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

type memoryMetric struct {
	desc      *prometheus.Desc
}


// NewMemoryCollector returns a collector that collecting memory statistics
func NewMemoryCollector(namespace string, redfishClient *gofish.ApiClient) (*MemoryCollector) {
	var (
		subsystem  = "memory"
//		labelNames = []string{"partition", "node"}
	)
	return &MemoryCollector{
		redfishClient: redfishClient,
		metrics: map[string]memoryMetric{
			"serverside_bytesOut": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "serverside_bytes_out"),
					"serverside_bytes_out",
					BaseLabelNames,
					nil,
				),
			},
		},
		collectorScrapeStatus: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      "collector_scrape_status",
					Help:      "collector_scrape_status",
				},
				[]string{"collector"},
			),
		collectorScrapeDuration: prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Namespace: namespace,
					Name:      "collector_scrape_duration",
					Help:      "collector_scrape_duration",
				},
				[]string{"collector"},
			),
		}
	}



func (m *MemoryCollector) Describe (ch chan<- *prometheus.Desc)  {
	for _, metric := range m.metrics {
		ch <- metric.desc
	}
	m.collectorScrapeStatus.Describe(ch)
	m.collectorScrapeDuration.Describe(ch)


}



func (m *MemoryCollector) Collect(ch chan<- prometheus.Metric) {
	m.collectorScrapeStatus.WithLabelValues("memory").Set(float64(1))
}