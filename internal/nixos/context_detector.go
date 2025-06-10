package nixos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// ContextDetector handles NixOS configuration context detection
type ContextDetector struct {
	logger *logger.Logger
}

// NewContextDetector creates a new context detector
func NewContextDetector(log *logger.Logger) *ContextDetector {
	return &ContextDetector{
		logger: log,
	}
}

// DetectNixOSContext performs comprehensive NixOS configuration detection
func (cd *ContextDetector) DetectNixOSContext(userConfig *config.UserConfig) (*config.NixOSContext, error) {
	cd.logger.Info("Starting NixOS context detection...")

	context := &config.NixOSContext{
		LastDetected:    time.Now(),
		CacheValid:      false,
		DetectionErrors: []string{},
	}

	// Run detection methods
	cd.detectSystemType(context)
	cd.detectNixVersion(context)
	cd.detectFlakesUsage(context, userConfig)
	cd.detectChannelsUsage(context)
	cd.detectHomeManager(context)
	cd.detectConfigurationFiles(context, userConfig)
	cd.detectEnabledServices(context)
	cd.detectInstalledPackages(context)

	// Mark cache as valid if no critical errors
	context.CacheValid = len(context.DetectionErrors) == 0

	cd.logger.Info("NixOS context detection completed")
	return context, nil
}

// detectSystemType determines if running on NixOS, nix-darwin, or other
func (cd *ContextDetector) detectSystemType(context *config.NixOSContext) {
	cd.logger.Debug("Detecting system type...")

	// Check for NixOS
	if _, err := os.Stat("/etc/nixos"); err == nil {
		context.SystemType = "nixos"
		cd.logger.Debug("Detected NixOS system")
		return
	}

	// Check for nix-darwin on macOS
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat("/etc/nix/nix.conf"); err == nil {
			context.SystemType = "nix-darwin"
			cd.logger.Debug("Detected nix-darwin system")
			return
		}
	}

	// Check if nix is available but not system-wide (home-manager only)
	if cmd := exec.Command("which", "nix"); cmd.Run() == nil {
		context.SystemType = "home-manager-only"
		cd.logger.Debug("Detected home-manager-only system")
		return
	}

	context.SystemType = "unknown"
	context.DetectionErrors = append(context.DetectionErrors, "Unable to determine system type")
	cd.logger.Warn("Unable to determine system type")
}

// detectNixVersion gets nix and NixOS version information
func (cd *ContextDetector) detectNixVersion(context *config.NixOSContext) {
	cd.logger.Debug("Detecting Nix version...")

	// Get Nix version
	if output, err := exec.Command("nix", "--version").Output(); err == nil {
		context.NixVersion = strings.TrimSpace(string(output))
		cd.logger.Debug("Detected Nix version: " + context.NixVersion)
	}

	// Get NixOS version (only on NixOS systems)
	if context.SystemType == "nixos" {
		if output, err := exec.Command("nixos-version").Output(); err == nil {
			context.NixOSVersion = strings.TrimSpace(string(output))
			cd.logger.Debug("Detected NixOS version: " + context.NixOSVersion)
		}
	}
}

// detectFlakesUsage checks if the system uses Nix flakes
func (cd *ContextDetector) detectFlakesUsage(context *config.NixOSContext, userConfig *config.UserConfig) {
	cd.logger.Debug("Detecting flakes usage...")

	// Priority order for flake detection:
	// 1. User-specified paths
	// 2. /etc/nixos/flake.nix
	// 3. ~/.config/nixos/flake.nix
	// 4. Current directory flake.nix
	// 5. Home directory flake.nix

	flakePaths := []string{
		userConfig.NixosFolder + "/flake.nix",
		"/etc/nixos/flake.nix",
		filepath.Join(os.Getenv("HOME"), ".config", "nixos", "flake.nix"),
		"./flake.nix",
		filepath.Join(os.Getenv("HOME"), "flake.nix"),
	}

	for _, flakePath := range flakePaths {
		// Expand ~ to home directory
		if strings.HasPrefix(flakePath, "~/") {
			flakePath = filepath.Join(os.Getenv("HOME"), flakePath[2:])
		}

		if _, err := os.Stat(flakePath); err == nil {
			context.UsesFlakes = true
			context.FlakeFile = flakePath
			cd.logger.Debug("Found flake.nix at: " + flakePath)
			return
		}
	}

	// Check if flakes are enabled in nix.conf
	nixConfPaths := []string{"/etc/nix/nix.conf", filepath.Join(os.Getenv("HOME"), ".config", "nix", "nix.conf")}
	for _, confPath := range nixConfPaths {
		if content, err := os.ReadFile(confPath); err == nil {
			if strings.Contains(string(content), "experimental-features") &&
				strings.Contains(string(content), "flakes") {
				context.UsesFlakes = true
				cd.logger.Debug("Flakes enabled in nix.conf: " + confPath)
				return
			}
		}
	}

	context.UsesFlakes = false
	cd.logger.Debug("No flakes usage detected")
}

// detectChannelsUsage checks if the system uses Nix channels
func (cd *ContextDetector) detectChannelsUsage(context *config.NixOSContext) {
	cd.logger.Debug("Detecting channels usage...")

	// Check for user channels
	if output, err := exec.Command("nix-channel", "--list").Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		context.UsesChannels = true
		cd.logger.Debug("User channels detected")
		return
	}

	// Check for system channels (on NixOS)
	if context.SystemType == "nixos" {
		if output, err := exec.Command("sudo", "nix-channel", "--list").Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
			context.UsesChannels = true
			cd.logger.Debug("System channels detected")
			return
		}
	}

	context.UsesChannels = false
	cd.logger.Debug("No channels usage detected")
}

// detectHomeManager checks for Home Manager installation and type
func (cd *ContextDetector) detectHomeManager(context *config.NixOSContext) {
	cd.logger.Debug("Detecting Home Manager...")

	// Check for standalone Home Manager
	if cmd := exec.Command("which", "home-manager"); cmd.Run() == nil {
		context.HasHomeManager = true
		context.HomeManagerType = "standalone"

		// Find Home Manager config
		hmConfigPaths := []string{
			filepath.Join(os.Getenv("HOME"), ".config", "home-manager", "home.nix"),
			filepath.Join(os.Getenv("HOME"), ".config", "nixpkgs", "home.nix"),
		}

		for _, hmPath := range hmConfigPaths {
			if _, err := os.Stat(hmPath); err == nil {
				context.HomeManagerConfigPath = hmPath
				break
			}
		}

		cd.logger.Debug("Detected standalone Home Manager")
		return
	}

	// Check for Home Manager as NixOS module
	if context.SystemType == "nixos" {
		configPaths := []string{
			"/etc/nixos/configuration.nix",
			context.ConfigurationNix,
		}

		for _, configPath := range configPaths {
			if configPath == "" {
				continue
			}

			if content, err := os.ReadFile(configPath); err == nil {
				if strings.Contains(string(content), "home-manager") {
					context.HasHomeManager = true
					context.HomeManagerType = "module"
					cd.logger.Debug("Detected Home Manager as NixOS module")
					return
				}
			}
		}
	}

	context.HasHomeManager = false
	context.HomeManagerType = "none"
	cd.logger.Debug("No Home Manager detected")
}

// detectConfigurationFiles finds and analyzes NixOS configuration files
func (cd *ContextDetector) detectConfigurationFiles(context *config.NixOSContext, userConfig *config.UserConfig) {
	cd.logger.Debug("Detecting configuration files...")

	var configPaths []string

	// Priority order for configuration detection:
	// 1. User-specified paths
	// 2. /etc/nixos/
	// 3. ~/.config/nixos/
	// 4. Current directory

	if userConfig.NixosFolder != "" {
		nixosFolder := userConfig.NixosFolder
		if strings.HasPrefix(nixosFolder, "~/") {
			nixosFolder = filepath.Join(os.Getenv("HOME"), nixosFolder[2:])
		}
		configPaths = append(configPaths, filepath.Join(nixosFolder, "configuration.nix"))
	}

	configPaths = append(configPaths,
		"/etc/nixos/configuration.nix",
		filepath.Join(os.Getenv("HOME"), ".config", "nixos", "configuration.nix"),
		"./configuration.nix",
	)

	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			context.ConfigurationNix = configPath
			context.NixOSConfigPath = filepath.Dir(configPath)
			context.ConfigurationFiles = append(context.ConfigurationFiles, configPath)
			cd.logger.Debug("Found configuration.nix at: " + configPath)
			break
		}
	}

	// Look for hardware-configuration.nix
	if context.NixOSConfigPath != "" {
		hwConfigPath := filepath.Join(context.NixOSConfigPath, "hardware-configuration.nix")
		if _, err := os.Stat(hwConfigPath); err == nil {
			context.HardwareConfigNix = hwConfigPath
			context.ConfigurationFiles = append(context.ConfigurationFiles, hwConfigPath)
			cd.logger.Debug("Found hardware-configuration.nix at: " + hwConfigPath)
		}
	}

	// Look for additional .nix files in the config directory
	if context.NixOSConfigPath != "" {
		if files, err := filepath.Glob(filepath.Join(context.NixOSConfigPath, "*.nix")); err == nil {
			for _, file := range files {
				found := false
				for _, existingFile := range context.ConfigurationFiles {
					if existingFile == file {
						found = true
						break
					}
				}
				if !found {
					context.ConfigurationFiles = append(context.ConfigurationFiles, file)
				}
			}
		}
	}
}

// detectEnabledServices parses configuration files to find enabled services
func (cd *ContextDetector) detectEnabledServices(context *config.NixOSContext) {
	cd.logger.Debug("Detecting enabled services...")

	if context.ConfigurationNix == "" {
		return
	}

	content, err := os.ReadFile(context.ConfigurationNix)
	if err != nil {
		context.DetectionErrors = append(context.DetectionErrors,
			fmt.Sprintf("Failed to read configuration.nix: %v", err))
		return
	}

	// Parse services from configuration
	serviceRegex := regexp.MustCompile(`services\.([a-zA-Z0-9_-]+)\.enable\s*=\s*true`)
	matches := serviceRegex.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) > 1 {
			context.EnabledServices = append(context.EnabledServices, match[1])
		}
	}

	cd.logger.Debug("Detected " + fmt.Sprintf("%d", len(context.EnabledServices)) + " enabled services")
}

// detectInstalledPackages attempts to get a list of installed packages
func (cd *ContextDetector) detectInstalledPackages(context *config.NixOSContext) {
	cd.logger.Debug("Detecting installed packages...")

	// Try to get system packages (limited to avoid performance issues)
	if output, err := exec.Command("nix-env", "--query", "--installed").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		count := 0
		for scanner.Scan() && count < 50 { // Limit to first 50 packages
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				context.InstalledPackages = append(context.InstalledPackages, line)
				count++
			}
		}
	}

	cd.logger.Debug("Detected " + fmt.Sprintf("%d", len(context.InstalledPackages)) + " installed packages")
}

// IsContextCacheValid checks if the cached context is still valid
func (cd *ContextDetector) IsContextCacheValid(context *config.NixOSContext) bool {
	if context == nil || !context.CacheValid {
		return false
	}

	// Check if cache is older than 1 hour
	if time.Since(context.LastDetected) > time.Hour {
		return false
	}

	// Check if key configuration files have been modified
	if context.ConfigurationNix != "" {
		if stat, err := os.Stat(context.ConfigurationNix); err == nil {
			if stat.ModTime().After(context.LastDetected) {
				return false
			}
		}
	}

	if context.FlakeFile != "" {
		if stat, err := os.Stat(context.FlakeFile); err == nil {
			if stat.ModTime().After(context.LastDetected) {
				return false
			}
		}
	}

	return true
}

// RefreshContext forces a refresh of the NixOS context
func (cd *ContextDetector) RefreshContext(userConfig *config.UserConfig) error {
	cd.logger.Info("Refreshing NixOS context...")

	newContext, err := cd.DetectNixOSContext(userConfig)
	if err != nil {
		return fmt.Errorf("failed to refresh context: %v", err)
	}

	userConfig.NixOSContext = *newContext

	// Save updated config
	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save updated context: %v", err)
	}

	cd.logger.Info("NixOS context refreshed successfully")
	return nil
}

// GetContext returns the current context, detecting if necessary
func (cd *ContextDetector) GetContext(userConfig *config.UserConfig) (*config.NixOSContext, error) {
	// Check if we have a valid cached context
	if cd.IsContextCacheValid(&userConfig.NixOSContext) {
		cd.logger.Debug("Using cached NixOS context")
		return &userConfig.NixOSContext, nil
	}

	// Detect new context
	cd.logger.Debug("Detecting fresh NixOS context")
	newContext, err := cd.DetectNixOSContext(userConfig)
	if err != nil {
		return nil, err
	}

	// Update and save the config
	userConfig.NixOSContext = *newContext
	if err := config.SaveUserConfig(userConfig); err != nil {
		cd.logger.Warn("Failed to save context to config: " + err.Error())
	}

	return newContext, nil
}

// ClearCache clears the cached context by invalidating it in the user config
func (cd *ContextDetector) ClearCache() error {
	cd.logger.Debug("Clearing context cache...")

	// Load current user config
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %v", err)
	}

	// Invalidate the cached context
	userConfig.NixOSContext.CacheValid = false
	userConfig.NixOSContext.LastDetected = time.Time{}
	userConfig.NixOSContext.DetectionErrors = []string{}

	// Save the updated config
	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save updated config: %v", err)
	}

	cd.logger.Debug("Context cache cleared successfully")
	return nil
}

// GetCacheLocation returns the location where context cache is stored
func (cd *ContextDetector) GetCacheLocation() string {
	// The context is cached in the user config file
	configPath, err := config.ConfigFilePath()
	if err != nil {
		return "unknown (config path unavailable)"
	}
	return configPath
}
