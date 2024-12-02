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
	options := runtime.Options{}
	cmd := &cobra.Command{
		Use:   "scriptcheck",
		Short: "Simple utility cli for working with pipeline scripts",
		Long:  "CLI allowing to check or extract inlined pipeline scripts",
	}

	cmd.PersistentFlags().StringVarP(
		&options.Pattern,
		"pattern",
		"p",
		"*.yml",
		"Filenames to extract script blocks from",
	)

	cmd.PersistentFlags().BoolVarP(
		&options.Debug,
		"debug",
		"d",
		false,
		"Test",
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
