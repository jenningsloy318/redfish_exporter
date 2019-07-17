package collector

import (
	"sync"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	gofish "github.com/stmcginnis/gofish/school"
	"fmt"
	"net/http"
	"crypto/tls"
	

)

// Metric name parts.
const (
	// Subsystem(s).
	exporter = "exporter"
)

// Metric descriptors.
var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil,
	)
	redfishUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"redfish host up.",
		[]string{"host"}, nil,
	)
	BaseLabelNames = []string{"host"}

	BaseLabelValues = make([]string, 1,1)
)

// Exporter collects redfish metrics. It implements prometheus.Collector.
type Exporter struct {
	redfishClientHost string 
	redfishClientUsername string 
	redfishClientPassword string
	error        prometheus.Gauge
	scrapers     []Scraper
	totalScrapes prometheus.Counter
	scrapeErrors *prometheus.CounterVec
}

var scrapers = []Scraper{
}

func New(host string, username string, password string ) *Exporter {
	return &Exporter{
		redfishClientHost: host,
		redfishClientUsername: username,
		redfishClientPassword: password,
		scrapers:     scrapers,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times redfish was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a redfish host.",
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from redfish resulted in an error (1 for error, 0 for success).",
		}),
	}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// We cannot know in advance what metrics the exporter will generate
	// from redfish. So we use the poor man's describe method: Run a collect
	// and send the descriptors of all the collected metrics. The problem
	// here is that we need to connect to the redfish . If it is currently
	// unavailable, the descriptors will be incomplete. Since this is a
	// stand-alone exporter and not used as a library within other code
	// implementing additional metrics, the worst that can happen is that we
	// don't detect inconsistent metrics created by this exporter
	// itself. Also, a change in the monitored redfish instance may change the
	// exported metrics during the runtime of the exporter.

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.scrape(ch)
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	scrapeTime := time.Now()
	
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), "connection")

	redfishClient, ok := newRedfishClient(e.redfishClientHost,e.redfishClientUsername,e.redfishClientPassword)
	if ok{
		ch <- prometheus.MustNewConstMetric(redfishUpDesc, prometheus.GaugeValue, 1, e.redfishClientHost)
		BaseLabelValues[0]=e.redfishClientHost
	
	} else {
		ch <- prometheus.MustNewConstMetric(redfishUpDesc, prometheus.GaugeValue, 0, e.redfishClientHost)
		BaseLabelValues[0]=e.redfishClientHost

	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for _, scraper := range e.scrapers {
		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()
			label := "collect." + scraper.Name()
			scrapeTime := time.Now()
			if err := scraper.Scrape(redfishClient, ch); err != nil {
				log.Errorln("Error scraping for "+label+":", err)
				e.scrapeErrors.WithLabelValues(label).Inc()
				e.error.Set(1)
			}

			ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), label)
		}(scraper)
	}
}





func newRedfishClient(host string, username string, password string) (*gofish.ApiClient, bool) {

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
		return redfishClient, false
	}

	 service, err := gofish.ServiceRoot(redfishClient)
	 if err != nil {
		log.Fatalf("Errors occours when Getting Services: %s",err)
		return redfishClient, false
	}

	// Generates a authenticated session
	auth, err := service.CreateSession(username, password)
	if err != nil {
		log.Fatalf("Errors occours when creating sessions: %s",err)
		return redfishClient, false
	}

	// Assign the token back to our gofish client
	redfishClient.Token = auth.Token	
	 
	return redfishClient,true
}
