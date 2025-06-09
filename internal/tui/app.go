package tui

import (
	"strings"

	"nix-ai-help/internal/tui/components"
	"nix-ai-help/internal/tui/models"
	"nix-ai-help/internal/tui/panels"
	"nix-ai-help/internal/tui/services"
	"nix-ai-help/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FocusedPanel represents which panel is currently focused
type FocusedPanel int

const (
	CommandsPanel FocusedPanel = iota
	ExecutionPanel
)

// App represents the main TUI application
type App struct {
	width  int
	height int

	// Panels
	commandsPanel  *panels.CommandsPanel
	executionPanel *panels.ExecutionPanel
	statusBar      *panels.StatusBar

	// Components
	changelogPopup *components.ChangelogPopup

	// State
	focused  FocusedPanel
	theme    *styles.Theme
	quitting bool

	// Services
	executor *services.CommandExecutor

	// Application state
	state *models.AppState
}

// NewApp creates a new TUI application instance
func NewApp() *App {
	theme := styles.NewDefaultTheme()
	state := models.NewAppState()
	executor := services.NewCommandExecutor()

	return &App{
		commandsPanel:  panels.NewCommandsPanel(theme),
		executionPanel: panels.NewExecutionPanel(theme),
		statusBar:      panels.NewStatusBar(theme),
		changelogPopup: components.NewChangelogPopup(theme),
		focused:        CommandsPanel,
		theme:          theme,
		executor:       executor,
		state:          state,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.commandsPanel.Init(),
		a.executionPanel.Init(),
		a.statusBar.Init(),
		a.changelogPopup.Init(),
	)
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updatePanelSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.quitting = true
			return a, tea.Quit

		case "tab":
			a.switchFocus()

		case "shift+tab":
			a.switchFocusReverse()

		case "f1", "ctrl+h":
			// Toggle changelog popup
			if a.changelogPopup.IsVisible() {
				a.changelogPopup.Hide()
			} else {
				a.changelogPopup.Show()
			}
			return a, nil

		default:
			// If changelog popup is visible, route keys to it first
			if a.changelogPopup.IsVisible() {
				var cmd tea.Cmd
				a.changelogPopup, cmd = a.changelogPopup.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			}

			// Route key messages to the focused panel
			cmd := a.handleFocusedPanelInput(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	default:
		// Handle command execution messages
		switch msg := msg.(type) {
		case panels.CommandSelectedMsg:
			// Convert CommandSelectedMsg to CommandExecutionStartMsg
			startMsg := panels.CommandExecutionStartMsg{
				Command: msg.Command,
			}
			cmd := a.executor.ExecuteCommand(startMsg.Command)
			cmds = append(cmds, cmd)
		case panels.CommandExecutionStartMsg:
			cmd := a.executor.ExecuteCommand(msg.Command)
			cmds = append(cmds, cmd)
		}

		// Update all panels with the message
		var cmd tea.Cmd
		a.commandsPanel, cmd = a.commandsPanel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		a.executionPanel, cmd = a.executionPanel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		a.statusBar, cmd = a.statusBar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		a.changelogPopup, cmd = a.changelogPopup.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if a.quitting {
		return "Goodbye!\n"
	}

	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Calculate panel dimensions
	commandsPanelWidth := int(float64(a.width) * 0.3)
	executionPanelWidth := a.width - commandsPanelWidth - 1 // -1 for border
	contentHeight := a.height - 3                           // Reserve space for status bar

	// Render panels
	commandsView := a.commandsPanel.View()
	executionView := a.executionPanel.View()
	statusView := a.statusBar.View()

	// Apply focus styling
	if a.focused == CommandsPanel {
		commandsView = a.theme.FocusedBorder.Render(commandsView)
		executionView = a.theme.UnfocusedBorder.Render(executionView)
	} else {
		commandsView = a.theme.UnfocusedBorder.Render(commandsView)
		executionView = a.theme.FocusedBorder.Render(executionView)
	}

	// Resize panels to fit
	commandsView = lipgloss.NewStyle().
		Width(commandsPanelWidth).
		Height(contentHeight).
		Render(commandsView)

	executionView = lipgloss.NewStyle().
		Width(executionPanelWidth).
		Height(contentHeight).
		Render(executionView)

	// Combine panels horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		commandsView,
		executionView,
	)

	// Add status bar
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		statusView,
	)

	// If changelog popup is visible, render it on top
	if a.changelogPopup.IsVisible() {
		popupView := a.changelogPopup.View()

		// Calculate position to center the popup
		popupWidth := int(float64(a.width) * 0.8)
		popupHeight := int(float64(a.height) * 0.8)
		xOffset := (a.width - popupWidth) / 2
		yOffset := (a.height - popupHeight) / 2

		// Simple overlay: place popup on top of content
		if xOffset >= 0 && yOffset >= 0 {
			lines := strings.Split(content, "\n")
			popupLines := strings.Split(popupView, "\n")

			// Ensure we have enough lines
			for len(lines) < a.height {
				lines = append(lines, "")
			}

			// Overlay popup lines
			for i, popupLine := range popupLines {
				lineIndex := yOffset + i
				if lineIndex < len(lines) {
					// Create line with popup content centered
					line := lines[lineIndex]
					if len(line) < a.width {
						line += strings.Repeat(" ", a.width-len(line))
					}

					if xOffset+len(popupLine) <= len(line) {
						runes := []rune(line)
						popupRunes := []rune(popupLine)
						copy(runes[xOffset:], popupRunes)
						lines[lineIndex] = string(runes)
					}
				}
			}
			content = strings.Join(lines, "\n")
		} else {
			// Fallback: just show popup centered
			content = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, popupView)
		}
	}

	return content
}

// switchFocus switches focus to the next panel
func (a *App) switchFocus() {
	switch a.focused {
	case CommandsPanel:
		a.focused = ExecutionPanel
		a.commandsPanel.SetFocused(false)
		a.executionPanel.SetFocused(true)
	case ExecutionPanel:
		a.focused = CommandsPanel
		a.executionPanel.SetFocused(false)
		a.commandsPanel.SetFocused(true)
	}
}

// switchFocusReverse switches focus to the previous panel
func (a *App) switchFocusReverse() {
	switch a.focused {
	case ExecutionPanel:
		a.focused = CommandsPanel
		a.executionPanel.SetFocused(false)
		a.commandsPanel.SetFocused(true)
	case CommandsPanel:
		a.focused = ExecutionPanel
		a.commandsPanel.SetFocused(false)
		a.executionPanel.SetFocused(true)
	}
}

// handleFocusedPanelInput routes input to the currently focused panel
func (a *App) handleFocusedPanelInput(msg tea.KeyMsg) tea.Cmd {
	switch a.focused {
	case CommandsPanel:
		var cmd tea.Cmd
		a.commandsPanel, cmd = a.commandsPanel.Update(msg)
		return cmd
	case ExecutionPanel:
		var cmd tea.Cmd
		a.executionPanel, cmd = a.executionPanel.Update(msg)
		return cmd
	}
	return nil
}

// updatePanelSizes updates the sizes of all panels based on current window size
func (a *App) updatePanelSizes() {
	commandsPanelWidth := int(float64(a.width) * 0.3)
	executionPanelWidth := a.width - commandsPanelWidth - 1
	contentHeight := a.height - 3

	a.commandsPanel.SetSize(commandsPanelWidth, contentHeight)
	a.executionPanel.SetSize(executionPanelWidth, contentHeight)
	a.statusBar.SetSize(a.width, 2)

	// Size changelog popup to take up most of the screen
	popupWidth := int(float64(a.width) * 0.8)
	popupHeight := int(float64(a.height) * 0.8)
	a.changelogPopup.SetSize(popupWidth, popupHeight)
}

// Run starts the TUI application
func Run() error {
	app := NewApp()

	program := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := program.Run()
	return err
}
