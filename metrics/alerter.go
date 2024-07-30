package metrics

type Alerter interface {
	NewAlert(metric Metric) error
	AlertOkAgain(metric Metric) error
}
