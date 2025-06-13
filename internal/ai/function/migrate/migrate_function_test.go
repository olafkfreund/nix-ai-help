package migrate

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewMigrateFunction(t *testing.T) {
	function := NewMigrateFunction()

	if function == nil {
		t.Fatal("NewMigrateFunction returned nil")
	}

	if function.Name() != "migrate" {
		t.Errorf("Expected function name 'migrate', got '%s'", function.Name())
	}

	if function.Description() != "AI-powered NixOS migration assistance for channels to flakes and configuration updates" {
		t.Errorf("Expected correct description")
	}

	// Test schema parameters
	schema := function.Schema()
	if len(schema.Parameters) != 9 {
		t.Errorf("Expected 9 parameters, got %d", len(schema.Parameters))
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

func TestMigrateFunction_ValidateParameters(t *testing.T) {
	function := NewMigrateFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid analyze operation",
			params: map[string]interface{}{
				"operation": "analyze",
			},
			expectError: false,
		},
		{
			name: "valid to-flakes operation with config path",
			params: map[string]interface{}{
				"operation":   "to-flakes",
				"config_path": "/etc/nixos",
				"dry_run":     true,
			},
			expectError: false,
		},
		{
			name: "valid backup operation with backup name",
			params: map[string]interface{}{
				"operation":   "backup",
				"backup_name": "test-backup",
			},
			expectError: false,
		},
		{
			name:        "missing operation parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"operation": "invalid-operation",
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

func TestMigrateFunction_Execute_Analyze(t *testing.T) {
	function := NewMigrateFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "analyze",
		"config_path": "/etc/nixos",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	// Verify expected fields in analysis
	expectedFields := []string{"operation", "config_path", "current_setup", "nixos_version", "migration_complexity"}
	for _, field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("Expected field '%s' in analysis data", field)
		}
	}
}

func TestMigrateFunction_Execute_ToFlakes(t *testing.T) {
	function := NewMigrateFunction()
	ctx := context.Background()

	tests := []struct {
		name     string
		params   map[string]interface{}
		dryRun   bool
		expected string
	}{
		{
			name: "to-flakes with dry run",
			params: map[string]interface{}{
				"operation": "to-flakes",
				"dry_run":   true,
			},
			dryRun:   true,
			expected: "dry_run",
		},
		{
			name: "to-flakes without dry run",
			params: map[string]interface{}{
				"operation": "to-flakes",
				"dry_run":   false,
			},
			dryRun:   false,
			expected: "in_progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := function.Execute(ctx, tt.params, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if !result.Success {
				t.Errorf("Expected successful execution, got failure: %s", result.Error)
			}

			data, ok := result.Data.(map[string]interface{})
			if !ok {
				t.Fatal("Expected data to be a map")
			}

			// Verify expected fields
			expectedFields := []string{"operation", "migration_steps", "status"}
			for _, field := range expectedFields {
				if _, exists := data[field]; !exists {
					t.Errorf("Expected field '%s' in to-flakes data", field)
				}
			}
		})
	}
}

func TestMigrateFunction_Execute_Backup(t *testing.T) {
	function := NewMigrateFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "backup",
		"backup_name": "test-backup",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	// Verify expected fields
	expectedFields := []string{"operation", "backup_path", "files_backed_up"}
	for _, field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("Expected field '%s' in backup data", field)
		}
	}
}

func TestMigrateFunction_Execute_InvalidOperation(t *testing.T) {
	function := NewMigrateFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "invalid-operation",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Error("Expected execution failure for invalid operation")
	}
}

func TestMigrateFunction_Execute_MissingOperation(t *testing.T) {
	function := NewMigrateFunction()
	ctx := context.Background()

	params := map[string]interface{}{}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Error("Expected execution failure for missing operation")
	}
}

func TestMigrateFunction_Execute_InvalidOperationType(t *testing.T) {
	function := NewMigrateFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": 123, // Invalid type
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Success {
		t.Error("Expected execution failure for invalid operation type")
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
