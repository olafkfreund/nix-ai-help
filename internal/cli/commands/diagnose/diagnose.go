package diagnose

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"nix-ai-help/internal/ai"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/cli/commands/interfaces"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// DiagnoseCommand implements the CommandHandler interface for the diagnose command
type DiagnoseCommand struct {
	aiProvider  ai.Provider
	config      *config.UserConfig
	contextFile string
	nixosPath   string
	logger      *logger.Logger
}

// NewDiagnoseCommand creates a new DiagnoseCommand instance
func NewDiagnoseCommand(provider ai.Provider, cfg *config.UserConfig, logger *logger.Logger) *DiagnoseCommand {
	return &DiagnoseCommand{
		aiProvider: provider,
		config:     cfg,
		logger:     logger,
	}
}

// GetCommand returns the cobra.Command for the diagnose command
func (dc *DiagnoseCommand) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnose [logfile]",
		Short: "Diagnose NixOS issues from logs or config",
		Long: `Diagnose NixOS issues by analyzing logs, configuration files, or piped input. Uses AI and documentation to suggest fixes.

Examples:
  nixai diagnose /var/log/messages
  journalctl -xe | nixai diagnose
  nixai diagnose --file /var/log/nixos-rebuild.log
  nixai diagnose --type system
  nixai diagnose --context "build failed with dependency error"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := dc.Execute(cmd.Context(), args)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Output)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("file", "f", "", "Specify log file path to analyze")
	cmd.Flags().StringP("type", "t", "", "Diagnostic type (system, config, services, network, hardware, performance)")
	cmd.Flags().StringP("output", "o", "markdown", "Output format (markdown, plain, json)")
	cmd.Flags().StringP("context", "c", "", "Additional context information to include in analysis")
	cmd.Flags().StringVar(&dc.nixosPath, "nixos-path", "", "Path to your NixOS configuration folder")

	return cmd
}

// Execute runs the diagnose command with the given arguments
func (dc *DiagnoseCommand) Execute(ctx context.Context, args []string) (*interfaces.CommandResult, error) {
	startTime := time.Now()

	// Parse command flags from context
	var inputFile, diagType, outputFormat, additionalContext string

	// Extract command flags from environment variables
	inputFile = os.Getenv("NIXAI_DIAGNOSE_FILE")
	diagType = os.Getenv("NIXAI_DIAGNOSE_TYPE")
	outputFormat = os.Getenv("NIXAI_DIAGNOSE_OUTPUT")
	if outputFormat == "" {
		outputFormat = "markdown" // default
	}
	additionalContext = os.Getenv("NIXAI_DIAGNOSE_CONTEXT")

	// Build initial output with header
	output := utils.FormatHeader("🩺 NixOS Diagnostics") + "\n\n"

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(dc.logger)
	nixosCtx, err := contextDetector.GetContext(dc.config)
	if err != nil {
		dc.logger.Warn("Context detection failed: " + err.Error())
		nixosCtx = nil
	}

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		output += utils.FormatNote("📋 "+contextSummary) + "\n\n"
	}

	var logData string

	// Determine input source based on flags and arguments
	if inputFile != "" {
		// Use --file flag
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return &interfaces.CommandResult{
				Output:   utils.FormatError("Failed to read file: " + err.Error()),
				Error:    err,
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, err
		}
		logData = string(data)
	} else if len(args) > 0 {
		// Use positional argument
		file := args[0]
		data, err := os.ReadFile(file)
		if err != nil {
			return &interfaces.CommandResult{
				Output:   utils.FormatError("Failed to read log file: " + err.Error()),
				Error:    err,
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, err
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
				output += fmt.Sprintf("Running %s diagnostics...\n", diagType)
				logData = fmt.Sprintf("Perform %s diagnostics for NixOS system", diagType)
			} else {
				message := utils.FormatWarning("No log file, piped input, or diagnostic type provided.") + "\n" +
					utils.FormatTip("Usage: nixai diagnose [logfile] or nixai diagnose --type system")

				return &interfaces.CommandResult{
					Output:   output + message,
					ExitCode: 1,
					Duration: time.Since(startTime),
				}, nil
			}
		}
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

	output += utils.FormatInfo("Querying AI provider... ")
	resp, err := dc.aiProvider.Query(ctx, contextualPrompt)
	output += utils.FormatSuccess("done") + "\n\n"
	if err != nil {
		return &interfaces.CommandResult{
			Output:   output + utils.FormatError("AI error: "+err.Error()),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Format output based on output format flag
	switch outputFormat {
	case "plain":
		output += resp
	case "json":
		// Simple JSON wrapper
		output += fmt.Sprintf(`{"diagnosis": %q}`, resp)
	default: // markdown
		output += utils.RenderMarkdown(resp)
	}

	return &interfaces.CommandResult{
		Output:   output,
		ExitCode: 0,
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"diagnosisType": diagType,
			"context":       additionalContext,
			"format":        outputFormat,
		},
	}, nil
}

// GetHelp returns detailed help for the diagnose command
func (dc *DiagnoseCommand) GetHelp() string {
	return `The diagnose command analyzes logs, error messages, and system state to identify and fix NixOS issues.

It accepts input in several ways:
- From a specified log file path
- From standard input (piped from other commands)
- By specifying a diagnostic type for general system analysis

The AI uses this input along with your NixOS system context to provide:
- A diagnosis of the underlying issue
- Likely root causes and explanations
- Step-by-step fix instructions tailored to your system
- Links to relevant documentation when available`
}

// GetExamples returns usage examples for the diagnose command
func (dc *DiagnoseCommand) GetExamples() []string {
	return []string{
		"nixai diagnose /var/log/nixos/nixos-rebuild.log",
		"journalctl -xe | nixai diagnose",
		"nixai diagnose --file /var/log/messages",
		"nixai diagnose --type system",
		"nixai diagnose --context \"NixOS build failed with Python package error\"",
	}
}
