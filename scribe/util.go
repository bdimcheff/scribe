package scribe

import (
	"fmt"
	"time"
)

func (s *scribe) logRaw(level string, message ...interface{}) {
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05.000")

	if !s.QuietMode || level == "ERROR" {
		fmt.Printf("%s - %s - scribe - %s\n", timestamp, level, message)
	}
}

func (s *scribe) logMessage(message ...interface{}) {
	s.logRaw("INFO", message)
}

func (s *scribe) logError(message ...interface{}) {
	s.logRaw("ERROR", message)
}

func (s *scribe) logDebug(message ...interface{}) {
	s.logRaw("DEBUG", message)
}
