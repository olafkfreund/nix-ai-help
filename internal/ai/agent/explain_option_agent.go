package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/mcp"
)

// ExplainOptionAgent is specialized for explaining NixOS configuration options.
type ExplainOptionAgent struct {
	BaseAgent
	mcpClient *mcp.MCPClient
}

// OptionContext contains structured information for explaining NixOS options.
type OptionContext struct {
	OptionPath   string            // e.g., "services.nginx.enable"
	OptionType   string            // e.g., "boolean", "string", "listOf attrs"
	DefaultValue string            // Default value if known
	Description  string            // Brief description
	Examples     []string          // Configuration examples
	RelatedOpts  []string          // Related options
	PackageName  string            // Associated package
	ServiceName  string            // Associated service
	UseCase      string            // When to use this option
	Category     string            // services, programs, system, etc.
	Metadata     map[string]string // Additional context
}

// NewExplainOptionAgent creates a new ExplainOptionAgent.
func NewExplainOptionAgent(provider ai.Provider, mcpClient *mcp.MCPClient) *ExplainOptionAgent {
	agent := &ExplainOptionAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleExplainOption,
		},
		mcpClient: mcpClient,
	}
	return agent
}

// Query handles NixOS option explanation requests with enhanced context.
func (a *ExplainOptionAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build enhanced context for the option
	optionCtx, err := a.buildOptionContext(ctx, question)
	if err != nil {
		return "", fmt.Errorf("failed to build option context: %w", err)
	}

	// Build the enhanced prompt
	prompt := a.buildOptionPrompt(question, optionCtx)

	// Query the AI provider
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to query provider: %w", err)
	}

	return response, nil
}

// GenerateResponse generates a response using the provider's GenerateResponse method.
func (a *ExplainOptionAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Enhance the prompt with role-specific instructions
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	return a.provider.GenerateResponse(ctx, enhancedPrompt)
}

// QueryWithContext queries with additional structured context.
func (a *ExplainOptionAgent) QueryWithContext(ctx context.Context, question string, optionCtx *OptionContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildOptionPrompt(question, optionCtx)
	return a.provider.Query(ctx, prompt)
}

// buildOptionContext builds comprehensive context for a NixOS option.
func (a *ExplainOptionAgent) buildOptionContext(ctx context.Context, question string) (*OptionContext, error) {
	optionCtx := &OptionContext{
		Metadata: make(map[string]string),
	}

	// Extract option path from question
	optionPath := a.extractOptionPath(question)
	if optionPath != "" {
		optionCtx.OptionPath = optionPath
		optionCtx.Category = a.categorizeOption(optionPath)
		optionCtx.PackageName = a.extractPackageName(optionPath)
		optionCtx.ServiceName = a.extractServiceName(optionPath)
	}

	// Try to get additional context from MCP if available
	if a.mcpClient != nil {
		mcpInfo, err := a.queryMCPForOption(ctx, optionPath)
		if err == nil && mcpInfo != "" {
			optionCtx.Description = mcpInfo
			optionCtx.Metadata["mcp_source"] = "nixos_options"
		}
	}

	// Determine use case based on option category
	optionCtx.UseCase = a.determineUseCase(optionCtx.Category, optionPath)

	// Add related options based on pattern matching
	optionCtx.RelatedOpts = a.findRelatedOptions(optionPath)

	return optionCtx, nil
}

// buildOptionPrompt constructs an enhanced prompt for option explanation.
func (a *ExplainOptionAgent) buildOptionPrompt(question string, optionCtx *OptionContext) string {
	var prompt strings.Builder

	// Start with role-specific prompt
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("## NixOS Option Explanation Request\n\n")
	prompt.WriteString(fmt.Sprintf("**User Question**: %s\n\n", question))

	if optionCtx != nil {
		prompt.WriteString("### Context Information:\n")

		if optionCtx.OptionPath != "" {
			prompt.WriteString(fmt.Sprintf("- **Option Path**: `%s`\n", optionCtx.OptionPath))
		}

		if optionCtx.Category != "" {
			prompt.WriteString(fmt.Sprintf("- **Category**: %s\n", optionCtx.Category))
		}

		if optionCtx.PackageName != "" {
			prompt.WriteString(fmt.Sprintf("- **Associated Package**: %s\n", optionCtx.PackageName))
		}

		if optionCtx.ServiceName != "" {
			prompt.WriteString(fmt.Sprintf("- **Associated Service**: %s\n", optionCtx.ServiceName))
		}

		if optionCtx.UseCase != "" {
			prompt.WriteString(fmt.Sprintf("- **Primary Use Case**: %s\n", optionCtx.UseCase))
		}

		if len(optionCtx.RelatedOpts) > 0 {
			prompt.WriteString(fmt.Sprintf("- **Related Options**: %s\n", strings.Join(optionCtx.RelatedOpts, ", ")))
		}

		if optionCtx.Description != "" {
			prompt.WriteString(fmt.Sprintf("- **Documentation**: %s\n", optionCtx.Description))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("### Instructions:\n")
	prompt.WriteString("Please provide a comprehensive explanation focusing on:\n")
	prompt.WriteString("1. What this option does and why it's useful\n")
	prompt.WriteString("2. Practical configuration examples (basic and advanced)\n")
	prompt.WriteString("3. Best practices and common pitfalls\n")
	prompt.WriteString("4. Integration with other NixOS components\n")
	prompt.WriteString("5. Real-world use cases and scenarios\n\n")

	return prompt.String()
}

// extractOptionPath attempts to extract a NixOS option path from the question.
func (a *ExplainOptionAgent) extractOptionPath(question string) string {
	// Look for common NixOS option patterns
	patterns := []string{
		`services\.[\w.]+`,
		`programs\.[\w.]+`,
		`system\.[\w.]+`,
		`environment\.[\w.]+`,
		`networking\.[\w.]+`,
		`hardware\.[\w.]+`,
		`boot\.[\w.]+`,
		`security\.[\w.]+`,
		`virtualisation\.[\w.]+`,
	}

	for _, pattern := range patterns {
		if match := findFirstMatch(question, pattern); match != "" {
			return match
		}
	}

	return ""
}

// categorizeOption determines the category of a NixOS option.
func (a *ExplainOptionAgent) categorizeOption(optionPath string) string {
	if strings.HasPrefix(optionPath, "services.") {
		return "System Services"
	} else if strings.HasPrefix(optionPath, "programs.") {
		return "System Programs"
	} else if strings.HasPrefix(optionPath, "environment.") {
		return "Environment Configuration"
	} else if strings.HasPrefix(optionPath, "networking.") {
		return "Network Configuration"
	} else if strings.HasPrefix(optionPath, "hardware.") {
		return "Hardware Configuration"
	} else if strings.HasPrefix(optionPath, "boot.") {
		return "Boot Configuration"
	} else if strings.HasPrefix(optionPath, "security.") {
		return "Security Configuration"
	} else if strings.HasPrefix(optionPath, "system.") {
		return "System Configuration"
	} else if strings.HasPrefix(optionPath, "virtualisation.") {
		return "Virtualisation"
	}
	return "General Configuration"
}

// extractPackageName tries to extract the package name from option path.
func (a *ExplainOptionAgent) extractPackageName(optionPath string) string {
	parts := strings.Split(optionPath, ".")
	if len(parts) >= 2 {
		if parts[0] == "services" || parts[0] == "programs" {
			return parts[1]
		}
	}
	return ""
}

// extractServiceName tries to extract the service name from option path.
func (a *ExplainOptionAgent) extractServiceName(optionPath string) string {
	if strings.HasPrefix(optionPath, "services.") {
		parts := strings.Split(optionPath, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return ""
}

// determineUseCase provides context about when to use specific options.
func (a *ExplainOptionAgent) determineUseCase(category, optionPath string) string {
	switch category {
	case "System Services":
		return "Configure and manage system services like web servers, databases, or network services"
	case "System Programs":
		return "Install and configure system-wide programs and utilities"
	case "Environment Configuration":
		return "Set up system environment variables, packages, and global settings"
	case "Network Configuration":
		return "Configure network interfaces, firewalls, and network services"
	case "Hardware Configuration":
		return "Configure hardware-specific settings, drivers, and device support"
	case "Boot Configuration":
		return "Configure bootloader, kernel parameters, and boot process"
	case "Security Configuration":
		return "Configure security policies, user permissions, and system hardening"
	case "Virtualisation":
		return "Set up containers, VMs, or virtualisation platforms"
	default:
		return "General system configuration and customization"
	}
}

// findRelatedOptions suggests related options based on the current option.
func (a *ExplainOptionAgent) findRelatedOptions(optionPath string) []string {
	var related []string

	if strings.HasPrefix(optionPath, "services.") {
		basePath := strings.Join(strings.Split(optionPath, ".")[:2], ".")
		related = append(related,
			basePath+".enable",
			basePath+".package",
			basePath+".user",
			basePath+".group",
			basePath+".configFile",
		)
	} else if strings.HasPrefix(optionPath, "programs.") {
		basePath := strings.Join(strings.Split(optionPath, ".")[:2], ".")
		related = append(related,
			basePath+".enable",
			basePath+".package",
			basePath+".settings",
			basePath+".extraConfig",
			basePath+".aliases",
		)
	}

	// Filter out the original option path
	filtered := make([]string, 0, len(related))
	for _, opt := range related {
		if opt != optionPath {
			filtered = append(filtered, opt)
		}
	}

	return filtered
}

// queryMCPForOption attempts to get option information from MCP server.
func (a *ExplainOptionAgent) queryMCPForOption(ctx context.Context, optionPath string) (string, error) {
	if a.mcpClient == nil || optionPath == "" {
		return "", fmt.Errorf("MCP client not available or option path empty")
	}

	// Query NixOS options documentation
	query := fmt.Sprintf("NixOS option %s", optionPath)
	response, err := a.mcpClient.QueryDocumentation(query)
	if err != nil {
		return "", err
	}

	return response, nil
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *ExplainOptionAgent) enhancePromptWithRole(prompt string) string {
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		return fmt.Sprintf("%s\n\n%s", template, prompt)
	}
	return prompt
}
