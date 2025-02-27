package main

import (
	"testing"
)

func TestSearchPhrase(t *testing.T) {
	content := "hello world hello"
	phrase := "hello world"
	totalWords, phraseCount := searchPhrase(content, phrase)

	if totalWords != 3 {
		t.Errorf("Expected total words to be 3, got %d", totalWords)
	}

	if phraseCount != 1 {
		t.Errorf("Expected phrase count to be 1, got %d", phraseCount)
	}
}

func TestSearchPhraseNotFound(t *testing.T) {
	content := "hello world"
	phrase := "test phrase"
	totalWords, phraseCount := searchPhrase(content, phrase)

	if totalWords != 2 {
		t.Errorf("Expected total words to be 2, got %d", totalWords)
	}

	if phraseCount != 0 {
		t.Errorf("Expected phrase count to be 0, got %d", phraseCount)
	}
}

func TestReadFile(t *testing.T) {
	filePath := "test.txt"
	content, err := readFile(filePath)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	expectedContent := "hello world hello again this is a test file hello world hello "
	if content != expectedContent {
		t.Errorf("Expected content to be '%s', got '%s'", expectedContent, content)
	}
}

func TestMultipleOccurrences(t *testing.T) {
	content := "hello world hello world hello"
	phrase := "hello world"
	_, phraseCount := searchPhrase(content, phrase)

	if phraseCount != 2 {
		t.Errorf("Expected phrase count to be 2, got %d", phraseCount)
	}
}
