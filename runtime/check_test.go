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
	cases := [2]struct {
		options       *Options
		script        reader.ScriptBlock
		expectSuccess bool
	}{
		{
			options:       newOptionsWithDefaults(reader.PipelineTypeGitlab),
			script:        exampleScript("cd TEST"),
			expectSuccess: false,
		},
		{
			options:       newOptionsWithDefaults(reader.PipelineTypeGitlab),
			script:        exampleScript("cd TEST || exit 1"),
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

func exampleScript(script string) reader.ScriptBlock {
	return reader.NewScriptBlock(
		"test",
		"key",
		"job",
		script,
		ast.String(token.String("example", "org", &token.Position{})),
	)
}
