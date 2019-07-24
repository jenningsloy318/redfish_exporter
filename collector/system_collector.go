package collector


import (
	//redfish "github.com/stmcginnis/gofish/school/redfish"
	gofish "github.com/stmcginnis/gofish/school"
	redfish "github.com/stmcginnis/gofish/school/redfish"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

)

// A SystemCollector implements the prometheus.Collector.
type SystemCollector struct {
	redfishClient *gofish.ApiClient
	metrics                   map[string]systemMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

type systemMetric struct {
	desc      *prometheus.Desc
}

var (
	SystemLabelNames =append(BaseLabelNames,"name","hostname")


)
// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(namespace string, redfishClient *gofish.ApiClient) (*SystemCollector) {
	var (
		subsystem  = "system"
	)
	return &SystemCollector{
		redfishClient: redfishClient,
		metrics: map[string]systemMetric{
			"system_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "state"),
					"system state",
					SystemLabelNames,
					nil,
				),
			},
			"system_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "health"),
					"system health",
					SystemLabelNames,
					nil,
				),
			},			
			"system_power_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_state"),
					"system power state",
					SystemLabelNames,
					nil,
				),
			},		
			"system_total_memory_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_memory_state"),
					"system overall memory state",
					SystemLabelNames,
					nil,
				),
			},				
			"system_total_memory_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_memory_health"),
					"system overall memory health",
					SystemLabelNames,
					nil,
				),
			},	
			"system_total_memory_size": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_memory_size"),
					"system total memory size, GiB",
					SystemLabelNames,
					nil,
				),
			},				
			"system_total_processor_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_processor_state"),
					"system overall processor state",
					SystemLabelNames,
					nil,
				),
			},						
			"system_total_processor_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_processor_health"),
					"system overall processor health",
					SystemLabelNames,
					nil,
				),
			},
			"system_total_processor_count": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_processor_count"),
					"system total  processor count",
					SystemLabelNames,
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



func (s *SystemCollector) Describe (ch chan<- *prometheus.Desc)  {
	for _, metric := range s.metrics {
		ch <- metric.desc
	}
	s.collectorScrapeStatus.Describe(ch)
	s.collectorScrapeDuration.Describe(ch)


}



func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	//get service
	service, err := gofish.ServiceRoot(s.redfishClient)
	if err != nil {
		log.Fatalf("Errors Getting Services for chassis metrics : %s",err)
	}

	// get a list of systems from service 
  systems, err :=  service.Systems()
	if err !=nil {
		log.Fatalf("Errors Getting systems from service : %s",err)
	}

	for _,system := range systems {
		// overall system metrics 
		systemName := system.Name
		systemHostName :=system.HostName
		systemPowerState :=system.PowerState
		systemState := system.Status.State
		systemHealth := system.Status.Health
		systemTotalProcessorCount := system.ProcessorSummary.Count
		systemTotalProcessorsState := system.ProcessorSummary.Status.State
		systemTotalProcessorsHealth := system.ProcessorSummary.Status.Health
		systemTotalMemoryState := system.MemorySummary.Status.State
		systemTotalMemoryHealth := system.MemorySummary.Status.Health
		systemTotalMemoryAmount := system.MemorySummary.TotalSystemMemoryGiB
		

		SystemLabelValues :=append(BaseLabelValues,systemName,systemHostName)
		
		ch <- prometheus.MustNewConstMetric(s.metrics["system_health"].desc, prometheus.GaugeValue, parseCommonStatusHealth(systemHealth), SystemLabelValues...)      
		ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue, parseCommonStatusState(systemState), SystemLabelValues...)      
		ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].desc, prometheus.GaugeValue, parseSystemPowerState(systemPowerState), SystemLabelValues...)  
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].desc, prometheus.GaugeValue, parseCommonStatusState(systemTotalProcessorsState), SystemLabelValues...)  
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health"].desc, prometheus.GaugeValue, parseCommonStatusHealth(systemTotalProcessorsHealth), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].desc, prometheus.GaugeValue,  float64(systemTotalProcessorCount), SystemLabelValues...)      
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].desc, prometheus.GaugeValue,  parseCommonStatusState(systemTotalMemoryState), SystemLabelValues...)      
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health"].desc, prometheus.GaugeValue,  parseCommonStatusHealth(systemTotalMemoryHealth), SystemLabelValues...)      
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].desc, prometheus.GaugeValue,  float64(systemTotalMemoryAmount), SystemLabelValues...)      
	
		// process memory
		memoryLink := string([]byte(system.MemoryLink))

		memories,err :=  redfish.GetMemory(s.redfishClient,memoryLink)
		if err !=nil {
			log.Fatalf("Errors Getting memory uri from computer system : %s",err)
		}
		log.Infof("memories %s",memories)
	}














}