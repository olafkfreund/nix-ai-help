package services

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"nix-ai-help/internal/cli"
	"nix-ai-help/internal/tui/models"
	"nix-ai-help/internal/tui/panels"

	tea "github.com/charmbracelet/bubbletea"
)

// CommandExecutor handles command execution for the TUI
type CommandExecutor struct {
	// Add any dependencies here if needed
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// ExecuteCommand executes a command and returns a tea command that sends the result
func (e *CommandExecutor) ExecuteCommand(command string) tea.Cmd {
	return func() tea.Msg {
		// Parse command and arguments
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return panels.CommandExecutionResultMsg{
				Result: models.ExecutionResult{
					Command:   command,
					Error:     "Empty command",
					ExitCode:  1,
					Duration:  0,
					Timestamp: time.Now(),
				},
			}
		}

		cmdName := parts[0]
		args := parts[1:]

		// Start execution
		startTime := time.Now()

		// Try to execute through nixai's command system first
		if e.isNixAICommand(cmdName) {
			return e.executeNixAICommand(cmdName, args, startTime)
		}

		// Fall back to system command execution
		return e.executeSystemCommand(cmdName, args, startTime)
	}
}

// isNixAICommand checks if the command is a nixai command
func (e *CommandExecutor) isNixAICommand(cmdName string) bool {
	nixaiCommands := []string{
		"ask", "build", "community", "config", "configure", "diagnose", "doctor",
		"explain-option", "explain-home-option", "flake", "gc", "hardware", "learn",
		"logs", "machines", "mcp-server", "migrate", "package-repo", "search",
		"snippets", "store", "templates", "interactive", "help", "version",
	}

	for _, cmd := range nixaiCommands {
		if cmd == cmdName {
			return true
		}
	}
	return false
}

// executeNixAICommand executes a nixai command using the existing command system
func (e *CommandExecutor) executeNixAICommand(cmdName string, args []string, startTime time.Time) panels.CommandExecutionResultMsg {
	// Use the existing RunDirectCommand function from the CLI package
	output, err := cli.RunDirectCommand(cmdName, args)

	duration := time.Since(startTime)
	exitCode := 0
	errorMsg := ""

	if err != nil {
		errorMsg = err.Error()
		exitCode = 1
	}

	return panels.CommandExecutionResultMsg{
		Result: models.ExecutionResult{
			Command:   cmdName,
			Args:      args,
			Output:    output,
			Error:     errorMsg,
			ExitCode:  exitCode,
			Duration:  duration,
			Timestamp: startTime,
			Streaming: false,
		},
	}
}

// executeSystemCommand executes a system command
func (e *CommandExecutor) executeSystemCommand(cmdName string, args []string, startTime time.Time) panels.CommandExecutionResultMsg {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, args...)
	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)
	exitCode := 0
	errorMsg := ""

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
		errorMsg = err.Error()
	}

	return panels.CommandExecutionResultMsg{
		Result: models.ExecutionResult{
			Command:   cmdName,
			Args:      args,
			Output:    string(output),
			Error:     errorMsg,
			ExitCode:  exitCode,
			Duration:  duration,
			Timestamp: startTime,
			Streaming: false,
		},
	}
}

// ExecuteCommandStream executes a command with streaming output
func (e *CommandExecutor) ExecuteCommandStream(command string) tea.Cmd {
	return func() tea.Msg {
		// For now, just use the regular execution
		// TODO: Implement streaming execution for long-running commands
		return e.ExecuteCommand(command)()
	}
}
