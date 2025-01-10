package runtime

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"scriptcheck/format"
	"scriptcheck/reader"
)

func ExtractScripts(options *Options, globPatterns []string) error {
	scripts, files, err := collectAndExtractScripts(options, globPatterns)
	if err != nil {
		return err
	}

	log.Printf(
		"Extracting %s script(s) from %s file(s)...\n",
		format.Color(len(scripts), format.Bold),
		format.Color(len(files), format.Bold),
	)

	if writeErr := writeFiles(options, scripts); writeErr != nil {
		return writeErr
	}

	log.Printf(
		"Successfully extracted %s scripts and saved into %s directory!",
		format.Color(len(scripts), format.Bold),
		format.Color(options.OutputDirectory, format.Bold),
	)

	return nil
}

func writeFiles(options *Options, scripts []reader.ScriptBlock) error {
	err := os.MkdirAll(options.OutputDirectory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create dir: %s", err.Error())
	}

	dirWriter := DirScriptWriter{
		directory: options.OutputDirectory,
	}

	for _, script := range scripts {
		_, err2 := dirWriter.WriteScriptTest(options.OutputDirectory+"/experimental", script)
		if err2 != nil {
			println("Err2", err2.Error())
			return err2
		}

		// todo: prepare for removal
		err := createAndWriteFile(script, options.OutputDirectory)
		if err != nil {
			return fmt.Errorf("unable to create file: %w", err)
		}
	}

	return nil
}

func createAndWriteFile(script reader.ScriptBlock, directory string) error {
	filePath := path.Join(directory, script.OutputFileName())

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
	return writeScriptBlock(file, script)
}
