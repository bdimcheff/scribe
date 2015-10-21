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

func parseOlarkLogFormat(logLine string) (logData olarkLogFormat) {
	timestamp, err := time.Parse("2006-01-02 15:04:05,999", logLine)

	if err != nil {
		fmt.Printf("ERROR: unable to parse timestamp from log line %s\n", logLine)
		log.Println(err)
	}

	parts := strings.SplitN(logLine, " ", 7)
	levelString := parts[3]
	serviceName := parts[5]
	message := parts[6]

	logData = olarkLogFormat{
		timestamp:   timestamp,
		level:       levelString,
		serviceName: serviceName,
		message:     message,
	}

	return logData
}

func main() {
	flag.StringVarP(&server, "server", "s", "localhost", "syslog server to log to")
	flag.BoolVarP(&reprintLogs, "print", "p", true, "reprint log lines to stdout for further capture")
	flag.Parse()

	logger, err := syslog.Dial("tcp", "localhost:10514", syslog.LOG_DEBUG, "brandon-test")

	if err != nil {
		log.Println("Error connecting to syslog, exiting")
		log.Println(err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		logData := parseOlarkLogFormat(line)
		log.Println(logData)
		logger.Info(line)
		if reprintLogs {
			fmt.Println(line) // Println will add back the final '\n'
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	os.Exit(0)
}
