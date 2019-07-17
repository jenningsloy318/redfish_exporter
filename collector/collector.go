package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"bytes"
	"strconv"
)

const (
	// Exporter namespace.
	namespace = "redfish"
	// Math constant for picoseconds to seconds.
	picoSeconds = 1e12
)


func newDesc(subsystem, name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, nil,
	)
}



func parseStatus(data string) (float64, bool) {

	// vserver state
	if bytes.Equal([]byte(data), []byte("running")) {
		return 1, true
	}
	if bytes.Equal([]byte(data), []byte("stopped")) {
		return 0, true
	}
	if bytes.Equal([]byte(data), []byte("starting")) {
		return 2, true
	}
	if bytes.Equal([]byte(data), []byte("stopping")) {
		return 3, true
	}
	if bytes.Equal([]byte(data), []byte("initializing")) {
		return 4, true
	}
	if bytes.Equal([]byte(data), []byte("deleting")) {
		return 5, true
	}
//volume state
	if bytes.Equal([]byte(data), []byte("online")) {
		return 1, true
	}
	if bytes.Equal([]byte(data), []byte("offline")) {
		return 0, true
	}
	if bytes.Equal([]byte(data), []byte("restricted")) {
		return 2, true
	}
	if bytes.Equal([]byte(data), []byte("mixed")) {
		return 3, true
	}

	value, err := strconv.ParseFloat(string(data), 64)
	return value, err == nil
}

func stringToFloat64Slice(data []string) ([]float64, bool ) {
  var numbers []float64
	for _, arg := range data {
		if n, err := strconv.ParseFloat(arg, 64); err == nil {
				numbers = append(numbers, n)
		}
		
	}
	return numbers, true 

}

func float64SliceSum(data []float64) float64 {
  var sum float64
	for _, value := range data {
		sum += value
}
return sum
}



func float64SliceToBucket(data []float64) map[float64]uint64 {

  var bucket map[float64]uint64
	for index, value := range data {
		bucket[float64(index)]=uint64(value)
}
	return bucket 
}

func boolToFloat64 (data bool) float64 {

	if data {
		return float64(1)
	}else {
		return float64(0)
	}
}