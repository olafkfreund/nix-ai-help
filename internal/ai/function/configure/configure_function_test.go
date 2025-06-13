package configure

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewConfigureFunction(t *testing.T) {
	cf := NewConfigureFunction()

	if cf == nil {
		t.Fatal("NewConfigureFunction returned nil")
	}

	if cf.Name() != "configure" {
		t.Errorf("Expected function name 'configure', got '%s'", cf.Name())
	}

	if cf.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestConfigureFunction_ValidateParameters(t *testing.T) {
	cf := NewConfigureFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"context": "Enable SSH service on NixOS",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"context":     "Configure firewall settings",
				"operation":   "add",
				"config_type": "nixos",
				"config_path": "/etc/nixos/configuration.nix",
				"module":      "networking.firewall",
				"option":      "enable",
				"value":       true,
				"dry_run":     true,
				"backup":      true,
				"validate":    true,
				"format":      "nix",
				"options": map[string]interface{}{
					"preserve_comments": "true",
				},
			},
			expectError: false,
		},
		{
			name: "Missing required context",
			params: map[string]interface{}{
				"operation": "add",
			},
			expectError: true,
		},
		{
			name: "Invalid operation",
			params: map[string]interface{}{
				"context":   "Configure service",
				"operation": "invalid-op",
			},
			expectError: true,
		},
		{
			name: "Invalid config_type",
			params: map[string]interface{}{
				"context":     "Configure service",
				"config_type": "invalid-type",
			},
			expectError: true,
		},
		{
			name: "Invalid format",
			params: map[string]interface{}{
				"context": "Configure service",
				"format":  "invalid-format",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cf.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestConfigureFunction_ParseRequest(t *testing.T) {
	cf := NewConfigureFunction()

	params := map[string]interface{}{
		"context":     "Enable SSH and configure firewall",
		"operation":   "update",
		"config_type": "nixos",
		"config_path": "/etc/nixos/configuration.nix",
		"module":      "services.openssh",
		"option":      "enable",
		"value":       true,
		"dry_run":     true,
		"backup":      true,
		"validate":    true,
		"format":      "nix",
		"options": map[string]interface{}{
			"preserve_comments": "true",
			"indent_size":       "2",
		},
	}

	request, err := cf.parseRequest(params)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if request.Context != "Enable SSH and configure firewall" {
		t.Errorf("Expected context 'Enable SSH and configure firewall', got '%s'", request.Context)
	}

	if request.Operation != "update" {
		t.Errorf("Expected operation 'update', got '%s'", request.Operation)
	}

	if request.ConfigType != "nixos" {
		t.Errorf("Expected config_type 'nixos', got '%s'", request.ConfigType)
	}

	if request.ConfigPath != "/etc/nixos/configuration.nix" {
		t.Errorf("Expected config_path '/etc/nixos/configuration.nix', got '%s'", request.ConfigPath)
	}

	if request.Module != "services.openssh" {
		t.Errorf("Expected module 'services.openssh', got '%s'", request.Module)
	}

	if request.Option != "enable" {
		t.Errorf("Expected option 'enable', got '%s'", request.Option)
	}

	if request.Value != true {
		t.Errorf("Expected value true, got %v", request.Value)
	}

	if !request.DryRun {
		t.Error("Expected dry_run to be true")
	}

	if !request.Backup {
		t.Error("Expected backup to be true")
	}

	if !request.Validate {
		t.Error("Expected validate to be true")
	}

	if request.Format != "nix" {
		t.Errorf("Expected format 'nix', got '%s'", request.Format)
	}

	if len(request.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(request.Options))
	}

	if request.Options["preserve_comments"] != "true" {
		t.Errorf("Expected preserve_comments 'true', got '%s'", request.Options["preserve_comments"])
	}
}

func TestConfigureFunction_DetermineOperation(t *testing.T) {
	cf := NewConfigureFunction()

	tests := []struct {
		name     string
		context  string
		expected string
	}{
		{
			name:     "Enable service context",
			context:  "Enable SSH service",
			expected: "add",
		},
		{
			name:     "Disable service context",
			context:  "Disable firewall",
			expected: "remove",
		},
		{
			name:     "Configure/update context",
			context:  "Configure nginx settings",
			expected: "update",
		},
		{
			name:     "Set value context",
			context:  "Set timezone to UTC",
			expected: "update",
		},
		{
			name:     "Remove option context",
			context:  "Remove old configuration",
			expected: "remove",
		},
		{
			name:     "Add new module context",
			context:  "Add docker support",
			expected: "add",
		},
		{
			name:     "Generic configuration context",
			context:  "Update system configuration",
			expected: "update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &ConfigureRequest{Context: tt.context}
			operation := cf.determineOperation(request)

			if operation != tt.expected {
				t.Errorf("Expected operation '%s', got '%s'", tt.expected, operation)
			}
		})
	}
}

func TestConfigureFunction_ValidateConfiguration(t *testing.T) {
	cf := NewConfigureFunction()

	tests := []struct {
		name          string
		configType    string
		configContent string
		expectValid   bool
		expectErrors  int
	}{
		{
			name:       "Valid NixOS configuration",
			configType: "nixos",
			configContent: `{
				services.openssh.enable = true;
				networking.firewall.enable = false;
			}`,
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name:       "Valid Home Manager configuration",
			configType: "home-manager",
			configContent: `{
				programs.git.enable = true;
				home.stateVersion = "23.05";
			}`,
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name:       "Invalid syntax",
			configType: "nixos",
			configContent: `{
				services.openssh.enable = true
				networking.firewall.enable = false;
			}`,
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name:          "Empty configuration",
			configType:    "nixos",
			configContent: "",
			expectValid:   false,
			expectErrors:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &ConfigureRequest{
				ConfigType: tt.configType,
			}

			validation := cf.validateConfiguration(request, tt.configContent)

			if validation.Valid != tt.expectValid {
				t.Errorf("Expected valid=%t, got valid=%t", tt.expectValid, validation.Valid)
			}

			if len(validation.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(validation.Errors))
			}

			if validation.Summary == "" {
				t.Error("Expected validation summary but got empty string")
			}
		})
	}
}

func TestConfigureFunction_GenerateBackupPath(t *testing.T) {
	cf := NewConfigureFunction()

	tests := []struct {
		name         string
		configPath   string
		expectBackup bool
	}{
		{
			name:         "NixOS configuration file",
			configPath:   "/etc/nixos/configuration.nix",
			expectBackup: true,
		},
		{
			name:         "Home Manager configuration",
			configPath:   "/home/user/.config/nixpkgs/home.nix",
			expectBackup: true,
		},
		{
			name:         "Flake file",
			configPath:   "/etc/nixos/flake.nix",
			expectBackup: true,
		},
		{
			name:         "Empty path",
			configPath:   "",
			expectBackup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &ConfigureRequest{
				ConfigPath: tt.configPath,
				Backup:     true,
			}

			backupPath := cf.generateBackupPath(request)

			if tt.expectBackup {
				if backupPath == "" {
					t.Error("Expected backup path but got empty string")
				}
				if backupPath == tt.configPath {
					t.Error("Backup path should be different from original path")
				}
			} else {
				if backupPath != "" {
					t.Errorf("Expected empty backup path but got '%s'", backupPath)
				}
			}
		})
	}
}

func TestConfigureFunction_ApplyConfigurationChange(t *testing.T) {
	cf := NewConfigureFunction()

	tests := []struct {
		name        string
		operation   string
		option      string
		value       interface{}
		expectError bool
	}{
		{
			name:        "Add boolean option",
			operation:   "add",
			option:      "services.openssh.enable",
			value:       true,
			expectError: false,
		},
		{
			name:        "Update string option",
			operation:   "update",
			option:      "time.timeZone",
			value:       "Europe/London",
			expectError: false,
		},
		{
			name:        "Remove option",
			operation:   "remove",
			option:      "services.nginx.enable",
			value:       nil,
			expectError: false,
		},
		{
			name:        "Invalid operation",
			operation:   "invalid",
			option:      "some.option",
			value:       "value",
			expectError: true,
		},
		{
			name:        "Empty option path",
			operation:   "add",
			option:      "",
			value:       true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &ConfigureRequest{
				Operation: tt.operation,
				Option:    tt.option,
				Value:     tt.value,
				DryRun:    true, // Always use dry run for tests
			}

			change, err := cf.applyConfigurationChange(request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if change == nil {
					t.Error("Expected configuration change but got nil")
				}
				if change != nil {
					if change.Option != tt.option {
						t.Errorf("Expected option '%s', got '%s'", tt.option, change.Option)
					}
					if change.Operation != tt.operation {
						t.Errorf("Expected operation '%s', got '%s'", tt.operation, change.Operation)
					}
				}
			}
		})
	}
}

func TestConfigureFunction_ExecuteWithMockAgent(t *testing.T) {
	cf := NewConfigureFunction()

	params := map[string]interface{}{
		"context":   "Enable SSH service",
		"operation": "add",
		"module":    "services.openssh",
		"option":    "enable",
		"value":     true,
		"dry_run":   true,
	}

	options := &functionbase.FunctionOptions{}

	result, err := cf.Execute(context.Background(), params, options)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got failure: %s", result.Error)
	}

	if result.Data == nil {
		t.Error("Expected result data but got nil")
	}
}
