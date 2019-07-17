package collector

import (

	"github.com/prometheus/client_golang/prometheus"
	gofish "github.com/stmcginnis/gofish/school"
)

// Scraper is minimal interface that let's you add new prometheus metrics to mysqld_exporter.
type Scraper interface {
	// Name of the Scraper. Should be unique.
	Name() string
	// Help describes the role of the Scraper.
	// Example: "Collect  node metrics"
	Help() string
	// Scrape collects data from netappClient connection.
	Scrape(redfishClient *gofish.ApiClient, ch chan<- prometheus.Metric) error
}
