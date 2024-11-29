package cmd

import (
	"log"
	"scriptcheck/reader"
)

type fileWriter func([]reader.ScriptBlock) ([]string, error)

func runForFiles(files []string, writer fileWriter) ([]string, error) {
	var scriptFilesNames = make([]string, 0)
	decoder := reader.NewDecoder(sharedOptions.pipelineType)

	for _, file := range files {
		fileScripts, err := decoder.DecodeFile(file)
		if err != nil {
			log.Printf("Error while running: %s\n", err.Error())
			return nil, err
		}

		scriptFileNames, err := writer(fileScripts)
		if err != nil {
			log.Printf("Error while running: %s\n", err.Error())
			return nil, err
		}

		scriptFilesNames = append(scriptFilesNames, scriptFileNames...)
	}

	return scriptFilesNames, nil
}
