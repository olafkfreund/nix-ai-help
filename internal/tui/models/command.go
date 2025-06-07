package models

import "strings"

// CommandMetadata represents detailed information about a command
type CommandMetadata struct {
	Name        string
	Icon        string
	Description string
	Usage       string
	Examples    []string
	Subcommands []CommandMetadata
	Category    string
	Hidden      bool
}

// CommandCategories represents different categories of commands
type CommandCategories struct {
	AI            []CommandMetadata
	System        []CommandMetadata
	Development   []CommandMetadata
	Configuration []CommandMetadata
	Utilities     []CommandMetadata
}

// GetAllCommands returns all available commands
func GetDefaultCommands() []CommandMetadata {
	return []CommandMetadata{
		{
			Name:        "ask",
			Icon:        "🤖",
			Description: "Ask AI questions about NixOS",
			Usage:       "ask \"your question\"",
			Examples:    []string{"ask \"how to enable SSH?\"", "ask \"what is a flake?\""},
			Category:    "AI",
		},
		{
			Name:        "build",
			Icon:        "🛠️",
			Description: "Build NixOS configuration",
			Usage:       "build [options]",
			Examples:    []string{"build", "build --check"},
			Category:    "System",
		},
		{
			Name:        "community",
			Icon:        "🌐",
			Description: "Access community resources",
			Usage:       "community [subcommand]",
			Examples:    []string{"community forums", "community discord"},
			Category:    "Utilities",
			Subcommands: []CommandMetadata{
				{Name: "forums", Description: "Browse community forums"},
				{Name: "discord", Description: "Join Discord server"},
				{Name: "github", Description: "Visit GitHub repositories"},
			},
		},
		{
			Name:        "config",
			Icon:        "⚙️",
			Description: "Manage nixai configuration",
			Usage:       "config [subcommand]",
			Examples:    []string{"config show", "config edit"},
			Category:    "Configuration",
		},
		{
			Name:        "configure",
			Icon:        "🧑‍💻",
			Description: "Configure NixOS system",
			Usage:       "configure [options]",
			Examples:    []string{"configure --help"},
			Category:    "Configuration",
		},
		{
			Name:        "diagnose",
			Icon:        "🩺",
			Description: "Diagnose system issues",
			Usage:       "diagnose [options]",
			Examples:    []string{"diagnose", "diagnose --verbose"},
			Category:    "System",
		},
		{
			Name:        "doctor",
			Icon:        "🩻",
			Description: "Run system health checks",
			Usage:       "doctor [options]",
			Examples:    []string{"doctor", "doctor --fix"},
			Category:    "System",
		},
		{
			Name:        "explain-option",
			Icon:        "🖥️",
			Description: "Explain NixOS configuration options",
			Usage:       "explain-option <option>",
			Examples:    []string{"explain-option services.openssh", "explain-option boot.loader"},
			Category:    "AI",
		},
		{
			Name:        "flake",
			Icon:        "🧊",
			Description: "Manage Nix flakes",
			Usage:       "flake [subcommand]",
			Examples:    []string{"flake init", "flake update"},
			Category:    "Development",
		},
		{
			Name:        "gc",
			Icon:        "🧹",
			Description: "Garbage collect Nix store",
			Usage:       "gc [options]",
			Examples:    []string{"gc", "gc --delete-older-than 30d"},
			Category:    "System",
		},
		{
			Name:        "hardware",
			Icon:        "💻",
			Description: "Manage hardware configuration",
			Usage:       "hardware [subcommand]",
			Examples:    []string{"hardware scan", "hardware gpu"},
			Category:    "Configuration",
		},
		{
			Name:        "learn",
			Icon:        "📚",
			Description: "Interactive learning system",
			Usage:       "learn [topic]",
			Examples:    []string{"learn basics", "learn flakes"},
			Category:    "AI",
		},
		{
			Name:        "logs",
			Icon:        "📝",
			Description: "View and analyze system logs",
			Usage:       "logs [options]",
			Examples:    []string{"logs", "logs --follow"},
			Category:    "System",
		},
		{
			Name:        "machines",
			Icon:        "🖧",
			Description: "Manage remote machines",
			Usage:       "machines [subcommand]",
			Examples:    []string{"machines list", "machines deploy"},
			Category:    "Development",
		},
		{
			Name:        "mcp-server",
			Icon:        "🛰️",
			Description: "Manage MCP server",
			Usage:       "mcp-server [subcommand]",
			Examples:    []string{"mcp-server start", "mcp-server status"},
			Category:    "System",
		},
		{
			Name:        "migrate",
			Icon:        "🔀",
			Description: "Migrate configurations",
			Usage:       "migrate [options]",
			Examples:    []string{"migrate --from-ubuntu", "migrate --backup"},
			Category:    "Configuration",
		},
		{
			Name:        "package-repo",
			Icon:        "📦",
			Description: "Analyze and package repositories",
			Usage:       "package-repo <url>",
			Examples:    []string{"package-repo https://github.com/user/repo"},
			Category:    "Development",
		},
		{
			Name:        "search",
			Icon:        "🔍",
			Description: "Search packages and options",
			Usage:       "search <query>",
			Examples:    []string{"search firefox", "search \"text editor\""},
			Category:    "Utilities",
		},
		{
			Name:        "snippets",
			Icon:        "🔖",
			Description: "Manage configuration snippets",
			Usage:       "snippets [subcommand]",
			Examples:    []string{"snippets list", "snippets add"},
			Category:    "Utilities",
		},
		{
			Name:        "store",
			Icon:        "💾",
			Description: "Manage Nix store",
			Usage:       "store [subcommand]",
			Examples:    []string{"store info", "store optimize"},
			Category:    "System",
		},
		{
			Name:        "templates",
			Icon:        "📄",
			Description: "Manage configuration templates",
			Usage:       "templates [subcommand]",
			Examples:    []string{"templates list", "templates apply"},
			Category:    "Configuration",
		},
		{
			Name:        "exit",
			Icon:        "❌",
			Description: "Exit interactive mode",
			Usage:       "exit",
			Examples:    []string{"exit"},
			Category:    "Utilities",
			Hidden:      true,
		},
	}
}

// FilterCommands filters commands based on a search query
func FilterCommands(commands []CommandMetadata, query string) []CommandMetadata {
	if query == "" {
		return commands
	}

	var filtered []CommandMetadata
	for _, cmd := range commands {
		if matchesQuery(cmd, query) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// matchesQuery checks if a command matches the search query
func matchesQuery(cmd CommandMetadata, query string) bool {
	// Simple case-insensitive matching
	queryLower := strings.ToLower(query)
	return strings.Contains(strings.ToLower(cmd.Name), queryLower) ||
		strings.Contains(strings.ToLower(cmd.Description), queryLower) ||
		strings.Contains(strings.ToLower(cmd.Category), queryLower)
}
