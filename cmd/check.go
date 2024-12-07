package cmd

import (
	"github.com/spf13/cobra"
	"scriptcheck/runtime"
)

func newCheckCommand(options *runtime.Options) *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check [pattern]",
		Short: "Run shellcheck against scripts in pipeline yml files",
		Long:  "Run shellcheck against scripts in pipeline yml files",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, files []string) {
			_ = runtime.CheckScripts(options, files)
		},
	}

	checkCmd.Flags().StringVarP(
		&options.OutputFile,
		"output",
		"o",
		runtime.StdoutOutput,
		"output file to write into",
	)

	checkCmd.Flags().StringVarP(
		&options.Shell,
		"shell",
		"s",
		"sh",
		"Shell to pass to shellcheck",
	)

	checkCmd.Flags().StringArrayVarP(
		&options.ShellCheckArgs,
		"flags",
		"f",
		[]string{},
		"shellcheck arguments",
	)

	return checkCmd
}
