package runtime

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scriptcheck/reader"
	"strings"
)

func ExtractScripts(options *Options, files []string) error {
	log.Printf("Extracting scripts from %d files...\n", len(files))
	if scripts, err := extractScriptsFromFiles(options, files, writeFiles); err != nil {
		log.Printf("Error extracting scripts: %v\n", err)
		return err
	} else {
		log.Printf(
			"Successfully extracted %d scripts and saved into '%s' directory!",
			len(scripts),
			options.OutputDirectory,
		)

		return nil
	}
}

func writeFiles(options *Options, scripts []reader.ScriptBlock) ([]string, error) {
	err := os.MkdirAll(options.OutputDirectory, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	var fileNames []string
	for _, script := range scripts {
		file, err := createAndWriteFile(options, script, options.OutputDirectory)
		if err != nil {
			return nil, fmt.Errorf("unable to create temporary file %w", err)
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

func createAndWriteFile(options *Options, script reader.ScriptBlock, directory string) (*os.File, error) {
	filePath := script.GetOutputFilePath(directory)

	// create nested directories
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	tempF, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %s", err.Error())
	}

	if !strings.HasPrefix(script.Script, "#!") {
		var scriptShell string
		if len(script.Shell) > 0 {
			scriptShell = script.Shell
		} else {
			scriptShell = options.Shell
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
