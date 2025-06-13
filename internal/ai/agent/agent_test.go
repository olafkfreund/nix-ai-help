package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockProvider implements the ai.Provider interface for testing
type MockProvider struct {
	response string
	err      error
}

func (m *MockProvider) Query(prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockProvider) GetPartialResponse() string {
	return ""
}

func (m *MockProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func TestOllamaAgent_Query_Diagnoser(t *testing.T) {
	mockProvider := &MockProvider{response: "diagnostics agent response"}
	agent := NewOllamaAgent(mockProvider)

	// Set role first
	err := agent.SetRole(roles.RoleDiagnoser)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	input := "systemctl status shows failed service"
	resp, err := agent.Query(context.Background(), input)

	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	if resp != "diagnostics agent response" {
		t.Errorf("Response = %s, expected 'diagnostics agent response'", resp)
	}
}

func TestOllamaAgent_Query_Explainer(t *testing.T) {
	mockProvider := &MockProvider{response: "explainer response"}
	agent := NewOllamaAgent(mockProvider)

	// Set role first
	err := agent.SetRole(roles.RoleExplainer)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	input := "nixos-rebuild switch"
	resp, err := agent.Query(context.Background(), input)

	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	if resp != "explainer response" {
		t.Errorf("Response = %s, expected 'explainer response'", resp)
	}
}

func TestOllamaAgent_Query_UnsupportedRole(t *testing.T) {
	mockProvider := &MockProvider{response: "response"}
	agent := NewOllamaAgent(mockProvider)

	// Try to set an invalid role
	err := agent.SetRole(roles.RoleType("unknown"))
	if err == nil {
		t.Error("Expected error for unsupported role, got none")
	}
}
