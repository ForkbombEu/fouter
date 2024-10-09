package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSlangFiles(t *testing.T) {
	// Sample slang files data for testing
	slangFiles := map[string][]string{
		"testdir":  {"file1.slang", "file2.slang"},
		"testdir2": {"file3.slang"},
	}

	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/slang/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create the handler
	handler := listSlangFilesHandler(slangFiles)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check that the response contains the expected content
	assert.Contains(t, rr.Body.String(), "<h1>Available contract files:</h1>")
	assert.Contains(t, rr.Body.String(), "<h2>Directory: testdir</h2>")
	assert.Contains(t, rr.Body.String(), "<h2>Directory: testdir2</h2>")
	assert.Contains(t, rr.Body.String(), "<a href=\"/slang/testdir/file1.slang\">testdir/file1.slang</a>")
	assert.Contains(t, rr.Body.String(), "<a href=\"/slang/testdir/file2.slang\">testdir/file2.slang</a>")
	assert.Contains(t, rr.Body.String(), "<a href=\"/slang/testdir2/file3.slang\">testdir2/file3.slang</a>")
}

func TestSlangFilePage(t *testing.T) {
	// Set up a test slang file
	testFileName := "testfile.slang"
	testFileContent := "Given nothing\nThen print the string 'Test Successful'"
	_ = os.WriteFile(testFileName, []byte(testFileContent), 0644)
	defer os.Remove(testFileName) // Cleanup after test

	req, err := http.NewRequest("GET", "/slang/"+testFileName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := slangFilePageHandler(testFileName)

	handler.ServeHTTP(rr, req)

	// Check the response code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check if the file content is in the response
	assert.Contains(t, rr.Body.String(), testFileContent)

	// Check for the execution button in the response
	assert.Contains(t, rr.Body.String(), `<button type="submit">Execute testfile.slang</button>`)
}

func TestExecuteSlangFile(t *testing.T) {
	// Set up a test slang file
	testFileName := "testexecfile.slang"
	testFileContent := `Given nothing
Then print the string 'Execution Successful'`
	_ = os.WriteFile(testFileName, []byte(testFileContent), 0644)
	defer os.Remove(testFileName) // Cleanup after test

	// Simulate a POST request to execute the slang file
	req, err := http.NewRequest("POST", "/slang/execute/"+testFileName, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := executeSlangFileHandler(testFileName)

	handler.ServeHTTP(rr, req)

	// Check the response code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the output of the execution (adapt based on your expected output)
	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err, "Failed to unmarshal output")
	assert.Contains(t, result["output"], "Execution_Successful")
}

func TestInvalidRequestMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/slang/execute/testfile.slang", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := executeSlangFileHandler("testfile.slang") // Assuming a dummy file for handler

	handler.ServeHTTP(rr, req)

	// Check the response code for invalid method
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request method")
}
