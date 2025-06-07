package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// DoctorAgent handles system health checks and comprehensive diagnostics
type DoctorAgent struct {
	BaseAgent
}

// DoctorContext contains system health and diagnostic information
type DoctorContext struct {
	SystemHealth     string   `json:"system_health,omitempty"`
	ServiceStatus    string   `json:"service_status,omitempty"`
	StorageInfo      string   `json:"storage_info,omitempty"`
	MemoryInfo       string   `json:"memory_info,omitempty"`
	NetworkStatus    string   `json:"network_status,omitempty"`
	NixStoreHealth   string   `json:"nix_store_health,omitempty"`
	ChannelStatus    string   `json:"channel_status,omitempty"`
	GenerationInfo   string   `json:"generation_info,omitempty"`
	SystemErrors     []string `json:"system_errors,omitempty"`
	WarningMessages  []string `json:"warning_messages,omitempty"`
	ConfigValidation string   `json:"config_validation,omitempty"`
	HardwareInfo     string   `json:"hardware_info,omitempty"`
	PerformanceInfo  string   `json:"performance_info,omitempty"`
}

// NewDoctorAgent creates a new DoctorAgent with the Doctor role
func NewDoctorAgent(provider ai.Provider) *DoctorAgent {
	agent := &DoctorAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleDoctor,
		},
	}
	return agent
}

// Query handles system health questions and diagnostics
func (a *DoctorAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt, ok := roles.RolePromptTemplate[a.role]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", a.role)
	}

	// Build context-aware prompt
	fullPrompt := a.buildContextualPrompt(prompt, question)

	return a.provider.Query(ctx, fullPrompt)
}

// GenerateResponse handles system health response generation
func (a *DoctorAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add health-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// buildContextualPrompt constructs a context-aware prompt for doctor operations
func (a *DoctorAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add doctor context if available
	if a.contextData != nil {
		if doctorCtx, ok := a.contextData.(*DoctorContext); ok {
			contextStr := a.formatDoctorContext(doctorCtx)
			if contextStr != "" {
				promptParts = append(promptParts, "System Health Context:")
				promptParts = append(promptParts, contextStr)
			}
		}
	}

	// Add user input
	promptParts = append(promptParts, "Health Check Request:")
	promptParts = append(promptParts, userInput)

	return strings.Join(promptParts, "\n\n")
}

// formatDoctorContext formats DoctorContext into a readable string
func (a *DoctorAgent) formatDoctorContext(ctx *DoctorContext) string {
	var parts []string

	if ctx.SystemHealth != "" {
		parts = append(parts, fmt.Sprintf("System Health: %s", ctx.SystemHealth))
	}

	if ctx.NixStoreHealth != "" {
		parts = append(parts, fmt.Sprintf("Nix Store Health: %s", ctx.NixStoreHealth))
	}

	if ctx.ChannelStatus != "" {
		parts = append(parts, fmt.Sprintf("Channel Status: %s", ctx.ChannelStatus))
	}

	if ctx.GenerationInfo != "" {
		parts = append(parts, fmt.Sprintf("Generation Info: %s", ctx.GenerationInfo))
	}

	if ctx.ServiceStatus != "" {
		parts = append(parts, fmt.Sprintf("Service Status: %s", ctx.ServiceStatus))
	}

	if ctx.NetworkStatus != "" {
		parts = append(parts, fmt.Sprintf("Network Status: %s", ctx.NetworkStatus))
	}

	if ctx.StorageInfo != "" {
		parts = append(parts, fmt.Sprintf("Storage Info: %s", ctx.StorageInfo))
	}

	if ctx.MemoryInfo != "" {
		parts = append(parts, fmt.Sprintf("Memory Info: %s", ctx.MemoryInfo))
	}

	if ctx.HardwareInfo != "" {
		parts = append(parts, fmt.Sprintf("Hardware Info: %s", ctx.HardwareInfo))
	}

	if ctx.PerformanceInfo != "" {
		parts = append(parts, fmt.Sprintf("Performance Info: %s", ctx.PerformanceInfo))
	}

	if ctx.ConfigValidation != "" {
		parts = append(parts, fmt.Sprintf("Config Validation: %s", ctx.ConfigValidation))
	}

	if len(ctx.SystemErrors) > 0 {
		parts = append(parts, fmt.Sprintf("System Errors: %s", strings.Join(ctx.SystemErrors, ", ")))
	}

	if len(ctx.WarningMessages) > 0 {
		parts = append(parts, fmt.Sprintf("Warnings: %s", strings.Join(ctx.WarningMessages, ", ")))
	}

	return strings.Join(parts, "\n")
}

// SetDoctorContext is a convenience method to set DoctorContext
func (a *DoctorAgent) SetDoctorContext(ctx *DoctorContext) {
	a.SetContext(ctx)
}
