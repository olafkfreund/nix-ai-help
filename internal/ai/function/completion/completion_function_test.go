package completion

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewCompletionFunction(t *testing.T) {
	cf := NewCompletionFunction()

	if cf == nil {
		t.Fatal("NewCompletionFunction returned nil")
	}

	if cf.Name() != "completion" {
		t.Errorf("Expected function name 'completion', got '%s'", cf.Name())
	}

	if cf.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestCompletionFunction_ValidateParameters(t *testing.T) {
	cf := NewCompletionFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"context": "nixos-rebuild switch",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"context":         "nix-env -i",
				"completion_type": "packages",
				"prefix":          "firefox",
				"language":        "nix",
				"shell":           "zsh",
				"position":        8,
				"max_results":     10,
				"include_doc":     true,
				"filter_type":     "fuzzy",
				"options": map[string]interface{}{
					"case_sensitive": "false",
				},
			},
			expectError: false,
		},
		{
			name: "Missing required context",
			params: map[string]interface{}{
				"completion_type": "shell",
			},
			expectError: true,
		},
		{
			name: "Invalid completion_type",
			params: map[string]interface{}{
				"context":         "command",
				"completion_type": "invalid-type",
			},
			expectError: true,
		},
		{
			name: "Invalid max_results (negative)",
			params: map[string]interface{}{
				"context":     "command",
				"max_results": -1,
			},
			expectError: true,
		},
		{
			name: "Invalid position (negative)",
			params: map[string]interface{}{
				"context":  "command",
				"position": -5,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cf.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestCompletionFunction_ParseRequest(t *testing.T) {
	cf := NewCompletionFunction()

	params := map[string]interface{}{
		"context":         "nix-env -iA nixpkgs.",
		"completion_type": "packages",
		"prefix":          "firefox",
		"language":        "nix",
		"shell":           "zsh",
		"position":        20,
		"max_results":     15,
		"include_doc":     true,
		"filter_type":     "prefix",
		"options": map[string]interface{}{
			"case_sensitive": "false",
			"show_hidden":    "true",
		},
	}

	request, err := cf.parseRequest(params)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if request.Context != "nix-env -iA nixpkgs." {
		t.Errorf("Expected context 'nix-env -iA nixpkgs.', got '%s'", request.Context)
	}

	if request.CompletionType != "packages" {
		t.Errorf("Expected completion_type 'packages', got '%s'", request.CompletionType)
	}

	if request.Prefix != "firefox" {
		t.Errorf("Expected prefix 'firefox', got '%s'", request.Prefix)
	}

	if request.Language != "nix" {
		t.Errorf("Expected language 'nix', got '%s'", request.Language)
	}

	if request.Shell != "zsh" {
		t.Errorf("Expected shell 'zsh', got '%s'", request.Shell)
	}

	if request.Position != 20 {
		t.Errorf("Expected position 20, got %d", request.Position)
	}

	if request.MaxResults != 15 {
		t.Errorf("Expected max_results 15, got %d", request.MaxResults)
	}

	if !request.IncludeDoc {
		t.Error("Expected include_doc to be true")
	}

	if request.FilterType != "prefix" {
		t.Errorf("Expected filter_type 'prefix', got '%s'", request.FilterType)
	}

	if len(request.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(request.Options))
	}

	if request.Options["case_sensitive"] != "false" {
		t.Errorf("Expected case_sensitive 'false', got '%s'", request.Options["case_sensitive"])
	}
}

func TestCompletionFunction_CompletionTypes(t *testing.T) {
	cf := NewCompletionFunction()

	tests := []struct {
		name           string
		context        string
		completionType string
		expectSuccess  bool
	}{
		{
			name:           "NixOS completion type",
			context:        "nixos-rebuild switch",
			completionType: "nixos",
			expectSuccess:  true,
		},
		{
			name:           "Package completion type",
			context:        "nix-env -iA nixpkgs.",
			completionType: "packages",
			expectSuccess:  true,
		},
		{
			name:           "Shell completion type",
			context:        "ls -la",
			completionType: "shell",
			expectSuccess:  true,
		},
		{
			name:           "Nix expression completion",
			context:        "let x = ",
			completionType: "nix",
			expectSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"context":         tt.context,
				"completion_type": tt.completionType,
			}

			result, err := cf.Execute(context.Background(), params, &functionbase.FunctionOptions{})

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected successful execution, got error: %v", err)
				}
				if !result.Success {
					t.Errorf("Expected successful result, got failure: %s", result.Error)
				}
			}
		})
	}
}

func TestCompletionFunction_FilterLogic(t *testing.T) {
	cf := NewCompletionFunction()

	tests := []struct {
		name          string
		prefix        string
		filterType    string
		maxResults    int
		expectResults bool
	}{
		{
			name:          "Prefix filtering",
			prefix:        "fire",
			filterType:    "prefix",
			maxResults:    10,
			expectResults: true,
		},
		{
			name:          "Exact matching",
			prefix:        "firefox",
			filterType:    "exact",
			maxResults:    10,
			expectResults: true,
		},
		{
			name:          "No prefix (all results)",
			prefix:        "",
			filterType:    "prefix",
			maxResults:    10,
			expectResults: true,
		},
		{
			name:          "Limited results",
			prefix:        "",
			filterType:    "prefix",
			maxResults:    1,
			expectResults: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"context":     "nix-env -i",
				"prefix":      tt.prefix,
				"filter_type": tt.filterType,
				"max_results": tt.maxResults,
			}

			result, err := cf.Execute(context.Background(), params, &functionbase.FunctionOptions{})

			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if !result.Success {
				t.Errorf("Expected successful execution, got failure: %s", result.Error)
			}

			// Check that we got result data
			if result.Data == nil && tt.expectResults {
				t.Error("Expected result data but got nil")
			}
		})
	}
}

func TestCompletionFunction_DocumentationSupport(t *testing.T) {
	cf := NewCompletionFunction()

	tests := []struct {
		name           string
		completionType string
		includeDoc     bool
		expectSuccess  bool
	}{
		{
			name:           "NixOS completion with documentation",
			completionType: "nixos",
			includeDoc:     true,
			expectSuccess:  true,
		},
		{
			name:           "Packages completion with documentation",
			completionType: "packages",
			includeDoc:     true,
			expectSuccess:  true,
		},
		{
			name:           "Shell completion without documentation",
			completionType: "shell",
			includeDoc:     false,
			expectSuccess:  true,
		},
		{
			name:           "Flakes completion with documentation",
			completionType: "flakes",
			includeDoc:     true,
			expectSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"context":         "test context",
				"completion_type": tt.completionType,
				"include_doc":     tt.includeDoc,
			}

			result, err := cf.Execute(context.Background(), params, &functionbase.FunctionOptions{})

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected successful execution, got error: %v", err)
				}
				if !result.Success {
					t.Errorf("Expected successful result, got failure: %s", result.Error)
				}
			}
		})
	}
}

func TestCompletionFunction_ExecuteWithMockAgent(t *testing.T) {
	cf := NewCompletionFunction()

	params := map[string]interface{}{
		"context":         "nix-env -i",
		"completion_type": "packages",
		"prefix":          "vim",
		"max_results":     5,
	}

	options := &functionbase.FunctionOptions{}

	result, err := cf.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	if result.Data == nil {
		t.Error("Expected result data but got nil")
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
