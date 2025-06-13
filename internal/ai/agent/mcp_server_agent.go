package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// McpServerAgent handles MCP (Model Context Protocol) server operations and management.
type McpServerAgent struct {
	provider    ai.Provider
	role        string
	contextData interface{}
}

// McpServerContext provides context for MCP server operations.
type McpServerContext struct {
	ServerConfig     map[string]interface{} // MCP server configuration
	ClientConfig     map[string]interface{} // MCP client configuration
	AvailableServers []string               // list of available MCP servers
	ServerStatus     map[string]string      // status of each server
	ErrorLogs        []string               // recent error logs
	PerformanceData  map[string]interface{} // performance metrics
	SecuritySettings map[string]interface{} // security configuration
	Metadata         map[string]string      // additional context
}

// NewMcpServerAgent creates a new MCP server agent.
func NewMcpServerAgent(provider ai.Provider) *McpServerAgent {
	return &McpServerAgent{
		provider: provider,
		role:     string(roles.RoleMcpServer),
	}
}

// Helper for all usages in this file:
func queryProviderWithContextOrFallback(provider ai.Provider, ctx context.Context, prompt string) (string, error) {
	if p, ok := provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		return p.QueryWithContext(ctx, prompt)
	}
	if p, ok := provider.(interface{ Query(string) (string, error) }); ok {
		return p.Query(prompt)
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// Query processes MCP server-related queries.
func (a *McpServerAgent) Query(ctx context.Context, input string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	// Build MCP server-specific prompt
	prompt := a.buildMcpServerPrompt(input)

	// Use provider to generate response
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// GenerateResponse generates a response for MCP server assistance.
func (a *McpServerAgent) GenerateResponse(ctx context.Context, input string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	// Enhance prompt with role and context
	prompt := a.enhancePromptWithRole(input)

	// Generate response using provider
	response, err := queryProviderWithContextOrFallback(a.provider, ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate MCP server response: %w", err)
	}

	return a.formatMcpServerResponse(response), nil
}

// SetRole sets the role for the agent.
func (a *McpServerAgent) SetRole(role string) {
	a.role = role
}

// SetContext sets the context for MCP server operations.
func (a *McpServerAgent) SetContext(context interface{}) error {
	if context == nil {
		a.contextData = nil
		return nil
	}

	if mcpCtx, ok := context.(*McpServerContext); ok {
		a.contextData = mcpCtx
		return nil
	}

	return fmt.Errorf("invalid context type for McpServerAgent")
}

// SetupMcpServer helps set up a new MCP server configuration.
func (a *McpServerAgent) SetupMcpServer(serverType string, requirements map[string]interface{}) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As an MCP Server Specialist, help set up an MCP server:
Server Type: %s
Requirements: %+v

%s

Provide complete setup instructions with configuration examples.`,
		serverType, requirements, a.formatMcpServerContext())

	ctx := context.Background()
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// DiagnoseMcpIssues diagnoses MCP server problems.
func (a *McpServerAgent) DiagnoseMcpIssues(issue string, symptoms []string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As an MCP Server Specialist, diagnose MCP server issues:
Issue: %s
Symptoms: %s

%s

Provide diagnosis and step-by-step troubleshooting instructions.`,
		issue, strings.Join(symptoms, ", "), a.formatMcpServerContext())

	ctx := context.Background()
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// OptimizeMcpPerformance provides MCP server performance optimization guidance.
func (a *McpServerAgent) OptimizeMcpPerformance(performanceGoals []string, currentMetrics map[string]interface{}) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As an MCP Server Specialist, optimize MCP server performance:
Performance Goals: %s
Current Metrics: %+v

%s

Provide specific optimization recommendations and implementation steps.`,
		strings.Join(performanceGoals, ", "), currentMetrics, a.formatMcpServerContext())

	ctx := context.Background()
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// ManageMcpSecurity handles MCP server security configuration.
func (a *McpServerAgent) ManageMcpSecurity(securityConcerns []string, currentConfig map[string]interface{}) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As an MCP Server Specialist, manage MCP server security:
Security Concerns: %s
Current Configuration: %+v

%s

Provide security recommendations and configuration updates.`,
		strings.Join(securityConcerns, ", "), currentConfig, a.formatMcpServerContext())

	ctx := context.Background()
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// IntegrateMcpServer helps integrate MCP server with applications.
func (a *McpServerAgent) IntegrateMcpServer(targetApp string, integrationRequirements []string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As an MCP Server Specialist, integrate MCP server with applications:
Target Application: %s
Integration Requirements: %s

%s

Provide integration guide with examples and best practices.`,
		targetApp, strings.Join(integrationRequirements, ", "), a.formatMcpServerContext())

	ctx := context.Background()
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// MonitorMcpServer provides monitoring and alerting guidance.
func (a *McpServerAgent) MonitorMcpServer(monitoringScope []string, alertingNeeds []string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As an MCP Server Specialist, set up MCP server monitoring:
Monitoring Scope: %s
Alerting Needs: %s

%s

Provide monitoring setup instructions and alerting configuration.`,
		strings.Join(monitoringScope, ", "), strings.Join(alertingNeeds, ", "), a.formatMcpServerContext())

	ctx := context.Background()
	return queryProviderWithContextOrFallback(a.provider, ctx, prompt)
}

// buildMcpServerPrompt constructs an MCP server-specific prompt.
func (a *McpServerAgent) buildMcpServerPrompt(input string) string {
	var prompt strings.Builder

	// Add role context
	if template, exists := roles.RolePromptTemplate[roles.RoleMcpServer]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	// Add MCP server context
	if a.contextData != nil {
		prompt.WriteString("MCP Server Context:\n")
		prompt.WriteString(a.formatMcpServerContext())
		prompt.WriteString("\n\n")
	}

	// Add user input
	prompt.WriteString("User Query: ")
	prompt.WriteString(input)

	return prompt.String()
}

// formatMcpServerContext formats the MCP server context for inclusion in prompts.
func (a *McpServerAgent) formatMcpServerContext() string {
	if a.contextData == nil {
		return "No specific MCP server context provided."
	}

	ctx, ok := a.contextData.(*McpServerContext)
	if !ok {
		return "Invalid MCP server context."
	}

	var context strings.Builder
	context.WriteString("MCP Server Environment:\n")

	if len(ctx.AvailableServers) > 0 {
		context.WriteString(fmt.Sprintf("- Available Servers: %s\n", strings.Join(ctx.AvailableServers, ", ")))
	}

	if len(ctx.ServerStatus) > 0 {
		context.WriteString("- Server Status:\n")
		for server, status := range ctx.ServerStatus {
			context.WriteString(fmt.Sprintf("  - %s: %s\n", server, status))
		}
	}

	if len(ctx.ErrorLogs) > 0 {
		context.WriteString("- Recent Errors:\n")
		for _, log := range ctx.ErrorLogs {
			context.WriteString(fmt.Sprintf("  - %s\n", log))
		}
	}

	if len(ctx.ServerConfig) > 0 {
		context.WriteString("- Server Configuration Available\n")
	}

	if len(ctx.SecuritySettings) > 0 {
		context.WriteString("- Security Settings Configured\n")
	}

	return context.String()
}

// enhancePromptWithRole enhances the prompt with role-specific information.
func (a *McpServerAgent) enhancePromptWithRole(input string) string {
	if template, exists := roles.RolePromptTemplate[roles.RoleMcpServer]; exists {
		return fmt.Sprintf("%s\n\nUser Request: %s", template, input)
	}
	return input
}

// formatMcpServerResponse formats the response for better readability.
func (a *McpServerAgent) formatMcpServerResponse(response string) string {
	// Add any MCP server-specific response formatting here
	return response
}
