package models

import "strings"

// CommandMetadata represents detailed information about a command
type CommandMetadata struct {
	Name        string
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
			Description: "Ask AI questions about NixOS",
			Usage:       "ask \"your question\"",
			Examples:    []string{"ask \"how to enable SSH?\"", "ask \"what is a flake?\"", "ask \"configure nginx\" --quiet"},
			Category:    "AI",
		},
		{
			Name:        "build",
			Description: "Build NixOS configuration",
			Usage:       "build [options]",
			Examples:    []string{"build", "build --check"},
			Category:    "System",
		},
		{
			Name:        "community",
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
			Description: "Manage nixai configuration",
			Usage:       "config [subcommand]",
			Examples:    []string{"config show", "config edit"},
			Category:    "Configuration",
		},
		{
			Name:        "configure",
			Description: "Configure NixOS system",
			Usage:       "configure [options]",
			Examples:    []string{"configure --help"},
			Category:    "Configuration",
		},
		{
			Name:        "diagnose",
			Description: "Diagnose system issues",
			Usage:       "diagnose [options]",
			Examples:    []string{"diagnose", "diagnose --verbose"},
			Category:    "System",
		},
		{
			Name:        "doctor",
			Description: "Run system health checks",
			Usage:       "doctor [options]",
			Examples:    []string{"doctor", "doctor --fix"},
			Category:    "System",
		},
		{
			Name:        "explain-option",
			Description: "Explain NixOS configuration options",
			Usage:       "explain-option <option>",
			Examples:    []string{"explain-option services.openssh", "explain-option boot.loader"},
			Category:    "AI",
		},
		{
			Name:        "flake",
			Description: "Manage Nix flakes",
			Usage:       "flake [subcommand]",
			Examples:    []string{"flake init", "flake update"},
			Category:    "Development",
		},
		{
			Name:        "gc",
			Description: "Garbage collect Nix store",
			Usage:       "gc [options]",
			Examples:    []string{"gc", "gc --delete-older-than 30d"},
			Category:    "System",
		},
		{
			Name:        "hardware",
			Description: "Manage hardware configuration",
			Usage:       "hardware [subcommand]",
			Examples:    []string{"hardware scan", "hardware gpu"},
			Category:    "Configuration",
		},
		{
			Name:        "learn",
			Description: "Interactive learning system",
			Usage:       "learn [topic]",
			Examples:    []string{"learn basics", "learn flakes"},
			Category:    "AI",
		},
		{
			Name:        "logs",
			Description: "View and analyze system logs",
			Usage:       "logs [options]",
			Examples:    []string{"logs", "logs --follow"},
			Category:    "System",
		},
		{
			Name:        "machines",
			Description: "Manage remote machines",
			Usage:       "machines [subcommand]",
			Examples:    []string{"machines list", "machines deploy"},
			Category:    "Development",
		},
		{
			Name:        "mcp-server",
			Description: "Manage MCP server",
			Usage:       "mcp-server [subcommand]",
			Examples:    []string{"mcp-server start", "mcp-server status"},
			Category:    "System",
		},
		{
			Name:        "migrate",
			Description: "Migrate configurations",
			Usage:       "migrate [options]",
			Examples:    []string{"migrate --from-ubuntu", "migrate --backup"},
			Category:    "Configuration",
		},
		{
			Name:        "package-repo",
			Description: "Analyze and package repositories",
			Usage:       "package-repo <url>",
			Examples:    []string{"package-repo https://github.com/user/repo"},
			Category:    "Development",
		},
		{
			Name:        "search",
			Description: "Search packages and options",
			Usage:       "search <query>",
			Examples:    []string{"search firefox", "search \"text editor\""},
			Category:    "Utilities",
		},
		{
			Name:        "snippets",
			Description: "Manage configuration snippets",
			Usage:       "snippets [subcommand]",
			Examples:    []string{"snippets list", "snippets add"},
			Category:    "Utilities",
		},
		{
			Name:        "store",
			Description: "Manage Nix store",
			Usage:       "store [subcommand]",
			Examples:    []string{"store info", "store optimize"},
			Category:    "System",
		},
		{
			Name:        "templates",
			Description: "Manage configuration templates",
			Usage:       "templates [subcommand]",
			Examples:    []string{"templates list", "templates apply"},
			Category:    "Configuration",
		},
		{
			Name:        "exit",
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
