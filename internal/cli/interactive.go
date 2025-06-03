package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/utils"
	"nix-ai-help/pkg/version"

	"math/rand"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// Global variable for NixOS config path (directory)
var nixosConfigPath string

var tipsOfTheDay = []string{
	"Use 'package-repo <url>' to generate a Nix derivation from a Git repo!",
	"Try 'explain-option <option>' to get a detailed explanation of any NixOS option.",
	"You can use up/down arrows to navigate your command history.",
	"Press Tab to autocomplete commands and options.",
	"Pipe logs into nixai for instant diagnostics!",
	"Use 'help' or '?' at any time for contextual help.",
}

// TODO: Implement AI provider and model switching in interactive mode
// var currentAIProvider string
// var currentModel string = "llama3"

// --- CLI command stubs for interactive mode ---
var healthCheckCmd = &cobra.Command{Use: "health", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] health check not implemented") }}
var upgradeAdvisorCmd = &cobra.Command{Use: "upgrade-advisor", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] upgrade advisor not implemented") }}
var serviceExamplesCmd = &cobra.Command{Use: "service-examples", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] service examples not implemented") }}
var lintConfigCmd = &cobra.Command{Use: "lint-config", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] lint config not implemented") }}
var decodeErrorCmd = &cobra.Command{Use: "decode-error", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] decode error not implemented") }}
var findOptionCmd = &cobra.Command{Use: "find-option", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] find-option not implemented") }}
var learnBasicsCmd = &cobra.Command{Use: "basics", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] learn basics not implemented") }}
var learnAdvancedCmd = &cobra.Command{Use: "advanced", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] learn advanced not implemented") }}
var learnQuizCmd = &cobra.Command{Use: "quiz", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] learn quiz not implemented") }}
var learnPathCmd = &cobra.Command{Use: "path <topic>", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] learn path not implemented") }}
var learnProgressCmd = &cobra.Command{Use: "progress", Run: func(cmd *cobra.Command, args []string) { fmt.Println("[interactive] learn progress not implemented") }}

// Ensure learn subcommands are added to learnCmd
func init() {
	learnCmd.AddCommand(learnBasicsCmd)
	learnCmd.AddCommand(learnAdvancedCmd)
	learnCmd.AddCommand(learnQuizCmd)
	learnCmd.AddCommand(learnPathCmd)
	learnCmd.AddCommand(learnProgressCmd)
}

// Stub for ExplainFlakeInputs
func ExplainFlakeInputs(args []string) {
	fmt.Println("[interactive] flake explain-inputs not implemented")
}

func printWelcomeScreen() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	tip := tipsOfTheDay[rnd.Intn(len(tipsOfTheDay))]
	welcome := utils.FormatHeader("Welcome to nixai - Your NixOS AI Assistant 🐧") +
		"\n" +
		"Type a command or question, or type 'help' for a list of commands.\n" +
		utils.FormatDivider() +
		"\n" +
		"Tip of the day: " + tip + "\n" +
		utils.FormatDivider() +
		"\nPopular commands:\n" +
		utils.FormatKeyValue("ask <question>", "Ask any NixOS question") +
		utils.FormatKeyValue("package-repo <url>", "Generate Nix derivation from repo") +
		utils.FormatKeyValue("explain-option <option>", "Explain a NixOS option") +
		utils.FormatKeyValue("explain-home-option <option>", "Explain a Home Manager option") +
		utils.FormatKeyValue("search <package>", "Search for a Nix package") +
		utils.FormatKeyValue("exit", "Quit interactive mode")
	out, _ := glamour.Render(welcome, "dark")
	fmt.Println(out)
}

// InteractiveMode starts the interactive command-line interface for nixai.
func InteractiveMode() {
	printWelcomeScreen()
	// Setup readline for input with history, autocomplete, and multi-line support (manual workaround)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            utils.FormatHeader("nixai> "),
		HistoryFile:       "/tmp/nixai_history.tmp",
		AutoComplete:      commandCompleter{},
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r == readline.CharCtrlZ {
				return r, false // disable Ctrl+Z
			}
			return r, true
		},
	})
	if err != nil {
		fmt.Println("Error initializing interactive mode:", err)
		return
	}
	defer rl.Close()

	var multilineBuf []string
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		}

		// Multi-line input: Shift+Enter (\n) appends, Enter submits
		if strings.HasSuffix(line, "\\") {
			multilineBuf = append(multilineBuf, strings.TrimSuffix(line, "\\"))
			rl.SetPrompt(utils.FormatHeader("... "))
			continue
		}
		if len(multilineBuf) > 0 {
			multilineBuf = append(multilineBuf, line)
			input := strings.Join(multilineBuf, "\n")
			multilineBuf = nil
			rl.SetPrompt(utils.FormatHeader("nixai> "))
			processInteractiveInput(input)
			continue
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		processInteractiveInput(input)
	}
}

// processInteractiveInput handles a single or multi-line input in interactive mode.
func processInteractiveInput(input string) {
	if input == "exit" || input == "quit" {
		fmt.Println(utils.FormatDivider() + "\nGoodbye! 👋")
		os.Exit(0)
	}
	if input == "help" || input == "?" {
		printWelcomeScreen()
		return
	}
	if strings.HasPrefix(input, "explain-option") && len(strings.Fields(input)) == 1 {
		fmt.Println(utils.FormatTip("Usage: explain-option <option>\nExample: explain-option services.nginx.enable"))
		return
	}
	if strings.HasPrefix(input, "search") && len(strings.Fields(input)) == 1 {
		fmt.Println(utils.FormatTip("Usage: search <package>\nExample: search libreoffice"))
		return
	}
	if strings.HasPrefix(input, "package-repo") && len(strings.Fields(input)) == 1 {
		fmt.Println(utils.FormatTip("Usage: package-repo <url>\nExample: package-repo https://github.com/NixOS/nixpkgs"))
		return
	}
	if strings.HasPrefix(input, "explain-home-option") && len(strings.Fields(input)) == 1 {
		fmt.Println(utils.FormatTip("Usage: explain-home-option <option>\nExample: explain-home-option programs.zsh.enable"))
		return
	}
	fmt.Println(utils.FormatDivider())
	fmt.Println("You entered:", input)
	fmt.Println(utils.FormatDivider())
}

var (
	packageAutocompleteCache = struct {
		sync.Mutex
		query   string
		results []string
	}{query: "", results: nil}

	optionAutocompleteCache = struct {
		sync.Mutex
		prefix  string
		results []string
	}{prefix: "", results: nil}
)

// Animated snowflake spinner for progress indication
var snowflakeFrames = []string{"❄️  ", "  ❄️", " ❄️ ", "❄️  ", "  ❄️"}

func showSnowflakeSpinner(stop <-chan struct{}) {
	for i := 0; ; i++ {
		select {
		case <-stop:
			fmt.Print("\r   \r")
			return
		default:
			fmt.Printf("\r%s", snowflakeFrames[i%len(snowflakeFrames)])
			time.Sleep(120 * time.Millisecond)
		}
	}
}

// commandCompleter implements readline.AutoCompleter for command palette/autocomplete
// Enhanced: supports fuzzy search for commands, package name completion for 'search', and option completion for 'explain-option' and 'explain-home-option'.
type commandCompleter struct{}

func (c commandCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line)
	fields := strings.Fields(input)
	cmds := []string{"ask", "package-repo", "explain-option", "explain-home-option", "search", "exit", "help", "?", "quit"}
	var suggestions [][]rune

	if len(fields) == 0 {
		for _, cmd := range cmds {
			suggestions = append(suggestions, []rune(cmd))
		}
		return suggestions, 0
	}

	// Fuzzy match for command name
	if len(fields) == 1 && pos <= len(fields[0]) {
		word := fields[0]
		for _, cmd := range cmds {
			if strings.Contains(cmd, word) {
				suggestions = append(suggestions, []rune(cmd))
			}
		}
		return suggestions, 0
	}

	// Autocomplete for 'search <package>'
	if fields[0] == "search" && len(fields) >= 2 {
		query := strings.Join(fields[1:], " ")
		maxResults := 10
		packageAutocompleteCache.Lock()
		if packageAutocompleteCache.query != query {
			executor := nixos.NewExecutor("")
			results, err := executor.SearchNixPackagesForAutocomplete(query, maxResults)
			if err == nil {
				packageAutocompleteCache.query = query
				packageAutocompleteCache.results = results
			}
		}
		pkgs := packageAutocompleteCache.results
		packageAutocompleteCache.Unlock()
		for _, pkg := range pkgs {
			if strings.HasPrefix(pkg, fields[len(fields)-1]) {
				suggestions = append(suggestions, []rune(pkg))
			}
		}
		return suggestions, len(fields[0]) + 1 // after 'search '
	}

	// Option completion for 'explain-option <option>' and 'explain-home-option <option>'
	if (fields[0] == "explain-option" || fields[0] == "explain-home-option") && len(fields) >= 2 {
		prefix := fields[len(fields)-1]
		optionAutocompleteCache.Lock()
		if optionAutocompleteCache.prefix != prefix {
			optionAutocompleteCache.prefix = prefix
			optionAutocompleteCache.results = nil
			optionAutocompleteCache.Unlock()
			// Show snowflake spinner while querying MCP
			stop := make(chan struct{})
			go showSnowflakeSpinner(stop)
			cfg, _ := config.LoadUserConfig()
			mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
			client := mcp.NewMCPClient(mcpURL)
			results, err := client.OptionCompletion(prefix)
			close(stop)
			fmt.Print("\r   \r") // Clear spinner
			optionAutocompleteCache.Lock()
			if err == nil {
				optionAutocompleteCache.results = results
			}
			optionAutocompleteCache.Unlock()
		}
		results := optionAutocompleteCache.results
		optionAutocompleteCache.Unlock()
		for _, opt := range results {
			if strings.HasPrefix(opt, prefix) {
				suggestions = append(suggestions, []rune(opt))
			}
		}
		return suggestions, len(fields[0]) + 1 // after command
	}

	return suggestions, 0
}

// handleCommand processes user commands entered in interactive mode.
func handleCommand(command string) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return
	}
	switch fields[0] {
	case "help":
		fmt.Println(utils.FormatHeader("📚 Available Commands"))

		commands := []string{
			"diagnose <log/config>      - Diagnose NixOS issues",
			"decode-error <error>       - AI-powered error decoder and fix generator",
			"search <package>           - Search for and install Nix packages",
			"explain-option <option>    - Get AI-powered explanations for NixOS options",
			"find-option <description>  - Find NixOS options from natural language description",
			"service-examples <service> - Get AI-powered service configuration examples",
			"lint-config <file>         - AI-powered configuration file analysis and linting",
			"package-repo <path>        - Analyze repositories and generate Nix derivations",
			"devenv <subcommand>        - Create and manage development environments",
			"  devenv list               - List available devenv templates",
			"  devenv create <template>  - Create development environment",
			"  devenv suggest <desc>     - Get AI template suggestions",
			"config <subcommand>        - AI-assisted Nix configuration management",
			"  config show               - Show current configuration with AI analysis",
			"  config set <key> <value>  - Set configuration option with AI guidance",
			"  config unset <key>        - Unset configuration option with safety checks",
			"  config explain <key>      - AI-powered explanation of config options",
			"  config analyze            - Comprehensive configuration analysis",
			"  config validate           - Validate configuration and suggest improvements",
			"  config optimize           - AI recommendations for performance optimization",
			"  config backup             - Create backup of current configuration",
			"  config restore <backup>   - Restore configuration from backup",
			"health                     - Run comprehensive system health check",
			"upgrade-advisor            - Get AI-powered upgrade guidance and safety checks",
			"show config                - Show current configuration and MCP sources",
			"set ai <provider> [model]  - Set AI provider (ollama, gemini, openai) and model (optional)",
			"set-nixos-path <path>      - Set path to NixOS config folder",
			"flake <subcommand>         - Manage Nix flakes (show, update, check, explain-inputs, explain <input>, ...)",
			"learn <subcommand>         - Learn Nix concepts (basics, advanced, quiz, path <topic>, progress)",
			"version                    - Display version information",
			"exit                       - Exit interactive mode",
		}

		fmt.Println(utils.FormatList(commands))
	case "config":
		if len(fields) < 2 {
			fmt.Println("Usage: config <show|set|unset|edit|explain|analyze|validate|optimize|backup|restore> [args...]")
			fmt.Println("Examples:")
			fmt.Println("  config show                              # Show current config with analysis")
			fmt.Println("  config set experimental-features \"nix-command flakes\"")
			fmt.Println("  config explain substituters             # Get AI explanation")
			fmt.Println("  config analyze                          # Full analysis")
			fmt.Println("  config validate                         # Validate config")
			fmt.Println("  config optimize                         # Performance tips")
			return
		}
		// Use the same logic as the CLI command by calling it directly
		configCmd.Run(configCmd, fields[1:])
		return
	case "health":
		fmt.Println("Running comprehensive system health check...")
		healthCheckCmd.Run(healthCheckCmd, []string{})
		return
	case "upgrade-advisor":
		fmt.Println("Starting upgrade advisor analysis...")
		upgradeAdvisorCmd.Run(upgradeAdvisorCmd, []string{})
		return
	case "service-examples":
		if len(fields) < 2 {
			fmt.Println("Usage: service-examples <service>")
			fmt.Println("Examples:")
			fmt.Println("  service-examples nginx")
			fmt.Println("  service-examples postgresql")
			fmt.Println("  service-examples docker")
			fmt.Println("  service-examples prometheus")
			return
		}
		service := strings.Join(fields[1:], " ")
		// Use the same logic as the CLI command by calling it directly
		serviceExamplesCmd.Run(serviceExamplesCmd, []string{service})
		return
	case "lint-config":
		if len(fields) < 2 {
			fmt.Println("Usage: lint-config <file>")
			fmt.Println("Examples:")
			fmt.Println("  lint-config /etc/nixos/configuration.nix")
			fmt.Println("  lint-config ~/.config/nixpkgs/home.nix")
			fmt.Println("  lint-config ./flake.nix")
			return
		}
		file := strings.Join(fields[1:], " ")
		// Use the same logic as the CLI command by calling it directly
		lintConfigCmd.Run(lintConfigCmd, []string{file})
		return
	case "decode-error":
		if len(fields) < 2 {
			fmt.Println("Usage: decode-error <error_message>")
			fmt.Println("Examples:")
			fmt.Println("  decode-error \"syntax error at line 42\"")
			fmt.Println("  decode-error \"error: function 'buildNodePackage' called without required argument\"")
			return
		}
		errorMessage := strings.Join(fields[1:], " ")
		// Use the same logic as the CLI command by calling it directly
		decodeErrorCmd.Run(decodeErrorCmd, []string{errorMessage})
		return
	case "search":
		if len(fields) < 2 {
			fmt.Println("Usage: search <package>")
			return
		}
		query := strings.Join(fields[1:], " ")
		cfg, _ := config.LoadUserConfig()
		configPath := ""
		if nixosConfigPath != "" {
			fmt.Printf("Using NixOS config folder: %s\n", nixosConfigPath)
			configPath = nixosConfigPath
		} else if cfg != nil && cfg.NixosFolder != "" {
			configPath = cfg.NixosFolder
		}
		executor := nixos.NewExecutor(configPath)
		fmt.Printf("Searching for Nix packages matching: %s\n", query)
		output, err := executor.SearchNixPackages(query)
		if err != nil {
			fmt.Printf("Error searching for packages: %v\n", err)
			return
		}
		lines := strings.Split(output, "\n")
		var pkgs []struct{ Attr, Name, Desc string }
		var lastAttr string
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "evaluating ") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) > 1 && (strings.Contains(fields[0], ".") || strings.Contains(fields[0], ":")) {
				attr := fields[0]
				attr = strings.TrimPrefix(attr, "legacyPackages.x86_64-linux.")
				attr = strings.TrimPrefix(attr, "nixpkgs.")
				name := attr
				desc := strings.Join(fields[1:], " ")
				pkgs = append(pkgs, struct{ Attr, Name, Desc string }{fields[0], name, desc})
				lastAttr = fields[0]
				continue
			}
			if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && lastAttr != "" {
				if len(pkgs) > 0 {
					pkgs[len(pkgs)-1].Desc += " " + strings.TrimSpace(line)
				}
				continue
			}
			if idx := strings.Index(line, " - "); idx > 0 {
				name := line[:idx]
				desc := line[idx+3:]
				pkgs = append(pkgs, struct{ Attr, Name, Desc string }{name, name, desc})
				lastAttr = name
			}
		}
		if len(pkgs) == 0 {
			fmt.Println("No packages found.")
			return
		}
		fmt.Println()
		for i, pkg := range pkgs {
			fmt.Printf("%2d. %s\n    %s\n", i+1, pkg.Name, pkg.Desc)
		}
		fmt.Print("\nEnter the number of the package to see configuration and test options (or leave blank to exit): ")
		reader := bufio.NewReader(os.Stdin)
		sel, _ := reader.ReadString('\n')
		sel = strings.TrimSpace(sel)
		if sel == "" {
			return
		}
		idx := -1
		fmt.Sscanf(sel, "%d", &idx)
		if idx < 1 || idx > len(pkgs) {
			fmt.Println("Invalid selection.")
			return
		}
		pkg := pkgs[idx-1]
		fmt.Printf("\nSelected: %s\n\n", pkg.Name)
		fmt.Println("NixOS (configuration.nix):")
		fmt.Printf("  environment.systemPackages = with pkgs; [ %s ];\n", pkg.Name)
		fmt.Println("Home Manager (home.nix):")
		fmt.Printf("  home.packages = with pkgs; [ %s ];\n", pkg.Name)
		fmt.Println("\nFetching available options with nixos-option --find...")
		optOut, err := executor.ListServiceOptions(pkg.Name)
		if err == nil && strings.TrimSpace(optOut) != "" {
			fmt.Println(optOut)
		} else {
			fmt.Println("No additional options found or nixos-option --find failed.")
		}
		fmt.Print("\nTest this package in a temporary shell? [y/N]: ")
		yn, _ := reader.ReadString('\n')
		yn = strings.TrimSpace(yn)
		if strings.ToLower(yn) == "y" {
			fmt.Printf("\nLaunching 'nix-shell -p %s'...\n", pkg.Name)
			cmdShell := exec.Command("nix-shell", "-p", pkg.Name)
			cmdShell.Stdin = os.Stdin
			cmdShell.Stdout = os.Stdout
			cmdShell.Stderr = os.Stderr
			err := cmdShell.Run()
			if err != nil {
				fmt.Printf("Error running nix-shell: %v\n", err)
			}
			return
		}
	case "package-repo":
		if len(fields) < 2 {
			fmt.Println("Usage: package-repo <path>")
			fmt.Println("Examples:")
			fmt.Println("  package-repo . --local")
			fmt.Println("  package-repo /path/to/project")
			fmt.Println("  package-repo . --analyze-only")
			fmt.Println("  package-repo https://github.com/user/project")
			return
		}
		// Pass all arguments after 'package-repo' to the CLI command
		packageRepoCmd.Run(packageRepoCmd, fields[1:])
		return
	case "show":
		if len(fields) > 1 && fields[1] == "config" {
			showConfig()
		} else {
			fmt.Println("Unknown show command. Try 'show config'.")
		}
	case "set":
		if len(fields) >= 3 && fields[1] == "ai" {
			provider := fields[2]
			model := ""
			if len(fields) > 3 {
				model = fields[3]
			}
			setAIProvider(provider, model)
		} else {
			fmt.Println("Usage: set ai <provider> [model]")
		}
	case "set-nixos-path":
		if len(fields) < 2 {
			fmt.Println("Usage: set-nixos-path <path-to-nixos-config-folder>")
			return
		}
		nixosConfigPath = fields[1]
		fmt.Printf("NixOS config folder set to: %s\n", nixosConfigPath)
		return
	case "flake":
		if len(fields) < 2 {
			fmt.Println("Usage: flake <show|update|check|explain-inputs|explain <input>|...>")
			return
		}
		if fields[1] == "explain-inputs" {
			ExplainFlakeInputs(nil)
			return
		}
		if fields[1] == "explain" {
			if len(fields) >= 3 {
				ExplainFlakeInputs(fields[2:3])
			} else {
				fmt.Println("Usage: flake explain <input>")
			}
			return
		}
		// fallback: pass to nix flake
		cmdArgs := append([]string{"flake"}, fields[1:]...)
		out, err := exec.Command("nix", cmdArgs...).CombinedOutput()
		fmt.Println(string(out))
		if err != nil {
			fmt.Printf("nix flake failed: %v\n", err)
			problemSummary := summarizeBuildOutput(string(out))
			if problemSummary != "" {
				fmt.Println("\nProblem summary:")
				fmt.Println(problemSummary)
			}
		}
		return
	case "explain-option":
		if len(fields) < 2 {
			fmt.Println("Usage: explain-option <option>")
			fmt.Println("Examples:")
			fmt.Println("  explain-option services.nginx.enable")
			fmt.Println("  explain-option networking.firewall.enable")
			return
		}
		// Use the same logic as the CLI command
		return
	case "find-option":
		if len(fields) < 2 {
			fmt.Println("Usage: find-option <description>")
			fmt.Println("Examples:")
			fmt.Println("  find-option \"enable SSH access\"")
			fmt.Println("  find-option \"configure firewall\"")
			fmt.Println("  find-option \"set up automatic updates\"")
			fmt.Println("  find-option \"enable docker\"")
			return
		}
		description := strings.Join(fields[1:], " ")
		// Use the same logic as the CLI command
		findOptionCmd.Run(findOptionCmd, []string{description})
		return
	case "devenv":
		if len(fields) < 2 {
			fmt.Println("Usage: devenv <list|create|suggest>")
			fmt.Println("Examples:")
			fmt.Println("  devenv list                              # List available templates")
			fmt.Println("  devenv create python myproject           # Create Python environment")
			fmt.Println("  devenv create rust --with-wasm          # Create Rust with WebAssembly")
			fmt.Println("  devenv suggest \"web app with database\"   # Get AI recommendations")
			return
		}

		switch fields[1] {
		case "list":
			devenvListCmd.Run(devenvListCmd, []string{})
		case "create":
			if len(fields) < 3 {
				fmt.Println("Usage: devenv create <template> [project-name]")
				fmt.Println("Examples:")
				fmt.Println("  devenv create python myapp")
				fmt.Println("  devenv create rust --with-wasm")
				fmt.Println("  devenv create nodejs --framework nextjs")
				return
			}
			devenvCreateCmd.Run(devenvCreateCmd, fields[2:])
		case "suggest":
			if len(fields) < 3 {
				fmt.Println("Usage: devenv suggest <description>")
				fmt.Println("Examples:")
				fmt.Println("  devenv suggest \"web application with database\"")
				fmt.Println("  devenv suggest \"machine learning project\"")
				return
			}
			description := strings.Join(fields[2:], " ")
			devenvSuggestCmd.Run(devenvSuggestCmd, []string{description})
		default:
			fmt.Println("Unknown devenv subcommand. Use: list, create, or suggest")
		}
		return
	case "learn":
		if len(fields) < 2 {
			fmt.Println("Usage: learn <basics|advanced|quiz|path <topic>|progress>")
			fmt.Println("Examples:")
			fmt.Println("  learn basics")
			fmt.Println("  learn advanced")
			fmt.Println("  learn quiz")
			fmt.Println("  learn path flakes")
			fmt.Println("  learn progress")
			return
		}
		sub := fields[1]
		switch sub {
		case "basics":
			learnBasicsCmd.Run(learnBasicsCmd, []string{})
		case "advanced":
			learnAdvancedCmd.Run(learnAdvancedCmd, []string{})
		case "quiz":
			learnQuizCmd.Run(learnQuizCmd, []string{})
		case "path":
			if len(fields) < 3 {
				fmt.Println("Usage: learn path <topic>")
				return
			}
			topic := strings.Join(fields[2:], " ")
			learnPathCmd.Run(learnPathCmd, []string{topic})
		case "progress":
			learnProgressCmd.Run(learnProgressCmd, []string{})
		default:
			fmt.Println("Unknown learn subcommand. Try: basics, advanced, quiz, path <topic>, progress")
		}
		return
	case "version":
		versionInfo := version.Get()
		fmt.Println(utils.FormatHeader("📦 Version Information"))
		fmt.Println(versionInfo.String())
		fmt.Println(utils.FormatKeyValue("Platform", versionInfo.Platform))
		fmt.Println(utils.FormatKeyValue("Go Version", versionInfo.GoVersion))
	case "exit":
		fmt.Println("Exiting nixai. Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("You entered:", command)
		fmt.Println("Type 'help' to see available commands.")
	}
}

func showConfig() {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println("Could not load config:", err)
		return
	}
	fmt.Println("Current nixai configuration:")
	fmt.Printf("  AI Provider: %s\n", cfg.AIProvider)
	fmt.Printf("  Log Level:   %s\n", cfg.LogLevel)
	fmt.Printf("  MCP Sources:\n")
	for _, src := range cfg.MCPServer.DocumentationSources {
		fmt.Printf("    - %s\n", src)
	}
}

func setAIProvider(provider, model string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println("Could not load config:", err)
		return
	}
	cfg.AIProvider = provider
	if provider == "ollama" && model != "" {
		fmt.Printf("Set AI provider to '%s' with model '%s'.\n", provider, model)
	} else if provider != "ollama" {
		fmt.Printf("Set AI provider to '%s'.\n", provider)
	}
	cfg.AIModel = model // set model if provided
	err = config.SaveUserConfig(cfg)
	if err != nil {
		fmt.Println("Failed to write user config:", err)
		return
	}
	fmt.Println("AI provider updated in user config. It will be used for future diagnoses.")
}
