package validation

import (
	"regexp"
	"strings"
)

// NixOSValidationResult represents the result of NixOS configuration validation
type NixOSValidationResult struct {
	IsValid  bool
	Errors   []NixOSValidationError
	Warnings []NixOSValidationWarning
	Severity string // "low", "medium", "high", "critical"
}

// NixOSValidationError represents an error in NixOS configuration
type NixOSValidationError struct {
	Type          string
	Message       string
	Suggestion    string
	LinePattern   string
	OptionName    string // The incorrect option name
	CorrectOption string // The correct option name
}

// NixOSValidationWarning represents a potential issue in NixOS configuration
type NixOSValidationWarning struct {
	Type        string
	Message     string
	Suggestion  string
	LinePattern string
	OptionName  string
}

// NixOSValidator validates NixOS configuration options and syntax
type NixOSValidator struct {
	// Maps incorrect options to correct ones
	optionCorrections map[string]string
	// Common incorrect patterns
	incorrectPatterns []NixOSIncorrectPattern
}

// NixOSIncorrectPattern represents a pattern that indicates incorrect NixOS configuration
type NixOSIncorrectPattern struct {
	Pattern    *regexp.Regexp
	ErrorType  string
	Message    string
	Suggestion string
	Severity   string
}

// NewNixOSValidator creates a new NixOS configuration validator
func NewNixOSValidator() *NixOSValidator {
	optionCorrections := map[string]string{
		// Hardware options
		"services.bluetooth.enable": "hardware.bluetooth.enable",
		"services.audio.enable":     "sound.enable",
		"services.wifi.enable":      "networking.wireless.enable",
		"services.graphics.enable":  "hardware.opengl.enable",
		"services.sound.enable":     "sound.enable",
		"hardware.audio.enable":     "sound.enable",

		// Networking options
		"services.network.enable":        "networking.networkmanager.enable",
		"services.networkmanager.enable": "networking.networkmanager.enable",
		"network.enable":                 "networking.networkmanager.enable",

		// Package management
		"environment.packages": "environment.systemPackages",
		"system.packages":      "environment.systemPackages",
		"packages":             "environment.systemPackages",

		// Service management
		"services.ssh.enable":  "services.openssh.enable",
		"services.sshd.enable": "services.openssh.enable",
		"ssh.enable":           "services.openssh.enable",

		// Display/X11 options
		"services.display.enable": "services.xserver.enable",
		"services.gui.enable":     "services.xserver.enable",
		"services.desktop.enable": "services.xserver.enable",

		// User management
		"users.user": "users.users",
		"user.users": "users.users",

		// Boot options
		"boot.grub.enable":    "boot.loader.grub.enable",
		"grub.enable":         "boot.loader.grub.enable",
		"systemd-boot.enable": "boot.loader.systemd-boot.enable",

		// Firewall options
		"firewall.enable":          "networking.firewall.enable",
		"services.firewall.enable": "networking.firewall.enable",
	}

	incorrectPatterns := []NixOSIncorrectPattern{
		{
			Pattern:    regexp.MustCompile(`services\.blueman.*=.*true`),
			ErrorType:  "incomplete_bluetooth_config",
			Message:    "Using blueman without enabling Bluetooth hardware first",
			Suggestion: "Enable Bluetooth with 'hardware.bluetooth.enable = true;' before configuring blueman",
			Severity:   "medium",
		},
		{
			Pattern:    regexp.MustCompile(`nix-env\s+-[iuq]`),
			ErrorType:  "deprecated_command",
			Message:    "Using deprecated 'nix-env' command in NixOS configuration advice",
			Suggestion: "Use declarative configuration in configuration.nix or flake.nix instead of imperative nix-env commands",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`sudo\s+nix-env`),
			ErrorType:  "deprecated_command",
			Message:    "Using deprecated 'sudo nix-env' command",
			Suggestion: "Use 'environment.systemPackages' in configuration.nix for system packages",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`apt\s+install|yum\s+install|pacman\s+-S`),
			ErrorType:  "wrong_package_manager",
			Message:    "Using non-NixOS package manager commands",
			Suggestion: "Use NixOS declarative configuration instead of traditional package managers",
			Severity:   "critical",
		},
		{
			Pattern:    regexp.MustCompile(`systemctl\s+enable.*\.service`),
			ErrorType:  "imperative_service_management",
			Message:    "Using imperative systemctl commands instead of declarative configuration",
			Suggestion: "Use 'services.<service>.enable = true;' in configuration.nix instead",
			Severity:   "medium",
		},
		{
			Pattern:    regexp.MustCompile(`/etc/\w+/.*\.conf`),
			ErrorType:  "direct_config_file_editing",
			Message:    "Suggesting direct editing of configuration files in /etc",
			Suggestion: "Use NixOS configuration options instead of directly editing files in /etc",
			Severity:   "medium",
		},
	}

	return &NixOSValidator{
		optionCorrections: optionCorrections,
		incorrectPatterns: incorrectPatterns,
	}
}

// ValidateNixOSContent validates NixOS configuration content
func (nv *NixOSValidator) ValidateNixOSContent(content string) *NixOSValidationResult {
	result := &NixOSValidationResult{
		IsValid:  true,
		Errors:   []NixOSValidationError{},
		Warnings: []NixOSValidationWarning{},
		Severity: "low",
	}

	// Special handling for Bluetooth configuration
	if strings.Contains(content, "services.blueman.enable") && !strings.Contains(content, "hardware.bluetooth.enable") {
		warning := NixOSValidationWarning{
			Type:        "incomplete_bluetooth_config",
			Message:     "Blueman service enabled without Bluetooth hardware support",
			Suggestion:  "Add 'hardware.bluetooth.enable = true;' to enable Bluetooth hardware support",
			LinePattern: "services.blueman.enable",
		}
		result.Warnings = append(result.Warnings, warning)
		if result.Severity == "low" {
			result.Severity = "medium"
		}
	}

	// Check for incorrect option names
	for incorrectOption, correctOption := range nv.optionCorrections {
		pattern := regexp.MustCompile(regexp.QuoteMeta(incorrectOption) + `\s*=`)
		if matches := pattern.FindAllString(content, -1); len(matches) > 0 {
			for _, match := range matches {
				error := NixOSValidationError{
					Type:          "incorrect_option_name",
					Message:       "Incorrect NixOS option: '" + incorrectOption + "' does not exist",
					Suggestion:    "Use '" + correctOption + "' instead",
					LinePattern:   strings.TrimSpace(match),
					OptionName:    incorrectOption,
					CorrectOption: correctOption,
				}
				result.Errors = append(result.Errors, error)
				result.IsValid = false
				if result.Severity == "low" {
					result.Severity = "high"
				}
			}
		}
	}

	// Check for incorrect patterns
	for _, pattern := range nv.incorrectPatterns {
		if matches := pattern.Pattern.FindAllString(content, -1); len(matches) > 0 {
			for _, match := range matches {
				if pattern.Severity == "critical" {
					error := NixOSValidationError{
						Type:        pattern.ErrorType,
						Message:     pattern.Message,
						Suggestion:  pattern.Suggestion,
						LinePattern: strings.TrimSpace(match),
					}
					result.Errors = append(result.Errors, error)
					result.IsValid = false
					result.Severity = "critical"
				} else {
					warning := NixOSValidationWarning{
						Type:        pattern.ErrorType,
						Message:     pattern.Message,
						Suggestion:  pattern.Suggestion,
						LinePattern: strings.TrimSpace(match),
					}
					result.Warnings = append(result.Warnings, warning)
					if result.Severity == "low" && (pattern.Severity == "medium" || pattern.Severity == "high") {
						result.Severity = pattern.Severity
					}
				}
			}
		}
	}

	return result
}

// FormatNixOSValidationResult formats the validation result for display
func (nv *NixOSValidator) FormatNixOSValidationResult(result *NixOSValidationResult) string {
	if result.IsValid && len(result.Warnings) == 0 {
		return ""
	}

	var output strings.Builder
	output.WriteString("## ⚠️ NixOS Configuration Validation\n\n")

	if len(result.Errors) > 0 {
		output.WriteString("### ❌ **Errors Found**\n")
		for _, err := range result.Errors {
			output.WriteString("- **" + err.Message + "**\n")
			if err.CorrectOption != "" {
				output.WriteString("  - ✅ **Correction**: `" + err.CorrectOption + "`\n")
			} else {
				output.WriteString("  - 💡 **Suggestion**: " + err.Suggestion + "\n")
			}
			if err.LinePattern != "" {
				output.WriteString("  - 🔍 **Found**: `" + err.LinePattern + "`\n")
			}
			output.WriteString("\n")
		}
	}

	if len(result.Warnings) > 0 {
		output.WriteString("### ⚠️ **Warnings**\n")
		for _, warning := range result.Warnings {
			output.WriteString("- **" + warning.Message + "**\n")
			output.WriteString("  - 💡 **Suggestion**: " + warning.Suggestion + "\n")
			if warning.LinePattern != "" {
				output.WriteString("  - 🔍 **Found**: `" + warning.LinePattern + "`\n")
			}
			output.WriteString("\n")
		}
	}

	return output.String()
}

// IsNixOSContent checks if the content appears to be NixOS configuration
func IsNixOSContent(content string) bool {
	// Look for common NixOS configuration patterns
	nixosPatterns := []string{
		`services\.`,
		`hardware\.`,
		`networking\.`,
		`environment\.systemPackages`,
		`boot\.loader`,
		`users\.users`,
		`nixos-rebuild`,
		`configuration\.nix`,
		`flake\.nix`,
	}

	for _, pattern := range nixosPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}

	return false
}
