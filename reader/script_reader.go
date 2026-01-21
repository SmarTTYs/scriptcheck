package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"log"
	"scriptcheck/color"
	"slices"
)

type PipelineType string

const (
	PipelineTypeGitlab PipelineType = "gitlab"
)

func NewDecoder(pipelineType PipelineType, debug bool, defaultShell string, experimentalFolding bool) ScriptDecoder {
	switch pipelineType {
	case PipelineTypeGitlab:
		return newGitlabDecoder(debug, defaultShell, experimentalFolding)
	}

	panic(fmt.Sprintf("unknown pipeline type: %s", pipelineType))
}

type aliasValueMap map[*ast.AliasNode]ast.Node
type scriptParser func(
	document *ast.DocumentNode,
	node ast.Node,
	aliasValueMap aliasValueMap,
	experimentalFolding bool,
) []ScriptNode

type ScriptNode struct {
	Script Script
	Line   int
}

type ScriptReader interface {
	readScriptsForAst(file *ast.File, aliasValueMap aliasValueMap) ([]ScriptBlock, error)
}

type ScriptDecoder struct {
	ScriptReader

	experimentalFolding bool
	defaultShell        string
	debug               bool

	parser scriptParser
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

	anchorWalker := &anchorWalker{
		anchorNodeMap: make(map[string]ast.Node),
		aliasValueMap: make(aliasValueMap),
	}

	// otherwise the current filter walker fails as body
	// will be null for empty yaml files
	for _, doc := range astFile.Docs {
		if doc.Body != nil {
			ast.Walk(anchorWalker, doc.Body)
		}
	}

	readerScripts, err := d.readScriptsForAst(astFile, anchorWalker.aliasValueMap)
	if d.debug {
		log.Printf(
			"Extracted %s script(s) from file '%s'\n",
			color.Color(len(readerScripts), color.Bold),
			color.Color(astFile.Name, color.Bold),
		)
	}

	directiveDecoder := newScriptCheckDirectiveDecoder(d)
	directiveScripts, err := directiveDecoder.readScriptsForAst(astFile, anchorWalker.aliasValueMap)

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

	scriptBlocks = append(scriptBlocks, readerScripts...)
	for _, directiveScript := range directiveScripts {
		contains := slices.ContainsFunc(scriptBlocks, func(block ScriptBlock) bool {
			return directiveScript.FileName == block.FileName && directiveScript.Path == block.Path
		})

		if !contains {
			scriptBlocks = append(scriptBlocks, directiveScript)
		}
	}

	return scriptBlocks, nil
}

func readFile(file string) (*ast.File, error) {
	astFile, err := parser.ParseFile(file, parser.ParseComments, parser.AllowDuplicateMapKey())
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
