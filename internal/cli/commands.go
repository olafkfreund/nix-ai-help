package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/agent"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/config"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check for global TUI flag and handle it for any command except interactive
		if globalTUI && cmd.Name() != "interactive" {
			// For non-interactive commands, launch TUI with the command pre-selected
			return LaunchTUIMode(cmd, append([]string{cmd.Name()}, args...))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for global TUI flag first
		if globalTUI {
			// If TUI mode is requested, launch the TUI with any provided args
			return LaunchTUIMode(cmd, args)
		}

		if askQuestion != "" {
			// Get current provider and model flag values directly from the command
			currentProvider, _ := cmd.PersistentFlags().GetString("provider")
			currentModel, _ := cmd.PersistentFlags().GetString("model")

			// DEBUG: Print what we found
			fmt.Printf("DEBUG: Root command - askQuestion='%s'\n", askQuestion)
			fmt.Printf("DEBUG: Root command - currentProvider='%s', aiProvider global='%s'\n", currentProvider, aiProvider)
			fmt.Printf("DEBUG: Root command - currentModel='%s', aiModel global='%s'\n", currentModel, aiModel)

			// If no provider specified in flags, fall back to global variables
			if currentProvider == "" {
				currentProvider = aiProvider
			}
			if currentModel == "" {
				currentModel = aiModel
			}

			// Set environment variables for provider and model flags so enhanced ask command can access them
			if currentProvider != "" {
				os.Setenv("NIXAI_PROVIDER", currentProvider)
			}
			if currentModel != "" {
				os.Setenv("NIXAI_MODEL", currentModel)
			}

			// Use the enhanced ask command implementation instead of simple version
			runAskCmd([]string{askQuestion}, os.Stdout)
			return nil
		}
		// If no --ask, show help
		return cmd.Help()
	},
}

var askQuestion string
var nixosPath string
var daemonMode bool
var agentRole string
var agentType string
var aiProvider string
var aiModel string
var contextFile string
var globalTUI bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&askQuestion, "ask", "a", "", "Ask a question about NixOS configuration")
	rootCmd.PersistentFlags().StringVarP(&nixosPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.PersistentFlags().StringVar(&agentRole, "role", "", "Specify the agent role (diagnoser, explainer, ask, build, flake, etc.)")
	rootCmd.PersistentFlags().StringVar(&agentType, "agent", "", "Specify the agent type (ask, build, diagnose, flake, etc.)")
	rootCmd.PersistentFlags().StringVar(&aiProvider, "provider", "", "Specify the AI provider (ollama, openai, gemini, etc.)")
	rootCmd.PersistentFlags().StringVar(&aiModel, "model", "", "Specify the AI model (llama3, gpt-4, gemini-1.5-pro, etc.)")
	rootCmd.PersistentFlags().StringVar(&contextFile, "context-file", "", "Path to a file containing context information (JSON or text)")
	rootCmd.PersistentFlags().BoolVar(&globalTUI, "tui", false, "Launch TUI mode for any command")
	mcpServerCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run MCP server in background/daemon mode")
	doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed output and progress information")

	// Add ask command flags
	askCmd.Flags().BoolP("quiet", "q", false, "Suppress validation output and show only the AI response")
	askCmd.Flags().BoolP("verbose", "v", false, "Show detailed validation output with multi-section layout")
	askCmd.Flags().BoolP("stream", "s", false, "Stream the response in real-time")

	// Add package-repo command flags
	packageRepoCmd.Flags().String("local", "", "Analyze local repository path instead of cloning")
	packageRepoCmd.Flags().String("output", "", "Output file path for generated derivation")
	packageRepoCmd.Flags().String("name", "", "Override package name for the derivation")
	packageRepoCmd.Flags().Bool("analyze-only", false, "Only analyze repository without generating derivation")

	// Add logs subcommands
	logsCmd.AddCommand(logsSystemCmd)
	logsCmd.AddCommand(logsBootCmd)
	logsCmd.AddCommand(logsServiceCmd)
	logsCmd.AddCommand(logsErrorsCmd)
	logsCmd.AddCommand(logsBuildCmd)
	logsCmd.AddCommand(logsAnalyzeCmd)
}

// Helper functions for agent/role/context handling
func loadContextFromFile(filepath string) (interface{}, error) {
	if filepath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}

	// Try to parse as JSON first
	var jsonContext interface{}
	if err := json.Unmarshal(data, &jsonContext); err == nil {
		return jsonContext, nil
	}

	// If not valid JSON, return as string
	return string(data), nil
}

func createAgentFromFlags(provider ai.Provider) (agent.Agent, error) {
	// If no agent type specified, determine from role or use default
	if agentType == "" {
		// Try to infer agent type from role
		if agentRole != "" {
			switch strings.ToLower(agentRole) {
			case "ask":
				return agent.NewAskAgent(provider), nil
			case "flake":
				return agent.NewFlakeAgent(provider), nil
			case "build":
				return agent.NewBuildAgent(provider), nil
			case "explain-option":
				return agent.NewExplainOptionAgent(provider, nil), nil
			case "explain-home-option":
				return agent.NewExplainHomeOptionAgent(provider, nil), nil
			case "configure":
				return agent.NewConfigureAgent(provider), nil
			case "hardware":
				return agent.NewHardwareAgent(provider), nil
			case "gc":
				return agent.NewGCAgent(provider), nil
			case "interactive":
				return agent.NewInteractiveAgent(provider), nil
			case "learn":
				return agent.NewLearnAgent(provider), nil
			case "migrate":
				return agent.NewMigrateAgent(provider), nil
			case "devenv":
				return agent.NewDevenvAgent(provider), nil
			case "package-repo":
				return agent.NewPackageRepoAgent(provider), nil
			case "templates":
				return agent.NewTemplatesAgent(provider), nil
			case "help":
				return agent.NewHelpAgent(provider), nil
			// These agents have interface compatibility issues, default to ask agent
			case "diagnose", "community", "machines", "store", "logs", "mcp-server", "neovim-setup", "snippets":
				return agent.NewAskAgent(provider), nil
			default:
				return agent.NewAskAgent(provider), nil // Default fallback
			}
		}
		// Default to ask agent if no role specified
		return agent.NewAskAgent(provider), nil
	}

	// Create agent based on explicit agent type
	switch strings.ToLower(agentType) {
	case "ask":
		return agent.NewAskAgent(provider), nil
	case "flake":
		return agent.NewFlakeAgent(provider), nil
	case "build":
		return agent.NewBuildAgent(provider), nil
	case "explain-option":
		return agent.NewExplainOptionAgent(provider, nil), nil
	case "explain-home-option":
		return agent.NewExplainHomeOptionAgent(provider, nil), nil
	case "configure":
		return agent.NewConfigureAgent(provider), nil
	case "hardware":
		return agent.NewHardwareAgent(provider), nil
	case "gc":
		return agent.NewGCAgent(provider), nil
	case "interactive":
		return agent.NewInteractiveAgent(provider), nil
	case "learn":
		return agent.NewLearnAgent(provider), nil
	case "migrate":
		return agent.NewMigrateAgent(provider), nil
	case "devenv":
		return agent.NewDevenvAgent(provider), nil
	case "package-repo":
		return agent.NewPackageRepoAgent(provider), nil
	case "templates":
		return agent.NewTemplatesAgent(provider), nil
	case "help":
		return agent.NewHelpAgent(provider), nil
	// These agents have interface compatibility issues, default to ask agent
	case "diagnose", "community", "machines", "store", "logs", "mcp-server", "neovim-setup", "snippets":
		return agent.NewAskAgent(provider), nil
	case "ollama", "openai", "gemini", "llamacpp", "custom":
		// These are provider names, not agent types - default to ask agent
		return agent.NewAskAgent(provider), nil
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

func validateAndSetRole(agentInstance agent.Agent) error {
	if agentRole == "" {
		return nil // No role specified, use default
	}

	// Validate role
	if !roles.ValidateRole(agentRole) {
		return fmt.Errorf("invalid role: %s", agentRole)
	}

	// Set role on agent
	return agentInstance.SetRole(roles.RoleType(agentRole))
}

func setAgentContext(agentInstance agent.Agent) error {
	contextData, err := loadContextFromFile(contextFile)
	if err != nil {
		return err
	}

	if contextData != nil {
		agentInstance.SetContext(contextData)
	}

	return nil
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
		// Validate provider using model registry
		registry := config.NewModelRegistry(cfg)
		availableProviders := registry.GetAvailableProviders()
		isValid := false
		for _, provider := range availableProviders {
			if value == provider {
				isValid = true
				break
			}
		}
		if !isValid {
			validOptions := strings.Join(availableProviders, ", ")
			fmt.Println(utils.FormatError("Invalid AI provider. Valid options: " + validOptions))
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
	Args:  conditionalArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}

		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err != nil {
			fmt.Println(utils.FormatWarning("Context detection failed: " + err.Error()))
			nixosCtx = nil
		}

		// Display detected context summary if available
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Println(utils.FormatNote("📋 " + contextSummary))
			fmt.Println()
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
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			os.Exit(1)
		}

		// Get provider name for context
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
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

		// Build context-aware prompt
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		basePrompt := fmt.Sprintf("How to search, install, and configure %s in NixOS?", query)
		contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt, nixosCtx)

		promptCtx := ai.PromptContext{
			Question:     contextualPrompt,
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
	Args:  conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		option := args[0]
		fmt.Println(utils.FormatHeader("🏠 Home Manager Option: " + option))
		fmt.Println()

		// Load configuration first
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}

		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err != nil {
			fmt.Println(utils.FormatWarning("Context detection failed: " + err.Error()))
			nixosCtx = nil
		}

		// Display detected context summary if available
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Println(utils.FormatNote("📋 " + contextSummary))
			fmt.Println()
		}

		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			os.Exit(1)
		}

		// Get provider name for context
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
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
		basePrompt, err := builder.BuildPrompt(promptCtx)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Prompt build error: "+err.Error()))
			os.Exit(1)
		}

		// Build context-aware prompt using the context builder
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt, nixosCtx)

		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		aiResp, aiErr := aiProvider.Query(contextualPrompt)
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
		Args:  conditionalExactArgsValidator(1),
		Run: func(cmd *cobra.Command, args []string) {
			option := args[0]
			format, _ := cmd.Flags().GetString("format")
			providerFlag, _ := cmd.Flags().GetString("provider")
			examplesOnly, _ := cmd.Flags().GetBool("examples-only")

			// Load configuration first
			cfg, err := config.LoadUserConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
				os.Exit(1)
			}

			// Initialize context detector and get NixOS context
			contextDetector := nixos.NewContextDetector(logger.NewLogger())
			nixosCtx, err := contextDetector.GetContext(cfg)
			if err != nil {
				fmt.Println(utils.FormatWarning("Context detection failed: " + err.Error()))
				nixosCtx = nil
			}

			// Display detected context summary if available
			if nixosCtx != nil && nixosCtx.CacheValid {
				contextBuilder := nixoscontext.NewNixOSContextBuilder()
				contextSummary := contextBuilder.GetContextSummary(nixosCtx)
				fmt.Println(utils.FormatNote("📋 " + contextSummary))
				fmt.Println()
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

			// Create a temporary config with the selected provider
			tempCfg := *cfg
			tempCfg.AIProvider = aiProviderName

			aiProvider, err := GetLegacyAIProvider(&tempCfg, logger.NewLogger())
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
				os.Exit(1)
			}

			// Build context-aware prompt using the context builder
			var basePrompt string
			if examplesOnly {
				basePrompt = buildExamplesOnlyPrompt(option, doc, format, source, version)
			} else {
				basePrompt = buildEnhancedExplainOptionPrompt(option, doc, format, source, version)
			}
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt, nixosCtx)

			fmt.Print(utils.FormatInfo("Querying AI provider... "))
			aiResp, aiErr := aiProvider.Query(contextualPrompt)
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
- Modern TUI interface with two-panel layout
- Live command search and filtering
- Parameter input for commands that need it
- Real-time command execution
- All advanced features available in non-interactive mode

Examples:
  nixai interactive              # Start modern TUI interface (default)
  nixai interactive --classic    # Start classic interactive mode
`,
	Run: func(cmd *cobra.Command, args []string) {
		useClassic, _ := cmd.Flags().GetBool("classic")
		if useClassic {
			InteractiveMode()
		} else {
			InteractiveModeTUI()
		}
	},
}

func init() {
	// Add the --classic flag to the interactive command (TUI is now default)
	interactiveCmd.Flags().Bool("classic", false, "Launch classic interactive mode instead of modern TUI")
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

// Logs subcommands
var logsSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Analyze system logs",
	Long:  "Analyze system logs for issues, patterns, and recommendations.",
	Run:   handleLogsSystem,
}

var logsBootCmd = &cobra.Command{
	Use:   "boot",
	Short: "Analyze boot logs",
	Long:  "Analyze boot logs for startup issues, errors, and performance insights.",
	Run:   handleLogsBoot,
}

var logsServiceCmd = &cobra.Command{
	Use:   "service [service-name]",
	Short: "Analyze service logs",
	Long:  "Analyze service-specific logs for issues, errors, and troubleshooting recommendations.",
	Run:   handleLogsService,
}

var logsErrorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Analyze error logs",
	Long:  "Analyze system error logs and provide troubleshooting recommendations.",
	Run:   handleLogsErrors,
}

var logsBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Analyze build logs",
	Long:  "Analyze build logs for compilation errors, dependency issues, and optimization suggestions.",
	Run:   handleLogsBuild,
}

var logsAnalyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze specific log file",
	Long:  "Analyze a specific log file with AI-powered diagnostics.",
	Run:   handleLogsAnalyze,
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
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var neovimSetupInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Neovim integration with nixai",
	Long: `Install Neovim integration by creating the necessary configuration files
and Lua modules to connect Neovim with the nixai MCP server.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleNeovimSetupInstall(cmd, args)
	},
}

var neovimSetupConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Neovim integration settings",
	Long: `Configure Neovim integration settings such as socket path,
key bindings, and other integration options.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleNeovimSetupConfigure(cmd, args)
	},
}

var neovimSetupStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Neovim integration status",
	Long: `Check the status of Neovim integration, including whether
configuration files exist and the MCP server is accessible.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleNeovimSetupStatus(cmd, args)
	},
}

var neovimSetupUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Neovim integration configuration",
	Long: `Update the Neovim integration configuration files to the
latest version and apply any new features or bug fixes.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleNeovimSetupUpdate(cmd, args)
	},
}

var neovimSetupRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove Neovim integration",
	Long: `Remove the Neovim integration by deleting configuration files
and modules that were created for nixai integration.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleNeovimSetupRemove(cmd, args)
	},
}

func init() {
	// Add subcommands to neovim-setup
	neovimSetupCmd.AddCommand(neovimSetupInstallCmd)
	neovimSetupCmd.AddCommand(neovimSetupConfigureCmd)
	neovimSetupCmd.AddCommand(neovimSetupStatusCmd)
	neovimSetupCmd.AddCommand(neovimSetupUpdateCmd)
	neovimSetupCmd.AddCommand(neovimSetupRemoveCmd)

	// Add flags to install and configure commands
	neovimSetupInstallCmd.Flags().String("config-dir", "", "Neovim configuration directory (default: auto-detect)")
	neovimSetupInstallCmd.Flags().String("socket-path", "/tmp/nixai-mcp.sock", "MCP server socket path")

	neovimSetupConfigureCmd.Flags().String("config-dir", "", "Neovim configuration directory (default: auto-detect)")
	neovimSetupConfigureCmd.Flags().String("socket-path", "/tmp/nixai-mcp.sock", "MCP server socket path")

	neovimSetupUpdateCmd.Flags().String("config-dir", "", "Neovim configuration directory (default: auto-detect)")
	neovimSetupUpdateCmd.Flags().String("socket-path", "/tmp/nixai-mcp.sock", "MCP server socket path")
}

// NewNeovimSetupCmd creates a new neovim-setup command for TUI mode
func NewNeovimSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     neovimSetupCmd.Use,
		Short:   neovimSetupCmd.Short,
		Long:    neovimSetupCmd.Long,
		Example: neovimSetupCmd.Example,
		Run:     neovimSetupCmd.Run,
	}

	// Add subcommands
	cmd.AddCommand(neovimSetupInstallCmd)
	cmd.AddCommand(neovimSetupConfigureCmd)
	cmd.AddCommand(neovimSetupStatusCmd)
	cmd.AddCommand(neovimSetupUpdateCmd)
	cmd.AddCommand(neovimSetupRemoveCmd)

	// Copy flags from existing commands
	cmd.PersistentFlags().AddFlagSet(neovimSetupCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(neovimSetupCmd.Flags())

	return cmd
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

// conditionalArgsValidator returns a validator that checks if TUI mode is requested
// and bypasses argument validation if so
func conditionalArgsValidator(minArgs int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		// If TUI mode is requested, don't validate args
		if globalTUI {
			return nil
		}
		// Otherwise, apply the minimum args validation
		return cobra.MinimumNArgs(minArgs)(cmd, args)
	}
}

// conditionalExactArgsValidator returns a validator that checks if TUI mode is requested
// and bypasses exact argument validation if so
func conditionalExactArgsValidator(exactArgs int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		// If TUI mode is requested, don't validate args
		if globalTUI {
			return nil
		}
		// Otherwise, apply the exact args validation
		return cobra.ExactArgs(exactArgs)(cmd, args)
	}
}

// conditionalRangeArgsValidator returns a validator that checks if TUI mode is requested
// and bypasses range argument validation if so
func conditionalRangeArgsValidator(min, max int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		// If TUI mode is requested, don't validate args
		if globalTUI {
			return nil
		}
		// Otherwise, apply the range args validation
		return cobra.RangeArgs(min, max)(cmd, args)
	}
}

// conditionalMaximumArgsValidator returns a validator that checks if TUI mode is requested
// and bypasses maximum argument validation if so
func conditionalMaximumArgsValidator(maxArgs int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		// If TUI mode is requested, don't validate args
		if globalTUI {
			return nil
		}
		// Otherwise, apply the maximum args validation
		return cobra.MaximumNArgs(maxArgs)(cmd, args)
	}
}

// Ask command - Now uses the enhanced implementation
var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask a question about NixOS configuration",
	Long: `Ask a direct question about NixOS configuration and get an AI-powered answer with comprehensive multi-source validation.

This command queries multiple information sources:
- Official NixOS documentation via MCP server
- Verified package search results
- Real-world GitHub configuration examples
- Response validation for common syntax errors

Output modes:
- Default: Concise progress indicators with footer-style summary
- --quiet: Show only the AI response without any validation output
- --verbose: Show detailed validation output with multi-section layout
- --stream: Stream the response in real-time (great for LlamaCpp with Vulkan support)

Examples:
  nixai ask "How do I configure nginx?"
  nixai ask "What is the difference between services.openssh.enable and programs.ssh.enable?"
  nixai ask "How do I set up a development environment with Python?" --provider gemini
  nixai ask "How do I enable SSH?" --quiet
  nixai ask "How do I enable nginx?" --verbose
  nixai ask "Help me troubleshoot my build" --stream`,
	Args: conditionalArgsValidator(1), Run: func(cmd *cobra.Command, args []string) {
		// Get the quiet, verbose, and stream flag values
		quiet, _ := cmd.Flags().GetBool("quiet")
		verbose, _ := cmd.Flags().GetBool("verbose")
		stream, _ := cmd.Flags().GetBool("stream")

		// Get current provider and model flag values - check both command and persistent flags
		currentProvider, _ := cmd.Root().PersistentFlags().GetString("provider")
		currentModel, _ := cmd.Root().PersistentFlags().GetString("model")

		// If no provider specified, fall back to global variables
		if currentProvider == "" {
			currentProvider = aiProvider
		}
		if currentModel == "" {
			currentModel = aiModel
		}

		// Route to appropriate version based on flags
		if stream {
			runAskCmdWithStreaming(args, cmd.OutOrStdout(), currentProvider, currentModel)
		} else if quiet {
			runAskCmdWithOptionsQuiet(args, cmd.OutOrStdout(), currentProvider, currentModel)
		} else if verbose {
			runAskCmdWithOptions(args, cmd.OutOrStdout(), currentProvider, currentModel)
		} else {
			// Default to concise mode for better user experience
			runAskCmdWithConciseMode(args, cmd.OutOrStdout(), currentProvider, currentModel)
		}
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
  nixai configure --search "web server nginx"
  nixai configure --output my-config.nix
  nixai configure --advanced --home --output home-config.nix
  nixai configure --search "desktop" --advanced --output desktop-config.nix
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🛠️  Interactive NixOS Configuration"))
		fmt.Println()

		// Get flag values
		searchQuery, _ := cmd.Flags().GetString("search")
		outputFile, _ := cmd.Flags().GetString("output")
		isAdvanced, _ := cmd.Flags().GetBool("advanced")
		isHome, _ := cmd.Flags().GetBool("home")

		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}

		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err != nil {
			fmt.Println(utils.FormatWarning("Failed to detect NixOS context: " + err.Error()))
			nixosCtx = nil
		}

		// Display detected context summary
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Println(utils.FormatNote("📋 " + contextSummary))
			fmt.Println()
		}

		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			os.Exit(1)
		}

		// Get provider name for context
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}

		var input string
		if searchQuery != "" {
			input = searchQuery
			fmt.Println(utils.FormatInfo("Using search query: " + searchQuery))
		} else {
			configType := "NixOS"
			if isHome {
				configType = "Home Manager"
			}
			fmt.Printf(utils.FormatInfo("Describe what you want to configure for %s (e.g. desktop, web server, development environment):\n"), configType)
			fmt.Print("> ")
			_, _ = fmt.Scanln(&input)
			if input == "" {
				fmt.Println(utils.FormatWarning("No input provided. Exiting."))
				return
			}
		}

		// Build the prompt based on configuration type and advanced options
		prompt := buildConfigurePrompt(input, isHome, isAdvanced)

		// Enhance prompt with context-aware information
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextualPrompt := contextBuilder.BuildContextualPrompt(prompt, nixosCtx)
			prompt = contextualPrompt
		}

		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}

		// Display or save the output
		if outputFile != "" {
			err := saveConfigurationToFile(resp, outputFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to save to file: "+err.Error()))
				os.Exit(1)
			}
			fmt.Println(utils.FormatSuccess("✅ Configuration saved to: " + outputFile))
			fmt.Println(utils.FormatTip("Review the generated configuration and customize as needed"))
		} else {
			fmt.Println(utils.RenderMarkdown(resp))
		}
	},
}

// buildConfigurePrompt builds an AI prompt for configuration generation
func buildConfigurePrompt(input string, isHome bool, isAdvanced bool) string {
	configType := "NixOS"
	if isHome {
		configType = "Home Manager"
	}

	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("You are an expert %s configuration assistant. ", configType))
	prompt.WriteString(fmt.Sprintf("Generate a complete, production-ready %s configuration based on the following request:\n\n", configType))
	prompt.WriteString(fmt.Sprintf("Request: %s\n\n", input))

	if isHome {
		prompt.WriteString("Generate Home Manager configuration that includes:\n")
		prompt.WriteString("- Appropriate program configurations\n")
		prompt.WriteString("- Service configurations if needed\n")
		prompt.WriteString("- Package installations\n")
		prompt.WriteString("- Dotfile management where relevant\n\n")
	} else {
		prompt.WriteString("Generate NixOS configuration that includes:\n")
		prompt.WriteString("- System-level service configurations\n")
		prompt.WriteString("- Hardware enablement where needed\n")
		prompt.WriteString("- Security and networking settings\n")
		prompt.WriteString("- Package installations\n")
		prompt.WriteString("- User and group configurations where relevant\n\n")
	}

	if isAdvanced {
		prompt.WriteString("Use advanced configuration options including:\n")
		prompt.WriteString("- Detailed service configurations with all relevant options\n")
		prompt.WriteString("- Security hardening configurations\n")
		prompt.WriteString("- Performance optimizations\n")
		prompt.WriteString("- Advanced networking and hardware configurations\n")
		prompt.WriteString("- Modular configuration structure\n")
		prompt.WriteString("- Comprehensive documentation and comments\n\n")
	}

	prompt.WriteString("Requirements:\n")
	prompt.WriteString("- Provide complete, syntactically correct Nix configuration\n")
	prompt.WriteString("- Include helpful comments explaining each section\n")
	prompt.WriteString("- Use best practices and idiomatic Nix expressions\n")
	prompt.WriteString("- Ensure compatibility with current NixOS/Home Manager versions\n")
	prompt.WriteString("- Include error handling and fallbacks where appropriate\n")

	if isAdvanced {
		prompt.WriteString("- Provide detailed explanations for advanced configurations\n")
		prompt.WriteString("- Include alternative configuration options as comments\n")
		prompt.WriteString("- Add troubleshooting notes where relevant\n")
	}

	return prompt.String()
}

// saveConfigurationToFile saves the generated configuration to a file
func saveConfigurationToFile(content, filename string) error {
	// Clean the content to extract just the configuration
	lines := strings.Split(content, "\n")
	var configLines []string
	inCodeBlock := false

	for _, line := range lines {
		// Look for code blocks
		if strings.HasPrefix(line, "```nix") || strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Include lines that are inside code blocks or look like Nix configuration
		if inCodeBlock || strings.Contains(line, "{") || strings.Contains(line, "}") ||
			strings.Contains(line, "=") || strings.HasPrefix(strings.TrimSpace(line), "#") ||
			strings.Contains(line, "enable") || strings.Contains(line, "programs.") ||
			strings.Contains(line, "services.") || strings.Contains(line, "environment.") {
			configLines = append(configLines, line)
		}
	}

	// If we didn't find a proper code block, save the original content
	if len(configLines) == 0 {
		configLines = lines
	}

	finalContent := strings.Join(configLines, "\n")

	// Ensure the file has a .nix extension
	if !strings.HasSuffix(filename, ".nix") {
		filename += ".nix"
	}

	return os.WriteFile(filename, []byte(finalContent), 0644)
}

func init() {
	// Add flags for the configure command
	configureCmd.Flags().StringP("search", "s", "", "Search query for configuration type (e.g., 'web server nginx', 'desktop')")
	configureCmd.Flags().StringP("output", "o", "", "Output file path for generated configuration (will add .nix extension)")
	configureCmd.Flags().Bool("advanced", false, "Generate advanced configuration with detailed options and optimizations")
	configureCmd.Flags().Bool("home", false, "Generate Home Manager configuration instead of NixOS system configuration")
}

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose [logfile]",
	Short: "Diagnose NixOS issues from logs or config",
	Long: `Diagnose NixOS issues by analyzing logs, configuration files, or piped input. Uses AI and documentation to suggest fixes.

Examples:
  nixai diagnose /var/log/messages
  journalctl -xe | nixai diagnose
  nixai diagnose --file /var/log/nixos-rebuild.log
  nixai diagnose --type system
  nixai diagnose --context "build failed with dependency error"
`,
	Args: conditionalMaximumArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🩺 NixOS Diagnostics"))
		fmt.Println()

		// Parse command flags
		inputFile, _ := cmd.Flags().GetString("file")
		diagType, _ := cmd.Flags().GetString("type")
		outputFormat, _ := cmd.Flags().GetString("output")
		additionalContext, _ := cmd.Flags().GetString("context")

		// Load configuration first
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
			os.Exit(1)
		}

		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err != nil {
			fmt.Println(utils.FormatWarning("Context detection failed: " + err.Error()))
			nixosCtx = nil
		}

		// Display detected context summary if available
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Println(utils.FormatNote("📋 " + contextSummary))
			fmt.Println()
		}

		var logData string

		// Determine input source based on flags and arguments
		if inputFile != "" {
			// Use --file flag
			data, err := os.ReadFile(inputFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to read file: "+err.Error()))
				os.Exit(1)
			}
			logData = string(data)
		} else if len(args) > 0 {
			// Use positional argument
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
				// No input provided, offer diagnostic options based on type flag
				if diagType != "" {
					fmt.Printf("Running %s diagnostics...\n", diagType)
					logData = fmt.Sprintf("Perform %s diagnostics for NixOS system", diagType)
				} else {
					fmt.Println(utils.FormatWarning("No log file, piped input, or diagnostic type provided."))
					fmt.Println(utils.FormatTip("Usage: nixai diagnose [logfile] or nixai diagnose --type system"))
					return
				}
			}
		}

		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			os.Exit(1)
		}

		// Build context-aware prompt using the context builder
		basePrompt := "You are a NixOS expert. Analyze the following log or error output and provide a diagnosis, root cause, and step-by-step fix instructions.\n\n"

		if diagType != "" {
			basePrompt += fmt.Sprintf("Focus on %s-related issues. ", diagType)
		}

		if additionalContext != "" {
			basePrompt += fmt.Sprintf("Additional context: %s\n\n", additionalContext)
		}

		basePrompt += "Log or error:\n" + logData

		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt, nixosCtx)

		fmt.Print(utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(contextualPrompt)
		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("AI error: "+err.Error()))
			os.Exit(1)
		}

		// Format output based on output format flag
		switch outputFormat {
		case "plain":
			fmt.Println(resp)
		case "json":
			// Simple JSON wrapper
			fmt.Printf(`{"diagnosis": %q}`, resp)
			fmt.Println()
		default: // markdown
			fmt.Println(utils.RenderMarkdown(resp))
		}
	},
}

func init() {
	// Add flags to diagnose command
	diagnoseCmd.Flags().StringP("file", "f", "", "Specify log file path to analyze")
	diagnoseCmd.Flags().StringP("type", "t", "", "Diagnostic type (system, config, services, network, hardware, performance)")
	diagnoseCmd.Flags().StringP("output", "o", "markdown", "Output format (markdown, plain, json)")
	diagnoseCmd.Flags().StringP("context", "c", "", "Additional context information to include in analysis")
}

var doctorCmd = &cobra.Command{
	Use:   "doctor [check_type]",
	Short: "Run comprehensive NixOS health checks and diagnostics",
	Long: `Run comprehensive NixOS health checks and get AI-powered diagnostics and recommendations.

Supports multiple check types:
  system      - Core system health checks
  nixos       - NixOS-specific configuration checks  
  packages    - Package and store integrity checks
  services    - System service status checks
  storage     - Storage and filesystem checks
  network     - Network connectivity checks
  security    - Security configuration checks
  all         - Run all available checks (default)

Examples:
  nixai doctor               # Run all health checks
  nixai doctor system        # Run only system checks
  nixai doctor packages      # Check package integrity
  nixai doctor --verbose     # Detailed output
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDoctorCommand(cmd, args)
	},
}

// runDoctorCommand executes the comprehensive doctor health checks
func runDoctorCommand(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🩻 NixOS Doctor: Comprehensive Health Check"))
	fmt.Println()

	// Load configuration first
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		fmt.Println(utils.FormatWarning("Context detection failed: " + err.Error()))
		nixosCtx = nil
	}

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		fmt.Println(utils.FormatNote("📋 " + contextSummary))
		fmt.Println()
	}

	// Determine check type
	checkType := "all"
	if len(args) > 0 {
		checkType = args[0]
	}

	// Get verbose flag
	verbose, _ := cmd.Flags().GetBool("verbose")

	fmt.Println(utils.FormatInfo("🔍 Performing health checks..."))
	fmt.Println()

	// Show what checks are being performed
	showChecksBeingPerformed(checkType, verbose)

	// Initialize AI provider for analysis
	aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		os.Exit(1)
	}

	// Perform actual health checks
	healthResults := performHealthChecks(checkType, cfg, verbose)

	// Display results
	displayHealthResults(healthResults, verbose)

	// Get AI analysis if provider is available
	if aiProvider != nil {
		fmt.Println()
		fmt.Println(utils.FormatHeader("🤖 AI-Powered Analysis"))
		fmt.Print(utils.FormatInfo("Analyzing results with AI... "))

		// Build context-aware prompt using the context builder
		baseAnalysisPrompt := buildAnalysisPrompt(healthResults, checkType)
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextualPrompt := contextBuilder.BuildContextualPrompt(baseAnalysisPrompt, nixosCtx)

		analysis, err := aiProvider.Query(contextualPrompt)

		fmt.Println(utils.FormatSuccess("done"))
		if err != nil {
			fmt.Println(utils.FormatWarning("AI analysis unavailable: " + err.Error()))
		} else {
			fmt.Println()
			fmt.Println(utils.RenderMarkdown(analysis))
		}
	}
}

// showChecksBeingPerformed displays what checks are being performed
func showChecksBeingPerformed(checkType string, verbose bool) {
	checkTypes := getCheckTypes(checkType)

	fmt.Println(utils.FormatSubsection("Health Check Categories", ""))
	for _, ct := range checkTypes {
		switch ct {
		case "system":
			fmt.Println("  🖥️  System Health - Core system components and boot status")
		case "nixos":
			fmt.Println("  🐧 NixOS Configuration - Config syntax and rebuild status")
		case "packages":
			fmt.Println("  📦 Package Integrity - Nix store and package health")
		case "services":
			fmt.Println("  🔧 System Services - Service status and failed units")
		case "storage":
			fmt.Println("  💾 Storage Health - Filesystem and disk usage")
		case "network":
			fmt.Println("  🌐 Network Status - Connectivity and DNS resolution")
		case "security":
			fmt.Println("  🔒 Security Audit - Permissions and security settings")
		}
	}
	fmt.Println()
}

// getCheckTypes returns the list of check types to perform
func getCheckTypes(checkType string) []string {
	if checkType == "all" {
		return []string{"system", "nixos", "packages", "services", "storage", "network", "security"}
	}
	return []string{checkType}
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Category    string
	Name        string
	Status      string // "pass", "warn", "fail", "info"
	Description string
	Details     string
	Command     string // Optional command suggestion
}

// performHealthChecks executes the actual health checks
func performHealthChecks(checkType string, cfg *config.UserConfig, verbose bool) []HealthCheckResult {
	var results []HealthCheckResult
	checkTypes := getCheckTypes(checkType)

	// Determine config path
	configPath := cfg.NixosFolder
	if nixosPath != "" {
		configPath = nixosPath
	}
	if configPath == "" {
		configPath = "/etc/nixos"
	}

	for _, ct := range checkTypes {
		fmt.Print(utils.FormatProgress("  Checking " + ct + "... "))

		switch ct {
		case "system":
			results = append(results, performSystemChecks(configPath, verbose)...)
		case "nixos":
			results = append(results, performNixOSChecks(configPath, verbose)...)
		case "packages":
			results = append(results, performPackageChecks(verbose)...)
		case "services":
			results = append(results, performServiceChecks(verbose)...)
		case "storage":
			results = append(results, performStorageChecks(verbose)...)
		case "network":
			results = append(results, performNetworkChecks(verbose)...)
		case "security":
			results = append(results, performSecurityChecks(verbose)...)
		}

		fmt.Println(utils.FormatSuccess("done"))
	}

	return results
}

// performSystemChecks checks core system health
func performSystemChecks(configPath string, verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	// Check if NixOS is running
	if _, err := os.Stat("/run/current-system"); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "system",
			Name:        "NixOS System",
			Status:      "pass",
			Description: "NixOS system is properly initialized",
			Details:     "Current system generation exists",
		})
	} else {
		results = append(results, HealthCheckResult{
			Category:    "system",
			Name:        "NixOS System",
			Status:      "fail",
			Description: "NixOS system may not be properly initialized",
			Details:     "/run/current-system not found",
			Command:     "sudo nixos-rebuild switch",
		})
	}

	// Check boot loader with comprehensive EFI and legacy support
	bootLoaderDetected := false
	bootLoaderDetails := []string{}
	permissionIssues := false

	// Check if this is an EFI system
	isEFISystem := false
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		isEFISystem = true
		bootLoaderDetails = append(bootLoaderDetails, "EFI system detected")
	}

	// Try to use bootctl to get boot loader information if available
	if isEFISystem {
		if output, err := exec.Command("bootctl", "status").CombinedOutput(); err == nil {
			outputStr := string(output)
			bootLoaderDetails = append(bootLoaderDetails, "bootctl command available")

			// Parse bootctl output for boot loader type
			if strings.Contains(outputStr, "systemd-boot") {
				bootLoaderDetected = true
				bootLoaderDetails = append(bootLoaderDetails, "systemd-boot detected via bootctl")
			} else if strings.Contains(outputStr, "GRUB") {
				bootLoaderDetected = true
				bootLoaderDetails = append(bootLoaderDetails, "GRUB detected via bootctl")
			}
		}
	}

	// Check for systemd-boot (EFI) via file system
	if _, err := os.Stat("/boot/loader/loader.conf"); err == nil {
		bootLoaderDetected = true
		bootLoaderDetails = append(bootLoaderDetails, "systemd-boot configuration found")

		// Check for boot entries
		if _, err := os.Stat("/boot/loader/entries"); err == nil {
			bootLoaderDetails = append(bootLoaderDetails, "boot entries directory exists")
		}
	} else if os.IsPermission(err) {
		permissionIssues = true
		bootLoaderDetails = append(bootLoaderDetails, "permission denied accessing /boot/loader")
	}

	// Check for GRUB (both EFI and legacy)
	grubDetected := false
	if _, err := os.Stat("/boot/grub"); err == nil {
		grubDetected = true
		bootLoaderDetected = true
		bootLoaderDetails = append(bootLoaderDetails, "GRUB directory found")
	} else if os.IsPermission(err) {
		permissionIssues = true
		bootLoaderDetails = append(bootLoaderDetails, "permission denied accessing /boot/grub")
	}

	// Check for GRUB EFI installation
	if isEFISystem {
		if _, err := os.Stat("/boot/EFI"); err == nil {
			bootLoaderDetails = append(bootLoaderDetails, "EFI boot directory exists")

			// Check for various EFI boot loaders
			efiDirs := []string{"nixos", "systemd", "BOOT", "Linux"}
			for _, dir := range efiDirs {
				if _, err := os.Stat("/boot/EFI/" + dir); err == nil {
					bootLoaderDetails = append(bootLoaderDetails, fmt.Sprintf("EFI/%s directory found", dir))
					if dir == "nixos" || dir == "systemd" {
						bootLoaderDetected = true
					}
				}
			}
		} else if os.IsPermission(err) {
			permissionIssues = true
			bootLoaderDetails = append(bootLoaderDetails, "permission denied accessing /boot/EFI")
		}
	}

	// Check via efibootmgr if available and EFI system
	if isEFISystem && !bootLoaderDetected {
		if output, err := exec.Command("efibootmgr").CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "nixos") || strings.Contains(outputStr, "systemd-boot") || strings.Contains(outputStr, "GRUB") {
				bootLoaderDetected = true
				bootLoaderDetails = append(bootLoaderDetails, "boot entries found via efibootmgr")
			}
		}
	}

	// Determine boot loader status and create result
	if bootLoaderDetected {
		bootType := "Unknown"
		if grubDetected && isEFISystem {
			bootType = "GRUB EFI"
		} else if grubDetected {
			bootType = "GRUB Legacy"
		} else if isEFISystem {
			bootType = "systemd-boot (EFI)"
		}

		results = append(results, HealthCheckResult{
			Category:    "system",
			Name:        "Boot Loader",
			Status:      "pass",
			Description: bootType + " boot loader detected",
			Details:     strings.Join(bootLoaderDetails, "; "),
		})
	} else if permissionIssues {
		results = append(results, HealthCheckResult{
			Category:    "system",
			Name:        "Boot Loader",
			Status:      "warn",
			Description: "Boot loader detection limited by permissions",
			Details:     strings.Join(bootLoaderDetails, "; ") + ". Run 'sudo nixai doctor' for complete detection",
			Command:     "sudo bootctl status",
		})
	} else {
		results = append(results, HealthCheckResult{
			Category:    "system",
			Name:        "Boot Loader",
			Status:      "warn",
			Description: "Boot loader configuration unclear",
			Details:     "Unable to detect boot loader: " + strings.Join(bootLoaderDetails, "; "),
			Command:     "sudo bootctl status",
		})
	}

	// Check system uptime
	if uptimeBytes, err := os.ReadFile("/proc/uptime"); err == nil {
		uptimeStr := strings.Fields(string(uptimeBytes))[0]
		results = append(results, HealthCheckResult{
			Category:    "system",
			Name:        "System Uptime",
			Status:      "info",
			Description: "System uptime information",
			Details:     "Uptime: " + uptimeStr + " seconds",
		})
	}

	return results
}

// performNixOSChecks checks NixOS-specific configuration
func performNixOSChecks(configPath string, verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	confNix := configPath
	flakeNix := configPath

	// If configPath is a directory, append file names
	if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
		confNix = configPath + "/configuration.nix"
		flakeNix = configPath + "/flake.nix"
	}

	// Check for configuration files
	confExists := false
	flakeExists := false

	if _, err := os.Stat(confNix); err == nil {
		confExists = true
		results = append(results, HealthCheckResult{
			Category:    "nixos",
			Name:        "Configuration File",
			Status:      "pass",
			Description: "configuration.nix found",
			Details:     "Traditional NixOS configuration detected at " + confNix,
		})
	}

	if _, err := os.Stat(flakeNix); err == nil {
		flakeExists = true
		results = append(results, HealthCheckResult{
			Category:    "nixos",
			Name:        "Flake Configuration",
			Status:      "pass",
			Description: "flake.nix found",
			Details:     "Flake-based configuration detected at " + flakeNix,
		})
	}

	if !confExists && !flakeExists {
		results = append(results, HealthCheckResult{
			Category:    "nixos",
			Name:        "Configuration Files",
			Status:      "fail",
			Description: "No NixOS configuration found",
			Details:     "Neither configuration.nix nor flake.nix found in " + configPath,
			Command:     "nixos-generate-config",
		})
	}

	// Check for hardware configuration
	hwConfPath := configPath + "/hardware-configuration.nix"
	if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
		if _, err := os.Stat(hwConfPath); err == nil {
			results = append(results, HealthCheckResult{
				Category:    "nixos",
				Name:        "Hardware Configuration",
				Status:      "pass",
				Description: "hardware-configuration.nix found",
				Details:     "Hardware-specific configuration is available",
			})
		}
	}

	return results
}

// performPackageChecks checks package and store integrity
func performPackageChecks(verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	// Check Nix store
	if _, err := os.Stat("/nix/store"); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "packages",
			Name:        "Nix Store",
			Status:      "pass",
			Description: "Nix store is accessible",
			Details:     "Package store appears healthy",
		})
	} else {
		results = append(results, HealthCheckResult{
			Category:    "packages",
			Name:        "Nix Store",
			Status:      "fail",
			Description: "Nix store not accessible",
			Details:     "/nix/store not found or inaccessible",
		})
	}

	// Check for nix-channel or flake registry
	if _, err := os.Stat(os.Getenv("HOME") + "/.nix-channels"); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "packages",
			Name:        "Package Channels",
			Status:      "info",
			Description: "Nix channels configured",
			Details:     "Traditional channel-based package management detected",
			Command:     "nix-channel --list",
		})
	}

	return results
}

// performServiceChecks checks system services
func performServiceChecks(verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	// Check systemctl availability
	if _, err := exec.LookPath("systemctl"); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "services",
			Name:        "Service Manager",
			Status:      "pass",
			Description: "systemd is available",
			Details:     "System service management is functional",
			Command:     "systemctl status",
		})

		// Check for failed services
		cmd := exec.Command("systemctl", "--failed", "--no-legend", "--no-pager")
		if output, err := cmd.Output(); err == nil {
			failedServices := strings.TrimSpace(string(output))
			if failedServices == "" {
				results = append(results, HealthCheckResult{
					Category:    "services",
					Name:        "Failed Services",
					Status:      "pass",
					Description: "No failed services detected",
					Details:     "All system services are running properly",
				})
			} else {
				results = append(results, HealthCheckResult{
					Category:    "services",
					Name:        "Failed Services",
					Status:      "warn",
					Description: "Some services have failed",
					Details:     "Failed services detected",
					Command:     "systemctl --failed",
				})
			}
		}
	} else {
		results = append(results, HealthCheckResult{
			Category:    "services",
			Name:        "Service Manager",
			Status:      "warn",
			Description: "systemctl not available",
			Details:     "Cannot check service status",
		})
	}

	return results
}

// performStorageChecks checks storage and filesystem health
func performStorageChecks(verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	// Check disk usage of root filesystem
	if _, err := exec.LookPath("df"); err == nil {
		cmd := exec.Command("df", "-h", "/")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 5 {
					usage := fields[4]
					results = append(results, HealthCheckResult{
						Category:    "storage",
						Name:        "Root Filesystem",
						Status:      "info",
						Description: "Root filesystem usage: " + usage,
						Details:     "Monitor disk space regularly",
						Command:     "df -h",
					})
				}
			}
		}
	}

	// Check for Nix store disk usage
	cmd := exec.Command("du", "-sh", "/nix/store")
	if output, err := cmd.Output(); err == nil {
		storeSize := strings.Fields(string(output))[0]
		results = append(results, HealthCheckResult{
			Category:    "storage",
			Name:        "Nix Store Size",
			Status:      "info",
			Description: "Nix store size: " + storeSize,
			Details:     "Consider garbage collection if size is large",
			Command:     "nix-collect-garbage",
		})
	}

	return results
}

// performNetworkChecks checks network connectivity
func performNetworkChecks(verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	// Check internet connectivity
	cmd := exec.Command("ping", "-c", "1", "-W", "3", "8.8.8.8")
	if err := cmd.Run(); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "network",
			Name:        "Internet Connectivity",
			Status:      "pass",
			Description: "Internet connection is working",
			Details:     "Successfully reached external DNS server",
		})
	} else {
		results = append(results, HealthCheckResult{
			Category:    "network",
			Name:        "Internet Connectivity",
			Status:      "warn",
			Description: "Internet connection issue",
			Details:     "Cannot reach external servers",
			Command:     "ping 8.8.8.8",
		})
	}

	// Check DNS resolution
	cmd = exec.Command("nslookup", "nixos.org")
	if err := cmd.Run(); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "network",
			Name:        "DNS Resolution",
			Status:      "pass",
			Description: "DNS resolution is working",
			Details:     "Successfully resolved nixos.org",
		})
	} else {
		results = append(results, HealthCheckResult{
			Category:    "network",
			Name:        "DNS Resolution",
			Status:      "warn",
			Description: "DNS resolution issue",
			Details:     "Cannot resolve domain names",
			Command:     "cat /etc/resolv.conf",
		})
	}

	return results
}

// performSecurityChecks checks security-related configurations
func performSecurityChecks(verbose bool) []HealthCheckResult {
	var results []HealthCheckResult

	// Check if running as root
	if os.Getuid() == 0 {
		results = append(results, HealthCheckResult{
			Category:    "security",
			Name:        "User Privileges",
			Status:      "warn",
			Description: "Running as root user",
			Details:     "Consider using a non-root user for daily operations",
		})
	} else {
		results = append(results, HealthCheckResult{
			Category:    "security",
			Name:        "User Privileges",
			Status:      "pass",
			Description: "Running as non-root user",
			Details:     "Good security practice",
		})
	}

	// Check SSH configuration if it exists
	if _, err := os.Stat("/etc/ssh/sshd_config"); err == nil {
		results = append(results, HealthCheckResult{
			Category:    "security",
			Name:        "SSH Configuration",
			Status:      "info",
			Description: "SSH server configuration found",
			Details:     "Review SSH security settings",
			Command:     "sudo sshd -T",
		})
	}

	// Check firewall status if available
	if _, err := exec.LookPath("iptables"); err == nil {
		cmd := exec.Command("iptables", "-L", "-n")
		if err := cmd.Run(); err == nil {
			results = append(results, HealthCheckResult{
				Category:    "security",
				Name:        "Firewall Rules",
				Status:      "info",
				Description: "iptables firewall detected",
				Details:     "Review firewall configuration",
				Command:     "sudo iptables -L",
			})
		}
	}

	return results
}

// displayHealthResults shows the health check results in a formatted way
func displayHealthResults(results []HealthCheckResult, verbose bool) {
	fmt.Println(utils.FormatHeader("📊 Health Check Results"))
	fmt.Println()

	categories := make(map[string][]HealthCheckResult)
	var passCount, warnCount, failCount, infoCount int

	// Group results by category
	for _, result := range results {
		categories[result.Category] = append(categories[result.Category], result)

		switch result.Status {
		case "pass":
			passCount++
		case "warn":
			warnCount++
		case "fail":
			failCount++
		case "info":
			infoCount++
		}
	}

	// Display results by category
	categoryOrder := []string{"system", "nixos", "packages", "services", "storage", "network", "security"}
	for _, category := range categoryOrder {
		if results, exists := categories[category]; exists {
			fmt.Println(utils.FormatSubsection(getCategoryTitle(category), ""))

			for _, result := range results {
				status := getStatusIcon(result.Status)
				fmt.Printf("  %s %s\n", status, result.Description)

				if verbose && result.Details != "" {
					fmt.Printf("      %s\n", utils.FormatKeyValue("Details", result.Details))
				}

				if result.Command != "" {
					fmt.Printf("      %s\n", utils.FormatKeyValue("Suggested command", result.Command))
				}
			}
			fmt.Println()
		}
	}

	// Display summary
	fmt.Println(utils.FormatHeader("📈 Health Summary"))
	fmt.Printf("  %s %d checks passed\n", getStatusIcon("pass"), passCount)
	if infoCount > 0 {
		fmt.Printf("  %s %d informational\n", getStatusIcon("info"), infoCount)
	}
	if warnCount > 0 {
		fmt.Printf("  %s %d warnings\n", getStatusIcon("warn"), warnCount)
	}
	if failCount > 0 {
		fmt.Printf("  %s %d failures\n", getStatusIcon("fail"), failCount)
	}

	overallStatus := "healthy"
	if failCount > 0 {
		overallStatus = "critical"
	} else if warnCount > 0 {
		overallStatus = "warnings detected"
	}

	fmt.Printf("\n  Overall Status: %s\n", utils.FormatKeyValue("", overallStatus))
}

// getCategoryTitle returns a formatted title for each category
func getCategoryTitle(category string) string {
	titles := map[string]string{
		"system":   "🖥️  System Health",
		"nixos":    "🐧 NixOS Configuration",
		"packages": "📦 Package Integrity",
		"services": "🔧 System Services",
		"storage":  "💾 Storage Health",
		"network":  "🌐 Network Status",
		"security": "🔒 Security Audit",
	}
	if title, exists := titles[category]; exists {
		return title
	}
	return strings.Title(category)
}

// getStatusIcon returns an appropriate icon for each status
func getStatusIcon(status string) string {
	switch status {
	case "pass":
		return "✅"
	case "warn":
		return "⚠️ "
	case "fail":
		return "❌"
	case "info":
		return "ℹ️ "
	default:
		return "❓"
	}
}

// buildAnalysisPrompt creates a prompt for AI analysis
func buildAnalysisPrompt(results []HealthCheckResult, checkType string) string {
	var promptParts []string

	promptParts = append(promptParts, "You are a NixOS system health expert. Analyze the following health check results and provide:")
	promptParts = append(promptParts, "1. Overall system assessment")
	promptParts = append(promptParts, "2. Priority issues that need immediate attention")
	promptParts = append(promptParts, "3. Recommended fixes with specific commands")
	promptParts = append(promptParts, "4. Prevention tips for maintaining system health")
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, "Health Check Results:")

	for _, result := range results {
		status := map[string]string{
			"pass": "PASS",
			"warn": "WARNING",
			"fail": "FAILURE",
			"info": "INFO",
		}[result.Status]

		promptParts = append(promptParts, fmt.Sprintf("- [%s] %s: %s", status, result.Name, result.Description))
		if result.Details != "" {
			promptParts = append(promptParts, fmt.Sprintf("  Details: %s", result.Details))
		}
		if result.Command != "" {
			promptParts = append(promptParts, fmt.Sprintf("  Suggested: %s", result.Command))
		}
	}

	return strings.Join(promptParts, "\n")
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
	Args: conditionalExactArgsValidator(1),
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

// Missing command handlers

// handleFlakeCommand handles the flake command
func handleFlakeCommand(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	// TODO: Implement flake command functionality
	fmt.Println("Flake command functionality is coming soon!")
}

// handleLearnCommand handles the learn command
func handleLearnCommand(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	// Use the proper implementation from direct_commands.go
	runLearnCmd(args, cmd.OutOrStdout())
}

// handleLogsCommand is the main handler for the logs command
func handleLogsCommand(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	// If no subcommand specified, show help
	if len(args) == 0 {
		cmd.Help()
		return
	}

	// This should not be reached as subcommands handle specific operations
	fmt.Println("Use 'nixai logs --help' to see available subcommands")
}

// handleLogsErrors handles error log analysis
func handleLogsErrors(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	fmt.Println(utils.FormatHeader("🚨 Error Logs Analysis"))
	fmt.Println()

	fmt.Println(utils.FormatProgress("Fetching error logs..."))

	// Get error logs with various patterns
	command := "journalctl --priority=err --lines=50 --no-pager"
	logData, err := runCommand(command)
	if err != nil {
		// Try with sudo if regular access fails
		fmt.Println(utils.FormatWarning("Standard access failed, trying with elevated privileges..."))
		if output, sudoErr := runCommandWithSudo(command); sudoErr == nil {
			logData = output
		} else {
			fmt.Println(utils.FormatError("Failed to fetch error logs: " + sudoErr.Error()))
			return
		}
	}

	if logData == "" {
		fmt.Println(utils.FormatSuccess("No error logs found - system appears healthy!"))
		return
	}

	// Initialize logs agent
	logsAgent, err := initializeLogsAgent()
	if err != nil {
		fmt.Println(utils.FormatWarning("Failed to initialize AI agent, using basic analysis: " + err.Error()))
		displayBasicLogSummary(logData, "errors")
		return
	}

	// Analyze with AI
	fmt.Print(utils.FormatInfo("Analyzing error logs with AI... "))

	ctx := context.Background()
	analysis, err := logsAgent.Query(ctx, fmt.Sprintf("Analyze these error logs, prioritize critical issues, and provide step-by-step troubleshooting recommendations:\n\n%s", logData))

	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Println(utils.FormatError("AI analysis failed: " + err.Error()))
		displayBasicLogSummary(logData, "errors")
		return
	}

	fmt.Println(utils.RenderMarkdown(analysis))
}

// handleLogsAnalyze handles analysis of specific log files
func handleLogsAnalyze(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	fmt.Println(utils.FormatHeader("🔍 Log File Analysis"))
	fmt.Println()

	var logData string
	var err error

	// Check if log file was provided
	if len(args) > 0 {
		logFile := args[0]
		fmt.Printf("Reading log file: %s\n", logFile)

		data, err := os.ReadFile(logFile)
		if err != nil {
			fmt.Println(utils.FormatError("Failed to read log file: " + err.Error()))
			return
		}
		logData = string(data)
	} else {
		// Read from stdin if no file provided
		fmt.Println(utils.FormatInfo("Reading log data from stdin... (press Ctrl+D to finish)"))

		reader := bufio.NewReader(os.Stdin)
		var logLines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break // EOF or error
			}
			logLines = append(logLines, line)
		}
		logData = strings.Join(logLines, "")
	}

	if logData == "" {
		fmt.Println(utils.FormatError("No log data to analyze"))
		fmt.Println(utils.FormatInfo("Usage: nixai logs analyze <file> or pipe data to stdin"))
		return
	}

	// Initialize logs agent
	logsAgent, err := initializeLogsAgent()
	if err != nil {
		fmt.Println(utils.FormatWarning("Failed to initialize AI agent, using basic analysis: " + err.Error()))
		displayBasicLogSummary(logData, "file")
		return
	}

	// Analyze with AI
	fmt.Print(utils.FormatInfo("Analyzing log file with AI... "))

	ctx := context.Background()
	analysis, err := logsAgent.Query(ctx, fmt.Sprintf("Analyze this log file, identify patterns, issues, and provide actionable recommendations:\n\n%s", logData))

	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Println(utils.FormatError("AI analysis failed: " + err.Error()))
		displayBasicLogSummary(logData, "file")
		return
	}

	fmt.Println(utils.RenderMarkdown(analysis))
}

// handleNeovimSetupCommand handles the neovim-setup command
func handleNeovimSetupCommand(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	// TODO: Implement neovim setup functionality
	fmt.Println("Neovim setup functionality is coming soon!")
}

// Neovim setup subcommand handlers
func handleNeovimSetupInstall(cmd *cobra.Command, args []string) {
	configDir, _ := cmd.Flags().GetString("config-dir")
	socketPath, _ := cmd.Flags().GetString("socket-path")

	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("🔧 Installing Neovim Integration"))
	fmt.Fprintln(cmd.OutOrStdout())

	// Use default config directory if not specified
	if configDir == "" {
		var err error
		configDir, err = neovim.GetUserConfigDir()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Failed to get config directory: "+err.Error()))
			return
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Config Directory", configDir))
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Socket Path", socketPath))
	fmt.Fprintln(cmd.OutOrStdout())

	// Create the Neovim module
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Creating nixai.lua module..."))
	err := neovim.CreateNeovimModule(socketPath, configDir)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Failed to create Neovim module: "+err.Error()))
		return
	}

	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess("Neovim integration installed successfully!"))
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("📝 Next Steps", ""))
	fmt.Fprintln(cmd.OutOrStdout(), "1. Add the following to your init.lua:")
	fmt.Fprintln(cmd.OutOrStdout())
	initConfig := neovim.GenerateInitConfig(socketPath)
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatCodeBlock(initConfig, "lua"))
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "2. Restart Neovim to load the integration")
	fmt.Fprintln(cmd.OutOrStdout(), "3. Use <leader>na to ask nixai questions from Neovim")
}

func handleNeovimSetupConfigure(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("⚙️ Configuring Neovim Integration"))
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Configuration options:")
	fmt.Fprintln(cmd.OutOrStdout(), "• Use --socket-path to specify custom MCP socket")
	fmt.Fprintln(cmd.OutOrStdout(), "• Use --config-dir to specify custom Neovim config directory")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Run 'nixai neovim-setup install' to apply configuration"))
}

func handleNeovimSetupStatus(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("📊 Neovim Integration Status"))
	fmt.Fprintln(cmd.OutOrStdout())

	// Check if config directory exists
	configDir, err := neovim.GetUserConfigDir()
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Config Directory", "❌ Not found"))
		return
	}

	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Config Directory", "✅ "+configDir))

	// Check if nixai.lua exists
	nixaiLuaPath := filepath.Join(configDir, "lua", "nixai.lua")
	if _, err := os.Stat(nixaiLuaPath); err == nil {
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("nixai.lua Module", "✅ Installed"))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("nixai.lua Module", "❌ Not found"))
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Run 'nixai neovim-setup install' to install integration"))
		return
	}

	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Integration Status", "✅ Ready"))
}

func handleNeovimSetupUpdate(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("🔄 Updating Neovim Integration"))
	fmt.Fprintln(cmd.OutOrStdout())

	// Re-run installation to update files
	handleNeovimSetupInstall(cmd, args)

	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatNote("Integration updated to latest version"))
}

func handleNeovimSetupRemove(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("🗑️ Removing Neovim Integration"))
	fmt.Fprintln(cmd.OutOrStdout())

	configDir, err := neovim.GetUserConfigDir()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Failed to get config directory: "+err.Error()))
		return
	}

	nixaiLuaPath := filepath.Join(configDir, "lua", "nixai.lua")
	if _, err := os.Stat(nixaiLuaPath); err == nil {
		err := os.Remove(nixaiLuaPath)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Failed to remove nixai.lua: "+err.Error()))
			return
		}
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess("nixai.lua module removed"))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatNote("nixai.lua module not found"))
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Manual cleanup required:"))
	fmt.Fprintln(cmd.OutOrStdout(), "• Remove nixai setup code from your init.lua")
	fmt.Fprintln(cmd.OutOrStdout(), "• Restart Neovim to complete removal")
}

// handlePackageRepoCommand handles the package-repo command
func handlePackageRepoCommand(cmd *cobra.Command, args []string) {
	// Parse command flags
	localPath, _ := cmd.Flags().GetString("local")
	outputPath, _ := cmd.Flags().GetString("output")
	packageName, _ := cmd.Flags().GetString("name")
	analyzeOnly, _ := cmd.Flags().GetBool("analyze-only")

	// Determine repository URL or local path
	var repoURL string
	if localPath == "" && len(args) > 0 {
		repoURL = args[0]
	}

	// Validate input
	if localPath == "" && repoURL == "" {
		fmt.Fprintln(os.Stderr, utils.FormatError("Either repository URL or --local path must be provided"))
		fmt.Fprintln(os.Stderr, utils.FormatTip("Usage: nixai package-repo <repo-url> [flags] or nixai package-repo --local <path> [flags]"))
		return
	}

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	// Initialize AI provider (using the legacy interface for packaging service)
	legacyAIProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}

	// Initialize MCP client
	mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	mcpClient := mcp.NewMCPClient(mcpURL)

	// Create packaging service
	tempDir := "/tmp/nixai-packaging"
	packagingService := packaging.NewPackagingService(
		legacyAIProvider, // Use legacy AI provider directly
		mcpClient,
		tempDir,
		logger.NewLogger(),
	)

	// Create package request
	request := &packaging.PackageRequest{
		RepoURL:     repoURL,
		LocalPath:   localPath,
		OutputPath:  outputPath,
		PackageName: packageName,
		Quiet:       false,
	}

	// Display header
	if localPath != "" {
		fmt.Println(utils.FormatHeader("📦 Analyzing Local Repository: " + localPath))
	} else {
		fmt.Println(utils.FormatHeader("📦 Analyzing Repository: " + repoURL))
	}
	fmt.Println()

	// Execute packaging
	ctx := context.Background()
	result, err := packagingService.PackageRepository(ctx, request)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to package repository: "+err.Error()))
		return
	}

	// Display analysis results
	fmt.Println(utils.FormatHeader("🔍 Repository Analysis"))
	fmt.Println(utils.FormatKeyValue("Project Name", result.Analysis.ProjectName))
	fmt.Println(utils.FormatKeyValue("Language", result.Analysis.Language))
	fmt.Println(utils.FormatKeyValue("Build System", string(result.Analysis.BuildSystem))) // Convert BuildSystem to string
	fmt.Println(utils.FormatKeyValue("Dependencies", fmt.Sprintf("%d found", len(result.Analysis.Dependencies))))

	if len(result.Analysis.Dependencies) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatHeader("📋 Dependencies"))
		for _, dep := range result.Analysis.Dependencies {
			fmt.Printf("  • %s (%s)\n", dep.Name, dep.Type)
		}
	}

	// Display validation issues if any
	if len(result.ValidationIssues) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatHeader("⚠️  Validation Issues"))
		for _, issue := range result.ValidationIssues {
			fmt.Println(utils.FormatWarning("• " + issue))
		}
	}

	// Display derivation if not analyze-only
	if !analyzeOnly && result.Derivation != "" {
		fmt.Println()
		fmt.Println(utils.FormatHeader("📜 Generated Nix Derivation"))
		fmt.Println(utils.RenderMarkdown("```nix\n" + result.Derivation + "\n```"))

		// Save to file if output path specified
		if outputPath != "" {
			err := os.WriteFile(outputPath, []byte(result.Derivation), 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to write derivation to file: "+err.Error()))
			} else {
				fmt.Println()
				fmt.Println(utils.FormatSuccess("✅ Derivation written to: " + outputPath))
			}
		}
	}

	// Display nixpkgs mappings if available
	if len(result.NixpkgsMappings) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatHeader("🗂️  Nixpkgs Mappings"))
		for dep, nixpkg := range result.NixpkgsMappings {
			fmt.Println(utils.FormatKeyValue(dep, nixpkg))
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatSuccess("✅ Repository analysis complete!"))
}

// initializeLogsAgent creates a logs agent with AI provider
func initializeLogsAgent() (*agent.LogsAgent, error) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	legacyProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI provider: %w", err)
	}

	// Adapt legacy provider to new Provider interface
	provider := ai.NewLegacyProviderAdapter(legacyProvider)
	logsAgent := agent.NewLogsAgent(provider)
	return logsAgent, nil
}

// Core log analysis functions (can be called from both CLI and TUI)

// analyzeSystemLogs performs system log analysis and outputs to the provided writer
func analyzeSystemLogs(out io.Writer) {
	fmt.Fprintln(out, utils.FormatHeader("🖥️ System Logs Analysis"))
	fmt.Fprintln(out)

	// Check if we need sudo for some system logs
	var logData string
	var err error

	fmt.Fprintln(out, utils.FormatProgress("Fetching system logs..."))

	// Try to get system logs
	if output, err := runCommand("journalctl --system --lines=100 --no-pager"); err == nil {
		logData = output
	} else {
		// Try with sudo if regular access fails
		fmt.Fprintln(out, utils.FormatWarning("Standard access failed, trying with elevated privileges..."))
		if output, sudoErr := runCommandWithSudo("journalctl --system --lines=100 --no-pager"); sudoErr == nil {
			logData = output
		} else {
			fmt.Fprintln(out, utils.FormatError("Failed to fetch system logs: "+sudoErr.Error()))
			return
		}
	}

	if logData == "" {
		fmt.Fprintln(out, utils.FormatWarning("No system log data found"))
		return
	}

	// Initialize logs agent
	logsAgent, err := initializeLogsAgent()
	if err != nil {
		fmt.Fprintln(out, utils.FormatWarning("Failed to initialize AI agent, using basic analysis: "+err.Error()))
		displayBasicLogSummaryToWriter(out, logData, "system")
		return
	}

	// Analyze with AI
	fmt.Fprint(out, utils.FormatInfo("Analyzing system logs with AI... "))

	ctx := context.Background()
	analysis, err := logsAgent.Query(ctx, fmt.Sprintf("Analyze these system logs for issues, patterns, and recommendations:\n\n%s", logData))

	fmt.Fprintln(out, utils.FormatSuccess("done"))

	if err != nil {
		fmt.Fprintln(out, utils.FormatError("AI analysis failed: "+err.Error()))
		displayBasicLogSummaryToWriter(out, logData, "system")
		return
	}

	fmt.Fprintln(out, utils.RenderMarkdown(analysis))
}

// handleLogsSystem handles system log analysis
func handleLogsSystem(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	analyzeSystemLogs(os.Stdout)
}

// handleLogsBoot handles boot log analysis
func handleLogsBoot(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	fmt.Println(utils.FormatHeader("🚀 Boot Logs Analysis"))
	fmt.Println()

	fmt.Println(utils.FormatProgress("Fetching boot logs..."))

	// Get boot logs
	var logData string
	var err error

	if output, err := runCommand("journalctl --boot --lines=200 --no-pager"); err == nil {
		logData = output
	} else {
		// Try with sudo if regular access fails
		fmt.Println(utils.FormatWarning("Standard access failed, trying with elevated privileges..."))
		if output, sudoErr := runCommandWithSudo("journalctl --boot --lines=200 --no-pager"); sudoErr == nil {
			logData = output
		} else {
			fmt.Println(utils.FormatError("Failed to fetch boot logs: " + sudoErr.Error()))
			return
		}
	}

	if logData == "" {
		fmt.Println(utils.FormatWarning("No boot log data found"))
		return
	}

	// Initialize logs agent
	logsAgent, err := initializeLogsAgent()
	if err != nil {
		fmt.Println(utils.FormatWarning("Failed to initialize AI agent, using basic analysis: " + err.Error()))
		displayBasicLogSummary(logData, "boot")
		return
	}

	// Analyze with AI
	fmt.Print(utils.FormatInfo("Analyzing boot logs with AI... "))

	ctx := context.Background()
	analysis, err := logsAgent.Query(ctx, fmt.Sprintf("Analyze these boot logs for startup issues, errors, and performance insights:\n\n%s", logData))

	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Println(utils.FormatError("AI analysis failed: " + err.Error()))
		displayBasicLogSummary(logData, "boot")
		return
	}

	fmt.Println(utils.RenderMarkdown(analysis))
}

// handleLogsService handles service-specific log analysis
func handleLogsService(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	fmt.Println(utils.FormatHeader("🔧 Service Logs Analysis"))
	fmt.Println()

	var serviceName string
	if len(args) > 0 {
		serviceName = args[0]
	} else {
		fmt.Print("Enter service name: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(utils.FormatError("Failed to read service name: " + err.Error()))
			return
		}
		serviceName = strings.TrimSpace(input)
	}

	if serviceName == "" {
		fmt.Println(utils.FormatError("Service name is required"))
		fmt.Println(utils.FormatInfo("Usage: nixai logs service <service-name>"))
		fmt.Println(utils.FormatInfo("Example: nixai logs service nginx"))
		return
	}

	fmt.Printf("Fetching logs for service: %s\n", utils.FormatKeyValue("Service", serviceName))

	// Get service logs
	command := fmt.Sprintf("journalctl --unit=%s --lines=100 --no-pager", serviceName)
	logData, err := runCommand(command)
	if err != nil {
		// Try with sudo if regular access fails
		fmt.Println(utils.FormatWarning("Standard access failed, trying with elevated privileges..."))
		if output, sudoErr := runCommandWithSudo(command); sudoErr == nil {
			logData = output
		} else {
			fmt.Printf("Failed to fetch logs for service %s: %s\n", serviceName, sudoErr.Error())
			return
		}
	}

	if logData == "" {
		fmt.Printf("No log data found for service: %s\n", serviceName)
		return
	}

	// Initialize logs agent
	logsAgent, err := initializeLogsAgent()
	if err != nil {
		fmt.Println(utils.FormatWarning("Failed to initialize AI agent, using basic analysis: " + err.Error()))
		displayBasicLogSummary(logData, "service")
		return
	}

	// Analyze with AI
	fmt.Print(utils.FormatInfo("Analyzing service logs with AI... "))

	ctx := context.Background()
	analysis, err := logsAgent.Query(ctx, fmt.Sprintf("Analyze these service logs for %s, identify issues, errors, and provide troubleshooting recommendations:\n\n%s", serviceName, logData))

	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Println(utils.FormatError("AI analysis failed: " + err.Error()))
		displayBasicLogSummary(logData, "service")
		return
	}

	fmt.Println(utils.RenderMarkdown(analysis))
}

// handleLogsBuild handles build log analysis
func handleLogsBuild(cmd *cobra.Command, args []string) {
	if globalTUI {
		LaunchTUIMode(cmd, args)
		return
	}

	fmt.Println(utils.FormatHeader("🔨 Build Logs Analysis"))
	fmt.Println()

	var logData string
	var err error

	// Check if log file was provided
	if len(args) > 0 {
		logFile := args[0]
		fmt.Printf("Reading build log from file: %s\n", logFile)

		data, err := os.ReadFile(logFile)
		if err != nil {
			fmt.Println(utils.FormatError("Failed to read log file: " + err.Error()))
			return
		}
		logData = string(data)
	} else {
		// Try to get recent build logs from nixos-rebuild
		fmt.Println(utils.FormatProgress("Searching for recent build logs..."))

		command := "journalctl --unit=nixos-rebuild --lines=200 --no-pager"
		if output, err := runCommand(command); err == nil && strings.TrimSpace(output) != "" {
			logData = output
		} else {
			// Check for nix build logs
			command = "journalctl --identifier=nix --lines=200 --no-pager"
			if output, err := runCommand(command); err == nil && strings.TrimSpace(output) != "" {
				logData = output
			} else {
				fmt.Println(utils.FormatWarning("No recent build logs found"))
				fmt.Println(utils.FormatInfo("Usage: nixai logs build [log-file]"))
				fmt.Println(utils.FormatInfo("Example: nixai logs build /var/log/nixos-rebuild.log"))
				return
			}
		}
	}

	if logData == "" {
		fmt.Println(utils.FormatWarning("No build log data found"))
		return
	}

	// Initialize logs agent
	logsAgent, err := initializeLogsAgent()
	if err != nil {
		fmt.Println(utils.FormatWarning("Failed to initialize AI agent, using basic analysis: " + err.Error()))
		displayBasicLogSummary(logData, "build")
		return
	}

	// Analyze with AI
	fmt.Print(utils.FormatInfo("Analyzing build logs with AI... "))

	ctx := context.Background()
	analysis, err := logsAgent.Query(ctx, fmt.Sprintf("Analyze these build logs for compilation errors, dependency issues, and build optimization suggestions:\n\n%s", logData))

	fmt.Println(utils.FormatSuccess("done"))

	if err != nil {
		fmt.Println(utils.FormatError("AI analysis failed: " + err.Error()))
		displayBasicLogSummary(logData, "build")
		return
	}

	fmt.Println(utils.RenderMarkdown(analysis))
}

// displayBasicLogSummary provides a basic log summary when AI is unavailable
func displayBasicLogSummary(logData, logType string) {
	lines := strings.Split(logData, "\n")

	fmt.Println(utils.FormatSubsection("📊 Basic Log Summary", ""))
	fmt.Println(utils.FormatKeyValue("Log Type", logType))
	fmt.Println(utils.FormatKeyValue("Total Lines", fmt.Sprintf("%d", len(lines))))

	// Count log levels
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "critical") {
			errorCount++
		} else if strings.Contains(lowerLine, "warning") || strings.Contains(lowerLine, "warn") {
			warningCount++
		} else if strings.Contains(lowerLine, "info") {
			infoCount++
		}
	}

	fmt.Println(utils.FormatKeyValue("Errors", fmt.Sprintf("%d", errorCount)))
	fmt.Println(utils.FormatKeyValue("Warnings", fmt.Sprintf("%d", warningCount)))
	fmt.Println(utils.FormatKeyValue("Info Messages", fmt.Sprintf("%d", infoCount)))

	// Show recent entries
	if len(lines) > 0 {
		fmt.Println(utils.FormatSubsection("📋 Recent Entries", ""))
		recentCount := 5
		if len(lines) < recentCount {
			recentCount = len(lines)
		}

		startIdx := len(lines) - recentCount
		for i := startIdx; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				fmt.Printf("  %s\n", lines[i])
			}
		}
	}
}

// displayBasicLogSummaryToWriter provides a basic log summary to an io.Writer when AI is unavailable
func displayBasicLogSummaryToWriter(out io.Writer, logData, logType string) {
	lines := strings.Split(logData, "\n")

	fmt.Fprintln(out, utils.FormatSubsection("📊 Basic Log Summary", ""))
	fmt.Fprintln(out, utils.FormatKeyValue("Log Type", logType))
	fmt.Fprintln(out, utils.FormatKeyValue("Total Lines", fmt.Sprintf("%d", len(lines))))

	// Count log levels
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "critical") {
			errorCount++
		} else if strings.Contains(lowerLine, "warning") || strings.Contains(lowerLine, "warn") {
			warningCount++
		} else if strings.Contains(lowerLine, "info") {
			infoCount++
		}
	}

	fmt.Fprintln(out, utils.FormatKeyValue("Errors", fmt.Sprintf("%d", errorCount)))
	fmt.Fprintln(out, utils.FormatKeyValue("Warnings", fmt.Sprintf("%d", warningCount)))
	fmt.Fprintln(out, utils.FormatKeyValue("Info Messages", fmt.Sprintf("%d", infoCount)))

	// Show recent entries
	if len(lines) > 0 {
		fmt.Fprintln(out, utils.FormatSubsection("📋 Recent Entries", ""))
		recentCount := 5
		if len(lines) < recentCount {
			recentCount = len(lines)
		}

		startIdx := len(lines) - recentCount
		for i := startIdx; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				fmt.Fprintf(out, "  %s\n", lines[i])
			}
		}
	}
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
	rootCmd.AddCommand(NewDepsCommand())
	rootCmd.AddCommand(contextCmd)
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
