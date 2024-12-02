package runtime

import (
	"log"
	"scriptcheck/reader"
)

const StdoutOutput = "stdout"

type FileWriter func(Options, []reader.ScriptBlock) ([]string, error)

func RunForFiles(options Options, files []string, writer FileWriter) ([]string, error) {
	var scriptFilesNames = make([]string, 0)
	decoder := reader.NewDecoder(options.PipelineType)

	for _, file := range files {
		fileScripts, err := decoder.DecodeFile(file)
		if err != nil {
			log.Printf("Error while running: %s\n", err.Error())
			return nil, err
		}

		scriptFileNames, err := writer(options, fileScripts)
		if err != nil {
			log.Printf("Error while running: %s\n", err.Error())
			return nil, err
		}

		scriptFilesNames = append(scriptFilesNames, scriptFileNames...)
	}

	return scriptFilesNames, nil
}
