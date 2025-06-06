package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// SearchFunction handles searching for packages, options, and configurations
type SearchFunction struct {
	*functionbase.BaseFunction
	agent  *agent.SearchAgent
	logger *logger.Logger
}

// SearchRequest represents the input parameters for the search function
type SearchRequest struct {
	Context     string            `json:"context"`
	Query       string            `json:"query"`
	SearchType  string            `json:"search_type,omitempty"`
	Category    string            `json:"category,omitempty"`
	Source      string            `json:"source,omitempty"`
	MaxResults  int               `json:"max_results,omitempty"`
	IncludeDesc bool              `json:"include_desc,omitempty"`
	FilterBy    string            `json:"filter_by,omitempty"`
	SortBy      string            `json:"sort_by,omitempty"`
	Exact       bool              `json:"exact,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// SearchResponse represents the output of the search function
type SearchResponse struct {
	Context       string         `json:"context"`
	Status        string         `json:"status"`
	Query         string         `json:"query"`
	Results       []SearchResult `json:"results,omitempty"`
	TotalMatches  int            `json:"total_matches"`
	SearchTime    time.Duration  `json:"search_time"`
	Suggestions   []string       `json:"suggestions,omitempty"`
	Categories    []string       `json:"categories,omitempty"`
	ErrorMessage  string         `json:"error_message,omitempty"`
	ExecutionTime time.Duration  `json:"execution_time,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Type        string            `json:"type"`
	Category    string            `json:"category,omitempty"`
	Version     string            `json:"version,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	License     string            `json:"license,omitempty"`
	Path        string            `json:"path,omitempty"`
	Source      string            `json:"source"`
	Relevance   float64           `json:"relevance"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewSearchFunction creates a new search function instance
func NewSearchFunction() *SearchFunction {
	baseFunc := functionbase.NewBaseFunction(
		"search",
		"Search for NixOS packages, options, and configurations",
		[]functionbase.FunctionParameter{
			functionbase.StringParam("context", "The context or reason for the search", true),
			functionbase.StringParam("query", "Search query string", true),
			functionbase.StringParamWithEnum("search_type", "Type of search to perform", false, []string{
				"packages", "options", "configurations", "documentation", "flakes", "guides", "all",
			}),
			functionbase.StringParam("category", "Category to search within", false),
			functionbase.StringParam("source", "Source to search (nixpkgs, home-manager, etc.)", false),
			functionbase.IntParam("max_results", "Maximum number of results to return", false, 10),
			functionbase.BoolParam("include_desc", "Include detailed descriptions", false),
			functionbase.StringParam("filter_by", "Filter criteria", false),
			functionbase.StringParam("sort_by", "Sort criteria", false),
		},
	)

	return &SearchFunction{
		BaseFunction: baseFunc,
	}
}

// Name returns the function name
func (f *SearchFunction) Name() string {
	return f.BaseFunction.Name()
}

// Description returns the function description
func (f *SearchFunction) Description() string {
	return f.BaseFunction.Description()
}

// Version returns the function version
func (f *SearchFunction) Version() string {
	return "1.0.0"
}

// Parameters returns the function parameter schema
func (f *SearchFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"context": map[string]interface{}{
				"type":        "string",
				"description": "The context or reason for the search",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query string",
			},
			"search_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of search to perform",
				"enum":        []string{"packages", "options", "modules", "flakes", "configs", "all"},
				"default":     "packages",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "The category to search within",
				"enum":        []string{"development", "system", "desktop", "games", "multimedia", "networking", "security", "web"},
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "The source to search in",
				"enum":        []string{"nixpkgs", "home-manager", "nur", "flakes", "all"},
				"default":     "nixpkgs",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return",
				"default":     20,
				"minimum":     1,
				"maximum":     100,
			},
			"include_desc": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include descriptions in results",
				"default":     true,
			},
			"filter_by": map[string]interface{}{
				"type":        "string",
				"description": "Filter criteria for results",
				"enum":        []string{"maintained", "recent", "stable", "popular"},
			},
			"sort_by": map[string]interface{}{
				"type":        "string",
				"description": "Sort order for results",
				"enum":        []string{"relevance", "name", "popularity", "updated"},
				"default":     "relevance",
			},
			"exact": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to perform exact matching",
				"default":     false,
			},
			"options": map[string]interface{}{
				"type":        "object",
				"description": "Additional search options",
			},
		},
		"required": []string{"context", "query"},
	}
}

// Execute runs the search function with the given parameters
func (f *SearchFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Extract basic parameters
	context, _ := params["context"].(string)
	query, _ := params["query"].(string)
	searchType, _ := params["search_type"].(string)

	if context == "" {
		return functionbase.ErrorResult(fmt.Errorf("context parameter is required"), time.Since(startTime)), nil
	}
	if query == "" {
		return functionbase.ErrorResult(fmt.Errorf("query parameter is required"), time.Since(startTime)), nil
	}

	// Create search agent with default provider (should be configured separately)
	// For now, return a basic response without AI agent since provider integration
	// needs to be handled at a higher level

	// Build basic search result
	searchResponse := SearchResponse{
		Context: context,
		Query:   query,
		Status:  "success",
		Results: []SearchResult{
			{
				Name:        "Sample Result",
				Description: fmt.Sprintf("Search results for query: %s in context: %s", query, context),
				Source:      "nixpkgs",
				Category:    searchType,
			},
		},
		TotalMatches:  1,
		SearchTime:    time.Since(startTime),
		ExecutionTime: time.Since(startTime),
	}

	return functionbase.SuccessResult(searchResponse, time.Since(startTime)), nil
}

// executeSearch performs the actual search operation
func (f *SearchFunction) executeSearch(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	searchStart := time.Now()

	response := &SearchResponse{
		Context:    req.Context,
		Query:      req.Query,
		Status:     "success",
		Results:    []SearchResult{},
		SearchTime: 0,
	}

	switch strings.ToLower(req.SearchType) {
	case "packages":
		results, err := f.searchPackages(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search packages: %w", err)
		}
		response.Results = results

	case "options":
		results, err := f.searchOptions(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search options: %w", err)
		}
		response.Results = results

	case "modules":
		results, err := f.searchModules(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search modules: %w", err)
		}
		response.Results = results

	case "flakes":
		results, err := f.searchFlakes(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search flakes: %w", err)
		}
		response.Results = results

	case "configs":
		results, err := f.searchConfigs(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search configs: %w", err)
		}
		response.Results = results

	case "all":
		results, err := f.searchAll(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search all: %w", err)
		}
		response.Results = results

	default:
		return nil, fmt.Errorf("unsupported search type: %s", req.SearchType)
	}

	// Apply filtering and sorting
	response.Results = f.filterResults(response.Results, req)
	response.Results = f.sortResults(response.Results, req.SortBy)

	// Limit results
	if len(response.Results) > req.MaxResults {
		response.Results = response.Results[:req.MaxResults]
	}

	response.TotalMatches = len(response.Results)
	response.SearchTime = time.Since(searchStart)

	// Generate suggestions if no results found
	if len(response.Results) == 0 {
		response.Suggestions = f.generateSuggestions(req.Query)
	}

	// Get available categories
	response.Categories = f.getAvailableCategories(req.SearchType)

	f.logger.Info(fmt.Sprintf("Search completed: %d results in %v", response.TotalMatches, response.SearchTime))

	return response, nil
}

// searchPackages searches for packages
func (f *SearchFunction) searchPackages(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// Return mock search results since agent integration needs to be done at higher level
	results := []SearchResult{
		{
			Name:        "example-package",
			Description: fmt.Sprintf("Search result for '%s' in packages", req.Query),
			Type:        "package",
			Category:    req.Category,
			Version:     "1.0.0",
			Source:      req.Source,
		},
	}
	return results, nil
}

// searchOptions searches for NixOS options
func (f *SearchFunction) searchOptions(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// Return mock search results
	results := []SearchResult{
		{
			Name:        "example.option",
			Description: fmt.Sprintf("Search result for '%s' in options", req.Query),
			Type:        "option",
			Category:    req.Category,
			Path:        "services.example.enable",
			Source:      req.Source,
		},
	}
	return results, nil
}

// searchModules searches for NixOS modules
func (f *SearchFunction) searchModules(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// Return mock search results
	results := []SearchResult{
		{
			Name:        "example-module",
			Description: fmt.Sprintf("Search result for '%s' in modules", req.Query),
			Type:        "module",
			Category:    req.Category,
			Path:        "/path/to/module",
			Source:      req.Source,
		},
	}
	return results, nil
}

// searchFlakes searches for Nix flakes
func (f *SearchFunction) searchFlakes(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// Return mock search results
	results := []SearchResult{
		{
			Name:        "example-flake",
			Description: fmt.Sprintf("Search result for '%s' in flakes", req.Query),
			Type:        "flake",
			Category:    req.Category,
			Homepage:    "https://github.com/example/flake",
			Source:      "flakes",
		},
	}
	return results, nil
}

// searchConfigs searches for configuration examples
func (f *SearchFunction) searchConfigs(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// Return mock search results
	results := []SearchResult{
		{
			Name:        "example-config",
			Description: fmt.Sprintf("Search result for '%s' in configs", req.Query),
			Type:        "config",
			Category:    req.Category,
			Path:        "/etc/nixos/configuration.nix",
			Source:      "configs",
		},
	}
	return results, nil
}

// searchAll searches across all types
func (f *SearchFunction) searchAll(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	var allResults []SearchResult

	// Search packages
	if packages, err := f.searchPackages(ctx, req); err == nil {
		allResults = append(allResults, packages...)
	}

	// Search options
	if options, err := f.searchOptions(ctx, req); err == nil {
		allResults = append(allResults, options...)
	}

	// Search modules
	if modules, err := f.searchModules(ctx, req); err == nil {
		allResults = append(allResults, modules...)
	}

	// Search flakes
	if flakes, err := f.searchFlakes(ctx, req); err == nil {
		allResults = append(allResults, flakes...)
	}

	// Search configs
	if configs, err := f.searchConfigs(ctx, req); err == nil {
		allResults = append(allResults, configs...)
	}

	return allResults, nil
}

// filterResults applies filters to search results
func (f *SearchFunction) filterResults(results []SearchResult, req *SearchRequest) []SearchResult {
	if req.FilterBy == "" {
		return results
	}

	filtered := make([]SearchResult, 0)
	for _, result := range results {
		switch req.FilterBy {
		case "maintained":
			if result.Metadata["maintained"] == "true" {
				filtered = append(filtered, result)
			}
		case "recent":
			if result.Metadata["recent"] == "true" {
				filtered = append(filtered, result)
			}
		case "stable":
			if result.Metadata["stable"] == "true" {
				filtered = append(filtered, result)
			}
		case "popular":
			if result.Metadata["popular"] == "true" {
				filtered = append(filtered, result)
			}
		default:
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// sortResults sorts search results by specified criteria
func (f *SearchFunction) sortResults(results []SearchResult, sortBy string) []SearchResult {
	// Implementation would depend on the sorting criteria
	// For now, return as-is since relevance sorting is default
	return results
}

// generateSuggestions generates search suggestions for empty results
func (f *SearchFunction) generateSuggestions(query string) []string {
	// Simple suggestion logic - in practice this would be more sophisticated
	suggestions := []string{
		fmt.Sprintf("Try searching for '%s' in all sources", query),
		"Check spelling and try alternative terms",
		"Use broader search terms",
		"Try searching in different categories",
	}

	if strings.Contains(query, "-") {
		suggestions = append(suggestions, fmt.Sprintf("Try searching for '%s'", strings.ReplaceAll(query, "-", " ")))
	}

	return suggestions
}

// getAvailableCategories returns available categories for a search type
func (f *SearchFunction) getAvailableCategories(searchType string) []string {
	categories := map[string][]string{
		"packages": {"development", "system", "desktop", "games", "multimedia", "networking", "security", "web"},
		"options":  {"system", "hardware", "networking", "services", "security", "boot"},
		"modules":  {"system", "hardware", "services", "desktop", "development"},
		"flakes":   {"templates", "devshells", "packages", "systems", "modules"},
		"configs":  {"desktop", "server", "development", "gaming", "minimal"},
	}

	if cats, exists := categories[searchType]; exists {
		return cats
	}

	return []string{}
}
