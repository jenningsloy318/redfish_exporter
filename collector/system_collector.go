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
	systems, err := service.Systems()

	if err != nil {
		log.Fatalf("Errors Getting systems from service : %s", err)
	}

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

		ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(systemHealthState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue, parseCommonStatusState(systemState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].desc, prometheus.GaugeValue, parseSystemPowerState(systemPowerState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].desc, prometheus.GaugeValue, parseCommonStatusState(systemTotalProcessorsState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(systemTotalProcessorsHealthState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].desc, prometheus.GaugeValue, float64(systemTotalProcessorCount), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].desc, prometheus.GaugeValue, parseCommonStatusState(systemTotalMemoryState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(systemTotalMemoryHealthState), SystemLabelValues...)
		ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].desc, prometheus.GaugeValue, float64(systemTotalMemoryAmount), SystemLabelValues...)

		// get system OdataID
		systemOdataID := system.ODataID

		// process memory metrics
		// construct memory Link
		memoriesLink := fmt.Sprintf("%sMemory/", systemOdataID)

		memories, err := redfish.ListReferencedMemorys(s.redfishClient, memoriesLink)
		if err != nil {
			log.Fatalf("Errors Getting memory from computer system : %s", err)
		}

		for _, memory := range memories {
			memoryName := memory.DeviceLocator
			//memoryDeviceLocator := memory.DeviceLocator
			memoryCapacityMiB := memory.CapacityMiB
			memoryState := memory.Status.State
			memoryHealthState := memory.Status.Health

			SystemMemoryLabelValues := append(BaseLabelValues, "memory", memoryName, systemHostName)

			ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_state"].desc, prometheus.GaugeValue, parseCommonStatusState(memoryState), SystemMemoryLabelValues...)
			ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(memoryHealthState), SystemMemoryLabelValues...)
			ch <- prometheus.MustNewConstMetric(s.metrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), SystemMemoryLabelValues...)

		}

		// process processor metrics

		//	processorsLink :=  fmt.Sprintf("%sMemory/",systemOdataID)

		//process storage
		storagesLink := fmt.Sprintf("%sStorage/", systemOdataID)

		storages, err := redfish.ListReferencedStorages(s.redfishClient, storagesLink)
		if err != nil {
			log.Fatalf("Errors Getting storages from system: %s", err)
		}

		for _, storage := range storages {

			volumes, err := storage.Volumes()
			if err != nil {
				log.Fatalf("Errors Getting volumes  from system storage : %s", err)
			}

			for _, volume := range volumes {
				volumeODataIDslice := strings.Split(volume.ODataID, "/")
				volumeName := volumeODataIDslice[len(volumeODataIDslice)-1]
				volumeCapacityBytes := volume.CapacityBytes
				volumeState := volume.Status.State
				volumeHealthState := volume.Status.Health
				SystemVolumeLabelValues := append(BaseLabelValues, "volume", volumeName, systemHostName)

				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_state"].desc, prometheus.GaugeValue, parseCommonStatusState(volumeState), SystemVolumeLabelValues...)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(volumeHealthState), SystemVolumeLabelValues...)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_volume_capacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), SystemVolumeLabelValues...)

			}

			drives, err := storage.Drives()
			if err != nil {
				log.Fatalf("Errors Getting volumes  from system storage : %s", err)
			}

			for _, drive := range drives {
				driveODataIDslice := strings.Split(drive.ODataID, "/")
				driveName := driveODataIDslice[len(driveODataIDslice)-1]
				driveCapacityBytes := drive.CapacityBytes
				driveState := drive.Status.State
				driveHealthState := drive.Status.Health
				SystemdriveLabelValues := append(BaseLabelValues, "drive", driveName, systemHostName)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_state"].desc, prometheus.GaugeValue, parseCommonStatusState(driveState), SystemdriveLabelValues...)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(driveHealthState), SystemdriveLabelValues...)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), SystemdriveLabelValues...)
			}

			storagecontrollers, err := storage.StorageControllers()

			if err != nil {
				log.Fatalf("Errors Getting storagecontrollers from system storage : %s", err)
			}

			for _, controller := range storagecontrollers {
				controllerODataIDslice := strings.Split(controller.ODataID, "/")
				controllerName := controllerODataIDslice[len(controllerODataIDslice)-1]
				controllerState := controller.Status.State
				controllerHealthState := controller.Status.Health
				controllerLabelValues := append(BaseLabelValues, "storagecontroller", controllerName, systemHostName)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_state"].desc, prometheus.GaugeValue, parseCommonStatusState(controllerState), controllerLabelValues...)
				ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_health_state"].desc, prometheus.GaugeValue, parseCommonStatusHealth(controllerHealthState), controllerLabelValues...)

			}

		}

	}

	s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))

}
