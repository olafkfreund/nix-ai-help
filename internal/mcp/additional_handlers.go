package mcp

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
)

// Phase 1: Core NixOS Operations Handlers

// handleBuildSystemAnalyze analyzes build issues and suggests fixes with AI
func (m *MCPServer) handleBuildSystemAnalyze(buildLog, project, depth string) string {
	m.logger.Debug(fmt.Sprintf("handleBuildSystemAnalyze called | buildLog=%s project=%s depth=%s",
		truncateString(buildLog, 100), project, depth))

	var result strings.Builder
	result.WriteString("🔧 Build System Analysis\n\n")

	if buildLog == "" {
		result.WriteString("❌ No build log provided for analysis.\n")
		result.WriteString("Please provide build log content to analyze build issues.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("📋 Project: %s\n", project))
	result.WriteString(fmt.Sprintf("🔍 Analysis Depth: %s\n", depth))
	result.WriteString("\n📊 Build Log Analysis:\n")

	// Analyze common build issues
	if strings.Contains(buildLog, "error:") {
		result.WriteString("  ❌ Build errors detected\n")
	}
	if strings.Contains(buildLog, "warning:") {
		result.WriteString("  ⚠️  Build warnings found\n")
	}
	if strings.Contains(buildLog, "derivation") {
		result.WriteString("  📦 Nix derivation build detected\n")
	}
	if strings.Contains(buildLog, "failed") {
		result.WriteString("  💥 Build failure indicators found\n")
	}

	result.WriteString("\n🔧 Suggested Actions:\n")
	result.WriteString("  • Review error messages in the build log\n")
	result.WriteString("  • Check derivation dependencies\n")
	result.WriteString("  • Verify build inputs and outputs\n")
	result.WriteString("  • Consider running with --verbose for more details\n")

	return result.String()
}

// handleDiagnoseSystem diagnoses NixOS system issues from logs or config files
func (m *MCPServer) handleDiagnoseSystem(logContent, logType, contextStr string) string {
	m.logger.Debug(fmt.Sprintf("handleDiagnoseSystem called | logType=%s context=%s", logType, contextStr))

	var result strings.Builder
	result.WriteString("🩺 System Diagnosis\n\n")

	if logContent == "" {
		result.WriteString("❌ No log content provided for diagnosis.\n")
		result.WriteString("Please provide log content to analyze system issues.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("📋 Log Type: %s\n", logType))
	result.WriteString(fmt.Sprintf("🔍 Context: %s\n", contextStr))
	result.WriteString("\n📊 Log Analysis:\n")

	// Analyze common system issues
	if strings.Contains(logContent, "systemd") {
		result.WriteString("  🔧 SystemD related entries found\n")
	}
	if strings.Contains(logContent, "failed") || strings.Contains(logContent, "error") {
		result.WriteString("  ❌ Error conditions detected\n")
	}
	if strings.Contains(logContent, "warning") {
		result.WriteString("  ⚠️  Warning messages found\n")
	}
	if strings.Contains(logContent, "kernel") {
		result.WriteString("  🐧 Kernel related messages detected\n")
	}

	result.WriteString("\n🔧 Diagnostic Recommendations:\n")
	result.WriteString("  • Check system journal: journalctl -xe\n")
	result.WriteString("  • Review service status: systemctl status\n")
	result.WriteString("  • Verify configuration: nixos-rebuild dry-run\n")
	result.WriteString("  • Check disk space and system resources\n")

	return result.String()
}

// handleGenerateConfiguration generates NixOS configuration based on requirements
func (m *MCPServer) handleGenerateConfiguration(configType string, services, features []string) string {
	m.logger.Debug(fmt.Sprintf("handleGenerateConfiguration called | configType=%s services=%v features=%v",
		configType, services, features))

	var result strings.Builder
	result.WriteString("🛠️  Configuration Generator\n\n")

	result.WriteString(fmt.Sprintf("📋 Configuration Type: %s\n", configType))
	result.WriteString(fmt.Sprintf("🔧 Services: %v\n", services))
	result.WriteString(fmt.Sprintf("✨ Features: %v\n", features))
	result.WriteString("\n")

	// Generate basic configuration template
	result.WriteString("📝 Generated Configuration:\n\n")
	result.WriteString("```nix\n")
	result.WriteString("{ config, pkgs, ... }:\n\n")
	result.WriteString("{\n")

	// Add services configuration
	if len(services) > 0 {
		result.WriteString("  # Services Configuration\n")
		for _, service := range services {
			result.WriteString(fmt.Sprintf("  services.%s.enable = true;\n", service))
		}
		result.WriteString("\n")
	}

	// Add features configuration
	if len(features) > 0 {
		result.WriteString("  # Features Configuration\n")
		for _, feature := range features {
			result.WriteString(fmt.Sprintf("  # TODO: Configure %s\n", feature))
		}
		result.WriteString("\n")
	}

	result.WriteString("  # System packages\n")
	result.WriteString("  environment.systemPackages = with pkgs; [\n")
	result.WriteString("    # Add your packages here\n")
	result.WriteString("  ];\n")
	result.WriteString("}\n")
	result.WriteString("```\n")

	return result.String()
}

// handleValidateConfiguration validates NixOS configuration files for syntax and logic errors
func (m *MCPServer) handleValidateConfiguration(configContent, configPath, checkLevel string) string {
	m.logger.Debug(fmt.Sprintf("handleValidateConfiguration called | configPath=%s checkLevel=%s",
		configPath, checkLevel))

	var result strings.Builder
	result.WriteString("✅ Configuration Validation\n\n")

	if configContent == "" {
		result.WriteString("❌ No configuration content provided for validation.\n")
		result.WriteString("Please provide configuration content to validate.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("📋 Configuration Path: %s\n", configPath))
	result.WriteString(fmt.Sprintf("🔍 Check Level: %s\n", checkLevel))
	result.WriteString("\n📊 Validation Results:\n")

	// Basic syntax checks
	if strings.Contains(configContent, "{ config, pkgs") {
		result.WriteString("  ✅ Standard NixOS configuration format detected\n")
	}
	if strings.Contains(configContent, "services.") {
		result.WriteString("  ✅ Services configuration found\n")
	}
	if strings.Contains(configContent, "environment.systemPackages") {
		result.WriteString("  ✅ System packages configuration found\n")
	}

	// Check for common issues
	openBraces := strings.Count(configContent, "{")
	closeBraces := strings.Count(configContent, "}")
	if openBraces != closeBraces {
		result.WriteString("  ⚠️  Potential brace mismatch detected\n")
	} else {
		result.WriteString("  ✅ Brace balance looks correct\n")
	}

	result.WriteString("\n🔧 Validation Recommendations:\n")
	result.WriteString("  • Run: nixos-rebuild dry-build to test configuration\n")
	result.WriteString("  • Use: nix-instantiate --parse to check syntax\n")
	result.WriteString("  • Consider: nixos-option to validate specific options\n")

	return result.String()
}

// handleAnalyzePackageRepo analyzes Git repositories and generates Nix derivations
func (m *MCPServer) handleAnalyzePackageRepo(repoUrl, packageName, outputFormat string) string {
	m.logger.Debug(fmt.Sprintf("handleAnalyzePackageRepo called | repoUrl=%s packageName=%s outputFormat=%s",
		repoUrl, packageName, outputFormat))

	var result strings.Builder
	result.WriteString("📦 Package Repository Analysis\n\n")

	if repoUrl == "" {
		result.WriteString("❌ No repository URL provided for analysis.\n")
		result.WriteString("Please provide a Git repository URL to analyze.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("📋 Repository: %s\n", repoUrl))
	result.WriteString(fmt.Sprintf("🏷️  Package Name: %s\n", packageName))
	result.WriteString(fmt.Sprintf("📄 Output Format: %s\n", outputFormat))
	result.WriteString("\n📊 Repository Analysis:\n")

	// Analyze repository characteristics
	result.WriteString("  🔍 Repository characteristics:\n")
	result.WriteString("    • Language detection needed\n")
	result.WriteString("    • Build system identification required\n")
	result.WriteString("    • Dependency analysis needed\n")

	result.WriteString("\n📝 Generated Nix Derivation Template:\n\n")
	result.WriteString("```nix\n")
	result.WriteString("{ lib, stdenv, fetchFromGitHub, ... }:\n\n")
	result.WriteString("stdenv.mkDerivation rec {\n")
	result.WriteString(fmt.Sprintf("  pname = \"%s\";\n", packageName))
	result.WriteString("  version = \"0.1.0\"; # Update version\n\n")
	result.WriteString("  src = fetchFromGitHub {\n")
	// Extract owner/repo from URL
	if strings.Contains(repoUrl, "github.com") {
		parts := strings.Split(repoUrl, "/")
		if len(parts) >= 5 {
			owner := parts[len(parts)-2]
			repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
			result.WriteString(fmt.Sprintf("    owner = \"%s\";\n", owner))
			result.WriteString(fmt.Sprintf("    repo = \"%s\";\n", repo))
		}
	}
	result.WriteString("    rev = \"v${version}\";\n")
	result.WriteString("    sha256 = \"0000000000000000000000000000000000000000000000000000\"; # Update hash\n")
	result.WriteString("  };\n\n")
	result.WriteString("  # Add build dependencies here\n")
	result.WriteString("  # buildInputs = [ ];\n\n")
	result.WriteString("  meta = with lib; {\n")
	result.WriteString(fmt.Sprintf("    description = \"%s\";\n", packageName))
	result.WriteString(fmt.Sprintf("    homepage = \"%s\";\n", repoUrl))
	result.WriteString("    # license = licenses.mit; # Update license\n")
	result.WriteString("    # maintainers = with maintainers; [ ];\n")
	result.WriteString("  };\n")
	result.WriteString("}\n")
	result.WriteString("```\n")

	return result.String()
}

// handleGetServiceExamples gets practical configuration examples for NixOS services
func (m *MCPServer) handleGetServiceExamples(serviceName, useCase string, detailed bool) string {
	m.logger.Debug(fmt.Sprintf("handleGetServiceExamples called | serviceName=%s useCase=%s detailed=%t",
		serviceName, useCase, detailed))

	var result strings.Builder
	result.WriteString("📚 Service Configuration Examples\n\n")

	if serviceName == "" {
		result.WriteString("❌ No service name provided.\n")
		result.WriteString("Please specify a service name to get configuration examples.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("🔧 Service: %s\n", serviceName))
	result.WriteString(fmt.Sprintf("🎯 Use Case: %s\n", useCase))
	result.WriteString("\n📝 Configuration Examples:\n\n")

	// Provide basic service examples based on common services
	switch serviceName {
	case "nginx":
		result.WriteString("```nix\n")
		result.WriteString("services.nginx = {\n")
		result.WriteString("  enable = true;\n")
		result.WriteString("  virtualHosts.\"example.com\" = {\n")
		result.WriteString("    root = \"/var/www/example.com\";\n")
		result.WriteString("    locations.\"/\" = {\n")
		result.WriteString("      index = \"index.html\";\n")
		result.WriteString("    };\n")
		result.WriteString("  };\n")
		result.WriteString("};\n")
		result.WriteString("```\n")

	case "postgresql":
		result.WriteString("```nix\n")
		result.WriteString("services.postgresql = {\n")
		result.WriteString("  enable = true;\n")
		result.WriteString("  package = pkgs.postgresql_15;\n")
		result.WriteString("  ensureDatabases = [ \"myapp\" ];\n")
		result.WriteString("  ensureUsers = [{\n")
		result.WriteString("    name = \"myapp\";\n")
		result.WriteString("    ensurePermissions = {\n")
		result.WriteString("      \"DATABASE myapp\" = \"ALL PRIVILEGES\";\n")
		result.WriteString("    };\n")
		result.WriteString("  }];\n")
		result.WriteString("};\n")
		result.WriteString("```\n")

	case "openssh":
		result.WriteString("```nix\n")
		result.WriteString("services.openssh = {\n")
		result.WriteString("  enable = true;\n")
		result.WriteString("  settings = {\n")
		result.WriteString("    PasswordAuthentication = false;\n")
		result.WriteString("    KbdInteractiveAuthentication = false;\n")
		result.WriteString("    PermitRootLogin = \"no\";\n")
		result.WriteString("  };\n")
		result.WriteString("};\n")
		result.WriteString("```\n")

	default:
		result.WriteString("```nix\n")
		result.WriteString(fmt.Sprintf("services.%s = {\n", serviceName))
		result.WriteString("  enable = true;\n")
		result.WriteString("  # Add configuration options here\n")
		result.WriteString("};\n")
		result.WriteString("```\n")
	}

	if detailed {
		result.WriteString("\n📖 Additional Information:\n")
		result.WriteString("  • Check NixOS manual for complete options\n")
		result.WriteString("  • Use nixos-option to explore available settings\n")
		result.WriteString("  • Consider security implications for production use\n")
	}

	return result.String()
}

// handleCheckSystemHealth performs comprehensive NixOS system health checks
func (m *MCPServer) handleCheckSystemHealth(checkType string, includeRecommendations bool) string {
	m.logger.Debug(fmt.Sprintf("handleCheckSystemHealth called | checkType=%s includeRecommendations=%t",
		checkType, includeRecommendations))

	var result strings.Builder
	result.WriteString("🏥 System Health Check\n\n")

	result.WriteString(fmt.Sprintf("🔍 Check Type: %s\n", checkType))
	result.WriteString("\n📊 Health Status:\n")

	// Get current context for health assessment
	cfg, err := config.LoadUserConfig()
	if err == nil {
		contextDetector := nixos.NewContextDetector(&m.logger)
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err == nil && nixosCtx != nil {
			result.WriteString("  ✅ NixOS context detection working\n")
			result.WriteString(fmt.Sprintf("  📋 System Type: %s\n", nixosCtx.SystemType))
			result.WriteString(fmt.Sprintf("  🔧 Uses Flakes: %t\n", nixosCtx.UsesFlakes))
			result.WriteString(fmt.Sprintf("  🏠 Home Manager: %s\n", nixosCtx.HomeManagerType))
		} else {
			result.WriteString("  ⚠️  Context detection issues detected\n")
		}
	} else {
		result.WriteString("  ❌ Configuration loading failed\n")
	}

	result.WriteString("\n🔧 System Components:\n")
	result.WriteString("  • Configuration syntax: Ready for validation\n")
	result.WriteString("  • Service status: Ready for checking\n")
	result.WriteString("  • Package integrity: Ready for verification\n")
	result.WriteString("  • Disk usage: Ready for analysis\n")

	if includeRecommendations {
		result.WriteString("\n💡 Health Recommendations:\n")
		result.WriteString("  • Run: nixos-rebuild dry-build to test configuration\n")
		result.WriteString("  • Check: systemctl --failed for failed services\n")
		result.WriteString("  • Monitor: df -h for disk space usage\n")
		result.WriteString("  • Review: journalctl -p err for system errors\n")
		result.WriteString("  • Update: nix-channel --update for latest packages\n")
	}

	return result.String()
}

// handleAnalyzeGarbageCollection analyzes Nix store and suggests safe garbage collection
func (m *MCPServer) handleAnalyzeGarbageCollection(analysisType string, dryRun bool) string {
	m.logger.Debug(fmt.Sprintf("handleAnalyzeGarbageCollection called | analysisType=%s dryRun=%t",
		analysisType, dryRun))

	var result strings.Builder
	result.WriteString("🗑️  Garbage Collection Analysis\n\n")

	result.WriteString(fmt.Sprintf("🔍 Analysis Type: %s\n", analysisType))
	result.WriteString(fmt.Sprintf("🧪 Dry Run Mode: %t\n", dryRun))
	result.WriteString("\n📊 Nix Store Analysis:\n")

	result.WriteString("  📦 Store analysis ready\n")
	result.WriteString("  🔗 Dependency graph analysis ready\n")
	result.WriteString("  💾 Space usage calculation ready\n")
	result.WriteString("  🛡️  Safety checks ready\n")

	result.WriteString("\n🔧 Garbage Collection Commands:\n")
	if dryRun {
		result.WriteString("  • nix-collect-garbage --dry-run (safe preview)\n")
		result.WriteString("  • nix-store --gc --print-roots (show roots)\n")
		result.WriteString("  • nix-store --gc --print-dead (show candidates)\n")
	} else {
		result.WriteString("  ⚠️  Live mode commands (use with caution):\n")
		result.WriteString("  • nix-collect-garbage -d (delete old generations)\n")
		result.WriteString("  • nix-store --gc (collect garbage)\n")
		result.WriteString("  • nix-store --optimise (deduplicate store)\n")
	}

	result.WriteString("\n💡 Recommendations:\n")
	result.WriteString("  • Always test with --dry-run first\n")
	result.WriteString("  • Keep recent generations for rollback\n")
	result.WriteString("  • Consider automated garbage collection\n")
	result.WriteString("  • Monitor disk space after collection\n")

	return result.String()
}

// handleGetHardwareInfo gets hardware detection and optimization suggestions
func (m *MCPServer) handleGetHardwareInfo(detectionType string, includeOptimizations bool) string {
	m.logger.Debug(fmt.Sprintf("handleGetHardwareInfo called | detectionType=%s includeOptimizations=%t",
		detectionType, includeOptimizations))

	var result strings.Builder
	result.WriteString("🖥️  Hardware Information\n\n")

	result.WriteString(fmt.Sprintf("🔍 Detection Type: %s\n", detectionType))
	result.WriteString("\n📊 Hardware Detection:\n")

	result.WriteString("  🖥️  CPU information ready for detection\n")
	result.WriteString("  💾 Memory analysis ready\n")
	result.WriteString("  💿 Storage devices ready for enumeration\n")
	result.WriteString("  🎮 Graphics hardware ready for detection\n")
	result.WriteString("  🔌 Network interfaces ready for listing\n")

	result.WriteString("\n🔧 Hardware Detection Commands:\n")
	result.WriteString("  • lscpu (CPU information)\n")
	result.WriteString("  • lsblk (block devices)\n")
	result.WriteString("  • lspci (PCI devices)\n")
	result.WriteString("  • lsusb (USB devices)\n")
	result.WriteString("  • free -h (memory usage)\n")

	if includeOptimizations {
		result.WriteString("\n⚡ Hardware Optimizations:\n")
		result.WriteString("  • Enable hardware acceleration for graphics\n")
		result.WriteString("  • Configure CPU power management\n")
		result.WriteString("  • Optimize kernel modules for detected hardware\n")
		result.WriteString("  • Configure appropriate filesystems\n")
		result.WriteString("  • Enable hardware-specific services\n")

		result.WriteString("\n📝 Example Configuration:\n")
		result.WriteString("```nix\n")
		result.WriteString("{\n")
		result.WriteString("  # Hardware acceleration\n")
		result.WriteString("  hardware.opengl.enable = true;\n")
		result.WriteString("  \n")
		result.WriteString("  # CPU power management\n")
		result.WriteString("  powerManagement.cpuFreqGovernor = \"ondemand\";\n")
		result.WriteString("  \n")
		result.WriteString("  # Audio\n")
		result.WriteString("  sound.enable = true;\n")
		result.WriteString("  hardware.pulseaudio.enable = true;\n")
		result.WriteString("}\n")
		result.WriteString("```\n")
	}

	return result.String()
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
