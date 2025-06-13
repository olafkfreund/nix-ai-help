package diagnose

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/nixos"
)

func TestNewDiagnoseFunction(t *testing.T) {
	df := NewDiagnoseFunction()

	if df.Name() != "diagnose" {
		t.Errorf("Expected function name 'diagnose', got '%s'", df.Name())
	}

	if df.Description() == "" {
		t.Error("Expected non-empty description")
	}

	schema := df.Schema()
	if len(schema.Parameters) == 0 {
		t.Error("Expected parameters in schema")
	}
}

func TestDiagnoseFunctionValidation(t *testing.T) {
	df := NewDiagnoseFunction()

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "valid basic parameters",
			params: map[string]interface{}{
				"user_description": "Service won't start",
				"include_steps":    true,
			},
			hasError: false,
		},
		{
			name: "valid with system info",
			params: map[string]interface{}{
				"error_message": "Permission denied",
				"system_info": map[string]interface{}{
					"nixos_version":   "23.11",
					"is_flake_system": true,
				},
			},
			hasError: false,
		},
		{
			name: "invalid analysis type",
			params: map[string]interface{}{
				"user_description": "Problem",
				"analysis_type":    "invalid_type",
			},
			hasError: true,
		},
		{
			name: "invalid severity",
			params: map[string]interface{}{
				"user_description": "Problem",
				"severity":         "invalid_severity",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := df.ValidateParameters(tt.params)
			if tt.hasError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestDiagnoseFunctionExecution(t *testing.T) {
	df := NewDiagnoseFunction()
	ctx := context.Background()

	// Test basic execution
	params := map[string]interface{}{
		"user_description":   "My nginx service won't start after rebuilding",
		"error_message":      "Failed to bind to port 80",
		"analysis_type":      "service",
		"include_steps":      true,
		"include_prevention": true,
	}

	options := &functionbase.FunctionOptions{
		Timeout: 30 * time.Second,
	}

	result, err := df.Execute(ctx, params, options)

	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	if !result.Success {
		t.Errorf("Expected successful result, got error: %s", result.Error)
	}

	// Check that result data is DiagnoseResponse
	response, ok := result.Data.(*DiagnoseResponse)
	if !ok {
		t.Errorf("Expected DiagnoseResponse, got %T", result.Data)
		return
	}

	if response.Summary == "" {
		t.Error("Expected non-empty summary")
	}

	if response.Severity == "" {
		t.Error("Expected severity to be set")
	}
}

func TestDiagnoseFunctionWithProgress(t *testing.T) {
	df := NewDiagnoseFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"user_description": "Build failed",
		"log_data":         "error: builder failed with exit code 1",
		"analysis_type":    "build",
	}

	progressChan := make(chan functionbase.Progress, 10)
	options := &functionbase.FunctionOptions{
		ProgressCallback: func(progress functionbase.Progress) {
			progressChan <- progress
		},
	}

	go func() {
		_, err := df.Execute(ctx, params, options)
		if err != nil {
			t.Errorf("Execution with progress failed: %v", err)
		}
		close(progressChan)
	}()

	progressCount := 0
	for progress := range progressChan {
		progressCount++
		if progress.Percentage < 0 || progress.Percentage > 100 {
			t.Errorf("Invalid progress percentage: %f", progress.Percentage)
		}
	}

	if progressCount == 0 {
		t.Error("Expected progress updates but got none")
	}
}

func TestParseRequest(t *testing.T) {
	df := NewDiagnoseFunction()

	params := map[string]interface{}{
		"user_description": "Service problem",
		"log_data":         "error logs here",
		"analysis_type":    "service",
		"include_steps":    true,
		"system_info": map[string]interface{}{
			"nixos_version":   "23.11",
			"nix_version":     "2.18.1",
			"is_flake_system": true,
		},
	}

	request, err := df.parseRequest(params)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if request.UserDescription != "Service problem" {
		t.Errorf("Expected user_description 'Service problem', got '%s'", request.UserDescription)
	}

	if request.LogData != "error logs here" {
		t.Errorf("Expected log_data 'error logs here', got '%s'", request.LogData)
	}

	if request.AnalysisType != "service" {
		t.Errorf("Expected analysis_type 'service', got '%s'", request.AnalysisType)
	}

	if !request.IncludeSteps {
		t.Error("Expected include_steps to be true")
	}

	if request.SystemInfo == nil {
		t.Error("Expected system_info to be parsed")
	} else {
		if request.SystemInfo.NixOSVersion != "23.11" {
			t.Errorf("Expected nixos_version '23.11', got '%s'", request.SystemInfo.NixOSVersion)
		}
		if !request.SystemInfo.IsFlakeSystem {
			t.Error("Expected is_flake_system to be true")
		}
	}
}

func TestParseRequestValidation(t *testing.T) {
	df := NewDiagnoseFunction()

	// Test with empty parameters
	emptyParams := map[string]interface{}{}
	_, err := df.parseRequest(emptyParams)
	if err == nil {
		t.Error("Expected error for empty parameters")
	}

	// Test with at least one valid input
	validParams := map[string]interface{}{
		"user_description": "Some problem",
	}
	_, err = df.parseRequest(validParams)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got: %v", err)
	}
}

func TestAnalyzeErrorMessage(t *testing.T) {
	df := NewDiagnoseFunction()

	tests := []struct {
		name         string
		errorMsg     string
		expectDiag   bool
		expectedType string
	}{
		{
			name:         "permission denied",
			errorMsg:     "Permission denied when accessing file",
			expectDiag:   true,
			expectedType: "permission",
		},
		{
			name:         "no space left",
			errorMsg:     "No space left on device",
			expectDiag:   true,
			expectedType: "storage",
		},
		{
			name:         "connection refused",
			errorMsg:     "Connection refused to server",
			expectDiag:   true,
			expectedType: "network",
		},
		{
			name:         "build failure",
			errorMsg:     "builder for '/nix/store/...' failed",
			expectDiag:   true,
			expectedType: "build",
		},
		{
			name:       "generic error",
			errorMsg:   "Some unknown error",
			expectDiag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := df.analyzeErrorMessage(tt.errorMsg)

			if tt.expectDiag && diag == nil {
				t.Error("Expected diagnostic but got nil")
			}

			if !tt.expectDiag && diag != nil {
				t.Error("Expected no diagnostic but got one")
			}

			if tt.expectDiag && diag != nil {
				if diag.ErrorType != tt.expectedType {
					t.Errorf("Expected error type '%s', got '%s'", tt.expectedType, diag.ErrorType)
				}
			}
		})
	}
}

func TestAnalyzeLogData(t *testing.T) {
	df := NewDiagnoseFunction()

	logData := `
2024-01-01 10:00:00 INFO: Service starting
2024-01-01 10:00:01 ERROR: Failed to connect to database
2024-01-01 10:00:02 INFO: Retrying connection
2024-01-01 10:00:03 FAILED: Connection attempt failed
2024-01-01 10:00:04 INFO: Service stopped
`

	diagnostics := df.analyzeLogData(logData)

	if len(diagnostics) == 0 {
		t.Error("Expected diagnostics from log analysis")
	}

	// Should find both error and failure
	hasError := false
	hasFailure := false

	for _, diag := range diagnostics {
		if diag.ErrorType == "error" {
			hasError = true
		}
		if diag.ErrorType == "failure" {
			hasFailure = true
		}
	}

	if !hasError {
		t.Error("Expected to find error diagnostic")
	}

	if !hasFailure {
		t.Error("Expected to find failure diagnostic")
	}
}

func TestGenerateRecommendations(t *testing.T) {
	df := NewDiagnoseFunction()

	request := &DiagnoseRequest{
		AnalysisType: "service",
	}

	diagnostics := []nixos.Diagnostic{
		{
			ErrorType: "permission",
			Issue:     "Permission issue",
			Severity:  "high",
		},
	}

	recommendations := df.generateRecommendations(request, diagnostics)

	if len(recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}

	// Should have service-specific and permission-specific recommendations
	hasServiceRec := false
	hasPermissionRec := false

	for _, rec := range recommendations {
		if rec.Type == "command" && rec.Description == "Check service status and logs" {
			hasServiceRec = true
		}
		if rec.Type == "command" && rec.Description == "Check file permissions and ownership" {
			hasPermissionRec = true
		}
	}

	if !hasServiceRec {
		t.Error("Expected service-specific recommendation")
	}

	if !hasPermissionRec {
		t.Error("Expected permission-specific recommendation")
	}
}
