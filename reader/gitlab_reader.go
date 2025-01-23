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
		parser:       readScriptsFromNode,
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
			blockName := jobName + "_" + eKey
			directive := scriptDirectiveFromComment(element.GetComment())
			for i, script := range readScriptsFromNode(r.document, eValue, r.anchorNodeMap) {
				var elementName string
				if i > 0 {
					elementName = blockName + fmt.Sprintf("_%d", i)
				} else {
					elementName = blockName
				}

				scriptBlock := NewScriptBlock(
					file,
					elementName,
					r.defaultShell,
					script,
					eValue,
				)

				if directive != nil {
					if directiveShell := directive.ShellDirective(); directiveShell != "" {
						scriptBlock.Shell = directiveShell
					}
				}

				scripts = append(scripts, scriptBlock)
			}
		}
	}

	return scripts
}

func readScriptsFromNode(document *ast.DocumentNode, node ast.Node, anchorNodeMap documentAnchorMap) []scriptNode {
	switch vType := node.(type) {
	case *ast.TagNode:
		if vType.Start.Value == gitlabReferenceTag {
			referencedNode := readNodeFromReference(document, vType)
			if referencedNode != nil {
				return readScriptsFromNode(document, *referencedNode, anchorNodeMap)
			} else {
				return nil
			}
		} else {
			log.Println("Unknown reference type")
			return nil
		}
	case *ast.AnchorNode:
		return readScriptsFromNode(document, vType.Value, anchorNodeMap)
	case *ast.AliasNode:
		aliasName := vType.Value.GetToken().Value
		if anchorValue, exists := anchorNodeMap[aliasName]; !exists {
			panic(fmt.Sprintf("anchor %s not found!", aliasName))
		} else {
			return readScriptsFromNode(document, anchorValue, anchorNodeMap)
		}
	case *ast.SequenceNode:
		elements := make([]scriptNode, 0)
		for _, listElement := range vType.Values {
			scripts := readScriptsFromNode(document, listElement, anchorNodeMap)
			elements = append(elements, scripts...)
		}
		return elements
	case *ast.LiteralNode:
		return readScriptsFromNode(document, vType.Value, anchorNodeMap)
	case *ast.StringNode:
		// transform gitlab specific input markers
		script := replaceJobInputReference(vType.Value)
		pos := vType.GetToken().Position.Line
		return []scriptNode{{script, pos}}
	default:
		return nil
	}
}

func replaceJobInputReference(script string) Script {
	transformedString := script
	res := jobInputRegex.FindAllStringSubmatch(transformedString, -1)
	for i := range res {
		match := res[i][0]
		env := strings.ToUpper(res[i][1])
		env = "${" + strings.TrimSpace(env) + "}"
		transformedString = strings.Replace(transformedString, match, env, 1)
	}

	return Script(transformedString)
}

func pathFromSequence(node *ast.SequenceNode) *yaml.Path {
	pathBuilder := (&yaml.PathBuilder{}).Root()
	for _, pathValue := range node.Values {
		pathBuilder.Child(pathValue.String())
	}

	return pathBuilder.Build()
}

func readNodeFromReference(document *ast.DocumentNode, tag *ast.TagNode) *ast.Node {
	pathValues := tag.Value.(*ast.SequenceNode)
	pathString := pathFromSequence(pathValues)

	pathNode, _ := pathString.FilterNode(document.Body)
	return &pathNode
}
