package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiskMetric(t *testing.T) {
	t.Run("should be alert if disk percentage is over 90%", func(t *testing.T) {
		metric := &DiskMetric{MetricValues{Type: Disk, Value: "91"}}
		nextState := metric.GetNextState()
		assert.Equal(t, Alert, nextState)
	})

	t.Run("should not be alert if disk percentage is 90% or less", func(t *testing.T) {
		metric := &DiskMetric{MetricValues{Type: Disk, Value: "90"}}
		nextState := metric.GetNextState()
		assert.Equal(t, OK, nextState)
	})
}
