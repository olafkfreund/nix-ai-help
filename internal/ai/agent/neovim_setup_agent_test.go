package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNeovimSetupAgent(t *testing.T) {
	provider := &MockProvider{response: "mock response"}
	agent := NewNeovimSetupAgent(provider)

	assert.NotNil(t, agent)

	// Check that context is properly initialized
	assert.NotNil(t, agent.context)
	assert.Equal(t, "nixvim", agent.context.ConfigType)
	assert.Equal(t, "nixvim", agent.context.PluginManager)
}

func TestNeovimSetupAgent_SetupNeovimConfig(t *testing.T) {
	provider := &MockProvider{response: "Neovim configuration setup complete"}
	agent := NewNeovimSetupAgent(provider)

	result, err := agent.SetupNeovimConfig(context.Background(), "nixvim", "go,python")
	require.NoError(t, err)
	assert.Contains(t, result, "configuration setup complete")
	assert.Equal(t, "nixvim", agent.context.ConfigType)
}

func TestNeovimSetupAgent_ConfigureLSP(t *testing.T) {
	provider := &MockProvider{response: "LSP configuration complete"}
	agent := NewNeovimSetupAgent(provider)

	result, err := agent.ConfigureLSP(context.Background(), []string{"go", "python"})
	require.NoError(t, err)
	assert.Contains(t, result, "LSP configuration complete")
	assert.Equal(t, []string{"go", "python"}, agent.context.Languages)
}

func TestNeovimSetupAgent_OptimizePerformance(t *testing.T) {
	provider := &MockProvider{response: "Performance optimization recommendations generated"}
	agent := NewNeovimSetupAgent(provider)

	result, err := agent.OptimizePerformance(context.Background(), 500, []string{"startup-time", "memory-usage"})
	require.NoError(t, err)
	assert.Contains(t, result, "optimization recommendations")
	assert.Equal(t, 500, agent.context.StartupTime)
	assert.Equal(t, []string{"startup-time", "memory-usage"}, agent.context.PerformanceGoals)
}

func TestNeovimSetupAgent_MigrateConfiguration(t *testing.T) {
	provider := &MockProvider{response: "Migration configuration complete"}
	agent := NewNeovimSetupAgent(provider)

	result, err := agent.MigrateConfiguration(context.Background(), "VSCode", "/path/to/vscode/config")
	require.NoError(t, err)
	assert.Contains(t, result, "Migration configuration complete")
}

func TestNeovimSetupAgent_TroubleshootConfig(t *testing.T) {
	provider := &MockProvider{response: "Troubleshooting recommendations"}
	agent := NewNeovimSetupAgent(provider)

	issue := "plugin_not_loading"
	errorMessage := "Plugin 'telescope' not found"
	result, err := agent.TroubleshootConfig(context.Background(), issue, errorMessage)
	require.NoError(t, err)
	assert.Contains(t, result, "recommendations")
}

func TestNeovimSetupAgent_SetContext(t *testing.T) {
	provider := &MockProvider{response: "mock response"}
	agent := NewNeovimSetupAgent(provider)

	neovimCtx := &NeovimSetupContext{
		ConfigType:    "home-manager",
		PluginManager: "lazy",
		Languages:     []string{"rust", "go"},
	}

	err := agent.SetContext(neovimCtx)
	require.NoError(t, err)
	assert.Equal(t, "home-manager", agent.context.ConfigType)
	assert.Equal(t, "lazy", agent.context.PluginManager)
}

func TestNeovimSetupAgent_SetContext_InvalidType(t *testing.T) {
	provider := &MockProvider{response: "mock response"}
	agent := NewNeovimSetupAgent(provider)

	err := agent.SetContext("invalid context")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *NeovimSetupContext")
}

func TestNeovimSetupAgent_formatContext(t *testing.T) {
	provider := &MockProvider{response: "mock response"}
	agent := NewNeovimSetupAgent(provider)

	agent.context.ConfigType = "nixvim"
	agent.context.Languages = []string{"rust", "go"}
	agent.context.WorkflowType = "systems-programming"

	formatted := agent.formatContext()
	assert.Contains(t, formatted, "Configuration Type: nixvim")
	assert.Contains(t, formatted, "Languages: rust, go")
	assert.Contains(t, formatted, "Workflow Type: systems-programming")
}

func TestNeovimSetupAgent_ContextInitialization(t *testing.T) {
	provider := &MockProvider{response: "mock response"}
	agent := NewNeovimSetupAgent(provider)

	// Verify all context slices are initialized
	assert.NotNil(t, agent.context.Languages)
	assert.NotNil(t, agent.context.LSPServers)
	assert.NotNil(t, agent.context.Formatters)
	assert.NotNil(t, agent.context.Linters)
	assert.NotNil(t, agent.context.UIPreferences)
	assert.NotNil(t, agent.context.PerformanceGoals)
	assert.NotNil(t, agent.context.ResourceLimits)
}
