package format

import (
	"encoding/json"
	"github.com/google/uuid"
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

func (f *CodeQualityReportFormatter) Format(reports []ScriptCheckReport) (string, error) {
	codeClimateReports := make([]codeClimateReport, 0)
	for _, report := range reports {
		codeClimateReport := codeClimateReport{}
		codeClimateReport.Description = report.Message
		codeClimateReport.CheckName = report.Reason
		codeClimateReport.Fingerprint = uuid.New().String()
		codeClimateReport.Location.Path = report.File
		codeClimateReport.Location.Lines.Begin = report.Line
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
