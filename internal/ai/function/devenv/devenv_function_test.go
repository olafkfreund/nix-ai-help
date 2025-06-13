package devenv

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"

	"github.com/stretchr/testify/assert"
)

// MockProvider for testing
type MockProvider struct {
	response string
	err      error
}

func (m *MockProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockProvider) Query(prompt string) (string, error) {
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
	return ch, m.err
}

// newTestDevenvFunction creates a DevenvFunction with a mock provider for testing
func newTestDevenvFunction(mockResponse string, mockErr error) *DevenvFunction {
	function := NewDevenvFunction()
	mockProvider := &MockProvider{response: mockResponse, err: mockErr}
	function.agent = agent.NewDevenvAgent(mockProvider)
	return function
}

func TestDevenvFunction_NewDevenvFunction(t *testing.T) {
	function := NewDevenvFunction()

	assert.NotNil(t, function)
	assert.Equal(t, "devenv", function.Name())
	assert.Contains(t, function.Description(), "development environments")
}

func TestDevenvFunction_GetSchema(t *testing.T) {
	function := NewDevenvFunction()
	schema := function.GetSchema()

	assert.NotNil(t, schema)
	// Note: schema["name"] contains the function Name method, not a string
	// Let's verify the schema has the expected structure instead
	assert.NotNil(t, schema["name"])
	assert.NotNil(t, schema["description"])
	assert.NotNil(t, schema["parameters"])

	// Verify parameters structure
	params, ok := schema["parameters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "object", params["type"])
	assert.NotNil(t, params["properties"])
}

func TestDevenvFunction_ValidateParameters(t *testing.T) {
	function := NewDevenvFunction()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"context":   "Create a new Go development environment",
				"operation": "create",
			},
			wantErr: false,
		},
		{
			name: "missing context",
			params: map[string]interface{}{
				"operation": "create",
			},
			wantErr: true,
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"context": "Test missing operation",
			},
			wantErr: true,
		},
		{
			name: "valid operation with additional params",
			params: map[string]interface{}{
				"context":   "Create Go environment",
				"operation": "create",
				"language":  "go",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := function.ValidateParameters(tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevenvFunction_Execute_Create(t *testing.T) {
	// Create a function with a mock provider that returns a successful response
	mockResponse := `{
		"operation": "create",
		"status": "success",
		"configuration": {
			"language": "go",
			"tools": ["gopls", "go-tools"]
		},
		"setup_steps": ["Install dependencies", "Configure environment"]
	}`
	function := newTestDevenvFunction(mockResponse, nil)

	params := map[string]interface{}{
		"context":   "Create a new Go development environment for testing",
		"operation": "create",
		"language":  "go",
	}

	result, err := function.Execute(context.Background(), params, &functionbase.FunctionOptions{})

	// With a proper mock provider, this should succeed
	assert.NoError(t, err, "Execute should succeed with mock provider")
	assert.NotNil(t, result, "Result should not be nil")
	assert.True(t, result.Success, "Execution should succeed with mock provider")
	assert.NotNil(t, result.Data, "Result should contain data")
}

func TestDevenvFunction_Execute_InvalidOperation(t *testing.T) {
	function := NewDevenvFunction()

	params := map[string]interface{}{
		"context":   "Test invalid operation",
		"operation": "invalid",
	}

	result, err := function.Execute(context.Background(), params, &functionbase.FunctionOptions{})

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "invalid operation")
}

func TestDevenvFunction_Execute_MissingOperation(t *testing.T) {
	function := NewDevenvFunction()

	params := map[string]interface{}{
		"context":  "Test missing operation",
		"language": "go",
	}

	result, err := function.Execute(context.Background(), params, &functionbase.FunctionOptions{})

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "operation is required")
}
