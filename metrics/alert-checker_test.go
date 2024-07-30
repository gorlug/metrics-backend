package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getAlertChecker(metrics []Metric, getAllMetricsError error) (*AlertChecker, *MockMetricsService, *MockAlerter) {
	service := &MockMetricsService{metrics: metrics, getAllMetricsError: getAllMetricsError}
	alerter := &MockAlerter{}
	return &AlertChecker{metricsService: service, alerter: alerter}, service, alerter
}

func TestAlertChecker(t *testing.T) {
	t.Run("should create a new alert", func(t *testing.T) {
		// arrange
		alertChecker, service, alerter := getAlertChecker([]Metric{
			&MockMetric{
				NextState:    Alert,
				MetricValues: MetricValues{Host: "host1", Name: "some metric", Type: Ping, State: OK},
			},
		}, nil)
		// act
		alertChecker.CheckAlerts()
		// assert
		assert.Equal(t, 1, len(service.stateSaved))
		expectedMetric := MetricValues{Host: "host1", Name: "some metric", Type: Ping, State: Alert}
		assert.EqualValues(t, expectedMetric, service.stateSaved[0])

		assert.Equal(t, 1, len(alerter.newAlerts))
		assert.EqualValues(t, expectedMetric, alerter.newAlerts[0].GetMetricValues())
		assert.Equal(t, 0, len(alerter.alertsOkAgain))
	})

	t.Run("should not create a new alert if the next state is not alert", func(t *testing.T) {
		// arrange
		alertChecker, service, alerter := getAlertChecker([]Metric{
			&MockMetric{
				NextState:    OK,
				MetricValues: MetricValues{Host: "host1", Name: "some metric", Type: Ping, State: OK},
			},
		}, nil)
		// act
		alertChecker.CheckAlerts()
		// assert
		assert.Equal(t, 0, len(service.stateSaved))
		assert.Equal(t, 0, len(alerter.newAlerts))
		assert.Equal(t, 0, len(alerter.alertsOkAgain))
	})

	t.Run("should not create a new alert if the current state is already an alert", func(t *testing.T) {
		// arrange
		alertChecker, service, alerter := getAlertChecker([]Metric{
			&MockMetric{
				NextState:    Alert,
				MetricValues: MetricValues{Host: "host1", Name: "some metric", Type: Ping, State: Alert},
			},
		}, nil)
		// act
		alertChecker.CheckAlerts()
		// assert
		assert.Equal(t, 0, len(service.stateSaved))
		assert.Equal(t, 0, len(alerter.newAlerts))
		assert.Equal(t, 0, len(alerter.alertsOkAgain))
	})

	t.Run("should not crash if GetMetricValues fails and send an alert", func(t *testing.T) {
		// arrange
		alertChecker, service, alerter := getAlertChecker(nil, assert.AnError)
		// act
		alertChecker.CheckAlerts()
		// assert
		assert.Equal(t, 0, len(service.stateSaved))
		assert.Equal(t, 1, len(alerter.newAlerts))
		assert.EqualValues(t, GetFailedToGetMetricsMetric(), alerter.newAlerts[0])
		assert.Equal(t, 0, len(alerter.alertsOkAgain))
	})

	t.Run("should send an ok message and save if it was in alert state and now is ok", func(t *testing.T) {
		// arrange
		alertChecker, service, alerter := getAlertChecker([]Metric{
			&MockMetric{
				NextState:    OK,
				MetricValues: MetricValues{Host: "host1", Name: "some metric", Type: Ping, State: Alert},
			},
		}, nil)
		// act
		alertChecker.CheckAlerts()
		// assert
		assert.Equal(t, 1, len(service.stateSaved))
		expectedMetric := MetricValues{Host: "host1", Name: "some metric", Type: Ping, State: OK}
		assert.EqualValues(t, expectedMetric, service.stateSaved[0])

		assert.Equal(t, 1, len(alerter.alertsOkAgain))
		assert.EqualValues(t, expectedMetric, alerter.alertsOkAgain[0].GetMetricValues())
	})
}

type MockMetric struct {
	NextState MetricState
	MetricValues
}

func (m *MockMetric) GetNextState() MetricState {
	return m.NextState
}

func (m *MockMetric) GetMetricValues() MetricValues {
	return m.MetricValues
}

type MockMetricsService struct {
	metrics            []Metric
	getAllMetricsError error
	metricsSaved       []MetricValues
	stateSaved         []MetricValues
}

func (m *MockMetricsService) GetAllMetrics() ([]Metric, error) {
	return m.metrics, m.getAllMetricsError
}

func (m *MockMetricsService) SaveMetric(metric MetricValues) error {
	if m.metricsSaved == nil {
		m.metricsSaved = []MetricValues{}
	}
	m.metricsSaved = append(m.metricsSaved, metric)
	return nil
}

func (m *MockMetricsService) SaveState(metric MetricValues, state MetricState) error {
	if m.stateSaved == nil {
		m.stateSaved = []MetricValues{}
	}
	updated := metric
	updated.State = state

	m.stateSaved = append(m.stateSaved, updated)
	return nil
}

type MockAlerter struct {
	newAlerts     []Metric
	alertsOkAgain []Metric
}

func (m *MockAlerter) NewAlert(metric Metric) error {
	if m.newAlerts == nil {
		m.newAlerts = []Metric{}
	}
	m.newAlerts = append(m.newAlerts, metric)
	return nil
}

func (m *MockAlerter) AlertOkAgain(metric Metric) error {
	if m.alertsOkAgain == nil {
		m.alertsOkAgain = []Metric{}
	}
	m.alertsOkAgain = append(m.alertsOkAgain, metric)
	return nil
}
