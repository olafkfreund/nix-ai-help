package validation

import (
	"fmt"
	"math"
	"time"

	"nix-ai-help/pkg/logger"
)

// ConfidenceScorer calculates comprehensive confidence scores for answers
type ConfidenceScorer struct {
	logger *logger.Logger
}

// NewConfidenceScorer creates a new confidence scorer
func NewConfidenceScorer() *ConfidenceScorer {
	return &ConfidenceScorer{
		logger: logger.NewLogger(),
	}
}

// CalculateConfidence calculates a comprehensive confidence score based on all validation results
func (cs *ConfidenceScorer) CalculateConfidence(result *EnhancedValidationResult) *AnswerConfidence {
	confidence := &AnswerConfidence{
		SourceVerification: 0.5,
		Recency:            0.5,
		CommunityConsensus: 0.5,
		ToolVerification:   0.5,
		SyntaxValidity:     0.5,
		Overall:            0.5,
	}

	// Calculate source verification score
	confidence.SourceVerification = cs.calculateSourceVerificationScore(result)

	// Calculate recency score
	confidence.Recency = cs.calculateRecencyScore(result)

	// Calculate community consensus score
	confidence.CommunityConsensus = cs.calculateCommunityConsensusScore(result)

	// Calculate tool verification score
	confidence.ToolVerification = cs.calculateToolVerificationScore(result)

	// Calculate syntax validity score
	confidence.SyntaxValidity = cs.calculateSyntaxValidityScore(result)

	// Calculate overall confidence (weighted average)
	confidence.Overall = cs.calculateOverallConfidence(confidence)

	cs.logger.Printf("Confidence calculated - source_verification: %.2f, recency: %.2f, community_consensus: %.2f, tool_verification: %.2f, syntax_validity: %.2f, overall: %.2f",
		confidence.SourceVerification,
		confidence.Recency,
		confidence.CommunityConsensus,
		confidence.ToolVerification,
		confidence.SyntaxValidity,
		confidence.Overall,
	)

	return confidence
}

// calculateSourceVerificationScore calculates confidence based on source verification
func (cs *ConfidenceScorer) calculateSourceVerificationScore(result *EnhancedValidationResult) float64 {
	score := 0.5 // Base score

	// Pre-answer validation contribution
	if result.PreAnswerValidation != nil {
		switch result.PreAnswerValidation.ConfidenceLevel {
		case "high":
			score += 0.3
		case "medium":
			score += 0.1
		case "low":
			score -= 0.1
		}

		// Bonus for verified sources
		sourceBonus := float64(len(result.PreAnswerValidation.VerifiedSources)) * 0.05
		if sourceBonus > 0.2 {
			sourceBonus = 0.2 // Cap bonus at 0.2
		}
		score += sourceBonus
	}

	// Source verification contribution
	if result.SourceVerification != nil {
		score += result.SourceVerification.Confidence * 0.3

		// Penalty for failed verifications
		if result.SourceVerification.PackageVerificationFailed {
			score -= 0.1
		}
		if result.SourceVerification.OptionVerificationFailed {
			score -= 0.15
		}

		// Bonus for verified packages and options
		verifiedItems := len(result.SourceVerification.PackagesVerified) + len(result.SourceVerification.OptionsVerified)
		if verifiedItems > 0 {
			score += float64(verifiedItems) * 0.02
		}
	}

	return cs.clampScore(score)
}

// calculateRecencyScore calculates confidence based on information recency
func (cs *ConfidenceScorer) calculateRecencyScore(result *EnhancedValidationResult) float64 {
	score := 0.7 // Assume moderately recent by default

	// Check for deprecated patterns or commands
	if result.NixOSValidation != nil {
		deprecatedCount := 0
		for _, err := range result.NixOSValidation.Errors {
			if err.Type == "deprecated_command" || err.Type == "deprecated_option" {
				deprecatedCount++
			}
		}

		// Penalty for deprecated content
		deprecatedPenalty := float64(deprecatedCount) * 0.2
		score -= deprecatedPenalty
	}

	// Check validation timestamps
	if result.PreAnswerValidation != nil {
		// If validation was recent, give bonus
		if result.PreAnswerValidation.ValidationTime < time.Minute {
			score += 0.1
		}
	}

	// Bonus for tool verification (indicates current system compatibility)
	if result.ToolValidation != nil && result.ToolValidation.Confidence > 0.7 {
		score += 0.2
	}

	return cs.clampScore(score)
}

// calculateCommunityConsensusScore calculates confidence based on community consensus
func (cs *ConfidenceScorer) calculateCommunityConsensusScore(result *EnhancedValidationResult) float64 {
	if result.CommunityValidation == nil {
		return 0.5 // Neutral when no community data
	}

	score := result.CommunityValidation.CommunityConsensus

	// Bonus for community best practices
	if len(result.CommunityValidation.BestPractices) > 0 {
		practiceBonus := float64(len(result.CommunityValidation.BestPractices)) * 0.05
		if practiceBonus > 0.2 {
			practiceBonus = 0.2
		}
		score += practiceBonus
	}

	// Penalty for common gotchas
	if len(result.CommunityValidation.CommonGotchas) > 0 {
		gotchaPenalty := float64(len(result.CommunityValidation.CommonGotchas)) * 0.1
		score -= gotchaPenalty
	}

	// Bonus for wiki validation
	if result.CommunityValidation.WikiValidation != nil {
		wikiConfidence := result.CommunityValidation.WikiValidation.Confidence
		score = (score + wikiConfidence) / 2.0 // Average with wiki confidence
	}

	// Bonus for GitHub validation
	if result.CommunityValidation.GitHubValidation != nil {
		githubConfidence := result.CommunityValidation.GitHubValidation.Confidence
		score = (score*0.7 + githubConfidence*0.3) // Weighted average
	}

	return cs.clampScore(score)
}

// calculateToolVerificationScore calculates confidence based on tool verification
func (cs *ConfidenceScorer) calculateToolVerificationScore(result *EnhancedValidationResult) float64 {
	if result.ToolValidation == nil {
		return 0.5 // Neutral when no tool data
	}

	score := result.ToolValidation.Confidence

	// Detailed analysis of tool validation results
	totalChecks := len(result.ToolValidation.PackageChecks) +
		len(result.ToolValidation.OptionChecks) +
		len(result.ToolValidation.SyntaxChecks) +
		len(result.ToolValidation.CommandChecks)

	if totalChecks > 0 {
		successfulChecks := len(result.ToolValidation.SuccessfulChecks)
		failedChecks := len(result.ToolValidation.FailedChecks)

		if successfulChecks+failedChecks > 0 {
			toolSuccessRate := float64(successfulChecks) / float64(successfulChecks+failedChecks)
			score = (score + toolSuccessRate) / 2.0 // Average with computed success rate
		}

		// Bonus for having many successful verifications
		if successfulChecks >= 3 {
			score += 0.1
		}

		// Penalty for critical failures
		criticalFailures := 0
		for _, check := range result.ToolValidation.OptionChecks {
			if !check.Valid {
				criticalFailures++
			}
		}

		if criticalFailures > 0 {
			score -= float64(criticalFailures) * 0.15
		}
	}

	// Bonus for syntax validation
	allSyntaxValid := true
	for _, syntaxCheck := range result.ToolValidation.SyntaxChecks {
		if !syntaxCheck.Valid {
			allSyntaxValid = false
			break
		}
	}

	if allSyntaxValid && len(result.ToolValidation.SyntaxChecks) > 0 {
		score += 0.1
	}

	return cs.clampScore(score)
}

// calculateSyntaxValidityScore calculates confidence based on syntax validation
func (cs *ConfidenceScorer) calculateSyntaxValidityScore(result *EnhancedValidationResult) float64 {
	score := 0.8 // Assume good syntax by default

	// NixOS validation syntax issues
	if result.NixOSValidation != nil {
		if !result.NixOSValidation.IsValid {
			syntaxErrors := 0
			for _, err := range result.NixOSValidation.Errors {
				if err.Type == "syntax_error" || err.Type == "incorrect_option_name" {
					syntaxErrors++
				}
			}

			if syntaxErrors > 0 {
				score -= float64(syntaxErrors) * 0.2
			}
		}

		// Penalty for severity
		switch result.NixOSValidation.Severity {
		case "critical":
			score -= 0.4
		case "high":
			score -= 0.3
		case "medium":
			score -= 0.1
		}
	}

	// Flake validation syntax issues
	if result.FlakeValidation != nil {
		if !result.FlakeValidation.IsValid {
			syntaxErrors := 0
			for _, err := range result.FlakeValidation.Errors {
				if err.Type == "syntax_error" || err.Type == "incorrect_structure" {
					syntaxErrors++
				}
			}

			if syntaxErrors > 0 {
				score -= float64(syntaxErrors) * 0.25
			}
		}
	}

	// Tool validation syntax checks
	if result.ToolValidation != nil {
		for _, syntaxCheck := range result.ToolValidation.SyntaxChecks {
			if !syntaxCheck.Valid {
				score -= 0.2
			}
		}
	}

	return cs.clampScore(score)
}

// calculateOverallConfidence calculates the overall confidence as a weighted average
func (cs *ConfidenceScorer) calculateOverallConfidence(confidence *AnswerConfidence) float64 {
	// Define weights for different aspects of confidence
	weights := map[string]float64{
		"source_verification": 0.25,
		"tool_verification":   0.25,
		"syntax_validity":     0.20,
		"community_consensus": 0.20,
		"recency":             0.10,
	}

	overall := confidence.SourceVerification*weights["source_verification"] +
		confidence.ToolVerification*weights["tool_verification"] +
		confidence.SyntaxValidity*weights["syntax_validity"] +
		confidence.CommunityConsensus*weights["community_consensus"] +
		confidence.Recency*weights["recency"]

	return cs.clampScore(overall)
}

// CalculateConfidenceLevel converts numerical confidence to categorical level
func (cs *ConfidenceScorer) CalculateConfidenceLevel(overallConfidence float64) string {
	switch {
	case overallConfidence >= 0.9:
		return "excellent"
	case overallConfidence >= 0.8:
		return "high"
	case overallConfidence >= 0.6:
		return "medium"
	case overallConfidence >= 0.4:
		return "low"
	default:
		return "very-low"
	}
}

// CalculateConfidenceFactors analyzes which factors contribute most to confidence
func (cs *ConfidenceScorer) CalculateConfidenceFactors(confidence *AnswerConfidence) map[string]string {
	factors := make(map[string]string)

	// Analyze strongest contributing factors
	scores := map[string]float64{
		"Source Verification": confidence.SourceVerification,
		"Tool Verification":   confidence.ToolVerification,
		"Syntax Validity":     confidence.SyntaxValidity,
		"Community Consensus": confidence.CommunityConsensus,
		"Recency":             confidence.Recency,
	}

	// Find highest and lowest scoring factors
	var highestFactor, lowestFactor string
	var highestScore, lowestScore float64 = 0.0, 1.0

	for factor, score := range scores {
		if score > highestScore {
			highestScore = score
			highestFactor = factor
		}
		if score < lowestScore {
			lowestScore = score
			lowestFactor = factor
		}
	}

	factors["strongest"] = highestFactor
	factors["weakest"] = lowestFactor

	// Categorize overall confidence pattern
	if confidence.ToolVerification > 0.8 && confidence.SyntaxValidity > 0.8 {
		factors["pattern"] = "technically-verified"
	} else if confidence.CommunityConsensus > 0.8 && confidence.SourceVerification > 0.7 {
		factors["pattern"] = "community-endorsed"
	} else if confidence.SourceVerification > 0.8 {
		factors["pattern"] = "documentation-backed"
	} else {
		factors["pattern"] = "mixed-signals"
	}

	return factors
}

// clampScore ensures score stays within 0.0 to 1.0 range
func (cs *ConfidenceScorer) clampScore(score float64) float64 {
	return math.Max(0.0, math.Min(1.0, score))
}

// CalculateConfidenceChange compares confidence scores over time
func (cs *ConfidenceScorer) CalculateConfidenceChange(previous, current *AnswerConfidence) map[string]float64 {
	changes := make(map[string]float64)

	if previous == nil {
		return changes
	}

	changes["source_verification"] = current.SourceVerification - previous.SourceVerification
	changes["tool_verification"] = current.ToolVerification - previous.ToolVerification
	changes["syntax_validity"] = current.SyntaxValidity - previous.SyntaxValidity
	changes["community_consensus"] = current.CommunityConsensus - previous.CommunityConsensus
	changes["recency"] = current.Recency - previous.Recency
	changes["overall"] = current.Overall - previous.Overall

	return changes
}

// GetConfidenceExplanation provides a human-readable explanation of the confidence score
func (cs *ConfidenceScorer) GetConfidenceExplanation(confidence *AnswerConfidence) string {
	level := cs.CalculateConfidenceLevel(confidence.Overall)
	factors := cs.CalculateConfidenceFactors(confidence)

	explanations := map[string]string{
		"excellent": "Exceptional confidence - answer is well-verified across multiple sources with strong technical validation.",
		"high":      "High confidence - answer is supported by reliable sources and passes most validation checks.",
		"medium":    "Moderate confidence - answer has good foundation but may need additional verification.",
		"low":       "Low confidence - answer has some issues or limited verification. Consider alternative sources.",
		"very-low":  "Very low confidence - answer has significant issues and should be used with caution.",
	}

	explanation := explanations[level]

	// Add factor-specific details
	if factors["strongest"] != "" {
		explanation += fmt.Sprintf(" Strongest validation: %s.", factors["strongest"])
	}

	if factors["weakest"] != "" && cs.getFactorScore(factors["weakest"], confidence) < 0.5 {
		explanation += fmt.Sprintf(" Area for improvement: %s.", factors["weakest"])
	}

	return explanation
}

// getFactorScore gets the score for a specific factor name
func (cs *ConfidenceScorer) getFactorScore(factorName string, confidence *AnswerConfidence) float64 {
	switch factorName {
	case "Source Verification":
		return confidence.SourceVerification
	case "Tool Verification":
		return confidence.ToolVerification
	case "Syntax Validity":
		return confidence.SyntaxValidity
	case "Community Consensus":
		return confidence.CommunityConsensus
	case "Recency":
		return confidence.Recency
	default:
		return 0.5
	}
}
