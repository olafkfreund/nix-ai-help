package explainoption

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

// ExplainOptionCommand implements the CommandHandler interface for the explain-option command
type ExplainOptionCommand struct {
	aiProvider    ai.Provider
	config        *config.UserConfig
	logger        *logger.Logger
	format        string
	providerFlag  string
	examplesOnly  bool
}

// NewExplainOptionCommand creates a new ExplainOptionCommand instance
func NewExplainOptionCommand(provider ai.Provider, cfg *config.UserConfig, logger *logger.Logger) *ExplainOptionCommand {
	return &ExplainOptionCommand{
		aiProvider: provider,
		config:     cfg,
		logger:     logger,
		format:     "markdown", // default format
	}
}

// GetCommand returns the cobra.Command for the explain-option command
func (ec *ExplainOptionCommand) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain-option <option>",
		Short: "Explain a NixOS option using AI and documentation",
		Long: `Explain a NixOS configuration option using AI and up-to-date documentation.

This command provides:
- Clear option explanations with type information
- Default values and examples of usage
- Documentation references with source attribution
- Module/package relationship details
- Modern Markdown-formatted output by default

Examples:
  nixai explain-option services.openssh.enable
  nixai explain-option networking.hostName --format plain
  nixai explain-option boot.loader.systemd-boot.enable --provider gemini
  nixai explain-option virtualisation.docker.enable --examples-only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get command flags and store on the command object
			ec.format, _ = cmd.Flags().GetString("format")
			ec.providerFlag, _ = cmd.Flags().GetString("provider")
			ec.examplesOnly, _ = cmd.Flags().GetBool("examples-only")

			// Execute the command
			result, err := ec.Execute(cmd.Context(), args)
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

// Execute runs the explain-option command with the given arguments
func (ec *ExplainOptionCommand) Execute(ctx context.Context, args []string) (*interfaces.CommandResult, error) {
	startTime := time.Now()
	
	if len(args) == 0 {
		return &interfaces.CommandResult{
			Output:   utils.FormatError("No option specified.") + "\n" + utils.FormatTip("Example: explain-option services.openssh.enable"),
			Error:    fmt.Errorf("no option specified"),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, fmt.Errorf("no option specified")
	}

	option := args[0]

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(ec.logger)
	nixosCtx, err := contextDetector.GetContext(ec.config)
	if err != nil {
		ec.logger.Warn("Context detection failed: " + err.Error())
		nixosCtx = nil
	}

	// Query MCP for documentation
	var output strings.Builder
	output.WriteString(utils.FormatHeader("NixOS Option: " + option))
	output.WriteString("\n\n")

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		output.WriteString(utils.FormatNote("📋 " + contextSummary))
		output.WriteString("\n\n")
	}

	mcpURL := fmt.Sprintf("http://%s:%d", ec.config.MCPServer.Host, ec.config.MCPServer.Port)
	mcpClient := mcp.NewMCPClient(mcpURL)

	output.WriteString(utils.FormatInfo("Querying documentation... "))
	doc, docErr := mcpClient.QueryDocumentation(option)

	if docErr != nil || doc == "" {
		output.WriteString(utils.FormatError("failed") + "\n\n")
		output.WriteString(utils.FormatError("No documentation found for option: " + option))

		return &interfaces.CommandResult{
			Output:   output.String(),
			Error:    fmt.Errorf("no documentation found for option: %s", option),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, fmt.Errorf("no documentation found")
	}

	output.WriteString(utils.FormatSuccess("done") + "\n\n")

	// Extract source and version information
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
	// Set up AI provider
	aiProviderName := ec.providerFlag
	if aiProviderName == "" {
		aiProviderName = ec.config.AIProvider
	}

	// Create a temporary config with the selected provider if needed
	var aiProvider ai.Provider
	if aiProviderName != ec.config.AIProvider {
		// Create a temporary config with the selected provider
		tempCfg := *ec.config
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
		aiProvider = ec.aiProvider
	}

	// Build context-aware prompt
	var basePrompt string
	if ec.examplesOnly {
		basePrompt = buildExamplesOnlyPrompt(option, doc, ec.format, source, version)
	} else {
		basePrompt = buildEnhancedExplainOptionPrompt(option, doc, ec.format, source, version)
	}

	contextBuilder := nixoscontext.NewNixOSContextBuilder()
	contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt, nixosCtx)

	output.WriteString(utils.FormatInfo("Querying AI provider... "))
	aiResp, aiErr := aiProvider.Query(ctx, contextualPrompt)

	if aiErr != nil {
		output.WriteString(utils.FormatError("failed") + "\n\n")
		output.WriteString(utils.FormatError("AI error: " + aiErr.Error()))

		return &interfaces.CommandResult{
			Output:   output.String(),
			Error:    aiErr,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, aiErr
	}

	output.WriteString(utils.FormatSuccess("done") + "\n\n")
	output.WriteString(utils.RenderMarkdown(aiResp))

	return &interfaces.CommandResult{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"option":        option,
			"format":        ec.format,
			"examplesOnly":  ec.examplesOnly,
			"aiProvider":    aiProviderName,
			"documentation": doc != "",
		},
	}, nil
}

// GetHelp returns detailed help for the explain-option command
func (ec *ExplainOptionCommand) GetHelp() string {
	return `The explain-option command explains NixOS configuration options in detail.

It provides:
- Clear explanation of the option's purpose and behavior
- Type information (boolean, string, etc.)
- Default values when available
- Usage examples and common patterns
- Documentation references with source attribution
- Related options and modules
- Module/package relationship details

The command combines official documentation with AI-powered explanations to give you comprehensive information about any NixOS option.`
}

// GetExamples returns usage examples for the explain-option command
func (ec *ExplainOptionCommand) GetExamples() []string {
	return []string{
		"nixai explain-option services.openssh.enable",
		"nixai explain-option networking.hostName --format plain",
		"nixai explain-option boot.loader.systemd-boot.enable --provider gemini",
		"nixai explain-option virtualisation.docker.enable --examples-only",
	}
}

// buildEnhancedExplainOptionPrompt creates a prompt for the AI to explain an option
func buildEnhancedExplainOptionPrompt(option, doc, format, source, version string) string {
	prompt := fmt.Sprintf(`You are a NixOS configuration expert assistant. Explain the NixOS option '%s' in detail.

DOCUMENTATION:
%s

FORMAT: %s

Provide a comprehensive explanation including:
1. What the option does and its purpose
2. Type information and default value
3. Practical examples of usage (with comments)
4. Common patterns and best practices
5. Related options or modules when relevant

Source attribution: %s
NixOS version: %s

Be accurate, informative, and helpful. Ensure any code examples work correctly.`, option, doc, format, source, version)

	return prompt
}

// buildExamplesOnlyPrompt creates a prompt for the AI focused only on examples
func buildExamplesOnlyPrompt(option, doc, format, source, version string) string {
	prompt := fmt.Sprintf(`You are a NixOS configuration expert assistant. Provide ONLY practical examples of the NixOS option '%s'.

DOCUMENTATION:
%s

FORMAT: %s

Provide ONLY:
1. Multiple practical usage examples with different values/settings
2. Each example should have helpful comments explaining what it does
3. Show common patterns and best practices
4. If relevant, show examples in both configuration.nix and flake.nix contexts

Source attribution: %s
NixOS version: %s

Focus exclusively on examples. Do not explain what the option is or provide other details.`, option, doc, format, source, version)

	return prompt
}
