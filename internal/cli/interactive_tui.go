package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// LaunchTUIMode launches TUI mode with support for any command and parameters
func LaunchTUIMode(cmd *cobra.Command, args []string) error {
	var initialCommand string
	var initialArgs []string

	// If specific command and args are provided, use them
	if len(args) > 0 {
		initialCommand = args[0]
		if len(args) > 1 {
			initialArgs = args[1:]
		}
	}

	// Create the TUI application with initial command context
	app := tea.NewProgram(
		initialModelWithCommand(initialCommand, initialArgs),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the TUI
	if _, err := app.Run(); err != nil {
		return fmt.Errorf("error running TUI: %v", err)
	}

	return nil
}

// initialModelWithCommand creates initial model with optional pre-selected command
func initialModelWithCommand(command string, args []string) tuiModel {
	model := initialModel()

	// If a command is specified, find and select it
	if command != "" {
		for i, cmd := range model.commands {
			if cmd.name == command {
				model.selectedCommand = i
				// Store initial command and args for execution in Init
				model.selectedCmdName = command
				if len(args) > 0 {
					// Store args in optionValues for later use
					if model.optionValues == nil {
						model.optionValues = make(map[string]string)
					}
					model.optionValues["__initial_args__"] = strings.Join(args, " ")
					model.commandOutput = fmt.Sprintf("Preparing to execute: %s %s", command, strings.Join(args, " "))
				}
				break
			}
		}
	}

	return model
}

// InteractiveModeTUI starts the modern TUI interface for nixai
func InteractiveModeTUI() {
	// Create the TUI application without AltScreen to avoid terminal compatibility issues
	app := tea.NewProgram(
		initialModel(),
		tea.WithMouseCellMotion(),
	)

	// Run the TUI
	if _, err := app.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// tuiModel represents the state of our TUI application
type tuiModel struct {
	commands           []commandItem
	selectedCommand    int
	selectedSubcommand int
	commandOutput      string
	isExecuting        bool
	focused            focusedPanel
	terminalWidth      int
	terminalHeight     int
	searchQuery        string
	searchMode         bool
	parameterInput     string
	inputMode          bool
	selectedCmdName    string
	currentState       tuiState
	commandOptions     []commandOption
	selectedOption     int
	optionValues       map[string]string

	// Streaming output support
	streamingOutput []string
	isStreaming     bool
	currentCommand  string
}

type commandItem struct {
	name        string
	description string
	needsInput  bool
	options     []commandOption
	subcommands []subcommandItem
}

type subcommandItem struct {
	name        string
	description string
	options     []commandOption
}

type commandOption struct {
	name         string
	flag         string
	description  string
	required     bool
	hasValue     bool
	defaultValue string
	optionType   string // "string", "bool", "int"
}

type tuiState int

const (
	stateCommandList tuiState = iota
	stateSubcommandSelection
	stateCommandOptions
	stateExecuting
	stateResults
)

type focusedPanel int

const (
	focusCommands focusedPanel = iota
	focusSubcommands
	focusOptions
	focusOutput
	focusInput
)

// executeCommandMsg represents a command execution result
type executeCommandMsg struct {
	command string
	output  string
}

// streamingOutputMsg represents streaming command output
type streamingOutputMsg struct {
	command string
	output  string
	isEnd   bool
}

// commandExecutionStartMsg represents the start of command execution
type commandExecutionStartMsg struct {
	command string
}

// Define styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7ebae4")).
			MarginLeft(1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#414868"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7ebae4")).
			Foreground(lipgloss.Color("#1a1b26")).
			Bold(true)

	commandStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272a4")).
				Italic(true)

	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#414868")).
			Foreground(lipgloss.Color("#a9b1d6")).
			PaddingLeft(1).
			PaddingRight(1)
)

// initialModel creates the initial state for our TUI
func initialModel() tuiModel {
	// Get available commands from the root command
	commands := getAvailableCommands()

	return tuiModel{
		commands:        commands,
		selectedCommand: 0,
		focused:         focusCommands,
		commandOutput:   "Welcome to nixai TUI! Select a command from the left panel to get started.",
		currentState:    stateCommandList,
		optionValues:    make(map[string]string),
		streamingOutput: make([]string, 0),
		isStreaming:     false,
		currentCommand:  "",
	}
}

// getAvailableCommands returns a list of available nixai commands
func getAvailableCommands() []commandItem {
	commands := []commandItem{
		{
			name:        "ask",
			description: "Ask any NixOS question",
			needsInput:  true,
			options: []commandOption{
				{name: "Question", flag: "question", description: "Your NixOS question", required: true, hasValue: true, optionType: "string"},
				{name: "Agent", flag: "agent", description: "AI provider (ollama, openai, gemini)", required: false, hasValue: true, defaultValue: "ollama", optionType: "string"},
				{name: "Role", flag: "role", description: "Agent role (diagnoser, explainer, etc.)", required: false, hasValue: true, optionType: "string"},
			},
			subcommands: []subcommandItem{},
		},
		{
			name:        "search",
			description: "Search for NixOS packages/services",
			needsInput:  true,
			options: []commandOption{
				{name: "Package", flag: "package", description: "Package name to search", required: true, hasValue: true, optionType: "string"},
				{name: "Channel", flag: "channel", description: "NixOS channel (stable, unstable)", required: false, hasValue: true, defaultValue: "unstable", optionType: "string"},
			},
			subcommands: []subcommandItem{},
		},
		{
			name:        "explain-option",
			description: "Explain a NixOS option",
			needsInput:  true,
			options: []commandOption{
				{name: "Option", flag: "option", description: "NixOS option to explain", required: true, hasValue: true, optionType: "string"},
				{name: "Format", flag: "format", description: "Output format (markdown, plain, table)", required: false, hasValue: true, defaultValue: "markdown", optionType: "string"},
				{name: "Examples Only", flag: "examples-only", description: "Show only usage examples", required: false, hasValue: false, optionType: "bool"},
			},
			subcommands: []subcommandItem{},
		},
		{
			name:        "community",
			description: "Community resources and support",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "search", description: "Search community configurations", options: []commandOption{}},
				{name: "share", description: "Share your configuration", options: []commandOption{}},
				{name: "validate", description: "Validate your configuration", options: []commandOption{}},
				{name: "trends", description: "Show community trends", options: []commandOption{}},
				{name: "rate", description: "Rate configurations", options: []commandOption{}},
			},
		},
		{
			name:        "devenv",
			description: "Create and manage development environments",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "list", description: "List available templates", options: []commandOption{}},
				{name: "create", description: "Create new development environment", options: []commandOption{}},
				{name: "suggest", description: "Get AI template suggestions", options: []commandOption{}},
			},
		},
		{
			name:        "mcp-server",
			description: "Start or manage the MCP server",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "start", description: "Start the MCP server", options: []commandOption{}},
				{name: "stop", description: "Stop the MCP server", options: []commandOption{}},
				{name: "status", description: "Check server status", options: []commandOption{}},
				{name: "restart", description: "Restart the server", options: []commandOption{}},
				{name: "query", description: "Query documentation", options: []commandOption{}},
			},
		},
		{
			name:        "machines",
			description: "Manage configurations across multiple machines",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "list", description: "List configured machines", options: []commandOption{}},
				{name: "deploy", description: "Deploy configurations", options: []commandOption{}},
				{name: "setup-deploy-rs", description: "Setup deploy-rs", options: []commandOption{}},
			},
		},
		{
			name:        "doctor",
			description: "Run comprehensive NixOS health checks and get AI-powered diagnostics",
			needsInput:  false,
			options: []commandOption{
				{name: "Verbose", flag: "verbose", description: "Show detailed output and progress information", required: false, hasValue: false, optionType: "bool"},
			},
			subcommands: []subcommandItem{
				{name: "system", description: "Core system health checks", options: []commandOption{}},
				{name: "nixos", description: "NixOS-specific configuration checks", options: []commandOption{}},
				{name: "packages", description: "Package and store integrity checks", options: []commandOption{}},
				{name: "services", description: "System service status checks", options: []commandOption{}},
				{name: "storage", description: "Storage and filesystem checks", options: []commandOption{}},
				{name: "network", description: "Network connectivity checks", options: []commandOption{}},
				{name: "security", description: "Security configuration checks", options: []commandOption{}},
				{name: "all", description: "Run all available checks (default)", options: []commandOption{}},
			},
		},
		{
			name:        "flake",
			description: "Nix flake utilities",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "init", description: "Initialize new flake", options: []commandOption{}},
				{name: "check", description: "Check flake validity", options: []commandOption{}},
				{name: "show", description: "Show flake outputs", options: []commandOption{}},
				{name: "update", description: "Update flake inputs", options: []commandOption{}},
				{name: "template", description: "Create from template", options: []commandOption{}},
				{name: "convert", description: "Convert to flake", options: []commandOption{}},
			},
		},
		{
			name:        "learn",
			description: "NixOS learning and training commands",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "basics", description: "Learn NixOS basics", options: []commandOption{}},
				{name: "flakes", description: "Learn about flakes", options: []commandOption{}},
				{name: "packages", description: "Learn package management", options: []commandOption{}},
				{name: "services", description: "Learn service configuration", options: []commandOption{}},
				{name: "advanced", description: "Advanced topics", options: []commandOption{}},
				{name: "troubleshooting", description: "Troubleshooting guide", options: []commandOption{}},
			},
		},
		{
			name:        "logs",
			description: "Analyze and parse NixOS logs",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "system", description: "System logs", options: []commandOption{}},
				{name: "boot", description: "Boot logs", options: []commandOption{}},
				{name: "service", description: "Service logs", options: []commandOption{}},
				{name: "errors", description: "Error logs", options: []commandOption{}},
				{name: "build", description: "Build logs", options: []commandOption{}},
				{name: "analyze", description: "Analyze logs with AI", options: []commandOption{}},
			},
		},
		{
			name:        "templates",
			description: "Manage NixOS configuration templates",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "list", description: "List templates", options: []commandOption{}},
				{name: "show", description: "Show template", options: []commandOption{}},
				{name: "apply", description: "Apply template", options: []commandOption{}},
				{name: "search", description: "Search templates", options: []commandOption{}},
				{name: "save", description: "Save template", options: []commandOption{}},
				{name: "categories", description: "List categories", options: []commandOption{}},
			},
		},
		{
			name:        "snippets",
			description: "Manage NixOS configuration snippets",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "list", description: "List snippets", options: []commandOption{}},
				{name: "add", description: "Add snippet", options: []commandOption{}},
				{name: "show", description: "Show snippet", options: []commandOption{}},
				{name: "remove", description: "Remove snippet", options: []commandOption{}},
				{name: "search", description: "Search snippets", options: []commandOption{}},
			},
		},
		{
			name:        "store",
			description: "Manage, backup, and analyze the Nix store",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "backup", description: "Backup store", options: []commandOption{}},
				{name: "restore", description: "Restore store", options: []commandOption{}},
				{name: "integrity", description: "Check integrity", options: []commandOption{}},
				{name: "performance", description: "Analyze performance", options: []commandOption{}},
			},
		},
		{
			name:        "deps",
			description: "Analyze NixOS configuration dependencies",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "analyze", description: "Analyze dependencies", options: []commandOption{}},
				{name: "why", description: "Explain package inclusion", options: []commandOption{}},
				{name: "conflicts", description: "Find conflicts", options: []commandOption{}},
				{name: "optimize", description: "Optimize dependencies", options: []commandOption{}},
				{name: "graph", description: "Generate dependency graph", options: []commandOption{}},
			},
		},
		{
			name:        "build",
			description: "Enhanced build troubleshooting and optimization",
			needsInput:  false,
			options:     []commandOption{},
			subcommands: []subcommandItem{
				{name: "debug", description: "Deep build failure analysis with pattern recognition", options: []commandOption{}},
				{name: "retry", description: "Intelligent retry with automated fixes", options: []commandOption{}},
				{name: "cache-miss", description: "Analyze cache miss reasons and optimization", options: []commandOption{}},
				{name: "sandbox-debug", description: "Debug sandbox-related build issues", options: []commandOption{}},
				{name: "profile", description: "Build performance analysis and optimization", options: []commandOption{}},
				{name: "watch", description: "Real-time build monitoring with AI insights", options: []commandOption{}},
				{name: "status", description: "Check status of background builds", options: []commandOption{}},
				{name: "stop", description: "Cancel a running background build", options: []commandOption{}},
				{name: "background", description: "Start a build in the background", options: []commandOption{}},
				{name: "queue", description: "Build multiple packages sequentially", options: []commandOption{}},
			},
		},
		// Simple commands without subcommands
		{
			name:        "package-repo",
			description: "Analyze Git repos and generate Nix derivations",
			needsInput:  true,
			options: []commandOption{
				{name: "Repository URL", flag: "repo-url", description: "Git repository URL to analyze", required: false, hasValue: true, optionType: "string"},
				{name: "Local Path", flag: "local", description: "Local repository path", required: false, hasValue: true, optionType: "string"},
				{name: "Output Path", flag: "output", description: "Output file path for derivation", required: false, hasValue: true, optionType: "string"},
				{name: "Package Name", flag: "name", description: "Custom package name", required: false, hasValue: true, optionType: "string"},
				{name: "Analyze Only", flag: "analyze-only", description: "Only analyze, don't generate derivation", required: false, hasValue: false, optionType: "bool"},
			},
			subcommands: []subcommandItem{},
		},
		{name: "diagnose", description: "Diagnose NixOS issues", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
		{name: "config", description: "Manage nixai configuration", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
		{name: "configure", description: "Configure NixOS interactively", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
		{name: "gc", description: "AI-powered garbage collection analysis", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
		{name: "hardware", description: "AI-powered hardware configuration optimizer", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
		{name: "migrate", description: "AI-powered migration assistant", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
		{name: "neovim-setup", description: "Neovim integration setup", needsInput: false, options: []commandOption{}, subcommands: []subcommandItem{}},
	}

	return commands
}

// Init is called when the model is first created
func (m tuiModel) Init() tea.Cmd {
	// If we have initial command args, execute the command immediately
	if args, exists := m.optionValues["__initial_args__"]; exists && m.selectedCmdName != "" {
		// Remove the initial args marker
		delete(m.optionValues, "__initial_args__")

		// Parse the args and execute
		argList := strings.Fields(args)
		return tea.Batch(
			func() tea.Msg {
				return commandExecutionStartMsg{
					command: fmt.Sprintf("%s %s", m.selectedCmdName, args),
				}
			},
			m.executeCommandWithParams(m.selectedCmdName, argList),
		)
	}
	return nil
}

// Update handles all incoming messages and updates the model state
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height

	case executeCommandMsg:
		m.commandOutput = msg.output
		m.isExecuting = false
		m.currentState = stateResults
		m.focused = focusOutput

	case commandExecutionStartMsg:
		m.isStreaming = true
		m.isExecuting = true
		m.currentCommand = msg.command
		m.streamingOutput = []string{fmt.Sprintf("$ %s", msg.command)}
		m.commandOutput = strings.Join(m.streamingOutput, "\n")
		m.currentState = stateExecuting

	case streamingOutputMsg:
		if m.isStreaming && msg.command == m.currentCommand {
			if msg.isEnd {
				m.isStreaming = false
				m.isExecuting = false
				m.currentState = stateResults
				m.focused = focusOutput
			} else {
				m.streamingOutput = append(m.streamingOutput, msg.output)
				m.commandOutput = strings.Join(m.streamingOutput, "\n")
			}
		}

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles key presses based on current state
func (m tuiModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle text input first when in input modes
	if m.inputMode || m.searchMode {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.handleEscape(), nil
		case "enter":
			return m.handleEnter()
		case "backspace":
			return m.handleBackspace(), nil
		default:
			return m.handleTextInput(msg), nil
		}
	}

	// Handle other keys when not in input mode
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "tab":
		return m.handleTabNavigation(), nil

	case "esc":
		return m.handleEscape(), nil

	case "enter":
		return m.handleEnter()

	// Arrow keys (always safe for navigation)
	case "up":
		return m.handleUpNavigation(), nil
	case "down":
		return m.handleDownNavigation(), nil
	case "left":
		return m.handleLeftNavigation(), nil
	case "right":
		return m.handleRightNavigation(), nil

	// Vim-style navigation with Ctrl (safe from text input conflicts)
	case "ctrl+k":
		return m.handleUpNavigation(), nil
	case "ctrl+j":
		return m.handleDownNavigation(), nil
	case "ctrl+h":
		return m.handleLeftNavigation(), nil
	case "ctrl+l":
		return m.handleRightNavigation(), nil

	// Execute command shortcuts
	case "ctrl+r", "ctrl+enter":
		// Quick execute shortcut in options state
		if m.currentState == stateCommandOptions {
			filteredCommands := m.filterCommands()
			if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
				cmd := filteredCommands[m.selectedCommand]
				args := m.buildCommandArgs()
				m.isExecuting = true
				m.currentState = stateExecuting
				return m, m.executeCommandWithParams(cmd.name, args)
			}
		}
		return m, nil

	case "/":
		return m.handleSearchMode(), nil

	case "backspace":
		return m.handleBackspace(), nil

	default:
		// Allow all other keys to be used for text input when appropriate
		return m.handleTextInput(msg), nil
	}
}

// handleTabNavigation switches focus between panels based on current state
func (m tuiModel) handleTabNavigation() tuiModel {
	switch m.currentState {
	case stateCommandList:
		if m.searchMode {
			return m
		}
		// Switch between commands and output
		if m.focused == focusCommands {
			m.focused = focusOutput
		} else {
			m.focused = focusCommands
		}

	case stateSubcommandSelection:
		// Switch between commands, subcommands, and output
		switch m.focused {
		case focusCommands:
			m.focused = focusSubcommands
		case focusSubcommands:
			m.focused = focusOutput
		case focusOutput:
			m.focused = focusCommands
		}

	case stateCommandOptions:
		// Switch between commands, options, and output
		switch m.focused {
		case focusCommands:
			m.focused = focusOptions
		case focusOptions:
			m.focused = focusOutput
		case focusOutput:
			m.focused = focusCommands
		}

	case stateResults:
		// In results state, tab goes back to command selection
		m.currentState = stateCommandList
		m.focused = focusCommands
		m.commandOutput = "Select a command to execute"
	}

	return m
}

// handleEscape handles escape key based on current state
func (m tuiModel) handleEscape() tuiModel {
	if m.searchMode {
		m.searchMode = false
		m.searchQuery = ""
		return m
	}

	if m.inputMode {
		m.inputMode = false
		m.parameterInput = ""
		m.focused = focusCommands
		return m
	}

	switch m.currentState {
	case stateSubcommandSelection:
		m.currentState = stateCommandList
		m.focused = focusCommands
		m.selectedSubcommand = 0
		m.commandOutput = "Select a command to execute"

	case stateCommandOptions:
		// Check if we came from subcommand selection
		filteredCommands := m.filterCommands()
		if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
			cmd := filteredCommands[m.selectedCommand]
			if len(cmd.subcommands) > 0 {
				// Go back to subcommand selection
				m.currentState = stateSubcommandSelection
				m.focused = focusSubcommands
			} else {
				// Go back to command list
				m.currentState = stateCommandList
				m.focused = focusCommands
			}
		} else {
			m.currentState = stateCommandList
			m.focused = focusCommands
		}
		m.optionValues = make(map[string]string)
		m.selectedOption = 0

	case stateResults:
		m.currentState = stateCommandList
		m.focused = focusCommands
		m.commandOutput = "Select a command to execute"
	}

	return m
}

// handleEnter handles enter key based on current state and focus
func (m tuiModel) handleEnter() (tuiModel, tea.Cmd) {
	// Handle input mode first, regardless of state
	if m.inputMode && len(m.commandOptions) > 0 && m.selectedOption < len(m.commandOptions) {
		opt := m.commandOptions[m.selectedOption]
		m.optionValues[opt.flag] = m.parameterInput
		m.inputMode = false
		m.parameterInput = ""
		m.focused = focusOptions
		m.commandOutput = fmt.Sprintf("Set '%s' to: %s", opt.name, m.optionValues[opt.flag])
		return m, nil
	}

	switch m.currentState {
	case stateCommandList:
		if m.focused == focusCommands && !m.searchMode {
			filteredCommands := m.filterCommands()
			if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
				cmd := filteredCommands[m.selectedCommand]

				// If command has subcommands, show subcommand selection
				if len(cmd.subcommands) > 0 {
					m.currentState = stateSubcommandSelection
					m.focused = focusSubcommands
					m.selectedSubcommand = 0
					m.commandOutput = fmt.Sprintf("Select a subcommand for '%s'\n\nInstructions:\n• Use ↑↓ to navigate subcommands\n• Press Enter to select a subcommand\n• Press Esc to go back", cmd.name)
				} else if len(cmd.options) > 0 {
					// If command has options, show options panel
					m.currentState = stateCommandOptions
					m.focused = focusOptions
					m.commandOptions = cmd.options
					m.selectedOption = 0
					m.optionValues = make(map[string]string)

					// Set default values
					for _, opt := range cmd.options {
						if opt.defaultValue != "" {
							m.optionValues[opt.flag] = opt.defaultValue
						}
					}

					m.commandOutput = fmt.Sprintf("Configure options for '%s' command\n\nInstructions:\n• Use Tab to switch between panels\n• Use ↑↓ to navigate options\n• Press Enter on an option to configure it\n• Press Enter on command panel to execute with current options", cmd.name)
				} else {
					// Execute command immediately if no options or subcommands
					m.isExecuting = true
					m.currentState = stateExecuting
					return m, m.executeCommand(cmd.name)
				}
			}
		}

	case stateSubcommandSelection:
		if m.focused == focusSubcommands {
			filteredCommands := m.filterCommands()
			if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
				cmd := filteredCommands[m.selectedCommand]
				if m.selectedSubcommand >= 0 && m.selectedSubcommand < len(cmd.subcommands) {
					subcmd := cmd.subcommands[m.selectedSubcommand]

					// If subcommand has options, show options panel
					if len(subcmd.options) > 0 {
						m.currentState = stateCommandOptions
						m.focused = focusOptions
						m.commandOptions = subcmd.options
						m.selectedOption = 0
						m.optionValues = make(map[string]string)

						// Set default values
						for _, opt := range subcmd.options {
							if opt.defaultValue != "" {
								m.optionValues[opt.flag] = opt.defaultValue
							}
						}

						m.commandOutput = fmt.Sprintf("Configure options for '%s %s' command\n\nInstructions:\n• Use Tab to switch between panels\n• Use ↑↓ to navigate options\n• Press Enter on an option to configure it\n• Press Enter on command panel to execute with current options", cmd.name, subcmd.name)
					} else {
						// Execute subcommand immediately if no options
						m.isExecuting = true
						m.currentState = stateExecuting
						return m, m.executeCommandWithSubcommand(cmd.name, subcmd.name, []string{})
					}
				}
			}
		}

	case stateCommandOptions:
		if m.focused == focusOptions {
			// Check if "Execute Command" option is selected
			if m.selectedOption == len(m.commandOptions) {
				// Execute command with current options
				filteredCommands := m.filterCommands()
				if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
					cmd := filteredCommands[m.selectedCommand]
					args := m.buildCommandArgs()
					m.isExecuting = true
					m.currentState = stateExecuting
					return m, m.executeCommandWithParams(cmd.name, args)
				}
			} else if len(m.commandOptions) > 0 && m.selectedOption < len(m.commandOptions) {
				// Handle individual option configuration
				opt := m.commandOptions[m.selectedOption]
				if opt.optionType == "bool" {
					// Toggle boolean value
					current := m.optionValues[opt.flag]
					if current == "true" {
						m.optionValues[opt.flag] = "false"
					} else {
						m.optionValues[opt.flag] = "true"
					}
					m.commandOutput = fmt.Sprintf("Toggled '%s' to: %s", opt.name, m.optionValues[opt.flag])
				} else {
					// Enter text input mode for string/int values
					m.inputMode = true
					m.parameterInput = m.optionValues[opt.flag]
					m.focused = focusInput
					m.commandOutput = fmt.Sprintf("Entering input mode for '%s'. Type your value and press Enter to confirm.", opt.name)
				}
			}
		} else if m.focused == focusCommands {
			// Execute command with configured options
			filteredCommands := m.filterCommands()
			if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
				cmd := filteredCommands[m.selectedCommand]
				args := m.buildCommandArgs()
				m.isExecuting = true
				m.currentState = stateExecuting
				return m, m.executeCommandWithParams(cmd.name, args)
			}
		}

	case stateResults:
		// In results state, enter goes back to command selection
		m.currentState = stateCommandList
		m.focused = focusCommands
		m.commandOutput = "Select a command to execute"
	}

	return m, nil
}

// buildCommandArgs builds command arguments from option values
func (m tuiModel) buildCommandArgs() []string {
	var args []string

	for _, opt := range m.commandOptions {
		value := m.optionValues[opt.flag]
		if value != "" {
			if opt.optionType == "bool" && value == "true" {
				args = append(args, "--"+opt.flag)
			} else if opt.optionType != "bool" {
				args = append(args, "--"+opt.flag, value)
			}
		}
	}

	return args
}

// Navigation helpers
func (m tuiModel) handleUpNavigation() tuiModel {
	if m.inputMode || m.searchMode {
		return m // Don't navigate when in input/search mode
	}

	switch m.currentState {
	case stateCommandList:
		if m.focused == focusCommands {
			if m.selectedCommand > 0 {
				m.selectedCommand--
			}
		}
	case stateSubcommandSelection:
		if m.focused == focusSubcommands {
			if m.selectedSubcommand > 0 {
				m.selectedSubcommand--
			}
		}
	case stateCommandOptions:
		if m.focused == focusOptions {
			if m.selectedOption > 0 {
				m.selectedOption--
			}
		} else if m.focused == focusCommands {
			if m.selectedCommand > 0 {
				m.selectedCommand--
			}
		}
	}
	return m
}

func (m tuiModel) handleDownNavigation() tuiModel {
	if m.inputMode || m.searchMode {
		return m // Don't navigate when in input/search mode
	}

	switch m.currentState {
	case stateCommandList:
		if m.focused == focusCommands {
			filteredCommands := m.filterCommands()
			if m.selectedCommand < len(filteredCommands)-1 {
				m.selectedCommand++
			}
		}
	case stateSubcommandSelection:
		if m.focused == focusSubcommands {
			filteredCommands := m.filterCommands()
			if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
				cmd := filteredCommands[m.selectedCommand]
				if m.selectedSubcommand < len(cmd.subcommands)-1 {
					m.selectedSubcommand++
				}
			}
		}
	case stateCommandOptions:
		if m.focused == focusOptions {
			// Include the "Execute Command" option at the bottom
			maxOptions := len(m.commandOptions) // Execute option is at the end
			if m.selectedOption < maxOptions {
				m.selectedOption++
			}
		} else if m.focused == focusCommands {
			filteredCommands := m.filterCommands()
			if m.selectedCommand < len(filteredCommands)-1 {
				m.selectedCommand++
			}
		}
	}
	return m
}

func (m tuiModel) handleLeftNavigation() tuiModel {
	if m.inputMode || m.searchMode {
		return m
	}

	switch m.currentState {
	case stateCommandOptions:
		m.focused = focusCommands
	case stateCommandList:
		// Left does nothing in command list state
	}
	return m
}

func (m tuiModel) handleRightNavigation() tuiModel {
	if m.inputMode || m.searchMode {
		return m
	}

	switch m.currentState {
	case stateCommandOptions:
		m.focused = focusOptions
	case stateCommandList:
		m.focused = focusOutput
	}
	return m
}

func (m tuiModel) handleSearchMode() tuiModel {
	if m.currentState == stateCommandList && m.focused == focusCommands {
		m.searchMode = true
		m.searchQuery = ""
	}
	return m
}

func (m tuiModel) handleBackspace() tuiModel {
	if m.inputMode && len(m.parameterInput) > 0 {
		m.parameterInput = m.parameterInput[:len(m.parameterInput)-1]
	} else if m.searchMode && len(m.searchQuery) > 0 {
		m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		m.selectedCommand = 0
	}
	return m
}

func (m tuiModel) handleTextInput(msg tea.KeyMsg) tuiModel {
	// Accept single characters, numbers, and common symbols for input
	key := msg.String()
	if len(key) == 1 || key == "space" {
		if key == "space" {
			key = " "
		}

		if m.inputMode {
			m.parameterInput += key
		} else if m.searchMode {
			m.searchQuery += key
			m.selectedCommand = 0
		}
	}
	return m
}

// View renders the TUI interface
func (m tuiModel) View() string {
	if m.terminalWidth == 0 {
		return "Loading..."
	}

	// Calculate panel dimensions based on current state
	var leftPanel, rightPanel string
	var title string

	switch m.currentState {
	case stateCommandList:
		// Two-panel layout: Commands + Output
		leftPanelWidth := m.terminalWidth * 40 / 100
		rightPanelWidth := m.terminalWidth - leftPanelWidth - 4
		panelHeight := m.terminalHeight - 4

		leftPanel = m.renderCommandsPanel(leftPanelWidth, panelHeight)
		rightPanel = m.renderOutputPanel(rightPanelWidth, panelHeight)
		title = "❄️ nixai: NixOS AI Assistant - Select Command"

	case stateSubcommandSelection:
		// Three-panel layout: Commands + Subcommands + Output
		leftPanelWidth := m.terminalWidth * 30 / 100
		middlePanelWidth := m.terminalWidth * 30 / 100
		rightPanelWidth := m.terminalWidth - leftPanelWidth - middlePanelWidth - 6
		panelHeight := m.terminalHeight - 4

		leftPanel = m.renderCommandsPanel(leftPanelWidth, panelHeight)
		middlePanel := m.renderSubcommandsPanel(middlePanelWidth, panelHeight)
		outputPanel := m.renderOutputPanel(rightPanelWidth, panelHeight)

		rightPanel = lipgloss.JoinHorizontal(lipgloss.Top, middlePanel, outputPanel)
		title = "❄️ nixai: NixOS AI Assistant - Select Subcommand"

	case stateCommandOptions:
		// Three-panel layout: Commands + Options + Output (stack options and output)
		leftPanelWidth := m.terminalWidth * 30 / 100
		rightPanelWidth := m.terminalWidth - leftPanelWidth - 4
		panelHeight := m.terminalHeight - 4

		leftPanel = m.renderCommandsPanel(leftPanelWidth, panelHeight)

		// Stack options and output vertically on the right
		optionsHeight := panelHeight / 2
		outputHeight := panelHeight - optionsHeight

		optionsPanel := m.renderOptionsPanel(rightPanelWidth, optionsHeight)
		outputPanel := m.renderOutputPanel(rightPanelWidth, outputHeight)

		rightPanel = lipgloss.JoinVertical(lipgloss.Left, optionsPanel, outputPanel)
		title = "❄️ nixai: Configure Options"

	case stateExecuting:
		// Two-panel layout with executing indicator
		leftPanelWidth := m.terminalWidth * 40 / 100
		rightPanelWidth := m.terminalWidth - leftPanelWidth - 4
		panelHeight := m.terminalHeight - 4

		leftPanel = m.renderCommandsPanel(leftPanelWidth, panelHeight)
		rightPanel = m.renderOutputPanel(rightPanelWidth, panelHeight)
		title = "❄️ nixai: Executing Command..."

	case stateResults:
		// Two-panel layout showing results
		leftPanelWidth := m.terminalWidth * 40 / 100
		rightPanelWidth := m.terminalWidth - leftPanelWidth - 4
		panelHeight := m.terminalHeight - 4

		leftPanel = m.renderCommandsPanel(leftPanelWidth, panelHeight)
		rightPanel = m.renderOutputPanel(rightPanelWidth, panelHeight)
		title = "❄️ nixai: Command Results (Tab to select new command)"
	}

	// Create the status bar
	statusBar := m.renderStatusBar(m.terminalWidth)

	// Combine panels horizontally
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Create the title
	titleFormatted := titleStyle.Render(title)

	// Combine everything vertically
	return lipgloss.JoinVertical(lipgloss.Left, titleFormatted, mainArea, statusBar)
}

// renderCommandsPanel renders the left commands panel
func (m tuiModel) renderCommandsPanel(width, height int) string {
	var content strings.Builder

	// Show input mode if active
	if m.inputMode {
		inputHeader := fmt.Sprintf("Enter parameter for '%s':", m.selectedCmdName)
		content.WriteString(inputHeader + "\n\n")

		inputLine := fmt.Sprintf("Input: %s_", m.parameterInput)
		if m.focused == focusInput {
			inputLine = selectedStyle.Render(inputLine)
		} else {
			inputLine = commandStyle.Render(inputLine)
		}
		content.WriteString(inputLine + "\n\n")
		content.WriteString("Press Enter to execute, Esc to cancel\n")
	} else if m.searchMode {
		// Add search bar if in search mode
		searchBar := fmt.Sprintf("Search: %s_", m.searchQuery)
		content.WriteString(searchBar + "\n\n")
	} else {
		content.WriteString("Commands (Press / to search):\n\n")
	}

	// Only show command list if not in input mode
	if !m.inputMode {
		// Filter commands based on search query
		filteredCommands := m.filterCommands()

		// Render command list
		for i, cmd := range filteredCommands {
			line := cmd.name

			// Add indicator for commands that need input
			if cmd.needsInput {
				line += " [INPUT]"
			}

			if i == m.selectedCommand && m.focused == focusCommands {
				line = selectedStyle.Render(line)
			} else {
				line = commandStyle.Render(line)
			}

			content.WriteString(line + "\n")

			// Add description on next line if not in search mode
			if !m.searchMode {
				desc := descriptionStyle.Render("  " + cmd.description)
				content.WriteString(desc + "\n")
			}
		}
	}

	// Create the panel with border
	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content.String())

	return panel
}

// renderOutputPanel renders the right output panel
func (m tuiModel) renderOutputPanel(width, height int) string {
	content := m.commandOutput

	if m.isStreaming {
		content = "⚡ Executing command (real-time output)...\n\n" + content
	} else if m.isExecuting {
		content = "⏳ Executing command...\n\n" + content
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// renderOptionsPanel renders the options configuration panel
func (m tuiModel) renderOptionsPanel(width, height int) string {
	var content strings.Builder

	if m.inputMode {
		// Show input mode for current option
		if len(m.commandOptions) > 0 && m.selectedOption < len(m.commandOptions) {
			opt := m.commandOptions[m.selectedOption]
			content.WriteString(fmt.Sprintf("Enter value for '%s':\n", opt.name))
			content.WriteString(fmt.Sprintf("Description: %s\n\n", opt.description))

			inputDisplay := fmt.Sprintf("Value: %s█", m.parameterInput)
			if m.focused == focusInput {
				inputDisplay = selectedStyle.Render(inputDisplay)
			} else {
				inputDisplay = commandStyle.Render(inputDisplay)
			}
			content.WriteString(inputDisplay + "\n\n")
			content.WriteString(descriptionStyle.Render("Press Enter to confirm, Esc to cancel"))
		}
	} else {
		if m.focused == focusOptions {
			content.WriteString("⚡ Options (Enter to configure, ↑↓ to navigate):\n\n")
		} else {
			content.WriteString("Options (Tab to focus here):\n\n")
		}

		for i, opt := range m.commandOptions {
			value := m.optionValues[opt.flag]
			if value == "" && opt.defaultValue != "" {
				value = opt.defaultValue
			}
			if value == "" {
				value = "<not set>"
			}

			var line string
			if opt.required {
				line = fmt.Sprintf("* %s: %s", opt.name, value)
			} else {
				line = fmt.Sprintf("  %s: %s", opt.name, value)
			}

			if i == m.selectedOption && m.focused == focusOptions {
				line = selectedStyle.Render(line)
			} else {
				line = commandStyle.Render(line)
			}

			content.WriteString(line + "\n")

			// Add description
			desc := descriptionStyle.Render(fmt.Sprintf("    %s", opt.description))
			content.WriteString(desc + "\n")
		}

		// Add "Execute Command" option at the bottom
		content.WriteString("\n")
		executeOption := "Execute Command"
		if m.selectedOption == len(m.commandOptions) && m.focused == focusOptions {
			executeOption = selectedStyle.Render(executeOption)
		} else {
			executeOption = commandStyle.Render(executeOption)
		}
		content.WriteString(executeOption + "\n")
		content.WriteString(descriptionStyle.Render("    Run command with configured options") + "\n")

		if len(m.commandOptions) > 0 {
			content.WriteString("\n")
			if m.focused == focusOptions {
				content.WriteString(descriptionStyle.Render("* = required, Enter = configure/execute, ↑↓/Ctrl+jk = navigate, Tab = switch panel"))
			} else {
				content.WriteString(descriptionStyle.Render("* = required, Tab to focus this panel"))
			}
		}
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content.String())

	return panel
}

// renderSubcommandsPanel renders the middle subcommands panel
func (m tuiModel) renderSubcommandsPanel(width, height int) string {
	var content strings.Builder

	// Show subcommand selection
	if m.focused == focusSubcommands {
		content.WriteString("⚡ Subcommands (Enter to select, ↑↓ to navigate):\n\n")
	} else {
		content.WriteString("Subcommands (Tab to focus here):\n\n")
	}

	// Get current command to show its subcommands
	filteredCommands := m.filterCommands()
	if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
		cmd := filteredCommands[m.selectedCommand]

		for i, subcmd := range cmd.subcommands {
			line := fmt.Sprintf("  %s", subcmd.name)

			if i == m.selectedSubcommand && m.focused == focusSubcommands {
				line = selectedStyle.Render(line)
			} else {
				line = commandStyle.Render(line)
			}

			content.WriteString(line + "\n")

			// Add description on next line
			desc := descriptionStyle.Render(fmt.Sprintf("    %s", subcmd.description))
			content.WriteString(desc + "\n")
		}

		if len(cmd.subcommands) > 0 {
			content.WriteString("\n")
			if m.focused == focusSubcommands {
				content.WriteString(descriptionStyle.Render("Enter = select, ↑↓/Ctrl+jk = navigate, Tab = switch panel"))
			} else {
				content.WriteString(descriptionStyle.Render("Tab to focus this panel"))
			}
		}
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content.String())

	return panel
}

// renderStatusBar renders the bottom status bar
func (m tuiModel) renderStatusBar(width int) string {
	var statusItems []string

	// Add current state and focused panel
	switch m.currentState {
	case stateCommandList:
		switch m.focused {
		case focusCommands:
			statusItems = append(statusItems, "Commands")
		case focusOutput:
			statusItems = append(statusItems, "Output")
		}
		statusItems = append(statusItems, "Tab: Switch Panel")
		statusItems = append(statusItems, "↑↓/Ctrl+jk: Navigate")
		statusItems = append(statusItems, "Enter: Select")
		statusItems = append(statusItems, "/: Search")
		statusItems = append(statusItems, "Ctrl+C: Exit")

	case stateSubcommandSelection:
		switch m.focused {
		case focusCommands:
			statusItems = append(statusItems, "Commands")
		case focusSubcommands:
			statusItems = append(statusItems, "Subcommands")
		case focusOutput:
			statusItems = append(statusItems, "Output")
		}
		statusItems = append(statusItems, "Tab: Switch Panel")
		statusItems = append(statusItems, "↑↓/Ctrl+jk: Navigate")
		statusItems = append(statusItems, "Enter: Select")
		statusItems = append(statusItems, "Esc: Back")
		statusItems = append(statusItems, "Ctrl+C: Exit")

	case stateCommandOptions:
		switch m.focused {
		case focusCommands:
			statusItems = append(statusItems, "Commands (Enter: Execute)")
		case focusOptions:
			statusItems = append(statusItems, "Options (Enter: Configure)")
		case focusOutput:
			statusItems = append(statusItems, "Output")
		case focusInput:
			statusItems = append(statusItems, "Input")
		}
		if m.inputMode {
			statusItems = append(statusItems, "Type text")
			statusItems = append(statusItems, "Enter: Confirm")
			statusItems = append(statusItems, "Esc: Cancel")
		} else {
			statusItems = append(statusItems, "Tab: Switch Panel")
			statusItems = append(statusItems, "↑↓/Ctrl+jk: Navigate")
			statusItems = append(statusItems, "Ctrl+R: Execute")
			statusItems = append(statusItems, "Esc: Back")
		}

	case stateExecuting:
		statusItems = append(statusItems, "⏳ Executing...")

	case stateResults:
		statusItems = append(statusItems, "✅ Results")
		statusItems = append(statusItems, "Tab: New Command")
		statusItems = append(statusItems, "Esc: Back")
		statusItems = append(statusItems, "Ctrl+C: Exit")
	}

	statusText := strings.Join(statusItems, " | ")

	return statusStyle.
		Width(width).
		Render(statusText)
}

// filterCommands filters commands based on search query
func (m tuiModel) filterCommands() []commandItem {
	if m.searchQuery == "" {
		return m.commands
	}

	var filtered []commandItem
	query := strings.ToLower(m.searchQuery)

	for _, cmd := range m.commands {
		if strings.Contains(strings.ToLower(cmd.name), query) ||
			strings.Contains(strings.ToLower(cmd.description), query) {
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}

// executeCommand executes a command and returns a tea.Cmd
func (m tuiModel) executeCommand(cmdName string) tea.Cmd {
	return func() tea.Msg {
		// Create a buffer to capture command output
		var outputBuffer bytes.Buffer

		// Execute the actual command using RunDirectCommand
		handled, err := RunDirectCommand(cmdName, []string{}, &outputBuffer)

		var output string
		if err != nil {
			output = fmt.Sprintf("❌ Error executing command '%s': %v", cmdName, err)
		} else if !handled {
			// Command not found in RunDirectCommand, provide a help message
			output = fmt.Sprintf("Command '%s' not yet implemented.\n\nUse 'help' to see available commands.", cmdName)
		} else {
			// Use the actual command output
			output = outputBuffer.String()
			if output == "" {
				output = fmt.Sprintf("✅ Command '%s' executed successfully (no output)", cmdName)
			}
		}

		return executeCommandMsg{
			command: cmdName,
			output:  output,
		}
	}
}

// executeCommandWithParams executes a command with parameters and returns a tea.Cmd
func (m tuiModel) executeCommandWithParams(cmdName string, args []string) tea.Cmd {
	// Check if this is a command that supports streaming
	if cmdName == "flake" && len(args) > 0 && args[0] == "validate" {
		return tea.Batch(
			func() tea.Msg {
				return commandExecutionStartMsg{
					command: fmt.Sprintf("%s %s", cmdName, strings.Join(args, " ")),
				}
			},
			m.executeFlakeValidateStreaming(args[1:]),
		)
	}

	return func() tea.Msg {
		// Create a buffer to capture command output
		var outputBuffer bytes.Buffer

		// Execute the actual command using RunDirectCommand with parameters
		handled, err := RunDirectCommand(cmdName, args, &outputBuffer)

		var output string
		if err != nil {
			output = fmt.Sprintf("❌ Error executing command '%s %s': %v", cmdName, strings.Join(args, " "), err)
		} else if !handled {
			// Command not found in RunDirectCommand, provide a help message
			output = fmt.Sprintf("Command '%s' with parameters not yet implemented.\n\nUse 'help' to see available commands.", cmdName)
		} else {
			// Use the actual command output
			output = outputBuffer.String()
			if output == "" {
				output = fmt.Sprintf("✅ Command '%s %s' executed successfully (no output)", cmdName, strings.Join(args, " "))
			}
		}

		return executeCommandMsg{
			command: fmt.Sprintf("%s %s", cmdName, strings.Join(args, " ")),
			output:  output,
		}
	}
}

// executeCommandWithSubcommand executes a command with a subcommand and returns a tea.Cmd
func (m tuiModel) executeCommandWithSubcommand(cmdName, subcommandName string, args []string) tea.Cmd {
	return func() tea.Msg {
		// Create a buffer to capture command output
		var outputBuffer bytes.Buffer

		// Build full command with subcommand
		fullArgs := append([]string{subcommandName}, args...)

		// Execute the actual command using RunDirectCommand with subcommand
		handled, err := RunDirectCommand(cmdName, fullArgs, &outputBuffer)

		var output string
		if err != nil {
			output = fmt.Sprintf("❌ Error executing command '%s %s %s': %v", cmdName, subcommandName, strings.Join(args, " "), err)
		} else if !handled {
			// Command not found in RunDirectCommand, provide a help message
			output = fmt.Sprintf("Subcommand '%s %s' not yet implemented.\n\nUse 'help' to see available commands.", cmdName, subcommandName)
		} else {
			// Use the actual command output
			output = outputBuffer.String()
			if output == "" {
				output = fmt.Sprintf("✅ Command '%s %s %s' executed successfully (no output)", cmdName, subcommandName, strings.Join(args, " "))
			}
		}

		return executeCommandMsg{
			command: fmt.Sprintf("%s %s %s", cmdName, subcommandName, strings.Join(args, " ")),
			output:  output,
		}
	}
}

// executeCommandWithStreaming executes a command with real-time streaming output
func (m tuiModel) executeCommandWithStreaming(cmdName string, args []string) tea.Cmd {
	return tea.Sequence(
		// Send start message
		func() tea.Msg {
			return commandExecutionStartMsg{
				command: fmt.Sprintf("%s %s", cmdName, strings.Join(args, " ")),
			}
		},
		// Execute command with streaming
		func() tea.Msg {
			return m.streamCommand(cmdName, args)()
		},
	)
}

// streamCommand creates a streaming command execution
func (m tuiModel) streamCommand(cmdName string, args []string) tea.Cmd {
	return func() tea.Msg {
		command := fmt.Sprintf("%s %s", cmdName, strings.Join(args, " "))

		// Check if this is a flake validate command that should use real-time execution
		if cmdName == "flake" && len(args) > 0 && args[0] == "validate" {
			return m.executeFlakeValidateStreaming(args[1:])()
		}

		// For other commands, fall back to regular execution
		var outputBuffer bytes.Buffer
		handled, err := RunDirectCommand(cmdName, args, &outputBuffer)

		var output string
		if err != nil {
			output = fmt.Sprintf("❌ Error executing command '%s': %v", command, err)
		} else if !handled {
			output = fmt.Sprintf("Command '%s' not yet implemented.\n\nUse 'help' to see available commands.", command)
		} else {
			output = outputBuffer.String()
			if output == "" {
				output = fmt.Sprintf("✅ Command '%s' executed successfully", command)
			}
		}

		return streamingOutputMsg{
			command: command,
			output:  output,
			isEnd:   true,
		}
	}
}

// executeFlakeValidateStreaming executes flake validate with real-time output
func (m tuiModel) executeFlakeValidateStreaming(args []string) tea.Cmd {
	return func() tea.Msg {
		command := "flake validate"

		// Determine the correct flake path using user config or arguments
		var flakePath string
		if len(args) > 0 {
			// Use argument if provided
			flakePath = args[0]
		} else {
			// Load user configuration to get NixOS path
			userCfg, err := config.LoadUserConfig()
			if err == nil && userCfg.NixosFolder != "" {
				configPath := utils.ExpandHome(userCfg.NixosFolder)

				// Check if the path is a directory containing flake.nix or a direct file path
				if utils.IsDirectory(configPath) {
					flakePath = filepath.Join(configPath, "flake.nix")
				} else if strings.HasSuffix(configPath, "flake.nix") {
					flakePath = configPath
				} else {
					flakePath = filepath.Join(configPath, "flake.nix")
				}
			} else {
				// Fallback to auto-detection
				commonPaths := []string{
					os.ExpandEnv("$HOME/.config/nixos/flake.nix"),
					"/etc/nixos/flake.nix",
					"./flake.nix",
				}

				for _, p := range commonPaths {
					if utils.IsFile(p) {
						flakePath = p
						break
					}
				}

				if flakePath == "" {
					flakePath = "./flake.nix"
				}
			}
		}

		// Check if flake.nix exists
		if !utils.IsFile(flakePath) {
			return streamingOutputMsg{
				command: command,
				output:  fmt.Sprintf("❌ No flake.nix found at: %s", flakePath),
				isEnd:   true,
			}
		}

		flakeDir := filepath.Dir(flakePath)

		// Execute the command with live output collection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "nix", "flake", "check")
		cmd.Dir = flakeDir

		// Set up pipes for real-time output capture
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return streamingOutputMsg{
				command: command,
				output:  fmt.Sprintf("❌ Failed to create stdout pipe: %v", err),
				isEnd:   true,
			}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return streamingOutputMsg{
				command: command,
				output:  fmt.Sprintf("❌ Failed to create stderr pipe: %v", err),
				isEnd:   true,
			}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return streamingOutputMsg{
				command: command,
				output:  fmt.Sprintf("❌ Failed to start command: %v", err),
				isEnd:   true,
			}
		}

		// Read combined output
		var outputBuilder strings.Builder
		outputBuilder.WriteString(fmt.Sprintf("🔍 Validating flake: %s\n\n", flakePath))

		// Read stdout
		stdoutScanner := bufio.NewScanner(stdout)
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if strings.TrimSpace(line) != "" {
				outputBuilder.WriteString(line + "\n")
			}
		}

		// Read stderr
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if strings.TrimSpace(line) != "" {
				outputBuilder.WriteString(line + "\n")
			}
		}

		// Wait for command completion
		err = cmd.Wait()

		// Final result
		var finalOutput string
		if err != nil {
			finalOutput = fmt.Sprintf("%s\n❌ Flake validation failed: %v", outputBuilder.String(), err)
		} else {
			finalOutput = fmt.Sprintf("%s\n✅ Flake validation completed successfully", outputBuilder.String())
		}

		return streamingOutputMsg{
			command: command,
			output:  finalOutput,
			isEnd:   true,
		}
	}
}
