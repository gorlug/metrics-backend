package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/robfig/cron"
	"log"
	"metrics-backend/metrics"
	"metrics-backend/rest"
	"os"
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

	journalLogService, err := metrics.NewJournalLogService(os.Getenv("TIMESCALE_DATABASE_URL"))
	CheckError(err)
	defer journalLogService.Close()

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

	rest.CreateRestApi(metricsService, journalLogService)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
