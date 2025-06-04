package community

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// GitHubClient handles integration with GitHub for community configurations
type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string // Optional, for authenticated requests
	logger     *logger.Logger
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	Description string     `json:"description"`
	URL         string     `json:"html_url"`
	CloneURL    string     `json:"clone_url"`
	Language    string     `json:"language"`
	Stars       int        `json:"stargazers_count"`
	Forks       int        `json:"forks_count"`
	Topics      []string   `json:"topics"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Owner       GitHubUser `json:"owner"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"html_url"`
}

// GitHubSearchResponse represents GitHub API search response
type GitHubSearchResponse struct {
	TotalCount        int                `json:"total_count"`
	IncompleteResults bool               `json:"incomplete_results"`
	Items             []GitHubRepository `json:"items"`
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient(apiToken string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  "https://api.github.com",
		apiToken: apiToken,
		logger:   logger.NewLoggerWithLevel("info"),
	}
}

// SearchRepositories searches for NixOS-related repositories
func (gc *GitHubClient) SearchRepositories(query string, maxResults int) ([]GitHubRepository, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	// Build search query
	searchQuery := fmt.Sprintf("nixos %s in:name,description,readme", query)

	params := url.Values{}
	params.Set("q", searchQuery)
	params.Set("sort", "stars")
	params.Set("order", "desc")
	params.Set("per_page", strconv.Itoa(maxResults))

	searchURL := fmt.Sprintf("%s/search/repositories?%s", gc.baseURL, params.Encode())

	gc.logger.Debug(fmt.Sprintf("Searching GitHub repositories: query=%s, url=%s", searchQuery, searchURL))

	req, err := http.NewRequestWithContext(context.Background(), "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token is provided
	if gc.apiToken != "" {
		req.Header.Set("Authorization", "token "+gc.apiToken)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixai-community-client")

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var searchResponse GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	gc.logger.Info(fmt.Sprintf("GitHub search completed: found %d total, returned %d", searchResponse.TotalCount, len(searchResponse.Items)))

	return searchResponse.Items, nil
}

// GetRepository fetches detailed information about a specific repository
func (gc *GitHubClient) GetRepository(owner, repo string) (*GitHubRepository, error) {
	repoURL := fmt.Sprintf("%s/repos/%s/%s", gc.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(context.Background(), "GET", repoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if gc.apiToken != "" {
		req.Header.Set("Authorization", "token "+gc.apiToken)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixai-community-client")

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var repository GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repository, nil
}

// GetFileContent fetches the content of a file from a repository
func (gc *GitHubClient) GetFileContent(owner, repo, path string) (string, error) {
	contentURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s", gc.baseURL, owner, repo, path)

	req, err := http.NewRequestWithContext(context.Background(), "GET", contentURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if gc.apiToken != "" {
		req.Header.Set("Authorization", "token "+gc.apiToken)
	}

	req.Header.Set("Accept", "application/vnd.github.v3.raw")
	req.Header.Set("User-Agent", "nixai-community-client")

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(content), nil
}

// SearchNixOSConfigurations searches specifically for NixOS configuration files
func (gc *GitHubClient) SearchNixOSConfigurations(topic string) ([]Configuration, error) {
	repos, err := gc.SearchRepositories(topic, 20)
	if err != nil {
		return nil, err
	}

	var configs []Configuration

	for _, repo := range repos {
		// Skip if not related to NixOS
		if !gc.isNixOSRelated(repo) {
			continue
		}

		config := Configuration{
			ID:          fmt.Sprintf("github_%d", repo.ID),
			Name:        repo.Name,
			Author:      repo.Owner.Login,
			Description: repo.Description,
			Tags:        repo.Topics,
			Rating:      gc.calculateRating(repo),
			Downloads:   repo.Forks, // Use forks as download metric
			Views:       repo.Stars,
			URL:         repo.URL,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
			Language:    "nix",
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// Helper methods

func (gc *GitHubClient) isNixOSRelated(repo GitHubRepository) bool {
	// Check if repository is related to NixOS
	keywords := []string{"nixos", "nix", "configuration", "flake"}

	repoText := strings.ToLower(repo.Name + " " + repo.Description)

	for _, keyword := range keywords {
		if strings.Contains(repoText, keyword) {
			return true
		}
	}

	// Check topics
	for _, topic := range repo.Topics {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(topic), keyword) {
				return true
			}
		}
	}

	return false
}

func (gc *GitHubClient) calculateRating(repo GitHubRepository) float64 {
	// Simple rating calculation based on stars and activity
	stars := float64(repo.Stars)

	// Base rating from stars (logarithmic scale)
	rating := 1.0
	if stars > 0 {
		rating = 1.0 + (4.0 * (stars / (stars + 100.0))) // Asymptotic to 5.0
	}

	// Bonus for recent activity
	daysSinceUpdate := time.Since(repo.UpdatedAt).Hours() / 24
	if daysSinceUpdate < 30 {
		rating += 0.2
	}

	// Cap at 5.0
	if rating > 5.0 {
		rating = 5.0
	}

	return rating
}
