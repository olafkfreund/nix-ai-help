package utils

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
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
