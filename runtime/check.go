package runtime

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"scriptcheck/reader"
	"strings"
)

func CheckScripts(options *Options, files []string) error {
	log.Printf("Checking scripts from [%d] files...\n", len(files))
	if scriptFilesNames, err := extractScriptsFromFiles(options, files, writeTempFiles); err != nil {
		os.Exit(1)
	} else {
		if len(scriptFilesNames) == 0 {
			os.Exit(0)
		}

		if err := runShellcheck(options, scriptFilesNames); err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				log.Println("Found shellcheck issues, exiting...")
				os.Exit(exitError.ExitCode())
			} else {
				log.Println("There was an error checking your files...")
				os.Exit(2)
			}
		}
	}
	log.Printf("Successfully checked files!")
	return nil
}

func runShellcheck(options *Options, fileNames []string) error {
	cmd := exec.Command("shellcheck", fileNames...)
	cmd.Dir, _ = os.Getwd()
	cmd.Args = append(cmd.Args, "--shell", options.Shell)
	cmd.Stderr = os.Stderr

	// append provided shell args
	for _, arg := range options.ShellCheckArgs {
		cmd.Args = append(cmd.Args, "--"+arg)
	}

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	switch options.OutputFile {
	case StdoutOutput:
		cmd.Stdout = os.Stdout
	default:
		file, err := os.Create(options.OutputFile)
		if err != nil {
			return fmt.Errorf("unable to create output file: %w", err)
		}
		cmd.Stdout = file
	}

	return cmd.Run()
}

func writeTempFiles(_ *Options, scripts []reader.ScriptBlock) ([]string, error) {
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

// todo: improve the dir / file names
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
