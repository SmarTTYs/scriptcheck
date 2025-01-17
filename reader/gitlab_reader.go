package reader

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"log"
	"regexp"
	"slices"
	"strings"
)

// jobs names prefixed with a dot get ignored by gitlab ci
const gitlabJobIgnoreMarker = "."
const gitlabReferenceTag = "!reference"

// regular expression to find gitlab input references
var jobInputRegex = regexp.MustCompile("\\$\\[\\[(\\s*inputs[^]]+)]]")

// sections that can contain scripts
var sections = []string{
	"script",
	"before_script",
	"after_script",
}

func newGitlabDecoder(debug bool, defaultShell string) ScriptDecoder {
	decoder := ScriptDecoder{
		ScriptReader: gitlabScriptReader{
			defaultShell:  defaultShell,
			anchorNodeMap: make(documentAnchorMap),
		},
		defaultShell: defaultShell,
		debug:        debug,
		parser:       readScriptFromNode,
		transformer:  replaceJobInputReference,
	}

	return decoder
}

type gitlabScriptReader struct {
	ScriptReader

	defaultShell string

	// currently looped document
	document      *ast.DocumentNode
	anchorNodeMap documentAnchorMap
}

func (r gitlabScriptReader) readScriptsForAst(file *ast.File) ([]ScriptBlock, error) {
	if len(file.Docs) > 1 {
		r.document = file.Docs[1]
	} else {
		r.document = file.Docs[0]
	}

	// otherwise the current filter walker fails as body
	// will be null for empty yaml files
	if r.document.Body != nil {
		for _, n := range ast.Filter(ast.AnchorType, r.document) {
			anchor := n.(*ast.AnchorNode)
			anchorName := anchor.Name.GetToken().Value
			r.anchorNodeMap[anchorName] = anchor.Value
		}
	}

	// read script blocks from given document
	return r.readFromDocument(file.Name)
}

func (r gitlabScriptReader) readFromDocument(fileName string) ([]ScriptBlock, error) {
	documentScripts := make([]ScriptBlock, 0)
	switch body := r.document.Body.(type) {
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
			script := readScriptFromNode(r.document, eValue, r.anchorNodeMap)
			script = replaceJobInputReference(script)

			blockName := jobName + "_" + eKey
			scriptBlock := NewScriptBlock(
				file,
				blockName,
				r.defaultShell,
				script,
				eValue,
			)

			if directive := scriptDirectiveFromComment(element.GetComment()); directive != nil {
				if directiveShell := directive.ShellDirective(); directiveShell != "" {
					scriptBlock.Shell = directiveShell
				}
			}

			scripts = append(scripts, scriptBlock)
		}
	}

	return scripts
}

func replaceJobInputReference(script Script) Script {
	transformedString := string(script)
	for _, match := range jobInputRegex.FindAllString(transformedString, -1) {
		replaced := strings.TrimPrefix(match, "$")
		replaced = strings.TrimFunc(replaced, func(r rune) bool {
			return r == '[' || r == ']' || r == ' '
		})
		replaced = strings.ToUpper(replaced)
		transformedString = strings.Replace(transformedString, match, "${"+replaced+"}", 1)
	}

	return Script(transformedString)
}

func readLineFromNode(node ast.Node) int {
	switch vType := node.(type) {
	case *ast.LiteralNode:
		return vType.GetToken().Position.Line + 1
	default:
		return node.GetToken().Position.Line
	}
}

func readScriptFromNode(document *ast.DocumentNode, node ast.Node, anchorNodeMap map[string]ast.Node) Script {
	switch vType := node.(type) {
	case *ast.TagNode:
		if vType.Start.Value == gitlabReferenceTag {
			return readScriptFromReference(document, vType, anchorNodeMap)
		} else {
			log.Println("Unknown reference type")
			return ""
		}
	case *ast.AnchorNode:
		return readScriptFromNode(document, vType.Value, anchorNodeMap)
	case *ast.AliasNode:
		aliasName := vType.Value.GetToken().Value
		if anchorValue, exists := anchorNodeMap[aliasName]; !exists {
			panic(fmt.Sprintf("anchor %s not found!", aliasName))
		} else {
			return readScriptFromNode(document, anchorValue, anchorNodeMap)
		}
	case *ast.StringNode:
		return Script(vType.Value)
	case *ast.LiteralNode:
		println(vType.Value.Value)
		println(vType.Value.GetToken().Value)
		return Script(vType.Value.Value)
	case *ast.SequenceNode:
		sb := new(strings.Builder)
		for _, listElement := range vType.Values {
			sb.WriteString(string(readScriptFromNode(document, listElement, anchorNodeMap)))
			sb.WriteString("\n")
		}
		return Script(sb.String())
	default:
		return ""
	}
}

func pathFromSequence(node *ast.SequenceNode) *yaml.Path {
	pathBuilder := (&yaml.PathBuilder{}).Root()
	for _, pathValue := range node.Values {
		pathBuilder.Child(pathValue.String())
	}

	return pathBuilder.Build()
}

func readScriptFromReference(document *ast.DocumentNode, tag *ast.TagNode, anchorNodeMap map[string]ast.Node) Script {
	pathValues := tag.Value.(*ast.SequenceNode)
	pathString := pathFromSequence(pathValues)

	pathNode, err := pathString.FilterNode(document.Body)
	if err != nil {
		return ""
	}

	return readScriptFromNode(document, pathNode, anchorNodeMap)
}
