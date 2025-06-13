package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// HardwareAgent handles hardware configuration and optimization operations
type HardwareAgent struct {
	BaseAgent
}

// HardwareContext contains hardware-specific context information
type HardwareContext struct {
	SystemInfo          string            `json:"system_info,omitempty"`          // System hardware overview
	CPUInfo             string            `json:"cpu_info,omitempty"`             // CPU details and specifications
	GPUInfo             string            `json:"gpu_info,omitempty"`             // Graphics card information
	MemoryInfo          string            `json:"memory_info,omitempty"`          // RAM and memory details
	StorageInfo         string            `json:"storage_info,omitempty"`         // Storage devices and configuration
	NetworkInfo         string            `json:"network_info,omitempty"`         // Network interfaces and adapters
	AudioInfo           string            `json:"audio_info,omitempty"`           // Audio hardware and configuration
	USBDevices          []string          `json:"usb_devices,omitempty"`          // Connected USB devices
	PCIDevices          []string          `json:"pci_devices,omitempty"`          // PCI devices and controllers
	KernelModules       []string          `json:"kernel_modules,omitempty"`       // Loaded kernel modules
	LoadedDrivers       []string          `json:"loaded_drivers,omitempty"`       // Currently loaded drivers
	MissingDrivers      []string          `json:"missing_drivers,omitempty"`      // Missing or problematic drivers
	HardwareIssues      []string          `json:"hardware_issues,omitempty"`      // Known hardware problems
	PowerManagement     string            `json:"power_management,omitempty"`     // Power management configuration
	ThermalInfo         string            `json:"thermal_info,omitempty"`         // Temperature and thermal management
	BIOS_UEFI           string            `json:"bios_uefi,omitempty"`            // BIOS/UEFI information
	SecureBoot          bool              `json:"secure_boot,omitempty"`          // Secure boot status
	VirtualizationInfo  string            `json:"virtualization_info,omitempty"`  // Virtualization support
	Architecture        string            `json:"architecture,omitempty"`         // System architecture
	Microcode           string            `json:"microcode,omitempty"`            // CPU microcode information
	Firmware            []string          `json:"firmware,omitempty"`             // Firmware packages and versions
	HardwareConfig      map[string]string `json:"hardware_config,omitempty"`      // Current hardware configuration
	OptimizationGoals   []string          `json:"optimization_goals,omitempty"`   // Performance optimization targets
	CompatibilityIssues []string          `json:"compatibility_issues,omitempty"` // Known compatibility problems
	RecommendedPackages []string          `json:"recommended_packages,omitempty"` // Suggested hardware packages
	ConfigOptions       map[string]string `json:"config_options,omitempty"`       // Suggested NixOS configuration options
}

// NewHardwareAgent creates a new hardware agent with the specified provider.
func NewHardwareAgent(provider ai.Provider) *HardwareAgent {
	agent := &HardwareAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleHardware,
		},
	}
	return agent
}

// Query handles hardware-related queries using the provider.
func (a *HardwareAgent) Query(ctx context.Context, question string) (string, error) {
	if a.role == "" {
		return "", fmt.Errorf("role not set for HardwareAgent")
	}

	// Build enhanced prompt with hardware context
	prompt := a.buildContextualPrompt(question)

	// Use provider to generate response
	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("hardware agent query failed: %w", err)
		}
		return a.enhanceResponseWithHardwareGuidance(response), nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("hardware agent query failed: %w", err)
		}
		return a.enhanceResponseWithHardwareGuidance(response), nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateResponse handles hardware-specific response generation.
func (a *HardwareAgent) GenerateResponse(ctx context.Context, input string) (string, error) {
	if a.role == "" {
		return "", fmt.Errorf("role not set for HardwareAgent")
	}

	// Build enhanced prompt with hardware context and role
	prompt := a.buildContextualPrompt(input)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	// Use provider to generate response
	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("hardware agent response generation failed: %w", err)
	}

	return a.enhanceResponseWithHardwareGuidance(response), nil
}

// buildContextualPrompt creates a comprehensive prompt with hardware context.
func (a *HardwareAgent) buildContextualPrompt(input string) string {
	prompt := fmt.Sprintf("Hardware Query: %s\n\n", input)

	// Add hardware context if available
	if a.contextData != nil {
		if hwCtx, ok := a.contextData.(*HardwareContext); ok {
			prompt += a.buildHardwareContextSection(hwCtx)
		}
	}

	return prompt
}

// buildHardwareContextSection creates a formatted context section for hardware information.
func (a *HardwareAgent) buildHardwareContextSection(ctx *HardwareContext) string {
	var contextStr string

	if ctx.SystemInfo != "" {
		contextStr += "## System Overview\n"
		contextStr += fmt.Sprintf("- System Info: %s\n", ctx.SystemInfo)
		if ctx.Architecture != "" {
			contextStr += fmt.Sprintf("- Architecture: %s\n", ctx.Architecture)
		}
		contextStr += "\n"
	}

	if ctx.CPUInfo != "" || ctx.Microcode != "" {
		contextStr += "## CPU Information\n"
		if ctx.CPUInfo != "" {
			contextStr += fmt.Sprintf("- CPU: %s\n", ctx.CPUInfo)
		}
		if ctx.Microcode != "" {
			contextStr += fmt.Sprintf("- Microcode: %s\n", ctx.Microcode)
		}
		contextStr += "\n"
	}

	if ctx.GPUInfo != "" {
		contextStr += "## Graphics\n"
		contextStr += fmt.Sprintf("- GPU: %s\n", ctx.GPUInfo)
		contextStr += "\n"
	}

	if ctx.MemoryInfo != "" {
		contextStr += "## Memory\n"
		contextStr += fmt.Sprintf("- Memory: %s\n", ctx.MemoryInfo)
		contextStr += "\n"
	}

	if ctx.StorageInfo != "" {
		contextStr += "## Storage\n"
		contextStr += fmt.Sprintf("- Storage: %s\n", ctx.StorageInfo)
		contextStr += "\n"
	}

	if ctx.NetworkInfo != "" || ctx.AudioInfo != "" {
		contextStr += "## Connectivity\n"
		if ctx.NetworkInfo != "" {
			contextStr += fmt.Sprintf("- Network: %s\n", ctx.NetworkInfo)
		}
		if ctx.AudioInfo != "" {
			contextStr += fmt.Sprintf("- Audio: %s\n", ctx.AudioInfo)
		}
		contextStr += "\n"
	}

	if len(ctx.USBDevices) > 0 {
		contextStr += "## USB Devices\n"
		for _, device := range ctx.USBDevices {
			contextStr += fmt.Sprintf("- %s\n", device)
		}
		contextStr += "\n"
	}

	if len(ctx.PCIDevices) > 0 {
		contextStr += "## PCI Devices\n"
		for _, device := range ctx.PCIDevices {
			contextStr += fmt.Sprintf("- %s\n", device)
		}
		contextStr += "\n"
	}

	if len(ctx.KernelModules) > 0 {
		contextStr += "## Loaded Kernel Modules\n"
		for _, module := range ctx.KernelModules {
			contextStr += fmt.Sprintf("- %s\n", module)
		}
		contextStr += "\n"
	}

	if len(ctx.LoadedDrivers) > 0 {
		contextStr += "## Loaded Drivers\n"
		for _, driver := range ctx.LoadedDrivers {
			contextStr += fmt.Sprintf("- %s\n", driver)
		}
		contextStr += "\n"
	}

	if len(ctx.MissingDrivers) > 0 {
		contextStr += "## Missing Drivers\n"
		for _, driver := range ctx.MissingDrivers {
			contextStr += fmt.Sprintf("- %s\n", driver)
		}
		contextStr += "\n"
	}

	if len(ctx.HardwareIssues) > 0 {
		contextStr += "## Hardware Issues\n"
		for _, issue := range ctx.HardwareIssues {
			contextStr += fmt.Sprintf("- %s\n", issue)
		}
		contextStr += "\n"
	}

	if ctx.PowerManagement != "" || ctx.ThermalInfo != "" {
		contextStr += "## Power & Thermal\n"
		if ctx.PowerManagement != "" {
			contextStr += fmt.Sprintf("- Power Management: %s\n", ctx.PowerManagement)
		}
		if ctx.ThermalInfo != "" {
			contextStr += fmt.Sprintf("- Thermal: %s\n", ctx.ThermalInfo)
		}
		contextStr += "\n"
	}

	if ctx.BIOS_UEFI != "" || ctx.SecureBoot {
		contextStr += "## Firmware\n"
		if ctx.BIOS_UEFI != "" {
			contextStr += fmt.Sprintf("- BIOS/UEFI: %s\n", ctx.BIOS_UEFI)
		}
		contextStr += fmt.Sprintf("- Secure Boot: %t\n", ctx.SecureBoot)
		contextStr += "\n"
	}

	if len(ctx.Firmware) > 0 {
		contextStr += "## Firmware Packages\n"
		for _, fw := range ctx.Firmware {
			contextStr += fmt.Sprintf("- %s\n", fw)
		}
		contextStr += "\n"
	}

	if ctx.VirtualizationInfo != "" {
		contextStr += "## Virtualization\n"
		contextStr += fmt.Sprintf("- Virtualization: %s\n", ctx.VirtualizationInfo)
		contextStr += "\n"
	}

	if len(ctx.HardwareConfig) > 0 {
		contextStr += "## Current Hardware Configuration\n"
		for key, value := range ctx.HardwareConfig {
			contextStr += fmt.Sprintf("- %s: %s\n", key, value)
		}
		contextStr += "\n"
	}

	if len(ctx.OptimizationGoals) > 0 {
		contextStr += "## Optimization Goals\n"
		for _, goal := range ctx.OptimizationGoals {
			contextStr += fmt.Sprintf("- %s\n", goal)
		}
		contextStr += "\n"
	}

	if len(ctx.CompatibilityIssues) > 0 {
		contextStr += "## Compatibility Issues\n"
		for _, issue := range ctx.CompatibilityIssues {
			contextStr += fmt.Sprintf("- %s\n", issue)
		}
		contextStr += "\n"
	}

	if len(ctx.RecommendedPackages) > 0 {
		contextStr += "## Recommended Packages\n"
		for _, pkg := range ctx.RecommendedPackages {
			contextStr += fmt.Sprintf("- %s\n", pkg)
		}
		contextStr += "\n"
	}

	if len(ctx.ConfigOptions) > 0 {
		contextStr += "## Suggested Configuration Options\n"
		for option, value := range ctx.ConfigOptions {
			contextStr += fmt.Sprintf("- %s = %s\n", option, value)
		}
		contextStr += "\n"
	}

	return contextStr
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *HardwareAgent) enhancePromptWithRole(prompt string) string {
	rolePrompt := roles.RolePromptTemplate[a.role]
	return fmt.Sprintf("%s\n\n%s", rolePrompt, prompt)
}

// enhanceResponseWithHardwareGuidance adds hardware-specific guidance to responses.
func (a *HardwareAgent) enhanceResponseWithHardwareGuidance(response string) string {
	guidance := "\n\n---\n**Hardware Configuration Tips:**\n"
	guidance += "- Test hardware changes in a safe environment first\n"
	guidance += "- Keep backups of working configurations\n"
	guidance += "- Check hardware compatibility before major changes\n"
	guidance += "- Monitor system stability after hardware configuration changes\n"
	guidance += "- Use `nixos-generate-config` to detect new hardware automatically\n"

	return response + guidance
}
