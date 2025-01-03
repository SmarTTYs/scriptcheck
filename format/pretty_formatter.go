package format

import (
	"bufio"
	"fmt"
	"strings"
)

type PrettyFormatter struct {
}

func (f *PrettyFormatter) Format(reports []ScriptCheckReport) (string, error) {
	builder := new(strings.Builder)

	fileReports := make(map[string]map[int][]ScriptCheckReport)
	for _, report := range reports {
		if lineMap, exists := fileReports[report.File]; !exists {
			lineMap := make(map[int][]ScriptCheckReport)
			lineMap[report.Line] = append(lineMap[report.Line], report)
			fileReports[report.File] = lineMap
		} else {
			lineMap[report.Line] = append(lineMap[report.Line], report)
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

func (f *PrettyFormatter) appendGroupedReport(builder *strings.Builder, file string, line int, reports []ScriptCheckReport) {
	builder.WriteString(Color(fmt.Sprintf("In %s line %d:", file, line), Bold))
	builder.WriteString("\n")
	builder.WriteString(f.getLine(reports[0]) + "\n")

	informationList := make([]string, 0)
	for _, report := range reports {
		builder.WriteString(f.formatReportLine(report))
		informationList = append(
			informationList,
			fmt.Sprintf("https://www.shellcheck.net/wiki/%s -- %s", report.Reason, report.Message),
		)
	}

	builder.WriteString("For more information:\n")
	for _, line := range informationList {
		builder.WriteString(" " + line + "\n")
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

func (f *PrettyFormatter) formatReportLine(report ScriptCheckReport) string {
	color := f.getLevelColor(report.Level)

	prefix := strings.Repeat(" ", report.Column-1)
	var marker string
	if report.EndColumn-report.Column > 2 {
		marker = "^" + strings.Repeat("-", report.EndColumn-report.Column-2) + "^"
	} else {
		marker = "^--"
	}

	return Color(fmt.Sprintf("%s%s %s (%s): %s\n", prefix, marker, report.Reason, report.Level, report.Message), color)
}

func (f *PrettyFormatter) getLevelColor(level string) string {
	var color string
	switch level {
	case "info":
		color = Green
	case "warning":
		color = Yellow
	case "error":
		color = Red
	case "style":
		color = Blue
	default:
		color = Red
	}

	return color
}

func (*PrettyFormatter) getLine(report ScriptCheckReport) string {
	lineScanner := bufio.NewScanner(strings.NewReader(report.script.ScriptString()))
	line := 0
	for lineScanner.Scan() {
		line++
		if line == report.report.Line {
			return lineScanner.Text()
		}
	}

	return ""
}
