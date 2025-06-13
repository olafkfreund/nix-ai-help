package community

import (
	"context"
	"strings"
	"testing"
)

func TestNewCommunityFunction(t *testing.T) {
	cf := NewCommunityFunction()

	if cf == nil {
		t.Fatal("NewCommunityFunction returned nil")
	}

	if cf.Name() != "community" {
		t.Errorf("Expected function name 'community', got '%s'", cf.Name())
	}

	if cf.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestCommunityFunction_ValidateParameters(t *testing.T) {
	cf := NewCommunityFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"query": "getting started",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"query":         "flakes tutorial",
				"resource_type": "tutorials",
				"tags":          []interface{}{"flakes", "nixos"},
				"difficulty":    "beginner",
				"sort_by":       "votes",
				"limit":         5,
			},
			expectError: false,
		},
		{
			name: "Missing query parameter",
			params: map[string]interface{}{
				"resource_type": "forum",
			},
			expectError: true,
		},
		{
			name: "Empty query parameter",
			params: map[string]interface{}{
				"query": "",
			},
			expectError: true,
		},
		{
			name: "Invalid resource_type",
			params: map[string]interface{}{
				"query":         "test",
				"resource_type": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid difficulty",
			params: map[string]interface{}{
				"query":      "test",
				"difficulty": "expert",
			},
			expectError: true,
		},
		{
			name: "Invalid sort_by",
			params: map[string]interface{}{
				"query":   "test",
				"sort_by": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid limit too low",
			params: map[string]interface{}{
				"query": "test",
				"limit": 0,
			},
			expectError: true,
		},
		{
			name: "Invalid limit too high",
			params: map[string]interface{}{
				"query": "test",
				"limit": 100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cf.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestCommunityFunction_ParseRequest(t *testing.T) {
	cf := NewCommunityFunction()

	params := map[string]interface{}{
		"query":         "nixos installation",
		"resource_type": "tutorials",
		"tags":          []interface{}{"nixos", "installation"},
		"difficulty":    "beginner",
		"sort_by":       "votes",
		"limit":         5.0, // JSON numbers are float64
	}

	request, err := cf.parseRequest(params)
	if err != nil {
		t.Fatalf("Unexpected error parsing request: %v", err)
	}

	if request.Query != "nixos installation" {
		t.Errorf("Expected query 'nixos installation', got '%s'", request.Query)
	}

	if request.ResourceType != "tutorials" {
		t.Errorf("Expected resource_type 'tutorials', got '%s'", request.ResourceType)
	}

	expectedTags := []string{"nixos", "installation"}
	if len(request.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(request.Tags))
	}

	for i, tag := range expectedTags {
		if i >= len(request.Tags) || request.Tags[i] != tag {
			t.Errorf("Expected tag[%d] '%s', got '%s'", i, tag, request.Tags[i])
		}
	}

	if request.Difficulty != "beginner" {
		t.Errorf("Expected difficulty 'beginner', got '%s'", request.Difficulty)
	}

	if request.SortBy != "votes" {
		t.Errorf("Expected sort_by 'votes', got '%s'", request.SortBy)
	}

	if request.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", request.Limit)
	}
}

func TestCommunityFunction_SearchCommunityResources(t *testing.T) {
	cf := NewCommunityFunction()

	tests := []struct {
		name          string
		request       *CommunityRequest
		minResults    int
		expectedTypes []string
	}{
		{
			name: "Installation query",
			request: &CommunityRequest{
				Query: "installation guide",
				Limit: 10,
			},
			minResults:    1,
			expectedTypes: []string{"forum"},
		},
		{
			name: "Flakes query",
			request: &CommunityRequest{
				Query: "flakes tutorial",
				Limit: 10,
			},
			minResults:    1,
			expectedTypes: []string{"forum"},
		},
		{
			name: "Options documentation",
			request: &CommunityRequest{
				Query:        "nixos options",
				ResourceType: "docs",
				Limit:        10,
			},
			minResults:    1,
			expectedTypes: []string{"docs"},
		},
		{
			name: "Package search",
			request: &CommunityRequest{
				Query:        "package search",
				ResourceType: "packages",
				Limit:        10,
			},
			minResults:    1,
			expectedTypes: []string{"packages"},
		},
		{
			name: "Getting started tutorial",
			request: &CommunityRequest{
				Query:        "getting started",
				ResourceType: "tutorials",
				Limit:        10,
			},
			minResults:    1,
			expectedTypes: []string{"tutorials"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := cf.searchCommunityResources(tt.request)

			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(results))
			}

			// Check that results match expected types
			if len(tt.expectedTypes) > 0 {
				found := false
				for _, result := range results {
					for _, expectedType := range tt.expectedTypes {
						if result.Type == expectedType {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Errorf("Expected at least one result of types %v, but found none", tt.expectedTypes)
				}
			}

			// Verify result structure
			for _, result := range results {
				if result.Type == "" {
					t.Error("Result should have a type")
				}
				if result.Title == "" {
					t.Error("Result should have a title")
				}
				if result.URL == "" {
					t.Error("Result should have a URL")
				}
				if result.Description == "" {
					t.Error("Result should have a description")
				}
			}
		})
	}
}

func TestCommunityFunction_FilterResults(t *testing.T) {
	cf := NewCommunityFunction()

	// Create test results
	results := []CommunityResource{
		{
			Type:       "forum",
			Title:      "Beginner Guide",
			Difficulty: "beginner",
			Tags:       []string{"tutorial", "nixos"},
		},
		{
			Type:       "docs",
			Title:      "Advanced Configuration",
			Difficulty: "advanced",
			Tags:       []string{"configuration", "expert"},
		},
		{
			Type:       "tutorial",
			Title:      "Intermediate Setup",
			Difficulty: "intermediate",
			Tags:       []string{"setup", "nixos"},
		},
	}

	tests := []struct {
		name           string
		request        *CommunityRequest
		expectedCount  int
		expectedTitles []string
	}{
		{
			name: "Filter by difficulty",
			request: &CommunityRequest{
				Difficulty: "beginner",
			},
			expectedCount:  1,
			expectedTitles: []string{"Beginner Guide"},
		},
		{
			name: "Filter by tags",
			request: &CommunityRequest{
				Tags: []string{"nixos"},
			},
			expectedCount:  2,
			expectedTitles: []string{"Beginner Guide", "Intermediate Setup"},
		},
		{
			name: "Filter by difficulty and tags",
			request: &CommunityRequest{
				Difficulty: "intermediate",
				Tags:       []string{"setup"},
			},
			expectedCount:  1,
			expectedTitles: []string{"Intermediate Setup"},
		},
		{
			name:          "No filters",
			request:       &CommunityRequest{},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := cf.filterResults(results, tt.request)

			if len(filtered) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(filtered))
			}

			if len(tt.expectedTitles) > 0 {
				for _, expectedTitle := range tt.expectedTitles {
					found := false
					for _, result := range filtered {
						if result.Title == expectedTitle {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected to find result with title '%s'", expectedTitle)
					}
				}
			}
		})
	}
}

func TestCommunityFunction_GenerateSuggestions(t *testing.T) {
	cf := NewCommunityFunction()

	tests := []struct {
		name             string
		query            string
		expectedContains []string
		minSuggestions   int
	}{
		{
			name:             "Installation query",
			query:            "how to install nixos",
			expectedContains: []string{"installation", "setup"},
			minSuggestions:   2,
		},
		{
			name:             "Flakes query",
			query:            "nix flakes tutorial",
			expectedContains: []string{"flakes", "tutorial"},
			minSuggestions:   2,
		},
		{
			name:             "Configuration query",
			query:            "nixos configuration",
			expectedContains: []string{"configuration", "options"},
			minSuggestions:   2,
		},
		{
			name:             "Package query",
			query:            "package management",
			expectedContains: []string{"package"},
			minSuggestions:   2,
		},
		{
			name:             "Generic query",
			query:            "help",
			expectedContains: []string{"getting started"},
			minSuggestions:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := cf.generateSuggestions(tt.query)

			if len(suggestions) < tt.minSuggestions {
				t.Errorf("Expected at least %d suggestions, got %d", tt.minSuggestions, len(suggestions))
			}

			for _, expected := range tt.expectedContains {
				found := false
				for _, suggestion := range suggestions {
					if strings.Contains(strings.ToLower(suggestion), strings.ToLower(expected)) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected suggestions to contain '%s', but found none in: %v", expected, suggestions)
				}
			}
		})
	}
}

func TestCommunityFunction_FindRelatedTags(t *testing.T) {
	cf := NewCommunityFunction()

	tests := []struct {
		name         string
		query        string
		existingTags []string
		minTags      int
		expectedTags []string
	}{
		{
			name:         "NixOS query",
			query:        "nixos installation",
			existingTags: []string{},
			minTags:      2,
			expectedTags: []string{"linux", "functional"},
		},
		{
			name:         "Flakes query",
			query:        "nix flakes",
			existingTags: []string{},
			minTags:      2,
			expectedTags: []string{"nix", "reproducible"},
		},
		{
			name:         "Home Manager query",
			query:        "home manager config",
			existingTags: []string{},
			minTags:      2,
			expectedTags: []string{"user-config", "dotfiles"},
		},
		{
			name:         "With existing tags",
			query:        "nixos",
			existingTags: []string{"linux"},
			minTags:      1,
			expectedTags: []string{"functional", "declarative"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relatedTags := cf.findRelatedTags(tt.query, tt.existingTags)

			if len(relatedTags) < tt.minTags {
				t.Errorf("Expected at least %d related tags, got %d", tt.minTags, len(relatedTags))
			}

			// Check that existing tags are not included
			for _, existingTag := range tt.existingTags {
				for _, relatedTag := range relatedTags {
					if strings.EqualFold(existingTag, relatedTag) {
						t.Errorf("Related tags should not contain existing tag '%s'", existingTag)
					}
				}
			}

			// Check for expected tags
			for _, expectedTag := range tt.expectedTags {
				found := false
				for _, relatedTag := range relatedTags {
					if strings.EqualFold(expectedTag, relatedTag) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find related tag '%s' in: %v", expectedTag, relatedTags)
				}
			}
		})
	}
}

func TestCommunityFunction_Execute(t *testing.T) {
	cf := NewCommunityFunction()
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Basic execution",
			params: map[string]interface{}{
				"query": "getting started with nixos",
			},
			expectError: false,
		},
		{
			name: "With resource type",
			params: map[string]interface{}{
				"query":         "flakes tutorial",
				"resource_type": "tutorials",
			},
			expectError: false,
		},
		{
			name: "With all parameters",
			params: map[string]interface{}{
				"query":         "nixos configuration",
				"resource_type": "docs",
				"tags":          []interface{}{"nixos", "configuration"},
				"difficulty":    "intermediate",
				"sort_by":       "votes",
				"limit":         3,
			},
			expectError: false,
		},
		{
			name: "Missing query",
			params: map[string]interface{}{
				"resource_type": "forum",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cf.Execute(ctx, tt.params, nil)

			if tt.expectError {
				if err != nil || (result != nil && result.Success) {
					t.Error("Expected execution error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected execution error: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if !result.Success {
				t.Error("Result should be successful")
			}

			if result.Data == nil {
				t.Error("Result data should not be nil")
			}

			// Verify the response structure
			response, ok := result.Data.(*CommunityResponse)
			if !ok {
				t.Errorf("Expected *CommunityResponse, got %T", result.Data)
				return
			}

			if response.Query == "" {
				t.Error("Response query should not be empty")
			}

			if len(response.Results) == 0 {
				t.Error("Response should include results")
			}

			// Check result structure
			for _, result := range response.Results {
				if result.Type == "" {
					t.Error("Result type should not be empty")
				}
				if result.Title == "" {
					t.Error("Result title should not be empty")
				}
				if result.URL == "" {
					t.Error("Result URL should not be empty")
				}
			}

			if len(response.Suggestions) == 0 {
				t.Error("Response should include suggestions")
			}

			if len(response.RelatedTags) == 0 {
				t.Error("Response should include related tags")
			}
		})
	}
}

func TestCommunityFunction_ExecuteWithMissingQuery(t *testing.T) {
	cf := NewCommunityFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"resource_type": "forum",
		// query is missing
	}

	result, err := cf.Execute(ctx, params, nil)

	// Should not return an error but result should indicate failure
	if err != nil {
		t.Errorf("Execute should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Success {
		t.Error("Result should not be successful when query is missing")
	}

	if result.Error == "" {
		t.Error("Result should have error message when query is missing")
	}
}

func TestCommunityFunction_ContainsHelper(t *testing.T) {
	testSlice := []string{"nixos", "flakes", "home-manager"}

	// Test case-insensitive matching
	if !contains(testSlice, "NixOS") {
		t.Error("contains should be case-insensitive")
	}

	if !contains(testSlice, "FLAKES") {
		t.Error("contains should be case-insensitive")
	}

	if contains(testSlice, "docker") {
		t.Error("contains should return false for non-existing items")
	}
}
