package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// SnippetsAgent handles code snippet creation, management, and organization.
type SnippetsAgent struct {
	BaseAgent
	context *SnippetsContext
}

// SnippetsContext contains information about the snippet management environment.
type SnippetsContext struct {
	// Snippet library management
	LibraryPath   string   `json:"library_path"`   // Path to snippet library
	SnippetFormat string   `json:"snippet_format"` // Snippet format (ultisnips, vscode, etc.)
	Categories    []string `json:"categories"`     // Snippet categories
	SnippetCount  int      `json:"snippet_count"`  // Total number of snippets

	// Editor integration
	EditorType     string `json:"editor_type"`     // Target editor (neovim, vscode, emacs)
	SnippetEngine  string `json:"snippet_engine"`  // Snippet engine being used
	TriggerSystem  string `json:"trigger_system"`  // How snippets are triggered
	ConfigLocation string `json:"config_location"` // Snippet configuration location

	// Language and framework support
	Languages       []string       `json:"languages"`        // Programming languages
	Frameworks      []string       `json:"frameworks"`       // Frameworks and libraries
	SnippetsByLang  map[string]int `json:"snippets_by_lang"` // Snippet count by language
	CustomVariables []string       `json:"custom_variables"` // Custom snippet variables

	// Organization and structure
	NamingConvention string `json:"naming_convention"` // Snippet naming convention
	Organization     string `json:"organization"`      // How snippets are organized
	Documentation    bool   `json:"documentation"`     // Whether snippets are documented
	SearchCapability bool   `json:"search_capability"` // Can search through snippets

	// Dynamic features
	PlaceholderTypes []string `json:"placeholder_types"` // Types of placeholders used
	Transformations  []string `json:"transformations"`   // Text transformations available
	ConditionalLogic bool     `json:"conditional_logic"` // Whether snippets use conditional logic
	ContextAwareness bool     `json:"context_awareness"` // Context-aware snippet behavior

	// Maintenance and quality
	LastUpdated       string            `json:"last_updated"`       // When snippets were last updated
	ConflictIssues    []string          `json:"conflict_issues"`    // Known snippet conflicts
	PerformanceIssues []string          `json:"performance_issues"` // Performance-related issues
	QualityMetrics    map[string]string `json:"quality_metrics"`    // Quality assessment metrics
}

// NewSnippetsAgent creates a new SnippetsAgent.
func NewSnippetsAgent(provider ai.Provider) *SnippetsAgent {
	return &SnippetsAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleSnippets,
		},
		context: &SnippetsContext{
			SnippetFormat:     "ultisnips",
			Categories:        []string{},
			Languages:         []string{},
			Frameworks:        []string{},
			SnippetsByLang:    make(map[string]int),
			CustomVariables:   []string{},
			PlaceholderTypes:  []string{},
			Transformations:   []string{},
			ConflictIssues:    []string{},
			PerformanceIssues: []string{},
			QualityMetrics:    make(map[string]string),
		},
	}
}

// CreateSnippet generates a new code snippet based on requirements.
func (a *SnippetsAgent) CreateSnippet(ctx context.Context, language, pattern, description string) (string, error) {
	prompt := a.buildCreateSnippetPrompt(language, pattern, description)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to create snippet: %w", err)
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to create snippet: %w", err)
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// OrganizeLibrary helps organize and structure snippet libraries.
func (a *SnippetsAgent) OrganizeLibrary(ctx context.Context, currentStructure string, organizationGoals []string) (string, error) {
	a.context.Organization = currentStructure

	prompt := a.buildOrganizeLibraryPrompt(currentStructure, organizationGoals)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to organize library: %w", err)
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to organize library: %w", err)
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// SetupSnippetEngine configures snippet engines for different editors.
func (a *SnippetsAgent) SetupSnippetEngine(ctx context.Context, editorType, engineType string) (string, error) {
	a.context.EditorType = editorType
	a.context.SnippetEngine = engineType

	prompt := a.buildSetupEnginePrompt(editorType, engineType)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to setup snippet engine: %w", err)
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to setup snippet engine: %w", err)
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// OptimizePerformance improves snippet library performance and loading times.
func (a *SnippetsAgent) OptimizePerformance(ctx context.Context, performanceIssues []string) (string, error) {
	a.context.PerformanceIssues = performanceIssues

	prompt := a.buildOptimizePrompt(performanceIssues)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to optimize performance: %w", err)
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to optimize performance: %w", err)
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateCollection creates a comprehensive snippet collection for specific use cases.
func (a *SnippetsAgent) GenerateCollection(ctx context.Context, useCase string, languages []string, requirements []string) (string, error) {
	a.context.Languages = languages

	prompt := a.buildGenerateCollectionPrompt([]string{useCase}, languages, requirements)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to generate collection: %w", err)
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to generate collection: %w", err)
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// MaintainSnippets provides maintenance and update recommendations for snippet libraries.
func (a *SnippetsAgent) MaintainSnippets(ctx context.Context, lastUpdated string, conflictIssues []string) (string, error) {
	a.context.LastUpdated = lastUpdated
	a.context.ConflictIssues = conflictIssues

	prompt := a.buildMaintenancePrompt(lastUpdated, conflictIssues)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		response, err := p.QueryWithContext(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to maintain snippets: %w", err)
		}
		return response, nil
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to maintain snippets: %w", err)
		}
		return response, nil
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GetContext returns the current agent context.
func (a *SnippetsAgent) GetContext() interface{} {
	return a.context
}

// SetContext updates the agent context.
func (a *SnippetsAgent) SetContext(ctx interface{}) error {
	if snippetsCtx, ok := ctx.(*SnippetsContext); ok {
		a.context = snippetsCtx
		return nil
	}
	return fmt.Errorf("invalid context type for SnippetsAgent")
}

// formatContext returns a formatted string representation of the current context.
func (a *SnippetsAgent) formatContext() string {
	return fmt.Sprintf(`Library Path: %s
Snippet Format: %s
Editor: %s
Snippet Engine: %s
Languages: %v
Categories: %v
Total Snippets: %d
Organization: %s
Documentation: %t`,
		a.context.LibraryPath,
		a.context.SnippetFormat,
		a.context.EditorType,
		a.context.SnippetEngine,
		a.context.Languages,
		a.context.Categories,
		a.context.SnippetCount,
		a.context.Organization,
		a.context.Documentation)
}

// --- Prompt builder methods for SnippetsAgent ---

// buildCreateSnippetPrompt builds a prompt for creating a new snippet.
func (a *SnippetsAgent) buildCreateSnippetPrompt(language, pattern, description string) string {
	return fmt.Sprintf(`Create a code snippet in %s that matches the following pattern: "%s". Description: %s. Use the %s snippet format.`,
		language, pattern, description, a.context.SnippetFormat)
}

// buildOrganizeLibraryPrompt builds a prompt for organizing the snippet library.
func (a *SnippetsAgent) buildOrganizeLibraryPrompt(currentStructure string, organizationGoals []string) string {
	return fmt.Sprintf(`Current snippet library structure: %s\nOrganization goals: %v. Suggest a new organization plan for the snippet library.`,
		currentStructure, organizationGoals)
}

// buildSetupEnginePrompt builds a prompt for setting up a snippet engine for an editor.
func (a *SnippetsAgent) buildSetupEnginePrompt(editorType, engineType string) string {
	return fmt.Sprintf(`Setup instructions for integrating the %s snippet engine with %s. Provide configuration steps and best practices.`,
		editorType, engineType)
}

// buildOptimizePrompt builds a prompt for optimizing snippet performance.
func (a *SnippetsAgent) buildOptimizePrompt(performanceIssues []string) string {
	return fmt.Sprintf(`Performance issues: %v. Suggest optimizations to improve snippet loading and usage performance.`,
		performanceIssues)
}

// buildGenerateCollectionPrompt builds a prompt for generating a snippet collection.
func (a *SnippetsAgent) buildGenerateCollectionPrompt(useCases []string, languages []string, requirements []string) string {
	return fmt.Sprintf(`Generate a collection of code snippets for use cases: %v, languages: %v, with these requirements: %v. Use the %s format.`,
		useCases, languages, requirements, a.context.SnippetFormat)
}

// buildMaintenancePrompt builds a prompt for snippet maintenance recommendations.
func (a *SnippetsAgent) buildMaintenancePrompt(lastUpdated string, conflictIssues []string) string {
	return fmt.Sprintf(`The snippet library was last updated on %s. Known conflict issues: %v. Recommend maintenance and update actions.`,
		lastUpdated, conflictIssues)
}
