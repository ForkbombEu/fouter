package fouter

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed test/*
var embeddedFiles embed.FS

// Helper function to create temporary .slang files in the filesystem for testing.
func createTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	filePath := filepath.Join(dir, name)
	err = os.WriteFile(filePath, []byte(content), os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	return filePath
}

func TestCreateFileRouter(t *testing.T) {
	tempDir := t.TempDir()
	rootFilePath := createTestFile(t, tempDir, "root.slang", "root file content")
	subFilePath := createTestFile(t, filepath.Join(tempDir, "subdir"), "subfile.slang", "subdir file content")

	var processedFiles []SlangFile
	handler := func(file SlangFile) {
		processedFiles = append(processedFiles, file)
	}
	err := CreateFileRouter(tempDir, &embeddedFiles, "test", handler)
	if err != nil {
		t.Fatalf("Error in CreateFileRouter: %v", err)
	}
	expectedFiles := []SlangFile{
		{
			Path:       "test/embedded.slang", // Embedded file
			Content:    "test embedded file content",
			FileName:   "embedded.slang",
			Dir:        "test",
			IsEmbedded: true,
		},
		{
			Path:       rootFilePath,
			Content:    "root file content",
			FileName:   "root.slang",
			Dir:        ".",
			IsEmbedded: false,
		},
		{
			Path:       subFilePath,
			Content:    "subdir file content",
			FileName:   "subfile.slang",
			Dir:        "subdir",
			IsEmbedded: false,
		},
	}
	if len(processedFiles) != len(expectedFiles) {
		t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(processedFiles))
	}
	for i, expected := range expectedFiles {
		processed := processedFiles[i]

		if processed.Path != expected.Path {
			t.Errorf("Expected path: %s, got: %s", expected.Path, processed.Path)
		}
		if processed.FileName != expected.FileName {
			t.Errorf("Expected file name: %s, got: %s", expected.FileName, processed.FileName)
		}
		if strings.TrimSpace(processed.Content) != strings.TrimSpace(expected.Content) {
			t.Errorf("Expected content for %s: %s, got: %s", expected.FileName, expected.Content, processed.Content)
		}
		if processed.Dir != expected.Dir {
			t.Errorf("Expected directory for %s: %s, got: %s", expected.FileName, expected.Dir, processed.Dir)
		}
	}
}
