package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/utils"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Command structure for the CLI
var rootCmd = &cobra.Command{
	Use:   "nixai [question]",
	Short: "NixAI helps solve Nix configuration problems",
	Long: `NixAI is a command-line tool that assists users in diagnosing and solving NixOS configuration issues using AI models and documentation queries.

You can also ask questions directly, e.g.:
  nixai "how can I configure curl?"`,
	Args: cobra.ArbitraryArgs,
}

var logFile string
var configSnippet string
var nixosConfigPath string
var nixLogTarget string          // New: for --nix-log flag
var nixosConfigPathGlobal string // Global path for build/flake context

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
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(flakeCmd)
	rootCmd.AddCommand(explainOptionCmd)
	rootCmd.AddCommand(mcpServerCmd)   // Register the MCP server command
	rootCmd.AddCommand(healthCheckCmd) // Register the health check command

	diagnoseCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Path to a log file to analyze")
	diagnoseCmd.Flags().StringVarP(&configSnippet, "config-snippet", "c", "", "NixOS configuration snippet to analyze")
	diagnoseCmd.Flags().StringVarP(&nixLogTarget, "nix-log", "g", "", "Run 'nix log' (optionally with a path or derivation) and analyze the output") // New flag
	searchCmd.Flags().StringVarP(&nixosConfigPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.PersistentFlags().StringVarP(&nixosConfigPathGlobal, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	configCmd.AddCommand(showUserConfig)
	mcpServerCmd.AddCommand(mcpServerStartCmd)
	mcpServerCmd.AddCommand(mcpServerStopCmd)
	mcpServerCmd.AddCommand(mcpServerStatusCmd)
	mcpServerStartCmd.Flags().BoolP("background", "d", false, "Run MCP server in background (daemon mode)")
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
		cfg, err := config.LoadUserConfig()
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
		cfg, _ := config.LoadUserConfig()
		configPath := ""
		if nixosConfigPath != "" {
			fmt.Printf("Using NixOS config folder: %s\n", nixosConfigPath)
			configPath = nixosConfigPath
		} else if cfg != nil && cfg.NixosFolder != "" {
			configPath = cfg.NixosFolder
		}
		if configPath == "" || !utils.IsDirectory(configPath) {
			fmt.Fprintf(os.Stderr, "[Error] NixOS config path is not set or does not exist: '%s'\n", configPath)
			fmt.Fprintln(os.Stderr, "Set the config path with --nixos-path/-n or in your config file. Example:")
			fmt.Fprintln(os.Stderr, "  nixai search --nixos-path /etc/nixos pkg <query>")
			fmt.Fprintln(os.Stderr, "Or set it interactively with 'set-nixos-path' in interactive mode.")
			os.Exit(1)
		}
		var output string
		var err error
		executor := nixos.NewExecutor(configPath)
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
			fmt.Println(utils.FormatWarning("No matches found. Please refine your query or check the spelling."))
			return
		}

		// Display search results with enhanced formatting
		fmt.Println(utils.FormatHeader("🔍 Search Results"))
		fmt.Println(utils.FormatKeyValue("Query", query))
		fmt.Println(utils.FormatKeyValue("Type", searchType))
		fmt.Println(utils.FormatDivider())

		for i, item := range items {
			title := fmt.Sprintf("%d. %s", i+1, item.Name)
			fmt.Println(utils.FormatSubsection(title, utils.MutedStyle.Render(item.Desc)))
		}

		fmt.Print(utils.FormatInfo("Enter the number of the " + searchType + " to see configuration options (or leave blank to exit): "))
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
			fmt.Println(utils.FormatError("Invalid selection."))
			return
		}
		item := items[idx-1]

		// Display selected item with enhanced formatting
		fmt.Println(utils.FormatHeader("📦 Selected " + strings.Title(searchType)))
		fmt.Println(utils.FormatKeyValue("Name", item.Name))
		fmt.Println(utils.FormatKeyValue("Description", item.Desc))
		fmt.Println(utils.FormatDivider())

		// Show config options (placeholder or MCP/doc search)
		// executor already defined above
		if searchType == "service" {
			fmt.Println(utils.FormatSection("Configuration Example", ""))
			fmt.Println(utils.FormatCodeBlock(fmt.Sprintf("services.%s.enable = true;", item.Name), "nix"))
			fmt.Println(utils.FormatNote("For more options, see the NixOS manual or run: nixos-option --find services." + item.Name))
			fmt.Println(utils.FormatProgress("Fetching available options with nixos-option --find..."))
			optOut, err := executor.ListServiceOptions(item.Name)
			if err == nil && strings.TrimSpace(optOut) != "" {
				fmt.Println(utils.FormatSection("Available Options", optOut))
			} else {
				fmt.Println(utils.FormatWarning("No additional options found or nixos-option --find failed."))
			}
		} else {
			fmt.Println(utils.FormatSection("Configuration Examples", ""))

			nixosConfig := fmt.Sprintf("environment.systemPackages = with pkgs; [ %s ];", item.Name)
			hmConfig := fmt.Sprintf("home.packages = with pkgs; [ %s ];", item.Name)

			fmt.Println(utils.FormatSubsection("NixOS (configuration.nix)", ""))
			fmt.Println(utils.FormatCodeBlock(nixosConfig, "nix"))
			fmt.Println(utils.FormatSubsection("Home Manager (home.nix)", ""))
			fmt.Println(utils.FormatCodeBlock(hmConfig, "nix"))

			fmt.Println(utils.FormatProgress("Fetching available options with nixos-option..."))
			optOut, err := executor.ShowNixOSOptions(item.Name)
			if err == nil && strings.TrimSpace(optOut) != "" {
				fmt.Println(utils.FormatSection("Available Options", optOut))
			} else {
				fmt.Println(utils.FormatWarning("No additional options found or nixos-option failed."))
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

// Config command for AI-assisted Nix configuration management
var configCmd = &cobra.Command{
	Use:   "config [show|set|unset|edit|explain] [key] [value]",
	Short: "AI-assisted Nix configuration management",
	Long:  `Manage and understand your Nix configuration with AI-powered help.\nExamples:\n  nixai config show\n  nixai config set experimental-features nix-command flakes\n  nixai config explain substituters`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
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
		if len(args) == 0 || args[0] == "show" {
			out, err := exec.Command("nix", "config", "show").CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to run 'nix config show': %v\nOutput: %s\n", err, string(out))
				os.Exit(1)
			}
			fmt.Println("Current Nix configuration:")
			fmt.Println(string(out))
			prompt := "Summarize and suggest improvements for this Nix config:\n" + string(out)
			aiResp, err := provider.Query(prompt)
			if err == nil && aiResp != "" {
				fmt.Println("\nAI suggestions:")
				fmt.Println(aiResp)
			}
			return
		}
		if args[0] == "set" && len(args) >= 3 {
			key := args[1]
			value := strings.Join(args[2:], " ")
			cmdOut, err := exec.Command("nix", "config", "set", key, value).CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to set config: %v\nOutput: %s\n", err, string(cmdOut))
				os.Exit(1)
			}
			fmt.Printf("Set %s = %s\n", key, value)
			return
		}
		if args[0] == "unset" && len(args) >= 2 {
			key := args[1]
			cmdOut, err := exec.Command("nix", "config", "unset", key).CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to unset config: %v\nOutput: %s\n", err, string(cmdOut))
				os.Exit(1)
			}
			fmt.Printf("Unset %s\n", key)
			return
		}
		if args[0] == "edit" {
			cmdOut, err := exec.Command("nix", "config", "edit").CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to edit config: %v\nOutput: %s\n", err, string(cmdOut))
				os.Exit(1)
			}
			fmt.Println("Opened config in editor.")
			return
		}
		if args[0] == "explain" && len(args) >= 2 {
			key := args[1]
			prompt := "Explain the Nix config option '" + key + "' and how to use it."
			aiResp, err := provider.Query(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "AI error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(aiResp)
			return
		}
		fmt.Println("Usage: nixai config [show|set|unset|edit|explain] [key] [value]")
	},
}

// Build command for AI-assisted nix build troubleshooting
var buildCmd = &cobra.Command{
	Use:   "build [args]",
	Short: "AI-assisted nix build/flakes troubleshooting and guidance",
	Long:  `Build or rebuild your NixOS system or packages, with AI-powered help for flakes and configuration issues.\nExamples:\n  nixai build\n  nixai build .#nixosConfigurations.myhost.config.system.build.toplevel\n  nixai build --flake .`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
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
		cmdArgs := []string{"build"}
		if len(args) > 0 {
			cmdArgs = append(cmdArgs, args...)
		}
		command := exec.Command("nix", cmdArgs...)
		if nixosConfigPathGlobal != "" {
			command.Dir = nixosConfigPathGlobal
		}
		out, err := command.CombinedOutput()
		fmt.Println(string(out))
		if err != nil {
			fmt.Fprintf(os.Stderr, "nix build failed: %v\n", err)
			// Parse and summarize the error output for the user (basic version)
			problemSummary := summarizeBuildOutput(string(out))
			if problemSummary != "" {
				fmt.Println("\nProblem summary:")
				fmt.Println(problemSummary)
			}
			prompt := "I ran 'nix build" + " " + strings.Join(args, " ") + "' and got this output:\n" + string(out) + "\nHow can I fix this build or configuration problem?"
			aiResp, aiErr := provider.Query(prompt)
			if aiErr == nil && aiResp != "" {
				fmt.Println("\nAI suggestions:")
				fmt.Println(aiResp)
			}
			os.Exit(1)
		}
	},
}

// summarizeBuildOutput provides a simple summary of common nix build errors.
func summarizeBuildOutput(output string) string {
	lines := strings.Split(output, "\n")
	var summary []string
	for _, line := range lines {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") || strings.Contains(line, "cannot") {
			summary = append(summary, line)
		}
	}
	return strings.Join(summary, "\n")
}

// Flake command for AI-assisted nix flake troubleshooting
var flakeCmd = &cobra.Command{
	Use:   "flake [args]",
	Short: "AI-assisted nix flake commands and troubleshooting",
	Long:  `Run nix flake commands (show, update, check, etc.) with AI-powered help for troubleshooting and configuration.\nExamples:\n  nixai flake show\n  nixai flake update\n  nixai flake check\n  nixai flake explain-inputs\n  nixai flake explain <input>`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && (args[0] == "explain-inputs" || args[0] == "explain") {
			ExplainFlakeInputs(args[1:])
			return
		}
		cmdArgs := []string{"flake"}
		if len(args) > 0 {
			cmdArgs = append(cmdArgs, args...)
		}
		command := exec.Command("nix", cmdArgs...)
		if nixosConfigPathGlobal != "" {
			command.Dir = nixosConfigPathGlobal
		}
		out, err := command.CombinedOutput()
		fmt.Println(string(out))
		if err != nil {
			fmt.Fprintf(os.Stderr, "nix flake failed: %v\n", err)
			problemSummary := summarizeBuildOutput(string(out))
			if problemSummary != "" {
				fmt.Println("\nProblem summary:")
				fmt.Println(problemSummary)
			}
			prompt := "I ran 'nix flake" + " " + strings.Join(args, " ") + "' and got this output:\n" + string(out) + "\nHow can I fix this flake or configuration problem?"
			cfg, _ := config.LoadYAMLConfig("configs/default.yaml")
			var provider ai.AIProvider
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
			aiResp, aiErr := provider.Query(prompt)
			if aiErr == nil && aiResp != "" {
				fmt.Println("\nAI suggestions:")
				fmt.Println(aiResp)
			}
			os.Exit(1)
		}
	},
}

// renderForTerminal formats markdown and simple HTML for terminal output with color
func renderForTerminal(input string) string {
	if input == "" {
		return ""
	}
	// Remove HTML tags except <b>, <i>, <code>, <pre>, <a>, <ul>, <ol>, <li>
	re := regexp.MustCompile(`<[^>]+>`)
	clean := re.ReplaceAllStringFunc(input, func(tag string) string {
		// Allow some tags to pass as markdown
		switch {
		case tag == "<b>" || tag == "</b>":
			return "**"
		case tag == "<i>" || tag == "</i>":
			return "_"
		case tag == "<code>" || tag == "</code>":
			return "`"
		case tag == "<pre>" || tag == "</pre>":
			return "\n```\n"
		case tag == "<ul>" || tag == "</ul>" || tag == "<ol>" || tag == "</ol>":
			return ""
		case tag == "<li>":
			return "- "
		case tag == "</li>":
			return "\n"
		case tag[:2] == "<a":
			return ""
		case tag == "</a>":
			return ""
		default:
			return ""
		}
	})
	// Use glamour to render markdown to ANSI
	out, err := glamour.Render(clean, "dark")
	if err != nil {
		return clean
	}
	return out
}

// ExplainFlakeInputs handles 'nixai flake explain-inputs' and 'nixai flake explain <input>'
func ExplainFlakeInputs(inputs []string) {
	// Use the correct flake.nix path based on nixosConfigPathGlobal
	flakePath := "flake.nix"
	if nixosConfigPathGlobal != "" {
		flakePath = nixosConfigPathGlobal + "/flake.nix"
	}
	data, err := os.ReadFile(flakePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read flake.nix at %s: %v\n", flakePath, err)
		os.Exit(1)
	}
	flakeInputs := parseFlakeInputs(string(data))
	if len(flakeInputs) == 0 {
		fmt.Println("No inputs found in flake.nix.")
		return
	}
	fmt.Println("Found the following flake inputs:")
	for i, inp := range flakeInputs {
		fmt.Printf("%2d. %s\n", i+1, inp.Name)
		fmt.Printf("    URL: %s\n", inp.URL)
	}
	// 2. If a specific input is requested, focus on that
	var targets []FlakeInput
	if len(inputs) > 0 && inputs[0] != "" {
		for _, inp := range flakeInputs {
			if inp.Name == inputs[0] {
				targets = []FlakeInput{inp}
				break
			}
		}
		if len(targets) == 0 {
			fmt.Printf("Input '%s' not found in flake.nix.\n", inputs[0])
			return
		}
	} else {
		targets = flakeInputs
	}
	// 3. For each input, try to fetch README.md and flake.nix from GitHub if possible
	for _, inp := range targets {
		readme, flake, repoURL := fetchGitHubReadmeAndFlake(inp.URL)
		fmt.Println(renderForTerminal("\n---\n"))
		fmt.Print(renderForTerminal(fmt.Sprintf("**Input:** `%s`\n**Source:** %s\n", inp.Name, inp.URL)))
		if repoURL != "" {
			fmt.Print(renderForTerminal(fmt.Sprintf("**Repo:** %s\n", repoURL)))
		}
		if readme != "" {
			fmt.Println(renderForTerminal("### README.md summary:"))
			fmt.Println(renderForTerminal(summarizeText(readme)))
		} else {
			fmt.Println(renderForTerminal("_No README.md found._"))
		}
		if flake != "" {
			fmt.Println(renderForTerminal("### flake.nix summary:"))
			fmt.Println(renderForTerminal(summarizeText(flake)))
		} else {
			fmt.Println(renderForTerminal("_No flake.nix found in repo._"))
		}
		// 4. Use AI to explain and suggest improvements
		cfg, _ := config.LoadYAMLConfig("configs/default.yaml")
		var provider ai.AIProvider
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
		prompt := "Explain the purpose and best practices for the Nix flake input '" + inp.Name + "' with this source URL: " + inp.URL + ".\n" +
			"README.md (if any):\n" + readme + "\nflake.nix (if any):\n" + flake + "\nSuggest improvements or highlight issues."
		aiResp, aiErr := provider.Query(prompt)
		if aiErr == nil && aiResp != "" {
			fmt.Println(renderForTerminal("\n**AI explanation and suggestions:**"))
			fmt.Println(renderForTerminal(aiResp))
		}
	}
}

type FlakeInput struct {
	Name string
	URL  string
}

// parseFlakeInputs extracts flake inputs from a flake.nix string (supports both 'name.url = ...;' and 'name = { url = ...; ... };' forms)
func parseFlakeInputs(flake string) []FlakeInput {
	var inputs []FlakeInput
	inInputs := false
	lines := strings.Split(flake, "\n")
	var i int
	for i = 0; i < len(lines); i++ {
		l := strings.TrimSpace(lines[i])
		if l == "" || strings.HasPrefix(l, "#") {
			continue // skip empty lines and comments
		}
		if strings.HasPrefix(l, "inputs =") || strings.HasPrefix(l, "inputs=") {
			inInputs = true
			continue
		}
		if inInputs && l == "}" {
			break
		}
		if !inInputs {
			continue
		}
		// Match 'name.url = ...;' form
		if strings.Contains(l, ".url") && strings.Contains(l, "=") {
			parts := strings.SplitN(l, "=", 2)
			namePart := strings.TrimSpace(parts[0])
			if !strings.HasSuffix(namePart, ".url") {
				continue
			}
			name := strings.TrimSuffix(namePart, ".url")
			url := strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
			url = strings.Trim(url, "\"'")
			inputs = append(inputs, FlakeInput{Name: name, URL: url})
			continue
		}
		// Match 'name = { url = ...; ... };' form
		if strings.HasSuffix(l, "=") && i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), "{") {
			name := strings.TrimSuffix(strings.TrimSpace(strings.TrimSuffix(l, "=")), " ")
			// Parse the attribute set block
			blockLevel := 0
			url := ""
			for j := i + 1; j < len(lines); j++ {
				line := strings.TrimSpace(lines[j])
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if strings.HasPrefix(line, "{") {
					blockLevel++
				}
				if strings.HasPrefix(line, "}") || strings.HasPrefix(line, "};") {
					blockLevel--
					if blockLevel <= 0 {
						i = j // advance outer loop
						break
					}
				}
				// Look for 'url = ...;' inside the block
				if strings.HasPrefix(line, "url") && strings.Contains(line, "=") {
					urlParts := strings.SplitN(line, "=", 2)
					urlVal := strings.TrimSpace(strings.TrimSuffix(urlParts[1], ";"))
					urlVal = strings.Trim(urlVal, "\"'")
					url = urlVal
				}
			}
			if name != "" && url != "" {
				inputs = append(inputs, FlakeInput{Name: name, URL: url})
			}
			continue
		}
	}
	return inputs
}

// fetchGitHubReadmeAndFlake tries to fetch README.md and flake.nix from a GitHub repo URL
func fetchGitHubReadmeAndFlake(url string) (readme, flake, repoURL string) {
	if !strings.HasPrefix(url, "github:") {
		return "", "", ""
	}
	// github:NixOS/nixpkgs/nixos-unstable -> NixOS/nixpkgs
	parts := strings.Split(url, ":")
	if len(parts) < 2 {
		return "", "", ""
	}
	repo := parts[1]
	if idx := strings.Index(repo, "/"); idx > 0 {
		if idx2 := strings.Index(repo[idx+1:], "/"); idx2 > 0 {
			repo = repo[:idx+1+idx2]
		}
	}
	repoURL = "https://github.com/" + repo
	readmeURL := "https://raw.githubusercontent.com/" + repo + "/master/README.md"
	flakeURL := "https://raw.githubusercontent.com/" + repo + "/master/flake.nix"
	readme = fetchURL(readmeURL)
	flake = fetchURL(flakeURL)
	return readme, flake, repoURL
}

// fetchURL fetches the content of a URL (returns empty string on error)
func fetchURL(url string) string {
	resp, err := exec.Command("curl", "-sL", url).Output()
	if err != nil {
		return ""
	}
	return string(resp)
}

// summarizeText returns the first 20 lines or 1500 chars of text
func summarizeText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) > 20 {
		return strings.Join(lines[:20], "\n") + "\n..."
	}
	if len(text) > 1500 {
		return text[:1500] + "..."
	}
	return text
}

// Execute runs the root command
func Execute() {
	// Set from env if not set by flag
	if nixosConfigPathGlobal == "" {
		if envPath := os.Getenv("NIXAI_NIXOS_PATH"); envPath != "" {
			nixosConfigPathGlobal = envPath
		}
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// MCP server command
var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Manage the MCP documentation/query server",
	Long:  `Manage the Model Context Protocol (MCP) server for NixOS documentation and option queries. Use start/stop/status subcommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// MCP server start command
var mcpServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		background, _ := cmd.Flags().GetBool("background")
		if background {
			absPath, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not determine nixai binary path: %v\n", err)
				os.Exit(1)
			}
			cmdStr := fmt.Sprintf("nohup %s mcp-server start > mcp.log 2>&1 &", absPath)
			fmt.Println("Starting MCP server in background...")
			shCmd := exec.Command("sh", "-c", cmdStr)
			if err := shCmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start MCP server in background: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("MCP server started in background. Logs: mcp.log")
			return
		}
		// Foreground
		server, err := mcp.NewServerFromConfig("configs/default.yaml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create MCP server: %v\n", err)
			os.Exit(1)
		}
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

// MCP server stop command
var mcpServerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadYAMLConfig("configs/default.yaml")
		if err != nil {
			fmt.Println(utils.FormatError("Failed to load config: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatProgress("Stopping MCP server..."))
		addr := fmt.Sprintf("http://%s:%d/shutdown", cfg.MCPServer.Host, cfg.MCPServer.Port)
		resp, err := http.Get(addr)
		if err != nil {
			fmt.Println(utils.FormatError("Failed to contact MCP server: " + err.Error()))
			os.Exit(1)
		}
		defer resp.Body.Close()
		msg, _ := io.ReadAll(resp.Body)

		if strings.TrimSpace(string(msg)) != "" {
			fmt.Println(utils.FormatSuccess("MCP server stopped successfully"))
			fmt.Println(utils.InfoStyle.Render(string(msg)))
		} else {
			fmt.Println(utils.FormatSuccess("MCP server stopped"))
		}
	},
}

// MCP server status command
var mcpServerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show MCP server status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadYAMLConfig("configs/default.yaml")
		if err != nil {
			fmt.Println(utils.FormatError("Failed to load config: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatProgress("Checking MCP server status..."))
		addr := fmt.Sprintf("http://%s:%d/healthz", cfg.MCPServer.Host, cfg.MCPServer.Port)
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(addr)
		if err != nil {
			fmt.Println(utils.FormatError("MCP server is NOT running"))
			fmt.Println(utils.FormatInfo("Start it with: nixai mcp-server start"))
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == 200 && strings.TrimSpace(string(body)) == "ok" {
			fmt.Println(utils.FormatSuccess("MCP server is running"))
			fmt.Println(utils.FormatKeyValue("Address", addr))
			fmt.Println(utils.FormatKeyValue("Status", "Healthy"))
		} else {
			fmt.Println(utils.FormatWarning("MCP server is responding but not healthy"))
			fmt.Println(utils.FormatKeyValue("Response", strings.TrimSpace(string(body))))
		}
	},
}

// Show user config command
var showUserConfig = &cobra.Command{
	Use:   "show-user",
	Short: "Show the current user config (~/.config/nixai/config.yaml)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Println("Error loading user config:", err)
			return
		}
		out, _ := yaml.Marshal(cfg)
		fmt.Println(string(out))
	},
}

// extractNixOSOption attempts to extract a NixOS option from a natural language query.
func extractNixOSOption(input string) string {
	// Try to find a pattern like services.nginx.enable or similar
	re := regexp.MustCompile(`([a-zA-Z0-9_.-]+\.[a-zA-Z0-9_.-]+(\.[a-zA-Z0-9_.-]+)*)`)
	matches := re.FindAllString(input, -1)
	if len(matches) > 0 {
		return matches[0]
	}
	// Fallback: return the input as-is (may be a direct option)
	return strings.TrimSpace(input)
}

// suggestSimilarOptions attempts to suggest similar or related options when an option isn't found
func suggestSimilarOptions(option string) []string {
	suggestions := []string{}

	// Extract the service/module name for suggestions
	parts := strings.Split(option, ".")
	if len(parts) >= 2 {
		if parts[0] == "services" && len(parts) >= 2 {
			serviceName := parts[1]
			suggestions = append(suggestions, []string{
				fmt.Sprintf("services.%s.enable", serviceName),
				fmt.Sprintf("services.%s.package", serviceName),
				fmt.Sprintf("services.%s.settings", serviceName),
				fmt.Sprintf("services.%s.extraConfig", serviceName),
			}...)
		}

		if parts[0] == "networking" {
			suggestions = append(suggestions, []string{
				"networking.firewall.enable",
				"networking.hostName",
				"networking.interfaces",
				"networking.nameservers",
			}...)
		}

		if parts[0] == "boot" {
			suggestions = append(suggestions, []string{
				"boot.loader.systemd-boot.enable",
				"boot.loader.grub.enable",
				"boot.kernelPackages",
				"boot.initrd.availableKernelModules",
			}...)
		}
	}

	return suggestions
}

// buildExplainOptionPrompt creates a comprehensive prompt for AI to explain NixOS options with usage examples
func buildExplainOptionPrompt(option, documentation string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users understand configuration options. Please explain the following NixOS option in a clear, practical manner.

**Option:** %s

**Official Documentation:**
%s

**Please provide:**

1. **Purpose & Overview**: What this option does and why you'd use it
2. **Type & Default**: The data type and default value (if any)
3. **Usage Examples**: Show 2-3 practical configuration examples:
   - Basic/minimal usage
   - Common real-world scenario
   - Advanced configuration (if applicable)
4. **Best Practices**: Tips, warnings, or recommendations
5. **Related Options**: Other options that work well with this one

**Format your response using Markdown with:**
- Clear headings (##)
- Code blocks for configuration examples
- Bullet points for lists
- **Bold** text for emphasis

Make it practical and actionable for someone configuring their NixOS system.`, option, documentation)
}

// Explain option command for AI-powered NixOS option explanation
var explainOptionCmd = &cobra.Command{
	Use:   "explain-option <option|question>",
	Short: "Explain a NixOS option using AI and documentation",
	Long:  `Get a concise, AI-generated explanation for any NixOS option, including type, default, and best practices. Accepts natural language queries.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		option := extractNixOSOption(query)
		fmt.Println(utils.FormatProgress("Analyzing NixOS option: " + option))

		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading user config: "+err.Error()))
			os.Exit(1)
		}

		// Check MCP server status before querying
		fmt.Print(utils.FormatProgress("Fetching official documentation..."))
		mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
		statusResp, err := http.Get(mcpURL + "/healthz")
		if err != nil || statusResp.StatusCode != 200 {
			fmt.Println(utils.FormatError("MCP server is not running"))
			fmt.Println(utils.FormatInfo("Please start it with 'nixai mcp-server start' or 'nixai mcp-server start -d'"))
			os.Exit(1)
		}
		if statusResp != nil {
			statusResp.Body.Close()
		}

		mcpClient := mcp.NewMCPClient(mcpURL)
		doc, err := mcpClient.QueryDocumentation(option)
		if err != nil {
			fmt.Println(utils.FormatError("Error querying documentation: " + err.Error()))
			os.Exit(1)
		}
		if strings.TrimSpace(doc) == "" || strings.Contains(doc, "No relevant documentation found") {
			fmt.Println(utils.FormatWarning("No relevant documentation found for this option"))
			suggestions := suggestSimilarOptions(option)
			if len(suggestions) > 0 {
				fmt.Println(utils.FormatSubsection("Did you mean one of these options?", ""))
				fmt.Println(utils.FormatList(suggestions))
			}
			return
		}
		fmt.Println(utils.FormatSuccess("Documentation found!"))

		// Select AI provider
		fmt.Print(utils.FormatProgress("Generating explanation with " + cfg.AIProvider + "..."))
		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			provider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://api.gemini.com")
		case "openai":
			provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			provider = ai.NewOllamaProvider("llama3")
		}

		prompt := buildExplainOptionPrompt(option, doc)
		aiResp, aiErr := provider.Query(prompt)
		if aiErr != nil {
			fmt.Println(utils.FormatError("AI error: " + aiErr.Error()))
			os.Exit(1)
		}
		if strings.TrimSpace(aiResp) == "" {
			fmt.Println("\n" + utils.FormatError("AI did not return an explanation."))
			return
		}
		fmt.Println(utils.FormatSuccess("Complete!"))

		// Create a beautiful header for the explanation
		fmt.Println("\n" + utils.FormatHeader("📋 NixOS Option Explanation"))
		fmt.Println(utils.FormatKeyValue("Option", option))
		fmt.Println(utils.FormatDivider())

		// Render output as markdown in terminal with enhanced styling
		out, err := glamour.Render(aiResp, "dark")
		if err != nil {
			// Fallback to plain text with basic formatting
			fmt.Println(utils.FormatSection("Explanation", aiResp))
		} else {
			fmt.Print(out)
		}

		fmt.Println("\n" + utils.FormatDivider())

		// Enhanced tips section
		fmt.Println(utils.FormatTip("Use 'nixai search service <name>' to find related services"))
		fmt.Println(utils.FormatNote("Run 'nixai mcp-server query <option>' for raw documentation"))
	},
}

// Health check command for comprehensive system checks
var healthCheckCmd = &cobra.Command{
	Use:   "health",
	Short: "Run comprehensive NixOS system health check",
	Long: `Performs a comprehensive health check of your NixOS system including:
- Configuration validation
- System services status
- Disk space analysis
- Nix channels status
- Boot system integrity
- Network connectivity
- Nix store health
- AI-powered analysis and recommendations`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		fmt.Print(utils.FormatProgress("🔍 Starting NixOS health check"))
		fmt.Println()

		runHealthCheck(cfg)
	},
}

// runHealthCheck performs comprehensive system health checks
func runHealthCheck(cfg *config.UserConfig) {
	var healthReport []string
	var issues []string
	var warnings []string

	fmt.Print(utils.FormatProgress("⚙️ Checking NixOS configuration validity"))
	fmt.Println()

	// 1. Check configuration validity with dry-run
	executor := nixos.NewExecutor(utils.ExpandHome(cfg.NixosFolder))
	configValid, configOutput, err := checkConfigurationValidity(executor)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Configuration validation error: %v", err))
	} else if !configValid {
		issues = append(issues, fmt.Sprintf("Configuration validation failed:\n%s", configOutput))
	} else {
		healthReport = append(healthReport, "✅ **Configuration**: Valid and ready for deployment")
	}

	fmt.Print(utils.FormatProgress("🔧 Analyzing system services"))
	fmt.Println()

	// 2. Check critical system services
	serviceStatus, serviceIssues := checkSystemServices(executor)
	healthReport = append(healthReport, serviceStatus...)
	issues = append(issues, serviceIssues...)

	fmt.Print(utils.FormatProgress("💾 Checking disk space"))
	fmt.Println()

	// 3. Check disk space
	diskStatus, diskWarnings := checkDiskSpace(executor)
	healthReport = append(healthReport, diskStatus...)
	warnings = append(warnings, diskWarnings...)

	fmt.Print(utils.FormatProgress("📡 Checking Nix channels"))
	fmt.Println()

	// 4. Check Nix channels
	channelStatus, channelIssues := checkNixChannels(executor)
	healthReport = append(healthReport, channelStatus...)
	issues = append(issues, channelIssues...)

	fmt.Print(utils.FormatProgress("🚀 Checking boot system"))
	fmt.Println()

	// 5. Check boot system integrity
	bootStatus, bootIssues := checkBootSystem(executor)
	healthReport = append(healthReport, bootStatus...)
	issues = append(issues, bootIssues...)

	fmt.Print(utils.FormatProgress("🌐 Testing network connectivity"))
	fmt.Println()

	// 6. Check network connectivity
	networkStatus, networkIssues := checkNetworkConnectivity()
	healthReport = append(healthReport, networkStatus...)
	issues = append(issues, networkIssues...)

	fmt.Print(utils.FormatProgress("📦 Checking Nix store health"))
	fmt.Println()

	// 7. Check Nix store health
	storeStatus, storeIssues := checkNixStore(executor)
	healthReport = append(healthReport, storeStatus...)
	issues = append(issues, storeIssues...)

	// Generate comprehensive report
	fmt.Print(utils.FormatProgress("🤖 Generating AI-powered analysis"))
	fmt.Println()

	report := generateHealthReport(healthReport, warnings, issues)

	// Get AI analysis if there are issues or warnings
	if len(issues) > 0 || len(warnings) > 0 {
		aiAnalysis, err := getAIHealthAnalysis(cfg, issues, warnings)
		if err == nil {
			report += "\n\n" + utils.FormatSection("🤖 AI Analysis & Recommendations", aiAnalysis)
		}
	}

	// Render the report with glamour
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Println(report)
		return
	}

	output, err := r.Render(report)
	if err != nil {
		fmt.Println(report)
		return
	}

	fmt.Print(output)
}

// checkConfigurationValidity checks if the NixOS configuration is valid
func checkConfigurationValidity(executor *nixos.Executor) (bool, string, error) {
	output, err := executor.ExecuteCommand("sudo", "nixos-rebuild", "dry-run", "--show-trace")
	if err != nil {
		return false, output, err
	}

	// Check for common error patterns
	errorPatterns := []string{
		"error:",
		"assertion failed",
		"infinite recursion",
		"stack overflow",
		"syntax error",
	}

	outputLower := strings.ToLower(output)
	for _, pattern := range errorPatterns {
		if strings.Contains(outputLower, pattern) {
			return false, output, nil
		}
	}

	return true, output, nil
}

// checkSystemServices checks the status of critical system services
func checkSystemServices(executor *nixos.Executor) ([]string, []string) {
	var status []string
	var issues []string

	// Critical services to check
	criticalServices := []string{
		"systemd-networkd",
		"sshd",
		"dbus",
		"systemd-resolved",
		"systemd-timesyncd",
	}

	for _, service := range criticalServices {
		output, err := executor.ExecuteCommand("systemctl", "is-active", service)
		if err != nil || strings.TrimSpace(output) != "active" {
			issues = append(issues, fmt.Sprintf("Service **%s** is not active: %s", service, strings.TrimSpace(output)))
		} else {
			status = append(status, fmt.Sprintf("✅ **Service %s**: Active", service))
		}
	}

	// Check failed services
	output, err := executor.ExecuteCommand("systemctl", "list-units", "--failed", "--no-pager")
	if err == nil && strings.Contains(output, "0 loaded units listed") {
		status = append(status, "✅ **Failed Services**: None detected")
	} else if err == nil {
		issues = append(issues, fmt.Sprintf("Failed services detected:\n```\n%s\n```", output))
	}

	return status, issues
}

// checkDiskSpace checks available disk space
func checkDiskSpace(executor *nixos.Executor) ([]string, []string) {
	var status []string
	var warnings []string

	output, err := executor.ExecuteCommand("df", "-h", "/", "/nix", "/tmp", "/var")
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not check disk space: %v", err))
		return status, warnings
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			mountPoint := fields[5]
			usage := fields[4]

			// Parse usage percentage
			usageStr := strings.TrimSuffix(usage, "%")
			if usageStr != usage { // Has % sign
				if usageStr >= "90" {
					warnings = append(warnings, fmt.Sprintf("**%s**: High disk usage (%s)", mountPoint, usage))
				} else if usageStr >= "80" {
					warnings = append(warnings, fmt.Sprintf("**%s**: Moderate disk usage (%s)", mountPoint, usage))
				} else {
					status = append(status, fmt.Sprintf("✅ **%s**: Good disk space (%s used)", mountPoint, usage))
				}
			}
		}
	}

	return status, warnings
}

// checkNixChannels checks the status of Nix channels
func checkNixChannels(executor *nixos.Executor) ([]string, []string) {
	var status []string
	var issues []string

	// Check system channels
	output, err := executor.ExecuteCommand("sudo", "nix-channel", "--list")
	if err != nil {
		issues = append(issues, fmt.Sprintf("Could not list system channels: %v", err))
	} else if strings.TrimSpace(output) == "" {
		issues = append(issues, "No system channels configured")
	} else {
		status = append(status, "✅ **System Channels**: Configured")
		// Check if channels are up to date
		updateOutput, updateErr := executor.ExecuteCommand("sudo", "nix-channel", "--update", "--dry-run")
		if updateErr == nil && !strings.Contains(updateOutput, "updating") {
			status = append(status, "✅ **System Channels**: Up to date")
		}
	}

	// Check user channels
	userOutput, err := executor.ExecuteCommand("nix-channel", "--list")
	if err == nil && strings.TrimSpace(userOutput) != "" {
		status = append(status, "✅ **User Channels**: Configured")
	}

	return status, issues
}

// checkBootSystem checks boot system integrity
func checkBootSystem(executor *nixos.Executor) ([]string, []string) {
	var status []string
	var issues []string

	// Check if current generation is bootable
	_, err := executor.ExecuteCommand("sudo", "nixos-rebuild", "boot", "--dry-run")
	if err != nil {
		issues = append(issues, fmt.Sprintf("Boot configuration check failed: %v", err))
	} else {
		status = append(status, "✅ **Boot System**: Configuration is bootable")
	}

	// Check available generations
	genOutput, err := executor.ExecuteCommand("sudo", "nix-env", "--list-generations", "-p", "/nix/var/nix/profiles/system")
	if err == nil {
		generations := strings.Split(strings.TrimSpace(genOutput), "\n")
		if len(generations) > 0 {
			status = append(status, fmt.Sprintf("✅ **System Generations**: %d available", len(generations)))
		}
	}

	return status, issues
}

// checkNetworkConnectivity checks basic network connectivity
func checkNetworkConnectivity() ([]string, []string) {
	var status []string
	var issues []string

	// Test basic connectivity
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get("https://cache.nixos.org")
	if err != nil {
		issues = append(issues, fmt.Sprintf("Cannot reach NixOS cache: %v", err))
	} else {
		status = append(status, "✅ **Network**: NixOS cache reachable")
	}

	_, err = client.Get("https://channels.nixos.org")
	if err != nil {
		issues = append(issues, fmt.Sprintf("Cannot reach NixOS channels: %v", err))
	} else {
		status = append(status, "✅ **Network**: NixOS channels reachable")
	}

	return status, issues
}

// checkNixStore checks Nix store health
func checkNixStore(executor *nixos.Executor) ([]string, []string) {
	var status []string
	var issues []string

	// Check store integrity
	output, err := executor.ExecuteCommand("nix-store", "--verify", "--check-contents")
	if err != nil {
		issues = append(issues, fmt.Sprintf("Nix store verification failed: %v", err))
	} else if strings.Contains(output, "error") {
		issues = append(issues, fmt.Sprintf("Nix store has integrity issues:\n```\n%s\n```", output))
	} else {
		status = append(status, "✅ **Nix Store**: Integrity verified")
	}

	// Check for garbage collection recommendations
	output, err = executor.ExecuteCommand("nix-store", "--gc", "--dry-run")
	if err == nil && strings.Contains(output, "bytes would be freed") {
		// Extract the amount that would be freed
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMGT]B)?\s*bytes would be freed`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 0 {
			status = append(status, fmt.Sprintf("💡 **Garbage Collection**: %s can be freed", matches[0]))
		}
	}

	return status, issues
}

// generateHealthReport creates a formatted health report
func generateHealthReport(healthReport, warnings, issues []string) string {
	var report strings.Builder

	report.WriteString(utils.FormatHeader("🏥 NixOS System Health Check Report"))
	report.WriteString("\n\n")

	// Overall status
	if len(issues) == 0 && len(warnings) == 0 {
		report.WriteString(utils.FormatSuccess("🎉 System is healthy! No issues detected."))
	} else if len(issues) == 0 {
		report.WriteString(utils.FormatInfo("ℹ️ System is mostly healthy with some minor warnings."))
	} else {
		report.WriteString(utils.FormatWarning("⚠️ System has issues that need attention."))
	}

	report.WriteString("\n\n")

	// Health status section
	if len(healthReport) > 0 {
		report.WriteString(utils.FormatSection("✅ Health Status", strings.Join(healthReport, "\n")))
		report.WriteString("\n\n")
	}

	// Warnings section
	if len(warnings) > 0 {
		report.WriteString(utils.FormatSection("⚠️ Warnings", strings.Join(warnings, "\n")))
		report.WriteString("\n\n")
	}

	// Issues section
	if len(issues) > 0 {
		report.WriteString(utils.FormatSection("❌ Issues Requiring Attention", strings.Join(issues, "\n")))
		report.WriteString("\n\n")
	}

	// Add helpful tips
	tips := []string{
		"Run `sudo nixos-rebuild switch` to apply pending configuration changes",
		"Use `nix-store --gc` to free up disk space by removing unused packages",
		"Check system logs with `journalctl -xe` for detailed error information",
		"Update channels with `sudo nix-channel --update` for latest packages",
	}

	report.WriteString(utils.FormatSection("💡 General Tips", strings.Join(tips, "\n")))

	return report.String()
}

// getAIHealthAnalysis gets AI analysis of health issues
func getAIHealthAnalysis(cfg *config.UserConfig, issues, warnings []string) (string, error) {
	var analysisInput strings.Builder

	analysisInput.WriteString("NixOS System Health Check Analysis Request:\n\n")

	if len(issues) > 0 {
		analysisInput.WriteString("ISSUES DETECTED:\n")
		for i, issue := range issues {
			analysisInput.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
		analysisInput.WriteString("\n")
	}

	if len(warnings) > 0 {
		analysisInput.WriteString("WARNINGS:\n")
		for i, warning := range warnings {
			analysisInput.WriteString(fmt.Sprintf("%d. %s\n", i+1, warning))
		}
		analysisInput.WriteString("\n")
	}

	prompt := `You are a NixOS system administrator expert. Analyze the health check results above and provide:

1. **Root Cause Analysis**: Explain what's causing each issue
2. **Priority Assessment**: Rank issues by urgency (Critical/High/Medium/Low)
3. **Step-by-Step Solutions**: Provide specific commands and actions to fix each issue
4. **Prevention Tips**: How to avoid these issues in the future
5. **Related Documentation**: Mention relevant NixOS wiki pages or manual sections

Format your response in clear Markdown with proper headings and code blocks for commands.
Be specific and actionable. Focus on NixOS-specific solutions.`

	// Select AI provider based on config
	var provider ai.AIProvider
	switch cfg.AIProvider {
	case "ollama":
		provider = ai.NewOllamaProvider(cfg.AIModel)
	case "gemini":
		provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://api.gemini.com")
	case "openai":
		provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	default:
		provider = ai.NewOllamaProvider("llama3")
	}

	response, err := provider.Query(analysisInput.String() + "\n\n" + prompt)
	if err != nil {
		return "", fmt.Errorf("AI analysis failed: %w", err)
	}

	return response, nil
}
