package validation

import (
	"regexp"
	"strings"
)

// FlakeValidationResult represents the result of flake syntax validation
type FlakeValidationResult struct {
	IsValid  bool
	Errors   []FlakeValidationError
	Warnings []FlakeValidationWarning
	Severity string // "low", "medium", "high", "critical"
}

// FlakeValidationError represents a syntax error in flake content
type FlakeValidationError struct {
	Type        string
	Message     string
	Suggestion  string
	LinePattern string // The problematic pattern found
}

// FlakeValidationWarning represents a potential issue in flake content
type FlakeValidationWarning struct {
	Type        string
	Message     string
	Suggestion  string
	LinePattern string
}

// FlakeValidator validates NixOS flake syntax and structure
type FlakeValidator struct {
	// Common incorrect patterns that AI models generate
	incorrectPatterns []IncorrectPattern
	// Required patterns for valid flakes
	requiredPatterns []RequiredPattern
}

// IncorrectPattern represents a pattern that indicates incorrect flake syntax
type IncorrectPattern struct {
	Pattern    *regexp.Regexp
	ErrorType  string
	Message    string
	Suggestion string
	Severity   string
}

// RequiredPattern represents a pattern that should be present in valid flakes
type RequiredPattern struct {
	Pattern     *regexp.Regexp
	Description string
	Required    bool
}

// NewFlakeValidator creates a new flake validator with common error patterns
func NewFlakeValidator() *FlakeValidator {
	incorrectPatterns := []IncorrectPattern{
		{
			Pattern:    regexp.MustCompile(`nixpkgs\.nix\s*=\s*\{[^}]*type\s*=\s*"github"`),
			ErrorType:  "incorrect_input_syntax",
			Message:    "Incorrect input syntax: using 'nixpkgs.nix = { type = \"github\" }' format",
			Suggestion: "Use 'nixpkgs.url = \"github:NixOS/nixpkgs/...\"' instead",
			Severity:   "critical",
		},
		{
			Pattern:    regexp.MustCompile(`devShell\s*=\s*\{\s*package\s*=`),
			ErrorType:  "incorrect_devshell_structure",
			Message:    "Incorrect devShell structure: using 'devShell = { package = ... }'",
			Suggestion: "Use 'devShells.default = pkgs.mkShell { ... }' or 'devShells.${system}.default = ...'",
			Severity:   "critical",
		},
		{
			Pattern:    regexp.MustCompile(`outputs\s*=\s*\{\s*self\s*=\s*\{`),
			ErrorType:  "incorrect_outputs_structure",
			Message:    "Incorrect outputs structure: using 'outputs = { self = { ... } }'",
			Suggestion: "Use 'outputs = { self, nixpkgs }: { ... }' function syntax",
			Severity:   "critical",
		},
		{
			Pattern:    regexp.MustCompile(`\.nix\s*=\s*\{[^}]*type\s*=`),
			ErrorType:  "incorrect_input_format",
			Message:    "Incorrect input format: using '.nix = { type = ... }' syntax",
			Suggestion: "Use '.url = \"...\"' format for inputs",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`release\s*=\s*\{[^}]*Build artifacts`),
			ErrorType:  "placeholder_content",
			Message:    "Contains placeholder/example content that should be customized",
			Suggestion: "Replace example content with actual implementation",
			Severity:   "medium",
		},
		{
			Pattern:    regexp.MustCompile(`pkgs\.mkShell\s*\{[^}]*buildInputs.*pkgs\.python3.*pkgs\.nodejs`),
			ErrorType:  "generic_dependencies",
			Message:    "Using generic example dependencies (python3, nodejs)",
			Suggestion: "Replace with dependencies specific to your project",
			Severity:   "low",
		},
		// NixOS Configuration Option Validation
		{
			Pattern:    regexp.MustCompile(`services\.bluetooth\.enable\s*=\s*true`),
			ErrorType:  "incorrect_nixos_option",
			Message:    "Incorrect NixOS option: 'services.bluetooth.enable' does not exist",
			Suggestion: "Use 'hardware.bluetooth.enable = true;' instead",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`services\.audio\.enable\s*=\s*true`),
			ErrorType:  "incorrect_nixos_option",
			Message:    "Incorrect NixOS option: 'services.audio.enable' does not exist",
			Suggestion: "Use 'sound.enable = true;' or 'security.rtkit.enable = true; services.pipewire.enable = true;' for modern audio",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`services\.wifi\.enable\s*=\s*true`),
			ErrorType:  "incorrect_nixos_option",
			Message:    "Incorrect NixOS option: 'services.wifi.enable' does not exist",
			Suggestion: "Use 'networking.wireless.enable = true;' or 'networking.networkmanager.enable = true;' instead",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`services\.graphics\.enable\s*=\s*true`),
			ErrorType:  "incorrect_nixos_option",
			Message:    "Incorrect NixOS option: 'services.graphics.enable' does not exist",
			Suggestion: "Use 'hardware.opengl.enable = true;' or 'services.xserver.enable = true;' instead",
			Severity:   "high",
		},
		{
			Pattern:    regexp.MustCompile(`environment\.packages\s*=`),
			ErrorType:  "incorrect_nixos_option",
			Message:    "Incorrect NixOS option: 'environment.packages' does not exist",
			Suggestion: "Use 'environment.systemPackages = with pkgs; [ ... ];' instead",
			Severity:   "high",
		},
	}

	requiredPatterns := []RequiredPattern{
		{
			Pattern:     regexp.MustCompile(`outputs\s*=\s*\{[^}]*\}\s*:\s*\{`),
			Description: "Function-style outputs definition",
			Required:    true,
		},
		{
			Pattern:     regexp.MustCompile(`description\s*=`),
			Description: "Flake description",
			Required:    true,
		},
		{
			Pattern:     regexp.MustCompile(`inputs\s*=\s*\{`),
			Description: "Inputs section",
			Required:    false,
		},
	}

	return &FlakeValidator{
		incorrectPatterns: incorrectPatterns,
		requiredPatterns:  requiredPatterns,
	}
}

// ValidateFlakeContent validates the content of a flake response
func (fv *FlakeValidator) ValidateFlakeContent(content string) *FlakeValidationResult {
	result := &FlakeValidationResult{
		IsValid:  true,
		Errors:   []FlakeValidationError{},
		Warnings: []FlakeValidationWarning{},
		Severity: "low",
	}

	// Check for incorrect patterns
	for _, pattern := range fv.incorrectPatterns {
		if matches := pattern.Pattern.FindAllString(content, -1); len(matches) > 0 {
			for _, match := range matches {
				error := FlakeValidationError{
					Type:        pattern.ErrorType,
					Message:     pattern.Message,
					Suggestion:  pattern.Suggestion,
					LinePattern: strings.TrimSpace(match),
				}
				result.Errors = append(result.Errors, error)

				// Update severity based on the most severe error found
				if pattern.Severity == "critical" {
					result.Severity = "critical"
					result.IsValid = false
				} else if pattern.Severity == "high" && result.Severity != "critical" {
					result.Severity = "high"
					result.IsValid = false
				} else if pattern.Severity == "medium" && result.Severity != "critical" && result.Severity != "high" {
					result.Severity = "medium"
				}
			}
		}
	}

	// Check for missing required patterns
	for _, required := range fv.requiredPatterns {
		if required.Required && !required.Pattern.MatchString(content) {
			warning := FlakeValidationWarning{
				Type:       "missing_required_pattern",
				Message:    "Missing required pattern: " + required.Description,
				Suggestion: "Ensure your flake includes " + required.Description,
			}
			result.Warnings = append(result.Warnings, warning)
		}
	}

	return result
}

// FormatValidationResult formats the validation result for display
func (fv *FlakeValidator) FormatValidationResult(result *FlakeValidationResult) string {
	if result.IsValid && len(result.Warnings) == 0 {
		return "✅ Flake syntax validation passed"
	}

	var output strings.Builder

	if !result.IsValid {
		output.WriteString("❌ **Flake Syntax Validation Failed**\n\n")
		output.WriteString("⚠️  **The AI response contains incorrect NixOS flake syntax!**\n\n")
	} else {
		output.WriteString("⚠️  **Flake Syntax Warnings**\n\n")
	}

	// Display errors
	if len(result.Errors) > 0 {
		output.WriteString("**Errors:**\n")
		for i, err := range result.Errors {
			output.WriteString(formatError(i+1, err))
		}
		output.WriteString("\n")
	}

	// Display warnings
	if len(result.Warnings) > 0 {
		output.WriteString("**Warnings:**\n")
		for i, warning := range result.Warnings {
			output.WriteString(formatWarning(i+1, warning))
		}
		output.WriteString("\n")
	}

	// Add recommendations
	if !result.IsValid {
		output.WriteString("**Recommendations:**\n")
		output.WriteString("- Refer to official NixOS flake documentation: https://nixos.wiki/wiki/Flakes\n")
		output.WriteString("- Use `nix flake init` to generate a basic template\n")
		output.WriteString("- Validate your flake with `nix flake check`\n")
		output.WriteString("- Consider the AI response as a starting point that needs correction\n\n")
	}

	return output.String()
}

// formatError formats an individual error for display
func formatError(index int, err FlakeValidationError) string {
	var output strings.Builder
	output.WriteString(strings.Repeat("  ", 1)) // Indent
	output.WriteString("• **Error " + string(rune('0'+index)) + "**: " + err.Message + "\n")
	if err.LinePattern != "" {
		output.WriteString(strings.Repeat("  ", 2))
		output.WriteString("Found: `" + err.LinePattern + "`\n")
	}
	output.WriteString(strings.Repeat("  ", 2))
	output.WriteString("Fix: " + err.Suggestion + "\n\n")
	return output.String()
}

// formatWarning formats an individual warning for display
func formatWarning(index int, warning FlakeValidationWarning) string {
	var output strings.Builder
	output.WriteString(strings.Repeat("  ", 1)) // Indent
	output.WriteString("• **Warning " + string(rune('0'+index)) + "**: " + warning.Message + "\n")
	if warning.LinePattern != "" {
		output.WriteString(strings.Repeat("  ", 2))
		output.WriteString("Found: `" + warning.LinePattern + "`\n")
	}
	output.WriteString(strings.Repeat("  ", 2))
	output.WriteString("Suggestion: " + warning.Suggestion + "\n\n")
	return output.String()
}

// IsFlakeContent checks if content appears to be about NixOS flakes
func IsFlakeContent(content string) bool {
	flakeKeywords := []string{
		"flake.nix",
		"outputs =",
		"inputs =",
		"nixpkgs",
		"devShell",
		"devShells",
		"nix flake",
		"flake init",
		"flake check",
	}

	content = strings.ToLower(content)
	keywordCount := 0

	for _, keyword := range flakeKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			keywordCount++
		}
	}

	// If we find 3 or more flake-related keywords, consider it flake content
	return keywordCount >= 3
}
