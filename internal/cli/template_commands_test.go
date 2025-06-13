package cli

import (
	"testing"

	"nix-ai-help/pkg/logger"
)

// TestTemplateManagerCreation tests template manager creation
func TestTemplateManagerCreation(t *testing.T) {
	log := logger.NewLoggerWithLevel("info")
	tm := NewTemplateManager("", log)

	if tm == nil {
		t.Error("Expected TemplateManager to be created, got nil")
	}
}

// TestLoadBuiltinTemplates tests loading builtin templates
func TestLoadBuiltinTemplates(t *testing.T) {
	log := logger.NewLoggerWithLevel("info")
	tm := NewTemplateManager("", log)

	templates := tm.LoadBuiltinTemplates()
	if len(templates) == 0 {
		t.Error("Expected builtin templates to be loaded, got empty list")
	}

	// Check that we have some expected templates
	foundTemplate := false
	for _, template := range templates {
		if template.Name == "desktop-minimal" {
			foundTemplate = true
			if template.Category != "Desktop" {
				t.Errorf("Expected desktop-minimal template to have category 'Desktop', got '%s'", template.Category)
			}
			if template.Content == "" {
				t.Error("Expected desktop-minimal template to have content")
			}
			break
		}
	}

	if !foundTemplate {
		t.Error("Expected to find 'desktop-minimal' template in builtin templates")
	}
}

// TestGetTemplate tests template retrieval
func TestGetTemplate(t *testing.T) {
	log := logger.NewLoggerWithLevel("info")
	tm := NewTemplateManager("", log)

	// Test getting a builtin template
	template, err := tm.GetTemplate("desktop-minimal")
	if err != nil {
		t.Errorf("Expected to get template, got error: %v", err)
	}

	if template == nil {
		t.Error("Expected template to be returned, got nil")
	}

	if template != nil {
		if template.Name != "desktop-minimal" {
			t.Errorf("Expected template name 'desktop-minimal', got '%s'", template.Name)
		}
	}

	// Test getting a non-existent template
	_, err = tm.GetTemplate("non-existent-template")
	if err == nil {
		t.Error("Expected error when getting non-existent template, got nil")
	}
}

// TestSearchTemplates tests template searching
func TestSearchTemplates(t *testing.T) {
	log := logger.NewLoggerWithLevel("info")
	tm := NewTemplateManager("", log)

	// Search for desktop templates
	results := tm.SearchTemplates("desktop")
	if len(results) == 0 {
		t.Error("Expected to find desktop templates, got empty results")
	}

	// Check that results contain desktop-related templates
	foundDesktop := false
	for _, template := range results {
		if template.Category == "Desktop" ||
			contains(template.Tags, "desktop") ||
			contains([]string{template.Name}, "desktop") {
			foundDesktop = true
			break
		}
	}

	if !foundDesktop {
		t.Error("Expected search results to contain desktop-related templates")
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
