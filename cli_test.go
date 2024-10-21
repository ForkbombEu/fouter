package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/ForkbombEu/fouter/fouter"
	slangroom "github.com/dyne/slangroom-exec/bindings/go"
	"github.com/spf13/cobra"
)

func TestMainCLI(t *testing.T) {

	var rootCmd = &cobra.Command{
		Use:   "fouter",
		Short: "A CLI tool for running slang files",
	}
	err := fouter.CreateFileRouter("", &embeddedFiles, "contracts", func(file fouter.SlangFile) {
		input := slangroom.SlangroomInput{Contract: file.Content}
		createCommand(rootCmd, file.Dir, file.FileName, input)
	})
	if err != nil {
		t.Fatalf("Error setting up file router: %v", err)
	}

	tests := []struct {
		command string
		output  string
	}{
		{"contracts hello", "hello"},
		{"contracts test", "timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			rootCmd.SetArgs(strings.Split(tt.command, " "))
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}
			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r)
			output := buf.String()
			if !strings.Contains(output, tt.output) {
				t.Errorf("Expected output to contain %q, got %q", tt.output, output)
			}
		})
	}
}
