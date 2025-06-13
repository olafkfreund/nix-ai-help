package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestFlakeAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "flake agent response"}
	agent := NewFlakeAgent(mockProvider)

	input := "How do I update my flake inputs?"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "flake agent")
}

func TestFlakeAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "flake agent response"}
	agent := NewFlakeAgent(mockProvider)

	input := "Help me debug this flake error"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "flake agent response")
}

func TestFlakeAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewFlakeAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RoleFlake)
	require.NoError(t, err)
	require.Equal(t, roles.RoleFlake, agent.role)

	// Test setting context
	flakeCtx := &FlakeContext{ProjectType: "nixos"}
	agent.SetContext(flakeCtx)
	require.Equal(t, flakeCtx, agent.contextData)
}

func TestFlakeAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewFlakeAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestFlakeContext_Formatting(t *testing.T) {
	flakeCtx := &FlakeContext{
		FlakePath:     "/home/user/project",
		FlakeNix:      "{ inputs = { nixpkgs.url = \"github:NixOS/nixpkgs\"; }; }",
		FlakeLock:     "{ \"nodes\": { \"nixpkgs\": { \"locked\": {} } } }",
		FlakeInputs:   map[string]string{"nixpkgs": "github:NixOS/nixpkgs/nixos-23.11", "home-manager": "github:nix-community/home-manager"},
		FlakeOutputs:  []string{"packages.x86_64-linux.default", "devShells.x86_64-linux.default"},
		FlakeMetadata: "path:/home/user/project?lastModified=1234567890",
		FlakeErrors:   []string{"error: flake output 'packages' is not a function", "warning: Git tree '/home/user/project' is dirty"},
		FlakeCommands: []string{"nix flake update", "nix build", "nix develop"},
		ProjectType:   "home-manager",
		FlakeSystem:   "x86_64-linux",
		Dependencies:  []string{"nixpkgs", "home-manager", "neovim-flake"},
		BuildOutputs:  "building '/nix/store/abc123-source'...",
	}

	// Test that context can be created and has expected fields
	require.NotEmpty(t, flakeCtx.FlakePath)
	require.NotEmpty(t, flakeCtx.FlakeNix)
	require.NotEmpty(t, flakeCtx.FlakeLock)
	require.Len(t, flakeCtx.FlakeInputs, 2)
	require.Len(t, flakeCtx.FlakeOutputs, 2)
	require.Len(t, flakeCtx.FlakeErrors, 2)
	require.Len(t, flakeCtx.FlakeCommands, 3)
	require.Equal(t, "home-manager", flakeCtx.ProjectType)
	require.Equal(t, "x86_64-linux", flakeCtx.FlakeSystem)
	require.Len(t, flakeCtx.Dependencies, 3)
}
