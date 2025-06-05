package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// HelpAgent provides guidance on nixai commands and features.
type HelpAgent struct {
	BaseAgent
}

// HelpContext provides context for help requests.
type HelpContext struct {
	UserGoal     string   // What the user is trying to accomplish
	UserLevel    string   // Beginner, Intermediate, Advanced
	CurrentTask  string   // Current task or problem
	AvailableCmd []string // Available nixai commands
	ErrorContext string   // Any error messages or context
	Environment  string   // System environment details
}

// NewHelpAgent creates a new HelpAgent with the specified provider.
func NewHelpAgent(provider ai.Provider) *HelpAgent {
	agent := &HelpAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleHelp,
		},
	}
	return agent
}

// Query provides command and feature guidance using the provider's Query method.
func (a *HelpAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build help-specific prompt with context
	prompt := a.buildHelpPrompt(question, a.getHelpContextFromData())

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", err
	}

	return a.enhanceResponseWithHelpGuidance(response), nil
}

// GenerateResponse generates a response using the provider's GenerateResponse method.
func (a *HelpAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Enhance the prompt with role-specific instructions
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", err
	}

	return a.enhanceResponseWithHelpGuidance(response), nil
}

// buildHelpPrompt constructs an enhanced prompt for help requests.
func (a *HelpAgent) buildHelpPrompt(question string, helpCtx *HelpContext) string {
	var prompt strings.Builder

	// Start with role-specific prompt
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("## Help Request\n\n")
	prompt.WriteString(fmt.Sprintf("**User Question**: %s\n\n", question))

	if helpCtx != nil {
		prompt.WriteString("### Context Information:\n")

		if helpCtx.UserGoal != "" {
			prompt.WriteString(fmt.Sprintf("- **User Goal**: %s\n", helpCtx.UserGoal))
		}

		if helpCtx.UserLevel != "" {
			prompt.WriteString(fmt.Sprintf("- **Experience Level**: %s\n", helpCtx.UserLevel))
		}

		if helpCtx.CurrentTask != "" {
			prompt.WriteString(fmt.Sprintf("- **Current Task**: %s\n", helpCtx.CurrentTask))
		}

		if len(helpCtx.AvailableCmd) > 0 {
			prompt.WriteString(fmt.Sprintf("- **Available Commands**: %s\n", strings.Join(helpCtx.AvailableCmd, ", ")))
		}

		if helpCtx.ErrorContext != "" {
			prompt.WriteString(fmt.Sprintf("- **Error Context**: %s\n", helpCtx.ErrorContext))
		}

		if helpCtx.Environment != "" {
			prompt.WriteString(fmt.Sprintf("- **Environment**: %s\n", helpCtx.Environment))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("### Available nixai Commands:\n")
	prompt.WriteString(a.buildCommandReference())

	prompt.WriteString("\n\n### Response Guidelines:\n")
	prompt.WriteString("- Recommend the most appropriate command(s) for the user's goal\n")
	prompt.WriteString("- Provide clear command syntax and examples\n")
	prompt.WriteString("- Explain why specific commands are recommended\n")
	prompt.WriteString("- Suggest workflow steps when multiple commands are needed\n")
	prompt.WriteString("- Include relevant tips and best practices\n")

	return prompt.String()
}

// buildCommandReference creates a comprehensive reference of nixai commands.
func (a *HelpAgent) buildCommandReference() string {
	commands := map[string]string{
		"nixai ask":                 "Ask direct questions about NixOS",
		"nixai diagnose":            "Diagnose NixOS problems from logs/configs",
		"nixai doctor":              "Perform comprehensive system health checks",
		"nixai search":              "Search for NixOS packages and options",
		"nixai explain-option":      "Explain NixOS configuration options",
		"nixai explain-home-option": "Explain Home Manager configuration options",
		"nixai build":               "Get help with NixOS build issues",
		"nixai flake":               "Manage and work with Nix flakes",
		"nixai gc":                  "Garbage collection and store cleanup",
		"nixai hardware":            "Hardware detection and configuration",
		"nixai interactive":         "Start interactive troubleshooting session",
		"nixai learn":               "Access learning resources and tutorials",
		"nixai logs":                "Analyze system and service logs",
		"nixai package-repo":        "Analyze repositories for Nix packaging",
		"nixai config":              "Configuration management assistance",
		"nixai community":           "Access community resources and support",
		"nixai machines":            "Manage multiple NixOS machines",
		"nixai devenv":              "Development environment setup",
		"nixai neovim-setup":        "Neovim configuration for NixOS",
		"nixai templates":           "Configuration templates and examples",
		"nixai migrate":             "Migration assistance between NixOS versions",
		"nixai mcp-server":          "MCP server management",
		"nixai snippets":            "Code snippets and configuration examples",
		"nixai store":               "Nix store operations and management",
	}

	var reference strings.Builder
	for cmd, description := range commands {
		reference.WriteString(fmt.Sprintf("- **%s**: %s\n", cmd, description))
	}

	return reference.String()
}

// getHelpContextFromData extracts HelpContext from the agent's context data.
func (a *HelpAgent) getHelpContextFromData() *HelpContext {
	if a.contextData == nil {
		return nil
	}

	if helpCtx, ok := a.contextData.(*HelpContext); ok {
		return helpCtx
	}

	return nil
}

// enhanceResponseWithHelpGuidance adds help-specific guidance to responses.
func (a *HelpAgent) enhanceResponseWithHelpGuidance(response string) string {
	guidance := "\n\n💡 **Additional Help Resources**:\n"
	guidance += "- Use `nixai help` to see all available commands\n"
	guidance += "- Use `nixai interactive` for step-by-step guidance\n"
	guidance += "- Use `nixai doctor` to diagnose system issues\n"
	guidance += "- Use `nixai learn` to access tutorials and documentation\n"
	guidance += "- Use `nixai community` to find community support\n"

	return response + guidance
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *HelpAgent) enhancePromptWithRole(prompt string) string {
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		return fmt.Sprintf("%s\n\n%s", template, prompt)
	}
	return prompt
}

// SetHelpContext is a convenience method to set HelpContext.
func (a *HelpAgent) SetHelpContext(ctx *HelpContext) {
	a.SetContext(ctx)
}

// GetAvailableCommands returns a list of available nixai commands.
func (a *HelpAgent) GetAvailableCommands() []string {
	return []string{
		"ask", "diagnose", "doctor", "search", "explain-option", "explain-home-option",
		"build", "flake", "gc", "hardware", "interactive", "learn", "logs",
		"package-repo", "config", "community", "machines", "devenv", "neovim-setup",
		"templates", "migrate", "mcp-server", "snippets", "store", "help",
	}
}

// SuggestCommand suggests the best command based on user input.
func (a *HelpAgent) SuggestCommand(ctx context.Context, userInput string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Based on the user input: "%s"

Suggest the single most appropriate nixai command from this list:
%s

Respond with just the command name (e.g., "ask", "diagnose", "doctor").`,
		userInput, strings.Join(a.GetAvailableCommands(), ", "))

	return a.provider.Query(ctx, prompt)
}

// ExplainWorkflow provides a workflow explanation for accomplishing a goal.
func (a *HelpAgent) ExplainWorkflow(ctx context.Context, goal string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	helpCtx := &HelpContext{
		UserGoal:     goal,
		AvailableCmd: a.GetAvailableCommands(),
	}
	a.SetHelpContext(helpCtx)

	prompt := fmt.Sprintf("Explain the recommended workflow to accomplish this goal: %s", goal)
	return a.Query(ctx, prompt)
}
