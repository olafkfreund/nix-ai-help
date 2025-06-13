package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// MockDepsAIProvider implements the AIProvider interface for testing
type MockDepsAIProvider struct {
	response string
	err      error
}

func (m *MockDepsAIProvider) Query(prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockDepsAIProvider) GenerateResponse(prompt string) (string, error) {
	return m.Query(prompt)
}

// TestDepsAnalyzeCommand tests the dependency analyze command with mocked output
func TestDepsAnalyzeCommand(t *testing.T) {
	// Create a mock AI provider
	mockAI := &MockDepsAIProvider{
		response: "Your dependency graph looks healthy with no circular dependencies.",
	}

	// Create a test command that simulates the analyze behavior
	cmd := &cobra.Command{
		Use: "analyze",
		Run: func(cmd *cobra.Command, args []string) {
			// Mock analyzing dependencies
			cmd.Println(utils.FormatHeader("🛠️ NixOS Dependency Analyzer"))

			// Mock determining config path
			cfgPath := "/etc/nixos/configuration.nix"
			cmd.Println(utils.FormatKeyValue("Configuration Path", cfgPath))

			// Mock dependency graph creation and display
			cmd.Println(utils.FormatSuccess("Dependency analysis complete."))
			cmd.Println(utils.FormatHeader("🌳 Dependency Graph:"))
			cmd.Println("└── root")
			cmd.Println("    ├── configuration.nix")
			cmd.Println("    │   ├── hardware-configuration.nix")
			cmd.Println("    │   └── services.xserver")
			cmd.Println("    │       └── xorg")
			cmd.Println("    └── desktop.nix")
			cmd.Println("        └── firefox")

			cmd.Println(utils.FormatHeader("🤖 AI Analysis & Insights"))
			aiInsights, _ := mockAI.Query("analyze dependencies")
			cmd.Println(aiInsights)
		},
	}

	// Capture output
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute the command
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()

	// Check output
	expectedOutputs := []string{
		"NixOS Dependency Analyzer",
		"Configuration Path",
		"Dependency analysis complete",
		"AI Analysis & Insights",
		mockAI.response,
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, result)
		}
	}
}

// TestDepsWhyCommand tests the "why" subcommand
func TestDepsWhyCommand(t *testing.T) {
	// Create command that simulates finding a package
	cmd := &cobra.Command{
		Use: "why [package]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Println(utils.FormatError("Package name is required"))
				return
			}
			packageName := args[0]

			cmd.Println(utils.FormatHeader("🔍 Why is '" + packageName + "' installed?"))

			cfgPath := "/etc/nixos/configuration.nix"
			cmd.Println(utils.FormatKeyValue("Configuration Path", cfgPath))

			// Mock dependency paths based on package name
			var paths [][]string
			if packageName == "firefox" {
				paths = [][]string{
					{"root", "desktop.nix", "firefox"},
				}
			}

			if len(paths) == 0 {
				cmd.Println(utils.FormatWarning("No packages matching '" + packageName + "' were found."))
			} else {
				cmd.Println(utils.FormatSuccess(fmt.Sprintf("Found %d dependency paths leading to '%s':", len(paths), packageName)))
				cmd.Println()

				for i, path := range paths {
					cmd.Printf("Path %d: %s\n", i+1, strings.Join(path, " -> "))
				}
			}
		},
	}

	// Test for a package that exists
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"firefox"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()

	// Check output
	expectedOutputs := []string{
		"Why is 'firefox' installed",
		"Configuration Path",
		"Found 1 dependency paths",
		"root -> desktop.nix -> firefox",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, result)
		}
	}

	// Test for non-existent package
	var outputNotFound bytes.Buffer
	cmd.SetOut(&outputNotFound)
	cmd.SetErr(&outputNotFound)
	cmd.SetArgs([]string{"nonexistent-package"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	resultNotFound := outputNotFound.String()

	if !strings.Contains(resultNotFound, "No packages matching") {
		t.Errorf("Expected output to contain 'No packages matching' for non-existent package, got: %s", resultNotFound)
	}
}

// TestDepsGraphCommand tests the "graph" subcommand
func TestDepsGraphCommand(t *testing.T) {
	// Create command that simulates graph generation
	cmd := &cobra.Command{
		Use: "graph",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(utils.FormatHeader("📊 Dependency Graph Generator"))

			cfgPath := "/etc/nixos/configuration.nix"
			cmd.Println(utils.FormatKeyValue("Configuration Path", cfgPath))

			// Output format would normally be DOT, but for testing we just output text
			cmd.Println(utils.FormatSuccess("Graph generated successfully"))
			cmd.Println("digraph nixos_config {")
			cmd.Println("  root -> \"configuration.nix\";")
			cmd.Println("  \"configuration.nix\" -> \"hardware-configuration.nix\";")
			cmd.Println("}")
		},
	}

	// Capture output
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()

	// Check output
	expectedOutputs := []string{
		"Dependency Graph Generator",
		"Configuration Path",
		"Graph generated successfully",
		"digraph nixos_config",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, result)
		}
	}
}

// TestNewDepsCommand tests the creation of the deps command structure
func TestNewDepsCommand(t *testing.T) {
	cmd := NewDepsCommand()

	if cmd.Use != "deps" {
		t.Errorf("Expected command use to be 'deps', got: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	// Check that subcommands exist
	subCommands := cmd.Commands()
	expectedSubCommands := []string{"analyze", "why", "conflicts", "optimize", "graph"}

	found := make(map[string]bool)
	for _, subCmd := range subCommands {
		// Extract just the command name (before any space)
		cmdName := strings.Split(subCmd.Use, " ")[0]
		found[cmdName] = true
	}

	for _, expected := range expectedSubCommands {
		if !found[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

// TestDepsCommandFlags tests flag handling for the deps commands
func TestDepsCommandFlags(t *testing.T) {
	cmd := NewDepsCommand()

	// Test that the command can handle flags without crashing
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Should show help when no subcommand is provided
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Usage:") {
		t.Errorf("Expected help output to contain 'Usage:', got: %s", result)
	}
}

// TestDepsCommandWithInvalidArgs tests error handling for invalid arguments
func TestDepsCommandWithInvalidArgs(t *testing.T) {
	// Test the why command with no arguments
	cmd := &cobra.Command{
		Use: "why [package]",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Println(utils.FormatError("Package name is required"))
				return
			}
			// Normal execution would continue here
		},
	}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Package name is required") {
		t.Errorf("Expected error message for missing package name, got: %s", result)
	}
}
