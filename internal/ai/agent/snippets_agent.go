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

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to create snippet: %w", err)
	}

	return response, nil
}

// OrganizeLibrary helps organize and structure snippet libraries.
func (a *SnippetsAgent) OrganizeLibrary(ctx context.Context, currentStructure string, organizationGoals []string) (string, error) {
	a.context.Organization = currentStructure

	prompt := a.buildOrganizeLibraryPrompt(currentStructure, organizationGoals)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to organize library: %w", err)
	}

	return response, nil
}

// SetupSnippetEngine configures snippet engines for different editors.
func (a *SnippetsAgent) SetupSnippetEngine(ctx context.Context, editorType, engineType string) (string, error) {
	a.context.EditorType = editorType
	a.context.SnippetEngine = engineType

	prompt := a.buildSetupEnginePrompt(editorType, engineType)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to setup snippet engine: %w", err)
	}

	return response, nil
}

// OptimizePerformance improves snippet library performance and loading times.
func (a *SnippetsAgent) OptimizePerformance(ctx context.Context, performanceIssues []string) (string, error) {
	a.context.PerformanceIssues = performanceIssues

	prompt := a.buildOptimizePrompt(performanceIssues)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to optimize performance: %w", err)
	}

	return response, nil
}

// GenerateCollection creates a comprehensive snippet collection for specific use cases.
func (a *SnippetsAgent) GenerateCollection(ctx context.Context, useCase string, languages []string, requirements []string) (string, error) {
	a.context.Languages = languages

	prompt := a.buildGenerateCollectionPrompt([]string{useCase}, languages, requirements)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate collection: %w", err)
	}

	return response, nil
}

// MaintainSnippets provides maintenance and update recommendations for snippet libraries.
func (a *SnippetsAgent) MaintainSnippets(ctx context.Context, lastUpdated string, conflictIssues []string) (string, error) {
	a.context.LastUpdated = lastUpdated
	a.context.ConflictIssues = conflictIssues

	prompt := a.buildMaintenancePrompt(lastUpdated, conflictIssues)

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to maintain snippets: %w", err)
	}

	return response, nil
}

// Helper methods for building prompts

func (a *SnippetsAgent) buildCreateSnippetPrompt(language, pattern, description string) string {
	return fmt.Sprintf(`Create a code snippet with the following specifications:

Language: %s
Pattern/Template: %s
Description: %s
Current Context: %s

Please provide:
1. Complete snippet code with proper syntax
2. Placeholder definitions and transformations
3. Trigger word and expansion logic
4. Variable substitutions and dynamic content
5. Documentation and usage examples
6. Integration instructions for the target editor

Focus on creating a useful, well-structured snippet that follows best practices and conventions.`,
		language, pattern, description, a.formatContext())
}

func (a *SnippetsAgent) buildOrganizeLibraryPrompt(currentStructure string, goals []string) string {
	return fmt.Sprintf(`Help organize a snippet library with the following requirements:

Current Structure: %s
Organization Goals: %v
Current Context: %s

Please provide:
1. Recommended directory and file structure
2. Naming conventions for snippets and categories
3. Categorization and tagging strategies
4. Documentation and metadata standards
5. Search and discovery improvements
6. Maintenance and update procedures

Focus on creating a maintainable, discoverable library structure.`,
		currentStructure, goals, a.formatContext())
}

func (a *SnippetsAgent) buildSetupEnginePrompt(editorType, engineType string) string {
	return fmt.Sprintf(`Setup snippet engine configuration for:

Editor: %s
Snippet Engine: %s
Current Context: %s

Please provide:
1. Installation and configuration instructions
2. Engine-specific configuration files and settings
3. Integration with editor features and workflows
4. Trigger and expansion configuration
5. Custom variable and function setup
6. Troubleshooting and debugging guidance

Ensure optimal integration and performance with the target editor.`,
		editorType, engineType, a.formatContext())
}

func (a *SnippetsAgent) buildOptimizePrompt(performanceIssues []string) string {
	return fmt.Sprintf(`Optimize snippet library performance addressing these issues:

Performance Issues: %v
Current Context: %s

Please provide:
1. Performance analysis and bottleneck identification
2. Loading time optimization strategies
3. Memory usage reduction techniques
4. Snippet organization for faster access
5. Caching and preloading optimizations
6. Measurement and monitoring recommendations

Focus on improving responsiveness and reducing resource usage.`,
		performanceIssues, a.formatContext())
}

func (a *SnippetsAgent) buildGenerateCollectionPrompt(useCase, languages []string, requirements []string) string {
	return fmt.Sprintf(`Generate a comprehensive snippet collection for:

Use Case: %s
Languages: %v
Requirements: %v
Current Context: %s

Please provide:
1. Complete snippet collection covering common patterns
2. Language-specific snippets and idioms
3. Framework and library specific templates
4. Workflow-optimized snippet sequences
5. Documentation and usage guides
6. Installation and setup instructions

Create a production-ready collection that enhances productivity.`,
		useCase, languages, requirements, a.formatContext())
}

func (a *SnippetsAgent) buildMaintenancePrompt(lastUpdated string, conflictIssues []string) string {
	return fmt.Sprintf(`Provide maintenance recommendations for snippet library:

Last Updated: %s
Conflict Issues: %v
Current Context: %s

Please provide:
1. Update and maintenance schedule recommendations
2. Conflict resolution and prevention strategies
3. Quality assessment and improvement suggestions
4. Deprecated snippet identification and removal
5. Performance monitoring and optimization
6. Backup and versioning strategies

Focus on maintaining library quality and reliability over time.`,
		lastUpdated, conflictIssues, a.formatContext())
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
