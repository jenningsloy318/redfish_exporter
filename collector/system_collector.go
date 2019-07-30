package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	gofish "github.com/stmcginnis/gofish/school"
	redfish "github.com/stmcginnis/gofish/school/redfish"
	"strings"
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
	SystemVolumeLabelNames            = append(BaseLabelNames, "name", "volume", "hostname")
	SystemDriveLabelNames             = append(BaseLabelNames, "name", "drive", "hostname")
	SystemStorageControllerLabelNames = append(BaseLabelNames, "name", "storagecontroller", "hostname")
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
		log.Fatalf("Errors Getting Services for chassis metrics : %s", err)
	}

	// get a list of systems from service
	 if systems, err := service.Systems(); err != nil {
		log.Fatalf("Errors Getting systems from service : %s", err)
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
			if systemHealthStateValue :=parseCommonStatusHealth(systemHealthState);systemHealthStateValue !=float64(0) {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].desc, prometheus.GaugeValue,systemHealthStateValue, SystemLabelValues...)
			}
			if systemStateValue:=parseCommonStatusState(systemState);systemStateValue!=float64(0) {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue,systemStateValue, SystemLabelValues...)
			}
			if systemPowerStateValue := parseSystemPowerState(systemPowerState);systemPowerStateValue != float64(0) {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].desc, prometheus.GaugeValue,systemPowerStateValue , SystemLabelValues...)

			}
			if systemTotalProcessorsStateValue :=parseCommonStatusState(systemTotalProcessorsState);systemTotalProcessorsStateValue !=float64(0) {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].desc, prometheus.GaugeValue,systemTotalProcessorsStateValue , SystemLabelValues...)

			}
			if systemTotalProcessorsHealthStateValue := parseCommonStatusHealth(systemTotalProcessorsHealthState);systemTotalProcessorsHealthStateValue !=float64(0){
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].desc, prometheus.GaugeValue,systemTotalProcessorsHealthStateValue , SystemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].desc, prometheus.GaugeValue, float64(systemTotalProcessorCount), SystemLabelValues...)

			if systemTotalMemoryStateValue:=parseCommonStatusState(systemTotalMemoryState);systemTotalMemoryStateValue !=float64(0) {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].desc, prometheus.GaugeValue, systemTotalMemoryStateValue, SystemLabelValues...)

			}
			if systemTotalMemoryHealthStateValue:=parseCommonStatusHealth(systemTotalMemoryHealthState);systemTotalMemoryHealthStateValue !=float64(0) {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].desc, prometheus.GaugeValue,systemTotalMemoryHealthStateValue , SystemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].desc, prometheus.GaugeValue, float64(systemTotalMemoryAmount), SystemLabelValues...)

			// get system OdataID
			systemOdataID := system.ODataID

			// process memory metrics
			// construct memory Link
			memoriesLink := fmt.Sprintf("%sMemory/", systemOdataID)

			if memories, err := redfish.ListReferencedMemorys(s.redfishClient, memoriesLink); err != nil {
				log.Infof("Errors Getting memory from computer system : %s", err)
			}else {
				for _, memory := range memories {
					memoryName := memory.DeviceLocator
					//memoryDeviceLocator := memory.DeviceLocator
					memoryCapacityMiB := memory.CapacityMiB
					memoryState := memory.Status.State
					memoryHealthState := memory.Status.Health

					SystemMemoryLabelValues := append(BaseLabelValues, "memory", memoryName, systemHostName)
					if  memoryStateValue := parseCommonStatusState(memoryState);memoryStateValue !=float64(0){
						ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, SystemMemoryLabelValues...)

					}
					if memoryHealthStateValue:=parseCommonStatusHealth(memoryHealthState);memoryHealthStateValue !=float64(0){
						ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_health_state"].desc, prometheus.GaugeValue,memoryHealthStateValue , SystemMemoryLabelValues...)

					}
					ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), SystemMemoryLabelValues...)

				}
		}

			// process processor metrics

			//	processorsLink :=  fmt.Sprintf("%sMemory/",systemOdataID)

			//process storage
			storagesLink := fmt.Sprintf("%sStorage/", systemOdataID)

			if storages, err := redfish.ListReferencedStorages(s.redfishClient, storagesLink); err != nil {
				log.Infof("Errors Getting storages from system: %s", err)
			} else {
			for _, storage := range storages {

				if volumes, err := storage.Volumes();err != nil {
					log.Infof("Errors Getting volumes  from system storage : %s", err)
				} else {
					for _, volume := range volumes {
						volumeODataIDslice := strings.Split(volume.ODataID, "/")
						volumeName := volumeODataIDslice[len(volumeODataIDslice)-1]
						volumeCapacityBytes := volume.CapacityBytes
						volumeState := volume.Status.State
						volumeHealthState := volume.Status.Health
						SystemVolumeLabelValues := append(BaseLabelValues, "volume", volumeName, systemHostName)
						if volumeStateValue :=parseCommonStatusState(volumeState);volumeStateValue !=float64(0){
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_state"].desc, prometheus.GaugeValue, volumeStateValue, SystemVolumeLabelValues...)

						}
						if volumeHealthStateValue :=parseCommonStatusHealth(volumeHealthState);volumeHealthStateValue !=float64(0) {
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_health_state"].desc, prometheus.GaugeValue,volumeHealthStateValue , SystemVolumeLabelValues...)

						}
						ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_capacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), SystemVolumeLabelValues...)

					}
				}

				if drives, err := storage.Drives(); err != nil {
					log.Infof("Errors Getting volumes  from system storage : %s", err)
				}else {
					for _, drive := range drives {
						driveODataIDslice := strings.Split(drive.ODataID, "/")
						driveName := driveODataIDslice[len(driveODataIDslice)-1]
						driveCapacityBytes := drive.CapacityBytes
						driveState := drive.Status.State
						driveHealthState := drive.Status.Health
						SystemdriveLabelValues := append(BaseLabelValues, "drive", driveName, systemHostName)
						if driveStateValue :=parseCommonStatusState(driveState);driveStateValue !=float64(0) {
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_state"].desc, prometheus.GaugeValue, driveStateValue, SystemdriveLabelValues...)

						}
						if driveHealthStateValue:=parseCommonStatusHealth(driveHealthState) ;driveHealthStateValue !=float64(0) {
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_health_state"].desc, prometheus.GaugeValue, driveHealthStateValue, SystemdriveLabelValues...)

						}
						ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), SystemdriveLabelValues...)
					}
				}
				
				if storagecontrollers, err := storage.StorageControllers(); err != nil {
					log.Fatalf("Errors Getting storagecontrollers from system storage : %s", err)
				} else {
					for _, controller := range storagecontrollers {
						controllerODataIDslice := strings.Split(controller.ODataID, "/")
						controllerName := controllerODataIDslice[len(controllerODataIDslice)-1]
						controllerState := controller.Status.State
						controllerHealthState := controller.Status.Health
						controllerLabelValues := append(BaseLabelValues, "storagecontroller", controllerName, systemHostName)
						if controllerStateValue:=parseCommonStatusState(controllerState);controllerStateValue !=float64(0) {
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_state"].desc, prometheus.GaugeValue,controllerStateValue , controllerLabelValues...)

						}
						if controllerHealthStateValue:=parseCommonStatusHealth(controllerHealthState);controllerHealthStateValue !=float64(0) {
							ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_health_state"].desc, prometheus.GaugeValue, controllerHealthStateValue, controllerLabelValues...)

						}

					}

			}

			}
			}

		}
	}
	s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))

}
