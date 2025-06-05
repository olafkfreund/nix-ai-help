package packaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitCloner handles cloning Git repositories for analysis
type GitCloner struct {
	tempDir string
}

// NewGitCloner creates a new Git cloner
func NewGitCloner(tempDir string) *GitCloner {
	return &GitCloner{
		tempDir: tempDir,
	}
}

// CloneRepository clones a Git repository to a temporary directory
func (gc *GitCloner) CloneRepository(repoURL string) (string, error) {
	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(gc.tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract repo name from URL
	repoName := extractRepoNameFromURL(repoURL)
	if repoName == "" {
		return "", fmt.Errorf("could not extract repository name from URL: %s", repoURL)
	}

	// Create target directory
	targetDir := filepath.Join(gc.tempDir, repoName)

	// Remove existing directory if it exists
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.RemoveAll(targetDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Clone the repository
	// #nosec G204 -- repoURL and targetDir are validated/trusted or controlled by CLI logic
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return targetDir, nil
}

// CloneRepositoryQuiet clones a repository without output
func (gc *GitCloner) CloneRepositoryQuiet(repoURL string) (string, error) {
	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(gc.tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract repo name from URL
	repoName := extractRepoNameFromURL(repoURL)
	if repoName == "" {
		return "", fmt.Errorf("could not extract repository name from URL: %s", repoURL)
	}

	// Create target directory
	targetDir := filepath.Join(gc.tempDir, repoName)

	// Remove existing directory if it exists
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.RemoveAll(targetDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Clone the repository quietly
	// #nosec G204 -- repoURL and targetDir are validated/trusted or controlled by CLI logic
	cmd := exec.Command("git", "clone", "--depth", "1", "--quiet", repoURL, targetDir)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return targetDir, nil
}

// Cleanup removes the temporary directory
func (gc *GitCloner) Cleanup() error {
	if gc.tempDir != "" {
		return os.RemoveAll(gc.tempDir)
	}
	return nil
}

// extractRepoNameFromURL extracts the repository name from a Git URL
func extractRepoNameFromURL(repoURL string) string {
	// Handle different URL formats
	url := strings.TrimSpace(repoURL)

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Extract the last part of the path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// IsGitRepository checks if the given path is a Git repository
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// GetRepositoryInfo extracts basic information from a Git repository
func GetRepositoryInfo(repoPath string) (*RepositoryInfo, error) {
	if !IsGitRepository(repoPath) {
		return nil, fmt.Errorf("not a Git repository: %s", repoPath)
	}

	info := &RepositoryInfo{
		Path: repoPath,
	}

	// Get remote origin URL
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	if output, err := cmd.Output(); err == nil {
		info.RemoteURL = strings.TrimSpace(string(output))
	}

	// Get current branch
	cmd = exec.Command("git", "-C", repoPath, "branch", "--show-current")
	if output, err := cmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(output))
	}

	// Get latest commit hash
	cmd = exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	if output, err := cmd.Output(); err == nil {
		info.CommitHash = strings.TrimSpace(string(output))
	}

	// Get commit count
	cmd = exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD")
	if output, err := cmd.Output(); err == nil {
		info.CommitCount = strings.TrimSpace(string(output))
	}

	return info, nil
}

// RepositoryInfo contains basic Git repository information
type RepositoryInfo struct {
	Path        string `json:"path"`
	RemoteURL   string `json:"remote_url"`
	Branch      string `json:"branch"`
	CommitHash  string `json:"commit_hash"`
	CommitCount string `json:"commit_count"`
}
