package explainoption

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewExplainOptionFunction(t *testing.T) {
	eof := NewExplainOptionFunction()

	if eof == nil {
		t.Fatal("NewExplainOptionFunction returned nil")
	}

	if eof.Name() != "explain-option" {
		t.Errorf("Expected function name 'explain-option', got '%s'", eof.Name())
	}

	if eof.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestExplainOptionFunction_ValidateParameters(t *testing.T) {
	eof := NewExplainOptionFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"option": "services.openssh.enable",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"option":        "boot.loader.systemd-boot.enable",
				"module":        "boot",
				"show_examples": true,
				"detailed":      true,
			},
			expectError: false,
		},
		{
			name: "Missing option parameter",
			params: map[string]interface{}{
				"module": "services",
			},
			expectError: true,
		},
		{
			name: "Empty option parameter",
			params: map[string]interface{}{
				"option": "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eof.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestExplainOptionFunction_ParseRequest(t *testing.T) {
	eof := NewExplainOptionFunction()

	params := map[string]interface{}{
		"option":        "services.nginx.enable",
		"module":        "services",
		"show_examples": true,
		"detailed":      false,
	}

	request, err := eof.parseRequest(params)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if request.Option != "services.nginx.enable" {
		t.Errorf("Expected option 'services.nginx.enable', got '%s'", request.Option)
	}

	if request.Module != "services" {
		t.Errorf("Expected module 'services', got '%s'", request.Module)
	}

	if !request.ShowExamples {
		t.Error("Expected ShowExamples to be true")
	}

	if request.Detailed {
		t.Error("Expected Detailed to be false")
	}
}

func TestExplainOptionFunction_GenerateBasicDescription(t *testing.T) {
	eof := NewExplainOptionFunction()

	tests := []struct {
		name     string
		option   string
		expected string
	}{
		{
			name:     "Service enable option",
			option:   "services.openssh.enable",
			expected: "Enables the services.openssh service/feature",
		},
		{
			name:     "Boot option",
			option:   "boot.loader.systemd-boot.enable",
			expected: "Boot-related configuration option: loader.systemd-boot.enable",
		},
		{
			name:     "Networking option",
			option:   "networking.hostName",
			expected: "Network configuration option: hostName",
		},
		{
			name:     "Generic option",
			option:   "system.stateVersion",
			expected: "NixOS configuration option: system.stateVersion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := eof.generateBasicDescription(tt.option)
			if description != tt.expected {
				t.Errorf("Expected description '%s', got '%s'", tt.expected, description)
			}
		})
	}
}

func TestExplainOptionFunction_GenerateExamples(t *testing.T) {
	eof := NewExplainOptionFunction()

	tests := []struct {
		name     string
		option   string
		expected []string
	}{
		{
			name:   "Enable option",
			option: "services.openssh.enable",
			expected: []string{
				"services.openssh.enable = true;",
				"services.openssh.enable = false;",
			},
		},
		{
			name:   "Port option",
			option: "services.openssh.port",
			expected: []string{
				"services.openssh.port = 22;",
				"services.openssh.port = 8080;",
			},
		},
		{
			name:   "ExtraConfig option",
			option: "services.nginx.extraConfig",
			expected: []string{
				"services.nginx.extraConfig = ''\n  # Additional configuration here\n'';",
			},
		},
		{
			name:   "Generic option",
			option: "networking.hostName",
			expected: []string{
				"networking.hostName = \"value\";",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			examples := eof.generateExamples(tt.option)
			if len(examples) != len(tt.expected) {
				t.Errorf("Expected %d examples, got %d", len(tt.expected), len(examples))
				return
			}

			for i, expected := range tt.expected {
				if examples[i] != expected {
					t.Errorf("Expected example %d to be '%s', got '%s'", i, expected, examples[i])
				}
			}
		})
	}
}

func TestExplainOptionFunction_FindRelatedOptions(t *testing.T) {
	eof := NewExplainOptionFunction()

	tests := []struct {
		name     string
		option   string
		expected []string
	}{
		{
			name:   "Service option",
			option: "services.openssh.enable",
			expected: []string{
				"services.openssh.package",
				"services.openssh.extraConfig",
			},
		},
		{
			name:   "Boot option",
			option: "boot.loader.systemd-boot.enable",
			expected: []string{
				"boot.loader.grub.enable",
				"boot.kernelPackages",
			},
		},
		{
			name:   "Networking option",
			option: "networking.hostName",
			expected: []string{
				"networking.firewall.enable",
				"networking.networkmanager.enable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			related := eof.findRelatedOptions(tt.option)

			// Check that all expected options are present
			for _, expected := range tt.expected {
				found := false
				for _, rel := range related {
					if rel == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected related option '%s' not found in %v", expected, related)
				}
			}

			// Check that the original option is not in the related list
			for _, rel := range related {
				if rel == tt.option {
					t.Errorf("Original option '%s' should not be in related options list", tt.option)
				}
			}
		})
	}
}

func TestExplainOptionFunction_Execute(t *testing.T) {
	eof := NewExplainOptionFunction()

	params := map[string]interface{}{
		"option":        "services.openssh.enable",
		"show_examples": true,
		"detailed":      true,
	}

	options := &functionbase.FunctionOptions{
		ProgressCallback: func(progress functionbase.Progress) {
			// Progress callback for testing
		},
	}

	// Note: This will work without MCP since we have fallback logic
	result, err := eof.Execute(context.Background(), params, options)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.Success {
		t.Errorf("Result should indicate success, got error: %s", result.Error)
	}

	// Verify response structure
	response, ok := result.Data.(*ExplainOptionResponse)
	if !ok {
		t.Fatal("Result data should be ExplainOptionResponse")
	}

	if response.Option != "services.openssh.enable" {
		t.Errorf("Expected option 'services.openssh.enable', got '%s'", response.Option)
	}

	if response.Description == "" {
		t.Error("Description should not be empty")
	}

	if len(response.Examples) == 0 {
		t.Error("Examples should not be empty when show_examples is true")
	}

	if len(response.RelatedOptions) == 0 {
		t.Error("RelatedOptions should not be empty")
	}
}

func TestExplainOptionFunction_ExecuteWithMissingOption(t *testing.T) {
	eof := NewExplainOptionFunction()

	params := map[string]interface{}{
		"module": "services",
	}

	result, err := eof.Execute(context.Background(), params, nil)

	if err != nil {
		t.Fatalf("Execute should not return error for validation issues: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Success {
		t.Error("Result should indicate failure due to missing option")
	}

	if !contains(result.Error, "option") {
		t.Errorf("Error message should mention missing option, got: %s", result.Error)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
