package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"log"
	"scriptcheck/color"
)

type PipelineType string

const (
	PipelineTypeGitlab PipelineType = "gitlab"
)

func NewDecoder(pipelineType PipelineType, debug bool, defaultShell string) ScriptDecoder {
	switch pipelineType {
	case PipelineTypeGitlab:
		return newGitlabDecoder(debug, defaultShell)
	}

	panic(fmt.Sprintf("unknown pipeline type: %s", pipelineType))
}

type documentAnchorMap map[string]ast.Node
type scriptParser func(document *ast.DocumentNode, node ast.Node, anchorMap map[string]ast.Node) Script
type scriptTransformer func(script Script) Script

type ScriptReader interface {
	readScriptsForAst(file *ast.File) ([]ScriptBlock, error)
}

type ScriptDecoder struct {
	ScriptReader

	defaultShell string
	debug        bool

	parser      scriptParser
	transformer scriptTransformer
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
		log.Printf(
			"Extracted %s script(s) from file '%s'\n",
			color.Color(len(readerScripts), color.Bold),
			color.Color(astFile.Name, color.Bold),
		)
	}

	directiveDecoder := newScriptCheckDirectiveDecoder(d)
	directiveScripts, err := directiveDecoder.readScriptsForAst(astFile)
	if d.debug {
		log.Printf(
			"Extracted %s script(s) from directives for file '%s'\n",
			color.Color(len(directiveScripts), color.Bold),
			color.Color(astFile.Name, color.Bold),
		)
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
