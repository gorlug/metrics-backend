package rest

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"html/template"
	"io"
	"log"
	"metrics-backend/dashboard"
	. "metrics-backend/journal"
	"metrics-backend/logger"
	. "metrics-backend/metrics"
	"metrics-backend/user"
	"net/http"
	"os"
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

func CreateRestApi(metricsService *DbMetricsService, journalService *JournalLogService, userService *user.UserService) {
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_CALLBACK_URL")),
	)
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options.HttpOnly = true
	gothic.Store = store

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(CreateAuthenticationMiddleware(userService))
	e.Renderer = newTemplate()

	api := NewApi(metricsService, journalService, store, userService)

	e.POST("/metric", api.createMetric)
	e.GET("/dashboard", api.ShowDashboard)
	e.POST("/delete/:id", api.DeleteMetric)
	log.Printf("journal service: %v", journalService)
	if journalService != nil {
		e.GET("/journal", api.ShowJournal)
		e.POST("/journal", api.PostJournal)
	}

	e.GET("/auth/:provider", api.Authenticate)
	e.GET("/auth/:provider/callback", api.AuthCallback)
	e.GET("/logout", api.Logout)

	e.Logger.Fatal(e.Start(":8080"))
}

type Api struct {
	metricsService *DbMetricsService
	journalService *JournalLogService
	store          sessions.Store
	userService    *user.UserService
}

func NewApi(metricsService *DbMetricsService, journalService *JournalLogService, store sessions.Store, userService *user.UserService) *Api {
	return &Api{metricsService: metricsService, journalService: journalService, store: store, userService: userService}
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

type JournalBody struct {
	Logs string `json:"logs"`
}

func (a *Api) PostJournal(c echo.Context) error {
	var journalBody JournalBody
	err := c.Bind(&journalBody)
	if err != nil {
		log.Println("failed to parse journal logs body", err)
		return err
	}

	err = a.journalService.SaveJournalLogs(journalBody.Logs)

	if err != nil {
		log.Println("failed to save journal logs", err)
		return err
	}
	return nil
}

func CreateAuthenticationMiddleware(userService *user.UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			paths := []string{"/auth/:provider", "/auth/:provider/callback", "/logout", "/metric"}
			for _, path := range paths {
				if c.Path() == path {
					return next(c)
				}
			}
			if c.Path() == "/journal" && c.Request().Method == "POST" {
				return next(c)
			}
			session, err := gothic.Store.Get(c.Request(), "session")
			if err != nil {
				return redirectToAuthentication(c)
			}
			if session.Values["user"] == nil {
				return redirectToAuthentication(c)
			}
			userEmail := session.Values["user"]

			userExists, err := userService.DoesUserEmailExist(fmt.Sprint(userEmail))
			if err != nil {
				return err
			}
			if !userExists {
				return redirectToAuthentication(c)
			}
			logger.LogDebug("user exists %v", userEmail)

			c.Set("user", userEmail)
			return next(c)
		}
	}
}

func redirectToAuthentication(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, "/auth/google")
}

func (a *Api) ShowJournal(c echo.Context) error {
	user := c.Get("user")
	logger.LogDebug("user %v", user)

	start := c.QueryParam("start")
	end := c.QueryParam("end")
	timezone := c.QueryParam("timezone")
	page := c.QueryParam("page")
	pageSize := c.QueryParam("pageSize")

	renderData := &JournalRenderData{
		Start:     ParseTime(start, 0, timezone),
		End:       ParseTime(end, 10, timezone),
		Page:      parseIntWithDefault(page, 1),
		PageSize:  parseIntWithDefault(pageSize, 10),
		Container: c.QueryParam("container"),
		Host:      c.QueryParam("host"),
		Filter:    c.QueryParam("filter"),
	}

	return NewJournalView(a.journalService).Render(c, renderData)
}

func (a *Api) Authenticate(c echo.Context) error {
	logger.LogDebug("authenticating with provider %v", c.Param("provider"))
	// try to get the user without re-authenticating
	query := c.Request().URL.Query()
	query.Add("provider", c.Param("provider"))
	c.Request().URL.RawQuery = query.Encode()
	gothUser, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err == nil {
		logger.LogDebug("user %v", gothUser.Email)
	} else {
		logger.LogDebug("user not found, starting auth %v", err)
		gothic.BeginAuthHandler(c.Response().Writer, c.Request())
	}
	return nil
}

func (a *Api) AuthCallback(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return err
	}
	logger.LogDebug("user %v", user.Name)
	// create a session cookie for the user
	s, err := gothic.Store.New(c.Request(), "session")
	if err != nil {
		return err
	}
	s.Values["user"] = user.Email
	err = s.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/journal")
}

func (a *Api) Logout(c echo.Context) error {
	err := gothic.Logout(c.Response(), c.Request())
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func ParseTime(timeString string, durationDifference int, timezone string) time.Time {
	if timezone == "" {
		timezone = "Europe/Berlin"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		println(fmt.Sprintf("failed to load location %v, error: %v", timezone, err))
		location = time.Local
	}
	timeObject, err := time.Parse("2006-01-02T15:04", timeString)
	if err == nil {
		timeObject, err = time.ParseInLocation("2006-01-02T15:04", timeString, location)
	}
	if err != nil {
		timeObject = time.Now().Add(time.Duration(-1) * time.Hour)
		timeObject = timeObject.Add(time.Duration(durationDifference) * time.Minute)
	}
	return timeObject
}

func parseIntWithDefault(pageString string, defaultValue int) int {
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return defaultValue
	}
	return page
}
