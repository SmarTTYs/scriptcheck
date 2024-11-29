package cmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestExtractCommand(t *testing.T) {
	_, err := ExecuteCommand(rootCmd, "extract", "../dir/first_yaml.yml")
	if err != nil {
		t.Errorf("Did not expect an error")
	}
}

func TestUnknownFile(T *testing.T) {
	consoleOutput, err := ExecuteCommand(rootCmd, "extract", "../unknown/first_yaml.yml")
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			println("is exit error with code %s", exitError.ExitCode())
			os.Exit(exitError.ExitCode())
		} else {
			log.Printf("error %s", err)
			log.Printf("test...")
			log.Printf("Error while running: %s\n", err.Error())
			os.Exit(1)
		}
	}
	println("test1")
	println("error", err)
	println(consoleOutput)
	println("test2")
}
