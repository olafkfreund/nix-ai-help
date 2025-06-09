package panels

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/tui/models"
	"nix-ai-help/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandsPanel represents the left panel showing available commands
type CommandsPanel struct {
	width   int
	height  int
	focused bool
	theme   *styles.Theme

	// Commands
	commands         []models.CommandMetadata
	filteredCommands []models.CommandMetadata
	selected         int

	// Search
	searchMode  bool
	searchInput textinput.Model
	searchQuery string

	// Display
	scrollOffset int
}

// NewCommandsPanel creates a new commands panel
func NewCommandsPanel(theme *styles.Theme) *CommandsPanel {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search commands..."
	searchInput.CharLimit = 50

	commands := models.GetDefaultCommands()

	return &CommandsPanel{
		theme:            theme,
		commands:         commands,
		filteredCommands: commands,
		selected:         0,
		searchInput:      searchInput,
		searchMode:       false,
	}
}

// Init initializes the commands panel
func (p *CommandsPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the commands panel
func (p *CommandsPanel) Update(msg tea.Msg) (*CommandsPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}

		if p.searchMode {
			return p.handleSearchInput(msg)
		}

		switch msg.String() {
		case "up", "k":
			p.moveUp()
		case "down", "j":
			p.moveDown()
		case "home":
			p.selected = 0
			p.scrollOffset = 0
		case "end":
			p.selected = len(p.filteredCommands) - 1
			p.adjustScroll()
		case "pgup":
			p.moveUp10()
		case "pgdown":
			p.moveDown10()
		case "/":
			p.enterSearchMode()
			return p, p.searchInput.Focus()
		case "enter":
			return p, p.executeSelectedCommand()
		case "esc":
			if p.searchQuery != "" {
				p.clearSearch()
			}
		}
	}

	return p, cmd
}

// View renders the commands panel
func (p *CommandsPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	var content strings.Builder

	// Header
	header := "Commands"
	if p.searchMode || p.searchQuery != "" {
		header = "Search Commands"
	}
	content.WriteString(p.theme.CommandsPanel.Header.Render(header))
	content.WriteString("\n")

	// Search input (if in search mode)
	if p.searchMode {
		searchView := p.theme.SearchBox.Width(p.width - 4).Render(p.searchInput.View())
		content.WriteString(searchView)
		content.WriteString("\n")
	} else if p.searchQuery != "" {
		// Show current search query
		queryView := p.theme.SearchBox.Width(p.width - 4).Render("Filter: " + p.searchQuery)
		content.WriteString(queryView)
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Commands list
	availableHeight := p.height - 6 // Account for header, spacing, help, and scroll indicator
	if p.searchMode || p.searchQuery != "" {
		availableHeight -= 2 // Account for search box
	}

	commandsView := p.renderCommandsList(availableHeight)
	content.WriteString(commandsView)

	// Scroll indicator
	if len(p.filteredCommands) > 0 {
		scrollInfo := p.renderScrollIndicator(availableHeight)
		content.WriteString("\n")
		content.WriteString(scrollInfo)
	}

	// Help text
	if !p.searchMode {
		helpText := p.theme.CommandsPanel.Base.
			Foreground(p.theme.Muted).
			Render("/ search • ↑↓ navigate • Enter execute • PgUp/PgDn scroll")
		content.WriteString("\n")
		content.WriteString(helpText)
	}

	return p.theme.CommandsPanel.Base.Render(content.String())
}

// renderCommandsList renders the list of commands
func (p *CommandsPanel) renderCommandsList(height int) string {
	if len(p.filteredCommands) == 0 {
		return p.theme.CommandsPanel.Base.
			Foreground(p.theme.Muted).
			Render("  No commands found")
	}

	var lines []string

	start := p.scrollOffset
	// Adjust for multi-line items - each command takes 2 lines now
	maxItems := height / 2
	end := start + maxItems
	if end > len(p.filteredCommands) {
		end = len(p.filteredCommands)
	}

	for i := start; i < end; i++ {
		cmd := p.filteredCommands[i]
		line := p.renderCommandItem(cmd, i == p.selected)
		lines = append(lines, line)

		// Add spacing between command items
		if i < end-1 {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// renderCommandItem renders a single command item
func (p *CommandsPanel) renderCommandItem(cmd models.CommandMetadata, selected bool) string {
	name := cmd.Name

	// Create more prominent styling for commands
	if selected {
		name = p.theme.CommandsPanel.Selected.Render(fmt.Sprintf("  %s  ", name))
	} else {
		name = p.theme.CommandsPanel.Content.Render(fmt.Sprintf("  %s", name))
	}

	// Add description if there's space
	if p.width > 25 && cmd.Description != "" {
		maxDescLen := p.width - len(cmd.Name) - 10 // Account for extra spacing
		if maxDescLen > 0 {
			desc := cmd.Description
			if len(desc) > maxDescLen {
				desc = desc[:maxDescLen-3] + "..."
			}
			if selected {
				// For selected items, show description below the name for better readability
				descLine := p.theme.CommandsPanel.Base.
					Foreground(p.theme.Muted).
					Render(fmt.Sprintf("    %s", desc))
				return name + "\n" + descLine
			} else {
				desc = p.theme.CommandsPanel.Base.
					Foreground(p.theme.Muted).
					Render(fmt.Sprintf("  %s", desc))
				return name + "\n" + desc
			}
		}
	}

	return name
}

// renderScrollIndicator renders a text-based scroll position indicator
func (p *CommandsPanel) renderScrollIndicator(availableHeight int) string {
	if len(p.filteredCommands) == 0 {
		return ""
	}

	maxVisibleItems := availableHeight / 3 // Account for multi-line items

	// Calculate visible range
	start := p.scrollOffset + 1
	end := p.scrollOffset + maxVisibleItems
	if end > len(p.filteredCommands) {
		end = len(p.filteredCommands)
	}

	// Only show indicator if there are more items than can be displayed
	if len(p.filteredCommands) <= maxVisibleItems {
		return ""
	}

	scrollText := fmt.Sprintf("(%d-%d of %d)", start, end, len(p.filteredCommands))

	return p.theme.CommandsPanel.Base.
		Foreground(p.theme.Muted).
		Render(fmt.Sprintf("  %s", scrollText))
}

// handleSearchInput handles input when in search mode
func (p *CommandsPanel) handleSearchInput(msg tea.KeyMsg) (*CommandsPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		p.exitSearchMode()
		return p, nil
	case "esc":
		p.exitSearchMode()
		return p, nil
	default:
		p.searchInput, cmd = p.searchInput.Update(msg)
		p.searchQuery = p.searchInput.Value()
		p.filterCommands()
	}

	return p, cmd
}

// enterSearchMode enters search mode
func (p *CommandsPanel) enterSearchMode() {
	p.searchMode = true
	p.searchInput.SetValue(p.searchQuery)
}

// exitSearchMode exits search mode
func (p *CommandsPanel) exitSearchMode() {
	p.searchMode = false
	p.searchInput.Blur()
}

// clearSearch clears the search query and shows all commands
func (p *CommandsPanel) clearSearch() {
	p.searchQuery = ""
	p.searchInput.SetValue("")
	p.filteredCommands = p.commands
	p.selected = 0
	p.scrollOffset = 0
}

// filterCommands filters commands based on search query
func (p *CommandsPanel) filterCommands() {
	p.filteredCommands = models.FilterCommands(p.commands, p.searchQuery)
	p.selected = 0
	p.scrollOffset = 0
}

// moveUp moves selection up
func (p *CommandsPanel) moveUp() {
	if p.selected > 0 {
		p.selected--
		p.adjustScroll()
	}
}

// moveDown moves selection down
func (p *CommandsPanel) moveDown() {
	if p.selected < len(p.filteredCommands)-1 {
		p.selected++
		p.adjustScroll()
	}
}

// moveUp10 moves selection up by 10
func (p *CommandsPanel) moveUp10() {
	p.selected -= 10
	if p.selected < 0 {
		p.selected = 0
	}
	p.adjustScroll()
}

// moveDown10 moves selection down by 10
func (p *CommandsPanel) moveDown10() {
	p.selected += 10
	if p.selected >= len(p.filteredCommands) {
		p.selected = len(p.filteredCommands) - 1
	}
	p.adjustScroll()
}

// adjustScroll adjusts scroll offset to keep selected item visible
func (p *CommandsPanel) adjustScroll() {
	visibleHeight := p.height - 6 // Account for header, search, help
	if p.searchMode || p.searchQuery != "" {
		visibleHeight -= 2
	}

	// Account for multi-line items (each command takes ~2-3 lines)
	maxVisibleItems := visibleHeight / 3

	if p.selected < p.scrollOffset {
		p.scrollOffset = p.selected
	} else if p.selected >= p.scrollOffset+maxVisibleItems {
		p.scrollOffset = p.selected - maxVisibleItems + 1
	}
}

// executeSelectedCommand returns a command to execute the selected command
func (p *CommandsPanel) executeSelectedCommand() tea.Cmd {
	if len(p.filteredCommands) == 0 {
		return nil
	}

	selectedCmd := p.filteredCommands[p.selected]

	// Create a command execution message
	return func() tea.Msg {
		return CommandSelectedMsg{
			Command: selectedCmd.Name,
		}
	}
}

// SetFocused sets the focus state of the panel
func (p *CommandsPanel) SetFocused(focused bool) {
	p.focused = focused
	if !focused && p.searchMode {
		p.exitSearchMode()
	}
}

// SetSize sets the size of the panel
func (p *CommandsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// CommandSelectedMsg represents a message when a command is selected
type CommandSelectedMsg struct {
	Command string
}
