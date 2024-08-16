package logger

import (
	"fmt"
	"log"
)

func LogDebug(message string, a ...any) {
	log.Printf(message, a...)
}

func LogError(message string, a ...any) {
	log.Printf("Error: %v", fmt.Sprintf(message, a...))
}
