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

type ScriptWriter interface {
	io.StringWriter

	WriteScriptTest(directory string, script reader.ScriptBlock) (*os.File, error)
	WriteScript(script reader.ScriptBlock) (*os.File, error)
}

func NewDirWriter(dir string) ScriptWriter {
	return &TempScriptWriter{
		directory: dir,
	}
}

func NewTempDirWriter(dir string) ScriptWriter {
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
		directory: dir,
	}
}

type TempScriptWriter struct {
	ScriptWriter

	directory string
}

type DirScriptWriter struct {
	ScriptWriter

	directory   string
	fileCreator func(string) (*os.File, error)
}

func (w *DirScriptWriter) WriteScriptTest(directory string, script reader.ScriptBlock) (*os.File, error) {
	w.directory = directory
	return w.WriteScript(script)
}

func (w *DirScriptWriter) WriteScript(script reader.ScriptBlock) (*os.File, error) {
	println("Directory", w.directory)
	filePath := path.Join(w.directory, script.OutputFileName())

	// create nested directories
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %s", err.Error())
	}

	defer func() {
		_ = file.Close()
	}()

	// write into file
	err = writeScriptBlock(file, script)
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

	writeErr := writeScriptBlock(tempF, script)
	return tempF, writeErr
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

func writeScriptBlock(writer io.StringWriter, script reader.ScriptBlock) error {
	scriptString := script.ScriptString()
	if _, err := writer.WriteString(scriptString); err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}

	return nil
}
