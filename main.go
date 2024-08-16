package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/robfig/cron"
	"log"
	"metrics-backend/journal"
	"metrics-backend/metrics"
	"metrics-backend/rest"
	"metrics-backend/user"
	"os"
	_ "time/tzdata"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
		log.Fatal("Error loading .env file")
	}

	telegramAlerter := &metrics.TelegramAlerter{
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatId: os.Getenv("TELEGRAM_CHAT_ID"),
	}

	metricsService, err := metrics.NewDBMetricsService(os.Getenv("DATABASE_URL"), telegramAlerter)
	CheckError(err)
	defer metricsService.Close()

	var journalService *journal.JournalLogService
	timescaleDbUrl := os.Getenv("TIMESCALE_DATABASE_URL")
	if timescaleDbUrl != "" {
		logService, err := journal.NewJournalLogService(timescaleDbUrl)
		if err != nil {
			log.Printf("Journal service is disabled: %v", err)
		} else {
			log.Print("Journal service is enabled")
		}
		CheckError(err)
		journalService = logService
		defer logService.Close()
	}

	userService, err := user.NewUserService(metricsService.ConnPool)
	CheckError(err)

	alertChecker := metrics.NewAlertChecker(metricsService, telegramAlerter)
	alertChecker.CheckAlerts()

	cronSpec := cron.New()
	interval, exists := os.LookupEnv("CHECK_INTERVAL")
	if !exists {
		interval = "5m"
	}
	err = cronSpec.AddFunc(fmt.Sprintf("@every %v", interval), alertChecker.CheckAlerts)
	CheckError(err)
	cronSpec.Start()
	defer cronSpec.Stop()

	rest.CreateRestApi(metricsService, journalService, userService)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
