package nixos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nix-ai-help/pkg/logger"
)

// UpgradeInfo contains information about the current system and available upgrades
type UpgradeInfo struct {
	CurrentVersion    string        `json:"current_version"`
	CurrentChannel    string        `json:"current_channel"`
	AvailableChannels []ChannelInfo `json:"available_channels"`
	PreChecks         []CheckResult `json:"pre_checks"`
	BackupAdvice      []string      `json:"backup_advice"`
	UpgradeSteps      []UpgradeStep `json:"upgrade_steps"`
	PostChecks        []string      `json:"post_checks"`
	Warnings          []string      `json:"warnings"`
	EstimatedTime     string        `json:"estimated_time"`
}

// ChannelInfo represents information about a NixOS channel
type ChannelInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	URL           string `json:"url"`
	IsCurrent     bool   `json:"is_current"`
	IsRecommended bool   `json:"is_recommended"`
	ReleaseDate   string `json:"release_date"`
	Description   string `json:"description"`
}

// CheckResult represents the result of a pre-upgrade check
type CheckResult struct {
	Name       string `json:"name"`
	Status     string `json:"status"` // "pass", "warn", "fail"
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Critical   bool   `json:"critical"`
}

// UpgradeStep represents a step in the upgrade process
type UpgradeStep struct {
	Title         string `json:"title"`
	Command       string `json:"command"`
	Description   string `json:"description"`
	Optional      bool   `json:"optional"`
	Dangerous     bool   `json:"dangerous"`
	EstimatedTime string `json:"estimated_time"`
}

// UpgradeAdvisor provides functionality for NixOS upgrade guidance
type UpgradeAdvisor struct {
	logger     logger.Logger
	configPath string
}

// NewUpgradeAdvisor creates a new UpgradeAdvisor instance
func NewUpgradeAdvisor(log logger.Logger) *UpgradeAdvisor {
	return &UpgradeAdvisor{
		logger: log,
	}
}

// NewUpgradeAdvisorWithConfig creates a new UpgradeAdvisor instance with a configuration path
func NewUpgradeAdvisorWithConfig(log logger.Logger, configPath string) *UpgradeAdvisor {
	return &UpgradeAdvisor{
		logger:     log,
		configPath: configPath,
	}
}

// isFlakeBased checks if the configuration is flake-based
func (ua *UpgradeAdvisor) isFlakeBased() bool {
	if ua.configPath == "" {
		return false
	}

	flakePath := filepath.Join(ua.configPath, "flake.nix")
	if _, err := os.Stat(flakePath); err == nil {
		return true
	}

	return false
}

// AnalyzeUpgradeOptions analyzes the current system and provides upgrade recommendations
func (ua *UpgradeAdvisor) AnalyzeUpgradeOptions(ctx context.Context) (*UpgradeInfo, error) {
	ua.logger.Info("Starting upgrade analysis...")

	info := &UpgradeInfo{
		PreChecks:    make([]CheckResult, 0),
		BackupAdvice: make([]string, 0),
		UpgradeSteps: make([]UpgradeStep, 0),
		PostChecks:   make([]string, 0),
		Warnings:     make([]string, 0),
	}

	isFlake := ua.isFlakeBased()

	// Get current version (always needed)
	if err := ua.getCurrentVersion(ctx, info); err != nil {
		return nil, fmt.Errorf("failed to get current system version: %w", err)
	}

	// For traditional configs only: get channel info and available channels
	if !isFlake {
		if err := ua.getCurrentChannelInfo(ctx, info); err != nil {
			ua.logger.Warn("Failed to get current channel info: " + err.Error())
		}

		if err := ua.getAvailableChannels(ctx, info); err != nil {
			ua.logger.Warn("Failed to get available channels: " + err.Error())
			// Continue anyway, this is not critical
		}
	} else {
		ua.logger.Info("Flake-based configuration detected - skipping channel checks")
		info.CurrentChannel = "flake-based"

		// Get flake input information instead
		if err := ua.getFlakeInputInfo(ctx, info); err != nil {
			ua.logger.Warn("Failed to get flake input info: " + err.Error())
		}
	}

	// Run pre-upgrade checks
	ua.runPreUpgradeChecks(ctx, info)

	// Analyze flake inputs (if applicable)
	if isFlake {
		if err := ua.getFlakeInputInfo(ctx, info); err != nil {
			ua.logger.Warn("Failed to analyze flake inputs: " + err.Error())
		}
	}

	// Generate backup advice
	ua.generateBackupAdvice(info)

	// Generate upgrade steps
	ua.generateUpgradeSteps(info)

	// Generate post-upgrade checks
	ua.generatePostUpgradeChecks(info)

	// Estimate upgrade time
	ua.estimateUpgradeTime(info)

	ua.logger.Info("Upgrade analysis completed")
	return info, nil
}

// getCurrentVersion gets the current NixOS version
func (ua *UpgradeAdvisor) getCurrentVersion(ctx context.Context, info *UpgradeInfo) error {
	// Get NixOS version
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "nixos-version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get nixos-version: %w", err)
	}

	versionStr := strings.TrimSpace(string(output))
	info.CurrentVersion = versionStr

	// Extract version number for comparison
	versionRegex := regexp.MustCompile(`(\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(versionStr)
	if len(matches) > 1 {
		info.CurrentVersion = matches[1]
	}

	return nil
}

// getCurrentChannelInfo gets the current channel information (for traditional configs only)
func (ua *UpgradeAdvisor) getCurrentChannelInfo(ctx context.Context, info *UpgradeInfo) error {
	// Get current channel
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "nix-channel", "--list")
	output, err := cmd.Output()
	if err != nil {
		ua.logger.Warn("Failed to get nix channels: " + err.Error())
		info.CurrentChannel = "unknown"
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.Contains(line, "nixos") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.CurrentChannel = parts[1]
				break
			}
		}
	}

	if info.CurrentChannel == "" {
		info.CurrentChannel = "unknown"
	}

	return nil
}

// getCurrentSystemInfo gets the current NixOS version and channel information (deprecated - use separate methods)
func (ua *UpgradeAdvisor) getCurrentSystemInfo(ctx context.Context, info *UpgradeInfo) error {
	if err := ua.getCurrentVersion(ctx, info); err != nil {
		return err
	}

	return ua.getCurrentChannelInfo(ctx, info)
}

// getAvailableChannels retrieves information about available NixOS channels (for traditional configs only)
func (ua *UpgradeAdvisor) getAvailableChannels(ctx context.Context, info *UpgradeInfo) error {
	// Skip for flake-based configurations
	if ua.isFlakeBased() {
		ua.logger.Info("Skipping channel information for flake-based configuration")
		return nil
	}

	// Define known stable channels with their information
	channels := []ChannelInfo{
		{
			Name:        "nixos-23.11",
			Version:     "23.11",
			URL:         "https://nixos.org/channels/nixos-23.11",
			Description: "NixOS 23.11 (Tapir) - Stable release",
			ReleaseDate: "2023-12-09",
		},
		{
			Name:          "nixos-24.05",
			Version:       "24.05",
			URL:           "https://nixos.org/channels/nixos-24.05",
			Description:   "NixOS 24.05 (Uakari) - Stable release",
			ReleaseDate:   "2024-05-31",
			IsRecommended: true,
		},
		{
			Name:          "nixos-24.11",
			Version:       "24.11",
			URL:           "https://nixos.org/channels/nixos-24.11",
			Description:   "NixOS 24.11 (Vicuña) - Latest stable release",
			ReleaseDate:   "2024-11-30",
			IsRecommended: true,
		},
		{
			Name:        "nixos-unstable",
			Version:     "unstable",
			URL:         "https://nixos.org/channels/nixos-unstable",
			Description: "NixOS Unstable - Rolling release with latest packages",
			ReleaseDate: "continuous",
		},
	}

	// Mark current channel
	for i := range channels {
		if strings.Contains(info.CurrentChannel, channels[i].Name) {
			channels[i].IsCurrent = true
		}
	}

	info.AvailableChannels = channels
	return nil
}

// runPreUpgradeChecks performs various system checks before upgrade
func (ua *UpgradeAdvisor) runPreUpgradeChecks(ctx context.Context, info *UpgradeInfo) {
	isFlake := ua.isFlakeBased()

	var checks []func(context.Context, *UpgradeInfo) CheckResult

	// Common checks for both flake and traditional configs
	checks = append(checks,
		ua.checkConfigurationFiles,
		ua.checkDiskSpace,
		ua.checkConfigValidity,
		ua.checkRunningServices,
		ua.checkBootLoader,
		ua.checkNetworkConnectivity,
		ua.checkNixStoreIntegrity,
	)

	// Conditional checks based on configuration type
	if isFlake {
		// Flake-specific checks
		checks = append(checks, ua.checkFlakeInputs)
		checks = append(checks, ua.checkFlakeLock)
	} else {
		// Traditional channel-based checks
		checks = append(checks, ua.checkChannelUpdates)
	}

	for _, check := range checks {
		result := check(ctx, info)
		info.PreChecks = append(info.PreChecks, result)

		if result.Status == "fail" && result.Critical {
			info.Warnings = append(info.Warnings,
				fmt.Sprintf("Critical issue detected: %s", result.Message))
		}
	}
}

// checkConfigurationFiles verifies that essential configuration files exist
func (ua *UpgradeAdvisor) checkConfigurationFiles(ctx context.Context, info *UpgradeInfo) CheckResult {
	if ua.configPath == "" {
		return CheckResult{
			Name:    "Configuration Files",
			Status:  "warn",
			Message: "Config path not set, skipping file checks",
		}
	}

	// Check for either flake.nix or configuration.nix
	flakePath := filepath.Join(ua.configPath, "flake.nix")
	configPath := filepath.Join(ua.configPath, "configuration.nix")

	hasFlake := false
	hasConfig := false

	if _, err := os.Stat(flakePath); err == nil {
		hasFlake = true
	}

	if _, err := os.Stat(configPath); err == nil {
		hasConfig = true
	}

	if !hasFlake && !hasConfig {
		return CheckResult{
			Name:       "Configuration Files",
			Status:     "fail",
			Message:    "No flake.nix or configuration.nix found",
			Suggestion: "Ensure your NixOS configuration files exist in the specified path",
			Critical:   true,
		}
	}

	if hasFlake {
		return CheckResult{
			Name:    "Configuration Files",
			Status:  "pass",
			Message: "Flake-based configuration detected",
		}
	}

	return CheckResult{
		Name:    "Configuration Files",
		Status:  "pass",
		Message: "Traditional configuration.nix detected",
	}
}

// checkDiskSpace verifies sufficient disk space for upgrade
func (ua *UpgradeAdvisor) checkDiskSpace(ctx context.Context, info *UpgradeInfo) CheckResult {
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "df", "-h", "/nix")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:     "Disk Space Check",
			Status:   "warn",
			Message:  "Could not check disk space",
			Critical: false,
		}
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return CheckResult{
			Name:     "Disk Space Check",
			Status:   "warn",
			Message:  "Could not parse disk space information",
			Critical: false,
		}
	}

	// Parse the available space (fourth column in df output)
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return CheckResult{
			Name:     "Disk Space Check",
			Status:   "warn",
			Message:  "Could not parse disk space information",
			Critical: false,
		}
	}

	availableStr := fields[3]

	// Extract numeric value (remove G, M, K suffixes)
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)([GMK]?)`)
	matches := re.FindStringSubmatch(availableStr)
	if len(matches) < 2 {
		return CheckResult{
			Name:     "Disk Space Check",
			Status:   "warn",
			Message:  "Could not parse available disk space",
			Critical: false,
		}
	}

	available, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return CheckResult{
			Name:     "Disk Space Check",
			Status:   "warn",
			Message:  "Could not parse available disk space",
			Critical: false,
		}
	}

	// Convert to GB
	switch matches[2] {
	case "K":
		available = available / (1024 * 1024)
	case "M":
		available = available / 1024
	case "G":
		// already in GB
	default:
		// Assume bytes
		available = available / (1024 * 1024 * 1024)
	}

	if available < 2 {
		return CheckResult{
			Name:       "Disk Space Check",
			Status:     "fail",
			Message:    fmt.Sprintf("Insufficient disk space: %.1fGB available", available),
			Suggestion: "Free up at least 2GB of disk space before upgrading",
			Critical:   true,
		}
	} else if available < 5 {
		return CheckResult{
			Name:       "Disk Space Check",
			Status:     "warn",
			Message:    fmt.Sprintf("Low disk space: %.1fGB available", available),
			Suggestion: "Consider freeing up more space for a smoother upgrade",
			Critical:   false,
		}
	}

	return CheckResult{
		Name:    "Disk Space Check",
		Status:  "pass",
		Message: fmt.Sprintf("Sufficient disk space: %.1fGB available", available),
	}
}

// checkConfigValidity verifies NixOS configuration validity
func (ua *UpgradeAdvisor) checkConfigValidity(ctx context.Context, info *UpgradeInfo) CheckResult {
	// If we have a config path, run the check from that directory
	var cmd *exec.Cmd
	if ua.configPath != "" {
		// #nosec G204 -- Arguments are constructed internally, not from user input
		cmd = exec.CommandContext(ctx, "nixos-rebuild", "dry-run", "--fast")
		cmd.Dir = ua.configPath
	} else {
		// #nosec G204 -- Arguments are constructed internally, not from user input
		cmd = exec.CommandContext(ctx, "nixos-rebuild", "dry-run", "--fast")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Parse the error output to provide more specific suggestions
		errorMsg := string(output)
		suggestion := "Fix configuration errors before upgrading"

		if strings.Contains(errorMsg, "syntax error") {
			suggestion = "Fix syntax errors in your configuration files"
		} else if strings.Contains(errorMsg, "infinite recursion") {
			suggestion = "Resolve infinite recursion in your configuration"
		} else if strings.Contains(errorMsg, "assertion failed") {
			suggestion = "Fix assertion failures in your configuration"
		} else if strings.Contains(errorMsg, "file not found") {
			suggestion = "Ensure all referenced files exist in your configuration"
		}

		return CheckResult{
			Name:       "Configuration Validity",
			Status:     "fail",
			Message:    "NixOS configuration has errors",
			Suggestion: suggestion,
			Critical:   true,
		}
	}

	return CheckResult{
		Name:    "Configuration Validity",
		Status:  "pass",
		Message: "NixOS configuration is valid",
	}
}

// checkRunningServices checks for critical running services
func (ua *UpgradeAdvisor) checkRunningServices(ctx context.Context, info *UpgradeInfo) CheckResult {
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--failed", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:     "Service Status Check",
			Status:   "warn",
			Message:  "Could not check service status",
			Critical: false,
		}
	}

	failedServices := strings.TrimSpace(string(output))
	if failedServices != "" {
		lines := strings.Split(failedServices, "\n")
		return CheckResult{
			Name:       "Service Status Check",
			Status:     "warn",
			Message:    fmt.Sprintf("%d failed services detected", len(lines)),
			Suggestion: "Review and fix failed services before upgrading",
			Critical:   false,
		}
	}

	return CheckResult{
		Name:    "Service Status Check",
		Status:  "pass",
		Message: "No failed services detected",
	}
}

// checkChannelUpdates verifies channel update status
func (ua *UpgradeAdvisor) checkChannelUpdates(ctx context.Context, info *UpgradeInfo) CheckResult {
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "nix-channel", "--update", "--dry-run")
	err := cmd.Run()
	if err != nil {
		return CheckResult{
			Name:       "Channel Updates",
			Status:     "warn",
			Message:    "Could not check for channel updates",
			Suggestion: "Ensure network connectivity and try updating channels",
			Critical:   false,
		}
	}

	return CheckResult{
		Name:    "Channel Updates",
		Status:  "pass",
		Message: "Channels can be updated successfully",
	}
}

// checkBootLoader verifies boot loader configuration
func (ua *UpgradeAdvisor) checkBootLoader(ctx context.Context, info *UpgradeInfo) CheckResult {
	// Check if systemd-boot or GRUB is configured
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "test", "-d", "/boot/loader")
	err := cmd.Run()
	if err == nil {
		return CheckResult{
			Name:    "Boot Loader Check",
			Status:  "pass",
			Message: "systemd-boot detected and configured",
		}
	}

	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd = exec.CommandContext(ctx, "test", "-f", "/boot/grub/grub.cfg")
	err = cmd.Run()
	if err == nil {
		return CheckResult{
			Name:    "Boot Loader Check",
			Status:  "pass",
			Message: "GRUB detected and configured",
		}
	}

	return CheckResult{
		Name:       "Boot Loader Check",
		Status:     "warn",
		Message:    "Boot loader configuration unclear",
		Suggestion: "Verify boot loader is properly configured",
		Critical:   false,
	}
}

// checkNetworkConnectivity verifies network access to Nix caches
func (ua *UpgradeAdvisor) checkNetworkConnectivity(ctx context.Context, info *UpgradeInfo) CheckResult {
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "curl", "-s", "--max-time", "10",
		"https://cache.nixos.org/nix-cache-info")
	err := cmd.Run()
	if err != nil {
		return CheckResult{
			Name:       "Network Connectivity",
			Status:     "warn",
			Message:    "Cannot reach NixOS binary cache",
			Suggestion: "Check internet connection and DNS resolution",
			Critical:   false,
		}
	}

	return CheckResult{
		Name:    "Network Connectivity",
		Status:  "pass",
		Message: "NixOS binary cache is accessible",
	}
}

// checkNixStoreIntegrity verifies Nix store integrity
func (ua *UpgradeAdvisor) checkNixStoreIntegrity(ctx context.Context, info *UpgradeInfo) CheckResult {
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "nix-store", "--verify", "--check-contents")
	cmd.Env = append(cmd.Env, "NIX_STORE_CHECK_LIMIT=100") // Limit check for performance
	err := cmd.Run()
	if err != nil {
		return CheckResult{
			Name:       "Nix Store Integrity",
			Status:     "warn",
			Message:    "Nix store integrity check found issues",
			Suggestion: "Run 'nix-store --verify --repair' to fix store issues",
			Critical:   false,
		}
	}

	return CheckResult{
		Name:    "Nix Store Integrity",
		Status:  "pass",
		Message: "Nix store integrity verified",
	}
}

// checkFlakeInputs verifies that flake inputs are accessible and up to date
func (ua *UpgradeAdvisor) checkFlakeInputs(ctx context.Context, info *UpgradeInfo) CheckResult {
	if ua.configPath == "" {
		return CheckResult{
			Name:    "Flake Inputs Check",
			Status:  "warn",
			Message: "Config path not set, skipping flake inputs check",
		}
	}

	// Test if we can evaluate the flake without building
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "nix", "flake", "show", ua.configPath, "--no-build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := string(output)
		suggestion := "Check flake inputs and network connectivity"

		if strings.Contains(errorMsg, "does not provide attribute") {
			suggestion = "Verify flake outputs are correctly defined"
		} else if strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "fetch") {
			suggestion = "Check network connectivity and input URLs"
		} else if strings.Contains(errorMsg, "syntax error") {
			suggestion = "Fix syntax errors in flake.nix"
		}

		return CheckResult{
			Name:       "Flake Inputs Check",
			Status:     "fail",
			Message:    "Flake inputs cannot be resolved",
			Suggestion: suggestion,
			Critical:   true,
		}
	}

	return CheckResult{
		Name:    "Flake Inputs Check",
		Status:  "pass",
		Message: "Flake inputs are accessible and valid",
	}
}

// checkFlakeLock verifies flake.lock file status
func (ua *UpgradeAdvisor) checkFlakeLock(ctx context.Context, info *UpgradeInfo) CheckResult {
	if ua.configPath == "" {
		return CheckResult{
			Name:    "Flake Lock Check",
			Status:  "warn",
			Message: "Config path not set, skipping flake lock check",
		}
	}

	lockPath := filepath.Join(ua.configPath, "flake.lock")
	if _, err := os.Stat(lockPath); err != nil {
		return CheckResult{
			Name:       "Flake Lock Check",
			Status:     "fail",
			Message:    "flake.lock file is missing",
			Suggestion: "Run 'nix flake lock' to generate the lock file",
			Critical:   true,
		}
	}

	// Check if lock file is outdated by comparing timestamps
	flakePath := filepath.Join(ua.configPath, "flake.nix")
	flakeInfo, err := os.Stat(flakePath)
	if err != nil {
		return CheckResult{
			Name:    "Flake Lock Check",
			Status:  "warn",
			Message: "Could not check flake.nix timestamp",
		}
	}

	lockInfo, err := os.Stat(lockPath)
	if err != nil {
		return CheckResult{
			Name:    "Flake Lock Check",
			Status:  "warn",
			Message: "Could not check flake.lock timestamp",
		}
	}

	if flakeInfo.ModTime().After(lockInfo.ModTime()) {
		return CheckResult{
			Name:       "Flake Lock Check",
			Status:     "warn",
			Message:    "flake.lock may be outdated",
			Suggestion: "Consider running 'nix flake lock' to update lock file",
			Critical:   false,
		}
	}

	return CheckResult{
		Name:    "Flake Lock Check",
		Status:  "pass",
		Message: "flake.lock file exists and appears current",
	}
}

// getFlakeInputInfo retrieves information about flake inputs for upgrade recommendations
func (ua *UpgradeAdvisor) getFlakeInputInfo(ctx context.Context, info *UpgradeInfo) error {
	if ua.configPath == "" {
		ua.logger.Warn("Config path not set, skipping flake input analysis")
		return nil
	}

	// Get flake inputs information
	// #nosec G204 -- Arguments are constructed internally, not from user input
	cmd := exec.CommandContext(ctx, "nix", "flake", "metadata", ua.configPath, "--json")
	output, err := cmd.Output()
	if err != nil {
		ua.logger.Warn("Failed to get flake metadata: " + err.Error())
		return nil
	}

	// Parse the JSON output to extract input information
	// This is a simplified analysis - in a full implementation you'd parse the JSON
	outputStr := string(output)
	if strings.Contains(outputStr, "nixpkgs") {
		info.Warnings = append(info.Warnings,
			"Flake-based configuration detected. Use 'nix flake update' to update inputs.")
	}

	// Add flake-specific recommendations
	ua.logger.Info("Flake input analysis completed")
	return nil
}

// generateBackupAdvice creates backup recommendations
func (ua *UpgradeAdvisor) generateBackupAdvice(info *UpgradeInfo) {
	advice := []string{
		"Create a system backup before upgrading",
		"Back up your NixOS configuration files (/etc/nixos/)",
		"Document current system generation: nixos-rebuild list-generations",
		"Export current package list: nix-env -q > current-packages.txt",
		"Back up important user data",
		"Create a recovery USB/DVD with current NixOS ISO",
		"Note current kernel version: uname -r",
		"Export current systemd services: systemctl list-unit-files --state=enabled",
	}
	info.BackupAdvice = advice
}

// generateUpgradeSteps creates step-by-step upgrade instructions
func (ua *UpgradeAdvisor) generateUpgradeSteps(info *UpgradeInfo) {
	isFlake := ua.isFlakeBased()

	var steps []UpgradeStep

	if isFlake {
		// Flake-based upgrade steps
		steps = []UpgradeStep{
			{
				Title:         "Update flake inputs",
				Command:       "nix flake update",
				Description:   "Update all flake inputs to their latest versions",
				Optional:      false,
				Dangerous:     false,
				EstimatedTime: "1-3 minutes",
			},
			{
				Title:         "Test configuration",
				Command:       "sudo nixos-rebuild test --flake .",
				Description:   "Test the new configuration without making it permanent",
				Optional:      true,
				Dangerous:     false,
				EstimatedTime: "5-15 minutes",
			},
			{
				Title:         "Switch to new generation",
				Command:       "sudo nixos-rebuild switch --flake .",
				Description:   "Apply the upgrade and switch to the new system generation",
				Optional:      false,
				Dangerous:     true,
				EstimatedTime: "10-30 minutes",
			},
			{
				Title:         "Reboot system",
				Command:       "sudo reboot",
				Description:   "Restart to ensure all changes take effect properly",
				Optional:      true,
				Dangerous:     false,
				EstimatedTime: "2-5 minutes",
			},
			{
				Title:         "Clean old generations",
				Command:       "sudo nix-collect-garbage -d",
				Description:   "Remove old system generations to free up space",
				Optional:      true,
				Dangerous:     false,
				EstimatedTime: "1-3 minutes",
			},
		}
	} else {
		// Traditional channel-based upgrade steps
		steps = []UpgradeStep{
			{
				Title:         "Update channels",
				Command:       "sudo nix-channel --update",
				Description:   "Download the latest channel information and packages",
				Optional:      false,
				Dangerous:     false,
				EstimatedTime: "2-5 minutes",
			},
			{
				Title:         "Test configuration",
				Command:       "sudo nixos-rebuild test",
				Description:   "Test the new configuration without making it permanent",
				Optional:      true,
				Dangerous:     false,
				EstimatedTime: "5-15 minutes",
			},
			{
				Title:         "Switch to new generation",
				Command:       "sudo nixos-rebuild switch",
				Description:   "Apply the upgrade and switch to the new system generation",
				Optional:      false,
				Dangerous:     true,
				EstimatedTime: "10-30 minutes",
			},
			{
				Title:         "Reboot system",
				Command:       "sudo reboot",
				Description:   "Restart to ensure all changes take effect properly",
				Optional:      true,
				Dangerous:     false,
				EstimatedTime: "2-5 minutes",
			},
			{
				Title:         "Clean old generations",
				Command:       "sudo nix-collect-garbage -d",
				Description:   "Remove old system generations to free up space",
				Optional:      true,
				Dangerous:     false,
				EstimatedTime: "1-3 minutes",
			},
		}
	}

	// If we have a config path, adjust the commands to use it
	if ua.configPath != "" && isFlake {
		for i := range steps {
			if steps[i].Command == "nix flake update" {
				steps[i].Command = fmt.Sprintf("cd %s && nix flake update", ua.configPath)
			} else if strings.Contains(steps[i].Command, "--flake .") {
				steps[i].Command = strings.Replace(steps[i].Command, "--flake .", fmt.Sprintf("--flake %s", ua.configPath), 1)
			}
		}
	}

	info.UpgradeSteps = steps
}

// generatePostUpgradeChecks creates post-upgrade validation steps
func (ua *UpgradeAdvisor) generatePostUpgradeChecks(info *UpgradeInfo) {
	checks := []string{
		"Verify system version: nixos-version",
		"Check system services: systemctl status",
		"Verify network connectivity: ping google.com",
		"Test user applications and configurations",
		"Check available disk space: df -h",
		"Verify boot loader: ls /boot/loader/entries/ (systemd-boot) or grub-probe /",
		"Review system logs: journalctl -p err -b",
		"Test package installation: nix-env -iA nixpkgs.hello",
	}
	info.PostChecks = checks
}

// estimateUpgradeTime estimates total upgrade time based on system analysis
func (ua *UpgradeAdvisor) estimateUpgradeTime(info *UpgradeInfo) {
	// Base time estimation
	baseTime := 20 // minutes

	// Adjust based on checks
	for _, check := range info.PreChecks {
		if check.Status == "warn" || check.Status == "fail" {
			baseTime += 5 // Add time for potential issues
		}
	}

	// Adjust based on channel change
	if info.CurrentChannel != "" && !strings.Contains(info.CurrentChannel, "unstable") {
		baseTime += 10 // Channel upgrades typically take longer
	}

	info.EstimatedTime = fmt.Sprintf("%d-60 minutes", baseTime)
}
