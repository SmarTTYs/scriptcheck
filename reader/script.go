package reader

import (
	"github.com/goccy/go-yaml/ast"
	"path/filepath"
	"strings"
)

type Script string

func NewScriptBlock(
	file, blockName, defaultShell string,
	script ScriptNode,
	node ast.Node,
) ScriptBlock {
	directive := script.NodeDirective
	block := ScriptBlock{
		FileName:  file,
		BlockName: blockName,
		Script:    script.Script,
		Path:      node.GetPath(),
		Shell:     defaultShell,
		directive: script.NodeDirective,

		// Column:   position.Column,
		StartPos: script.Line,
	}

	if directive != nil {
		if directiveShell := directive.ShellDirective(); directiveShell != "" {
			block.Shell = directiveShell
		} else {
			shellDirective := make(ScriptDirective)
			shellDirective["shell"] = block.Shell
			for k, v := range *directive {
				shellDirective[k] = v
			}
			block.directive = &shellDirective
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

func (script ScriptBlock) ScriptString() string {
	builder := new(strings.Builder)

	directive := script.directive
	if directive == nil && len(script.Shell) > 0 {
		shellDirective := make(ScriptDirective)
		shellDirective["shell"] = script.Shell
		directive = &shellDirective
	}

	if directive != nil {
		if shellcheckDirective := directive.asShellcheckDirective(script); shellcheckDirective != nil {
			builder.WriteString(*shellcheckDirective)
		}
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
