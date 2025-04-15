package utils

import (
	"io/ioutil"
	"os"
)

// ReadFileContents reads the contents of a file and returns it as a string.
// It returns an error if the file cannot be read.
func ReadFileContents(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FileExists checks if a file exists at the given path.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
