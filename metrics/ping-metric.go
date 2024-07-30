package metrics

import (
	"os"
	"strconv"
	"time"
)

type PingMetric struct {
	MetricValues
}

const defaultMinutesTillAlert = 5

func (m *PingMetric) GetNextState() MetricState {
	minutes := getTimeMinusMinutes(m.getMinutesTillAlert())
	if m.Timestamp.Before(minutes) {
		return Alert
	}
	return OK
}

func (m *PingMetric) getMinutesTillAlert() int {
	if m.Value != "" {
		minutes, err := strconv.Atoi(m.Value)
		if err == nil {
			return minutes
		}
	}
	return getSystemMinutesTillAlert()
}

func getTimeMinusMinutes(minutes int) time.Time {
	return time.Now().Add(time.Duration(-minutes) * time.Minute)
}

func (m *PingMetric) GetMetricValues() MetricValues {
	return m.MetricValues
}

func getSystemMinutesTillAlert() int {
	value, exists := os.LookupEnv("MINUTES_TILL_ALERT")
	if !exists {
		return defaultMinutesTillAlert
	}
	minutes, err := strconv.Atoi(value)
	if err != nil {
		return defaultMinutesTillAlert
	}
	return minutes
}
