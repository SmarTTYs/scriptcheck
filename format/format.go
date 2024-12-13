package format

import (
	"encoding/json"
	"fmt"
	"scriptcheck/reader"
)

type Format string

const (
	StandardFormat    Format = "standard"
	CodeQualityFormat Format = "code_quality"
)

type ShellCheckReportFormatter interface {
	Format(report []byte, scriptMap map[string]reader.ScriptBlock) (string, error)
}

type shellCheckReport []reportEntry

type reportEntry struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	EndLine   int    `json:"endLine"`
	Column    int    `json:"column"`
	EndColumn int    `json:"endColumn"`
	Level     string `json:"level"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

func shellCheckReportFromString(bytes []byte) (shellCheckReport, error) {
	var report shellCheckReport
	err := json.Unmarshal(bytes, &report)

	if err != nil {
		return nil, fmt.Errorf("unable to parse shellcheck output: %w", err)
	}

	return report, nil
}

type StandardReportFormatter struct{}

func (f *StandardReportFormatter) Format(reportString []byte, _ map[string]reader.ScriptBlock) (string, error) {
	return string(reportString), nil
}

func NewFormatter(format Format) ShellCheckReportFormatter {
	switch format {
	case CodeQualityFormat:
		return &CodeQualityReportFormatter{}
	case StandardFormat:
		return &StandardReportFormatter{}
	}

	panic("")
}
