package panels

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar represents the bottom status bar
type StatusBar struct {
	width  int
	height int
	theme  *styles.Theme

	// Status information
	mcpServerRunning bool
	aiProvider       string
	nixosPath        string
	currentCommand   string
	systemLoad       string
}

// NewStatusBar creates a new status bar
func NewStatusBar(theme *styles.Theme) *StatusBar {
	return &StatusBar{
		theme:            theme,
		mcpServerRunning: false,
		aiProvider:       "Unknown",
		nixosPath:        "/etc/nixos",
		systemLoad:       "Normal",
	}
}

// Init initializes the status bar
func (s *StatusBar) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status bar
func (s *StatusBar) Update(msg tea.Msg) (*StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusUpdateMsg:
		s.mcpServerRunning = msg.MCPServerRunning
		s.aiProvider = msg.AIProvider
		s.nixosPath = msg.NixOSPath
		s.systemLoad = msg.SystemLoad

	case CommandExecutionStartMsg:
		s.currentCommand = msg.Command

	case CommandExecutionResultMsg:
		s.currentCommand = ""
	}

	return s, nil
}

// View renders the status bar
func (s *StatusBar) View() string {
	if s.width == 0 {
		return ""
	}

	// Left side - Status indicators
	leftStatus := s.renderLeftStatus()

	// Right side - Help text
	rightStatus := s.renderRightStatus()

	// Calculate spacing
	leftWidth := lipgloss.Width(leftStatus)
	rightWidth := lipgloss.Width(rightStatus)
	spacing := s.width - leftWidth - rightWidth

	if spacing < 0 {
		spacing = 0
	}

	spacer := strings.Repeat(" ", spacing)

	statusLine := leftStatus + spacer + rightStatus

	return s.theme.StatusBar.Base.
		Width(s.width).
		Render(statusLine)
}

// renderLeftStatus renders the left side status indicators
func (s *StatusBar) renderLeftStatus() string {
	var indicators []string

	// Current command indicator
	if s.currentCommand != "" {
		cmdIndicator := s.theme.StatusBar.Active.Render(fmt.Sprintf("⏳ %s", s.currentCommand))
		indicators = append(indicators, cmdIndicator)
	}

	// MCP Server status
	mcpStatus := "● MCP"
	if s.mcpServerRunning {
		mcpStatus = s.theme.StatusBar.Success.Render("● MCP Running")
	} else {
		mcpStatus = s.theme.StatusBar.Error.Render("● MCP Stopped")
	}
	indicators = append(indicators, mcpStatus)

	// AI Provider status
	aiStatus := s.theme.StatusBar.Active.Render(fmt.Sprintf("● AI: %s", s.aiProvider))
	indicators = append(indicators, aiStatus)

	// NixOS path
	pathStatus := s.theme.StatusBar.Base.Render(fmt.Sprintf("● NixOS: %s", s.nixosPath))
	indicators = append(indicators, pathStatus)

	return lipgloss.JoinHorizontal(lipgloss.Left, indicators...)
}

// renderRightStatus renders the right side help text
func (s *StatusBar) renderRightStatus() string {
	helpItems := []string{
		"F1:Help",
		"Tab:Switch",
		"Ctrl+C:Exit",
	}

	var styledItems []string
	for _, item := range helpItems {
		styled := s.theme.StatusBar.Base.
			Foreground(s.theme.Muted).
			Render(item)
		styledItems = append(styledItems, styled)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		styledItems[0], " │ ",
		styledItems[1], " │ ",
		styledItems[2],
	)
}

// SetSize sets the size of the status bar
func (s *StatusBar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetMCPStatus updates the MCP server status
func (s *StatusBar) SetMCPStatus(running bool) {
	s.mcpServerRunning = running
}

// SetAIProvider updates the AI provider information
func (s *StatusBar) SetAIProvider(provider string) {
	s.aiProvider = provider
}

// SetNixOSPath updates the NixOS configuration path
func (s *StatusBar) SetNixOSPath(path string) {
	s.nixosPath = path
}

// StatusUpdateMsg represents a status update message
type StatusUpdateMsg struct {
	MCPServerRunning bool
	AIProvider       string
	NixOSPath        string
	SystemLoad       string
}
