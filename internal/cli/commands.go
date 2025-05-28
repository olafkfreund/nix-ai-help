package cli

import (
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/nixos"

	"github.com/spf13/cobra"
)

// Command structure for the CLI
var rootCmd = &cobra.Command{
	Use:   "nixai",
	Short: "NixAI helps solve Nix configuration problems",
	Long:  `NixAI is a command-line tool that assists users in diagnosing and solving NixOS configuration issues using AI models and documentation queries.`,
}

var logFile string
var configSnippet string

// Tail the last n lines of a string
func tailLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// Initialize the CLI commands
func init() {
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(logsCmd)

	diagnoseCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Path to a log file to analyze")
	diagnoseCmd.Flags().StringVarP(&configSnippet, "config-snippet", "c", "", "NixOS configuration snippet to analyze")
}

// Diagnose command to analyze NixOS configuration issues
var diagnoseCmd = &cobra.Command{
	Use:   "diagnose [log or config snippet]",
	Short: "Diagnose NixOS configuration issues",
	Long:  `Diagnose NixOS configuration issues using logs, config snippets, or stdin.`,
	Run: func(cmd *cobra.Command, args []string) {
		var logInput, userInput string

		// 1. If --log-file is set, read log from file
		if logFile != "" {
			data, err := os.ReadFile(logFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read log file: %v\n", err)
				os.Exit(1)
			}
			logInput = tailLines(string(data), 200)
		}

		// 2. If --config-snippet is set, use as user input
		if configSnippet != "" {
			userInput = configSnippet
		}

		// 3. If stdin is not a terminal, read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinData, _ := os.ReadFile("/dev/stdin")
			if logInput == "" {
				logInput = tailLines(string(stdinData), 200)
			} else {
				userInput = string(stdinData)
			}
		}

		// 4. If args are provided, treat as log or config snippet
		if len(args) > 0 {
			if logInput == "" {
				logInput = tailLines(args[0], 200)
			} else if userInput == "" {
				userInput = args[0]
			}
		}

		if logInput == "" && userInput == "" {
			fmt.Fprintln(os.Stderr, "No log or config snippet provided. Use --log-file, --config-snippet, pipe input, or pass as argument.")
			os.Exit(1)
		}

		// Load config and AI provider (simplified, real code should use internal/config)
		provider := ai.NewOllamaProvider("llama3") // Changed from llama2 to llama3
		diagnostics := nixos.Diagnose(logInput, userInput, provider)
		fmt.Print(nixos.FormatDiagnostics(diagnostics))
	},
}

// Configure command to set up NixOS configurations
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure NixOS settings",
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation for configuring NixOS
		fmt.Println("Configuring NixOS settings...")
		// Call the appropriate functions to configure NixOS
	},
}

// Logs command to parse and analyze log outputs
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Parse and analyze log outputs",
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation for parsing logs
		fmt.Println("Parsing log outputs...")
		// Call the appropriate functions to parse logs
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
