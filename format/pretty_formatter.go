package format

import (
	"bufio"
	"fmt"
	"scriptcheck/color"
	"scriptcheck/report"
	"strings"
)

type PrettyFormatter struct {
}

func (f *PrettyFormatter) Format(reports []report.ScriptCheckReport) (string, error) {
	builder := new(strings.Builder)

	fileReports := make(map[string]map[int][]report.ScriptCheckReport)
	for _, scriptReport := range reports {
		if lineMap, exists := fileReports[scriptReport.File]; !exists {
			lineMap := make(map[int][]report.ScriptCheckReport)
			lineMap[scriptReport.Line] = append(lineMap[scriptReport.Line], scriptReport)
			fileReports[scriptReport.File] = lineMap
		} else {
			lineMap[scriptReport.Line] = append(lineMap[scriptReport.Line], scriptReport)
		}
	}

	for file, reports := range fileReports {
		for line, lineReports := range reports {
			f.appendGroupedReport(builder, file, line, lineReports)
		}
	}

	/*
		builder.WriteString("\n TEST CURRENT \n")
		for _, report := range reports {
			f.appendReport(builder, report)
		}
	*/

	return builder.String(), nil
}

func (f *PrettyFormatter) appendGroupedReport(builder *strings.Builder, file string, line int, reports []report.ScriptCheckReport) {
	builder.WriteString(color.Color(fmt.Sprintf("In %s line %d:", file, line), color.Bold))
	builder.WriteString("\n")
	builder.WriteString(f.getLine(reports[0]) + "\n")

	informationList := make([]string, 0)
	for _, scriptReport := range reports {
		builder.WriteString(f.formatReportLine(scriptReport))
		informationList = append(
			informationList,
			fmt.Sprintf("https://www.shellcheck.net/wiki/%s -- %s", scriptReport.Reason, scriptReport.Message),
		)
	}

	builder.WriteString("For more information:\n")
	for _, line := range informationList {
		builder.WriteString(" " + line + "\n\n")
	}
}

/*
func (f *PrettyFormatter) appendReport(builder *strings.Builder, report ScriptCheckReport) {
	builder.WriteString(Color(fmt.Sprintf("In %s line %d", report.File, report.Line), Bold))
	builder.WriteString("\n")
	builder.WriteString(f.getLine(report) + "\n")
	builder.WriteString(f.formatReportLine(report))
	builder.WriteString("\n")
}
*/

func (f *PrettyFormatter) formatReportLine(report report.ScriptCheckReport) string {
	levelColor := f.getLevelColor(report.Level)

	prefix := strings.Repeat(" ", report.Column-1)
	var marker string
	if report.EndColumn-report.Column > 2 {
		marker = "^" + strings.Repeat("-", report.EndColumn-report.Column-2) + "^"
	} else {
		marker = "^--"
	}

	return color.Color(fmt.Sprintf("%s%s %s (%s): %s\n", prefix, marker, report.Reason, report.Level, report.Message), levelColor)
}

func (f *PrettyFormatter) getLevelColor(level string) string {
	var levelColor string
	switch level {
	case "info":
		levelColor = color.Green
	case "warning":
		levelColor = color.Yellow
	case "error":
		levelColor = color.Red
	case "style":
		levelColor = color.Blue
	default:
		levelColor = color.Red
	}

	return levelColor
}

func (*PrettyFormatter) getLine(report report.ScriptCheckReport) string {
	lineScanner := bufio.NewScanner(strings.NewReader(report.Script.ScriptString()))
	line := 0
	for lineScanner.Scan() {
		line++
		if line == report.Report.Line {
			return lineScanner.Text()
		}
	}

	return ""
}
