package store

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

// TestNewStoreFunction tests the creation of a new store function
func TestNewStoreFunction(t *testing.T) {
	function := NewStoreFunction()
	if function == nil {
		t.Fatal("NewStoreFunction() returned nil")
	}

	if function.log == nil {
		t.Error("NewStoreFunction() should initialize logger")
	}
}

// TestStoreFunction_Name tests the Name method
func TestStoreFunction_Name(t *testing.T) {
	function := NewStoreFunction()
	expected := "store"
	if got := function.Name(); got != expected {
		t.Errorf("Name() = %v, want %v", got, expected)
	}
}

// TestStoreFunction_Description tests the Description method
func TestStoreFunction_Description(t *testing.T) {
	function := NewStoreFunction()
	description := function.Description()
	if description == "" {
		t.Error("Description() should not return empty string")
	}
	if !containsKeywords(description, []string{"store", "management", "analysis"}) {
		t.Error("Description() should contain relevant keywords")
	}
}

// TestStoreFunction_Schema tests the Schema method
func TestStoreFunction_Schema(t *testing.T) {
	function := NewStoreFunction()
	schema := function.Schema()

	// Test basic schema properties
	if schema.Name != "store" {
		t.Errorf("Schema.Name = %v, want store", schema.Name)
	}

	if schema.Description == "" {
		t.Error("Schema.Description should not be empty")
	}

	// Test required parameters
	if len(schema.Parameters) == 0 {
		t.Error("Schema should have parameters")
	}

	// Find and test operation parameter
	var operationParam *functionbase.FunctionParameter
	for i := range schema.Parameters {
		if schema.Parameters[i].Name == "operation" {
			operationParam = &schema.Parameters[i]
			break
		}
	}

	if operationParam == nil {
		t.Fatal("Schema should have 'operation' parameter")
	}

	if !operationParam.Required {
		t.Error("Operation parameter should be required")
	}

	expectedOperations := []string{
		"query", "usage", "optimize", "verify", "paths", "deps",
		"roots", "repair", "export", "import", "diff", "vacuum",
	}

	if len(operationParam.Enum) != len(expectedOperations) {
		t.Errorf("Operation parameter should have %d enum values, got %d", len(expectedOperations), len(operationParam.Enum))
	}

	for _, expected := range expectedOperations {
		found := false
		for _, actual := range operationParam.Enum {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Operation parameter should include '%s' in enum", expected)
		}
	}
}

// TestStoreFunction_ValidateParameters tests parameter validation
func TestStoreFunction_ValidateParameters(t *testing.T) {
	function := NewStoreFunction()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "missing operation parameter",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid operation parameter type",
			params: map[string]interface{}{
				"operation": 123,
			},
			wantErr: true,
		},
		{
			name: "invalid operation value",
			params: map[string]interface{}{
				"operation": "invalid_operation",
			},
			wantErr: true,
		},
		{
			name: "valid query operation",
			params: map[string]interface{}{
				"operation": "query",
			},
			wantErr: false,
		},
		{
			name: "valid usage operation",
			params: map[string]interface{}{
				"operation": "usage",
			},
			wantErr: false,
		},
		{
			name: "valid optimize operation",
			params: map[string]interface{}{
				"operation": "optimize",
			},
			wantErr: false,
		},
		{
			name: "valid verify operation",
			params: map[string]interface{}{
				"operation": "verify",
			},
			wantErr: false,
		},
		{
			name: "valid paths operation",
			params: map[string]interface{}{
				"operation": "paths",
			},
			wantErr: false,
		},
		{
			name: "valid deps operation",
			params: map[string]interface{}{
				"operation": "deps",
			},
			wantErr: false,
		},
		{
			name: "valid roots operation",
			params: map[string]interface{}{
				"operation": "roots",
			},
			wantErr: false,
		},
		{
			name: "valid repair operation",
			params: map[string]interface{}{
				"operation": "repair",
			},
			wantErr: false,
		},
		{
			name: "valid export operation",
			params: map[string]interface{}{
				"operation": "export",
			},
			wantErr: false,
		},
		{
			name: "valid import operation",
			params: map[string]interface{}{
				"operation": "import",
			},
			wantErr: false,
		},
		{
			name: "valid diff operation",
			params: map[string]interface{}{
				"operation": "diff",
			},
			wantErr: false,
		},
		{
			name: "valid vacuum operation",
			params: map[string]interface{}{
				"operation": "vacuum",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := function.ValidateParameters(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStoreFunction_Execute_Query tests the query operation
func TestStoreFunction_Execute_Query(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "query",
		"path":      "/nix/store/test-path",
		"pattern":   "hello*",
		"recursive": true,
		"format":    "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_query")

	// Validate query-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_query" {
		t.Error("Result should have type 'store_query'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Query result should contain 'results'")
	}

	if _, ok := data["recommendations"]; !ok {
		t.Error("Query result should contain 'recommendations'")
	}
}

// TestStoreFunction_Execute_Usage tests the usage operation
func TestStoreFunction_Execute_Usage(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":      "usage",
		"size_threshold": "100M",
		"format":         "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_usage")

	// Validate usage-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_usage" {
		t.Error("Result should have type 'store_usage'")
	}

	if _, ok := data["statistics"]; !ok {
		t.Error("Usage result should contain 'statistics'")
	}
}

// TestStoreFunction_Execute_Optimize tests the optimize operation
func TestStoreFunction_Execute_Optimize(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "optimize",
		"dry_run":   true,
		"force":     false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_optimization")

	// Validate optimize-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_optimization" {
		t.Error("Result should have type 'store_optimization'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Optimize result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Verify tests the verify operation
func TestStoreFunction_Execute_Verify(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "verify",
		"path":      "/nix/store/test-path",
		"force":     false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_verification")

	// Validate verify-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_verification" {
		t.Error("Result should have type 'store_verification'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Verify result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Paths tests the paths operation
func TestStoreFunction_Execute_Paths(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "paths",
		"pattern":   "hello*",
		"format":    "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_paths")

	// Validate paths-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_paths" {
		t.Error("Result should have type 'store_paths'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Paths result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Deps tests the deps operation
func TestStoreFunction_Execute_Deps(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "deps",
		"path":      "/nix/store/test-path",
		"recursive": true,
		"format":    "tree",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "dependency_analysis")

	// Validate deps-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "dependency_analysis" {
		t.Error("Result should have type 'dependency_analysis'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Deps result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Roots tests the roots operation
func TestStoreFunction_Execute_Roots(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "roots",
		"dry_run":   true,
		"format":    "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "roots_management")

	// Validate roots-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "roots_management" {
		t.Error("Result should have type 'roots_management'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Roots result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Repair tests the repair operation
func TestStoreFunction_Execute_Repair(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "repair",
		"path":      "/nix/store/corrupted-path",
		"force":     false,
		"dry_run":   true,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_repair")

	// Validate repair-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_repair" {
		t.Error("Result should have type 'store_repair'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Repair result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Export tests the export operation
func TestStoreFunction_Execute_Export(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "export",
		"path":        "/nix/store/test-path",
		"output_file": "/tmp/export.nar",
		"compression": "xz",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_export")

	// Validate export-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_export" {
		t.Error("Result should have type 'store_export'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Export result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Import tests the import operation
func TestStoreFunction_Execute_Import(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "import",
		"path":      "/tmp/import.nar",
		"force":     false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_import")

	// Validate import-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_import" {
		t.Error("Result should have type 'store_import'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Import result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Diff tests the diff operation
func TestStoreFunction_Execute_Diff(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "diff",
		"format":    "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_diff")

	// Validate diff-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_diff" {
		t.Error("Result should have type 'store_diff'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Diff result should contain 'results'")
	}
}

// TestStoreFunction_Execute_Vacuum tests the vacuum operation
func TestStoreFunction_Execute_Vacuum(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "vacuum",
		"dry_run":   true,
		"force":     false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateStoreResult(t, result, "store_vacuum")

	// Validate vacuum-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "store_vacuum" {
		t.Error("Result should have type 'store_vacuum'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Vacuum result should contain 'results'")
	}
}

// TestStoreFunction_Execute_InvalidOperation tests invalid operation handling
func TestStoreFunction_Execute_InvalidOperation(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "invalid_operation",
	}

	_, err := function.Execute(ctx, params, nil)
	if err == nil {
		t.Error("Execute() should return error for invalid operation")
	}
}

// TestStoreFunction_Execute_MissingOperation tests missing operation parameter
func TestStoreFunction_Execute_MissingOperation(t *testing.T) {
	function := NewStoreFunction()
	ctx := context.Background()

	params := map[string]interface{}{}

	_, err := function.Execute(ctx, params, nil)
	if err == nil {
		t.Error("Execute() should return error for missing operation parameter")
	}
}

// Helper function to validate common store result structure
func validateStoreResult(t *testing.T, result *functionbase.FunctionResult, expectedType string) {
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.Success {
		t.Error("Result.Success should be true")
	}

	if result.Data == nil {
		t.Fatal("Result.Data should not be nil")
	}

	if result.Duration <= 0 {
		t.Error("Result.Duration should be positive")
	}

	if result.Timestamp.IsZero() {
		t.Error("Result.Timestamp should be set")
	}

	if result.Metadata == nil {
		t.Error("Result.Metadata should not be nil")
	}

	// Validate data structure
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	// Check for common fields
	if _, ok := data["operation"]; !ok {
		t.Error("Result data should contain 'operation'")
	}

	if _, ok := data["duration"]; !ok {
		t.Error("Result data should contain 'duration'")
	}

	if _, ok := data["timestamp"]; !ok {
		t.Error("Result data should contain 'timestamp'")
	}

	if data["type"] != expectedType {
		t.Errorf("Result data type = %v, want %v", data["type"], expectedType)
	}
}

// Helper function to check if a string contains keywords
func containsKeywords(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if !contains(text, keyword) {
			return false
		}
	}
	return true
}

// Helper function to check if a string contains a substring
func contains(text, substr string) bool {
	return len(text) >= len(substr) &&
		(text == substr ||
			(len(text) > len(substr) &&
				(text[:len(substr)] == substr ||
					text[len(text)-len(substr):] == substr ||
					containsInner(text, substr))))
}

// Helper function to check if substr exists within text
func containsInner(text, substr string) bool {
	for i := 1; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
