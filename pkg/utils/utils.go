package utils

import (
	"errors"
	"os"
	"strings"
)

// IsFile checks if the given path is a file.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirectory checks if the given path is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// SplitLines splits a string into a slice of lines.
func SplitLines(input string) []string {
	return strings.Split(strings.TrimSpace(input), "\n")
}

// Contains checks if a slice of strings contains a specific string.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidatePath checks if the provided path is valid and returns an error if not.
func ValidatePath(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}
	if !IsFile(path) && !IsDirectory(path) {
		return errors.New("path does not exist")
	}
	return nil
}
