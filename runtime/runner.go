package runtime

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"log"
	"scriptcheck/color"
	"scriptcheck/reader"
)

const StdoutOutput = "stdout"

func collectAndExtractScripts(options *Options, globPatterns []string) ([]reader.ScriptBlock, []string, error) {
	files, err := collectFiles(globPatterns)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Reading %s file(s)...\n", color.Color(len(files), color.Bold))
	if scripts, err := extractScriptsFromFiles(options, files); err != nil {
		log.Printf("Error extracting scripts: %v\n", err)
		return nil, nil, fmt.Errorf("unable to extract files: %w", err)
	} else {
		if len(scripts) == 0 && options.Strict {
			return nil, nil, fmt.Errorf("no scripts found in %s", color.Color(len(files), color.Bold))
		}

		return scripts, files, nil
	}
}

func extractScriptsFromFiles(options *Options, files []string) ([]reader.ScriptBlock, error) {
	decoder := reader.NewDecoder(options.PipelineType, options.Debug)
	scripts := make([]reader.ScriptBlock, 0)

	if options.Merge {
		fileScripts, err := decoder.MergeAndDecode(files)
		if err != nil {
			log.Printf("Error while merging: %s\n", err.Error())
			return nil, err
		}
		scripts = append(scripts, fileScripts...)
	} else {
		for _, file := range files {
			fileScripts, err := decoder.DecodeFile(file)
			if err != nil {
				log.Printf("Error while running: %s\n", err.Error())
				return nil, err
			}
			scripts = append(scripts, fileScripts...)
		}
	}

	return scripts, nil
}

func collectFiles(globPatterns []string) ([]string, error) {
	files := make([]string, 0)
	for _, pattern := range globPatterns {
		globFiles, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return nil, err
		}
		files = append(files, globFiles...)
	}

	return files, nil
}
