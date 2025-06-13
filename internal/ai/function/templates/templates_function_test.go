package templates

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

// TestNewTemplatesFunction tests the creation of a new templates function
func TestNewTemplatesFunction(t *testing.T) {
	function := NewTemplatesFunction()
	if function == nil {
		t.Fatal("NewTemplatesFunction() returned nil")
	}

	if function.log == nil {
		t.Error("NewTemplatesFunction() should initialize logger")
	}
}

// TestTemplatesFunction_Name tests the Name method
func TestTemplatesFunction_Name(t *testing.T) {
	function := NewTemplatesFunction()
	expected := "templates"
	if got := function.Name(); got != expected {
		t.Errorf("Name() = %v, want %v", got, expected)
	}
}

// TestTemplatesFunction_Description tests the Description method
func TestTemplatesFunction_Description(t *testing.T) {
	function := NewTemplatesFunction()
	description := function.Description()
	if description == "" {
		t.Error("Description() should not return empty string")
	}
	if !containsKeywords(description, []string{"template", "project", "scaffolding"}) {
		t.Error("Description() should contain relevant keywords")
	}
}

// TestTemplatesFunction_Schema tests the Schema method
func TestTemplatesFunction_Schema(t *testing.T) {
	function := NewTemplatesFunction()
	schema := function.Schema()

	// Test basic schema properties
	if schema.Name != "templates" {
		t.Errorf("Schema.Name = %v, want templates", schema.Name)
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
		"list", "search", "show", "create", "init", "add", "remove",
		"update", "validate", "customize", "preview", "export", "import", "registry",
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

	// Test language parameter enum
	var languageParam *functionbase.FunctionParameter
	for i := range schema.Parameters {
		if schema.Parameters[i].Name == "language" {
			languageParam = &schema.Parameters[i]
			break
		}
	}

	if languageParam != nil {
		expectedLanguages := []string{
			"rust", "go", "python", "javascript", "typescript",
			"haskell", "c", "cpp", "java", "scala", "ruby",
			"php", "dart", "kotlin", "swift", "zig", "elm",
		}

		for _, expected := range expectedLanguages {
			found := false
			for _, actual := range languageParam.Enum {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Language parameter should include '%s' in enum", expected)
			}
		}
	}

	// Test category parameter enum
	var categoryParam *functionbase.FunctionParameter
	for i := range schema.Parameters {
		if schema.Parameters[i].Name == "category" {
			categoryParam = &schema.Parameters[i]
			break
		}
	}

	if categoryParam != nil {
		expectedCategories := []string{
			"web", "cli", "library", "service", "desktop",
			"mobile", "game", "data", "ml", "devops",
			"documentation", "config", "minimal",
		}

		for _, expected := range expectedCategories {
			found := false
			for _, actual := range categoryParam.Enum {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Category parameter should include '%s' in enum", expected)
			}
		}
	}
}

// TestTemplatesFunction_ValidateParameters tests parameter validation
func TestTemplatesFunction_ValidateParameters(t *testing.T) {
	function := NewTemplatesFunction()

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
			name: "valid list operation",
			params: map[string]interface{}{
				"operation": "list",
			},
			wantErr: false,
		},
		{
			name: "valid search operation",
			params: map[string]interface{}{
				"operation": "search",
			},
			wantErr: false,
		},
		{
			name: "valid show operation",
			params: map[string]interface{}{
				"operation": "show",
			},
			wantErr: false,
		},
		{
			name: "valid create operation",
			params: map[string]interface{}{
				"operation": "create",
			},
			wantErr: false,
		},
		{
			name: "valid init operation",
			params: map[string]interface{}{
				"operation": "init",
			},
			wantErr: false,
		},
		{
			name: "valid add operation",
			params: map[string]interface{}{
				"operation": "add",
			},
			wantErr: false,
		},
		{
			name: "valid remove operation",
			params: map[string]interface{}{
				"operation": "remove",
			},
			wantErr: false,
		},
		{
			name: "valid update operation",
			params: map[string]interface{}{
				"operation": "update",
			},
			wantErr: false,
		},
		{
			name: "valid validate operation",
			params: map[string]interface{}{
				"operation": "validate",
			},
			wantErr: false,
		},
		{
			name: "valid customize operation",
			params: map[string]interface{}{
				"operation": "customize",
			},
			wantErr: false,
		},
		{
			name: "valid preview operation",
			params: map[string]interface{}{
				"operation": "preview",
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
			name: "valid registry operation",
			params: map[string]interface{}{
				"operation": "registry",
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

// TestTemplatesFunction_Execute_List tests the list operation
func TestTemplatesFunction_Execute_List(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "list",
		"language":  "rust",
		"category":  "cli",
		"format":    "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_list")

	// Validate list-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_list" {
		t.Error("Result should have type 'template_list'")
	}

	if _, ok := data["templates"]; !ok {
		t.Error("List result should contain 'templates'")
	}

	if _, ok := data["recommendations"]; !ok {
		t.Error("List result should contain 'recommendations'")
	}
}

// TestTemplatesFunction_Execute_Search tests the search operation
func TestTemplatesFunction_Execute_Search(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "search",
		"template_name": "rust",
		"language":      "rust",
		"category":      "cli",
		"format":        "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_search")

	// Validate search-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_search" {
		t.Error("Result should have type 'template_search'")
	}

	if _, ok := data["results"]; !ok {
		t.Error("Search result should contain 'results'")
	}
}

// TestTemplatesFunction_Execute_Show tests the show operation
func TestTemplatesFunction_Execute_Show(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "show",
		"template_name": "rust-cli",
		"format":        "yaml",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_details")

	// Validate show-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_details" {
		t.Error("Result should have type 'template_details'")
	}

	if _, ok := data["template"]; !ok {
		t.Error("Show result should contain 'template'")
	}
}

// TestTemplatesFunction_Execute_Create tests the create operation
func TestTemplatesFunction_Execute_Create(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "create",
		"template_name": "rust-cli",
		"project_name":  "my-rust-project",
		"output_dir":    "./test-project",
		"interactive":   false,
		"git_init":      true,
		"force":         false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "project_creation")

	// Validate create-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "project_creation" {
		t.Error("Result should have type 'project_creation'")
	}

	if _, ok := data["project"]; !ok {
		t.Error("Create result should contain 'project'")
	}
}

// TestTemplatesFunction_Execute_Init tests the init operation
func TestTemplatesFunction_Execute_Init(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":    "init",
		"template_url": "github:nixos/templates#rust",
		"project_name": "my-project",
		"output_dir":   "./",
		"interactive":  true,
		"git_init":     true,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_init")

	// Validate init-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_init" {
		t.Error("Result should have type 'template_init'")
	}

	if _, ok := data["initialization"]; !ok {
		t.Error("Init result should contain 'initialization'")
	}
}

// TestTemplatesFunction_Execute_Add tests the add operation
func TestTemplatesFunction_Execute_Add(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "add",
		"template_name": "custom-rust",
		"template_url":  "github:user/custom-template",
		"force":         false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_addition")

	// Validate add-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_addition" {
		t.Error("Result should have type 'template_addition'")
	}

	if _, ok := data["added_template"]; !ok {
		t.Error("Add result should contain 'added_template'")
	}
}

// TestTemplatesFunction_Execute_Remove tests the remove operation
func TestTemplatesFunction_Execute_Remove(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "remove",
		"template_name": "custom-rust",
		"force":         false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_removal")

	// Validate remove-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_removal" {
		t.Error("Result should have type 'template_removal'")
	}

	if _, ok := data["removed_template"]; !ok {
		t.Error("Remove result should contain 'removed_template'")
	}
}

// TestTemplatesFunction_Execute_Update tests the update operation
func TestTemplatesFunction_Execute_Update(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "update",
		"template_name": "rust-cli",
		"force":         false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_update")

	// Validate update-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_update" {
		t.Error("Result should have type 'template_update'")
	}

	if _, ok := data["update_info"]; !ok {
		t.Error("Update result should contain 'update_info'")
	}
}

// TestTemplatesFunction_Execute_Validate tests the validate operation
func TestTemplatesFunction_Execute_Validate(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "validate",
		"template_name": "rust-cli",
		"format":        "json",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_validation")

	// Validate validation-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_validation" {
		t.Error("Result should have type 'template_validation'")
	}

	if _, ok := data["validation"]; !ok {
		t.Error("Validate result should contain 'validation'")
	}
}

// TestTemplatesFunction_Execute_Customize tests the customize operation
func TestTemplatesFunction_Execute_Customize(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "customize",
		"template_name": "rust-cli",
		"interactive":   true,
		"format":        "yaml",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_customization")

	// Validate customize-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_customization" {
		t.Error("Result should have type 'template_customization'")
	}

	if _, ok := data["customization"]; !ok {
		t.Error("Customize result should contain 'customization'")
	}
}

// TestTemplatesFunction_Execute_Preview tests the preview operation
func TestTemplatesFunction_Execute_Preview(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "preview",
		"template_name": "rust-cli",
		"project_name":  "preview-project",
		"format":        "tree",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_preview")

	// Validate preview-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_preview" {
		t.Error("Result should have type 'template_preview'")
	}

	if _, ok := data["preview"]; !ok {
		t.Error("Preview result should contain 'preview'")
	}
}

// TestTemplatesFunction_Execute_Export tests the export operation
func TestTemplatesFunction_Execute_Export(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":     "export",
		"template_name": "rust-cli",
		"output_file":   "/tmp/template-export.tar.gz",
		"format":        "json",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_export")

	// Validate export-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_export" {
		t.Error("Result should have type 'template_export'")
	}

	if _, ok := data["export_info"]; !ok {
		t.Error("Export result should contain 'export_info'")
	}
}

// TestTemplatesFunction_Execute_Import tests the import operation
func TestTemplatesFunction_Execute_Import(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":    "import",
		"template_url": "/tmp/template-import.tar.gz",
		"force":        false,
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_import")

	// Validate import-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_import" {
		t.Error("Result should have type 'template_import'")
	}

	if _, ok := data["import_info"]; !ok {
		t.Error("Import result should contain 'import_info'")
	}
}

// TestTemplatesFunction_Execute_Registry tests the registry operation
func TestTemplatesFunction_Execute_Registry(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "registry",
		"format":    "table",
	}

	result, err := function.Execute(ctx, params, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	validateTemplatesResult(t, result, "template_registry")

	// Validate registry-specific data
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result.Data should be a map")
	}

	if data["type"] != "template_registry" {
		t.Error("Result should have type 'template_registry'")
	}

	if _, ok := data["registry"]; !ok {
		t.Error("Registry result should contain 'registry'")
	}
}

// TestTemplatesFunction_Execute_InvalidOperation tests invalid operation handling
func TestTemplatesFunction_Execute_InvalidOperation(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "invalid_operation",
	}

	_, err := function.Execute(ctx, params, nil)
	if err == nil {
		t.Error("Execute() should return error for invalid operation")
	}
}

// TestTemplatesFunction_Execute_MissingOperation tests missing operation parameter
func TestTemplatesFunction_Execute_MissingOperation(t *testing.T) {
	function := NewTemplatesFunction()
	ctx := context.Background()

	params := map[string]interface{}{}

	_, err := function.Execute(ctx, params, nil)
	if err == nil {
		t.Error("Execute() should return error for missing operation parameter")
	}
}

// Helper function to validate common templates result structure
func validateTemplatesResult(t *testing.T, result *functionbase.FunctionResult, expectedType string) {
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
