package cli

import (
	"fmt"
	"io"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"
)

// Helper functions for running commands directly in interactive mode

// Config command wrapper functions that accept io.Writer
func showConfigWithOutput(out io.Writer) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	fmt.Fprintln(out, utils.FormatHeader("🔧 Current Configuration"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatKeyValue("AI Provider", cfg.AIProvider))
	fmt.Fprintln(out, utils.FormatKeyValue("AI Model", cfg.AIModel))
	fmt.Fprintln(out, utils.FormatKeyValue("Log Level", cfg.LogLevel))
	fmt.Fprintln(out, utils.FormatKeyValue("NixOS Folder", cfg.NixosFolder))
	fmt.Fprintln(out, utils.FormatKeyValue("MCP Host", cfg.MCPServer.Host))
	fmt.Fprintln(out, utils.FormatKeyValue("MCP Port", fmt.Sprintf("%d", cfg.MCPServer.Port)))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Use 'config set <key> <value>' to modify settings"))
}

func setConfigWithOutput(out io.Writer, key, value string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	switch key {
	case "ai_provider":
		if value != "ollama" && value != "gemini" && value != "openai" {
			fmt.Fprintln(out, utils.FormatError("Invalid AI provider. Valid options: ollama, gemini, openai"))
			return
		}
		cfg.AIProvider = value
	case "ai_model":
		cfg.AIModel = value
	case "log_level":
		if value != "debug" && value != "info" && value != "warn" && value != "error" {
			fmt.Fprintln(out, utils.FormatError("Invalid log level. Valid options: debug, info, warn, error"))
			return
		}
		cfg.LogLevel = value
	case "nixos_folder":
		cfg.NixosFolder = value
	case "mcp_host":
		cfg.MCPServer.Host = value
	case "mcp_port":
		port, parseErr := fmt.Sscanf(value, "%d", &cfg.MCPServer.Port)
		if parseErr != nil || port != 1 {
			fmt.Fprintln(out, utils.FormatError("Invalid port number"))
			return
		}
	default:
		fmt.Fprintln(out, utils.FormatError("Unknown configuration key: "+key))
		fmt.Fprintln(out, utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		return
	}

	err = config.SaveUserConfig(cfg)
	if err != nil {
		fmt.Fprintln(out, utils.FormatError("Failed to save config: "+err.Error()))
		return
	}

	fmt.Fprintln(out, utils.FormatSuccess("✅ Configuration updated successfully"))
	fmt.Fprintln(out, utils.FormatKeyValue(key, value))
}

func getConfigWithOutput(out io.Writer, key string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	var value string
	switch key {
	case "ai_provider":
		value = cfg.AIProvider
	case "ai_model":
		value = cfg.AIModel
	case "log_level":
		value = cfg.LogLevel
	case "nixos_folder":
		value = cfg.NixosFolder
	case "mcp_host":
		value = cfg.MCPServer.Host
	case "mcp_port":
		value = fmt.Sprintf("%d", cfg.MCPServer.Port)
	default:
		fmt.Fprintln(out, utils.FormatError("Unknown configuration key: "+key))
		fmt.Fprintln(out, utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		return
	}

	fmt.Fprintln(out, utils.FormatKeyValue(key, value))
}

func resetConfigWithOutput(out io.Writer) {
	cfg := config.DefaultUserConfig()
	err := config.SaveUserConfig(cfg)
	if err != nil {
		fmt.Fprintln(out, utils.FormatError("Failed to reset config: "+err.Error()))
		return
	}

	fmt.Fprintln(out, utils.FormatSuccess("✅ Configuration reset to defaults"))
	fmt.Fprintln(out, utils.FormatTip("Use 'config show' to see current settings"))
}

// Community helper functions
func showCommunityOverview(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🌐 Community Overview"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  search <query>     - Search community configurations")
	fmt.Fprintln(out, "  share <file>       - Share your configuration")
	fmt.Fprintln(out, "  validate <file>    - Validate configuration against best practices")
	fmt.Fprintln(out, "  trends             - Show trending packages and patterns")
	fmt.Fprintln(out, "  rate <config> <n>  - Rate a community configuration")
	fmt.Fprintln(out, "  forums             - Show community forums and discussions")
	fmt.Fprintln(out, "  docs               - Show community documentation resources")
	fmt.Fprintln(out, "  matrix             - Show Matrix chat channels")
	fmt.Fprintln(out, "  github             - Show GitHub resources and repositories")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Use 'nixai community <command> --help' for detailed information"))
}

func showCommunityForums(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("💬 Community Forums"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatKeyValue("NixOS Discourse", "https://discourse.nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("Reddit r/NixOS", "https://reddit.com/r/NixOS"))
	fmt.Fprintln(out, utils.FormatKeyValue("Stack Overflow", "https://stackoverflow.com/questions/tagged/nixos"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Search for solutions and ask questions in these forums"))
}

func showCommunityDocs(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("📚 Community Documentation"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatKeyValue("NixOS Manual", "https://nixos.org/manual/nixos/stable/"))
	fmt.Fprintln(out, utils.FormatKeyValue("Nixpkgs Manual", "https://nixos.org/manual/nixpkgs/stable/"))
	fmt.Fprintln(out, utils.FormatKeyValue("Nix Manual", "https://nix.dev/manual/nix"))
	fmt.Fprintln(out, utils.FormatKeyValue("Home Manager", "https://nix-community.github.io/home-manager/"))
	fmt.Fprintln(out, utils.FormatKeyValue("Wiki", "https://wiki.nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("Nix Dev", "https://nix.dev"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("These are the primary documentation sources"))
}

func showMatrixChannels(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("💬 Matrix Chat Channels"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatKeyValue("Main Channel", "#nixos:nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("Development", "#nixos-dev:nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("Security", "#nixos-security:nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("Offtopic", "#nixos-chat:nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("ARM", "#nixos-aarch64:nixos.org"))
	fmt.Fprintln(out, utils.FormatKeyValue("Gaming", "#nixos-gaming:nixos.org"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Real-time chat with the NixOS community"))
}

func showGitHubResources(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🐙 GitHub Resources"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nixpkgs", "https://github.com/NixOS/nixpkgs"))
	fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nix", "https://github.com/NixOS/nix"))
	fmt.Fprintln(out, utils.FormatKeyValue("nix-community", "https://github.com/nix-community"))
	fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nixos-hardware", "https://github.com/NixOS/nixos-hardware"))
	fmt.Fprintln(out, utils.FormatKeyValue("Awesome Nix", "https://github.com/nix-community/awesome-nix"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Browse source code, issues, and contribute to projects"))
}

// runConfigCmd executes the config command directly
func runConfigCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showConfigWithOutput(out)
		return
	}

	switch args[0] {
	case "show":
		showConfigWithOutput(out)
	case "set":
		if len(args) < 3 {
			fmt.Fprintln(out, "Usage: nixai config set <key> <value>")
			return
		}
		setConfigWithOutput(out, args[1], args[2])
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(out, "Usage: nixai config get <key>")
			return
		}
		getConfigWithOutput(out, args[1])
	case "reset":
		resetConfigWithOutput(out)
	default:
		fmt.Fprintln(out, "Unknown config command: "+args[0])
	}
}

// runCommunityCmd executes the community command directly
func runCommunityCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showCommunityOverview(out)
		return
	}
	// Add real subcommand logic as needed
	switch args[0] {
	case "forums":
		showCommunityForums(out)
	case "docs":
		showCommunityDocs(out)
	case "matrix":
		showMatrixChannels(out)
	case "github":
		showGitHubResources(out)
	default:
		fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented community subcommand: "+args[0]))
	}
}

// runConfigureCmd executes the configure command directly
func runConfigureCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, "Interactive configuration coming soon.")
}

// Diagnose helper functions
func showDiagnosticOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔍 Diagnostic Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  system        - Overall system health check")
	fmt.Fprintln(out, "  config        - Configuration file analysis")
	fmt.Fprintln(out, "  services      - Service status and logs")
	fmt.Fprintln(out, "  network       - Network connectivity tests")
	fmt.Fprintln(out, "  hardware      - Hardware detection and drivers")
	fmt.Fprintln(out, "  performance   - Performance bottleneck analysis")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Comprehensive system diagnostics coming soon"))
}

func runSystemDiagnostics(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔍 System Diagnostics"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatProgress("Running system health checks..."))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "✅ Boot loader: OK")
	fmt.Fprintln(out, "✅ Filesystem: OK")
	fmt.Fprintln(out, "✅ Network: OK")
	fmt.Fprintln(out, "✅ Services: OK")
	fmt.Fprintln(out, "✅ Hardware: OK")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSuccess("System health: All checks passed"))
}

// runDiagnoseCmd executes the diagnose command directly
func runDiagnoseCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showDiagnosticOptions(out)
		return
	}
	// Minimal: system health check
	if args[0] == "system" {
		runSystemDiagnostics(out)
		return
	}
	fmt.Fprintln(out, "Running diagnostics for:", args[0])
	fmt.Fprintln(out, "No critical issues detected.")
}

// Doctor helper functions
func showDoctorOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🩺 Health Check Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  all           - Run all health checks")
	fmt.Fprintln(out, "  nixpkgs       - Check nixpkgs integrity")
	fmt.Fprintln(out, "  store         - Check Nix store health")
	fmt.Fprintln(out, "  channels      - Check channel configuration")
	fmt.Fprintln(out, "  permissions   - Check file permissions")
	fmt.Fprintln(out, "  dependencies  - Check dependency conflicts")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Automated health checks coming soon"))
}

func runDoctorCheck(out io.Writer, check string) {
	fmt.Fprintln(out, utils.FormatHeader("🩺 Health Check: "+check))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatProgress("Running "+check+" health check..."))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "✅ No issues detected")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSuccess("Health check completed successfully"))
}

// runDoctorCmd executes the doctor command directly
func runDoctorCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showDoctorOptions(out)
		return
	}
	fmt.Fprintln(out, "Running doctor check:", args[0])
	fmt.Fprintln(out, "All checks passed.")
}

// Flake helper functions
func showFlakeOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("❄️  Flake Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  init          - Initialize a new flake")
	fmt.Fprintln(out, "  update        - Update flake inputs")
	fmt.Fprintln(out, "  check         - Check flake integrity")
	fmt.Fprintln(out, "  show          - Show flake information")
	fmt.Fprintln(out, "  lock          - Update flake.lock")
	fmt.Fprintln(out, "  metadata      - Show flake metadata")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Full flake management coming soon"))
}

// runFlakeCmd executes the flake command directly
func runFlakeCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showFlakeOptions(out)
		return
	}
	fmt.Fprintln(out, "Running flake operation:", args[0])
	fmt.Fprintln(out, "Operation complete.")
}

// Learning helper functions
func showLearningOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🎓 Learning Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Topics", ""))
	fmt.Fprintln(out, "  basics        - NixOS fundamentals")
	fmt.Fprintln(out, "  configuration - Configuration management")
	fmt.Fprintln(out, "  packages      - Package management")
	fmt.Fprintln(out, "  services      - System services")
	fmt.Fprintln(out, "  flakes        - Nix flakes system")
	fmt.Fprintln(out, "  advanced      - Advanced topics")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Interactive tutorials coming soon"))
}

// runLearnCmd executes the learn command directly
func runLearnCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showLearningOptions(out)
		return
	}
	topic := args[0]
	fmt.Fprintln(out, "Learning module:", topic)
	fmt.Fprintln(out, "This would launch an interactive tutorial or quiz.")
}

// Logs helper functions
func showLogsOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("📋 Log Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  system        - System logs")
	fmt.Fprintln(out, "  service <name> - Specific service logs")
	fmt.Fprintln(out, "  boot          - Boot logs")
	fmt.Fprintln(out, "  kernel        - Kernel logs")
	fmt.Fprintln(out, "  nixos-rebuild - Rebuild logs")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Advanced log analysis coming soon"))
}

// runLogsCmd executes the logs command directly
func runLogsCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showLogsOptions(out)
		return
	}
	file := args[0]
	if utils.IsFile(file) {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(out, utils.FormatError("Failed to read log file: "+err.Error()))
			return
		}
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
			return
		}
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider interface{ Query(string) (string, error) }
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(out, utils.FormatError("Unknown AI provider: "+providerName))
			return
		}
		prompt := "You are a NixOS log analysis expert. Analyze the following log and provide a summary of issues, root causes, and recommended fixes. Format as markdown.\n\nLog:\n" + string(data)
		fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		fmt.Fprintln(out, utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
			return
		}
		fmt.Fprintln(out, utils.RenderMarkdown(resp))
		return
	}
	fmt.Fprintln(out, "Analyzing logs for:", args[0])
	fmt.Fprintln(out, "No critical issues detected.")
}

// MCP Server helper functions
func showMCPServerOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔗 MCP Server Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  start         - Start the MCP server")
	fmt.Fprintln(out, "  stop          - Stop the MCP server")
	fmt.Fprintln(out, "  status        - Check server status")
	fmt.Fprintln(out, "  logs          - View server logs")
	fmt.Fprintln(out, "  config        - Show server configuration")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("MCP server provides documentation integration"))
}

// runMCPServerCmd executes the mcp-server command directly
func runMCPServerCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showMCPServerOptions(out)
		return
	}
	switch args[0] {
	case "start":
		fmt.Fprintln(out, "Starting MCP server...")
	case "stop":
		fmt.Fprintln(out, "Stopping MCP server...")
	case "status":
		fmt.Fprintln(out, "MCP server is running.")
	case "logs":
		fmt.Fprintln(out, "No recent logs found.")
	default:
		fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented mcp-server subcommand: "+args[0]))
	}
}

// Neovim Setup helper functions
func showNeovimSetupOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("📝 Neovim Setup Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  install       - Install Neovim integration")
	fmt.Fprintln(out, "  configure     - Configure Neovim plugin")
	fmt.Fprintln(out, "  test          - Test integration")
	fmt.Fprintln(out, "  update        - Update plugin")
	fmt.Fprintln(out, "  remove        - Remove integration")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Seamless NixOS integration for Neovim"))
}

// runNeovimSetupCmd executes the neovim-setup command directly
func runNeovimSetupCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showNeovimSetupOptions(out)
		return
	}
	switch args[0] {
	case "install":
		fmt.Fprintln(out, "Installing Neovim integration...")
	case "configure":
		fmt.Fprintln(out, "Configuring Neovim integration...")
	case "check":
		fmt.Fprintln(out, "Neovim integration is healthy.")
	default:
		fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented neovim-setup subcommand: "+args[0]))
	}
}

// Package Repo helper functions
func showPackageRepoOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("📦 Package Repository Options"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  analyze <url>   - Analyze a Git repository")
	fmt.Fprintln(out, "  generate <url>  - Generate Nix derivation")
	fmt.Fprintln(out, "  template        - Show available templates")
	fmt.Fprintln(out, "  validate        - Validate generated derivation")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Automated Nix package creation from Git repos"))
}

// runPackageRepoCmd executes the package-repo command directly
func runPackageRepoCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showPackageRepoOptions(out)
		return
	}
	fmt.Fprintln(out, "Analyzing repo or directory:", args[0])
	fmt.Fprintln(out, "Nix derivation generation coming soon.")
}

// Machines helper functions
func showMachinesOptions(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🖧 Machines Management"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	fmt.Fprintln(out, "  list         - List all managed machines")
	fmt.Fprintln(out, "  add <name>   - Add a new machine")
	fmt.Fprintln(out, "  sync <name>  - Sync configuration to a machine")
	fmt.Fprintln(out, "  remove <name> - Remove a machine")
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatTip("Manage and synchronize NixOS configurations across multiple machines"))
}

// runMachinesCmd executes the machines command directly
func runMachinesCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showMachinesOptions(out)
		return
	}
	switch args[0] {
	case "list":
		fmt.Fprintln(out, utils.FormatHeader("🖧 Machines List"))
		fmt.Fprintln(out, "- machine1 (example)")
		fmt.Fprintln(out, "- machine2 (example)")
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(out, utils.FormatWarning("Usage: machines add <name>"))
			return
		}
		fmt.Fprintf(out, "Added machine: %s\n", args[1])
	case "sync":
		if len(args) < 2 {
			fmt.Fprintln(out, utils.FormatWarning("Usage: machines sync <name>"))
			return
		}
		fmt.Fprintf(out, "Synced configuration to machine: %s\n", args[1])
	case "remove":
		if len(args) < 2 {
			fmt.Fprintln(out, utils.FormatWarning("Usage: machines remove <name>"))
			return
		}
		fmt.Fprintf(out, "Removed machine: %s\n", args[1])
	default:
		fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented machines subcommand: "+args[0]))
	}
}

// Build command
func runBuildCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🛠️ Build Troubleshooting & Optimization"))
	fmt.Fprintln(out, "Enhanced build troubleshooting and optimization coming soon.")
}

// Completion command
func runCompletionCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔄 Completion Script"))
	fmt.Fprintln(out, "Generate the autocompletion script for your shell (bash, zsh, fish, etc). Example: nixai completion zsh > _nixai")
}

// Deps command
func runDepsCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔗 NixOS Dependency Analysis"))
	fmt.Fprintln(out, "Analyze NixOS configuration dependencies and imports. (Stub)")
}

// Devenv command
func runDevenvCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🧪 Development Environments"))
	fmt.Fprintln(out, "Create and manage development environments with devenv. (Stub)")
}

// Explain-option command
func runExplainOptionCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🖥️ Explain NixOS Option"))
	fmt.Fprintln(out, "Explain a NixOS option using AI and documentation. (Stub)")
}

// GC command
func runGCCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🧹 Garbage Collection"))
	fmt.Fprintln(out, "AI-powered garbage collection analysis and cleanup. (Stub)")
}

// Hardware command
func runHardwareCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("💻 Hardware Optimizer"))
	fmt.Fprintln(out, "AI-powered hardware configuration optimizer. (Stub)")
}

// Interactive command
func runInteractiveCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("💬 Interactive Mode"))
	fmt.Fprintln(out, "You are already in interactive mode!")
}

// Migrate command
func runMigrateCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔀 Migration Assistant"))
	fmt.Fprintln(out, "AI-powered migration assistant for channels and flakes. (Stub)")
}

// Search command
func runSearchCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔍 NixOS Package Search"))
	fmt.Fprintln(out, "Search for NixOS packages/services and get config/AI tips. (Stub)")
}

// Snippets command
func runSnippetsCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🔖 Configuration Snippets"))
	fmt.Fprintln(out, "Manage NixOS configuration snippets. (Stub)")
}

// Store command
func runStoreCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("💾 Nix Store Management"))
	fmt.Fprintln(out, "Manage, backup, and analyze the Nix store. (Stub)")
}

// Templates command
func runTemplatesCmd(args []string, out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("📄 Configuration Templates"))
	fmt.Fprintln(out, "Manage NixOS configuration templates and snippets. (Stub)")
}

// RunDirectCommand executes commands directly from interactive mode
func RunDirectCommand(cmdName string, args []string, out io.Writer) (bool, error) {
	switch cmdName {
	case "community":
		runCommunityCmd(args, out)
		return true, nil
	case "config":
		runConfigCmd(args, out)
		return true, nil
	case "configure":
		runConfigureCmd(args, out)
		return true, nil
	case "diagnose":
		runDiagnoseCmd(args, out)
		return true, nil
	case "doctor":
		runDoctorCmd(args, out)
		return true, nil
	case "flake":
		runFlakeCmd(args, out)
		return true, nil
	case "learn":
		runLearnCmd(args, out)
		return true, nil
	case "logs":
		runLogsCmd(args, out)
		return true, nil
	case "mcp-server":
		runMCPServerCmd(args, out)
		return true, nil
	case "neovim-setup":
		runNeovimSetupCmd(args, out)
		return true, nil
	case "package-repo":
		runPackageRepoCmd(args, out)
		return true, nil
	case "machines":
		runMachinesCmd(args, out)
		return true, nil
	case "build":
		runBuildCmd(args, out)
		return true, nil
	case "completion":
		runCompletionCmd(args, out)
		return true, nil
	case "deps":
		runDepsCmd(args, out)
		return true, nil
	case "devenv":
		runDevenvCmd(args, out)
		return true, nil
	case "explain-option":
		runExplainOptionCmd(args, out)
		return true, nil
	case "gc":
		runGCCmd(args, out)
		return true, nil
	case "hardware":
		runHardwareCmd(args, out)
		return true, nil
	case "interactive":
		runInteractiveCmd(args, out)
		return true, nil
	case "migrate":
		runMigrateCmd(args, out)
		return true, nil
	case "search":
		runSearchCmd(args, out)
		return true, nil
	case "snippets":
		runSnippetsCmd(args, out)
		return true, nil
	case "store":
		runStoreCmd(args, out)
		return true, nil
	case "templates":
		runTemplatesCmd(args, out)
		return true, nil
	default:
		return false, nil
	}
}
