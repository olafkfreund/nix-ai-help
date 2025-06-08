package packaging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/packaging/templates"
	"nix-ai-help/pkg/logger"
)

// EnhancedDerivationGenerator combines template-based and AI-assisted derivation generation
type EnhancedDerivationGenerator struct {
	analyzer        *RepositoryAnalyzer
	templateManager *templates.TemplateManager
	aiProvider      ai.AIProvider
	mcpClient       *mcp.MCPClient
	logger          *logger.Logger
}

// GenerationMode defines how the derivation should be generated
type GenerationMode int

const (
	ModeTemplate GenerationMode = iota // Use template-based generation
	ModeAI                             // Use AI-assisted generation
	ModeHybrid                         // Use template as base, enhance with AI
)

// GenerationOptions controls derivation generation behavior
type GenerationOptions struct {
	Mode        GenerationMode
	OutputPath  string
	Interactive bool
	UseCache    bool
}

// NewEnhancedDerivationGenerator creates a new enhanced derivation generator
func NewEnhancedDerivationGenerator(aiProvider ai.AIProvider, mcpClient *mcp.MCPClient, log *logger.Logger) *EnhancedDerivationGenerator {
	return &EnhancedDerivationGenerator{
		analyzer:        NewRepositoryAnalyzer(log),
		templateManager: templates.NewTemplateManager(),
		aiProvider:      aiProvider,
		mcpClient:       mcpClient,
		logger:          log,
	}
}

// GenerateFromRepo generates a Nix derivation from a repository with the specified options
func (edg *EnhancedDerivationGenerator) GenerateFromRepo(ctx context.Context, repoPath string, opts GenerationOptions) (string, error) {
	// Analyze the repository
	analysis, err := edg.analyzer.AnalyzeRepository(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to analyze repository: %w", err)
	}

	edg.logger.Info(fmt.Sprintf("Analyzed repository: %s (Language: %s, Build System: %s)",
		analysis.ProjectName, analysis.Language, analysis.BuildSystem))

	// Generate derivation based on mode
	var derivation string
	switch opts.Mode {
	case ModeTemplate:
		derivation, err = edg.generateTemplate(analysis)
	case ModeAI:
		derivation, err = edg.generateAI(ctx, analysis)
	case ModeHybrid:
		derivation, err = edg.generateHybrid(ctx, analysis)
	default:
		return "", fmt.Errorf("unknown generation mode: %d", opts.Mode)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate derivation: %w", err)
	}

	// Write to file if output path specified
	if opts.OutputPath != "" {
		err = edg.writeDerivationToFile(derivation, opts.OutputPath)
		if err != nil {
			return "", fmt.Errorf("failed to write derivation to file: %w", err)
		}
		edg.logger.Info(fmt.Sprintf("Derivation written to: %s", opts.OutputPath))
	}

	return derivation, nil
}

// generateTemplate generates a derivation using template-based approach
func (edg *EnhancedDerivationGenerator) generateTemplate(analysis *RepoAnalysis) (string, error) {
	// Get appropriate template
	template, err := edg.templateManager.GetTemplate(analysis.Language, string(analysis.BuildSystem))
	if err != nil {
		edg.logger.Warn(fmt.Sprintf("No specific template found, using default: %v", err))
		// Fall back to default template
		template, err = edg.templateManager.GetTemplate("", "default")
		if err != nil {
			return "", fmt.Errorf("failed to get default template: %w", err)
		}
	}

	// Create template context from analysis
	context := edg.createTemplateContext(analysis)

	// Validate context against template requirements
	errors := edg.templateManager.ValidateContext(template, context)
	if len(errors) > 0 {
		edg.logger.Warn(fmt.Sprintf("Template validation warnings: %v", errors))
		// Fill in missing required variables with defaults
		edg.fillMissingRequiredVariables(context, template)
	}

	// Apply template
	derivation, err := edg.templateManager.ApplyTemplate(template, context)
	if err != nil {
		return "", fmt.Errorf("failed to apply template: %w", err)
	}

	return derivation, nil
}

// generateAI generates a derivation using AI-assisted approach (delegates to existing generator)
func (edg *EnhancedDerivationGenerator) generateAI(ctx context.Context, analysis *RepoAnalysis) (string, error) {
	// Create legacy generator for AI-based generation
	legacyGenerator := NewDerivationGenerator(edg.aiProvider, edg.mcpClient)
	return legacyGenerator.GenerateDerivation(ctx, analysis)
}

// generateHybrid generates a derivation using both template and AI (template as base, AI for enhancement)
func (edg *EnhancedDerivationGenerator) generateHybrid(ctx context.Context, analysis *RepoAnalysis) (string, error) {
	// First generate base derivation with template
	templateDerivation, err := edg.generateTemplate(analysis)
	if err != nil {
		return "", fmt.Errorf("failed to generate template base: %w", err)
	}

	// Get nixpkgs context for AI enhancement
	nixpkgsContext, err := edg.getNixpkgsContext(ctx, analysis.BuildSystem, analysis.Language)
	if err != nil {
		edg.logger.Warn(fmt.Sprintf("Failed to get nixpkgs context: %v", err))
		nixpkgsContext = ""
	}

	// Create enhancement prompt
	prompt := edg.createEnhancementPrompt(analysis, templateDerivation, nixpkgsContext)

	// Enhance with AI
	response, err := edg.aiProvider.Query(prompt)
	if err != nil {
		edg.logger.Warn(fmt.Sprintf("AI enhancement failed, returning template: %v", err))
		return templateDerivation, nil
	}

	// Extract enhanced derivation
	enhanced := edg.extractDerivation(response)
	if enhanced == "" {
		edg.logger.Warn("AI enhancement returned empty, using template")
		return templateDerivation, nil
	}

	return enhanced, nil
}

// createTemplateContext creates a template context from repository analysis
func (edg *EnhancedDerivationGenerator) createTemplateContext(analysis *RepoAnalysis) *templates.TemplateContext {
	context := &templates.TemplateContext{
		ProjectName:       analysis.ProjectName,
		Language:          analysis.Language,
		BuildSystem:       string(analysis.BuildSystem),
		Description:       analysis.Description,
		License:           analysis.License,
		Dependencies:      make(map[string]string),
		DevDependencies:   make(map[string]string),
		BuildInputs:       []string{},
		NativeBuildInputs: []string{},
		Custom:            make(map[string]interface{}),
	}

	// Set default version if not available
	if context.Version == "" {
		context.Version = "0.1.0"
	}

	// Extract owner from repo URL if available
	if analysis.RepoURL != "" {
		owner := edg.extractOwnerFromURL(analysis.RepoURL)
		if owner != "" {
			context.Custom["Owner"] = owner
		}
	}

	// Process dependencies
	for _, dep := range analysis.Dependencies {
		if dep.Type == "runtime" || dep.Type == "build" {
			context.Dependencies[dep.Name] = dep.Version
		} else if dep.Type == "dev" {
			context.DevDependencies[dep.Name] = dep.Version
		}

		// Add system dependencies to build inputs
		if dep.System {
			context.BuildInputs = append(context.BuildInputs, dep.Name)
		}
	}

	// Set build system specific inputs
	edg.setBuildSystemInputs(context, analysis.BuildSystem)

	// Add custom phases based on detected patterns
	edg.setCustomPhases(context, analysis)

	return context
}

// extractOwnerFromURL extracts the owner/organization from a Git URL
func (edg *EnhancedDerivationGenerator) extractOwnerFromURL(url string) string {
	// Handle GitHub URLs
	if strings.Contains(url, "github.com") {
		parts := strings.Split(url, "/")
		for i, part := range parts {
			if part == "github.com" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}

// setBuildSystemInputs sets build system specific native build inputs
func (edg *EnhancedDerivationGenerator) setBuildSystemInputs(context *templates.TemplateContext, buildSystem BuildSystem) {
	switch buildSystem {
	case BuildSystemNpm:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "nodejs", "npm")
	case BuildSystemYarn:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "nodejs", "yarn")
	case BuildSystemCMake:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "cmake")
	case BuildSystemMeson:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "meson", "ninja")
	case BuildSystemAutotools:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "autoconf", "automake", "libtool")
	case BuildSystemCargoRust:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "rustc", "cargo")
	case BuildSystemGo:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "go")
	case BuildSystemPython:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "python3")
	case BuildSystemMaven:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "maven")
	case BuildSystemGradle:
		context.NativeBuildInputs = append(context.NativeBuildInputs, "gradle")
	}
}

// setCustomPhases sets custom build phases based on analysis
func (edg *EnhancedDerivationGenerator) setCustomPhases(context *templates.TemplateContext, analysis *RepoAnalysis) {
	// Set check phase if tests are detected
	if analysis.HasTests {
		switch context.BuildSystem {
		case "npm", "yarn":
			context.CheckPhase = "npm test"
		case "cargo":
			context.CheckPhase = "cargo test"
		case "go":
			context.CheckPhase = "go test ./..."
		case "python":
			context.CheckPhase = "python -m pytest"
		case "cmake":
			context.CheckPhase = "make test"
		}
	}
}

// fillMissingRequiredVariables fills missing required variables with sensible defaults
func (edg *EnhancedDerivationGenerator) fillMissingRequiredVariables(context *templates.TemplateContext, template *templates.DerivationTemplate) {
	for _, variable := range template.Variables {
		if !variable.Required {
			continue
		}

		switch variable.Name {
		case "Version":
			if context.Version == "" {
				context.Version = "0.1.0"
			}
		case "Owner":
			if context.Custom == nil {
				context.Custom = make(map[string]interface{})
			}
			if _, exists := context.Custom["Owner"]; !exists {
				context.Custom["Owner"] = "unknown"
			}
		case "ProjectName":
			if context.ProjectName == "" {
				context.ProjectName = "unknown-project"
			}
		case "Language":
			if context.Language == "" {
				context.Language = "unknown"
			}
		case "BuildSystem":
			if context.BuildSystem == "" {
				context.BuildSystem = "unknown"
			}
		}
	}
}

// writeDerivationToFile writes the derivation to a file
func (edg *EnhancedDerivationGenerator) writeDerivationToFile(derivation, outputPath string) error {
	// Ensure the output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the derivation to file
	return os.WriteFile(outputPath, []byte(derivation), 0644)
}

// getNixpkgsContext and other AI-related methods (delegate to existing implementation)
func (edg *EnhancedDerivationGenerator) getNixpkgsContext(ctx context.Context, buildSystem BuildSystem, language string) (string, error) {
	// Create legacy generator to reuse existing logic
	legacyGenerator := NewDerivationGenerator(edg.aiProvider, edg.mcpClient)
	return legacyGenerator.GetNixpkgsContext(ctx, buildSystem, language)
}

// createEnhancementPrompt creates a prompt for AI to enhance a template-based derivation
func (edg *EnhancedDerivationGenerator) createEnhancementPrompt(analysis *RepoAnalysis, templateDerivation, nixpkgsContext string) string {
	var prompt strings.Builder

	prompt.WriteString(`You are an expert Nix package maintainer. I have a template-based Nix derivation that needs enhancement and optimization.

CURRENT TEMPLATE-BASED DERIVATION:
`)
	prompt.WriteString("```nix\n")
	prompt.WriteString(templateDerivation)
	prompt.WriteString("\n```\n\n")

	prompt.WriteString("PROJECT ANALYSIS:\n")
	prompt.WriteString(fmt.Sprintf("- Project Name: %s\n", analysis.ProjectName))
	prompt.WriteString(fmt.Sprintf("- Build System: %s\n", analysis.BuildSystem))
	prompt.WriteString(fmt.Sprintf("- Primary Language: %s\n", analysis.Language))
	if analysis.License != "" {
		prompt.WriteString(fmt.Sprintf("- License: %s\n", analysis.License))
	}
	if analysis.Description != "" {
		prompt.WriteString(fmt.Sprintf("- Description: %s\n", analysis.Description))
	}

	if len(analysis.Dependencies) > 0 {
		prompt.WriteString(fmt.Sprintf("- Dependencies: %d found\n", len(analysis.Dependencies)))
	}

	if nixpkgsContext != "" {
		prompt.WriteString("\nRELEVANT NIXPKGS DOCUMENTATION:\n")
		prompt.WriteString(nixpkgsContext)
		prompt.WriteString("\n")
	}

	prompt.WriteString(`
ENHANCEMENT REQUIREMENTS:
1. Review and optimize the template-based derivation
2. Add missing build inputs or native build inputs based on the project analysis
3. Improve build phases, install phases, or check phases if needed
4. Ensure proper dependency handling
5. Add any missing meta attributes
6. Follow nixpkgs best practices and conventions
7. Keep the structure clean and maintainable

Please provide the enhanced Nix derivation. Return ONLY the complete .nix file content, no explanations.`)

	return prompt.String()
}

// extractDerivation extracts clean derivation from AI response (delegate to existing implementation)
func (edg *EnhancedDerivationGenerator) extractDerivation(response string) string {
	// Create legacy generator to reuse existing logic
	legacyGenerator := NewDerivationGenerator(edg.aiProvider, edg.mcpClient)
	return legacyGenerator.ExtractDerivation(response)
}
