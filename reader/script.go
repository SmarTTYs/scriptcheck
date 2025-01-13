package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"path/filepath"
	"strings"
)

type Script string

func NewScriptBlock(file, blockName, defaultShell string, script Script, node ast.Node) ScriptBlock {
	line := readLineFromNode(node)
	// pos := readPositionFromNode(node)
	return ScriptBlock{
		FileName:  file,
		BlockName: blockName,
		Script:    script,
		Path:      node.GetPath(),
		Shell:     defaultShell,

		// Column:   position.Column,
		StartPos: line,
	}
}

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
	if script.HasShell() && !script.Script.hasShell() {
		builder.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", script.Shell))
	}

	builder.WriteString(string(script.Script))
	return builder.String()
}

func (script ScriptBlock) HasShell() bool {
	return len(script.Shell) > 0
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
