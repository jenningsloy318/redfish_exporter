package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"sync"
)

// A SystemCollector implements the prometheus.Collector.

type systemMetric struct {
	desc *prometheus.Desc
}

var (
	subsystem_s                         = "system"
	SystemLabelNames                  = []string{"hostname", "resource", "system_id"}
	SystemMemoryLabelNames            = []string{"hostname", "resource", "memory", "memory_id"}
	SystemProcessorLabelNames         = []string{"hostname", "resource", "processor", "processor_id"}
	SystemVolumeLabelNames            = []string{"hostname", "resource", "volume", "volume_id"}
	SystemDriveLabelNames             = []string{"hostname", "resource", "drive", "drive_id"}
	SystemStorageControllerLabelNames = []string{"hostname", "resource", "storage_controller", "storage_controller_id"}
	SystemPCIeDeviceLabelNames        = []string{"hostname", "resource", "pcie_device", "pcie_device_id"}
	SystemNetworkInterfaceLabelNames  = []string{"hostname", "resource", "network_interface", "network_interface_id"}
	SystemEthernetInterfaceLabelNames = []string{"hostname", "resource", "ethernet_interface", "ethernet_interface_id", "ethernet_interface_speed"}
	systemMetrics                     = map[string]systemMetric{
		"system_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "state"),
				"system state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "health_state"),
				"system health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_power_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "power_state"),
				"system power state",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "total_memory_state"),
				"system overall memory state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "total_memory_health_state"),
				"system overall memory health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_size": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "total_memory_size"),
				"system total memory size, GiB",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_processor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "total_processor_state"),
				"system overall processor state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_processor_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "total_processor_health_state"),
				"system overall processor health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_processor_count": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "total_processor_count"),
				"system total  processor count",
				SystemLabelNames,
				nil,
			),
		},
		"system_memory_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "memory_state"),
				"system memory state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_memory_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "memory_health_state"),
				"system memory  health state,1(OK),2(Warning),3(Critical)",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_memory_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "memory_capacity"),
				"system memory capacity, MiB",
				SystemMemoryLabelNames,
				nil,
			),
		},

		"system_processor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "processor_state"),
				"system processor state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "processor_health_state"),
				"system processor  health state,1(OK),2(Warning),3(Critical)",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_total_threads": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "processor_total_threads"),
				"system processor total threads",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_total_cores": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "processor_total_cores"),
				"system processor total cores",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_storage_volume_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_volume_state"),
				"system storage volume state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_volume_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_volume_health_state"),
				"system storage volume health state,1(OK),2(Warning),3(Critical)",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_volume_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_volume_capacity"),
				"system storage volume capacity,Bytes",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_drive_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_drive_state"),
				"system storage drive state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_drive_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_drive_health_state"),
				"system storage volume health state,1(OK),2(Warning),3(Critical)",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_drive_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_drive_capacity"),
				"system storage drive capacity,Bytes",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_controller_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_controller_state"),
				"system storage controller state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemStorageControllerLabelNames,
				nil,
			),
		},
		"system_storage_controller_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "storage_controller_health_state"),
				"system storage controller health state,1(OK),2(Warning),3(Critical)",
				SystemStorageControllerLabelNames,
				nil,
			),
		},
		"system_pcie_device_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "pcie_device_state"),
				"system pcie device state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemPCIeDeviceLabelNames,
				nil,
			),
		},
		"system_pcie_device_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "pcie_device_health_state"),
				"system pcie device health state,1(OK),2(Warning),3(Critical)",
				SystemPCIeDeviceLabelNames,
				nil,
			),
		},
		"system_network_interface_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "network_interface_state"),
				"system network interface state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemNetworkInterfaceLabelNames,
				nil,
			),
		},
		"system_network_interface_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "network_interface_health_state"),
				"system network interface health state,1(OK),2(Warning),3(Critical)",
				SystemNetworkInterfaceLabelNames,
				nil,
			),
		},
		"system_etherenet_interface_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "etherenet_interface_state"),
				"system ethernet interface state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
		"system_etherenet_interface_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "etherenet_interface_health_state"),
				"system ethernet interface health state,1(OK),2(Warning),3(Critical)",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
		"system_etherenet_interface_link_status": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "etherenet_interface_link_status"),
				"system ethernet interface link status，1(LinkUp),2(NoLink),3(LinkDown)",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
		"system_etherenet_interface_link_enabled": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem_s, "etherenet_interface_link_enabled"),
				"system ethernet interface if the link is enabled",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
	}
)

type SystemCollector struct {
	redfishClient           *gofish.APIClient
	metrics                 map[string]systemMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(namespace string, redfishClient *gofish.APIClient) *SystemCollector {
	var ()
	return &SystemCollector{
		redfishClient: redfishClient,
		metrics:       systemMetrics,
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

func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.desc
	}
	s.collectorScrapeStatus.Describe(ch)
	s.collectorScrapeDuration.Describe(ch)

}

func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	//get service
	service := s.redfishClient.Service

	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		log.Infof("Errors Getting systems from service : %s", err)
	} else {

		for _, system := range systems {
			// overall system metrics

			SystemID := system.ID
			systemHostName := system.HostName
			systemPowerState := system.PowerState
			systemState := system.Status.State
			systemHealthState := system.Status.Health
			systemTotalProcessorCount := system.ProcessorSummary.Count
			systemTotalProcessorsState := system.ProcessorSummary.Status.State
			systemTotalProcessorsHealthState := system.ProcessorSummary.Status.Health
			systemTotalMemoryState := system.MemorySummary.Status.State
			systemTotalMemoryHealthState := system.MemorySummary.Status.Health
			systemTotalMemoryAmount := system.MemorySummary.TotalSystemMemoryGiB

			systemLabelValues := []string{systemHostName, "system", SystemID}
			if systemHealthStateValue, ok := parseCommonStatusHealth(systemHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].desc, prometheus.GaugeValue, systemHealthStateValue, systemLabelValues...)
			}
			if systemStateValue, ok := parseCommonStatusState(systemState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue, systemStateValue, systemLabelValues...)
			}
			if systemPowerStateValue, ok := parseSystemPowerState(systemPowerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].desc, prometheus.GaugeValue, systemPowerStateValue, systemLabelValues...)

			}
			if systemTotalProcessorsStateValue, ok := parseCommonStatusState(systemTotalProcessorsState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].desc, prometheus.GaugeValue, systemTotalProcessorsStateValue, systemLabelValues...)

			}
			if systemTotalProcessorsHealthStateValue, ok := parseCommonStatusHealth(systemTotalProcessorsHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].desc, prometheus.GaugeValue, systemTotalProcessorsHealthStateValue, systemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].desc, prometheus.GaugeValue, float64(systemTotalProcessorCount), systemLabelValues...)

			if systemTotalMemoryStateValue, ok := parseCommonStatusState(systemTotalMemoryState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].desc, prometheus.GaugeValue, systemTotalMemoryStateValue, systemLabelValues...)

			}
			if systemTotalMemoryHealthStateValue, ok := parseCommonStatusHealth(systemTotalMemoryHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].desc, prometheus.GaugeValue, systemTotalMemoryHealthStateValue, systemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].desc, prometheus.GaugeValue, float64(systemTotalMemoryAmount), systemLabelValues...)

			// get system OdataID
			//systemOdataID := system.ODataID

			// process memory metrics
			// construct memory Link
			//memoriesLink := fmt.Sprintf("%sMemory/", systemOdataID)

			//if memories, err := redfish.ListReferencedMemorys(s.redfishClient, memoriesLink); err != nil {
			if memories, err := system.Memory(); err != nil {
				log.Infof("Errors Getting memory from computer system : %s", err)
			} else {
				wg1 := &sync.WaitGroup{}
				wg1.Add(len(memories))

				for _, memory := range memories {
					go parseMemory(ch, systemHostName, memory, wg1)

				}
			}

			// process processor metrics

			//processorsLink := fmt.Sprintf("%sProcessors/", systemOdataID)

			//if processors, err := redfish.ListReferencedProcessors(s.redfishClient, processorsLink); err != nil {
			if processors, err := system.Processors(); err != nil {
				log.Infof("Errors Getting Processors from system: %s", err)
			} else {
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(processors))

				for _, processor := range processors {
					go parsePorcessor(ch, systemHostName, processor, wg2)

				}

			}

			//process storage
			//storagesLink := fmt.Sprintf("%sStorage/", systemOdataID)

			//if storages, err := redfish.ListReferencedStorages(s.redfishClient, storagesLink); err != nil {
			if storages, err := system.Storage(); err != nil {
				log.Infof("Errors Getting storages from system: %s", err)
			} else {
				for _, storage := range storages {

					if volumes, err := storage.Volumes(); err != nil {
						log.Infof("Errors Getting volumes  from system storage : %s", err)
					} else {
						wg3 := &sync.WaitGroup{}
						wg3.Add(len(volumes))

						for _, volume := range volumes {
							go parseVolume(ch, systemHostName, volume, wg3)

						}
					}

					if drives, err := storage.Drives(); err != nil {
						log.Infof("Errors Getting volumes  from system storage : %s", err)
					} else {
						wg4 := &sync.WaitGroup{}
						wg4.Add(len(drives))
						for _, drive := range drives {
							go parseDrive(ch, systemHostName, drive, wg4)
						}
					}

					//					if storagecontrollers, err := storage.StorageControllers(); err != nil {
					//						log.Infof("Errors Getting storagecontrollers from system storage : %s", err)
					//					} else {
					//
					//						for _, controller := range storagecontrollers {
					//
					//							controllerODataIDslice := strings.Split(controller.ODataID, "/")
					//							controllerName := controllerODataIDslice[len(controllerODataIDslice)-1]
					//							controllerState := controller.Status.State
					//							controllerHealthState := controller.Status.Health
					//							controllerLabelValues := []string{ "storage_controller", controllerName, systemHostName)
					//							if controllerStateValue,ok := parseCommonStatusState(controllerState); ok {
					//								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_state"].desc, prometheus.GaugeValue, controllerStateValue, //controllerLabelValues...)
					//
					//							}
					//							if controllerHealthStateValue,ok := parseCommonStatusHealth(controllerHealthState); ok {
					//								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_health_state"].desc, prometheus.GaugeValue, controllerHealthStateValue, //controllerLabelValues...)
					//
					//							}
					//
					//						}
					//
					//					}

				}
			}
			//process pci devices
			//pciDevicesLink := fmt.Sprintf("%sPcidevice/", systemOdataID)
			if pcieDevices, err := system.PCIeDevices(); err != nil {
				log.Infof("Errors Getting PCI-E devices from system: %s", err)
			} else {
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(pcieDevices))
				for _, pcieDevice := range pcieDevices {
					go parsePcieDevice(ch, systemHostName, pcieDevice, wg5)
				}
			}

			//process networkinterfaces
			if networkInterfaces, err := system.NetworkInterfaces(); err != nil {
				log.Infof("Errors Getting network Interfaces from system: %s", err)
			} else {
				wg6 := &sync.WaitGroup{}
				wg6.Add(len(networkInterfaces))
				for _, networkInterface := range networkInterfaces {
					go parseNetworkInterface(ch, systemHostName, networkInterface, wg6)
				}

			}

			//process nethernetinterfaces
			if ethernetInterfaces, err := system.EthernetInterfaces(); err != nil {
				log.Infof("Errors Getting ethernet Interfaces from system: %s", err)
			} else {
				wg7 := &sync.WaitGroup{}
				wg7.Add(len(ethernetInterfaces))

				for _, ethernetInterface := range ethernetInterfaces {
					go parseEthernetInterface(ch, systemHostName, ethernetInterface, wg7)
				}

			}
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))

	}

}

func parseMemory(ch chan<- prometheus.Metric, systemHostName string, memory *redfish.Memory, wg *sync.WaitGroup) {
	defer wg.Done()
	memoryName := memory.Name
	memoryId := memory.ID
	//memoryDeviceLocator := memory.DeviceLocator
	memoryCapacityMiB := memory.CapacityMiB
	memoryState := memory.Status.State
	memoryHealthState := memory.Status.Health

	systemMemoryLabelValues := []string{systemHostName, "memory", memoryName, memoryId}
	if memoryStateValue, ok := parseCommonStatusState(memoryState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, systemMemoryLabelValues...)

	}
	if memoryHealthStateValue, ok := parseCommonStatusHealth(memoryHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_health_state"].desc, prometheus.GaugeValue, memoryHealthStateValue, systemMemoryLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), systemMemoryLabelValues...)

}

func parsePorcessor(ch chan<- prometheus.Metric, systemHostName string, processor *redfish.Processor, wg *sync.WaitGroup) {
	defer wg.Done()
	processorName := processor.Name
	processorID := processor.ID
	processorTotalCores := processor.TotalCores
	processorTotalThreads := processor.TotalThreads
	processorState := processor.Status.State
	processorHelathState := processor.Status.Health

	systemProcessorLabelValues := []string{systemHostName, "processor", processorName, processorID}

	if processorStateValue, ok := parseCommonStatusState(processorState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_state"].desc, prometheus.GaugeValue, processorStateValue, systemProcessorLabelValues...)

	}
	if processorHelathStateValue, ok := parseCommonStatusHealth(processorHelathState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_health_state"].desc, prometheus.GaugeValue, processorHelathStateValue, systemProcessorLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_threads"].desc, prometheus.GaugeValue, float64(processorTotalThreads), systemProcessorLabelValues...)
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_cores"].desc, prometheus.GaugeValue, float64(processorTotalCores), systemProcessorLabelValues...)
}
func parseVolume(ch chan<- prometheus.Metric, systemHostName string, volume *redfish.Volume, wg *sync.WaitGroup) {
	defer wg.Done()
	volumeName := volume.Name
	volumeID := volume.ID
	volumeCapacityBytes := volume.CapacityBytes
	volumeState := volume.Status.State
	volumeHealthState := volume.Status.Health
	systemVolumeLabelValues := []string{systemHostName, "volume", volumeName, volumeID}
	if volumeStateValue, ok := parseCommonStatusState(volumeState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_state"].desc, prometheus.GaugeValue, volumeStateValue, systemVolumeLabelValues...)

	}
	if volumeHealthStateValue, ok := parseCommonStatusHealth(volumeHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_health_state"].desc, prometheus.GaugeValue, volumeHealthStateValue, systemVolumeLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_capacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), systemVolumeLabelValues...)
}
func parseDrive(ch chan<- prometheus.Metric, systemHostName string, drive *redfish.Drive, wg *sync.WaitGroup) {
	defer wg.Done()
	driveName := drive.Name
	driveID := drive.ID
	driveCapacityBytes := drive.CapacityBytes
	driveState := drive.Status.State
	driveHealthState := drive.Status.Health
	systemdriveLabelValues := []string{systemHostName, "drive", driveName, driveID}
	if driveStateValue, ok := parseCommonStatusState(driveState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_state"].desc, prometheus.GaugeValue, driveStateValue, systemdriveLabelValues...)

	}
	if driveHealthStateValue, ok := parseCommonStatusHealth(driveHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_health_state"].desc, prometheus.GaugeValue, driveHealthStateValue, systemdriveLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), systemdriveLabelValues...)
}

func parsePcieDevice(ch chan<- prometheus.Metric, systemHostName string, pcieDevice *redfish.PCIeDevice, wg *sync.WaitGroup) {

	defer wg.Done()
	pcieDeviceName := pcieDevice.Name
	pcieDeviceID := pcieDevice.ID
	pcieDeviceState := pcieDevice.Status.State
	pcieDeviceHealthState := pcieDevice.Status.Health
	systemPCIeDeviceLabelValues := []string{systemHostName, "pcie_device", pcieDeviceName, pcieDeviceID}

	if pcieStateVaule, ok := parseCommonStatusState(pcieDeviceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_device_state"].desc, prometheus.GaugeValue, pcieStateVaule, systemPCIeDeviceLabelValues...)

	}
	if pcieHealthStateVaule, ok := parseCommonStatusHealth(pcieDeviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_device_health_state"].desc, prometheus.GaugeValue, pcieHealthStateVaule, systemPCIeDeviceLabelValues...)

	}
}

func parseNetworkInterface(ch chan<- prometheus.Metric, systemHostName string, networkInterface *redfish.NetworkInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	networkInterfaceName := networkInterface.Name
	networkInterfaceID := networkInterface.ID
	networkInterfaceState := networkInterface.Status.State
	networkInterfaceHealthState := networkInterface.Status.Health
	systemNetworkInterfaceLabelValues := []string{systemHostName, "network_interface", networkInterfaceName, networkInterfaceID}

	if networknetworkInterfaceStateVaule, ok := parseCommonStatusState(networkInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_network_interface_state"].desc, prometheus.GaugeValue, networknetworkInterfaceStateVaule, systemNetworkInterfaceLabelValues...)

	}
	if networknetworkInterfaceHealthStateVaule, ok := parseCommonStatusHealth(networkInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_network_interface_health_state"].desc, prometheus.GaugeValue, networknetworkInterfaceHealthStateVaule, systemNetworkInterfaceLabelValues...)

	}
}

func parseEthernetInterface(ch chan<- prometheus.Metric, systemHostName string, ethernetInterface *redfish.EthernetInterface, wg *sync.WaitGroup) {
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
	systemEthernetInterfaceLabelValues := []string{systemHostName, "ethernet_interface", ethernetInterfaceName, ethernetInterfaceID, ethernetInterfaceSpeed}
	if ethernetInterfaceStateValue, ok := parseCommonStatusState(ethernetInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_etherenet_interface_state"].desc, prometheus.GaugeValue, ethernetInterfaceStateValue, systemEthernetInterfaceLabelValues...)

	}
	if ethernetInterfaceHealthStateValue, ok := parseCommonStatusHealth(ethernetInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_etherenet_interface_health_state"].desc, prometheus.GaugeValue, ethernetInterfaceHealthStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceLinkStatusValue, ok := parseLinkStatus(ethernetInterfaceLinkStatus); ok {

		ch <- prometheus.MustNewConstMetric(systemMetrics["system_etherenet_interface_link_status"].desc, prometheus.GaugeValue, ethernetInterfaceLinkStatusValue, systemEthernetInterfaceLabelValues...)

	}

	ch <- prometheus.MustNewConstMetric(systemMetrics["system_etherenet_interface_link_enabled"].desc, prometheus.GaugeValue, boolToFloat64(ethernetInterfaceEnabled), systemEthernetInterfaceLabelValues...)

}
