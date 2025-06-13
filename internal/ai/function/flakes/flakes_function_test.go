package flakes

import (
"context"
"testing"
"time"

"nix-ai-help/internal/ai/functionbase"
"github.com/stretchr/testify/assert"
)

func TestNewFlakesFunction(t *testing.T) {
	fn := NewFlakesFunction()
	assert.NotNil(t, fn)
	assert.Equal(t, "flakes", fn.Name())
	assert.NotEmpty(t, fn.Description())
	
	schema := fn.Schema()
	assert.Equal(t, "flakes", schema.Name)
	assert.NotEmpty(t, schema.Parameters)
}

func TestFlakesFunction_ValidateParameters(t *testing.T) {
	fn := NewFlakesFunction()

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
	}{
		{
			name: "valid init operation",
			input: map[string]interface{}{
				"operation": "init",
				"path":      "/tmp/test",
			},
			expectError: false,
		},
		{
			name: "valid build operation",
			input: map[string]interface{}{
				"operation": "build",
				"flake_ref": "nixpkgs#hello",
			},
			expectError: false,
		},
		{
			name: "invalid operation",
			input: map[string]interface{}{
				"operation": "invalid",
			},
			expectError: true,
		},
		{
			name: "missing operation",
			input: map[string]interface{}{
				"flake_ref": "nixpkgs#hello",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
err := fn.ValidateParameters(tt.input)
if tt.expectError {
assert.Error(t, err)
} else {
assert.NoError(t, err)
}
})
	}
}

func TestFlakesFunction_Execute_BasicOperations(t *testing.T) {
	fn := NewFlakesFunction()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "help operation",
			input: map[string]interface{}{
				"operation": "help",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
options := &functionbase.FunctionOptions{}
			result, err := fn.Execute(ctx, tt.input, options)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}
