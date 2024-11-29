package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"scriptcheck/reader"
	"strings"
)

var extractOptions struct {
	shell           string
	outputDirectory string
}

func init() {
	rootCmd.AddCommand(extractCommand)
	extractCommand.Flags().StringVarP(&extractOptions.shell, "shell", "s", "sh", "Shell to pass to shellcheck")
	extractCommand.Flags().StringVarP(&extractOptions.outputDirectory, "output", "o", "scripts", "TODO")
}

var extractCommand = &cobra.Command{
	Use:   "extract [pattern]",
	Short: "Extract script blocks from pipeline yaml files",
	Long:  "Extract script blocks from pipeline yaml files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := runForFiles(args, writeFiles); err != nil {
			os.Exit(1)
		}
	},
}

func writeFiles(scripts []reader.ScriptBlock) ([]string, error) {
	err := os.MkdirAll(extractOptions.outputDirectory, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	var fileNames []string
	for _, script := range scripts {
		file, err := createAndWriteFile(script, extractOptions.outputDirectory)
		if err != nil {
			return nil, fmt.Errorf("unable to create temporary file %w", err)
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

func createAndWriteFile(script reader.ScriptBlock, directory string) (*os.File, error) {
	fileName := script.GetOutputFileName(directory)
	tempF, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %s", err.Error())
	}

	if !strings.HasPrefix(script.Script, "#!") {
		var scriptShell string
		if len(script.Shell) > 0 {
			scriptShell = script.Shell
		} else {
			scriptShell = extractOptions.shell
		}

		if _, err := tempF.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", scriptShell)); err != nil {
			return nil, fmt.Errorf("unable to write to temp file: %s", err.Error())
		}
	}

	if _, err = tempF.WriteString(script.Script); err != nil {
		return nil, fmt.Errorf("unable to write to temp file: %s", err.Error())
	}

	return tempF, nil
}
