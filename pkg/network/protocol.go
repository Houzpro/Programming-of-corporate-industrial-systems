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

type Response struct {
	Filename       string `json:"filename"`
	WordCount      int    `json:"word_count"`
	CharacterCount int    `json:"character_count"`
	LineCount      int    `json:"line_count"`
}

type BatchResponse struct {
	Results []Response `json:"results"`
}

func UploadFileToServer(serverURL string, filePath string) (*Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", serverURL+"/analyze", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func UploadFilesToServer(serverURL string, filePaths []string) (*BatchResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		part, err := writer.CreateFormFile("files", filepath.Base(filePath))
		if err != nil {
			file.Close()
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			file.Close()
			return nil, err
		}

		file.Close()
	}
	writer.Close()

	req, err := http.NewRequest("POST", serverURL+"/analyze-batch", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var batchResponse BatchResponse
	err = json.NewDecoder(resp.Body).Decode(&batchResponse)
	if err != nil {
		return nil, err
	}

	return &batchResponse, nil
}
