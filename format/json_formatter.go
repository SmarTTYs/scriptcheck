package format

import (
	"encoding/json"
	"scriptcheck/report"
)

type JsonFormatter struct{}

func (f *JsonFormatter) Format(reports []report.ScriptCheckReport) (string, error) {
	if bytes, err := json.Marshal(reports); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}
