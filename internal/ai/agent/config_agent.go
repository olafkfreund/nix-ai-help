package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/pkg/logger"
)

// ConfigAgent is specialized for NixOS configuration management and analysis.
type ConfigAgent struct {
	role        string
	contextData interface{}
	logger      *logger.Logger
}

// ConfigContext contains structured information for configuration analysis.
type ConfigContext struct {
	ConfigPath      string            // Path to configuration file
	ConfigContent   string            // Content of configuration file
	ConfigType      string            // Type: system, user, module, etc.
	Issues          []string          // Identified configuration issues
	Suggestions     []string          // Improvement suggestions
	Dependencies    []string          // Configuration dependencies
	ConflictsWith   []string          // Potential conflicts
	SecurityIssues  []string          // Security-related issues
	Performance     []string          // Performance considerations
	Maintainability []string          // Maintainability concerns
	Metadata        map[string]string // Additional context
}

// NewConfigAgent creates a new ConfigAgent.
func NewConfigAgent() *ConfigAgent {
	return &ConfigAgent{
		role:   string(roles.RoleConfig),
		logger: logger.NewLogger(),
	}
}

// Query handles configuration analysis queries with role-based prompting.
func (a *ConfigAgent) Query(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	if !roles.ValidateRole(role) {
		return "", fmt.Errorf("unsupported role: %s", role)
	}

	// Get role-specific prompt template
	prompt, ok := roles.RolePromptTemplate[roles.RoleType(role)]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", role)
	}

	// Build enhanced configuration prompt with context
	enhancedPrompt := a.buildConfigPrompt(prompt, input, contextData)

	a.logger.Debug("ConfigAgent: Built enhanced configuration prompt")
	return enhancedPrompt, nil
}

func (a *ConfigAgent) GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	// For now, GenerateResponse behaves the same as Query
	// In the future, this could integrate with actual LLM providers
	return a.Query(ctx, input, role, contextData)
}

func (a *ConfigAgent) SetRole(role string) {
	a.role = role
}

func (a *ConfigAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}

// AnalyzeConfiguration analyzes a NixOS configuration file or content.
func (a *ConfigAgent) AnalyzeConfiguration(ctx context.Context, configPath, configContent string) (string, error) {
	context := &ConfigContext{
		ConfigPath:    configPath,
		ConfigContent: configContent,
		ConfigType:    determineConfigType(configPath),
		Metadata:      make(map[string]string),
	}

	prompt := fmt.Sprintf("Analyze this NixOS configuration:\n\nFile: %s\nContent:\n%s", configPath, configContent)
	return a.Query(ctx, prompt, string(roles.RoleConfig), context)
}

// ReviewConfiguration provides a comprehensive review of configuration quality.
func (a *ConfigAgent) ReviewConfiguration(ctx context.Context, configPath, configContent string, reviewType string) (string, error) {
	context := &ConfigContext{
		ConfigPath:    configPath,
		ConfigContent: configContent,
		ConfigType:    determineConfigType(configPath),
		Metadata: map[string]string{
			"review_type": reviewType,
		},
	}

	prompt := fmt.Sprintf("Review this NixOS configuration for %s:\n\nFile: %s\nContent:\n%s",
		reviewType, configPath, configContent)

	return a.Query(ctx, prompt, string(roles.RoleConfig), context)
}

// SuggestImprovements provides specific improvement suggestions for a configuration.
func (a *ConfigAgent) SuggestImprovements(ctx context.Context, configPath, configContent string, focusAreas []string) (string, error) {
	context := &ConfigContext{
		ConfigPath:    configPath,
		ConfigContent: configContent,
		ConfigType:    determineConfigType(configPath),
		Metadata: map[string]string{
			"focus_areas": strings.Join(focusAreas, ","),
		},
	}

	prompt := fmt.Sprintf("Suggest improvements for this NixOS configuration, focusing on %s:\n\nFile: %s\nContent:\n%s",
		strings.Join(focusAreas, ", "), configPath, configContent)

	return a.Query(ctx, prompt, string(roles.RoleConfig), context)
}

// ValidateConfiguration checks configuration syntax and structure.
func (a *ConfigAgent) ValidateConfiguration(ctx context.Context, configPath, configContent string) (string, error) {
	context := &ConfigContext{
		ConfigPath:    configPath,
		ConfigContent: configContent,
		ConfigType:    determineConfigType(configPath),
		Metadata: map[string]string{
			"validation_type": "syntax_and_structure",
		},
	}

	prompt := fmt.Sprintf("Validate this NixOS configuration for syntax errors and structural issues:\n\nFile: %s\nContent:\n%s",
		configPath, configContent)

	return a.Query(ctx, prompt, string(roles.RoleConfig), context)
}

// buildConfigPrompt constructs a comprehensive configuration prompt with context
func (a *ConfigAgent) buildConfigPrompt(basePrompt, input string, contextData interface{}) string {
	var prompt strings.Builder

	// Start with role-specific base prompt
	prompt.WriteString(basePrompt)
	prompt.WriteString("\n\n")

	// Add structured context if available
	if configCtx, ok := contextData.(*ConfigContext); ok {
		prompt.WriteString("## CONFIGURATION CONTEXT\n\n")

		// Configuration information
		prompt.WriteString("### Configuration Information\n")
		prompt.WriteString(fmt.Sprintf("- **File Path**: %s\n", configCtx.ConfigPath))
		prompt.WriteString(fmt.Sprintf("- **Config Type**: %s\n", configCtx.ConfigType))

		if len(configCtx.Issues) > 0 {
			prompt.WriteString("### Known Issues\n")
			for _, issue := range configCtx.Issues {
				prompt.WriteString(fmt.Sprintf("- %s\n", issue))
			}
			prompt.WriteString("\n")
		}

		if len(configCtx.Dependencies) > 0 {
			prompt.WriteString("### Dependencies\n")
			for _, dep := range configCtx.Dependencies {
				prompt.WriteString(fmt.Sprintf("- %s\n", dep))
			}
			prompt.WriteString("\n")
		}

		if len(configCtx.ConflictsWith) > 0 {
			prompt.WriteString("### Potential Conflicts\n")
			for _, conflict := range configCtx.ConflictsWith {
				prompt.WriteString(fmt.Sprintf("- %s\n", conflict))
			}
			prompt.WriteString("\n")
		}

		if len(configCtx.SecurityIssues) > 0 {
			prompt.WriteString("### Security Issues\n")
			for _, security := range configCtx.SecurityIssues {
				prompt.WriteString(fmt.Sprintf("- %s\n", security))
			}
			prompt.WriteString("\n")
		}

		if len(configCtx.Performance) > 0 {
			prompt.WriteString("### Performance Considerations\n")
			for _, perf := range configCtx.Performance {
				prompt.WriteString(fmt.Sprintf("- %s\n", perf))
			}
			prompt.WriteString("\n")
		}

		if len(configCtx.Maintainability) > 0 {
			prompt.WriteString("### Maintainability Concerns\n")
			for _, maint := range configCtx.Maintainability {
				prompt.WriteString(fmt.Sprintf("- %s\n", maint))
			}
			prompt.WriteString("\n")
		}

		// Metadata
		if len(configCtx.Metadata) > 0 {
			prompt.WriteString("### Additional Context\n")
			for key, value := range configCtx.Metadata {
				prompt.WriteString(fmt.Sprintf("- **%s**: %s\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value))
			}
			prompt.WriteString("\n")
		}

		// Configuration content
		if configCtx.ConfigContent != "" {
			prompt.WriteString("### Configuration Content\n```nix\n")
			prompt.WriteString(configCtx.ConfigContent)
			prompt.WriteString("\n```\n\n")
		}
	}

	// Add the main input/question
	prompt.WriteString("## USER REQUEST\n")
	prompt.WriteString(input)
	prompt.WriteString("\n")

	return prompt.String()
}

// determineConfigType determines the type of configuration file based on its path.
func determineConfigType(configPath string) string {
	switch {
	case strings.Contains(configPath, "/etc/nixos/"):
		return "system"
	case strings.Contains(configPath, "home-manager"):
		return "user"
	case strings.Contains(configPath, "flake.nix"):
		return "flake"
	case strings.Contains(configPath, "default.nix"):
		return "derivation"
	case strings.Contains(configPath, "shell.nix"):
		return "shell"
	case strings.HasSuffix(configPath, ".nix"):
		return "module"
	default:
		return "unknown"
	}
}
