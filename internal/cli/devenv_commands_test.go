package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nix-ai-help/internal/devenv"
	"nix-ai-help/pkg/logger"

	"github.com/spf13/cobra"
)

// MockAIProvider for testing CLI commands
type MockDevenvAIProvider struct {
	responses map[string]string
}

func NewMockDevenvAIProvider() *MockDevenvAIProvider {
	return &MockDevenvAIProvider{
		responses: make(map[string]string),
	}
}

func (m *MockDevenvAIProvider) SetResponse(prompt string, response string) {
	m.responses[prompt] = response
}

func (m *MockDevenvAIProvider) Query(prompt string) (string, error) {
	if response, exists := m.responses[prompt]; exists {
		return response, nil
	}
	return "python", nil // Default response
}

func (m *MockDevenvAIProvider) GenerateResponse(prompt string) (string, error) {
	return m.Query(prompt)
}

func TestDevenvListCommand(t *testing.T) {
	// Capture stdout
	var buf bytes.Buffer

	// Create a new command for testing
	cmd := &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration (using test config)
			// Create logger
			log := logger.NewLoggerWithLevel("debug")

			// Initialize AI provider
			aiProvider := NewMockDevenvAIProvider()

			// Create devenv service
			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				t.Fatalf("Error creating devenv service: %v", err)
			}

			templates := service.ListTemplates()
			if len(templates) == 0 {
				buf.WriteString("No templates available")
				return
			}

			for name, description := range templates {
				buf.WriteString(name + ": " + description + "\n")
			}
		},
	}

	// Execute command
	cmd.SetOut(&buf)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Check that output contains expected templates
	expectedTemplates := []string{"python", "rust", "nodejs", "golang"}
	for _, template := range expectedTemplates {
		if !strings.Contains(output, template) {
			t.Errorf("Expected output to contain template: %s", template)
		}
	}
}

func TestDevenvCreateCommand_ValidInput(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")

	// Create a new command for testing
	cmd := &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				t.Fatal("Expected template name argument")
			}

			templateName := args[0]
			projectName := "test-project"
			if len(args) > 1 {
				projectName = args[1]
			}

			// Create logger
			log := logger.NewLoggerWithLevel("debug")

			// Initialize AI provider
			aiProvider := NewMockDevenvAIProvider()

			// Create devenv service
			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				t.Fatalf("Error creating devenv service: %v", err)
			}

			// Create the project
			err = service.CreateProject(templateName, projectName, projectDir,
				map[string]string{}, []string{})
			if err != nil {
				t.Fatalf("Error creating project: %v", err)
			}
		},
	}

	// Execute command with arguments
	cmd.SetArgs([]string{"python"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Check that devenv.nix was created
	devenvPath := filepath.Join(projectDir, "devenv.nix")
	if !fileExists(devenvPath) {
		t.Errorf("Expected devenv.nix to be created at %s", devenvPath)
	}
}

func TestDevenvCreateCommand_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")

	cmd := &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				return
			}

			templateName := args[0]
			projectName := "test-project"

			log := logger.NewLoggerWithLevel("debug")
			aiProvider := NewMockDevenvAIProvider()

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				t.Fatalf("Error creating devenv service: %v", err)
			}

			err = service.CreateProject(templateName, projectName, projectDir,
				map[string]string{}, []string{})
			if err == nil {
				t.Error("Expected error for invalid template")
			}
		},
	}

	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()
	// We expect this to not return an error at the command level
	// The error should be handled within the command function
	if err != nil && !strings.Contains(err.Error(), "template not found") {
		t.Fatalf("Unexpected command execution error: %v", err)
	}
}

func TestDevenvSuggestCommand(t *testing.T) {
	var buf bytes.Buffer

	cmd := &cobra.Command{
		Use: "suggest",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				t.Fatal("Expected description argument")
			}

			description := strings.Join(args, " ")

			log := logger.NewLoggerWithLevel("debug")
			aiProvider := NewMockDevenvAIProvider()
			aiProvider.SetResponse("", "python") // Mock response

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				t.Fatalf("Error creating devenv service: %v", err)
			}

			suggestion, err := service.SuggestTemplate(description)
			if err != nil {
				t.Fatalf("Error getting suggestion: %v", err)
			}

			buf.WriteString("Suggested template: " + suggestion)
		},
	}

	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"web", "application", "with", "database"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "python") {
		t.Errorf("Expected output to contain 'python', got: %s", output)
	}
}

func TestDevenvCommandValidation(t *testing.T) {
	tests := []struct {
		name           string
		templateName   string
		projectName    string
		services       []string
		expectError    bool
		errorSubstring string
	}{
		{
			name:         "Valid input",
			templateName: "python",
			projectName:  "my-project",
			services:     []string{"postgres"},
			expectError:  false,
		},
		{
			name:           "Empty template name",
			templateName:   "",
			projectName:    "my-project",
			services:       []string{},
			expectError:    true,
			errorSubstring: "template name cannot be empty",
		},
		{
			name:           "Invalid project name",
			templateName:   "python",
			projectName:    "123invalid",
			services:       []string{},
			expectError:    true,
			errorSubstring: "invalid characters",
		},
		{
			name:           "Invalid service",
			templateName:   "python",
			projectName:    "my-project",
			services:       []string{"invalidservice"},
			expectError:    true,
			errorSubstring: "invalid service name",
		},
		{
			name:           "Unsupported service for template",
			templateName:   "rust",
			projectName:    "my-project",
			services:       []string{"mongodb"}, // rust template doesn't support mongodb
			expectError:    true,
			errorSubstring: "not supported by template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, tt.projectName)

			log := logger.NewLoggerWithLevel("debug")
			aiProvider := NewMockDevenvAIProvider()

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				t.Fatalf("Error creating devenv service: %v", err)
			}

			err = service.CreateProject(tt.templateName, tt.projectName, projectDir,
				map[string]string{}, tt.services)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestDevenvCommandFlags(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "flag-test")

	// Test that flags are properly parsed and used
	cmd := &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			templateName := "python"
			projectName := "flag-test"

			// Get flags
			servicesFlag, _ := cmd.Flags().GetString("services")
			var services []string
			if servicesFlag != "" {
				services = strings.Split(servicesFlag, ",")
				for i, service := range services {
					services[i] = strings.TrimSpace(service)
				}
			}

			log := logger.NewLoggerWithLevel("debug")
			aiProvider := NewMockDevenvAIProvider()

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				t.Fatalf("Error creating devenv service: %v", err)
			}

			err = service.CreateProject(templateName, projectName, projectDir,
				map[string]string{}, services)
			if err != nil {
				t.Fatalf("Error creating project: %v", err)
			}
		},
	}

	// Add flags
	cmd.Flags().String("services", "", "Comma-separated list of services")

	// Execute with flags
	cmd.SetArgs([]string{"--services", "postgres,redis"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Check that devenv.nix was created
	devenvPath := filepath.Join(projectDir, "devenv.nix")
	if !fileExists(devenvPath) {
		t.Errorf("Expected devenv.nix to be created at %s", devenvPath)
	}
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Test interactive mode integration
func TestInteractiveDevenvCommands(t *testing.T) {
	// This would test the interactive mode handlers for devenv commands
	// Since the interactive mode uses the same underlying commands,
	// we mainly need to test that command parsing works correctly

	tests := []struct {
		input       string
		expectError bool
		description string
	}{
		{"devenv list", false, "List command"},
		{"devenv create python myproject", false, "Create command"},
		{"devenv suggest web app", false, "Suggest command"},
		{"devenv", false, "Base command shows help"},
		{"devenv invalid", true, "Invalid subcommand"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Parse the command input
			parts := strings.Fields(tt.input)
			if len(parts) == 0 {
				t.Fatal("Empty command")
			}

			if parts[0] != "devenv" {
				t.Fatal("Not a devenv command")
			}

			// Test command parsing logic
			if len(parts) == 1 {
				// Base devenv command - should show help
				return
			}

			subcommand := parts[1]
			validSubcommands := []string{"list", "create", "suggest"}

			valid := false
			for _, validCmd := range validSubcommands {
				if subcommand == validCmd {
					valid = true
					break
				}
			}

			if tt.expectError && valid {
				t.Errorf("Expected invalid subcommand but '%s' is valid", subcommand)
			} else if !tt.expectError && !valid {
				t.Errorf("Expected valid subcommand but '%s' is invalid", subcommand)
			}
		})
	}
}
