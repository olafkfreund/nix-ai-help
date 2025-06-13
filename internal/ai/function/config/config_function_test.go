package config

import (
	"context"
	"testing"
)

func TestConfigFunction_NewConfigFunction(t *testing.T) {
	cf := NewConfigFunction()

	if cf == nil {
		t.Fatal("NewConfigFunction returned nil")
	}

	if cf.Name() != "config" {
		t.Errorf("Expected function name 'config', got '%s'", cf.Name())
	}

	if cf.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestConfigFunction_ValidateParameters(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid generate operation",
			params: map[string]interface{}{
				"operation":   "generate",
				"config_type": "nixos",
			},
			expectError: false,
		},
		{
			name: "valid validate operation",
			params: map[string]interface{}{
				"operation":   "validate",
				"config_path": "/etc/nixos/configuration.nix",
			},
			expectError: false,
		},
		{
			name: "missing required operation",
			params: map[string]interface{}{
				"config_type": "nixos",
			},
			expectError: true,
		},
		{
			name: "valid with all options",
			params: map[string]interface{}{
				"operation":    "generate",
				"config_type":  "nixos",
				"system":       "x86_64-linux",
				"home_manager": true,
				"flakes":       true,
				"validate":     true,
				"options": map[string]interface{}{
					"hostname": "mynixos",
					"timezone": "America/New_York",
				},
				"services": []interface{}{"ssh", "docker"},
				"packages": []interface{}{"git", "vim"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cf.ValidateParameters(tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateParameters() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestConfigFunction_parseRequest(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkFunc   func(*ConfigRequest) bool
	}{
		{
			name: "basic generate request",
			params: map[string]interface{}{
				"operation":   "generate",
				"config_type": "nixos",
			},
			expectError: false,
			checkFunc: func(req *ConfigRequest) bool {
				return req.Operation == "generate" && req.ConfigType == "nixos"
			},
		},
		{
			name: "complete config request",
			params: map[string]interface{}{
				"operation":    "generate",
				"config_type":  "nixos",
				"system":       "x86_64-linux",
				"home_manager": true,
				"flakes":       true,
				"validate":     true,
				"options": map[string]interface{}{
					"hostname": "mynixos",
					"timezone": "America/New_York",
				},
				"services": []interface{}{"ssh", "docker"},
				"packages": []interface{}{"git", "vim"},
			},
			expectError: false,
			checkFunc: func(req *ConfigRequest) bool {
				return req.Operation == "generate" &&
					req.ConfigType == "nixos" &&
					req.System == "x86_64-linux" &&
					req.HomeManager == true &&
					req.Flakes == true &&
					req.Validate == true &&
					len(req.Options) == 2 &&
					len(req.Services) == 2 &&
					len(req.Packages) == 2
			},
		},
		{
			name: "validate request",
			params: map[string]interface{}{
				"operation":      "validate",
				"config_path":    "/etc/nixos/configuration.nix",
				"config_content": "{ ... }",
			},
			expectError: false,
			checkFunc: func(req *ConfigRequest) bool {
				return req.Operation == "validate" &&
					req.ConfigPath == "/etc/nixos/configuration.nix" &&
					req.ConfigContent == "{ ... }"
			},
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"config_type": "nixos",
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
			req, err := cf.parseRequest(tt.params)
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

func TestConfigFunction_validateOperation(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name        string
		req         *ConfigRequest
		expectError bool
	}{
		{
			name: "valid generate operation",
			req: &ConfigRequest{
				Operation:  "generate",
				ConfigType: "nixos",
			},
			expectError: false,
		},
		{
			name: "valid validate operation with path",
			req: &ConfigRequest{
				Operation:  "validate",
				ConfigPath: "/etc/nixos/configuration.nix",
			},
			expectError: false,
		},
		{
			name: "valid validate operation with content",
			req: &ConfigRequest{
				Operation:     "validate",
				ConfigContent: "{ ... }",
			},
			expectError: false,
		},
		{
			name: "valid update operation",
			req: &ConfigRequest{
				Operation:  "update",
				ConfigPath: "/etc/nixos/configuration.nix",
			},
			expectError: false,
		},
		{
			name: "valid diff operation",
			req: &ConfigRequest{
				Operation:  "diff",
				ConfigPath: "/etc/nixos/configuration.nix",
			},
			expectError: false,
		},
		{
			name: "generate without config_type",
			req: &ConfigRequest{
				Operation: "generate",
			},
			expectError: true,
		},
		{
			name: "validate without path or content",
			req: &ConfigRequest{
				Operation: "validate",
			},
			expectError: true,
		},
		{
			name: "update without path",
			req: &ConfigRequest{
				Operation: "update",
			},
			expectError: true,
		},
		{
			name: "invalid operation",
			req: &ConfigRequest{
				Operation: "invalid",
			},
			expectError: true,
		},
		{
			name: "valid system architecture",
			req: &ConfigRequest{
				Operation:  "generate",
				ConfigType: "nixos",
				System:     "x86_64-linux",
			},
			expectError: false,
		},
		{
			name: "invalid system architecture",
			req: &ConfigRequest{
				Operation:  "generate",
				ConfigType: "nixos",
				System:     "invalid-system",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cf.validateOperation(tt.req)
			if (err != nil) != tt.expectError {
				t.Errorf("validateOperation() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestConfigFunction_parseAgentResponse(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name      string
		response  string
		req       *ConfigRequest
		checkFunc func(*ConfigResponse) bool
	}{
		{
			name:     "generate response with config",
			response: "Generated NixOS configuration:\n```nix\n{ config, pkgs, ... }:\n{\n  services.openssh.enable = true;\n  environment.systemPackages = with pkgs; [ git vim ];\n}\n```",
			req:      &ConfigRequest{Operation: "generate"},
			checkFunc: func(resp *ConfigResponse) bool {
				return resp.Status == "success" && resp.ConfigContent != ""
			},
		},
		{
			name: "validate response",
			response: `Configuration validation completed.
Valid: Configuration syntax is correct
Warning: Some deprecated options detected
Error: Missing required option services.xserver.enable`,
			req: &ConfigRequest{Operation: "validate"},
			checkFunc: func(resp *ConfigResponse) bool {
				return resp.ValidationResult != ""
			},
		},
		{
			name: "response with commands",
			response: `Configuration generated successfully. Apply with:
nixos-rebuild switch
home-manager switch`,
			req: &ConfigRequest{Operation: "generate"},
			checkFunc: func(resp *ConfigResponse) bool {
				return len(resp.SuggestedCommands) >= 1
			},
		},
		{
			name: "response with recommendations",
			response: `Configuration analysis complete.
Recommendations:
- Enable automatic garbage collection
- Use binary caches for faster builds
- Consider using flakes for reproducible builds`,
			req: &ConfigRequest{Operation: "analyze"},
			checkFunc: func(resp *ConfigResponse) bool {
				return len(resp.Recommendations) >= 2
			},
		},
		{
			name:     "backup response",
			response: `Configuration backed up successfully to /etc/nixos/configuration.nix.bak`,
			req:      &ConfigRequest{Operation: "backup"},
			checkFunc: func(resp *ConfigResponse) bool {
				return resp.BackupPath != ""
			},
		},
		{
			name: "diff response",
			response: `Configuration differences:
--- old/configuration.nix
+++ new/configuration.nix
@@ -1,3 +1,4 @@
 { config, pkgs, ... }:
 {
+  services.openssh.enable = true;
   environment.systemPackages = with pkgs; [ git ];
 }`,
			req: &ConfigRequest{Operation: "diff"},
			checkFunc: func(resp *ConfigResponse) bool {
				return resp.ConfigDiff != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := cf.parseAgentResponse(tt.response, tt.req)
			if resp == nil {
				t.Fatal("parseAgentResponse returned nil")
			}

			if tt.checkFunc != nil && !tt.checkFunc(resp) {
				t.Errorf("parseAgentResponse() returned response that failed validation check")
			}
		})
	}
}

func TestConfigFunction_extractValidationResult(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name     string
		response string
		expected bool // whether we expect meaningful validation result
	}{
		{
			name: "response with validation details",
			response: `Configuration validation completed.
Valid: Syntax is correct
Warning: Deprecated option detected
Error: Missing required option`,
			expected: true,
		},
		{
			name:     "response without validation details",
			response: "Configuration processed successfully.",
			expected: true, // Should return default message
		},
		{
			name:     "empty response",
			response: "",
			expected: true, // Should return default message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cf.extractValidationResult(tt.response)
			hasContent := result != ""
			if hasContent != tt.expected {
				t.Errorf("extractValidationResult() hasContent = %v, expected %v", hasContent, tt.expected)
			}
		})
	}
}

func TestConfigFunction_Execute(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name          string
		params        map[string]interface{}
		expectError   bool
		expectSuccess bool
	}{
		{
			name: "valid generate execution",
			params: map[string]interface{}{
				"operation":   "generate",
				"config_type": "nixos",
			},
			expectError:   false,
			expectSuccess: true,
		},
		{
			name: "valid validate execution",
			params: map[string]interface{}{
				"operation":   "validate",
				"config_path": "/etc/nixos/configuration.nix",
			},
			expectError:   false,
			expectSuccess: true,
		},
		{
			name: "invalid parameters",
			params: map[string]interface{}{
				"operation": "generate",
				// Missing config_type for generate operation
			},
			expectError:   false, // Should return error result, not execution error
			expectSuccess: false,
		},
		{
			name: "missing required operation",
			params: map[string]interface{}{
				"config_type": "nixos",
			},
			expectError:   false, // Should return error result, not execution error
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := cf.Execute(ctx, tt.params, nil)

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

func TestConfigFunction_Schema(t *testing.T) {
	cf := NewConfigFunction()
	schema := cf.Schema()

	if schema.Name != "config" {
		t.Errorf("Expected schema name 'config', got '%s'", schema.Name)
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
