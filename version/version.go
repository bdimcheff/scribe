package version

import (
	"fmt"
)

const (
	VERSION = "0.0.1"
)

var (
	GitCommit string
)

// Version string including git commit hash
func GetVersion() string {
	return fmt.Sprintf("v%s-%s", VERSION, GitCommit)
}

// Version string prefixed with program name
func GetFullVersion() string {
	return "scribe " + GetVersion()
}
