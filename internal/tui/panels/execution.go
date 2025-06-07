package panels

import (
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/tui/models"
	"nix-ai-help/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ExecutionPanel represents the right panel for command execution
type ExecutionPanel struct {
	width   int
	height  int
	focused bool
	theme   *styles.Theme

	// Input
	input        textinput.Model
	inputHistory []string
	historyIndex int

	// Output
	output   []string
	viewport viewport.Model

	// State
	executing   bool
	lastCommand string
	lastResult  *models.ExecutionResult
}

// NewExecutionPanel creates a new execution panel
func NewExecutionPanel(theme *styles.Theme) *ExecutionPanel {
	input := textinput.New()
	input.Placeholder = "Enter command..."
	input.CharLimit = 200

	vp := viewport.New(0, 0)

	return &ExecutionPanel{
		theme:        theme,
		input:        input,
		viewport:     vp,
		inputHistory: make([]string, 0),
		output:       make([]string, 0),
		historyIndex: -1,
	}
}

// Init initializes the execution panel
func (p *ExecutionPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the execution panel
func (p *ExecutionPanel) Update(msg tea.Msg) (*ExecutionPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}

		switch msg.String() {
		case "enter":
			return p.handleEnter()
		case "up":
			if p.input.Focused() {
				p.navigateHistoryUp()
				return p, nil
			}
			// Scroll output up
			p.viewport.LineUp(1)
		case "down":
			if p.input.Focused() {
				p.navigateHistoryDown()
				return p, nil
			}
			// Scroll output down
			p.viewport.LineDown(1)
		case "pgup":
			p.viewport.HalfViewUp()
		case "pgdown":
			p.viewport.HalfViewDown()
		case "ctrl+l":
			p.clearOutput()
		case "ctrl+k":
			p.clearInput()
		case "ctrl+r":
			// TODO: Implement history search
		case "tab":
			// TODO: Implement auto-completion
		default:
			p.input, cmd = p.input.Update(msg)
		}

	case CommandSelectedMsg:
		// Handle command selected from commands panel
		p.input.SetValue(msg.Command)
		return p.handleEnter()

	case CommandExecutionStartMsg:
		p.executing = true
		p.addOutput(fmt.Sprintf("$ %s", msg.Command))

	case CommandExecutionResultMsg:
		p.executing = false
		p.lastResult = &msg.Result
		if msg.Result.Output != "" {
			p.addOutput(msg.Result.Output)
		}
		if msg.Result.Error != "" {
			errorText := p.theme.ExecutionPanel.Base.
				Foreground(p.theme.Error).
				Render(fmt.Sprintf("Error: %s", msg.Result.Error))
			p.addOutput(errorText)
		}

		// Add completion message
		duration := msg.Result.Duration.Round(time.Millisecond)
		status := "completed"
		if msg.Result.ExitCode != 0 {
			status = "failed"
		}
		completionMsg := p.theme.ExecutionPanel.Base.
			Foreground(p.theme.Muted).
			Render(fmt.Sprintf("Command %s in %v", status, duration))
		p.addOutput(completionMsg)
		p.addOutput("") // Empty line for separation

	case CommandExecutionOutputMsg:
		// Handle streaming output
		if msg.Output != "" {
			p.addOutput(msg.Output)
		}
	}

	// Update viewport
	p.viewport, cmd = p.viewport.Update(msg)

	return p, cmd
}

// View renders the execution panel
func (p *ExecutionPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	var content strings.Builder

	// Header
	header := "Command Execution"
	if p.executing {
		header = "⏳ Executing..."
	} else if p.lastResult != nil {
		if p.lastResult.ExitCode == 0 {
			header = "✅ Ready"
		} else {
			header = "❌ Error"
		}
	}
	content.WriteString(p.theme.ExecutionPanel.Header.Render(header))
	content.WriteString("\n")

	// Output area
	outputHeight := p.height - 6 // Account for header, input, buttons
	p.viewport.Width = p.width - 2
	p.viewport.Height = outputHeight

	outputContent := strings.Join(p.output, "\n")
	p.viewport.SetContent(outputContent)
	content.WriteString(p.viewport.View())
	content.WriteString("\n")

	// Input area
	inputLabel := "Command Input:"
	content.WriteString(p.theme.ExecutionPanel.Base.Render(inputLabel))
	content.WriteString("\n")

	inputView := p.theme.Input.Width(p.width - 4).Render(p.input.View())
	content.WriteString(inputView)
	content.WriteString("\n")

	// Buttons
	buttons := p.renderButtons()
	content.WriteString(buttons)

	return p.theme.ExecutionPanel.Base.Render(content.String())
}

// renderButtons renders the action buttons
func (p *ExecutionPanel) renderButtons() string {
	executeBtn := p.theme.Button.Render("Execute")
	clearBtn := p.theme.Button.
		Background(p.theme.Secondary).
		Render("Clear")
	historyBtn := p.theme.Button.
		Background(p.theme.Secondary).
		Render("History ▼")
	helpBtn := p.theme.Button.
		Background(p.theme.Secondary).
		Render("Help")

	if p.executing {
		executeBtn = p.theme.Button.
			Background(p.theme.Muted).
			Render("Executing...")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		executeBtn, clearBtn, historyBtn, helpBtn,
	)
}

// handleEnter handles the enter key press
func (p *ExecutionPanel) handleEnter() (*ExecutionPanel, tea.Cmd) {
	command := strings.TrimSpace(p.input.Value())
	if command == "" {
		return p, nil
	}

	// Add to history
	p.addToHistory(command)

	// Clear input
	p.input.SetValue("")
	p.historyIndex = -1

	// Execute command
	return p, p.executeCommand(command)
}

// executeCommand creates a command to execute the given command
func (p *ExecutionPanel) executeCommand(command string) tea.Cmd {
	return func() tea.Msg {
		return CommandExecutionStartMsg{
			Command: command,
		}
	}
}

// addToHistory adds a command to the input history
func (p *ExecutionPanel) addToHistory(command string) {
	// Don't add duplicate of last command
	if len(p.inputHistory) > 0 && p.inputHistory[len(p.inputHistory)-1] == command {
		return
	}

	p.inputHistory = append(p.inputHistory, command)

	// Limit history size
	maxHistory := 100
	if len(p.inputHistory) > maxHistory {
		p.inputHistory = p.inputHistory[len(p.inputHistory)-maxHistory:]
	}
}

// navigateHistoryUp navigates up in command history
func (p *ExecutionPanel) navigateHistoryUp() {
	if len(p.inputHistory) == 0 {
		return
	}

	if p.historyIndex == -1 {
		p.historyIndex = len(p.inputHistory) - 1
	} else if p.historyIndex > 0 {
		p.historyIndex--
	}

	if p.historyIndex >= 0 && p.historyIndex < len(p.inputHistory) {
		p.input.SetValue(p.inputHistory[p.historyIndex])
	}
}

// navigateHistoryDown navigates down in command history
func (p *ExecutionPanel) navigateHistoryDown() {
	if len(p.inputHistory) == 0 || p.historyIndex == -1 {
		return
	}

	p.historyIndex++
	if p.historyIndex >= len(p.inputHistory) {
		p.historyIndex = -1
		p.input.SetValue("")
	} else {
		p.input.SetValue(p.inputHistory[p.historyIndex])
	}
}

// addOutput adds a line to the output
func (p *ExecutionPanel) addOutput(line string) {
	p.output = append(p.output, line)

	// Limit output size
	maxLines := 1000
	if len(p.output) > maxLines {
		p.output = p.output[len(p.output)-maxLines:]
	}

	// Auto-scroll to bottom
	p.viewport.GotoBottom()
}

// clearOutput clears the output area
func (p *ExecutionPanel) clearOutput() {
	p.output = make([]string, 0)
	p.viewport.SetContent("")
}

// clearInput clears the input field
func (p *ExecutionPanel) clearInput() {
	p.input.SetValue("")
	p.historyIndex = -1
}

// SetFocused sets the focus state of the panel
func (p *ExecutionPanel) SetFocused(focused bool) {
	p.focused = focused
	if focused {
		p.input.Focus()
	} else {
		p.input.Blur()
	}
}

// SetSize sets the size of the panel
func (p *ExecutionPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2
	p.viewport.Height = height - 6
}

// Message types for command execution
type CommandExecutionStartMsg struct {
	Command string
}

type CommandExecutionResultMsg struct {
	Result models.ExecutionResult
}

type CommandExecutionOutputMsg struct {
	Output string
}
