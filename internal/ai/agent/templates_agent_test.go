package agent

import (
	"context"
	"fmt"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockProvider for TemplatesAgent testing
type mockTemplatesProvider struct {
	queryResponse     string
	generateResponse  string
	shouldReturnError bool
}

func (m *mockTemplatesProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if m.shouldReturnError {
		return "", fmt.Errorf("provider unavailable")
	}
	return m.generateResponse, nil
}

func (m *mockTemplatesProvider) Query(prompt string) (string, error) {
	if m.shouldReturnError {
		return "", fmt.Errorf("provider unavailable")
	}
	return m.queryResponse, nil
}

func (m *mockTemplatesProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	if m.shouldReturnError {
		return "", fmt.Errorf("provider unavailable")
	}
	return m.queryResponse, nil
}

func (m *mockTemplatesProvider) GetPartialResponse() string {
	return ""
}

func (m *mockTemplatesProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: "mock stream response", Done: true}
	close(ch)
	return ch, nil
}

func TestNewTemplatesAgent(t *testing.T) {
	provider := &mockTemplatesProvider{}
	agent := NewTemplatesAgent(provider)

	if agent == nil {
		t.Fatal("Expected TemplatesAgent, got nil")
	}

	if agent.role != roles.RoleTemplates {
		t.Errorf("Expected role %s, got %s", roles.RoleTemplates, agent.role)
	}

	if agent.provider != provider {
		t.Error("Expected provider to be set")
	}
}

func TestTemplatesAgent_Query(t *testing.T) {
	tests := []struct {
		name              string
		question          string
		mockResponse      string
		shouldReturnError bool
		expectError       bool
	}{
		{
			name:         "successful template query",
			question:     "How do I create a flake template for a Rust project?",
			mockResponse: "Here's how to create a Rust flake template...",
			expectError:  false,
		},
		{
			name:         "query about NixOS templates",
			question:     "What's the best template for a web server configuration?",
			mockResponse: "For web server templates, consider these options...",
			expectError:  false,
		},
		{
			name:              "provider error",
			question:          "Create a template",
			shouldReturnError: true,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockTemplatesProvider{
				queryResponse:     tt.mockResponse,
				shouldReturnError: tt.shouldReturnError,
			}
			agent := NewTemplatesAgent(provider)

			response, err := agent.Query(context.Background(), tt.question)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if response == "" {
				t.Error("Expected non-empty response")
			}

			// Check if response contains template guidance
			if !contains(response, "Template Management Tips") {
				t.Error("Expected response to contain template guidance")
			}
		})
	}
}

func TestTemplatesAgent_GenerateResponse(t *testing.T) {
	provider := &mockTemplatesProvider{
		generateResponse: "Generated template response",
	}
	agent := NewTemplatesAgent(provider)

	response, err := agent.GenerateResponse(context.Background(), "Generate a template")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestTemplatesAgent_GenerateTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateCtx  *TemplateContext
		mockResponse string
		expectError  bool
	}{
		{
			name: "generate flake template",
			templateCtx: &TemplateContext{
				TemplateType: "flake",
				ProjectName:  "my-rust-project",
				Language:     "rust",
				Features:     []string{"development", "CI/CD"},
			},
			mockResponse: "# Rust Flake Template\n{...}",
			expectError:  false,
		},
		{
			name: "generate nixos template",
			templateCtx: &TemplateContext{
				TemplateType: "nixos",
				Purpose:      "web server",
				Services:     []string{"nginx", "postgresql"},
			},
			mockResponse: "# NixOS Web Server Configuration\n{...}",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockTemplatesProvider{
				queryResponse: tt.mockResponse,
			}
			agent := NewTemplatesAgent(provider)

			response, err := agent.GenerateTemplate(context.Background(), tt.templateCtx)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if response == "" {
				t.Error("Expected non-empty response")
			}

			// Check if response is formatted properly
			if !contains(response, "Generated") && !contains(response, "Template") {
				t.Error("Expected response to be formatted as template output")
			}
		})
	}
}

func TestTemplatesAgent_CustomizeTemplate(t *testing.T) {
	provider := &mockTemplatesProvider{
		queryResponse: "Customized template with new features",
	}
	agent := NewTemplatesAgent(provider)

	baseTemplate := "# Basic flake template\n{...}"
	templateCtx := &TemplateContext{
		TemplateType:  "flake",
		Customization: "Add PostgreSQL support",
		Services:      []string{"postgresql"},
	}

	response, err := agent.CustomizeTemplate(context.Background(), baseTemplate, templateCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestTemplatesAgent_ExplainTemplate(t *testing.T) {
	provider := &mockTemplatesProvider{
		queryResponse: "This template configures a basic Rust development environment...",
	}
	agent := NewTemplatesAgent(provider)

	template := "{ description = \"Rust development environment\"; }"
	templateCtx := &TemplateContext{
		TemplateType: "flake",
		Language:     "rust",
	}

	response, err := agent.ExplainTemplate(context.Background(), template, templateCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Check if response contains template guidance
	if !contains(response, "Template Management Tips") {
		t.Error("Expected response to contain template guidance")
	}
}

func TestTemplatesAgent_ValidateTemplate(t *testing.T) {
	provider := &mockTemplatesProvider{
		queryResponse: "Template validation: syntax is correct, no issues found",
	}
	agent := NewTemplatesAgent(provider)

	template := "{ description = \"Valid template\"; inputs = {}; outputs = {}; }"
	templateCtx := &TemplateContext{
		TemplateType: "flake",
	}

	response, err := agent.ValidateTemplate(context.Background(), template, templateCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestTemplatesAgent_SuggestImprovements(t *testing.T) {
	provider := &mockTemplatesProvider{
		queryResponse: "Consider adding better error handling and documentation...",
	}
	agent := NewTemplatesAgent(provider)

	template := "{ description = \"Basic template\"; }"
	templateCtx := &TemplateContext{
		TemplateType: "flake",
	}

	response, err := agent.SuggestImprovements(context.Background(), template, templateCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestTemplatesAgent_ContextManagement(t *testing.T) {
	provider := &mockTemplatesProvider{}
	agent := NewTemplatesAgent(provider)

	// Test setting context
	templateCtx := &TemplateContext{
		TemplateType: "flake",
		ProjectName:  "test-project",
		Language:     "python",
	}

	agent.SetContext(templateCtx)

	// Test getting context
	retrievedCtx := agent.getTemplateContextFromData()
	if retrievedCtx.TemplateType != "flake" {
		t.Error("Context not properly stored or retrieved")
	}
}

func TestTemplatesAgent_RoleValidation(t *testing.T) {
	provider := &mockTemplatesProvider{}
	agent := NewTemplatesAgent(provider)

	// Test with correct role
	err := agent.validateRole()
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	// Test with incorrect role
	agent.role = "invalid-role"
	err = agent.validateRole()
	if err == nil {
		t.Error("Expected validation error for invalid role")
	}
}

func TestTemplatesAgent_PromptBuilding(t *testing.T) {
	provider := &mockTemplatesProvider{}
	agent := NewTemplatesAgent(provider)

	templateCtx := &TemplateContext{
		TemplateType:  "nixos",
		ProjectName:   "web-server",
		Purpose:       "production web server",
		Features:      []string{"nginx", "ssl", "monitoring"},
		Architecture:  "x86_64-linux",
		Services:      []string{"nginx", "postgresql"},
		Customization: "High availability setup",
	}

	// Test template prompt building
	prompt := agent.buildTemplatePrompt("How do I configure nginx?", templateCtx)
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Test generation prompt building
	genPrompt := agent.buildGenerateTemplatePrompt(templateCtx)
	if genPrompt == "" {
		t.Error("Expected non-empty generation prompt")
	}

	// Test customization prompt building
	baseTemplate := "{ services.nginx.enable = true; }"
	customPrompt := agent.buildCustomizeTemplatePrompt(baseTemplate, templateCtx)
	if customPrompt == "" {
		t.Error("Expected non-empty customization prompt")
	}

	// Test explanation prompt building
	explainPrompt := agent.buildExplainTemplatePrompt(baseTemplate, templateCtx)
	if explainPrompt == "" {
		t.Error("Expected non-empty explanation prompt")
	}

	// Test validation prompt building
	validatePrompt := agent.buildValidateTemplatePrompt(baseTemplate, templateCtx)
	if validatePrompt == "" {
		t.Error("Expected non-empty validation prompt")
	}

	// Test improvement prompt building
	improvePrompt := agent.buildImprovementPrompt(baseTemplate, templateCtx)
	if improvePrompt == "" {
		t.Error("Expected non-empty improvement prompt")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
