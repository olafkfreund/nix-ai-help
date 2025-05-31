package devenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// Service handles devenv operations
type Service struct {
	registry  *Registry
	generator *Generator
	logger    *logger.Logger
	ai        ai.AIProvider
}

// NewService creates a new devenv service
func NewService(aiProvider ai.AIProvider, log *logger.Logger) (*Service, error) {
	generator, err := NewGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	service := &Service{
		registry:  GlobalRegistry,
		generator: generator,
		logger:    log,
		ai:        aiProvider,
	}

	// Register built-in templates
	if err := service.registerBuiltinTemplates(); err != nil {
		return nil, fmt.Errorf("failed to register builtin templates: %w", err)
	}

	return service, nil
}

// registerBuiltinTemplates registers all built-in language templates
func (s *Service) registerBuiltinTemplates() error {
	templates := []Template{
		&PythonTemplate{},
		&RustTemplate{},
		&NodejsTemplate{},
		&GolangTemplate{},
	}

	for _, template := range templates {
		if err := s.registry.RegisterIfNotExists(template); err != nil {
			return fmt.Errorf("failed to register template %s: %w", template.Name(), err)
		}
		s.logger.Debug(fmt.Sprintf("Registered devenv template: %s", template.Name()))
	}

	return nil
}

// CreateProject creates a new devenv project from a template
func (s *Service) CreateProject(templateName, projectName, directory string, options map[string]string, services []string) error {
	// Validate input parameters
	if err := s.validateCreateProjectInput(templateName, projectName, directory, services); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	template, err := s.registry.Get(templateName)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Set up directory
	if directory == "" {
		directory = projectName
	}

	absDir, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists and is not empty
	if err := s.validateTargetDirectory(absDir); err != nil {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if devenv.nix already exists
	devenvPath := filepath.Join(absDir, "devenv.nix")
	if utils.IsFile(devenvPath) {
		return fmt.Errorf("devenv.nix already exists in %s", absDir)
	}

	// Build template configuration
	config := TemplateConfig{
		ProjectName: projectName,
		Directory:   absDir,
		Language:    templateName,
		Options:     options,
		Services:    services,
		EnvVars:     make(map[string]string),
	}

	// Validate configuration
	if err := template.Validate(config); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Validate services compatibility
	if err := s.validateServicesCompatibility(template, services); err != nil {
		return fmt.Errorf("services validation failed: %w", err)
	}

	// Generate devenv configuration
	devenvConfig, err := template.Generate(config)
	if err != nil {
		return fmt.Errorf("failed to generate devenv config: %w", err)
	}

	// Generate devenv.nix file
	if err := s.generator.Generate(devenvConfig, devenvPath); err != nil {
		return fmt.Errorf("failed to generate devenv.nix: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Created devenv.nix at %s", devenvPath))

	// Generate additional files based on template
	if err := s.generateAdditionalFiles(template, config, absDir); err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to generate additional files: %v", err))
	}

	return nil
}

// validateCreateProjectInput validates the input parameters for CreateProject
func (s *Service) validateCreateProjectInput(templateName, projectName, directory string, services []string) error {
	// Validate template name
	if templateName == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	// Validate project name
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Validate project name format (no special characters)
	if !isValidProjectName(projectName) {
		return fmt.Errorf("project name '%s' contains invalid characters. Use only letters, numbers, hyphens, and underscores", projectName)
	}

	// Validate services
	for _, service := range services {
		if !isValidServiceName(service) {
			return fmt.Errorf("invalid service name: %s", service)
		}
	}

	return nil
}

// validateTargetDirectory checks if the target directory is suitable for project creation
func (s *Service) validateTargetDirectory(directory string) error {
	if !utils.DirExists(directory) {
		// Directory doesn't exist, which is fine - we'll create it
		return nil
	}

	// Directory exists - check if it's empty or only contains hidden files
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Check for non-hidden files that would conflict
	for _, entry := range entries {
		name := entry.Name()
		// Allow hidden files (starting with .) but warn about important files
		if !strings.HasPrefix(name, ".") {
			if name == "devenv.nix" || name == "devenv.lock" || name == ".devenv" {
				return fmt.Errorf("directory already contains devenv files")
			}
		}
	}

	return nil
}

// validateServicesCompatibility checks if the requested services are compatible with the template
func (s *Service) validateServicesCompatibility(template Template, services []string) error {
	supportedServices := template.SupportedServices()
	if len(supportedServices) == 0 && len(services) > 0 {
		return fmt.Errorf("template %s does not support any services", template.Name())
	}

	// Create a map for quick lookup
	supported := make(map[string]bool)
	for _, service := range supportedServices {
		supported[service] = true
	}

	// Check each requested service
	for _, service := range services {
		if !supported[service] {
			return fmt.Errorf("service '%s' is not supported by template '%s'. Supported services: %v",
				service, template.Name(), supportedServices)
		}
	}

	return nil
}

// isValidProjectName checks if a project name is valid
func isValidProjectName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}

	// Allow letters, numbers, hyphens, and underscores
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	// Cannot start with a number or special character
	first := rune(name[0])
	return (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
}

// isValidServiceName checks if a service name is valid
func isValidServiceName(name string) bool {
	validServices := map[string]bool{
		"postgres": true,
		"redis":    true,
		"mysql":    true,
		"mongodb":  true,
		"nginx":    true,
		"caddy":    true,
		"traefik":  true,
		"mailhog":  true,
		"adminer":  true,
		"minio":    true,
	}

	return validServices[name]
}

// generateAdditionalFiles creates template-specific additional files
func (s *Service) generateAdditionalFiles(template Template, config TemplateConfig, directory string) error {
	templateName := template.Name()

	switch templateName {
	case "python":
		// Create requirements.txt if it doesn't exist
		reqPath := filepath.Join(directory, "requirements.txt")
		if !utils.IsFile(reqPath) {
			if err := os.WriteFile(reqPath, []byte("# Add your Python dependencies here\n"), 0644); err != nil {
				return err
			}
		}

		// Create main.py if it doesn't exist
		mainPath := filepath.Join(directory, "main.py")
		if !utils.IsFile(mainPath) {
			content := `#!/usr/bin/env python3
"""
Main application entry point for ` + config.ProjectName + `
"""

def main():
    print("Hello from ` + config.ProjectName + `!")

if __name__ == "__main__":
    main()
`
			if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
				return err
			}
		}

	case "rust":
		// Initialize Cargo project if Cargo.toml doesn't exist
		cargoPath := filepath.Join(directory, "Cargo.toml")
		if !utils.IsFile(cargoPath) {
			content := `[package]
name = "` + config.ProjectName + `"
version = "0.1.0"
edition = "2021"

[dependencies]
`
			if err := os.WriteFile(cargoPath, []byte(content), 0644); err != nil {
				return err
			}

			// Create src directory and main.rs
			srcDir := filepath.Join(directory, "src")
			if err := os.MkdirAll(srcDir, 0755); err != nil {
				return err
			}

			mainPath := filepath.Join(srcDir, "main.rs")
			if !utils.IsFile(mainPath) {
				content := `fn main() {
    println!("Hello from ` + config.ProjectName + `!");
}
`
				if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
					return err
				}
			}
		}

	case "nodejs":
		// Create package.json if it doesn't exist
		packagePath := filepath.Join(directory, "package.json")
		if !utils.IsFile(packagePath) {
			content := `{
  "name": "` + config.ProjectName + `",
  "version": "1.0.0",
  "description": "",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js",
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "keywords": [],
  "author": "",
  "license": "ISC"
}
`
			if err := os.WriteFile(packagePath, []byte(content), 0644); err != nil {
				return err
			}
		}

		// Create index.js if it doesn't exist
		indexPath := filepath.Join(directory, "index.js")
		if !utils.IsFile(indexPath) {
			content := `console.log('Hello from ` + config.ProjectName + `!');
`
			if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
				return err
			}
		}

	case "golang":
		// Initialize Go module if go.mod doesn't exist
		goModPath := filepath.Join(directory, "go.mod")
		if !utils.IsFile(goModPath) {
			content := `module ` + config.ProjectName + `

go 1.21
`
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				return err
			}
		}

		// Create main.go if it doesn't exist
		mainPath := filepath.Join(directory, "main.go")
		if !utils.IsFile(mainPath) {
			content := `package main

import "fmt"

func main() {
	fmt.Println("Hello from ` + config.ProjectName + `!")
}
`
			if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// ListTemplates returns all available templates with descriptions
func (s *Service) ListTemplates() map[string]string {
	return s.registry.ListWithDescriptions()
}

// GetTemplate returns a specific template
func (s *Service) GetTemplate(name string) (Template, error) {
	return s.registry.Get(name)
}

// SuggestTemplate uses AI to suggest the best template based on user input
func (s *Service) SuggestTemplate(description string) (string, error) {
	if s.ai == nil {
		return "", fmt.Errorf("AI provider not available")
	}

	templates := s.registry.ListWithDescriptions()
	prompt := buildTemplateSuggestionPrompt(description, templates)

	response, err := s.ai.Query(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get AI suggestion: %w", err)
	}

	// Parse the AI response to extract the template name
	// This is a simple implementation - could be made more sophisticated
	for templateName := range templates {
		if contains(response, templateName) {
			return templateName, nil
		}
	}

	return "", fmt.Errorf("no suitable template found for description: %s", description)
}

// buildTemplateSuggestionPrompt creates a prompt for AI template suggestion
func buildTemplateSuggestionPrompt(description string, templates map[string]string) string {
	prompt := fmt.Sprintf(`Based on the following project description, suggest the most appropriate devenv template:

Description: %s

Available templates:
`, description)

	for name, desc := range templates {
		prompt += fmt.Sprintf("- %s: %s\n", name, desc)
	}

	prompt += `
Please respond with just the template name that best matches the project description.
Consider the programming language, framework preferences, and project type mentioned in the description.
`

	return prompt
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
