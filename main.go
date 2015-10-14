package main

import (
	"log"
	"os"

	syslog "github.com/olark/scribe/syslog"

	//"github.com/olark/scribe/version"
	flag "github.com/spf13/pflag"
)

var server string
var reprintLogs bool

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

	logger.Info("shouldbeinfo ||| {\"conversation_id\": \"abc123\"}")

	os.Exit(0)
}
