package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// CompletionContext provides context for shell completion generation and management.
type CompletionContext struct {
	// Shell information
	ShellType    string `json:"shell_type"`    // bash, zsh, fish, powershell
	ShellVersion string `json:"shell_version"` // Shell version information
	ShellConfig  string `json:"shell_config"`  // Current shell configuration

	// Completion system information
	CompletionType      string   `json:"completion_type"`      // bash-completion, zsh-completions, fish-completions
	CompletionPath      string   `json:"completion_path"`      // Installation path for completions
	ExistingCompletions []string `json:"existing_completions"` // Currently installed completions

	// Command information
	CommandName      string            `json:"command_name"`      // Command to generate completions for
	Subcommands      []string          `json:"subcommands"`       // Available subcommands
	Flags            []string          `json:"flags"`             // Available flags and options
	FlagDescriptions map[string]string `json:"flag_descriptions"` // Flag descriptions

	// Context-aware completion
	CompletionScope  string   `json:"completion_scope"`  // global, project, user
	CustomFunctions  []string `json:"custom_functions"`  // Custom completion functions
	CompletionScript string   `json:"completion_script"` // Generated or existing completion script

	// Integration information
	PackageManager  string `json:"package_manager"`  // nix, homebrew, apt, etc.
	InstallMethod   string `json:"install_method"`   // how completions are installed
	UpdateMechanism string `json:"update_mechanism"` // how completions are updated

	// User preferences
	VerboseCompletions bool     `json:"verbose_completions"` // Show detailed descriptions
	CaseSensitive      bool     `json:"case_sensitive"`      // Case sensitive completions
	PreferredStyle     string   `json:"preferred_style"`     // completion style preferences
	ExcludePatterns    []string `json:"exclude_patterns"`    // Patterns to exclude from completion

	// Quality and performance
	CompletionSpeed string `json:"completion_speed"` // fast, normal, comprehensive
	CacheEnabled    bool   `json:"cache_enabled"`    // Whether to use completion caching
	CachePath       string `json:"cache_path"`       // Path to completion cache
	MaxSuggestions  int    `json:"max_suggestions"`  // Maximum number of suggestions

	// Error and diagnostic information
	CompletionErrors []string `json:"completion_errors"` // Current completion issues
	DiagnosticInfo   string   `json:"diagnostic_info"`   // Diagnostic information
	SystemInfo       string   `json:"system_info"`       // System information for context
}

// CompletionAgent handles shell completion generation, installation, and management tasks.
type CompletionAgent struct {
	BaseAgent
	contextData *CompletionContext
}

// NewCompletionAgent creates a new completion management agent.
func NewCompletionAgent(provider ai.Provider) *CompletionAgent {
	agent := &CompletionAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleCompletion,
		},
	}
	return agent
}

// SetContext sets the completion context for the agent.
func (a *CompletionAgent) SetContext(ctx interface{}) error {
	if completionCtx, ok := ctx.(*CompletionContext); ok {
		a.contextData = completionCtx
		return nil
	}
	return fmt.Errorf("invalid context type for CompletionAgent")
}

// GenerateCompletionScript generates shell completion scripts for commands.
func (a *CompletionAgent) GenerateCompletionScript(ctx context.Context, command string, shellType string) (string, error) {
	if a.contextData == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionGenerationPrompt(command, shellType)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate completion script: %w", err)
	}

	return response, nil
}

// InstallCompletions provides guidance for installing shell completions.
func (a *CompletionAgent) InstallCompletions(ctx context.Context, shellType string, installPath string) (string, error) {
	if a.contextData == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionInstallationPrompt(shellType, installPath)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to provide installation guidance: %w", err)
	}

	return response, nil
}

// DiagnoseCompletionIssues analyzes and provides solutions for completion problems.
func (a *CompletionAgent) DiagnoseCompletionIssues(ctx context.Context, issues string) (string, error) {
	if a.contextData == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionDiagnosisPrompt(issues)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to diagnose completion issues: %w", err)
	}

	return response, nil
}

// OptimizeCompletions provides recommendations for improving completion performance.
func (a *CompletionAgent) OptimizeCompletions(ctx context.Context, currentSetup string) (string, error) {
	if a.contextData == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionOptimizationPrompt(currentSetup)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to optimize completions: %w", err)
	}

	return response, nil
}

// enhancePromptWithRole adds role-specific context to prompts.
func (a *CompletionAgent) enhancePromptWithRole(prompt string) string {
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		return fmt.Sprintf("%s\n\n%s", template, prompt)
	}
	return prompt
}

// buildCompletionGenerationPrompt constructs a prompt for generating completion scripts.
func (a *CompletionAgent) buildCompletionGenerationPrompt(command string, shellType string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(fmt.Sprintf("Generate a %s shell completion script for the command '%s'.\n\n", shellType, command))

	if a.contextData != nil {
		if len(a.contextData.Subcommands) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("Available subcommands: %s\n", strings.Join(a.contextData.Subcommands, ", ")))
		}

		if len(a.contextData.Flags) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("Available flags: %s\n", strings.Join(a.contextData.Flags, ", ")))
		}

		if len(a.contextData.FlagDescriptions) > 0 {
			promptBuilder.WriteString("Flag descriptions:\n")
			for flag, desc := range a.contextData.FlagDescriptions {
				promptBuilder.WriteString(fmt.Sprintf("  %s: %s\n", flag, desc))
			}
		}

		if a.contextData.VerboseCompletions {
			promptBuilder.WriteString("Include detailed descriptions for completions.\n")
		}

		if a.contextData.MaxSuggestions > 0 {
			promptBuilder.WriteString(fmt.Sprintf("Limit suggestions to %d items.\n", a.contextData.MaxSuggestions))
		}
	}

	promptBuilder.WriteString("\nProvide a complete, functional completion script with proper error handling.")

	return promptBuilder.String()
}

// buildCompletionInstallationPrompt constructs a prompt for installation guidance.
func (a *CompletionAgent) buildCompletionInstallationPrompt(shellType string, installPath string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(fmt.Sprintf("Provide step-by-step instructions for installing %s shell completions", shellType))

	if installPath != "" {
		promptBuilder.WriteString(fmt.Sprintf(" at path: %s", installPath))
	}

	promptBuilder.WriteString(".\n\n")

	if a.contextData != nil {
		if a.contextData.PackageManager != "" {
			promptBuilder.WriteString(fmt.Sprintf("Package manager: %s\n", a.contextData.PackageManager))
		}

		if a.contextData.InstallMethod != "" {
			promptBuilder.WriteString(fmt.Sprintf("Preferred install method: %s\n", a.contextData.InstallMethod))
		}

		if a.contextData.CompletionPath != "" {
			promptBuilder.WriteString(fmt.Sprintf("Completion path: %s\n", a.contextData.CompletionPath))
		}

		if len(a.contextData.ExistingCompletions) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("Existing completions: %s\n", strings.Join(a.contextData.ExistingCompletions, ", ")))
		}
	}

	promptBuilder.WriteString("\nInclude verification steps and troubleshooting tips.")

	return promptBuilder.String()
}

// buildCompletionDiagnosisPrompt constructs a prompt for diagnosing completion issues.
func (a *CompletionAgent) buildCompletionDiagnosisPrompt(issues string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(fmt.Sprintf("Diagnose and provide solutions for the following shell completion issues:\n\n%s\n\n", issues))

	if a.contextData != nil {
		if a.contextData.ShellType != "" {
			promptBuilder.WriteString(fmt.Sprintf("Shell type: %s\n", a.contextData.ShellType))
		}

		if a.contextData.ShellVersion != "" {
			promptBuilder.WriteString(fmt.Sprintf("Shell version: %s\n", a.contextData.ShellVersion))
		}

		if a.contextData.CompletionType != "" {
			promptBuilder.WriteString(fmt.Sprintf("Completion system: %s\n", a.contextData.CompletionType))
		}

		if len(a.contextData.CompletionErrors) > 0 {
			promptBuilder.WriteString("Current errors:\n")
			for _, err := range a.contextData.CompletionErrors {
				promptBuilder.WriteString(fmt.Sprintf("  - %s\n", err))
			}
		}

		if a.contextData.SystemInfo != "" {
			promptBuilder.WriteString(fmt.Sprintf("System info: %s\n", a.contextData.SystemInfo))
		}

		if a.contextData.DiagnosticInfo != "" {
			promptBuilder.WriteString(fmt.Sprintf("Diagnostic info: %s\n", a.contextData.DiagnosticInfo))
		}
	}

	promptBuilder.WriteString("\nProvide step-by-step troubleshooting solutions.")

	return promptBuilder.String()
}

// buildCompletionOptimizationPrompt constructs a prompt for optimizing completions.
func (a *CompletionAgent) buildCompletionOptimizationPrompt(currentSetup string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(fmt.Sprintf("Analyze and provide optimization recommendations for this completion setup:\n\n%s\n\n", currentSetup))

	if a.contextData != nil {
		if a.contextData.CompletionSpeed != "" {
			promptBuilder.WriteString(fmt.Sprintf("Target speed: %s\n", a.contextData.CompletionSpeed))
		}

		if a.contextData.CacheEnabled {
			promptBuilder.WriteString("Caching is enabled\n")
			if a.contextData.CachePath != "" {
				promptBuilder.WriteString(fmt.Sprintf("Cache path: %s\n", a.contextData.CachePath))
			}
		}

		if a.contextData.MaxSuggestions > 0 {
			promptBuilder.WriteString(fmt.Sprintf("Max suggestions: %d\n", a.contextData.MaxSuggestions))
		}

		if len(a.contextData.ExcludePatterns) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("Exclude patterns: %s\n", strings.Join(a.contextData.ExcludePatterns, ", ")))
		}

		if a.contextData.PreferredStyle != "" {
			promptBuilder.WriteString(fmt.Sprintf("Preferred style: %s\n", a.contextData.PreferredStyle))
		}
	}

	promptBuilder.WriteString("\nFocus on performance, usability, and maintainability improvements.")

	return promptBuilder.String()
}
