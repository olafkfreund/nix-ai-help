package community

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// Configuration represents a shared NixOS configuration
type Configuration struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Rating      float64   `json:"rating"`
	Downloads   int       `json:"downloads"`
	Views       int       `json:"views"`
	URL         string    `json:"url"`
	FilePath    string    `json:"file_path"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Size        int64     `json:"size"`
	Language    string    `json:"language"`
}

// BestPractice represents a community best practice
type BestPractice struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"` // low, medium, high, critical
	Examples    []string  `json:"examples"`
	Violations  []string  `json:"violations"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TrendData represents trending packages and configurations
type TrendData struct {
	PopularPackages     []PackageStats  `json:"popular_packages"`
	TrendingConfigs     []Configuration `json:"trending_configs"`
	TotalConfigurations int             `json:"total_configurations"`
	ActiveContributors  int             `json:"active_contributors"`
	PackagesTracked     int             `json:"packages_tracked"`
	LastUpdated         time.Time       `json:"last_updated"`
}

// PackageStats represents statistics for a package
type PackageStats struct {
	Name        string  `json:"name"`
	Downloads   int     `json:"downloads"`
	Rating      float64 `json:"rating"`
	Description string  `json:"description"`
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	IsValid       bool           `json:"is_valid"`
	Score         float64        `json:"score"`
	Issues        []string       `json:"issues"`
	Suggestions   []string       `json:"suggestions"`
	BestPractices []BestPractice `json:"best_practices"`
}

// Manager handles community integration functionality
type Manager struct {
	config       *config.UserConfig
	cache        *CacheManager
	githubClient *GitHubClient
	logger       *logger.Logger
}

// NewManager creates a new community manager instance
func NewManager(cfg *config.UserConfig) *Manager {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "nixai", "community")
	cache := NewCacheManager(cacheDir)

	githubClient := NewGitHubClient("")
	log := logger.NewLoggerWithLevel(cfg.LogLevel)

	return &Manager{
		config:       cfg,
		cache:        cache,
		githubClient: githubClient,
		logger:       log,
	}
}

// SearchConfigurations searches for configurations based on query
func (m *Manager) SearchConfigurations(query string) ([]Configuration, error) {
	m.logger.Info("Searching configurations for query: " + query)

	// Check cache first
	cacheKey := GetCacheKey("search", query)
	var cachedResults []Configuration
	if found, err := m.cache.Get(cacheKey, &cachedResults); err == nil && found {
		return cachedResults, nil
	}

	// Simulate search results (in real implementation, this would query external sources)
	configs := m.generateMockConfigurations(query)

	// Filter based on query
	var results []Configuration
	queryLower := strings.ToLower(query)
	for _, config := range configs {
		if m.matchesQuery(config, queryLower) {
			results = append(results, config)
		}
	}

	// Sort by relevance (rating for now)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	// Cache results
	m.cache.Set(cacheKey, results, "search")

	return results, nil
}

// ShareConfiguration shares a configuration with the community
func (m *Manager) ShareConfiguration(config *Configuration) error {
	m.logger.Info("Sharing configuration: " + config.Name)

	// Read file content
	if config.FilePath != "" {
		content, err := os.ReadFile(config.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read configuration file: %w", err)
		}
		config.Content = string(content)

		// Get file info
		info, err := os.Stat(config.FilePath)
		if err == nil {
			config.Size = info.Size()
		}
	}

	// Generate ID and set metadata
	config.ID = utils.GenerateID()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Language = "nix"

	// In real implementation, this would upload to a community platform
	// For now, we'll simulate success and cache the configuration
	sharedConfigs := []Configuration{*config}
	cacheKey := GetCacheKey("shared", config.Author)
	m.cache.Set(cacheKey, sharedConfigs, "shared")

	m.logger.Info("Configuration shared successfully with ID: " + config.ID)
	return nil
}

// ValidateConfiguration validates a configuration against best practices
func (m *Manager) ValidateConfiguration(filePath string) (*ValidationResult, error) {
	m.logger.Info("Validating configuration file: " + filePath)

	// Read configuration content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	contentStr := string(content)
	result := &ValidationResult{
		IsValid:       true,
		Score:         8.0,
		Issues:        []string{},
		Suggestions:   []string{},
		BestPractices: m.getBestPractices(),
	}

	// Run validation checks
	m.validateSecurity(contentStr, result)
	m.validatePerformance(contentStr, result)
	m.validateMaintainability(contentStr, result)

	// Calculate final score
	issueCount := len(result.Issues)
	if issueCount > 0 {
		result.Score = 10.0 - float64(issueCount)*0.5
		if result.Score < 0 {
			result.Score = 0
		}
	}

	result.IsValid = result.Score >= 7.0

	return result, nil
}

// GetTrends returns trending packages and configurations
func (m *Manager) GetTrends() (*TrendData, error) {
	m.logger.Info("Fetching community trends")

	// Check cache first
	cacheKey := GetCacheKey("trends", "current")
	var cachedTrends TrendData
	if found, err := m.cache.Get(cacheKey, &cachedTrends); err == nil && found {
		return &cachedTrends, nil
	}

	// Generate mock trend data
	trends := &TrendData{
		PopularPackages: []PackageStats{
			{Name: "home-manager", Downloads: 15420, Rating: 4.8, Description: "Declarative home configuration"},
			{Name: "nixpkgs", Downloads: 12890, Rating: 4.9, Description: "Nix packages collection"},
			{Name: "flake-utils", Downloads: 8765, Rating: 4.7, Description: "Pure Nix flake utility functions"},
			{Name: "nixos-hardware", Downloads: 7543, Rating: 4.6, Description: "Hardware-specific NixOS modules"},
			{Name: "nix-direnv", Downloads: 6234, Rating: 4.5, Description: "Fast direnv integration"},
		},
		TrendingConfigs:     m.generateTrendingConfigurations(),
		TotalConfigurations: 2847,
		ActiveContributors:  456,
		PackagesTracked:     89123,
		LastUpdated:         time.Now(),
	}

	// Cache trends
	m.cache.Set(cacheKey, trends, "trends")

	return trends, nil
}

// RateConfiguration submits a rating for a configuration
func (m *Manager) RateConfiguration(configName string, rating float64, comment string) error {
	m.logger.Info("Rating configuration: " + configName + " with rating " + fmt.Sprintf("%.1f", rating))

	// In real implementation, this would submit to a community platform
	// For now, we'll simulate success

	// Cache the rating
	ratingData := map[string]interface{}{
		"config":  configName,
		"rating":  rating,
		"comment": comment,
		"time":    time.Now(),
	}

	cacheKey := GetCacheKey("rating", configName, fmt.Sprintf("%.1f", rating))
	m.cache.Set(cacheKey, ratingData, "rating")

	return nil
}

// Helper methods

func (m *Manager) matchesQuery(config Configuration, query string) bool {
	// Check name, description, tags, and author
	if strings.Contains(strings.ToLower(config.Name), query) ||
		strings.Contains(strings.ToLower(config.Description), query) ||
		strings.Contains(strings.ToLower(config.Author), query) {
		return true
	}

	// Check tags
	for _, tag := range config.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}

func (m *Manager) generateMockConfigurations(query string) []Configuration {
	// This would be replaced with actual API calls in production
	configs := []Configuration{
		{
			ID: "config1", Name: "Gaming NixOS Setup", Author: "gamer123",
			Description: "Optimized NixOS configuration for gaming with NVIDIA drivers",
			Tags:        []string{"gaming", "nvidia", "performance"}, Rating: 4.8, Views: 1250,
			URL: "https://github.com/gamer123/nixos-gaming",
		},
		{
			ID: "config2", Name: "Docker Development Environment", Author: "devops_pro",
			Description: "Complete Docker setup with development tools",
			Tags:        []string{"docker", "development", "containers"}, Rating: 4.6, Views: 890,
			URL: "https://github.com/devops_pro/nixos-docker-dev",
		},
		{
			ID: "config3", Name: "Minimal Server Configuration", Author: "sysadmin",
			Description: "Lightweight server setup for production deployment",
			Tags:        []string{"server", "minimal", "production"}, Rating: 4.7, Views: 2100,
			URL: "https://github.com/sysadmin/nixos-minimal-server",
		},
		{
			ID: "config4", Name: "Home Lab Setup", Author: "homelab_enthusiast",
			Description: "Complete home lab configuration with services",
			Tags:        []string{"homelab", "services", "self-hosted"}, Rating: 4.5, Views: 750,
			URL: "https://github.com/homelab_enthusiast/nixos-homelab",
		},
	}
	return configs
}

func (m *Manager) generateTrendingConfigurations() []Configuration {
	return []Configuration{
		{
			ID: "trend1", Name: "Modern Development Workstation", Author: "dev_expert",
			Description: "Complete development environment with modern tools",
			Rating:      4.9, Views: 3200,
		},
		{
			ID: "trend2", Name: "Media Server Setup", Author: "media_guru",
			Description: "Plex, Sonarr, Radarr, and more for media management",
			Rating:      4.7, Views: 2850,
		},
		{
			ID: "trend3", Name: "Kubernetes Node Configuration", Author: "k8s_admin",
			Description: "Production-ready Kubernetes worker node setup",
			Rating:      4.8, Views: 2100,
		},
	}
}

func (m *Manager) getBestPractices() []BestPractice {
	return []BestPractice{
		{
			ID: "bp1", Title: "Use Declarative Package Management",
			Description: "Prefer environment.systemPackages over imperative package installation",
			Category:    "packages", Priority: "high",
		},
		{
			ID: "bp2", Title: "Enable Automatic Garbage Collection",
			Description: "Configure nix.gc to automatically clean old generations",
			Category:    "maintenance", Priority: "medium",
		},
		{
			ID: "bp3", Title: "Pin Nixpkgs Version",
			Description: "Use a specific nixpkgs commit or release for reproducibility",
			Category:    "reproducibility", Priority: "high",
		},
	}
}

func (m *Manager) validateSecurity(content string, result *ValidationResult) {
	if strings.Contains(content, "permitRootLogin = \"yes\"") {
		result.Issues = append(result.Issues, "SSH root login is enabled - security risk")
		result.Suggestions = append(result.Suggestions, "Set services.openssh.permitRootLogin = \"no\"")
	}

	if !strings.Contains(content, "firewall.enable = true") {
		result.Issues = append(result.Issues, "Firewall is not explicitly enabled")
		result.Suggestions = append(result.Suggestions, "Add networking.firewall.enable = true")
	}
}

func (m *Manager) validatePerformance(content string, result *ValidationResult) {
	if !strings.Contains(content, "nix.gc") {
		result.Issues = append(result.Issues, "Automatic garbage collection not configured")
		result.Suggestions = append(result.Suggestions, "Add nix.gc.automatic = true and nix.gc.dates = \"weekly\"")
	}
}

func (m *Manager) validateMaintainability(content string, result *ValidationResult) {
	if !strings.Contains(content, "system.stateVersion") {
		result.Issues = append(result.Issues, "State version not specified")
		result.Suggestions = append(result.Suggestions, "Add system.stateVersion = \"YY.MM\" for your NixOS version")
	}
}
