package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"regexp"
	"slices"
	"strings"
)

// jobs names prefixed with a dot get ignored by gitlab ci
const gitlabJobIgnoreMarker = "."

// regular expression to find gitlab input references
var jobInputRegex = regexp.MustCompile("\\$\\[\\[(\\s*inputs[^]]+)]]")

// sections that can contain scripts
var sections = []string{
	"script",
	"before_script",
	"after_script",
}

func newGitlabDecoder() ScriptDecoder {
	decoder := ScriptDecoder{
		gitlabScriptReader{
			anchorNodeMap: make(DocumentAnchorMap),
		},
		readScriptFromNode,
		replaceJobInputReference,
	}

	return decoder
}

type gitlabScriptReader struct {
	ScriptReader
	anchorNodeMap DocumentAnchorMap
}

func (r gitlabScriptReader) readScriptsForAst(file *ast.File) ([]ScriptBlock, error) {
	var documentToRead *ast.DocumentNode
	if len(file.Docs) > 1 {
		documentToRead = file.Docs[1]
	} else {
		documentToRead = file.Docs[0]
	}

	// special support for anchor types
	for _, n := range ast.Filter(ast.AnchorType, documentToRead) {
		anchor := n.(*ast.AnchorNode)
		anchorName := anchor.Name.GetToken().Value
		r.anchorNodeMap[anchorName] = anchor.Value
	}

	// read script blocks from given document
	return r.readFromDocument(file.Name, documentToRead)
}

func (r gitlabScriptReader) readFromDocument(fileName string, doc *ast.DocumentNode) ([]ScriptBlock, error) {
	documentScripts := make([]ScriptBlock, 0)
	switch body := doc.Body.(type) {
	case *ast.MappingNode:
		for _, vNode := range body.Values {
			if scripts := r.readScriptsFromMappingNode(fileName, vNode); scripts != nil {
				documentScripts = append(documentScripts, scripts...)
			}
		}
	case *ast.MappingValueNode:
		if scripts := r.readScriptsFromMappingNode(fileName, body); scripts != nil {
			documentScripts = append(documentScripts, scripts...)
		}
	}

	return documentScripts, nil
}

func (r gitlabScriptReader) readScriptsFromMappingNode(fileName string, mappingValueNode *ast.MappingValueNode) []ScriptBlock {
	vNode := mappingValueNode.Value
	if vNode.Type() == ast.MappingType {
		jobName := mappingValueNode.Key.String()
		if strings.HasPrefix(jobName, gitlabJobIgnoreMarker) {
			return nil
		}

		return r.readScriptsFromJob(fileName, jobName, vNode.(*ast.MappingNode))
	}

	return nil
}

func (r gitlabScriptReader) readScriptsFromJob(file, jobName string, node *ast.MappingNode) []ScriptBlock {
	scripts := make([]ScriptBlock, 0)
	for _, element := range node.Values {
		eKey := element.Key.String()
		eValue := element.Value
		if slices.Contains(sections, eKey) {
			script := readScriptFromNode(eValue, r.anchorNodeMap)
			script = replaceJobInputReference(script)

			scriptBlock := ScriptBlock{
				FileName:  file,
				BlockName: jobName + "_" + eKey,
				Script:    script,
			}

			if directive := scriptDirectiveFromComment(element.GetComment()); directive != nil {
				scriptBlock.Shell = directive.ShellDirective()
			}

			scripts = append(scripts, scriptBlock)
		}
	}

	return scripts
}

func replaceJobInputReference(script string) string {
	for _, match := range jobInputRegex.FindAllString(script, -1) {
		replaced := strings.TrimPrefix(match, "$")
		replaced = strings.TrimFunc(replaced, func(r rune) bool {
			return r == '[' || r == ']' || r == ' '
		})
		replaced = strings.ToUpper(replaced)
		script = strings.Replace(script, match, "$"+replaced, -1)
	}

	return script
}

func readScriptFromNode(node ast.Node, anchorNodeMap map[string]ast.Node) string {
	switch vType := node.(type) {
	case *ast.TagNode:
		return readScriptFromNode(vType.Value, anchorNodeMap)
	case *ast.AnchorNode:
		return readScriptFromNode(vType.Value, anchorNodeMap)
	case *ast.AliasNode:
		aliasName := vType.Value.GetToken().Value
		anchorValue, exists := anchorNodeMap[aliasName]
		if !exists {
			panic(fmt.Sprintf("anchor %s not found!", aliasName))
		}

		return readScriptFromNode(anchorValue, anchorNodeMap)
	case *ast.StringNode:
		return vType.Value
	case *ast.LiteralNode:
		return vType.Value.Value
	case *ast.SequenceNode:
		sb := new(strings.Builder)
		for _, listElement := range vType.Values {
			sb.WriteString(readScriptFromNode(listElement, anchorNodeMap))
			sb.WriteString("\n")
		}
		return sb.String()
	default:
		return ""
	}
}
