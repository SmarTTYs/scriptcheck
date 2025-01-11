package reader

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Script string

type ScriptBlock struct {
	FileName  string
	BlockName string
	Script    Script
	Shell     string
	Path      string

	// todo: currently does not work as expected
	//  as positional information seem to be incorrect
	//  in some cases
	Column   int
	StartPos int
}

func (script ScriptBlock) ScriptString() string {
	builder := new(strings.Builder)
	if !script.Script.HasShell() {
		var scriptShell string
		if len(script.Shell) > 0 {
			scriptShell = script.Shell
		} else {
			scriptShell = "sh"
		}

		builder.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", scriptShell))
	}

	builder.WriteString(string(script.Script))
	return builder.String()
}

func (s Script) HasShell() bool {
	return strings.HasPrefix(string(s), "#!")
}

func (script ScriptBlock) OutputFileName() string {
	sBuilder := new(strings.Builder)
	extension := filepath.Ext(script.FileName)
	sBuilder.WriteString(script.FileName[:len(script.FileName)-len(extension)])
	sBuilder.WriteRune('-')
	sBuilder.WriteString(script.BlockName)
	sBuilder.WriteString(".sh")

	return sBuilder.String()
}
