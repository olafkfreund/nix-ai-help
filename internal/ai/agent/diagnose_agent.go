package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
)

// DiagnosticContext contains structured context for NixOS diagnostics
type DiagnosticContext struct {
	LogData             string             `json:"log_data,omitempty"`
	ConfigSnippet       string             `json:"config_snippet,omitempty"`
	ErrorMessage        string             `json:"error_message,omitempty"`
	SystemInfo          *SystemInfo        `json:"system_info,omitempty"`
	ExistingDiagnostics []nixos.Diagnostic `json:"existing_diagnostics,omitempty"`
	CommandOutput       string             `json:"command_output,omitempty"`
	UserDescription     string             `json:"user_description,omitempty"`
}

// SystemInfo contains relevant NixOS system information
type SystemInfo struct {
	NixVersion    string `json:"nix_version,omitempty"`
	NixOSVersion  string `json:"nixos_version,omitempty"`
	Channel       string `json:"channel,omitempty"`
	Generation    string `json:"generation,omitempty"`
	Architecture  string `json:"architecture,omitempty"`
	IsFlakeSystem bool   `json:"is_flake_system"`
}

// DiagnoseAgent is a specialized agent for the 'diagnose' command.
type DiagnoseAgent struct {
	role        string
	contextData interface{}
	logger      *logger.Logger
}

func NewDiagnoseAgent() *DiagnoseAgent {
	return &DiagnoseAgent{
		role:   string(roles.RoleDiagnose),
		logger: logger.NewLogger(),
	}
}

func (a *DiagnoseAgent) Query(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	if !roles.ValidateRole(role) {
		return "", fmt.Errorf("unsupported role: %s", role)
	}

	// Get role-specific prompt template
	prompt, ok := roles.RolePromptTemplate[roles.RoleType(role)]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", role)
	}

	// Build enhanced diagnostic prompt with context
	enhancedPrompt := a.buildDiagnosticPrompt(prompt, input, contextData)

	a.logger.Debug("DiagnoseAgent: Built enhanced diagnostic prompt")
	return enhancedPrompt, nil
}

func (a *DiagnoseAgent) GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	// For now, GenerateResponse behaves the same as Query
	// In the future, this could integrate with actual LLM providers
	return a.Query(ctx, input, role, contextData)
}

func (a *DiagnoseAgent) SetRole(role string) {
	a.role = role
}

func (a *DiagnoseAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}

// buildDiagnosticPrompt constructs a comprehensive diagnostic prompt with context
func (a *DiagnoseAgent) buildDiagnosticPrompt(basePrompt, input string, contextData interface{}) string {
	var prompt strings.Builder

	// Start with role-specific base prompt
	prompt.WriteString(basePrompt)
	prompt.WriteString("\n\n")

	// Add structured context if available
	if ctx, ok := contextData.(*DiagnosticContext); ok {
		prompt.WriteString("## DIAGNOSTIC CONTEXT\n\n")

		// System information
		if ctx.SystemInfo != nil {
			prompt.WriteString("### System Information\n")
			if ctx.SystemInfo.NixOSVersion != "" {
				prompt.WriteString(fmt.Sprintf("- NixOS Version: %s\n", ctx.SystemInfo.NixOSVersion))
			}
			if ctx.SystemInfo.NixVersion != "" {
				prompt.WriteString(fmt.Sprintf("- Nix Version: %s\n", ctx.SystemInfo.NixVersion))
			}
			if ctx.SystemInfo.Channel != "" {
				prompt.WriteString(fmt.Sprintf("- Channel: %s\n", ctx.SystemInfo.Channel))
			}
			if ctx.SystemInfo.Generation != "" {
				prompt.WriteString(fmt.Sprintf("- Generation: %s\n", ctx.SystemInfo.Generation))
			}
			if ctx.SystemInfo.Architecture != "" {
				prompt.WriteString(fmt.Sprintf("- Architecture: %s\n", ctx.SystemInfo.Architecture))
			}
			prompt.WriteString(fmt.Sprintf("- Flake System: %t\n", ctx.SystemInfo.IsFlakeSystem))
			prompt.WriteString("\n")
		}

		// Error information
		if ctx.ErrorMessage != "" {
			prompt.WriteString("### Error Message\n```\n")
			prompt.WriteString(ctx.ErrorMessage)
			prompt.WriteString("\n```\n\n")
		}

		// Log data
		if ctx.LogData != "" {
			prompt.WriteString("### Log Output\n```\n")
			prompt.WriteString(ctx.LogData)
			prompt.WriteString("\n```\n\n")
		}

		// Configuration snippet
		if ctx.ConfigSnippet != "" {
			prompt.WriteString("### Configuration Snippet\n```nix\n")
			prompt.WriteString(ctx.ConfigSnippet)
			prompt.WriteString("\n```\n\n")
		}

		// Command output
		if ctx.CommandOutput != "" {
			prompt.WriteString("### Command Output\n```\n")
			prompt.WriteString(ctx.CommandOutput)
			prompt.WriteString("\n```\n\n")
		}

		// User description
		if ctx.UserDescription != "" {
			prompt.WriteString("### User Description\n")
			prompt.WriteString(ctx.UserDescription)
			prompt.WriteString("\n\n")
		}

		// Existing diagnostics from automated analysis
		if len(ctx.ExistingDiagnostics) > 0 {
			prompt.WriteString("### Automated Analysis Results\n")
			prompt.WriteString("The following issues were automatically detected:\n\n")
			for i, diag := range ctx.ExistingDiagnostics {
				if diag.ErrorType != "info" && diag.ErrorType != "ai_analysis" {
					prompt.WriteString(fmt.Sprintf("%d. **%s** (Severity: %s)\n", i+1, diag.Issue, diag.Severity))
					prompt.WriteString(fmt.Sprintf("   - Type: %s\n", diag.ErrorType))
					if diag.Details != "" {
						prompt.WriteString(fmt.Sprintf("   - Details: %s\n", diag.Details))
					}
				}
			}
			prompt.WriteString("\n")
		}
	}

	// Add the main user input/question
	if input != "" {
		prompt.WriteString("## USER INPUT\n")
		prompt.WriteString(input)
		prompt.WriteString("\n\n")
	}

	// Add specific instructions for NixOS diagnosis
	prompt.WriteString("## INSTRUCTIONS\n")
	prompt.WriteString("Please provide a comprehensive diagnosis with:\n")
	prompt.WriteString("1. **Problem Summary**: Clear description of the issue\n")
	prompt.WriteString("2. **Root Cause Analysis**: Technical explanation of why this occurred\n")
	prompt.WriteString("3. **Step-by-Step Fix**: Numbered, actionable steps to resolve the issue\n")
	prompt.WriteString("4. **Verification**: How to confirm the fix worked\n")
	prompt.WriteString("5. **Prevention**: How to avoid this issue in the future\n")
	prompt.WriteString("6. **Additional Resources**: Relevant NixOS documentation or community resources\n\n")
	prompt.WriteString("Focus on NixOS-specific solutions and leverage your knowledge of common patterns and troubleshooting techniques.")

	return prompt.String()
}

// BuildDiagnosticContext is a helper function to create DiagnosticContext from various inputs
func BuildDiagnosticContext(logData, configSnippet, errorMessage, userDescription string, existingDiagnostics []nixos.Diagnostic) *DiagnosticContext {
	return &DiagnosticContext{
		LogData:             logData,
		ConfigSnippet:       configSnippet,
		ErrorMessage:        errorMessage,
		UserDescription:     userDescription,
		ExistingDiagnostics: existingDiagnostics,
	}
}

// AddSystemInfo adds system information to the diagnostic context
func (dc *DiagnosticContext) AddSystemInfo(info *SystemInfo) {
	dc.SystemInfo = info
}

// AddCommandOutput adds command output to the diagnostic context
func (dc *DiagnosticContext) AddCommandOutput(output string) {
	dc.CommandOutput = output
}
