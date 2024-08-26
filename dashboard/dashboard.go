package dashboard

import (
	. "github.com/gorlug/metrics-backend/metrics"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strconv"
)

type MetricRow struct {
	Id      string
	Host    string
	Name    string
	IsAlert bool
	State   string
	Values  []string
}

type Table struct {
	Headers     []string
	Rows        []MetricRow
	DeleteLabel string
}

type Dashboard struct {
	metricsService *DbMetricsService
}

func NewDashboard(metricsService *DbMetricsService) *Dashboard {
	return &Dashboard{metricsService: metricsService}
}

func (d *Dashboard) Render(c echo.Context) error {
	metrics, err := d.metricsService.GetAllMetrics()
	if err != nil {
		log.Println("failed to get metrics", err)
		return err
	}
	table := &Table{
		Headers:     []string{"Host", "Name", "Type", "Value", "Timestamp", "State"},
		Rows:        []MetricRow{},
		DeleteLabel: "Delete",
	}
	for _, metricObject := range metrics {
		table.Rows = append(table.Rows, metricToMetricRow(metricObject))
	}

	return c.Render(http.StatusOK, "dashboard", table)
}

func metricToMetricRow(metricObject Metric) MetricRow {
	metric := metricObject.GetMetricValues()
	return MetricRow{Id: strconv.Itoa(metric.Id),
		IsAlert: metric.State == Alert,
		Host:    metric.Host, Name: metric.Name,
		State: string(metric.State),
		Values: []string{
			metric.Host,
			metric.Name,
			string(metric.Type),
			metric.Value,
			metric.Timestamp.Format("2006-01-02 15:04:05"),
		}}
}
