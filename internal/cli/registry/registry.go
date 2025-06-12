package registry

import (
	"nix-ai-help/internal/cli/commands/interfaces"
	"sync"

	"github.com/spf13/cobra"
)

// CommandRegistry manages the registration and retrieval of command handlers.
// It provides a central location for registering, accessing, and executing commands.
type CommandRegistry struct {
	// commands is a map of command names to their handlers
	commands map[string]interfaces.CommandHandler

	// rootCmd is the root Cobra command that all other commands are added to
	rootCmd *cobra.Command

	// mu protects concurrent access to the commands map
	mu sync.RWMutex
}

// NewRegistry creates a new command registry with the given root command
func NewRegistry(rootCmd *cobra.Command) *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]interfaces.CommandHandler),
		rootCmd:  rootCmd,
	}
}

// Register adds a command handler to the registry and attaches the Cobra command to the root command
func (r *CommandRegistry) Register(name string, handler interfaces.CommandHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store the handler in our registry
	r.commands[name] = handler

	// Get the Cobra command and add it to the root
	cmd := handler.GetCommand()
	r.rootCmd.AddCommand(cmd)
}

// Get retrieves a command handler by name
func (r *CommandRegistry) Get(name string) (interfaces.CommandHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.commands[name]
	return handler, exists
}

// GetAll returns all registered command handlers
func (r *CommandRegistry) GetAll() map[string]interfaces.CommandHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Make a copy to avoid concurrent access issues
	result := make(map[string]interfaces.CommandHandler, len(r.commands))
	for k, v := range r.commands {
		result[k] = v
	}

	return result
}

// GetRootCommand returns the root cobra.Command
func (r *CommandRegistry) GetRootCommand() *cobra.Command {
	return r.rootCmd
}
