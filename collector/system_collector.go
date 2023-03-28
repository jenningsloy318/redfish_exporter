package collector

import (
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// SystemSubsystem is the system subsystem
var (
	SystemSubsystem                   = "system"
	SystemLabelNames                  = []string{"hostname", "resource", "system_id"}
	SystemMemoryLabelNames            = []string{"hostname", "resource", "memory", "memory_id"}
	SystemProcessorLabelNames         = []string{"hostname", "resource", "processor", "processor_id"}
	SystemVolumeLabelNames            = []string{"hostname", "resource", "volume", "volume_id"}
	SystemDeviceLabelNames            = []string{"hostname", "resource", "device"}
	SystemDriveLabelNames             = []string{"hostname", "resource", "drive", "drive_id"}
	SystemStorageControllerLabelNames = []string{"hostname", "resource", "storage_controller", "storage_controller_id"}
	SystemPCIeDeviceLabelNames        = []string{"hostname", "resource", "pcie_device", "pcie_device_id", "pcie_device_partnumber", "pcie_device_type", "pcie_serial_number"}
	SystemNetworkInterfaceLabelNames  = []string{"hostname", "resource", "network_interface", "network_interface_id"}
	SystemEthernetInterfaceLabelNames = []string{"hostname", "resource", "ethernet_interface", "ethernet_interface_id", "ethernet_interface_speed"}
	SystemPCIeFunctionLabelNames      = []string{"hostname", "resource", "pcie_function_name", "pcie_function_id", "pci_function_deviceclass", "pci_function_type"}

	SystemLogServiceLabelNames = []string{"system_id", "log_service", "log_service_id", "log_service_enabled", "log_service_overwrite_policy"}
	SystemLogEntryLabelNames   = []string{"system_id", "log_service", "log_service_id", "log_entry", "log_entry_id", "log_entry_code", "log_entry_type", "log_entry_message_id", "log_entry_sensor_number", "log_entry_sensor_type"}

	systemMetrics = createSystemMetricMap()
)

// SystemCollector implements the prometheus.Collector.
type SystemCollector struct {
	redfishClient *gofish.APIClient
	metrics       map[string]Metric
	prometheus.Collector
	collectorScrapeStatus *prometheus.GaugeVec
	Log                   *log.Entry
}

func createSystemMetricMap() map[string]Metric {
	systemMetrics := make(map[string]Metric)

	addToMetricMap(systemMetrics, SystemSubsystem, "state", fmt.Sprintf("system state,%s", CommonStateHelp), SystemLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "health_state", fmt.Sprintf("system health,%s", CommonHealthHelp), SystemLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "power_state", "system power state", SystemLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "total_memory_state", fmt.Sprintf("system overall memory state,%s", CommonStateHelp), SystemLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "total_memory_health_state", fmt.Sprintf("system overall memory health,%s", CommonHealthHelp), SystemLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "total_memory_size", "system total memory size, GiB", SystemLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "total_processor_state", fmt.Sprintf("system overall processor state,%s", CommonStateHelp), SystemLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "total_processor_health_state", fmt.Sprintf("system overall processor health,%s", CommonHealthHelp), SystemLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "total_processor_count", "system total processor count", SystemLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "memory_state", fmt.Sprintf("system memory state,%s", CommonStateHelp), SystemMemoryLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "memory_health_state", fmt.Sprintf("system memory health state,%s", CommonHealthHelp), SystemMemoryLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "memory_capacity", "system memory capacity, MiB", SystemMemoryLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "processor_state", fmt.Sprintf("system processor state,%s", CommonStateHelp), SystemProcessorLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "processor_health_state", fmt.Sprintf("system processor health state,%s", CommonHealthHelp), SystemProcessorLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "processor_total_threads", "system processor total threads", SystemProcessorLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "processor_total_cores", "system processor total cores", SystemProcessorLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "simple_storage_device_state", fmt.Sprintf("system simple storage device state,%s", CommonStateHelp), SystemDeviceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "simple_storage_device_health_state", fmt.Sprintf("system simple storage device health state,%s", CommonHealthHelp), SystemDeviceLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "storage_volume_state", fmt.Sprintf("system storage volume state,%s", CommonStateHelp), SystemVolumeLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "storage_volume_health_state", fmt.Sprintf("system storage volume health state,%s", CommonHealthHelp), SystemVolumeLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "storage_volume_capacity", "system storage volume capacity, Bytes", SystemVolumeLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "storage_drive_state", fmt.Sprintf("system storage drive state,%s", CommonStateHelp), SystemDriveLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "storage_drive_health_state", fmt.Sprintf("system storage drive health state,%s", CommonHealthHelp), SystemDriveLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "storage_drive_capacity", "system storage drive capacity, Bytes", SystemDriveLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "storage_controller_state", fmt.Sprintf("system storage controller state,%s", CommonStateHelp), SystemStorageControllerLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "storage_controller_health_state", fmt.Sprintf("system storage controller health state,%s", CommonHealthHelp), SystemStorageControllerLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "pcie_device_state", fmt.Sprintf("system pcie device state,%s", CommonStateHelp), SystemPCIeDeviceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "pcie_device_health_state", fmt.Sprintf("system pcie device health state,%s", CommonHealthHelp), SystemPCIeDeviceLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "pcie_function_state", fmt.Sprintf("system pcie function state,%s", CommonStateHelp), SystemPCIeFunctionLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "pcie_function_health_state", fmt.Sprintf("system pcie device function state,%s", CommonHealthHelp), SystemPCIeFunctionLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "network_interface_state", fmt.Sprintf("system network interface state,%s", CommonStateHelp), SystemNetworkInterfaceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "network_interface_health_state", fmt.Sprintf("system network interface health state,%s", CommonHealthHelp), SystemNetworkInterfaceLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_state", fmt.Sprintf("system ethernet interface state,%s", CommonStateHelp), SystemEthernetInterfaceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_health_state", fmt.Sprintf("system ethernet interface health state,%s", CommonHealthHelp), SystemEthernetInterfaceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_link_status", fmt.Sprintf("system ethernet interface link status,%s", CommonLinkHelp), SystemEthernetInterfaceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "ethernet_interface_link_enabled", "system ethernet interface if the link is enabled", SystemEthernetInterfaceLabelNames)

	addToMetricMap(systemMetrics, SystemSubsystem, "log_service_state", fmt.Sprintf("system log service state,%s", CommonStateHelp), SystemLogServiceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "log_service_health_state", fmt.Sprintf("system log service health state,%s", CommonHealthHelp), SystemLogServiceLabelNames)
	addToMetricMap(systemMetrics, SystemSubsystem, "log_entry_severity_state", fmt.Sprintf("system log entry severity state,%s", CommonSeverityHelp), SystemLogEntryLabelNames)

	return systemMetrics
}

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(redfishClient *gofish.APIClient, logger *log.Entry) *SystemCollector {
	return &SystemCollector{
		redfishClient: redfishClient,
		metrics:       systemMetrics,
		Log: logger.WithFields(log.Fields{
			"collector": "SystemCollector",
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

// Describe implements prometheus.Collector.
func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.desc
	}
	s.collectorScrapeStatus.Describe(ch)
}

// Collect implements prometheus.Collector.
func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := s.Log
	//get service
	service := s.redfishClient.Service

	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		collectorLogContext.WithField("operation", "service.Systems()").WithError(err).Error("error getting systems from service")
	} else {
		for _, system := range systems {
			systemLogContext := collectorLogContext.WithField("System", system.ID)
			systemLogContext.Info("collector scrape started")
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
			if systemPowerStateValue, ok := parseCommonPowerState(systemPowerState); ok {
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
			memories, err := system.Memory()
			if err != nil {
				systemLogContext.WithField("operation", "system.Memory()").WithError(err).Error("error getting memory data from system")
			} else if memories == nil {
				systemLogContext.WithField("operation", "system.Memory()").Info("no memory data found")
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
			processors, err := system.Processors()
			if err != nil {
				systemLogContext.WithField("operation", "system.Processors()").WithError(err).Error("error getting processor data from system")
			} else if processors == nil {
				systemLogContext.WithField("operation", "system.Processors()").Info("no processor data found")
			} else {
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(processors))

				for _, processor := range processors {
					go parseProcessor(ch, systemHostName, processor, wg2)

				}
			}

			//process storage
			//storagesLink := fmt.Sprintf("%sStorage/", systemOdataID)

			//if storages, err := redfish.ListReferencedStorages(s.redfishClient, storagesLink); err != nil {
			storages, err := system.Storage()
			if err != nil {
				systemLogContext.WithField("operation", "system.Storage()").WithError(err).Error("error getting storage data from system")
			} else if storages == nil {
				systemLogContext.WithField("operation", "system.Storage()").Info("no storage data found")
			} else {
				for _, storage := range storages {
					if volumes, err := storage.Volumes(); err != nil {
						systemLogContext.WithField("operation", "system.Volumes()").WithError(err).Error("error getting storage data from system")
					} else {
						wg3 := &sync.WaitGroup{}
						wg3.Add(len(volumes))

						for _, volume := range volumes {
							go parseVolume(ch, systemHostName, volume, wg3)
						}
					}

					drives, err := storage.Drives()
					if err != nil {
						systemLogContext.WithField("operation", "system.Drives()").WithError(err).Error("error getting drive data from system")
					} else if drives == nil {
						systemLogContext.WithFields(log.Fields{"operation": "system.Drives()", "storage": storage.ID}).Info("no drive data found")
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
			pcieDevices, err := system.PCIeDevices()
			if err != nil {
				systemLogContext.WithField("operation", "system.PCIeDevices()").WithError(err).Error("error getting PCI-E device data from system")
			} else if pcieDevices == nil {
				systemLogContext.WithField("operation", "system.PCIeDevices()").Info("no PCI-E device data found")
			} else {
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(pcieDevices))
				for _, pcieDevice := range pcieDevices {
					go parsePcieDevice(ch, systemHostName, pcieDevice, wg5)
				}
			}

			//process networkinterfaces
			networkInterfaces, err := system.NetworkInterfaces()
			if err != nil {
				systemLogContext.WithField("operation", "system.NetworkInterfaces()").WithError(err).Error("error getting network interface data from system")
			} else if networkInterfaces == nil {
				systemLogContext.WithField("operation", "system.NetworkInterfaces()").Info("no network interface data found")
			} else {
				wg6 := &sync.WaitGroup{}
				wg6.Add(len(networkInterfaces))
				for _, networkInterface := range networkInterfaces {
					go parseNetworkInterface(ch, systemHostName, networkInterface, wg6)
				}
			}

			//process ethernetinterfaces
			ethernetInterfaces, err := system.EthernetInterfaces()
			if err != nil {
				systemLogContext.WithField("operation", "system.EthernetInterfaces()").WithError(err).Error("error getting ethernet interface data from system")
			} else if ethernetInterfaces == nil {
				systemLogContext.WithField("operation", "system.PCIeDevices()").Info("no ethernet interface data found")
			} else {
				wg7 := &sync.WaitGroup{}
				wg7.Add(len(ethernetInterfaces))
				for _, ethernetInterface := range ethernetInterfaces {
					go parseEthernetInterface(ch, systemHostName, ethernetInterface, wg7)
				}
			}

			//process simple storage
			simpleStorages, err := system.SimpleStorages()
			if err != nil {
				systemLogContext.WithField("operation", "system.SimpleStorages()").WithError(err).Error("error getting simple storage data from system")
			} else if simpleStorages == nil {
				systemLogContext.WithField("operation", "system.SimpleStorages()").Info("no simple storage data found")
			} else {
				for _, simpleStorage := range simpleStorages {
					devices := simpleStorage.Devices
					wg8 := &sync.WaitGroup{}
					wg8.Add(len(devices))
					for _, device := range devices {
						go parseDevice(ch, systemHostName, device, wg8)
					}
				}
			}
			//process pci functions
			pcieFunctions, err := system.PCIeFunctions()
			if err != nil {
				systemLogContext.WithField("operation", "system.PCIeFunctions()").WithError(err).Error("error getting PCI-E device function data from system")
			} else if pcieFunctions == nil {
				systemLogContext.WithField("operation", "system.PCIeFunctions()").Info("no PCI-E device function data found")
			} else {
				wg9 := &sync.WaitGroup{}
				wg9.Add(len(pcieFunctions))
				for _, pcieFunction := range pcieFunctions {
					go parsePcieFunction(ch, systemHostName, pcieFunction, wg9)
				}
			}

			// process log services
			logServices, err := system.LogServices()
			if err != nil {
				systemLogContext.WithField("operation", "system.LogServices()").WithError(err).Error("error getting log services from system")
			} else if logServices == nil {
				systemLogContext.WithField("operation", "system.LogServices()").Info("no log services found")
			} else {
				wg10 := &sync.WaitGroup{}
				wg10.Add(len(logServices))

				for _, logService := range logServices {
					if err = parseLogService(ch, systemMetrics, SystemSubsystem, SystemID, logService, wg10); err != nil {
						systemLogContext.WithField("operation", "system.LogServices()").WithError(err).Error("error getting log entries from log service")
					}
				}
			}

			systemLogContext.Info("collector scrape completed")
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	}
}

func parseMemory(ch chan<- prometheus.Metric, systemHostName string, memory *redfish.Memory, wg *sync.WaitGroup) {
	defer wg.Done()
	memoryName := memory.Name
	memoryID := memory.ID
	//memoryDeviceLocator := memory.DeviceLocator
	memoryCapacityMiB := memory.CapacityMiB
	memoryState := memory.Status.State
	memoryHealthState := memory.Status.Health

	systemMemoryLabelValues := []string{systemHostName, "memory", memoryName, memoryID}
	if memoryStateValue, ok := parseCommonStatusState(memoryState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, systemMemoryLabelValues...)
	}
	if memoryHealthStateValue, ok := parseCommonStatusHealth(memoryHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_health_state"].desc, prometheus.GaugeValue, memoryHealthStateValue, systemMemoryLabelValues...)
	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), systemMemoryLabelValues...)

}

func parseProcessor(ch chan<- prometheus.Metric, systemHostName string, processor *redfish.Processor, wg *sync.WaitGroup) {
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
func parseDevice(ch chan<- prometheus.Metric, systemHostName string, device redfish.Device, wg *sync.WaitGroup) {
	defer wg.Done()
	deviceName := device.Name
	deviceState := device.Status.State
	deviceHealthState := device.Status.Health
	systemDeviceLabelValues := []string{systemHostName, "device", deviceName}
	if deviceStateValue, ok := parseCommonStatusState(deviceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_simple_storage_device_state"].desc, prometheus.GaugeValue, deviceStateValue, systemDeviceLabelValues...)
	}
	if deviceHealthStateValue, ok := parseCommonStatusHealth(deviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_simple_storage_device_health_state"].desc, prometheus.GaugeValue, deviceHealthStateValue, systemDeviceLabelValues...)
	}
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
	pcieDevicePartNumber := pcieDevice.PartNumber
	pcieDeviceType := fmt.Sprintf("%v,", pcieDevice.DeviceType)
	pcieSerialNumber := pcieDevice.SerialNumber
	systemPCIeDeviceLabelValues := []string{systemHostName, "pcie_device", pcieDeviceName, pcieDeviceID, pcieDevicePartNumber, pcieDeviceType, pcieSerialNumber}

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
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_state"].desc, prometheus.GaugeValue, ethernetInterfaceStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceHealthStateValue, ok := parseCommonStatusHealth(ethernetInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_health_state"].desc, prometheus.GaugeValue, ethernetInterfaceHealthStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceLinkStatusValue, ok := parseLinkStatus(ethernetInterfaceLinkStatus); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_link_status"].desc, prometheus.GaugeValue, ethernetInterfaceLinkStatusValue, systemEthernetInterfaceLabelValues...)
	}

	ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_link_enabled"].desc, prometheus.GaugeValue, boolToFloat64(ethernetInterfaceEnabled), systemEthernetInterfaceLabelValues...)
}

func parsePcieFunction(ch chan<- prometheus.Metric, systemHostName string, pcieFunction *redfish.PCIeFunction, wg *sync.WaitGroup) {
	defer wg.Done()
	pcieFunctionName := pcieFunction.Name
	pcieFunctionID := fmt.Sprintf("%v", pcieFunction.ID)
	pciFunctionDeviceclass := fmt.Sprintf("%v", pcieFunction.DeviceClass)
	pciFunctionType := fmt.Sprintf("%v", pcieFunction.FunctionType)
	pciFunctionState := pcieFunction.Status.State
	pciFunctionHealthState := pcieFunction.Status.Health

	systemPCIeFunctionLabelLabelValues := []string{systemHostName, "pcie_function", pcieFunctionName, pcieFunctionID, pciFunctionDeviceclass, pciFunctionType}

	if pciFunctionStateValue, ok := parseCommonStatusState(pciFunctionState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_state"].desc, prometheus.GaugeValue, pciFunctionStateValue, systemPCIeFunctionLabelLabelValues...)
	}

	if pciFunctionHealthStateValue, ok := parseCommonStatusHealth(pciFunctionHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_health_state"].desc, prometheus.GaugeValue, pciFunctionHealthStateValue, systemPCIeFunctionLabelLabelValues...)
	}
}
