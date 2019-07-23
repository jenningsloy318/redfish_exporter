package collector


import (
	//redfish "github.com/stmcginnis/gofish/school/redfish"
	gofish "github.com/stmcginnis/gofish/school"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"


)

// A ChassisCollector implements the prometheus.Collector.
type ChassisCollector struct {
	redfishClient *gofish.ApiClient
	metrics                   map[string]chassisMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

type chassisMetric struct {
	desc      *prometheus.Desc
}

var (
	ChassisLabelNames =append(BaseLabelNames,"name")

	ChassisTemperatureStatusLabelNames =append(BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
	ChassisFanStatusLabelNames =append(BaseLabelNames,"fan_name","fan_member_id")

)
// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(namespace string, redfishClient *gofish.ApiClient) (*ChassisCollector) {
	var (
		subsystem  = "chassis"
	)
	// get service from redfish client
	

	return &ChassisCollector{
		redfishClient: redfishClient,
		metrics: map[string]chassisMetric{
			"chassis_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "health"),
					"health of chassis ",
					BaseLabelNames,
					nil,
				),
			},
			"chassis_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "health"),
					"state of chassis ",
					BaseLabelNames,
					nil,
				),
			},
//			"chassis_temperature_status_health":{
//				desc: prometheus.NewDesc(
//					prometheus.BuildFQName(namespace, subsystem, "chassis_temperature_status_health"),
//					"status health of temperature on this chassis component ",
//					BaseLabelNames,
//					nil,
//				),
//			},
			"chassis_temperature_sensor_state":{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "chassis_temperature_sensor_state"),
					"status state of temperature on this chassis component ",
					ChassisTemperatureStatusLabelNames,
					nil,
				),
			},
			"chassis_temperature_celsius":{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "chassis_temperature_celsius"),
					"celsius of temperature on this chassis component",
					ChassisTemperatureStatusLabelNames,
					nil,
				),
			},			
			"chassis_fan_health":{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "chassis_fan_health"),
					"fan health on this chassis component",
					ChassisFanStatusLabelNames,
					nil,
				),
			},
			"chassis_fan_state":{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "chassis_fan_state"),
					"fan state on this chassis component",
					ChassisFanStatusLabelNames,
					nil,
				),
			},			
			"chassis_fan_rpm":{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "chassis_fan_rpm"),
					"fan rpm on this chassis component",
					ChassisFanStatusLabelNames,
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



func (c *ChassisCollector) Describe (ch chan<- *prometheus.Desc)  {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
	c.collectorScrapeStatus.Describe(ch)
	c.collectorScrapeDuration.Describe(ch)


}



func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {

	service, err := gofish.ServiceRoot(c.redfishClient)
	if err != nil {
		log.Fatalf("Errors Getting Services for chassis metrics : %s",err)
	}

	// get a list of chassis from service 
  chassises, err :=  service.Chassis()
	if err !=nil {
		log.Fatalf("Errors Getting chassis from service : %s",err)
	}
	// process the chassises 
	for _, chasssis := range chassises {
			chassisStatus :=chasssis.Status
			chassisStatusState :=chassisStatus.State
			chassisStatusHealth :=chassisStatus.Health
			ChassisLabelValues :=append(BaseLabelValues,"chassis")
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].desc, prometheus.GaugeValue, pareCommonStatusHealth(chassisStatusHealth), ChassisLabelValues...)      
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].desc, prometheus.GaugeValue, pareCommonStatusHealth(chassisStatusState), ChassisLabelValues...)      

		chassisThermal, err := chasssis.Thermal()
		if err !=nil {
			log.Fatalf("Errors Getting   Thermal from chassis : %s",err)
		}
		// process temperature
		chassisTemperatures := chassisThermal.Temperatures
		for _,chassisTemperature := range chassisTemperatures {
			chassisTemperatureSensorName := chassisTemperature.Name   
			chassisTemperatureSensorMemberID := chassisTemperature.MemberID   
			chassisTemperatureStatus := chassisTemperature.Status
//			chassisTemperatureStatusHealth :=chassisTemperatureStatus.Health
			chassisTemperatureStatusState :=chassisTemperatureStatus.State
//			chassisTemperatureStatusLabelNames :=append(BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
			chassisTemperatureStatusLabelvalues :=append(BaseLabelValues,chassisTemperatureSensorName,chassisTemperatureSensorMemberID)

	//		ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_status_health"].desc, prometheus.GaugeValue, pareCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureStatusLabelvalues...)      

			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_sensor_state"].desc, prometheus.GaugeValue, pareCommonStatusState(chassisTemperatureStatusState), chassisTemperatureStatusLabelvalues...)      

			chassisTemperatureReadingCelsius := chassisTemperature.ReadingCelsius 
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_celsius"].desc, prometheus.GaugeValue, float64(chassisTemperatureReadingCelsius), chassisTemperatureStatusLabelvalues...)      
		}

		// process fans

		chassisFans := chassisThermal.Fans 
		for _,chassisFan := range chassisFans {
			chassisFanMemberID := chassisFan.MemberID
			chassisFanName := chassisFan.FanName
			chassisFanStaus := chassisFan.Status
			chassisFanStausHealth :=chassisFanStaus.Health
			chassisFanStausState :=chassisFanStaus.State
			chassisFanRPM :=chassisFan.ReadingRPM

//			chassisFanStatusLabelNames :=append(BaseLabelNames,"fan_name","fan_member_id")
			chassisFanStatusLabelvalues :=append(BaseLabelValues,chassisFanName,chassisFanMemberID)


			chassisFanReadingRPM := chassisFan.ReadingRPM    
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_health"].desc, prometheus.GaugeValue, pareCommonStatusHealth(chassisFanStausHealth), chassisFanStatusLabelvalues...)      
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_state"].desc, prometheus.GaugeValue, pareCommonStatusState(chassisFanStausState), chassisFanStatusLabelvalues...)      

			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_rpm"].desc, prometheus.GaugeValue, float64(chassisFanRPM), chassisFanStatusLabelvalues...)      

		}
		chassisPower, err :=chasssis.Power()
		if err != nil {
			log.Fatalf("Errors Getting powerinf from chassis : %s",err)
		}
	}





	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}