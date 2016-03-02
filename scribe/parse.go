package scribe

import (
	"errors"
	"fmt"
	"strings"
	"time"

	syslog "github.com/olark/scribe/syslog"
)

// we need to parse logs of the form 2015-10-14 15:58:24,543 - INFO - servicename - message

type OlarkLogFormat struct {
	timestamp   time.Time
	level       string
	serviceName string
	message     string
}

func getPriorityFromString(s string) syslog.Priority {
	switch s {
	case "INFO":
		return syslog.LOG_INFO
	case "ERROR":
		return syslog.LOG_ERR
	case "WARNING":
		return syslog.LOG_WARNING
	}

	return syslog.LOG_DEBUG
}

func parseOlarkLogFormat(logLine string) (logData *OlarkLogFormat, e error) {
	parts := strings.SplitN(logLine, " ", 8)

	if len(parts) < 8 {
		return nil, errors.New("not enough whitespace-separated strings on this line")
	}

	dateString := parts[0]
	timeString := parts[1]

	// golang doesn't properly support ISO8601 dates, so we have to use a
	// slightly different format, replacing comma with period
	// https://github.com/golang/go/issues/6189
	timeString = strings.Replace(timeString, ",", ".", 1)
	datetimeString := strings.Join([]string{dateString, timeString}, " ")
	levelString := parts[3]
	serviceName := parts[5]
	message := parts[7]

	timestamp, err := time.Parse("2006-01-02 15:04:05.000", datetimeString)

	if err != nil {
		logDebug(fmt.Sprintf("Unable to parse timestamp from %s\n", datetimeString), err)
		return nil, err
	}

	if parts[2] != "-" || parts[4] != "-" || parts[6] != "-" {
		return nil, errors.New("Line is not formatted according to spec")
	}

	logData = &OlarkLogFormat{
		timestamp:   timestamp,
		level:       levelString,
		serviceName: serviceName,
		message:     message,
	}

	return logData, nil
}
