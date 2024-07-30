package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPingMetricAlert(t *testing.T) {
	t.Run("should return alert if the timestamp is older than 5 minutes", func(t *testing.T) {
		metric := &PingMetric{MetricValues{Type: Ping, State: OK, Timestamp: getTimeMinusMinutes(6)}}
		nextState := metric.GetNextState()
		assert.Equal(t, Alert, nextState)
	})

	t.Run("should not return alert", func(t *testing.T) {
		t.Run("if the timestamp is younger than 5 minutes ago", func(t *testing.T) {
			metric := &PingMetric{MetricValues{Type: Ping, State: Alert, Timestamp: getTimeMinusMinutes(4)}}
			nextState := metric.GetNextState()
			assert.Equal(t, OK, nextState)
		})
	})

	t.Run("if a value is provided use that as the time in minutes till alert", func(t *testing.T) {
		t.Run("should alert after 24h", func(t *testing.T) {
			metric := &PingMetric{MetricValues{
				Type: Ping, State: OK, Timestamp: getTimeMinusMinutes(1445), Value: "1440"}}
			nextState := metric.GetNextState()
			assert.Equal(t, Alert, nextState)
		})

		t.Run("should not alert if it is before 24h", func(t *testing.T) {
			metric := &PingMetric{MetricValues{
				Type: Ping, State: OK, Timestamp: getTimeMinusMinutes(1439), Value: "1440"}}
			nextState := metric.GetNextState()
			assert.Equal(t, OK, nextState)
		})
	})
}
