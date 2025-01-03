package reader

import (
	"github.com/goccy/go-yaml/ast"
)

func newScriptCheckDirectiveDecoder(decoder ScriptDecoder) ScriptDecoder {
	return ScriptDecoder{
		ScriptReader: &scriptcheckDirectiveReader{
			parser:      decoder.Parser,
			transformer: decoder.Transformer,
		},
		Debug:       decoder.Debug,
		Parser:      decoder.Parser,
		Transformer: decoder.Transformer,
	}
}

//goland:noinspection SpellCheckingInspection
type scriptcheckDirectiveReader struct {
	ScriptReader

	parser      ScriptParser
	transformer ScriptTransformer
}

type scriptCheckDirectiveVisitor struct {
	ast.Visitor
	file *ast.File

	// currently looped document
	document *ast.DocumentNode

	Scripts []ScriptBlock

	parser        ScriptParser
	transformer   ScriptTransformer
	anchorNodeMap DocumentAnchorMap
}

func (r *scriptcheckDirectiveReader) readScriptsForAst(file *ast.File) ([]ScriptBlock, error) {
	directiveWalker := &scriptCheckDirectiveVisitor{
		file:          file,
		parser:        r.parser,
		transformer:   r.transformer,
		anchorNodeMap: make(DocumentAnchorMap),
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

	if directive := ScriptDirectiveFromComment(node.GetComment()); directive != nil {
		mappingValueNode := node.(*ast.MappingValueNode)
		name := mappingValueNode.Key.String()

		if script := v.parser(v.document, mappingValueNode.Value, v.anchorNodeMap); len(script) > 0 {
			script = v.transformer(script)
			scriptBlock := ScriptBlock{
				FileName:  v.file.Name,
				BlockName: "directive-" + name,
				Script:    Script(script),
				Shell:     directive.ShellDirective(),
				Path:      mappingValueNode.Value.GetPath(),
			}
			v.Scripts = append(v.Scripts, scriptBlock)
		}
	}

	return v
}
