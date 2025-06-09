package components

import (
	"fmt"
	"io/ioutil"
	"strings"

	"nix-ai-help/internal/tui/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// ChangelogEntry represents a single changelog entry
type ChangelogEntry struct {
	Version      string        `yaml:"version"`
	Date         string        `yaml:"date"`
	Highlights   []string      `yaml:"highlights"`
	Features     []FeatureItem `yaml:"features"`
	Improvements []string      `yaml:"improvements"`
	Fixes        []string      `yaml:"fixes"`
}

// FeatureItem represents a feature with title and description
type FeatureItem struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

// ChangelogData represents the entire changelog structure
type ChangelogData struct {
	Changelog []ChangelogEntry `yaml:"changelog"`
}

// ChangelogPopup represents the changelog popup component
type ChangelogPopup struct {
	width    int
	height   int
	theme    *styles.Theme
	viewport viewport.Model
	visible  bool
	content  string
}

// NewChangelogPopup creates a new changelog popup
func NewChangelogPopup(theme *styles.Theme) *ChangelogPopup {
	vp := viewport.New(0, 0)

	return &ChangelogPopup{
		theme:    theme,
		viewport: vp,
		visible:  false,
	}
}

// Init initializes the changelog popup
func (c *ChangelogPopup) Init() tea.Cmd {
	return c.loadChangelog()
}

// Update handles messages for the changelog popup
func (c *ChangelogPopup) Update(msg tea.Msg) (*ChangelogPopup, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !c.visible {
			return c, nil
		}

		switch msg.String() {
		case "esc", "enter", "q":
			c.visible = false
			return c, nil
		case "up", "k":
			c.viewport.LineUp(1)
		case "down", "j":
			c.viewport.LineDown(1)
		case "pgup":
			c.viewport.HalfViewUp()
		case "pgdown":
			c.viewport.HalfViewDown()
		case "home":
			c.viewport.GotoTop()
		case "end":
			c.viewport.GotoBottom()
		}

	case ChangelogLoadedMsg:
		c.content = msg.Content
		c.viewport.SetContent(c.content)
		return c, nil
	}

	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

// View renders the changelog popup
func (c *ChangelogPopup) View() string {
	if !c.visible {
		return ""
	}

	// Create the popup border
	border := lipgloss.RoundedBorder()

	// Header
	header := c.theme.CommandsPanel.Header.
		Width(c.width - 4).
		Align(lipgloss.Center).
		Render("nixai Changelog")

	// Content area
	content := c.viewport.View()

	// Footer with controls
	footer := c.theme.CommandsPanel.Base.
		Foreground(c.theme.Muted).
		Width(c.width - 4).
		Align(lipgloss.Center).
		Render("[Esc] Close  [↑↓] Scroll  [PgUp/PgDn] Page  [Home/End] Top/Bottom")

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
		BorderForeground(c.theme.Primary).
		Background(c.theme.Background).
		Padding(1).
		Width(c.width).
		Height(c.height).
		Render(popup)

	return styledPopup
}

// Show displays the changelog popup
func (c *ChangelogPopup) Show() {
	c.visible = true
}

// Hide hides the changelog popup
func (c *ChangelogPopup) Hide() {
	c.visible = false
}

// IsVisible returns whether the popup is visible
func (c *ChangelogPopup) IsVisible() bool {
	return c.visible
}

// SetSize sets the size of the popup
func (c *ChangelogPopup) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width - 6   // Account for border and padding
	c.viewport.Height = height - 8 // Account for header, footer, border, padding
}

// loadChangelog loads the changelog from the YAML file
func (c *ChangelogPopup) loadChangelog() tea.Cmd {
	return func() tea.Msg {
		// Try to load the changelog file
		data, err := ioutil.ReadFile("configs/changelog.yaml")
		if err != nil {
			return ChangelogLoadedMsg{
				Content: c.generateErrorMessage(err),
			}
		}

		var changelog ChangelogData
		err = yaml.Unmarshal(data, &changelog)
		if err != nil {
			return ChangelogLoadedMsg{
				Content: c.generateErrorMessage(err),
			}
		}

		content := c.formatChangelog(changelog)
		return ChangelogLoadedMsg{
			Content: content,
		}
	}
}

// formatChangelog formats the changelog data into readable text
func (c *ChangelogPopup) formatChangelog(changelog ChangelogData) string {
	var content strings.Builder

	for i, entry := range changelog.Changelog {
		// Version header
		content.WriteString(fmt.Sprintf("Version %s (%s)\n", entry.Version, entry.Date))
		content.WriteString(strings.Repeat("─", 50) + "\n\n")

		// Highlights
		if len(entry.Highlights) > 0 {
			content.WriteString("HIGHLIGHTS:\n")
			for _, highlight := range entry.Highlights {
				content.WriteString(fmt.Sprintf("  * %s\n", highlight))
			}
			content.WriteString("\n")
		}

		// Features
		if len(entry.Features) > 0 {
			content.WriteString("NEW FEATURES:\n")
			for _, feature := range entry.Features {
				content.WriteString(fmt.Sprintf("  * %s\n", feature.Title))
				content.WriteString(fmt.Sprintf("    %s\n", feature.Description))
			}
			content.WriteString("\n")
		}

		// Improvements
		if len(entry.Improvements) > 0 {
			content.WriteString("IMPROVEMENTS:\n")
			for _, improvement := range entry.Improvements {
				content.WriteString(fmt.Sprintf("  * %s\n", improvement))
			}
			content.WriteString("\n")
		}

		// Fixes
		if len(entry.Fixes) > 0 {
			content.WriteString("BUG FIXES:\n")
			for _, fix := range entry.Fixes {
				content.WriteString(fmt.Sprintf("  * %s\n", fix))
			}
			content.WriteString("\n")
		}

		// Add separator between versions
		if i < len(changelog.Changelog)-1 {
			content.WriteString(strings.Repeat("═", 60) + "\n\n")
		}
	}

	return content.String()
}

// generateErrorMessage generates an error message for display
func (c *ChangelogPopup) generateErrorMessage(err error) string {
	return fmt.Sprintf(`Error loading changelog: %s

This might be because:
  • The changelog.yaml file is missing
  • The file format is invalid
  • File permissions prevent access

You can still use nixai normally. The changelog will be
available when the configuration is properly set up.`, err.Error())
}

// ChangelogLoadedMsg represents a message when changelog is loaded
type ChangelogLoadedMsg struct {
	Content string
}
