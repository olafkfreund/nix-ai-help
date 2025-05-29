package nixos

import (
	"fmt"
	"os/exec"
	"strings"
)

// Executor provides functionality to execute local commands related to NixOS configuration.
type Executor struct {
	ConfigPath string
}

// NewExecutor creates a new instance of Executor with an optional config path.
func NewExecutor(configPath string) *Executor {
	return &Executor{ConfigPath: configPath}
}

// ExecuteCommand executes a given command and returns its output.
func (e *Executor) ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	if e.ConfigPath != "" {
		cmd.Dir = e.ConfigPath
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExecuteNixCommand executes a NixOS specific command and returns the output.
func (e *Executor) ExecuteNixCommand(command string) (string, error) {
	return e.ExecuteCommand("nix", strings.Fields(command)...)
}

// SearchNixPackages searches for Nix packages using `nix search nixpkgs <query>` and returns the output.
func (e *Executor) SearchNixPackages(query string) (string, error) {
	args := []string{"search", "nixpkgs"}
	if strings.TrimSpace(query) != "" {
		qargs := strings.Fields(query)
		fmt.Printf("[DEBUG] SearchNixPackages: args = %v\n", append(args, qargs...))
		args = append(args, qargs...)
	} else {
		fmt.Printf("[DEBUG] SearchNixPackages: args = %v\n", args)
	}
	output, err := e.ExecuteCommand("nix", args...)
	fmt.Printf("[DEBUG] SearchNixPackages: raw output =\n%s\n", output)
	return output, err
}

// SearchNixServices searches for NixOS services using `nix search nixos` and returns the output.
func (e *Executor) SearchNixServices(query string) (string, error) {
	return e.ExecuteCommand("nix", "search", "nixos", query)
}

// InstallNixPackage installs a package using `nix-env -iA` and returns the output.
func (e *Executor) InstallNixPackage(attr string) (string, error) {
	return e.ExecuteCommand("nix-env", "-iA", attr)
}

// ShowNixOSOptions runs 'nixos-option <option>' and returns the output.
func (e *Executor) ShowNixOSOptions(option string) (string, error) {
	return e.ExecuteCommand("nixos-option", option)
}

// ListServiceOptions lists all available options for a given NixOS service using 'nixos-option --find services.<name>'.
func (e *Executor) ListServiceOptions(service string) (string, error) {
	if strings.TrimSpace(service) == "" {
		return "", fmt.Errorf("service name is required")
	}
	return e.ExecuteCommand("nixos-option", "--find", fmt.Sprintf("services.%s", service))
}
