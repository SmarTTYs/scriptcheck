package format

import "encoding/json"

type JsonFormatter struct{}

func (f *JsonFormatter) Format(reports []ScriptCheckReport) (string, error) {
	if bytes, err := json.Marshal(reports); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}
