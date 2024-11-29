package reader

import (
	"github.com/goccy/go-yaml/ast"
)

func NewScriptCheckDirectiveReader(decoder ScriptDecoder) ScriptDecoder {
	return ScriptDecoder{
		&scriptcheckDirectiveReader{
			parser:      decoder.parser,
			transformer: decoder.transformer,
		},
		decoder.parser,
		decoder.transformer,
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
	file    *ast.File
	Scripts []ScriptBlock

	parser        ScriptParser
	transformer   ScriptTransformer
	anchorNodeMap DocumentAnchorMap
}

func (r *scriptcheckDirectiveReader) readScriptsForAst(file *ast.File) ([]ScriptBlock, error) {
	walker := &scriptCheckDirectiveVisitor{
		file:          file,
		parser:        r.parser,
		transformer:   r.transformer,
		anchorNodeMap: make(DocumentAnchorMap),
	}

	for _, doc := range file.Docs {
		for _, n := range ast.Filter(ast.AnchorType, doc) {
			anchor := n.(*ast.AnchorNode)
			anchorName := anchor.Name.GetToken().Value
			walker.anchorNodeMap[anchorName] = anchor.Value
		}
		ast.Walk(walker, doc)
	}

	return walker.Scripts, nil
}

func (v *scriptCheckDirectiveVisitor) Visit(node ast.Node) ast.Visitor {
	// currently only mapping value types are supported
	if node.Type() != ast.MappingValueType {
		return v
	}

	if directive := scriptDirectiveFromComment(node.GetComment()); directive != nil {
		mappingValueNode := node.(*ast.MappingValueNode)
		name := mappingValueNode.Key.String()
		script := readScriptFromNode(mappingValueNode.Value, v.anchorNodeMap)

		script = v.transformer(script)
		scriptBlock := ScriptBlock{
			FileName:  v.file.Name,
			BlockName: "directive-" + name,
			Script:    script,
			Shell:     directive.ShellDirective(),
		}
		v.Scripts = append(v.Scripts, scriptBlock)
	}

	return v
}
