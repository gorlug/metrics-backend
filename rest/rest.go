package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"log"
	"metrics-backend/dashboard"
	. "metrics-backend/metrics"
	"net/http"
	"strconv"
	"time"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.gohtml")),
	}
}

func CreateRestApi(metricsService *DbMetricsService) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = newTemplate()

	api := NewApi(metricsService)

	e.POST("/metric", api.createMetric)
	e.GET("/dashboard", api.ShowDashboard)
	e.POST("/delete/:id", api.DeleteMetric)

	e.Logger.Fatal(e.Start(":8080"))
}

type Api struct {
	metricsService *DbMetricsService
}

func NewApi(metricsService *DbMetricsService) *Api {
	return &Api{metricsService: metricsService}
}

func (a *Api) createMetric(c echo.Context) error {
	var metric MetricValues

	if err := c.Bind(&metric); err != nil {
		return err
	}

	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	log.Printf("received metric %v", metric.String())
	err := a.metricsService.SaveMetric(metric)
	if err != nil {
		log.Println("failed to save metric", err)
		return err
	}

	return c.String(http.StatusOK, "ok")
}

type DashboardMetric struct {
	MetricValues
	Id        string
	Timestamp string
}

type DashboardData struct {
	Metrics []DashboardMetric
}

func (a *Api) ShowDashboard(c echo.Context) error {
	return dashboard.NewDashboard(a.metricsService).Render(c)
}

func (a *Api) DeleteMetric(c echo.Context) error {
	id := c.Param("id")
	log.Printf("deleting metric with id %v", id)

	intId, err := strconv.Atoi(id)
	if err != nil {
		log.Println("failed to convert id to int", err)
		return err
	}

	err = a.metricsService.DeleteMetric(intId)
	if err != nil {
		log.Println("failed to delete metric", err)
		return err
	}
	return a.ShowDashboard(c)
}