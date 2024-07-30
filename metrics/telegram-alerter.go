package metrics

import (
	"fmt"
	"log"
	"net/http"
)

// https://api.telegram.org/bot${telegramToken.Parameter?.Value}/sendMessage?chat_id=${telegramChatId.Parameter?.Value}&text=${alarmDescription}

type TelegramAlerter struct {
	TelegramToken  string
	TelegramChatId string
}

func (a *TelegramAlerter) NewAlert(metric Metric) error {
	log.Printf("Sending alert for metric: %v\n", metric.String())
	alertMessage := fmt.Sprintf("Alert: %v", getFormatedMetricMessage(metric))
	return a.sendTelegramMessage(alertMessage)
}

func (a *TelegramAlerter) AlertOkAgain(metric Metric) error {
	log.Printf("Sending ok again for metric: %v\n", metric.String())
	okMessage := fmt.Sprintf("OK again: %v", getFormatedMetricMessage(metric))
	return a.sendTelegramMessage(okMessage)
}

func (a *TelegramAlerter) sendTelegramMessage(alertMessage string) error {
	requestUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", a.TelegramToken, a.TelegramChatId, alertMessage)
	_, err := http.Get(requestUrl)
	return err
}

func getFormatedMetricMessage(metric Metric) string {
	valueMessage := ""
	if metric.GetMetricValues().Value != "" {
		valueMessage = fmt.Sprintf(" Value: %v", metric.GetMetricValues().Value)
	}
	return fmt.Sprintf("%v - %v%v", metric.GetMetricValues().Host, metric.GetMetricValues().Name, valueMessage)
}
