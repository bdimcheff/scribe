package scribe

import (
	"fmt"
	"time"
)

func logRaw(level string, message ...interface{}) {
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05.000")

	fmt.Printf("%s - %s - scribe - %s\n", timestamp, level, message)
}

func logMessage(message ...interface{}) {
	logRaw("INFO", message)
}

func logError(message ...interface{}) {
	logRaw("ERROR", message)
}

func logDebug(message ...interface{}) {
	logRaw("DEBUG", message)
}
