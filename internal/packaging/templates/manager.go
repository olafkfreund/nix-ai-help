package templates

import (
	"fmt"
	"strings"
	"text/template"
)

// TemplateVariable represents a variable that can be substituted in templates
type TemplateVariable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "bool", "list", "map"
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required"`
}

// DerivationTemplate represents a Nix derivation template
type DerivationTemplate struct {
	Name        string             `json:"name"`
	Language    string             `json:"language"`
	BuildSystem string             `json:"build_system"`
	Template    string             `json:"template"`
	Variables   []TemplateVariable `json:"variables"`
	Description string             `json:"description"`
	Examples    []string           `json:"examples,omitempty"`
}

// TemplateContext contains variables for template substitution
type TemplateContext struct {
	// Project information
	ProjectName string `json:"project_name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	License     string `json:"license"`
	Owner       string `json:"owner"` // GitHub owner/organization

	// Build information
	Language          string   `json:"language"`
	BuildSystem       string   `json:"build_system"`
	BuildInputs       []string `json:"build_inputs"`
	NativeBuildInputs []string `json:"native_build_inputs"`

	// Dependencies
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"dev_dependencies"`

	// Build configuration
	BuildPhase     string   `json:"build_phase,omitempty"`
	InstallPhase   string   `json:"install_phase,omitempty"`
	CheckPhase     string   `json:"check_phase,omitempty"`
	ConfigureFlags []string `json:"configure_flags,omitempty"`
	MakeFlags      []string `json:"make_flags,omitempty"`

	// Custom variables
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// TemplateManager manages derivation templates
type TemplateManager struct {
	templates map[string]*DerivationTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*DerivationTemplate),
	}

	// Load built-in templates
	tm.loadBuiltinTemplates()

	return tm
}

// GetTemplate retrieves a template by language and build system
func (tm *TemplateManager) GetTemplate(language, buildSystem string) (*DerivationTemplate, error) {
	// Try exact match first
	key := fmt.Sprintf("%s-%s", language, buildSystem)
	if template, exists := tm.templates[key]; exists {
		return template, nil
	}

	// Try language-only match
	if template, exists := tm.templates[language]; exists {
		return template, nil
	}

	// Try build system only match
	if template, exists := tm.templates[buildSystem]; exists {
		return template, nil
	}

	// Return default template
	if template, exists := tm.templates["default"]; exists {
		return template, nil
	}

	return nil, fmt.Errorf("no template found for language '%s' and build system '%s'", language, buildSystem)
}

// ApplyTemplate applies a template with the given context
func (tm *TemplateManager) ApplyTemplate(derivTemplate *DerivationTemplate, context *TemplateContext) (string, error) {
	// Parse template with custom functions
	tmpl, err := template.New(derivTemplate.Name).Funcs(templateFuncs).Parse(derivTemplate.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var result strings.Builder
	err = tmpl.Execute(&result, context)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() []*DerivationTemplate {
	templates := make([]*DerivationTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// ValidateContext validates that all required variables are present in context
func (tm *TemplateManager) ValidateContext(template *DerivationTemplate, context *TemplateContext) []string {
	var errors []string

	for _, variable := range template.Variables {
		if !variable.Required {
			continue
		}

		// Check if variable exists in context
		if !tm.hasVariable(context, variable.Name) {
			errors = append(errors, fmt.Sprintf("required variable '%s' is missing", variable.Name))
		}
	}

	return errors
}

// hasVariable checks if a variable exists in the context
func (tm *TemplateManager) hasVariable(context *TemplateContext, name string) bool {
	switch name {
	case "ProjectName":
		return context.ProjectName != ""
	case "Version":
		return context.Version != ""
	case "Description":
		return context.Description != ""
	case "Owner":
		return context.Owner != ""
	case "Language":
		return context.Language != ""
	case "BuildSystem":
		return context.BuildSystem != ""
	default:
		// Check custom variables
		if context.Custom != nil {
			_, exists := context.Custom[name]
			return exists
		}
		return false
	}
}

// loadBuiltinTemplates loads the built-in templates from embedded content
func (tm *TemplateManager) loadBuiltinTemplates() {
	embeddedTemplates := getEmbeddedTemplates()

	// Create DerivationTemplate objects from embedded content
	for key, content := range embeddedTemplates {
		parts := strings.Split(key, "-")
		language := parts[0]
		buildSystem := ""
		if len(parts) > 1 {
			buildSystem = parts[1]
		}

		template := &DerivationTemplate{
			Name:        key,
			Language:    language,
			BuildSystem: buildSystem,
			Template:    content,
			Description: fmt.Sprintf("Template for %s projects", key),
			Variables:   getDefaultVariables(),
		}

		tm.templates[key] = template
	}

	// Add language-only aliases
	tm.templates["javascript"] = tm.templates["javascript-npm"]
	tm.templates["typescript"] = tm.templates["typescript-npm"]
	tm.templates["python"] = tm.templates["python-pip"]
	tm.templates["rust"] = tm.templates["rust-cargo"]
	tm.templates["go"] = tm.templates["go-modules"]
	tm.templates["c"] = tm.templates["c-cmake"]
	tm.templates["cpp"] = tm.templates["cpp-cmake"]
}

// getDefaultVariables returns the default set of variables for templates
func getDefaultVariables() []TemplateVariable {
	return []TemplateVariable{
		{Name: "ProjectName", Type: "string", Description: "Name of the project", Required: true},
		{Name: "Version", Type: "string", Description: "Version of the project", Required: true},
		{Name: "Owner", Type: "string", Description: "GitHub owner/organization", Required: true},
		{Name: "Description", Type: "string", Description: "Project description", Required: false},
		{Name: "Homepage", Type: "string", Description: "Project homepage URL", Required: false},
		{Name: "License", Type: "string", Description: "Project license", Required: false},
		{Name: "Language", Type: "string", Description: "Primary programming language", Required: true},
		{Name: "BuildSystem", Type: "string", Description: "Build system used", Required: true},
		{Name: "BuildInputs", Type: "list", Description: "Build dependencies", Required: false},
		{Name: "NativeBuildInputs", Type: "list", Description: "Native build dependencies", Required: false},
		{Name: "Dependencies", Type: "map", Description: "Runtime dependencies", Required: false},
		{Name: "DevDependencies", Type: "map", Description: "Development dependencies", Required: false},
		{Name: "BuildPhase", Type: "string", Description: "Custom build phase", Required: false},
		{Name: "InstallPhase", Type: "string", Description: "Custom install phase", Required: false},
		{Name: "CheckPhase", Type: "string", Description: "Custom check phase", Required: false},
		{Name: "ConfigureFlags", Type: "list", Description: "Configure flags", Required: false},
		{Name: "MakeFlags", Type: "list", Description: "Make flags", Required: false},
	}
}
