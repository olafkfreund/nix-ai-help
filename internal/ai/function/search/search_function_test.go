package search

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSearchFunction(t *testing.T) {
	function := NewSearchFunction()

	assert.NotNil(t, function)
	assert.Equal(t, "search", function.Name())
	assert.Equal(t, "Search for NixOS packages, options, and configurations", function.Description())
	assert.Equal(t, "1.0.0", function.Version())

	// Test parameters
	params := function.Parameters()
	require.NotNil(t, params)

	properties, ok := params["properties"].(map[string]interface{})
	require.True(t, ok)

	// Check required parameters
	contextParam, exists := properties["context"]
	assert.True(t, exists)
	assert.NotNil(t, contextParam)

	queryParam, exists := properties["query"]
	assert.True(t, exists)
	assert.NotNil(t, queryParam)

	searchTypeParam, exists := properties["search_type"]
	assert.True(t, exists)
	assert.NotNil(t, searchTypeParam)
}

func TestSearchFunction_ValidateParameters(t *testing.T) {
	function := NewSearchFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid search for packages",
			params: map[string]interface{}{
				"context":     "Looking for a text editor",
				"query":       "vim",
				"search_type": "packages",
			},
			expectError: false,
		},
		{
			name: "valid search for options",
			params: map[string]interface{}{
				"context":     "Configuring services",
				"query":       "nginx",
				"search_type": "options",
				"source":      "nixpkgs",
			},
			expectError: false,
		},
		{
			name: "valid search with filters",
			params: map[string]interface{}{
				"context":      "Development tools",
				"query":        "compiler",
				"search_type":  "packages",
				"category":     "development",
				"max_results":  50,
				"include_desc": true,
				"sort_by":      "popularity",
			},
			expectError: false,
		},
		{
			name:        "missing context parameter",
			params:      map[string]interface{}{"query": "vim"},
			expectError: true,
		},
		{
			name:        "missing query parameter",
			params:      map[string]interface{}{"context": "Looking for editor"},
			expectError: true,
		},
		{
			name:        "empty parameters",
			params:      map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := function.ValidateParameters(tt.params)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchFunction_Execute_BasicSearch(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"context":     "Looking for a text editor",
		"query":       "vim",
		"search_type": "packages",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	// Verify response structure
	response, ok := result.Data.(SearchResponse)
	require.True(t, ok)

	assert.Equal(t, "Looking for a text editor", response.Context)
	assert.Equal(t, "vim", response.Query)
	assert.Equal(t, "success", response.Status)
	assert.Greater(t, len(response.Results), 0)
	assert.Greater(t, response.SearchTime, time.Duration(0))
}

func TestSearchFunction_Execute_SearchTypes(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	searchTypes := []string{"packages", "options", "modules", "flakes", "configs", "all"}

	for _, searchType := range searchTypes {
		t.Run("search_type_"+searchType, func(t *testing.T) {
			params := map[string]interface{}{
				"context":     "Testing search functionality",
				"query":       "test",
				"search_type": searchType,
			}
			result, err := function.Execute(ctx, params, nil)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Success)

			response, ok := result.Data.(SearchResponse)
			require.True(t, ok)
			assert.NotEmpty(t, response.Query)
		})
	}
}

func TestSearchFunction_Execute_WithFilters(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"context":      "Development environment setup",
		"query":        "compiler",
		"search_type":  "packages",
		"category":     "development",
		"source":       "nixpkgs",
		"max_results":  25,
		"include_desc": true,
		"filter_by":    "maintained",
		"sort_by":      "popularity",
		"exact":        false,
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	response, ok := result.Data.(SearchResponse)
	require.True(t, ok)

	assert.Equal(t, "Development environment setup", response.Context)
	assert.Equal(t, "compiler", response.Query)
	assert.GreaterOrEqual(t, len(response.Results), 0)
}

func TestSearchFunction_Execute_MissingContext(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"query":       "vim",
		"search_type": "packages",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "context parameter is required")
}

func TestSearchFunction_Execute_MissingQuery(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"context":     "Looking for editor",
		"search_type": "packages",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "query parameter is required")
}

func TestSearchFunction_searchPackages(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	req := &SearchRequest{
		Context:     "Testing package search",
		Query:       "vim",
		SearchType:  "packages",
		Category:    "development",
		Source:      "nixpkgs",
		MaxResults:  10,
		IncludeDesc: true,
	}

	results, err := function.searchPackages(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check result structure
	for _, result := range results {
		assert.NotEmpty(t, result.Name)
		assert.Equal(t, "package", result.Type)
		assert.NotEmpty(t, result.Source)
	}
}

func TestSearchFunction_searchOptions(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	req := &SearchRequest{
		Context:    "Testing options search",
		Query:      "nginx",
		SearchType: "options",
		Source:     "nixpkgs",
	}

	results, err := function.searchOptions(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check result structure
	for _, result := range results {
		assert.NotEmpty(t, result.Name)
		assert.Equal(t, "option", result.Type)
		assert.NotEmpty(t, result.Path)
	}
}

func TestSearchFunction_searchModules(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	req := &SearchRequest{
		Context:    "Testing modules search",
		Query:      "docker",
		SearchType: "modules",
		Source:     "nixpkgs",
	}

	results, err := function.searchModules(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check result structure
	for _, result := range results {
		assert.NotEmpty(t, result.Name)
		assert.Equal(t, "module", result.Type)
	}
}

func TestSearchFunction_searchFlakes(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	req := &SearchRequest{
		Context:    "Testing flakes search",
		Query:      "home-manager",
		SearchType: "flakes",
		Source:     "flakes",
	}

	results, err := function.searchFlakes(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check result structure
	for _, result := range results {
		assert.NotEmpty(t, result.Name)
		assert.Equal(t, "flake", result.Type)
	}
}

func TestSearchFunction_searchConfigs(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	req := &SearchRequest{
		Context:    "Testing configs search",
		Query:      "desktop",
		SearchType: "configs",
		Source:     "configs",
	}

	results, err := function.searchConfigs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check result structure
	for _, result := range results {
		assert.NotEmpty(t, result.Name)
		assert.Equal(t, "config", result.Type)
	}
}

func TestSearchFunction_searchAll(t *testing.T) {
	function := NewSearchFunction()
	ctx := context.Background()

	req := &SearchRequest{
		Context:    "Testing comprehensive search",
		Query:      "git",
		SearchType: "all",
		MaxResults: 50,
	}

	results, err := function.searchAll(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Should contain results from different types
	types := make(map[string]bool)
	for _, result := range results {
		types[result.Type] = true
	}
	assert.Greater(t, len(types), 1) // Should have multiple result types
}

// Test request/response structs
func TestSearchStructs(t *testing.T) {
	// Test SearchRequest struct
	req := SearchRequest{
		Context:     "Development setup",
		Query:       "vim",
		SearchType:  "packages",
		Category:    "development",
		Source:      "nixpkgs",
		MaxResults:  20,
		IncludeDesc: true,
		FilterBy:    "maintained",
		SortBy:      "popularity",
		Exact:       false,
		Options:     map[string]string{"test": "value"},
	}

	assert.Equal(t, "Development setup", req.Context)
	assert.Equal(t, "vim", req.Query)
	assert.Equal(t, "packages", req.SearchType)
	assert.Equal(t, 20, req.MaxResults)
	assert.True(t, req.IncludeDesc)

	// Test SearchResult struct
	result := SearchResult{
		Name:        "vim",
		Description: "Vi IMproved - enhanced vi editor",
		Type:        "package",
		Category:    "development",
		Version:     "9.0.0",
		Homepage:    "https://www.vim.org/",
		License:     "MIT",
		Path:        "vim",
		Source:      "nixpkgs",
		Relevance:   0.95,
		Metadata:    map[string]string{"maintainer": "nixos"},
	}

	assert.Equal(t, "vim", result.Name)
	assert.Equal(t, "package", result.Type)
	assert.Equal(t, 0.95, result.Relevance)
	assert.Contains(t, result.Metadata, "maintainer")

	// Test SearchResponse struct
	response := SearchResponse{
		Context:       "Development setup",
		Status:        "success",
		Query:         "vim",
		Results:       []SearchResult{result},
		TotalMatches:  1,
		SearchTime:    time.Millisecond * 100,
		Suggestions:   []string{"neovim", "emacs"},
		Categories:    []string{"development", "editor"},
		ExecutionTime: time.Millisecond * 150,
	}

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, 1, response.TotalMatches)
	assert.Len(t, response.Results, 1)
	assert.Contains(t, response.Suggestions, "neovim")
	assert.Greater(t, response.SearchTime, time.Duration(0))
}
