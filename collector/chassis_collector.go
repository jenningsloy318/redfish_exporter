package collector

import (
	//redfish "github.com/stmcginnis/gofish/school/redfish"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	gofish "github.com/stmcginnis/gofish/school"
)

// A ChassisCollector implements the prometheus.Collector.
type ChassisCollector struct {
	redfishClient           *gofish.ApiClient
	metrics                 map[string]chassisMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

type chassisMetric struct {
	desc *prometheus.Desc
}

var (
	ChassisLabelNames = append(BaseLabelNames, "name")

	ChassisTemperatureLabelNames = append(BaseLabelNames, "name", "temperature_sensor_name", "temperature_sensor_member_id")
	ChassisFanLabelNames         = append(BaseLabelNames, "name", "fan_name", "fan_member_id")
	ChassisPowerVotageLabelNames = append(BaseLabelNames, "name", "power_name")
	ChassisPowerSupplyLabelNames = append(BaseLabelNames, "name", "power_name")
)

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(namespace string, redfishClient *gofish.ApiClient) *ChassisCollector {
	var (
		subsystem = "chassis"
	)
	// get service from redfish client

	return &ChassisCollector{
		redfishClient: redfishClient,
		metrics: map[string]chassisMetric{
			"chassis_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "health"),
					"health of chassis, 1(OK),2(Warning),3(Critical)",
					ChassisLabelNames,
					nil,
				),
			},
			"chassis_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "state"),
					"state of chassis,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
					ChassisLabelNames,
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
			"chassis_temperature_sensor_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "temperature_sensor_state"),
					"status state of temperature on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
					ChassisTemperatureLabelNames,
					nil,
				),
			},
			"chassis_temperature_celsius": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "temperature_celsius"),
					"celsius of temperature on this chassis component",
					ChassisTemperatureLabelNames,
					nil,
				),
			},
			"chassis_fan_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "fan_health"),
					"fan health on this chassis component,1(OK),2(Warning),3(Critical)",
					ChassisFanLabelNames,
					nil,
				),
			},
			"chassis_fan_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "fan_state"),
					"fan state on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
					ChassisFanLabelNames,
					nil,
				),
			},
			"chassis_fan_rpm": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "fan_rpm"),
					"fan rpm on this chassis component",
					ChassisFanLabelNames,
					nil,
				),
			},
			"chassis_power_voltage_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_voltage_state"),
					"power voltage state of chassis component",
					ChassisPowerVotageLabelNames,
					nil,
				),
			},
			"chassis_power_voltage_volts": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_voltage_volts"),
					"power voltage volts number of chassis component",
					ChassisPowerVotageLabelNames,
					nil,
				),
			},
			"chassis_power_powersupply_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_powersupply_state"),
					"powersupply state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
					ChassisPowerSupplyLabelNames,
					nil,
				),
			},
			"chassis_power_powersupply_health": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_powersupply_health"),
					"powersupply health of chassis component,1(OK),2(Warning),3(Critical)",
					ChassisPowerSupplyLabelNames,
					nil,
				),
			},
			"chassis_power_powersupply_last_power_output_watts": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_powersupply_last_power_output_watts"),
					"last_power_output_watts of powersupply on this chassis",
					ChassisPowerSupplyLabelNames,
					nil,
				),
			},
			"chassis_power_powersupply_power_capacity_watts": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_powersupply_power_capacity_watts"),
					"power_capacity_watts of powersupply on this chassis",
					ChassisPowerSupplyLabelNames,
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

func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
	c.collectorScrapeStatus.Describe(ch)
	c.collectorScrapeDuration.Describe(ch)

}

func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {

	service, err := gofish.ServiceRoot(c.redfishClient)
	if err != nil {
		log.Fatalf("Errors Getting Services for chassis metrics : %s", err)
	}

	// get a list of chassis from service
	if chassises, err := service.Chassis(); err != nil {
		log.Fatalf("Errors Getting chassis from service : %s", err)
	} else {
		// process the chassises
		for _, chasssis := range chassises {
			chassisStatus := chasssis.Status
			chassisStatusState := chassisStatus.State
			chassisStatusHealth := chassisStatus.Health
			ChassisLabelValues := append(BaseLabelValues, "chassis")
			if chassisStatusHealthValue := parseCommonStatusHealth(chassisStatusHealth); chassisStatusHealthValue != float64(0) {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
			}
			if chassisStatusStateValue := parseCommonStatusState(chassisStatusState); chassisStatusStateValue != float64(0) {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
			}
			if chassisThermal, err := chasssis.Thermal(); err != nil {
				log.Infof("Errors Getting Thermal from chassis : %s", err)
			} else {
				// process temperature
				chassisTemperatures := chassisThermal.Temperatures
				for _, chassisTemperature := range chassisTemperatures {
					chassisTemperatureSensorName := chassisTemperature.Name
					chassisTemperatureSensorMemberID := chassisTemperature.MemberID
					chassisTemperatureStatus := chassisTemperature.Status
					//			chassisTemperatureStatusHealth :=chassisTemperatureStatus.Health
					chassisTemperatureStatusState := chassisTemperatureStatus.State
					//			chassisTemperatureStatusLabelNames :=append(BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
					chassisTemperatureLabelvalues := append(BaseLabelValues, "temperature", chassisTemperatureSensorName, chassisTemperatureSensorMemberID)

					//		ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_status_health"].desc, prometheus.GaugeValue, parseCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureLabelvalues...)
					if chassisTemperatureStatusStateValue := parseCommonStatusState(chassisTemperatureStatusState); chassisTemperatureStatusStateValue != float64(0) {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_sensor_state"].desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
					}

					chassisTemperatureReadingCelsius := chassisTemperature.ReadingCelsius
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_celsius"].desc, prometheus.GaugeValue, float64(chassisTemperatureReadingCelsius), chassisTemperatureLabelvalues...)
				}

				// process fans

				chassisFans := chassisThermal.Fans
				for _, chassisFan := range chassisFans {
					chassisFanMemberID := chassisFan.MemberID
					chassisFanName := chassisFan.FanName
					chassisFanStaus := chassisFan.Status
					chassisFanStausHealth := chassisFanStaus.Health
					chassisFanStausState := chassisFanStaus.State
					chassisFanRPM := chassisFan.ReadingRPM

					//			chassisFanStatusLabelNames :=append(BaseLabelNames,"fan_name","fan_member_id")
					chassisFanLabelvalues := append(BaseLabelValues, "fan", chassisFanName, chassisFanMemberID)

					if chassisFanStausHealthValue := parseCommonStatusHealth(chassisFanStausHealth); chassisFanStausHealthValue != float64(0) {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_health"].desc, prometheus.GaugeValue, chassisFanStausHealthValue, chassisFanLabelvalues...)
					}

					if chassisFanStausStateValue := parseCommonStatusState(chassisFanStausState); chassisFanStausStateValue != float64(0) {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_state"].desc, prometheus.GaugeValue, chassisFanStausStateValue, chassisFanLabelvalues...)
					}
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_rpm"].desc, prometheus.GaugeValue, float64(chassisFanRPM), chassisFanLabelvalues...)

				}
			}
			if chassisPowerInfo, err := chasssis.Power(); err != nil {
				log.Infof("Errors Getting powerinf from chassis : %s", err)
			} else {
				// power votages
				chassisPowerInfoVoltages := chassisPowerInfo.Voltages
				for _, chassisPowerInfoVoltage := range chassisPowerInfoVoltages {
					chassisPowerInfoVoltageName := chassisPowerInfoVoltage.Name
					chassisPowerInfoVoltageNameReadingVolts := chassisPowerInfoVoltage.ReadingVolts
					chassisPowerInfoVoltageState := chassisPowerInfoVoltage.Status.State
					chassisPowerVotageLabelvalues := append(BaseLabelValues, "power_votage", chassisPowerInfoVoltageName)
					if chassisPowerInfoVoltageStateValue := parseCommonStatusState(chassisPowerInfoVoltageState); chassisPowerInfoVoltageStateValue != float64(0) {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_voltage_state"].desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, chassisPowerVotageLabelvalues...)
					}
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_voltage_volts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoVoltageNameReadingVolts), chassisPowerVotageLabelvalues...)

				}

				// powerSupply
				chassisPowerInfoPowerSupplies := chassisPowerInfo.PowerSupplies
				for _, chassisPowerInfoPowerSupply := range chassisPowerInfoPowerSupplies {
					chassisPowerInfoPowerSupplyName := chassisPowerInfoPowerSupply.Name
					chassisPowerInfoPowerSupplyPowerCapacityWatts := chassisPowerInfoPowerSupply.PowerCapacityWatts
					chassisPowerInfoPowerSupplyLastPowerOutputWatts := chassisPowerInfoPowerSupply.LastPowerOutputWatts
					chassisPowerInfoPowerSupplyState := chassisPowerInfoPowerSupply.Status.State
					chassisPowerInfoPowerSupplyHealth := chassisPowerInfoPowerSupply.Status.Health
					chassisPowerSupplyLabelvalues := append(BaseLabelValues, "power_supply", chassisPowerInfoPowerSupplyName)
					if chassisPowerInfoPowerSupplyStateValue := parseCommonStatusState(chassisPowerInfoPowerSupplyState); chassisPowerInfoPowerSupplyStateValue != float64(0) {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_state"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, chassisPowerSupplyLabelvalues...)
					}
					if chassisPowerInfoPowerSupplyHealthValue := parseCommonStatusHealth(chassisPowerInfoPowerSupplyHealth); chassisPowerInfoPowerSupplyHealthValue != float64(0) {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_health"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, chassisPowerSupplyLabelvalues...)
					}
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_last_power_output_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyLastPowerOutputWatts), chassisPowerSupplyLabelvalues...)
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_power_capacity_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyPowerCapacityWatts), chassisPowerSupplyLabelvalues...)
				}
			}
		}
	}
	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}
