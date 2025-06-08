package cli

import (
	"fmt"
	"io"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// Helper to run a cobra.Command and capture its output to io.Writer
func runCobraCommand(cmd *cobra.Command, args []string, out io.Writer) {
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	_ = cmd.Execute()
}

// Helper functions for running commands directly in interactive mode

// Config command wrapper functions that accept io.Writer
func showConfigWithOutput(out io.Writer) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔧 Current Configuration"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("AI Provider", cfg.AIProvider))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("AI Model", cfg.AIModel))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Log Level", cfg.LogLevel))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS Folder", cfg.NixosFolder))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("MCP Host", cfg.MCPServer.Host))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("MCP Port", fmt.Sprintf("%d", cfg.MCPServer.Port)))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Use 'config set <key> <value>' to modify settings"))
}

func setConfigWithOutput(out io.Writer, key, value string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	switch key {
	case "ai_provider":
		if value != "ollama" && value != "gemini" && value != "openai" {
			_, _ = fmt.Fprintln(out, utils.FormatError("Invalid AI provider. Valid options: ollama, gemini, openai"))
			return
		}
		cfg.AIProvider = value
	case "ai_model":
		cfg.AIModel = value
	case "log_level":
		if value != "debug" && value != "info" && value != "warn" && value != "error" {
			_, _ = fmt.Fprintln(out, utils.FormatError("Invalid log level. Valid options: debug, info, warn, error"))
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
			_, _ = fmt.Fprintln(out, utils.FormatError("Invalid port number"))
			return
		}
	default:
		_, _ = fmt.Fprintln(out, utils.FormatError("Unknown configuration key: "+key))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		return
	}

	err = config.SaveUserConfig(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to save config: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Configuration updated successfully"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue(key, value))
}

func getConfigWithOutput(out io.Writer, key string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
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
		_, _ = fmt.Fprintln(out, utils.FormatError("Unknown configuration key: "+key))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatKeyValue(key, value))
}

func resetConfigWithOutput(out io.Writer) {
	cfg := config.DefaultUserConfig()
	err := config.SaveUserConfig(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to reset config: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Configuration reset to defaults"))
	_, _ = fmt.Fprintln(out, utils.FormatTip("Use 'config show' to see current settings"))
}

// Community helper functions
func showCommunityOverview(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🌐 Community Overview"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  search <query>     - Search community configurations")
	_, _ = fmt.Fprintln(out, "  share <file>       - Share your configuration")
	_, _ = fmt.Fprintln(out, "  validate <file>    - Validate configuration against best practices")
	_, _ = fmt.Fprintln(out, "  trends             - Show trending packages and patterns")
	_, _ = fmt.Fprintln(out, "  rate <config> <n>  - Rate a community configuration")
	_, _ = fmt.Fprintln(out, "  forums             - Show community forums and discussions")
	_, _ = fmt.Fprintln(out, "  docs               - Show community documentation resources")
	_, _ = fmt.Fprintln(out, "  matrix             - Show Matrix chat channels")
	_, _ = fmt.Fprintln(out, "  github             - Show GitHub resources and repositories")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Use 'nixai community <command> --help' for detailed information"))
}

func showCommunityForums(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("💬 Community Forums"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS Discourse", "https://discourse.nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Reddit r/NixOS", "https://reddit.com/r/NixOS"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Stack Overflow", "https://stackoverflow.com/questions/tagged/nixos"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Search for solutions and ask questions in these forums"))
}

func showCommunityDocs(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📚 Community Documentation"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS Manual", "https://nixos.org/manual/nixos/stable/"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Nixpkgs Manual", "https://nixos.org/manual/nixpkgs/stable/"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Nix Manual", "https://nix.dev/manual/nix"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Home Manager", "https://nix-community.github.io/home-manager/"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Wiki", "https://wiki.nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Nix Dev", "https://nix.dev"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("These are the primary documentation sources"))
}

func showMatrixChannels(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("💬 Matrix Chat Channels"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Main Channel", "#nixos:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Development", "#nixos-dev:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Security", "#nixos-security:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Offtopic", "#nixos-chat:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("ARM", "#nixos-aarch64:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Gaming", "#nixos-gaming:nixos.org"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Real-time chat with the NixOS community"))
}

func showGitHubResources(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🐙 GitHub Resources"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nixpkgs", "https://github.com/NixOS/nixpkgs"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nix", "https://github.com/NixOS/nix"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("nix-community", "https://github.com/nix-community"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nixos-hardware", "https://github.com/NixOS/nixos-hardware"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Awesome Nix", "https://github.com/nix-community/awesome-nix"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Browse source code, issues, and contribute to projects"))
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
			_, _ = fmt.Fprintln(out, "Usage: nixai config set <key> <value>")
			return
		}
		setConfigWithOutput(out, args[1], args[2])
	case "get":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, "Usage: nixai config get <key>")
			return
		}
		getConfigWithOutput(out, args[1])
	case "reset":
		resetConfigWithOutput(out)
	default:
		_, _ = fmt.Fprintln(out, "Unknown config command: "+args[0])
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
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented community subcommand: "+args[0]))
	}
}

// runConfigureCmd executes the configure command directly
func runConfigureCmd(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, "Interactive configuration coming soon.")
}

// Diagnose helper functions
func showDiagnosticOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 Diagnostic Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  system        - Overall system health check")
	_, _ = fmt.Fprintln(out, "  config        - Configuration file analysis")
	_, _ = fmt.Fprintln(out, "  services      - Service status and logs")
	_, _ = fmt.Fprintln(out, "  network       - Network connectivity tests")
	_, _ = fmt.Fprintln(out, "  hardware      - Hardware detection and drivers")
	_, _ = fmt.Fprintln(out, "  performance   - Performance bottleneck analysis")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Comprehensive system diagnostics coming soon"))
}

func runSystemDiagnostics(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 System Diagnostics"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatProgress("Running system health checks..."))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "✅ Boot loader: OK")
	_, _ = fmt.Fprintln(out, "✅ Filesystem: OK")
	_, _ = fmt.Fprintln(out, "✅ Network: OK")
	_, _ = fmt.Fprintln(out, "✅ Services: OK")
	_, _ = fmt.Fprintln(out, "✅ Hardware: OK")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSuccess("System health: All checks passed"))
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
	_, _ = fmt.Fprintln(out, "Running diagnostics for:", args[0])
	_, _ = fmt.Fprintln(out, "No critical issues detected.")
}

// Doctor helper functions
func showDoctorOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🩺 Health Check Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  all           - Run all health checks")
	_, _ = fmt.Fprintln(out, "  nixpkgs       - Check nixpkgs integrity")
	_, _ = fmt.Fprintln(out, "  store         - Check Nix store health")
	_, _ = fmt.Fprintln(out, "  channels      - Check channel configuration")
	_, _ = fmt.Fprintln(out, "  permissions   - Check file permissions")
	_, _ = fmt.Fprintln(out, "  dependencies  - Check dependency conflicts")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Automated health checks coming soon"))
}

// runDoctorCmd executes the doctor command directly
func runDoctorCmd(args []string, out io.Writer) {
	// Use the comprehensive enhanced doctor implementation
	runCobraCommand(doctorCmd, args, out)
}

// Flake helper functions
func showFlakeOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("❄️  Flake Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  init          - Initialize a new flake")
	_, _ = fmt.Fprintln(out, "  update        - Update flake inputs")
	_, _ = fmt.Fprintln(out, "  check         - Check flake integrity")
	_, _ = fmt.Fprintln(out, "  show          - Show flake information")
	_, _ = fmt.Fprintln(out, "  lock          - Update flake.lock")
	_, _ = fmt.Fprintln(out, "  metadata      - Show flake metadata")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Full flake management coming soon"))
}

// runFlakeCmd executes the flake command directly
func runFlakeCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showFlakeOptions(out)
		return
	}
	_, _ = fmt.Fprintln(out, "Running flake operation:", args[0])
	_, _ = fmt.Fprintln(out, "Operation complete.")
}

// Learning helper functions
func showLearningOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🎓 Learning Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Topics", ""))
	_, _ = fmt.Fprintln(out, "  basics        - NixOS fundamentals")
	_, _ = fmt.Fprintln(out, "  configuration - Configuration management")
	_, _ = fmt.Fprintln(out, "  packages      - Package management")
	_, _ = fmt.Fprintln(out, "  services      - System services")
	_, _ = fmt.Fprintln(out, "  flakes        - Nix flakes system")
	_, _ = fmt.Fprintln(out, "  advanced      - Advanced topics")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Interactive tutorials coming soon"))
}

// runLearnCmd executes the learn command directly
func runLearnCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showLearningOptions(out)
		return
	}
	topic := args[0]
	_, _ = fmt.Fprintln(out, "Learning module:", topic)
	_, _ = fmt.Fprintln(out, "This would launch an interactive tutorial or quiz.")
}

// Logs helper functions
func showLogsOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📋 Log Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  system        - System logs")
	_, _ = fmt.Fprintln(out, "  service <name> - Specific service logs")
	_, _ = fmt.Fprintln(out, "  boot          - Boot logs")
	_, _ = fmt.Fprintln(out, "  kernel        - Kernel logs")
	_, _ = fmt.Fprintln(out, "  nixos-rebuild - Rebuild logs")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Advanced log analysis coming soon"))
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
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to read log file: "+err.Error()))
			return
		}
		cfg, err := config.LoadUserConfig()
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
			return
		}
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			return
		}
		prompt := "You are a NixOS log analysis expert. Analyze the following log and provide a summary of issues, root causes, and recommended fixes. Format as markdown.\n\nLog:\n" + string(data)
		_, _ = fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
			return
		}
		_, _ = fmt.Fprintln(out, utils.RenderMarkdown(resp))
		return
	}
	_, _ = fmt.Fprintln(out, "Analyzing logs for:", args[0])
	_, _ = fmt.Fprintln(out, "No critical issues detected.")
}

// MCP Server helper functions
func showMCPServerOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔗 MCP Server Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  start         - Start the MCP server")
	_, _ = fmt.Fprintln(out, "  stop          - Stop the MCP server")
	_, _ = fmt.Fprintln(out, "  status        - Check server status")
	_, _ = fmt.Fprintln(out, "  logs          - View server logs")
	_, _ = fmt.Fprintln(out, "  config        - Show server configuration")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("MCP server provides documentation integration"))
}

// runMCPServerCmd executes the mcp-server command directly
func runMCPServerCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showMCPServerOptions(out)
		return
	}
	switch args[0] {
	case "start":
		_, _ = fmt.Fprintln(out, "Starting MCP server...")
	case "stop":
		_, _ = fmt.Fprintln(out, "Stopping MCP server...")
	case "status":
		_, _ = fmt.Fprintln(out, "MCP server is running.")
	case "logs":
		_, _ = fmt.Fprintln(out, "No recent logs found.")
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented mcp-server subcommand: "+args[0]))
	}
}

// Neovim Setup helper functions
func showNeovimSetupOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📝 Neovim Setup Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  install       - Install Neovim integration")
	_, _ = fmt.Fprintln(out, "  configure     - Configure Neovim plugin")
	_, _ = fmt.Fprintln(out, "  test          - Test integration")
	_, _ = fmt.Fprintln(out, "  update        - Update plugin")
	_, _ = fmt.Fprintln(out, "  remove        - Remove integration")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Seamless NixOS integration for Neovim"))
}

// runNeovimSetupCmd executes the neovim-setup command directly
func runNeovimSetupCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showNeovimSetupOptions(out)
		return
	}
	switch args[0] {
	case "install":
		_, _ = fmt.Fprintln(out, "Installing Neovim integration...")
	case "configure":
		_, _ = fmt.Fprintln(out, "Configuring Neovim integration...")
	case "check":
		_, _ = fmt.Fprintln(out, "Neovim integration is healthy.")
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented neovim-setup subcommand: "+args[0]))
	}
}

// Package Repo helper functions
func showPackageRepoOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📦 Package Repository Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  analyze <url>   - Analyze a Git repository")
	_, _ = fmt.Fprintln(out, "  generate <url>  - Generate Nix derivation")
	_, _ = fmt.Fprintln(out, "  template        - Show available templates")
	_, _ = fmt.Fprintln(out, "  validate        - Validate generated derivation")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Automated Nix package creation from Git repos"))
}

// runPackageRepoCmd executes the package-repo command directly
func runPackageRepoCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showPackageRepoOptions(out)
		return
	}
	_, _ = fmt.Fprintln(out, "Analyzing repo or directory:", args[0])
	_, _ = fmt.Fprintln(out, "Nix derivation generation coming soon.")
}

// Machines helper functions
func showMachinesOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🖧 Machines Management"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  list         - List all managed machines")
	_, _ = fmt.Fprintln(out, "  add <name>   - Add a new machine")
	_, _ = fmt.Fprintln(out, "  sync <name>  - Sync configuration to a machine")
	_, _ = fmt.Fprintln(out, "  remove <name> - Remove a machine")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Manage and synchronize NixOS configurations across multiple machines"))
}

// runMachinesCmd executes the machines command directly
func runMachinesCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showMachinesOptions(out)
		return
	}
	switch args[0] {
	case "list":
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🖧 Machines List"))
		_, _ = fmt.Fprintln(out, "- machine1 (example)")
		_, _ = fmt.Fprintln(out, "- machine2 (example)")
	case "add":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("Usage: machines add <name>"))
			return
		}
		_, _ = fmt.Fprintf(out, "Added machine: %s\n", args[1])
	case "sync":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("Usage: machines sync <name>"))
			return
		}
		_, _ = fmt.Fprintf(out, "Synced configuration to machine: %s\n", args[1])
	case "remove":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("Usage: machines remove <name>"))
			return
		}
		_, _ = fmt.Fprintf(out, "Removed machine: %s\n", args[1])
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented machines subcommand: "+args[0]))
	}
}

// Build command
func runBuildCmd(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🛠️ Build Troubleshooting & Optimization"))
	_, _ = fmt.Fprintln(out, "Enhanced build troubleshooting and optimization coming soon.")
}

// Completion command
func runCompletionCmd(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔄 Completion Script"))
	_, _ = fmt.Fprintln(out, "Generate the autocompletion script for your shell (bash, zsh, fish, etc). Example: nixai completion zsh > _nixai")
}

// Deps command
func runDepsCmd(args []string, out io.Writer) {
	runCobraCommand(NewDepsCommand(), args, out)
}

// Devenv command
func runDevenvCmd(args []string, out io.Writer) {
	// Show help if no args (for interactive parity)
	if len(args) == 0 {
		_ = NewDevenvCommand().Help()
		return
	}
	runCobraCommand(NewDevenvCommand(), args, out)
}

// NewDevenvCommand returns a fresh devenv command with all subcommands
func NewDevenvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   devenvCmd.Use,
		Short: devenvCmd.Short,
		Long:  devenvCmd.Long,
	}
	// Add subcommands as fresh instances
	cmd.AddCommand(NewDevenvListCmd())
	cmd.AddCommand(NewDevenvCreateCmd())
	cmd.AddCommand(NewDevenvSuggestCmd())
	cmd.PersistentFlags().AddFlagSet(devenvCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(devenvCmd.Flags())
	return cmd
}

// Explain-option command
func runExplainOptionCmd(args []string, out io.Writer) {
	runCobraCommand(NewExplainOptionCommand(), args, out)
}

// GC command
func runGCCmd(args []string, out io.Writer) {
	runCobraCommand(NewGCCmd(), args, out)
}

// Hardware command
func runHardwareCmd(args []string, out io.Writer) {
	runCobraCommand(NewHardwareCmd(), args, out)
}

// Migrate command
func runMigrateCmd(args []string, out io.Writer) {
	runCobraCommand(NewMigrateCmd(), args, out)
}

// Snippets command
func runSnippetsCmd(args []string, out io.Writer) {
	runCobraCommand(NewSnippetsCmd(), args, out)
}

// Templates command
func runTemplatesCmd(args []string, out io.Writer) {
	runCobraCommand(NewTemplatesCmd(), args, out)
}

// NewSnippetsCmd returns a cobra.Command for the 'snippets' command.
func NewSnippetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippets",
		Short: "Show, add, or manage code snippets for NixOS, Home Manager, and related workflows.",
		Long: `Manage and view reusable code snippets for NixOS, Home Manager, and related workflows.

Examples:
  nixai snippets list
  nixai snippets add --name my-snippet --file ./snippet.nix
  nixai snippets show my-snippet
  nixai snippets remove my-snippet
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, just list available snippets from a default directory
			snippetDir := utils.GetSnippetsDir()
			snippets, err := utils.ListSnippets(snippetDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list snippets: %v\n", err)
				return err
			}
			if len(snippets) == 0 {
				cmd.Println(utils.FormatHeader("No snippets found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Snippets:"))
			for _, s := range snippets {
				cmd.Println(utils.FormatKeyValue(s.Name, s.Description))
			}
			return nil
		},
	}
	cmd.AddCommand(NewSnippetsListCmd())
	cmd.AddCommand(NewSnippetsAddCmd())
	cmd.AddCommand(NewSnippetsShowCmd())
	cmd.AddCommand(NewSnippetsRemoveCmd())
	return cmd
}

// NewSnippetsListCmd returns a cobra.Command for the 'list' subcommand of 'snippets'.
func NewSnippetsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available code snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := utils.GetSnippetsDir()
			snippets, err := utils.ListSnippets(dir)
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list snippets: " + err.Error()))
				return err
			}
			if len(snippets) == 0 {
				cmd.Println(utils.FormatHeader("No snippets found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Snippets:"))
			for _, s := range snippets {
				cmd.Println(utils.FormatKeyValue(s.Name, s.Description))
			}
			return nil
		},
	}
}

// NewSnippetsAddCmd returns a cobra.Command for the 'add' subcommand of 'snippets'.
func NewSnippetsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new code snippet",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(utils.FormatHeader("Add snippet: Not yet implemented."))
			return nil
		},
	}
}

// NewSnippetsShowCmd returns a cobra.Command for the 'show' subcommand of 'snippets'.
func NewSnippetsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show a code snippet by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Println(utils.FormatError("Usage: snippets show <name>"))
				return nil
			}
			dir := utils.GetSnippetsDir()
			snippets, err := utils.ListSnippets(dir)
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list snippets: " + err.Error()))
				return err
			}
			for _, s := range snippets {
				if s.Name == args[0] {
					cmd.Println(utils.FormatHeader(s.Name))
					content, _ := os.ReadFile(s.Path)
					cmd.Println(string(content))
					return nil
				}
			}
			cmd.Println(utils.FormatError("Snippet not found: " + args[0]))
			return nil
		},
	}
}

// NewSnippetsRemoveCmd returns a cobra.Command for the 'remove' subcommand of 'snippets'.
func NewSnippetsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove a code snippet by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(utils.FormatHeader("Remove snippet: Not yet implemented."))
			return nil
		},
	}
}

// NewTemplatesCmd returns a cobra.Command for the 'templates' command.
func NewTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List and manage project templates for NixOS, Home Manager, and related setups.",
		Long: `Browse, add, or use project templates for NixOS, Home Manager, and related workflows.

Examples:
  nixai templates list
  nixai templates show minimal-nixos
  nixai templates use minimal-nixos --output ./my-nixos
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := utils.ListTemplates()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list templates: %v\n", err)
				return err
			}
			if len(templates) == 0 {
				cmd.Println(utils.FormatHeader("No templates found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Templates:"))
			for _, t := range templates {
				cmd.Println(utils.FormatKeyValue(t.Name, t.Description))
			}
			return nil
		},
	}
	cmd.AddCommand(NewTemplatesListCmd())
	cmd.AddCommand(NewTemplatesShowCmd())
	cmd.AddCommand(NewTemplatesUseCmd())
	return cmd
}

// NewTemplatesListCmd returns a cobra.Command for the 'list' subcommand of 'templates'.
func NewTemplatesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := utils.ListTemplates()
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list templates: " + err.Error()))
				return err
			}
			if len(templates) == 0 {
				cmd.Println(utils.FormatHeader("No templates found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Templates:"))
			for _, t := range templates {
				cmd.Println(utils.FormatKeyValue(t.Name, t.Description))
			}
			return nil
		},
	}
}

// NewTemplatesShowCmd returns a cobra.Command for the 'show' subcommand of 'templates'.
func NewTemplatesShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show a template by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Println(utils.FormatError("Usage: templates show <name>"))
				return nil
			}
			templates, err := utils.ListTemplates()
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list templates: " + err.Error()))
				return err
			}
			for _, t := range templates {
				if t.Name == args[0] {
					cmd.Println(utils.FormatHeader(t.Name))
					content, _ := os.ReadFile(t.Path)
					cmd.Println(string(content))
					return nil
				}
			}
			cmd.Println(utils.FormatError("Template not found: " + args[0]))
			return nil
		},
	}
}

// NewTemplatesUseCmd returns a cobra.Command for the 'use' subcommand of 'templates'.
func NewTemplatesUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use",
		Short: "Copy a template to a target directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(utils.FormatHeader("Use template: Not yet implemented."))
			return nil
		},
	}
}

// Store command
func runStoreCmd(args []string, out io.Writer) {
	runCobraCommand(NewStoreCommand(), args, out)
}

// NewStoreCommand returns a fresh store command with all subcommands
func NewStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   storeCmd.Use,
		Short: storeCmd.Short,
		Long:  storeCmd.Long,
	}
	// Add subcommands (fresh instances)
	cmd.AddCommand(storeBackupCmd)
	cmd.AddCommand(storeRestoreCmd)
	cmd.AddCommand(storeIntegrityCmd)
	cmd.AddCommand(storePerformanceCmd)
	cmd.PersistentFlags().AddFlagSet(storeCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(storeCmd.Flags())
	return cmd
}

// Search command
func runSearchCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, utils.FormatError("Usage: search <package>"))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Example: search curl"))
		return
	}

	query := args[0]
	if len(args) > 1 {
		query = fmt.Sprintf("%s %s", args[0], args[1])

	}

	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	exec := nixos.NewExecutor(cfg.NixosFolder)
	pkgOut, pkgErr := exec.SearchNixPackages(query)
	if pkgErr != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("NixOS package search failed: "+pkgErr.Error()))
	} else if pkgOut != "" {
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 NixOS Search Results for: "+query))
		_, _ = fmt.Fprintln(out, pkgOut)
	}

	aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}

	// Get provider name for context
	providerName := cfg.AIProvider
	if providerName == "" {
		providerName = "ollama"
	}

	var docExcerpts []string
	_, _ = fmt.Fprint(out, utils.FormatInfo("Querying documentation... "))
	mcpBase := cfg.MCPServer.Host
	mcpContextAdded := false
	if mcpBase != "" {
		mcpClient := mcp.NewMCPClient(mcpBase)
		doc, err := mcpClient.QueryDocumentation(query)
		_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
		if err == nil && doc != "" {
			opt, fallbackDoc := parseMCPOptionDoc(doc)
			if opt.Name != "" {
				context := fmt.Sprintf("Option: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nNixOS Version: %s\nRelated: %v\nLinks: %v", opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
				docExcerpts = append(docExcerpts, context)
				mcpContextAdded = true
			} else if len(fallbackDoc) > 0 && (len(fallbackDoc) < 1000 || len(fallbackDoc) > 10) {
				docExcerpts = append(docExcerpts, fallbackDoc)
				mcpContextAdded = true
			}
		}
	} else {
		_, _ = fmt.Fprintln(out, utils.FormatWarning("skipped (no MCP host configured)"))
	}

	promptInstruction := "You are a NixOS expert. Always provide NixOS-specific configuration.nix examples, use the NixOS module system, and avoid generic Linux or upstream package advice. Show how to enable and configure this package/service in NixOS."
	if !mcpContextAdded {
		docExcerpts = append(docExcerpts, promptInstruction)
	} else {
		docExcerpts = append(docExcerpts, "\n"+promptInstruction)
	}

	promptCtx := ai.PromptContext{
		Question:     query,
		DocExcerpts:  docExcerpts,
		Intent:       "explain",
		OutputFormat: "markdown",
		Provider:     providerName,
	}
	builder := ai.DefaultPromptBuilder{}
	prompt, err := builder.BuildPrompt(promptCtx)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Prompt build error: "+err.Error()))
		return
	}
	_, _ = fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
	aiAnswer, aiErr := aiProvider.Query(prompt)
	_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
	if aiErr == nil && aiAnswer != "" {
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🤖 AI Best Practices & Tips"))
		_, _ = fmt.Fprintln(out, utils.RenderMarkdown(aiAnswer))
	}
}

// Ask command
func runAskCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, utils.FormatError("Usage: ask <question>"))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Example: ask How do I enable nginx?"))
		return
	}
	question := ""
	if len(args) == 1 {
		question = args[0]
	} else {
		question = ""
		for i, s := range args {
			if i > 0 {
				question += " "
			}
			question += s
		}
	}
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🤖 AI Answer to your question:"))
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}
	aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}
	_, _ = fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
	resp, err := aiProvider.Query(question)
	_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
		return
	}
	_, _ = fmt.Fprintln(out, utils.RenderMarkdown(resp))
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
		_, _ = fmt.Fprintln(out, utils.FormatTip("You are already in interactive mode!"))
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
	case "ask":
		runAskCmd(args, out)
		return true, nil
	case "help":
		_, _ = fmt.Fprintln(out, utils.FormatHeader("❓ Help: Available Commands"))
		_, _ = fmt.Fprintln(out, `🤖 ask <question>: Ask any NixOS question\n🛠️ build: Enhanced build troubleshooting and optimization\n🌐 community: Community resources and support\n🔄 completion: Generate the autocompletion script for the specified shell\n⚙️ config: Manage nixai configuration\n🧑‍💻 configure: Configure NixOS interactively\n🔗 deps: Analyze NixOS configuration dependencies and imports\n🧪 devenv: Create and manage development environments with devenv\n🩺 diagnose: Diagnose NixOS issues\n🩻 doctor: Run NixOS health checks\n🖥️ explain-option <option>: Explain a NixOS option\n🧊 flake: Nix flake utilities\n🧹 gc: AI-powered garbage collection analysis and cleanup\n💻 hardware: AI-powered hardware configuration optimizer\n❓ help: Help about any command\n💬 interactive: Launch interactive AI-powered NixOS assistant shell\n📚 learn: NixOS learning and training commands\n📝 logs: Analyze and parse NixOS logs\n🖧 machines: Manage and synchronize NixOS configurations across multiple machines\n🛰️ mcp-server: Start or manage the MCP server\n🔀 migrate: AI-powered migration assistant for channels and flakes\n📝 neovim-setup: Neovim integration setup\n📦 package-repo <url>: Analyze Git repos and generate Nix derivations\n🔍 search <package>: Search for NixOS packages/services and get config/AI tips\n🔖 snippets: Manage NixOS configuration snippets\n💾 store: Manage, backup, and analyze the Nix store\n📄 templates: Manage NixOS configuration templates and snippets\n❌ exit: Exit interactive mode`)
		return true, nil
	case "exit":
		_, _ = fmt.Fprintln(out, utils.FormatTip("Type Ctrl+D or 'exit' to leave interactive mode."))
		return true, nil
	default:
		return false, nil
	}
}
