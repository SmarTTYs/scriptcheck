package reader

import (
	"github.com/goccy/go-yaml/ast"
)

func newScriptCheckDirectiveDecoder(decoder ScriptDecoder) ScriptDecoder {
	return ScriptDecoder{
		ScriptReader: &scriptcheckDirectiveReader{
			parser:      decoder.parser,
			transformer: decoder.transformer,

			defaultShell: decoder.defaultShell,
		},
		defaultShell: decoder.defaultShell,
		debug:        decoder.debug,
		parser:       decoder.parser,
		transformer:  decoder.transformer,
	}
}

//goland:noinspection SpellCheckingInspection
type scriptcheckDirectiveReader struct {
	ScriptReader

	defaultShell string
	parser       scriptParser
	transformer  scriptTransformer
}

type scriptCheckDirectiveVisitor struct {
	ast.Visitor
	file *ast.File

	defaultShell string

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
		defaultShell:  r.defaultShell,
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

		if scripts := v.parser(v.document, nodeValue, v.anchorNodeMap); len(scripts) > 0 {
			for _, script := range scripts {
				script.script = v.transformer(script.script)
				blockName := "directive_" + name
				scriptBlock := NewScriptBlock(
					v.file.Name,
					blockName,
					v.defaultShell,
					script,
					nodeValue,
				)

				if directiveShell := directive.ShellDirective(); directiveShell != "" {
					scriptBlock.Shell = directiveShell
				}

				v.Scripts = append(v.Scripts, scriptBlock)
			}
		}
	}

	return v
}
