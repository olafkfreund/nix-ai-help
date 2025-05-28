package nixos

import (
	"os/exec"
	"strings"
)

// Executor provides functionality to execute local commands related to NixOS configuration.
type Executor struct{}

// NewExecutor creates a new instance of Executor.
func NewExecutor() *Executor {
	return &Executor{}
}

// ExecuteCommand executes a given command and returns its output.
func (e *Executor) ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExecuteNixCommand executes a NixOS specific command and returns the output.
func (e *Executor) ExecuteNixCommand(command string) (string, error) {
	return e.ExecuteCommand("nix", strings.Fields(command)...)
}
