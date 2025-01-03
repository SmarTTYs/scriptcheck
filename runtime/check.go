package runtime

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"scriptcheck/format"
	"scriptcheck/reader"
	"slices"
	"strings"
)

func CheckScripts(options *Options, fileNames []string) error {
	files, err := collectFiles(fileNames)
	if err != nil {
		return err
	}

	log.Printf("Checking scripts from [%d] files...\n", len(files))
	if scripts, err := extractScriptsFromFiles(options, files); err != nil {
		os.Exit(1)
	} else {
		if len(scripts) == 0 {
			os.Exit(0)
		}

		tempDir, fileScriptBlockMap, err := writeTempFiles(options, scripts)
		if err != nil {
			return err
		}
		fileNames := slices.Collect(maps.Keys(fileScriptBlockMap))

		err = runShellcheck(options, fileNames, fileScriptBlockMap)
		_ = os.RemoveAll(*tempDir)

		if err != nil {
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

func runShellcheck(options *Options, fileNames []string, scriptMap map[string]reader.ScriptBlock) error {
	report, err := executeShellCheck(scriptMap, options, fileNames)
	if err != nil {
		return fmt.Errorf("unable to parse shellcheck report: %w", err)
	}

	formatter := format.NewFormatter(options.Format)
	reportString, err := formatter.Format(report)
	if err != nil {
		return fmt.Errorf("unable to format shellcheck report: %w", err)
	}

	if writeErr := writeReport(options, reportString); writeErr != nil {
		return writeErr
	}

	return nil
}

func executeShellCheck(scriptMap map[string]reader.ScriptBlock, options *Options, fileNames []string) ([]format.ScriptCheckReport, error) {
	out := new(bytes.Buffer)
	cmd := exec.Command("shellcheck", fileNames...)
	cmd.Dir, _ = os.Getwd()
	cmd.Args = append(cmd.Args, "--shell", options.Shell)
	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	// append provided shell args
	for _, arg := range options.ShellCheckArgs {
		cmd.Args = append(cmd.Args, "--"+arg)
	}

	// always force json format in order to parse it afterward
	cmd.Args = append(cmd.Args, "--format", "json")

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	if runErr := cmd.Run(); runErr != nil {
		return format.NewScriptCheckReport(out.Bytes(), scriptMap)
	} else {
		return nil, runErr
	}
}

func writeReport(options *Options, report string) error {
	var writerFile *os.File
	switch options.OutputFile {
	case StdoutOutput:
		writerFile = os.Stdout
	default:
		file, err := os.Create(options.OutputFile)
		if err != nil {
			return err
		}
		writerFile = file

		defer func() {
			_ = file.Close()
		}()
	}

	_, err := writerFile.WriteString(report)
	if err != nil {
		return fmt.Errorf("unable to write shellcheck output: %w", err)
	}

	return nil
}

func writeTempFiles(options *Options, scripts []reader.ScriptBlock) (*string, map[string]reader.ScriptBlock, error) {
	tempDir, err := os.MkdirTemp("", "scripts")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	var fileNames = make(map[string]reader.ScriptBlock)
	for _, script := range scripts {
		file, err := createTempFile(options, tempDir, script)
		if err != nil {
			return &tempDir, nil, fmt.Errorf("unable to create temporary file %w", err)
		}
		fileNames[file.Name()] = script
	}

	return &tempDir, fileNames, nil
}

func createTempFile(options *Options, tempDir string, script reader.ScriptBlock) (*os.File, error) {
	filePath := script.GetOutputFilePath("")
	transformedFileName := strings.ReplaceAll(filePath, "/", "")
	filePattern := fmt.Sprintf("script-*-%s", transformedFileName)
	tempF, err := os.CreateTemp(tempDir, filePattern)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %s", err.Error())
	}

	// write into file
	err = writeScriptBlock(tempF, options, script)
	if err != nil {
		return nil, err
	}

	return tempF, nil
}

func writeScriptBlock(file *os.File, options *Options, script reader.ScriptBlock) error {
	// ensure every script starts either with a script or shellcheck directive
	if !script.Script.HasShell() {
		var scriptShell string
		if len(script.Shell) > 0 {
			scriptShell = script.Shell
		} else {
			scriptShell = options.Shell
		}

		if _, err := file.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", scriptShell)); err != nil {
			return fmt.Errorf("unable to write to file: %s", err.Error())
		}
	}

	if _, err := file.WriteString(script.ScriptString()); err != nil {
		return fmt.Errorf("unable to write to file: %s", err.Error())
	}

	return nil
}
