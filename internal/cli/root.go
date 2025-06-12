package cli

import (
	"fmt"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/cli/commands/ask"
	"nix-ai-help/internal/cli/commands/diagnose"
	explainhomeoption "nix-ai-help/internal/cli/commands/explain-home-option"
	explainoption "nix-ai-help/internal/cli/commands/explain-option"
	"nix-ai-help/internal/cli/registry"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
	"nix-ai-help/pkg/version"

	"github.com/spf13/cobra"
)

// RootFlags holds global flags for the root command
type RootFlags struct {
	AskQuestion string // used for the direct ask functionality
	NixosPath   string
	ContextFile string
	AIProvider  string
	AIModel     string
	TUIMode     bool
}

var (
	// Global flags instance
	globalFlags RootFlags
	// Global registry instance
	globalRegistry *registry.CommandRegistry
)

// NewRootCommand creates the root cobra.Command for the application
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
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
			if globalFlags.TUIMode && cmd.Name() != "interactive" {
				// For non-interactive commands, launch TUI with the command pre-selected
				return LaunchTUIMode(cmd, append([]string{cmd.Name()}, args...))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for global TUI flag first
			if globalFlags.TUIMode {
				// If TUI mode is requested, launch the TUI with any provided args
				return LaunchTUIMode(cmd, args)
			}

			// Get the ask flag value
			askQuestion, _ := cmd.Flags().GetString("ask")

			if askQuestion != "" {
				// Set environment variables for provider and model flags
				if globalFlags.AIProvider != "" {
					os.Setenv("NIXAI_PROVIDER", globalFlags.AIProvider)
				}
				if globalFlags.AIModel != "" {
					os.Setenv("NIXAI_MODEL", globalFlags.AIModel)
				}

				// Get the ask command handler from the registry and execute it
				handler, ok := globalRegistry.Get("ask")
				if ok {
					result, err := handler.Execute(cmd.Context(), []string{askQuestion})
					if err != nil {
						return err
					}
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Output)
					return nil
				}

				// Fallback to the old implementation if handler is not found
				runAskCmdWithConciseMode([]string{askQuestion}, os.Stdout, "", "")
				return nil
			}

			// Direct question mode (when args are provided without flags)
			if len(args) > 0 {
				question := args[0]
				// Get the ask command handler from the registry
				handler, ok := globalRegistry.Get("ask")
				if ok {
					result, err := handler.Execute(cmd.Context(), []string{question})
					if err != nil {
						return err
					}
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Output)
					return nil
				}

				// Fallback to the old implementation
				runAskCmdWithConciseMode([]string{question}, os.Stdout, "", "")
				return nil
			}

			// If no question or --ask, show help
			return cmd.Help()
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVarP(&globalFlags.AskQuestion, "ask", "a", "", "Ask a question about NixOS configuration")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.NixosPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.ContextFile, "context-file", "", "Path to a file containing context information (JSON or text)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.AIProvider, "provider", "", "Specify the AI provider (ollama, openai, gemini, claude, etc.)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.AIModel, "model", "", "Specify the AI model (llama3, gpt-4, gemini-1.5-pro, claude-3-opus, etc.)")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.TUIMode, "tui", false, "Launch TUI mode for any command")

	return rootCmd
}

// InitializeCommandRegistry creates and returns a command registry with all commands registered
func InitializeCommandRegistry() *registry.CommandRegistry {
	// Create the root command
	rootCmd := NewRootCommand()

	// Create registry with root command
	globalRegistry = registry.NewRegistry(rootCmd)

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to load config: "+err.Error()))
		os.Exit(1)
	}

	// Create AI provider
	provider, err := ai.GetProvider(cfg.AIProvider, cfg.AIModel)
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		os.Exit(1)
	}

	// Create logger
	log := logger.NewLogger()

	// Register commands
	// Register the 'ask' command
	askCommand := ask.NewAskCommand(provider, cfg, log)
	globalRegistry.Register("ask", askCommand)

	// Register the 'diagnose' command
	diagnoseCommand := diagnose.NewDiagnoseCommand(provider, cfg, log)
	globalRegistry.Register("diagnose", diagnoseCommand)

	// Register the 'explain-option' command
	explainOptionCommand := explainoption.NewExplainOptionCommand(provider, cfg, log)
	globalRegistry.Register("explain-option", explainOptionCommand)

	// Register the 'explain-home-option' command
	explainHomeOptionCommand := explainhomeoption.NewExplainHomeOptionCommand(provider, cfg, log)
	globalRegistry.Register("explain-home-option", explainHomeOptionCommand)

	// TODO: Register more commands here

	return globalRegistry
}

// Execute runs the root command with appropriate initialization
func Execute() {
	// Set environment variable for nixosPath if provided
	if globalFlags.NixosPath != "" {
		if err := os.Setenv("NIXAI_NIXOS_PATH", globalFlags.NixosPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set NIXAI_NIXOS_PATH: %v\n", err)
		}
	}

	// Initialize command registry and get the root command
	reg := InitializeCommandRegistry()

	// Execute root command
	rootCmd := reg.GetRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
