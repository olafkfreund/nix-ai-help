package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/pkg/logger"
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
	context *CompletionContext
}

// NewCompletionAgent creates a new completion management agent.
func NewCompletionAgent(provider AIProvider) *CompletionAgent {
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
		a.context = completionCtx
		return nil
	}
	return fmt.Errorf("invalid context type for CompletionAgent")
}

// GenerateCompletionScript generates shell completion scripts for commands.
func (a *CompletionAgent) GenerateCompletionScript(ctx context.Context, command string, shellType string) (string, error) {
	if a.context == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionGenerationPrompt(command, shellType)

	logger.Debug("CompletionAgent generating completion script",
		"command", command,
		"shell", shellType,
		"role", string(a.role))

	response, err := a.provider.GenerateResponse(ctx, prompt, string(a.role), a.context)
	if err != nil {
		return "", fmt.Errorf("failed to generate completion script: %w", err)
	}

	return response, nil
}

// InstallCompletions provides guidance for installing shell completions.
func (a *CompletionAgent) InstallCompletions(ctx context.Context, shellType string, installPath string) (string, error) {
	if a.context == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionInstallationPrompt(shellType, installPath)

	logger.Debug("CompletionAgent providing installation guidance",
		"shell", shellType,
		"path", installPath,
		"role", string(a.role))

	response, err := a.provider.GenerateResponse(ctx, prompt, string(a.role), a.context)
	if err != nil {
		return "", fmt.Errorf("failed to provide installation guidance: %w", err)
	}

	return response, nil
}

// DiagnoseCompletionIssues analyzes and provides solutions for completion problems.
func (a *CompletionAgent) DiagnoseCompletionIssues(ctx context.Context, issues string) (string, error) {
	if a.context == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionDiagnosisPrompt(issues)

	logger.Debug("CompletionAgent diagnosing completion issues",
		"issues", issues,
		"role", string(a.role))

	response, err := a.provider.GenerateResponse(ctx, prompt, string(a.role), a.context)
	if err != nil {
		return "", fmt.Errorf("failed to diagnose completion issues: %w", err)
	}

	return response, nil
}

// OptimizeCompletions provides recommendations for improving completion performance.
func (a *CompletionAgent) OptimizeCompletions(ctx context.Context, currentSetup string) (string, error) {
	if a.context == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCompletionOptimizationPrompt(currentSetup)

	logger.Debug("CompletionAgent optimizing completion setup",
		"current_setup", currentSetup,
		"role", string(a.role))

	response, err := a.provider.GenerateResponse(ctx, prompt, string(a.role), a.context)
	if err != nil {
		return "", fmt.Errorf("failed to optimize completions: %w", err)
	}

	return response, nil
}

// SetupAdvancedCompletions provides guidance for advanced completion features.
func (a *CompletionAgent) SetupAdvancedCompletions(ctx context.Context, features []string) (string, error) {
	if a.context == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildAdvancedCompletionPrompt(features)

	logger.Debug("CompletionAgent setting up advanced completions",
		"features", strings.Join(features, ", "),
		"role", string(a.role))

	response, err := a.provider.GenerateResponse(ctx, prompt, string(a.role), a.context)
	if err != nil {
		return "", fmt.Errorf("failed to setup advanced completions: %w", err)
	}

	return response, nil
}

// ManageCompletionCache provides cache management strategies and commands.
func (a *CompletionAgent) ManageCompletionCache(ctx context.Context, operation string) (string, error) {
	if a.context == nil {
		return "", fmt.Errorf("completion context not set")
	}

	prompt := a.buildCacheManagementPrompt(operation)

	logger.Debug("CompletionAgent managing completion cache",
		"operation", operation,
		"role", string(a.role))

	response, err := a.provider.GenerateResponse(ctx, prompt, string(a.role), a.context)
	if err != nil {
		return "", fmt.Errorf("failed to manage completion cache: %w", err)
	}

	return response, nil
}

// buildCompletionGenerationPrompt creates a prompt for generating completion scripts.
func (a *CompletionAgent) buildCompletionGenerationPrompt(command string, shellType string) string {
	var prompt strings.Builder

	// Add role-specific context
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString(fmt.Sprintf("Generate a comprehensive %s completion script for the '%s' command.\n\n", shellType, command))

	// Add context information
	if a.context != nil {
		prompt.WriteString("**Current Context:**\n")
		prompt.WriteString(fmt.Sprintf("- Shell: %s (%s)\n", a.context.ShellType, a.context.ShellVersion))

		if len(a.context.Subcommands) > 0 {
			prompt.WriteString(fmt.Sprintf("- Subcommands: %s\n", strings.Join(a.context.Subcommands, ", ")))
		}

		if len(a.context.Flags) > 0 {
			prompt.WriteString(fmt.Sprintf("- Flags: %s\n", strings.Join(a.context.Flags, ", ")))
		}

		if a.context.CompletionPath != "" {
			prompt.WriteString(fmt.Sprintf("- Installation Path: %s\n", a.context.CompletionPath))
		}

		if a.context.VerboseCompletions {
			prompt.WriteString("- Style: Verbose with descriptions\n")
		}

		if a.context.MaxSuggestions > 0 {
			prompt.WriteString(fmt.Sprintf("- Max Suggestions: %d\n", a.context.MaxSuggestions))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. **Complete Completion Script**: Full, ready-to-use completion script\n")
	prompt.WriteString("2. **Installation Instructions**: How to install and enable the completions\n")
	prompt.WriteString("3. **Testing Guide**: How to test that completions work correctly\n")
	prompt.WriteString("4. **Customization Options**: Ways to customize or extend the completions\n")
	prompt.WriteString("5. **Troubleshooting**: Common issues and solutions\n\n")

	prompt.WriteString("Focus on creating efficient, user-friendly completions that enhance the command-line experience.")

	return prompt.String()
}

// buildCompletionInstallationPrompt creates a prompt for completion installation guidance.
func (a *CompletionAgent) buildCompletionInstallationPrompt(shellType string, installPath string) string {
	var prompt strings.Builder

	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString(fmt.Sprintf("Provide comprehensive installation guidance for %s shell completions.\n\n", shellType))

	if a.context != nil {
		prompt.WriteString("**Installation Context:**\n")
		prompt.WriteString(fmt.Sprintf("- Target Shell: %s (%s)\n", a.context.ShellType, a.context.ShellVersion))
		prompt.WriteString(fmt.Sprintf("- Installation Path: %s\n", installPath))

		if a.context.PackageManager != "" {
			prompt.WriteString(fmt.Sprintf("- Package Manager: %s\n", a.context.PackageManager))
		}

		if len(a.context.ExistingCompletions) > 0 {
			prompt.WriteString(fmt.Sprintf("- Existing Completions: %s\n", strings.Join(a.context.ExistingCompletions, ", ")))
		}

		if a.context.SystemInfo != "" {
			prompt.WriteString(fmt.Sprintf("- System: %s\n", a.context.SystemInfo))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. **Installation Steps**: Detailed, step-by-step installation process\n")
	prompt.WriteString("2. **Path Configuration**: How to set up correct paths and permissions\n")
	prompt.WriteString("3. **Shell Configuration**: Required changes to shell config files\n")
	prompt.WriteString("4. **Verification**: How to verify the installation was successful\n")
	prompt.WriteString("5. **Alternative Methods**: Different installation approaches\n")
	prompt.WriteString("6. **Uninstallation**: How to remove completions if needed\n\n")

	prompt.WriteString("Focus on providing clear, safe installation procedures with proper error handling.")

	return prompt.String()
}

// buildCompletionDiagnosisPrompt creates a prompt for diagnosing completion issues.
func (a *CompletionAgent) buildCompletionDiagnosisPrompt(issues string) string {
	var prompt strings.Builder

	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("Diagnose and provide solutions for the following shell completion issues:\n\n")
	prompt.WriteString(fmt.Sprintf("**Reported Issues:**\n%s\n\n", issues))

	if a.context != nil {
		prompt.WriteString("**Diagnostic Context:**\n")
		prompt.WriteString(fmt.Sprintf("- Shell: %s (%s)\n", a.context.ShellType, a.context.ShellVersion))
		prompt.WriteString(fmt.Sprintf("- Completion Type: %s\n", a.context.CompletionType))

		if a.context.CompletionPath != "" {
			prompt.WriteString(fmt.Sprintf("- Completion Path: %s\n", a.context.CompletionPath))
		}

		if len(a.context.CompletionErrors) > 0 {
			prompt.WriteString(fmt.Sprintf("- Error Messages: %s\n", strings.Join(a.context.CompletionErrors, "; ")))
		}

		if a.context.DiagnosticInfo != "" {
			prompt.WriteString(fmt.Sprintf("- Diagnostic Info: %s\n", a.context.DiagnosticInfo))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. **Problem Analysis**: Root cause analysis of the completion issues\n")
	prompt.WriteString("2. **Solution Steps**: Specific steps to resolve each issue\n")
	prompt.WriteString("3. **Verification Commands**: Commands to test that fixes work\n")
	prompt.WriteString("4. **Prevention**: How to avoid these issues in the future\n")
	prompt.WriteString("5. **Alternative Solutions**: Backup approaches if primary fixes fail\n\n")

	prompt.WriteString("Focus on systematic troubleshooting and reliable solutions.")

	return prompt.String()
}

// buildCompletionOptimizationPrompt creates a prompt for optimizing completion performance.
func (a *CompletionAgent) buildCompletionOptimizationPrompt(currentSetup string) string {
	var prompt strings.Builder

	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("Analyze and optimize the current shell completion setup for better performance and user experience.\n\n")
	prompt.WriteString(fmt.Sprintf("**Current Setup:**\n%s\n\n", currentSetup))

	if a.context != nil {
		prompt.WriteString("**Optimization Context:**\n")
		prompt.WriteString(fmt.Sprintf("- Shell: %s (%s)\n", a.context.ShellType, a.context.ShellVersion))
		prompt.WriteString(fmt.Sprintf("- Completion Speed: %s\n", a.context.CompletionSpeed))
		prompt.WriteString(fmt.Sprintf("- Cache Enabled: %t\n", a.context.CacheEnabled))

		if a.context.CachePath != "" {
			prompt.WriteString(fmt.Sprintf("- Cache Path: %s\n", a.context.CachePath))
		}

		if a.context.MaxSuggestions > 0 {
			prompt.WriteString(fmt.Sprintf("- Max Suggestions: %d\n", a.context.MaxSuggestions))
		}

		if len(a.context.ExistingCompletions) > 0 {
			prompt.WriteString(fmt.Sprintf("- Installed Completions: %d\n", len(a.context.ExistingCompletions)))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. **Performance Analysis**: Assessment of current completion performance\n")
	prompt.WriteString("2. **Optimization Recommendations**: Specific improvements to implement\n")
	prompt.WriteString("3. **Caching Strategies**: How to improve completion caching\n")
	prompt.WriteString("4. **Configuration Tuning**: Shell-specific optimization settings\n")
	prompt.WriteString("5. **Cleanup Suggestions**: Removing unused or slow completions\n")
	prompt.WriteString("6. **Monitoring**: How to track completion performance over time\n\n")

	prompt.WriteString("Focus on practical optimizations that provide measurable improvements.")

	return prompt.String()
}

// buildAdvancedCompletionPrompt creates a prompt for setting up advanced completion features.
func (a *CompletionAgent) buildAdvancedCompletionPrompt(features []string) string {
	var prompt strings.Builder

	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("Set up advanced shell completion features for enhanced user experience.\n\n")
	prompt.WriteString(fmt.Sprintf("**Requested Features:**\n%s\n\n", strings.Join(features, "\n- ")))

	if a.context != nil {
		prompt.WriteString("**Setup Context:**\n")
		prompt.WriteString(fmt.Sprintf("- Shell: %s (%s)\n", a.context.ShellType, a.context.ShellVersion))
		prompt.WriteString(fmt.Sprintf("- Completion Scope: %s\n", a.context.CompletionScope))

		if len(a.context.CustomFunctions) > 0 {
			prompt.WriteString(fmt.Sprintf("- Custom Functions: %s\n", strings.Join(a.context.CustomFunctions, ", ")))
		}

		if a.context.PreferredStyle != "" {
			prompt.WriteString(fmt.Sprintf("- Preferred Style: %s\n", a.context.PreferredStyle))
		}

		prompt.WriteString(fmt.Sprintf("- Verbose Completions: %t\n", a.context.VerboseCompletions))
		prompt.WriteString(fmt.Sprintf("- Case Sensitive: %t\n", a.context.CaseSensitive))

		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. **Feature Implementation**: How to implement each requested feature\n")
	prompt.WriteString("2. **Configuration Examples**: Complete configuration examples\n")
	prompt.WriteString("3. **Integration Guide**: How features work together\n")
	prompt.WriteString("4. **Customization Options**: Ways to tailor features to user needs\n")
	prompt.WriteString("5. **Best Practices**: Recommended patterns and approaches\n")
	prompt.WriteString("6. **Troubleshooting**: Common issues with advanced features\n\n")

	prompt.WriteString("Focus on creating powerful, user-friendly completion experiences.")

	return prompt.String()
}

// buildCacheManagementPrompt creates a prompt for completion cache management.
func (a *CompletionAgent) buildCacheManagementPrompt(operation string) string {
	var prompt strings.Builder

	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString(fmt.Sprintf("Provide guidance for completion cache management operation: %s\n\n", operation))

	if a.context != nil {
		prompt.WriteString("**Cache Context:**\n")
		prompt.WriteString(fmt.Sprintf("- Shell: %s (%s)\n", a.context.ShellType, a.context.ShellVersion))
		prompt.WriteString(fmt.Sprintf("- Cache Enabled: %t\n", a.context.CacheEnabled))

		if a.context.CachePath != "" {
			prompt.WriteString(fmt.Sprintf("- Cache Path: %s\n", a.context.CachePath))
		}

		prompt.WriteString(fmt.Sprintf("- Completion Speed: %s\n", a.context.CompletionSpeed))

		if len(a.context.ExistingCompletions) > 0 {
			prompt.WriteString(fmt.Sprintf("- Active Completions: %d\n", len(a.context.ExistingCompletions)))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. **Operation Steps**: Detailed steps for the cache operation\n")
	prompt.WriteString("2. **Safety Checks**: How to safely perform cache operations\n")
	prompt.WriteString("3. **Performance Impact**: Expected impact on completion performance\n")
	prompt.WriteString("4. **Recovery Procedures**: How to recover from cache issues\n")
	prompt.WriteString("5. **Maintenance Schedule**: Recommended cache maintenance practices\n")
	prompt.WriteString("6. **Monitoring**: How to monitor cache health and performance\n\n")

	prompt.WriteString("Focus on safe, effective cache management that maintains completion reliability.")

	return prompt.String()
}
