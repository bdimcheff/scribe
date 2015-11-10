package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"

	syslog "github.com/olark/scribe/syslog"

	//"github.com/olark/scribe/version"
	flag "github.com/spf13/pflag"
)

var server string
var reprintLogs bool
var dryRun bool
var tag string
var bufferLength int

// we need to parse logs of the form 2015-10-14 15:58:24,543 - INFO - servicename - message

type olarkLogFormat struct {
	timestamp   time.Time
	level       string
	serviceName string
	message     string
}

func getLogFunction(writer *syslog.Writer, s string) (logFunction func(string) error) {
	switch s {
	case "INFO":
		return writer.Info
	case "ERROR":
		return writer.Err
	case "WARNING":
		return writer.Warning
	}

	return writer.Debug
}

//Mon Jan 2 15:04:05 -0700 MST 2006

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
		fmt.Printf("ERROR: unable to parse timestamp from log line %s\n", logLine)
		log.Println(err)
		return olarkLogFormat{}, err
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
		fmt.Fprintln(os.Stderr, "connect to remote syslog failed")
	}

	connect := func() error {
		var connectError error
		logger, connectError = syslog.Dial("tcp", server, syslog.LOG_DEBUG, tag)

		return connectError
	}

	backoffConfig := backoff.NewExponentialBackOff()
	// retry forever
	backoffConfig.MaxElapsedTime = 0

	backoff.RetryNotify(connect, backoffConfig, errorCallback)

	if err != nil {
		return nil, err
	}

	//successfully connected to a logger

	return logger, nil
}

func parseCommandLineOptions() {
	flag.StringVarP(&server, "server", "s", "localhost", "syslog server to log to")
	flag.BoolVarP(&reprintLogs, "print", "p", true, "reprint log lines to stdout for further capture")
	flag.BoolVarP(&dryRun, "dry", "d", false, "don't actually log to syslog")
	flag.StringVarP(&tag, "tag", "t", "scribe", "override the service/component from logs with this tag")
	flag.IntVarP(&bufferLength, "buffer-length", "b", 100000, "number of log lines to buffer before dropping them")
	flag.Parse()
}

func main() {
	parseCommandLineOptions()

	scanner := bufio.NewScanner(os.Stdin)

	var logger *syslog.Writer
	var err error

	if !dryRun {
		logger, err = connectToLogger()

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error connecting to logger.  Not exiting, but logs are being dropped.")
			dryRun = true
		}
	}

	logChannel := make(chan string, bufferLength)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			select {
			case logChannel <- line:
			default:
				fmt.Fprintln(os.Stderr, "Channel full, unable to log")
			}
		}
	}()

	for {
		line := <-logChannel

		logData, err := parseOlarkLogFormat(line)

		if reprintLogs {
			fmt.Println(line) // Println will add back the final '\n'
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, line)
			fmt.Fprintln(os.Stderr, "Unable to process previous line due to formatting error:")
			fmt.Fprintln(os.Stderr, err)

			continue
		}
		if !dryRun {
			loggerFunction := getLogFunction(logger, logData.level)
			loggerFunction(logData.message)
		}
	}
}
