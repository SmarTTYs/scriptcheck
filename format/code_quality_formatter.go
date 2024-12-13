package format

import (
	"encoding/json"
	"github.com/google/uuid"
	"scriptcheck/reader"
	"strconv"
)

type CodeQualityReportFormatter struct{}

type codeClimateReport struct {
	Description string `json:"description"`
	CheckName   string `json:"check_name"`
	Fingerprint string `json:"fingerprint"`
	Severity    string `json:"severity"`
	Location    struct {
		Path  string `json:"path"`
		Lines struct {
			Begin int `json:"begin"`
		} `json:"lines"`
	} `json:"location"`
}

func (f *CodeQualityReportFormatter) Format(report ShellCheckReport, scriptMap map[string]reader.ScriptBlock) (string, error) {
	codeClimateReports := make([]codeClimateReport, 0)
	for _, report := range report {
		scriptBlock := scriptMap[report.File]
		var offset int
		if scriptBlock.HasShell {
			offset = 0
		} else {
			offset = 1
		}

		codeClimateReport := codeClimateReport{}
		codeClimateReport.Description = report.Message
		codeClimateReport.CheckName = strconv.Itoa(report.Code)
		codeClimateReport.Fingerprint = uuid.New().String()
		codeClimateReport.Location.Path = scriptBlock.FileName
		codeClimateReport.Location.Lines.Begin = scriptBlock.StartPos + report.Line - offset
		codeClimateReport.Severity = severityFromShellcheck(report.Level)

		codeClimateReports = append(codeClimateReports, codeClimateReport)
	}

	marshal, err := json.Marshal(codeClimateReports)
	return string(marshal), err
}

func severityFromShellcheck(shellCheckSeverity string) string {
	switch shellCheckSeverity {
	case "error":
		return "major"
	case "warning":
		return "minor"
	case "info":
		return "info"
	case "style":
		return "minor"
	default:
		return "critical"
	}
}
