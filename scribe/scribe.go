package scribe

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"

	syslog "github.com/olark/scribe/syslog"
)

type scribe struct {
	*Options
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

		logData, err := parseOlarkLogFormat(line)

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
