package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	slang "github.com/FilippoTrotter/slangroom-go" // Change this to dyne/slangroom-exec once merged
	"github.com/gorilla/mux"
)

const baseDir = "."

// Function to find all .slang files recursively in the directory and return a map of directories to their files
func findSlangFiles(dir string) (map[string][]string, error) {
	slangFiles := make(map[string][]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".slang") {
			relativePath, _ := filepath.Rel(baseDir, path) // Get the relative path
			dir := filepath.Dir(relativePath)
			slangFiles[dir] = append(slangFiles[dir], relativePath)
		}
		return nil
	})
	return slangFiles, err
}

// Handler to print the content of the .slang file and provide a button to execute it
func slangFilePageHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><body><h1>%s</h1><pre>%s</pre>", filepath.Base(filePath), string(content))

		relativePath, _ := filepath.Rel(baseDir, filePath)
		fmt.Fprintf(w, `<form method="POST" action="/slang/execute/%s">
                            <button type="submit">Execute %s</button>
                        </form>`, relativePath, filepath.Base(filePath))

		fmt.Fprintln(w, "</body></html>")
	}
}

// Handler to execute the .slang file content on a POST request
func executeSlangFileHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		result, success := slang.SlangroomExec("", string(content), "", "", "", "")
		if !success {
			http.Error(w, "Error executing slang file: "+result.Logs, http.StatusInternalServerError)
			return
		}

		output := map[string]interface{}{
			"output": result.Output,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(output); err != nil {
			http.Error(w, "Error encoding output to JSON", http.StatusInternalServerError)
		}
	}
}

// Handler to list all available .slang files, grouped by directory
func listSlangFilesHandler(slangFiles map[string][]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "<html><body><h1>Available contract files:</h1>")

		for dir, files := range slangFiles {
			fmt.Fprintf(w, "<h2>Directory: %s</h2><ul>", dir)

			for _, file := range files {
				fileName := filepath.Base(file)
				link := fmt.Sprintf("/slang/%s/%s", dir, fileName)
				fmt.Fprintf(w, `<li><a href="%s">%s/%s</a></li>`, link, dir, fileName)
			}

			fmt.Fprintln(w, "</ul>")
		}

		fmt.Fprintln(w, "</body></html>")
	}
}

func main() {
	r := mux.NewRouter()

	slangFiles, err := findSlangFiles(baseDir)
	if err != nil {
		log.Fatalf("Error finding .slang files: %v", err)
	}

	r.HandleFunc("/slang/", listSlangFilesHandler(slangFiles)).Methods("GET")

	// For each file, create an API endpoint to show its content and add an execution button
	for _, files := range slangFiles {
		for _, file := range files {
			// Create a handler to show file content and an execute button
			r.HandleFunc(fmt.Sprintf("/slang/%s", file), slangFilePageHandler(file)).Methods("GET")

			// Create a handler to execute the file
			r.HandleFunc(fmt.Sprintf("/slang/execute/%s", file), executeSlangFileHandler(file)).Methods("POST")
		}
	}

	fmt.Println("Starting server on :3000")
	fmt.Println("Access the contract files at: http://localhost:3000/slang/")
	log.Fatal(http.ListenAndServe(":3000", r))
}
