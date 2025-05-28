package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/config"

	"gopkg.in/yaml.v3"
)

var currentAIProvider string
var currentModel string = "llama3"

// InteractiveMode starts the interactive command-line interface for nixai.
func InteractiveMode() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to nixai! Type 'help' for commands, 'exit' to quit.")

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
		fmt.Println("  show config                - Show current configuration and MCP sources")
		fmt.Println("  set ai <provider> [model]  - Set AI provider (ollama, gemini, openai) and model (optional)")
		fmt.Println("  exit                       - Exit interactive mode")
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
	case "exit":
		fmt.Println("Exiting nixai. Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("You entered:", command)
	}
}

func showConfig() {
	cfg, err := config.LoadYAMLConfig("configs/default.yaml")
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
	cfg, err := config.LoadYAMLConfig("configs/default.yaml")
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
	tmp := struct {
		Default *config.YAMLConfig `yaml:"default"`
	}{Default: cfg}
	// Save back to YAML
	data, err := yaml.Marshal(&tmp)
	if err != nil {
		fmt.Println("Failed to marshal config:", err)
		return
	}
	err = os.WriteFile("configs/default.yaml", data, 0644)
	if err != nil {
		fmt.Println("Failed to write config:", err)
		return
	}
	fmt.Println("AI provider updated in config. It will be used for future diagnoses.")
}
