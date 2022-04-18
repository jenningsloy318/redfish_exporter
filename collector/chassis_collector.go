package collector

import (
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// ChassisSubsystem is the chassis subsystem
var (
	ChassisSubsystem                  = "chassis"
	ChassisLabelNames                 = []string{"resource", "chassis_id"}
	ChassisModel                      = []string{"resource", "chassis_id", "manufacturer", "model", "part_number"}
	ChassisTemperatureLabelNames      = []string{"resource", "chassis_id", "sensor", "sensor_id"}
	ChassisFanLabelNames              = []string{"resource", "chassis_id", "fan", "fan_id"}
	ChassisPowerVoltageLabelNames     = []string{"resource", "chassis_id", "power_voltage", "power_voltage_id"}
	ChassisPowerSupplyLabelNames      = []string{"resource", "chassis_id", "power_supply", "power_supply_id"}
	ChassisNetworkAdapterLabelNames   = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id"}
	ChassisNetworkPortLabelNames      = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id", "network_port", "network_port_id", "network_port_type", "network_port_speed"}
	ChassisPhysicalSecurityLabelNames = []string{"resource", "chassis_id", "intrusion_sensor_number", "intrusion_sensor_rearm"}

	chassisMetrics = map[string]chassisMetric{
		"chassis_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "health"),
				"health of chassis, 1(OK),2(Warning),3(Critical)",
				ChassisLabelNames,
				nil,
			),
		},
		"chassis_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "state"),
				"state of chassis,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisLabelNames,
				nil,
			),
		},
		"chassis_model_info": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "model_info"),
				"organization responsible for producing the chassis, the name by which the manufacturer generally refers to the chassis, and a part number assigned by the organization that is responsible for producing or manufacturing the chassis",
				ChassisModel,
				nil,
			),
		},
		"chassis_temperature_sensor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_sensor_state"),
				"status state of temperature on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_temperature_celsius": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_celsius"),
				"celsius of temperature on this chassis component",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_fan_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_health"),
				"fan health on this chassis component,1(OK),2(Warning),3(Critical)",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_fan_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_state"),
				"fan state on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_fan_rpm": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_rpm_percentage"),
				"fan rpm percentage on this chassis component",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_power_voltage_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_voltage_state"),
				"power voltage state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisPowerVoltageLabelNames,
				nil,
			),
		},
		"chassis_power_voltage_volts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_voltage_volts"),
				"power voltage volts number of chassis component",
				ChassisPowerVoltageLabelNames,
				nil,
			),
		},
		"chassis_power_average_consumed_watts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_average_consumed_watts"),
				"power wattage watts number of chassis component",
				ChassisPowerVoltageLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_state"),
				"powersupply state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_health"),
				"powersupply health of chassis component,1(OK),2(Warning),3(Critical)",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_last_power_output_watts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_last_power_output_watts"),
				"last_power_output_watts of powersupply on this chassis",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_power_capacity_watts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_power_capacity_watts"),
				"power_capacity_watts of powersupply on this chassis",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_network_adapter_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_adapter_state"),
				"chassis network adapter state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisNetworkAdapterLabelNames,
				nil,
			),
		},
		"chassis_network_adapter_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_adapter_health_state"),
				"chassis network adapter health state,1(OK),2(Warning),3(Critical)",
				ChassisNetworkAdapterLabelNames,
				nil,
			),
		},
		"chassis_network_port_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_port_state"),
				"chassis network port state state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisNetworkPortLabelNames,
				nil,
			),
		},
		"chassis_network_port_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_port_health_state"),
				"chassis network port state state,1(OK),2(Warning),3(Critical)",
				ChassisNetworkPortLabelNames,
				nil,
			),
		},
		"chassis_physical_security_sensor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "physical_security_sensor_state"),
				"indicates the known state of the physical security sensor, such as if it is hardware intrusion detected, 1(Normal),2(TamperingDetected),3(HardwareIntrusion)",
				ChassisPhysicalSecurityLabelNames,
				nil,
			),
		},
	}
)

//ChassisCollector implements the prometheus.Collector.
type ChassisCollector struct {
	redfishClient         *gofish.APIClient
	metrics               map[string]chassisMetric
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *log.Entry
}

type chassisMetric struct {
	desc *prometheus.Desc
}

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(namespace string, redfishClient *gofish.APIClient, logger *log.Entry) *ChassisCollector {
	// get service from redfish client

	return &ChassisCollector{
		redfishClient: redfishClient,
		metrics:       chassisMetrics,
		Log: logger.WithFields(log.Fields{
			"collector": "ChassisCollector",
		}),
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

// Describe implemented prometheus.Collector
func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
	c.collectorScrapeStatus.Describe(ch)

}

// Collect implemented prometheus.Collector
func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := c.Log
	service := c.redfishClient.Service

	// get a list of chassis from service
	if chassises, err := service.Chassis(); err != nil {
		collectorLogContext.WithField("operation", "service.Chassis()").WithError(err).Error("error getting chassis from service")
	} else {
		// process the chassises
		for _, chassis := range chassises {
			chassisLogContext := collectorLogContext.WithField("Chassis", chassis.ID)
			chassisLogContext.Info("collector scrape started")
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

			chassisManufacturer := chassis.Manufacturer
			chassisModel := chassis.Model
			chassisPartNumber := chassis.PartNumber
			ChassisModelLabelValues := []string{"chassis", chassisID, chassisManufacturer, chassisModel, chassisPartNumber}
			ch <- prometheus.MustNewConstMetric(c.metrics["chassis_model_info"].desc, prometheus.GaugeValue, 1, ChassisModelLabelValues...)

			chassisThermal, err := chassis.Thermal()
			if err != nil {
				chassisLogContext.WithField("operation", "chassis.Thermal()").WithError(err).Error("error getting thermal data from chassis")
			} else if chassisThermal == nil {
				chassisLogContext.WithField("operation", "chassis.Thermal()").Info("no thermal data found")
			} else {
				// process temperature
				chassisTemperatures := chassisThermal.Temperatures
				wg := &sync.WaitGroup{}
				wg.Add(len(chassisTemperatures))

				for _, chassisTemperature := range chassisTemperatures {
					go parseChassisTemperature(ch, chassisID, chassisTemperature, wg)
				}

				// process fans

				chassisFans := chassisThermal.Fans
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(chassisFans))
				for _, chassisFan := range chassisFans {
					go parseChassisFan(ch, chassisID, chassisFan, wg2)
				}
			}

			chassisPowerInfo, err := chassis.Power()
			if err != nil {
				chassisLogContext.WithField("operation", "chassis.Power()").WithError(err).Error("error getting power data from chassis")
			} else if chassisPowerInfo == nil {
				chassisLogContext.WithField("operation", "chassis.Power()").Info("no power data found")
			} else {
				// power voltages
				chassisPowerInfoVoltages := chassisPowerInfo.Voltages
				wg3 := &sync.WaitGroup{}
				wg3.Add(len(chassisPowerInfoVoltages))
				for _, chassisPowerInfoVoltage := range chassisPowerInfoVoltages {
					go parseChassisPowerInfoVoltage(ch, chassisID, chassisPowerInfoVoltage, wg3)

				}

				// power control
				chassisPowerInfoPowerControls := chassisPowerInfo.PowerControl
				wg4 := &sync.WaitGroup{}
				wg4.Add(len(chassisPowerInfoPowerControls))
				for _, chassisPowerInfoPowerControl := range chassisPowerInfoPowerControls {
					go parseChassisPowerInfoPowerControl(ch, chassisID, chassisPowerInfoPowerControl, wg4)
				}

				// powerSupply
				chassisPowerInfoPowerSupplies := chassisPowerInfo.PowerSupplies
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(chassisPowerInfoPowerSupplies))
				for _, chassisPowerInfoPowerSupply := range chassisPowerInfoPowerSupplies {

					go parseChassisPowerInfoPowerSupply(ch, chassisID, chassisPowerInfoPowerSupply, wg5)
				}
			}

			// process NetapAdapter

			networkAdapters, err := chassis.NetworkAdapters()
			if err != nil {
				chassisLogContext.WithField("operation", "chassis.NetworkAdapters()").WithError(err).Error("error getting network adapters data from chassis")
			} else if networkAdapters == nil {
				chassisLogContext.WithField("operation", "chassis.NetworkAdapters()").Info("no network adapters data found")
			} else {
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(networkAdapters))

				for _, networkAdapter := range networkAdapters {
					if err = parseNetworkAdapter(ch, chassisID, networkAdapter, wg5); err != nil {
						chassisLogContext.WithField("operation", "chassis.NetworkAdapters()").WithError(err).Error("error getting network ports from network adapter")
					}
				}
			}

			physicalSecurity := chassis.PhysicalSecurity
			if physicalSecurity != (redfish.PhysicalSecurity{}) {
				physicalSecurityIntrusionSensor := physicalSecurity.IntrusionSensor
				physicalSecurityIntrusionSensorNumber := fmt.Sprintf("%d", physicalSecurity.IntrusionSensorNumber)
				physicalSecurityIntrusionSensorReArmMethod := string(physicalSecurity.IntrusionSensorReArm)

				if phySecIntrusionSensor, ok := parsePhySecIntrusionSensor(physicalSecurityIntrusionSensor); ok {
					ChassisPhysicalSecurityLabelValues := []string{"physical_security", chassisID, physicalSecurityIntrusionSensorNumber, physicalSecurityIntrusionSensorReArmMethod}
					ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_physical_security_sensor_state"].desc, prometheus.GaugeValue, phySecIntrusionSensor, ChassisPhysicalSecurityLabelValues...)

				}
			}
			chassisLogContext.Info("collector scrape completed")
		}
	}

	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}

func parseChassisTemperature(ch chan<- prometheus.Metric, chassisID string, chassisTemperature redfish.Temperature, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisTemperatureSensorName := chassisTemperature.Name
	chassisTemperatureSensorID := chassisTemperature.MemberID
	chassisTemperatureStatus := chassisTemperature.Status
	//			chassisTemperatureStatusHealth :=chassisTemperatureStatus.Health
	chassisTemperatureStatusState := chassisTemperatureStatus.State
	//			chassisTemperatureStatusLabelNames :=[]string{BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
	chassisTemperatureLabelvalues := []string{"temperature", chassisID, chassisTemperatureSensorName, chassisTemperatureSensorID}

	//		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_status_health"].desc, prometheus.GaugeValue, parseCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureLabelvalues...)
	if chassisTemperatureStatusStateValue, ok := parseCommonStatusState(chassisTemperatureStatusState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_sensor_state"].desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
	}

	chassisTemperatureReadingCelsius := chassisTemperature.ReadingCelsius
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_celsius"].desc, prometheus.GaugeValue, float64(chassisTemperatureReadingCelsius), chassisTemperatureLabelvalues...)
}
func parseChassisFan(ch chan<- prometheus.Metric, chassisID string, chassisFan redfish.Fan, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisFanID := chassisFan.MemberID
	chassisFanName := chassisFan.Name
	chassisFanStaus := chassisFan.Status
	chassisFanStausHealth := chassisFanStaus.Health
	chassisFanStausState := chassisFanStaus.State
	chassisFanRPM := chassisFan.Reading

	//			chassisFanStatusLabelNames :=[]string{BaseLabelNames,"fan_name","fan_member_id")
	chassisFanLabelvalues := []string{"fan", chassisID, chassisFanName, chassisFanID}

	if chassisFanStausHealthValue, ok := parseCommonStatusHealth(chassisFanStausHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_health"].desc, prometheus.GaugeValue, chassisFanStausHealthValue, chassisFanLabelvalues...)
	}

	if chassisFanStausStateValue, ok := parseCommonStatusState(chassisFanStausState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_state"].desc, prometheus.GaugeValue, chassisFanStausStateValue, chassisFanLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm"].desc, prometheus.GaugeValue, float64(chassisFanRPM), chassisFanLabelvalues...)

}

func parseChassisPowerInfoVoltage(ch chan<- prometheus.Metric, chassisID string, chassisPowerInfoVoltage redfish.Voltage, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisPowerInfoVoltageName := chassisPowerInfoVoltage.Name
	chassisPowerInfoVoltageID := chassisPowerInfoVoltage.MemberID
	chassisPowerInfoVoltageNameReadingVolts := chassisPowerInfoVoltage.ReadingVolts
	chassisPowerInfoVoltageState := chassisPowerInfoVoltage.Status.State
	chassisPowerVoltageLabelvalues := []string{"power_voltage", chassisID, chassisPowerInfoVoltageName, chassisPowerInfoVoltageID}
	if chassisPowerInfoVoltageStateValue, ok := parseCommonStatusState(chassisPowerInfoVoltageState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_voltage_state"].desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, chassisPowerVoltageLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_voltage_volts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoVoltageNameReadingVolts), chassisPowerVoltageLabelvalues...)
}
func parseChassisPowerInfoPowerControl(ch chan<- prometheus.Metric, chassisID string, chassisPowerInfoPowerControl redfish.PowerControl, wg *sync.WaitGroup) {
	defer wg.Done()
	name := chassisPowerInfoPowerControl.Name
	id := chassisPowerInfoPowerControl.MemberID
	pm := chassisPowerInfoPowerControl.PowerMetrics
	chassisPowerVoltageLabelvalues := []string{"power_wattage", chassisID, name, id}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_average_consumed_watts"].desc, prometheus.GaugeValue, float64(pm.AverageConsumedWatts), chassisPowerVoltageLabelvalues...)
}

func parseChassisPowerInfoPowerSupply(ch chan<- prometheus.Metric, chassisID string, chassisPowerInfoPowerSupply redfish.PowerSupply, wg *sync.WaitGroup) {

	defer wg.Done()
	chassisPowerInfoPowerSupplyName := chassisPowerInfoPowerSupply.Name
	chassisPowerInfoPowerSupplyID := chassisPowerInfoPowerSupply.MemberID
	chassisPowerInfoPowerSupplyPowerCapacityWatts := chassisPowerInfoPowerSupply.PowerCapacityWatts
	chassisPowerInfoPowerSupplyLastPowerOutputWatts := chassisPowerInfoPowerSupply.LastPowerOutputWatts
	chassisPowerInfoPowerSupplyState := chassisPowerInfoPowerSupply.Status.State
	chassisPowerInfoPowerSupplyHealth := chassisPowerInfoPowerSupply.Status.Health
	chassisPowerSupplyLabelvalues := []string{"power_supply", chassisID, chassisPowerInfoPowerSupplyName, chassisPowerInfoPowerSupplyID}
	if chassisPowerInfoPowerSupplyStateValue, ok := parseCommonStatusState(chassisPowerInfoPowerSupplyState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_state"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, chassisPowerSupplyLabelvalues...)
	}
	if chassisPowerInfoPowerSupplyHealthValue, ok := parseCommonStatusHealth(chassisPowerInfoPowerSupplyHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_health"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, chassisPowerSupplyLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_last_power_output_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyLastPowerOutputWatts), chassisPowerSupplyLabelvalues...)
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_capacity_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyPowerCapacityWatts), chassisPowerSupplyLabelvalues...)
}

func parseNetworkAdapter(ch chan<- prometheus.Metric, chassisID string, networkAdapter *redfish.NetworkAdapter, wg *sync.WaitGroup) error {

	defer wg.Done()
	networkAdapterName := networkAdapter.Name
	networkAdapterID := networkAdapter.ID
	networkAdapterState := networkAdapter.Status.State
	networkAdapterHealthState := networkAdapter.Status.Health
	chassisNetworkAdapterLabelValues := []string{"network_adapter", chassisID, networkAdapterName, networkAdapterID}
	if networkAdapterStateValue, ok := parseCommonStatusState(networkAdapterState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_adapter_state"].desc, prometheus.GaugeValue, networkAdapterStateValue, chassisNetworkAdapterLabelValues...)
	}
	if networkAdapterHealthStateValue, ok := parseCommonStatusHealth(networkAdapterHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_adapter_health_state"].desc, prometheus.GaugeValue, networkAdapterHealthStateValue, chassisNetworkAdapterLabelValues...)
	}

	if networkPorts, err := networkAdapter.NetworkPorts(); err != nil {
		return err
	} else {
		wg6 := &sync.WaitGroup{}
		wg6.Add(len(networkPorts))
		for _, networkPort := range networkPorts {
			go parseNetworkPort(ch, chassisID, networkPort, networkAdapterName, networkAdapterID, wg6)
		}
	}
	return nil
}

func parseNetworkPort(ch chan<- prometheus.Metric, chassisID string, networkPort *redfish.NetworkPort, networkAdapterName string, networkAdapterID string, wg *sync.WaitGroup) {
	defer wg.Done()
	networkPortName := networkPort.Name
	networkPortID := networkPort.ID
	networkPortState := networkPort.Status.State
	networkPortLinkType := networkPort.ActiveLinkTechnology
	networkPortLinkSpeed := fmt.Sprintf("%d Mbps", networkPort.CurrentLinkSpeedMbps)
	networkPortHealthState := networkPort.Status.Health
	chassisNetworkPortLabelValues := []string{"network_port", chassisID, networkAdapterName, networkAdapterID, networkPortName, networkPortID, string(networkPortLinkType), networkPortLinkSpeed}
	if networkPortStateValue, ok := parseCommonStatusState(networkPortState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_state"].desc, prometheus.GaugeValue, networkPortStateValue, chassisNetworkPortLabelValues...)
	}
	if networkPortHealthStateValue, ok := parseCommonStatusHealth(networkPortHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_health_state"].desc, prometheus.GaugeValue, networkPortHealthStateValue, chassisNetworkPortLabelValues...)
	}
}
