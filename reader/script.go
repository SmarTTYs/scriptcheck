package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"path/filepath"
	"strings"
)

type Script string

func NewScriptBlock(
	file, blockName, defaultShell string,
	script scriptNode,
	node ast.Node,
	directive *ScriptDirective,
) ScriptBlock {
	block := ScriptBlock{
		FileName:  file,
		BlockName: blockName,
		Script:    script.script,
		Path:      node.GetPath(),
		Shell:     defaultShell,
		directive: directive,

		// Column:   position.Column,
		StartPos: script.line,
	}

	if directive != nil {
		if directiveShell := directive.ShellDirective(); directiveShell != "" {
			block.Shell = directiveShell
		}
	}

	return block
}

type ScriptBlock struct {
	FileName  string
	BlockName string
	Script    Script
	Shell     string
	Path      string

	directive *ScriptDirective

	// todo: currently does not work as expected
	//  as positional information seem to be incorrect
	//  in some cases
	Column   int
	StartPos int
}

func (d ScriptDirective) asShellcheckDirective(script ScriptBlock) string {
	directiveBuilder := new(strings.Builder)
	directiveBuilder.WriteString("# shellcheck")

	if script.HasShell() {
		directiveBuilder.WriteString(fmt.Sprintf(" shell=%s", script.Shell))
	}

	if len(d.DisabledRules()) > 0 {
		rulesString := strings.Join(d.DisabledRules(), ",")
		directiveBuilder.WriteString(fmt.Sprintf(" disable=%s", rulesString))
	}
	directiveBuilder.WriteString("\n")

	return directiveBuilder.String()
}

func (script ScriptBlock) ScriptString() string {
	builder := new(strings.Builder)

	if script.directive != nil {
		builder.WriteString(script.directive.asShellcheckDirective(script))
	} else if script.Shell != "" {
		builder.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", script.Shell))
	}

	builder.WriteString(string(script.Script))
	return builder.String()
}

func (script ScriptBlock) HasShell() bool {
	return len(script.Shell) > 0
}

func (script ScriptBlock) HasShellDirective() bool {
	return script.directive != nil
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
