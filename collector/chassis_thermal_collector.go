package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/stmcginnis/gofish/redfish"
	"sync"
)

// A ChassisThermalCollector implements the prometheus.Collector.

var (
	ChassisSubsystem         = "chassis"
	ChassisThermalLabelNames = []string{"resource", "chassis_id"}
	Metrics := map[string]chassisMetric{
		"thermal_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "thermal_state"),
				"State of Chassis Thermal",
				ChassisThermalLabelNames,
				nil,
			),
		},
		"thermal_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "thermal_health"),
				"Health of Chassis Thermal",
				ChassisThermalLabelNames,
				nil,
			),
		}
	}
)

type ChassisThermalCollector struct {
	chassisID             string
	thermal               *redfish.Thermal
	metrics               map[string]ChassisThermalMetric
	collectors            map[string]prometheus.Collector
	collectorScrapeStatus *prometheus.GaugeVec
}

type ChassisThermalMetric struct {
	desc *prometheus.Desc
}

// NewChassisThermalCollector returns a collector that collecting chassis statistics
func NewChassisThermalCollector(namespace string, chassisID string, thermal *redfish.Thermal) *ChassisThermalCollector {
	fans := thermal.Fans
	temperatures := thermal.Temperatures
	// get service from redfish client
	collectors := map[string]prometheus.Collector{"fan": NewFanCollector(namespace, chassisID, fans), "temperature": NewTemperatureCollecor(namespace, chassisID, temperatures)}
	return &ChassisThermalCollector{
		chassisID:  chassisID,
		thermal:    thermal,
		metrics: Metrics,
		collectors: collectors,
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

func (c *ChassisThermalCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}

	for _, collector := range c.collectors {
		collector.Describe(ch)
	}

	c.collectorScrapeStatus.Describe(ch)

}

func (c *ChassisThermalCollector) Collect(ch chan<- prometheus.Metric) {
	// process thermal itslef metrics
	labelValues := []string{"thermal", c.chassisID}
	thermalState := c.thermal.Status.State 
	thermalHealth := c.thermal.Status.Health 
	
	if thermalStateValue, ok := parseCommonStatusState(thermalState); ok {
		ch <- prometheus.MustNewConstMetric(c.metrics["thermal_state"], prometheus.GaugeValue,thermalStateValue , labelValues...)
	}
	
	if thermalHealthValue, ok := parseCommonStatusState(thermalHealth); ok {
		ch <- prometheus.MustNewConstMetric(c.metrics["thermal_health"], prometheus.GaugeValue,thermalHealthValue , labelValues...)
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

	c.collectorScrapeStatus.WithLabelValues("chassis_thermal").Set(float64(1))
}
