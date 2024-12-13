package runtime

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scriptcheck/reader"
)

func ExtractScripts(options *Options, files []string) error {
	log.Printf("Extracting scripts from %d files...\n", len(files))
	if scripts, err := extractScriptsFromFiles(options, files); err != nil {
		log.Printf("Error extracting scripts: %v\n", err)
		return err
	} else {
		_, err = writeFiles(options, scripts)
		if err != nil {
			return err
		}

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

	// write into file
	err = writeScriptBlock(tempF, options, script)
	if err != nil {
		return nil, err
	}

	return tempF, nil
}
