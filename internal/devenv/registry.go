package devenv

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Registry manages available devenv templates
type Registry struct {
	templates map[string]Template
	mutex     sync.RWMutex
}

// NewRegistry creates a new template registry
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]Template),
	}
}

// Register adds a template to the registry
func (r *Registry) Register(template Template) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := template.Name()
	if _, exists := r.templates[name]; exists {
		return fmt.Errorf("template '%s' is already registered", name)
	}

	r.templates[name] = template
	return nil
}

// Clear removes all templates from the registry
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.templates = make(map[string]Template)
}

// RegisterIfNotExists adds a template to the registry only if it doesn't exist
func (r *Registry) RegisterIfNotExists(template Template) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := template.Name()
	if _, exists := r.templates[name]; exists {
		return nil // Already exists, no error
	}

	r.templates[name] = template
	return nil
}

// Get retrieves a template by name
func (r *Registry) Get(name string) (Template, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	template, exists := r.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}

// List returns all registered template names
func (r *Registry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.templates))
	for name := range r.templates {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// ListWithDescriptions returns all templates with their descriptions
func (r *Registry) ListWithDescriptions() map[string]string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[string]string)
	for name, template := range r.templates {
		result[name] = template.Description()
	}

	return result
}

// Search finds templates by keyword in name or description
func (r *Registry) Search(keyword string) []Template {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	keyword = strings.ToLower(keyword)
	var matches []Template

	for _, template := range r.templates {
		name := strings.ToLower(template.Name())
		desc := strings.ToLower(template.Description())

		if strings.Contains(name, keyword) || strings.Contains(desc, keyword) {
			matches = append(matches, template)
		}
	}

	return matches
}

// GetTemplatesByLanguage returns templates that support a specific language
func (r *Registry) GetTemplatesByLanguage(language string) []Template {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	language = strings.ToLower(language)
	var matches []Template

	for _, template := range r.templates {
		templateName := strings.ToLower(template.Name())
		if strings.Contains(templateName, language) {
			matches = append(matches, template)
		}
	}

	return matches
}

// Global registry instance
var GlobalRegistry = NewRegistry()
