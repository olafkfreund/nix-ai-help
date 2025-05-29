package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/utils"
)

// TODO: Implement AI provider and model switching in interactive mode
// var currentAIProvider string
// var currentModel string = "llama3"

// InteractiveMode starts the interactive command-line interface for nixai.
func InteractiveMode() {
	reader := bufio.NewReader(os.Stdin)

	// Enhanced welcome message
	fmt.Println(utils.FormatHeader("🚀 Welcome to nixai Interactive Mode"))
	fmt.Println(utils.FormatInfo("Type 'help' for commands, 'exit' to quit."))

	if nixosConfigPath != "" {
		fmt.Println(utils.FormatKeyValue("NixOS Config Path", nixosConfigPath))
	}
	fmt.Println(utils.FormatDivider())

	for {
		fmt.Print(utils.AccentStyle.Render("> "))
		input, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nExiting nixai. Goodbye!")
				break
			}
			fmt.Println("Error reading input:", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("Exiting nixai. Goodbye!")
			break
		}

		handleCommand(input)
	}
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
		option := strings.Join(fields[1:], " ")
		// Use the same logic as the CLI command
		explainOptionCmd.Run(explainOptionCmd, []string{option})
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
	case "exit":
		fmt.Println("Exiting nixai. Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("You entered:", command)
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
