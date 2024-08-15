package logger

import (
	"log"
)

func LogDebug(message string, a ...any) {
	log.Printf(message, a...)
}
