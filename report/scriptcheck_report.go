package report

import (
	"encoding/json"
	"fmt"
	"scriptcheck/reader"
	"strconv"
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
	// currently ignored for json format as we
	// do not support 100% correct columns in
	// the checked yaml file
	Column    int `json:"-"`
	EndColumn int `json:"-"`

	// shellcheck reason (code prefixed by SC)
	Reason string `json:"reason"`

	// shellcheck message
	Message string `json:"message"`

	// original shellcheck report
	Report ShellcheckReport `json:"-"`

	// script for which a violation was found
	Script reader.ScriptBlock `json:"-"`
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

func newScriptCheckReport(reports []ShellcheckReport, scriptMap map[string]reader.ScriptBlock) []ScriptCheckReport {
	scriptCheckReports := make([]ScriptCheckReport, 0)
	for _, report := range reports {
		scriptBlock := scriptMap[report.File]

		var offset = 0

		// when the scriptblock defined a shell directive we need
		// to subtract one line from the reported line
		if scriptBlock.HasShell() || scriptBlock.HasShellDirective() {
			offset = 1
		}

		reason := "SC" + strconv.Itoa(report.Code)
		// the report starts at 1 so we need to subtract one in order
		// to get the correct line inside the yaml file
		reportLineBase := report.Line - 1
		scriptReport := ScriptCheckReport{
			File:    scriptBlock.FileName,
			Report:  report,
			Level:   report.Level,
			Message: report.Message,

			Line:      scriptBlock.StartPos + reportLineBase - offset,
			Column:    report.Column,
			EndColumn: report.EndColumn,

			Path:   scriptBlock.Path,
			Reason: reason,
			Script: scriptBlock,
		}
		scriptCheckReports = append(scriptCheckReports, scriptReport)
	}

	return scriptCheckReports
}

func shellCheckReportFromString(bytes []byte) ([]ShellcheckReport, error) {
	var report []ShellcheckReport
	err := json.Unmarshal(bytes, &report)

	if err != nil {
		return nil, fmt.Errorf("unable to parse shellcheck output: %w", err)
	}

	return report, nil
}
