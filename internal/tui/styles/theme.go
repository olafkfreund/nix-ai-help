package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the TUI
type Theme struct {
	// Colors
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Background lipgloss.Color
	Text       lipgloss.Color
	Muted      lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color

	// Borders
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style

	// Panels
	CommandsPanel  PanelStyles
	ExecutionPanel PanelStyles
	StatusBar      StatusBarStyles

	// Components
	SearchBox lipgloss.Style
	Button    lipgloss.Style
	Input     lipgloss.Style
}

// PanelStyles represents styling for panels
type PanelStyles struct {
	Base      lipgloss.Style
	Header    lipgloss.Style
	Content   lipgloss.Style
	Footer    lipgloss.Style
	Selected  lipgloss.Style
	Highlight lipgloss.Style
}

// StatusBarStyles represents styling for the status bar
type StatusBarStyles struct {
	Base     lipgloss.Style
	Active   lipgloss.Style
	Inactive lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
}

// NewDefaultTheme creates the default nixai theme
func NewDefaultTheme() *Theme {
	primary := lipgloss.Color("#5c7cfa")
	secondary := lipgloss.Color("#495057")
	accent := lipgloss.Color("#51cf66")
	background := lipgloss.Color("#1a1b26")
	text := lipgloss.Color("#a9b1d6")
	muted := lipgloss.Color("#565f89")
	errorColor := lipgloss.Color("#f7768e")
	successColor := lipgloss.Color("#9ece6a")
	warningColor := lipgloss.Color("#e0af68")

	theme := &Theme{
		Primary:    primary,
		Secondary:  secondary,
		Accent:     accent,
		Background: background,
		Text:       text,
		Muted:      muted,
		Error:      errorColor,
		Success:    successColor,
		Warning:    warningColor,
	}

	// Borders
	theme.FocusedBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		Padding(0, 1)

	theme.UnfocusedBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(muted).
		Padding(0, 1)

	// Commands Panel
	theme.CommandsPanel = PanelStyles{
		Base: lipgloss.NewStyle().
			Background(background).
			Foreground(text).
			Padding(1),
		Header: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Background(secondary).
			Padding(0, 2).
			MarginBottom(1),
		Content: lipgloss.NewStyle().
			Foreground(text).
			Padding(0, 2).
			MarginLeft(1),
		Selected: lipgloss.NewStyle().
			Background(primary).
			Foreground(background).
			Bold(true).
			Padding(0, 2).
			MarginLeft(1).
			MarginRight(1),
		Highlight: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 1),
	}

	// Execution Panel
	theme.ExecutionPanel = PanelStyles{
		Base: lipgloss.NewStyle().
			Background(background).
			Foreground(text),
		Header: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 1),
		Content: lipgloss.NewStyle().
			Foreground(text).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 1),
	}

	// Status Bar
	theme.StatusBar = StatusBarStyles{
		Base: lipgloss.NewStyle().
			Background(secondary).
			Foreground(text).
			Padding(0, 1),
		Active: lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true),
		Inactive: lipgloss.NewStyle().
			Foreground(muted),
		Error: lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true),
	}

	// Components
	theme.SearchBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(muted).
		Padding(0, 1).
		Foreground(text)

	theme.Button = lipgloss.NewStyle().
		Background(primary).
		Foreground(background).
		Padding(0, 2).
		Margin(0, 1).
		Bold(true)

	theme.Input = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		Padding(0, 1).
		Foreground(text)

	return theme
}

// NewNixOSTheme creates a NixOS-inspired theme
func NewNixOSTheme() *Theme {
	primary := lipgloss.Color("#7ebae4")
	secondary := lipgloss.Color("#414868")
	accent := lipgloss.Color("#9ece6a")
	background := lipgloss.Color("#24283b")
	text := lipgloss.Color("#c0caf5")
	muted := lipgloss.Color("#565f89")
	errorColor := lipgloss.Color("#f7768e")
	successColor := lipgloss.Color("#9ece6a")
	warningColor := lipgloss.Color("#e0af68")

	theme := NewDefaultTheme() // Start with default structure

	// Override colors
	theme.Primary = primary
	theme.Secondary = secondary
	theme.Accent = accent
	theme.Background = background
	theme.Text = text
	theme.Muted = muted
	theme.Error = errorColor
	theme.Success = successColor
	theme.Warning = warningColor

	// Update border colors
	theme.FocusedBorder = theme.FocusedBorder.BorderForeground(primary)
	theme.UnfocusedBorder = theme.UnfocusedBorder.BorderForeground(muted)

	return theme
}

// NewLightTheme creates a light theme
func NewLightTheme() *Theme {
	primary := lipgloss.Color("#364fc7")
	secondary := lipgloss.Color("#e9ecef")
	accent := lipgloss.Color("#2b8a3e")
	background := lipgloss.Color("#ffffff")
	text := lipgloss.Color("#212529")
	muted := lipgloss.Color("#6c757d")
	errorColor := lipgloss.Color("#dc3545")
	successColor := lipgloss.Color("#28a745")
	warningColor := lipgloss.Color("#ffc107")

	theme := NewDefaultTheme() // Start with default structure

	// Override colors
	theme.Primary = primary
	theme.Secondary = secondary
	theme.Accent = accent
	theme.Background = background
	theme.Text = text
	theme.Muted = muted
	theme.Error = errorColor
	theme.Success = successColor
	theme.Warning = warningColor

	// Update all styles with new colors
	theme.FocusedBorder = theme.FocusedBorder.BorderForeground(primary)
	theme.UnfocusedBorder = theme.UnfocusedBorder.BorderForeground(muted)

	// Update panel styles
	theme.CommandsPanel.Base = theme.CommandsPanel.Base.Background(background).Foreground(text)
	theme.ExecutionPanel.Base = theme.ExecutionPanel.Base.Background(background).Foreground(text)

	return theme
}
