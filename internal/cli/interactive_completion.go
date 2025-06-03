package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"nix-ai-help/pkg/utils"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// CommandCompleter provides tab completion for nixai commands
type CommandCompleter struct {
	commands    map[string]*cobra.Command
	subcommands map[string][]string
}

// NewCommandCompleter creates a new command completer
func NewCommandCompleter() *CommandCompleter {
	cc := &CommandCompleter{
		commands:    make(map[string]*cobra.Command),
		subcommands: make(map[string][]string),
	}

	// Build command map from root command
	for _, cmd := range rootCmd.Commands() {
		cc.commands[cmd.Name()] = cmd
	}

	// Define known subcommands for enhanced completion
	cc.subcommands = map[string][]string{
		"community":    {"forums", "docs", "matrix", "github"},
		"configure":    {"wizard", "hardware", "desktop", "services", "users"},
		"diagnose":     {"system", "config", "services", "network", "hardware", "performance"},
		"doctor":       {"full", "quick", "store", "config", "security"},
		"flake":        {"init", "check", "show", "update", "template", "convert"},
		"learn":        {"basics", "flakes", "packages", "services", "advanced", "troubleshooting"},
		"logs":         {"system", "boot", "service", "errors", "build", "analyze"},
		"mcp-server":   {"start", "stop", "status", "logs", "config"},
		"neovim-setup": {"install", "configure", "test", "update", "remove"},
		"package-repo": {"analyze", "generate", "template", "validate"},
		"build":        {"troubleshoot", "optimize", "fix", "analyze"},
		"config":       {"show", "set", "get", "reset"},
		"devenv":       {"list", "create", "suggest"},
		"gc":           {"analyze", "clean", "safe-clean"},
		"hardware":     {"analyze", "optimize", "detect"},
		"machines":     {"list", "add", "sync", "remove"},
		"migrate":      {"analyze", "channels", "flakes"},
		"search":       {}, // Takes package names, no fixed subcommands
		"snippets":     {"list", "add", "remove", "edit"},
		"store":        {"analyze", "backup", "restore", "gc"},
		"templates":    {"list", "apply", "create", "remove"},
	}

	return cc
}

// Complete implements readline.AutoCompleter interface
func (cc *CommandCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])
	fields := strings.Fields(lineStr)

	var completions []string

	if len(fields) == 0 || (len(fields) == 1 && !strings.HasSuffix(lineStr, " ")) {
		// Complete command names
		for cmd := range cc.commands {
			if strings.HasPrefix(cmd, lineStr) {
				completions = append(completions, cmd)
			}
		}

		// Add special commands
		specialCommands := []string{"help", "exit", "quit", "?"}
		for _, cmd := range specialCommands {
			if strings.HasPrefix(cmd, lineStr) {
				completions = append(completions, cmd)
			}
		}
	} else if len(fields) >= 1 {
		// Complete subcommands
		cmdName := fields[0]
		if subcmds, exists := cc.subcommands[cmdName]; exists {
			var partial string
			if len(fields) > 1 && !strings.HasSuffix(lineStr, " ") {
				partial = fields[len(fields)-1]
			}

			for _, subcmd := range subcmds {
				if partial == "" || strings.HasPrefix(subcmd, partial) {
					completions = append(completions, subcmd)
				}
			}
		}
	}

	// Sort completions
	sort.Strings(completions)

	// Convert to readline format
	if len(completions) > 0 {
		newLine = make([][]rune, len(completions))
		for i, completion := range completions {
			newLine[i] = []rune(completion)
		}

		// Calculate the length to replace
		if len(fields) > 0 && !strings.HasSuffix(lineStr, " ") {
			length = len(fields[len(fields)-1])
		}
	}

	return newLine, length
}

// InteractiveModeWithCompletion starts an enhanced interactive mode with tab completion
func InteractiveModeWithCompletion() {
	debug := os.Getenv("NIXAI_DEBUG") == "1"

	// Create completer
	completer := NewCommandCompleter()

	// Configure readline
	config := &readline.Config{
		Prompt:            "\033[31mnixai>\033[0m ",
		HistoryFile:       "/tmp/nixai_history",
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		// Fallback to basic interactive mode if readline fails
		fmt.Printf("Tab completion unavailable: %v\n", err)
		InteractiveMode()
		return
	}
	defer rl.Close()

	// Print welcome message
	printInteractiveWelcome()
	fmt.Println(utils.FormatTip("Tab completion is enabled! Press Tab to see available commands and options."))
	fmt.Println(utils.FormatDivider())

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else {
				break
			}
		}

		input := strings.TrimSpace(line)
		if debug {
			fmt.Println(utils.FormatInfo("DEBUG: Executing: " + input))
		}

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println(utils.FormatDivider() + "\nGoodbye! 👋")
			break
		}

		if input == "help" || input == "?" {
			printInteractiveWelcome()
			continue
		}

		// Process the command using existing logic
		processInteractiveCommand(input, debug)
	}
}

// processInteractiveCommand handles command execution logic (extracted from original InteractiveMode)
func processInteractiveCommand(input string, debug bool) {
	fields := parseCommandArgs(input)
	if len(fields) == 0 {
		return
	}

	// Support: 'nixai interactive <command> ...' (e.g., 'nixai interactive store')
	if fields[0] == "interactive" && len(fields) > 1 {
		fields = fields[1:]
	}

	// Build command map dynamically from root command
	knownCommands := make(map[string]*cobra.Command)
	for _, cmd := range rootCmd.Commands() {
		knownCommands[cmd.Name()] = cmd
	}

	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Available commands in interactive mode: ")
		for name := range knownCommands {
			fmt.Fprintf(os.Stderr, "%s, ", name)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Handle special interactive-only commands first
	switch fields[0] {
	case "help", "?":
		printInteractiveWelcome()
		return
	case "exit", "quit":
		fmt.Println(utils.FormatDivider() + "\nGoodbye! 👋")
		os.Exit(0)
	case "interactive":
		fmt.Println(utils.FormatTip("You are already in interactive mode!"))
		return
	default:
		// First try using the direct command execution for better interactive support
		var directBuf strings.Builder
		cmdHandled, _ := RunDirectCommand(fields[0], fields[1:], &directBuf)

		if cmdHandled {
			output := directBuf.String()
			if strings.TrimSpace(output) != "" {
				fmt.Println(output)
			}
		} else if cmd, ok := knownCommands[fields[0]]; ok {
			// Fall back to the standard command execution
			if debug {
				fmt.Fprintf(os.Stderr, "DEBUG: Found command '%s' in knownCommands\n", fields[0])
			}
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

// SetOutput allows setting a custom output for the completer (useful for testing)
func (cc *CommandCompleter) SetOutput(w io.Writer) {
	// This method can be used for testing or custom output handling
}
