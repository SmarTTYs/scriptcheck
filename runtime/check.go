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

func CheckScripts(options *Options, files []string) error {
	log.Printf("Checking scripts from [%d] files...\n", len(files))
	if scriptFilesNames, err := extractScriptsFromFiles(options, files); err != nil {
		os.Exit(1)
	} else {
		if len(scriptFilesNames) == 0 {
			os.Exit(0)
		}

		fileScriptBlockMap, err := writeTempFiles(options, scriptFilesNames)
		if err != nil {
			return err
		}

		if err := runShellcheck(options, fileScriptBlockMap); err != nil {
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

func runShellcheck(options *Options, scriptMap map[string]reader.ScriptBlock) error {
	var outB bytes.Buffer
	fileNames := slices.Collect(maps.Keys(scriptMap))

	cmd := exec.Command("shellcheck", fileNames...)
	cmd.Dir, _ = os.Getwd()
	cmd.Args = append(cmd.Args, "--shell", options.Shell)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outB

	// append provided shell args
	for _, arg := range options.ShellCheckArgs {
		cmd.Args = append(cmd.Args, "--"+arg)
	}

	if options.Format != format.StandardFormat {
		cmd.Args = append(cmd.Args, "--format", "json")
	}

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	executionErr := cmd.Run()

	var reportString string
	if options.Format != format.StandardFormat {
		formatter := format.NewFormatter(options.Format)
		reportString, _ = formatter.Format(outB.Bytes(), scriptMap)
	} else {
		reportString = outB.String()
	}

	err := writeReport(options, reportString)
	if err != nil {
		return err
	}

	return executionErr
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
	}

	_, err := writerFile.WriteString(report)
	if err != nil {
		return fmt.Errorf("unable to write shellcheck output: %w", err)
	}

	return nil
}

func writeTempFiles(options *Options, scripts []reader.ScriptBlock) (map[string]reader.ScriptBlock, error) {
	tempDir, err := os.MkdirTemp("", "scripts")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	var fileNames = make(map[string]reader.ScriptBlock)
	for _, script := range scripts {
		file, err := createTempFile(options, tempDir, script)
		if err != nil {
			return nil, fmt.Errorf("unable to create temporary file %w", err)
		}
		fileNames[file.Name()] = script
	}

	return fileNames, nil
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
	if !strings.HasPrefix(script.Script, "#!") {
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

	if _, err := file.WriteString(script.Script); err != nil {
		return fmt.Errorf("unable to write to file: %s", err.Error())
	}

	return nil
}
