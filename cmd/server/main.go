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
	// Create a temporary directory for file uploads
	tempDir := filepath.Join(os.TempDir(), "file-analyzer")
	os.MkdirAll(tempDir, 0755)

	// Set up HTTP server with file upload handler
	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the multipart form (10 MB max)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}

		// Get the file from the form
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the content
		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		// Process file content
		contentStr := string(content)
		wordCount := len(strings.Fields(contentStr))
		charCount := len(contentStr)
		lineCount := len(strings.Split(contentStr, "\n"))

		// Create response
		response := network.Response{
			Filename:       handler.Filename,
			WordCount:      wordCount,
			CharacterCount: charCount,
			LineCount:      lineCount,
		}

		// Send JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	// Add batch analysis endpoint
	http.HandleFunc("/analyze-batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the multipart form (50 MB max)
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}

		// Get all files from the form
		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			http.Error(w, "No files provided", http.StatusBadRequest)
			return
		}

		// Create a slice to store results
		results := make([]network.Response, 0, len(files))
		var wg sync.WaitGroup
		var mu sync.Mutex

		// Process each file concurrently
		for _, fileHeader := range files {
			wg.Add(1)
			go func(header *multipart.FileHeader) {
				defer wg.Done()

				// Open the file
				file, err := header.Open()
				if err != nil {
					fmt.Printf("Error opening file %s: %v\n", header.Filename, err)
					return
				}
				defer file.Close()

				// Read the content
				content, err := io.ReadAll(file)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", header.Filename, err)
					return
				}

				// Process file content
				contentStr := string(content)
				wordCount := len(strings.Fields(contentStr))
				charCount := len(contentStr)
				lineCount := len(strings.Split(contentStr, "\n"))

				// Create response
				fileResult := network.Response{
					Filename:       header.Filename,
					WordCount:      wordCount,
					CharacterCount: charCount,
					LineCount:      lineCount,
				}

				// Add to results with mutex lock
				mu.Lock()
				results = append(results, fileResult)
				mu.Unlock()
			}(fileHeader)
		}

		// Wait for all files to be processed
		wg.Wait()

		// Create batch response
		batchResponse := network.BatchResponse{
			Results: results,
		}

		// Send JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(batchResponse); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	// Add a simple status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("File Analyzer Server is running"))
	})

	// Start HTTP server
	port := ":8080"
	fmt.Printf("Server is listening on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// AnalyzeFiles analyzes multiple files using the analyzer package
func AnalyzeFiles(filenames []string) []models.FileAnalysis {
	var results []analyzer.FileAnalysis
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, filename := range filenames {
		wg.Add(1)
		go analyzer.AnalyzeFile(filename, &wg, &mu, &results)
	}

	wg.Wait()

	// Convert analyzer.FileAnalysis to models.FileAnalysis
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
