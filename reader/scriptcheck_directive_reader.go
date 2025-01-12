package reader

import (
	"github.com/goccy/go-yaml/ast"
)

func newScriptCheckDirectiveDecoder(decoder ScriptDecoder) ScriptDecoder {
	return ScriptDecoder{
		ScriptReader: &scriptcheckDirectiveReader{
			parser:      decoder.parser,
			transformer: decoder.transformer,
		},
		debug:       decoder.debug,
		parser:      decoder.parser,
		transformer: decoder.transformer,
	}
}

//goland:noinspection SpellCheckingInspection
type scriptcheckDirectiveReader struct {
	ScriptReader

	parser      scriptParser
	transformer scriptTransformer
}

type scriptCheckDirectiveVisitor struct {
	ast.Visitor
	file *ast.File

	// currently looped document
	document *ast.DocumentNode

	Scripts []ScriptBlock

	parser        scriptParser
	transformer   scriptTransformer
	anchorNodeMap documentAnchorMap
}

func (r *scriptcheckDirectiveReader) readScriptsForAst(file *ast.File) ([]ScriptBlock, error) {
	directiveWalker := &scriptCheckDirectiveVisitor{
		file:          file,
		parser:        r.parser,
		transformer:   r.transformer,
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

		if script := v.parser(v.document, nodeValue, v.anchorNodeMap); len(script) > 0 {
			script = v.transformer(script)
			blockName := "directive-" + name
			scriptBlock := NewScriptBlock(
				v.file.Name,
				blockName,
				script,
				nodeValue,
			)
			scriptBlock.Shell = directive.ShellDirective()
			v.Scripts = append(v.Scripts, scriptBlock)
		}
	}

	return v
}
