package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// BuildAgent handles NixOS build-related queries and issues
type BuildAgent struct {
	BaseAgent
}

// BuildContext contains build-specific context information
type BuildContext struct {
	BuildOutput    string   `json:"build_output,omitempty"`
	ErrorLogs      string   `json:"error_logs,omitempty"`
	ConfigPath     string   `json:"config_path,omitempty"`
	DerivationPath string   `json:"derivation_path,omitempty"`
	FailedPackages []string `json:"failed_packages,omitempty"`
	BuildSystem    string   `json:"build_system,omitempty"` // nixos-rebuild, nix-build, etc.
	Architecture   string   `json:"architecture,omitempty"` // x86_64-linux, aarch64-linux
	NixChannels    []string `json:"nix_channels,omitempty"`
	SystemInfo     string   `json:"system_info,omitempty"`
}

// NewBuildAgent creates a new BuildAgent with the Build role
func NewBuildAgent(provider ai.Provider) *BuildAgent {
	agent := &BuildAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleBuild,
		},
	}
	return agent
}

// Query handles build-related questions and analysis
func (a *BuildAgent) Query(ctx context.Context, question string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("AI provider not configured")
	}

	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt, ok := roles.RolePromptTemplate[a.role]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", a.role)
	}

	// Build context-aware prompt
	fullPrompt := a.buildContextualPrompt(prompt, question)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		return p.QueryWithContext(ctx, fullPrompt)
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		return p.Query(fullPrompt)
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateResponse handles build-related response generation
func (a *BuildAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add build-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// buildContextualPrompt constructs a context-aware prompt for build operations
func (a *BuildAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add build context if available
	if a.contextData != nil {
		if buildCtx, ok := a.contextData.(*BuildContext); ok {
			contextStr := a.formatBuildContext(buildCtx)
			if contextStr != "" {
				promptParts = append(promptParts, "Build Context:")
				promptParts = append(promptParts, contextStr)
			}
		}
	}

	// Add user input
	promptParts = append(promptParts, "User Request:")
	promptParts = append(promptParts, userInput)

	return strings.Join(promptParts, "\n\n")
}

// formatBuildContext formats BuildContext into a readable string
func (a *BuildAgent) formatBuildContext(ctx *BuildContext) string {
	var parts []string

	if ctx.BuildSystem != "" {
		parts = append(parts, fmt.Sprintf("Build System: %s", ctx.BuildSystem))
	}

	if ctx.Architecture != "" {
		parts = append(parts, fmt.Sprintf("Architecture: %s", ctx.Architecture))
	}

	if len(ctx.NixChannels) > 0 {
		parts = append(parts, fmt.Sprintf("Channels: %s", strings.Join(ctx.NixChannels, ", ")))
	}

	if ctx.ConfigPath != "" {
		parts = append(parts, fmt.Sprintf("Config Path: %s", ctx.ConfigPath))
	}

	if ctx.DerivationPath != "" {
		parts = append(parts, fmt.Sprintf("Derivation: %s", ctx.DerivationPath))
	}

	if len(ctx.FailedPackages) > 0 {
		parts = append(parts, fmt.Sprintf("Failed Packages: %s", strings.Join(ctx.FailedPackages, ", ")))
	}

	if ctx.BuildOutput != "" {
		parts = append(parts, fmt.Sprintf("Build Output:\n%s", ctx.BuildOutput))
	}

	if ctx.ErrorLogs != "" {
		parts = append(parts, fmt.Sprintf("Error Logs:\n%s", ctx.ErrorLogs))
	}

	if ctx.SystemInfo != "" {
		parts = append(parts, fmt.Sprintf("System Info:\n%s", ctx.SystemInfo))
	}

	return strings.Join(parts, "\n")
}

// SetBuildContext is a convenience method to set BuildContext
func (a *BuildAgent) SetBuildContext(ctx *BuildContext) {
	a.SetContext(ctx)
}
