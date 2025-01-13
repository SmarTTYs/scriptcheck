package runtime

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"scriptcheck/reader"
)

func NewReportWriter(options *Options) io.StringWriter {
	switch options.OutputFile {
	case StdoutOutput:
		return &StdoutWriter{}
	default:
		return &FileWriter{
			fileName: options.OutputFile,
		}
	}
}

func NewTempDirScriptWriter(tempDir string) ScriptWriter {
	return &TempScriptWriter{
		directory: tempDir,
	}
}

func NewDirScriptWriter(dir string) ScriptWriter {
	return &DirScriptWriter{
		directory: dir,
	}
}

type ScriptWriter interface {
	WriteScript(script reader.ScriptBlock) (*os.File, error)
}

type TempScriptWriter struct {
	ScriptWriter

	directory string
}

type DirScriptWriter struct {
	ScriptWriter

	directory string
}

func (w *DirScriptWriter) WriteScript(script reader.ScriptBlock) (*os.File, error) {
	filePath := path.Join(w.directory, script.OutputFileName())

	// create nested directories
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	// write into file
	writeErr := writeScriptBlock(file, script)
	return file, writeErr
}

func (w *TempScriptWriter) WriteScript(script reader.ScriptBlock) (*os.File, error) {
	tempF, err := os.CreateTemp(w.directory, "script-*")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %s", err.Error())
	}

	defer func() {
		_ = tempF.Close()
	}()

	writeErr := writeScriptBlock(tempF, script)
	return tempF, writeErr
}

type StdoutWriter struct{}

func (StdoutWriter) WriteString(s string) (int, error) {
	return os.Stdout.WriteString(s)
}

type FileWriter struct {
	fileName string
}

func (writer FileWriter) WriteString(s string) (int, error) {
	file, err := os.Create(writer.fileName)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = file.Close()
	}()

	return file.WriteString(s)
}

func writeScriptBlock(writer io.StringWriter, script reader.ScriptBlock) error {
	scriptString := script.ScriptString()
	if _, err := writer.WriteString(scriptString); err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}

	return nil
}
