package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"log"
)

type PipelineType string

const (
	PipelineTypeGitlab PipelineType = "gitlab"
)

func NewDecoder(pipelineType PipelineType, debug bool) ScriptDecoder {
	switch pipelineType {
	case PipelineTypeGitlab:
		return NewGitlabDecoder(debug)
	}

	panic(fmt.Sprintf("unknown pipeline type: %s", pipelineType))
}

type documentAnchorMap map[string]ast.Node
type scriptParser func(document *ast.DocumentNode, node ast.Node, anchorMap map[string]ast.Node) string
type scriptTransformer func(script string) string

type ScriptReader interface {
	readScriptsForAst(file *ast.File) ([]ScriptBlock, error)
}

type ScriptDecoder struct {
	ScriptReader

	debug       bool
	parser      scriptParser
	transformer scriptTransformer
}

func NewScriptBlock(file, key, jobName, script string, node ast.Node) ScriptBlock {
	line := readLineFromNode(node)
	// pos := readPositionFromNode(node)
	return ScriptBlock{
		FileName:  file,
		BlockName: jobName + "_" + key,
		Script:    Script(script),
		Path:      node.GetPath(),

		// Column:   position.Column,
		StartPos: line,
	}
}

func (d ScriptDecoder) DecodeFile(file string) ([]ScriptBlock, error) {
	if astFile, err := readFile(file); err != nil {
		return nil, err
	} else {
		return d.decodeAstFile(astFile)
	}
}

func (d ScriptDecoder) MergeAndDecode(files []string) ([]ScriptBlock, error) {
	if mergedFile, err := mergeFiles(files); err != nil {
		return nil, err
	} else {
		return d.decodeAstFile(mergedFile)
	}
}

func (d ScriptDecoder) decodeAstFile(astFile *ast.File) ([]ScriptBlock, error) {
	scriptBlocks := make([]ScriptBlock, 0)
	readerScripts, err := d.readScriptsForAst(astFile)
	if d.debug {
		log.Printf("Extracted %d script(s) from file '%s'\n", len(readerScripts), astFile.Name)
	}

	directiveDecoder := newScriptCheckDirectiveDecoder(d)
	directiveScripts, err := directiveDecoder.readScriptsForAst(astFile)
	if d.debug {
		log.Printf("Extracted %d script(s) from directives for file '%s'\n", len(readerScripts), astFile.Name)
	}

	if err != nil {
		return nil, err
	}

	scriptBlocks = append(scriptBlocks, directiveScripts...)
	scriptBlocks = append(scriptBlocks, readerScripts...)

	return scriptBlocks, nil
}

func readFile(file string) (*ast.File, error) {
	astFile, err := parser.ParseFile(file, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file %s: %w", file, err)
	}

	return astFile, nil
}

func mergeFiles(files []string) (*ast.File, error) {
	var mergedNode *ast.DocumentNode
	for index, file := range files {
		astFile, err := readFile(file)
		if err != nil {
			return nil, err
		}

		if index == 0 {
			mergedNode = astFile.Docs[0]
			err = mergeDocsIntoDst(mergedNode, astFile.Docs[1:])
		} else {
			err = mergeDocsIntoDst(mergedNode, astFile.Docs)
		}

		if err != nil {
			return nil, err
		}
	}

	mergedFile := ast.File{
		Name: "merged_pipeline_yaml.yml",
		Docs: []*ast.DocumentNode{mergedNode},
	}

	return &mergedFile, nil
}

func mergeDocsIntoDst(dst *ast.DocumentNode, docs []*ast.DocumentNode) error {
	for _, document := range docs {
		err := ast.Merge(dst, document)
		if err != nil {
			return err
		}
	}

	return nil
}
