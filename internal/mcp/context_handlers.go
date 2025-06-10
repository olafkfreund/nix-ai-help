package mcp

import (
	"fmt"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
)

// handleGetContext gets current NixOS system context information
func (m *MCPServer) handleGetContext(format string, detailed bool) string {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Sprintf("❌ Error loading config: %v", err)
	}

	// Create context detector (need pointer to logger)
	contextDetector := nixos.NewContextDetector(&m.logger)

	// Get context (will use cache if valid)
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		return fmt.Sprintf("❌ Failed to get context: %v", err)
	}

	// Format response
	response := FormatContextResponse(nixosCtx, format, detailed)
	return response.Content[0].Text
}

// handleDetectContext forces re-detection of NixOS system context
func (m *MCPServer) handleDetectContext(verbose bool) string {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Sprintf("❌ Error loading config: %v", err)
	}

	// Create context detector (need pointer to logger)
	contextDetector := nixos.NewContextDetector(&m.logger)

	// Force context detection by clearing cache first
	if err := contextDetector.ClearCache(); err != nil {
		return fmt.Sprintf("❌ Failed to clear cache: %v", err)
	}

	// Get fresh context (will detect since cache was cleared)
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		return fmt.Sprintf("❌ Context detection failed: %v", err)
	}

	// Format response
	response := FormatDetectionResponse(nixosCtx, verbose)
	return response.Content[0].Text
}

// handleResetContext clears cached context and forces refresh
func (m *MCPServer) handleResetContext(confirm bool) string {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Sprintf("❌ Error loading config: %v", err)
	}

	// Create context detector (need pointer to logger)
	contextDetector := nixos.NewContextDetector(&m.logger)

	// Clear cache
	if err := contextDetector.ClearCache(); err != nil {
		return fmt.Sprintf("❌ Failed to clear cache: %v", err)
	}

	// Get fresh context
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		return fmt.Sprintf("❌ Context re-detection failed: %v", err)
	}

	// Format successful response
	response := FormatResetResponse(true, nixosCtx)
	return response.Content[0].Text
}

// handleContextStatus shows context detection system status and health
func (m *MCPServer) handleContextStatus(includeMetrics bool) string {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Sprintf("❌ Error loading config: %v", err)
	}

	// Create context detector (need pointer to logger)
	contextDetector := nixos.NewContextDetector(&m.logger)

	// Check system health
	status := &StatusInfo{
		LastCheck:      time.Now(),
		Healthy:        true,
		Issues:         []string{},
		Metrics:        make(map[string]string),
		ConfigLocation: contextDetector.GetCacheLocation(),
	}

	// Try to get current context to check health
	startTime := time.Now()
	nixosCtx, err := contextDetector.GetContext(cfg)
	detectionTime := time.Since(startTime)

	// Check for issues
	if err != nil {
		status.Healthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("Context detection failed: %v", err))
	}

	// Check cache validity
	if nixosCtx != nil {
		if !nixosCtx.CacheValid {
			status.Issues = append(status.Issues, "Context cache is invalid or expired")
		}

		// Check for missing critical information
		if nixosCtx.SystemType == "" {
			status.Issues = append(status.Issues, "System type detection failed")
		}
		if nixosCtx.ConfigurationNix == "" {
			status.Issues = append(status.Issues, "NixOS configuration path not found")
		}
	} else {
		status.Healthy = false
		status.Issues = append(status.Issues, "No context information available")
	}

	// Add metrics if requested
	if includeMetrics {
		status.Metrics["detection_time_ms"] = fmt.Sprintf("%.0f", detectionTime.Seconds()*1000)
		status.Metrics["cache_valid"] = fmt.Sprintf("%t", nixosCtx != nil && nixosCtx.CacheValid)

		if nixosCtx != nil {
			status.Metrics["enabled_services_count"] = fmt.Sprintf("%d", len(nixosCtx.EnabledServices))
			status.Metrics["installed_packages_count"] = fmt.Sprintf("%d", len(nixosCtx.InstalledPackages))
			status.Metrics["uses_flakes"] = fmt.Sprintf("%t", nixosCtx.UsesFlakes)
			status.Metrics["has_home_manager"] = fmt.Sprintf("%t", nixosCtx.HasHomeManager)

			if nixosCtx.CacheValid {
				age := time.Since(nixosCtx.LastDetected)
				status.Metrics["cache_age_seconds"] = fmt.Sprintf("%.0f", age.Seconds())
			}
		}
	}

	// Set detection speed assessment
	if detectionTime < 100*time.Millisecond {
		status.DetectionSpeed = "Fast (< 100ms)"
	} else if detectionTime < 500*time.Millisecond {
		status.DetectionSpeed = "Normal (< 500ms)"
	} else if detectionTime < 1*time.Second {
		status.DetectionSpeed = "Slow (< 1s)"
	} else {
		status.DetectionSpeed = "Very Slow (> 1s)"
		status.Issues = append(status.Issues, "Context detection is unusually slow")
	}

	// If there are no issues, system is healthy
	if len(status.Issues) == 0 {
		status.Healthy = true
	}

	// Format response
	response := FormatStatusResponse(status, includeMetrics)
	return response.Content[0].Text
}

// Helper function to validate context detection capability
func (m *MCPServer) validateContextSystem(cfg *config.UserConfig) []string {
	var issues []string

	// Check if context detection dependencies are available
	contextDetector := nixos.NewContextDetector(&m.logger)

	// Try a quick context check
	_, err := contextDetector.GetContext(cfg)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Context detection error: %v", err))
	}

	// Check cache location
	if contextDetector.GetCacheLocation() == "" {
		issues = append(issues, "Configuration cache location unavailable")
	}

	return issues
}

// Helper function to generate context summary for AI integration
func (m *MCPServer) generateContextSummary(nixosCtx *config.NixOSContext) string {
	if nixosCtx == nil {
		return "No NixOS context available"
	}

	summary := fmt.Sprintf("System: %s, Flakes: %t, Home Manager: %s",
		nixosCtx.SystemType,
		nixosCtx.UsesFlakes,
		nixosCtx.HomeManagerType)

	if len(nixosCtx.EnabledServices) > 0 {
		summary += fmt.Sprintf(", Services: %d enabled", len(nixosCtx.EnabledServices))
	}

	return summary
}
