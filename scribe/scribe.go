package scribe

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"

	syslog "github.com/olark/scribe/syslog"
)

type scribe struct {
	*Options
}

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

func (s *scribe) parseOlarkLogFormat(logLine string) (logData *OlarkLogFormat, e error) {
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

func (s *scribe) connectToLogger() (logger *syslog.Writer, err error) {
	errorCallback := func(err error, backoffTime time.Duration) {
		logError("Connect to remote syslog failed, retrying")
	}

	connect := func() error {
		var connectError error
		logger, connectError = syslog.Dial("tcp", s.Server, syslog.LOG_DEBUG, s.Tag)

		return connectError
	}

	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.MaxElapsedTime = 0

	backoff.RetryNotify(connect, backoffConfig, errorCallback)

	if err != nil {
		return nil, err
	}

	logMessage("Connected to logger")

	return logger, nil
}

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

func Run(opts *Options) {
	s := &scribe{Options: opts}

	logMessage("scribe started")

	scanner := bufio.NewScanner(os.Stdin)

	var logger *syslog.Writer
	var err error

	logChannel := make(chan string, opts.BufferLength)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			if !s.QuietMode {
				fmt.Println(line)
			}

			select {
			case logChannel <- line:
				// line successfully enqueued to channel, so we can do nothing
			default:
				if !s.QuietMode {
					logError("Buffer full, dropping log line.")
				}
			}
		}
		close(logChannel)
	}()

	if !s.DryRun {
		logger, err = s.connectToLogger()

		if err != nil {
			// this should really never happen because connectToLogger should
			// retry forever
			logError("Error connecting to logger.  Not exiting, but logs are not being sent remotely.")
			s.DryRun = true
		}
	}

	for {
		line, more := <-logChannel
		if !more {
			break
		}

		logData, err := s.parseOlarkLogFormat(line)

		if err != nil && s.Verbose {
			logDebug("Unable to process previous line due to formatting error:", err)
			continue
		}

		if logger != nil && logData != nil && !s.DryRun {
			priority := getPriorityFromString(logData.level)
			logger.WriteDetailed(priority, &logData.timestamp, logData.serviceName, logData.message)
		}
	}
}
