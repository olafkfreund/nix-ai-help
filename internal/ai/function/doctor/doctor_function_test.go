package doctor

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewDoctorFunction(t *testing.T) {
	df := NewDoctorFunction()

	if df == nil {
		t.Fatal("NewDoctorFunction returned nil")
	}

	if df.Name() != "doctor" {
		t.Errorf("Expected function name 'doctor', got '%s'", df.Name())
	}

	if df.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestDoctorFunction_GetSchema(t *testing.T) {
	df := NewDoctorFunction()
	schema := df.Schema()

	if schema.Name != "doctor" {
		t.Errorf("Expected schema name 'doctor', got '%s'", schema.Name)
	}

	if schema.Description == "" {
		t.Error("Schema description should not be empty")
	}

	if len(schema.Parameters) == 0 {
		t.Error("Schema should have parameters")
	}
}

func TestDoctorFunction_ValidateParameters(t *testing.T) {
	df := NewDoctorFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid basic parameters",
			params: map[string]interface{}{
				"operation": "check",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"operation":     "diagnose",
				"check_type":    "system",
				"severity":      "warning",
				"component":     "nixos",
				"category":      "configuration",
				"auto_fix":      true,
				"verbose":       true,
				"output_format": "detailed",
			},
			expectError: false,
		},
		{
			name: "Missing required operation",
			params: map[string]interface{}{
				"check_type": "system",
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
		{
			name: "Invalid check_type",
			params: map[string]interface{}{
				"operation":  "check",
				"check_type": "invalid-type",
			},
			expectError: true,
		},
		{
			name: "Invalid severity",
			params: map[string]interface{}{
				"operation": "check",
				"severity":  "invalid-severity",
			},
			expectError: true,
		},
		{
			name: "Invalid output_format",
			params: map[string]interface{}{
				"operation":     "check",
				"output_format": "invalid-format",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := df.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestDoctorFunction_Execute_Check(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation":  "check",
		"check_type": "system",
		"verbose":    true,
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	if result.Data == nil {
		t.Error("Expected result data but got nil")
	}

	// Check response structure
	response, ok := result.Data.(*DoctorResponse)
	if !ok {
		t.Error("Expected DoctorResponse type")
		return
	}

	if response.Operation != "check" {
		t.Errorf("Expected operation 'check', got '%s'", response.Operation)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}
}

func TestDoctorFunction_Execute_Diagnose(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation": "diagnose",
		"severity":  "warning",
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	// Check response structure
	response, ok := result.Data.(*DoctorResponse)
	if !ok {
		t.Error("Expected DoctorResponse type")
		return
	}

	if response.Operation != "diagnose" {
		t.Errorf("Expected operation 'diagnose', got '%s'", response.Operation)
	}
}

func TestDoctorFunction_Execute_Fix(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation": "fix",
		"auto_fix":  true,
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	// Check response structure
	response, ok := result.Data.(*DoctorResponse)
	if !ok {
		t.Error("Expected DoctorResponse type")
		return
	}

	if response.Operation != "fix" {
		t.Errorf("Expected operation 'fix', got '%s'", response.Operation)
	}
}

func TestDoctorFunction_Execute_Status(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation": "status",
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	// Check response structure
	response, ok := result.Data.(*DoctorResponse)
	if !ok {
		t.Error("Expected DoctorResponse type")
		return
	}

	if response.Operation != "status" {
		t.Errorf("Expected operation 'status', got '%s'", response.Operation)
	}

	if response.OverallHealth == "" {
		t.Error("Expected overall health to be set")
	}
}

func TestDoctorFunction_Execute_Summary(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation": "summary",
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	// Check response structure
	response, ok := result.Data.(*DoctorResponse)
	if !ok {
		t.Error("Expected DoctorResponse type")
		return
	}

	if response.Operation != "summary" {
		t.Errorf("Expected operation 'summary', got '%s'", response.Operation)
	}

	if response.Summary == nil {
		t.Error("Expected summary to be set")
	}
}

func TestDoctorFunction_Execute_FullScan(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation": "full-scan",
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	// Check response structure
	response, ok := result.Data.(*DoctorResponse)
	if !ok {
		t.Error("Expected DoctorResponse type")
		return
	}

	if response.Operation != "full-scan" {
		t.Errorf("Expected operation 'full-scan', got '%s'", response.Operation)
	}

	if len(response.Checks) == 0 {
		t.Error("Expected health checks in full scan")
	}
}

func TestDoctorFunction_Execute_InvalidOperation(t *testing.T) {
	df := NewDoctorFunction()

	params := map[string]interface{}{
		"operation": "invalid-operation",
	}

	options := &functionbase.FunctionOptions{}

	result, err := df.Execute(context.Background(), params, options)

	// Should return error for invalid operation
	if err == nil {
		t.Error("Expected error for invalid operation")
	}

	if result == nil || result.Success {
		t.Error("Expected failed result for invalid operation")
	}
}
