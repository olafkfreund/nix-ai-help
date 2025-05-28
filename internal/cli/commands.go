package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
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
var nixosConfigPath string
var nixLogTarget string // New: for --nix-log flag

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
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(searchCmd)

	diagnoseCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Path to a log file to analyze")
	diagnoseCmd.Flags().StringVarP(&configSnippet, "config-snippet", "c", "", "NixOS configuration snippet to analyze")
	diagnoseCmd.Flags().StringVarP(&nixLogTarget, "nix-log", "g", "", "Run 'nix log' (optionally with a path or derivation) and analyze the output") // New flag
	searchCmd.Flags().StringVarP(&nixosConfigPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
}

// Diagnose command to analyze NixOS configuration issues
var diagnoseCmd = &cobra.Command{
	Use:   "diagnose [log or config snippet]",
	Short: "Diagnose NixOS configuration issues",
	Long: `Diagnose NixOS configuration issues using logs, config snippets, nix log output, or stdin.

Options:
  --log-file, -l: Path to a log file to analyze
  --nix-log, -g: Run 'nix log' (optionally with a path or derivation) and analyze the output
  --config-snippet, -c: NixOS configuration snippet to analyze`,
	Run: func(cmd *cobra.Command, args []string) {
		var logInput, userInput string

		// 1. If --nix-log is set, run 'nix log' and use its output
		if nixLogTarget != "" {
			cmdArgs := []string{"log"}
			if strings.TrimSpace(nixLogTarget) != "" {
				cmdArgs = append(cmdArgs, nixLogTarget)
			}
			out, err := exec.Command("nix", cmdArgs...).CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to run 'nix log': %v\nOutput: %s\n", err, string(out))
				os.Exit(1)
			}
			logInput = tailLines(string(out), 200)
		}

		// 2. If --log-file is set, read log from file (unless --nix-log already used)
		if logFile != "" && logInput == "" {
			data, err := os.ReadFile(logFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read log file: %v\n", err)
				os.Exit(1)
			}
			logInput = tailLines(string(data), 200)
		}

		// 3. If --config-snippet is set, use as user input
		if configSnippet != "" {
			userInput = configSnippet
		}

		// 4. If stdin is not a terminal, read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinData, _ := os.ReadFile("/dev/stdin")
			if logInput == "" {
				logInput = tailLines(string(stdinData), 200)
			} else {
				userInput = string(stdinData)
			}
		}

		// 5. If args are provided, treat as log or config snippet
		if len(args) > 0 {
			if logInput == "" {
				logInput = tailLines(args[0], 200)
			} else if userInput == "" {
				userInput = args[0]
			}
		}

		if logInput == "" && userInput == "" {
			fmt.Fprintln(os.Stderr, "No log, nix log, or config snippet provided. Use --nix-log, --log-file, --config-snippet, pipe input, or pass as argument.")
			os.Exit(1)
		}

		// Load config and select AI provider
		cfg, err := config.LoadYAMLConfig("configs/default.yaml")
		var provider ai.AIProvider
		if err == nil {
			switch cfg.AIProvider {
			case "ollama":
				provider = ai.NewOllamaProvider("llama3")
			case "gemini":
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://api.gemini.com")
			case "openai":
				provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			default:
				provider = ai.NewOllamaProvider("llama3")
			}
		} else {
			provider = ai.NewOllamaProvider("llama3")
		}

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

// Interactive command to start interactive mode
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive mode",
	Long:  `Start an interactive shell for diagnosing and configuring NixOS with nixai. Type 'help' for commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		InteractiveMode()
	},
}

// Search command to find and list Nix packages or services
var searchCmd = &cobra.Command{
	Use:   "search <pkg|service> <query>",
	Short: "Search for Nix packages or services and show configuration options",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchType := "package"
		if len(args) > 1 && (args[0] == "service" || args[0] == "pkg" || args[0] == "package") {
			searchType = args[0]
			args = args[1:]
		}
		query := strings.Join(args, " ")
		fmt.Printf("Searching for Nix %ss matching: %s\n", searchType, query)
		if nixosConfigPath != "" {
			fmt.Printf("Using NixOS config folder: %s\n", nixosConfigPath)
			// Optionally: pass this to executor or use for context
		}
		var output string
		var err error
		executor := nixos.NewExecutor()
		if searchType == "service" {
			// Instead of using 'nix search nixos', which fails, provide a helpful message and example config
			fmt.Println("Service search is not yet fully automated. For most NixOS services, use:")
			fmt.Printf("  services.%s.enable = true;\n", query)
			fmt.Println("For more options, see the NixOS manual or run: nixos-option services.", query)
			return
		} else {
			output, err = executor.SearchNixPackages(query)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
			os.Exit(1)
		}
		lines := strings.Split(output, "\n")
		var items []struct{ Attr, Name, Desc string }
		var lastAttr string
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
				// Hide 'evaluating ...' lines from output and parsing
			}
			if strings.HasPrefix(line, "evaluating ") {
				continue
			}
			// Try to match: attr [tab or space] description
			fields := strings.Fields(line)
			if len(fields) > 1 && (strings.Contains(fields[0], ".") || strings.Contains(fields[0], ":")) {
				attr := fields[0]
				attr = strings.TrimPrefix(attr, "legacyPackages.x86_64-linux.")
				attr = strings.TrimPrefix(attr, "nixpkgs.")
				name := attr
				desc := strings.Join(fields[1:], " ")
				items = append(items, struct{ Attr, Name, Desc string }{fields[0], name, desc})
				lastAttr = fields[0]
				continue
			}
			// If the line is indented, treat as a description for the last item
			if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && lastAttr != "" {
				if len(items) > 0 {
					items[len(items)-1].Desc += " " + strings.TrimSpace(line)
				}
				continue
			}
			// Fallback: try to match lines like 'name - description'
			if idx := strings.Index(line, " - "); idx > 0 {
				name := line[:idx]
				desc := line[idx+3:]
				items = append(items, struct{ Attr, Name, Desc string }{name, name, desc})
				lastAttr = name
			}
		}
		// Only print package results, not any skipped lines
		if len(items) == 0 {
			fmt.Println("No matches found. Please refine your query or check the spelling.")
			return
		}
		fmt.Println()
		for i, item := range items {
			fmt.Printf("%2d. %s\n    %s\n", i+1, item.Name, item.Desc)
		}
		fmt.Print("\nEnter the number of the ", searchType, " to see configuration options (or leave blank to exit): ")
		var sel string
		scan := bufio.NewScanner(os.Stdin)
		if scan.Scan() {
			sel = strings.TrimSpace(scan.Text())
		}
		if sel == "" {
			return
		}
		idx := -1
		fmt.Sscanf(sel, "%d", &idx)
		if idx < 1 || idx > len(items) {
			fmt.Println("Invalid selection.")
			return
		}
		item := items[idx-1]
		fmt.Printf("\nSelected: %s\n\n", item.Name)
		// Show config options (placeholder or MCP/doc search)
		// executor already defined above
		if searchType == "service" {
			fmt.Printf("  services.%s.enable = true;\n", item.Name)
			fmt.Println("  # For more options, see the NixOS manual or run: nixos-option services.", item.Name)
			fmt.Println("\nFetching available options with nixos-option...")
			optOut, err := executor.ShowNixOSOptions("services." + item.Name)
			if err == nil && strings.TrimSpace(optOut) != "" {
				fmt.Println(optOut)
			} else {
				fmt.Println("No additional options found or nixos-option failed.")
			}
		} else {
			fmt.Println("NixOS (configuration.nix):")
			fmt.Printf("  environment.systemPackages = with pkgs; [ %s ];\n", item.Name)
			fmt.Println("Home Manager (home.nix):")
			fmt.Printf("  home.packages = with pkgs; [ %s ];\n", item.Name)
			fmt.Println("\nFetching available options with nixos-option...")
			optOut, err := executor.ShowNixOSOptions(item.Name)
			if err == nil && strings.TrimSpace(optOut) != "" {
				fmt.Println(optOut)
			} else {
				fmt.Println("No additional options found or nixos-option failed.")
			}
		}
		fmt.Print("\nWould you like to install this ", searchType, " [i], run in nix-shell [r], or exit [Enter]? ")
		var action string
		if scan.Scan() {
			action = strings.TrimSpace(scan.Text())
		}
		if action == "i" && searchType == "package" {
			fmt.Printf("\nInstalling %s...\n", item.Name)
			installOut, err := executor.InstallNixPackage(item.Attr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error installing: %v\n", err)
			} else {
				fmt.Println(installOut)
			}
		} else if action == "r" && searchType == "package" {
			fmt.Printf("\nLaunching 'nix-shell -p %s'...\n", item.Name)
			cmdShell := exec.Command("nix-shell", "-p", item.Name)
			cmdShell.Stdin = os.Stdin
			cmdShell.Stdout = os.Stdout
			cmdShell.Stderr = os.Stderr
			err := cmdShell.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running nix-shell: %v\n", err)
			}
		}
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
