package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/olark/scribe/scribe"
	"github.com/olark/scribe/version"
)

var server string
var quietMode bool
var dryRun bool
var tag string
var bufferLength int
var verbose bool
var showVersion bool

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
	parseCommandLineOptions()

	if showVersion {
		fmt.Println(version.GetFullVersion())
		os.Exit(0)
	}

	opts := &scribe.Options{
		Server:       server,
		QuietMode:    quietMode,
		DryRun:       dryRun,
		Tag:          tag,
		BufferLength: bufferLength,
		Verbose:      verbose,
	}

	scribe.Run(opts)
}
