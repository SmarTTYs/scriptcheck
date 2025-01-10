package runtime

import (
	"scriptcheck/format"
	"scriptcheck/reader"
)

func NewOptions() *Options {
	return &Options{}
}

type Options struct {
	OutputFile string

	PipelineType reader.PipelineType
	Debug        bool
	Merge        bool

	Strict bool

	ShellCheckArgs []string
	Format         format.Format

	OutputDirectory string
}
