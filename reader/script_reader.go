package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"path/filepath"
	"strings"
)

type PipelineType string

const (
	PipelineTypeGitlab PipelineType = "gitlab"
)

func NewDecoder(pipelineType PipelineType) ScriptDecoder {
	switch pipelineType {
	case PipelineTypeGitlab:
		return newGitlabDecoder()
	}

	panic(fmt.Sprintf("unknown pipeline type: %s", pipelineType))
}

type DocumentAnchorMap map[string]ast.Node
type ScriptParser func(node ast.Node, anchorMap map[string]ast.Node) string
type ScriptTransformer func(script string) string

type ScriptReader interface {
	readScriptsForAst(file *ast.File) ([]ScriptBlock, error)
}

type ScriptDecoder struct {
	ScriptReader

	parser      ScriptParser
	transformer ScriptTransformer
}

type ScriptBlock struct {
	FileName  string
	BlockName string
	Script    string
	Shell     string
}

func (d ScriptDecoder) DecodeFile(file string) ([]ScriptBlock, error) {
	if astFile, err := readFile(file); err != nil {
		return nil, err
	} else {
		scriptBlocks := make([]ScriptBlock, 0)
		readerScripts, err := d.readScriptsForAst(astFile)

		directiveDecoder := NewScriptCheckDirectiveReader(d)
		directiveScripts, err := directiveDecoder.readScriptsForAst(astFile)

		if err != nil {
			return nil, err
		}

		scriptBlocks = append(scriptBlocks, directiveScripts...)
		scriptBlocks = append(scriptBlocks, readerScripts...)

		return scriptBlocks, nil
	}
}

func (script ScriptBlock) GetOutputFileName(parentDir string) string {
	nameWithoutExtension := script.FileName[:len(script.FileName)-len(filepath.Ext(script.FileName))]
	transformedFileName := strings.ReplaceAll(nameWithoutExtension, string(filepath.Separator), "-")
	return parentDir + "/" + transformedFileName + "-" + script.BlockName + ".sh"
}

func readFile(file string) (*ast.File, error) {
	astFile, err := parser.ParseFile(file, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file %s: %w", file, err)
	}

	return astFile, nil
}
