package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestNewConfigureAgent(t *testing.T) {
	provider := &MockProvider{response: "test response"}
	agent := NewConfigureAgent(provider)

	if agent == nil {
		t.Fatal("Expected non-nil ConfigureAgent")
	}

	if agent.provider != provider {
		t.Error("Provider not set correctly")
	}

	if agent.context == nil {
		t.Fatal("Expected non-nil configuration context")
	}

	if agent.role != roles.RoleConfigure {
		t.Errorf("Expected role %s, got %s", roles.RoleConfigure, agent.role)
	}
}

func TestConfigureAgent_Query(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		setupContext  func(*ConfigureAgent)
		expectedError bool
		checkResponse func(string) bool
	}{
		{
			name:  "basic configuration query",
			input: "How do I set up a basic NixOS desktop system?",
			setupContext: func(agent *ConfigureAgent) {
				agent.SetConfigurationContext(&ConfigurationContext{
					DesktopEnvironment: "GNOME",
					Architecture:       "x86_64-linux",
					SecurityLevel:      "standard",
				})
			},
			expectedError: false,
			checkResponse: func(response string) bool {
				return strings.Contains(response, "Configuration Assistant") &&
					strings.Contains(response, "Safety Tips")
			},
		},
		{
			name:  "configuration with hardware context",
			input: "Configure my system for high-performance computing",
			setupContext: func(agent *ConfigureAgent) {
				agent.SetConfigurationContext(&ConfigurationContext{
					Hardware:           "NVIDIA GPU, 64GB RAM",
					PerformanceProfile: "high-performance",
					Services:           []string{"CUDA", "OpenMPI"},
				})
			},
			expectedError: false,
			checkResponse: func(response string) bool {
				return strings.Contains(response, "mock response")
			},
		},
		{
			name:  "server configuration query",
			input: "Set up a secure web server configuration",
			setupContext: func(agent *ConfigureAgent) {
				agent.SetConfigurationContext(&ConfigurationContext{
					Services:      []string{"nginx", "postgresql"},
					SecurityLevel: "high",
					NetworkConfig: "public",
				})
			},
			expectedError: false,
			checkResponse: func(response string) bool {
				return strings.Contains(response, "Configuration Assistant")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{
				response: "mock response for " + tt.input,
			}
			agent := NewConfigureAgent(provider)

			if tt.setupContext != nil {
				tt.setupContext(agent)
			}

			response, err := agent.Query(context.Background(), tt.input)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectedError && tt.checkResponse != nil && !tt.checkResponse(response) {
				t.Errorf("Response check failed for response: %s", response)
			}
		})
	}
}

func TestConfigureAgent_GenerateResponse(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "initial configuration generation",
			input: "Generate initial configuration for a development workstation",
		},
		{
			name:  "server configuration generation",
			input: "Create configuration for a web application server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{
				response: "Generated configuration for " + tt.input,
			}
			agent := NewConfigureAgent(provider)

			response, err := agent.GenerateResponse(context.Background(), tt.input)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !strings.Contains(response, "Configuration Assistant") {
				t.Error("Expected formatted response")
			}
		})
	}
}

func TestConfigureAgent_SetConfigurationContext(t *testing.T) {
	agent := NewConfigureAgent(&MockProvider{response: "test"})

	context := &ConfigurationContext{
		Hardware:           "ThinkPad X1 Carbon",
		DesktopEnvironment: "i3",
		SecurityLevel:      "high",
		Services:           []string{"docker", "postgresql"},
	}

	agent.SetConfigurationContext(context)

	retrieved := agent.GetConfigurationContext()
	if retrieved.Hardware != context.Hardware {
		t.Errorf("Expected hardware %s, got %s", context.Hardware, retrieved.Hardware)
	}

	if retrieved.DesktopEnvironment != context.DesktopEnvironment {
		t.Errorf("Expected desktop environment %s, got %s", context.DesktopEnvironment, retrieved.DesktopEnvironment)
	}
}

func TestConfigureAgent_GetConfigurationContext(t *testing.T) {
	agent := NewConfigureAgent(&MockProvider{response: "test"})

	// Should have empty context initially
	context := agent.GetConfigurationContext()
	if context == nil {
		t.Error("Expected non-nil context")
	}

	// Set a context and verify retrieval
	newContext := &ConfigurationContext{
		Architecture: "aarch64-linux",
		BootLoader:   "systemd-boot",
		FileSystem:   "btrfs",
	}

	agent.SetConfigurationContext(newContext)
	retrieved := agent.GetConfigurationContext()

	if retrieved.Architecture != newContext.Architecture {
		t.Errorf("Expected architecture %s, got %s", newContext.Architecture, retrieved.Architecture)
	}
}

func TestConfigureAgent_AnalyzeSystemRequirements(t *testing.T) {
	tests := []struct {
		name       string
		systemInfo string
		context    *ConfigurationContext
	}{
		{
			name:       "laptop system analysis",
			systemInfo: "Laptop: Intel i7-12700H, 32GB RAM, NVIDIA RTX 3070, 1TB NVMe",
			context: &ConfigurationContext{
				InstallationType: "desktop",
				SecurityLevel:    "standard",
			},
		},
		{
			name:       "server system analysis",
			systemInfo: "Server: AMD EPYC 7542, 128GB ECC RAM, 10GbE networking",
			context: &ConfigurationContext{
				InstallationType: "server",
				SecurityLevel:    "high",
				Services:         []string{"nginx", "postgresql", "redis"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{
				response: "System analysis for " + tt.systemInfo,
			}
			agent := NewConfigureAgent(mockProvider)

			if tt.context != nil {
				agent.SetConfigurationContext(tt.context)
			}

			response, err := agent.AnalyzeSystemRequirements(context.Background(), tt.systemInfo)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !strings.Contains(response, "Configuration Assistant") {
				t.Error("Expected formatted response")
			}
		})
	}
}

func TestConfigureAgent_GenerateInitialConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		requirements string
		context      *ConfigurationContext
	}{
		{
			name:         "desktop workstation configuration",
			requirements: "Desktop workstation for development with GNOME, Docker, and development tools",
			context: &ConfigurationContext{
				DesktopEnvironment: "GNOME",
				Services:           []string{"docker", "postgresql"},
				Users:              []string{"developer"},
			},
		},
		{
			name:         "minimal server configuration",
			requirements: "Minimal server for web hosting with nginx and SSL",
			context: &ConfigurationContext{
				Services:      []string{"nginx", "acme"},
				SecurityLevel: "high",
				NetworkConfig: "public",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{
				response: "Generated configuration for " + tt.requirements,
			}
			agent := NewConfigureAgent(mockProvider)

			if tt.context != nil {
				agent.SetConfigurationContext(tt.context)
			}

			response, err := agent.GenerateInitialConfiguration(context.Background(), tt.requirements)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !strings.Contains(response, "Configuration Assistant") {
				t.Error("Expected formatted response")
			}
		})
	}
}

func TestConfigureAgent_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		context       *ConfigurationContext
	}{
		{
			name: "basic configuration validation",
			configContent: `{
				boot.loader.systemd-boot.enable = true;
				services.openssh.enable = true;
				users.users.alice = {
					isNormalUser = true;
					extraGroups = [ "wheel" ];
				};
			}`,
			context: &ConfigurationContext{
				SecurityLevel: "standard",
			},
		},
		{
			name: "flake configuration validation",
			configContent: `{
				inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
				outputs = { self, nixpkgs }: {
					nixosConfigurations.hostname = nixpkgs.lib.nixosSystem {
						system = "x86_64-linux";
					};
				};
			}`,
			context: &ConfigurationContext{
				InstallationType: "flake-based",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{
				response: "Validation results for configuration",
			}
			agent := NewConfigureAgent(mockProvider)

			if tt.context != nil {
				agent.SetConfigurationContext(tt.context)
			}

			response, err := agent.ValidateConfiguration(context.Background(), tt.configContent)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !strings.Contains(response, "Configuration Assistant") {
				t.Error("Expected formatted response")
			}
		})
	}
}

func TestConfigureAgent_OptimizeConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		context       *ConfigurationContext
	}{
		{
			name: "performance optimization",
			configContent: `{
				services.xserver.enable = true;
				services.postgresql.enable = true;
			}`,
			context: &ConfigurationContext{
				PerformanceProfile: "high-performance",
				Hardware:           "High-end workstation",
			},
		},
		{
			name: "security optimization",
			configContent: `{
				services.openssh.enable = true;
				networking.firewall.enable = false;
			}`,
			context: &ConfigurationContext{
				SecurityLevel: "high",
				NetworkConfig: "public",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{
				response: "Optimization suggestions for configuration",
			}
			agent := NewConfigureAgent(mockProvider)

			if tt.context != nil {
				agent.SetConfigurationContext(tt.context)
			}

			response, err := agent.OptimizeConfiguration(context.Background(), tt.configContent)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !strings.Contains(response, "Configuration Assistant") {
				t.Error("Expected formatted response")
			}
		})
	}
}

func TestConfigureAgent_TroubleshootConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		issue   string
		context *ConfigurationContext
	}{
		{
			name:  "service startup failure",
			issue: "PostgreSQL service fails to start after configuration change",
			context: &ConfigurationContext{
				Services:          []string{"postgresql"},
				Issues:            []string{"postgresql startup failure"},
				ConfigurationFile: "/etc/nixos/configuration.nix",
			},
		},
		{
			name:  "boot loader configuration issue",
			issue: "System won't boot after enabling systemd-boot",
			context: &ConfigurationContext{
				BootLoader: "systemd-boot",
				Issues:     []string{"boot failure"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{
				response: "Troubleshooting guidance for " + tt.issue,
			}
			agent := NewConfigureAgent(mockProvider)

			if tt.context != nil {
				agent.SetConfigurationContext(tt.context)
			}

			response, err := agent.TroubleshootConfiguration(context.Background(), tt.issue)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !strings.Contains(response, "Configuration Assistant") {
				t.Error("Expected formatted response")
			}
		})
	}
}

func TestConfigureAgent_buildConfigurationPrompt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		context *ConfigurationContext
		checks  []string
	}{
		{
			name:  "basic prompt construction",
			input: "How do I configure wireless networking?",
			context: &ConfigurationContext{
				NetworkConfig: "wireless",
				SecurityLevel: "standard",
			},
			checks: []string{"User Query:", "How do I configure wireless networking?"},
		},
		{
			name:  "prompt with full context",
			input: "Optimize my server configuration",
			context: &ConfigurationContext{
				Hardware:           "Dell PowerEdge R750",
				Services:           []string{"nginx", "postgresql", "redis"},
				SecurityLevel:      "high",
				PerformanceProfile: "server",
			},
			checks: []string{"Configuration Context:", "Hardware:", "Services:", "Security Level:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewConfigureAgent(&MockProvider{response: "test"})

			if tt.context != nil {
				agent.SetConfigurationContext(tt.context)
			}

			prompt := agent.buildConfigurationPrompt(tt.input)

			for _, check := range tt.checks {
				if !strings.Contains(prompt, check) {
					t.Errorf("Expected prompt to contain '%s', but it didn't. Prompt: %s", check, prompt)
				}
			}
		})
	}
}

func TestConfigureAgent_formatConfigurationResponse(t *testing.T) {
	agent := NewConfigureAgent(&MockProvider{response: "test"})

	input := "This is a test response about NixOS configuration."
	formatted := agent.formatConfigurationResponse(input)

	expectedElements := []string{
		"## NixOS Configuration Assistant",
		input,
		"Configuration Safety Tips",
		"backup your current configuration",
		"nixos-rebuild test",
	}

	for _, element := range expectedElements {
		if !strings.Contains(formatted, element) {
			t.Errorf("Expected formatted response to contain '%s'", element)
		}
	}
}

func TestConfigureAgent_validateRole(t *testing.T) {
	// Test with valid role
	agent := NewConfigureAgent(&MockProvider{response: "test"})
	err := agent.validateRole()
	if err != nil {
		t.Errorf("Expected no error for valid role, got: %v", err)
	}

	// Test with empty role
	agent.role = ""
	err = agent.validateRole()
	if err == nil {
		t.Error("Expected error for empty role")
	}

	// Test with invalid role
	agent.role = "invalid-role"
	err = agent.validateRole()
	if err == nil {
		t.Error("Expected error for invalid role")
	}
}
