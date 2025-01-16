package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
	"os"
	"scriptcheck/color"
	"scriptcheck/format"
	"scriptcheck/runtime"
)

func newCheckCommand(options *runtime.Options) *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check [pattern]",
		Short: "Run shellcheck against scripts in pipeline yml files",
		Long:  "Run shellcheck against scripts in pipeline yml files",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, globPatterns []string) {
			if err := runtime.CheckFiles(options, globPatterns); err != nil {
				var scriptCheckError *runtime.ScriptCheckError
				if errors.As(err, &scriptCheckError) {
					log.Printf(
						"Found %s issues, exiting...",
						color.Color(scriptCheckError.ReportCount(), color.Bold),
					)
					os.Exit(1)
				} else {
					log.Println("There was an error checking your files...")
					os.Exit(2)
				}
			} else {
				log.Printf("Successfully checked files!")
			}
		},
	}

	checkCmd.Flags().StringVar(
		&options.DefaultShell,
		"default-shell",
		"",
		"Defines default shell dialect to use in case no shebang or scriptcheck directive is used. Per default NO dialect will get specified",
	)

	checkCmd.Flags().StringVarP(
		&options.OutputFile,
		"output",
		"o",
		runtime.StdoutOutput,
		"output file to write into",
	)

	checkCmd.Flags().StringArrayVarP(
		&options.ShellCheckArgs,
		"args",
		"a",
		[]string{},
		"shellcheck arguments",
	)

	formatOptions := []format.Format{format.StandardFormat, format.CodeQualityFormat, format.JsonFormat}
	enumVarP(
		checkCmd.PersistentFlags(),
		formatOptions,
		&options.Format,
		format.StandardFormat,
		"format",
		"f",
		"Format in which you want to print shellcheck results",
	)

	return checkCmd
}
