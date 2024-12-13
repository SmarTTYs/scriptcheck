package runtime

import (
	"github.com/bmatcuk/doublestar/v4"
	"log"
	"scriptcheck/reader"
)

const StdoutOutput = "stdout"

type FileWriter func(*Options, []reader.ScriptBlock) ([]string, error)

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
