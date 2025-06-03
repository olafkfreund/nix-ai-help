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

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nixai",
	Short: "NixOS AI Assistant",
	Long:  `nixai: AI-powered CLI for NixOS diagnostics, search, and configuration.`,
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

func buildExplainOptionPrompt(option, documentation string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users understand configuration options. Please explain the following NixOS option in a clear, practical manner.\n\n**Option:** %s\n\n**Official Documentation:**\n%s\n\n**Please provide:**\n\n1. **Purpose & Overview**: What this option does and why you'd use it\n2. **Type & Default**: The data type and default value (if any)\n3. **Usage Examples**: Show 2-3 practical configuration examples\n4. **Best Practices**: How to use this option effectively\n5. **Related Options**: Other options that are commonly used with this one\n6. **Common Issues**: Potential problems and their solutions\n\nFormat your response using Markdown with section headings and code blocks for examples.`, option, documentation)
}

// Execute runs the root command
func Execute() {
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(explainOptionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
