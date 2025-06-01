package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/neovim"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/internal/packaging"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
	"nix-ai-help/pkg/version"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Helper function to get a valid Ollama model name
func getOllamaModel(configModel string) string {
	if configModel != "" {
		return configModel
	}
	return "llama3" // Default model
}

// filterDocumentationContent filters out HTML tags, wiki navigation, and other unwanted content
func filterDocumentationContent(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip HTML content
		if strings.Contains(trimmed, "<!DOCTYPE") ||
			strings.Contains(trimmed, "<html") ||
			strings.Contains(trimmed, "<head") ||
			strings.Contains(trimmed, "</head>") ||
			strings.Contains(trimmed, "<body") ||
			strings.Contains(trimmed, "</body>") ||
			strings.Contains(trimmed, "</html>") ||
			strings.Contains(trimmed, "<meta") ||
			strings.Contains(trimmed, "<link") ||
			strings.Contains(trimmed, "<script") ||
			strings.Contains(trimmed, "</script>") ||
			strings.Contains(trimmed, "<style") ||
			strings.Contains(trimmed, "</style>") {
			continue
		}

		// Skip wiki navigation and interface elements
		if strings.Contains(trimmed, "Navigation") ||
			strings.Contains(trimmed, "Edit this page") ||
			strings.Contains(trimmed, "Recent changes") ||
			strings.Contains(trimmed, "Random page") ||
			strings.Contains(trimmed, "Special pages") ||
			strings.Contains(trimmed, "Main Page") ||
			strings.Contains(trimmed, "Community portal") ||
			strings.Contains(trimmed, "Help") ||
			strings.Contains(trimmed, "Search") ||
			strings.Contains(trimmed, "Go") ||
			strings.Contains(trimmed, "Tools") ||
			strings.Contains(trimmed, "Print/export") ||
			strings.Contains(trimmed, "In other languages") {
			continue
		}

		// Skip lines that are just URL references to wiki with HTML content
		if strings.HasPrefix(trimmed, "https://wiki.nixos.org/wiki/") &&
			(strings.Contains(trimmed, "<!DOCTYPE") || strings.Contains(trimmed, "<head") ||
				strings.Contains(trimmed, "Navigation") || strings.Contains(trimmed, "<")) {
			continue
		}

		// Remove HTML tags from remaining content
		// Basic HTML tag removal regex
		re := regexp.MustCompile(`<[^>]*>`)
		cleaned := re.ReplaceAllString(trimmed, "")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned != "" {
			filtered = append(filtered, cleaned)
		}
	}

	return strings.Join(filtered, "\n")
}

// Command structure for the CLI
var rootCmd = &cobra.Command{
	Use:     "nixai [question]",
	Short:   "NixAI helps solve Nix configuration problems",
	Version: version.Version,
	Long: `NixAI is a command-line tool that assists users in diagnosing and solving NixOS configuration issues using AI models and documentation queries.

You can also ask questions directly, e.g.:
  nixai "how can I configure curl?"`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if --ask flag was used
		askFlag, _ := cmd.Flags().GetString("ask")

		var question string
		if askFlag != "" {
			// Use the question from the --ask flag
			question = askFlag
		} else if len(args) > 0 {
			// Join all arguments into a single question
			question = strings.Join(args, " ")
		} else {
			// If no arguments provided and no --ask flag, show help
			cmd.Help()
			return
		}

		fmt.Println(utils.FormatHeader("🤖 NixAI Question Assistant"))
		fmt.Println(utils.FormatKeyValue("Question", question))
		fmt.Println(utils.FormatDivider())

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading user config: "+err.Error()))
			os.Exit(1)
		}

		// Initialize AI provider
		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			provider = ai.NewOllamaProvider(getOllamaModel(cfg.AIModel))
		case "gemini":
			provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			provider = ai.NewOllamaProvider(getOllamaModel(""))
		}

		// Check MCP server status and query documentation if available
		fmt.Print(utils.FormatProgress("Checking documentation server..."))
		mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
		var docContext string

		statusResp, err := http.Get(mcpURL + "/healthz")
		if err == nil && statusResp.StatusCode == 200 {
			fmt.Print(utils.FormatProgress("Querying NixOS documentation..."))

			// Query MCP server for relevant documentation
			queryURL := fmt.Sprintf("%s/query?q=%s", mcpURL, question)
			docResp, err := http.Get(queryURL)
			if err == nil && docResp.StatusCode == 200 {
				defer docResp.Body.Close()
				body, err := io.ReadAll(docResp.Body)
				if err == nil && len(body) > 50 {
					docContext = string(body)
					fmt.Println(utils.FormatSuccess("Documentation retrieved"))
				}
			}
			statusResp.Body.Close()
		}

		// Prepare AI prompt with documentation context
		prompt := fmt.Sprintf(`You are a NixOS expert assistant. Please provide a helpful, accurate answer to this question about NixOS configuration:

Question: %s

Please provide:
1. A clear, practical answer
2. Relevant configuration examples if applicable  
3. Any important considerations or best practices
4. Links to official documentation when relevant

Keep the response concise but comprehensive.`, question)

		if docContext != "" {
			prompt += fmt.Sprintf(`

Here is relevant documentation context to help inform your answer:

%s`, docContext)
		}

		// Get AI response
		fmt.Print(utils.FormatProgress("Generating answer..."))
		response, err := provider.Query(prompt)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Failed to generate response: "+err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("Response generated"))
		fmt.Println()

		// Format and display the response
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		if err != nil {
			fmt.Println(response)
		} else {
			formatted, err := renderer.Render(response)
			if err != nil {
				fmt.Println(response)
			} else {
				fmt.Print(formatted)
			}
		}
	},
}

var logFile string
var configSnippet string
var nixosConfigPath string
var nixLogTarget string          // New: for --nix-log flag
var nixosConfigPathGlobal string // Global path for build/flake context
var askQuestion string           // New: for --ask/-a flag

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
	rootCmd.AddCommand(findOptionCmd)        // Register the find-option command
	rootCmd.AddCommand(mcpServerCmd)         // Register the MCP server command
	rootCmd.AddCommand(healthCheckCmd)       // Register the health check command
	rootCmd.AddCommand(decodeErrorCmd)       // Register the decode-error command
	rootCmd.AddCommand(upgradeAdvisorCmd)    // Register the upgrade-advisor command
	rootCmd.AddCommand(serviceExamplesCmd)   // Register the service-examples command
	rootCmd.AddCommand(lintConfigCmd)        // Register the lint-config command
	rootCmd.AddCommand(explainHomeOptionCmd) // Register the explain-home-option command
	rootCmd.AddCommand(packageRepoCmd)       // Register the package-repo command
	rootCmd.AddCommand(versionCmd)           // Register the version command
	rootCmd.AddCommand(migrateCmd)           // Register the migrate command

	diagnoseCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Path to a log file to analyze")
	diagnoseCmd.Flags().StringVarP(&configSnippet, "config-snippet", "c", "", "NixOS configuration snippet to analyze")
	diagnoseCmd.Flags().StringVarP(&nixLogTarget, "nix-log", "g", "", "Run 'nix log' (optionally with a path or derivation) and analyze the output") // New flag
	decodeErrorCmd.Flags().StringP("log-file", "l", "", "Path to a log file containing error messages to analyze")
	searchCmd.Flags().StringVarP(&nixosConfigPath, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.PersistentFlags().StringVarP(&nixosConfigPathGlobal, "nixos-path", "n", "", "Path to your NixOS configuration folder (containing flake.nix or configuration.nix)")
	rootCmd.Flags().StringVarP(&askQuestion, "ask", "a", "", "Ask a question about NixOS configuration")

	// Version command flags
	versionCmd.Flags().BoolP("json", "j", false, "Output version information in JSON format")
	versionCmd.Flags().BoolP("short", "s", false, "Output only the version number")
	configCmd.AddCommand(showUserConfig)
	mcpServerCmd.AddCommand(mcpServerStartCmd)
	mcpServerCmd.AddCommand(mcpServerStopCmd)
	mcpServerCmd.AddCommand(mcpServerStatusCmd)
	mcpServerStartCmd.Flags().BoolP("background", "d", false, "Run MCP server in background (daemon mode)")
	mcpServerStartCmd.Flags().Bool("daemon", false, "Alias for --background")
	mcpServerStartCmd.Flags().String("socket-path", "", "Custom path for the MCP server Unix socket")
	mcpServerStartCmd.Flags().String("config", "", "Path to custom config file")

	// Package repository command flags
	packageRepoCmd.Flags().StringP("local", "l", "", "Local path to repository (instead of cloning)")
	packageRepoCmd.Flags().StringP("output", "o", "", "Output directory for generated derivation file")
	packageRepoCmd.Flags().String("name", "", "Override package name")
	packageRepoCmd.Flags().BoolP("quiet", "q", false, "Suppress progress output")
	packageRepoCmd.Flags().Bool("analyze-only", false, "Only analyze repository without generating derivation")

	// Add neovim integration command
	rootCmd.AddCommand(neovimCmd)

	// Add neovim command flags
	neovimCmd.Flags().String("socket-path", "", "Custom path for the MCP server Unix socket")
	neovimCmd.Flags().String("config-dir", "", "Custom path for the Neovim configuration directory")
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
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
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
		fmt.Println(utils.FormatHeader("📦 Selected " + strings.ToUpper(searchType[:1]) + searchType[1:]))
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
	Use:   "config [show|set|unset|edit|explain|analyze|validate|optimize|backup|restore] [key] [value]",
	Short: "AI-assisted Nix configuration management",
	Long: `Manage and understand your Nix configuration with AI-powered help and intelligent recommendations.

Commands:
  show                     - Show current Nix configuration with AI analysis
  set <key> <value>       - Set configuration option with AI guidance  
  unset <key>             - Unset configuration option with safety checks
  edit                    - Open configuration in editor with AI tips
  explain <key>           - AI-powered explanation of configuration options
  analyze                 - Comprehensive analysis of current configuration
  validate                - Validate configuration and suggest improvements
  optimize                - AI recommendations for performance optimization
  backup                  - Create backup of current configuration
  restore <backup>        - Restore configuration from backup
  
Examples:
  nixai config show                              # Show and analyze current config
  nixai config set experimental-features "nix-command flakes"
  nixai config explain substituters             # Get AI explanation of option
  nixai config analyze                          # Full configuration analysis
  nixai config validate                         # Validate and suggest improvements
  nixai config optimize                         # Performance optimization tips`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration and initialize AI provider
		cfg, err := config.LoadUserConfig()
		var provider ai.AIProvider
		if err == nil {
			switch cfg.AIProvider {
			case "ollama":
				provider = ai.NewOllamaProvider(getOllamaModel(cfg.AIModel))
			case "gemini":
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
			case "openai":
				provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			default:
				provider = ai.NewOllamaProvider("llama3")
			}
		} else {
			provider = ai.NewOllamaProvider("llama3")
		}

		// Handle different subcommands
		if len(args) == 0 || args[0] == "show" {
			handleConfigShow(provider)
			return
		}

		switch args[0] {
		case "set":
			handleConfigSet(args, provider)
		case "unset":
			handleConfigUnset(args, provider)
		case "edit":
			handleConfigEdit(provider)
		case "explain":
			handleConfigExplain(args, provider)
		case "analyze":
			handleConfigAnalyze(provider)
		case "validate":
			handleConfigValidate(provider)
		case "optimize":
			handleConfigOptimize(provider)
		case "backup":
			handleConfigBackup()
		case "restore":
			handleConfigRestore(args)
		default:
			fmt.Println(utils.FormatError("Unknown config command: " + args[0]))
			fmt.Println(utils.FormatInfo("Run 'nixai config --help' for available commands"))
		}
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
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
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
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
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
			provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
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
		daemon, _ := cmd.Flags().GetBool("daemon")
		socketPath, _ := cmd.Flags().GetString("socket-path")
		configPath, _ := cmd.Flags().GetString("config")

		// If either background or daemon flag is set, run in background
		background = background || daemon

		// Check for daemon flag as alias for background
		daemonFlag, _ := cmd.Flags().GetBool("daemon")
		if daemonFlag {
			background = true
		}

		// Use default config path if not specified
		if configPath == "" {
			configPath = "configs/default.yaml"
		}

		// Check if NIXAI_SOCKET_PATH environment variable is set
		if socketPath == "" {
			envSocketPath := os.Getenv("NIXAI_SOCKET_PATH")
			if envSocketPath != "" {
				socketPath = envSocketPath
			}
		}

		if background {
			absPath, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not determine nixai binary path: %v\n", err)
				os.Exit(1)
			}

			// Build command with any custom flags
			startCmd := fmt.Sprintf("%s mcp-server start", absPath)
			if socketPath != "" {
				startCmd += fmt.Sprintf(" --socket-path=\"%s\"", socketPath)
			}
			if configPath != "configs/default.yaml" {
				startCmd += fmt.Sprintf(" --config=\"%s\"", configPath)
			}

			cmdStr := fmt.Sprintf("nohup %s > mcp.log 2>&1 &", startCmd)
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
		server, err := mcp.NewServerFromConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create MCP server: %v\n", err)
			os.Exit(1)
		}

		// Override socket path from command line if provided
		if socketPath != "" {
			server.SetSocketPath(socketPath)
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

// isHomeManagerOption determines if documentation refers to a Home Manager option
func isHomeManagerOption(doc string) bool {
	// Strong indicators of Home Manager documentation
	if strings.Contains(doc, "home-manager-options.extranix.com") ||
		strings.Contains(doc, "Location:") ||
		strings.Contains(doc, "home-manager/modules") ||
		strings.Contains(doc, "nix-community.github.io/home-manager") {
		return true
	}

	// Check for Home Manager context keywords
	if strings.Contains(doc, "Home Manager") {
		return true
	}

	// Check for common Home Manager option prefixes but be more specific
	// Only consider it Home Manager if it starts with home. prefix
	if strings.Contains(doc, "home.") {
		return true
	}

	// For programs.* options, we need more context to distinguish
	// NixOS programs.* from Home Manager programs.*
	if strings.Contains(doc, "programs.") {
		// If it mentions nixos modules path, it's likely NixOS
		if strings.Contains(doc, "nixos/modules") {
			return false
		}
		// If it mentions user-specific configuration, it might be Home Manager
		if strings.Contains(doc, "user configuration") ||
			strings.Contains(doc, "user's home") ||
			strings.Contains(doc, "per-user") {
			return true
		}
	}

	return false
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
		client := http.Client{Timeout: 5 * time.Second}
		statusResp, err := client.Get(mcpURL + "/healthz")
		if err != nil {
			fmt.Println(utils.FormatError("MCP server is not running"))
			fmt.Println(utils.FormatInfo("Please start it with 'nixai mcp-server start' or 'nixai mcp-server start -d'"))
			os.Exit(1)
		}
		if statusResp.StatusCode != 200 {
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

		// Determine if this is a Home Manager or NixOS option
		isHomeManager := isHomeManagerOption(doc)

		// Display appropriate header and info box
		if strings.Contains(doc, "Option:") {
			if isHomeManager {
				fmt.Println(utils.FormatHeader("🏠 Home Manager Option"))
				fmt.Println(utils.FormatBox("Home Manager Option", "This option is managed by Home Manager. See: https://nix-community.github.io/home-manager/options.html"))
			} else {
				fmt.Println(utils.FormatHeader("🖥️ NixOS Option"))
				fmt.Println(utils.FormatBox("NixOS Option", "This option is managed by NixOS. See: https://search.nixos.org/options"))
			}
		}

		// Filter and display the documentation
		filteredDoc := filterDocumentationContent(doc)
		if strings.TrimSpace(filteredDoc) != "" {
			fmt.Println(utils.FormatHeader("📋 Documentation"))
			fmt.Println(utils.FormatDivider())
			// Render the documentation as markdown for better formatting
			rendered, err := glamour.Render(filteredDoc, "dark")
			if err != nil {
				fmt.Println(filteredDoc)
			} else {
				fmt.Print(rendered)
			}
			fmt.Println(utils.FormatDivider())
		}

		// Select AI provider
		fmt.Print(utils.FormatProgress("Generating explanation with " + cfg.AIProvider + "..."))
		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			provider = ai.NewOllamaProvider(getOllamaModel(cfg.AIModel))
		case "gemini":
			provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
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

		// Create dynamic header based on option type
		var explanationHeader string
		if isHomeManager {
			explanationHeader = "🏠 Home Manager Option Explanation"
		} else {
			explanationHeader = "🖥️ NixOS Option Explanation"
		}
		fmt.Println("\n" + utils.FormatHeader(explanationHeader))
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

// explainHomeOptionCmd provides explanations for Home Manager options
var explainHomeOptionCmd = &cobra.Command{
	Use:   "explain-home-option <option>",
	Short: "Explain a Home Manager option using AI and documentation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		option := args[0]
		fmt.Println(utils.FormatProgress("Analyzing Home Manager option: " + option))

		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading user config: "+err.Error()))
			os.Exit(1)
		}

		// Check MCP server status before querying
		fmt.Print(utils.FormatProgress("Fetching official documentation..."))
		mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
		client := http.Client{Timeout: 5 * time.Second}
		statusResp, err := client.Get(mcpURL + "/healthz")
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
			return
		}
		fmt.Println(utils.FormatSuccess("Documentation found!"))

		// After fetching documentation from MCP, determine the source and label the result
		if strings.Contains(doc, "Option:") {
			if strings.Contains(doc, "Location:") || strings.Contains(doc, "home-manager-options.extranix.com") || strings.Contains(doc, "Home Manager") || strings.Contains(doc, "programs.") || strings.Contains(doc, "home.") {
				fmt.Println(utils.FormatHeader("🏠 Home Manager Option"))
				fmt.Println(utils.FormatBox("Home Manager Option", "This option is managed by Home Manager. See: https://nix-community.github.io/home-manager/options.html"))
			} else {
				fmt.Println(utils.FormatHeader("🖥️ NixOS Option"))
				fmt.Println(utils.FormatBox("NixOS Option", "This option is managed by NixOS. See: https://search.nixos.org/options"))
			}
		}

		// Filter and display the documentation
		filteredDoc := filterDocumentationContent(doc)
		if strings.TrimSpace(filteredDoc) != "" {
			fmt.Println(utils.FormatHeader("📋 Documentation"))
			fmt.Println(utils.FormatDivider())
			// Render the documentation as markdown for better formatting
			rendered, err := glamour.Render(filteredDoc, "dark")
			if err != nil {
				fmt.Println(filteredDoc)
			} else {
				fmt.Print(rendered)
			}
			fmt.Println(utils.FormatDivider())
		}

		// Select AI provider
		fmt.Print(utils.FormatProgress("Generating explanation with " + cfg.AIProvider + "..."))
		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			provider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			provider = ai.NewOllamaProvider("llama3")
		}

		response, err := provider.Query("Explain this Home Manager option in detail, including type, default, best practices, and usage examples.\n" + doc)
		if err != nil {
			fmt.Println(utils.FormatError("Error getting AI response: " + err.Error()))
			os.Exit(1)
		}

		// Render the response with markdown formatting
		rendered := renderForTerminal(response)
		fmt.Println(rendered)
	},
}

// Find option command for reverse NixOS option lookup
var findOptionCmd = &cobra.Command{
	Use:   "find-option <description>",
	Short: "Find NixOS options from natural language description",
	Long: `Find relevant NixOS options and configuration snippets based on what you want to achieve.
Describe your goal in plain English and get suggested options, examples, and documentation.

Examples:
  nixai find-option "enable SSH access"
  nixai find-option "configure firewall"
  nixai find-option "set up automatic updates"
  nixai find-option "enable docker"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")

		fmt.Println(utils.FormatHeader("🔍 Finding NixOS Options"))
		fmt.Println(utils.FormatKeyValue("Looking for", description))
		fmt.Println(utils.FormatDivider())

		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading user config: "+err.Error()))
			os.Exit(1)
		}

		// Select AI provider
		fmt.Print(utils.FormatProgress("Analyzing your request with AI..."))
		var provider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			provider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			provider = ai.NewOllamaProvider("llama3")
		}

		prompt := buildFindOptionPrompt(description)
		aiResp, err := provider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("AI error: " + err.Error()))
			os.Exit(1)
		}
		if strings.TrimSpace(aiResp) == "" {
			fmt.Println("\n" + utils.FormatError("AI did not return any suggestions."))
			return
		}
		fmt.Println(" ✓")

		fmt.Println("\n" + utils.FormatHeader("💡 Suggested NixOS Options"))
		fmt.Println(utils.FormatDivider())

		// Render output as markdown in terminal with enhanced styling
		out, err := glamour.Render(aiResp, "dark")
		if err != nil {
			// Fallback to plain text with basic formatting
			fmt.Println(utils.FormatSection("Suggestions", aiResp))
		} else {
			fmt.Print(out)
		}

		fmt.Println("\n" + utils.FormatDivider())

		// Enhanced tips section
		fmt.Println(utils.FormatTip("Use 'nixai explain-option <option>' to learn more about specific options"))
		fmt.Println(utils.FormatTip("Use 'nixai search service <name>' to find related services"))
		fmt.Println(utils.FormatNote("Always test configuration changes in a safe environment first"))
	},
}

// buildFindOptionPrompt creates a comprehensive prompt for AI to find relevant NixOS options
func buildFindOptionPrompt(description string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users find the right configuration options for their needs. 

**User Request:** "%s"

Please provide a helpful response with:

1. **Primary Options**: The main NixOS option(s) that address this request
2. **Configuration Examples**: Complete, working configuration snippets
3. **Related Options**: Additional options that work well together or provide enhanced functionality
4. **Best Practices**: Important tips, warnings, or recommendations
5. **Alternative Approaches**: If applicable, mention different ways to achieve the same goal

**Format your response using Markdown with:**
- Clear headings (##)
- Code blocks for configuration examples (use `+"`"+`nix code blocks`+"`"+`)
- Bullet points for lists
- **Bold** text for emphasis on option names

**Important Guidelines:**
- Provide complete, working configuration examples
- Focus on the most common and recommended approaches
- Include any security considerations or best practices
- Mention if the option requires system rebuilds or service restarts
- If the request is vague, provide multiple relevant options

Make it practical and actionable for someone configuring their NixOS system.`, description)
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
		provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
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

// Decode Error command - specialized AI-driven error analysis
var decodeErrorCmd = &cobra.Command{
	Use:   "decode-error [error_message]",
	Short: "AI-powered NixOS error decoder and fix generator",
	Long: `Analyze NixOS error messages using AI to provide detailed explanations,
root cause analysis, and step-by-step fix instructions.

The command can accept error messages in multiple ways:
  - As a command line argument
  - Through stdin (pipe or redirect)
  - From a log file using --log-file
  - Interactively when no input is provided

Examples:
  nixai decode-error "syntax error at line 42"
  journalctl -u nginx | nixai decode-error
  nixai decode-error --log-file /var/log/nixos-rebuild.log
  nixos-rebuild switch 2>&1 | nixai decode-error`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var errorInput string

		// 1. Check for log file input
		logFile, _ := cmd.Flags().GetString("log-file")
		if logFile != "" {
			data, err := os.ReadFile(logFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read log file: %v\n", err)
				os.Exit(1)
			}
			errorInput = tailLines(string(data), 100) // Get last 100 lines
		}

		// 2. Check for stdin input
		if errorInput == "" {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				stdinData, _ := os.ReadFile("/dev/stdin")
				errorInput = tailLines(string(stdinData), 100)
			}
		}

		// 3. Check for command line arguments
		if errorInput == "" && len(args) > 0 {
			errorInput = strings.Join(args, " ")
		}

		// 4. Interactive mode if no input provided
		if errorInput == "" {
			fmt.Println(utils.FormatHeader("🔍 NixOS Error Decoder"))
			fmt.Println(utils.FormatInfo("Please paste your NixOS error message below (press Ctrl+D when done):"))
			fmt.Println()

			var lines []string
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			if len(lines) == 0 {
				fmt.Fprintln(os.Stderr, utils.FormatError("No error message provided"))
				os.Exit(1)
			}

			errorInput = strings.Join(lines, "\n")
		}

		if strings.TrimSpace(errorInput) == "" {
			fmt.Fprintln(os.Stderr, utils.FormatError("No error message to analyze"))
			os.Exit(1)
		}

		// Load configuration and initialize AI provider
		fmt.Print(utils.FormatProgress("Loading configuration..."))
		cfg, err := config.LoadUserConfig()
		var provider ai.AIProvider
		if err == nil {
			switch cfg.AIProvider {
			case "ollama":
				provider = ai.NewOllamaProvider(cfg.AIModel)
			case "gemini":
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
			case "openai":
				provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			default:
				provider = ai.NewOllamaProvider("llama3")
			}
		} else {
			provider = ai.NewOllamaProvider("llama3")
		}
		fmt.Println(" ✓")

		// Perform enhanced diagnostics
		fmt.Print(utils.FormatProgress("Analyzing error patterns..."))
		diagnostics := nixos.Diagnose(errorInput, "", provider)
		fmt.Println(" ✓")

		// Display results with enhanced formatting
		fmt.Println()
		result := nixos.FormatDiagnostics(diagnostics)
		fmt.Print(result)

		// Provide additional context and suggestions
		if len(diagnostics) > 0 {
			fmt.Println()
			fmt.Println(utils.FormatHeader("💡 Additional Recommendations"))

			// Check for common follow-up actions
			hasHighSeverity := false
			for _, diag := range diagnostics {
				if diag.Severity == "high" || diag.Severity == "critical" {
					hasHighSeverity = true
					break
				}
			}

			if hasHighSeverity {
				fmt.Println(utils.FormatWarning("High severity issues detected - prioritize fixing these first"))
			}

			fmt.Println(utils.FormatTip("Use 'nixai explain-option <option>' to understand specific NixOS options"))
			fmt.Println(utils.FormatTip("Run 'nixai health-check' for a comprehensive system analysis"))
			fmt.Println(utils.FormatNote("Join the NixOS community for additional help: https://discourse.nixos.org/"))
		}
	},
}

// upgradeAdvisorCmd provides AI-powered NixOS upgrade guidance
var upgradeAdvisorCmd = &cobra.Command{
	Use:   "upgrade-advisor",
	Short: "Get AI-powered guidance for upgrading NixOS",
	Long: `Analyze your current NixOS system and get comprehensive guidance for upgrading to newer versions.

This command performs:
- System health checks and compatibility analysis
- Available upgrade options and recommendations  
- Pre-upgrade backup advice and preparation steps
- Step-by-step upgrade instructions with AI explanations
- Post-upgrade validation checklist

Examples:
  nixai upgrade-advisor                    # Get comprehensive upgrade analysis
  nixai upgrade-advisor --target 24.11    # Get guidance for specific version
  nixai upgrade-advisor --dry-run          # Show analysis without recommendations`,
	Run: func(cmd *cobra.Command, args []string) {
		targetVersion, _ := cmd.Flags().GetString("target")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Display header
		fmt.Println(utils.FormatHeader("🚀 NixOS Upgrade Advisor"))
		fmt.Println()

		// Load configuration and initialize AI provider
		fmt.Print(utils.FormatProgress("Loading configuration and initializing AI..."))
		cfg, err := config.LoadUserConfig()
		var provider ai.AIProvider
		if err == nil {
			switch cfg.AIProvider {
			case "ollama":
				provider = ai.NewOllamaProvider(cfg.AIModel)
			case "gemini":
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
			case "openai":
				provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			default:
				provider = ai.NewOllamaProvider("llama3")
			}
		} else {
			provider = ai.NewOllamaProvider("llama3")
		}
		fmt.Println(" ✓")

		// Validate NixOS configuration path before proceeding
		fmt.Print(utils.FormatProgress("Validating NixOS configuration path..."))
		var configPath string
		if nixosConfigPathGlobal != "" {
			configPath = nixosConfigPathGlobal
		} else if cfg != nil && cfg.NixosFolder != "" {
			configPath = utils.ExpandHome(cfg.NixosFolder)
		}

		if configPath == "" || !utils.IsDirectory(configPath) {
			fmt.Printf("\n%s NixOS config path is not set or does not exist: '%s'\n",
				utils.FormatError("Error:"), configPath)
			fmt.Println(utils.FormatNote("The upgrade advisor needs access to your NixOS configuration to:"))
			fmt.Println("  • Validate current configuration compatibility")
			fmt.Println("  • Check for potential upgrade issues")
			fmt.Println("  • Analyze system-specific requirements")
			fmt.Println()
			fmt.Println(utils.FormatTip("Set the config path using one of these methods:"))
			fmt.Println("  • CLI flag: --nixos-path /etc/nixos")
			fmt.Println("  • Config file: Edit ~/.config/nixai/config.yaml")
			fmt.Println("  • Interactive: Run 'nixai interactive' and use 'set-nixos-path'")
			fmt.Println()
			fmt.Println(utils.FormatWarning("Common paths:"))
			fmt.Println("  • /etc/nixos (traditional)")
			fmt.Println("  • ~/nixos-config (flake-based)")
			fmt.Println("  • ~/.config/nixos (user configuration)")
			os.Exit(1)
		}
		fmt.Println(" ✓")

		// Initialize upgrade advisor with validated path
		log := logger.NewLogger()
		upgradeAdvisor := nixos.NewUpgradeAdvisorWithConfig(*log, configPath)

		// Analyze upgrade options with detailed progress
		fmt.Println(utils.FormatProgress("🔍 Analyzing current system..."))

		// Step 1: System Information
		fmt.Print("  • [1/7] Detecting NixOS version and channel...")
		upgradeInfo, err := upgradeAdvisor.AnalyzeUpgradeOptions(cmd.Context())
		if err != nil {
			fmt.Printf("\n%s Failed to analyze upgrade options: %v\n", utils.FormatError("Error:"), err)
			os.Exit(1)
		}
		fmt.Println(" ✓")

		// Step 2: Channel Analysis
		fmt.Print("  • [2/7] Scanning available upgrade channels...")
		fmt.Println(" ✓")

		// Step 3: System Health Checks
		fmt.Print("  • [3/7] Running system health checks...")
		fmt.Println(" ✓")

		// Step 4: Disk Space Analysis
		fmt.Print("  • [4/7] Analyzing disk space requirements...")
		fmt.Println(" ✓")

		// Step 5: Configuration Validation
		fmt.Print("  • [5/7] Validating current configuration...")
		fmt.Println(" ✓")

		// Step 6: Service Status Check
		fmt.Print("  • [6/7] Checking critical system services...")
		fmt.Println(" ✓")

		// Step 7: Time Estimation
		fmt.Print("  • [7/7] Calculating upgrade time estimates...")
		fmt.Println(" ✓")

		// Display current system information
		fmt.Println()
		fmt.Println(utils.FormatHeader("📊 Current System Information"))
		fmt.Printf("🔧 NixOS Version: %s\n", utils.AccentStyle.Render(upgradeInfo.CurrentVersion))
		fmt.Printf("📡 Current Channel: %s\n", utils.AccentStyle.Render(upgradeInfo.CurrentChannel))
		fmt.Printf("⏱️  Estimated Upgrade Time: %s\n", utils.AccentStyle.Render(upgradeInfo.EstimatedTime))

		// Display available channels if not in dry-run mode
		if !dryRun && len(upgradeInfo.AvailableChannels) > 0 {
			fmt.Println()
			fmt.Println(utils.FormatHeader("🔄 Available Upgrade Options"))

			for i, channel := range upgradeInfo.AvailableChannels {
				if channel.IsCurrent {
					continue // Skip current channel
				}

				status := ""
				if channel.IsRecommended {
					status = " " + utils.FormatSuccess("(Recommended)")
				}

				fmt.Printf("%d. %s%s\n", i+1, utils.AccentStyle.Render(channel.Name), status)
				fmt.Printf("   📅 Released: %s\n", channel.ReleaseDate)
				fmt.Printf("   📖 %s\n", channel.Description)
				fmt.Println()
			}
		}

		// Display pre-upgrade checks with real-time progress
		fmt.Println(utils.FormatHeader("🔍 Pre-Upgrade System Checks"))
		fmt.Print(utils.FormatProgress("Running comprehensive system analysis..."))
		fmt.Println()

		hasFailures := false
		hasCritical := false

		for i, check := range upgradeInfo.PreChecks {
			// Show progress for each check
			fmt.Printf("  • [%d/%d] %s...", i+1, len(upgradeInfo.PreChecks), check.Name)

			var statusIcon, statusColor string
			switch check.Status {
			case "pass":
				statusIcon = "✅"
				statusColor = utils.FormatSuccess(check.Message)
				fmt.Print(" ✓")
			case "warn":
				statusIcon = "⚠️"
				statusColor = utils.FormatWarning(check.Message)
				hasFailures = true
				fmt.Print(" ⚠️")
			case "fail":
				statusIcon = "❌"
				statusColor = utils.FormatError(check.Message)
				hasFailures = true
				if check.Critical {
					hasCritical = true
				}
				fmt.Print(" ❌")
			}
			fmt.Println()

			fmt.Printf("    %s %s: %s\n", statusIcon, utils.AccentStyle.Render(check.Name), statusColor)
			if check.Suggestion != "" {
				fmt.Printf("    💡 %s\n", utils.FormatNote(check.Suggestion))
			}
			fmt.Println()
		}

		// Display warnings if any critical issues
		if hasCritical {
			fmt.Println()
			fmt.Println(utils.FormatError("⚠️  CRITICAL ISSUES DETECTED"))
			fmt.Println(utils.FormatWarning("Please resolve critical issues before proceeding with the upgrade."))
			fmt.Println()
		}

		// Stop here if dry-run
		if dryRun {
			fmt.Println()
			fmt.Println(utils.FormatNote("Dry-run complete. Use 'nixai upgrade-advisor' without --dry-run for full guidance."))
			return
		}

		// Display backup advice with progress
		fmt.Println()
		fmt.Print(utils.FormatProgress("💾 Generating backup recommendations..."))
		fmt.Println(" ✓")
		fmt.Println(utils.FormatHeader("💾 Pre-Upgrade Backup Checklist"))
		for i, advice := range upgradeInfo.BackupAdvice {
			fmt.Printf("%d. %s\n", i+1, advice)
		}

		// Generate AI-powered upgrade explanation
		fmt.Println()
		fmt.Print(utils.FormatProgress("🤖 Generating AI-powered upgrade guidance..."))
		fmt.Print("\n  • Analyzing system compatibility...")

		prompt := fmt.Sprintf(`As a NixOS expert, provide comprehensive upgrade guidance for a user upgrading from NixOS %s (channel: %s).

Current System Analysis:
- Version: %s
- Channel: %s
- Pre-check results: %d passed, %d warnings, %d failures
- Estimated time: %s

Please provide:
1. **Upgrade Strategy**: Best approach for this specific upgrade
2. **Risk Assessment**: Potential issues and how to mitigate them  
3. **Channel Recommendations**: Which target version to choose and why
4. **Special Considerations**: Any version-specific gotchas or breaking changes
5. **Recovery Plan**: What to do if something goes wrong

Target version consideration: %s

Format the response in clear markdown with appropriate headers, lists, and emphasis.`,
			upgradeInfo.CurrentVersion,
			upgradeInfo.CurrentChannel,
			upgradeInfo.CurrentVersion,
			upgradeInfo.CurrentChannel,
			countChecksByStatus(upgradeInfo.PreChecks, "pass"),
			countChecksByStatus(upgradeInfo.PreChecks, "warn"),
			countChecksByStatus(upgradeInfo.PreChecks, "fail"),
			upgradeInfo.EstimatedTime,
			func() string {
				if targetVersion != "" {
					return targetVersion
				}
				return "latest stable"
			}())

		fmt.Print(" ✓\n  • Generating upgrade recommendations...")
		response, err := provider.Query(prompt)
		if err != nil {
			fmt.Printf("\n%s Failed to generate AI guidance: %v\n", utils.FormatError("Error:"), err)
		} else {
			fmt.Print(" ✓\n  • Formatting guidance response...")
			fmt.Println(" ✓")
			fmt.Println()
			fmt.Println(utils.FormatHeader("🤖 AI-Powered Upgrade Guidance"))

			// Render the AI response as markdown
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(100),
			)
			if err == nil {
				if formatted, err := renderer.Render(response); err == nil {
					fmt.Print(formatted)
				} else {
					fmt.Println(response)
				}
			} else {
				fmt.Println(response)
			}
		}

		// Display upgrade steps with progress feedback
		fmt.Println()
		fmt.Print(utils.FormatProgress("📋 Preparing step-by-step upgrade instructions..."))
		fmt.Println(" ✓")
		fmt.Println(utils.FormatHeader("📋 Upgrade Steps"))
		for i, step := range upgradeInfo.UpgradeSteps {
			stepNum := fmt.Sprintf("Step %d", i+1)
			fmt.Printf("%s %s\n", utils.AccentStyle.Render(stepNum+":"), utils.InfoStyle.Render(step.Title))
			fmt.Printf("   📝 %s\n", step.Description)
			fmt.Printf("   💻 %s\n", utils.FormatCode(step.Command))
			fmt.Printf("   ⏱️  %s", step.EstimatedTime)

			if step.Optional {
				fmt.Printf(" %s", utils.FormatNote("(Optional)"))
			}
			if step.Dangerous {
				fmt.Printf(" %s", utils.FormatWarning("(Potentially Disruptive)"))
			}
			fmt.Println()
			fmt.Println()
		}

		// Display post-upgrade checks with progress
		fmt.Print(utils.FormatProgress("✅ Generating post-upgrade validation checklist..."))
		fmt.Println(" ✓")
		fmt.Println(utils.FormatHeader("✅ Post-Upgrade Validation"))
		for i, check := range upgradeInfo.PostChecks {
			fmt.Printf("%d. %s\n", i+1, check)
		}

		// Final recommendations with progress feedback
		fmt.Println()
		fmt.Print(utils.FormatProgress("🎯 Generating final recommendations..."))
		fmt.Println(" ✓")
		fmt.Println(utils.FormatHeader("🎯 Final Recommendations"))

		if hasCritical {
			fmt.Println(utils.FormatError("🚨 Do NOT proceed with upgrade until critical issues are resolved"))
		} else if hasFailures {
			fmt.Println(utils.FormatWarning("⚠️  Consider fixing warnings before upgrading for best results"))
		} else {
			fmt.Println(utils.FormatSuccess("✅ System appears ready for upgrade"))
		}

		fmt.Println()
		fmt.Println(utils.FormatTip("💡 Always test upgrades in a VM or spare system first"))
		fmt.Println(utils.FormatTip("💡 Keep installation media handy for recovery if needed"))
		fmt.Println(utils.FormatTip("💡 Consider upgrading incrementally through intermediate versions"))

		if len(upgradeInfo.Warnings) > 0 {
			fmt.Println()
			fmt.Println(utils.FormatHeader("⚠️  Important Warnings"))
			for _, warning := range upgradeInfo.Warnings {
				fmt.Printf("• %s\n", utils.FormatWarning(warning))
			}
		}

		// Completion summary
		fmt.Println()
		fmt.Println(utils.FormatSuccess("🎉 Upgrade analysis complete! Review the guidance above carefully before proceeding."))
		fmt.Println(utils.FormatNote("💡 Consider running 'nixai health' for additional system health insights"))
	},
}

// Helper function to count checks by status
func countChecksByStatus(checks []nixos.CheckResult, status string) int {
	count := 0
	for _, check := range checks {
		if check.Status == status {
			count++
		}
	}
	return count
}

func init() {
	upgradeAdvisorCmd.Flags().String("target", "", "Target NixOS version (e.g., 24.11)")
	upgradeAdvisorCmd.Flags().Bool("dry-run", false, "Show system analysis without upgrade recommendations")
}

// Package repository command for generating Nix derivations from Git repositories
var packageRepoCmd = &cobra.Command{
	Use:   "package-repo <git-url>",
	Short: "Generate Nix derivations from Git repositories using AI",
	Long: `Automatically analyze a Git repository and generate a Nix derivation for packaging it in nixpkgs.

This command performs:
- Repository analysis to detect build systems, dependencies, and project structure
- AI-powered derivation generation with nixpkgs best practices
- Dependency mapping to nixpkgs equivalents
- Validation and suggestions for the generated derivation

Examples:
  nixai package-repo https://github.com/user/project
  nixai package-repo https://github.com/user/rust-app --output ./derivations
  nixai package-repo /path/to/local/repo --local --name my-package

The generated derivation will follow nixpkgs conventions and include:
- Appropriate fetcher (fetchFromGitHub, fetchgit, etc.)
- Correct build function (buildGoModule, buildRustPackage, etc.)
- Proper dependency declarations and build phases
- Meta attributes with license, description, and maintainer info`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}

		// Create logger
		log := logger.NewLoggerWithLevel(cfg.LogLevel)

		// Get flags
		localPath, _ := cmd.Flags().GetString("local")
		outputPath, _ := cmd.Flags().GetString("output")
		packageName, _ := cmd.Flags().GetString("name")
		quiet, _ := cmd.Flags().GetBool("quiet")
		analyzeOnly, _ := cmd.Flags().GetBool("analyze-only")

		// Initialize AI provider
		var aiProvider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			aiProvider = ai.NewOllamaProvider("llama3")
		}

		// Initialize MCP client
		var mcpClient *mcp.MCPClient
		mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
		mcpClient = mcp.NewMCPClient(mcpURL)

		// Create packaging service
		tempDir := "/tmp/nixai-packaging"
		packagingService := packaging.NewPackagingService(aiProvider, mcpClient, tempDir, log)
		defer packagingService.Cleanup()

		// Prepare package request
		request := &packaging.PackageRequest{
			LocalPath:   localPath,
			OutputPath:  outputPath,
			PackageName: packageName,
			Quiet:       quiet,
		}

		// Set repository URL or local path
		if localPath != "" {
			// Resolve relative path to absolute path
			absPath, err := filepath.Abs(localPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to resolve path: %v\n", err)
				os.Exit(1)
			}
			if !utils.IsDirectory(absPath) {
				fmt.Fprintf(os.Stderr, "Local path does not exist or is not a directory: %s\n", absPath)
				os.Exit(1)
			}
			request.LocalPath = absPath
		} else if len(args) > 0 {
			request.RepoURL = args[0]
		} else {
			fmt.Fprintf(os.Stderr, "Either provide a Git URL as argument or use --local flag with a path\n")
			os.Exit(1)
		}

		if !quiet {
			fmt.Println("🔍 Analyzing repository for Nix packaging...")
		}

		// If analyze-only mode, just do analysis
		if analyzeOnly {
			var analysis *packaging.RepoAnalysis
			var err error

			if request.LocalPath != "" {
				analysis, err = packagingService.AnalyzeLocalRepository(request.LocalPath)
			} else {
				// For analyze-only with remote repo, we still need to clone temporarily
				result, err := packagingService.PackageRepository(cmd.Context(), request)
				if err == nil {
					analysis = result.Analysis
				}
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to analyze repository: %v\n", err)
				os.Exit(1)
			}

			// Display analysis results
			fmt.Println("\n📊 Repository Analysis Results:")
			fmt.Printf("Project Name: %s\n", analysis.ProjectName)
			fmt.Printf("Build System: %s\n", analysis.BuildSystem)
			fmt.Printf("Language: %s\n", analysis.Language)
			fmt.Printf("Dependencies: %d\n", len(analysis.Dependencies))
			fmt.Printf("Has Tests: %t\n", analysis.HasTests)
			if analysis.License != "" {
				fmt.Printf("License: %s\n", analysis.License)
			}
			if analysis.Description != "" {
				fmt.Printf("Description: %s\n", analysis.Description)
			}

			if len(analysis.Dependencies) > 0 {
				fmt.Println("\nDependencies:")
				for _, dep := range analysis.Dependencies {
					depType := "runtime"
					if dep.Type != "" {
						depType = dep.Type
					}
					fmt.Printf("  - %s (%s)", dep.Name, depType)
					if dep.Version != "" {
						fmt.Printf(" v%s", dep.Version)
					}
					if dep.System {
						fmt.Printf(" [system]")
					}
					fmt.Println()
				}
			}

			if len(analysis.BuildFiles) > 0 {
				fmt.Println("\nBuild Files:")
				for _, file := range analysis.BuildFiles {
					fmt.Printf("  - %s\n", file)
				}
			}
			return
		}

		// Generate derivation
		result, err := packagingService.PackageRepository(cmd.Context(), request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to package repository: %v\n", err)
			os.Exit(1)
		}

		if !quiet {
			fmt.Println("✅ Repository analysis complete!")
			fmt.Printf("📦 Generated derivation for: %s\n", result.Analysis.ProjectName)
			fmt.Printf("🔧 Build system: %s\n", result.Analysis.BuildSystem)
			fmt.Printf("📝 Language: %s\n", result.Analysis.Language)
		}

		// Display validation issues if any
		if len(result.ValidationIssues) > 0 {
			fmt.Println("\n⚠️  Validation Issues:")
			for _, issue := range result.ValidationIssues {
				fmt.Printf("  - %s\n", issue)
			}
		}

		// Display nixpkgs mappings if available
		if len(result.NixpkgsMappings) > 0 {
			fmt.Println("\n🔗 Suggested nixpkgs mappings:")
			for dep, nixpkg := range result.NixpkgsMappings {
				fmt.Printf("  %s → %s\n", dep, nixpkg)
			}
		}

		// Save or display derivation
		if result.OutputFile != "" {
			fmt.Printf("\n💾 Derivation saved to: %s\n", result.OutputFile)
		} else {
			fmt.Println("\n📄 Generated Nix derivation:")
			fmt.Println("```nix")
			fmt.Println(result.Derivation)
			fmt.Println("```")
		}

		if !quiet {
			fmt.Println("\n💡 Next steps:")
			fmt.Println("  1. Review the generated derivation")
			fmt.Println("  2. Test building with: nix-build -E 'with import <nixpkgs> {}; callPackage ./your-derivation.nix {}'")
			fmt.Println("  3. Submit to nixpkgs following contribution guidelines")
		}
	},
}

// Service examples command for showing real-world config examples
var serviceExamplesCmd = &cobra.Command{
	Use:   "service-examples <service>",
	Short: "Show real-world configuration examples for NixOS services",
	Long: `Get AI-powered, real-world configuration examples and explanations for any NixOS service.

Examples:
  nixai service-examples nginx
  nixai service-examples postgresql
  nixai service-examples docker
  nixai service-examples openssh`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		// Load configuration and initialize AI provider
		cfg, err := config.LoadUserConfig()
		var provider ai.AIProvider
		if err == nil {
			switch cfg.AIProvider {
			case "ollama":
				provider = ai.NewOllamaProvider(cfg.AIModel)
			case "gemini":
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
			case "openai":
				provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			default:
				provider = ai.NewOllamaProvider("llama3")
			}
		} else {
			provider = ai.NewOllamaProvider("llama3")
		}

		fmt.Println(utils.FormatHeader("🛠️ NixOS Service Examples: " + serviceName))
		fmt.Println(utils.FormatProgress("Analyzing service and generating examples..."))

		// Check if MCP server is running for documentation queries
		mcpURL := "http://localhost:8080"
		statusResp, err := http.Get(mcpURL + "/healthz")
		if err == nil && statusResp != nil {
			defer statusResp.Body.Close()
			if statusResp.StatusCode == 200 {
				// Query documentation for the service
				mcpClient := mcp.NewMCPClient(mcpURL)
				doc, err := mcpClient.QueryDocumentation("services." + serviceName)
				if err == nil && strings.TrimSpace(doc) != "" {
					fmt.Println(utils.FormatSuccess("Found service documentation!"))
				}
			}
		}

		// Build comprehensive prompt for service examples
		prompt := buildServiceExamplesPrompt(serviceName)

		// Get AI response
		response, err := provider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("Error getting AI response: " + err.Error()))
			os.Exit(1)
		}

		// Render the response with markdown formatting
		rendered := renderForTerminal(response)
		fmt.Println(rendered)

		fmt.Println()
		fmt.Println(utils.FormatNote("💡 Tip: Use 'nixai explain-option services." + serviceName + ".*' to get detailed explanations of specific options"))
	},
}

// Config linter command for linting and formatting NixOS config files
var lintConfigCmd = &cobra.Command{
	Use:   "lint-config <file>",
	Short: "Lint and analyze NixOS configuration files",
	Long: `Lint, analyze, and suggest improvements for NixOS configuration files using AI-powered analysis.

Features:
- Syntax and structure validation
- Best practices analysis
- Security and performance recommendations
- Anti-pattern detection
- Formatting suggestions

Examples:
  nixai lint-config /etc/nixos/configuration.nix
  nixai lint-config ./flake.nix
  nixai lint-config /home/user/.config/nixpkgs/home.nix`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		fmt.Println(utils.FormatHeader("🔍 NixOS Configuration Linter"))
		fmt.Println(utils.FormatKeyValue("File", filePath))
		fmt.Println(utils.FormatDivider())

		// Check if file exists
		if !utils.IsFile(filePath) {
			fmt.Println(utils.FormatError("File not found: " + filePath))
			os.Exit(1)
		}

		// Read the configuration file
		fmt.Println(utils.FormatProgress("Reading configuration file..."))
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println(utils.FormatError("Error reading file: " + err.Error()))
			os.Exit(1)
		}

		configContent := string(content)
		if strings.TrimSpace(configContent) == "" {
			fmt.Println(utils.FormatWarning("Configuration file is empty"))
			return
		}

		// Initialize AI provider
		cfg, err := config.LoadUserConfig()
		var provider ai.AIProvider
		if err == nil {
			switch cfg.AIProvider {
			case "ollama":
				provider = ai.NewOllamaProvider(cfg.AIModel)
			case "gemini":
				provider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
			case "openai":
				provider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
			default:
				provider = ai.NewOllamaProvider("llama3")
			}
		} else {
			provider = ai.NewOllamaProvider("llama3")
		}

		// Perform basic syntax validation
		fmt.Println(utils.FormatProgress("Performing syntax validation..."))
		syntaxIssues := performBasicSyntaxCheck(configContent)

		// Get file statistics
		lines := strings.Split(configContent, "\n")
		stats := fmt.Sprintf("Lines: %d, Size: %d bytes", len(lines), len(content))

		// Run AI analysis
		fmt.Println(utils.FormatProgress("Running AI-powered configuration analysis..."))
		prompt := buildConfigLintPrompt(configContent, filePath, stats, syntaxIssues)

		response, err := provider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("Error getting AI analysis: " + err.Error()))
			os.Exit(1)
		}

		// Render the response
		rendered := renderForTerminal(response)
		fmt.Println(rendered)

		// Show basic syntax issues if any
		if len(syntaxIssues) > 0 {
			fmt.Println()
			fmt.Println(utils.FormatHeader("⚠️ Syntax Issues Detected"))
			for _, issue := range syntaxIssues {
				fmt.Println(utils.FormatError("• " + issue))
			}
		}

		fmt.Println()
		fmt.Println(utils.FormatNote("💡 Tip: Test your configuration with 'sudo nixos-rebuild dry-run' before applying changes"))
	},
}

// buildServiceExamplesPrompt creates a comprehensive prompt for service examples
func buildServiceExamplesPrompt(serviceName string) string {
	return fmt.Sprintf(`You are a NixOS expert providing real-world configuration examples for services.

**Service:** %s

Please provide comprehensive examples and explanations:

## 🛠️ Service Overview
- What this service does and why you'd use it
- Common use cases and scenarios

## ⚙️ Basic Configuration
- Minimal working configuration
- Essential options that must be set
- Default behavior when enabled

## 🔧 Common Configurations
- Popular real-world setups
- Different use case scenarios
- Configuration variations

## 🚀 Advanced Examples
- Complex configurations
- Integration with other services
- Performance and security optimizations

## 💡 Best Practices
- Security considerations
- Performance recommendations
- Common pitfalls to avoid
- Maintenance tips

## 🔗 Related Services
- Services that work well together
- Dependencies or prerequisites
- Complementary tools

**Format your response using clear Markdown with:**
- Code blocks for all configuration examples (use `+"`nix`"+` code blocks)
- Clear headings and sections
- Practical, working examples
- Explanations for each configuration option

Focus on real-world, practical examples that users can actually use and adapt.`, serviceName)
}

// performBasicSyntaxCheck performs basic Nix syntax validation
func performBasicSyntaxCheck(content string) []string {
	var issues []string

	lines := strings.Split(content, "\n")
	openBraces := 0
	openParens := 0
	openBrackets := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Count braces, parentheses, and brackets
		for _, char := range line {
			switch char {
			case '{':
				openBraces++
			case '}':
				openBraces--
				if openBraces < 0 {
					issues = append(issues, fmt.Sprintf("Line %d: Unmatched closing brace", lineNum+1))
					openBraces = 0
				}
			case '(':
				openParens++
			case ')':
				openParens--
				if openParens < 0 {
					issues = append(issues, fmt.Sprintf("Line %d: Unmatched closing parenthesis", lineNum+1))
					openParens = 0
				}
			case '[':
				openBrackets++
			case ']':
				openBrackets--
				if openBrackets < 0 {
					issues = append(issues, fmt.Sprintf("Line %d: Unmatched closing bracket", lineNum+1))
					openBrackets = 0
				}
			}
		}

		// Check for common syntax issues
		if strings.Contains(line, "=;") {
			issues = append(issues, fmt.Sprintf("Line %d: Empty assignment (=;)", lineNum+1))
		}

		if strings.Count(line, "\"")%2 != 0 {
			issues = append(issues, fmt.Sprintf("Line %d: Unmatched quote", lineNum+1))
		}
	}

	// Check for unclosed braces, parentheses, brackets
	if openBraces > 0 {
		issues = append(issues, fmt.Sprintf("Unclosed braces: %d", openBraces))
	}
	if openParens > 0 {
		issues = append(issues, fmt.Sprintf("Unclosed parentheses: %d", openParens))
	}
	if openBrackets > 0 {
		issues = append(issues, fmt.Sprintf("Unclosed brackets: %d", openBrackets))
	}

	return issues
}

// buildConfigLintPrompt creates a comprehensive prompt for config linting
func buildConfigLintPrompt(configContent, filePath, stats string, syntaxIssues []string) string {
	syntaxSection := ""
	if len(syntaxIssues) > 0 {
		syntaxSection = fmt.Sprintf("\n**Syntax Issues Found:**\n%s\n", strings.Join(syntaxIssues, "\n"))
	}

	return fmt.Sprintf(`You are a NixOS configuration expert performing a comprehensive lint and analysis.

**File:** %s
**Statistics:** %s
%s
**Configuration Content:**
%s

Please provide a detailed analysis:

## 🔍 Syntax Analysis
- Overall syntax quality and correctness
- Any additional syntax issues not caught by basic checks
- Formatting and style recommendations

## 📋 Structure Analysis
- Configuration organization and readability
- Use of proper Nix idioms and patterns
- Module structure and imports

## 🔒 Security Review
- Security best practices compliance
- Potential security issues or vulnerabilities
- Recommended security improvements

## ⚡ Performance Analysis
- Performance implications of current settings
- Resource usage considerations
- Optimization opportunities

## 🎯 Best Practices
- Alignment with NixOS best practices
- Modern Nix language features usage
- Maintainability improvements

## ⚠️ Issues Found
- Deprecated options or patterns
- Conflicting configurations
- Missing recommended settings
- Anti-patterns to avoid

## 💡 Recommendations
- Specific improvements to implement
- Modern alternatives to current approaches
- Additional options to consider

## 📝 Action Items
- Priority fixes to implement
- Optional improvements
- Commands to run for testing

Use clear Markdown formatting with code blocks for configuration examples.
Rate the overall configuration quality (1-10) and provide specific, actionable feedback.`, filePath, stats, syntaxSection, configContent)
}

// Neovim integration command
var neovimCmd = &cobra.Command{
	Use:   "neovim-setup",
	Short: "Set up Neovim integration for nixai",
	Long: `Set up Neovim integration for nixai.

This command creates a Neovim Lua module in your Neovim configuration directory
that integrates with the nixai MCP server for NixOS documentation and assistance.

Examples:
  nixai neovim-setup
  nixai neovim-setup --socket-path=/custom/path/to/socket
  nixai neovim-setup --config-dir=/custom/path/to/neovim/config`,
	Run: func(cmd *cobra.Command, args []string) {
		socketPath, _ := cmd.Flags().GetString("socket-path")
		configDir, _ := cmd.Flags().GetString("config-dir")

		// If no socket path specified, check for environment variable
		if socketPath == "" {
			if envPath := os.Getenv("NIXAI_SOCKET_PATH"); envPath != "" {
				socketPath = envPath
			} else {
				socketPath = "/tmp/nixai-mcp.sock" // Default
			}
		}

		fmt.Println(utils.FormatHeader("🔧 Setting up Neovim integration"))
		fmt.Println(utils.FormatKeyValue("Socket Path", socketPath))
		if configDir != "" {
			fmt.Println(utils.FormatKeyValue("Config Directory", configDir))
		}

		err := neovim.CreateNeovimModule(socketPath, configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, utils.FormatError("Error setting up Neovim integration: "+err.Error()))
			os.Exit(1)
		}

		// Generate init snippet
		initSnippet := neovim.GenerateInitConfig(socketPath)

		// Get config dir that was actually used
		usedConfigDir, _ := neovim.GetUserConfigDir()

		fmt.Println(utils.FormatSuccess("\nNeovim integration set up successfully!"))
		fmt.Printf("\nCreated module: %s/lua/nixai.lua\n\n", usedConfigDir)

		fmt.Println(utils.FormatHeader("Add to your Neovim configuration:"))
		fmt.Print("\nAdd this to your init.lua:\n\n")
		fmt.Println(initSnippet)

		fmt.Println("\nOr if you're using init.vim, add this:")
		fmt.Println("\nlua << EOF")
		fmt.Println(initSnippet)
		fmt.Println("EOF")

		fmt.Println(utils.FormatHeader("\nAvailable Commands:"))
		fmt.Println("  <leader>nq - Ask a NixOS question")
		fmt.Println("  <leader>ns - Get context-aware suggestions")
		fmt.Println("  <leader>no - Explain a NixOS option")
		fmt.Println("  <leader>nh - Explain a Home Manager option")
	},
}

// Version command to display version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version information",
	Long: `Display version information for nixai.

By default, shows detailed version information including commit hash and build date.
Use --short for just the version number, or --json for JSON output.`,
	Run: func(cmd *cobra.Command, args []string) {
		versionInfo := version.Get()

		jsonOutput, _ := cmd.Flags().GetBool("json")
		shortOutput, _ := cmd.Flags().GetBool("short")

		if jsonOutput {
			versionJSON, err := json.MarshalIndent(versionInfo, "", "  ")
			if err != nil {
				fmt.Println(utils.FormatError("Failed to retrieve version information"))
				os.Exit(1)
			}
			fmt.Println(string(versionJSON))
		} else if shortOutput {
			fmt.Println(versionInfo.Short())
		} else {
			fmt.Println(versionInfo.String())
		}
	},
}

// --- Config command handler stubs ---
func handleConfigShow(provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigShow called")
}

func handleConfigSet(args []string, provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigSet called with args:", args)
}

func handleConfigUnset(args []string, provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigUnset called with args:", args)
}

func handleConfigEdit(provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigEdit called")
}

func handleConfigExplain(args []string, provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigExplain called with args:", args)
}

func handleConfigAnalyze(provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigAnalyze called")
}

func handleConfigValidate(provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigValidate called")
}

func handleConfigOptimize(provider ai.AIProvider) {
	fmt.Println("[STUB] handleConfigOptimize called")
}

func handleConfigBackup() {
	fmt.Println("[STUB] handleConfigBackup called")
}

func handleConfigRestore(args []string) {
	fmt.Println("[STUB] handleConfigRestore called with args:", args)
}
