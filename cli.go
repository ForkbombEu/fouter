package main

import (
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ForkbombEu/fouter/fouter"

	slangroom "github.com/dyne/slangroom-exec/bindings/go"
	"github.com/spf13/cobra"
)

//go:embed contracts
var embeddedFiles embed.FS

// Function to dynamically create commands for each .slang file
func createCommand(rootCmd *cobra.Command, relativePath, fileName string, content slangroom.SlangroomInput) {
	parts := strings.Split(relativePath, string(filepath.Separator))
	commandName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	parentCmd := rootCmd
	for _, part := range parts {
		var foundCmd *cobra.Command
		for _, cmd := range parentCmd.Commands() {
			if cmd.Name() == part {
				foundCmd = cmd
				break
			}
		}

		if foundCmd == nil {
			foundCmd = &cobra.Command{
				Use:   part,
				Short: fmt.Sprintf("Folder: %s", part),
			}
			parentCmd.AddCommand(foundCmd)
		}

		parentCmd = foundCmd
	}

	parentCmd.AddCommand(&cobra.Command{
		Use:   commandName,
		Short: fmt.Sprintf("Run %s", fileName),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := slangroom.Exec(content)
			if err != nil {
				fmt.Println(res.Logs)
			} else {
				fmt.Println(res.Output)
			}
		},
	})
}

func main() {
	// Create the root command
	var rootCmd = &cobra.Command{
		Use:   "fouter",
		Short: "A CLI tool for running slang files",
	}

	err := fouter.CreateFileRouter("", &embeddedFiles, "contracts", func(file fouter.SlangFile) {

		input := slangroom.SlangroomInput{Contract: file.Content}
		createCommand(rootCmd, file.Dir, file.FileName, input)
	})

	if err != nil {
		log.Fatalf("Error setting up file router: %v", err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
