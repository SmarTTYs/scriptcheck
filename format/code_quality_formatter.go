package format

import (
	"encoding/json"
	"github.com/google/uuid"
	"scriptcheck/report"
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

func (f *CodeQualityReportFormatter) Format(reports []report.ScriptCheckReport) (string, error) {
	codeClimateReports := make([]codeClimateReport, 0)
	for _, scriptReport := range reports {
		codeClimateReport := codeClimateReport{}
		codeClimateReport.Description = scriptReport.Message
		codeClimateReport.CheckName = scriptReport.Reason
		codeClimateReport.Fingerprint = uuid.New().String()
		codeClimateReport.Location.Path = scriptReport.File
		codeClimateReport.Location.Lines.Begin = scriptReport.Line
		codeClimateReport.Severity = severityFromShellcheck(scriptReport.Level)

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
