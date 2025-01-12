package format

import (
	"fmt"
	"scriptcheck/report"
)

type Format string

const (
	JsonFormat        Format = "json"
	StandardFormat    Format = "standard"
	CodeQualityFormat Format = "code_quality"
)

type ShellCheckReportFormatter interface {
	Format(reports []report.ScriptCheckReport) (string, error)
}

func NewFormatter(format Format) ShellCheckReportFormatter {
	switch format {
	case CodeQualityFormat:
		return &CodeQualityReportFormatter{}
	case JsonFormat:
		return &JsonFormatter{}
	case StandardFormat:
		return &PrettyFormatter{}
	}

	panic(fmt.Sprintf("Unknown format %s", format))
}
