package neovim

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/functionbase"
)

// MockProvider implements a simple mock provider for testing
type MockProvider struct{}

func (m *MockProvider) Query(prompt string) (string, error) {
	return "Mock neovim configuration response", nil
}

func (m *MockProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return "Mock neovim configuration response", nil
}

func (m *MockProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return "Mock neovim response", nil
}

func (m *MockProvider) GetPartialResponse() string {
	return ""
}

func (m *MockProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: "Mock neovim configuration response", Done: true}
	close(ch)
	return ch, nil
}

func TestNewNeovimFunction(t *testing.T) {
	function := NewNeovimFunction()

	if function == nil {
		t.Fatal("NewNeovimFunction returned nil")
	}

	if function.Name() != "neovim" {
		t.Errorf("Expected function name 'neovim', got '%s'", function.Name())
	}

	if function.Description() == "" {
		t.Error("Function description should not be empty")
	}

	// Test schema parameters
	schema := function.Schema()
	if len(schema.Parameters) != 20 {
		t.Errorf("Expected 20 parameters, got %d", len(schema.Parameters))
	}

	// Verify required parameters
	operationParam := findParameter(schema.Parameters, "operation")
	if operationParam == nil {
		t.Error("operation parameter not found")
	}
	if operationParam != nil && !operationParam.Required {
		t.Error("operation parameter should be required")
	}
	if operationParam != nil && len(operationParam.Enum) == 0 {
		t.Error("operation parameter should have enum options")
	}
}

func TestNeovimFunction_ValidateParameters(t *testing.T) {
	function := NewNeovimFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid configure operation",
			params: map[string]interface{}{
				"operation":   "configure",
				"config_type": "lua",
				"language":    "go",
			},
			expectError: false,
		},
		{
			name: "valid plugins operation",
			params: map[string]interface{}{
				"operation": "plugins",
				"category":  "lsp",
			},
			expectError: false,
		},
		{
			name: "missing required operation",
			params: map[string]interface{}{
				"config_type": "lua",
			},
			expectError: true,
		},
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"operation": "invalid-op",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := function.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestNeovimFunction_Execute_Configure(t *testing.T) {
	function := NewNeovimFunction()
	// Set up a mock provider for the neovim agent
	mockProvider := &MockProvider{}
	function.neovimAgent.SetProvider(mockProvider)

	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "configure",
		"config_type": "lua",
		"language":    "go",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	data, ok := result.Data.(*NeovimResponse)
	if !ok {
		t.Fatal("Expected data to be a NeovimResponse")
	}

	if data.Operation != "configure" {
		t.Errorf("Expected operation 'configure', got '%s'", data.Operation)
	}
}

func TestNeovimFunction_Execute_Plugins(t *testing.T) {
	function := NewNeovimFunction()
	// Set up a mock provider for the neovim agent
	mockProvider := &MockProvider{}
	function.neovimAgent.SetProvider(mockProvider)

	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "plugins",
		"category":  "lsp",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	data, ok := result.Data.(*NeovimResponse)
	if !ok {
		t.Fatal("Expected data to be a NeovimResponse")
	}

	if data.Operation != "plugins" {
		t.Errorf("Expected operation 'plugins', got '%s'", data.Operation)
	}
}

func TestNeovimFunction_Execute_InvalidOperation(t *testing.T) {
	function := NewNeovimFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "invalid-operation",
	}

	_, err := function.Execute(ctx, params, nil)
	if err == nil {
		t.Fatal("Expected execution to fail for invalid operation")
	}
}

func TestNeovimFunction_Execute_MissingOperation(t *testing.T) {
	function := NewNeovimFunction()
	ctx := context.Background()

	params := map[string]interface{}{}

	_, err := function.Execute(ctx, params, nil)
	if err == nil {
		t.Fatal("Expected execution to fail for missing operation")
	}
}

// Helper function to find a parameter by name
func findParameter(params []functionbase.FunctionParameter, name string) *functionbase.FunctionParameter {
	for i := range params {
		if params[i].Name == name {
			return &params[i]
		}
	}
	return nil
}
