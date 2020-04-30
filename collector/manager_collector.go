package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"sync"
)

// A ManagerCollector implements the prometheus.Collector.

type managerMetric struct {
	desc *prometheus.Desc
}

var (
	ManagerSubmanager                   = "manager"
	ManagerLabelNames                  = []string{"manager_id","name", "model", "type" }
	managerMetrics                     = map[string]managerMetric{
		"manager_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ManagerSubmanager, "state"),
				"manager state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ManagerLabelNames,
				nil,
			),
		},
		"manager_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ManagerSubmanager, "health_state"),
				"manager health,1(OK),2(Warning),3(Critical)",
				ManagerLabelNames,
				nil,
			),
		},		
		"manager_firmware_version": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ManagerSubmanager, "firmware_version"),
				"firmware version of the manager",
				ManagerLabelNames,
				nil,
			),
		},
		"manager_power_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ManagerSubmanager, "power_state"),
				"manager power state",
				ManagerLabelNames,
				nil,
			),
		},
	}
)

type ManagerCollector struct {
	redfishClient           *gofish.APIClient
	metrics                 map[string]managerMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

// NewManagerCollector returns a collector that collecting memory statistics
func NewManagerCollector(namespace string, redfishClient *gofish.APIClient) *ManagerCollector {
	return &ManagerCollector{
		redfishClient: redfishClient,
		metrics:       managerMetrics,
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

func (s *ManagerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.desc
	}
	s.collectorScrapeStatus.Describe(ch)

}

func (s *ManagerCollector) Collect(ch chan<- prometheus.Metric) {
	//get service
	service := s.redfishClient.Service

	// get a list of managers from service
	if managers, err := service.Managers(); err != nil {
		log.Infof("Errors Getting managers from service : %s", err)
	} else {

		for _, manager := range managers {
			// overall manager metrics
			[]string{"manager_id","name", "model", "type" }

			ManagerID := manager.ID
			managerName := manager.Name
			managerModel := manager.model
			managerType := manager.Type
			managerPowerState := manager.PowerState
			managerState := manager.Status.State
			managerHealthState := manager.Status.Health

			ManagerLabelValues := []string{ManagerID,managerName, managerModel,managerType}
			if managerHealthStateValue, ok := parseCommonStatusHealth(managerHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["manager_state"].desc, prometheus.GaugeValue, managerHealthStateValue, ManagerLabelValues...)
			}
			if managerStateValue, ok := parseCommonStatusState(managerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["manager_state"].desc, prometheus.GaugeValue, managerStateValue, ManagerLabelValues...)
			}
			if managerPowerStateValue, ok := parseManagerPowerState(managerPowerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["manager_power_state"].desc, prometheus.GaugeValue, managerPowerStateValue, ManagerLabelValues...)

			}
		}
		s.collectorScrapeStatus.WithLabelValues("manager").Set(float64(1))

	}

}

func parseMemory(ch chan<- prometheus.Metric, managerHostName string, memory *redfish.Memory, wg *sync.WaitGroup) {
	defer wg.Done()
	memoryName := memory.Name
	memoryId := memory.ID
	//memoryDeviceLocator := memory.DeviceLocator
	memoryCapacityMiB := memory.CapacityMiB
	memoryState := memory.Status.State
	memoryHealthState := memory.Status.Health

	managerMemoryLabelValues := []string{managerHostName, "memory", memoryName, memoryId}
	if memoryStateValue, ok := parseCommonStatusState(memoryState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, managerMemoryLabelValues...)

	}
	if memoryHealthStateValue, ok := parseCommonStatusHealth(memoryHealthState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_memory_health_state"].desc, prometheus.GaugeValue, memoryHealthStateValue, managerMemoryLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(managerMetrics["manager_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), managerMemoryLabelValues...)

}

func parsePorcessor(ch chan<- prometheus.Metric, managerHostName string, processor *redfish.Processor, wg *sync.WaitGroup) {
	defer wg.Done()
	processorName := processor.Name
	processorID := processor.ID
	processorTotalCores := processor.TotalCores
	processorTotalThreads := processor.TotalThreads
	processorState := processor.Status.State
	processorHelathState := processor.Status.Health

	managerProcessorLabelValues := []string{managerHostName, "processor", processorName, processorID}

	if processorStateValue, ok := parseCommonStatusState(processorState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_processor_state"].desc, prometheus.GaugeValue, processorStateValue, managerProcessorLabelValues...)

	}
	if processorHelathStateValue, ok := parseCommonStatusHealth(processorHelathState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_processor_health_state"].desc, prometheus.GaugeValue, processorHelathStateValue, managerProcessorLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(managerMetrics["manager_processor_total_threads"].desc, prometheus.GaugeValue, float64(processorTotalThreads), managerProcessorLabelValues...)
	ch <- prometheus.MustNewConstMetric(managerMetrics["manager_processor_total_cores"].desc, prometheus.GaugeValue, float64(processorTotalCores), managerProcessorLabelValues...)
}
func parseVolume(ch chan<- prometheus.Metric, managerHostName string, volume *redfish.Volume, wg *sync.WaitGroup) {
	defer wg.Done()
	volumeName := volume.Name
	volumeID := volume.ID
	volumeCapacityBytes := volume.CapacityBytes
	volumeState := volume.Status.State
	volumeHealthState := volume.Status.Health
	managerVolumeLabelValues := []string{managerHostName, "volume", volumeName, volumeID}
	if volumeStateValue, ok := parseCommonStatusState(volumeState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_storage_volume_state"].desc, prometheus.GaugeValue, volumeStateValue, managerVolumeLabelValues...)

	}
	if volumeHealthStateValue, ok := parseCommonStatusHealth(volumeHealthState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_storage_volume_health_state"].desc, prometheus.GaugeValue, volumeHealthStateValue, managerVolumeLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(managerMetrics["manager_storage_volume_capacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), managerVolumeLabelValues...)
}
func parseDrive(ch chan<- prometheus.Metric, managerHostName string, drive *redfish.Drive, wg *sync.WaitGroup) {
	defer wg.Done()
	driveName := drive.Name
	driveID := drive.ID
	driveCapacityBytes := drive.CapacityBytes
	driveState := drive.Status.State
	driveHealthState := drive.Status.Health
	managerdriveLabelValues := []string{managerHostName, "drive", driveName, driveID}
	if driveStateValue, ok := parseCommonStatusState(driveState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_storage_drive_state"].desc, prometheus.GaugeValue, driveStateValue, managerdriveLabelValues...)

	}
	if driveHealthStateValue, ok := parseCommonStatusHealth(driveHealthState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_storage_drive_health_state"].desc, prometheus.GaugeValue, driveHealthStateValue, managerdriveLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(managerMetrics["manager_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), managerdriveLabelValues...)
}

func parsePcieDevice(ch chan<- prometheus.Metric, managerHostName string, pcieDevice *redfish.PCIeDevice, wg *sync.WaitGroup) {

	defer wg.Done()
	pcieDeviceName := pcieDevice.Name
	pcieDeviceID := pcieDevice.ID
	pcieDeviceState := pcieDevice.Status.State
	pcieDeviceHealthState := pcieDevice.Status.Health
	managerPCIeDeviceLabelValues := []string{managerHostName, "pcie_device", pcieDeviceName, pcieDeviceID}

	if pcieStateVaule, ok := parseCommonStatusState(pcieDeviceState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_pcie_device_state"].desc, prometheus.GaugeValue, pcieStateVaule, managerPCIeDeviceLabelValues...)

	}
	if pcieHealthStateVaule, ok := parseCommonStatusHealth(pcieDeviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_pcie_device_health_state"].desc, prometheus.GaugeValue, pcieHealthStateVaule, managerPCIeDeviceLabelValues...)

	}
}

func parseNetworkInterface(ch chan<- prometheus.Metric, managerHostName string, networkInterface *redfish.NetworkInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	networkInterfaceName := networkInterface.Name
	networkInterfaceID := networkInterface.ID
	networkInterfaceState := networkInterface.Status.State
	networkInterfaceHealthState := networkInterface.Status.Health
	managerNetworkInterfaceLabelValues := []string{managerHostName, "network_interface", networkInterfaceName, networkInterfaceID}

	if networknetworkInterfaceStateVaule, ok := parseCommonStatusState(networkInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_network_interface_state"].desc, prometheus.GaugeValue, networknetworkInterfaceStateVaule, managerNetworkInterfaceLabelValues...)

	}
	if networknetworkInterfaceHealthStateVaule, ok := parseCommonStatusHealth(networkInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_network_interface_health_state"].desc, prometheus.GaugeValue, networknetworkInterfaceHealthStateVaule, managerNetworkInterfaceLabelValues...)

	}
}

func parseEthernetInterface(ch chan<- prometheus.Metric, managerHostName string, ethernetInterface *redfish.EthernetInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	//ethernetInterfaceODataIDslice := strings.Split(ethernetInterface.ODataID, "/")
	//ethernetInterfaceName := ethernetInterfaceODataIDslice[len(ethernetInterfaceODataIDslice)-1]
	ethernetInterfaceName := ethernetInterface.Name
	ethernetInterfaceID := ethernetInterface.ID
	ethernetInterfaceLinkStatus := ethernetInterface.LinkStatus
	ethernetInterfaceEnabled := ethernetInterface.InterfaceEnabled
	ethernetInterfaceSpeed := fmt.Sprintf("%d Mbps", ethernetInterface.SpeedMbps)
	ethernetInterfaceState := ethernetInterface.Status.State
	ethernetInterfaceHealthState := ethernetInterface.Status.Health
	managerEthernetInterfaceLabelValues := []string{managerHostName, "ethernet_interface", ethernetInterfaceName, ethernetInterfaceID, ethernetInterfaceSpeed}
	if ethernetInterfaceStateValue, ok := parseCommonStatusState(ethernetInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_etherenet_interface_state"].desc, prometheus.GaugeValue, ethernetInterfaceStateValue, managerEthernetInterfaceLabelValues...)

	}
	if ethernetInterfaceHealthStateValue, ok := parseCommonStatusHealth(ethernetInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_etherenet_interface_health_state"].desc, prometheus.GaugeValue, ethernetInterfaceHealthStateValue, managerEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceLinkStatusValue, ok := parseLinkStatus(ethernetInterfaceLinkStatus); ok {

		ch <- prometheus.MustNewConstMetric(managerMetrics["manager_etherenet_interface_link_status"].desc, prometheus.GaugeValue, ethernetInterfaceLinkStatusValue, managerEthernetInterfaceLabelValues...)

	}

	ch <- prometheus.MustNewConstMetric(managerMetrics["manager_etherenet_interface_link_enabled"].desc, prometheus.GaugeValue, boolToFloat64(ethernetInterfaceEnabled), managerEthernetInterfaceLabelValues...)

}
