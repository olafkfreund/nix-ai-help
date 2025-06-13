package build

import (
	"context"
	"testing"
)

func TestBuildFunction_NewBuildFunction(t *testing.T) {
	bf := NewBuildFunction()

	if bf == nil {
		t.Fatal("NewBuildFunction returned nil")
	}

	if bf.Name() != "build" {
		t.Errorf("Expected function name 'build', got '%s'", bf.Name())
	}

	if bf.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestBuildFunction_ValidateParameters(t *testing.T) {
	bf := NewBuildFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid build operation",
			params: map[string]interface{}{
				"operation": "build",
				"package":   "hello",
			},
			expectError: false,
		},
		{
			name: "valid troubleshoot operation",
			params: map[string]interface{}{
				"operation":  "troubleshoot",
				"error_logs": "build failed with error...",
			},
			expectError: false,
		},
		{
			name: "missing required operation",
			params: map[string]interface{}{
				"package": "hello",
			},
			expectError: true,
		},
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"operation": "invalid_operation",
			},
			expectError: true, // ValidateParameters should catch enum violations
		},
		{
			name: "valid with all options",
			params: map[string]interface{}{
				"operation":     "build",
				"package":       "hello",
				"configuration": "flake.nix",
				"system":        "x86_64-linux",
				"flake":         true,
				"verbose":       true,
				"max_jobs":      8,
				"cores":         4,
				"build_options": []interface{}{"--impure", "--show-trace"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bf.ValidateParameters(tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateParameters() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestBuildFunction_parseRequest(t *testing.T) {
	bf := NewBuildFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkFunc   func(*BuildRequest) bool
	}{
		{
			name: "basic build request",
			params: map[string]interface{}{
				"operation": "build",
				"package":   "hello",
			},
			expectError: false,
			checkFunc: func(req *BuildRequest) bool {
				return req.Operation == "build" && req.Package == "hello"
			},
		},
		{
			name: "complete build request",
			params: map[string]interface{}{
				"operation":     "build",
				"package":       "hello",
				"configuration": "flake.nix",
				"system":        "x86_64-linux",
				"flake":         true,
				"verbose":       true,
				"max_jobs":      float64(8),
				"cores":         float64(4),
				"build_options": []interface{}{"--impure", "--show-trace"},
			},
			expectError: false,
			checkFunc: func(req *BuildRequest) bool {
				return req.Operation == "build" &&
					req.Package == "hello" &&
					req.Configuration == "flake.nix" &&
					req.System == "x86_64-linux" &&
					req.Flake == true &&
					req.Verbose == true &&
					req.MaxJobs == 8 &&
					req.Cores == 4 &&
					len(req.BuildOptions) == 2
			},
		},
		{
			name: "troubleshoot request",
			params: map[string]interface{}{
				"operation":  "troubleshoot",
				"error_logs": "build failed with dependency error",
				"show_trace": true,
			},
			expectError: false,
			checkFunc: func(req *BuildRequest) bool {
				return req.Operation == "troubleshoot" &&
					req.ErrorLogs == "build failed with dependency error" &&
					req.ShowTrace == true
			},
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"package": "hello",
			},
			expectError: true,
			checkFunc:   nil,
		},
		{
			name: "invalid operation type",
			params: map[string]interface{}{
				"operation": 123,
			},
			expectError: true,
			checkFunc:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := bf.parseRequest(tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("parseRequest() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && tt.checkFunc != nil && !tt.checkFunc(req) {
				t.Errorf("parseRequest() returned request that failed validation check")
			}
		})
	}
}

func TestBuildFunction_validateOperation(t *testing.T) {
	bf := NewBuildFunction()

	tests := []struct {
		name        string
		req         *BuildRequest
		expectError bool
	}{
		{
			name: "valid build operation",
			req: &BuildRequest{
				Operation: "build",
				Package:   "hello",
			},
			expectError: false,
		},
		{
			name: "valid rebuild operation",
			req: &BuildRequest{
				Operation:     "rebuild",
				Configuration: "configuration.nix",
			},
			expectError: false,
		},
		{
			name: "valid troubleshoot operation with logs",
			req: &BuildRequest{
				Operation: "troubleshoot",
				ErrorLogs: "build failed...",
			},
			expectError: false,
		},
		{
			name: "valid troubleshoot operation with package",
			req: &BuildRequest{
				Operation: "troubleshoot",
				Package:   "hello",
			},
			expectError: false,
		},
		{
			name: "valid optimize operation",
			req: &BuildRequest{
				Operation: "optimize",
			},
			expectError: false,
		},
		{
			name: "valid clean operation",
			req: &BuildRequest{
				Operation: "clean",
			},
			expectError: false,
		},
		{
			name: "build without package or config",
			req: &BuildRequest{
				Operation: "build",
			},
			expectError: true,
		},
		{
			name: "troubleshoot without logs or package",
			req: &BuildRequest{
				Operation: "troubleshoot",
			},
			expectError: true,
		},
		{
			name: "invalid operation",
			req: &BuildRequest{
				Operation: "invalid",
			},
			expectError: true,
		},
		{
			name: "valid system architecture",
			req: &BuildRequest{
				Operation: "build",
				Package:   "hello",
				System:    "x86_64-linux",
			},
			expectError: false,
		},
		{
			name: "invalid system architecture",
			req: &BuildRequest{
				Operation: "build",
				Package:   "hello",
				System:    "invalid-system",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bf.validateOperation(tt.req)
			if (err != nil) != tt.expectError {
				t.Errorf("validateOperation() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestBuildFunction_parseAgentResponse(t *testing.T) {
	bf := NewBuildFunction()

	tests := []struct {
		name      string
		response  string
		req       *BuildRequest
		checkFunc func(*BuildResponse) bool
	}{
		{
			name:     "simple response",
			response: "Build completed successfully. The package was built without issues.",
			req:      &BuildRequest{Operation: "build"},
			checkFunc: func(resp *BuildResponse) bool {
				return resp.Status == "success" && resp.Solution != ""
			},
		},
		{
			name: "response with commands",
			response: `Build failed. Try the following commands:
nix build --impure
nix-build -A hello`,
			req: &BuildRequest{Operation: "build"},
			checkFunc: func(resp *BuildResponse) bool {
				return len(resp.SuggestedCommands) >= 2
			},
		},
		{
			name: "response with tips",
			response: `Build optimization tips:
- Use binary caches to speed up builds
- Enable parallel building with -j flag
- Consider using distributed builds`,
			req: &BuildRequest{Operation: "optimize"},
			checkFunc: func(resp *BuildResponse) bool {
				return len(resp.OptimizationTips) >= 2
			},
		},
		{
			name: "troubleshooting response",
			response: `Diagnosis: The build failed due to missing dependencies.
Analysis shows that the package requires libssl which is not available.
Solution: Add openssl to your build inputs.`,
			req: &BuildRequest{Operation: "troubleshoot"},
			checkFunc: func(resp *BuildResponse) bool {
				return resp.DiagnosisDetails != ""
			},
		},
		{
			name:     "response with time estimate",
			response: "This build will take approximately 15 minutes to complete.",
			req:      &BuildRequest{Operation: "build"},
			checkFunc: func(resp *BuildResponse) bool {
				return resp.EstimatedTime != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := bf.parseAgentResponse(tt.response, tt.req)
			if resp == nil {
				t.Fatal("parseAgentResponse returned nil")
			}

			if tt.checkFunc != nil && !tt.checkFunc(resp) {
				t.Errorf("parseAgentResponse() returned response that failed validation check")
			}
		})
	}
}

func TestBuildFunction_extractDiagnosisDetails(t *testing.T) {
	bf := NewBuildFunction()

	tests := []struct {
		name     string
		response string
		expected bool // whether we expect non-empty diagnosis
	}{
		{
			name: "response with diagnosis section",
			response: `Build troubleshooting analysis:
The issue appears to be related to missing dependencies.
The package requires libssl but it's not in the build environment.
This commonly occurs when using impure builds.

Solution: Add openssl to buildInputs`,
			expected: true,
		},
		{
			name:     "response without diagnosis section",
			response: "Simple build error message without analysis.",
			expected: true, // Should fallback to returning the message
		},
		{
			name:     "empty response",
			response: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bf.extractDiagnosisDetails(tt.response)
			hasContent := result != ""
			if hasContent != tt.expected {
				t.Errorf("extractDiagnosisDetails() hasContent = %v, expected %v", hasContent, tt.expected)
			}
		})
	}
}

func TestBuildFunction_Execute(t *testing.T) {
	bf := NewBuildFunction()

	tests := []struct {
		name          string
		params        map[string]interface{}
		expectError   bool
		expectSuccess bool
	}{
		{
			name: "valid build execution",
			params: map[string]interface{}{
				"operation": "build",
				"package":   "hello",
			},
			expectError:   false,
			expectSuccess: true,
		},
		{
			name: "invalid parameters",
			params: map[string]interface{}{
				"operation": "build",
				// Missing package for build operation
			},
			expectError:   false, // Should return error result, not execution error
			expectSuccess: false,
		},
		{
			name: "missing required operation",
			params: map[string]interface{}{
				"package": "hello",
			},
			expectError:   false, // Should return error result, not execution error
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := bf.Execute(ctx, tt.params, nil)

			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if result == nil {
				t.Fatal("Execute() returned nil result")
			}

			if result.Success != tt.expectSuccess {
				t.Errorf("Execute() success = %v, expectSuccess %v", result.Success, tt.expectSuccess)
			}
		})
	}
}

func TestBuildFunction_Schema(t *testing.T) {
	bf := NewBuildFunction()
	schema := bf.Schema()

	if schema.Name != "build" {
		t.Errorf("Expected schema name 'build', got '%s'", schema.Name)
	}

	if len(schema.Parameters) == 0 {
		t.Error("Expected non-empty parameters in schema")
	}

	// Check required parameters
	var hasOperation bool
	for _, param := range schema.Parameters {
		if param.Name == "operation" {
			hasOperation = true
			if !param.Required {
				t.Error("Expected 'operation' parameter to be required")
			}
		}
	}

	if !hasOperation {
		t.Error("Expected 'operation' parameter in schema")
	}
}
