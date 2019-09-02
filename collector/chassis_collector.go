package collector

import (
	//redfish "github.com/stmcginnis/gofish/school/redfish"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	gofish "github.com/stmcginnis/gofish/school"
	"fmt"
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
	ChassisLabelNames = []string{"resource", "chassis_id"}

	ChassisTemperatureLabelNames    = []string{ "resource", "chassis_id", "sensor", "sensor_id"}
	ChassisFanLabelNames            = []string{ "resource", "chassis_id", "fan", "fan_id"}
	ChassisPowerVotageLabelNames    = []string{ "resource", "chassis_id", "power_votage", "power_votage_id"}
	ChassisPowerSupplyLabelNames    = []string{ "resource", "chassis_id", "power_supply", "power_supply_id"}
	ChassisNetworkAdapterLabelNames = []string{ "resource", "chassis_id", "network_adapter", "network_adapter_id"}
	ChassisNetworkPortLabelNames    = []string{ "resource", "chassis_id", "network_adapter", "network_adapter_id","network_port", "network_port_id","network_port_type","network_port_speed"}
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
					prometheus.BuildFQName(namespace, subsystem, "fan_rpm_percentage"),
					"fan rpm percentage on this chassis component",
					ChassisFanLabelNames,
					nil,
				),
			},
			"chassis_power_voltage_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "power_voltage_state"),
					"power voltage state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
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
			"chassis_network_adapter_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "network_adapter_state"),
					"chassis network adapter state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
					ChassisNetworkAdapterLabelNames,
					nil,
				),
			},
			"chassis_network_adapter_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "network_adapter_health_state"),
					"chassis network adapter health state,1(OK),2(Warning),3(Critical)",
					ChassisNetworkAdapterLabelNames,
					nil,
				),
			},
			"chassis_network_port_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "network_port_state"),
					"chassis network port state state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
					ChassisNetworkPortLabelNames,
					nil,
				),
			},
			"chassis_network_port_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "network_port_health_state"),
					"chassis network port state state,1(OK),2(Warning),3(Critical)",
					ChassisNetworkPortLabelNames,
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
		log.Infof("Errors Getting Services for chassis metrics : %s", err)
	}

	// get a list of chassis from service
	if chassises, err := service.Chassis(); err != nil {
		log.Infof("Errors Getting chassis from service : %s", err)
	} else {
		// process the chassises
		for _, chassis := range chassises {
			chassisID := chassis.ID
			chassisStatus := chassis.Status
			chassisStatusState := chassisStatus.State
			chassisStatusHealth := chassisStatus.Health
			ChassisLabelValues := []string{ "chassis", chassisID}
			if chassisStatusHealthValue, ok := parseCommonStatusHealth(chassisStatusHealth); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
			}
			if chassisStatusStateValue, ok := parseCommonStatusState(chassisStatusState); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
			}
			
			if chassisThermal, err := chassis.Thermal(); err != nil {
				log.Infof("Errors Getting Thermal from chassis : %s", err)
			} else {
				// process temperature
				chassisTemperatures := chassisThermal.Temperatures
				for _, chassisTemperature := range chassisTemperatures {
					chassisTemperatureSensorName := chassisTemperature.Name
					//chassisTemperatureSensorName := chassisTemperature.MemberID
					chassisTemperatureSensorID := chassisTemperature.ID
					chassisTemperatureStatus := chassisTemperature.Status
					//			chassisTemperatureStatusHealth :=chassisTemperatureStatus.Health
					chassisTemperatureStatusState := chassisTemperatureStatus.State
					//			chassisTemperatureStatusLabelNames :=[]string{BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
					chassisTemperatureLabelvalues := []string{ "temperature", chassisID, chassisTemperatureSensorName, chassisTemperatureSensorID}

					//		ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_status_health"].desc, prometheus.GaugeValue, parseCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureLabelvalues...)
					if chassisTemperatureStatusStateValue, ok := parseCommonStatusState(chassisTemperatureStatusState); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_sensor_state"].desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
					}

					chassisTemperatureReadingCelsius := chassisTemperature.ReadingCelsius
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_temperature_celsius"].desc, prometheus.GaugeValue, float64(chassisTemperatureReadingCelsius), chassisTemperatureLabelvalues...)
				}

				// process fans

				chassisFans := chassisThermal.Fans
				for _, chassisFan := range chassisFans {
					chassisFanID := chassisFan.ID
					chassisFanName := chassisFan.Name
					//chassisFanName := chassisFan.MemberID
					chassisFanStaus := chassisFan.Status
					chassisFanStausHealth := chassisFanStaus.Health
					chassisFanStausState := chassisFanStaus.State
					chassisFanRPM := chassisFan.Reading

					//			chassisFanStatusLabelNames :=[]string{BaseLabelNames,"fan_name","fan_member_id")
					chassisFanLabelvalues := []string{ "fan", chassisID, chassisFanName, chassisFanID}

					if chassisFanStausHealthValue, ok := parseCommonStatusHealth(chassisFanStausHealth); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_health"].desc, prometheus.GaugeValue, chassisFanStausHealthValue, chassisFanLabelvalues...)
					}

					if chassisFanStausStateValue, ok := parseCommonStatusState(chassisFanStausState); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_state"].desc, prometheus.GaugeValue, chassisFanStausStateValue, chassisFanLabelvalues...)
					}
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_fan_rpm"].desc, prometheus.GaugeValue, float64(chassisFanRPM), chassisFanLabelvalues...)

				}
			}
			if chassisPowerInfo, err := chassis.Power(); err != nil {
				log.Infof("Errors Getting powerinf from chassis : %s", err)
			} else {
				// power votages
				chassisPowerInfoVoltages := chassisPowerInfo.Voltages
				for _, chassisPowerInfoVoltage := range chassisPowerInfoVoltages {
					chassisPowerInfoVoltageName := chassisPowerInfoVoltage.Name
					//chassisPowerInfoVoltageName := chassisPowerInfoVoltage.MemberID
					chassisPowerInfoVoltageID := chassisPowerInfoVoltage.ID
					chassisPowerInfoVoltageNameReadingVolts := chassisPowerInfoVoltage.ReadingVolts
					chassisPowerInfoVoltageState := chassisPowerInfoVoltage.Status.State
					chassisPowerVotageLabelvalues := []string{ "power_votage", chassisID, chassisPowerInfoVoltageName, chassisPowerInfoVoltageID}
					if chassisPowerInfoVoltageStateValue, ok := parseCommonStatusState(chassisPowerInfoVoltageState); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_voltage_state"].desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, chassisPowerVotageLabelvalues...)
					}
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_voltage_volts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoVoltageNameReadingVolts), chassisPowerVotageLabelvalues...)

				}

				// powerSupply
				chassisPowerInfoPowerSupplies := chassisPowerInfo.PowerSupplies
				for _, chassisPowerInfoPowerSupply := range chassisPowerInfoPowerSupplies {
					chassisPowerInfoPowerSupplyName := chassisPowerInfoPowerSupply.Name
					//chassisPowerInfoPowerSupplyName := chassisPowerInfoPowerSupply.MemberID
					chassisPowerInfoPowerSupplyID := chassisPowerInfoPowerSupply.ID
					chassisPowerInfoPowerSupplyPowerCapacityWatts := chassisPowerInfoPowerSupply.PowerCapacityWatts
					chassisPowerInfoPowerSupplyLastPowerOutputWatts := chassisPowerInfoPowerSupply.LastPowerOutputWatts
					chassisPowerInfoPowerSupplyState := chassisPowerInfoPowerSupply.Status.State
					chassisPowerInfoPowerSupplyHealth := chassisPowerInfoPowerSupply.Status.Health
					chassisPowerSupplyLabelvalues := []string{ "power_supply", chassisID, chassisPowerInfoPowerSupplyName, chassisPowerInfoPowerSupplyID}
					if chassisPowerInfoPowerSupplyStateValue, ok := parseCommonStatusState(chassisPowerInfoPowerSupplyState); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_state"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, chassisPowerSupplyLabelvalues...)
					}
					if chassisPowerInfoPowerSupplyHealthValue, ok := parseCommonStatusHealth(chassisPowerInfoPowerSupplyHealth); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_health"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, chassisPowerSupplyLabelvalues...)
					}
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_last_power_output_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyLastPowerOutputWatts), chassisPowerSupplyLabelvalues...)
					ch <- prometheus.MustNewConstMetric(c.metrics["chassis_power_powersupply_power_capacity_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyPowerCapacityWatts), chassisPowerSupplyLabelvalues...)
				}
			}

			// process NetapAdapter

			if networkAdapters, err := chassis.NetworkAdapters(); err != nil {
				log.Infof("Errors Getting NetworkAdapters from chassis : %s", err)
			} else {

				for _, networkAdapter := range networkAdapters {

					networkAdapterName := networkAdapter.Name
					networkAdapterID := networkAdapter.ID
					networkAdapterState := networkAdapter.Status.State
					networkAdapterHealthState := networkAdapter.Status.Health
					chassisNetworkAdapterLabelValues := []string{ "network_adapter",chassisID,  networkAdapterName, networkAdapterID}
					if networkAdapterStateValue, ok := parseCommonStatusState(networkAdapterState); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_network_adapter_state"].desc, prometheus.GaugeValue, networkAdapterStateValue, chassisNetworkAdapterLabelValues...)
					}
					if networkAdapterHealthStateValue, ok := parseCommonStatusHealth(networkAdapterHealthState); ok {
						ch <- prometheus.MustNewConstMetric(c.metrics["chassis_network_adapter_health_state"].desc, prometheus.GaugeValue, networkAdapterHealthStateValue, chassisNetworkAdapterLabelValues...)
					}

					if networkPorts, err := networkAdapter.NetworkPorts(); err != nil {
						log.Infof("Errors Getting Network port from networkAdapter : %s", err)
					} else {
						for _, networkPort := range networkPorts {
							networkPortName := networkPort.Name
							networkPortID := networkPort.ID
							networkPortState := networkPort.Status.State
							networkPortLinkType :=networkPort.ActiveLinkTechnology
							networkPortLinkSpeed := fmt.Sprintf("%d Mbps",networkPort.CurrentLinkSpeedMbps)
							networkPortHealthState := networkPort.Status.Health
							chassisNetworkPortLabelValues := []string{ "network_port",chassisID, networkAdapterName, networkAdapterID,networkPortName, networkPortID,string(networkPortLinkType),networkPortLinkSpeed}
							if networkPortStateValue, ok := parseCommonStatusState(networkPortState); ok {
								ch <- prometheus.MustNewConstMetric(c.metrics["chassis_network_port_state"].desc, prometheus.GaugeValue, networkPortStateValue, chassisNetworkPortLabelValues...)
							}
							if networkPortHealthStateValue, ok := parseCommonStatusHealth(networkPortHealthState); ok {
								ch <- prometheus.MustNewConstMetric(c.metrics["chassis_network_port_health_state"].desc, prometheus.GaugeValue, networkPortHealthStateValue, chassisNetworkPortLabelValues...)
							}
						}

					}

				}
			}

		}
	}
	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}





