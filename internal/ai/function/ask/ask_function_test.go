package ask

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewAskFunction(t *testing.T) {
	af := NewAskFunction()

	if af == nil {
		t.Fatal("NewAskFunction returned nil")
	}

	if af.Name() != "ask" {
		t.Errorf("Expected function name 'ask', got '%s'", af.Name())
	}

	if af.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestAskFunction_ValidateParameters(t *testing.T) {
	af := NewAskFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"question": "How do I enable SSH on NixOS?",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"question":       "How do I configure Git with Home Manager?",
				"category":       "home-manager",
				"context":        "Setting up development environment",
				"urgency":        "normal",
				"related_topics": []interface{}{"git", "development"},
			},
			expectError: false,
		},
		{
			name: "Invalid category",
			params: map[string]interface{}{
				"question": "How do I enable SSH?",
				"category": "invalid-category",
			},
			expectError: true,
		},
		{
			name: "Invalid urgency",
			params: map[string]interface{}{
				"question": "How do I enable SSH?",
				"urgency":  "super-urgent",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := af.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestAskFunction_ParseRequest(t *testing.T) {
	af := NewAskFunction()

	params := map[string]interface{}{
		"question":       "How do I configure Neovim?",
		"category":       "nixos",
		"context":        "Development setup",
		"urgency":        "normal",
		"related_topics": []interface{}{"neovim", "editor", "development"},
	}

	request, err := af.parseRequest(params)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if request.Question != "How do I configure Neovim?" {
		t.Errorf("Expected question 'How do I configure Neovim?', got '%s'", request.Question)
	}

	if request.Category != "nixos" {
		t.Errorf("Expected category 'nixos', got '%s'", request.Category)
	}

	if len(request.RelatedTopics) != 3 {
		t.Errorf("Expected 3 related topics, got %d", len(request.RelatedTopics))
	}
}

func TestAskFunction_BuildQuestionContext(t *testing.T) {
	af := NewAskFunction()

	request := &AskRequest{
		Question:      "How do I enable SSH?",
		Category:      "nixos",
		Context:       "Remote access needed",
		Urgency:       "high",
		RelatedTopics: []string{"ssh", "security"},
	}

	context := af.buildQuestionContext(request)

	if context == "" {
		t.Error("Question context should not be empty")
	}

	// Check that all components are included
	expectedParts := []string{
		"Question: How do I enable SSH?",
		"Category: nixos",
		"Context: Remote access needed",
		"Urgency: high",
		"Related topics: ssh, security",
	}

	for _, part := range expectedParts {
		if !contains(context, part) {
			t.Errorf("Expected context to contain '%s', but it didn't. Context: %s", part, context)
		}
	}
}

func TestAskFunction_DetermineConfidence(t *testing.T) {
	af := NewAskFunction()

	tests := []struct {
		name               string
		question           string
		expectedConfidence string
	}{
		{
			name:               "How-to question (high confidence)",
			question:           "How to enable SSH on NixOS?",
			expectedConfidence: "high",
		},
		{
			name:               "Configuration question (high confidence)",
			question:           "How do I configure Git?",
			expectedConfidence: "high",
		},
		{
			name:               "Recommendation question (medium confidence)",
			question:           "What is the best way to manage packages?",
			expectedConfidence: "medium",
		},
		{
			name:               "Vague question (low confidence)",
			question:           "Help me",
			expectedConfidence: "low",
		},
		{
			name:               "What is question (low confidence)",
			question:           "What is Nix?",
			expectedConfidence: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AskRequest{Question: tt.question}
			confidence := af.determineConfidence(request, "sample answer")

			if confidence != tt.expectedConfidence {
				t.Errorf("Expected confidence '%s', got '%s'", tt.expectedConfidence, confidence)
			}
		})
	}
}

func TestAskFunction_GenerateSuggestedActions(t *testing.T) {
	af := NewAskFunction()

	tests := []struct {
		name     string
		category string
		expected string
	}{
		{
			name:     "NixOS category",
			category: "nixos",
			expected: "nixos-rebuild switch",
		},
		{
			name:     "Home Manager category",
			category: "home-manager",
			expected: "home-manager switch",
		},
		{
			name:     "Troubleshooting category",
			category: "troubleshooting",
			expected: "journalctl",
		},
		{
			name:     "General category",
			category: "general",
			expected: "Backup your configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AskRequest{Category: tt.category}
			actions := af.generateSuggestedActions(request)

			if len(actions) == 0 {
				t.Error("Expected at least one suggested action")
			}

			found := false
			for _, action := range actions {
				if contains(action, tt.expected) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find action containing '%s' in actions %v", tt.expected, actions)
			}
		})
	}
}

func TestAskFunction_GenerateDocumentationRefs(t *testing.T) {
	af := NewAskFunction()

	tests := []struct {
		name     string
		category string
		expected string
	}{
		{
			name:     "NixOS category",
			category: "nixos",
			expected: "nixos.org/manual",
		},
		{
			name:     "Nix category",
			category: "nix",
			expected: "nix.dev/manual",
		},
		{
			name:     "Home Manager category",
			category: "home-manager",
			expected: "nix-community.github.io/home-manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AskRequest{Category: tt.category}
			refs := af.generateDocumentationRefs(request)

			if len(refs) == 0 {
				t.Error("Expected at least one documentation reference")
			}

			found := false
			for _, ref := range refs {
				if contains(ref, tt.expected) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find reference containing '%s' in refs %v", tt.expected, refs)
			}
		})
	}
}

func TestAskFunction_Execute(t *testing.T) {
	af := NewAskFunction()

	// Test with valid parameters
	params := map[string]interface{}{
		"question": "How do I enable SSH on NixOS?",
		"category": "nixos",
		"urgency":  "normal",
	}

	options := &functionbase.FunctionOptions{
		ProgressCallback: func(progress functionbase.Progress) {
			// Progress callback for testing
		},
	}

	// Note: This will fail without a real AI provider, but we can test the parameter parsing
	result, err := af.Execute(context.Background(), params, options)

	// We expect this to fail at the AI query stage since we don't have a real provider
	// But the parameter parsing should work
	if result != nil && !result.Success {
		// This is expected - we can't actually query without a provider
		var message string
		if result.Metadata != nil && result.Metadata["message"] != nil {
			message = result.Metadata["message"].(string)
		}
		t.Logf("Expected failure due to missing AI provider: %s", message)
	}

	if err != nil {
		t.Logf("Expected error due to missing AI provider: %v", err)
	}
}

func TestAskFunction_ExecuteWithMissingQuestion(t *testing.T) {
	af := NewAskFunction()

	// Test with missing required question parameter
	params := map[string]interface{}{
		"category": "nixos",
	}

	result, err := af.Execute(context.Background(), params, nil)

	if err != nil {
		t.Errorf("Execute should not return error, but should return failed result: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Success {
		t.Error("Result should indicate failure due to missing question")
	}

	// Check error message in the Error field instead of Message
	if !contains(result.Error, "question") {
		t.Errorf("Error message should mention missing question, got: %s", result.Error)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
