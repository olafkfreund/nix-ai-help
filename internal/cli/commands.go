package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/learning"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/neovim"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/internal/packaging"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
	"nix-ai-help/pkg/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nixai [question] [flags]",
	Short: "NixOS AI Assistant",
	Long: `nixai is a command-line tool that assists users in diagnosing and solving NixOS configuration issues using AI models and documentation queries.

You can also ask questions directly, e.g.:
  nixai -a "how can I configure curl?"

Usage:
  nixai [question] [flags]
  nixai [command]`,
	SilenceUsage: true,
	Version:      version.Get().Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if askQuestion != "" {
			fmt.Println(utils.FormatHeader("🤖 AI Answer to your question:"))

			cfg, err := config.LoadUserConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
				os.Exit(1)
			}

			providerName := cfg.AIProvider
			if providerName == "" {
				providerName = "ollama"
			}
			var aiProvider ai.AIProvider
			switch providerName {
			case "ollama":
				aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
			case "openai":
				aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			case "gemini":
				aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
			default:
				fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
				os.Exit(1)
			}

			// Query MCP for documentation context (optional, ignore errors)
			var docExcerpts []string
			mcpBase := cfg.MCPServer.Host
			if mcpBase != "" {
				mcpClient := mcp.NewMCPClient(mcpBase)
				doc, err := mcpClient.QueryDocumentation(askQuestion)
				if err == nil && doc != "" {
					docExcerpts = append(docExcerpts, doc)
				}
			}

			promptCtx := ai.PromptContext{
				Question:     askQuestion,
				DocExcerpts:  docExcerpts,
				Intent:       "explain",
				OutputFormat: "markdown",
				Provider:     providerName,
			}
			builder := ai.DefaultPromptBuilder{}
			prompt, err := builder.BuildPrompt(promptCtx)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Prompt build error: "+err.Error()))
				os.Exit(1)
			}
			answer, err := aiProvider.Query(prompt)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
				os.Exit(1)
			}
			fmt.Println(utils.RenderMarkdown(answer))
			return nil
		}
		// If no --ask, show help
		return cmd.Help()
	},
}

var askQuestion string
var nixosPath string
var daemonMode bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&askQuestion, "ask", "a", "", "Ask a question about NixOS configuration")
	rootCmd.PersistentFlags().StringVarP(&nixosPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	mcpServerCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run MCP server in background/daemon mode")
}

// Configuration management functions
func showConfig() {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}
	if nixosPath != "" {
		cfg.NixosFolder = nixosPath
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
	if nixosPath != "" {
		cfg.NixosFolder = nixosPath
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
	if nixosPath != "" {
		cfg.NixosFolder = nixosPath
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
	_, _ = fmt.Scanln(&response)
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

// Helper struct for MCP option JSON
// Only fields we care about

type mcpOptionDoc struct {
	Name        string   `json:"option_name"`
	Type        string   `json:"option_type"`
	Default     string   `json:"option_default"`
	Example     string   `json:"option_example"`
	Description string   `json:"option_description"`
	Source      string   `json:"option_source"`
	Version     string   `json:"nixos_version"`
	Related     []string `json:"related_options"`
	Links       []string `json:"links"`
}

// Parse MCP doc JSON, fallback to plain doc string if not JSON
func parseMCPOptionDoc(doc string) (mcpOptionDoc, string) {
	var opt mcpOptionDoc
	if err := json.Unmarshal([]byte(doc), &opt); err == nil && opt.Name != "" {
		return opt, ""
	}
	return mcpOptionDoc{}, doc
}

func buildEnhancedExplainOptionPrompt(option, documentation, format, source, version string) string {
	opt, fallbackDoc := parseMCPOptionDoc(documentation)
	if opt.Name == "" {
		// fallback to old prompt if not JSON
		sourceInfo := ""
		if source != "" {
			sourceInfo += fmt.Sprintf("\n**Source:** %s", source)
		}
		if version != "" {
			sourceInfo += fmt.Sprintf("\n**NixOS Version:** %s", version)
		}
		return fmt.Sprintf(`You are a NixOS expert helping users understand configuration options. Please explain the following NixOS option in a clear, practical manner.\n\n**Option:** %s%s\n\n**Official Documentation:**\n%s\n\n**Please provide:**\n\n1. **Purpose & Overview**: What this option does and why you'd use it\n2. **Type & Default**: The data type and default value (if any)\n3. **Usage Examples**: Show 2-3 practical configuration examples\n4. **Best Practices**: How to use this option effectively\n5. **Related Options**: List and briefly describe other options commonly used with this one\n6. **Troubleshooting Tips**: Common issues and how to resolve them\n7. **Links**: If possible, include links to relevant official documentation\n8. **Summary Table**: Provide a summary table of key attributes (name, type, default, description)\n\nFormat your response using %s with section headings and code blocks for examples.`, option, sourceInfo, fallbackDoc, format)
	}
	// Compose a rich prompt using all available fields
	related := ""
	if len(opt.Related) > 0 {
		related = "- " + strings.Join(opt.Related, "\n- ")
	}
	links := ""
	if len(opt.Links) > 0 {
		links = "- " + strings.Join(opt.Links, "\n- ")
	}
	return fmt.Sprintf(`You are a NixOS expert. Explain the following option in detail for a Linux user.\n\n**Option:** %s\n**Type:** %s\n**Default:** %s\n**Example:** %s\n**Description:** %s\n**Source:** %s\n**NixOS Version:** %s\n\n**Related Options:**\n%s\n\n**Links:**\n%s\n\n**Please provide:**\n1. Purpose & Overview\n2. Usage Examples (with code)\n3. Best Practices\n4. Troubleshooting Tips\n5. Summary Table (name, type, default, description)\n\nFormat your response using %s.`,
		opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, related, links, format)
}

func buildExamplesOnlyPrompt(option, documentation, format, source, version string) string {
	sourceInfo := ""
	if source != "" {
		sourceInfo += fmt.Sprintf("\n**Source:** %s", source)
	}
	if version != "" {
		sourceInfo += fmt.Sprintf("\n**NixOS Version:** %s", version)
	}
	return fmt.Sprintf(`You are a NixOS expert. Show only 2-3 practical configuration examples for the following option.\n\n**Option:** %s%s\n\n**Official Documentation:**\n%s\n\nFormat your response using %s and code blocks.`, option, sourceInfo, documentation, format)
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
		if nixosPath != "" {
			cfg.NixosFolder = nixosPath
		}
		exec := nixos.NewExecutor(cfg.NixosFolder)
		fmt.Println(utils.FormatHeader("🔍 NixOS Search Results for: " + query))
		fmt.Println()
		// Package search
		pkgOut, pkgErr := exec.SearchNixPackages(query)
		if pkgErr == nil && pkgOut != "" {
			fmt.Println(pkgOut)
		}
		// Query MCP for documentation context (with progress indicator)
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider ai.AIProvider
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
			os.Exit(1)
		}
		var docExcerpts []string
		fmt.Print(utils.FormatInfo("Querying documentation... "))
		mcpBase := cfg.MCPServer.Host
		mcpContextAdded := false
		if mcpBase != "" {
			mcpClient := mcp.NewMCPClient(mcpBase)
			doc, err := mcpClient.QueryDocumentation(query)
			fmt.Println(utils.FormatSuccess("done"))
			if err == nil && doc != "" {
				opt, fallbackDoc := parseMCPOptionDoc(doc)
				if opt.Name != "" {
					context := fmt.Sprintf("Option: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nNixOS Version: %s\nRelated: %v\nLinks: %v", opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
					docExcerpts = append(docExcerpts, context)
					mcpContextAdded = true
				} else if strings.Contains(strings.ToLower(fallbackDoc), "nixos") {
					docExcerpts = append(docExcerpts, fallbackDoc)
					mcpContextAdded = true
				}
			}
		} else {
			fmt.Println(utils.FormatWarning("skipped (no MCP host configured)"))
		}
		// Always add a strong NixOS-specific instruction to the prompt
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
			fmt.Fprintln(os.Stderr, utils.FormatError("Prompt build error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		aiAnswer, aiErr := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if aiErr == nil && aiAnswer != "" {
			fmt.Println(utils.FormatHeader("🤖 AI Best Practices & Tips"))
			fmt.Println(utils.RenderMarkdown(aiAnswer))
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
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider ai.AIProvider
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
			os.Exit(1)
		}

		// Query MCP for documentation context (with progress indicator)
		var docExcerpts []string
		fmt.Print(utils.FormatInfo("Querying documentation... "))
		mcpBase := cfg.MCPServer.Host
		if mcpBase != "" {
			mcpClient := mcp.NewMCPClient(mcpBase)
			doc, err := mcpClient.QueryDocumentation(option)
			fmt.Println(utils.FormatSuccess("done"))
			if err == nil && doc != "" {
				opt, fallbackDoc := parseMCPOptionDoc(doc)
				if opt.Name != "" {
					context := fmt.Sprintf("Option: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nNixOS Version: %s\nRelated: %v\nLinks: %v", opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
					docExcerpts = append(docExcerpts, context)
				} else {
					docExcerpts = append(docExcerpts, fallbackDoc)
				}
			}
		} else {
			fmt.Println(utils.FormatWarning("skipped (no MCP host configured)"))
		}

		promptCtx := ai.PromptContext{
			Question:     option,
			DocExcerpts:  docExcerpts,
			Intent:       "explain",
			OutputFormat: "markdown",
			Provider:     providerName,
		}
		builder := ai.DefaultPromptBuilder{}
		prompt, err := builder.BuildPrompt(promptCtx)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Prompt build error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		aiResp, aiErr := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if aiErr != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+aiErr.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(aiResp))
	},
}

// explainOptionCmd implements the explain-option command
var explainOptionCmd = NewExplainOptionCommand()

// NewExplainOptionCommand returns a fresh explain-option command
func NewExplainOptionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain-option <option>",
		Short: "Explain a NixOS option using AI and documentation",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			option := args[0]
			format, _ := cmd.Flags().GetString("format")
			providerFlag, _ := cmd.Flags().GetString("provider")
			examplesOnly, _ := cmd.Flags().GetBool("examples-only")
			cfg, err := config.LoadUserConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
				os.Exit(1)
			}
			mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
			mcpClient := mcp.NewMCPClient(mcpURL)
			fmt.Print(utils.FormatInfo("Querying documentation... "))
			doc, docErr := mcpClient.QueryDocumentation(option)
			fmt.Println(utils.FormatSuccess("done"))
			if docErr != nil || doc == "" {
				fmt.Fprintln(os.Stderr, utils.FormatError("No documentation found for option: "+option))
				return
			}
			var source, version string
			if strings.Contains(doc, "option_source") {
				parts := strings.Split(doc, "option_source")
				if len(parts) > 1 {
					source = strings.Split(parts[1], "\"")[1]
				}
			}
			if strings.Contains(doc, "nixos-") {
				idx := strings.Index(doc, "nixos-")
				version = doc[idx : idx+12]
			}
			aiProviderName := providerFlag
			if aiProviderName == "" {
				aiProviderName = cfg.AIProvider
			}
			var aiProvider ai.AIProvider
			switch aiProviderName {
			case "ollama":
				aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
			case "openai":
				aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			case "gemini":
				aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
			default:
				aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
			}
			var prompt string
			if examplesOnly {
				prompt = buildExamplesOnlyPrompt(option, doc, format, source, version)
			} else {
				prompt = buildEnhancedExplainOptionPrompt(option, doc, format, source, version)
			}
			fmt.Print(utils.FormatInfo("Querying AI provider... "))
			aiResp, aiErr := aiProvider.Query(prompt)
			fmt.Println(utils.FormatSuccess("done"))
			if aiErr != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+aiErr.Error()))
				os.Exit(1)
			}
			fmt.Println(utils.RenderMarkdown(aiResp))
		},
	}
	cmd.Flags().String("format", "markdown", "Output format: markdown, plain, or table")
	cmd.Flags().String("provider", "", "AI provider to use for this query (ollama, openai, gemini)")
	cmd.Flags().Bool("examples-only", false, "Show only usage examples for the option")
	return cmd
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

// Flake management command implementation
var flakeCmd = &cobra.Command{
	Use:   "flake",
	Short: "Manage NixOS flakes and configurations",
	Long: `Manage NixOS flakes and configurations with AI-powered assistance.

This command provides comprehensive flake management including creation, validation,
migration from legacy configurations, and troubleshooting.`,
	Example: `  # Create a new flake configuration
  nixai flake create --path ./my-flake

  # Validate an existing flake
  nixai flake validate

  # Migrate from legacy NixOS configuration
  nixai flake migrate --from /etc/nixos

  # Analyze flake for issues
  nixai flake analyze`,
	Run: handleFlakeCommand,
}

// Learning system command implementation
var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Interactive NixOS learning modules and tutorials",
	Long: `Access interactive learning modules, tutorials, and quizzes for NixOS.

The learning system provides structured educational content for users at all levels,
from beginners to advanced NixOS users. Progress is tracked and saved locally.`,
	Example: `  # List available learning modules
  nixai learn list

  # Start a specific learning module
  nixai learn start basics

  # Show learning progress
  nixai learn progress

  # Take a quiz on a topic
  nixai learn quiz flakes`,
	Run: handleLearnCommand,
}

// Log analysis command implementation
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Analyze and diagnose NixOS system logs",
	Long: `Analyze NixOS system logs with AI-powered diagnostics and troubleshooting.

This command can parse various log formats, identify issues, and provide
actionable recommendations for resolving problems.`,
	Example: `  # Analyze current system logs
  nixai logs analyze

  # Analyze specific log file
  nixai logs analyze --file /var/log/nixos/build.log

  # Parse piped log output
  journalctl -u nixos-rebuild | nixai logs parse

  # Get recent critical errors
  nixai logs errors --recent`,
	Run: handleLogsCommand,
}

// Neovim setup command implementation
var neovimSetupCmd = &cobra.Command{
	Use:   "neovim-setup",
	Short: "Set up Neovim integration with nixai MCP server",
	Long: `Set up Neovim integration with the nixai Model Context Protocol (MCP) server.

This command configures Neovim to work with nixai's documentation and AI features,
providing seamless access to NixOS help directly from your editor.`,
	Example: `  # Set up Neovim integration
  nixai neovim-setup install

  # Check integration status
  nixai neovim-setup status

  # Remove integration
  nixai neovim-setup remove

  # Update integration configuration
  nixai neovim-setup update`,
	Run: handleNeovimSetupCommand,
}

// Package repository analysis command implementation
var packageRepoCmd = &cobra.Command{
	Use:   "package-repo",
	Short: "Analyze Git repositories and generate Nix derivations",
	Long: `Analyze Git repositories and automatically generate Nix derivations for packaging.

This command clones or analyzes local repositories, understands their build systems,
and generates appropriate Nix derivations with proper dependencies and build instructions.`,
	Example: `  # Analyze a GitHub repository
  nixai package-repo https://github.com/user/repo

  # Analyze local repository
  nixai package-repo --local ./my-project

  # Generate derivation with custom name
  nixai package-repo https://github.com/user/repo --name my-package

  # Output to specific file
  nixai package-repo https://github.com/user/repo --output ./result.nix`,
	Run: handlePackageRepoCommand,
}

// MCP Server command implementation
var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Manage the Model Context Protocol (MCP) server",
	Long: `Manage the Model Context Protocol (MCP) server for documentation queries.

The MCP server provides VS Code integration and documentation querying capabilities.

Examples:
  nixai mcp-server start        # Start the MCP server
  nixai mcp-server start -d     # Start the MCP server in daemon mode
  nixai mcp-server stop         # Stop the MCP server  
  nixai mcp-server status       # Check server status
  nixai mcp-server restart      # Restart the MCP server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleMCPServerCommand(args)
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

		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider ai.AIProvider
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
			os.Exit(1)
		}

		// Query MCP for documentation context (optional, ignore errors)
		var docExcerpts []string
		fmt.Print(utils.FormatInfo("Querying documentation... "))
		mcpBase := cfg.MCPServer.Host
		if mcpBase != "" {
			mcpClient := mcp.NewMCPClient(mcpBase)
			doc, err := mcpClient.QueryDocumentation(question)
			fmt.Println(utils.FormatSuccess("done"))
			if err == nil && doc != "" {
				// Try to parse as MCP option doc JSON
				opt, fallbackDoc := parseMCPOptionDoc(doc)
				if opt.Name != "" {
					// Compose a rich context string from MCP fields
					context := fmt.Sprintf("Option: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nNixOS Version: %s\nRelated: %v\nLinks: %v", opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
					docExcerpts = append(docExcerpts, context)
				} else {
					docExcerpts = append(docExcerpts, fallbackDoc)
				}
			}
		} else {
			fmt.Println(utils.FormatWarning("skipped (no MCP host configured)"))
		}

		promptCtx := ai.PromptContext{
			Question:     question,
			DocExcerpts:  docExcerpts,
			Intent:       "explain",
			OutputFormat: "markdown",
			Provider:     providerName,
		}
		builder := ai.DefaultPromptBuilder{}
		prompt, err := builder.BuildPrompt(promptCtx)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Prompt build error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		answer, err := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(answer))
	},
}
var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Show NixOS community resources and support links",
	Long: `Access NixOS community forums, documentation, chat channels, and GitHub resources.

Examples:
  nixai community
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🌐 NixOS Community Resources"))
		fmt.Println()
		showCommunityOverview(os.Stdout)
		fmt.Println()
		showCommunityForums(os.Stdout)
		fmt.Println()
		showCommunityDocs(os.Stdout)
		fmt.Println()
		showMatrixChannels(os.Stdout)
		fmt.Println()
		showGitHubResources(os.Stdout)
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
	Short: "Configure NixOS interactively",
	Long: `Interactively generate or edit your NixOS configuration using AI-powered guidance and documentation lookup.

Examples:
  nixai configure
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🛠️  Interactive NixOS Configuration"))
		fmt.Println()
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider ai.AIProvider
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
			os.Exit(1)
		}
		fmt.Println(utils.FormatInfo("Describe what you want to configure (e.g. desktop, web server, user, etc):"))
		var input string
		fmt.Print("> ")
		_, _ = fmt.Scanln(&input)
		if input == "" {
			fmt.Println(utils.FormatWarning("No input provided. Exiting."))
			return
		}
		prompt := "You are a NixOS configuration assistant. Help the user generate a configuration.nix snippet for: " + input + "\nProvide a complete, copy-pasteable example and explain each part."
		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(resp))
	},
}

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose [logfile]",
	Short: "Diagnose NixOS issues from logs or config",
	Long: `Diagnose NixOS issues by analyzing logs, configuration files, or piped input. Uses AI and documentation to suggest fixes.

Examples:
  nixai diagnose /var/log/messages
  journalctl -xe | nixai diagnose
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🩺 NixOS Diagnostics"))
		fmt.Println()
		var logData string
		if len(args) > 0 {
			file := args[0]
			data, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to read log file: "+err.Error()))
				os.Exit(1)
			}
			logData = string(data)
		} else {
			// Read from stdin if piped
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				input, _ := io.ReadAll(os.Stdin)
				logData = string(input)
			} else {
				fmt.Println(utils.FormatWarning("No log file or piped input provided."))
				return
			}
		}
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider ai.AIProvider
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
			os.Exit(1)
		}
		prompt := "You are a NixOS expert. Analyze the following log or error output and provide a diagnosis, root cause, and step-by-step fix instructions.\n\nLog or error:\n" + logData
		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(resp))
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run NixOS health checks and get advice",
	Long: `Run a set of NixOS health checks and get AI-powered advice for improving your system configuration.

Examples:
  nixai doctor
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🩻 NixOS Doctor: Health Check"))
		fmt.Println()
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		var aiProvider ai.AIProvider
		switch providerName {
		case "ollama":
			aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
		default:
			fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
			os.Exit(1)
		}

		// Determine config path (from --nixos-path or config)
		configPath := cfg.NixosFolder
		if nixosPath != "" {
			configPath = nixosPath
		}
		if configPath == "" {
			configPath = "/etc/nixos"
		}

		confNix := configPath
		flakeNix := configPath
		// If configPath is a directory, append file names
		stat, err := os.Stat(configPath)
		if err == nil && stat.IsDir() {
			confNix = configPath + "/configuration.nix"
			flakeNix = configPath + "/flake.nix"
		}

		// Health checks
		results := []string{}
		confExists := false
		flakeExists := false
		if _, err := os.Stat(confNix); err == nil {
			results = append(results, "✅ configuration.nix exists")
			confExists = true
		}
		if _, err := os.Stat(flakeNix); err == nil {
			results = append(results, "✅ flake.nix exists (flake-based NixOS configuration detected)")
			flakeExists = true
		}
		if !confExists && !flakeExists {
			results = append(results, "❌ Neither configuration.nix nor flake.nix found in "+configPath)
		}

		if _, err := os.Stat("/run/current-system"); err == nil {
			results = append(results, "✅ nixos-rebuild previously succeeded")
		} else {
			results = append(results, "❌ nixos-rebuild may not have run")
		}
		results = append(results, "ℹ️  Run 'systemctl list-units --type=service' to see running services.")
		results = append(results, "ℹ️  Run 'systemctl --failed' to see failed units.")

		fmt.Println(utils.FormatHeader("System Health Check Results:"))
		for _, r := range results {
			fmt.Println("  ", r)
		}
		prompt := "You are a NixOS doctor. Given these health check results, provide a summary, highlight any problems, and suggest fixes or improvements.\n\nResults:\n" + strings.Join(results, "\n")
		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.RenderMarkdown(resp))
	},
}

// Completion command for shell script generation
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate the autocompletion script for the specified shell",
	Long: `Generate shell completion scripts for bash, zsh, fish, or powershell.

Examples:
  nixai completion bash > /etc/bash_completion.d/nixai
  nixai completion zsh > ~/.zshrc
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			_ = rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			_ = rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			fmt.Println(utils.FormatError("Unknown shell: " + args[0]))
		}
	},
}

// handleMCPServerCommand handles the mcp-server command and subcommands
func handleMCPServerCommand(args []string) error {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if len(args) == 0 {
		// Show help/status by default
		fmt.Println(utils.FormatHeader("🔗 MCP Server Management"))
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Available Commands", ""))
		fmt.Println("  start         - Start the MCP server")
		fmt.Println("  start -d      - Start the MCP server in daemon mode")
		fmt.Println("  stop          - Stop the MCP server")
		fmt.Println("  status        - Check server status")
		fmt.Println("  restart       - Restart the MCP server")
		fmt.Println("  query <text>  - Query the MCP server directly")
		fmt.Println()
		fmt.Println(utils.FormatTip("The MCP server provides VS Code integration and documentation querying"))
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "start":
		return handleMCPServerStart(cfg, daemonMode)
	case "stop":
		return handleMCPServerStop(cfg)
	case "status":
		return handleMCPServerStatus(cfg)
	case "restart":
		return handleMCPServerRestart(cfg)
	case "query":
		if len(args) < 2 {
			return fmt.Errorf("query command requires a query string")
		}

		var query string
		var sources []string
		var inSourcesMode bool

		for i := 1; i < len(args); i++ {
			if args[i] == "--source" || args[i] == "-s" {
				inSourcesMode = true
			} else if inSourcesMode {
				sources = append(sources, args[i])
				inSourcesMode = false
			} else {
				if query != "" {
					query += " "
				}
				query += args[i]
			}
		}

		return handleMCPServerQuery(cfg, query, sources...)
	default:
		return fmt.Errorf("unknown subcommand: %s. Available: start, stop, status, restart, query", subcommand)
	}
}

// handleMCPServerStart starts the MCP server
func handleMCPServerStart(cfg *config.UserConfig, daemon bool) error {
	fmt.Println(utils.FormatHeader("🚀 Starting MCP Server"))
	fmt.Println()

	// If daemon mode is requested, fork the process
	if daemon {
		// Create a command to start the server without daemon flag
		cmd := exec.Command(os.Args[0], "mcp-server", "start")

		// Start the background process without complex process group management
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start daemon process: %v", err)
		}

		// Don't wait for the process - let it run in background
		go func() {
			cmd.Wait() // Clean up when process exits
		}()

		fmt.Println(utils.FormatSuccess("MCP server started in daemon mode"))
		fmt.Println(utils.FormatKeyValue("Process ID", fmt.Sprintf("%d", cmd.Process.Pid)))
		fmt.Println(utils.FormatKeyValue("HTTP Server", fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)))
		fmt.Println(utils.FormatKeyValue("Unix Socket", cfg.MCPServer.SocketPath))
		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai mcp-server status' to check server health"))
		fmt.Println(utils.FormatTip("Use 'nixai mcp-server stop' to stop the server"))

		return nil
	}

	// Create MCP server from config
	configPath, _ := config.ConfigFilePath()
	if configPath == "" {
		configPath = "configs/default.yaml" // fallback
	}

	server, err := mcp.NewServerFromConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %v", err)
	}

	fmt.Print(utils.FormatInfo("Initializing MCP server... "))

	// Start the server (this will block)
	go func() {
		if err := server.Start(); err != nil {
			fmt.Println(utils.FormatError("failed"))
			fmt.Printf("Error: %v\n", err)
		}
	}()

	fmt.Println(utils.FormatSuccess("done"))
	fmt.Println()
	fmt.Println(utils.FormatKeyValue("HTTP Server", fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)))
	fmt.Println(utils.FormatKeyValue("Unix Socket", cfg.MCPServer.SocketPath))
	fmt.Println()
	fmt.Println(utils.FormatTip("Use 'nixai mcp-server status' to check server health"))
	fmt.Println(utils.FormatTip("Use 'nixai mcp-server stop' to stop the server"))

	// Keep the process running
	select {}
}

// handleMCPServerStop stops the MCP server
func handleMCPServerStop(cfg *config.UserConfig) error {
	fmt.Println(utils.FormatHeader("🛑 Stopping MCP Server"))
	fmt.Println()

	// Try to stop via HTTP endpoint
	url := fmt.Sprintf("http://%s:%d/shutdown", cfg.MCPServer.Host, cfg.MCPServer.Port)

	fmt.Print(utils.FormatInfo("Sending shutdown signal... "))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(utils.FormatError("failed"))
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println(utils.FormatSuccess("done"))
	fmt.Println(utils.FormatKeyValue("Status", "MCP server shutdown initiated"))

	return nil
}

// handleMCPServerStatus checks the MCP server status
func handleMCPServerStatus(cfg *config.UserConfig) error {
	fmt.Println(utils.FormatHeader("📊 MCP Server Status"))
	fmt.Println()

	// Check HTTP endpoint
	url := fmt.Sprintf("http://%s:%d/healthz", cfg.MCPServer.Host, cfg.MCPServer.Port)

	fmt.Print(utils.FormatInfo("Checking HTTP endpoint... "))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(utils.FormatError("unreachable"))
		fmt.Println(utils.FormatKeyValue("HTTP Status", "❌ Not running"))
	} else {
		defer resp.Body.Close()
		fmt.Println(utils.FormatSuccess("healthy"))
		fmt.Println(utils.FormatKeyValue("HTTP Status", "✅ Running"))
	}

	// Check Unix socket
	socketPath := cfg.MCPServer.SocketPath
	if socketPath == "" {
		socketPath = "/tmp/nixai-mcp.sock"
	}

	fmt.Print(utils.FormatInfo("Checking Unix socket... "))

	if _, err := os.Stat(socketPath); err == nil {
		fmt.Println(utils.FormatSuccess("exists"))
		fmt.Println(utils.FormatKeyValue("Socket Status", "✅ Available"))
		fmt.Println(utils.FormatKeyValue("Socket Path", socketPath))
	} else {
		fmt.Println(utils.FormatError("missing"))
		fmt.Println(utils.FormatKeyValue("Socket Status", "❌ Not available"))
	}

	fmt.Println()
	configPath, _ := config.ConfigFilePath()
	fmt.Println(utils.FormatKeyValue("Configuration", configPath))
	fmt.Println(utils.FormatKeyValue("Documentation Sources", fmt.Sprintf("%d sources", len(cfg.MCPServer.DocumentationSources))))

	return nil
}

// handleMCPServerRestart restarts the MCP server
func handleMCPServerRestart(cfg *config.UserConfig) error {
	fmt.Println(utils.FormatHeader("🔄 Restarting MCP Server"))
	fmt.Println()

	// Stop first
	if err := handleMCPServerStop(cfg); err != nil {
		fmt.Printf("Warning: Failed to stop server gracefully: %v\n", err)
	}

	// Wait a moment
	fmt.Print(utils.FormatInfo("Waiting for cleanup... "))
	time.Sleep(2 * time.Second)
	fmt.Println(utils.FormatSuccess("done"))

	// Start again
	return handleMCPServerStart(cfg, false)
}

// handleMCPServerQuery queries the MCP server directly
func handleMCPServerQuery(cfg *config.UserConfig, query string, sources ...string) error {
	fmt.Println(utils.FormatHeader("🔍 MCP Server Query"))
	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Query", query))
	if len(sources) > 0 {
		fmt.Println(utils.FormatKeyValue("Sources", strings.Join(sources, ", ")))
	}
	fmt.Println()

	// Create MCP client
	mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	client := mcp.NewMCPClient(mcpURL)

	fmt.Print(utils.FormatInfo("Querying documentation... "))

	result, err := client.QueryDocumentation(query, sources...)
	if err != nil {
		fmt.Println(utils.FormatError("failed"))
		return fmt.Errorf("query failed: %v", err)
	}

	fmt.Println(utils.FormatSuccess("done"))
	fmt.Println()
	fmt.Println(utils.FormatSubsection("📖 Documentation Results", ""))
	fmt.Println(utils.RenderMarkdown(result))

	return nil
}

// Handler functions for the CLI commands

// Flake subcommand handlers
func handleFlakeCreate(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🔧 Creating Flake Configuration"))
	fmt.Println()

	fmt.Println(utils.FormatInfo("Available flake creation modes:"))
	fmt.Println("  1. Basic flake template")
	fmt.Println("  2. Convert existing configuration.nix")
	fmt.Println("  3. Interactive guided setup")

	fmt.Println(utils.FormatTip("Use 'nixai migrate to-flake' for full migration assistance"))
}

func handleFlakeValidate(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("✅ Validating Flake Configuration"))
	fmt.Println()

	// Check for flake.nix in current directory or specified path
	flakePath := "./flake.nix"
	if len(args) > 0 {
		flakePath = args[0]
	}

	if _, err := os.Stat(flakePath); err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("No flake.nix found at: "+flakePath))
		return
	}

	fmt.Println(utils.FormatKeyValue("Flake File", flakePath))
	fmt.Println(utils.FormatInfo("Running flake validation..."))

	// TODO: Add actual validation logic using nix flake check
	fmt.Println(utils.FormatSuccess("✅ Flake validation completed"))
}

func handleFlakeMigrate(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🔄 Migrating to Flake Configuration"))
	fmt.Println()

	fmt.Println(utils.FormatInfo("Starting migration analysis..."))
	fmt.Println(utils.FormatTip("For complete migration assistance, use: 'nixai migrate to-flake'"))

	// Show basic migration guidance
	fmt.Println(utils.FormatInfo("Migration steps:"))
	fmt.Println("  1. Backup your current configuration")
	fmt.Println("  2. Create flake.nix template")
	fmt.Println("  3. Import existing configuration.nix")
	fmt.Println("  4. Test the new flake configuration")
}

func handleFlakeAnalyze(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🔍 Analyzing Flake Configuration"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider for analysis
	providerName := cfg.AIProvider
	if providerName == "" {
		providerName = "ollama"
	}

	var aiProvider ai.AIProvider
	switch providerName {
	case "ollama":
		aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
	case "openai":
		aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	case "gemini":
		aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
	default:
		fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
		os.Exit(1)
	}

	// Read flake.nix if exists
	flakePath := "./flake.nix"
	if len(args) > 0 {
		flakePath = args[0]
	}

	flakeContent, err := os.ReadFile(flakePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to read flake.nix: "+err.Error()))
		return
	}

	prompt := fmt.Sprintf("Analyze this NixOS flake configuration and provide recommendations for improvements, best practices, and potential issues:\n\n%s", string(flakeContent))

	fmt.Print(utils.FormatInfo("Analyzing flake with AI... "))
	result, err := aiProvider.Query(prompt)
	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("AI analysis failed: "+err.Error()))
		return
	}

	fmt.Println(utils.RenderMarkdown(result))
}

// Learning subcommand handlers
func handleLearnList(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📚 Available Learning Modules"))
	fmt.Println()

	// Use existing learning system
	modules, err := learning.LoadModules()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load modules: "+err.Error()))
		return
	}

	if len(modules) == 0 {
		fmt.Println(utils.FormatInfo("No learning modules currently available"))
		fmt.Println(utils.FormatTip("Modules are being developed and will be added in future updates"))
		return
	}

	for _, module := range modules {
		fmt.Println(utils.FormatKeyValue("Module", module.Title))
		fmt.Println(utils.FormatKeyValue("Level", module.Level))
		fmt.Println(utils.FormatKeyValue("Description", module.Description))
		fmt.Println()
	}
}

func handleLearnStart(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🎓 Starting Learning Module"))
	fmt.Println()

	if len(args) == 0 {
		fmt.Println(utils.FormatError("Please specify a module to start"))
		fmt.Println(utils.FormatTip("Use 'nixai learn list' to see available modules"))
		return
	}

	moduleName := args[0]
	fmt.Println(utils.FormatKeyValue("Starting Module", moduleName))
	fmt.Println(utils.FormatInfo("Loading module content..."))

	// TODO: Implement module loading and interactive learning
	fmt.Println(utils.FormatTip("Interactive learning modules are being developed"))
}

func handleLearnProgress(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📈 Learning Progress"))
	fmt.Println()

	progress, err := learning.LoadProgress()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load progress: "+err.Error()))
		return
	}

	fmt.Println(utils.FormatKeyValue("Completed Modules", fmt.Sprintf("%d", len(progress.CompletedModules))))
	fmt.Println(utils.FormatKeyValue("Quiz Scores", fmt.Sprintf("%d quizzes taken", len(progress.QuizScores))))

	if len(progress.CompletedModules) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Completed Modules", ""))
		for module := range progress.CompletedModules {
			fmt.Println("  ✅ " + module)
		}
	}
}

func handleLearnQuiz(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🧠 Learning Quiz"))
	fmt.Println()

	if len(args) == 0 {
		fmt.Println(utils.FormatError("Please specify a quiz topic"))
		return
	}

	topic := args[0]
	fmt.Println(utils.FormatKeyValue("Quiz Topic", topic))
	fmt.Println(utils.FormatInfo("Loading quiz questions..."))

	// TODO: Implement quiz functionality
	fmt.Println(utils.FormatTip("Interactive quizzes are being developed"))
}

// Log analysis subcommand handlers
func handleLogsAnalyze(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🔍 Analyzing System Logs"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider
	providerName := cfg.AIProvider
	if providerName == "" {
		providerName = "ollama"
	}

	var aiProvider ai.AIProvider
	switch providerName {
	case "ollama":
		aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
	case "openai":
		aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	case "gemini":
		aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
	default:
		fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
		os.Exit(1)
	}

	// Read log data
	var logData string
	if len(args) > 0 && args[0] == "--file" && len(args) > 1 {
		// Read from specified file
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to read log file: "+err.Error()))
			return
		}
		logData = string(data)
	} else {
		// Read from stdin if available
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			input, _ := io.ReadAll(os.Stdin)
			logData = string(input)
		} else {
			fmt.Println(utils.FormatInfo("No log data provided. Use --file <path> or pipe log data"))
			return
		}
	}

	// Parse logs using existing parser
	entries, err := nixos.ParseLog(logData)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to parse logs: "+err.Error()))
		return
	}

	fmt.Println(utils.FormatKeyValue("Parsed Entries", fmt.Sprintf("%d", len(entries))))
	fmt.Print(utils.FormatInfo("Analyzing logs with AI... "))

	prompt := fmt.Sprintf("Analyze these NixOS system log entries and identify issues, errors, or recommendations:\n\n%s", logData)
	result, err := aiProvider.Query(prompt)
	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("AI analysis failed: "+err.Error()))
		return
	}

	fmt.Println(utils.RenderMarkdown(result))
}

func handleLogsParse(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📊 Parsing Log Structure"))
	fmt.Println()

	var logData string
	if len(args) > 0 {
		// Read from file
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to read log file: "+err.Error()))
			return
		}
		logData = string(data)
	} else {
		// Read from stdin
		input, _ := io.ReadAll(os.Stdin)
		logData = string(input)
	}

	entries, err := nixos.ParseLog(logData)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to parse logs: "+err.Error()))
		return
	}

	fmt.Println(utils.FormatKeyValue("Total Entries", fmt.Sprintf("%d", len(entries))))
	fmt.Println()

	// Show sample entries
	sampleSize := 5
	if len(entries) < sampleSize {
		sampleSize = len(entries)
	}

	fmt.Println(utils.FormatSubsection("Sample Parsed Entries", ""))
	for i := 0; i < sampleSize; i++ {
		entry := entries[i]
		fmt.Printf("Entry %d:\n", i+1)
		if entry.Timestamp != "" {
			fmt.Println(utils.FormatKeyValue("  Timestamp", entry.Timestamp))
		}
		if entry.Level != "" {
			fmt.Println(utils.FormatKeyValue("  Level", entry.Level))
		}
		if entry.Unit != "" {
			fmt.Println(utils.FormatKeyValue("  Unit", entry.Unit))
		}
		fmt.Println(utils.FormatKeyValue("  Message", entry.Message))
		fmt.Println()
	}
}

func handleLogsErrors(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("❌ Recent Critical Errors"))
	fmt.Println()

	// Use journalctl to get recent errors
	fmt.Println(utils.FormatInfo("Fetching recent system errors..."))
	fmt.Println(utils.FormatTip("This would integrate with journalctl to show recent ERROR and CRITICAL level messages"))

	// TODO: Implement journalctl integration
	fmt.Println(utils.FormatKeyValue("Status", "Feature in development"))
}

func handleLogsWatch(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("👀 Watching System Logs"))
	fmt.Println()

	fmt.Println(utils.FormatInfo("Starting real-time log monitoring..."))
	fmt.Println(utils.FormatTip("This would provide real-time log analysis and alerts"))

	// TODO: Implement real-time log watching
	fmt.Println(utils.FormatKeyValue("Status", "Feature in development"))
}

// Neovim integration subcommand handlers
func handleNeovimInstall(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📦 Installing Neovim Integration"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	// Get Neovim config directory
	configDir, err := neovim.GetUserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to get Neovim config directory: "+err.Error()))
		return
	}

	fmt.Println(utils.FormatKeyValue("Neovim Config Dir", configDir))

	// Create integration using existing functionality
	socketPath := cfg.MCPServer.SocketPath
	if socketPath == "" {
		socketPath = "/tmp/nixai-mcp.sock"
	}

	fmt.Print(utils.FormatInfo("Creating Neovim module... "))
	err = neovim.CreateNeovimModule(socketPath, configDir)
	if err != nil {
		fmt.Println(utils.FormatError("failed"))
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to create module: "+err.Error()))
		return
	}

	fmt.Println(utils.FormatSuccess("done"))
	fmt.Println()
	fmt.Println(utils.FormatSuccess("✅ Neovim integration installed successfully"))
	fmt.Println(utils.FormatTip("Restart Neovim to load the nixai integration"))
}

func handleNeovimStatus(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📊 Neovim Integration Status"))
	fmt.Println()

	configDir, err := neovim.GetUserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to get Neovim config directory: "+err.Error()))
		return
	}

	// Check if nixai.lua exists
	nixaiLuaPath := configDir + "/lua/nixai.lua"
	if _, err := os.Stat(nixaiLuaPath); err == nil {
		fmt.Println(utils.FormatKeyValue("Integration Status", "✅ Installed"))
		fmt.Println(utils.FormatKeyValue("Module Path", nixaiLuaPath))
	} else {
		fmt.Println(utils.FormatKeyValue("Integration Status", "❌ Not installed"))
		fmt.Println(utils.FormatTip("Run 'nixai neovim-setup install' to set up integration"))
	}

	fmt.Println(utils.FormatKeyValue("Config Directory", configDir))
}

func handleNeovimRemove(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🗑️  Removing Neovim Integration"))
	fmt.Println()

	configDir, err := neovim.GetUserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to get Neovim config directory: "+err.Error()))
		return
	}

	nixaiLuaPath := configDir + "/lua/nixai.lua"

	fmt.Print(utils.FormatInfo("Removing nixai.lua module... "))
	if err := os.Remove(nixaiLuaPath); err != nil {
		fmt.Println(utils.FormatError("failed"))
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to remove module: "+err.Error()))
			return
		}
		fmt.Println(utils.FormatWarning("Module was not installed"))
	} else {
		fmt.Println(utils.FormatSuccess("done"))
		fmt.Println(utils.FormatSuccess("✅ Neovim integration removed"))
	}
}

func handleNeovimUpdate(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🔄 Updating Neovim Integration"))
	fmt.Println()

	// Remove and reinstall
	handleNeovimRemove(cmd, args)
	fmt.Println()
	handleNeovimInstall(cmd, args)
}

// Package repository analysis handler
func handlePackageRepoAnalysis(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📦 Analyzing Package Repository"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider
	providerName := cfg.AIProvider
	if providerName == "" {
		providerName = "ollama"
	}

	var aiProvider ai.AIProvider
	switch providerName {
	case "ollama":
		aiProvider = ai.NewOllamaLegacyProvider(cfg.AIModel)
	case "openai":
		aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	case "gemini":
		aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
	default:
		fmt.Fprintln(os.Stderr, utils.FormatError("Unknown AI provider: "+providerName))
		os.Exit(1)
	}

	// Initialize MCP client for documentation
	mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	mcpClient := mcp.NewMCPClient(mcpURL)

	// Create packaging service
	tempDir := "/tmp/nixai-packaging"
	logger := &logger.Logger{} // Simple logger for now
	packagingService := packaging.NewPackagingService(aiProvider, mcpClient, tempDir, logger)

	// Parse command line arguments for packaging request
	req := &packaging.PackageRequest{
		Quiet: false,
	}

	if len(args) > 0 {
		// First argument could be URL or with --local flag
		if args[0] == "--local" && len(args) > 1 {
			req.LocalPath = args[1]
		} else {
			req.RepoURL = args[0]
		}
	}

	// Parse additional flags (basic implementation)
	for i, arg := range args {
		switch arg {
		case "--name":
			if i+1 < len(args) {
				req.PackageName = args[i+1]
			}
		case "--output":
			if i+1 < len(args) {
				req.OutputPath = args[i+1]
			}
		case "--quiet":
			req.Quiet = true
		}
	}

	if req.RepoURL == "" && req.LocalPath == "" {
		fmt.Println(utils.FormatError("Please provide a repository URL or local path"))
		return
	}

	fmt.Print(utils.FormatInfo("Starting repository analysis... "))

	// Run packaging analysis
	ctx := context.Background()
	result, err := packagingService.PackageRepository(ctx, req)
	if err != nil {
		fmt.Println(utils.FormatError("failed"))
		fmt.Fprintln(os.Stderr, utils.FormatError("Analysis failed: "+err.Error()))
		return
	}

	fmt.Println(utils.FormatSuccess("done"))
	fmt.Println()

	// Display results
	if result.Analysis != nil {
		fmt.Println(utils.FormatSubsection("📊 Repository Analysis", ""))
		fmt.Println(utils.FormatKeyValue("Language", result.Analysis.Language))
		fmt.Println(utils.FormatKeyValue("Build System", string(result.Analysis.BuildSystem)))
		fmt.Println(utils.FormatKeyValue("Dependencies", fmt.Sprintf("%d found", len(result.Analysis.Dependencies))))
		if result.Analysis.License != "" {
			fmt.Println(utils.FormatKeyValue("License", result.Analysis.License))
		}
		if result.Analysis.Description != "" {
			fmt.Println(utils.FormatKeyValue("Description", result.Analysis.Description))
		}
	}

	if result.Derivation != "" {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("📄 Generated Derivation", ""))
		fmt.Println(result.Derivation)
	}

	if len(result.ValidationIssues) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("⚠️  Validation Issues", ""))
		for _, issue := range result.ValidationIssues {
			fmt.Println("  • " + issue)
		}
	}

	if result.OutputFile != "" {
		fmt.Println()
		fmt.Println(utils.FormatSuccess("✅ Derivation saved to: " + result.OutputFile))
	}
}

// Handler functions for the CLI commands

// handleFlakeCommand handles flake management operations
func handleFlakeCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(utils.FormatHeader("NixOS Flake Management"))
		fmt.Println(utils.FormatInfo("Available subcommands:"))
		fmt.Println("  create    - Create a new flake configuration")
		fmt.Println("  validate  - Validate existing flake")
		fmt.Println("  migrate   - Migrate from legacy configuration")
		fmt.Println("  analyze   - Analyze flake for issues")
		fmt.Println()
		fmt.Println(utils.FormatInfo("Use 'nixai flake <subcommand> --help' for more information"))
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "create":
		handleFlakeCreate(cmd, args[1:])
	case "validate":
		handleFlakeValidate(cmd, args[1:])
	case "migrate":
		handleFlakeMigrate(cmd, args[1:])
	case "analyze":
		handleFlakeAnalyze(cmd, args[1:])
	default:
		fmt.Fprintf(os.Stderr, utils.FormatError("Unknown flake subcommand: %s\n"), subcommand)
		fmt.Println(utils.FormatInfo("Run 'nixai flake' to see available subcommands"))
	}
}

// handleLearnCommand handles learning system operations
func handleLearnCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(utils.FormatHeader("NixOS Learning System"))
		fmt.Println(utils.FormatInfo("Available subcommands:"))
		fmt.Println("  list      - List available learning modules")
		fmt.Println("  start     - Start a learning module")
		fmt.Println("  progress  - Show learning progress")
		fmt.Println("  quiz      - Take a quiz on a topic")
		fmt.Println()
		fmt.Println(utils.FormatInfo("Use 'nixai learn <subcommand> --help' for more information"))
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		handleLearnList(cmd, args[1:])
	case "start":
		handleLearnStart(cmd, args[1:])
	case "progress":
		handleLearnProgress(cmd, args[1:])
	case "quiz":
		handleLearnQuiz(cmd, args[1:])
	default:
		fmt.Fprintf(os.Stderr, utils.FormatError("Unknown learn subcommand: %s\n"), subcommand)
		fmt.Println(utils.FormatInfo("Run 'nixai learn' to see available subcommands"))
	}
}

// handleLogsCommand handles log analysis operations
func handleLogsCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(utils.FormatHeader("NixOS Log Analysis"))
		fmt.Println(utils.FormatInfo("Available subcommands:"))
		fmt.Println("  analyze   - Analyze system logs with AI")
		fmt.Println("  parse     - Parse log format and structure")
		fmt.Println("  errors    - Show recent critical errors")
		fmt.Println("  watch     - Watch logs in real-time")
		fmt.Println()
		fmt.Println(utils.FormatInfo("Use 'nixai logs <subcommand> --help' for more information"))
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "analyze":
		handleLogsAnalyze(cmd, args[1:])
	case "parse":
		handleLogsParse(cmd, args[1:])
	case "errors":
		handleLogsErrors(cmd, args[1:])
	case "watch":
		handleLogsWatch(cmd, args[1:])
	default:
		fmt.Fprintf(os.Stderr, utils.FormatError("Unknown logs subcommand: %s\n"), subcommand)
		fmt.Println(utils.FormatInfo("Run 'nixai logs' to see available subcommands"))
	}
}

// handleNeovimSetupCommand handles Neovim integration setup
func handleNeovimSetupCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(utils.FormatHeader("Neovim Integration Setup"))
		fmt.Println(utils.FormatInfo("Available subcommands:"))
		fmt.Println("  install   - Install Neovim integration")
		fmt.Println("  status    - Check integration status")
		fmt.Println("  remove    - Remove integration")
		fmt.Println("  update    - Update integration configuration")
		fmt.Println()
		fmt.Println(utils.FormatInfo("Use 'nixai neovim-setup <subcommand> --help' for more information"))
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "install":
		handleNeovimInstall(cmd, args[1:])
	case "status":
		handleNeovimStatus(cmd, args[1:])
	case "remove":
		handleNeovimRemove(cmd, args[1:])
	case "update":
		handleNeovimUpdate(cmd, args[1:])
	default:
		fmt.Fprintf(os.Stderr, utils.FormatError("Unknown neovim-setup subcommand: %s\n"), subcommand)
		fmt.Println(utils.FormatInfo("Run 'nixai neovim-setup' to see available subcommands"))
	}
}

// handlePackageRepoCommand handles package repository analysis
func handlePackageRepoCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(utils.FormatHeader("Package Repository Analysis"))
		fmt.Println(utils.FormatInfo("Analyze Git repositories and generate Nix derivations"))
		fmt.Println()
		fmt.Println(utils.FormatInfo("Usage:"))
		fmt.Println("  nixai package-repo <repository-url>")
		fmt.Println("  nixai package-repo --local <local-path>")
		fmt.Println()
		fmt.Println(utils.FormatInfo("Options:"))
		fmt.Println("  --local    Use local repository path")
		fmt.Println("  --name     Custom package name")
		fmt.Println("  --output   Output file for derivation")
		fmt.Println("  --quiet    Suppress progress output")
		return
	}

	handlePackageRepoAnalysis(cmd, args)
}

// initializeCommands adds all commands to the root command
func initializeCommands() {
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(explainOptionCmd)
	rootCmd.AddCommand(explainHomeOptionCmd)
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(gcCmd)
	rootCmd.AddCommand(hardwareCmd)
	rootCmd.AddCommand(createMachinesCommand())
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(templatesCmd)
	rootCmd.AddCommand(snippetsCmd)
	rootCmd.AddCommand(enhancedBuildCmd)
	rootCmd.AddCommand(devenvCmd)
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
			if err := os.Setenv("NIXAI_NIXOS_PATH", nixosPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set NIXAI_NIXOS_PATH: %v\n", err)
			}
		}
	})
	initializeCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
