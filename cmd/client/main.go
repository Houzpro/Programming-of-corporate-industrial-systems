package main

import (
	"fmt"
	"os"
	"pr2/internal/utils"
	"pr2/pkg/network"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <file1> <file2> ...")
		return
	}

	serverAddress := "http://localhost:8080"

	// Collect valid files
	var validFiles []string
	for _, filePath := range os.Args[1:] {
		if !utils.FileExists(filePath) {
			fmt.Printf("File not found: %s\n", filePath)
			continue
		}
		validFiles = append(validFiles, filePath)
	}

	if len(validFiles) == 0 {
		fmt.Println("No valid files to process.")
		return
	}

	// Process files in batch
	processFiles(validFiles, serverAddress)
}

func processFiles(filePaths []string, serverAddress string) {
	// Use the batch upload function
	response, err := network.UploadFilesToServer(serverAddress, filePaths)
	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
		return
	}

	// Display results
	fmt.Println("Analysis results:")
	for _, result := range response.Results {
		fmt.Printf("\nAnalysis for %s:\n", result.Filename)
		fmt.Printf("  Words: %d\n", result.WordCount)
		fmt.Printf("  Characters: %d\n", result.CharacterCount)
		fmt.Printf("  Lines: %d\n", result.LineCount)
	}
}

// For backwards compatibility and single file processing
func processFile(filePath, serverAddress string) {
	// Upload file to server
	response, err := network.UploadFileToServer(serverAddress, filePath)
	if err != nil {
		fmt.Printf("Error processing file %s: %v\n", filePath, err)
		return
	}

	// Display results
	fmt.Printf("Analysis for %s:\n", response.Filename)
	fmt.Printf("  Words: %d\n", response.WordCount)
	fmt.Printf("  Characters: %d\n", response.CharacterCount)
	fmt.Printf("  Lines: %d\n", response.LineCount)
	fmt.Println()
}
