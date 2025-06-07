package cli

import (
"bytes"
"fmt"
"os"
"strings"

tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
)

// InteractiveModeTUI starts the modern TUI interface for nixai
func InteractiveModeTUI() {
	// Create the TUI application
	app := tea.NewProgram(
initialModel(),
		tea.WithAltScreen(),
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
	commands         []commandItem
	selectedCommand  int
	commandOutput    string
	isExecuting      bool
	focused          focusedPanel
	terminalWidth    int
	terminalHeight   int
	searchQuery      string
	searchMode       bool
}

type commandItem struct {
	name        string
	description string
	icon        string
}

type focusedPanel int

const (
focusCommands focusedPanel = iota
focusOutput
)

// executeCommandMsg represents a command execution result
type executeCommandMsg struct {
	command string
	output  string
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
	}
}

// getAvailableCommands returns a list of available nixai commands
func getAvailableCommands() []commandItem {
	commands := []commandItem{
		{"ask", "Ask any NixOS question", "🤖"},
		{"build", "Enhanced build troubleshooting and optimization", "🛠️"},
		{"community", "Community resources and support", "🌐"},
		{"config", "Manage nixai configuration", "⚙️"},
		{"configure", "Configure NixOS interactively", "🧑‍💻"},
		{"deps", "Analyze NixOS configuration dependencies", "🔗"},
		{"devenv", "Create and manage development environments", "🧪"},
		{"diagnose", "Diagnose NixOS issues", "🩺"},
		{"doctor", "Run NixOS health checks", "🩻"},
		{"explain-option", "Explain a NixOS option", "🖥️"},
		{"flake", "Nix flake utilities", "🧊"},
		{"gc", "AI-powered garbage collection analysis", "🧹"},
		{"hardware", "AI-powered hardware configuration optimizer", "💻"},
		{"learn", "NixOS learning and training commands", "📚"},
		{"logs", "Analyze and parse NixOS logs", "📝"},
		{"machines", "Manage configurations across multiple machines", "🖧"},
		{"mcp-server", "Start or manage the MCP server", "🛰️"},
		{"migrate", "AI-powered migration assistant", "🔀"},
		{"neovim-setup", "Neovim integration setup", "📝"},
		{"package-repo", "Analyze Git repos and generate Nix derivations", "📦"},
		{"search", "Search for NixOS packages/services", "🔍"},
		{"snippets", "Manage NixOS configuration snippets", "🔖"},
		{"store", "Manage, backup, and analyze the Nix store", "��"},
		{"templates", "Manage NixOS configuration templates", "📄"},
	}
	
	return commands
}

// Init is called when the model is first created
func (m tuiModel) Init() tea.Cmd {
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Switch focus between panels
			switch m.focused {
			case focusCommands:
				m.focused = focusOutput
			case focusOutput:
				m.focused = focusCommands
			}

		case "/":
			// Enter search mode
			if m.focused == focusCommands {
				m.searchMode = true
				m.searchQuery = ""
			}

		case "esc":
			// Exit search mode
			m.searchMode = false
			m.searchQuery = ""

		case "enter":
			if m.focused == focusCommands && !m.searchMode {
				// Execute selected command
				filteredCommands := m.filterCommands()
				if m.selectedCommand >= 0 && m.selectedCommand < len(filteredCommands) {
					cmd := filteredCommands[m.selectedCommand]
					m.isExecuting = true
					return m, m.executeCommand(cmd.name)
				}
			}

		case "up", "k":
			if m.focused == focusCommands && !m.searchMode {
				if m.selectedCommand > 0 {
					m.selectedCommand--
				}
			}

		case "down", "j":
			if m.focused == focusCommands && !m.searchMode {
				filteredCommands := m.filterCommands()
				if m.selectedCommand < len(filteredCommands)-1 {
					m.selectedCommand++
				}
			}

		case "backspace":
			if m.searchMode && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.selectedCommand = 0 // Reset selection when search changes
			}

		default:
			// Handle search input
			if m.searchMode && len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.selectedCommand = 0 // Reset selection when search changes
			}
		}
	}

	return m, nil
}

// View renders the TUI interface
func (m tuiModel) View() string {
	if m.terminalWidth == 0 {
		return "Loading..."
	}

	// Calculate panel dimensions
	leftPanelWidth := m.terminalWidth * 30 / 100
	rightPanelWidth := m.terminalWidth - leftPanelWidth - 4 // Account for borders
	panelHeight := m.terminalHeight - 4 // Account for title and status

	// Create the left panel (commands)
	leftPanel := m.renderCommandsPanel(leftPanelWidth, panelHeight)
	
	// Create the right panel (output)
	rightPanel := m.renderOutputPanel(rightPanelWidth, panelHeight)

	// Create the status bar
	statusBar := m.renderStatusBar(m.terminalWidth)

	// Combine panels horizontally
	mainArea := lipgloss.JoinHorizontal(
lipgloss.Top,
leftPanel,
rightPanel,
)

	// Create the title
	title := titleStyle.Render("❄️ nixai: NixOS AI Assistant - TUI Mode")

	// Combine everything vertically
	return lipgloss.JoinVertical(
lipgloss.Left,
title,
mainArea,
statusBar,
)
}

// renderCommandsPanel renders the left commands panel
func (m tuiModel) renderCommandsPanel(width, height int) string {
	var content strings.Builder

	// Add search bar if in search mode
	if m.searchMode {
		searchBar := fmt.Sprintf("🔍 Search: %s_", m.searchQuery)
		content.WriteString(searchBar + "\n\n")
	} else {
		content.WriteString("Commands (Press / to search):\n\n")
	}

	// Filter commands based on search query
	filteredCommands := m.filterCommands()

	// Render command list
	for i, cmd := range filteredCommands {
		line := fmt.Sprintf("%s %s", cmd.icon, cmd.name)
		
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
	
	if m.isExecuting {
		content = "⏳ Executing command...\n\n" + content
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// renderStatusBar renders the bottom status bar
func (m tuiModel) renderStatusBar(width int) string {
	var statusItems []string

	// Add focused panel indicator
	switch m.focused {
	case focusCommands:
		statusItems = append(statusItems, "📋 Commands")
	case focusOutput:
		statusItems = append(statusItems, "💻 Output")
	}

	// Add key bindings
	statusItems = append(statusItems, "Tab: Switch Panel")
	statusItems = append(statusItems, "Enter: Execute")
	statusItems = append(statusItems, "/: Search")
	statusItems = append(statusItems, "Ctrl+C: Exit")

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
