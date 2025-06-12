package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// CrossReferenceValidator performs cross-validation between multiple sources
type CrossReferenceValidator struct {
	logger *logger.Logger
}

// CrossReferenceResult represents the result of cross-reference validation
type CrossReferenceResult struct {
	ConsistencyScore   float64                          `json:"consistency_score"`
	Contradictions     []Contradiction                  `json:"contradictions"`
	Confirmations      []Confirmation                   `json:"confirmations"`
	SourceAgreement    map[string]float64               `json:"source_agreement"`
	QualityAssessment  *CrossReferenceQualityAssessment `json:"quality_assessment"`
	RecommendedSources []string                         `json:"recommended_sources"`
	ValidationTime     time.Duration                    `json:"validation_time"`
}

// Contradiction represents a contradiction found between sources
type Contradiction struct {
	Type        string `json:"type"` // "option", "package", "syntax", "approach"
	Description string `json:"description"`
	Source1     string `json:"source1"`
	Source2     string `json:"source2"`
	Evidence1   string `json:"evidence1"`
	Evidence2   string `json:"evidence2"`
	Severity    string `json:"severity"` // "low", "medium", "high", "critical"
	Resolution  string `json:"resolution"`
}

// Confirmation represents confirmation of information across sources
type Confirmation struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Sources     []string `json:"sources"`
	Evidence    []string `json:"evidence"`
	Confidence  float64  `json:"confidence"`
}

// CrossReferenceQualityAssessment represents an overall quality assessment for cross-reference validation
type CrossReferenceQualityAssessment struct {
	Accuracy     float64 `json:"accuracy"`     // 0-1: How accurate is the information?
	Completeness float64 `json:"completeness"` // 0-1: How complete is the answer?
	Clarity      float64 `json:"clarity"`      // 0-1: How clear is the explanation?
	Practicality float64 `json:"practicality"` // 0-1: How practical is the solution?
	UpToDate     float64 `json:"up_to_date"`   // 0-1: How current is the information?
	Overall      float64 `json:"overall"`      // Weighted average
}

// ValidationRule represents a rule for cross-reference validation
type ValidationRule struct {
	Name      string
	Type      string // "consistency", "contradiction", "best_practice"
	Pattern   *regexp.Regexp
	Validator func(context.Context, string, *EnhancedValidationResult) []ValidationIssue
	Weight    float64
	Enabled   bool
}

// ValidationIssue represents an issue found during validation
type ValidationIssue struct {
	Type        string
	Severity    string
	Description string
	Evidence    string
	Suggestion  string
	Source      string
}

// SourceResult represents the result from a specific validation source
type SourceResult struct {
	Source      string
	Content     string
	Confidence  float64
	Verified    bool
	LastUpdated time.Time
}

// NewCrossReferenceValidator creates a new cross-reference validator
func NewCrossReferenceValidator() *CrossReferenceValidator {
	return &CrossReferenceValidator{
		logger: logger.NewLogger(),
	}
}

// ValidateConsistency performs cross-reference validation across all sources
func (crv *CrossReferenceValidator) ValidateConsistency(ctx context.Context, question, answer string, result *EnhancedValidationResult) *CrossReferenceResult {
	startTime := time.Now()

	crossRefResult := &CrossReferenceResult{
		ConsistencyScore:   1.0,
		Contradictions:     []Contradiction{},
		Confirmations:      []Confirmation{},
		SourceAgreement:    make(map[string]float64),
		RecommendedSources: []string{},
	}

	// Check for contradictions between different validation sources
	crv.checkSourceContradictions(result, crossRefResult)

	// Find confirmations across sources
	crv.findSourceConfirmations(result, crossRefResult)

	// Calculate source agreement scores
	crv.calculateSourceAgreement(result, crossRefResult)

	// Perform quality assessment
	crossRefResult.QualityAssessment = crv.performQualityAssessment(question, answer, result)

	// Calculate overall consistency score
	crossRefResult.ConsistencyScore = crv.calculateConsistencyScore(crossRefResult)

	// Generate source recommendations
	crossRefResult.RecommendedSources = crv.generateSourceRecommendations(result, crossRefResult)

	crossRefResult.ValidationTime = time.Since(startTime)

	crv.logger.Printf("Cross-reference validation completed - consistency_score: %.2f, contradictions: %d, confirmations: %d, validation_time: %v",
		crossRefResult.ConsistencyScore,
		len(crossRefResult.Contradictions),
		len(crossRefResult.Confirmations),
		crossRefResult.ValidationTime,
	)

	return crossRefResult
}

// checkSourceContradictions identifies contradictions between different validation sources
func (crv *CrossReferenceValidator) checkSourceContradictions(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	// Check NixOS vs Tool validation contradictions
	if result.NixOSValidation != nil && result.ToolValidation != nil {
		crv.checkNixOSvsToolContradictions(result, crossRefResult)
	}

	// Check Pre-answer vs Community validation contradictions
	if result.PreAnswerValidation != nil && result.CommunityValidation != nil {
		crv.checkPreAnswerVsCommunityContradictions(result, crossRefResult)
	}

	// Check Source verification vs Tool validation contradictions
	if result.SourceVerification != nil && result.ToolValidation != nil {
		crv.checkSourceVsToolContradictions(result, crossRefResult)
	}

	// Check for internal contradictions within the answer itself
	crv.checkInternalContradictions(result, crossRefResult)
}

// checkNixOSvsToolContradictions checks for contradictions between NixOS validation and tool validation
func (crv *CrossReferenceValidator) checkNixOSvsToolContradictions(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	nixosResult := result.NixOSValidation
	toolResult := result.ToolValidation

	// Check package contradictions
	for _, packageCheck := range toolResult.PackageChecks {
		if !packageCheck.Available {
			// If tool says package is not available, but NixOS validator didn't catch it
			if !crv.packageMentionedInNixOSErrors(packageCheck.PackageName, nixosResult) {
				contradiction := Contradiction{
					Type:        "package",
					Description: fmt.Sprintf("Package '%s' not found by nix tools but not flagged by NixOS validator", packageCheck.PackageName),
					Source1:     "nix-tools",
					Source2:     "nixos-validator",
					Evidence1:   fmt.Sprintf("nix search failed for '%s'", packageCheck.PackageName),
					Evidence2:   "No errors reported for this package",
					Severity:    "medium",
					Resolution:  "Verify package name and availability in current channel",
				}
				crossRefResult.Contradictions = append(crossRefResult.Contradictions, contradiction)
			}
		}
	}

	// Check option contradictions
	for _, optionCheck := range toolResult.OptionChecks {
		if !optionCheck.Valid {
			// If tool says option is invalid, but NixOS validator didn't catch it
			if !crv.optionMentionedInNixOSErrors(optionCheck.OptionName, nixosResult) {
				contradiction := Contradiction{
					Type:        "option",
					Description: fmt.Sprintf("Option '%s' not found by nixos-option but not flagged by NixOS validator", optionCheck.OptionName),
					Source1:     "nixos-option",
					Source2:     "nixos-validator",
					Evidence1:   fmt.Sprintf("nixos-option failed for '%s'", optionCheck.OptionName),
					Evidence2:   "No errors reported for this option",
					Severity:    "high",
					Resolution:  "Verify option name and check for typos",
				}
				crossRefResult.Contradictions = append(crossRefResult.Contradictions, contradiction)
			}
		}
	}
}

// checkPreAnswerVsCommunityContradictions checks for contradictions between pre-answer and community validation
func (crv *CrossReferenceValidator) checkPreAnswerVsCommunityContradictions(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	preResult := result.PreAnswerValidation
	communityResult := result.CommunityValidation

	if preResult.ConfidenceLevel == "high" && communityResult.CommunityConsensus < 0.5 {
		contradiction := Contradiction{
			Type:        "approach",
			Description: "High confidence from documentation sources but low community consensus",
			Source1:     "pre-answer-validation",
			Source2:     "community-validation",
			Evidence1:   fmt.Sprintf("Confidence level: %s with %d verified sources", preResult.ConfidenceLevel, len(preResult.VerifiedSources)),
			Evidence2:   fmt.Sprintf("Community consensus: %.2f", communityResult.CommunityConsensus),
			Severity:    "medium",
			Resolution:  "Consider community best practices alongside official documentation",
		}
		crossRefResult.Contradictions = append(crossRefResult.Contradictions, contradiction)
	}
}

// checkSourceVsToolContradictions checks for contradictions between source verification and tool validation
func (crv *CrossReferenceValidator) checkSourceVsToolContradictions(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	sourceResult := result.SourceVerification
	toolResult := result.ToolValidation

	// Check if official sources say packages exist but tools can't find them
	for _, verifiedPkg := range sourceResult.PackagesVerified {
		toolFoundPkg := false
		for _, toolPkg := range toolResult.PackageChecks {
			if toolPkg.PackageName == verifiedPkg.Name && toolPkg.Available {
				toolFoundPkg = true
				break
			}
		}

		if !toolFoundPkg {
			contradiction := Contradiction{
				Type:        "package",
				Description: fmt.Sprintf("Package '%s' verified by search.nixos.org but not available locally", verifiedPkg.Name),
				Source1:     "search-nixos-org",
				Source2:     "nix-tools",
				Evidence1:   fmt.Sprintf("Package found in official repository: %s", verifiedPkg.Description),
				Evidence2:   "Package not found by local nix search",
				Severity:    "medium",
				Resolution:  "Check NixOS channel version or update package index",
			}
			crossRefResult.Contradictions = append(crossRefResult.Contradictions, contradiction)
		}
	}
}

// checkInternalContradictions checks for contradictions within the answer itself
func (crv *CrossReferenceValidator) checkInternalContradictions(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	// Check for contradictory recommendations in the same answer
	if result.NixOSValidation != nil {
		for i, error1 := range result.NixOSValidation.Errors {
			for j, error2 := range result.NixOSValidation.Errors {
				if i != j && crv.areContradictoryErrors(error1, error2) {
					contradiction := Contradiction{
						Type:        "syntax",
						Description: "Contradictory configuration recommendations in the same answer",
						Source1:     "internal-check-1",
						Source2:     "internal-check-2",
						Evidence1:   error1.Message,
						Evidence2:   error2.Message,
						Severity:    "high",
						Resolution:  "Review answer for consistency",
					}
					crossRefResult.Contradictions = append(crossRefResult.Contradictions, contradiction)
				}
			}
		}
	}
}

// findSourceConfirmations identifies confirmations across multiple sources
func (crv *CrossReferenceValidator) findSourceConfirmations(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	// Check for package confirmations
	crv.findPackageConfirmations(result, crossRefResult)

	// Check for option confirmations
	crv.findOptionConfirmations(result, crossRefResult)

	// Check for approach confirmations
	crv.findApproachConfirmations(result, crossRefResult)
}

// findPackageConfirmations finds packages confirmed by multiple sources
func (crv *CrossReferenceValidator) findPackageConfirmations(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	packageSources := make(map[string][]string)
	packageEvidence := make(map[string][]string)

	// Collect packages from source verification
	if result.SourceVerification != nil {
		for _, pkg := range result.SourceVerification.PackagesVerified {
			packageSources[pkg.Name] = append(packageSources[pkg.Name], "search-nixos-org")
			packageEvidence[pkg.Name] = append(packageEvidence[pkg.Name], fmt.Sprintf("Official repository: %s", pkg.Description))
		}
	}

	// Collect packages from tool validation
	if result.ToolValidation != nil {
		for _, pkg := range result.ToolValidation.PackageChecks {
			if pkg.Available {
				packageSources[pkg.PackageName] = append(packageSources[pkg.PackageName], "nix-tools")
				packageEvidence[pkg.PackageName] = append(packageEvidence[pkg.PackageName], "Available via nix search")
			}
		}
	}

	// Create confirmations for packages found by multiple sources
	for pkgName, sources := range packageSources {
		if len(sources) > 1 {
			confirmation := Confirmation{
				Type:        "package",
				Description: fmt.Sprintf("Package '%s' confirmed by multiple sources", pkgName),
				Sources:     sources,
				Evidence:    packageEvidence[pkgName],
				Confidence:  float64(len(sources)) / 3.0, // Normalize based on max expected sources
			}
			crossRefResult.Confirmations = append(crossRefResult.Confirmations, confirmation)
		}
	}
}

// findOptionConfirmations finds options confirmed by multiple sources
func (crv *CrossReferenceValidator) findOptionConfirmations(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	optionSources := make(map[string][]string)
	optionEvidence := make(map[string][]string)

	// Collect options from source verification
	if result.SourceVerification != nil {
		for _, opt := range result.SourceVerification.OptionsVerified {
			optionSources[opt.Name] = append(optionSources[opt.Name], "search-nixos-org")
			optionEvidence[opt.Name] = append(optionEvidence[opt.Name], fmt.Sprintf("Official option: %s", opt.Description))
		}
	}

	// Collect options from tool validation
	if result.ToolValidation != nil {
		for _, opt := range result.ToolValidation.OptionChecks {
			if opt.Valid {
				optionSources[opt.OptionName] = append(optionSources[opt.OptionName], "nixos-option")
				optionEvidence[opt.OptionName] = append(optionEvidence[opt.OptionName], "Valid via nixos-option")
			}
		}
	}

	// Create confirmations for options found by multiple sources
	for optName, sources := range optionSources {
		if len(sources) > 1 {
			confirmation := Confirmation{
				Type:        "option",
				Description: fmt.Sprintf("Option '%s' confirmed by multiple sources", optName),
				Sources:     sources,
				Evidence:    optionEvidence[optName],
				Confidence:  float64(len(sources)) / 2.0, // Normalize based on max expected sources
			}
			crossRefResult.Confirmations = append(crossRefResult.Confirmations, confirmation)
		}
	}
}

// findApproachConfirmations finds approaches confirmed by multiple sources
func (crv *CrossReferenceValidator) findApproachConfirmations(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	// Check if community and documentation sources agree on approach
	if result.PreAnswerValidation != nil && result.CommunityValidation != nil {
		if result.PreAnswerValidation.ConfidenceLevel == "high" && result.CommunityValidation.CommunityConsensus >= 0.7 {
			confirmation := Confirmation{
				Type:        "approach",
				Description: "Solution approach confirmed by both documentation and community sources",
				Sources:     []string{"pre-answer-validation", "community-validation"},
				Evidence: []string{
					fmt.Sprintf("Documentation confidence: %s", result.PreAnswerValidation.ConfidenceLevel),
					fmt.Sprintf("Community consensus: %.2f", result.CommunityValidation.CommunityConsensus),
				},
				Confidence: 0.9,
			}
			crossRefResult.Confirmations = append(crossRefResult.Confirmations, confirmation)
		}
	}
}

// calculateSourceAgreement calculates agreement scores between different sources
func (crv *CrossReferenceValidator) calculateSourceAgreement(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) {
	sources := []string{"nixos-validator", "tool-validation", "community-validation", "source-verification"}

	for _, source := range sources {
		agreement := crv.calculateSourceAgreementScore(source, result, crossRefResult)
		crossRefResult.SourceAgreement[source] = agreement
	}
}

// calculateSourceAgreementScore calculates how well a source agrees with others
func (crv *CrossReferenceValidator) calculateSourceAgreementScore(source string, result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) float64 {
	// Count confirmations and contradictions involving this source
	confirmations := 0
	contradictions := 0

	for _, confirmation := range crossRefResult.Confirmations {
		for _, confirmationSource := range confirmation.Sources {
			if confirmationSource == source {
				confirmations++
				break
			}
		}
	}

	for _, contradiction := range crossRefResult.Contradictions {
		if contradiction.Source1 == source || contradiction.Source2 == source {
			contradictions++
		}
	}

	total := confirmations + contradictions
	if total == 0 {
		return 0.5 // Neutral when no data
	}

	return float64(confirmations) / float64(total)
}

// performQualityAssessment performs a comprehensive quality assessment
func (crv *CrossReferenceValidator) performQualityAssessment(question, answer string, result *EnhancedValidationResult) *CrossReferenceQualityAssessment {
	assessment := &CrossReferenceQualityAssessment{}

	// Calculate accuracy based on validation results
	assessment.Accuracy = crv.calculateAccuracy(result)

	// Calculate completeness based on how well the answer addresses the question
	assessment.Completeness = crv.calculateCompleteness(question, answer, result)

	// Calculate clarity based on structure and explanation quality
	assessment.Clarity = crv.calculateClarity(answer)

	// Calculate practicality based on whether the solution is actionable
	assessment.Practicality = crv.calculatePracticality(answer, result)

	// Calculate how up-to-date the information is
	assessment.UpToDate = crv.calculateRecency(result)

	// Calculate overall score (weighted average)
	assessment.Overall = (assessment.Accuracy*0.3 + assessment.Completeness*0.2 +
		assessment.Clarity*0.2 + assessment.Practicality*0.2 + assessment.UpToDate*0.1)

	return assessment
}

// calculateConsistencyScore calculates the overall consistency score
func (crv *CrossReferenceValidator) calculateConsistencyScore(crossRefResult *CrossReferenceResult) float64 {
	if len(crossRefResult.Contradictions) == 0 && len(crossRefResult.Confirmations) == 0 {
		return 1.0 // Perfect score when no contradictions and no confirmations to check
	}

	confirmationScore := float64(len(crossRefResult.Confirmations))
	contradictionPenalty := 0.0

	for _, contradiction := range crossRefResult.Contradictions {
		switch contradiction.Severity {
		case "critical":
			contradictionPenalty += 1.0
		case "high":
			contradictionPenalty += 0.8
		case "medium":
			contradictionPenalty += 0.5
		case "low":
			contradictionPenalty += 0.2
		}
	}

	totalEvents := confirmationScore + contradictionPenalty
	if totalEvents == 0 {
		return 1.0
	}

	score := confirmationScore / totalEvents
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// generateSourceRecommendations generates recommendations for which sources to trust
func (crv *CrossReferenceValidator) generateSourceRecommendations(result *EnhancedValidationResult, crossRefResult *CrossReferenceResult) []string {
	recommendations := []string{}

	// Recommend sources with high agreement scores
	for source, agreement := range crossRefResult.SourceAgreement {
		if agreement >= 0.8 {
			recommendations = append(recommendations, fmt.Sprintf("High confidence in %s (%.0f%% agreement)", source, agreement*100))
		} else if agreement <= 0.3 {
			recommendations = append(recommendations, fmt.Sprintf("Low confidence in %s (%.0f%% agreement) - verify independently", source, agreement*100))
		}
	}

	// Add specific recommendations based on validation results
	if result.ToolValidation != nil && result.ToolValidation.Confidence >= 0.8 {
		recommendations = append(recommendations, "Local tool validation shows high confidence - solution likely works in current environment")
	}

	if result.CommunityValidation != nil && result.CommunityValidation.CommunityConsensus >= 0.8 {
		recommendations = append(recommendations, "Strong community consensus - solution follows established best practices")
	}

	return recommendations
}

// Helper methods for specific checks

func (crv *CrossReferenceValidator) packageMentionedInNixOSErrors(packageName string, nixosResult *NixOSValidationResult) bool {
	for _, err := range nixosResult.Errors {
		if strings.Contains(err.Message, packageName) || strings.Contains(err.LinePattern, packageName) {
			return true
		}
	}
	return false
}

func (crv *CrossReferenceValidator) optionMentionedInNixOSErrors(optionName string, nixosResult *NixOSValidationResult) bool {
	for _, err := range nixosResult.Errors {
		if strings.Contains(err.Message, optionName) || strings.Contains(err.OptionName, optionName) {
			return true
		}
	}
	return false
}

func (crv *CrossReferenceValidator) areContradictoryErrors(error1, error2 NixOSValidationError) bool {
	// Simple check for contradictory error messages
	return strings.Contains(error1.Message, "enable") && strings.Contains(error2.Message, "disable") ||
		strings.Contains(error1.Message, "use") && strings.Contains(error2.Message, "avoid")
}

func (crv *CrossReferenceValidator) calculateAccuracy(result *EnhancedValidationResult) float64 {
	if !result.IsAccurate {
		return 0.2 // Low accuracy if validation failed
	}

	if result.ConfidenceScore != nil {
		return result.ConfidenceScore.Overall
	}

	return 0.7 // Default moderate accuracy
}

func (crv *CrossReferenceValidator) calculateCompleteness(question, answer string, result *EnhancedValidationResult) float64 {
	// Simple heuristic based on answer length and structure
	answerLength := len(strings.Fields(answer))

	score := 0.5 // Base score

	if answerLength > 50 {
		score += 0.2 // Bonus for detailed answer
	}

	if strings.Contains(answer, "```") {
		score += 0.2 // Bonus for code examples
	}

	if strings.Contains(answer, "configuration.nix") || strings.Contains(answer, "flake.nix") {
		score += 0.1 // Bonus for specific file references
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (crv *CrossReferenceValidator) calculateClarity(answer string) float64 {
	score := 0.5 // Base score

	// Check for clear structure
	if strings.Contains(answer, "##") || strings.Contains(answer, "###") {
		score += 0.2 // Bonus for headers
	}

	if strings.Contains(answer, "1.") || strings.Contains(answer, "- ") {
		score += 0.2 // Bonus for lists
	}

	// Penalize very short or very long answers
	wordCount := len(strings.Fields(answer))
	if wordCount < 20 {
		score -= 0.3 // Penalty for too brief
	} else if wordCount > 500 {
		score -= 0.1 // Small penalty for being too verbose
	}

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

func (crv *CrossReferenceValidator) calculatePracticality(answer string, result *EnhancedValidationResult) float64 {
	score := 0.5 // Base score

	// Check for actionable content
	if strings.Contains(answer, "nixos-rebuild") {
		score += 0.2 // Bonus for rebuild instructions
	}

	if strings.Contains(answer, "enable = true") || strings.Contains(answer, "enable = false") {
		score += 0.2 // Bonus for specific configuration
	}

	// Check if tools can validate the solution
	if result.ToolValidation != nil && result.ToolValidation.Confidence > 0.7 {
		score += 0.2 // Bonus for tool-verifiable solution
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (crv *CrossReferenceValidator) calculateRecency(result *EnhancedValidationResult) float64 {
	// This would ideally check timestamps of sources
	// For now, return moderate score
	score := 0.7

	// Penalize if using deprecated patterns
	if result.NixOSValidation != nil {
		for _, err := range result.NixOSValidation.Errors {
			if err.Type == "deprecated_command" {
				score -= 0.2
			}
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}
