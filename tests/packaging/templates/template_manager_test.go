package templates_test

import (
	"testing"

	"nix-ai-help/internal/packaging/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateManager(t *testing.T) {
	manager := templates.NewTemplateManager()

	t.Run("Load embedded templates", func(t *testing.T) {
		templateList := manager.ListTemplates()
		assert.NotEmpty(t, templateList, "Should load embedded templates")

		// Check that we have expected templates
		expectedTemplates := []string{"javascript-npm", "typescript-npm", "python-pip", "rust-cargo", "go-modules", "default"}
		for _, expected := range expectedTemplates {
			template, err := manager.GetTemplate(expected, "")
			assert.NoError(t, err, "Should find template: %s", expected)
			assert.NotNil(t, template, "Template should not be nil: %s", expected)
			assert.NotEmpty(t, template.Template, "Template content should not be empty: %s", expected)
		}
	})

	t.Run("Language aliases work", func(t *testing.T) {
		// Test language-only aliases
		jsTemplate, err := manager.GetTemplate("javascript", "")
		require.NoError(t, err)
		assert.Equal(t, "javascript-npm", jsTemplate.Name)

		rustTemplate, err := manager.GetTemplate("rust", "")
		require.NoError(t, err)
		assert.Equal(t, "rust-cargo", rustTemplate.Name)
	})

	t.Run("Apply JavaScript template", func(t *testing.T) {
		template, err := manager.GetTemplate("javascript", "npm")
		require.NoError(t, err)

		context := &templates.TemplateContext{
			ProjectName: "test-app",
			Version:     "1.0.0",
			Description: "A test application",
			Owner:       "testuser",
			Language:    "javascript",
			BuildSystem: "npm",
		}

		result, err := manager.ApplyTemplate(template, context)
		require.NoError(t, err)
		assert.Contains(t, result, "test-app", "Should contain project name")
		assert.Contains(t, result, "1.0.0", "Should contain version")
		assert.Contains(t, result, "testuser", "Should contain owner")
		assert.Contains(t, result, "buildNpmPackage", "Should use npm builder")
	})

	t.Run("Apply Rust template", func(t *testing.T) {
		template, err := manager.GetTemplate("rust", "cargo")
		require.NoError(t, err)

		context := &templates.TemplateContext{
			ProjectName: "rust-cli",
			Version:     "0.2.0",
			Description: "A Rust CLI tool",
			Owner:       "rustdev",
			Language:    "rust",
			BuildSystem: "cargo",
		}

		result, err := manager.ApplyTemplate(template, context)
		require.NoError(t, err)
		assert.Contains(t, result, "rust-cli", "Should contain project name")
		assert.Contains(t, result, "0.2.0", "Should contain version")
		assert.Contains(t, result, "rustdev", "Should contain owner")
		assert.Contains(t, result, "buildRustPackage", "Should use Rust builder")
		assert.Contains(t, result, "cargoHash", "Should have cargo hash")
	})

	t.Run("Validate context requirements", func(t *testing.T) {
		template, err := manager.GetTemplate("javascript", "npm")
		require.NoError(t, err)

		// Context missing required fields
		context := &templates.TemplateContext{
			ProjectName: "", // Missing required field
			Language:    "javascript",
		}

		errors := manager.ValidateContext(template, context)
		assert.NotEmpty(t, errors, "Should have validation errors for missing required fields")
	})

	t.Run("Fall back to default template", func(t *testing.T) {
		template, err := manager.GetTemplate("unknown-language", "unknown-build-system")
		require.NoError(t, err)
		assert.Equal(t, "default", template.Name, "Should fall back to default template")
	})
}

func TestTemplateContext(t *testing.T) {
	t.Run("Create context with all fields", func(t *testing.T) {
		context := &templates.TemplateContext{
			ProjectName:       "my-project",
			Version:           "1.0.0",
			Description:       "My awesome project",
			Homepage:          "https://github.com/user/project",
			License:           "MIT",
			Owner:             "myuser",
			Language:          "go",
			BuildSystem:       "modules",
			BuildInputs:       []string{"pkg-config"},
			NativeBuildInputs: []string{"go"},
			Dependencies:      map[string]string{"lib1": "1.0", "lib2": "2.0"},
			DevDependencies:   map[string]string{"test-lib": "1.0"},
			BuildPhase:        "go build",
			InstallPhase:      "cp binary $out/bin/",
			CheckPhase:        "go test ./...",
			ConfigureFlags:    []string{"--enable-feature"},
			MakeFlags:         []string{"-j4"},
			Custom:            map[string]interface{}{"ExtraFlag": "value"},
		}

		assert.Equal(t, "my-project", context.ProjectName)
		assert.Equal(t, "go", context.Language)
		assert.Equal(t, "myuser", context.Owner)
		assert.Len(t, context.Dependencies, 2)
		assert.Equal(t, "value", context.Custom["ExtraFlag"])
	})
}

func TestEmbeddedTemplates(t *testing.T) {
	manager := templates.NewTemplateManager()

	t.Run("All embedded templates load correctly", func(t *testing.T) {
		templateList := manager.ListTemplates()
		assert.NotEmpty(t, templateList, "Should have embedded templates")

		// Check that specific templates exist and have content
		expectedTemplates := map[string]string{
			"javascript-npm": "buildNpmPackage",
			"typescript-npm": "buildNpmPackage",
			"python-pip":     "buildPythonApplication",
			"rust-cargo":     "buildRustPackage",
			"go-modules":     "buildGoModule",
			"default":        "stdenv.mkDerivation",
		}

		for templateName, expectedBuilder := range expectedTemplates {
			template, err := manager.GetTemplate(templateName, "")
			require.NoError(t, err, "Should find template: %s", templateName)
			assert.Contains(t, template.Template, expectedBuilder, "Template %s should use %s", templateName, expectedBuilder)
			assert.Contains(t, template.Template, "{{.ProjectName}}", "Template should contain project name variable: %s", templateName)
			assert.Contains(t, template.Template, "meta", "Template should contain meta section: %s", templateName)
		}
	})

	t.Run("Template variables are defined", func(t *testing.T) {
		template, err := manager.GetTemplate("javascript-npm", "")
		require.NoError(t, err)

		// Check that the template has defined variables
		assert.NotEmpty(t, template.Variables, "Template should have defined variables")

		// Check for required variables
		requiredVars := []string{"ProjectName", "Version", "Owner"}
		for _, reqVar := range requiredVars {
			found := false
			for _, variable := range template.Variables {
				if variable.Name == reqVar && variable.Required {
					found = true
					break
				}
			}
			assert.True(t, found, "Required variable %s should be defined", reqVar)
		}
	})
}
