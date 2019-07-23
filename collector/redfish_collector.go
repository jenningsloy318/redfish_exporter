package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	gofish "github.com/stmcginnis/gofish/school"
	"fmt"
	"net/http"
	"crypto/tls"
	"time"
	"strings"
	

)

// Metric name parts.
const (
	// Exporter namespace.
	namespace = "redfish"
	// Subsystem(s).
	exporter = "exporter"
	// Math constant for picoseconds to seconds.
	picoSeconds = 1e12
)


// Metric descriptors.
var (
	BaseLabelNames = []string{"host"}
	BaseLabelValues = make([]string, 1,1)	
	totalScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		BaseLabelNames, nil,
	)

)

// Exporter collects redfish metrics. It implements prometheus.Collector.
type RedfishCollector struct {
	redfishClient *gofish.ApiClient
	collectors    map[string]prometheus.Collector
	redfishUp     prometheus.Gauge
	redfishUpValue  float64
}


func NewRedfishCollector(host string, username string, password string ) *RedfishCollector {	
	BaseLabelValues[0]=host
	redfishClient, redfishUpValue := newRedfishClient(host,username,password)
//	memoryCollector :=NewMemoryCollector(namespace,redfishClient)
	chassisCollector := NewChassisCollector(namespace,redfishClient)
	return &RedfishCollector{
		redfishClient: redfishClient,
		collectors:    map[string]prometheus.Collector{"chassis": chassisCollector},
		redfishUp:	prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "",
				Name:      "up",
				Help:      "redfish up",
			},
		),
		redfishUpValue: redfishUpValue,
	}
}

// Describe implements prometheus.Collector.
func (r *RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range r.collectors {
		collector.Describe(ch)
	}


}

// Collect implements prometheus.Collector.
func (r *RedfishCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeTime := time.Now()
	 if r.redfishUpValue == float64(1) {
		r.redfishUp.Set(r.redfishUpValue)
		ch <- r.redfishUp
		for _, collector := range r.collectors {
			collector.Collect(ch)
		}
	 }else {
		r.redfishUp.Set(r.redfishUpValue)
		ch <- r.redfishUp
	 }
	 ch <- prometheus.MustNewConstMetric(totalScrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), BaseLabelValues...)
}






func newRedfishClient(host string, username string, password string) (*gofish.ApiClient, float64) {

	url := fmt.Sprintf("https://%s", host)

	// skip ssl verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := &http.Client{Transport: tr}

	log.Infof(url)
	// Create a new instance of gofish client
	 redfishClient,err := gofish.APIClient(url,httpClient)
	 if  err != nil {
		log.Fatalf("Errors occours when creating redfish client: %s",err)
		return redfishClient, float64(0)
	}

	 service, err := gofish.ServiceRoot(redfishClient)
	 if err != nil {
		log.Fatalf("Errors occours when Getting Services: %s",err)
		return redfishClient, float64(0)
	}

	// Generates a authenticated session
	auth, err := service.CreateSession(username, password)
	if err != nil {
		log.Fatalf("Errors occours when creating sessions: %s",err)
		return redfishClient, float64(0)
	}

	// Assign the token back to our gofish client
	redfishClient.Token = auth.Token	
	 
	return redfishClient,float64(1)
}



func pareCommonStatusHealth(status string) float64{
if strings.EqualFold(status,"OK"){
	return float64(0)
} else if strings.EqualFold(status,"Warning") {
	return float64(1)
}else if strings.EqualFold(status,"Critical") {
	return float64(2)
}
return float64(0)
}


func pareCommonStatusState(status string) float64{
	if strings.EqualFold(status,"Enabled"){
		return float64(0)
	} else if strings.EqualFold(status,"Disabled") {
		return float64(1)
	}else if strings.EqualFold(status,"StandbyOffinline") {
		return float64(2)
	}	else if strings.EqualFold(status,"StandbySpare") {
		return float64(3)
	}else if strings.EqualFold(status,"InTest") {
		return float64(4)
	}else if strings.EqualFold(status,"Starting") {
		return float64(5)
	}else if strings.EqualFold(status,"Absent") {
		return float64(6)
	}else if strings.EqualFold(status,"UnavailableOffline") {
		return float64(7)
	}else if strings.EqualFold(status,"Deferring") {
		return float64(8)
	}else if strings.EqualFold(status,"Quiesced") {
		return float64(9)
	}else if strings.EqualFold(status,"Updating") {
		return float64(10)
	}
	return float64(0)
	}