package network

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Response represents the structure of the response sent back after analysis.
type Response struct {
	Filename       string `json:"filename"`
	WordCount      int    `json:"word_count"`
	CharacterCount int    `json:"character_count"`
	LineCount      int    `json:"line_count"`
}

// BatchResponse represents multiple file analysis responses
type BatchResponse struct {
	Results []Response `json:"results"`
}

// UploadFileToServer uploads a file to the HTTP server for analysis.
func UploadFileToServer(serverURL string, filePath string) (*Response, error) {
	// Create a new file buffer
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create buffer for the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	// Copy file content to form field
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	// Create HTTP request
	req, err := http.NewRequest("POST", serverURL+"/analyze", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// UploadFilesToServer uploads multiple files to the HTTP server for analysis.
func UploadFilesToServer(serverURL string, filePaths []string) (*BatchResponse, error) {
	// Create buffer for the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add each file to the form
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		// Create form file field
		part, err := writer.CreateFormFile("files", filepath.Base(filePath))
		if err != nil {
			file.Close()
			return nil, err
		}

		// Copy file content to form field
		_, err = io.Copy(part, file)
		if err != nil {
			file.Close()
			return nil, err
		}

		file.Close()
	}
	writer.Close()

	// Create HTTP request
	req, err := http.NewRequest("POST", serverURL+"/analyze-batch", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var batchResponse BatchResponse
	err = json.NewDecoder(resp.Body).Decode(&batchResponse)
	if err != nil {
		return nil, err
	}

	return &batchResponse, nil
}
