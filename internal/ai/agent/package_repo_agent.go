package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// PackageRepoAgent handles Git repository analysis and Nix derivation generation
type PackageRepoAgent struct {
	BaseAgent
}

// PackageRepoContext contains repository and packaging context information
type PackageRepoContext struct {
	RepositoryURL    string            `json:"repository_url,omitempty"`
	RepositoryPath   string            `json:"repository_path,omitempty"`
	ProjectLanguage  string            `json:"project_language,omitempty"`
	BuildSystem      string            `json:"build_system,omitempty"` // cmake, cargo, npm, etc.
	Dependencies     []string          `json:"dependencies,omitempty"`
	PackageManagers  []string          `json:"package_managers,omitempty"` // cargo.lock, package.json, etc.
	LicenseInfo      string            `json:"license_info,omitempty"`
	ProjectMetadata  map[string]string `json:"project_metadata,omitempty"`
	SourceFiles      []string          `json:"source_files,omitempty"`
	ConfigFiles      []string          `json:"config_files,omitempty"`
	BuildScripts     []string          `json:"build_scripts,omitempty"`
	TestCommands     []string          `json:"test_commands,omitempty"`
	Documentation    string            `json:"documentation,omitempty"`
	ExistingNix      string            `json:"existing_nix,omitempty"` // existing nix files
	PackageVersion   string            `json:"package_version,omitempty"`
	ArchitectureReqs []string          `json:"architecture_reqs,omitempty"`
}

// NewPackageRepoAgent creates a new PackageRepoAgent with the PackageRepo role
func NewPackageRepoAgent(provider ai.Provider) *PackageRepoAgent {
	agent := &PackageRepoAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RolePackageRepo,
		},
	}
	return agent
}

// Query handles repository analysis and packaging questions
func (a *PackageRepoAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt, ok := roles.RolePromptTemplate[a.role]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", a.role)
	}

	// Build context-aware prompt
	fullPrompt := a.buildContextualPrompt(prompt, question)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		return p.QueryWithContext(ctx, fullPrompt)
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		return p.Query(fullPrompt)
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateResponse handles packaging response generation
func (a *PackageRepoAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add packaging-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// buildContextualPrompt constructs a context-aware prompt for packaging operations
func (a *PackageRepoAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add packaging context if available
	if a.contextData != nil {
		if repoCtx, ok := a.contextData.(*PackageRepoContext); ok {
			contextStr := a.formatPackageRepoContext(repoCtx)
			if contextStr != "" {
				promptParts = append(promptParts, "Repository Context:")
				promptParts = append(promptParts, contextStr)
			}
		}
	}

	// Add user input
	promptParts = append(promptParts, "Packaging Request:")
	promptParts = append(promptParts, userInput)

	return strings.Join(promptParts, "\n\n")
}

// formatPackageRepoContext formats PackageRepoContext into a readable string
func (a *PackageRepoAgent) formatPackageRepoContext(ctx *PackageRepoContext) string {
	var parts []string

	if ctx.RepositoryURL != "" {
		parts = append(parts, fmt.Sprintf("Repository URL: %s", ctx.RepositoryURL))
	}

	if ctx.RepositoryPath != "" {
		parts = append(parts, fmt.Sprintf("Repository Path: %s", ctx.RepositoryPath))
	}

	if ctx.ProjectLanguage != "" {
		parts = append(parts, fmt.Sprintf("Language: %s", ctx.ProjectLanguage))
	}

	if ctx.BuildSystem != "" {
		parts = append(parts, fmt.Sprintf("Build System: %s", ctx.BuildSystem))
	}

	if ctx.PackageVersion != "" {
		parts = append(parts, fmt.Sprintf("Version: %s", ctx.PackageVersion))
	}

	if ctx.LicenseInfo != "" {
		parts = append(parts, fmt.Sprintf("License: %s", ctx.LicenseInfo))
	}

	if len(ctx.PackageManagers) > 0 {
		parts = append(parts, fmt.Sprintf("Package Managers: %s", strings.Join(ctx.PackageManagers, ", ")))
	}

	if len(ctx.Dependencies) > 0 {
		parts = append(parts, fmt.Sprintf("Dependencies: %s", strings.Join(ctx.Dependencies, ", ")))
	}

	if len(ctx.ArchitectureReqs) > 0 {
		parts = append(parts, fmt.Sprintf("Architecture Requirements: %s", strings.Join(ctx.ArchitectureReqs, ", ")))
	}

	if len(ctx.SourceFiles) > 0 {
		parts = append(parts, fmt.Sprintf("Source Files: %s", strings.Join(ctx.SourceFiles, ", ")))
	}

	if len(ctx.ConfigFiles) > 0 {
		parts = append(parts, fmt.Sprintf("Config Files: %s", strings.Join(ctx.ConfigFiles, ", ")))
	}

	if len(ctx.BuildScripts) > 0 {
		parts = append(parts, fmt.Sprintf("Build Scripts: %s", strings.Join(ctx.BuildScripts, ", ")))
	}

	if len(ctx.TestCommands) > 0 {
		parts = append(parts, fmt.Sprintf("Test Commands: %s", strings.Join(ctx.TestCommands, ", ")))
	}

	if len(ctx.ProjectMetadata) > 0 {
		metadataStr := make([]string, 0, len(ctx.ProjectMetadata))
		for key, value := range ctx.ProjectMetadata {
			metadataStr = append(metadataStr, fmt.Sprintf("%s: %s", key, value))
		}
		parts = append(parts, fmt.Sprintf("Metadata:\n%s", strings.Join(metadataStr, "\n")))
	}

	if ctx.Documentation != "" {
		parts = append(parts, fmt.Sprintf("Documentation:\n%s", ctx.Documentation))
	}

	if ctx.ExistingNix != "" {
		parts = append(parts, fmt.Sprintf("Existing Nix Files:\n%s", ctx.ExistingNix))
	}

	return strings.Join(parts, "\n")
}

// SetPackageRepoContext is a convenience method to set PackageRepoContext
func (a *PackageRepoAgent) SetPackageRepoContext(ctx *PackageRepoContext) {
	a.SetContext(ctx)
}
