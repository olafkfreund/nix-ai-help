package models

import (
	"time"
)

// AppState represents the global application state
type AppState struct {
	// Configuration
	PanelRatio float64
	MaxHistory int

	// Current state
	CurrentCommand string
	LastExecution  *ExecutionResult
	History        []HistoryEntry

	// System status
	MCPServerRunning bool
	AIProvider       string
	NixOSPath        string
}

// Command represents a nixai command with metadata
type Command struct {
	Name        string
	Description string
	Usage       string
	Examples    []string
	Subcommands []string
	Category    string
}

// ExecutionResult represents the result of a command execution
type ExecutionResult struct {
	Command   string
	Args      []string
	Output    string
	Error     string
	ExitCode  int
	Duration  time.Duration
	Timestamp time.Time
	Streaming bool
}

// HistoryEntry represents a command history entry
type HistoryEntry struct {
	Command   string
	Timestamp time.Time
	Success   bool
	Duration  time.Duration
}

// Session represents a command session (for multi-session support)
type Session struct {
	ID      string
	Name    string
	History []HistoryEntry
	Output  []string
	Current string
	Created time.Time
	Active  bool
}

// NewAppState creates a new application state with default values
func NewAppState() *AppState {
	return &AppState{
		PanelRatio:       0.3,
		MaxHistory:       1000,
		History:          make([]HistoryEntry, 0),
		MCPServerRunning: false,
		AIProvider:       "Unknown",
		NixOSPath:        "/etc/nixos",
	}
}

// AddHistoryEntry adds a new entry to the command history
func (s *AppState) AddHistoryEntry(entry HistoryEntry) {
	s.History = append(s.History, entry)

	// Limit history size
	if len(s.History) > s.MaxHistory {
		s.History = s.History[len(s.History)-s.MaxHistory:]
	}
}

// GetRecentCommands returns the most recent commands
func (s *AppState) GetRecentCommands(limit int) []string {
	if len(s.History) == 0 {
		return []string{}
	}

	start := len(s.History) - limit
	if start < 0 {
		start = 0
	}

	commands := make([]string, 0, limit)
	for i := start; i < len(s.History); i++ {
		commands = append(commands, s.History[i].Command)
	}

	// Reverse to show most recent first
	for i, j := 0, len(commands)-1; i < j; i, j = i+1, j-1 {
		commands[i], commands[j] = commands[j], commands[i]
	}

	return commands
}
