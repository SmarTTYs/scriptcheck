package runtime

import "scriptcheck/reader"

func NewOptions() *Options {
	return &Options{}
}

type Options struct {
	Shell      string
	OutputFile string

	PipelineType reader.PipelineType
	Debug        bool
	Merge        bool

	ShellCheckArgs []string

	OutputDirectory string
}
