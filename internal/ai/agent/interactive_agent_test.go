package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockProviderForInteractive implements the Provider interface for testing
type MockProviderForInteractive struct {
	response  string
	err       error
	LastQuery string
}

func (m *MockProviderForInteractive) Query(prompt string) (string, error) {
	m.LastQuery = prompt
	return m.response, m.err
}

func (m *MockProviderForInteractive) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	m.LastQuery = prompt
	return m.response, m.err
}

func (m *MockProviderForInteractive) GetPartialResponse() string {
	return ""
}

func (m *MockProviderForInteractive) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: "mock stream response", Done: true}
	close(ch)
	return ch, nil
}

func (m *MockProviderForInteractive) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	m.LastQuery = prompt
	return m.response, m.err
}

func TestInteractiveAgent_NewInteractiveAgent(t *testing.T) {
	mockProvider := &MockProviderForInteractive{}
	agent := NewInteractiveAgent(mockProvider)

	if agent == nil {
		t.Fatal("NewInteractiveAgent returned nil")
	}

	if agent.role != roles.RoleInteractive {
		t.Errorf("Expected role %v, got %v", roles.RoleInteractive, agent.role)
	}
}

func TestInteractiveAgent_SetRole(t *testing.T) {
	mockProvider := &MockProviderForInteractive{}
	agent := NewInteractiveAgent(mockProvider)

	err := agent.SetRole(roles.RoleInteractive)

	if err != nil {
		t.Errorf("SetRole failed: %v", err)
	}
}

func TestInteractiveAgent_Query(t *testing.T) {
	mockProvider := &MockProviderForInteractive{
		response: "I can help you with your NixOS configuration. What specific issue are you facing?",
	}
	agent := NewInteractiveAgent(mockProvider)

	// Set role
	err := agent.SetRole(roles.RoleInteractive)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	input := "I'm having trouble with my NixOS configuration and need help debugging it"
	response, err := agent.Query(context.Background(), input)

	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Check that interactive guidance is added
	if !strings.Contains(response, "Interactive Tips") {
		t.Error("Expected response to contain interactive guidance")
	}

	// Verify prompt contains expected elements
	if !strings.Contains(mockProvider.LastQuery, input) {
		t.Error("Expected query to contain original input")
	}
	if !strings.Contains(mockProvider.LastQuery, "Interactive Session") {
		t.Error("Expected query to contain interactive session header")
	}
}

func TestInteractiveAgent_QueryWithContext(t *testing.T) {
	mockProvider := &MockProviderForInteractive{
		response: "Based on your beginner level, let me guide you through this step by step.",
	}
	agent := NewInteractiveAgent(mockProvider)

	// Set role
	err := agent.SetRole(roles.RoleInteractive)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	input := "How do I install a package in NixOS?"
	interactiveCtx := &InteractiveContext{
		SessionID:    "test-session-123",
		UserLevel:    "Beginner",
		CurrentTask:  "Package Installation",
		StepNumber:   1,
		ErrorContext: "",
		Preferences:  make(map[string]string),
		Metadata:     make(map[string]string),
	}

	response, err := agent.QueryWithContext(context.Background(), input, interactiveCtx)

	if err != nil {
		t.Errorf("QueryWithContext failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Verify prompt contains expected context elements
	if !strings.Contains(mockProvider.LastQuery, "test-session-123") {
		t.Error("Expected query to contain session ID")
	}
	if !strings.Contains(mockProvider.LastQuery, "Beginner") {
		t.Error("Expected query to contain user level")
	}
	if !strings.Contains(mockProvider.LastQuery, "Package Installation") {
		t.Error("Expected query to contain current task")
	}
	if !strings.Contains(mockProvider.LastQuery, input) {
		t.Error("Expected query to contain original input")
	}
}

func TestInteractiveAgent_StartSession(t *testing.T) {
	mockProvider := &MockProviderForInteractive{
		response: "Welcome! I'm here to help with your NixOS journey. What would you like to work on?",
	}
	agent := NewInteractiveAgent(mockProvider)

	// Set role
	err := agent.SetRole(roles.RoleInteractive)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	response, err := agent.StartSession(context.Background(), "Intermediate")

	if err != nil {
		t.Errorf("StartSession failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Verify session was initialized
	if !strings.Contains(mockProvider.LastQuery, "Intermediate") {
		t.Error("Expected query to contain user level")
	}
	if !strings.Contains(mockProvider.LastQuery, "starting session") {
		t.Error("Expected query to contain starting session task")
	}
}

func TestInteractiveAgent_DetermineUserLevel(t *testing.T) {
	agent := NewInteractiveAgent(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "advanced user with derivation",
			input:    "I need help with a custom derivation for my package",
			expected: "Advanced",
		},
		{
			name:     "advanced user with overlay",
			input:    "How do I create an overlay for this nix expression?",
			expected: "Advanced",
		},
		{
			name:     "intermediate user with configuration.nix",
			input:    "I want to modify my configuration.nix file",
			expected: "Intermediate",
		},
		{
			name:     "intermediate user with home-manager",
			input:    "How do I set up home-manager?",
			expected: "Intermediate",
		},
		{
			name:     "beginner user",
			input:    "I'm new to NixOS and need help getting started",
			expected: "Beginner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.determineUserLevel(tt.input)
			if result != tt.expected {
				t.Errorf("determineUserLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInteractiveAgent_ExtractCurrentTask(t *testing.T) {
	agent := NewInteractiveAgent(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "package installation task",
			input:    "I want to install firefox on my system",
			expected: "Package Installation",
		},
		{
			name:     "configuration task",
			input:    "How do I configure my desktop environment?",
			expected: "Configuration Management",
		},
		{
			name:     "build task",
			input:    "My system rebuild is failing",
			expected: "System Building",
		},
		{
			name:     "troubleshooting task",
			input:    "I'm getting an error when I try to boot",
			expected: "Troubleshooting",
		},
		{
			name:     "setup task",
			input:    "I need help with initial setup",
			expected: "System Setup",
		},
		{
			name:     "general assistance",
			input:    "What's the best way to learn NixOS?",
			expected: "General Assistance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.extractCurrentTask(tt.input)
			if result != tt.expected {
				t.Errorf("extractCurrentTask() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInteractiveAgent_SessionHistory(t *testing.T) {
	mockProvider := &MockProviderForInteractive{
		response: "Test response",
	}
	agent := NewInteractiveAgent(mockProvider)

	// Set role
	err := agent.SetRole(roles.RoleInteractive)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	// Initially empty
	history := agent.GetSessionHistory()
	if len(history) != 0 {
		t.Error("Expected empty session history initially")
	}

	// Add some interactions
	_, err = agent.Query(context.Background(), "First question")
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}

	_, err = agent.Query(context.Background(), "Second question")
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}

	history = agent.GetSessionHistory()
	if len(history) != 4 { // 2 questions + 2 responses
		t.Errorf("Expected 4 history entries, got %d", len(history))
	}

	// Check history content
	if !strings.Contains(history[0], "First question") {
		t.Error("Expected first entry to contain first question")
	}
	if !strings.Contains(history[2], "Second question") {
		t.Error("Expected third entry to contain second question")
	}

	// Clear history
	agent.ClearSessionHistory()
	history = agent.GetSessionHistory()
	if len(history) != 0 {
		t.Error("Expected empty session history after clearing")
	}
}

func TestInteractiveAgent_ExtractErrorContext(t *testing.T) {
	agent := NewInteractiveAgent(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "error with message",
			input:    "I'm getting this error: cannot build derivation",
			expected: "Error: cannot build derivation",
		},
		{
			name:     "general error mention",
			input:    "There's an error but I don't know what it means",
			expected: "User reported an error condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.extractErrorContext(tt.input)
			if result != tt.expected {
				t.Errorf("extractErrorContext() = %v, want %v", result, tt.expected)
			}
		})
	}
}
