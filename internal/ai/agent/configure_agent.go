package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// ConfigurationContext represents context for system configuration operations.
type ConfigurationContext struct {
	// System information
	Hardware     string `json:"hardware,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	BootLoader   string `json:"boot_loader,omitempty"`
	FileSystem   string `json:"file_system,omitempty"`

	// Configuration requirements
	DesktopEnvironment string   `json:"desktop_environment,omitempty"`
	Services           []string `json:"services,omitempty"`
	Users              []string `json:"users,omitempty"`
	NetworkConfig      string   `json:"network_config,omitempty"`

	// Setup preferences
	InstallationType   string `json:"installation_type,omitempty"`
	SecurityLevel      string `json:"security_level,omitempty"`
	PerformanceProfile string `json:"performance_profile,omitempty"`

	// Current status
	CurrentConfig     string   `json:"current_config,omitempty"`
	ConfigurationFile string   `json:"configuration_file,omitempty"`
	Issues            []string `json:"issues,omitempty"`
}

// ConfigureAgent represents an agent specialized in NixOS system configuration.
type ConfigureAgent struct {
	BaseAgent
	context *ConfigurationContext
}

// NewConfigureAgent creates a new ConfigureAgent instance.
func NewConfigureAgent(provider ai.Provider) *ConfigureAgent {
	agent := &ConfigureAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleConfigure,
		},
		context: &ConfigurationContext{},
	}
	return agent
}

// Query handles configuration-related queries with context awareness.
func (a *ConfigureAgent) Query(ctx context.Context, input string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build configuration-specific prompt
	prompt := a.buildConfigurationPrompt(input)

	// Query the provider
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("configuration query failed: %w", err)
	}

	// Format and enhance the response
	formattedResponse := a.formatConfigurationResponse(response)

	return formattedResponse, nil
}

// GenerateResponse generates responses for configuration tasks.
func (a *ConfigureAgent) GenerateResponse(ctx context.Context, input string) (string, error) {
	return a.Query(ctx, input)
}

// SetConfigurationContext sets the configuration context for the agent.
func (a *ConfigureAgent) SetConfigurationContext(context *ConfigurationContext) {
	a.context = context
}

// GetConfigurationContext returns the current configuration context.
func (a *ConfigureAgent) GetConfigurationContext() *ConfigurationContext {
	return a.context
}

// AnalyzeSystemRequirements analyzes system requirements for configuration.
func (a *ConfigureAgent) AnalyzeSystemRequirements(ctx context.Context, systemInfo string) (string, error) {
	prompt := fmt.Sprintf(`Based on the following system information, analyze the configuration requirements and provide recommendations:

System Information:
%s

Please provide:
1. Hardware-specific configuration needs
2. Recommended service setup
3. Security considerations
4. Performance optimizations
5. Potential compatibility issues

Current Configuration Context:
%s`, systemInfo, a.formatConfigurationContext())

	return a.Query(ctx, prompt)
}

// GenerateInitialConfiguration generates initial NixOS configuration.
func (a *ConfigureAgent) GenerateInitialConfiguration(ctx context.Context, requirements string) (string, error) {
	prompt := fmt.Sprintf(`Generate an initial NixOS configuration.nix file based on these requirements:

Requirements:
%s

Configuration Context:
%s

Please provide:
1. Complete configuration.nix file
2. Explanation of each major section
3. Optional configurations to consider
4. Next steps for customization

Focus on creating a working, secure, and maintainable configuration.`, requirements, a.formatConfigurationContext())

	return a.Query(ctx, prompt)
}

// ValidateConfiguration validates a NixOS configuration.
func (a *ConfigureAgent) ValidateConfiguration(ctx context.Context, configContent string) (string, error) {
	prompt := fmt.Sprintf(`Validate the following NixOS configuration and identify any issues:

Configuration:
%s

Current Context:
%s

Please check for:
1. Syntax errors and typos
2. Deprecated or invalid options
3. Security vulnerabilities
4. Performance issues
5. Missing dependencies
6. Conflicting configurations

Provide specific fixes for any issues found.`, configContent, a.formatConfigurationContext())

	return a.Query(ctx, prompt)
}

// OptimizeConfiguration suggests optimizations for a configuration.
func (a *ConfigureAgent) OptimizeConfiguration(ctx context.Context, configContent string) (string, error) {
	prompt := fmt.Sprintf(`Analyze and suggest optimizations for this NixOS configuration:

Configuration:
%s

System Context:
%s

Please provide:
1. Performance optimization opportunities
2. Security hardening suggestions
3. Maintainability improvements
4. Resource usage optimizations
5. Modern NixOS best practices

Include specific configuration changes and explanations.`, configContent, a.formatConfigurationContext())

	return a.Query(ctx, prompt)
}

// TroubleshootConfiguration helps troubleshoot configuration issues.
func (a *ConfigureAgent) TroubleshootConfiguration(ctx context.Context, issue string) (string, error) {
	prompt := fmt.Sprintf(`Help troubleshoot this NixOS configuration issue:

Issue Description:
%s

Configuration Context:
%s

Please provide:
1. Root cause analysis
2. Step-by-step resolution steps
3. Prevention strategies
4. Alternative approaches
5. Testing and validation steps

Include specific commands and configuration changes needed.`, issue, a.formatConfigurationContext())

	return a.Query(ctx, prompt)
}

// buildConfigurationPrompt builds a configuration-specific prompt.
func (a *ConfigureAgent) buildConfigurationPrompt(input string) string {
	var builder strings.Builder

	// Add role template
	if template, exists := roles.RolePromptTemplate[roles.RoleConfigure]; exists {
		builder.WriteString(template)
		builder.WriteString("\n\n")
	}

	// Add configuration context if available
	if a.context != nil {
		contextStr := a.formatConfigurationContext()
		if contextStr != "" {
			builder.WriteString("Configuration Context:\n")
			builder.WriteString(contextStr)
			builder.WriteString("\n\n")
		}
	}

	// Add user input
	builder.WriteString("User Query: ")
	builder.WriteString(input)

	return builder.String()
}

// formatConfigurationContext formats the configuration context for inclusion in prompts.
func (a *ConfigureAgent) formatConfigurationContext() string {
	if a.context == nil {
		return ""
	}

	var parts []string

	// System information
	if a.context.Hardware != "" {
		parts = append(parts, fmt.Sprintf("Hardware: %s", a.context.Hardware))
	}
	if a.context.Architecture != "" {
		parts = append(parts, fmt.Sprintf("Architecture: %s", a.context.Architecture))
	}
	if a.context.BootLoader != "" {
		parts = append(parts, fmt.Sprintf("Boot Loader: %s", a.context.BootLoader))
	}
	if a.context.FileSystem != "" {
		parts = append(parts, fmt.Sprintf("File System: %s", a.context.FileSystem))
	}

	// Configuration requirements
	if a.context.DesktopEnvironment != "" {
		parts = append(parts, fmt.Sprintf("Desktop Environment: %s", a.context.DesktopEnvironment))
	}
	if len(a.context.Services) > 0 {
		parts = append(parts, fmt.Sprintf("Required Services: %s", strings.Join(a.context.Services, ", ")))
	}
	if len(a.context.Users) > 0 {
		parts = append(parts, fmt.Sprintf("Users: %s", strings.Join(a.context.Users, ", ")))
	}
	if a.context.NetworkConfig != "" {
		parts = append(parts, fmt.Sprintf("Network Configuration: %s", a.context.NetworkConfig))
	}

	// Setup preferences
	if a.context.InstallationType != "" {
		parts = append(parts, fmt.Sprintf("Installation Type: %s", a.context.InstallationType))
	}
	if a.context.SecurityLevel != "" {
		parts = append(parts, fmt.Sprintf("Security Level: %s", a.context.SecurityLevel))
	}
	if a.context.PerformanceProfile != "" {
		parts = append(parts, fmt.Sprintf("Performance Profile: %s", a.context.PerformanceProfile))
	}

	// Current status
	if a.context.CurrentConfig != "" {
		parts = append(parts, fmt.Sprintf("Current Configuration: %s", a.context.CurrentConfig))
	}
	if a.context.ConfigurationFile != "" {
		parts = append(parts, fmt.Sprintf("Configuration File: %s", a.context.ConfigurationFile))
	}
	if len(a.context.Issues) > 0 {
		parts = append(parts, fmt.Sprintf("Known Issues: %s", strings.Join(a.context.Issues, "; ")))
	}

	return strings.Join(parts, "\n")
}

// formatConfigurationResponse formats the agent response with configuration-specific enhancements.
func (a *ConfigureAgent) formatConfigurationResponse(response string) string {
	var builder strings.Builder

	builder.WriteString("## NixOS Configuration Assistant\n\n")
	builder.WriteString(response)

	// Add configuration safety reminder
	builder.WriteString("\n\n---\n")
	builder.WriteString("**⚠️ Configuration Safety Tips:**\n")
	builder.WriteString("- Always backup your current configuration before making changes\n")
	builder.WriteString("- Test configuration changes with `nixos-rebuild test` first\n")
	builder.WriteString("- Use `nixos-rebuild switch` only after testing\n")
	builder.WriteString("- Keep previous generations available for rollback\n")
	builder.WriteString("- Validate configuration syntax before applying changes\n")

	return builder.String()
}

// validateRole ensures the agent has a valid role set.
func (a *ConfigureAgent) validateRole() error {
	if a.role == "" {
		return fmt.Errorf("configure agent role not set")
	}

	if !roles.ValidateRole(string(a.role)) {
		return fmt.Errorf("invalid configure agent role: %s", a.role)
	}

	return nil
}
