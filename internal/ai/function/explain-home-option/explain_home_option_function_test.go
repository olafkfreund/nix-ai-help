package explainhomeoption

import (
	"context"
	"strings"
	"testing"
)

func TestNewExplainHomeOptionFunction(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()

	if ehof == nil {
		t.Fatal("NewExplainHomeOptionFunction returned nil")
	}

	if ehof.Name() != "explain-home-option" {
		t.Errorf("Expected function name 'explain-home-option', got '%s'", ehof.Name())
	}

	if ehof.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestExplainHomeOptionFunction_ValidateParameters(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"option": "programs.git.enable",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"option":        "programs.bash.enable",
				"module":        "programs",
				"show_examples": true,
				"detailed":      true,
			},
			expectError: false,
		},
		{
			name: "Missing option parameter",
			params: map[string]interface{}{
				"module": "programs",
			},
			expectError: true,
		},
		{
			name: "Empty option parameter",
			params: map[string]interface{}{
				"option": "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ehof.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestExplainHomeOptionFunction_ParseRequest(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()

	params := map[string]interface{}{
		"option":        "programs.git.enable",
		"module":        "programs",
		"show_examples": true,
		"detailed":      false,
	}

	request, err := ehof.parseRequest(params)
	if err != nil {
		t.Fatalf("Unexpected error parsing request: %v", err)
	}

	if request.Option != "programs.git.enable" {
		t.Errorf("Expected option 'programs.git.enable', got '%s'", request.Option)
	}

	if request.Module != "programs" {
		t.Errorf("Expected module 'programs', got '%s'", request.Module)
	}

	if !request.ShowExamples {
		t.Error("Expected ShowExamples to be true")
	}

	if request.Detailed {
		t.Error("Expected Detailed to be false")
	}
}

func TestExplainHomeOptionFunction_GenerateBasicDescription(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()

	tests := []struct {
		name     string
		option   string
		contains []string
	}{
		{
			name:     "Git enable option",
			option:   "programs.git.enable",
			contains: []string{"git", "program", "Home Manager"},
		},
		{
			name:     "Bash enable option",
			option:   "programs.bash.enable",
			contains: []string{"bash", "program", "Home Manager"},
		},
		{
			name:     "Home packages option",
			option:   "home.packages",
			contains: []string{"packages", "install", "user"},
		},
		{
			name:     "Session variables option",
			option:   "home.sessionVariables",
			contains: []string{"environment", "variables", "shell"},
		},
		{
			name:     "XDG enable option",
			option:   "xdg.enable",
			contains: []string{"XDG", "Base Directory", "specification"},
		},
		{
			name:     "Service enable option",
			option:   "services.gpg-agent.enable",
			contains: []string{"gpg-agent", "service", "Home Manager"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := ehof.generateBasicDescription(tt.option)

			if description == "" {
				t.Error("Description should not be empty")
			}

			for _, word := range tt.contains {
				if !containsIgnoreCase(description, word) {
					t.Errorf("Description should contain '%s': %s", word, description)
				}
			}
		})
	}
}

func TestExplainHomeOptionFunction_GenerateExamples(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()

	tests := []struct {
		name     string
		option   string
		minCount int
		contains []string
	}{
		{
			name:     "Git enable option",
			option:   "programs.git.enable",
			minCount: 1,
			contains: []string{"programs.git.enable"},
		},
		{
			name:     "Bash enable option",
			option:   "programs.bash.enable",
			minCount: 1,
			contains: []string{"programs.bash.enable"},
		},
		{
			name:     "Home packages option",
			option:   "home.packages",
			minCount: 1,
			contains: []string{"home.packages", "with pkgs"},
		},
		{
			name:     "Session variables option",
			option:   "home.sessionVariables",
			minCount: 1,
			contains: []string{"home.sessionVariables"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			examples := ehof.generateExamples(tt.option)

			if len(examples) < tt.minCount {
				t.Errorf("Expected at least %d examples, got %d", tt.minCount, len(examples))
			}

			// Check if any example contains the required strings
			for _, required := range tt.contains {
				found := false
				for _, example := range examples {
					if containsIgnoreCase(example, required) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("No example contains '%s': %v", required, examples)
				}
			}
		})
	}
}

func TestExplainHomeOptionFunction_FindRelatedOptions(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()

	tests := []struct {
		name     string
		option   string
		minCount int
		contains []string
		excludes []string
	}{
		{
			name:     "Git enable option",
			option:   "programs.git.enable",
			minCount: 2,
			contains: []string{"programs.git.userName", "programs.git.userEmail"},
			excludes: []string{"programs.git.enable"}, // Should not include itself
		},
		{
			name:     "Bash enable option",
			option:   "programs.bash.enable",
			minCount: 2,
			contains: []string{"programs.bash.enableCompletion", "programs.bash.shellAliases"},
			excludes: []string{"programs.bash.enable"},
		},
		{
			name:     "Home packages option",
			option:   "home.packages",
			minCount: 2,
			contains: []string{"home.sessionVariables", "home.file"},
			excludes: []string{"home.packages"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			related := ehof.findRelatedOptions(tt.option)

			if len(related) < tt.minCount {
				t.Errorf("Expected at least %d related options, got %d", tt.minCount, len(related))
			}

			// Check if required options are included
			for _, required := range tt.contains {
				found := false
				for _, rel := range related {
					if rel == required {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Related options should contain '%s': %v", required, related)
				}
			}

			// Check that excluded options are not included
			for _, excluded := range tt.excludes {
				for _, rel := range related {
					if rel == excluded {
						t.Errorf("Related options should not contain '%s': %v", excluded, related)
					}
				}
			}
		})
	}
}

func TestExplainHomeOptionFunction_Execute(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Basic execution",
			params: map[string]interface{}{
				"option": "programs.git.enable",
			},
			expectError: false,
		},
		{
			name: "With examples",
			params: map[string]interface{}{
				"option":        "programs.bash.enable",
				"show_examples": true,
			},
			expectError: false,
		},
		{
			name: "Missing option",
			params: map[string]interface{}{
				"module": "programs",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ehof.Execute(ctx, tt.params, nil)

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
			response, ok := result.Data.(*ExplainHomeOptionResponse)
			if !ok {
				t.Errorf("Expected *ExplainHomeOptionResponse, got %T", result.Data)
				return
			}

			if response.Option == "" {
				t.Error("Response option should not be empty")
			}

			if response.Description == "" {
				t.Error("Response description should not be empty")
			}

			// Check examples if requested
			if showExamples, ok := tt.params["show_examples"].(bool); ok && showExamples {
				if len(response.Examples) == 0 {
					t.Error("Response should include examples when requested")
				}
			}
		})
	}
}

func TestExplainHomeOptionFunction_ExecuteWithMissingOption(t *testing.T) {
	ehof := NewExplainHomeOptionFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"module": "programs",
		// option is missing
	}

	result, err := ehof.Execute(ctx, params, nil)

	// Should not return an error but result should indicate failure
	if err != nil {
		t.Errorf("Execute should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Success {
		t.Error("Result should not be successful when option is missing")
	}

	if result.Error == "" {
		t.Error("Result should have error message when option is missing")
	}
}

// Helper function to check if a string contains another string (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
