package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"strings"
)

const scriptCheckPrefix = "scriptcheck"

type ScriptDirective map[string]string

type ScriptDirectiveName struct {
	shell         string
	disabledRules []string
}

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

	var disabledRules []string
	disabled, ok := directives["disable"]
	if ok {
		disabledRules = strings.Split(disabled, ",")
	} else {
		disabledRules = []string{}
	}
	test := ScriptDirectiveName{
		shell:         directives["shell"],
		disabledRules: disabledRules,
	}
	println("shell", test.shell)
	println("rules", test.disabledRules)
	println("---")

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

	directiveBuilderNew := new(strings.Builder)
	directiveBuilderNew.WriteString("# shellcheck")
	for key, value := range d {
		if len(value) > 0 {
			directiveBuilderNew.WriteString(fmt.Sprintf(" %v=%v", key, value))
		}
	}
	directiveBuilderNew.WriteString("\n")
	println("New", directiveBuilderNew.String())

	/*
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
	*/
	directive := directiveBuilderNew.String()
	return &directive
}

func merge(base, other *ScriptDirective) *ScriptDirective {
	if base == nil {
		return other
	}

	if other == nil {
		return base
	}

	disabledRule := base.DisabledRules()
	disabledRule = append(disabledRule, other.DisabledRules()...)
	directive := make(ScriptDirective)
	directive["disable"] = strings.Join(disabledRule, ",")
	directive["shell"] = base.ShellDirective()

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
