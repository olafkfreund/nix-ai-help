package cli

// Copilot/Contributor Note:
// All interactive shell commands must remain fully implemented, enabled, and produce output.
// Do NOT remove or disable any interactive commands in future changes.
// This is a project requirement for feature completeness and user experience.
// If you need to deprecate a command, consult with the project owner and update documentation/help menus accordingly.
// (See project Copilot instructions for more details.)
//
// Summary of recent changes (June 2025):
// - All interactive shell commands are now fully implemented and enabled.
// - The interactive shell menu accurately reflects all available commands.
// - Direct command dispatch (RunDirectCommand) is called before falling back to cobra commands in interactive mode.
// - All menu commands (logs, community, config, configure, diagnose, doctor, flake, learn, logs, mcp-server, neovim-setup, package-repo, machines, build, completion, deps, devenv, explain-option, gc, hardware, interactive, migrate, search, snippets, store, templates) work and produce output.
// - Updated docs/MANUAL.md and README.md to reflect the current state of interactive mode and list all commands.
// - Fixed TestCommandCompleter to expect all valid completions (including 'completion').
// - Do not remove or stub out interactive commands in the future.

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
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
			// Try direct command dispatch first
			if ok, _ := RunDirectCommand(fields[0], fields[1:], os.Stdout); ok {
				continue
			}
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
		utils.FormatKeyValue("🌐 community", "Community resources and support"),
		utils.FormatKeyValue("🔄 completion", "Generate the autocompletion script for the specified shell"),
		utils.FormatKeyValue("⚙️ config", "Manage nixai configuration"),
		utils.FormatKeyValue("🧑‍💻 configure", "Configure NixOS interactively"),
		utils.FormatKeyValue("🔗 deps", "Analyze NixOS configuration dependencies and imports"),
		utils.FormatKeyValue("🧪 devenv", "Create and manage development environments with devenv"),
		utils.FormatKeyValue("🩺 diagnose", "Diagnose NixOS issues"),
		utils.FormatKeyValue("🩻 doctor", "Run NixOS health checks"),
		utils.FormatKeyValue("🖥️ explain-option <option>", "Explain a NixOS option"),
		utils.FormatKeyValue("🧊 flake", "Nix flake utilities"),
		utils.FormatKeyValue("🧹 gc", "AI-powered garbage collection analysis and cleanup"),
		utils.FormatKeyValue("💻 hardware", "AI-powered hardware configuration optimizer"),
		utils.FormatKeyValue("❓ help", "Help about any command"),
		utils.FormatKeyValue("💬 interactive", "Launch interactive AI-powered NixOS assistant shell"),
		utils.FormatKeyValue("📚 learn", "NixOS learning and training commands"),
		utils.FormatKeyValue("📝 logs", "Analyze and parse NixOS logs"),
		utils.FormatKeyValue("🖧 machines", "Manage and synchronize NixOS configurations across multiple machines"),
		utils.FormatKeyValue("🛰️ mcp-server", "Start or manage the MCP server"),
		utils.FormatKeyValue("🔀 migrate", "AI-powered migration assistant for channels and flakes"),
		utils.FormatKeyValue("📝 neovim-setup", "Neovim integration setup"),
		utils.FormatKeyValue("📦 package-repo <url>", "Analyze Git repos and generate Nix derivations"),
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
	provider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		return "", err
	}
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
