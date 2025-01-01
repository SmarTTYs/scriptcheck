package format

import (
	"encoding/json"
	"fmt"
	"scriptcheck/reader"
	"strconv"
)

type Format string

const (
	JsonFormat        Format = "json"
	StandardFormat    Format = "standard"
	CodeQualityFormat Format = "code_quality"
)

type ScriptCheckReport struct {
	// name of the yaml file
	File string `json:"file"`

	// path inside the yaml file
	Path string `json:"path"`

	// shellcheck level of the report
	Level string `json:"level"`

	// line of the found violation inside yaml file
	Line int `json:"line"`

	// column where the violation was found
	Column    int `json:"column"`
	EndColumn int `json:"endColumn"`

	// shellcheck reason (code prefixed by SC)
	Reason string `json:"reason"`

	// shellcheck message
	Message string `json:"message"`

	// original shellcheck report
	report shellcheckReport

	// script for which a violation was found
	script reader.ScriptBlock
}

type shellcheckReport struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	EndLine   int    `json:"endLine"`
	Column    int    `json:"column"`
	EndColumn int    `json:"endColumn"`
	Level     string `json:"level"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

type ShellCheckReportFormatter interface {
	Format(reports []ScriptCheckReport) (string, error)
}

func NewScriptCheckReport(
	reportBytes []byte,
	scriptMap map[string]reader.ScriptBlock,
) ([]ScriptCheckReport, error) {
	if shellCheckReport, err := shellCheckReportFromString(reportBytes); err != nil {
		return nil, fmt.Errorf("unable to parse shellcheck report: %w", err)
	} else {
		return newScriptCheckReport(shellCheckReport, scriptMap), nil
	}
}

func newScriptCheckReport(reports []shellcheckReport, scriptMap map[string]reader.ScriptBlock) []ScriptCheckReport {
	scriptCheckReports := make([]ScriptCheckReport, 0)
	for _, report := range reports {
		scriptBlock := scriptMap[report.File]
		var offset int
		if scriptBlock.HasShell {
			offset = 0
		} else {
			offset = 1
		}

		reason := "SC" + strconv.Itoa(report.Code)
		scriptReport := ScriptCheckReport{
			File:    scriptBlock.FileName,
			report:  report,
			Level:   report.Level,
			Message: report.Message,

			Line:      scriptBlock.StartPos + report.Line - offset,
			Column:    report.Column,
			EndColumn: report.EndColumn,

			Path:   scriptBlock.Path,
			Reason: reason,
			script: scriptBlock,
		}
		scriptCheckReports = append(scriptCheckReports, scriptReport)
	}

	return scriptCheckReports
}

func shellCheckReportFromString(bytes []byte) ([]shellcheckReport, error) {
	var report []shellcheckReport
	err := json.Unmarshal(bytes, &report)

	if err != nil {
		return nil, fmt.Errorf("unable to parse shellcheck output: %w", err)
	}

	return report, nil
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
