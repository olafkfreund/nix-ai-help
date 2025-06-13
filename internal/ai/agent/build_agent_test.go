package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestBuildAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "build agent response"}
	agent := NewBuildAgent(mockProvider)

	input := "Why is my NixOS build failing?"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "build agent")
}

func TestBuildAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "build agent response"}
	agent := NewBuildAgent(mockProvider)

	input := "Help me fix this build error"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "build agent response")
}

func TestBuildAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewBuildAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RoleBuild)
	require.NoError(t, err)
	require.Equal(t, roles.RoleBuild, agent.role)

	// Test setting context
	buildCtx := &BuildContext{BuildSystem: "nixos-rebuild"}
	agent.SetContext(buildCtx)
	require.Equal(t, buildCtx, agent.contextData)
}

func TestBuildAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewBuildAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestBuildContext_Formatting(t *testing.T) {
	buildCtx := &BuildContext{
		BuildOutput:    "building derivation",
		ErrorLogs:      "error occurred",
		ConfigPath:     "/etc/nixos/configuration.nix",
		DerivationPath: "/nix/store/abc123.drv",
		FailedPackages: []string{"pkg1", "pkg2"},
		BuildSystem:    "nixos-rebuild",
		Architecture:   "x86_64-linux",
		NixChannels:    []string{"nixos-23.11", "nixpkgs-unstable"},
		SystemInfo:     "NixOS 23.11",
	}

	// Test that context can be created and has expected fields
	require.NotEmpty(t, buildCtx.BuildOutput)
	require.NotEmpty(t, buildCtx.ErrorLogs)
	require.NotEmpty(t, buildCtx.ConfigPath)
	require.Len(t, buildCtx.FailedPackages, 2)
	require.Equal(t, "nixos-rebuild", buildCtx.BuildSystem)
}
