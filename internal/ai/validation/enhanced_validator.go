package validation

import (
	"context"
	"time"

	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
)

// EnhancedValidator combines multiple validation sources for comprehensive accuracy checking
type EnhancedValidator struct {
	// Existing components
	preValidator   *PreAnswerValidator
	nixosValidator *NixOSValidator
	flakeValidator *FlakeValidator

	// New components
	searchNixOSClient  *nixos.SearchNixOSClient
	nixToolsExecutor   *nixos.ToolsExecutor
	communityValidator *CommunityValidator
	factChecker        *FactChecker
	crossRefValidator  *CrossReferenceValidator
	qualityMetrics     *QualityMetrics
	confidenceScorer   *ConfidenceScorer
	automatedScorer    *AutomatedQualityScorer

	logger *logger.Logger
}

// EnhancedValidationResult represents comprehensive validation results
type EnhancedValidationResult struct {
	// Overall assessment
	IsAccurate      bool              `json:"is_accurate"`
	ConfidenceScore *AnswerConfidence `json:"confidence_score"`
	QualityLevel    string            `json:"quality_level"` // "excellent", "good", "fair", "poor"

	// Source verification
	SourceVerification  *nixos.VerificationResult `json:"source_verification"`
	CrossReferenceCheck *CrossReferenceResult     `json:"cross_reference_check"`

	// Component validations
	PreAnswerValidation *FactualValidationResult    `json:"pre_answer_validation"`
	NixOSValidation     *NixOSValidationResult      `json:"nixos_validation"`
	FlakeValidation     *FlakeValidationResult      `json:"flake_validation"`
	CommunityValidation *CommunityValidationResult  `json:"community_validation"`
	ToolValidation      *nixos.ToolValidationResult `json:"tool_validation"`

	// Meta information
	ValidationTime   time.Duration  `json:"validation_time"`
	SourcesConsulted []string       `json:"sources_consulted"`
	QualityIssues    []QualityIssue `json:"quality_issues"`
	Recommendations  []string       `json:"recommendations"`

	// Enhanced validation with automated quality scoring
	AutomatedQualityScore *AutomatedQualityScore `json:"automated_quality_score,omitempty"`
}

// AnswerConfidence represents a comprehensive confidence scoring system
type AnswerConfidence struct {
	SourceVerification float64 `json:"source_verification"` // 0-1: How many sources confirm this?
	Recency            float64 `json:"recency"`             // 0-1: How recent is the information?
	CommunityConsensus float64 `json:"community_consensus"` // 0-1: Do community sources agree?
	ToolVerification   float64 `json:"tool_verification"`   // 0-1: Do local tools confirm this?
	SyntaxValidity     float64 `json:"syntax_validity"`     // 0-1: Is the syntax correct?

	Overall float64 `json:"overall"` // Weighted average
}

// QualityIssue represents specific quality concerns found during validation
type QualityIssue struct {
	Type       string `json:"type"`     // "syntax", "factual", "outdated", "inconsistent"
	Severity   string `json:"severity"` // "low", "medium", "high", "critical"
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Source     string `json:"source"` // Which validator found this issue
}

// NewEnhancedValidator creates a new enhanced validator with all components
func NewEnhancedValidator(mcpHost string, mcpPort int, githubToken string, logger *logger.Logger) *EnhancedValidator {
	// Initialize existing validators
	preValidator := NewPreAnswerValidator(mcpHost, mcpPort, githubToken)
	nixosValidator := NewNixOSValidator()
	flakeValidator := NewFlakeValidator()

	// Initialize new components
	searchNixOSClient := nixos.NewSearchNixOSClient()
	nixToolsExecutor := nixos.NewToolsExecutor()
	communityValidator := NewCommunityValidator(githubToken)
	factChecker := NewFactChecker()
	crossRefValidator := NewCrossReferenceValidator()
	qualityMetrics := NewQualityMetrics()
	confidenceScorer := NewConfidenceScorer()
	automatedScorer := NewAutomatedQualityScorer()

	return &EnhancedValidator{
		preValidator:       preValidator,
		nixosValidator:     nixosValidator,
		flakeValidator:     flakeValidator,
		searchNixOSClient:  searchNixOSClient,
		nixToolsExecutor:   nixToolsExecutor,
		communityValidator: communityValidator,
		factChecker:        factChecker,
		crossRefValidator:  crossRefValidator,
		qualityMetrics:     qualityMetrics,
		confidenceScorer:   confidenceScorer,
		automatedScorer:    automatedScorer,
		logger:             logger,
	}
}

// ValidateAnswer performs comprehensive validation of an AI-generated answer
func (ev *EnhancedValidator) ValidateAnswer(ctx context.Context, question, answer string) (*EnhancedValidationResult, error) {
	startTime := time.Now()

	result := &EnhancedValidationResult{
		IsAccurate:       true,
		QualityLevel:     "excellent",
		SourcesConsulted: []string{},
		QualityIssues:    []QualityIssue{},
		Recommendations:  []string{},
	}

	// Step 1: Pre-answer validation (what we should have known beforehand)
	if ev.preValidator != nil {
		preResult, err := ev.preValidator.ValidateQuestionFactually(ctx, question)
		if err != nil {
			ev.logger.Println("Pre-answer validation failed:", err)
		} else {
			result.PreAnswerValidation = preResult
			result.SourcesConsulted = append(result.SourcesConsulted, "pre-answer-validation")
		}
	}

	// Step 2: Content-specific validation
	if IsNixOSContent(answer) {
		nixosResult := ev.nixosValidator.ValidateNixOSContent(answer)
		result.NixOSValidation = nixosResult
		result.SourcesConsulted = append(result.SourcesConsulted, "nixos-validator")

		if !nixosResult.IsValid {
			result.IsAccurate = false
			for _, err := range nixosResult.Errors {
				result.QualityIssues = append(result.QualityIssues, QualityIssue{
					Type:       "nixos-config",
					Severity:   "high",
					Message:    err.Message,
					Suggestion: err.Suggestion,
					Source:     "nixos-validator",
				})
			}
		}
	}

	if IsFlakeContent(answer) {
		flakeResult := ev.flakeValidator.ValidateFlakeContent(answer)
		result.FlakeValidation = flakeResult
		result.SourcesConsulted = append(result.SourcesConsulted, "flake-validator")

		if !flakeResult.IsValid {
			result.IsAccurate = false
			for _, err := range flakeResult.Errors {
				result.QualityIssues = append(result.QualityIssues, QualityIssue{
					Type:       "flake-syntax",
					Severity:   "high", // Default severity since err.Severity doesn't exist
					Message:    err.Message,
					Suggestion: err.Suggestion,
					Source:     "flake-validator",
				})
			}
		}
	}

	// Step 3: Real-time tool validation
	if ev.nixToolsExecutor != nil {
		toolResult, err := ev.nixToolsExecutor.ValidateAnswer(ctx, answer)
		if err != nil {
			ev.logger.Println("Tool validation failed:", err)
		} else {
			result.ToolValidation = toolResult
			result.SourcesConsulted = append(result.SourcesConsulted, "nix-tools")
		}
	}

	// Step 4: Community consensus validation
	if ev.communityValidator != nil {
		communityResult, err := ev.communityValidator.ValidateAgainstCommunity(ctx, question, answer)
		if err != nil {
			ev.logger.Println("Community validation failed:", err)
		} else {
			result.CommunityValidation = communityResult
			result.SourcesConsulted = append(result.SourcesConsulted, "community-sources")
		}
	}

	// Step 5: Search.nixos.org verification
	if ev.searchNixOSClient != nil {
		sourceVerification, err := ev.searchNixOSClient.VerifyAnswer(ctx, answer)
		if err != nil {
			ev.logger.Println("Search.nixos.org verification failed:", err)
		} else {
			result.SourceVerification = sourceVerification
			result.SourcesConsulted = append(result.SourcesConsulted, "search-nixos-org")
		}
	}

	// Step 6: Cross-reference validation
	if ev.crossRefValidator != nil {
		crossRefResult := ev.crossRefValidator.ValidateConsistency(ctx, question, answer, result)
		result.CrossReferenceCheck = crossRefResult
		result.SourcesConsulted = append(result.SourcesConsulted, "cross-reference")
	}

	// Step 7: Calculate comprehensive confidence score
	if ev.confidenceScorer != nil {
		confidenceScore := ev.confidenceScorer.CalculateConfidence(result)
		result.ConfidenceScore = confidenceScore
	}

	// Step 8: Perform automated quality scoring using local Nix commands
	if ev.automatedScorer != nil {
		qualityScore, err := ev.automatedScorer.ScoreAnswer(ctx, question, answer)
		if err != nil {
			ev.logger.Printf("Automated quality scoring failed: %v", err)
		} else {
			result.AutomatedQualityScore = qualityScore
			result.SourcesConsulted = append(result.SourcesConsulted, "automated-quality-scorer")

			// Incorporate automated quality issues into overall result
			for _, issue := range qualityScore.Issues {
				result.QualityIssues = append(result.QualityIssues, QualityIssue{
					Type:       issue.Category,
					Severity:   issue.Severity,
					Message:    issue.Message,
					Suggestion: issue.Suggestion,
					Source:     "automated-quality-scorer",
				})
			}

			// Only mark as inaccurate if there are critical validation failures
			// Automated score affects quality level, not accuracy per se
			criticalFailures := 0
			for _, issue := range qualityScore.Issues {
				if issue.Severity == "critical" {
					criticalFailures++
				}
			}
			if criticalFailures > 0 {
				result.IsAccurate = false
			}
		}
	}

	// Step 9: Determine overall quality level
	result.QualityLevel = ev.determineQualityLevel(result)

	// Step 10: Generate recommendations
	result.Recommendations = ev.generateRecommendations(result)

	result.ValidationTime = time.Since(startTime)

	// Log validation summary
	ev.logger.Printf("Enhanced validation completed - Quality: %s, Confidence: %.2f, Sources: %d, Issues: %d, Time: %v",
		result.QualityLevel,
		result.ConfidenceScore.Overall,
		len(result.SourcesConsulted),
		len(result.QualityIssues),
		result.ValidationTime,
	)

	return result, nil
}

// ValidateQuestionPreAnswer performs pre-answer validation to guide response generation
func (ev *EnhancedValidator) ValidateQuestionPreAnswer(ctx context.Context, question string) (*FactualValidationResult, error) {
	if ev.preValidator == nil {
		return nil, nil
	}
	return ev.preValidator.ValidateQuestionFactually(ctx, question)
}

// determineQualityLevel determines the overall quality level based on validation results
func (ev *EnhancedValidator) determineQualityLevel(result *EnhancedValidationResult) string {
	if !result.IsAccurate {
		return "poor"
	}

	confidence := 0.5 // Default confidence if not set
	if result.ConfidenceScore != nil {
		confidence = result.ConfidenceScore.Overall
	}

	automatedScore := 50 // Default score if not set
	if result.AutomatedQualityScore != nil {
		automatedScore = result.AutomatedQualityScore.OverallScore
	}

	criticalIssues := 0
	highIssues := 0

	for _, issue := range result.QualityIssues {
		switch issue.Severity {
		case "critical":
			criticalIssues++
		case "high":
			highIssues++
		}
	}

	// Critical issues always result in poor quality
	if criticalIssues > 0 {
		return "poor"
	}

	// Combine confidence score with automated quality score
	combinedScore := (confidence*100 + float64(automatedScore)) / 2

	// Quality determination based on combined metrics
	if highIssues > 2 || combinedScore < 50 || automatedScore < 40 {
		return "poor"
	}

	if highIssues > 1 || combinedScore < 70 || automatedScore < 60 {
		return "fair"
	}

	if combinedScore >= 90 && automatedScore >= 85 && highIssues == 0 {
		return "excellent"
	}

	return "good"
}

// generateRecommendations generates actionable recommendations based on validation results
func (ev *EnhancedValidator) generateRecommendations(result *EnhancedValidationResult) []string {
	var recommendations []string

	if result.ConfidenceScore != nil && result.ConfidenceScore.Overall < 0.7 {
		recommendations = append(recommendations, "Consider seeking additional verification from official NixOS documentation")
	}

	if len(result.QualityIssues) > 0 {
		recommendations = append(recommendations, "Review and address the identified quality issues before implementing")
	}

	if result.ToolValidation != nil && len(result.ToolValidation.FailedChecks) > 0 {
		recommendations = append(recommendations, "Verify the suggested commands work in your specific NixOS environment")
	}

	if result.SourceVerification != nil && result.SourceVerification.PackageVerificationFailed {
		recommendations = append(recommendations, "Double-check package names and availability in your NixOS channel")
	}

	// Add automated quality score recommendations
	if result.AutomatedQualityScore != nil {
		automatedScore := result.AutomatedQualityScore

		if automatedScore.OverallScore < 70 {
			recommendations = append(recommendations, "The automated validation suggests several improvements are needed")
		}

		if automatedScore.BreakdownScores.SyntaxScore < 20 {
			recommendations = append(recommendations, "Syntax validation failed - check Nix expression syntax and formatting")
		}

		if automatedScore.BreakdownScores.PackageScore < 15 {
			recommendations = append(recommendations, "Package verification failed - ensure all referenced packages exist and are spelled correctly")
		}

		if automatedScore.BreakdownScores.OptionScore < 15 {
			recommendations = append(recommendations, "Option validation failed - verify NixOS configuration options are valid and properly formatted")
		}

		if automatedScore.BreakdownScores.CommandScore < 7 {
			recommendations = append(recommendations, "Command availability check failed - ensure all referenced commands are available on the system")
		}

		// Include specific automated recommendations
		recommendations = append(recommendations, automatedScore.Recommendations...)
	}

	return recommendations
}

// FormatValidationResult formats the enhanced validation result for user display
func (ev *EnhancedValidator) FormatValidationResult(result *EnhancedValidationResult) string {
	if result == nil {
		return ""
	}

	// This will be implemented to create a beautiful, informative display
	// following the UX design from the plan
	return ev.formatEnhancedValidationDisplay(result)
}

// formatEnhancedValidationDisplay creates the enhanced answer format with confidence indicators
func (ev *EnhancedValidator) formatEnhancedValidationDisplay(result *EnhancedValidationResult) string {
	// Implementation will follow the UX design from the enhanced-accuracy-plan.md
	// This includes confidence indicators, source attribution, and quality metrics
	return ""
}
