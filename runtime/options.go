package runtime

import "scriptcheck/reader"

type Options struct {
	Shell      string
	OutputFile string

	PipelineType reader.PipelineType
	Pattern      string
	Debug        bool

	ShellCheckArgs []string

	OutputDirectory string
}
