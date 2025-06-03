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
var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Community resources and support (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'community' command is not yet implemented.")
	},
}
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage nixai configuration (not yet implemented)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[nixai] The 'config' command is not yet implemented.")
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

func buildExplainOptionPrompt(option, documentation string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users understand configuration options. Please explain the following NixOS option in a clear, practical manner.\n\n**Option:** %s\n\n**Official Documentation:**\n%s\n\n**Please provide:**\n\n1. **Purpose & Overview**: What this option does and why you'd use it\n2. **Type & Default**: The data type and default value (if any)\n3. **Usage Examples**: Show 2-3 practical configuration examples\n4. **Best Practices**: How to use this option effectively\n5. **Related Options**: Other options that are commonly used with this one\n6. **Common Issues**: Potential problems and their solutions\n\nFormat your response using Markdown with section headings and code blocks for examples.`, option, documentation)
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
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(explainOptionCmd)
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
