package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"scriptcheck/reader"
	"scriptcheck/runtime"
)

var rootCmd = newRootCmd()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

func newRootCmd() *cobra.Command {
	options := runtime.NewOptions()
	cmd := &cobra.Command{
		Use:   "scriptcheck",
		Short: "Simple utility cli for working with pipeline scripts",
		Long:  "CLI allowing to check or extract inlined pipeline scripts",
	}

	cmd.PersistentFlags().BoolVar(
		&options.Debug,
		"verbose",
		false,
		"Verbose output",
	)

	cmd.PersistentFlags().BoolVar(
		&options.Merge,
		"merge",
		false,
		"Whether to merge all input files into one file",
	)

	typeOptions := []reader.PipelineType{reader.PipelineTypeGitlab}
	enumVarP(
		cmd.PersistentFlags(),
		typeOptions,
		&options.PipelineType,
		reader.PipelineTypeGitlab,
		"type",
		"t",
		"YAMl file pipeline type",
	)

	cmd.AddCommand(
		newCheckCommand(options),
		newExtractCommand(options),
	)

	return cmd
}
