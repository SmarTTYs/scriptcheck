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

func (d ScriptDirective) shellcheckString(script ScriptBlock) string {
	directiveBuilder := new(strings.Builder)
	directiveBuilder.WriteString("# shellcheck")

	if script.HasShell() && !script.Script.hasShell() {
		directiveBuilder.WriteString(fmt.Sprintf(" shell=%s", script.Shell))
	}

	if len(d.DisabledRules()) > 0 {
		rulesString := strings.Join(script.directive.DisabledRules(), ",")
		directiveBuilder.WriteString(fmt.Sprintf(" disable=%s", rulesString))
	}
	directiveBuilder.WriteString("\n")

	return directiveBuilder.String()
}

func (script ScriptBlock) ScriptString() string {
	builder := new(strings.Builder)

	/*
		if script.HasShell() && !script.Script.hasShell() {
			builder.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", script.Shell))
		}
	*/

	if script.directive != nil {
		builder.WriteString(script.directive.shellcheckString(script))
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

func (s Script) hasShell() bool {
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
