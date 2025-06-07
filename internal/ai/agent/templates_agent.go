package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// TemplatesAgent manages NixOS configuration templates and scaffolding.
type TemplatesAgent struct {
	BaseAgent
}

// TemplateContext provides context for template operations.
type TemplateContext struct {
	TemplateType  string            // "flake", "nixos", "home-manager", "devenv", "package"
	ProjectName   string            // Name of the project/configuration
	Purpose       string            // What the template will be used for
	Features      []string          // Required features (GUI, development, server, etc.)
	Architecture  string            // Target architecture
	Language      string            // Primary programming language (if applicable)
	Framework     string            // Framework or technology stack
	Services      []string          // Services to include (postgres, redis, nginx, etc.)
	Customization string            // Specific customization requirements
	BaseTemplate  string            // Base template to start from
	OutputPath    string            // Where to generate the template
	Metadata      map[string]string // Additional template metadata
}

// NewTemplatesAgent creates a new TemplatesAgent with the specified provider.
func NewTemplatesAgent(provider ai.Provider) *TemplatesAgent {
	agent := &TemplatesAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleTemplates,
		},
	}
	return agent
}

// Query provides template guidance using the provider's Query method.
func (a *TemplatesAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build template-specific prompt with context
	prompt := a.buildTemplatePrompt(question, a.getTemplateContextFromData())

	// Query the AI provider
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to query provider: %w", err)
	}

	// Enhance response with template guidance
	return a.enhanceResponseWithTemplateGuidance(response), nil
}

// GenerateResponse generates a response using the provider's GenerateResponse method.
func (a *TemplatesAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Enhance the prompt with role-specific instructions
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", err
	}

	return a.enhanceResponseWithTemplateGuidance(response), nil
}

// GenerateTemplate creates a new template configuration based on requirements.
func (a *TemplatesAgent) GenerateTemplate(ctx context.Context, templateCtx *TemplateContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildGenerateTemplatePrompt(templateCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate template: %w", err)
	}

	return a.formatTemplateOutput(response, templateCtx), nil
}

// CustomizeTemplate adapts an existing template to specific requirements.
func (a *TemplatesAgent) CustomizeTemplate(ctx context.Context, baseTemplate string, templateCtx *TemplateContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildCustomizeTemplatePrompt(baseTemplate, templateCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to customize template: %w", err)
	}

	return a.formatTemplateOutput(response, templateCtx), nil
}

// ExplainTemplate provides detailed explanation of template structure and components.
func (a *TemplatesAgent) ExplainTemplate(ctx context.Context, template string, templateCtx *TemplateContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildExplainTemplatePrompt(template, templateCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to explain template: %w", err)
	}

	return a.enhanceResponseWithTemplateGuidance(response), nil
}

// ValidateTemplate checks template syntax and structure for correctness.
func (a *TemplatesAgent) ValidateTemplate(ctx context.Context, template string, templateCtx *TemplateContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildValidateTemplatePrompt(template, templateCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to validate template: %w", err)
	}

	return a.enhanceResponseWithTemplateGuidance(response), nil
}

// SuggestImprovements provides template optimization recommendations.
func (a *TemplatesAgent) SuggestImprovements(ctx context.Context, template string, templateCtx *TemplateContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildImprovementPrompt(template, templateCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to suggest improvements: %w", err)
	}

	return a.enhanceResponseWithTemplateGuidance(response), nil
}

// buildTemplatePrompt builds the prompt for general template queries.
func (a *TemplatesAgent) buildTemplatePrompt(question string, templateCtx *TemplateContext) string {
	var promptBuilder strings.Builder

	// Get role-specific prompt template
	rolePrompt, exists := roles.RolePromptTemplate[a.role]
	if exists {
		promptBuilder.WriteString(rolePrompt)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Template Query: ")
	promptBuilder.WriteString(question)
	promptBuilder.WriteString("\n\n")

	// Add template context
	if templateCtx != nil {
		a.addTemplateContext(&promptBuilder, templateCtx)
	}

	promptBuilder.WriteString("\nProvide comprehensive template guidance with specific examples and best practices.")

	return promptBuilder.String()
}

// buildGenerateTemplatePrompt builds prompt for template generation.
func (a *TemplatesAgent) buildGenerateTemplatePrompt(templateCtx *TemplateContext) string {
	var promptBuilder strings.Builder

	rolePrompt, exists := roles.RolePromptTemplate[a.role]
	if exists {
		promptBuilder.WriteString(rolePrompt)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Generate a complete ")
	promptBuilder.WriteString(templateCtx.TemplateType)
	promptBuilder.WriteString(" template with the following requirements:\n\n")

	a.addTemplateContext(&promptBuilder, templateCtx)

	promptBuilder.WriteString("\nProvide:\n")
	promptBuilder.WriteString("1. Complete configuration files with clear comments\n")
	promptBuilder.WriteString("2. Directory structure and file organization\n")
	promptBuilder.WriteString("3. Setup instructions and usage examples\n")
	promptBuilder.WriteString("4. Customization guidance and best practices\n")

	return promptBuilder.String()
}

// buildCustomizeTemplatePrompt builds prompt for template customization.
func (a *TemplatesAgent) buildCustomizeTemplatePrompt(baseTemplate string, templateCtx *TemplateContext) string {
	var promptBuilder strings.Builder

	rolePrompt, exists := roles.RolePromptTemplate[a.role]
	if exists {
		promptBuilder.WriteString(rolePrompt)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Customize the following template according to the specified requirements:\n\n")
	promptBuilder.WriteString("Base Template:\n```\n")
	promptBuilder.WriteString(baseTemplate)
	promptBuilder.WriteString("\n```\n\n")

	promptBuilder.WriteString("Customization Requirements:\n")
	a.addTemplateContext(&promptBuilder, templateCtx)

	promptBuilder.WriteString("\nProvide the customized template with:\n")
	promptBuilder.WriteString("1. Modified configuration with required features\n")
	promptBuilder.WriteString("2. Explanation of changes made\n")
	promptBuilder.WriteString("3. Additional setup steps if needed\n")

	return promptBuilder.String()
}

// buildExplainTemplatePrompt builds prompt for template explanation.
func (a *TemplatesAgent) buildExplainTemplatePrompt(template string, templateCtx *TemplateContext) string {
	var promptBuilder strings.Builder

	rolePrompt, exists := roles.RolePromptTemplate[a.role]
	if exists {
		promptBuilder.WriteString(rolePrompt)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Explain the following template in detail:\n\n")
	promptBuilder.WriteString("Template:\n```\n")
	promptBuilder.WriteString(template)
	promptBuilder.WriteString("\n```\n\n")

	if templateCtx != nil {
		promptBuilder.WriteString("Context:\n")
		a.addTemplateContext(&promptBuilder, templateCtx)
	}

	promptBuilder.WriteString("\nProvide:\n")
	promptBuilder.WriteString("1. Overview of template purpose and structure\n")
	promptBuilder.WriteString("2. Section-by-section explanation of configuration\n")
	promptBuilder.WriteString("3. Key concepts and NixOS principles demonstrated\n")
	promptBuilder.WriteString("4. Customization opportunities and extension points\n")

	return promptBuilder.String()
}

// buildValidateTemplatePrompt builds prompt for template validation.
func (a *TemplatesAgent) buildValidateTemplatePrompt(template string, templateCtx *TemplateContext) string {
	var promptBuilder strings.Builder

	rolePrompt, exists := roles.RolePromptTemplate[a.role]
	if exists {
		promptBuilder.WriteString(rolePrompt)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Validate the following template for syntax, structure, and best practices:\n\n")
	promptBuilder.WriteString("Template:\n```\n")
	promptBuilder.WriteString(template)
	promptBuilder.WriteString("\n```\n\n")

	if templateCtx != nil {
		promptBuilder.WriteString("Context:\n")
		a.addTemplateContext(&promptBuilder, templateCtx)
	}

	promptBuilder.WriteString("\nCheck for:\n")
	promptBuilder.WriteString("1. Syntax errors and configuration issues\n")
	promptBuilder.WriteString("2. NixOS best practices compliance\n")
	promptBuilder.WriteString("3. Security considerations and recommendations\n")
	promptBuilder.WriteString("4. Performance optimization opportunities\n")
	promptBuilder.WriteString("5. Compatibility with specified requirements\n")

	return promptBuilder.String()
}

// buildImprovementPrompt builds prompt for template improvement suggestions.
func (a *TemplatesAgent) buildImprovementPrompt(template string, templateCtx *TemplateContext) string {
	var promptBuilder strings.Builder

	rolePrompt, exists := roles.RolePromptTemplate[a.role]
	if exists {
		promptBuilder.WriteString(rolePrompt)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Analyze the following template and suggest improvements:\n\n")
	promptBuilder.WriteString("Template:\n```\n")
	promptBuilder.WriteString(template)
	promptBuilder.WriteString("\n```\n\n")

	if templateCtx != nil {
		promptBuilder.WriteString("Context:\n")
		a.addTemplateContext(&promptBuilder, templateCtx)
	}

	promptBuilder.WriteString("\nSuggest improvements for:\n")
	promptBuilder.WriteString("1. Code organization and modularity\n")
	promptBuilder.WriteString("2. Performance and resource usage\n")
	promptBuilder.WriteString("3. Security hardening and best practices\n")
	promptBuilder.WriteString("4. Maintainability and documentation\n")
	promptBuilder.WriteString("5. Modern NixOS features and patterns\n")

	return promptBuilder.String()
}

// addTemplateContext adds template context information to the prompt.
func (a *TemplatesAgent) addTemplateContext(builder *strings.Builder, templateCtx *TemplateContext) {
	if templateCtx.TemplateType != "" {
		builder.WriteString("Template Type: ")
		builder.WriteString(templateCtx.TemplateType)
		builder.WriteString("\n")
	}

	if templateCtx.ProjectName != "" {
		builder.WriteString("Project Name: ")
		builder.WriteString(templateCtx.ProjectName)
		builder.WriteString("\n")
	}

	if templateCtx.Purpose != "" {
		builder.WriteString("Purpose: ")
		builder.WriteString(templateCtx.Purpose)
		builder.WriteString("\n")
	}

	if len(templateCtx.Features) > 0 {
		builder.WriteString("Required Features: ")
		builder.WriteString(strings.Join(templateCtx.Features, ", "))
		builder.WriteString("\n")
	}

	if templateCtx.Architecture != "" {
		builder.WriteString("Target Architecture: ")
		builder.WriteString(templateCtx.Architecture)
		builder.WriteString("\n")
	}

	if templateCtx.Language != "" {
		builder.WriteString("Programming Language: ")
		builder.WriteString(templateCtx.Language)
		builder.WriteString("\n")
	}

	if templateCtx.Framework != "" {
		builder.WriteString("Framework: ")
		builder.WriteString(templateCtx.Framework)
		builder.WriteString("\n")
	}

	if len(templateCtx.Services) > 0 {
		builder.WriteString("Services: ")
		builder.WriteString(strings.Join(templateCtx.Services, ", "))
		builder.WriteString("\n")
	}

	if templateCtx.Customization != "" {
		builder.WriteString("Customization Requirements: ")
		builder.WriteString(templateCtx.Customization)
		builder.WriteString("\n")
	}

	if templateCtx.BaseTemplate != "" {
		builder.WriteString("Base Template: ")
		builder.WriteString(templateCtx.BaseTemplate)
		builder.WriteString("\n")
	}
}

// getTemplateContextFromData extracts template context from stored data.
func (a *TemplatesAgent) getTemplateContextFromData() *TemplateContext {
	if a.contextData == nil {
		return &TemplateContext{}
	}

	if templateCtx, ok := a.contextData.(*TemplateContext); ok {
		return templateCtx
	}

	return &TemplateContext{}
}

// enhanceResponseWithTemplateGuidance adds template-specific guidance to responses.
func (a *TemplatesAgent) enhanceResponseWithTemplateGuidance(response string) string {
	var enhanced strings.Builder
	enhanced.WriteString(response)

	// Add helpful template guidance
	enhanced.WriteString("\n\n## Template Management Tips\n")
	enhanced.WriteString("- Use `nix flake init` to start with official templates\n")
	enhanced.WriteString("- Create reusable modules for common configuration patterns\n")
	enhanced.WriteString("- Test templates in virtual machines before deployment\n")
	enhanced.WriteString("- Keep templates minimal and document customization points\n")
	enhanced.WriteString("- Use template parameters to make configurations flexible\n")
	enhanced.WriteString("- Version control templates for collaboration and rollback\n")

	return enhanced.String()
}

// formatTemplateOutput formats template generation output for clarity.
func (a *TemplatesAgent) formatTemplateOutput(response string, templateCtx *TemplateContext) string {
	var formatted strings.Builder

	formatted.WriteString("# Generated ")
	formatted.WriteString(strings.Title(templateCtx.TemplateType))
	formatted.WriteString(" Template")

	if templateCtx.ProjectName != "" {
		formatted.WriteString(" for ")
		formatted.WriteString(templateCtx.ProjectName)
	}

	formatted.WriteString("\n\n")
	formatted.WriteString(response)

	// Add usage instructions
	formatted.WriteString("\n\n## Usage Instructions\n")
	formatted.WriteString("1. Save the configuration to appropriate files\n")
	formatted.WriteString("2. Review and customize the settings for your environment\n")
	formatted.WriteString("3. Test the configuration before applying to production\n")
	formatted.WriteString("4. Use `nixos-rebuild switch` or appropriate commands to apply\n")

	return formatted.String()
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *TemplatesAgent) enhancePromptWithRole(prompt string) string {
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		return fmt.Sprintf("%s\n\n%s", template, prompt)
	}
	return prompt
}
