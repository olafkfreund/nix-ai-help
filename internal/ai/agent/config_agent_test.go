package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestConfigAgent_NewConfigAgent(t *testing.T) {
	agent := NewConfigAgent()

	if agent == nil {
		t.Fatal("NewConfigAgent returned nil")
	}

	if agent.role != string(roles.RoleConfig) {
		t.Errorf("Role = %v, expected %v", agent.role, string(roles.RoleConfig))
	}
}

func TestConfigAgent_AnalyzeConfiguration(t *testing.T) {
	agent := NewConfigAgent()

	configPath := "/etc/nixos/configuration.nix"
	configContent := "{ config, pkgs, ... }: { services.openssh.enable = true; }"

	result, err := agent.AnalyzeConfiguration(context.Background(), configPath, configContent)

	if err != nil {
		t.Errorf("AnalyzeConfiguration failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
	if !strings.Contains(result, configPath) {
		t.Errorf("Result should contain config path %s", configPath)
	}
}

func TestConfigAgent_ReviewConfiguration(t *testing.T) {
	agent := NewConfigAgent()

	configPath := "/etc/nixos/configuration.nix"
	configContent := "{ config, pkgs, ... }: { }"
	reviewType := "security"

	result, err := agent.ReviewConfiguration(context.Background(), configPath, configContent, reviewType)

	if err != nil {
		t.Errorf("ReviewConfiguration failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
	if !strings.Contains(result, reviewType) {
		t.Errorf("Result should contain review type %s", reviewType)
	}
}

func TestConfigAgent_SuggestImprovements(t *testing.T) {
	agent := NewConfigAgent()

	configPath := "/etc/nixos/configuration.nix"
	configContent := "{ config, pkgs, ... }: { }"
	focusAreas := []string{"security", "performance"}

	result, err := agent.SuggestImprovements(context.Background(), configPath, configContent, focusAreas)

	if err != nil {
		t.Errorf("SuggestImprovements failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
	for _, area := range focusAreas {
		if !strings.Contains(result, area) {
			t.Errorf("Result should contain focus area %s", area)
		}
	}
}

func TestConfigAgent_ValidateConfiguration(t *testing.T) {
	agent := NewConfigAgent()

	configPath := "/etc/nixos/configuration.nix"
	configContent := "{ config, pkgs, ... }: { }"

	result, err := agent.ValidateConfiguration(context.Background(), configPath, configContent)

	if err != nil {
		t.Errorf("ValidateConfiguration failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
	if !strings.Contains(result, "syntax") {
		t.Error("Result should mention syntax validation")
	}
}

func TestConfigAgent_SetRole(t *testing.T) {
	agent := NewConfigAgent()
	newRole := "custom-role"

	agent.SetRole(newRole)

	if agent.role != newRole {
		t.Errorf("Role = %s, expected %s", agent.role, newRole)
	}
}

func TestConfigAgent_SetContext(t *testing.T) {
	agent := NewConfigAgent()
	context := map[string]string{"key": "value"}

	agent.SetContext(context)

	if agent.contextData == nil {
		t.Error("Context data should not be nil")
	}
}
