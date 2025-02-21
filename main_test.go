package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan bool)
	go func() {
		_, _ = io.Copy(&buf, r)
		done <- true
	}()

	f()
	_ = w.Close()
	os.Stdout = oldStdout
	<-done

	return buf.String()
}

func callMain() (int, string) {
	var exitCode int
	oldExit := exit
	defer func() { exit = oldExit }()
	exit = func(code int) {
		exitCode = code
		panic("exit")
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main in a goroutine
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if r != "exit" {
					panic(r)
				}
			}
			done <- true
		}()
		RealMain()
	}()

	// Copy output in another goroutine
	outputDone := make(chan bool)
	go func() {
		_, _ = io.Copy(&buf, r)
		outputDone <- true
	}()

	// Wait for main to finish
	<-done
	w.Close()
	os.Stdout = oldStdout
	<-outputDone

	return exitCode, buf.String()
}

func TestMain(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name           string
		args           []string
		expectedExit   int
		expectedOutput string
	}{
		{
			name:           "no arguments",
			args:           []string{"cheeseburger"},
			expectedExit:   1,
			expectedOutput: "Usage: cheeseburger <command>",
		},
		{
			name:           "help command",
			args:           []string{"cheeseburger", "help"},
			expectedExit:   0,
			expectedOutput: "Usage: cheeseburger <command> [options]",
		},
		{
			name:           "version command",
			args:           []string{"cheeseburger", "version"},
			expectedExit:   0,
			expectedOutput: "cheeseburger version " + CliVersion,
		},
		{
			name:           "unknown command",
			args:           []string{"cheeseburger", "unknown"},
			expectedExit:   1,
			expectedOutput: "Unknown command: unknown",
		},
		{
			name:           "serve without directory",
			args:           []string{"cheeseburger", "serve"},
			expectedExit:   1,
			expectedOutput: "Error: static directory path required for serve command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test args
			os.Args = tt.args

			exitCode, output := callMain()

			// Verify output and exit code
			assert.Contains(t, output, tt.expectedOutput)
			if tt.expectedExit > 0 {
				assert.Equal(t, tt.expectedExit, exitCode)
			}
		})
	}
}

func TestPrintHelp(t *testing.T) {
	output := captureOutput(func() {
		printHelp()
	})

	// Verify help text contains all commands
	assert.Contains(t, output, "Usage: cheeseburger")
	assert.Contains(t, output, "help")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "vanity")
	assert.Contains(t, output, "serve")
	assert.Contains(t, output, "mvc")

	// Verify MVC commands are listed
	assert.Contains(t, output, "serve [--vanity-name")
	assert.Contains(t, output, "clean")
	assert.Contains(t, output, "init")
	assert.Contains(t, output, "backup")
	assert.Contains(t, output, "restore")
}

// Mock os.Exit to prevent test termination
var osExit = os.Exit
