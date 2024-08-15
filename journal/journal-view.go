package journal

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

type DateRangeInput struct {
	Timestamp string
	Name      string
	Label     string
}

func formatDateRangeInputTimeStamp(time time.Time) string {
	return time.Format("2006-01-02T15:04")
}

type PaginationButton struct {
	Shown bool
	Url   string
	Page  int
	Label string
	Name  string
}

type JournalTable struct {
	Headers        []string
	Rows           [][]string
	NextUrl        string
	StartInput     *DateRangeInput
	EndInput       *DateRangeInput
	PageSize       int
	PreviousButton *PaginationButton
	NextButton     *PaginationButton
	Container      string
	Host           string
	Filter         string
}

type JournalView struct {
	journalService *JournalLogService
}

func NewJournalView(journalService *JournalLogService) *JournalView {
	return &JournalView{journalService: journalService}
}

type JournalRenderData struct {
	Start     time.Time
	End       time.Time
	Page      int
	PageSize  int
	Container string
	Host      string
	Filter    string
}

func (j *JournalView) CreateJournalTable(data *JournalRenderData) (*JournalTable, error) {
	println(fmt.Sprintf("CreateJournalTable start: %v, end: %v, page: %v, pageSize: %v, container: %v, host: %v", data.Start, data.End, data.Page, data.PageSize, data.Container, data.Host))

	location := GetLocation()

	pageData := &LogPageData{
		StartTime:     data.Start,
		EndTime:       data.End,
		Limit:         data.PageSize,
		TimeId:        data.PageSize * (data.Page - 1),
		ContainerName: data.Container,
		Hostname:      data.Host,
		Filter:        data.Filter,
	}
	var journalLogs []LogsEntry
	var err error

	journalLogs, err = j.journalService.GetLogPage(pageData)
	if err != nil {
		log.Printf("failed to get journal logs: %v", err)
		return nil, err
	}

	previousPage := 1
	if data.Page > 1 {
		previousPage = data.Page - 1
	}

	nextPage := data.Page
	if len(journalLogs) == data.PageSize {
		nextPage = data.Page + 1
	}

	const journalUrl = "/journal"
	headers := []string{"Time", "Log"}
	if data.Container != "" {
		headers = append(headers, "Host")
	}
	table := &JournalTable{
		Headers: headers,
		Rows:    [][]string{},
		NextUrl: journalUrl,
		StartInput: &DateRangeInput{
			Timestamp: formatDateRangeInputTimeStamp(data.Start),
			Name:      "start",
			Label:     "Start",
		},
		EndInput: &DateRangeInput{
			Timestamp: formatDateRangeInputTimeStamp(data.End),
			Name:      "end",
			Label:     "End",
		},
		PageSize: data.PageSize,
		NextButton: &PaginationButton{
			Shown: nextPage != data.Page,
			Url:   journalUrl,
			Page:  nextPage,
			Label: "Next",
			Name:  "next",
		},
		PreviousButton: &PaginationButton{
			Shown: data.Page != 1,
			Url:   journalUrl,
			Page:  previousPage,
			Label: "Previous",
			Name:  "previous",
		},
		Container: data.Container,
		Host:      data.Host,
		Filter:    data.Filter,
	}
	log.Printf("next url: %v", table.NextUrl)

	for _, logEntry := range journalLogs {
		var logValue string
		if data.Container != "" {
			logValue = logEntry.Log["MESSAGE"]
		} else {
			jsonObject, err := json.Marshal(logEntry.Log)
			if err != nil {
				log.Printf("failed to marshal log entry: %v", err)
				continue
			}
			logValue = fmt.Sprint(string(jsonObject))
		}
		rowContent := []string{logEntry.Time.In(location).Format("15:04:05.000 02.01.2006"), logValue}
		if data.Container != "" {
			rowContent = append(rowContent, logEntry.Log["_HOSTNAME"])
		}
		table.Rows = append(table.Rows, rowContent)

	}
	return table, nil
}

func (j *JournalView) Render(c echo.Context, data *JournalRenderData) error {
	table, err := j.CreateJournalTable(data)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "journal", table)
}
