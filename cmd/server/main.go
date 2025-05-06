package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"pr2/internal/analyzer"
	"pr2/internal/models"
	"pr2/pkg/network"
	"strings"
	"sync"
)

func main() {
	tempDir := filepath.Join(os.TempDir(), "file-analyzer")
	os.MkdirAll(tempDir, 0755)

	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		contentStr := string(content)
		wordCount := len(strings.Fields(contentStr))
		charCount := len(contentStr)
		lineCount := len(strings.Split(contentStr, "\n"))

		response := network.Response{
			Filename:       handler.Filename,
			WordCount:      wordCount,
			CharacterCount: charCount,
			LineCount:      lineCount,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/analyze-batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			http.Error(w, "No files provided", http.StatusBadRequest)
			return
		}

		results := make([]network.Response, 0, len(files))
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, fileHeader := range files {
			wg.Add(1)
			go func(header *multipart.FileHeader) {
				defer wg.Done()

				file, err := header.Open()
				if err != nil {
					return
				}
				defer file.Close()

				content, err := io.ReadAll(file)
				if err != nil {
					return
				}

				contentStr := string(content)
				wordCount := len(strings.Fields(contentStr))
				charCount := len(contentStr)
				lineCount := len(strings.Split(contentStr, "\n"))

				fileResult := network.Response{
					Filename:       header.Filename,
					WordCount:      wordCount,
					CharacterCount: charCount,
					LineCount:      lineCount,
				}

				mu.Lock()
				results = append(results, fileResult)
				mu.Unlock()
			}(fileHeader)
		}

		wg.Wait()

		batchResponse := network.BatchResponse{
			Results: results,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(batchResponse); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("File Analyzer Server is running"))
	})

	port := ":8080"
	fmt.Printf("Server is listening on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func AnalyzeFiles(filenames []string) []models.FileAnalysis {
	var results []analyzer.FileAnalysis
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, filename := range filenames {
		wg.Add(1)
		go analyzer.AnalyzeFile(filename, &wg, &mu, &results)
	}

	wg.Wait()

	modelResults := make([]models.FileAnalysis, len(results))
	for i, result := range results {
		modelResults[i] = models.FileAnalysis{
			Filename:  result.Filename,
			WordCount: result.WordCount,
			CharCount: result.CharCount,
			LineCount: result.LineCount,
		}
	}

	return modelResults
}
