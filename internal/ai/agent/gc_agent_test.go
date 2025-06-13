package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestGCAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "gc agent response"}
	agent := NewGCAgent(mockProvider)

	gcCtx := &GCContext{
		StoreSize:        "15.2 GB",
		StoreUsage:       "62%",
		GenerationCount:  25,
		OldestGeneration: "2024-11-15",
		LastCleanup:      "2024-12-28",
		LargeItems:       []string{"/nix/store/abc123-large-package", "/nix/store/def456-another-large"},
		UnusedItems:      []string{"/nix/store/ghi789-unused-package"},
		RootsCount:       12,
		GCRoots:          []string{"/nix/var/nix/gcroots/auto/1", "/nix/var/nix/gcroots/auto/2"},
		DiskSpace:        "8.5 GB available",
		AutoGCEnabled:    true,
		GCSchedule:       "weekly",
		DryRun:           true,
	}
	agent.SetContext(gcCtx)

	input := "How can I free up space in my Nix store?"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "gc agent")
}

func TestGCAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "gc agent response"}
	agent := NewGCAgent(mockProvider)

	gcCtx := &GCContext{
		StoreSize:       "22.8 GB",
		StoreUsage:      "85%",
		GenerationCount: 40,
		DiskSpace:       "3.2 GB available",
		KeepOutputs:     false,
		KeepDerivations: true,
		MaxFreed:        "10 GB",
		MinAge:          "30 days",
	}
	agent.SetContext(gcCtx)

	input := "My Nix store is nearly full, what cleanup strategy should I use?"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "gc agent response")
}

func TestGCAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewGCAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RoleGC)
	require.NoError(t, err)
	require.Equal(t, roles.RoleGC, agent.role)

	// Test setting context
	gcCtx := &GCContext{StoreSize: "10 GB"}
	agent.SetContext(gcCtx)
	require.Equal(t, gcCtx, agent.contextData)
}

func TestGCAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewGCAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestGCContext_Formatting(t *testing.T) {
	gcCtx := &GCContext{
		StoreSize:        "18.7 GB",
		StoreUsage:       "73%",
		GenerationCount:  30,
		OldestGeneration: "2024-10-01",
		LastCleanup:      "2024-12-20",
		LargeItems:       []string{"/nix/store/abc123-chromium", "/nix/store/def456-llvm", "/nix/store/ghi789-gcc"},
		UnusedItems:      []string{"/nix/store/jkl012-old-kernel", "/nix/store/mno345-unused-lib"},
		RootsCount:       15,
		GCRoots:          []string{"/nix/var/nix/gcroots/auto/1", "/nix/var/nix/gcroots/profiles", "/nix/var/nix/gcroots/booted-system"},
		DiskSpace:        "6.8 GB available",
		CleanupOptions:   []string{"--delete-older-than 30d", "--delete-generations +5", "--max-freed 5G"},
		AutoGCEnabled:    true,
		GCSchedule:       "daily",
		StorePaths:       []string{"/nix/store/abc123-target", "/nix/store/def456-another"},
		DryRun:           false,
		KeepOutputs:      true,
		KeepDerivations:  false,
		MaxFreed:         "8 GB",
		MinAge:           "14 days",
		SystemProfile:    "/nix/var/nix/profiles/system",
		UserProfiles:     []string{"/nix/var/nix/profiles/per-user/user1", "/nix/var/nix/profiles/per-user/user2"},
	}

	// Test that context can be created and has expected fields
	require.NotEmpty(t, gcCtx.StoreSize)
	require.Equal(t, "73%", gcCtx.StoreUsage)
	require.Equal(t, 30, gcCtx.GenerationCount)
	require.Equal(t, "2024-10-01", gcCtx.OldestGeneration)
	require.Equal(t, "2024-12-20", gcCtx.LastCleanup)
	require.Len(t, gcCtx.LargeItems, 3)
	require.Len(t, gcCtx.UnusedItems, 2)
	require.Equal(t, 15, gcCtx.RootsCount)
	require.Len(t, gcCtx.GCRoots, 3)
	require.NotEmpty(t, gcCtx.DiskSpace)
	require.Len(t, gcCtx.CleanupOptions, 3)
	require.True(t, gcCtx.AutoGCEnabled)
	require.Equal(t, "daily", gcCtx.GCSchedule)
	require.Len(t, gcCtx.StorePaths, 2)
	require.False(t, gcCtx.DryRun)
	require.True(t, gcCtx.KeepOutputs)
	require.False(t, gcCtx.KeepDerivations)
	require.Equal(t, "8 GB", gcCtx.MaxFreed)
	require.Equal(t, "14 days", gcCtx.MinAge)
	require.NotEmpty(t, gcCtx.SystemProfile)
	require.Len(t, gcCtx.UserProfiles, 2)
}
