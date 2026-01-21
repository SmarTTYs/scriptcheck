package runtime

import (
	"errors"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
	"scriptcheck/format"
	"scriptcheck/reader"
	"testing"
)

func TestChecking(t *testing.T) {
	cases := [4]struct {
		options       *Options
		script        reader.ScriptBlock
		expectSuccess bool
	}{
		{
			options:       newOptionsWithDefaults(reader.PipelineTypeGitlab),
			script:        exampleScript("cd TEST", nil),
			expectSuccess: false,
		},
		{
			options:       newOptionsWithDefaults(reader.PipelineTypeGitlab),
			script:        exampleScript("cd TEST || exit 1", nil),
			expectSuccess: true,
		},
		{
			options: newOptionsWithDefaults(reader.PipelineTypeGitlab),
			script: exampleScript(
				"cd TEST || exit 1",
				&reader.ScriptDirective{"shell": "bash"},
			),
			expectSuccess: true,
		},
		{
			options: newOptionsWithDefaults(reader.PipelineTypeGitlab),
			script: exampleScript(
				"cd TEST || exit 1",
				&reader.ScriptDirective{"disable": "SC12345"},
			),
			expectSuccess: true,
		},
	}

	for _, c := range cases {
		err := checkScripts(c.options, []reader.ScriptBlock{c.script})
		var scriptCheckError *ScriptCheckError
		if errors.As(err, &scriptCheckError) == c.expectSuccess {
			t.Errorf("error should be ScriptCheckError")
		}
	}
}

func newOptionsWithDefaults(pipelineType reader.PipelineType) *Options {
	return &Options{
		Format:       format.JsonFormat,
		OutputFile:   StdoutOutput,
		PipelineType: pipelineType,
	}
}

func exampleScript(script string, nodeDirective *reader.ScriptDirective) reader.ScriptBlock {
	scriptNode := reader.ScriptNode{
		Script:        reader.Script(script),
		Line:          10,
		NodeDirective: nodeDirective,
	}

	return reader.NewScriptBlock(
		"test",
		"key",
		"sh",
		scriptNode,
		ast.String(token.String("example", "org", &token.Position{})),
	)
}
