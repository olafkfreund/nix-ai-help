package context

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/config"
)

// NixOSContextBuilder builds context-aware prompts for AI interactions
type NixOSContextBuilder struct{}

// NewNixOSContextBuilder creates a new context builder
func NewNixOSContextBuilder() *NixOSContextBuilder {
	return &NixOSContextBuilder{}
}

// BuildContextualPrompt creates a context-aware prompt based on detected NixOS configuration
func (cb *NixOSContextBuilder) BuildContextualPrompt(basePrompt string, context *config.NixOSContext) string {
	if context == nil || !context.CacheValid {
		return basePrompt + "\n\n" + cb.buildGenericNixOSPrompt()
	}

	contextPrompt := cb.buildDetectedContextPrompt(context)
	return basePrompt + "\n\n" + contextPrompt
}

// buildDetectedContextPrompt creates a prompt based on detected context
func (cb *NixOSContextBuilder) buildDetectedContextPrompt(context *config.NixOSContext) string {
	var prompt strings.Builder

	prompt.WriteString("=== USER'S NIXOS CONTEXT ===\n")
	prompt.WriteString(fmt.Sprintf("System Type: %s\n", context.SystemType))

	// System configuration approach
	if context.UsesFlakes {
		prompt.WriteString("✅ USES FLAKES - Always suggest flake-based solutions\n")
		prompt.WriteString("❌ NEVER suggest nix-channel commands\n")
		if context.FlakeFile != "" {
			prompt.WriteString(fmt.Sprintf("Flake location: %s\n", context.FlakeFile))
		}
	} else if context.UsesChannels {
		prompt.WriteString("Uses legacy channels - suggest channel-compatible solutions\n")
		prompt.WriteString("Prefer nix-channel and nixos-rebuild commands\n")
	} else {
		prompt.WriteString("Configuration approach unclear - provide both flake and channel options\n")
	}

	// Home Manager integration
	if context.HasHomeManager {
		switch context.HomeManagerType {
		case "standalone":
			prompt.WriteString("✅ HAS STANDALONE HOME MANAGER\n")
			prompt.WriteString("Use 'home-manager switch' commands\n")
			if context.HomeManagerConfigPath != "" {
				prompt.WriteString(fmt.Sprintf("Home Manager config: %s\n", context.HomeManagerConfigPath))
			}
		case "module":
			prompt.WriteString("✅ HAS HOME MANAGER AS NIXOS MODULE\n")
			prompt.WriteString("Use home-manager.users.<username> syntax in configuration.nix\n")
		}
	} else {
		prompt.WriteString("❌ NO HOME MANAGER - Only suggest system-level configuration\n")
	}

	// Version information
	if context.NixOSVersion != "" {
		prompt.WriteString(fmt.Sprintf("NixOS Version: %s\n", context.NixOSVersion))
	}
	if context.NixVersion != "" {
		prompt.WriteString(fmt.Sprintf("Nix Version: %s\n", context.NixVersion))
	}

	// Configuration files
	if len(context.ConfigurationFiles) > 0 {
		prompt.WriteString("Configuration files:\n")
		for _, file := range context.ConfigurationFiles {
			prompt.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}

	// Currently enabled services (limit to important ones)
	if len(context.EnabledServices) > 0 {
		importantServices := cb.filterImportantServices(context.EnabledServices)
		if len(importantServices) > 0 {
			prompt.WriteString("Currently enabled services: ")
			prompt.WriteString(strings.Join(importantServices, ", "))
			prompt.WriteString("\n")
		}
	}

	// Detection warnings
	if len(context.DetectionErrors) > 0 {
		prompt.WriteString("⚠️  Detection warnings: ")
		prompt.WriteString(strings.Join(context.DetectionErrors, "; "))
		prompt.WriteString("\n")
	}

	prompt.WriteString("=== END CONTEXT ===\n")

	return prompt.String()
}

// buildGenericNixOSPrompt creates a generic prompt when context detection fails
func (cb *NixOSContextBuilder) buildGenericNixOSPrompt() string {
	return `=== NIXOS CONFIGURATION GUIDANCE ===
Context detection unavailable - provide comprehensive guidance:

1. Ask user about their setup (flakes vs channels, Home Manager)
2. Provide both flake-based and channel-based solutions when applicable  
3. Include Home Manager options if user has it
4. Always verify configuration syntax before suggesting changes
5. Never suggest nix-env commands for system configuration
=== END GUIDANCE ===`
}

// filterImportantServices filters to show only commonly relevant services
func (cb *NixOSContextBuilder) filterImportantServices(services []string) []string {
	important := []string{
		"openssh", "sshd", "nginx", "apache", "postgresql", "mysql",
		"docker", "containerd", "firewall", "sound", "xserver", "gnome",
		"kde", "plasma", "networkmanager", "bluetooth", "printing",
	}

	var filtered []string
	for _, service := range services {
		for _, imp := range important {
			if strings.Contains(strings.ToLower(service), imp) {
				filtered = append(filtered, service)
				break
			}
		}
		// Limit to first 10 important services to avoid overwhelming the prompt
		if len(filtered) >= 10 {
			break
		}
	}

	return filtered
}

// BuildFlakeContextPrompt creates specific prompts for flake-based configurations
func (cb *NixOSContextBuilder) BuildFlakeContextPrompt(basePrompt string, context *config.NixOSContext) string {
	if context == nil || !context.UsesFlakes {
		return basePrompt
	}

	flakePrompt := fmt.Sprintf(`
%s

=== FLAKE-SPECIFIC GUIDANCE ===
User is using Nix flakes. Follow these rules:
- Always suggest flake-based commands (nixos-rebuild switch --flake)
- Never suggest nix-channel operations
- Reference inputs from flake.nix when suggesting packages
- Use flake syntax for package specifications
- Suggest flake.lock updates when needed
=== END FLAKE GUIDANCE ===`, basePrompt)

	return flakePrompt
}

// BuildHomeManagerContextPrompt creates specific prompts for Home Manager configurations
func (cb *NixOSContextBuilder) BuildHomeManagerContextPrompt(basePrompt string, context *config.NixOSContext) string {
	if context == nil || !context.HasHomeManager {
		return basePrompt
	}

	var hmType string
	switch context.HomeManagerType {
	case "standalone":
		hmType = "standalone Home Manager - use 'home-manager switch' commands"
	case "module":
		hmType = "Home Manager as NixOS module - use home-manager.users.<username> syntax"
	default:
		hmType = "Home Manager detected"
	}

	hmPrompt := fmt.Sprintf(`
%s

=== HOME MANAGER GUIDANCE ===
User has %s
- Provide Home Manager-specific solutions for user-level configuration
- Distinguish between system-level (configuration.nix) and user-level (Home Manager) options
- Include both approaches when relevant
=== END HOME MANAGER GUIDANCE ===`, basePrompt, hmType)

	return hmPrompt
}

// BuildSystemSpecificPrompt creates prompts specific to the detected system type
func (cb *NixOSContextBuilder) BuildSystemSpecificPrompt(basePrompt string, context *config.NixOSContext) string {
	if context == nil {
		return basePrompt
	}

	var systemGuidance string
	switch context.SystemType {
	case "nixos":
		systemGuidance = `
=== NIXOS SYSTEM GUIDANCE ===
User is on NixOS:
- Suggest system-level configuration changes via configuration.nix
- Use nixos-rebuild switch/test commands
- Reference NixOS modules and options
- Consider both system and user-level solutions
=== END SYSTEM GUIDANCE ===`

	case "nix-darwin":
		systemGuidance = `
=== NIX-DARWIN GUIDANCE ===
User is on nix-darwin (macOS):
- Use darwin-rebuild switch commands
- Reference nix-darwin modules and options
- Consider macOS-specific configurations
- Home Manager integration is common on nix-darwin
=== END DARWIN GUIDANCE ===`

	case "home-manager-only":
		systemGuidance = `
=== HOME MANAGER ONLY GUIDANCE ===
User has only Home Manager (not NixOS):
- Only suggest user-level configurations
- Use home-manager switch commands
- No system-level nixos-rebuild commands available
- Focus on user packages and dotfiles management
=== END HOME MANAGER GUIDANCE ===`

	default:
		systemGuidance = `
=== GENERIC NIX GUIDANCE ===
System type unclear:
- Ask user about their Nix setup
- Provide multiple solution approaches
- Include both NixOS and Home Manager options
=== END GENERIC GUIDANCE ===`
	}

	return basePrompt + "\n" + systemGuidance
}

// GetContextSummary returns a brief summary of the detected context
func (cb *NixOSContextBuilder) GetContextSummary(context *config.NixOSContext) string {
	if context == nil || !context.CacheValid {
		return "Context: Unknown/Not detected"
	}

	var parts []string

	parts = append(parts, fmt.Sprintf("System: %s", context.SystemType))

	if context.UsesFlakes {
		parts = append(parts, "Flakes: Yes")
	} else if context.UsesChannels {
		parts = append(parts, "Channels: Yes")
	}

	if context.HasHomeManager {
		parts = append(parts, fmt.Sprintf("Home Manager: %s", context.HomeManagerType))
	}

	if len(context.EnabledServices) > 0 {
		parts = append(parts, fmt.Sprintf("Services: %d", len(context.EnabledServices)))
	}

	return strings.Join(parts, " | ")
}
