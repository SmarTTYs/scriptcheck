package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"scriptcheck/reader"
	"strings"
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

type ScriptWriter interface {
	io.StringWriter

	WriteScript(script reader.ScriptBlock) (*os.File, error)
}

func NewDirWriter(dir, defaultShell string) ScriptWriter {
	return &TempScriptWriter{
		directory:    dir,
		defaultShell: defaultShell,
	}
}

func NewTempDirWriter(dir, defaultShell string) ScriptWriter {
	/*
		test := &DirScriptWriter{
			fileCreator: func(s string) (*os.File, error) {
				return os.Create(s)
			},
		}

		test2 := &DirScriptWriter{
			fileCreator: func(s string) (*os.File, error) {
				return os.CreateTemp(dir, "script-*")
			},
		}
	*/

	return &TempScriptWriter{
		directory:    dir,
		defaultShell: defaultShell,
	}
}

type TempScriptWriter struct {
	ScriptWriter

	directory    string
	defaultShell string
}

type DirScriptWriter struct {
	ScriptWriter

	directory    string
	defaultShell string

	fileCreator func(string) (*os.File, error)
}

func (w *DirScriptWriter) WriteScript(script reader.ScriptBlock) (*os.File, error) {
	filePath := script.GetOutputFilePath(w.directory)

	// create nested directories
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// todo: this could be done by using the new filewriter
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %s", err.Error())
	}

	defer func() {
		_ = file.Close()
	}()

	// write into file
	err = writeScriptBlock(file, w.defaultShell, script)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (w *TempScriptWriter) WriteScript(script reader.ScriptBlock) (*os.File, error) {
	tempF, err := os.CreateTemp(w.directory, "script-*")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %s", err.Error())
	}

	writeErr := writeScriptBlock(tempF, w.defaultShell, script)
	if writeErr != nil {
		return nil, writeErr
	}

	return tempF, nil
}

func (w *TempScriptWriter) WriteString(_ string) (n int, err error) {
	return 0, err
}

type FileWriter struct {
	fileName string
}

type StdoutWriter struct{}

func (StdoutWriter) WriteString(s string) (int, error) {
	return os.Stdout.WriteString(s)
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

func writeScriptBlock(writer io.StringWriter, defaultShell string, script reader.ScriptBlock) error {
	scriptString := scriptBlockString(script, defaultShell)
	if _, err := writer.WriteString(scriptString); err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}

	return nil
}

func scriptBlockString(script reader.ScriptBlock, defaultShell string) string {
	builder := new(strings.Builder)
	if !script.Script.HasShell() {
		var scriptShell string
		if len(script.Shell) > 0 {
			scriptShell = script.Shell
		} else {
			scriptShell = defaultShell
		}

		builder.WriteString(fmt.Sprintf("# shellcheck shell=%s\n", scriptShell))
	}

	builder.WriteString(script.ScriptString())
	return builder.String()
}
