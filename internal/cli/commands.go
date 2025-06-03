package cli

import (
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/utils"
	"nix-ai-help/pkg/version"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nixai [question] [flags]",
	Short: "NixOS AI Assistant",
	Long: `nixai is a command-line tool that assists users in diagnosing and solving NixOS configuration issues using AI models and documentation queries.

You can also ask questions directly, e.g.:
  nixai "how can I configure curl?"

Usage:
  nixai [question] [flags]
  nixai [command]`,
	SilenceUsage: true,
}

var askQuestion string
var nixosPath string
var showVersion bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&askQuestion, "ask", "a", "", "Ask a question about NixOS configuration")
	rootCmd.PersistentFlags().StringVarP(&nixosPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version for nixai")
}

// Configuration management functions
func showConfig() {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	fmt.Println(utils.FormatHeader("🔧 Current nixai Configuration"))
	fmt.Println()
	fmt.Println(utils.FormatKeyValue("AI Provider", cfg.AIProvider))
	fmt.Println(utils.FormatKeyValue("AI Model", cfg.AIModel))
	fmt.Println(utils.FormatKeyValue("Log Level", cfg.LogLevel))
	fmt.Println(utils.FormatKeyValue("NixOS Folder", cfg.NixosFolder))
	fmt.Println(utils.FormatKeyValue("MCP Server Host", cfg.MCPServer.Host))
	fmt.Println(utils.FormatKeyValue("MCP Server Port", fmt.Sprintf("%d", cfg.MCPServer.Port)))
	if len(cfg.MCPServer.DocumentationSources) > 0 {
		fmt.Println(utils.FormatKeyValue("Documentation Sources", strings.Join(cfg.MCPServer.DocumentationSources, ", ")))
	}
}

func setConfig(key, value string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	switch key {
	case "ai_provider":
		if value != "ollama" && value != "gemini" && value != "openai" {
			fmt.Println(utils.FormatError("Invalid AI provider. Valid options: ollama, gemini, openai"))
			os.Exit(1)
		}
		cfg.AIProvider = value
	case "ai_model":
		cfg.AIModel = value
	case "log_level":
		if value != "debug" && value != "info" && value != "warn" && value != "error" {
			fmt.Println(utils.FormatError("Invalid log level. Valid options: debug, info, warn, error"))
			os.Exit(1)
		}
		cfg.LogLevel = value
	case "nixos_folder":
		cfg.NixosFolder = value
	case "mcp_host":
		cfg.MCPServer.Host = value
	case "mcp_port":
		port, err := fmt.Sscanf(value, "%d", &cfg.MCPServer.Port)
		if err != nil || port != 1 {
			fmt.Println(utils.FormatError("Invalid port number"))
			os.Exit(1)
		}
	default:
		fmt.Println(utils.FormatError("Unknown configuration key: " + key))
		fmt.Println(utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		os.Exit(1)
	}

	err = config.SaveUserConfig(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to save config: "+err.Error()))
		os.Exit(1)
	}

	fmt.Println(utils.FormatSuccess("✅ Configuration updated successfully"))
	fmt.Println(utils.FormatKeyValue(key, value))
}

func getConfig(key string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
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
		fmt.Println(utils.FormatError("Unknown configuration key: " + key))
		fmt.Println(utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		os.Exit(1)
	}

	fmt.Println(utils.FormatKeyValue(key, value))
}

func resetConfig() {
	fmt.Println(utils.FormatWarning("⚠️  This will reset all configuration to defaults. Continue? (y/N)"))
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println(utils.FormatInfo("Operation cancelled"))
		return
	}

	// Create default config
	defaultCfg := &config.UserConfig{
		AIProvider:  "ollama",
		AIModel:     "llama3",
		LogLevel:    "info",
		NixosFolder: "/etc/nixos",
		MCPServer: config.MCPServerConfig{
			Host: "localhost",
			Port: 8081,
			DocumentationSources: []string{
				"https://wiki.nixos.org/wiki/NixOS_Wiki",
				"https://nix.dev/manual/nix",
				"https://nixos.org/manual/nixpkgs/stable/",
				"https://nix.dev/manual/nix/2.28/language/",
				"https://nix-community.github.io/home-manager/",
			},
		},
	}

	err := config.SaveUserConfig(defaultCfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to reset config: "+err.Error()))
		os.Exit(1)
	}

	fmt.Println(utils.FormatSuccess("✅ Configuration reset to defaults successfully"))
}

func buildExplainHomeOptionPrompt(option, documentation string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users understand Home Manager configuration options. Please explain the following Home Manager option in a clear, practical manner.\n\n**Option:** %s\n\n**Official Documentation:**\n%s\n\n**Please provide:**\n\n1. **Purpose & Overview**: What this option does and why you'd use it\n2. **Type & Default**: The data type and default value (if any)\n3. **Usage Examples**: Show 2-3 practical configuration examples\n4. **Best Practices**: How to use this option effectively\n5. **Related Options**: Other options that are commonly used with this one\n6. **Common Issues**: Potential problems and their solutions\n\nFormat your response using Markdown with section headings and code blocks for examples.`, option, documentation)
}

func buildExplainOptionPrompt(option, documentation string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users understand configuration options. Please explain the following NixOS option in a clear, practical manner.\n\n**Option:** %s\n\n**Official Documentation:**\n%s\n\n**Please provide:**\n\n1. **Purpose & Overview**: What this option does and why you'd use it\n2. **Type & Default**: The data type and default value (if any)\n3. **Usage Examples**: Show 2-3 practical configuration examples\n4. **Best Practices**: How to use this option effectively\n5. **Related Options**: Other options that are commonly used with this one\n6. **Common Issues**: Potential problems and their solutions\n\nFormat your response using Markdown with section headings and code blocks for examples.`, option, documentation)
}

// searchCmd implements the enhanced search logic
var searchCmd = &cobra.Command{
	Use:   "search [package]",
	Short: "Search for NixOS packages/services and get config/AI tips",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}
		exec := nixos.NewExecutor(cfg.NixosFolder)
		fmt.Println(utils.FormatHeader("🔍 NixOS Search Results for: " + query))
		fmt.Println()
		// Package search
		pkgOut, pkgErr := exec.SearchNixPackages(query)
		if pkgErr == nil && pkgOut != "" {
			fmt.Println(pkgOut)
		}
		// Optionally: Service search, etc.
		// AI-powered answer
		aiProvider := ai.NewOllamaProvider("llama3") // Default, or use config
		aiPrompt := "Provide best practices, advanced usage, and pitfalls for NixOS package or service: " + query
		aiAnswer, aiErr := aiProvider.Query(aiPrompt)
		if aiErr == nil && aiAnswer != "" {
			aiBox := utils.FormatBox("🤖 AI Best Practices & Tips", aiAnswer)
			renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(120))
			rendered, err := renderer.Render(aiBox)
			if err != nil {
				fmt.Println(aiBox)
			} else {
				fmt.Print(rendered)
			}
		}
	},
}

// explainHomeOptionCmd implements the explain-home-option command
var explainHomeOptionCmd = &cobra.Command{
	Use:   "explain-home-option <option>",
	Short: "Explain a Home Manager option using AI and documentation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		option := args[0]
		fmt.Println(utils.FormatHeader("🏠 Home Manager Option: " + option))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}
		mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
		mcpClient := mcp.NewMCPClient(mcpURL)
		doc, docErr := mcpClient.QueryDocumentation(option)
		if docErr != nil || doc == "" {
			fmt.Fprintln(os.Stderr, utils.FormatError("No documentation found for Home Manager option: "+option))
			os.Exit(1)
		}
		aiProvider := ai.NewOllamaProvider("llama3") // Default, or use config
		prompt := buildExplainHomeOptionPrompt(option, doc)
		aiResp, aiErr := aiProvider.Query(prompt)
		if aiErr != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+aiErr.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(aiResp))
	},
}

// explainOptionCmd implements the explain-option command
var explainOptionCmd = &cobra.Command{
	Use:   "explain-option <option>",
	Short: "Explain a NixOS option using AI and documentation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		option := args[0]
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}
		mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
		mcpClient := mcp.NewMCPClient(mcpURL)
		doc, docErr := mcpClient.QueryDocumentation(option)
		if docErr != nil || doc == "" {
			fmt.Fprintln(os.Stderr, utils.FormatError("No documentation found for option: "+option))
			os.Exit(1)
		}
		aiProvider := ai.NewOllamaProvider("llama3") // Default, or use config
		prompt := buildExplainOptionPrompt(option, doc)
		aiResp, aiErr := aiProvider.Query(prompt)
		if aiErr != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+aiErr.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(aiResp))
	},
}

// interactiveCmd implements the interactive CLI mode
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive AI-powered NixOS assistant shell",
	Long: `Start an interactive shell for NixOS troubleshooting, package search, option explanation, and more.

Features:
- Live MCP-powered option completion
- Animated snowflake progress indicator
- Multi-line input, contextual help, and advanced autocomplete
- All advanced features available in non-interactive mode

Examples:
  nixai interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		InteractiveMode()
	},
}

// Stub commands for missing top-level commands
var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask a question about NixOS configuration",
	Long: `Ask a direct question about NixOS configuration and get an AI-powered answer.

Examples:
  nixai ask "How do I configure nginx?"
  nixai ask "What is the difference between services.openssh.enable and programs.ssh.enable?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := strings.Join(args, " ")
		fmt.Println(utils.FormatHeader("🤖 AI Answer to your question:"))

		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}

		provider := InitializeAIProvider(cfg)
		answer, err := provider.Query(question)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(answer))
	},
}
var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Community resources and support (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'community' command is not yet implemented.")
	},
}
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage nixai configuration settings",
	Long: `Manage nixai configuration settings including AI provider, model, and other options.

Available subcommands:
  show                    - Show current configuration
  set <key> <value>       - Set a configuration value
  get <key>               - Get a configuration value
  reset                   - Reset to default configuration

Examples:
  nixai config show
  nixai config set ai_provider ollama
  nixai config set ai_model llama3
  nixai config get ai_provider`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		switch args[0] {
		case "show":
			showConfig()
		case "set":
			if len(args) < 3 {
				fmt.Println(utils.FormatError("Usage: nixai config set <key> <value>"))
				os.Exit(1)
			}
			setConfig(args[1], args[2])
		case "get":
			if len(args) < 2 {
				fmt.Println(utils.FormatError("Usage: nixai config get <key>"))
				os.Exit(1)
			}
			getConfig(args[1])
		case "reset":
			resetConfig()
		default:
			fmt.Println(utils.FormatError("Unknown config command: " + args[0]))
			_ = cmd.Help()
			os.Exit(1)
		}
	},
}
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure NixOS interactively (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'configure' command is not yet implemented.")
	},
}
var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose NixOS issues (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'diagnose' command is not yet implemented.")
	},
}
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run NixOS health checks (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'doctor' command is not yet implemented.")
	},
}
var flakeCmd = &cobra.Command{
	Use:   "flake",
	Short: "Nix flake utilities (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'flake' command is not yet implemented.")
	},
}
var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "NixOS learning and training commands (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'learn' command is not yet implemented.")
	},
}
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Analyze and parse NixOS logs (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'logs' command is not yet implemented.")
	},
}
var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start or manage the MCP server (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'mcp-server' command is not yet implemented.")
	},
}
var neovimSetupCmd = &cobra.Command{
	Use:   "neovim-setup",
	Short: "Neovim integration setup (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'neovim-setup' command is not yet implemented.")
	},
}
var packageRepoCmd = &cobra.Command{
	Use:   "package-repo",
	Short: "Analyze Git repos and generate Nix derivations (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'package-repo' command is not yet implemented.")
	},
}

// initializeCommands adds all commands to the root command
func initializeCommands() {
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(explainOptionCmd)
	rootCmd.AddCommand(explainHomeOptionCmd)
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(enhancedBuildCmd)
	rootCmd.AddCommand(NewDepsCommand())
	rootCmd.AddCommand(devenvCmd)
	rootCmd.AddCommand(gcCmd)
	rootCmd.AddCommand(hardwareCmd)
	rootCmd.AddCommand(createMachinesCommand())
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(templatesCmd)
	rootCmd.AddCommand(snippetsCmd)
	// Register stub commands for missing features
	rootCmd.AddCommand(communityCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(flakeCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(mcpServerCmd)
	rootCmd.AddCommand(neovimSetupCmd)
	rootCmd.AddCommand(packageRepoCmd)
}

// Execute runs the root command
func Execute() {
	cobra.OnInitialize(func() {
		if nixosPath != "" {
			os.Setenv("NIXAI_NIXOS_PATH", nixosPath)
		}
	})
	if showVersion {
		v := version.Get()
		fmt.Println(utils.FormatHeader("nixai version:"))
		fmt.Println(utils.FormatKeyValue("Version", v.Version))
		fmt.Println(utils.FormatKeyValue("Git Commit", v.GitCommit))
		fmt.Println(utils.FormatKeyValue("Build Date", v.BuildDate))
		fmt.Println(utils.FormatKeyValue("Go Version", v.GoVersion))
		fmt.Println(utils.FormatKeyValue("Platform", v.Platform))
		os.Exit(0)
	}
	if askQuestion != "" {
		fmt.Println(utils.FormatHeader("🤖 AI Answer to your question:"))
		aiProvider := ai.NewOllamaProvider("llama3") // Default, or use config
		answer, err := aiProvider.Query(askQuestion)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(answer))
		os.Exit(0)
	}
	initializeCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
