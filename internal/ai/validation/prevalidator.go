package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/internal/community"
	"nix-ai-help/internal/mcp"
)

// FactualValidationResult represents the result of factual validation against sources
type FactualValidationResult struct {
	IsFactuallyAccurate bool
	ConfidenceLevel     string // "high", "medium", "low", "unknown"
	VerifiedSources     []VerifiedSource
	Warnings            []FactualWarning
	Suggestions         []string
	ValidationTime      time.Duration
}

// VerifiedSource represents a source that verified information
type VerifiedSource struct {
	Type        string // "nixos-wiki", "github-code", "mcp-docs", "manual"
	URL         string
	Description string
	Relevance   float64 // 0.0 to 1.0
}

// FactualWarning represents a potential factual issue
type FactualWarning struct {
	Type       string
	Message    string
	Suggestion string
	Severity   string // "low", "medium", "high", "critical"
	Source     string
}

// PreAnswerValidator performs factual validation before generating responses
type PreAnswerValidator struct {
	mcpClient    *mcp.MCPClient
	githubClient *community.GitHubClient

	// Common NixOS option patterns to validate
	nixosOptionPatterns map[string][]string

	// Known deprecated/incorrect patterns
	deprecatedPatterns []DeprecatedPattern
}

// DeprecatedPattern represents patterns that are known to be deprecated or incorrect
type DeprecatedPattern struct {
	Pattern     *regexp.Regexp
	Reason      string
	Alternative string
	Severity    string
}

// NewPreAnswerValidator creates a new pre-answer validator
func NewPreAnswerValidator(mcpHost string, mcpPort int, githubToken string) *PreAnswerValidator {
	var mcpClient *mcp.MCPClient
	if mcpHost != "" {
		mcpClient = mcp.NewMCPClient(fmt.Sprintf("http://%s:%d", mcpHost, mcpPort))
	}

	githubClient := community.NewGitHubClient(githubToken)

	// Define NixOS option patterns to validate
	nixosOptionPatterns := map[string][]string{
		"bluetooth": {"hardware.bluetooth.enable", "services.blueman.enable"},
		"audio":     {"sound.enable", "security.rtkit.enable", "services.pipewire", "hardware.pulseaudio"},
		"wifi":      {"networking.wireless.enable", "networking.networkmanager.enable"},
		"ssh":       {"services.openssh.enable", "services.openssh.settings"},
		"firewall":  {"networking.firewall.enable", "networking.firewall.allowedTCPPorts"},
		"graphics":  {"hardware.opengl.enable", "services.xserver.enable", "hardware.graphics.enable"},
		"docker":    {"virtualisation.docker.enable", "virtualisation.containers.enable"},
		"nvidia":    {"hardware.nvidia.modesetting.enable", "services.xserver.videoDrivers"},
	}

	// Define deprecated patterns
	deprecatedPatterns := []DeprecatedPattern{
		{
			Pattern:     regexp.MustCompile(`services\.bluetooth\.enable`),
			Reason:      "services.bluetooth.enable does not exist in NixOS",
			Alternative: "hardware.bluetooth.enable",
			Severity:    "critical",
		},
		{
			Pattern:     regexp.MustCompile(`nix-env\s+-[iuq]`),
			Reason:      "nix-env is deprecated in favor of declarative configuration",
			Alternative: "Use environment.systemPackages or nix profile",
			Severity:    "high",
		},
		{
			Pattern:     regexp.MustCompile(`services\.audio\.enable`),
			Reason:      "services.audio.enable does not exist in NixOS",
			Alternative: "sound.enable or services.pipewire.enable",
			Severity:    "critical",
		},
	}

	return &PreAnswerValidator{
		mcpClient:           mcpClient,
		githubClient:        githubClient,
		nixosOptionPatterns: nixosOptionPatterns,
		deprecatedPatterns:  deprecatedPatterns,
	}
}

// ValidateQuestionFactually performs factual validation before generating an answer
func (pav *PreAnswerValidator) ValidateQuestionFactually(ctx context.Context, question string) (*FactualValidationResult, error) {
	startTime := time.Now()

	result := &FactualValidationResult{
		IsFactuallyAccurate: true,
		ConfidenceLevel:     "unknown",
		VerifiedSources:     []VerifiedSource{},
		Warnings:            []FactualWarning{},
		Suggestions:         []string{},
	}

	// Extract NixOS-related terms from the question
	nixosTerms := pav.extractNixOSTerms(question)

	if len(nixosTerms) == 0 {
		result.ConfidenceLevel = "low"
		result.Warnings = append(result.Warnings, FactualWarning{
			Type:     "insufficient_context",
			Message:  "Question does not contain recognizable NixOS terms",
			Severity: "medium",
		})
	}

	// Validate against MCP documentation
	if pav.mcpClient != nil {
		mcpSources, err := pav.validateAgainstMCP(ctx, question, nixosTerms)
		if err == nil {
			result.VerifiedSources = append(result.VerifiedSources, mcpSources...)
		}
	}

	// Validate against GitHub code search
	if pav.githubClient != nil {
		githubSources, err := pav.validateAgainstGitHub(ctx, question, nixosTerms)
		if err == nil {
			result.VerifiedSources = append(result.VerifiedSources, githubSources...)
		}
	}

	// Check for deprecated patterns in the question
	for _, pattern := range pav.deprecatedPatterns {
		if pattern.Pattern.MatchString(question) {
			result.Warnings = append(result.Warnings, FactualWarning{
				Type:       "deprecated_pattern_in_question",
				Message:    "Question contains deprecated pattern: " + pattern.Reason,
				Suggestion: "Consider asking about: " + pattern.Alternative,
				Severity:   pattern.Severity,
			})
		}
	}

	// Determine confidence level based on verified sources
	result.ConfidenceLevel = pav.calculateConfidenceLevel(result.VerifiedSources, result.Warnings)

	// Generate suggestions based on validation results
	result.Suggestions = pav.generateSuggestions(question, result.VerifiedSources, nixosTerms)

	result.ValidationTime = time.Since(startTime)

	return result, nil
}

// extractNixOSTerms extracts NixOS-related terms from the question
func (pav *PreAnswerValidator) extractNixOSTerms(question string) []string {
	lowerQuestion := strings.ToLower(question)
	var terms []string

	// Check for known NixOS terms
	nixosTerms := map[string]bool{
		"bluetooth":      true,
		"audio":          true,
		"sound":          true,
		"wifi":           true,
		"wireless":       true,
		"network":        true,
		"ssh":            true,
		"openssh":        true,
		"firewall":       true,
		"graphics":       true,
		"opengl":         true,
		"nvidia":         true,
		"docker":         true,
		"virtualisation": true,
		"services":       true,
		"hardware":       true,
		"networking":     true,
		"environment":    true,
		"flake":          true,
		"configuration":  true,
		"nixpkgs":        true,
		"systempackages": true,
		"enable":         true,
	}

	words := strings.Fields(lowerQuestion)
	for _, word := range words {
		// Remove punctuation
		clean := regexp.MustCompile(`[^\w]`).ReplaceAllString(word, "")
		if nixosTerms[clean] {
			terms = append(terms, clean)
		}
	}

	return terms
}

// validateAgainstMCP validates question context against MCP documentation
func (pav *PreAnswerValidator) validateAgainstMCP(ctx context.Context, question string, terms []string) ([]VerifiedSource, error) {
	var sources []VerifiedSource

	// Query MCP for each term
	for _, term := range terms {
		doc, err := pav.mcpClient.QueryDocumentation(term)
		if err != nil {
			continue
		}

		if doc != "" && len(doc) > 50 {
			relevance := pav.calculateRelevance(question, doc)
			sources = append(sources, VerifiedSource{
				Type:        "mcp-docs",
				Description: fmt.Sprintf("MCP documentation for '%s'", term),
				Relevance:   relevance,
			})
		}
	}

	return sources, nil
}

// validateAgainstGitHub validates against GitHub code search
func (pav *PreAnswerValidator) validateAgainstGitHub(ctx context.Context, question string, terms []string) ([]VerifiedSource, error) {
	var sources []VerifiedSource

	// Search for real-world NixOS configurations
	for _, term := range terms {
		if len(term) < 4 { // Skip very short terms
			continue
		}

		configs, err := pav.githubClient.SearchNixOSConfigurations(term)
		if err != nil {
			continue
		}

		for i, config := range configs {
			if i >= 3 { // Limit to top 3 results per term
				break
			}

			relevance := pav.calculateRelevance(question, config.Content)
			sources = append(sources, VerifiedSource{
				Type:        "github-code",
				URL:         config.URL,
				Description: fmt.Sprintf("GitHub config: %s", config.Description),
				Relevance:   relevance,
			})
		}
	}

	return sources, nil
}

// calculateRelevance calculates how relevant a source is to the question
func (pav *PreAnswerValidator) calculateRelevance(question, content string) float64 {
	questionWords := strings.Fields(strings.ToLower(question))
	contentLower := strings.ToLower(content)

	matches := 0
	for _, word := range questionWords {
		if len(word) > 3 && strings.Contains(contentLower, word) {
			matches++
		}
	}

	if len(questionWords) == 0 {
		return 0.0
	}

	return float64(matches) / float64(len(questionWords))
}

// calculateConfidenceLevel determines confidence based on sources and warnings
func (pav *PreAnswerValidator) calculateConfidenceLevel(sources []VerifiedSource, warnings []FactualWarning) string {
	// Count high-relevance sources
	highRelevanceSources := 0
	for _, source := range sources {
		if source.Relevance > 0.5 {
			highRelevanceSources++
		}
	}

	// Count critical warnings
	criticalWarnings := 0
	for _, warning := range warnings {
		if warning.Severity == "critical" {
			criticalWarnings++
		}
	}

	if criticalWarnings > 0 {
		return "low"
	}

	if highRelevanceSources >= 3 {
		return "high"
	} else if highRelevanceSources >= 1 {
		return "medium"
	}

	return "low"
}

// generateSuggestions generates helpful suggestions based on validation
func (pav *PreAnswerValidator) generateSuggestions(question string, sources []VerifiedSource, terms []string) []string {
	var suggestions []string

	if len(sources) == 0 {
		suggestions = append(suggestions, "Consider refining your question with more specific NixOS terms")
		suggestions = append(suggestions, "Use 'nixai search <term>' to find relevant packages or options first")
	}

	// Suggest specific options based on detected terms
	for _, term := range terms {
		if options, exists := pav.nixosOptionPatterns[term]; exists {
			suggestions = append(suggestions, fmt.Sprintf("For '%s', consider options: %s", term, strings.Join(options, ", ")))
		}
	}

	if len(sources) > 0 {
		suggestions = append(suggestions, "This question has verified sources - answer should be reliable")
	}

	return suggestions
}

// FormatFactualValidationResult formats the validation result for display
func (pav *PreAnswerValidator) FormatFactualValidationResult(result *FactualValidationResult) string {
	var output strings.Builder

	output.WriteString("## 🔍 **Pre-Answer Factual Validation**\n\n")

	// Confidence level
	confidenceEmoji := map[string]string{
		"high":    "🟢",
		"medium":  "🟡",
		"low":     "🔴",
		"unknown": "⚪",
	}

	output.WriteString(fmt.Sprintf("**Confidence Level**: %s %s\n",
		confidenceEmoji[result.ConfidenceLevel],
		strings.ToUpper(result.ConfidenceLevel)))

	// Verified sources
	if len(result.VerifiedSources) > 0 {
		output.WriteString(fmt.Sprintf("\n**Verified Sources** (%d found):\n", len(result.VerifiedSources)))
		for _, source := range result.VerifiedSources {
			relevanceStr := fmt.Sprintf("%.0f%%", source.Relevance*100)
			output.WriteString(fmt.Sprintf("- %s (%s relevance): %s\n",
				source.Type, relevanceStr, source.Description))
		}
	}

	// Warnings
	if len(result.Warnings) > 0 {
		output.WriteString("\n**⚠️ Validation Warnings**:\n")
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("- **%s**: %s\n", warning.Severity, warning.Message))
			if warning.Suggestion != "" {
				output.WriteString(fmt.Sprintf("  💡 *%s*\n", warning.Suggestion))
			}
		}
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		output.WriteString("\n**💡 Suggestions**:\n")
		for _, suggestion := range result.Suggestions {
			output.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
	}

	output.WriteString(fmt.Sprintf("\n*Validation completed in %v*\n", result.ValidationTime))

	return output.String()
}
