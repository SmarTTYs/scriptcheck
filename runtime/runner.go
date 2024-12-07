package runtime

import (
	"log"
	"scriptcheck/reader"
)

const StdoutOutput = "stdout"

type FileWriter func(*Options, []reader.ScriptBlock) ([]string, error)

func extractScriptsFromFiles(options *Options, files []string, writer FileWriter) ([]string, error) {
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

	return writer(options, scripts)
}
