package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"scriptcheck/reader"
	"strings"
)

const stdoutOutput = "stdout"

// Shell to pass to shellcheck command
var checkOptions struct {
	shell          string
	outputFile     string
	shellCheckArgs []string
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVarP(&checkOptions.outputFile, "output", "o", stdoutOutput, "output file")
	checkCmd.Flags().StringVarP(&checkOptions.shell, "shell", "s", "sh", "Shell to pass to shellcheck")
	checkCmd.Flags().StringArrayVarP(&checkOptions.shellCheckArgs, "flags", "f", []string{}, "shellcheck arguments")
}

var checkCmd = &cobra.Command{
	Use:   "check [pattern]",
	Short: "Run shellcheck against scripts in pipeline yml files",
	Long:  "Run shellcheck against scripts in pipeline yml files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, files []string) {
		scriptFilesNames, err := runForFiles(files, writeTempFiles)
		if err != nil {
			os.Exit(1)
		}

		if len(scriptFilesNames) == 0 {
			os.Exit(0)
		}

		if err := runShellcheck(scriptFilesNames); err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				os.Exit(exitError.ExitCode())
			} else {
				os.Exit(2)
			}
		}
	},
}

func writeTempFiles(scripts []reader.ScriptBlock) ([]string, error) {
	tempDir, err := os.MkdirTemp("", "scripts")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	var fileNames []string
	for _, script := range scripts {
		file, err := createTempFile(tempDir, script)
		if err != nil {
			return nil, fmt.Errorf("unable to create temporary file %w", err)
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

func runShellcheck(fileNames []string) error {
	cmd := exec.Command("shellcheck", fileNames...)
	cmd.Dir, _ = os.Getwd()
	cmd.Args = append(cmd.Args, "--shell", checkOptions.shell)
	cmd.Stderr = os.Stderr

	// append provided shell args
	for _, arg := range checkOptions.shellCheckArgs {
		cmd.Args = append(cmd.Args, "--"+arg)
	}

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	switch checkOptions.outputFile {
	case stdoutOutput:
		cmd.Stdout = os.Stdout
	default:
		file, err := os.Create(checkOptions.outputFile)
		if err != nil {
			return fmt.Errorf("unable to create output file: %w", err)
		}
		cmd.Stdout = file
	}

	if err := cmd.Run(); err != nil {
		return err
		/*
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				os.Exit(exitError.ExitCode())
			} else {
				log.Printf("Error while running: %s\n", err.Error())
				os.Exit(1)
			}
		*/
	}

	return nil
}

func createTempFile(tempDir string, script reader.ScriptBlock) (*os.File, error) {
	transformedFileName := strings.ReplaceAll(script.GetOutputFileName(""), "/", "")
	tempF, err := os.CreateTemp(tempDir, fmt.Sprintf("script-%s", transformedFileName))
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %s", err.Error())
	}

	_, err = tempF.WriteString(script.Script)
	if err != nil {
		return nil, fmt.Errorf("unable to write to temp file: %s", err.Error())
	}

	return tempF, nil
}
