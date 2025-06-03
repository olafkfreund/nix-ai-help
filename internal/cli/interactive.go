package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

// InteractiveMode starts the interactive command-line interface for nixai.
func InteractiveMode() {
	printInteractiveWelcome()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("nixai> ")
		if !scanner.Scan() {
			fmt.Println("\nExiting nixai. Goodbye!")
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println(utils.FormatDivider() + "\nGoodbye! 👋")
			os.Exit(0)
		}
		if input == "help" || input == "?" {
			printInteractiveWelcome()
			continue
		}
		fields := strings.Fields(input)
		if len(fields) == 0 {
			continue
		}
		// Support: 'nixai interactive <command> ...' (e.g., 'nixai interactive store')
		if fields[0] == "interactive" && len(fields) > 1 {
			fields = fields[1:]
		}

		// Build command map dynamically from root command - this gets all registered commands
		knownCommands := make(map[string]*cobra.Command)
		for _, cmd := range rootCmd.Commands() {
			knownCommands[cmd.Name()] = cmd
		}

		// Handle special interactive-only commands first
		switch fields[0] {
		case "help", "?":
			printInteractiveWelcome()
			continue
		case "exit", "quit":
			fmt.Println(utils.FormatDivider() + "\nGoodbye! 👋")
			os.Exit(0)
		case "interactive":
			fmt.Println(utils.FormatTip("You are already in interactive mode!"))
			continue
		default:
			// Try to run any registered command
			if cmd, ok := knownCommands[fields[0]]; ok {
				output, err := runCommandAndCaptureOutput(cmd, fields[1:])
				if err != nil {
					fmt.Println(utils.FormatTip("Error: " + err.Error()))
				} else if strings.TrimSpace(output) == "" {
					fmt.Println(utils.FormatTip("No output from command. Try a subcommand like 'list', 'show', or 'add'."))
				} else {
					fmt.Println(output)
				}
			} else {
				// Handle questions directly without a command
				if len(fields) > 0 {
					question := strings.Join(fields, " ")
					answer, err := handleAsk(question)
					if err != nil {
						fmt.Println(utils.FormatTip("Error: " + err.Error()))
					} else {
						if strings.TrimSpace(answer) == "" {
							fmt.Println(utils.FormatTip("Unknown command: " + fields[0] + ". Type 'help' to see available commands."))
						} else {
							fmt.Println(answer)
						}
					}
				} else {
					fmt.Println(utils.FormatTip("Unknown command. Type 'help' to see available commands."))
				}
			}
		}
	}
}

func printInteractiveWelcome() {
	header := utils.FormatHeader("❄️ nixai: NixOS AI Assistant")
	intro := "Welcome to the interactive shell! Type a command or question, or type 'help' for a list of commands."
	divider := utils.FormatDivider()

	// Build a visually appealing, non-Markdown menu using utils formatting
	menu := strings.Join([]string{
		utils.FormatKeyValue("🤖 ask <question>", "Ask any NixOS question"),
		utils.FormatKeyValue("🛠️ build", "Enhanced build troubleshooting and optimization"),
		utils.FormatKeyValue("🌐 community", "Community resources and support (not yet implemented)"),
		utils.FormatKeyValue("🔄 completion", "Generate the autocompletion script for the specified shell"),
		utils.FormatKeyValue("⚙️ config", "Manage nixai configuration"),
		utils.FormatKeyValue("🧑‍💻 configure", "Configure NixOS interactively (not yet implemented)"),
		utils.FormatKeyValue("🔗 deps", "Analyze NixOS configuration dependencies and imports"),
		utils.FormatKeyValue("🧪 devenv", "Create and manage development environments with devenv"),
		utils.FormatKeyValue("🩺 diagnose", "Diagnose NixOS issues (not yet implemented)"),
		utils.FormatKeyValue("🩻 doctor", "Run NixOS health checks (not yet implemented)"),
		utils.FormatKeyValue("🖥️ explain-option <option>", "Explain a NixOS option"),
		utils.FormatKeyValue("🧊 flake", "Nix flake utilities (not yet implemented)"),
		utils.FormatKeyValue("🧹 gc", "AI-powered garbage collection analysis and cleanup"),
		utils.FormatKeyValue("💻 hardware", "AI-powered hardware configuration optimizer"),
		utils.FormatKeyValue("❓ help", "Help about any command"),
		utils.FormatKeyValue("💬 interactive", "Launch interactive AI-powered NixOS assistant shell"),
		utils.FormatKeyValue("📚 learn", "NixOS learning and training commands (not yet implemented)"),
		utils.FormatKeyValue("📝 logs", "Analyze and parse NixOS logs (not yet implemented)"),
		utils.FormatKeyValue("🖧 machines", "Manage and synchronize NixOS configurations across multiple machines"),
		utils.FormatKeyValue("🛰️ mcp-server", "Start or manage the MCP server (not yet implemented)"),
		utils.FormatKeyValue("🔀 migrate", "AI-powered migration assistant for channels and flakes"),
		utils.FormatKeyValue("📝 neovim-setup", "Neovim integration setup (not yet implemented)"),
		utils.FormatKeyValue("📦 package-repo <url>", "Analyze Git repos and generate Nix derivations (not yet implemented)"),
		utils.FormatKeyValue("🔍 search <package>", "Search for NixOS packages/services and get config/AI tips"),
		utils.FormatKeyValue("🔖 snippets", "Manage NixOS configuration snippets"),
		utils.FormatKeyValue("💾 store", "Manage, backup, and analyze the Nix store"),
		utils.FormatKeyValue("📄 templates", "Manage NixOS configuration templates and snippets"),
		utils.FormatKeyValue("❌ exit", "Exit interactive mode"),
	}, "\n")

	// Use glamour for header/intro, but print menu as styled plain text
	customStyle := `{
	  "header": {"color": "#8be9fd", "bold": true},
	  "hr": {"color": "#44475a"},
	  "blockquote": {"color": "#6272a4"}
	}`
	r, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(customStyle)),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Println(header + "\n" + intro + "\n" + divider + "\n" + menu)
		return
	}
	out, err := r.Render(header + "\n" + intro + "\n" + divider)
	if err != nil {
		fmt.Println(header + "\n" + intro + "\n" + divider + "\n" + menu)
		return
	}
	fmt.Print(out)
	fmt.Println(menu)
	fmt.Println(divider)
}

// Handler for 'ask' command - handles direct questions in interactive mode
func handleAsk(question string) (string, error) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return "", err
	}
	provider := InitializeAIProvider(cfg)
	resp, err := provider.Query(question)
	if err != nil {
		return "", err
	}
	return utils.RenderMarkdown(resp), nil
}

// Helper to run a cobra.Command and capture its output as string
func runCommandAndCaptureOutput(cmd *cobra.Command, args []string) (string, error) {
	origOut := cmd.OutOrStdout()
	origErr := cmd.OutOrStderr()
	var sb strings.Builder
	cmd.SetOut(&sb)
	cmd.SetErr(&sb)
	cmd.SetArgs(args)
	err := cmd.Execute()
	cmd.SetOut(origOut)
	cmd.SetErr(origErr)
	return sb.String(), err
}
