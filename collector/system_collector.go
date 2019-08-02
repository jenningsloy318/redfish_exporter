package collector

import (
	"strings"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	gofish "github.com/stmcginnis/gofish/school"
)

// A SystemCollector implements the prometheus.Collector.
type SystemCollector struct {
	redfishClient           *gofish.ApiClient
	metrics                 map[string]systemMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
}

type systemMetric struct {
	desc *prometheus.Desc
}

var (
	SystemLabelNames                  = append(BaseLabelNames, "name", "hostname")
	SystemMemoryLabelNames            = append(BaseLabelNames, "name", "memory", "hostname")
	SystemProcessorLabelNames         = append(BaseLabelNames, "name", "processor", "hostname")
	SystemVolumeLabelNames            = append(BaseLabelNames, "name", "volume", "hostname")
	SystemDriveLabelNames             = append(BaseLabelNames, "name", "drive", "hostname")
	SystemStorageControllerLabelNames = append(BaseLabelNames, "name", "storagecontroller", "hostname")
	SystemPCIeDeviceLabelNames				= append(BaseLabelNames, "name", "pcie_device", "hostname")
)

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(namespace string, redfishClient *gofish.ApiClient) *SystemCollector {
	var (
		subsystem = "system"
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
			"system_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "health_state"),
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
			"system_total_memory_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_memory_health_state"),
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
			"system_total_processor_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_processor_health_state"),
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
			"system_memory_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "memory_state"),
					"system memory state",
					SystemMemoryLabelNames,
					nil,
				),
			},
			"system_memory_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "memory_health_state"),
					"system memory  health state",
					SystemMemoryLabelNames,
					nil,
				),
			},
			"system_memory_capacity": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "memory_capacity"),
					"system memory capacity, MiB",
					SystemMemoryLabelNames,
					nil,
				),
			},

			"system_processor_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "processor_state"),
					"system processor state",
					SystemProcessorLabelNames,
					nil,
				),
			},
			"system_processor_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "processor_health_state"),
					"system processor  health state",
					SystemProcessorLabelNames,
					nil,
				),
			},
			"system_processor_total_threads": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "processor_total_threads"),
					"system processor total threads",
					SystemProcessorLabelNames,
					nil,
				),
			},
			"system_processor_total_cores": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "processor_total_cores"),
					"system processor total cores",
					SystemProcessorLabelNames,
					nil,
				),
			},
			"system_storage_volume_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_volume_state"),
					"system storage volume state",
					SystemVolumeLabelNames,
					nil,
				),
			},
			"system_storage_volume_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_volume_health_state"),
					"system storage volume health state",
					SystemVolumeLabelNames,
					nil,
				),
			},
			"system_storage_volume_capacity": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_volume_capacity"),
					"system storage volume capacity,Bytes",
					SystemVolumeLabelNames,
					nil,
				),
			},
			"system_storage_drive_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_drive_state"),
					"system storage drive state",
					SystemDriveLabelNames,
					nil,
				),
			},
			"system_storage_drive_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_drive_health_state"),
					"system storage volume health state",
					SystemDriveLabelNames,
					nil,
				),
			},
			"system_storage_drive_capacity": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_drive_capacity"),
					"system storage drive capacity,Bytes",
					SystemDriveLabelNames,
					nil,
				),
			},
			"system_storage_controller_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_controller_state"),
					"system storage controller state",
					SystemStorageControllerLabelNames,
					nil,
				),
			},
			"system_storage_controller_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "storage_controller_health_state"),
					"system storage controller health state",
					SystemStorageControllerLabelNames,
					nil,
				),
			},
			"system_pcie_device_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "pcie_device_state"),
					"system pcie device state",
					SystemPCIeDeviceLabelNames,
					nil,
				),
			},
			"system_pcie_device_health_state": {
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "pcie_device_health_state"),
					"system pcie device health state",
					SystemPCIeDeviceLabelNames,
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

func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
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
		log.Infof("Errors Getting Services for chassis metrics : %s", err)
	}

	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		log.Infof("Errors Getting systems from service : %s", err)
	} else {

		for _, system := range systems {
			// overall system metrics

			systemName := system.Name
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

			SystemLabelValues := append(BaseLabelValues, systemName, systemHostName)
			if systemHealthStateValue,ok := parseCommonStatusHealth(systemHealthState); ok{
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].desc, prometheus.GaugeValue, systemHealthStateValue, SystemLabelValues...)
			}
			if systemStateValue,ok := parseCommonStatusState(systemState); ok{
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue, systemStateValue, SystemLabelValues...)
			}
			if systemPowerStateValue,ok := parseSystemPowerState(systemPowerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].desc, prometheus.GaugeValue, systemPowerStateValue, SystemLabelValues...)

			}
			if systemTotalProcessorsStateValue,ok := parseCommonStatusState(systemTotalProcessorsState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].desc, prometheus.GaugeValue, systemTotalProcessorsStateValue, SystemLabelValues...)

			}
			if systemTotalProcessorsHealthStateValue,ok := parseCommonStatusHealth(systemTotalProcessorsHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].desc, prometheus.GaugeValue, systemTotalProcessorsHealthStateValue, SystemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].desc, prometheus.GaugeValue, float64(systemTotalProcessorCount), SystemLabelValues...)

			if systemTotalMemoryStateValue,ok := parseCommonStatusState(systemTotalMemoryState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].desc, prometheus.GaugeValue, systemTotalMemoryStateValue, SystemLabelValues...)

			}
			if systemTotalMemoryHealthStateValue,ok := parseCommonStatusHealth(systemTotalMemoryHealthState); ok{
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].desc, prometheus.GaugeValue, systemTotalMemoryHealthStateValue, SystemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].desc, prometheus.GaugeValue, float64(systemTotalMemoryAmount), SystemLabelValues...)

			// get system OdataID
			//systemOdataID := system.ODataID

			// process memory metrics
			// construct memory Link
			//memoriesLink := fmt.Sprintf("%sMemory/", systemOdataID)

			//if memories, err := redfish.ListReferencedMemorys(s.redfishClient, memoriesLink); err != nil {
			if memories, err := system.Memory(); err != nil {
				log.Infof("Errors Getting memory from computer system : %s", err)
			} else {
				for _, memory := range memories {
					memoryName := memory.DeviceLocator
					//memoryDeviceLocator := memory.DeviceLocator
					memoryCapacityMiB := memory.CapacityMiB
					memoryState := memory.Status.State
					memoryHealthState := memory.Status.Health

					SystemMemoryLabelValues := append(BaseLabelValues, "memory", memoryName, systemHostName)
					if memoryStateValue,ok := parseCommonStatusState(memoryState); ok {
						ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, SystemMemoryLabelValues...)

					}
					if memoryHealthStateValue,ok := parseCommonStatusHealth(memoryHealthState); ok{
						ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_health_state"].desc, prometheus.GaugeValue, memoryHealthStateValue, SystemMemoryLabelValues...)

					}
					ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), SystemMemoryLabelValues...)

				}
			}

			// process processor metrics

			//processorsLink := fmt.Sprintf("%sProcessors/", systemOdataID)

			//if processors, err := redfish.ListReferencedProcessors(s.redfishClient, processorsLink); err != nil {
			if processors, err := system.Processors(); err != nil {
				log.Infof("Errors Getting Processors from system: %s", err)
			} else {

				for _, processor := range processors {

					processorName := processor.Socket
					processorTotalCores := processor.TotalCores
					processorTotalThreads := processor.TotalThreads
					processorState := processor.Status.State
					processorHelathState := processor.Status.Health

					SystemProcessorLabelValues := append(BaseLabelValues, "processor", processorName, systemHostName)

					if processorStateValue,ok := parseCommonStatusState(processorState); ok {
						ch <- prometheus.MustNewConstMetric(s.metrics["system_processor_state"].desc, prometheus.GaugeValue, processorStateValue, SystemProcessorLabelValues...)

					}
					if processorHelathStateValue,ok := parseCommonStatusHealth(processorHelathState); ok {
						ch <- prometheus.MustNewConstMetric(s.metrics["system_processor_health_state"].desc, prometheus.GaugeValue, processorHelathStateValue, SystemProcessorLabelValues...)

					}
					ch <- prometheus.MustNewConstMetric(s.metrics["system_processor_total_threads"].desc, prometheus.GaugeValue, float64(processorTotalThreads), SystemProcessorLabelValues...)
					ch <- prometheus.MustNewConstMetric(s.metrics["system_processor_total_cores"].desc, prometheus.GaugeValue, float64(processorTotalCores), SystemProcessorLabelValues...)

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
						for _, volume := range volumes {
							volumeODataIDslice := strings.Split(volume.ODataID, "/")
							volumeName := volumeODataIDslice[len(volumeODataIDslice)-1]
							volumeCapacityBytes := volume.CapacityBytes
							volumeState := volume.Status.State
							volumeHealthState := volume.Status.Health
							SystemVolumeLabelValues := append(BaseLabelValues, "volume", volumeName, systemHostName)
							if volumeStateValue,ok := parseCommonStatusState(volumeState); ok {
								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_state"].desc, prometheus.GaugeValue, volumeStateValue, SystemVolumeLabelValues...)

							}
							if volumeHealthStateValue,ok := parseCommonStatusHealth(volumeHealthState); ok {
								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_health_state"].desc, prometheus.GaugeValue, volumeHealthStateValue, SystemVolumeLabelValues...)

							}
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_capacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), SystemVolumeLabelValues...)

						}
					}

					if drives, err := storage.Drives(); err != nil {
						log.Infof("Errors Getting volumes  from system storage : %s", err)
					} else {
						for _, drive := range drives {
							driveODataIDslice := strings.Split(drive.ODataID, "/")
							driveName := driveODataIDslice[len(driveODataIDslice)-1]
							driveCapacityBytes := drive.CapacityBytes
							driveState := drive.Status.State
							driveHealthState := drive.Status.Health
							SystemdriveLabelValues := append(BaseLabelValues, "drive", driveName, systemHostName)
							if driveStateValue,ok := parseCommonStatusState(driveState); ok {
								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_state"].desc, prometheus.GaugeValue, driveStateValue, SystemdriveLabelValues...)

							}
							if driveHealthStateValue,ok := parseCommonStatusHealth(driveHealthState); ok {
								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_health_state"].desc, prometheus.GaugeValue, driveHealthStateValue, SystemdriveLabelValues...)

							}
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), SystemdriveLabelValues...)
						}
					}

					if storagecontrollers, err := storage.StorageControllers(); err != nil {
						log.Infof("Errors Getting storagecontrollers from system storage : %s", err)
					} else {
						for _, controller := range storagecontrollers {
							controllerODataIDslice := strings.Split(controller.ODataID, "/")
							controllerName := controllerODataIDslice[len(controllerODataIDslice)-1]
							controllerState := controller.Status.State
							controllerHealthState := controller.Status.Health
							controllerLabelValues := append(BaseLabelValues, "storagecontroller", controllerName, systemHostName)
							if controllerStateValue,ok := parseCommonStatusState(controllerState); ok {
								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_state"].desc, prometheus.GaugeValue, controllerStateValue, controllerLabelValues...)

							}
							if controllerHealthStateValue,ok := parseCommonStatusHealth(controllerHealthState); ok {
								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_health_state"].desc, prometheus.GaugeValue, controllerHealthStateValue, controllerLabelValues...)

							}

						}

					}

				}
			}
			//process pci devices
			//pciDevicesLink := fmt.Sprintf("%sPcidevice/", systemOdataID)
			if pcieDevices, err := system.PCIeDevices(); err != nil {
				log.Infof("Errors Getting PCI-E devices from system: %s", err)
			} else {
				for _, pcieDevice := range pcieDevices {
					pcieDeviceODataIDslice :=strings.Split(pcieDevice.ODataID, "/")
					pcieDeviceID := pcieDeviceODataIDslice[len(pcieDeviceODataIDslice)-1]
					pcieDeviceState := pcieDevice.Status.State
					pcieDeviceHealthState := pcieDevice.Status.Health
					SystemPCIeDeviceLabelValues := append(BaseLabelValues, "PCIeDevice", pcieDeviceID, systemHostName)

					if pcieStateVaule,ok :=parseCommonStatusState(pcieDeviceState);ok {
						ch <- prometheus.MustNewConstMetric(s.metrics["system_pcie_device_state"].desc, prometheus.GaugeValue, pcieStateVaule, SystemPCIeDeviceLabelValues...)

					}
					if pcieHealthStateVaule,ok :=parseCommonStatusHealth(pcieDeviceHealthState);ok {
						ch <- prometheus.MustNewConstMetric(s.metrics["system_pcie_device_health_state"].desc, prometheus.GaugeValue, pcieHealthStateVaule, SystemPCIeDeviceLabelValues...)

					}
				}
			}

			//process networkinterfaces 
		}
	}
	s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))

}
