package mcp

import (
	"fmt"
	"strings"
)

// Phase 2: Development & Workflow Tools Handlers

// handleCreateDevenv creates development environment using devenv templates
func (m *MCPServer) handleCreateDevenv(language, framework, projectName string, services []string) string {
	m.logger.Debug(fmt.Sprintf("handleCreateDevenv called | language=%s framework=%s projectName=%s services=%v",
		language, framework, projectName, services))

	var result strings.Builder
	result.WriteString("🚀 Development Environment Generator\n\n")

	if language == "" {
		result.WriteString("❌ No programming language specified.\n")
		result.WriteString("Please specify a programming language for the development environment.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("💻 Language: %s\n", language))
	result.WriteString(fmt.Sprintf("🏗️  Framework: %s\n", framework))
	result.WriteString(fmt.Sprintf("📦 Project: %s\n", projectName))
	result.WriteString(fmt.Sprintf("🔧 Services: %v\n", services))
	result.WriteString("\n📝 Generated devenv.nix:\n\n")

	// Generate devenv configuration
	result.WriteString("```nix\n")
	result.WriteString("{ pkgs, ... }:\n\n")
	result.WriteString("{\n")
	result.WriteString("  # Development environment configuration\n")
	result.WriteString(fmt.Sprintf("  name = \"%s\";\n\n", projectName))

	// Language-specific configuration
	switch strings.ToLower(language) {
	case "python":
		result.WriteString("  languages.python = {\n")
		result.WriteString("    enable = true;\n")
		result.WriteString("    version = \"3.11\";\n")
		result.WriteString("  };\n\n")

	case "nodejs", "javascript", "typescript":
		result.WriteString("  languages.javascript = {\n")
		result.WriteString("    enable = true;\n")
		result.WriteString("    npm.enable = true;\n")
		result.WriteString("  };\n\n")

	case "rust":
		result.WriteString("  languages.rust = {\n")
		result.WriteString("    enable = true;\n")
		result.WriteString("    channel = \"stable\";\n")
		result.WriteString("  };\n\n")

	case "go":
		result.WriteString("  languages.go = {\n")
		result.WriteString("    enable = true;\n")
		result.WriteString("    version = \"1.21\";\n")
		result.WriteString("  };\n\n")

	default:
		result.WriteString(fmt.Sprintf("  # %s language configuration\n", language))
		result.WriteString("  # Add language-specific settings here\n\n")
	}

	// Framework-specific additions
	if framework != "" {
		result.WriteString(fmt.Sprintf("  # %s framework configuration\n", framework))
		result.WriteString("  # Add framework-specific settings here\n\n")
	}

	// Services configuration
	if len(services) > 0 {
		result.WriteString("  services = {\n")
		for _, service := range services {
			switch service {
			case "postgres", "postgresql":
				result.WriteString("    postgres.enable = true;\n")
			case "redis":
				result.WriteString("    redis.enable = true;\n")
			case "mysql":
				result.WriteString("    mysql.enable = true;\n")
			default:
				result.WriteString(fmt.Sprintf("    # %s.enable = true;\n", service))
			}
		}
		result.WriteString("  };\n\n")
	}

	result.WriteString("  packages = with pkgs; [\n")
	result.WriteString("    git\n")
	result.WriteString("    curl\n")
	result.WriteString("    wget\n")
	result.WriteString("    # Add additional packages here\n")
	result.WriteString("  ];\n\n")

	result.WriteString("  enterShell = ''\n")
	result.WriteString(fmt.Sprintf("    echo \"Welcome to %s development environment!\"\n", projectName))
	result.WriteString("    echo \"Language: " + language + "\"\n")
	if framework != "" {
		result.WriteString("    echo \"Framework: " + framework + "\"\n")
	}
	result.WriteString("  '';\n")
	result.WriteString("}\n")
	result.WriteString("```\n")

	result.WriteString("\n🔧 Setup Instructions:\n")
	result.WriteString("1. Save the configuration as `devenv.nix`\n")
	result.WriteString("2. Run: `devenv shell` to enter the environment\n")
	result.WriteString("3. Customize packages and services as needed\n")

	return result.String()
}

// handleSuggestDevenvTemplate gets AI-powered development environment template suggestions
func (m *MCPServer) handleSuggestDevenvTemplate(description string, requirements []string) string {
	m.logger.Debug(fmt.Sprintf("handleSuggestDevenvTemplate called | description=%s requirements=%v",
		description, requirements))

	var result strings.Builder
	result.WriteString("💡 Development Environment Template Suggestions\n\n")

	if description == "" {
		result.WriteString("❌ No project description provided.\n")
		result.WriteString("Please provide a project description to get template suggestions.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("📋 Description: %s\n", description))
	result.WriteString(fmt.Sprintf("📝 Requirements: %v\n", requirements))
	result.WriteString("\n🎯 Template Suggestions:\n\n")

	// Analyze description for template suggestions
	desc := strings.ToLower(description)

	if strings.Contains(desc, "web") || strings.Contains(desc, "frontend") {
		result.WriteString("🌐 **Web Development Template**\n")
		result.WriteString("   • Languages: JavaScript/TypeScript\n")
		result.WriteString("   • Frameworks: React, Vue, or Angular\n")
		result.WriteString("   • Tools: Node.js, npm/yarn, webpack\n\n")
	}

	if strings.Contains(desc, "api") || strings.Contains(desc, "backend") || strings.Contains(desc, "server") {
		result.WriteString("⚙️ **Backend API Template**\n")
		result.WriteString("   • Languages: Python, Go, or Rust\n")
		result.WriteString("   • Frameworks: FastAPI, Gin, or Axum\n")
		result.WriteString("   • Services: PostgreSQL, Redis\n\n")
	}

	if strings.Contains(desc, "data") || strings.Contains(desc, "machine learning") || strings.Contains(desc, "ml") {
		result.WriteString("📊 **Data Science Template**\n")
		result.WriteString("   • Languages: Python, R\n")
		result.WriteString("   • Tools: Jupyter, pandas, numpy\n")
		result.WriteString("   • Services: PostgreSQL for data storage\n\n")
	}

	if strings.Contains(desc, "mobile") || strings.Contains(desc, "app") {
		result.WriteString("📱 **Mobile Development Template**\n")
		result.WriteString("   • Languages: Dart, JavaScript\n")
		result.WriteString("   • Frameworks: Flutter, React Native\n")
		result.WriteString("   • Tools: Android SDK, iOS tools\n\n")
	}

	// Generic full-stack template
	result.WriteString("🚀 **Full-Stack Template**\n")
	result.WriteString("   • Frontend: React/Vue + TypeScript\n")
	result.WriteString("   • Backend: Node.js or Python\n")
	result.WriteString("   • Database: PostgreSQL\n")
	result.WriteString("   • Cache: Redis\n")
	result.WriteString("   • Tools: Docker, git\n\n")

	result.WriteString("💡 **Recommendation:**\n")
	result.WriteString("Based on your description, consider starting with the ")
	if strings.Contains(desc, "web") {
		result.WriteString("Web Development Template")
	} else if strings.Contains(desc, "api") {
		result.WriteString("Backend API Template")
	} else if strings.Contains(desc, "data") {
		result.WriteString("Data Science Template")
	} else {
		result.WriteString("Full-Stack Template")
	}
	result.WriteString(" and customize as needed.\n")

	return result.String()
}

// handleSetupNeovimIntegration sets up and configures Neovim integration with nixai MCP
func (m *MCPServer) handleSetupNeovimIntegration(configType, socketPath string) string {
	m.logger.Debug(fmt.Sprintf("handleSetupNeovimIntegration called | configType=%s socketPath=%s",
		configType, socketPath))

	var result strings.Builder
	result.WriteString("🎯 Neovim Integration Setup\n\n")

	if configType == "" {
		configType = "lua"
	}
	if socketPath == "" {
		socketPath = "/tmp/nixai-mcp.sock"
	}

	result.WriteString(fmt.Sprintf("🔧 Configuration Type: %s\n", configType))
	result.WriteString(fmt.Sprintf("🔌 Socket Path: %s\n", socketPath))
	result.WriteString("\n📝 Neovim Configuration:\n\n")

	if configType == "lua" {
		result.WriteString("```lua\n")
		result.WriteString("-- nixai MCP integration for Neovim\n")
		result.WriteString("local M = {}\n\n")

		result.WriteString("-- Configuration\n")
		result.WriteString(fmt.Sprintf("M.socket_path = '%s'\n", socketPath))
		result.WriteString("M.timeout = 5000\n\n")

		result.WriteString("-- Call MCP function\n")
		result.WriteString("function M.call_mcp(tool_name, args)\n")
		result.WriteString("  local cmd = string.format(\n")
		result.WriteString("    'echo \\'{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"tools/call\", \"params\": {\"name\": \"%s\", \"arguments\": %s}}\\' | socat - UNIX-CONNECT:%s',\n")
		result.WriteString("    tool_name, vim.fn.json_encode(args or {}), M.socket_path\n")
		result.WriteString("  )\n")
		result.WriteString("  \n")
		result.WriteString("  local result = vim.fn.system(cmd)\n")
		result.WriteString("  local success, parsed = pcall(vim.fn.json_decode, result)\n")
		result.WriteString("  \n")
		result.WriteString("  if success and parsed.result then\n")
		result.WriteString("    return parsed.result.content[1].text, nil\n")
		result.WriteString("  else\n")
		result.WriteString("    return nil, 'MCP call failed'\n")
		result.WriteString("  end\n")
		result.WriteString("end\n\n")

		result.WriteString("-- Get NixOS context\n")
		result.WriteString("function M.get_context(format, detailed)\n")
		result.WriteString("  local args = {}\n")
		result.WriteString("  if format then args.format = format end\n")
		result.WriteString("  if detailed then args.detailed = detailed end\n")
		result.WriteString("  \n")
		result.WriteString("  local result, err = M.call_mcp('get_nixos_context', args)\n")
		result.WriteString("  if err then\n")
		result.WriteString("    vim.notify('NixAI Error: ' .. err, vim.log.levels.ERROR)\n")
		result.WriteString("    return nil\n")
		result.WriteString("  end\n")
		result.WriteString("  return result\n")
		result.WriteString("end\n\n")

		result.WriteString("-- Ask nixai question\n")
		result.WriteString("function M.ask_question(question)\n")
		result.WriteString("  if not question or question == '' then\n")
		result.WriteString("    question = vim.fn.input('Ask nixai: ')\n")
		result.WriteString("  end\n")
		result.WriteString("  \n")
		result.WriteString("  if question == '' then return end\n")
		result.WriteString("  \n")
		result.WriteString("  local args = { query = question }\n")
		result.WriteString("  local result, err = M.call_mcp('query_nixos_docs', args)\n")
		result.WriteString("  \n")
		result.WriteString("  if err then\n")
		result.WriteString("    vim.notify('NixAI Error: ' .. err, vim.log.levels.ERROR)\n")
		result.WriteString("  else\n")
		result.WriteString("    -- Show result in new buffer\n")
		result.WriteString("    local buf = vim.api.nvim_create_buf(false, true)\n")
		result.WriteString("    local lines = vim.split(result, '\\n')\n")
		result.WriteString("    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)\n")
		result.WriteString("    vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')\n")
		result.WriteString("    vim.api.nvim_win_set_buf(0, buf)\n")
		result.WriteString("  end\n")
		result.WriteString("end\n\n")

		result.WriteString("-- Key mappings\n")
		result.WriteString("vim.keymap.set('n', '<leader>na', M.ask_question, { desc = 'Ask nixai' })\n")
		result.WriteString("vim.keymap.set('n', '<leader>nc', function() M.get_context('text', false) end, { desc = 'Get NixOS context' })\n\n")

		result.WriteString("return M\n")
		result.WriteString("```\n")

	} else if configType == "vimscript" {
		result.WriteString("```vim\n")
		result.WriteString("\" nixai MCP integration for Neovim\n")
		result.WriteString(fmt.Sprintf("let g:nixai_socket_path = '%s'\n", socketPath))
		result.WriteString("let g:nixai_timeout = 5000\n\n")

		result.WriteString("function! NixaiAsk(question)\n")
		result.WriteString("  if empty(a:question)\n")
		result.WriteString("    let question = input('Ask nixai: ')\n")
		result.WriteString("  else\n")
		result.WriteString("    let question = a:question\n")
		result.WriteString("  endif\n\n")

		result.WriteString("  if empty(question)\n")
		result.WriteString("    return\n")
		result.WriteString("  endif\n\n")

		result.WriteString("  let cmd = printf(\n")
		result.WriteString("    \\ 'echo ''{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"tools/call\", \"params\": {\"name\": \"query_nixos_docs\", \"arguments\": {\"query\": \"%s\"}}}'' | socat - UNIX-CONNECT:%s',\n")
		result.WriteString("    \\ shellescape(question), g:nixai_socket_path)\n\n")

		result.WriteString("  let result = system(cmd)\n")
		result.WriteString("  echo result\n")
		result.WriteString("endfunction\n\n")

		result.WriteString("\" Key mappings\n")
		result.WriteString("nnoremap <leader>na :call NixaiAsk('')<CR>\n")
		result.WriteString("```\n")
	}

	result.WriteString("\n🔧 Setup Instructions:\n")
	result.WriteString("1. Add the configuration to your Neovim config\n")
	result.WriteString("2. Ensure the MCP server is running\n")
	result.WriteString("3. Use <leader>na to ask questions\n")
	result.WriteString("4. Use <leader>nc to get system context\n")

	return result.String()
}

// handleFlakeOperations performs NixOS flake operations and management
func (m *MCPServer) handleFlakeOperations(operation, flakePath string, options []string) string {
	m.logger.Debug(fmt.Sprintf("handleFlakeOperations called | operation=%s flakePath=%s options=%v",
		operation, flakePath, options))

	var result strings.Builder
	result.WriteString("❄️  Flake Operations\n\n")

	if operation == "" {
		result.WriteString("❌ No operation specified.\n")
		result.WriteString("Available operations: init, update, build, check, show\n")
		return result.String()
	}

	if flakePath == "" {
		flakePath = "."
	}

	result.WriteString(fmt.Sprintf("🔧 Operation: %s\n", operation))
	result.WriteString(fmt.Sprintf("📁 Flake Path: %s\n", flakePath))
	result.WriteString(fmt.Sprintf("⚙️  Options: %v\n", options))
	result.WriteString("\n📋 Operation Details:\n\n")

	switch operation {
	case "init":
		result.WriteString("🆕 **Initialize New Flake**\n")
		result.WriteString("Command: `nix flake init`\n")
		result.WriteString("Creates a basic flake.nix template\n\n")

		result.WriteString("📝 Basic flake.nix template:\n")
		result.WriteString("```nix\n")
		result.WriteString("{\n")
		result.WriteString("  description = \"A NixOS flake\";\n\n")
		result.WriteString("  inputs = {\n")
		result.WriteString("    nixpkgs.url = \"github:NixOS/nixpkgs/nixos-unstable\";\n")
		result.WriteString("  };\n\n")
		result.WriteString("  outputs = { self, nixpkgs }: {\n")
		result.WriteString("    nixosConfigurations.hostname = nixpkgs.lib.nixosSystem {\n")
		result.WriteString("      system = \"x86_64-linux\";\n")
		result.WriteString("      modules = [ ./configuration.nix ];\n")
		result.WriteString("    };\n")
		result.WriteString("  };\n")
		result.WriteString("}\n")
		result.WriteString("```\n")

	case "update":
		result.WriteString("🔄 **Update Flake Inputs**\n")
		result.WriteString("Command: `nix flake update`\n")
		result.WriteString("Updates all inputs to their latest versions\n")
		result.WriteString("Creates/updates flake.lock file\n\n")

	case "build":
		result.WriteString("🔨 **Build Flake**\n")
		result.WriteString("Command: `nix build`\n")
		result.WriteString("Builds the default package/system\n")
		result.WriteString("Options for specific outputs available\n\n")

	case "check":
		result.WriteString("✅ **Check Flake**\n")
		result.WriteString("Command: `nix flake check`\n")
		result.WriteString("Validates flake syntax and evaluability\n")
		result.WriteString("Checks all outputs for errors\n\n")

	case "show":
		result.WriteString("👁️  **Show Flake Info**\n")
		result.WriteString("Command: `nix flake show`\n")
		result.WriteString("Displays flake metadata and outputs\n")
		result.WriteString("Shows available packages and systems\n\n")

	default:
		result.WriteString(fmt.Sprintf("❓ Unknown operation: %s\n", operation))
		result.WriteString("Available operations:\n")
		result.WriteString("  • init - Initialize new flake\n")
		result.WriteString("  • update - Update flake inputs\n")
		result.WriteString("  • build - Build flake\n")
		result.WriteString("  • check - Check flake validity\n")
		result.WriteString("  • show - Show flake information\n")
	}

	if len(options) > 0 {
		result.WriteString("🔧 Additional Options:\n")
		for _, opt := range options {
			result.WriteString(fmt.Sprintf("  • %s\n", opt))
		}
	}

	return result.String()
}

// handleMigrateToFlakes migrates NixOS configuration from channels to flakes
func (m *MCPServer) handleMigrateToFlakes(backupName string, dryRun, includeHomeManager bool) string {
	m.logger.Debug(fmt.Sprintf("handleMigrateToFlakes called | backupName=%s dryRun=%t includeHomeManager=%t",
		backupName, dryRun, includeHomeManager))

	var result strings.Builder
	result.WriteString("🔄 Migrate to Flakes\n\n")

	if backupName == "" {
		backupName = "pre-flake-backup"
	}

	result.WriteString(fmt.Sprintf("💾 Backup Name: %s\n", backupName))
	result.WriteString(fmt.Sprintf("🧪 Dry Run: %t\n", dryRun))
	result.WriteString(fmt.Sprintf("🏠 Include Home Manager: %t\n", includeHomeManager))
	result.WriteString("\n📋 Migration Steps:\n\n")

	result.WriteString("**1. Backup Current Configuration**\n")
	if dryRun {
		result.WriteString("   [DRY RUN] Would backup /etc/nixos to " + backupName + "\n")
	} else {
		result.WriteString("   • cp -r /etc/nixos /etc/nixos-backup-" + backupName + "\n")
	}
	result.WriteString("\n")

	result.WriteString("**2. Create flake.nix**\n")
	result.WriteString("   📝 Generated flake.nix:\n")
	result.WriteString("   ```nix\n")
	result.WriteString("   {\n")
	result.WriteString("     description = \"NixOS configuration flake\";\n\n")
	result.WriteString("     inputs = {\n")
	result.WriteString("       nixpkgs.url = \"github:NixOS/nixpkgs/nixos-unstable\";\n")
	if includeHomeManager {
		result.WriteString("       home-manager = {\n")
		result.WriteString("         url = \"github:nix-community/home-manager\";\n")
		result.WriteString("         inputs.nixpkgs.follows = \"nixpkgs\";\n")
		result.WriteString("       };\n")
	}
	result.WriteString("     };\n\n")
	result.WriteString("     outputs = { self, nixpkgs")
	if includeHomeManager {
		result.WriteString(", home-manager")
	}
	result.WriteString(" }: {\n")
	result.WriteString("       nixosConfigurations.hostname = nixpkgs.lib.nixosSystem {\n")
	result.WriteString("         system = \"x86_64-linux\";\n")
	result.WriteString("         modules = [\n")
	result.WriteString("           ./hardware-configuration.nix\n")
	result.WriteString("           ./configuration.nix\n")
	if includeHomeManager {
		result.WriteString("           home-manager.nixosModules.home-manager\n")
		result.WriteString("           {\n")
		result.WriteString("             home-manager.useGlobalPkgs = true;\n")
		result.WriteString("             home-manager.useUserPackages = true;\n")
		result.WriteString("           }\n")
	}
	result.WriteString("         ];\n")
	result.WriteString("       };\n")
	result.WriteString("     };\n")
	result.WriteString("   }\n")
	result.WriteString("   ```\n\n")

	result.WriteString("**3. Update Configuration**\n")
	result.WriteString("   • Remove channel-specific imports\n")
	result.WriteString("   • Update package references if needed\n")
	result.WriteString("   • Test configuration syntax\n\n")

	result.WriteString("**4. Build and Test**\n")
	if dryRun {
		result.WriteString("   [DRY RUN] Would run: nixos-rebuild dry-build --flake .\n")
	} else {
		result.WriteString("   • nixos-rebuild dry-build --flake .\n")
		result.WriteString("   • nixos-rebuild test --flake .\n")
	}
	result.WriteString("\n")

	result.WriteString("**5. Switch to Flakes**\n")
	if dryRun {
		result.WriteString("   [DRY RUN] Would run: nixos-rebuild switch --flake .\n")
	} else {
		result.WriteString("   • nixos-rebuild switch --flake .\n")
	}
	result.WriteString("\n")

	result.WriteString("⚠️  **Important Notes:**\n")
	result.WriteString("• Test thoroughly before switching permanently\n")
	result.WriteString("• Keep backup until migration is confirmed working\n")
	result.WriteString("• Update bootloader to use flake path\n")
	if includeHomeManager {
		result.WriteString("• Migrate Home Manager configuration separately if needed\n")
	}

	return result.String()
}

// handleAnalyzeDependencies analyzes NixOS configuration dependencies and relationships
func (m *MCPServer) handleAnalyzeDependencies(configPath, scope, format string) string {
	m.logger.Debug(fmt.Sprintf("handleAnalyzeDependencies called | configPath=%s scope=%s format=%s",
		configPath, scope, format))

	var result strings.Builder
	result.WriteString("🔗 Dependency Analysis\n\n")

	if configPath == "" {
		configPath = "/etc/nixos/configuration.nix"
	}
	if scope == "" {
		scope = "full"
	}
	if format == "" {
		format = "text"
	}

	result.WriteString(fmt.Sprintf("📁 Configuration: %s\n", configPath))
	result.WriteString(fmt.Sprintf("🔍 Scope: %s\n", scope))
	result.WriteString(fmt.Sprintf("📄 Format: %s\n", format))
	result.WriteString("\n📊 Dependency Analysis:\n\n")

	result.WriteString("**Configuration Dependencies:**\n")
	result.WriteString("  • hardware-configuration.nix\n")
	result.WriteString("  • System packages from nixpkgs\n")
	result.WriteString("  • Service configurations\n")
	result.WriteString("  • User configurations\n\n")

	result.WriteString("**Package Dependencies:**\n")
	result.WriteString("  • Direct package references\n")
	result.WriteString("  • Service-required packages\n")
	result.WriteString("  • Development tools\n")
	result.WriteString("  • System utilities\n\n")

	result.WriteString("**Service Dependencies:**\n")
	result.WriteString("  • Inter-service dependencies\n")
	result.WriteString("  • Network dependencies\n")
	result.WriteString("  • File system dependencies\n")
	result.WriteString("  • User/group dependencies\n\n")

	result.WriteString("🔧 Analysis Commands:\n")
	result.WriteString("  • nix-store --query --references (package refs)\n")
	result.WriteString("  • nix-store --query --referrers (reverse deps)\n")
	result.WriteString("  • nix-store --query --tree (dependency tree)\n")
	result.WriteString("  • systemd-analyze (service dependencies)\n\n")

	result.WriteString("💡 Optimization Suggestions:\n")
	result.WriteString("  • Remove unused packages\n")
	result.WriteString("  • Consolidate similar services\n")
	result.WriteString("  • Review optional dependencies\n")
	result.WriteString("  • Consider package overlays for customizations\n")

	return result.String()
}

// handleExplainDependencyChain explains why a specific package is included in the system
func (m *MCPServer) handleExplainDependencyChain(packageName, depth, includeOptional string) string {
	m.logger.Debug(fmt.Sprintf("handleExplainDependencyChain called | packageName=%s depth=%s includeOptional=%s",
		packageName, depth, includeOptional))

	var result strings.Builder
	result.WriteString("🔍 Dependency Chain Analysis\n\n")

	if packageName == "" {
		result.WriteString("❌ No package name provided.\n")
		result.WriteString("Please specify a package name to analyze its dependency chain.\n")
		return result.String()
	}

	if depth == "" {
		depth = "5"
	}

	result.WriteString(fmt.Sprintf("📦 Package: %s\n", packageName))
	result.WriteString(fmt.Sprintf("🔢 Analysis Depth: %s\n", depth))
	result.WriteString(fmt.Sprintf("🔧 Include Optional: %s\n", includeOptional))
	result.WriteString("\n🔗 Dependency Chain:\n\n")

	result.WriteString(fmt.Sprintf("**Why %s is in your system:**\n\n", packageName))

	result.WriteString("1. **Direct Reference**\n")
	result.WriteString("   • Explicitly listed in environment.systemPackages\n")
	result.WriteString("   • Required by enabled services\n")
	result.WriteString("   • Part of system configuration\n\n")

	result.WriteString("2. **Indirect Dependencies**\n")
	result.WriteString("   • Runtime dependency of another package\n")
	result.WriteString("   • Build-time dependency\n")
	result.WriteString("   • Optional dependency pulled in\n\n")

	result.WriteString("3. **Service Dependencies**\n")
	result.WriteString("   • Required by systemd services\n")
	result.WriteString("   • Network service dependencies\n")
	result.WriteString("   • User session dependencies\n\n")

	result.WriteString("🔧 Investigation Commands:\n")
	result.WriteString(fmt.Sprintf("  • nix why-depends /run/current-system %s\n", packageName))
	result.WriteString(fmt.Sprintf("  • nix-store --query --referrers $(which %s)\n", packageName))
	result.WriteString("  • nix-store --query --tree /run/current-system | grep " + packageName + "\n\n")

	result.WriteString("💡 **Next Steps:**\n")
	result.WriteString("• Use the investigation commands to get actual dependency paths\n")
	result.WriteString("• Check if the package is really needed\n")
	result.WriteString("• Consider alternatives if removing the dependency\n")
	result.WriteString("• Review service configurations that might require it\n")

	return result.String()
}

// handleStoreOperations performs Nix store backup, restore, and analysis operations
func (m *MCPServer) handleStoreOperations(operation string, paths, options []string) string {
	m.logger.Debug(fmt.Sprintf("handleStoreOperations called | operation=%s paths=%v options=%v",
		operation, paths, options))

	var result strings.Builder
	result.WriteString("🗃️  Nix Store Operations\n\n")

	if operation == "" {
		result.WriteString("❌ No operation specified.\n")
		result.WriteString("Available operations: backup, restore, analyze, verify, optimize\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("🔧 Operation: %s\n", operation))
	result.WriteString(fmt.Sprintf("📁 Paths: %v\n", paths))
	result.WriteString(fmt.Sprintf("⚙️  Options: %v\n", options))
	result.WriteString("\n📋 Operation Details:\n\n")

	switch operation {
	case "backup":
		result.WriteString("💾 **Store Backup**\n")
		result.WriteString("Commands:\n")
		result.WriteString("  • nix copy --to file:///backup/path /nix/store/...\n")
		result.WriteString("  • nix-store --export $(nix-store -qR path) > backup.nar\n")
		result.WriteString("  • rsync -av /nix/store/ /backup/nix-store/\n\n")

	case "restore":
		result.WriteString("📥 **Store Restore**\n")
		result.WriteString("Commands:\n")
		result.WriteString("  • nix copy --from file:///backup/path store-path\n")
		result.WriteString("  • nix-store --import < backup.nar\n")
		result.WriteString("  • rsync -av /backup/nix-store/ /nix/store/\n\n")

	case "analyze":
		result.WriteString("📊 **Store Analysis**\n")
		result.WriteString("Commands:\n")
		result.WriteString("  • nix path-info --store-size -rS /run/current-system\n")
		result.WriteString("  • du -sh /nix/store\n")
		result.WriteString("  • nix-store --query --tree /run/current-system\n")
		result.WriteString("  • nix store diff-closures old-system new-system\n\n")

	case "verify":
		result.WriteString("✅ **Store Verification**\n")
		result.WriteString("Commands:\n")
		result.WriteString("  • nix store verify --all\n")
		result.WriteString("  • nix-store --verify --check-contents\n")
		result.WriteString("  • nix-store --verify --repair\n\n")

	case "optimize":
		result.WriteString("⚡ **Store Optimization**\n")
		result.WriteString("Commands:\n")
		result.WriteString("  • nix-store --optimise (deduplicate)\n")
		result.WriteString("  • nix-collect-garbage -d (remove old generations)\n")
		result.WriteString("  • nix store optimise (new command)\n\n")

	default:
		result.WriteString(fmt.Sprintf("❓ Unknown operation: %s\n", operation))
		result.WriteString("Available operations:\n")
		result.WriteString("  • backup - Backup store paths\n")
		result.WriteString("  • restore - Restore store paths\n")
		result.WriteString("  • analyze - Analyze store usage\n")
		result.WriteString("  • verify - Verify store integrity\n")
		result.WriteString("  • optimize - Optimize store space\n")
	}

	result.WriteString("⚠️  **Safety Notes:**\n")
	result.WriteString("• Always backup before major operations\n")
	result.WriteString("• Test restore procedures regularly\n")
	result.WriteString("• Monitor disk space during operations\n")
	result.WriteString("• Use --dry-run when available\n")

	return result.String()
}

// handlePerformanceAnalysis analyzes NixOS system performance and suggests optimizations
func (m *MCPServer) handlePerformanceAnalysis(analysisType string, metrics []string, suggestions bool) string {
	m.logger.Debug(fmt.Sprintf("handlePerformanceAnalysis called | analysisType=%s metrics=%v suggestions=%t",
		analysisType, metrics, suggestions))

	var result strings.Builder
	result.WriteString("📈 Performance Analysis\n\n")

	if analysisType == "" {
		analysisType = "general"
	}

	result.WriteString(fmt.Sprintf("🔍 Analysis Type: %s\n", analysisType))
	result.WriteString(fmt.Sprintf("📊 Metrics: %v\n", metrics))
	result.WriteString(fmt.Sprintf("💡 Include Suggestions: %t\n", suggestions))
	result.WriteString("\n📋 Performance Assessment:\n\n")

	result.WriteString("**System Performance Areas:**\n")
	result.WriteString("  • Boot time and startup services\n")
	result.WriteString("  • Memory usage and management\n")
	result.WriteString("  • CPU utilization and scheduling\n")
	result.WriteString("  • I/O performance and disk usage\n")
	result.WriteString("  • Network performance\n\n")

	result.WriteString("🔧 **Analysis Commands:**\n")
	result.WriteString("  • systemd-analyze (boot performance)\n")
	result.WriteString("  • systemd-analyze blame (slow services)\n")
	result.WriteString("  • free -h (memory usage)\n")
	result.WriteString("  • top/htop (CPU and process monitoring)\n")
	result.WriteString("  • iotop (I/O monitoring)\n")
	result.WriteString("  • nethogs (network usage)\n\n")

	if suggestions {
		result.WriteString("⚡ **Performance Optimizations:**\n\n")

		result.WriteString("**Boot Performance:**\n")
		result.WriteString("  • Disable unnecessary services\n")
		result.WriteString("  • Use systemd service dependencies properly\n")
		result.WriteString("  • Enable parallel service startup\n")
		result.WriteString("  • Optimize initrd modules\n\n")

		result.WriteString("**Memory Optimization:**\n")
		result.WriteString("  • Enable zram for swap compression\n")
		result.WriteString("  • Adjust swappiness value\n")
		result.WriteString("  • Configure appropriate swap size\n")
		result.WriteString("  • Use memory-efficient desktop environments\n\n")

		result.WriteString("**CPU Optimization:**\n")
		result.WriteString("  • Set appropriate CPU governor\n")
		result.WriteString("  • Enable CPU microcode updates\n")
		result.WriteString("  • Configure CPU scaling\n")
		result.WriteString("  • Use performance-oriented kernel\n\n")

		result.WriteString("**I/O Optimization:**\n")
		result.WriteString("  • Use SSD-optimized filesystems (ext4, btrfs)\n")
		result.WriteString("  • Enable fstrim for SSDs\n")
		result.WriteString("  • Optimize mount options\n")
		result.WriteString("  • Configure appropriate I/O schedulers\n\n")

		result.WriteString("📝 **Example Configuration:**\n")
		result.WriteString("```nix\n")
		result.WriteString("{\n")
		result.WriteString("  # Boot optimization\n")
		result.WriteString("  boot.kernelParams = [ \"quiet\" \"loglevel=3\" ];\n")
		result.WriteString("  boot.loader.timeout = 1;\n\n")

		result.WriteString("  # Memory optimization\n")
		result.WriteString("  zramSwap.enable = true;\n")
		result.WriteString("  boot.kernel.sysctl.\"vm.swappiness\" = 10;\n\n")

		result.WriteString("  # CPU optimization\n")
		result.WriteString("  powerManagement.cpuFreqGovernor = \"performance\";\n")
		result.WriteString("  hardware.cpu.intel.updateMicrocode = true;\n\n")

		result.WriteString("  # I/O optimization\n")
		result.WriteString("  services.fstrim.enable = true;\n")
		result.WriteString("}\n")
		result.WriteString("```\n")
	}

	return result.String()
}

// handleSearchAdvanced performs advanced multi-source search for packages, options, and configurations
func (m *MCPServer) handleSearchAdvanced(query string, sources []string, filters map[string]string) string {
	m.logger.Debug(fmt.Sprintf("handleSearchAdvanced called | query=%s sources=%v filters=%v",
		query, sources, filters))

	var result strings.Builder
	result.WriteString("🔍 Advanced Search\n\n")

	if query == "" {
		result.WriteString("❌ No search query provided.\n")
		result.WriteString("Please provide a search query.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("🔎 Query: %s\n", query))
	result.WriteString(fmt.Sprintf("📚 Sources: %v\n", sources))
	result.WriteString(fmt.Sprintf("🔧 Filters: %v\n", filters))
	result.WriteString("\n📊 Search Results:\n\n")

	// Default sources if none specified
	if len(sources) == 0 {
		sources = []string{"packages", "options", "wiki"}
	}

	for _, source := range sources {
		result.WriteString(fmt.Sprintf("**%s Search Results:**\n", strings.Title(source)))

		switch source {
		case "packages":
			result.WriteString("  📦 Package search results would appear here\n")
			result.WriteString(fmt.Sprintf("     Command: nix search nixpkgs %s\n", query))

		case "options":
			result.WriteString("  ⚙️  NixOS option search results would appear here\n")
			result.WriteString(fmt.Sprintf("     Command: nixos-option -r %s\n", query))

		case "wiki":
			result.WriteString("  📖 Wiki search results would appear here\n")
			result.WriteString("     Source: https://wiki.nixos.org\n")

		case "home-manager":
			result.WriteString("  🏠 Home Manager option search results would appear here\n")
			result.WriteString("     Source: Home Manager options database\n")

		case "manual":
			result.WriteString("  📚 Manual search results would appear here\n")
			result.WriteString("     Source: NixOS manual and documentation\n")

		default:
			result.WriteString(fmt.Sprintf("  ❓ Unknown source: %s\n", source))
		}
		result.WriteString("\n")
	}

	// Apply filters if specified
	if len(filters) > 0 {
		result.WriteString("🔧 **Applied Filters:**\n")
		for key, value := range filters {
			result.WriteString(fmt.Sprintf("  • %s: %s\n", key, value))
		}
		result.WriteString("\n")
	}

	result.WriteString("💡 **Search Tips:**\n")
	result.WriteString("• Use specific keywords for better results\n")
	result.WriteString("• Combine multiple sources for comprehensive search\n")
	result.WriteString("• Apply filters to narrow down results\n")
	result.WriteString("• Check both packages and options for complete coverage\n")

	return result.String()
}
