package metrics

import (
	"os"
)

func GetFailedToGetMetricsMetric() Metric {
	return NewMetricBuilder().WithHost(getHostname()).WithName("Failed to get metrics").WithType(Ping).Build()
}

func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}
