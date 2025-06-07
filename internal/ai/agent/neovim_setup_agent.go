package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// NeovimSetupAgent handles Neovim configuration and setup operations.
type NeovimSetupAgent struct {
	BaseAgent
	context *NeovimSetupContext
}

// NeovimSetupContext contains information about the Neovim setup environment.
type NeovimSetupContext struct {
	// Configuration management
	ConfigType     string `json:"config_type"`     // nixvim, home-manager, traditional
	ConfigLocation string `json:"config_location"` // Path to configuration files
	NeovimVersion  string `json:"neovim_version"`  // Neovim version to use
	ExistingConfig bool   `json:"existing_config"` // Whether user has existing config

	// Plugin management
	PluginManager    string   `json:"plugin_manager"`    // Plugin management system (nixvim, lazy, packer)
	InstalledPlugins []string `json:"installed_plugins"` // Currently installed plugins
	RequiredPlugins  []string `json:"required_plugins"`  // Plugins needed for workflow
	PluginConflicts  []string `json:"plugin_conflicts"`  // Known plugin conflicts

	// Development environment
	Languages  []string `json:"languages"`   // Programming languages used
	LSPServers []string `json:"lsp_servers"` // Required LSP servers
	Formatters []string `json:"formatters"`  // Code formatters needed
	Linters    []string `json:"linters"`     // Linters to configure

	// Workflow preferences
	Keybindings   string   `json:"keybindings"`    // Keybinding style (vim, emacs, vscode)
	Theme         string   `json:"theme"`          // Color scheme preference
	UIPreferences []string `json:"ui_preferences"` // UI customization preferences
	WorkflowType  string   `json:"workflow_type"`  // Development workflow type

	// System integration
	SystemOS         string `json:"system_os"`         // Operating system
	TerminalEmulator string `json:"terminal_emulator"` // Terminal being used
	ShellType        string `json:"shell_type"`        // Shell environment
	NixConfiguration bool   `json:"nix_configuration"` // Whether using Nix for config

	// Performance settings
	StartupTime      int      `json:"startup_time"`      // Current startup time (ms)
	PerformanceGoals []string `json:"performance_goals"` // Performance optimization goals
	ResourceLimits   []string `json:"resource_limits"`   // System resource constraints
}

// NewNeovimSetupAgent creates a new NeovimSetupAgent.
func NewNeovimSetupAgent(provider ai.Provider) *NeovimSetupAgent {
	return &NeovimSetupAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleNeovimSetup,
		},
		context: &NeovimSetupContext{
			ConfigType:       "nixvim",
			PluginManager:    "nixvim",
			Languages:        []string{},
			LSPServers:       []string{},
			Formatters:       []string{},
			Linters:          []string{},
			UIPreferences:    []string{},
			PerformanceGoals: []string{},
			ResourceLimits:   []string{},
		},
	}
}

// SetupNeovimConfig helps set up a new Neovim configuration.
func (a *NeovimSetupAgent) SetupNeovimConfig(ctx context.Context, configType, targetLanguages string) (string, error) {
	a.context.ConfigType = configType

	prompt := a.buildSetupPrompt(configType, targetLanguages)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate Neovim setup: %w", err)
	}

	return response, nil
}

// ConfigureLSP sets up Language Server Protocol configuration.
func (a *NeovimSetupAgent) ConfigureLSP(ctx context.Context, languages []string) (string, error) {
	a.context.Languages = languages

	prompt := a.buildLSPPrompt(languages)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to configure LSP: %w", err)
	}

	return response, nil
}

// OptimizePerformance provides performance optimization recommendations.
func (a *NeovimSetupAgent) OptimizePerformance(ctx context.Context, currentStartupTime int, goals []string) (string, error) {
	a.context.StartupTime = currentStartupTime
	a.context.PerformanceGoals = goals

	prompt := a.buildPerformancePrompt(currentStartupTime, goals)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to optimize performance: %w", err)
	}

	return response, nil
}

// MigrateConfiguration helps migrate from existing editor configurations.
func (a *NeovimSetupAgent) MigrateConfiguration(ctx context.Context, sourceEditor, configPath string) (string, error) {
	prompt := a.buildMigrationPrompt(sourceEditor, configPath)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate migration guide: %w", err)
	}

	return response, nil
}

// CustomizeWorkflow provides workflow-specific customization advice.
func (a *NeovimSetupAgent) CustomizeWorkflow(ctx context.Context, workflowType string, requirements []string) (string, error) {
	a.context.WorkflowType = workflowType

	prompt := a.buildWorkflowPrompt(workflowType, requirements)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to customize workflow: %w", err)
	}

	return response, nil
}

// TroubleshootConfig helps diagnose and fix Neovim configuration issues.
func (a *NeovimSetupAgent) TroubleshootConfig(ctx context.Context, issue, errorMessage string) (string, error) {
	prompt := a.buildTroubleshootingPrompt(issue, errorMessage)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to troubleshoot configuration: %w", err)
	}

	return response, nil
}

// Helper methods for building prompts

func (a *NeovimSetupAgent) buildSetupPrompt(configType, targetLanguages string) string {
	return fmt.Sprintf(`Help set up a Neovim configuration with the following requirements:

Configuration Type: %s
Target Languages: %s
Current Context: %s

Please provide:
1. Complete configuration structure and files needed
2. Plugin recommendations for the specified languages
3. LSP server setup instructions
4. Key bindings and UI customizations
5. Installation and setup steps
6. Basic usage examples

Focus on creating a maintainable, performant configuration that follows best practices.`,
		configType, targetLanguages, a.formatContext())
}

func (a *NeovimSetupAgent) buildLSPPrompt(languages []string) string {
	return fmt.Sprintf(`Configure Language Server Protocol support for: %v

Current Context: %s

Please provide:
1. Required LSP servers for each language
2. Installation instructions (preferably through Nix)
3. Configuration examples for each server
4. Formatter and linter integration
5. Debugging setup where applicable
6. Troubleshooting common LSP issues

Ensure configurations are optimized for performance and developer experience.`,
		languages, a.formatContext())
}

func (a *NeovimSetupAgent) buildPerformancePrompt(startupTime int, goals []string) string {
	return fmt.Sprintf(`Optimize Neovim performance with current startup time: %dms

Performance Goals: %v
Current Context: %s

Please provide:
1. Startup time optimization strategies
2. Plugin loading optimization
3. Memory usage reduction techniques
4. Configuration cleanup recommendations
5. Benchmarking and measurement tools
6. Specific configuration changes to implement

Focus on maintaining functionality while improving performance metrics.`,
		startupTime, goals, a.formatContext())
}

func (a *NeovimSetupAgent) buildMigrationPrompt(sourceEditor, configPath string) string {
	return fmt.Sprintf(`Help migrate from %s to Neovim configuration.

Source Configuration: %s
Current Context: %s

Please provide:
1. Migration strategy and timeline
2. Configuration mapping and equivalents
3. Plugin alternatives and replacements
4. Workflow adaptation recommendations
5. Step-by-step migration process
6. Backup and rollback procedures

Ensure a smooth transition with minimal productivity loss.`,
		sourceEditor, configPath, a.formatContext())
}

func (a *NeovimSetupAgent) buildWorkflowPrompt(workflowType string, requirements []string) string {
	return fmt.Sprintf(`Customize Neovim for %s workflow.

Requirements: %v
Current Context: %s

Please provide:
1. Workflow-specific plugin recommendations
2. Custom key bindings for common tasks
3. Automation and macro suggestions
4. Integration with external tools
5. Productivity enhancement features
6. Configuration examples

Focus on optimizing the editing experience for this specific workflow.`,
		workflowType, requirements, a.formatContext())
}

func (a *NeovimSetupAgent) buildTroubleshootingPrompt(issue, errorMessage string) string {
	return fmt.Sprintf(`Troubleshoot Neovim configuration issue:

Issue: %s
Error Message: %s
Current Context: %s

Please provide:
1. Root cause analysis of the issue
2. Step-by-step debugging process
3. Specific fixes and configuration changes
4. Prevention strategies for similar issues
5. Alternative approaches if needed
6. Testing and validation steps

Focus on providing clear, actionable solutions.`,
		issue, errorMessage, a.formatContext())
}

// GetContext returns the current agent context.
func (a *NeovimSetupAgent) GetContext() interface{} {
	return a.context
}

// SetContext updates the agent context.
func (a *NeovimSetupAgent) SetContext(ctx interface{}) error {
	if neovimCtx, ok := ctx.(*NeovimSetupContext); ok {
		a.context = neovimCtx
		return nil
	}
	return fmt.Errorf("invalid context type for NeovimSetupAgent")
}

// formatContext returns a formatted string representation of the current context.
func (a *NeovimSetupAgent) formatContext() string {
	return fmt.Sprintf(`Configuration Type: %s
Plugin Manager: %s
Languages: %v
LSP Servers: %v
Workflow Type: %s
System: %s
Performance Goals: %v`,
		a.context.ConfigType,
		a.context.PluginManager,
		a.context.Languages,
		a.context.LSPServers,
		a.context.WorkflowType,
		a.context.SystemOS,
		a.context.PerformanceGoals)
}
