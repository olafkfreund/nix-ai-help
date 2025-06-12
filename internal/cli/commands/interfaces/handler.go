package interfaces

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// CommandHandler defines the interface that all command modules must implement.
// This interface provides methods for getting command information and execution.
type CommandHandler interface {
	// GetCommand returns the cobra.Command instance for this handler
	GetCommand() *cobra.Command

	// Execute runs the command with the provided context and arguments
	Execute(ctx context.Context, args []string) (*CommandResult, error)

	// GetHelp returns detailed help information for the command
	GetHelp() string

	// GetExamples returns usage examples for the command
	GetExamples() []string
}

// CommandResult encapsulates the result of executing a command
type CommandResult struct {
	// Output contains the command's output text
	Output string

	// Error contains any error that occurred during execution
	Error error

	// ExitCode is the exit status code for the command
	ExitCode int

	// Duration is how long the command took to execute
	Duration time.Duration

	// Metadata contains additional command-specific information
	Metadata map[string]interface{}
}
