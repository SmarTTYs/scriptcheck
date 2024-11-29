package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"scriptcheck/reader"
)

var sharedOptions struct {
	pipelineType reader.PipelineType
	pattern      string
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&sharedOptions.pattern,
		"pattern",
		"p",
		"*.yml",
		"Filenames to extract script blocks from",
	)

	typeOptions := []reader.PipelineType{reader.PipelineTypeGitlab}
	enumVarP(
		rootCmd.PersistentFlags(),
		typeOptions,
		&sharedOptions.pipelineType,
		reader.PipelineTypeGitlab,
		"type",
		"t",
		"YAMl file pipeline type",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

var rootCmd = &cobra.Command{
	Use:   "scriptcheck",
	Short: "Simple utility cli for working with pipeline scripts",
	Long:  "CLI allowing to check or extract inlined pipeline scripts",
}
