package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestSearchAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "search agent response"}
	agent := NewSearchAgent(mockProvider)

	searchCtx := &SearchContext{
		SearchQuery:    "firefox browser",
		SearchType:     "packages",
		SearchResults:  []string{"firefox", "firefox-esr", "firefox-bin"},
		SearchSources:  []string{"nixpkgs", "wiki"},
		ChannelVersion: "nixos-23.11",
		SystemArch:     "x86_64-linux",
		PackageFilters: []string{"license:mpl", "maintainer:mozilla"},
		SearchLimit:    10,
		SortBy:         "relevance",
		IncludeUnfree:  false,
		SearchHistory:  []string{"chromium", "browser", "web"},
		RelatedQueries: []string{"web browser", "mozilla firefox", "browser security"},
		DocSections:    []string{"packages.firefox", "programs.firefox"},
	}
	agent.SetContext(searchCtx)

	input := "Search for Firefox browser packages"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "search agent")
}

func TestSearchAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "search agent response"}
	agent := NewSearchAgent(mockProvider)

	searchCtx := &SearchContext{
		SearchQuery:    "nixos configuration options",
		SearchType:     "options",
		SearchResults:  []string{"services.openssh.enable", "networking.firewall.enable", "boot.loader.systemd-boot.enable"},
		SearchSources:  []string{"nixos-options", "manual"},
		ChannelVersion: "nixos-unstable",
		SystemArch:     "x86_64-linux",
		SearchLimit:    20,
		SortBy:         "name",
		MCPResults:     "Found 156 configuration options matching your query",
		DocSections:    []string{"configuration.nix", "module-system", "options-reference"},
		RelatedQueries: []string{"systemd configuration", "network settings", "boot options"},
	}
	agent.SetContext(searchCtx)

	input := "Find configuration options for system services"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "search agent response")
}

func TestSearchAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewSearchAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RoleSearch)
	require.NoError(t, err)
	require.Equal(t, roles.RoleSearch, agent.role)

	// Test setting context
	searchCtx := &SearchContext{SearchType: "packages"}
	agent.SetContext(searchCtx)
	require.Equal(t, searchCtx, agent.contextData)
}

func TestSearchAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewSearchAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestSearchContext_Formatting(t *testing.T) {
	searchCtx := &SearchContext{
		SearchQuery:    "development tools programming",
		SearchType:     "packages",
		SearchResults:  []string{"gcc", "clang", "rustc", "nodejs", "python3", "go"},
		SearchSources:  []string{"nixpkgs", "flakes", "community"},
		ChannelVersion: "nixos-unstable",
		SystemArch:     "x86_64-linux",
		PackageFilters: []string{"license:free", "category:development"},
		SearchLimit:    50,
		SortBy:         "popularity",
		IncludeUnfree:  true,
		SearchHistory:  []string{"compilers", "interpreters", "build tools", "IDE"},
		RelatedQueries: []string{"programming languages", "build systems", "development environment"},
		MCPResults:     "Found 342 development packages across multiple categories",
		DocSections:    []string{"development", "programming", "languages", "tools"},
	}

	// Test that context can be created and has expected fields
	require.NotEmpty(t, searchCtx.SearchQuery)
	require.Equal(t, "packages", searchCtx.SearchType)
	require.Len(t, searchCtx.SearchResults, 6)
	require.Len(t, searchCtx.SearchSources, 3)
	require.Equal(t, "nixos-unstable", searchCtx.ChannelVersion)
	require.Equal(t, "x86_64-linux", searchCtx.SystemArch)
	require.Len(t, searchCtx.PackageFilters, 2)
	require.Equal(t, 50, searchCtx.SearchLimit)
	require.Equal(t, "popularity", searchCtx.SortBy)
	require.True(t, searchCtx.IncludeUnfree)
	require.Len(t, searchCtx.SearchHistory, 4)
	require.Len(t, searchCtx.RelatedQueries, 3)
	require.NotEmpty(t, searchCtx.MCPResults)
	require.Len(t, searchCtx.DocSections, 4)
}
