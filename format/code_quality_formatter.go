package format

import (
	"encoding/json"
	"github.com/google/uuid"
	"scriptcheck/reader"
	"strconv"
)

type CodeQualityReportFormatter struct{}

type codeClimateLocation struct {
	Path  string           `json:"path"`
	Lines codeClimateLines `json:"lines"`
}

type codeClimateLines struct {
	Begin int `json:"begin"`
}

type codeClimateReport struct {
	Description string              `json:"description"`
	CheckName   string              `json:"check_name"`
	Fingerprint string              `json:"fingerprint"`
	Severity    string              `json:"severity"`
	Location    codeClimateLocation `json:"location"`
}

func (f *CodeQualityReportFormatter) Format(reportString []byte, scriptMap map[string]reader.ScriptBlock) (string, error) {
	report, _ := shellCheckReportFromString(reportString)
	reports := make([]codeClimateReport, 0)
	for _, issue := range report {
		scriptBlock := scriptMap[issue.File]
		var offset int
		if scriptBlock.HasShell {
			offset = 0
		} else {
			offset = 1
		}

		report := codeClimateReport{
			Description: issue.Message,
			CheckName:   strconv.Itoa(issue.Code),
			Fingerprint: uuid.New().String(),
			Location: codeClimateLocation{
				Path: scriptBlock.FileName + "#" + scriptBlock.Path,
				Lines: codeClimateLines{
					Begin: scriptBlock.StartPos + issue.Line - offset,
				},
			},
		}

		reports = append(reports, report)
	}

	marshal, err := json.Marshal(reports)
	return string(marshal), err
}
