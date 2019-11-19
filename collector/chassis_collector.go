package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"sync"
)

// A ChassisCollector implements the prometheus.Collector.

var (
	ChassisSubsystem  = "chassis"
	ChassisLabelNames = []string{"resource", "chassis_id"}
)

type ChassisCollector struct {
	chassis               *gofish.Chassis
	metrics               map[string]chassisMetric
	collectors            map[string]prometheus.Collector
	collectorScrapeStatus *prometheus.GaugeVec
}

type chassisMetric struct {
	desc *prometheus.Desc
}

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(namespace string, chassis *gofish.Chassis) *ChassisCollector {
	var collectors map[string]prometheus.Collector
	// get all collectors
	chassisID := chassis.ID
	thermal, err := chassis.Thermal()
	if err != nil {
		log.Infof("Errors Getting thermal from chassis : %s", err)
	} else {
		chassisThermalCollector := NewChassisThermalCollector(namespace, chassisID, thermal)
		chassisThermalcollectorName := fmt.Sprintf("chassis_%s_thermal", chassisID)
		collectors[ChassisThermalcollectorName] = chassisThermalCollector
	}

	power, err := chassis.Power()
	if err != nil {
		log.Infof("Errors Getting power infomation from chassis : %s", err)
	} else {
		chassisPowerCollector := NewChassisPowerCollector(namespace, chassisID, thermal)
		chassisPowercollectorName := fmt.Sprintf("chassis_%s_power", chassisID)
		collectors[ChassisPowercollectorName] = chassisPowerCollector
	}

	networkAdapters, err := chassis.NetworkAdapters()
	if err != nil {
		log.Infof("Errors Getting power infomation from chassis : %s", err)
	} else {
		chassisNetworkAdaptersCollector := NewChassisPowerCollector(namespace, chassisID, thermal)
		chassisNetworkAdapterscollectorName := fmt.Sprintf("chassis_%s_networkAdapters", chassisID)
		collectors[chassisNetworkAdapterscollectorName] = chassisNetworkAdaptersCollector
	}

	return &ChassisCollector{
		chassises:  chassises,
		metrics:    chassisMetrics,
		collecotrs: collectors,
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

func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}

	for _, collector := range c.collectors {
		collector.Describe(ch)
	}
	
	c.collectorScrapeStatus.Describe(ch)

}

func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {
	// process the chassises
	chassisID := chassis.ID
	chassisStatus := chassis.Status
	chassisStatusState := chassisStatus.State
	chassisStatusHealth := chassisStatus.Health
	ChassisLabelValues := []string{"chassis", chassisID}
	if chassisStatusHealthValue, ok := parseCommonStatusHealth(chassisStatusHealth); ok {
		ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
	}
	if chassisStatusStateValue, ok := parseCommonStatusState(chassisStatusState); ok {
		ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(c.collectors))

	defer wg.Wait()
	for _, collector := range c.collectors {
		go func(collector prometheus.Collector) {
			defer wg.Done()
			collector.Collect(ch)
		}(collector)
	}

	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}
