package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"

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

	diagnoseCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Path to a log file to analyze")
	diagnoseCmd.Flags().StringVarP(&configSnippet, "config-snippet", "c", "", "NixOS configuration snippet to analyze")
	diagnoseCmd.Flags().StringVarP(&nixLogTarget, "nix-log", "g", "", "Run 'nix log' (optionally with a path or derivation) and analyze the output") // New flag
	searchCmd.Flags().StringVarP(&nixosConfigPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.PersistentFlags().StringVarP(&nixosConfigPathGlobal, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	configCmd.AddCommand(showUserConfig)
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
