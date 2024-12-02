package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"scriptcheck/runtime"
)

func newExtractCommand(options runtime.Options) *cobra.Command {
	var extractCommand = &cobra.Command{
		Use:   "extract [pattern]",
		Short: "Extract script blocks from pipeline yaml files",
		Long:  "Extract script blocks from pipeline yaml files",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runtime.ExtractScripts(options, args); err != nil {
				os.Exit(1)
			}
		},
	}

	extractCommand.Flags().StringVarP(&options.Shell, "shell", "s", "sh", "Standard shell used for shellcheck directives")
	extractCommand.Flags().StringVarP(&options.OutputDirectory, "output", "o", "scripts", "Directory to extract files to")

	return extractCommand
}
