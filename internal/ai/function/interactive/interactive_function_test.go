package interactive

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewInteractiveFunction(t *testing.T) {
	inf := NewInteractiveFunction()

	if inf == nil {
		t.Fatal("NewInteractiveFunction returned nil")
	}

	if inf.Name() != "interactive" {
		t.Errorf("Expected function name 'interactive', got '%s'", inf.Name())
	}

	if inf.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestInteractiveFunction_ValidateParameters(t *testing.T) {
	inf := NewInteractiveFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid start operation",
			params: map[string]interface{}{
				"operation": "start",
			},
			expectError: false,
		},
		{
			name: "Valid execute operation with command",
			params: map[string]interface{}{
				"operation": "execute",
				"command":   "nixos-rebuild",
				"args":      []string{"switch", "--flake", "."},
				"mode":      "shell",
			},
			expectError: false,
		},
		{
			name: "Valid status operation with session",
			params: map[string]interface{}{
				"operation":  "status",
				"session_id": "session-123",
			},
			expectError: false,
		},
		{
			name: "Missing required operation",
			params: map[string]interface{}{
				"command": "nixos-rebuild",
			},
			expectError: true,
		},
		{
			name: "Invalid operation",
			params: map[string]interface{}{
				"operation": "invalid-operation",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inf.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestInteractiveFunction_Execute_Start(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "start",
		"mode":      "shell",
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute start operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "start" {
		t.Errorf("Expected operation 'start', got '%v'", data["operation"])
	}

	if data["status"] != "success" {
		t.Errorf("Expected status 'success', got '%v'", data["status"])
	}
}

func TestInteractiveFunction_Execute_Status(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation":  "status",
		"session_id": "test-session",
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute status operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "status" {
		t.Errorf("Expected operation 'status', got '%v'", data["operation"])
	}
}

func TestInteractiveFunction_Execute_Execute(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "execute",
		"command":   "nix",
		"args":      []string{"--version"},
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute execute operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "execute" {
		t.Errorf("Expected operation 'execute', got '%v'", data["operation"])
	}
}

func TestInteractiveFunction_Execute_History(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "history",
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute history operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "history" {
		t.Errorf("Expected operation 'history', got '%v'", data["operation"])
	}
}

func TestInteractiveFunction_Execute_Help(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "help",
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute help operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "help" {
		t.Errorf("Expected operation 'help', got '%v'", data["operation"])
	}

	// Check that metadata contains message
	if result.Metadata == nil || result.Metadata["message"] == nil {
		t.Error("Help operation should include message in metadata")
	}
}

func TestInteractiveFunction_Execute_Commands(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "commands",
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute commands operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "commands" {
		t.Errorf("Expected operation 'commands', got '%v'", data["operation"])
	}
}

func TestInteractiveFunction_Execute_Settings(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "settings",
		"settings": map[string]interface{}{
			"verbose": true,
			"timeout": "30s",
		},
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute settings operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "settings" {
		t.Errorf("Expected operation 'settings', got '%v'", data["operation"])
	}
}

func TestInteractiveFunction_Execute_Shortcuts(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "shortcuts",
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute shortcuts operation: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "shortcuts" {
		t.Errorf("Expected operation 'shortcuts', got '%v'", data["operation"])
	}
}

func TestInteractiveFunction_Execute_WithComplexParams(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation":  "execute",
		"command":    "nixos-rebuild",
		"args":       []string{"switch", "--flake", "."},
		"mode":       "shell",
		"session_id": "session-123",
		"settings": map[string]interface{}{
			"verbose": true,
			"timeout": "30s",
		},
	}

	result, err := inf.Execute(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result data to be a map")
	}

	if data["operation"] != "execute" {
		t.Errorf("Expected operation 'execute', got '%v'", data["operation"])
	}

	if data["status"] != "success" {
		t.Errorf("Expected status 'success', got '%v'", data["status"])
	}
}

func TestInteractiveFunction_Execute_InvalidOperation(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"operation": "invalid-operation",
	}

	result, err := inf.Execute(context.Background(), params, nil)

	if err == nil {
		t.Error("Expected error for invalid operation")
	}

	if result != nil {
		t.Error("Expected nil result for invalid operation")
	}

	expectedError := "unsupported interactive operation"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestInteractiveFunction_Execute_MissingOperation(t *testing.T) {
	inf := NewInteractiveFunction()

	params := map[string]interface{}{
		"command": "nix",
	}

	result, err := inf.Execute(context.Background(), params, nil)

	if err == nil {
		t.Error("Expected error for missing operation")
	}

	if result != nil {
		t.Error("Expected nil result for missing operation")
	}

	expectedError := "operation parameter is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestInteractiveFunction_Schema(t *testing.T) {
	inf := NewInteractiveFunction()

	schema := inf.Schema()

	if schema.Name != "interactive" {
		t.Errorf("Expected schema name 'interactive', got '%s'", schema.Name)
	}

	if schema.Description == "" {
		t.Error("Schema description should not be empty")
	}

	if len(schema.Parameters) == 0 {
		t.Error("Schema should have parameters")
	}

	// Check for required operation parameter
	found := false
	for _, param := range schema.Parameters {
		if param.Name == "operation" && param.Required {
			found = true
			break
		}
	}
	if !found {
		t.Error("Schema should have required 'operation' parameter")
	}
}

func TestInteractiveFunction_WithOptions(t *testing.T) {
	inf := NewInteractiveFunction()

	options := &functionbase.FunctionOptions{
		Timeout: 30,
	}

	params := map[string]interface{}{
		"operation": "start",
	}

	result, err := inf.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Failed to execute with options: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution with options")
	}
}
