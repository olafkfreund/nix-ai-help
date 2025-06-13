package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MachineContext represents context for machine management operations.
type MachineContext struct {
	// Machine information
	MachineName  string `json:"machine_name,omitempty"`
	HostName     string `json:"host_name,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`

	// Network and connectivity
	IPAddress     string `json:"ip_address,omitempty"`
	SSHUser       string `json:"ssh_user,omitempty"`
	SSHPort       int    `json:"ssh_port,omitempty"`
	NetworkStatus string `json:"network_status,omitempty"`

	// Configuration context
	FlakePath    string `json:"flake_path,omitempty"`
	ConfigHash   string `json:"config_hash,omitempty"`
	Generation   int    `json:"generation,omitempty"`
	DeployMethod string `json:"deploy_method,omitempty"`

	// Machine group and management
	MachineGroup string   `json:"machine_group,omitempty"`
	MachineRole  string   `json:"machine_role,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`

	// Deployment status
	LastDeploy   string   `json:"last_deploy,omitempty"`
	DeployStatus string   `json:"deploy_status,omitempty"`
	HealthStatus string   `json:"health_status,omitempty"`
	Issues       []string `json:"issues,omitempty"`

	// Performance and resources
	CPUUsage    string `json:"cpu_usage,omitempty"`
	MemoryUsage string `json:"memory_usage,omitempty"`
	DiskUsage   string `json:"disk_usage,omitempty"`
	LoadAverage string `json:"load_average,omitempty"`

	// Operation context
	OperationType string `json:"operation_type,omitempty"`
}

// MachinesAgent represents an agent specialized in multi-machine management.
type MachinesAgent struct {
	BaseAgent
	context *MachineContext
}

// NewMachinesAgent creates a new MachinesAgent instance.
func NewMachinesAgent(provider ai.Provider) *MachinesAgent {
	agent := &MachinesAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleMachines,
		},
		context: &MachineContext{},
	}
	return agent
}

// SetContext sets the machine context for operations.
func (a *MachinesAgent) SetContext(ctx *MachineContext) {
	a.context = ctx
}

// GetContext returns the current machine context.
func (a *MachinesAgent) GetContext() *MachineContext {
	return a.context
}

// Query handles machine management questions and operations.
func (a *MachinesAgent) Query(ctx context.Context, question string) (string, error) {
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

// GenerateResponse handles machine management response generation.
func (a *MachinesAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add machine-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// PlanDeployment provides intelligent deployment planning and strategy.
func (a *MachinesAgent) PlanDeployment(ctx context.Context, machines []string, deployMethod string) (string, error) {
	a.context.DeployMethod = deployMethod
	a.context.OperationType = "deployment_planning"

	prompt := a.buildPrompt("Plan and optimize deployment strategy for multiple machines", map[string]interface{}{
		"target_machines": machines,
		"deploy_method":   deployMethod,
		"operation":       "deployment planning",
		"include":         "dependency order, rollback strategy, health checks, monitoring",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to plan deployment: %w", err)
	}

	return a.formatMachineResponse(response, "Deployment Planning"), nil
}

// DiagnoseDeploymentIssues diagnoses deployment problems across machines.
func (a *MachinesAgent) DiagnoseDeploymentIssues(ctx context.Context, machineName string, issues []string) (string, error) {
	a.context.MachineName = machineName
	a.context.Issues = issues
	a.context.OperationType = "deployment_diagnosis"

	prompt := a.buildPrompt("Diagnose deployment issues and provide resolution strategies", map[string]interface{}{
		"machine_name": machineName,
		"issues":       issues,
		"operation":    "deployment troubleshooting",
		"include":      "root cause analysis, fix strategies, prevention measures",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to diagnose deployment issues: %w", err)
	}

	return a.formatMachineResponse(response, "Deployment Issue Diagnosis"), nil
}

// OptimizeMachineConfiguration provides configuration optimization recommendations.
func (a *MachinesAgent) OptimizeMachineConfiguration(ctx context.Context, machineName, machineRole string, requirements []string) (string, error) {
	a.context.MachineName = machineName
	a.context.MachineRole = machineRole
	a.context.OperationType = "configuration_optimization"

	prompt := a.buildPrompt("Optimize machine configuration for specific role and requirements", map[string]interface{}{
		"machine_name": machineName,
		"machine_role": machineRole,
		"requirements": requirements,
		"operation":    "configuration optimization",
		"include":      "performance tuning, resource allocation, service configuration, security hardening",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to optimize machine configuration: %w", err)
	}

	return a.formatMachineResponse(response, "Machine Configuration Optimization"), nil
}

// MonitorMachineHealth provides health monitoring and alerting recommendations.
func (a *MachinesAgent) MonitorMachineHealth(ctx context.Context, machines []string, healthMetrics []string) (string, error) {
	a.context.OperationType = "health_monitoring"

	prompt := a.buildPrompt("Design health monitoring and alerting strategy for machine fleet", map[string]interface{}{
		"machines":       machines,
		"health_metrics": healthMetrics,
		"operation":      "health monitoring setup",
		"include":        "monitoring tools, alert thresholds, dashboards, automated responses",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to design health monitoring: %w", err)
	}

	return a.formatMachineResponse(response, "Machine Health Monitoring"), nil
}

// ManageFlakeMigration provides guidance for flake-based multi-machine setups.
func (a *MachinesAgent) ManageFlakeMigration(ctx context.Context, flakePath string, machines []string) (string, error) {
	a.context.FlakePath = flakePath
	a.context.OperationType = "flake_migration"

	prompt := a.buildPrompt("Guide migration to flake-based multi-machine configuration", map[string]interface{}{
		"flake_path": flakePath,
		"machines":   machines,
		"operation":  "flake migration planning",
		"include":    "flake structure, machine configurations, deployment setup, testing strategy",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to plan flake migration: %w", err)
	}

	return a.formatMachineResponse(response, "Flake Migration Planning"), nil
}

// SetupDeployRs provides guidance for deploy-rs configuration and setup.
func (a *MachinesAgent) SetupDeployRs(ctx context.Context, hosts []string, interactive bool) (string, error) {
	a.context.DeployMethod = "deploy-rs"
	a.context.OperationType = "deploy_rs_setup"

	prompt := a.buildPrompt("Configure deploy-rs for multi-machine NixOS deployment", map[string]interface{}{
		"hosts":       hosts,
		"interactive": interactive,
		"operation":   "deploy-rs configuration",
		"include":     "flake integration, host configuration, SSH setup, deployment profiles",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to setup deploy-rs: %w", err)
	}

	return a.formatMachineResponse(response, "Deploy-rs Configuration"), nil
}

// buildContextualPrompt creates a context-aware prompt with machine information.
func (a *MachinesAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add machine context if available
	if a.context != nil {
		contextStr := a.formatMachineContext(a.context)
		if contextStr != "" {
			promptParts = append(promptParts, "Machine Context:")
			promptParts = append(promptParts, contextStr)
		}
	}

	// Add user input
	if userInput != "" {
		promptParts = append(promptParts, "User Request:")
		promptParts = append(promptParts, userInput)
	}

	return strings.Join(promptParts, "\n\n")
}

// formatMachineContext formats machine context for prompt inclusion.
func (a *MachinesAgent) formatMachineContext(ctx *MachineContext) string {
	var context strings.Builder

	if ctx.MachineName != "" {
		context.WriteString(fmt.Sprintf("- Machine: %s\n", ctx.MachineName))
	}
	if ctx.HostName != "" {
		context.WriteString(fmt.Sprintf("- Hostname: %s\n", ctx.HostName))
	}
	if ctx.Architecture != "" {
		context.WriteString(fmt.Sprintf("- Architecture: %s\n", ctx.Architecture))
	}
	if ctx.MachineRole != "" {
		context.WriteString(fmt.Sprintf("- Role: %s\n", ctx.MachineRole))
	}
	if ctx.DeployMethod != "" {
		context.WriteString(fmt.Sprintf("- Deploy Method: %s\n", ctx.DeployMethod))
	}
	if ctx.DeployStatus != "" {
		context.WriteString(fmt.Sprintf("- Deploy Status: %s\n", ctx.DeployStatus))
	}
	if ctx.HealthStatus != "" {
		context.WriteString(fmt.Sprintf("- Health Status: %s\n", ctx.HealthStatus))
	}
	if len(ctx.Issues) > 0 {
		context.WriteString(fmt.Sprintf("- Issues: %v\n", ctx.Issues))
	}

	return context.String()
}

// buildPrompt creates a specialized prompt for machine operations.
func (a *MachinesAgent) buildPrompt(task string, details map[string]interface{}) string {
	var prompt strings.Builder

	// Add role-specific context
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	// Add task description
	prompt.WriteString(fmt.Sprintf("**Task**: %s\n\n", task))

	// Add machine context
	prompt.WriteString("**Machine Context**:\n")
	if a.context.MachineName != "" {
		prompt.WriteString(fmt.Sprintf("- Machine Name: %s\n", a.context.MachineName))
	}
	if a.context.MachineRole != "" {
		prompt.WriteString(fmt.Sprintf("- Machine Role: %s\n", a.context.MachineRole))
	}
	if a.context.DeployMethod != "" {
		prompt.WriteString(fmt.Sprintf("- Deployment Method: %s\n", a.context.DeployMethod))
	}
	if a.context.HealthStatus != "" {
		prompt.WriteString(fmt.Sprintf("- Health Status: %s\n", a.context.HealthStatus))
	}
	if len(a.context.Issues) > 0 {
		prompt.WriteString(fmt.Sprintf("- Known Issues: %v\n", a.context.Issues))
	}

	// Add specific task details
	if len(details) > 0 {
		prompt.WriteString("\n**Operation Details**:\n")
		for key, value := range details {
			prompt.WriteString(fmt.Sprintf("- %s: %v\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value))
		}
	}

	// Add requirements
	prompt.WriteString("\n**Requirements**:\n")
	prompt.WriteString("- Provide specific commands and configuration examples\n")
	prompt.WriteString("- Include safety measures and rollback strategies\n")
	prompt.WriteString("- Consider network connectivity and SSH access\n")
	prompt.WriteString("- Plan for service dependencies and startup order\n")
	prompt.WriteString("- Include monitoring and health check recommendations\n")
	prompt.WriteString("- Ensure reproducible and declarative configurations\n")

	return prompt.String()
}

// formatMachineResponse formats the AI response for machine management operations.
func (a *MachinesAgent) formatMachineResponse(response, operation string) string {
	var formatted strings.Builder

	formatted.WriteString(fmt.Sprintf("# %s\n\n", operation))
	formatted.WriteString(response)

	// Add context-specific footer
	formatted.WriteString("\n\n---\n")
	formatted.WriteString("**🔧 Machine Management Best Practices**:\n")
	formatted.WriteString("- Test deployments in staging environment first\n")
	formatted.WriteString("- Maintain rollback strategies for all deployments\n")
	formatted.WriteString("- Monitor machine health and resource usage\n")
	formatted.WriteString("- Keep machine configurations in version control\n")
	formatted.WriteString("- Document machine roles and dependencies\n")
	formatted.WriteString("- Use automated deployment tools for consistency\n")
	formatted.WriteString("- Implement proper security and access controls\n")

	return formatted.String()
}
