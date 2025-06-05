package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
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
				aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&askQuestion, "ask", "a", "", "Ask a question about NixOS configuration")
	rootCmd.PersistentFlags().StringVarP(&nixosPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
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
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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
				aiProvider = ai.NewOllamaProvider(cfg.AIModel)
			case "openai":
				aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			case "gemini":
				aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "")
			default:
				aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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

// Stub commands to fix undefined errors. Replace with real implementations as needed.
var flakeCmd = &cobra.Command{Use: "flake", Short: "Stub flake command"}
var learnCmd = &cobra.Command{Use: "learn", Short: "Stub learn command"}
var logsCmd = &cobra.Command{Use: "logs", Short: "Stub logs command"}
var mcpServerCmd = &cobra.Command{Use: "mcp-server", Short: "Stub mcp-server command"}
var neovimSetupCmd = &cobra.Command{Use: "neovim-setup", Short: "Stub neovim-setup command"}
var packageRepoCmd = &cobra.Command{Use: "package-repo", Short: "Stub package-repo command"}

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
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
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
