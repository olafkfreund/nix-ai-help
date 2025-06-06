package community

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// CommunityFunction provides access to NixOS community resources and discussions
type CommunityFunction struct {
	*functionbase.BaseFunction
	logger logger.Logger
}

// CommunityRequest represents the input parameters for community resource access
type CommunityRequest struct {
	Query        string   `json:"query"`
	ResourceType string   `json:"resource_type,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Difficulty   string   `json:"difficulty,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"`
	Limit        int      `json:"limit,omitempty"`
}

// CommunityResponse represents the output of the community function
type CommunityResponse struct {
	Query        string              `json:"query"`
	ResourceType string              `json:"resource_type"`
	Results      []CommunityResource `json:"results"`
	Suggestions  []string            `json:"suggestions,omitempty"`
	RelatedTags  []string            `json:"related_tags,omitempty"`
}

// CommunityResource represents a single community resource
type CommunityResource struct {
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	URL         string            `json:"url"`
	Description string            `json:"description"`
	Author      string            `json:"author,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Difficulty  string            `json:"difficulty,omitempty"`
	Votes       int               `json:"votes,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewCommunityFunction creates a new community function
func NewCommunityFunction() *CommunityFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("query", "Search query for community resources", true),
		functionbase.StringParam("resource_type", "Type of resource: 'forum', 'docs', 'packages', 'tutorials', 'issues' (default: all)", false),
		functionbase.ArrayParam("tags", "Filter by tags (e.g., ['nixos', 'flakes', 'home-manager'])", false),
		functionbase.StringParam("difficulty", "Filter by difficulty: 'beginner', 'intermediate', 'advanced'", false),
		functionbase.StringParam("sort_by", "Sort results by: 'relevance' (default), 'date', 'votes', 'popularity'", false),
		functionbase.IntParam("limit", "Maximum number of results to return (default: 10, max: 50)", false, 10),
	}

	baseFunc := functionbase.NewBaseFunction(
		"community",
		"Access NixOS community resources including forums, documentation, packages, and tutorials",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Search for beginner tutorials",
			Parameters: map[string]interface{}{
				"query":         "getting started",
				"resource_type": "tutorials",
				"difficulty":    "beginner",
				"limit":         5,
			},
			Expected: "List of beginner-friendly NixOS tutorials and guides",
		},
		{
			Description: "Find forum discussions about flakes",
			Parameters: map[string]interface{}{
				"query":         "flakes configuration",
				"resource_type": "forum",
				"tags":          []interface{}{"flakes", "nixos"},
				"sort_by":       "votes",
			},
			Expected: "Recent forum discussions about Nix flakes sorted by votes",
		},
	}
	baseFunc.SetSchema(schema)

	return &CommunityFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
	}
}

// ValidateParameters validates the function parameters with custom checks
func (cf *CommunityFunction) ValidateParameters(params map[string]interface{}) error {
	// First run base validation
	if err := cf.BaseFunction.ValidateParameters(params); err != nil {
		return err
	}

	// Custom validation for query parameter
	if query, ok := params["query"].(string); ok {
		if strings.TrimSpace(query) == "" {
			return fmt.Errorf("query parameter cannot be empty")
		}
	}

	// Validate resource_type if provided
	if resourceType, ok := params["resource_type"].(string); ok && resourceType != "" {
		validTypes := []string{"forum", "docs", "packages", "tutorials", "issues"}
		if !contains(validTypes, resourceType) {
			return fmt.Errorf("invalid resource_type '%s', must be one of: %s", resourceType, strings.Join(validTypes, ", "))
		}
	}

	// Validate difficulty if provided
	if difficulty, ok := params["difficulty"].(string); ok && difficulty != "" {
		validDifficulties := []string{"beginner", "intermediate", "advanced"}
		if !contains(validDifficulties, difficulty) {
			return fmt.Errorf("invalid difficulty '%s', must be one of: %s", difficulty, strings.Join(validDifficulties, ", "))
		}
	}

	// Validate sort_by if provided
	if sortBy, ok := params["sort_by"].(string); ok && sortBy != "" {
		validSorts := []string{"relevance", "date", "votes", "popularity"}
		if !contains(validSorts, sortBy) {
			return fmt.Errorf("invalid sort_by '%s', must be one of: %s", sortBy, strings.Join(validSorts, ", "))
		}
	}

	// Validate limit if provided
	if limit, ok := params["limit"].(float64); ok {
		if limit < 1 || limit > 50 {
			return fmt.Errorf("limit must be between 1 and 50, got %.0f", limit)
		}
	}

	return nil
}

// Execute runs the community function
func (cf *CommunityFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	cf.logger.Debug("Starting community function execution")

	// Parse parameters into structured request
	request, err := cf.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate that we have a query
	if request.Query == "" {
		return functionbase.CreateErrorResult(
			fmt.Errorf("query parameter is required and cannot be empty"),
			"Missing required parameter",
		), nil
	}

	// Build the response
	response := &CommunityResponse{
		Query:        request.Query,
		ResourceType: request.ResourceType,
		Results:      cf.searchCommunityResources(request),
		Suggestions:  cf.generateSuggestions(request.Query),
		RelatedTags:  cf.findRelatedTags(request.Query, request.Tags),
	}

	cf.logger.Debug("Community function execution completed successfully")

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts raw parameters to structured CommunityRequest
func (cf *CommunityFunction) parseRequest(params map[string]interface{}) (*CommunityRequest, error) {
	request := &CommunityRequest{
		Limit: 10, // default
	}

	// Extract query (required)
	if query, ok := params["query"].(string); ok {
		request.Query = strings.TrimSpace(query)
	}

	// Extract resource_type (optional)
	if resourceType, ok := params["resource_type"].(string); ok {
		request.ResourceType = strings.TrimSpace(resourceType)
	}

	// Extract tags (optional)
	if tags, ok := params["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				request.Tags = append(request.Tags, strings.TrimSpace(tagStr))
			}
		}
	}

	// Extract difficulty (optional)
	if difficulty, ok := params["difficulty"].(string); ok {
		request.Difficulty = strings.TrimSpace(difficulty)
	}

	// Extract sort_by (optional)
	if sortBy, ok := params["sort_by"].(string); ok {
		request.SortBy = strings.TrimSpace(sortBy)
	}

	// Extract limit (optional)
	if limit, ok := params["limit"].(float64); ok {
		request.Limit = int(limit)
	}

	return request, nil
}

// searchCommunityResources simulates searching for community resources
func (cf *CommunityFunction) searchCommunityResources(request *CommunityRequest) []CommunityResource {
	var results []CommunityResource

	// Simulate different types of community resources based on query
	query := strings.ToLower(request.Query)

	// Generate forum discussions
	if request.ResourceType == "" || request.ResourceType == "forum" {
		if strings.Contains(query, "install") || strings.Contains(query, "setup") {
			results = append(results, CommunityResource{
				Type:        "forum",
				Title:       "NixOS Installation Guide Discussion",
				URL:         "https://discourse.nixos.org/t/nixos-installation-guide/123",
				Description: "Community discussion about best practices for NixOS installation",
				Author:      "nixos-community",
				Tags:        []string{"installation", "guide", "nixos"},
				Difficulty:  "beginner",
				Votes:       45,
				Metadata: map[string]string{
					"created": "2024-01-15",
					"replies": "23",
				},
			})
		}

		if strings.Contains(query, "flake") {
			results = append(results, CommunityResource{
				Type:        "forum",
				Title:       "Understanding Nix Flakes",
				URL:         "https://discourse.nixos.org/t/understanding-nix-flakes/456",
				Description: "In-depth discussion about Nix flakes architecture and usage patterns",
				Author:      "flake-expert",
				Tags:        []string{"flakes", "nix", "advanced"},
				Difficulty:  "intermediate",
				Votes:       78,
				Metadata: map[string]string{
					"created": "2024-02-10",
					"replies": "56",
				},
			})
		}
	}

	// Generate documentation
	if request.ResourceType == "" || request.ResourceType == "docs" {
		if strings.Contains(query, "option") || strings.Contains(query, "config") {
			results = append(results, CommunityResource{
				Type:        "docs",
				Title:       "NixOS Configuration Options Reference",
				URL:         "https://search.nixos.org/options",
				Description: "Comprehensive reference for all NixOS configuration options",
				Author:      "nixos-team",
				Tags:        []string{"configuration", "options", "reference"},
				Difficulty:  "intermediate",
				Metadata: map[string]string{
					"updated": "2024-03-01",
					"type":    "official",
				},
			})
		}

		if strings.Contains(query, "home") && strings.Contains(query, "manager") {
			results = append(results, CommunityResource{
				Type:        "docs",
				Title:       "Home Manager Manual",
				URL:         "https://nix-community.github.io/home-manager/",
				Description: "Complete documentation for Home Manager user environment management",
				Author:      "nix-community",
				Tags:        []string{"home-manager", "user-config", "documentation"},
				Difficulty:  "beginner",
				Metadata: map[string]string{
					"updated": "2024-02-28",
					"type":    "community",
				},
			})
		}
	}

	// Generate package resources
	if request.ResourceType == "" || request.ResourceType == "packages" {
		if strings.Contains(query, "search") || strings.Contains(query, "package") {
			results = append(results, CommunityResource{
				Type:        "packages",
				Title:       "NixOS Package Search",
				URL:         "https://search.nixos.org/packages",
				Description: "Official package search for nixpkgs repository",
				Author:      "nixos-team",
				Tags:        []string{"packages", "search", "nixpkgs"},
				Difficulty:  "beginner",
				Metadata: map[string]string{
					"packages": "80000+",
					"updated":  "daily",
				},
			})
		}
	}

	// Generate tutorials
	if request.ResourceType == "" || request.ResourceType == "tutorials" {
		if strings.Contains(query, "getting") && strings.Contains(query, "started") {
			results = append(results, CommunityResource{
				Type:        "tutorials",
				Title:       "NixOS Beginner's Guide",
				URL:         "https://nixos.org/learn.html",
				Description: "Step-by-step tutorial for new NixOS users",
				Author:      "nixos-community",
				Tags:        []string{"tutorial", "beginner", "guide"},
				Difficulty:  "beginner",
				Votes:       120,
				Metadata: map[string]string{
					"duration": "2 hours",
					"updated":  "2024-01-20",
				},
			})
		}

		if strings.Contains(query, "docker") || strings.Contains(query, "container") {
			results = append(results, CommunityResource{
				Type:        "tutorials",
				Title:       "Using Docker with NixOS",
				URL:         "https://nixos.wiki/wiki/Docker",
				Description: "Tutorial on running and managing Docker containers in NixOS",
				Author:      "docker-nix",
				Tags:        []string{"docker", "containers", "nixos"},
				Difficulty:  "intermediate",
				Votes:       67,
				Metadata: map[string]string{
					"duration": "1 hour",
					"updated":  "2024-02-15",
				},
			})
		}
	}

	// Apply filtering and sorting
	results = cf.filterResults(results, request)
	results = cf.sortResults(results, request.SortBy)

	// Apply limit
	if len(results) > request.Limit {
		results = results[:request.Limit]
	}

	return results
}

// filterResults applies filters to the search results
func (cf *CommunityFunction) filterResults(results []CommunityResource, request *CommunityRequest) []CommunityResource {
	var filtered []CommunityResource

	for _, result := range results {
		// Filter by difficulty
		if request.Difficulty != "" && result.Difficulty != request.Difficulty {
			continue
		}

		// Filter by tags
		if len(request.Tags) > 0 {
			hasMatchingTag := false
			for _, requestTag := range request.Tags {
				for _, resultTag := range result.Tags {
					if strings.EqualFold(requestTag, resultTag) {
						hasMatchingTag = true
						break
					}
				}
				if hasMatchingTag {
					break
				}
			}
			if !hasMatchingTag {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// sortResults sorts the search results based on the specified criteria
func (cf *CommunityFunction) sortResults(results []CommunityResource, sortBy string) []CommunityResource {
	// For now, return as-is since we're generating mock data
	// In a real implementation, this would sort by the specified criteria
	return results
}

// generateSuggestions creates search suggestions based on the query
func (cf *CommunityFunction) generateSuggestions(query string) []string {
	var suggestions []string

	query = strings.ToLower(query)

	switch {
	case strings.Contains(query, "install"):
		suggestions = append(suggestions, []string{
			"installation guide",
			"setup tutorial",
			"first steps",
			"hardware configuration",
		}...)

	case strings.Contains(query, "flake"):
		suggestions = append(suggestions, []string{
			"flakes tutorial",
			"flake.nix examples",
			"nix flakes best practices",
			"migrating to flakes",
		}...)

	case strings.Contains(query, "config"):
		suggestions = append(suggestions, []string{
			"configuration.nix examples",
			"nixos options",
			"system configuration",
			"user configuration",
		}...)

	case strings.Contains(query, "package"):
		suggestions = append(suggestions, []string{
			"package search",
			"custom packages",
			"package development",
			"package overrides",
		}...)

	default:
		suggestions = append(suggestions, []string{
			"getting started",
			"installation guide",
			"configuration examples",
			"package management",
		}...)
	}

	return suggestions
}

// findRelatedTags finds tags related to the query and existing tags
func (cf *CommunityFunction) findRelatedTags(query string, existingTags []string) []string {
	var relatedTags []string

	query = strings.ToLower(query)

	// Common tag associations
	tagMap := map[string][]string{
		"nixos":         {"linux", "functional", "declarative", "reproducible"},
		"flakes":        {"nix", "nixos", "reproducible", "development"},
		"home-manager":  {"user-config", "dotfiles", "nixos", "declarative"},
		"docker":        {"containers", "virtualization", "development"},
		"installation":  {"setup", "guide", "beginner", "hardware"},
		"configuration": {"options", "nixos", "system", "declarative"},
		"package":       {"nixpkgs", "software", "installation"},
		"development":   {"programming", "devenv", "tools"},
	}

	// Find related tags based on query content
	for key, tags := range tagMap {
		if strings.Contains(query, key) {
			for _, tag := range tags {
				if !contains(existingTags, tag) && !contains(relatedTags, tag) {
					relatedTags = append(relatedTags, tag)
				}
			}
		}
	}

	// Add common NixOS-related tags if none found
	if len(relatedTags) == 0 {
		relatedTags = []string{"nixos", "nix", "functional", "declarative"}
	}

	// Limit to 6 related tags
	if len(relatedTags) > 6 {
		relatedTags = relatedTags[:6]
	}

	return relatedTags
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
