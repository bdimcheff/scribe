package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	flag "github.com/spf13/pflag"

	syslog "github.com/olark/scribe/syslog"
	"github.com/olark/scribe/version"
)

var server string
var quietMode bool
var dryRun bool
var tag string
var bufferLength int
var verbose bool
var showVersion bool

// we need to parse logs of the form 2015-10-14 15:58:24,543 - INFO - servicename - message

type olarkLogFormat struct {
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

func parseOlarkLogFormat(logLine string) (logData olarkLogFormat, e error) {
	parts := strings.SplitN(logLine, " ", 8)

	if len(parts) < 8 {
		return olarkLogFormat{}, errors.New("not enough whitespace-separated strings on this line")
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
		logDebug(fmt.Sprintf("Unable to parse timestamp from %s\n", datetimeString))
		logDebug(err)
		return olarkLogFormat{}, err
	}

	if parts[2] != "-" || parts[4] != "-" || parts[6] != "-" {
		return olarkLogFormat{}, errors.New("Line is not formatted according to spec")
	}

	logData = olarkLogFormat{
		timestamp:   timestamp,
		level:       levelString,
		serviceName: serviceName,
		message:     message,
	}

	return logData, nil
}

func connectToLogger() (logger *syslog.Writer, err error) {
	errorCallback := func(err error, backoffTime time.Duration) {
		logError("Connect to remote syslog failed, retrying")
	}

	connect := func() error {
		var connectError error
		logger, connectError = syslog.Dial("tcp", server, syslog.LOG_DEBUG, tag)

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

func logRaw(level, message interface{}) {
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05.000")

	fmt.Printf("%s - %s - scribe - %s", timestamp, level, message)
	fmt.Println("")

}

func logMessage(message interface{}) {
	logRaw("INFO", message)
}

func logError(message interface{}) {
	logRaw("ERROR", message)
}

func logDebug(message interface{}) {
	if verbose {
		logRaw("DEBUG", message)
	}
}

func parseCommandLineOptions() {
	flag.StringVarP(&server, "server", "s", "localhost", "syslog server to log to")
	flag.BoolVarP(&quietMode, "quiet", "q", false, "don't reprint log lines to stdout for further capture")
	flag.BoolVarP(&dryRun, "dry", "d", false, "don't actually log to syslog")
	flag.StringVarP(&tag, "tag", "t", "scribe", "override the service/component from logs with this tag")
	flag.IntVarP(&bufferLength, "buffer-length", "b", 100000, "number of log lines to buffer before dropping them")
	flag.BoolVarP(&verbose, "verbose", "v", false, "log scribe messages/errors")
	flag.BoolVarP(&showVersion, "version", "", false, "display scribe version")
	flag.Parse()
}

func main() {
	logMessage("scribe started")

	parseCommandLineOptions()

	if showVersion {
		fmt.Println(version.GetFullVersion())
		os.Exit(0)
	}

	scanner := bufio.NewScanner(os.Stdin)

	var logger *syslog.Writer
	var err error

	logChannel := make(chan string, bufferLength)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			if !quietMode {
				fmt.Println(line)
			}

			select {
			case logChannel <- line:
				// line successfully enqueued to channel, so we can do nothing
			default:
				if !quietMode {
					logError("Buffer full, dropping log line.")
				}
			}
		}
		close(logChannel)
	}()

	if !dryRun {
		logger, err = connectToLogger()

		if err != nil {
			// this should really never happen because connectToLogger should
			// retry forever
			logError("Error connecting to logger.  Not exiting, but logs are not being sent remotely.")
			dryRun = true
		}
	}

	for {
		line, more := <-logChannel
		if !more {
			break
		}

		logData, err := parseOlarkLogFormat(line)

		if err != nil {
			logDebug("Unable to process previous line due to formatting error:")
			logDebug(err)

			continue
		}

		if logger != nil && !dryRun {
			priority := getPriorityFromString(logData.level)
			logger.WriteDetailed(priority, &logData.timestamp, logData.serviceName, logData.message)
		}
	}
}
