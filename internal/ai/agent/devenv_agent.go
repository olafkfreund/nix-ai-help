package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// DevenvAgent assists with development environment setup and management.
type DevenvAgent struct {
	BaseAgent
}

// DevenvContext provides context for development environment operations.
type DevenvContext struct {
	ProjectType       string            // "web", "rust", "go", "python", "nodejs", etc.
	Languages         []string          // Programming languages needed
	Tools             []string          // Development tools required
	Services          []string          // Services needed (databases, redis, etc.)
	Frameworks        []string          // Frameworks being used
	Dependencies      map[string]string // Package dependencies with versions
	Environment       string            // "development", "testing", "ci", etc.
	NixShell          bool              // Whether using nix-shell
	Flakes            bool              // Whether using flakes
	Direnv            bool              // Whether using direnv
	ProjectRoot       string            // Project root directory
	BuildSystem       string            // "cargo", "npm", "make", "cmake", etc.
	DevServices       []string          // Development services (hot reload, etc.)
	ContainerNeeds    bool              // Whether containers are needed
	EditorConfig      string            // Editor/IDE configuration needs
	TestingFramework  string            // Testing tools needed
	LintingTools      []string          // Linting and formatting tools
	DebugTools        []string          // Debugging tools needed
	Documentation     []string          // Documentation generation tools
}

// NewDevenvAgent creates a new DevenvAgent with the specified provider.
func NewDevenvAgent(provider ai.Provider) *DevenvAgent {
	agent := &DevenvAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleDevenv,
		},
	}
	return agent
}

// Query provides development environment guidance using the provider's Query method.
func (a *DevenvAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build devenv-specific prompt with context
	prompt := a.buildDevenvPrompt(question, a.getDevenvContextFromData())

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", err
	}

	return a.formatDevenvResponse(response), nil
}

// GenerateResponse provides detailed development environment setup using the provider's GenerateResponse method.
func (a *DevenvAgent) GenerateResponse(ctx context.Context, request string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build comprehensive devenv prompt
	prompt := a.buildDevenvPrompt(request, a.getDevenvContextFromData())

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", err
	}

	return a.formatDevenvResponse(response), nil
}

// SetDevenvContext sets development environment-specific context.
func (a *DevenvAgent) SetDevenvContext(context *DevenvContext) {
	a.contextData = context
}

// GetDevenvContext returns the current development environment context.
func (a *DevenvAgent) GetDevenvContext() *DevenvContext {
	if ctx, ok := a.contextData.(*DevenvContext); ok {
		return ctx
	}
	return &DevenvContext{}
}

// buildDevenvPrompt constructs a development environment-specific prompt.
func (a *DevenvAgent) buildDevenvPrompt(question string, context *DevenvContext) string {
	var prompt strings.Builder

	// Get role-specific prompt template
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	// Add development environment context
	prompt.WriteString("Development Environment Context:\n")
	
	if context.ProjectType != "" {
		prompt.WriteString(fmt.Sprintf("- Project Type: %s\n", context.ProjectType))
	}
	if len(context.Languages) > 0 {
		prompt.WriteString(fmt.Sprintf("- Languages: %s\n", strings.Join(context.Languages, ", ")))
	}
	if len(context.Tools) > 0 {
		prompt.WriteString(fmt.Sprintf("- Tools: %s\n", strings.Join(context.Tools, ", ")))
	}
	if len(context.Services) > 0 {
		prompt.WriteString(fmt.Sprintf("- Services: %s\n", strings.Join(context.Services, ", ")))
	}
	if len(context.Frameworks) > 0 {
		prompt.WriteString(fmt.Sprintf("- Frameworks: %s\n", strings.Join(context.Frameworks, ", ")))
	}
	if len(context.Dependencies) > 0 {
		var deps []string
		for pkg, version := range context.Dependencies {
			if version != "" {
				deps = append(deps, fmt.Sprintf("%s@%s", pkg, version))
			} else {
				deps = append(deps, pkg)
			}
		}
		prompt.WriteString(fmt.Sprintf("- Dependencies: %s\n", strings.Join(deps, ", ")))
	}
	if context.Environment != "" {
		prompt.WriteString(fmt.Sprintf("- Environment: %s\n", context.Environment))
	}
	if context.NixShell {
		prompt.WriteString("- Using nix-shell: Yes\n")
	}
	if context.Flakes {
		prompt.WriteString("- Using Flakes: Yes\n")
	}
	if context.Direnv {
		prompt.WriteString("- Using direnv: Yes\n")
	}
	if context.ProjectRoot != "" {
		prompt.WriteString(fmt.Sprintf("- Project Root: %s\n", context.ProjectRoot))
	}
	if context.BuildSystem != "" {
		prompt.WriteString(fmt.Sprintf("- Build System: %s\n", context.BuildSystem))
	}
	if len(context.DevServices) > 0 {
		prompt.WriteString(fmt.Sprintf("- Dev Services: %s\n", strings.Join(context.DevServices, ", ")))
	}
	if context.ContainerNeeds {
		prompt.WriteString("- Container Support: Required\n")
	}
	if context.EditorConfig != "" {
		prompt.WriteString(fmt.Sprintf("- Editor Config: %s\n", context.EditorConfig))
	}
	if context.TestingFramework != "" {
		prompt.WriteString(fmt.Sprintf("- Testing Framework: %s\n", context.TestingFramework))
	}
	if len(context.LintingTools) > 0 {
		prompt.WriteString(fmt.Sprintf("- Linting Tools: %s\n", strings.Join(context.LintingTools, ", ")))
	}
	if len(context.DebugTools) > 0 {
		prompt.WriteString(fmt.Sprintf("- Debug Tools: %s\n", strings.Join(context.DebugTools, ", ")))
	}
	if len(context.Documentation) > 0 {
		prompt.WriteString(fmt.Sprintf("- Documentation Tools: %s\n", strings.Join(context.Documentation, ", ")))
	}

	prompt.WriteString("\nDevelopment Environment Question:\n")
	prompt.WriteString(question)

	return prompt.String()
}

// formatDevenvResponse formats the AI response for development environment guidance.
func (a *DevenvAgent) formatDevenvResponse(response string) string {
	// Add devenv-specific formatting and guidance
	var formatted strings.Builder
	
	formatted.WriteString("🛠️ Development Environment Guidance:\n\n")
	formatted.WriteString(response)
	
	// Add common devenv reminders
	formatted.WriteString("\n\n📋 Development Environment Best Practices:")
	formatted.WriteString("\n• Use declarative configuration with nix-shell or flakes")
	formatted.WriteString("\n• Pin package versions for reproducible builds")
	formatted.WriteString("\n• Use direnv for automatic environment activation")
	formatted.WriteString("\n• Include all development tools in the environment")
	formatted.WriteString("\n• Document environment setup in README or docs")
	formatted.WriteString("\n• Test the environment on a fresh system")
	formatted.WriteString("\n• Consider using cachix for faster builds")
	
	return formatted.String()
}

// getDevenvContextFromData extracts devenv context from stored data.
func (a *DevenvAgent) getDevenvContextFromData() *DevenvContext {
	if ctx, ok := a.contextData.(*DevenvContext); ok {
		return ctx
	}
	return &DevenvContext{}
}

// AnalyzeProject analyzes a project and suggests development environment setup.
func (a *DevenvAgent) AnalyzeProject(ctx context.Context, projectPath, projectType string) (string, error) {
	devenvCtx := &DevenvContext{
		ProjectType: projectType,
		ProjectRoot: projectPath,
		Environment: "development",
	}
	
	a.SetDevenvContext(devenvCtx)
	
	question := fmt.Sprintf("Analyze the %s project at %s and recommend a complete development environment setup including tools, dependencies, and configuration.", projectType, projectPath)
	
	return a.GenerateResponse(ctx, question)
}

// GenerateShellNix creates a shell.nix configuration for the project.
func (a *DevenvAgent) GenerateShellNix(ctx context.Context, devenvCtx *DevenvContext) (string, error) {
	a.SetDevenvContext(devenvCtx)
	
	var request strings.Builder
	request.WriteString("Generate a comprehensive shell.nix file that includes:")
	request.WriteString("\n1. All required packages and dependencies")
	request.WriteString("\n2. Development tools and utilities")
	request.WriteString("\n3. Environment variables and shell hooks")
	request.WriteString("\n4. Build system integration")
	request.WriteString("\n5. Editor/IDE support tools")
	request.WriteString("\n6. Testing and debugging tools")
	
	return a.GenerateResponse(ctx, request.String())
}

// GenerateFlakeNix creates a flake.nix configuration for the project.
func (a *DevenvAgent) GenerateFlakeNix(ctx context.Context, devenvCtx *DevenvContext) (string, error) {
	devenvCtx.Flakes = true
	a.SetDevenvContext(devenvCtx)
	
	var request strings.Builder
	request.WriteString("Generate a comprehensive flake.nix file that includes:")
	request.WriteString("\n1. Development shell with all required packages")
	request.WriteString("\n2. Build outputs for the project")
	request.WriteString("\n3. Multiple development environments (dev, testing, ci)")
	request.WriteString("\n4. Proper input management and version pinning")
	request.WriteString("\n5. Cross-platform compatibility")
	request.WriteString("\n6. Integration with common development workflows")
	
	return a.GenerateResponse(ctx, request.String())
}

// SetupDirenv provides direnv setup guidance.
func (a *DevenvAgent) SetupDirenv(ctx context.Context, context *DevenvContext) (string, error) {
	context.Direnv = true
	a.SetDevenvContext(context)
	
	question := "Provide complete setup instructions for direnv integration including .envrc configuration, shell integration, and usage best practices."
	
	return a.GenerateResponse(ctx, question)
}

// OptimizeBuildPerformance provides build optimization suggestions.
func (a *DevenvAgent) OptimizeBuildPerformance(ctx context.Context, context *DevenvContext) (string, error) {
	a.SetDevenvContext(context)
	
	question := "Analyze the development environment setup and provide recommendations for optimizing build performance, including caching strategies, parallel builds, and dependency management."
	
	return a.GenerateResponse(ctx, question)
}

// TroubleshootEnvironment helps debug development environment issues.
func (a *DevenvAgent) TroubleshootEnvironment(ctx context.Context, issues []string, context *DevenvContext) (string, error) {
	a.SetDevenvContext(context)
	
	question := fmt.Sprintf("Help troubleshoot these development environment issues: %s. Provide specific debugging steps and solutions.", strings.Join(issues, ", "))
	
	return a.GenerateResponse(ctx, question)
}
