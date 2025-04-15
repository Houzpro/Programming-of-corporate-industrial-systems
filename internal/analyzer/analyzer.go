package analyzer

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// FileAnalysis holds the results of the file analysis.
type FileAnalysis struct {
	Filename  string
	WordCount int
	CharCount int
	LineCount int
}

// AnalyzeFile analyzes the contents of a given file and returns a FileAnalysis struct.
func AnalyzeFile(filename string, wg *sync.WaitGroup, mu *sync.Mutex, results *[]FileAnalysis) {
	defer wg.Done()

	file, err := os.Open(filename)
	if err != nil {
		return // Handle error (e.g., log it)
	}
	defer file.Close()

	var wordCount, charCount, lineCount int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		charCount += len(line)
		wordCount += len(strings.Fields(line))
	}

	mu.Lock()
	*results = append(*results, FileAnalysis{
		Filename:  filename,
		WordCount: wordCount,
		CharCount: charCount,
		LineCount: lineCount,
	})
	mu.Unlock()
}
