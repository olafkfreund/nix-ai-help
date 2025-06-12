package components

import (
	"nix-ai-help/internal/tui/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AskResponseLoadedMsg represents when AI response content is loaded
type AskResponseLoadedMsg struct {
	Content string
}

// AskResponsePopup represents the AI response popup component
type AskResponsePopup struct {
	width    int
	height   int
	theme    *styles.Theme
	viewport viewport.Model
	visible  bool
	content  string
	question string
}

// NewAskResponsePopup creates a new AI response popup
func NewAskResponsePopup(theme *styles.Theme) *AskResponsePopup {
	vp := viewport.New(0, 0)

	return &AskResponsePopup{
		theme:    theme,
		viewport: vp,
		visible:  false,
	}
}

// Init initializes the AI response popup
func (a *AskResponsePopup) Init() tea.Cmd {
	return nil
}

// Update handles messages for the AI response popup
func (a *AskResponsePopup) Update(msg tea.Msg) (*AskResponsePopup, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !a.visible {
			return a, nil
		}

		switch msg.String() {
		case "esc", "enter", "q":
			a.visible = false
			return a, nil
		case "up", "k":
			a.viewport.LineUp(1)
		case "down", "j":
			a.viewport.LineDown(1)
		case "pgup":
			a.viewport.HalfViewUp()
		case "pgdown":
			a.viewport.HalfViewDown()
		case "home", "g":
			a.viewport.GotoTop()
		case "end", "G":
			a.viewport.GotoBottom()
		}

	case AskResponseLoadedMsg:
		a.content = msg.Content
		a.viewport.SetContent(a.content)
		return a, nil
	}

	a.viewport, cmd = a.viewport.Update(msg)
	return a, cmd
}

// View renders the AI response popup
func (a *AskResponsePopup) View() string {
	if !a.visible {
		return ""
	}

	// Create the popup border
	border := lipgloss.RoundedBorder()

	// Header with question
	headerText := "🤖 AI Response"
	if a.question != "" {
		headerText = "🤖 AI Response: " + a.question
		// Truncate long questions
		if len(headerText) > a.width-8 {
			headerText = headerText[:a.width-11] + "..."
		}
	}

	header := a.theme.CommandsPanel.Header.
		Width(a.width - 4).
		Align(lipgloss.Center).
		Render(headerText)

	// Content area
	content := a.viewport.View()

	// Footer with controls
	scrollPercent := 0
	if a.viewport.TotalLineCount() > 0 {
		scrollPercent = int((float64(a.viewport.YOffset) / float64(max(1, a.viewport.TotalLineCount()-a.viewport.Height))) * 100)
		if a.viewport.AtTop() {
			scrollPercent = 0
		}
		if a.viewport.AtBottom() {
			scrollPercent = 100
		}
	}

	footerText := "[Esc] Close  [↑↓/jk] Scroll  [PgUp/PgDn] Page  [Home/End] Top/Bottom"
	if scrollPercent >= 0 {
		if scrollPercent >= 100 {
			footerText += " (100%)"
		} else if scrollPercent >= 10 {
			footerText += " (" + string(rune('0'+scrollPercent/10)) + string(rune('0'+scrollPercent%10)) + "%)"
		} else {
			footerText += " (" + string(rune('0'+scrollPercent)) + "%)"
		}
	}

	footer := a.theme.CommandsPanel.Base.
		Foreground(a.theme.Muted).
		Width(a.width - 4).
		Align(lipgloss.Center).
		Render(footerText)

	// Combine all parts
	popup := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		footer,
	)

	// Apply border and styling
	styledPopup := lipgloss.NewStyle().
		Border(border).
		BorderForeground(a.theme.Primary).
		Background(a.theme.Background).
		Padding(1).
		Width(a.width).
		Height(a.height).
		Render(popup)

	return styledPopup
}

// Show displays the AI response popup with the given question and response
func (a *AskResponsePopup) Show(question, response string) {
	a.question = question
	a.content = response
	a.viewport.SetContent(a.content)
	a.viewport.GotoTop()
	a.visible = true
}

// Hide hides the AI response popup
func (a *AskResponsePopup) Hide() {
	a.visible = false
}

// IsVisible returns whether the popup is visible
func (a *AskResponsePopup) IsVisible() bool {
	return a.visible
}

// HasContent returns whether the popup has content to show
func (a *AskResponsePopup) HasContent() bool {
	return a.content != ""
}

// SetSize sets the size of the popup
func (a *AskResponsePopup) SetSize(width, height int) {
	a.width = width
	a.height = height
	a.viewport.Width = width - 6   // Account for border and padding
	a.viewport.Height = height - 8 // Account for header, footer, border, padding
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
