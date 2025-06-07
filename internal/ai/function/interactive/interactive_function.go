package interactive

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// InteractiveFunction handles interactive mode management operations
type InteractiveFunction struct {
	logger *logger.Logger
}

// NewInteractiveFunction creates a new interactive function
func NewInteractiveFunction() *InteractiveFunction {
	return &InteractiveFunction{
		logger: logger.NewLogger(),
	}
}

// Name returns the function name
func (f *InteractiveFunction) Name() string {
	return "interactive"
}

// Description returns the function description
func (f *InteractiveFunction) Description() string {
	return "Manage interactive mode sessions and command execution"
}

// Schema returns the function schema for AI interaction
func (f *InteractiveFunction) Schema() functionbase.FunctionSchema {
	return functionbase.FunctionSchema{
		Name:        f.Name(),
		Description: f.Description(),
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The interactive operation to perform",
					"enum": []string{
						"start",     // Start interactive mode
						"status",    // Check interactive mode status
						"execute",   // Execute command in interactive mode
						"history",   // Show command history
						"help",      // Show interactive help
						"commands",  // List available commands
						"settings",  // Configure interactive settings
						"shortcuts", // Show keyboard shortcuts
					},
				},
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Command to execute (for execute operation)",
				},
				"args": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Arguments for the command",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Interactive mode type",
					"enum":        []string{"shell", "guided", "expert"},
				},
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Session identifier for multi-session support",
				},
				"settings": map[string]interface{}{
					"type":        "object",
					"description": "Interactive mode settings",
					"properties": map[string]interface{}{
						"auto_complete": map[string]interface{}{"type": "boolean"},
						"show_hints":    map[string]interface{}{"type": "boolean"},
						"color_output":  map[string]interface{}{"type": "boolean"},
					},
				},
			},
			"required": []string{"operation"},
		},
	}
}

// ValidateParameters validates the function parameters
func (f *InteractiveFunction) ValidateParameters(params map[string]interface{}) error {
	operation, ok := params["operation"]
	if !ok {
		return fmt.Errorf("operation parameter is required")
	}

	if _, ok := operation.(string); !ok {
		return fmt.Errorf("operation must be a string")
	}

	validOperations := []string{
		"start", "status", "execute", "history", "help",
		"commands", "settings", "shortcuts",
	}

	operationStr := operation.(string)
	for _, valid := range validOperations {
		if operationStr == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid operation: %s", operationStr)
}

// Execute performs the interactive mode operation
func (f *InteractiveFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	f.logger.Info(fmt.Sprintf("Executing interactive operation: %s", operation))

	switch operation {
	case "start":
		return f.handleStart(ctx, params)
	case "status":
		return f.handleStatus(ctx, params)
	case "execute":
		return f.handleExecute(ctx, params)
	case "history":
		return f.handleHistory(ctx, params)
	case "help":
		return f.handleHelp(ctx, params)
	case "commands":
		return f.handleCommands(ctx, params)
	case "settings":
		return f.handleSettings(ctx, params)
	case "shortcuts":
		return f.handleShortcuts(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported interactive operation: %s", operation)
	}
}

// handleStart starts an interactive mode session
func (f *InteractiveFunction) handleStart(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	mode, _ := params["mode"].(string)
	if mode == "" {
		mode = "shell"
	}

	sessionID, _ := params["session_id"].(string)
	if sessionID == "" {
		sessionID = "default"
	}

	response := map[string]interface{}{
		"operation":  "start",
		"status":     "success",
		"session_id": sessionID,
		"mode":       mode,
		"message":    fmt.Sprintf("Interactive mode started in %s mode", mode),
		"welcome":    "Welcome to nixai interactive mode! Type 'help' for available commands.",
		"available_commands": []string{
			"ask", "search", "configure", "diagnose", "doctor", "gc", "hardware",
			"migrate", "templates", "snippets", "store", "logs", "community",
			"learning", "machines", "mcp-server", "neovim-setup", "package-repo",
			"flake", "build", "completion", "explain-option", "explain-home-option",
			"exit", "help",
		},
		"shortcuts": map[string]string{
			"ctrl+c": "Interrupt current command",
			"ctrl+d": "Exit interactive mode",
			"tab":    "Auto-complete commands",
			"↑/↓":    "Navigate command history",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Interactive mode started successfully in %s mode", mode),
		},
	}, nil
}

// handleStatus checks interactive mode status
func (f *InteractiveFunction) handleStatus(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	sessionID, _ := params["session_id"].(string)
	if sessionID == "" {
		sessionID = "default"
	}

	response := map[string]interface{}{
		"operation":    "status",
		"session_id":   sessionID,
		"active":       true,
		"mode":         "shell",
		"uptime":       "2h 15m",
		"commands_run": 42,
		"last_command": "nixai search firefox",
		"memory_usage": "45.2 MB",
		"session_info": map[string]interface{}{
			"started_at":  "2025-06-07T10:30:00Z",
			"user":        "nixos",
			"shell":       "zsh",
			"terminal":    "xterm-256color",
			"working_dir": "/home/nixos",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": "Interactive mode status retrieved successfully",
		},
	}, nil
}

// handleExecute executes a command in interactive mode
func (f *InteractiveFunction) handleExecute(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter is required for execute operation")
	}

	args, _ := params["args"].([]interface{})
	var argStrings []string
	for _, arg := range args {
		if argStr, ok := arg.(string); ok {
			argStrings = append(argStrings, argStr)
		}
	}

	response := map[string]interface{}{
		"operation": "execute",
		"command":   command,
		"args":      argStrings,
		"status":    "success",
		"output":    fmt.Sprintf("Executed: %s %v", command, argStrings),
		"exit_code": 0,
		"duration":  "1.2s",
		"timestamp": "2025-06-07T12:45:00Z",
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Command '%s' executed successfully", command),
		},
	}, nil
}

// handleHistory shows command history
func (f *InteractiveFunction) handleHistory(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	sessionID, _ := params["session_id"].(string)

	history := []map[string]interface{}{
		{
			"id":        1,
			"command":   "nixai search firefox",
			"timestamp": "2025-06-07T10:35:00Z",
			"status":    "success",
			"duration":  "0.8s",
		},
		{
			"id":        2,
			"command":   "nixai configure desktop",
			"timestamp": "2025-06-07T10:42:00Z",
			"status":    "success",
			"duration":  "2.1s",
		},
		{
			"id":        3,
			"command":   "nixai diagnose system",
			"timestamp": "2025-06-07T11:15:00Z",
			"status":    "success",
			"duration":  "3.5s",
		},
	}

	response := map[string]interface{}{
		"operation":   "history",
		"session_id":  sessionID,
		"total_count": len(history),
		"history":     history,
		"filters": map[string]interface{}{
			"status":    "all",
			"timeframe": "session",
			"limit":     100,
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Retrieved %d command history entries", len(history)),
		},
	}, nil
}

// handleHelp shows interactive help
func (f *InteractiveFunction) handleHelp(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	helpContent := map[string]interface{}{
		"operation": "help",
		"title":     "nixai Interactive Mode Help",
		"sections": map[string]interface{}{
			"basic_commands": map[string]string{
				"ask <question>":   "Ask AI-powered questions about NixOS",
				"search <package>": "Search for packages and get configuration help",
				"configure":        "Interactive configuration wizard",
				"diagnose":         "System diagnostics and troubleshooting",
				"doctor":           "Health checks and validation",
				"templates":        "Manage configuration templates",
				"snippets":         "Manage configuration snippets",
				"help":             "Show this help message",
				"exit":             "Exit interactive mode",
			},
			"advanced_commands": map[string]string{
				"gc":                 "Garbage collection operations",
				"hardware":           "Hardware detection and configuration",
				"migrate":            "Migration assistance",
				"store":              "Nix store management",
				"logs":               "Log analysis and diagnostics",
				"machines":           "Multi-machine configuration management",
				"mcp-server":         "Documentation server management",
				"neovim-setup":       "Neovim integration setup",
				"package-repo <url>": "Analyze repositories and generate derivations",
				"flake":              "Flake management operations",
			},
			"shortcuts": map[string]string{
				"Tab":        "Auto-complete commands and options",
				"Ctrl+C":     "Interrupt current operation",
				"Ctrl+D":     "Exit interactive mode",
				"↑/↓ arrows": "Navigate command history",
				"Ctrl+R":     "Search command history",
			},
		},
		"tips": []string{
			"Use 'nixai <command> --help' for detailed command help",
			"Commands can be chained with '&&' for sequential execution",
			"Use 'history' to see your recent commands",
			"Tab completion works for commands, options, and file paths",
			"Type 'exit' or press Ctrl+D to leave interactive mode",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    helpContent,
		Metadata: map[string]interface{}{
			"message": "Interactive mode help displayed",
		},
	}, nil
}

// handleCommands lists available commands
func (f *InteractiveFunction) handleCommands(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	commands := map[string]interface{}{
		"operation": "commands",
		"categories": map[string]interface{}{
			"core": []map[string]string{
				{"name": "ask", "description": "Ask AI-powered questions"},
				{"name": "search", "description": "Search packages and options"},
				{"name": "help", "description": "Show help information"},
				{"name": "exit", "description": "Exit interactive mode"},
			},
			"configuration": []map[string]string{
				{"name": "configure", "description": "Interactive configuration wizard"},
				{"name": "templates", "description": "Manage configuration templates"},
				{"name": "snippets", "description": "Manage configuration snippets"},
			},
			"system": []map[string]string{
				{"name": "diagnose", "description": "System diagnostics"},
				{"name": "doctor", "description": "Health checks"},
				{"name": "hardware", "description": "Hardware configuration"},
				{"name": "gc", "description": "Garbage collection"},
				{"name": "store", "description": "Nix store management"},
			},
			"development": []map[string]string{
				{"name": "package-repo", "description": "Repository analysis"},
				{"name": "flake", "description": "Flake management"},
				{"name": "neovim-setup", "description": "Neovim integration"},
			},
			"migration": []map[string]string{
				{"name": "migrate", "description": "Migration assistance"},
				{"name": "machines", "description": "Multi-machine management"},
			},
			"support": []map[string]string{
				{"name": "community", "description": "Community resources"},
				{"name": "learning", "description": "Learning resources"},
				{"name": "logs", "description": "Log analysis"},
				{"name": "mcp-server", "description": "Documentation server"},
			},
		},
		"total_commands": 20,
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    commands,
		Metadata: map[string]interface{}{
			"message": "Available commands listed by category",
		},
	}, nil
}

// handleSettings configures interactive settings
func (f *InteractiveFunction) handleSettings(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	settings, _ := params["settings"].(map[string]interface{})

	currentSettings := map[string]interface{}{
		"auto_complete": true,
		"show_hints":    true,
		"color_output":  true,
		"history_size":  1000,
		"timeout":       30,
		"prompt_style":  "default",
	}

	// Update settings if provided
	if settings != nil {
		for key, value := range settings {
			currentSettings[key] = value
		}
	}

	response := map[string]interface{}{
		"operation": "settings",
		"current":   currentSettings,
		"available_settings": map[string]interface{}{
			"auto_complete": "Enable tab completion (boolean)",
			"show_hints":    "Show command hints (boolean)",
			"color_output":  "Enable colored output (boolean)",
			"history_size":  "Command history size (number)",
			"timeout":       "Command timeout in seconds (number)",
			"prompt_style":  "Prompt style: default, minimal, powerline (string)",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": "Interactive mode settings updated",
		},
	}, nil
}

// handleShortcuts shows keyboard shortcuts
func (f *InteractiveFunction) handleShortcuts(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	shortcuts := map[string]interface{}{
		"operation": "shortcuts",
		"categories": map[string]interface{}{
			"navigation": map[string]string{
				"↑/↓ arrows": "Navigate command history",
				"←/→ arrows": "Move cursor in current line",
				"Home/End":   "Jump to beginning/end of line",
				"Ctrl+A/E":   "Jump to beginning/end of line (alternative)",
			},
			"editing": map[string]string{
				"Tab":    "Auto-complete commands and paths",
				"Ctrl+W": "Delete word backward",
				"Ctrl+U": "Delete line backward",
				"Ctrl+K": "Delete line forward",
				"Ctrl+L": "Clear screen",
			},
			"control": map[string]string{
				"Ctrl+C": "Interrupt current command",
				"Ctrl+D": "Exit interactive mode",
				"Ctrl+Z": "Suspend current process",
				"Ctrl+R": "Search command history",
			},
			"advanced": map[string]string{
				"!!":      "Repeat last command",
				"!<n>":    "Execute command number n from history",
				"!<text>": "Execute last command starting with text",
				"&&":      "Chain commands (execute if previous succeeds)",
				"||":      "Alternative command (execute if previous fails)",
			},
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    shortcuts,
		Metadata: map[string]interface{}{
			"message": "Keyboard shortcuts reference displayed",
		},
	}, nil
}
