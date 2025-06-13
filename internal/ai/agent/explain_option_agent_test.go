package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockProvider for testing
type MockProviderForOptions struct {
	mock.Mock
}

func (m *MockProviderForOptions) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockProviderForOptions) GetPartialResponse() string {
	return ""
}

func (m *MockProviderForOptions) Query(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}

func (m *MockProviderForOptions) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockProviderForOptions) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: "mock stream response", Done: true}
	close(ch)
	return ch, nil
}

func TestExplainOptionAgent_Query(t *testing.T) {
	tests := []struct {
		name          string
		question      string
		mockResponse  string
		expectedError bool
		shouldContain []string
	}{
		{
			name:          "nginx service option",
			question:      "What does services.nginx.enable do?",
			mockResponse:  "services.nginx.enable is a boolean option that enables the nginx web server service",
			expectedError: false,
			shouldContain: []string{"services.nginx.enable", "nginx", "web server"},
		},
		{
			name:          "git program option",
			question:      "Explain programs.git.enable",
			mockResponse:  "programs.git.enable is a boolean option that enables git version control system",
			expectedError: false,
			shouldContain: []string{"programs.git.enable", "git", "version control"},
		},
		{
			name:          "hardware option",
			question:      "What is hardware.bluetooth.enable?",
			mockResponse:  "hardware.bluetooth.enable is a boolean option that enables bluetooth support",
			expectedError: false,
			shouldContain: []string{"hardware.bluetooth.enable", "bluetooth"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockProviderForOptions)
			agent := NewExplainOptionAgent(mockProvider, nil)

			// Set role
			err := agent.SetRole(roles.RoleExplainOption)
			assert.NoError(t, err)

			// Build the prompt as the agent would
			optionCtx, _ := agent.buildOptionContext(context.Background(), tt.question)
			prompt := agent.buildOptionPrompt(tt.question, optionCtx)

			// Setup mock expectation for QueryWithContext only
			mockProvider.On("QueryWithContext", mock.Anything, prompt).Return(tt.mockResponse, nil)

			// Execute
			response, err := agent.Query(context.Background(), tt.question)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, response)
				assert.Equal(t, tt.mockResponse, response)
			}

			// Verify mock expectations
			mockProvider.AssertExpectations(t)
		})
	}
}

func TestExplainOptionAgent_BuildOptionContext(t *testing.T) {
	agent := NewExplainOptionAgent(nil, nil)

	tests := []struct {
		name               string
		question           string
		expectedOptionPath string
		expectedCategory   string
		expectedPackage    string
		expectedService    string
	}{
		{
			name:               "nginx service",
			question:           "services.nginx.enable",
			expectedOptionPath: "services.nginx.enable",
			expectedCategory:   "System Services",
			expectedPackage:    "nginx",
			expectedService:    "nginx",
		},
		{
			name:               "git program",
			question:           "programs.git.enable",
			expectedOptionPath: "programs.git.enable",
			expectedCategory:   "System Programs",
			expectedPackage:    "git",
			expectedService:    "",
		},
		{
			name:               "networking option",
			question:           "networking.firewall.enable",
			expectedOptionPath: "networking.firewall.enable",
			expectedCategory:   "Network Configuration",
			expectedPackage:    "",
			expectedService:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := agent.buildOptionContext(context.Background(), tt.question)

			assert.NoError(t, err)
			assert.NotNil(t, ctx)
			assert.Equal(t, tt.expectedOptionPath, ctx.OptionPath)
			assert.Equal(t, tt.expectedCategory, ctx.Category)
			assert.Equal(t, tt.expectedPackage, ctx.PackageName)
			assert.Equal(t, tt.expectedService, ctx.ServiceName)
		})
	}
}

func TestExplainOptionAgent_ExtractOptionPath(t *testing.T) {
	agent := NewExplainOptionAgent(nil, nil)

	tests := []struct {
		name     string
		question string
		expected string
	}{
		{
			name:     "simple service option",
			question: "What does services.nginx.enable do?",
			expected: "services.nginx.enable",
		},
		{
			name:     "program option",
			question: "Explain programs.git.enable",
			expected: "programs.git.enable",
		},
		{
			name:     "nested option",
			question: "services.postgresql.settings.port",
			expected: "services.postgresql.settings.port",
		},
		{
			name:     "no option found",
			question: "How do I configure nginx?",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.extractOptionPath(tt.question)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExplainOptionAgent_CategorizeOption(t *testing.T) {
	agent := NewExplainOptionAgent(nil, nil)

	tests := []struct {
		name             string
		optionPath       string
		expectedCategory string
	}{
		{
			name:             "service option",
			optionPath:       "services.nginx.enable",
			expectedCategory: "System Services",
		},
		{
			name:             "program option",
			optionPath:       "programs.git.enable",
			expectedCategory: "System Programs",
		},
		{
			name:             "networking option",
			optionPath:       "networking.firewall.enable",
			expectedCategory: "Network Configuration",
		},
		{
			name:             "hardware option",
			optionPath:       "hardware.bluetooth.enable",
			expectedCategory: "Hardware Configuration",
		},
		{
			name:             "unknown option",
			optionPath:       "unknown.option.path",
			expectedCategory: "General Configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.categorizeOption(tt.optionPath)
			assert.Equal(t, tt.expectedCategory, result)
		})
	}
}

func TestExplainOptionAgent_BuildOptionPrompt(t *testing.T) {
	agent := NewExplainOptionAgent(nil, nil)

	// Set role first
	err := agent.SetRole(roles.RoleExplainOption)
	assert.NoError(t, err)

	question := "What does services.nginx.enable do?"
	optionCtx := &OptionContext{
		OptionPath:  "services.nginx.enable",
		Category:    "System Services",
		PackageName: "nginx",
		ServiceName: "nginx",
		UseCase:     "Configure web server",
		Description: "Enable nginx web server",
	}

	prompt := agent.buildOptionPrompt(question, optionCtx)

	// Verify prompt contains expected elements
	assert.Contains(t, prompt, "NixOS Option Explanation Request")
	assert.Contains(t, prompt, question)
	assert.Contains(t, prompt, "services.nginx.enable")
	assert.Contains(t, prompt, "System Services")
	assert.Contains(t, prompt, "nginx")
	assert.Contains(t, prompt, "Practical configuration examples")
}

func TestExplainOptionAgent_FindRelatedOptions(t *testing.T) {
	agent := NewExplainOptionAgent(nil, nil)

	tests := []struct {
		name          string
		optionPath    string
		expectedCount int
		shouldContain []string
	}{
		{
			name:          "service option",
			optionPath:    "services.nginx.enable",
			expectedCount: 4, // enable, package, user, group, configFile minus the original
			shouldContain: []string{"services.nginx.package", "services.nginx.user"},
		},
		{
			name:          "program option",
			optionPath:    "programs.git.enable",
			expectedCount: 4,
			shouldContain: []string{"programs.git.package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			related := agent.findRelatedOptions(tt.optionPath)

			assert.Len(t, related, tt.expectedCount)

			for _, shouldContain := range tt.shouldContain {
				assert.Contains(t, related, shouldContain)
			}

			// Verify original option is not included
			assert.NotContains(t, related, tt.optionPath)
		})
	}
}

func TestExplainOptionAgent_QueryWithContext(t *testing.T) {
	mockProvider := new(MockProviderForOptions)
	agent := NewExplainOptionAgent(mockProvider, nil)

	// Set role
	err := agent.SetRole(roles.RoleExplainOption)
	assert.NoError(t, err)

	question := "What does services.nginx.enable do?"
	optionCtx := &OptionContext{
		OptionPath:  "services.nginx.enable",
		Category:    "System Services",
		PackageName: "nginx",
		ServiceName: "nginx",
	}

	expectedResponse := "nginx service explanation"

	// Setup mock - it should be called with a prompt that contains our context
	mockProvider.On("QueryWithContext", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return strings.Contains(prompt, "services.nginx.enable") &&
			strings.Contains(prompt, "System Services") &&
			strings.Contains(prompt, "nginx")
	})).Return(expectedResponse, nil)

	// Execute
	response, err := agent.QueryWithContext(context.Background(), question, optionCtx)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockProvider.AssertExpectations(t)
}

func TestExplainOptionAgent_SetRole(t *testing.T) {
	agent := NewExplainOptionAgent(nil, nil)

	// Test valid role
	err := agent.SetRole(roles.RoleExplainOption)
	assert.NoError(t, err)
	assert.Equal(t, roles.RoleExplainOption, agent.role)

	// Test invalid role (this should still work as ValidateRole accepts it)
	err = agent.SetRole(roles.RoleAsk)
	assert.NoError(t, err) // Should work since RoleAsk is valid
}

func TestExplainOptionAgent_InvalidRole(t *testing.T) {
	mockProvider := new(MockProviderForOptions)
	agent := NewExplainOptionAgent(mockProvider, nil)

	// Clear the role to test the no-role scenario
	agent.role = ""

	// Execute
	_, err := agent.Query(context.Background(), "test question")

	// Should fail due to no role set
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent role not set")
}
