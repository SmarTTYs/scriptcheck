package runtime

import (
	"log"
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
	scriptWriter := NewDirScriptWriter(options.OutputDirectory)
	for _, script := range scripts {
		_, err := scriptWriter.WriteScript(script)
		if err != nil {
			return err
		}
	}

	return nil
}
