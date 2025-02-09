package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"strings"
)

const scriptCheckPrefix = "scriptcheck"

type ScriptDirective map[string]string

func scriptDirectiveFromString(dataString string) ScriptDirective {
	data := strings.TrimPrefix(dataString, scriptCheckPrefix)
	data = strings.TrimSpace(data)
	markerParts := strings.Split(data, " ")

	directives := map[string]string{}
	for _, markerPart := range markerParts {
		if len(markerPart) > 0 {
			keyValue := strings.SplitN(markerPart, "=", 2)
			key := keyValue[0]

			var value string
			if len(keyValue) > 1 {
				value = keyValue[1]
			}
			directives[key] = value
		}
	}

	return directives
}

func (d ScriptDirective) ShellDirective() string {
	return d["shell"]
}

func (d ScriptDirective) DisabledRules() []string {
	disabled, ok := d["disable"]
	if ok {
		return strings.Split(disabled, ",")
	} else {
		return []string{}
	}
}

// todo: improve empty directive handling
func (d ScriptDirective) asShellcheckDirective(script ScriptBlock) *string {
	if !script.HasShell() && len(d) == 0 {
		return nil
	}

	directiveBuilder := new(strings.Builder)
	directiveBuilder.WriteString("# shellcheck")

	if script.HasShell() {
		directiveBuilder.WriteString(fmt.Sprintf(" shell=%s", script.Shell))
	}

	if len(d.DisabledRules()) > 0 {
		rulesString := strings.Join(d.DisabledRules(), ",")
		directiveBuilder.WriteString(fmt.Sprintf(" disable=%s", rulesString))
	}

	directiveBuilder.WriteString("\n")

	directive := directiveBuilder.String()
	return &directive
}

func (d ScriptDirective) merge(other *ScriptDirective) *ScriptDirective {
	if other == nil {
		return &d
	}

	disabledRule := d.DisabledRules()
	disabledRule = append(disabledRule, other.DisabledRules()...)
	directive := make(ScriptDirective)
	directive["disable"] = strings.Join(disabledRule, ",")
	directive["shell"] = d.ShellDirective()

	return &directive
}

func findSequenceElementDirective(
	sequence *ast.SequenceNode,
	elementIndex int,
) *ScriptDirective {
	var comment *ast.CommentGroupNode
	if elementIndex == 0 {
		comment = sequence.Comment
	} else {
		comment = sequence.ValueHeadComments[elementIndex]
	}

	return scriptDirectiveFromComment(comment)
}

func scriptDirectiveFromComment(comment *ast.CommentGroupNode) *ScriptDirective {
	if marker := findScriptCheckMarker(comment); marker != nil {
		directive := scriptDirectiveFromString(*marker)
		return &directive
	}

	return nil
}

func findScriptCheckMarker(comment *ast.CommentGroupNode) *string {
	if comment == nil {
		return nil
	}

	for _, comment := range comment.Comments {
		trimmed := strings.TrimSpace(comment.Token.Value)
		if strings.HasPrefix(trimmed, scriptCheckPrefix) {
			return &trimmed
		}
	}

	return nil
}
