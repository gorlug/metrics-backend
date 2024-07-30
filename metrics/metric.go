package metrics

import (
	"fmt"
	"log"
	"time"
)

type MetricType string

const (
	Disk MetricType = "disk"
	Ping MetricType = "ping"
)

func IsValidMetricType(metricType string) bool {
	for _, t := range []MetricType{Disk, Ping} {
		if MetricType(metricType) == t {
			return true
		}
	}
	return false
}

type MetricState string

const (
	OK    MetricState = "ok"
	Alert MetricState = "alert"
)

func IsValidMetricState(metricState string) bool {
	for _, s := range []MetricState{OK, Alert} {
		if MetricState(metricState) == s {
			return true
		}
	}
	return false
}

type MetricValues struct {
	Host      string     `json:"host"`
	Name      string     `json:"name"`
	Type      MetricType `json:"type"`
	Timestamp time.Time  `json:"timestamp"`
	// optional
	Value string      `json:"value,omitempty"`
	State MetricState `json:"state,omitempty"`
	Id    int
}

type Metric interface {
	GetNextState() MetricState
	String() string
	GetMetricValues() MetricValues
}

type MetricBuilder struct {
	MetricValues
}

func NewMetricBuilder() *MetricBuilder {
	return &MetricBuilder{}
}

func (m *MetricBuilder) WithHost(host string) *MetricBuilder {
	m.Host = host
	return m
}

func (m *MetricBuilder) WithName(name string) *MetricBuilder {
	m.Name = name
	return m
}

func (m *MetricBuilder) WithType(metricType MetricType) *MetricBuilder {
	if !IsValidMetricType(string(metricType)) {
		log.Fatalf("Invalid metric type: %v", metricType)
	}
	m.Type = metricType
	return m
}

func (m *MetricBuilder) WithTimestamp(timestamp time.Time) *MetricBuilder {
	m.Timestamp = timestamp
	return m
}

func (m *MetricBuilder) WithValue(value string) *MetricBuilder {
	m.Value = value
	return m
}

func (m *MetricBuilder) WithState(state MetricState) *MetricBuilder {
	if !IsValidMetricState(string(state)) {
		log.Fatalf("Invalid metric state: %v", state)
	}
	m.State = state
	return m
}

func (m *MetricBuilder) WithMetricValues(metricValues MetricValues) *MetricBuilder {
	m.MetricValues = metricValues
	return m
}

func (m *MetricBuilder) Build() Metric {
	if m.State == "" {
		m.WithState(OK)
	}
	switch m.Type {
	case Disk:
		return &DiskMetric{
			MetricValues: m.MetricValues,
		}
	default:
		return &PingMetric{
			MetricValues: m.MetricValues,
		}
	}
}

func (m MetricValues) String() string {
	return fmt.Sprintf("Metric{Host: %v, Name: %v, Type: %v, Timestamp: %v, Value: %v, State: %v}", m.Host, m.Name, m.Type, m.Timestamp, m.Value, m.State)
}

func PrintMetrics(metrics []Metric) {
	for _, metric := range metrics {
		fmt.Println(metric)
	}
}
