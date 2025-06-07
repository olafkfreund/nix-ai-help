package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// LogsAgent handles log analysis and system monitoring operations
type LogsAgent struct {
	BaseAgent
}

// LogsContext contains log analysis and monitoring context information
type LogsContext struct {
	LogSources        []string          `json:"log_sources,omitempty"`        // Available log sources
	LogFiles          []string          `json:"log_files,omitempty"`          // Specific log files to analyze
	LogContent        string            `json:"log_content,omitempty"`        // Raw log content
	SystemdJournal    string            `json:"systemd_journal,omitempty"`    // Systemd journal entries
	ServiceLogs       map[string]string `json:"service_logs,omitempty"`       // Service-specific logs
	ErrorMessages     []string          `json:"error_messages,omitempty"`     // Extracted error messages
	WarningMessages   []string          `json:"warning_messages,omitempty"`   // Warning messages
	CriticalMessages  []string          `json:"critical_messages,omitempty"`  // Critical system messages
	LogLevel          string            `json:"log_level,omitempty"`          // Current log level filter
	TimeRange         string            `json:"time_range,omitempty"`         // Time range for log analysis
	ServiceNames      []string          `json:"service_names,omitempty"`      // Services to monitor
	LogPatterns       []string          `json:"log_patterns,omitempty"`       // Patterns to search for
	KernelMessages    string            `json:"kernel_messages,omitempty"`    // Kernel log messages
	BootMessages      string            `json:"boot_messages,omitempty"`      // Boot-time log messages
	NetworkLogs       string            `json:"network_logs,omitempty"`       // Network-related logs
	AuthLogs          string            `json:"auth_logs,omitempty"`          // Authentication logs
	SystemErrors      []string          `json:"system_errors,omitempty"`      // System-level errors
	PerformanceIssues []string          `json:"performance_issues,omitempty"` // Performance-related log entries
	SecurityEvents    []string          `json:"security_events,omitempty"`    // Security-related events
	LogRotation       string            `json:"log_rotation,omitempty"`       // Log rotation configuration
	LogSize           string            `json:"log_size,omitempty"`           // Current log size
	LogRetention      string            `json:"log_retention,omitempty"`      // Log retention policy
	MonitoringAlerts  []string          `json:"monitoring_alerts,omitempty"`  // Active monitoring alerts
	LogAnalysisGoals  []string          `json:"log_analysis_goals,omitempty"` // Analysis objectives
	FilterCriteria    map[string]string `json:"filter_criteria,omitempty"`    // Log filtering criteria
	OutputFormat      string            `json:"output_format,omitempty"`      // Desired output format
}

// NewLogsAgent creates a new logs agent with the specified provider.
func NewLogsAgent(provider ai.Provider) *LogsAgent {
	agent := &LogsAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleLogs,
		},
	}
	return agent
}

// Query handles log-related queries using the provider.
func (a *LogsAgent) Query(ctx context.Context, question string) (string, error) {
	if a.role == "" {
		return "", fmt.Errorf("role not set for LogsAgent")
	}

	// Build enhanced prompt with logs context
	prompt := a.buildContextualPrompt(question)

	// Use provider to generate response
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("logs agent query failed: %w", err)
	}

	return a.enhanceResponseWithLogsGuidance(response), nil
}

// GenerateResponse handles logs-specific response generation.
func (a *LogsAgent) GenerateResponse(ctx context.Context, input string) (string, error) {
	if a.role == "" {
		return "", fmt.Errorf("role not set for LogsAgent")
	}

	// Build enhanced prompt with logs context and role
	prompt := a.buildContextualPrompt(input)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	// Use provider to generate response
	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("logs agent response generation failed: %w", err)
	}

	return a.enhanceResponseWithLogsGuidance(response), nil
}

// buildContextualPrompt creates a comprehensive prompt with logs context.
func (a *LogsAgent) buildContextualPrompt(input string) string {
	prompt := fmt.Sprintf("Logs Query: %s\n\n", input)

	// Add logs context if available
	if a.contextData != nil {
		if logsCtx, ok := a.contextData.(*LogsContext); ok {
			prompt += a.buildLogsContextSection(logsCtx)
		}
	}

	return prompt
}

// buildLogsContextSection creates a formatted context section for logs information.
func (a *LogsAgent) buildLogsContextSection(ctx *LogsContext) string {
	var contextStr string

	if len(ctx.LogSources) > 0 {
		contextStr += "## Available Log Sources\n"
		for _, source := range ctx.LogSources {
			contextStr += fmt.Sprintf("- %s\n", source)
		}
		contextStr += "\n"
	}

	if len(ctx.LogFiles) > 0 {
		contextStr += "## Log Files\n"
		for _, file := range ctx.LogFiles {
			contextStr += fmt.Sprintf("- %s\n", file)
		}
		contextStr += "\n"
	}

	if ctx.LogContent != "" {
		contextStr += "## Log Content\n"
		contextStr += "```\n"
		contextStr += ctx.LogContent
		contextStr += "\n```\n\n"
	}

	if ctx.SystemdJournal != "" {
		contextStr += "## Systemd Journal\n"
		contextStr += "```\n"
		contextStr += ctx.SystemdJournal
		contextStr += "\n```\n\n"
	}

	if len(ctx.ServiceLogs) > 0 {
		contextStr += "## Service Logs\n"
		for service, logs := range ctx.ServiceLogs {
			contextStr += fmt.Sprintf("### %s\n", service)
			contextStr += "```\n"
			contextStr += logs
			contextStr += "\n```\n\n"
		}
	}

	if len(ctx.ErrorMessages) > 0 {
		contextStr += "## Error Messages\n"
		for _, msg := range ctx.ErrorMessages {
			contextStr += fmt.Sprintf("- %s\n", msg)
		}
		contextStr += "\n"
	}

	if len(ctx.WarningMessages) > 0 {
		contextStr += "## Warning Messages\n"
		for _, msg := range ctx.WarningMessages {
			contextStr += fmt.Sprintf("- %s\n", msg)
		}
		contextStr += "\n"
	}

	if len(ctx.CriticalMessages) > 0 {
		contextStr += "## Critical Messages\n"
		for _, msg := range ctx.CriticalMessages {
			contextStr += fmt.Sprintf("- %s\n", msg)
		}
		contextStr += "\n"
	}

	if ctx.LogLevel != "" || ctx.TimeRange != "" {
		contextStr += "## Log Analysis Parameters\n"
		if ctx.LogLevel != "" {
			contextStr += fmt.Sprintf("- Log Level: %s\n", ctx.LogLevel)
		}
		if ctx.TimeRange != "" {
			contextStr += fmt.Sprintf("- Time Range: %s\n", ctx.TimeRange)
		}
		if ctx.LogSize != "" {
			contextStr += fmt.Sprintf("- Log Size: %s\n", ctx.LogSize)
		}
		contextStr += "\n"
	}

	if len(ctx.ServiceNames) > 0 {
		contextStr += "## Services to Monitor\n"
		for _, service := range ctx.ServiceNames {
			contextStr += fmt.Sprintf("- %s\n", service)
		}
		contextStr += "\n"
	}

	if len(ctx.LogPatterns) > 0 {
		contextStr += "## Search Patterns\n"
		for _, pattern := range ctx.LogPatterns {
			contextStr += fmt.Sprintf("- %s\n", pattern)
		}
		contextStr += "\n"
	}

	if ctx.KernelMessages != "" {
		contextStr += "## Kernel Messages\n"
		contextStr += "```\n"
		contextStr += ctx.KernelMessages
		contextStr += "\n```\n\n"
	}

	if ctx.BootMessages != "" {
		contextStr += "## Boot Messages\n"
		contextStr += "```\n"
		contextStr += ctx.BootMessages
		contextStr += "\n```\n\n"
	}

	if ctx.NetworkLogs != "" {
		contextStr += "## Network Logs\n"
		contextStr += "```\n"
		contextStr += ctx.NetworkLogs
		contextStr += "\n```\n\n"
	}

	if ctx.AuthLogs != "" {
		contextStr += "## Authentication Logs\n"
		contextStr += "```\n"
		contextStr += ctx.AuthLogs
		contextStr += "\n```\n\n"
	}

	if len(ctx.SystemErrors) > 0 {
		contextStr += "## System Errors\n"
		for _, err := range ctx.SystemErrors {
			contextStr += fmt.Sprintf("- %s\n", err)
		}
		contextStr += "\n"
	}

	if len(ctx.PerformanceIssues) > 0 {
		contextStr += "## Performance Issues\n"
		for _, issue := range ctx.PerformanceIssues {
			contextStr += fmt.Sprintf("- %s\n", issue)
		}
		contextStr += "\n"
	}

	if len(ctx.SecurityEvents) > 0 {
		contextStr += "## Security Events\n"
		for _, event := range ctx.SecurityEvents {
			contextStr += fmt.Sprintf("- %s\n", event)
		}
		contextStr += "\n"
	}

	if ctx.LogRotation != "" || ctx.LogRetention != "" {
		contextStr += "## Log Management\n"
		if ctx.LogRotation != "" {
			contextStr += fmt.Sprintf("- Log Rotation: %s\n", ctx.LogRotation)
		}
		if ctx.LogRetention != "" {
			contextStr += fmt.Sprintf("- Log Retention: %s\n", ctx.LogRetention)
		}
		contextStr += "\n"
	}

	if len(ctx.MonitoringAlerts) > 0 {
		contextStr += "## Active Monitoring Alerts\n"
		for _, alert := range ctx.MonitoringAlerts {
			contextStr += fmt.Sprintf("- %s\n", alert)
		}
		contextStr += "\n"
	}

	if len(ctx.LogAnalysisGoals) > 0 {
		contextStr += "## Analysis Goals\n"
		for _, goal := range ctx.LogAnalysisGoals {
			contextStr += fmt.Sprintf("- %s\n", goal)
		}
		contextStr += "\n"
	}

	if len(ctx.FilterCriteria) > 0 {
		contextStr += "## Filter Criteria\n"
		for key, value := range ctx.FilterCriteria {
			contextStr += fmt.Sprintf("- %s: %s\n", key, value)
		}
		contextStr += "\n"
	}

	if ctx.OutputFormat != "" {
		contextStr += "## Output Format\n"
		contextStr += fmt.Sprintf("- Format: %s\n", ctx.OutputFormat)
		contextStr += "\n"
	}

	return contextStr
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *LogsAgent) enhancePromptWithRole(prompt string) string {
	rolePrompt := roles.RolePromptTemplate[a.role]
	return fmt.Sprintf("%s\n\n%s", rolePrompt, prompt)
}

// enhanceResponseWithLogsGuidance adds logs-specific guidance to responses.
func (a *LogsAgent) enhanceResponseWithLogsGuidance(response string) string {
	guidance := "\n\n---\n**Log Analysis Tips:**\n"
	guidance += "- Use `journalctl` for systemd service logs\n"
	guidance += "- Filter logs by time with `--since` and `--until`\n"
	guidance += "- Use `grep` and `awk` for pattern matching\n"
	guidance += "- Monitor logs in real-time with `journalctl -f`\n"
	guidance += "- Check log rotation and retention policies\n"
	guidance += "- Consider log aggregation tools for complex analysis\n"

	return response + guidance
}
