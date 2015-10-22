package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	syslog "github.com/olark/scribe/syslog"

	//"github.com/olark/scribe/version"
	flag "github.com/spf13/pflag"
)

var server string
var reprintLogs bool
var dryRun bool

// we need to parse logs of the form 2015-10-14 15:58:24,543 - INFO - servicename - message

type olarkLogFormat struct {
	timestamp   time.Time
	level       string
	serviceName string
	message     string
}

// func getLogFunction(s string) (logFunction, err error) {
//
// }
//Mon Jan 2 15:04:05 -0700 MST 2006

func parseOlarkLogFormat(logLine string) (logData olarkLogFormat, e error) {
	parts := strings.SplitN(logLine, " ", 7)
	dateString := parts[0]
	timeString := parts[1]
	timeString = strings.Replace(timeString, ",", ".", 1)
	datetimeString := strings.Join([]string{dateString, timeString}, " ")
	levelString := parts[3]
	serviceName := parts[5]
	message := parts[6]

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
	logger, err = syslog.Dial("tcp", "localhost:10514", syslog.LOG_DEBUG, "brandon-test")

	if err != nil {
		log.Println("Error connecting to syslog")
		log.Println(err)

		return nil, err
	}

	return logger, nil
}

func main() {
	flag.StringVarP(&server, "server", "s", "localhost", "syslog server to log to")
	flag.BoolVarP(&reprintLogs, "print", "p", true, "reprint log lines to stdout for further capture")
	flag.BoolVarP(&dryRun, "dry", "d", false, "don't actually log to syslog")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
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
		log.Println(logData)
		if !dryRun {
			logger, err := connectToLogger()

			if err != nil {
				os.Exit(1)
			}

			logger.Info(line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	os.Exit(0)
}
