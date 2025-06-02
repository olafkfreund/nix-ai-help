package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// DirExists checks if the given path exists and is a directory.
func DirExists(path string) bool {
	return IsDirectory(path)
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

// ExpandHome expands the '~/' prefix in a path to the user's home directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

// GetConfigDir returns the config directory for nixai, respecting XDG_CONFIG_HOME or defaulting to $HOME/.config/nixai
func GetConfigDir() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return filepath.Join(xdg, "nixai"), nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".config", "nixai"), nil
}

// GenerateID generates a unique ID for community resources
func GenerateID() string {
	return fmt.Sprintf("nixai_%d_%s", time.Now().Unix(), randomString(8))
}

// ParseTags parses a comma-separated tag string into a slice
func ParseTags(tagStr string) []string {
	if tagStr == "" {
		return []string{}
	}

	tags := strings.Split(tagStr, ",")
	var result []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ParseFloat safely parses a string to float64, returning 0 on error
func ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ParseInt safely parses a string to int, returning 0 on error
func ParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// FindJSONStart finds the starting position of JSON content in a string
func FindJSONStart(s string) int {
	return strings.Index(s, "{")
}

// FindJSONEnd finds the ending position of JSON content starting from a given position
func FindJSONEnd(s string, start int) int {
	if start < 0 || start >= len(s) {
		return -1
	}

	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
