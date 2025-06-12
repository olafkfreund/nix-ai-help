package explainhomeoption

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/cli/commands/interfaces"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// ExplainHomeOptionCommand implements the CommandHandler interface for the explain-home-option command
type ExplainHomeOptionCommand struct {
	aiProvider   ai.Provider
	config       *config.UserConfig
	logger       *logger.Logger
	format       string
	providerFlag string
	examplesOnly bool
}

// NewExplainHomeOptionCommand creates a new ExplainHomeOptionCommand instance
func NewExplainHomeOptionCommand(provider ai.Provider, cfg *config.UserConfig, logger *logger.Logger) *ExplainHomeOptionCommand {
	return &ExplainHomeOptionCommand{
		aiProvider: provider,
		config:     cfg,
		logger:     logger,
		format:     "markdown", // default format
	}
}

// GetCommand returns the cobra.Command for the explain-home-option command
func (ehc *ExplainHomeOptionCommand) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain-home-option <option>",
		Short: "Explain a Home Manager option using AI and documentation",
		Long: `Explain a Home Manager configuration option using AI and up-to-date documentation.

This command provides:
- Clear option explanations with type information
- Default values and examples of usage
- Documentation references with source attribution
- Module relationship details
- Modern Markdown-formatted output by default

Examples:
  nixai explain-home-option programs.git.enable
  nixai explain-home-option programs.firefox.enable --format plain
  nixai explain-home-option services.gpg-agent.enable --provider gemini
  nixai explain-home-option programs.neovim.enable --examples-only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get command flags
			ehc.format, _ = cmd.Flags().GetString("format")
			ehc.providerFlag, _ = cmd.Flags().GetString("provider")
			ehc.examplesOnly, _ = cmd.Flags().GetBool("examples-only")

			// Execute the command
			result, err := ehc.Execute(cmd.Context(), args)
			if err != nil {
				return err
			}

			// Write the output
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Output)
			return nil
		},
	}

	// Add flags
	cmd.Flags().String("format", "markdown", "Output format: markdown, plain, or table")
	cmd.Flags().String("provider", "", "AI provider to use for this query (ollama, openai, gemini, claude)")
	cmd.Flags().Bool("examples-only", false, "Show only usage examples for the option")

	return cmd
}

// parseMCPOptionDoc parses the documentation from MCP and extracts structured information
func parseMCPOptionDoc(doc string) (opt struct {
	Name        string
	Type        string
	Default     string
	Example     string
	Description string
	Source      string
	Version     string
	Related     []string
	Links       []string
}, fallback string) {

	fallback = doc // Default fallback to the original doc

	// Extract fields if they exist in the documentation
	if strings.Contains(doc, "option_name") {
		parts := strings.Split(doc, "option_name")
		if len(parts) > 1 {
			opt.Name = strings.Split(parts[1], "\"")[1]
		}
	}

	if strings.Contains(doc, "option_type") {
		parts := strings.Split(doc, "option_type")
		if len(parts) > 1 {
			opt.Type = strings.Split(parts[1], "\"")[1]
		}
	}

	if strings.Contains(doc, "option_default") {
		parts := strings.Split(doc, "option_default")
		if len(parts) > 1 {
			opt.Default = strings.Split(parts[1], "\"")[1]
		}
	}

	if strings.Contains(doc, "option_example") {
		parts := strings.Split(doc, "option_example")
		if len(parts) > 1 {
			opt.Example = strings.Split(parts[1], "\"")[1]
		}
	}

	if strings.Contains(doc, "option_description") {
		parts := strings.Split(doc, "option_description")
		if len(parts) > 1 {
			opt.Description = strings.Split(parts[1], "\"")[1]
		}
	}

	if strings.Contains(doc, "option_source") {
		parts := strings.Split(doc, "option_source")
		if len(parts) > 1 {
			opt.Source = strings.Split(parts[1], "\"")[1]
		}
	}

	if strings.Contains(doc, "nixos-") {
		idx := strings.Index(doc, "nixos-")
		opt.Version = doc[idx : idx+12]
	}

	return opt, fallback
}

// Execute runs the explain-home-option command with the given arguments
func (ehc *ExplainHomeOptionCommand) Execute(ctx context.Context, args []string) (*interfaces.CommandResult, error) {
	startTime := time.Now()

	if len(args) == 0 {
		return &interfaces.CommandResult{
			Output:   utils.FormatError("No option specified.") + "\n" + utils.FormatTip("Example: explain-home-option programs.git.enable"),
			Error:    fmt.Errorf("no option specified"),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, fmt.Errorf("no option specified")
	}

	option := args[0]

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(ehc.logger)
	nixosCtx, err := contextDetector.GetContext(ehc.config)
	if err != nil {
		ehc.logger.Warn("Context detection failed: " + err.Error())
		nixosCtx = nil
	}

	// Start building the output
	var output strings.Builder
	output.WriteString(utils.FormatHeader("🏠 Home Manager Option: " + option))
	output.WriteString("\n\n")

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		output.WriteString(utils.FormatNote("📋 " + contextSummary))
		output.WriteString("\n\n")
	}

	// Set up AI provider
	aiProviderName := ehc.providerFlag
	if aiProviderName == "" {
		aiProviderName = ehc.config.AIProvider
	}

	// Create a temporary config with the selected provider if needed
	var aiProvider ai.Provider
	if aiProviderName != ehc.config.AIProvider {
		// Create a temporary config with the selected provider
		tempCfg := *ehc.config
		tempCfg.AIProvider = aiProviderName

		var err error
		aiProvider, err = ai.GetProvider(tempCfg.AIProvider, tempCfg.AIModel)
		if err != nil {
			output.WriteString(utils.FormatError("Failed to initialize AI provider: " + err.Error()))
			return &interfaces.CommandResult{
				Output:   output.String(),
				Error:    err,
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, err
		}
	} else {
		aiProvider = ehc.aiProvider
	}

	// Query MCP for documentation
	var docExcerpts []string
	output.WriteString(utils.FormatInfo("Querying documentation... "))

	mcpBase := ehc.config.MCPServer.Host
	if mcpBase != "" {
		mcpURL := fmt.Sprintf("http://%s:%d", ehc.config.MCPServer.Host, ehc.config.MCPServer.Port)
		mcpClient := mcp.NewMCPClient(mcpURL)
		doc, err := mcpClient.QueryDocumentation(option)

		if err == nil && doc != "" {
			output.WriteString(utils.FormatSuccess("done"))
			output.WriteString("\n\n")

			opt, fallbackDoc := parseMCPOptionDoc(doc)
			if opt.Name != "" {
				context := fmt.Sprintf("Option: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nHome Manager Version: %s",
					opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version)
				docExcerpts = append(docExcerpts, context)
			} else {
				docExcerpts = append(docExcerpts, fallbackDoc)
			}
		} else {
			output.WriteString(utils.FormatWarning("no documentation found"))
			output.WriteString("\n\n")
		}
	} else {
		output.WriteString(utils.FormatWarning("skipped (no MCP host configured)"))
		output.WriteString("\n\n")
	}

	// Build the AI prompt
	promptCtx := ai.PromptContext{
		Question:     option,
		DocExcerpts:  docExcerpts,
		Intent:       "explain",
		OutputFormat: ehc.format,
		Provider:     aiProviderName,
	}

	builder := ai.DefaultPromptBuilder{}
	basePrompt, err := builder.BuildPrompt(promptCtx)
	if err != nil {
		output.WriteString(utils.FormatError("Prompt build error: " + err.Error()))
		return &interfaces.CommandResult{
			Output:   output.String(),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Build context-aware prompt using the context builder
	contextBuilder := nixoscontext.NewNixOSContextBuilder()
	contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt, nixosCtx)

	output.WriteString(utils.FormatInfo("Querying AI provider... "))
	aiResp, aiErr := aiProvider.Query(ctx, contextualPrompt)

	if aiErr != nil {
		output.WriteString(utils.FormatError("failed"))
		output.WriteString("\n\n")
		output.WriteString(utils.FormatError("AI error: " + aiErr.Error()))

		return &interfaces.CommandResult{
			Output:   output.String(),
			Error:    aiErr,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, aiErr
	}

	output.WriteString(utils.FormatSuccess("done"))
	output.WriteString("\n\n")
	output.WriteString(utils.RenderMarkdown(aiResp))

	return &interfaces.CommandResult{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"option":        option,
			"format":        ehc.format,
			"examplesOnly":  ehc.examplesOnly,
			"aiProvider":    aiProviderName,
			"documentation": len(docExcerpts) > 0,
		},
	}, nil
}

// GetHelp returns detailed help for the explain-home-option command
func (ehc *ExplainHomeOptionCommand) GetHelp() string {
	return `The explain-home-option command explains Home Manager configuration options in detail.

It provides:
- Clear explanation of the option's purpose and behavior
- Type information (boolean, string, etc.)
- Default values when available 
- Usage examples and common patterns
- Documentation references with source attribution
- Related options and modules
- Integration details with NixOS when relevant

The command combines official Home Manager documentation with AI-powered explanations to give you comprehensive information about any Home Manager option.`
}

// GetExamples returns usage examples for the explain-home-option command
func (ehc *ExplainHomeOptionCommand) GetExamples() []string {
	return []string{
		"nixai explain-home-option programs.git.enable",
		"nixai explain-home-option programs.firefox.enable --format plain",
		"nixai explain-home-option services.gpg-agent.enable --provider gemini",
		"nixai explain-home-option programs.neovim.enable --examples-only",
	}
}
