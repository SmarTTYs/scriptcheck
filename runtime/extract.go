package runtime

import (
	"log"
	"scriptcheck/color"
	"scriptcheck/reader"
)

func ExtractScripts(options *Options, globPatterns []string) error {
	scripts, files, err := collectAndExtractScripts(options, globPatterns)
	if err != nil {
		return err
	}

	log.Printf(
		"Extracting %s script(s) from %s file(s)...\n",
		color.Color(len(scripts), color.Bold),
		color.Color(len(files), color.Bold),
	)

	if writeErr := writeFiles(options, scripts); writeErr != nil {
		return writeErr
	}

	log.Printf(
		"Successfully extracted %s scripts and saved into %s directory!",
		color.Color(len(scripts), color.Bold),
		color.Color(options.OutputDirectory, color.Bold),
	)

	return nil
}

func writeFiles(options *Options, scripts []reader.ScriptBlock) error {
	scriptWriter := NewDirScriptWriter(options.OutputDirectory)
	for _, script := range scripts {
		_, err := scriptWriter.WriteScript(script)
		if err != nil {
			return err
		}
	}

	return nil
}
