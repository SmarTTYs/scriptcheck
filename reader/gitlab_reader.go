package reader

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
	"regexp"
	"slices"
	"strings"
	"unicode"
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

func newGitlabDecoder(debug bool, defaultShell string, experimentalFolding bool) ScriptDecoder {
	decoder := ScriptDecoder{
		ScriptReader: gitlabScriptReader{
			defaultShell:        defaultShell,
			anchorNodeMap:       make(documentAnchorMap),
			experimentalFolding: experimentalFolding,
		},
		defaultShell:        defaultShell,
		debug:               debug,
		parser:              readScriptsFromNode,
		experimentalFolding: experimentalFolding,
	}

	return decoder
}

type gitlabScriptReader struct {
	ScriptReader

	defaultShell string

	experimentalFolding bool

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
			for i, script := range readScriptsFromNode(r.document, eValue, r.anchorNodeMap, r.experimentalFolding) {
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
					directive,
				)

				scripts = append(scripts, scriptBlock)
			}
		}
	}

	return scripts
}

func readScriptsFromNode(
	document *ast.DocumentNode,
	node ast.Node,
	anchorNodeMap documentAnchorMap,
	experimentalFolding bool,
) []scriptNode {
	switch vType := node.(type) {
	case *ast.TagNode:
		if vType.Start.Value == gitlabReferenceTag {
			referencedNode := readNodeFromReference(document, vType)
			if referencedNode != nil {
				return readScriptsFromNode(document, *referencedNode, anchorNodeMap, experimentalFolding)
			} else {
				return nil
			}
		} else {
			return nil
		}
	case *ast.AnchorNode:
		return readScriptsFromNode(document, vType.Value, anchorNodeMap, experimentalFolding)
	case *ast.AliasNode:
		aliasName := vType.Value.GetToken().Value
		if anchorValue, exists := anchorNodeMap[aliasName]; !exists {
			panic(fmt.Sprintf("anchor %s not found!", aliasName))
		} else {
			return readScriptsFromNode(document, anchorValue, anchorNodeMap, experimentalFolding)
		}
	case *ast.SequenceNode:
		elements := make([]scriptNode, 0)
		for _, listElement := range vType.Values {
			scripts := readScriptsFromNode(document, listElement, anchorNodeMap, experimentalFolding)
			elements = append(elements, scripts...)
		}
		return elements
	// currently we do not directly create the script
	// for the literals value as in this case the
	// position (line) seems to be off. So we use the
	// literals position and increment it by 1 as yaml
	// expects a line break
	case *ast.LiteralNode:
		var scriptString string
		if vType.Start.Type == token.FoldedType && experimentalFolding {
			origin := strings.TrimFunc(vType.Value.GetToken().Origin, unicode.IsSpace)
			scriptString = unfoldFoldedLiteral(origin)
		} else {
			scriptString = vType.Value.Value
		}
		script := replaceJobInputReference(scriptString)
		pos := vType.Start.Position.Line + 1
		return []scriptNode{{script, pos}}
	case *ast.StringNode:
		// transform gitlab specific input markers
		script := replaceJobInputReference(vType.Value)
		pos := vType.GetToken().Position.Line
		return []scriptNode{{script, pos}}
	default:
		return nil
	}
}

func unfoldFoldedLiteral(literal string) string {
	sb := new(strings.Builder)
	/*
		lineScanner := bufio.NewScanner(strings.NewReader(literal))
		for lineScanner.Scan() {
			println("line", lineScanner.Text())
		}
	*/
	lbc := token.DetectLineBreakCharacter(literal)
	lines := strings.Split(literal, lbc)
	for index, element := range lines {
		trimmed := strings.TrimLeftFunc(element, unicode.IsSpace)
		if strings.ContainsFunc(element, isNoWhitespace) {
			// for comments do not append a trailing slash
			if strings.HasPrefix(trimmed, "#") {
				sb.WriteString(trimmed + "\n")
			} else {
				sb.WriteString(trimmed)
				if index != len(lines)-1 {
					sb.WriteString(" \\\n")
				}
			}
		} else {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func isNoWhitespace(rune rune) bool {
	return !unicode.IsSpace(rune)
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
