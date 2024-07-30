package metrics

import (
	"log"
)

type AlertChecker struct {
	metricsService          MetricsService
	alerter                 Alerter
	MetricsServiceErrorSent bool
}

func NewAlertChecker(metricsService MetricsService, alerter Alerter) *AlertChecker {
	return &AlertChecker{metricsService: metricsService, alerter: alerter, MetricsServiceErrorSent: false}
}

func (a *AlertChecker) CheckAlerts() {
	log.Println("Checking alerts")
	metricsArr, err := a.metricsService.GetAllMetrics()
	if err != nil {
		a.sendFailedToGetMetrics(err)
		return
	}
	if a.MetricsServiceErrorSent {
		a.sendGettingMetricsOkAgain()
	}
	a.MetricsServiceErrorSent = false
	for _, metric := range metricsArr {
		log.Printf("Checking metric %v", metric.String())
		if IsMetricInNewStateAlert(metric) {
			log.Printf("setting alert for metric %v", metric.String())
			updatedMetricValues, err := a.saveNewState(metric, Alert)
			err = a.alerter.NewAlert(NewMetricBuilder().WithMetricValues(updatedMetricValues).Build())
			if err != nil {
				log.Println("Failed to send alert", err)
			}
		}
		if IsMetricOkAgain(metric) {
			log.Printf("setting ok for metric %v", metric.String())
			updatedMetricValues, err := a.saveNewState(metric, OK)
			err = a.alerter.AlertOkAgain(NewMetricBuilder().WithMetricValues(updatedMetricValues).Build())
			if err != nil {
				log.Println("Failed to send alert", err)
			}
		}
	}
}

func (a *AlertChecker) sendGettingMetricsOkAgain() {
	err := a.alerter.AlertOkAgain(GetFailedToGetMetricsMetric())
	if err != nil {
		log.Println("Failed to send alert", err)
	}
}

func (a *AlertChecker) sendFailedToGetMetrics(err error) {
	log.Println("Failed to get metrics", err)
	if !a.MetricsServiceErrorSent {
		err := a.alerter.NewAlert(GetFailedToGetMetricsMetric())
		if err != nil {
			log.Println("Failed to send alert", err)
		}
		a.MetricsServiceErrorSent = true
	}
}

func IsMetricOkAgain(metric Metric) bool {
	return metric.GetNextState() == OK && metric.GetMetricValues().State == Alert
}

func (a *AlertChecker) saveNewState(metric Metric, newState MetricState) (MetricValues, error) {
	updatedMetricValues := metric.GetMetricValues()
	updatedMetricValues.State = newState
	err := a.metricsService.SaveState(metric.GetMetricValues(), newState)
	if err != nil {
		log.Println("Failed to save metric", err)
	}
	return updatedMetricValues, err
}

func IsMetricInNewStateAlert(metric Metric) bool {
	return metric.GetNextState() == Alert && metric.GetMetricValues().State != Alert
}
