package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/ai/validation"
)

// FlakeAgent handles Nix flake-related operations and guidance
type FlakeAgent struct {
	BaseAgent
	validator *validation.FlakeValidator
}

// FlakeContext contains flake-specific context information
type FlakeContext struct {
	FlakePath     string            `json:"flake_path,omitempty"`
	FlakeNix      string            `json:"flake_nix,omitempty"`
	FlakeLock     string            `json:"flake_lock,omitempty"`
	FlakeInputs   map[string]string `json:"flake_inputs,omitempty"`
	FlakeOutputs  []string          `json:"flake_outputs,omitempty"`
	FlakeMetadata string            `json:"flake_metadata,omitempty"`
	FlakeErrors   []string          `json:"flake_errors,omitempty"`
	FlakeCommands []string          `json:"flake_commands,omitempty"`
	ProjectType   string            `json:"project_type,omitempty"` // nixos, home-manager, dev-shell, etc.
	FlakeSystem   string            `json:"flake_system,omitempty"` // x86_64-linux, etc.
	Dependencies  []string          `json:"dependencies,omitempty"`
	BuildOutputs  string            `json:"build_outputs,omitempty"`
}

// NewFlakeAgent creates a new FlakeAgent with the Flake role
func NewFlakeAgent(provider ai.Provider) *FlakeAgent {
	agent := &FlakeAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleFlake,
		},
		validator: validation.NewFlakeValidator(),
	}
	return agent
}

// Query handles flake-related questions and guidance
func (a *FlakeAgent) Query(ctx context.Context, question string) (string, error) {
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

	// Get response from AI provider
	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, fullPrompt)
		if err != nil {
			return "", err
		}
		// Validate the response for flake syntax issues
		if a.validator != nil && validation.IsFlakeContent(response) {
			validationResult := a.validator.ValidateFlakeContent(response)
			if !validationResult.IsValid || len(validationResult.Warnings) > 0 {
				validationSummary := a.validator.FormatValidationResult(validationResult)
				// Append validation summary to the response
				response += "\n\n---\n\n" + validationSummary
			}
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(fullPrompt)
		if err != nil {
			return "", err
		}
		if a.validator != nil && validation.IsFlakeContent(response) {
			validationResult := a.validator.ValidateFlakeContent(response)
			if !validationResult.IsValid || len(validationResult.Warnings) > 0 {
				validationSummary := a.validator.FormatValidationResult(validationResult)
				response += "\n\n---\n\n" + validationSummary
			}
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateResponse handles flake-related response generation
func (a *FlakeAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("AI provider not configured")
	}

	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add flake-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// buildContextualPrompt constructs a context-aware prompt for flake operations
func (a *FlakeAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add flake context if available
	if a.contextData != nil {
		if flakeCtx, ok := a.contextData.(*FlakeContext); ok {
			contextStr := a.formatFlakeContext(flakeCtx)
			if contextStr != "" {
				promptParts = append(promptParts, "Flake Context:")
				promptParts = append(promptParts, contextStr)
			}
		}
	}

	// Add user input
	promptParts = append(promptParts, "Flake Request:")
	promptParts = append(promptParts, userInput)

	return strings.Join(promptParts, "\n\n")
}

// formatFlakeContext formats FlakeContext into a readable string
func (a *FlakeAgent) formatFlakeContext(ctx *FlakeContext) string {
	var parts []string

	if ctx.FlakePath != "" {
		parts = append(parts, fmt.Sprintf("Flake Path: %s", ctx.FlakePath))
	}

	if ctx.ProjectType != "" {
		parts = append(parts, fmt.Sprintf("Project Type: %s", ctx.ProjectType))
	}

	if ctx.FlakeSystem != "" {
		parts = append(parts, fmt.Sprintf("System: %s", ctx.FlakeSystem))
	}

	if len(ctx.FlakeInputs) > 0 {
		inputsStr := make([]string, 0, len(ctx.FlakeInputs))
		for name, url := range ctx.FlakeInputs {
			inputsStr = append(inputsStr, fmt.Sprintf("%s: %s", name, url))
		}
		parts = append(parts, fmt.Sprintf("Inputs:\n%s", strings.Join(inputsStr, "\n")))
	}

	if len(ctx.FlakeOutputs) > 0 {
		parts = append(parts, fmt.Sprintf("Outputs: %s", strings.Join(ctx.FlakeOutputs, ", ")))
	}

	if len(ctx.Dependencies) > 0 {
		parts = append(parts, fmt.Sprintf("Dependencies: %s", strings.Join(ctx.Dependencies, ", ")))
	}

	if ctx.FlakeNix != "" {
		parts = append(parts, fmt.Sprintf("flake.nix:\n%s", ctx.FlakeNix))
	}

	if ctx.FlakeLock != "" {
		parts = append(parts, fmt.Sprintf("flake.lock:\n%s", ctx.FlakeLock))
	}

	if ctx.FlakeMetadata != "" {
		parts = append(parts, fmt.Sprintf("Metadata:\n%s", ctx.FlakeMetadata))
	}

	if len(ctx.FlakeErrors) > 0 {
		parts = append(parts, fmt.Sprintf("Errors: %s", strings.Join(ctx.FlakeErrors, ", ")))
	}

	if len(ctx.FlakeCommands) > 0 {
		parts = append(parts, fmt.Sprintf("Commands: %s", strings.Join(ctx.FlakeCommands, ", ")))
	}

	if ctx.BuildOutputs != "" {
		parts = append(parts, fmt.Sprintf("Build Outputs:\n%s", ctx.BuildOutputs))
	}

	return strings.Join(parts, "\n")
}

// SetFlakeContext is a convenience method to set FlakeContext
func (a *FlakeAgent) SetFlakeContext(ctx *FlakeContext) {
	a.SetContext(ctx)
}
