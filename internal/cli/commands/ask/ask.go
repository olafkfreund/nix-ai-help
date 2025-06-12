package ask

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/agent"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/cli/commands/interfaces"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// AskCommand implements the CommandHandler interface for the ask command
type AskCommand struct {
	aiProvider  ai.Provider
	config      *config.UserConfig
	contextFile string
	nixosPath   string
	logger      *logger.Logger
}

// NewAskCommand creates a new AskCommand instance
func NewAskCommand(provider ai.Provider, cfg *config.UserConfig, logger *logger.Logger) *AskCommand {
	return &AskCommand{
		aiProvider: provider,
		config:     cfg,
		logger:     logger,
	}
}

// GetCommand returns the cobra.Command for the ask command
func (ac *AskCommand) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
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

Examples:
  nixai ask "How do I configure nginx?"
  nixai ask "What is the difference between services.openssh.enable and programs.ssh.enable?"
  nixai ask "How do I set up a development environment with Python?" --provider gemini
  nixai ask "How do I enable SSH?" --quiet
  nixai ask "How do I enable nginx?" --verbose`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get command flags
			quiet, _ := cmd.Flags().GetBool("quiet")
			verbose, _ := cmd.Flags().GetBool("verbose")
			provider, _ := cmd.Flags().GetString("provider")
			model, _ := cmd.Flags().GetString("model")

			// Execute the command
			result, err := ac.Execute(cmd.Context(), args)
			if err != nil {
				return err
			}

			// Write the output
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Output)
			return result.Error
		},
	}

	// Add flags
	cmd.Flags().BoolP("quiet", "q", false, "Suppress validation output and show only the AI response")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed validation output with multi-section layout")
	cmd.Flags().String("provider", "", "AI provider to use")
	cmd.Flags().String("model", "", "AI model to use")
	cmd.Flags().StringVar(&ac.contextFile, "context-file", "", "Path to a file containing context information (JSON or text)")
	cmd.Flags().StringVar(&ac.nixosPath, "nixos-path", "", "Path to your NixOS configuration folder")

	return cmd
}

// Execute runs the ask command with the given arguments
func (ac *AskCommand) Execute(ctx context.Context, args []string) (*interfaces.CommandResult, error) {
	startTime := time.Now()

	// Get output and other parameters from context
	quiet := false
	verbose := false
	var out io.Writer = os.Stdout

	// Parse the question from arguments
	if len(args) == 0 {
		return &interfaces.CommandResult{
			Output:   utils.FormatError("Usage: ask <question>\n") + utils.FormatTip("Example: ask How do I enable nginx?"),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, fmt.Errorf("no question provided")
	}

	question := strings.Join(args, " ")

	// Create a new agent for handling the ask command
	askAgent := agent.NewAskAgent(ac.aiProvider)

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(ac.logger)
	nixosCtx, err := contextDetector.GetContext(ac.config)
	if err != nil {
		ac.logger.Warn("Context detection failed: " + err.Error())
		nixosCtx = nil
	}

	// Generate contextual summary if available
	var contextSummary string
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary = contextBuilder.GetContextSummary(nixosCtx)
	}

	// Execute the ask agent
	result, err := askAgent.Execute(question, contextSummary, nixosCtx)
	if err != nil {
		return &interfaces.CommandResult{
			Output:   utils.FormatError("Failed to get answer: " + err.Error()),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Format the output based on quiet/verbose flags
	output := result
	if !quiet {
		// Add header and context
		output = utils.FormatHeader("🤖 NixOS AI Assistant") + "\n\n" +
			utils.FormatSubHeader("Question: "+question) + "\n\n"

		// Add context if available
		if contextSummary != "" {
			output += utils.FormatNote("📋 "+contextSummary) + "\n\n"
		}

		// Add the AI's answer
		output += utils.FormatSubHeader("Answer:") + "\n\n" + result
	}

	return &interfaces.CommandResult{
		Output:   output,
		ExitCode: 0,
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"question": question,
			"context":  contextSummary,
		},
	}, nil
}

// GetHelp returns detailed help for the ask command
func (ac *AskCommand) GetHelp() string {
	return `The ask command lets you get AI-powered answers to questions about NixOS.

It provides intelligent responses by:
- Accessing NixOS documentation through the MCP server
- Querying package information from nixpkgs
- Analyzing GitHub examples for real-world configurations
- Validating the responses for accuracy and correctness

You can customize the response style using flags:
- Default: Balanced output with summarized validation
- --quiet: Just the answer without validation details
- --verbose: Detailed validation steps and reasoning`
}

// GetExamples returns usage examples for the ask command
func (ac *AskCommand) GetExamples() []string {
	return []string{
		"nixai ask \"How do I configure nginx?\"",
		"nixai ask \"What's the difference between services.openssh and programs.ssh?\"",
		"nixai ask \"How do I set up a Python environment?\" --provider openai",
		"nixai ask \"How do I enable SSH?\" --quiet",
		"nixai ask \"Show me an example flake.nix\" --verbose",
	}
}
