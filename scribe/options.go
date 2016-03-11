package scribe

type Options struct {
	Server       string
	QuietMode    bool
	DryRun       bool
	Tag          string
	BufferLength int
	Verbose      bool
}
