package metrics

import (
	"log"
	"strconv"
)

type DiskMetric struct {
	MetricValues
}

func (m *DiskMetric) GetNextState() MetricState {
	floatValue, err := strconv.ParseFloat(m.Value, 64)
	if err != nil {
		log.Printf("Invalid value for disk metric: %v\n", m.Value)
		return Alert
	}
	if floatValue > 90 {
		return Alert
	}
	return OK
}

func (m *DiskMetric) GetMetricValues() MetricValues {
	return m.MetricValues
}
