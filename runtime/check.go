package runtime

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"os/exec"
	"scriptcheck/format"
	"scriptcheck/reader"
	"slices"
)

func removeIntermediateScripts(path string) {
	log.Printf("Removing intermediate directory %s", path)
	_ = os.RemoveAll(path)
}

func CheckFiles(options *Options, globPatterns []string) error {
	scripts, files, err := collectAndExtractScripts(options, globPatterns)
	if err != nil {
		return err
	}

	log.Printf(
		"Checking %s script(s) from %s file(s)...\n",
		format.Color(len(scripts), format.Bold),
		format.Color(len(files), format.Bold),
	)

	return checkScripts(options, scripts)
}

func checkScripts(options *Options, scripts []reader.ScriptBlock) error {
	tempDir, fileScriptBlockMap, err := writeTempFiles(options, scripts)
	if err != nil {
		return err
	}

	defer removeIntermediateScripts(*tempDir)

	fileNames := slices.Collect(maps.Keys(fileScriptBlockMap))
	err = runShellcheck(options, fileNames, fileScriptBlockMap)
	if err != nil {
		return err
	}

	log.Printf("Successfully checked files!")
	return nil
}

type ScriptCheckError struct {
	reports []format.ScriptCheckReport
}

func (e ScriptCheckError) ReportCount() int {
	return len(e.reports)
}

func (e ScriptCheckError) Error() string {
	return fmt.Sprintf("Found %d issues", len(e.reports))
}

func runShellcheck(options *Options, fileNames []string, scriptMap map[string]reader.ScriptBlock) error {
	report, err := executeShellCheckCommand(scriptMap, options, fileNames)
	if err != nil {
		return fmt.Errorf("unable to parse shellcheck report: %w", err)
	}

	if len(report) == 0 {
		return nil
	}

	formatter := format.NewFormatter(options.Format)
	reportString, err := formatter.Format(report)
	if err != nil {
		return fmt.Errorf("unable to format shellcheck report: %w", err)
	}

	writer := NewReportWriter(options)
	if writeErr := writeReport(writer, reportString); writeErr != nil {
		return writeErr
	}

	return &ScriptCheckError{report}
}

func executeShellCheckCommand(scriptMap map[string]reader.ScriptBlock, options *Options, fileNames []string) ([]format.ScriptCheckReport, error) {
	out := new(bytes.Buffer)
	cmd := exec.Command("shellcheck", fileNames...)
	cmd.Dir, _ = os.Getwd()
	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	// append provided shell args
	for _, arg := range options.ShellCheckArgs {
		cmd.Args = append(cmd.Args, "--"+arg)
	}

	// always force json format in order to parse it afterward
	cmd.Args = append(cmd.Args, "--format", "json")

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	if runErr := cmd.Run(); runErr != nil {
		var exitError *exec.ExitError
		if errors.As(runErr, &exitError) && exitError.ExitCode() == 2 {
			return nil, runErr
		}
		return format.NewScriptCheckReport(out.Bytes(), scriptMap)
	} else {
		// nothing to do in this case
		return nil, nil
	}
}

func writeReport(writer io.StringWriter, report string) error {
	_, err := writer.WriteString(report)
	if err != nil {
		return fmt.Errorf("unable to write shellcheck output: %w", err)
	}

	return nil
}

func writeTempFiles(options *Options, scripts []reader.ScriptBlock) (*string, map[string]reader.ScriptBlock, error) {
	tempDir, err := os.MkdirTemp("", "scripts")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}

	if options.Debug {
		log.Printf(
			"Created intermediate directory %s",
			format.Color(tempDir, format.Bold),
		)
	}

	writer := NewTempDirWriter(tempDir)

	var fileNames = make(map[string]reader.ScriptBlock)
	for _, script := range scripts {
		file, err := writer.WriteScript(script)
		if err != nil {
			return &tempDir, nil, fmt.Errorf("unable to create temporary file %w", err)
		}
		fileNames[file.Name()] = script
	}

	return &tempDir, fileNames, nil
}
