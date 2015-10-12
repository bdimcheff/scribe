package main

import (
	//"log"
	"os"

	//"github.com/olark/scribe/version"
)

import flag "github.com/spf13/pflag"

var server string
var reprintLogs bool

func main() {
  flag.StringVarP(&server, "server", "s", "localhost", "syslog server to log to")
  flag.BoolVarP(&reprintLogs, "print", "p", true, "reprint log lines to stdout for further capture")
  flag.Parse()

	os.Exit(0)
}
