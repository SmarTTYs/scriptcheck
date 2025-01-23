package format

import (
	"encoding/json"
	"fmt"
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
		reportLine := fmt.Sprintf("%s/%s", scriptReport.Reason, scriptReport.Message)

		codeClimateReport := codeClimateReport{}
		codeClimateReport.Description = reportLine
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

/*
func (f *CodeQualityReportFormatter) createFingerprint(violation string) string {
	hasher := sha256.New()
	hasher.Write([]byte(violation))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}
*/

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
