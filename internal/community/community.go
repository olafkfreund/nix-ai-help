package community

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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
	config          *config.UserConfig
	cache           *CacheManager
	githubClient    *GitHubClient
	discourseClient *DiscourseClient
	logger          *logger.Logger
}

// NewManager creates a new community manager instance
func NewManager(cfg *config.UserConfig) *Manager {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "nixai", "community")
	cache := NewCacheManager(cacheDir)

	githubClient := NewGitHubClient("")

	// Initialize Discourse client with environment variables or config values
	discourseAPIKey := os.Getenv("DISCOURSE_API_KEY")
	if discourseAPIKey == "" {
		discourseAPIKey = cfg.Discourse.APIKey
	}

	discourseUsername := os.Getenv("DISCOURSE_USERNAME")
	if discourseUsername == "" {
		discourseUsername = cfg.Discourse.Username
	}

	discourseClient := NewDiscourseClient(cfg.Discourse.BaseURL, discourseAPIKey, discourseUsername)

	log := logger.NewLoggerWithLevel(cfg.LogLevel)

	return &Manager{
		config:          cfg,
		cache:           cache,
		githubClient:    githubClient,
		discourseClient: discourseClient,
		logger:          log,
	}
}

// SearchConfigurations searches for configurations based on query
func (m *Manager) SearchConfigurations(query string, limit int) ([]Configuration, error) {
	m.logger.Info("Searching configurations for query: " + query)

	// Check cache first
	cacheKey := GetCacheKey("search", query)
	var cachedResults []Configuration
	if found, err := m.cache.Get(cacheKey, &cachedResults); err == nil && found {
		return cachedResults, nil
	}

	var results []Configuration

	// Search Discourse if enabled
	if m.config.Discourse.Enabled {
		ctx := context.Background()
		discourseResults, err := m.searchDiscourse(ctx, query)
		if err != nil {
			m.logger.Info("Discourse search failed: " + err.Error())
		} else {
			results = append(results, discourseResults...)
		}
	}

	// Also get mock configurations (keeping existing functionality)
	mockConfigs := m.generateMockConfigurations(query)

	// Filter mock configs based on query
	queryLower := strings.ToLower(query)
	for _, config := range mockConfigs {
		if m.matchesQuery(config, queryLower) {
			results = append(results, config)
		}
	}

	// Sort by relevance (rating for now)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	// Cache results
	_ = m.cache.Set(cacheKey, results, "search")

	return results, nil
}

// SearchByCategory searches configurations within a specific category
func (m *Manager) SearchByCategory(category string, query string, limit int) ([]Configuration, error) {
	m.logger.Info(fmt.Sprintf("Searching in category '%s' for: %s", category, query))

	// Check cache first
	cacheKey := GetCacheKey("category_search", category+":"+query)
	var cachedResults []Configuration
	if found, err := m.cache.Get(cacheKey, &cachedResults); err == nil && found {
		return cachedResults, nil
	}

	var results []Configuration

	// Search Discourse if available and category is discourse-related
	if m.discourseClient != nil && m.config.Discourse.Enabled {
		if discourseResults, err := m.searchDiscourseByCategory(category, query, limit); err == nil {
			results = append(results, discourseResults...)
		} else {
			m.logger.Warn("Failed to search Discourse by category: " + err.Error())
		}
	}

	// Add mock results for demonstration
	mockResults := m.generateMockCategoryResults(category, query, limit)
	results = append(results, mockResults...)

	// Sort by relevance
	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	// Cache results
	_ = m.cache.Set(cacheKey, results, "category_search")

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
	_ = m.cache.Set(cacheKey, sharedConfigs, "shared")

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

	// Enhance with Discourse trending topics if available
	if m.discourseClient != nil && m.config.Discourse.Enabled {
		if err := m.enhanceTrendsWithDiscourse(trends); err != nil {
			m.logger.Warn("Failed to fetch Discourse trends: " + err.Error())
		}
	}

	// Cache trends
	_ = m.cache.Set(cacheKey, trends, "trends")

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
	_ = m.cache.Set(cacheKey, ratingData, "rating")

	return nil
}

// GetDiscourseStatus returns the status of Discourse integration
func (m *Manager) GetDiscourseStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":       m.config.Discourse.Enabled,
		"base_url":      m.config.Discourse.BaseURL,
		"authenticated": m.config.Discourse.APIKey != "" && m.config.Discourse.Username != "",
		"available":     false,
		"last_error":    nil,
	}

	if m.discourseClient != nil && m.config.Discourse.Enabled {
		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := m.discourseClient.GetCategories(ctx)
		if err != nil {
			status["last_error"] = err.Error()
			m.logger.Warn("Discourse connection test failed: " + err.Error())
		} else {
			status["available"] = true
		}
	}

	return status
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

func (m *Manager) generateMockCategoryResults(category string, query string, limit int) []Configuration {
	// This would be replaced with real category-specific search logic
	mockConfigs := []Configuration{
		{
			ID:          utils.GenerateID(),
			Name:        fmt.Sprintf("%s Configuration for %s", func(s string) string { caser := cases.Title(language.English); return caser.String(s) }(category), query),
			Author:      "community-user",
			Description: fmt.Sprintf("Sample %s configuration matching %s", category, query),
			Tags:        []string{category, "mock", "example"},
			Rating:      4.2,
			Downloads:   150,
			Views:       500,
			CreatedAt:   time.Now().AddDate(0, -1, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -5),
			Language:    "nix",
		},
	}

	if len(mockConfigs) > limit {
		return mockConfigs[:limit]
	}
	return mockConfigs
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

// enhanceTrendsWithDiscourse enhances trend data with Discourse popular topics
func (m *Manager) enhanceTrendsWithDiscourse(trends *TrendData) error {
	ctx := context.Background()

	// Get top topics from Discourse (weekly period)
	topTopicsResp, err := m.discourseClient.GetTopTopics(ctx, "weekly", 10)
	if err != nil {
		return fmt.Errorf("failed to get top topics: %w", err)
	}

	// Get latest topics for additional trending content
	latestTopicsResp, err := m.discourseClient.GetLatestTopics(ctx, 5)
	if err != nil {
		m.logger.Warn("Failed to get latest topics: " + err.Error())
	}

	// Combine topics from both responses
	allTopics := make([]DiscourseTopic, 0)
	if topTopicsResp != nil {
		allTopics = append(allTopics, topTopicsResp.TopicList.Topics...)
	}
	if latestTopicsResp != nil {
		allTopics = append(allTopics, latestTopicsResp.TopicList.Topics...)
	}

	// Convert topics to configurations
	discourseConfigs := make([]Configuration, 0, len(allTopics))

	for _, topic := range allTopics {
		// Get category name - we'll need to fetch this separately or use ID
		categoryName := fmt.Sprintf("category-%d", topic.CategoryID)

		config := Configuration{
			ID:          fmt.Sprintf("discourse-%d", topic.ID),
			Name:        topic.Title,
			Author:      "discourse-user", // Will be enhanced later with proper user lookup
			Description: m.extractDescription(topic.Excerpt, 200),
			Tags:        []string{"discourse", "community", categoryName},
			Rating:      m.calculateTopicRating(topic),
			Views:       topic.Views,
			URL:         fmt.Sprintf("%s/t/%s/%d", m.config.Discourse.BaseURL, topic.Slug, topic.ID),
			CreatedAt:   topic.CreatedAt,
			UpdatedAt:   topic.LastPostedAt,
			Language:    "markdown",
		}

		// Add any existing tags from the topic
		if len(topic.Tags) > 0 {
			config.Tags = append(config.Tags, topic.Tags...)
		}

		// Add common NixOS related tags
		config.Tags = append(config.Tags, "nixos", "configuration")

		discourseConfigs = append(discourseConfigs, config)
	}

	// Add top Discourse configurations to trending configs
	// Sort by rating and add best ones
	sort.Slice(discourseConfigs, func(i, j int) bool {
		return discourseConfigs[i].Rating > discourseConfigs[j].Rating
	})

	// Add up to 5 top Discourse topics to trending configs
	maxDiscourseConfigs := 5
	if len(discourseConfigs) < maxDiscourseConfigs {
		maxDiscourseConfigs = len(discourseConfigs)
	}

	if len(discourseConfigs) > 0 {
		trends.TrendingConfigs = append(trends.TrendingConfigs, discourseConfigs[:maxDiscourseConfigs]...)
	}

	// Update statistics
	trends.TotalConfigurations += len(discourseConfigs)

	return nil
}

// searchDiscourse searches Discourse for relevant posts and topics
func (m *Manager) searchDiscourse(ctx context.Context, query string) ([]Configuration, error) {
	// Search Discourse posts
	searchResults, err := m.discourseClient.SearchPosts(ctx, query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search Discourse: %w", err)
	}

	var configs []Configuration
	userMap := make(map[int]DiscourseUser)

	// Build user map for easy lookup
	for _, user := range searchResults.Users {
		userMap[user.ID] = user
	}

	// Convert posts to configurations
	for _, post := range searchResults.Posts {
		if _, exists := userMap[post.UserID]; exists {
			config := Configuration{
				ID:          fmt.Sprintf("discourse_post_%d", post.ID),
				Name:        fmt.Sprintf("Post: %s", post.TopicSlug),
				Author:      post.Username,
				Description: m.extractDescription(post.Cooked, 200),
				Tags:        []string{"discourse", "community", "help"},
				Rating:      4.0, // Default rating for Discourse posts
				Downloads:   0,
				Views:       0,
				URL:         fmt.Sprintf("%s/t/%s/%d/%d", m.config.Discourse.BaseURL, post.TopicSlug, post.TopicID, post.PostNumber),
				FilePath:    "",
				Content:     post.Raw,
				CreatedAt:   post.CreatedAt,
				UpdatedAt:   post.UpdatedAt,
				Size:        int64(len(post.Raw)),
				Language:    "markdown",
			}
			configs = append(configs, config)
		}
	}

	// Convert topics to configurations
	for _, topic := range searchResults.Topics {
		config := Configuration{
			ID:          fmt.Sprintf("discourse_topic_%d", topic.ID),
			Name:        topic.Title,
			Author:      "", // Will be filled from first post
			Description: topic.Excerpt,
			Tags:        append([]string{"discourse", "community"}, topic.Tags...),
			Rating:      m.calculateTopicRating(topic),
			Downloads:   0,
			Views:       topic.Views,
			URL:         fmt.Sprintf("%s/t/%s/%d", m.config.Discourse.BaseURL, topic.Slug, topic.ID),
			FilePath:    "",
			Content:     "",
			CreatedAt:   topic.CreatedAt,
			UpdatedAt:   topic.LastPostedAt,
			Size:        int64(topic.WordCount),
			Language:    "markdown",
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// searchDiscourseByCategory searches Discourse within a specific category
func (m *Manager) searchDiscourseByCategory(category string, query string, limit int) ([]Configuration, error) {
	ctx := context.Background()

	// Map common category names to potential Discourse categories
	categoryMap := map[string][]string{
		"guides":      {"guides", "tutorials", "howto"},
		"help":        {"help", "support", "troubleshooting"},
		"development": {"development", "dev", "programming"},
		"packages":    {"packages", "nixpkgs"},
		"flakes":      {"flakes", "nix-flakes"},
		"hardware":    {"hardware", "nixos-hardware"},
	}

	// Get category search terms
	searchTerms, exists := categoryMap[strings.ToLower(category)]
	if !exists {
		searchTerms = []string{category}
	}

	var allResults []Configuration

	// Search for each category term
	for _, term := range searchTerms {
		searchQuery := fmt.Sprintf("%s category:%s", query, term)
		searchResp, err := m.discourseClient.SearchPosts(ctx, searchQuery, limit)
		if err != nil {
			continue // Skip failed searches
		}

		// Convert posts to configurations
		for _, post := range searchResp.Posts {
			config := Configuration{
				ID:          fmt.Sprintf("discourse-post-%d", post.ID),
				Name:        fmt.Sprintf("Post #%d in topic %s", post.PostNumber, post.TopicSlug),
				Author:      post.Username,
				Description: m.extractDescription(post.Cooked, 200),
				Tags:        []string{"discourse", category, "post"},
				Rating:      3.5, // Default rating for posts
				URL:         fmt.Sprintf("%s/t/%s/%d/%d", m.config.Discourse.BaseURL, post.TopicSlug, post.TopicID, post.PostNumber),
				CreatedAt:   post.CreatedAt,
				UpdatedAt:   post.UpdatedAt,
				Language:    "markdown",
			}
			allResults = append(allResults, config)
		}
	}

	return allResults, nil
}

// extractDescription extracts a description from HTML content with a maximum length
func (m *Manager) extractDescription(htmlContent string, maxLength int) string {
	// Simple HTML tag removal (in production, use a proper HTML parser)
	description := strings.ReplaceAll(htmlContent, "<p>", "")
	description = strings.ReplaceAll(description, "</p>", " ")
	description = strings.ReplaceAll(description, "<br>", " ")
	description = strings.ReplaceAll(description, "<div>", "")
	description = strings.ReplaceAll(description, "</div>", " ")

	// Remove any remaining HTML tags
	for strings.Contains(description, "<") && strings.Contains(description, ">") {
		start := strings.Index(description, "<")
		end := strings.Index(description[start:], ">")
		if end == -1 {
			break
		}
		description = description[:start] + description[start+end+1:]
	}

	// Trim whitespace and limit length
	description = strings.TrimSpace(description)
	if len(description) > maxLength {
		description = description[:maxLength] + "..."
	}

	return description
}

// calculateTopicRating calculates a rating for a Discourse topic based on engagement
func (m *Manager) calculateTopicRating(topic DiscourseTopic) float64 {
	// Simple rating calculation based on likes, views, and posts
	likesScore := float64(topic.LikeCount) * 0.1
	viewsScore := float64(topic.Views) * 0.001
	postsScore := float64(topic.PostsCount) * 0.05

	rating := 3.0 + likesScore + viewsScore + postsScore

	// Cap at 5.0
	if rating > 5.0 {
		rating = 5.0
	}

	return rating
}
