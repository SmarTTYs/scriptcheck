package runtime

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scriptcheck/reader"
)

func ExtractScripts(options *Options, fileNames []string) error {
	files, err := collectFiles(fileNames)
	if err != nil {
		return err
	}

	log.Printf("Extracting scripts from %d files...\n", len(files))
	if scripts, err := extractScriptsFromFiles(options, files); err != nil {
		log.Printf("Error extracting scripts: %v\n", err)
		return err
	} else {
		err = writeFiles(options, scripts)
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

func writeFiles(options *Options, scripts []reader.ScriptBlock) error {
	err := os.MkdirAll(options.OutputDirectory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	for _, script := range scripts {
		err := createAndWriteFile(options, script, options.OutputDirectory)
		if err != nil {
			return fmt.Errorf("unable to create temporary file %w", err)
		}
	}

	return nil
}

func createAndWriteFile(options *Options, script reader.ScriptBlock, directory string) error {
	filePath := script.GetOutputFilePath(directory)

	// create nested directories
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create file: %s", err.Error())
	}

	defer func() {
		_ = file.Close()
	}()

	// write into file
	err = writeScriptBlock(file, options, script)
	if err != nil {
		return err
	}

	return nil
}
