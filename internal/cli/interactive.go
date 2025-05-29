package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
)

var currentAIProvider string
var currentModel string = "llama3"

// InteractiveMode starts the interactive command-line interface for nixai.
func InteractiveMode() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to nixai! Type 'help' for commands, 'exit' to quit.")
	if nixosConfigPath != "" {
		fmt.Printf("Using NixOS config folder: %s\n", nixosConfigPath)
	}

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
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
		fmt.Println("Available commands:")
		fmt.Println("  diagnose <log/config>      - Diagnose NixOS issues")
		fmt.Println("  search <package>           - Search for and install Nix packages")
		fmt.Println("  show config                - Show current configuration and MCP sources")
		fmt.Println("  set ai <provider> [model]  - Set AI provider (ollama, gemini, openai) and model (optional)")
		fmt.Println("  set-nixos-path <path>      - Set path to NixOS config folder")
		fmt.Println("  flake <subcommand>         - Manage Nix flakes (show, update, check, explain-inputs, explain <input>, ...)")
		fmt.Println("  exit                       - Exit interactive mode")
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
		currentModel = model
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
