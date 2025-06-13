package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestAskAgent_NewAskAgent(t *testing.T) {
	mockProvider := &MockProvider{response: "test"}
	agent := NewAskAgent(mockProvider)

	if agent == nil {
		t.Fatal("NewAskAgent returned nil")
	}
	if agent.role != roles.RoleAsk {
		t.Errorf("Expected role %v, got %v", roles.RoleAsk, agent.role)
	}
}

func TestAskAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "This is how you enable flakes in NixOS..."}
	agent := NewAskAgent(mockProvider)

	question := "How do I enable flakes in NixOS?"
	response, err := agent.Query(context.Background(), question)

	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Verify response contains enhanced guidance
	if len(response) <= len(mockProvider.response) {
		t.Error("Expected response to be enhanced with guidance")
	}
}

func TestAskAgent_QueryWithContext(t *testing.T) {
	mockProvider := &MockProvider{response: "Flakes are a new feature..."}
	agent := NewAskAgent(mockProvider)

	question := "What are Nix flakes?"
	askCtx := &AskContext{
		Question:      question,
		Category:      "Nix Flakes",
		Urgency:       "Medium - Configuration",
		RelatedTopics: []string{"flake.nix", "flake inputs", "flake outputs"},
	}

	response, err := agent.QueryWithContext(context.Background(), question, askCtx)

	if err != nil {
		t.Errorf("QueryWithContext failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestAskAgent_CategorizeQuestion(t *testing.T) {
	agent := NewAskAgent(nil)

	tests := []struct {
		question string
		expected string
	}{
		{"How do I configure NixOS?", "NixOS System Configuration"},
		{"What is Home Manager?", "Home Manager Configuration"},
		{"How do flakes work?", "Nix Flakes"},
		{"How to install a package?", "Package Management"},
		{"Service not starting", "Service Management"},
		{"General Nix question", "General Nix/NixOS"},
	}

	for _, tt := range tests {
		result := agent.categorizeQuestion(tt.question)
		if result != tt.expected {
			t.Errorf("categorizeQuestion(%q) = %q, want %q", tt.question, result, tt.expected)
		}
	}
}

func TestAskAgent_DetermineUrgency(t *testing.T) {
	agent := NewAskAgent(nil)

	tests := []struct {
		question string
		expected string
	}{
		{"System is broken!", "High - System Issue"},
		{"How to setup flakes?", "Medium - Configuration"},
		{"What is Nix?", "Low - General Question"},
	}

	for _, tt := range tests {
		result := agent.determineUrgency(tt.question)
		if result != tt.expected {
			t.Errorf("determineUrgency(%q) = %q, want %q", tt.question, result, tt.expected)
		}
	}
}
