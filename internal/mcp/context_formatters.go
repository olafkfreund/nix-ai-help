package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"
)

// ContextResponse represents the structured response format for context data
type ContextResponse struct {
	Content []ContextContent `json:"content"`
	Context *ContextData     `json:"context,omitempty"`
}

// ContextContent represents the text content of the response
type ContextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ContextData represents the structured context information
type ContextData struct {
	SystemType        string            `json:"systemType"`
	UsesFlakes        bool              `json:"usesFlakes"`
	UsesChannels      bool              `json:"usesChannels"`
	HomeManagerType   string            `json:"homeManagerType"`
	HasHomeManager    bool              `json:"hasHomeManager"`
	NixOSVersion      string            `json:"nixosVersion"`
	NixVersion        string            `json:"nixVersion"`
	SystemArch        string            `json:"systemArch"`
	EnabledServices   []string          `json:"enabledServices"`
	InstalledPackages []string          `json:"installedPackages"`
	ConfigPaths       map[string]string `json:"configPaths"`
	CacheInfo         *CacheInfo        `json:"cacheInfo,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// CacheInfo represents context cache information
type CacheInfo struct {
	Valid        bool      `json:"valid"`
	LastDetected time.Time `json:"lastDetected"`
	AgeSeconds   int64     `json:"ageSeconds"`
	Location     string    `json:"location,omitempty"`
}

// StatusInfo represents context system status information
type StatusInfo struct {
	Healthy        bool              `json:"healthy"`
	LastCheck      time.Time         `json:"lastCheck"`
	Issues         []string          `json:"issues,omitempty"`
	Metrics        map[string]string `json:"metrics,omitempty"`
	ConfigLocation string            `json:"configLocation"`
	DetectionSpeed string            `json:"detectionSpeed,omitempty"`
}

// FormatContextResponse formats NixOS context for MCP response
func FormatContextResponse(nixosCtx *config.NixOSContext, format string, detailed bool) *ContextResponse {
	if nixosCtx == nil {
		return &ContextResponse{
			Content: []ContextContent{{
				Type: "text",
				Text: "❌ No context available. Run context detection first.",
			}},
		}
	}

	// Create structured context data
	contextData := &ContextData{
		SystemType:        nixosCtx.SystemType,
		UsesFlakes:        nixosCtx.UsesFlakes,
		UsesChannels:      nixosCtx.UsesChannels,
		HomeManagerType:   nixosCtx.HomeManagerType,
		HasHomeManager:    nixosCtx.HasHomeManager,
		NixOSVersion:      nixosCtx.NixOSVersion,
		NixVersion:        nixosCtx.NixVersion,
		EnabledServices:   nixosCtx.EnabledServices,
		InstalledPackages: nixosCtx.InstalledPackages,
		ConfigPaths: map[string]string{
			"nixos":         nixosCtx.NixOSConfigPath,
			"homeManager":   nixosCtx.HomeManagerConfigPath,
			"flake":         nixosCtx.FlakeFile,
			"configuration": nixosCtx.ConfigurationNix,
			"hardware":      nixosCtx.HardwareConfigNix,
		},
		Metadata: make(map[string]string),
	}

	// Add cache information if available
	if nixosCtx.CacheValid {
		contextData.CacheInfo = &CacheInfo{
			Valid:        nixosCtx.CacheValid,
			LastDetected: nixosCtx.LastDetected,
			AgeSeconds:   int64(time.Since(nixosCtx.LastDetected).Seconds()),
		}
	}

	var textContent string
	if format == "json" {
		// JSON format response
		jsonData, _ := json.MarshalIndent(contextData, "", "  ")
		textContent = string(jsonData)
	} else {
		// Human-readable text format
		textContent = formatContextText(nixosCtx, detailed)
	}

	return &ContextResponse{
		Content: []ContextContent{{
			Type: "text",
			Text: textContent,
		}},
		Context: contextData,
	}
}

// formatContextText formats context as human-readable text
func formatContextText(nixosCtx *config.NixOSContext, detailed bool) string {
	var b strings.Builder

	// Header with system summary
	b.WriteString(utils.FormatHeader("📋 NixOS System Context"))
	b.WriteString("\n\n")

	// System summary line (matches CLI output)
	summary := fmt.Sprintf("📋 System: %s | Flakes: %s | Home Manager: %s",
		nixosCtx.SystemType,
		formatBoolYesNo(nixosCtx.UsesFlakes),
		nixosCtx.HomeManagerType)
	b.WriteString(summary)
	b.WriteString("\n\n")

	// System Information
	b.WriteString("### System Information\n")
	b.WriteString(utils.FormatKeyValue("System Type", nixosCtx.SystemType))
	if nixosCtx.NixOSVersion != "" {
		b.WriteString(utils.FormatKeyValue("NixOS Version", nixosCtx.NixOSVersion))
	}
	if nixosCtx.NixVersion != "" {
		b.WriteString(utils.FormatKeyValue("Nix Version", nixosCtx.NixVersion))
	}
	b.WriteString("\n")

	// Configuration
	b.WriteString("### Configuration\n")
	b.WriteString(utils.FormatKeyValue("Uses Flakes", formatBoolCheck(nixosCtx.UsesFlakes)))
	b.WriteString(utils.FormatKeyValue("Uses Channels", formatBoolCheck(nixosCtx.UsesChannels)))
	b.WriteString(utils.FormatKeyValue("Has Home Manager", formatBoolCheck(nixosCtx.HasHomeManager)))
	if nixosCtx.HasHomeManager {
		b.WriteString(utils.FormatKeyValue("Home Manager Type", nixosCtx.HomeManagerType))
	}
	b.WriteString("\n")

	// File Paths
	b.WriteString("### File Paths\n")
	if nixosCtx.NixOSConfigPath != "" {
		b.WriteString(utils.FormatKeyValue("NixOS Config", nixosCtx.NixOSConfigPath))
	}
	if nixosCtx.HomeManagerConfigPath != "" {
		b.WriteString(utils.FormatKeyValue("Home Manager Config", nixosCtx.HomeManagerConfigPath))
	}
	if nixosCtx.FlakeFile != "" {
		b.WriteString(utils.FormatKeyValue("Flake File", nixosCtx.FlakeFile))
	}
	if nixosCtx.ConfigurationNix != "" {
		b.WriteString(utils.FormatKeyValue("Configuration.nix", nixosCtx.ConfigurationNix))
	}
	if nixosCtx.HardwareConfigNix != "" {
		b.WriteString(utils.FormatKeyValue("Hardware Config", nixosCtx.HardwareConfigNix))
	}
	b.WriteString("\n")

	if detailed {
		// Enabled Services (limited list for MCP)
		if len(nixosCtx.EnabledServices) > 0 {
			b.WriteString(fmt.Sprintf("### Enabled Services (%d total)\n", len(nixosCtx.EnabledServices)))
			serviceCount := 10 // Limit for MCP response
			if len(nixosCtx.EnabledServices) < serviceCount {
				serviceCount = len(nixosCtx.EnabledServices)
			}
			for i := 0; i < serviceCount; i++ {
				b.WriteString(fmt.Sprintf("  • %s\n", nixosCtx.EnabledServices[i]))
			}
			if len(nixosCtx.EnabledServices) > serviceCount {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(nixosCtx.EnabledServices)-serviceCount))
			}
			b.WriteString("\n")
		}

		// Installed Packages (limited list for MCP)
		if len(nixosCtx.InstalledPackages) > 0 {
			b.WriteString(fmt.Sprintf("### Installed Packages (%d total)\n", len(nixosCtx.InstalledPackages)))
			packageCount := 15 // Limit for MCP response
			if len(nixosCtx.InstalledPackages) < packageCount {
				packageCount = len(nixosCtx.InstalledPackages)
			}
			for i := 0; i < packageCount; i++ {
				b.WriteString(fmt.Sprintf("  • %s\n", nixosCtx.InstalledPackages[i]))
			}
			if len(nixosCtx.InstalledPackages) > packageCount {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(nixosCtx.InstalledPackages)-packageCount))
			}
			b.WriteString("\n")
		}
	}

	// Cache Information
	if nixosCtx.CacheValid {
		b.WriteString("### Cache Information\n")
		b.WriteString(utils.FormatKeyValue("Cache Valid", "✅ Yes"))
		b.WriteString(utils.FormatKeyValue("Last Detected", nixosCtx.LastDetected.Format("2006-01-02 15:04:05")))
		age := time.Since(nixosCtx.LastDetected)
		b.WriteString(utils.FormatKeyValue("Cache Age", formatDuration(age)))
	}

	return b.String()
}

// FormatStatusResponse formats context system status for MCP response
func FormatStatusResponse(status *StatusInfo, includeMetrics bool) *ContextResponse {
	var b strings.Builder

	b.WriteString(utils.FormatHeader("📊 Context System Status"))
	b.WriteString("\n\n")

	// Health status
	healthIcon := "✅"
	healthText := "Healthy"
	if !status.Healthy {
		healthIcon = "❌"
		healthText = "Unhealthy"
	}
	b.WriteString(utils.FormatKeyValue("System Health", fmt.Sprintf("%s %s", healthIcon, healthText)))
	b.WriteString(utils.FormatKeyValue("Last Check", status.LastCheck.Format("2006-01-02 15:04:05")))

	if status.ConfigLocation != "" {
		b.WriteString(utils.FormatKeyValue("Config Location", status.ConfigLocation))
	}

	if status.DetectionSpeed != "" {
		b.WriteString(utils.FormatKeyValue("Detection Speed", status.DetectionSpeed))
	}

	if len(status.Issues) > 0 {
		b.WriteString("\n### Issues Detected\n")
		for i, issue := range status.Issues {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
	}

	if includeMetrics && len(status.Metrics) > 0 {
		b.WriteString("\n### Metrics\n")
		for key, value := range status.Metrics {
			b.WriteString(utils.FormatKeyValue(key, value))
		}
	}

	return &ContextResponse{
		Content: []ContextContent{{
			Type: "text",
			Text: b.String(),
		}},
		Context: &ContextData{
			Metadata: map[string]string{
				"type":    "status",
				"healthy": fmt.Sprintf("%t", status.Healthy),
			},
		},
	}
}

// Helper functions for formatting
func formatBoolYesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func formatBoolCheck(value bool) string {
	if value {
		return "✅ Yes"
	}
	return "❌ No"
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

// FormatDetectionResponse formats context detection results
func FormatDetectionResponse(nixosCtx *config.NixOSContext, verbose bool) *ContextResponse {
	if nixosCtx == nil {
		return &ContextResponse{
			Content: []ContextContent{{
				Type: "text",
				Text: "❌ Context detection failed. Please check system configuration and permissions.",
			}},
		}
	}

	var b strings.Builder

	b.WriteString(utils.FormatHeader("🔍 NixOS Context Detection"))
	b.WriteString("\n\n")

	if verbose {
		b.WriteString("Starting context detection process...\n")
		b.WriteString("Clearing context cache...\n")
		b.WriteString("Re-detecting system context...\n\n")
	}

	// Show the detected context
	b.WriteString(formatContextText(nixosCtx, false))

	b.WriteString("\n")
	b.WriteString("✅ Context detection completed\n")

	return FormatContextResponse(nixosCtx, "text", false)
}

// FormatResetResponse formats context reset results
func FormatResetResponse(success bool, nixosCtx *config.NixOSContext) *ContextResponse {
	var b strings.Builder

	b.WriteString(utils.FormatHeader("🔄 Reset NixOS Context"))
	b.WriteString("\n\n")

	if success {
		b.WriteString("Clearing context cache...\n")
		b.WriteString("Re-detecting system context...\n\n")

		if nixosCtx != nil {
			summary := fmt.Sprintf("📋 System: %s | Flakes: %s | Home Manager: %s",
				nixosCtx.SystemType,
				formatBoolYesNo(nixosCtx.UsesFlakes),
				nixosCtx.HomeManagerType)
			b.WriteString(summary)
			b.WriteString("\n\n")
		}

		b.WriteString("✅ Context reset completed\n")
	} else {
		b.WriteString("❌ Context reset failed\n")
		b.WriteString("Please check system configuration and permissions.\n")
	}

	return &ContextResponse{
		Content: []ContextContent{{
			Type: "text",
			Text: b.String(),
		}},
		Context: func() *ContextData {
			if nixosCtx != nil {
				return &ContextData{
					SystemType:      nixosCtx.SystemType,
					UsesFlakes:      nixosCtx.UsesFlakes,
					HomeManagerType: nixosCtx.HomeManagerType,
					Metadata: map[string]string{
						"operation": "reset",
						"success":   fmt.Sprintf("%t", success),
					},
				}
			}
			return nil
		}(),
	}
}
