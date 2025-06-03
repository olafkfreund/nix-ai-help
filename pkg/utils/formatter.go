package utils

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Color and style definitions using lipgloss for consistent, beautiful formatting
var (
	// Color palette
	primaryColor = lipgloss.Color("#7C3AED") // Purple
	successColor = lipgloss.Color("#10B981") // Green
	warningColor = lipgloss.Color("#F59E0B") // Orange
	errorColor   = lipgloss.Color("#EF4444") // Red
	infoColor    = lipgloss.Color("#3B82F6") // Blue
	mutedColor   = lipgloss.Color("#6B7280") // Gray
	accentColor  = lipgloss.Color("#EC4899") // Pink

	// Base styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 0)

	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(infoColor)

	MutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	AccentStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	CodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A3A3A3")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1).
			Margin(0, 1)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			Margin(1, 0)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Margin(1, 0)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Background(lipgloss.Color("#2a003f")).
			Bold(true).
			Padding(0, 1)
)

// FormatHeader creates a prominent header with decorative borders
func FormatHeader(title string) string {
	border := strings.Repeat("━", len(title)+4)
	return fmt.Sprintf("%s\n  %s  \n%s",
		HeaderStyle.Render(border),
		HeaderStyle.Render(title),
		HeaderStyle.Render(border))
}

// FormatSection creates a section with a title and content
func FormatSection(title, content string) string {
	return fmt.Sprintf("%s\n%s\n", TitleStyle.Render("## "+title), content)
}

// FormatSubsection creates a subsection with a subtitle and content
func FormatSubsection(subtitle, content string) string {
	return fmt.Sprintf("%s\n%s\n", SubtitleStyle.Render("### "+subtitle), content)
}

// FormatSubheader creates a bold subheader (used for section headers)
func FormatSubheader(msg string) string {
	return SubtitleStyle.Render(msg + ":")
}

// FormatSuccess creates a success message with checkmark
func FormatSuccess(message string) string {
	return SuccessStyle.Render("✅ " + message)
}

// FormatWarning creates a warning message with warning icon
func FormatWarning(message string) string {
	return WarningStyle.Render("⚠️  " + message)
}

// FormatError creates an error message with error icon
func FormatError(message string) string {
	return ErrorStyle.Render("❌ " + message)
}

// FormatInfo creates an info message with info icon
func FormatInfo(message string) string {
	return InfoStyle.Render("ℹ️  " + message)
}

// FormatProgress creates a progress indicator
func FormatProgress(message string) string {
	return InfoStyle.Render("🔄 " + message)
}

// FormatCode creates inline code formatting
func FormatCode(code string) string {
	return CodeStyle.Render(code)
}

// FormatCodeBlock creates a code block with optional language label
func FormatCodeBlock(code, language string) string {
	var header string
	if language != "" {
		header = MutedStyle.Render(fmt.Sprintf("┌─ %s", language)) + "\n"
	}

	lines := strings.Split(strings.TrimSpace(code), "\n")
	var formattedLines []string

	for _, line := range lines {
		formattedLines = append(formattedLines, CodeStyle.Render(line))
	}

	footer := MutedStyle.Render("└" + strings.Repeat("─", 40))

	return header + strings.Join(formattedLines, "\n") + "\n" + footer
}

// FormatList creates a bulleted list
func FormatList(items []string) string {
	var formatted []string
	for _, item := range items {
		formatted = append(formatted, InfoStyle.Render("  • "+item))
	}
	return strings.Join(formatted, "\n")
}

// FormatNumberedList creates a numbered list
func FormatNumberedList(items []string) string {
	var formatted []string
	for i, item := range items {
		formatted = append(formatted, InfoStyle.Render(fmt.Sprintf("  %d. %s", i+1, item)))
	}
	return strings.Join(formatted, "\n")
}

// FormatKeyValue creates a key-value pair display
func FormatKeyValue(key, value string) string {
	return fmt.Sprintf("%s %s",
		AccentStyle.Render(key+":"),
		InfoStyle.Render(value))
}

// FormatBox creates a boxed content area
func FormatBox(title, content string) string {
	if title != "" {
		titleLine := AccentStyle.Render("┌─ " + title + " ")
		titleLine += MutedStyle.Render(strings.Repeat("─", max(0, 60-len(title)-3)) + "┐")

		lines := strings.Split(content, "\n")
		var boxedLines []string
		boxedLines = append(boxedLines, titleLine)

		for _, line := range lines {
			boxedLines = append(boxedLines, MutedStyle.Render("│ ")+line)
		}

		boxedLines = append(boxedLines, MutedStyle.Render("└"+strings.Repeat("─", 60)+"┘"))
		return strings.Join(boxedLines, "\n")
	}

	return BoxStyle.Render(content)
}

// FormatTable creates a simple table
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}
	// Minimal stub to fix build
	return ""
}

// FormatDivider creates a visual divider line
func FormatDivider() string {
	return strings.Repeat("─", 60)
}

// FormatTip creates a tip message with a lightbulb icon
func FormatTip(message string) string {
	return InfoStyle.Render("💡 " + message)
}

// FormatNote creates a note message with a note icon
func FormatNote(message string) string {
	return MutedStyle.Render("📝 " + message)
}

// RenderMarkdown renders markdown using glamour with fallback to plain text
func RenderMarkdown(markdown string) string {
	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(120))
	if err != nil {
		return markdown
	}
	rendered, err := renderer.Render(markdown)
	if err != nil {
		return markdown
	}
	return rendered
}
