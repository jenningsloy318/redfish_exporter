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
	labelValues := []string{"thermal", c.chassisID, "interface", "ifnet"}
	thermalStatus := c.thermal.Status
	valueOfthermalStatus := reflect.ValueOf(&thermalStatus).Elem()
	typeOfthermalStatus := valueOfthermalStatus.Type()

	for index := 0; index < valueOfthermalStatus.NumField(); index++ {
		var floatType = reflect.TypeOf(float64(0))

		metricName := fmt.Sprintf("%s", strings.ToLower(typeOfthermalStatus.Field(index).Name))
		metricValue := valueOfthermalStatus.Field(index)
		if !metricValue.Type().ConvertibleTo(floatType) {
			fmt.Errorf("cannot convert %v to float64", metricValue.Type())
			continue
		}
		metricDesc := fmt.Sprintf("%s of Thermal", strings.Title(metricName))
		newChassisThermalMetric := ChassisThermalMetric{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, metricName),
				metricDesc,
				ChassisThermalLabelNames,
				nil,
			),
		}
		c.metrics[metricName] = newChassisThermalMetric
		ch <- prometheus.MustNewConstMetric(newChassisThermalMetric.desc, prometheus.GaugeValue, metricValue.Convert(floatType).Float(), labelValues...)
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
