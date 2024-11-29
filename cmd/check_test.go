package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"testing"
)

func TestCheckExample(t *testing.T) {
	_, err := ExecuteCommand(rootCmd, "check", "../dir/first_yaml.yml")
	println("error", err)
	// println(consoleOutput)
}

func ExecuteCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = ExecuteCommandC(root, args...)
	return output, err
}

func ExecuteCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
