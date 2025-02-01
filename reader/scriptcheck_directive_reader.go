package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
)

func newScriptCheckDirectiveDecoder(decoder ScriptDecoder) ScriptDecoder {
	return ScriptDecoder{
		ScriptReader: &scriptcheckDirectiveReader{
			parser:                decoder.parser,
			defaultShell:          decoder.defaultShell,
			experirementalFolding: decoder.experimentalFolding,
		},
		defaultShell:        decoder.defaultShell,
		debug:               decoder.debug,
		parser:              decoder.parser,
		experimentalFolding: decoder.experimentalFolding,
	}
}

//goland:noinspection SpellCheckingInspection
type scriptcheckDirectiveReader struct {
	ScriptReader

	experirementalFolding bool
	defaultShell          string
	parser                scriptParser
}

type scriptCheckDirectiveVisitor struct {
	ast.Visitor
	file *ast.File

	// currently looped document
	document *ast.DocumentNode

	Scripts             []ScriptBlock
	experimentalFolding bool

	reader        *scriptcheckDirectiveReader
	anchorNodeMap documentAnchorMap
}

func (reader *scriptcheckDirectiveReader) readScriptsForAst(file *ast.File) ([]ScriptBlock, error) {
	directiveWalker := &scriptCheckDirectiveVisitor{
		file:          file,
		reader:        reader,
		anchorNodeMap: make(documentAnchorMap),
	}

	for _, doc := range file.Docs {
		if doc.Body != nil {
			directiveWalker.document = doc
			for _, n := range ast.Filter(ast.AnchorType, doc) {
				anchor := n.(*ast.AnchorNode)
				anchorName := anchor.Name.GetToken().Value
				directiveWalker.anchorNodeMap[anchorName] = anchor.Value
			}
			ast.Walk(directiveWalker, doc)
		}
	}

	return directiveWalker.Scripts, nil
}

func (v *scriptCheckDirectiveVisitor) Visit(node ast.Node) ast.Visitor {
	// currently only mapping value types are supported
	if node.Type() != ast.MappingValueType {
		return v
	}

	if directive := scriptDirectiveFromComment(node.GetComment()); directive != nil {
		mappingValueNode := node.(*ast.MappingValueNode)
		name := mappingValueNode.Key.String()
		nodeValue := mappingValueNode.Value

		if scripts := v.reader.parser(v.document, nodeValue, v.anchorNodeMap, v.experimentalFolding); len(scripts) > 0 {
			blockName := "directive_" + name
			for i, script := range scripts {
				var elementName string
				if i > 0 {
					elementName = blockName + fmt.Sprintf("_%d", i)
				} else {
					elementName = blockName
				}

				scriptBlock := NewScriptBlock(
					v.file.Name,
					elementName,
					v.reader.defaultShell,
					script,
					nodeValue,
					directive,
				)

				v.Scripts = append(v.Scripts, scriptBlock)
			}
		}
	}

	return v
}
