package validation

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"nix-ai-help/pkg/logger"
)

// QualityMetrics evaluates the quality of AI responses across multiple dimensions
type QualityMetrics struct {
	logger *logger.Logger
}

// QualityAssessment represents a comprehensive quality evaluation
type QualityAssessment struct {
	OverallScore    float64            `json:"overall_score"`    // 0.0 to 1.0
	QualityLevel    string             `json:"quality_level"`    // excellent, good, fair, poor
	DimensionScores map[string]float64 `json:"dimension_scores"` // Individual dimension scores
	Strengths       []string           `json:"strengths"`
	Weaknesses      []string           `json:"weaknesses"`
	Suggestions     []string           `json:"suggestions"`
}

// QualityDimensions defines the dimensions of quality assessment
type QualityDimensions struct {
	Clarity      float64 `json:"clarity"`      // How clear and understandable is the response
	Completeness float64 `json:"completeness"` // How complete is the response relative to the question
	Accuracy     float64 `json:"accuracy"`     // Technical accuracy (syntax, facts)
	Practicality float64 `json:"practicality"` // How actionable and practical is the response
	Structure    float64 `json:"structure"`    // How well-structured and organized is the response
}

// NewQualityMetrics creates a new quality metrics evaluator
func NewQualityMetrics() *QualityMetrics {
	return &QualityMetrics{
		logger: logger.NewLogger(),
	}
}

// AssessQuality performs comprehensive quality assessment of an AI response
func (qm *QualityMetrics) AssessQuality(ctx context.Context, question, answer string) (*QualityAssessment, error) {
	qm.logger.Printf("Starting quality assessment for response")

	// Evaluate each dimension
	dimensions := &QualityDimensions{
		Clarity:      qm.assessClarity(answer),
		Completeness: qm.assessCompleteness(question, answer),
		Accuracy:     qm.assessAccuracy(answer),
		Practicality: qm.assessPracticality(answer),
		Structure:    qm.assessStructure(answer),
	}

	// Calculate overall score (weighted average)
	overallScore := qm.calculateOverallScore(dimensions)

	// Determine quality level
	qualityLevel := qm.determineQualityLevel(overallScore)

	// Generate insights
	strengths, weaknesses, suggestions := qm.generateInsights(dimensions, answer)

	assessment := &QualityAssessment{
		OverallScore: overallScore,
		QualityLevel: qualityLevel,
		DimensionScores: map[string]float64{
			"clarity":      dimensions.Clarity,
			"completeness": dimensions.Completeness,
			"accuracy":     dimensions.Accuracy,
			"practicality": dimensions.Practicality,
			"structure":    dimensions.Structure,
		},
		Strengths:   strengths,
		Weaknesses:  weaknesses,
		Suggestions: suggestions,
	}

	qm.logger.Printf("Quality assessment completed with overall score: %.2f (%s)", overallScore, qualityLevel)
	return assessment, nil
}

// assessClarity evaluates how clear and understandable the response is
func (qm *QualityMetrics) assessClarity(answer string) float64 {
	score := 1.0 // Start with perfect score

	// Check for overly complex sentences
	sentences := strings.Split(answer, ".")
	avgSentenceLength := 0
	if len(sentences) > 0 {
		totalLength := len(answer)
		avgSentenceLength = totalLength / len(sentences)
	}

	if avgSentenceLength > 200 {
		score -= 0.3
	} else if avgSentenceLength > 150 {
		score -= 0.2
	} else if avgSentenceLength > 100 {
		score -= 0.1
	}

	// Check for technical jargon without explanation
	jargonTerms := []string{
		"derivation", "closure", "store path", "impure", "IFD",
		"sandbox", "fixed-output", "content-addressed",
	}

	jargonCount := 0
	lowerAnswer := strings.ToLower(answer)
	for _, term := range jargonTerms {
		if strings.Contains(lowerAnswer, term) {
			jargonCount++
		}
	}

	if jargonCount > 3 && !strings.Contains(lowerAnswer, "means") && !strings.Contains(lowerAnswer, "refers to") {
		score -= 0.2
	}

	// Check for clear structure indicators
	if strings.Contains(answer, "Step 1") || strings.Contains(answer, "1.") || strings.Contains(answer, "First,") {
		score += 0.1
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// assessCompleteness evaluates how complete the response is relative to the question
func (qm *QualityMetrics) assessCompleteness(question, answer string) float64 {
	score := 0.5 // Start with neutral score

	questionLower := strings.ToLower(question)
	answerLower := strings.ToLower(answer)

	// Check if the response addresses the main question components
	questionWords := strings.Fields(questionLower)
	answerWords := strings.Fields(answerLower)

	// Create a map for faster lookup
	answerWordMap := make(map[string]bool)
	for _, word := range answerWords {
		if len(word) > 3 {
			answerWordMap[word] = true
		}
	}

	// Count how many question words are addressed
	addressedWords := 0
	importantWords := 0
	for _, word := range questionWords {
		if len(word) > 3 {
			importantWords++
			if answerWordMap[word] {
				addressedWords++
			}
		}
	}

	if importantWords > 0 {
		wordCoverage := float64(addressedWords) / float64(importantWords)
		score = wordCoverage
	}

	// Bonus points for providing examples
	if strings.Contains(answerLower, "example") || strings.Contains(answerLower, "for instance") {
		score += 0.1
	}

	// Bonus points for providing alternatives
	if strings.Contains(answerLower, "alternatively") || strings.Contains(answerLower, "also") {
		score += 0.1
	}

	// Check for code examples in NixOS-related questions
	if (strings.Contains(questionLower, "config") || strings.Contains(questionLower, "nix")) &&
		(strings.Contains(answer, "```") || strings.Contains(answer, "{") || strings.Contains(answer, "=")) {
		score += 0.2
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// assessAccuracy evaluates technical accuracy (syntax, facts)
func (qm *QualityMetrics) assessAccuracy(answer string) float64 {
	score := 0.8 // Start with high confidence

	// Check for common NixOS syntax errors
	if strings.Contains(answer, "nix-env -i") && strings.Contains(answer, "recommended") {
		score -= 0.2
	}

	// Check for proper Nix syntax patterns
	if strings.Contains(answer, "{") && strings.Contains(answer, "}") {
		if strings.Count(answer, "{") != strings.Count(answer, "}") {
			score -= 0.3
		}
	}

	// Check for proper attribute paths
	nixAttributePattern := regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9_]*\.[a-zA-Z][a-zA-Z0-9_]*\b`)
	if nixAttributePattern.MatchString(answer) {
		score += 0.1
	}

	// Check for deprecated patterns
	deprecatedPatterns := []string{
		"with pkgs;",
		"imports = [ <",
	}

	for _, pattern := range deprecatedPatterns {
		if strings.Contains(answer, pattern) {
			score -= 0.1
		}
	}

	// Check for proper configuration.nix structure
	if strings.Contains(answer, "configuration.nix") {
		if strings.Contains(answer, "{ config, pkgs, ... }:") {
			score += 0.1
		}
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// assessPracticality evaluates how actionable and practical the response is
func (qm *QualityMetrics) assessPracticality(answer string) float64 {
	score := 0.5 // Start neutral

	answerLower := strings.ToLower(answer)

	// Check for actionable commands
	commands := []string{
		"nixos-rebuild", "nix-shell", "nix run", "nix build",
		"nix flake", "home-manager", "systemctl",
	}

	commandCount := 0
	for _, cmd := range commands {
		if strings.Contains(answerLower, cmd) {
			commandCount++
		}
	}

	if commandCount > 0 {
		score += 0.3
	}

	// Check for step-by-step instructions
	stepIndicators := []string{
		"step 1", "first,", "then,", "next,", "finally,",
		"1.", "2.", "3.",
	}

	stepCount := 0
	for _, indicator := range stepIndicators {
		if strings.Contains(answerLower, indicator) {
			stepCount++
		}
	}

	if stepCount >= 2 {
		score += 0.2
	}

	// Check for file paths and specific locations
	if strings.Contains(answer, "/etc/nixos/") || strings.Contains(answer, "~/.config/") {
		score += 0.1
	}

	// Check for warnings about potential issues
	if strings.Contains(answerLower, "warning") || strings.Contains(answerLower, "careful") ||
		strings.Contains(answerLower, "backup") {
		score += 0.1
	}

	// Penalty for overly theoretical responses
	if strings.Contains(answerLower, "in theory") || strings.Contains(answerLower, "conceptually") {
		if !strings.Contains(answerLower, "example") && !strings.Contains(answerLower, "step") {
			score -= 0.2
		}
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// assessStructure evaluates how well-structured and organized the response is
func (qm *QualityMetrics) assessStructure(answer string) float64 {
	score := 0.5 // Start neutral

	// Check for clear sections/headers
	if strings.Contains(answer, "##") || strings.Contains(answer, "**") {
		score += 0.2
	}

	// Check for lists
	listIndicators := []string{"- ", "* ", "1. ", "2. "}
	hasLists := false
	for _, indicator := range listIndicators {
		if strings.Contains(answer, indicator) {
			hasLists = true
			break
		}
	}

	if hasLists {
		score += 0.2
	}

	// Check for code blocks
	if strings.Contains(answer, "```") || strings.Count(answer, "`") >= 4 {
		score += 0.2
	}

	// Check for logical flow indicators
	flowIndicators := []string{
		"first", "then", "next", "finally", "however", "therefore",
		"additionally", "meanwhile", "consequently",
	}

	flowCount := 0
	answerLower := strings.ToLower(answer)
	for _, indicator := range flowIndicators {
		if strings.Contains(answerLower, indicator) {
			flowCount++
		}
	}

	if flowCount >= 2 {
		score += 0.2
	}

	// Check for paragraph structure
	paragraphs := strings.Split(strings.TrimSpace(answer), "\n\n")
	if len(paragraphs) >= 2 && len(paragraphs) <= 6 {
		score += 0.1
	} else if len(paragraphs) == 1 && len(answer) > 500 {
		score -= 0.2
	}

	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// calculateOverallScore computes weighted average of all dimensions
func (qm *QualityMetrics) calculateOverallScore(dimensions *QualityDimensions) float64 {
	// Weights for different dimensions
	weights := map[string]float64{
		"clarity":      0.20,
		"completeness": 0.25,
		"accuracy":     0.30,
		"practicality": 0.15,
		"structure":    0.10,
	}

	totalScore := 0.0
	totalScore += dimensions.Clarity * weights["clarity"]
	totalScore += dimensions.Completeness * weights["completeness"]
	totalScore += dimensions.Accuracy * weights["accuracy"]
	totalScore += dimensions.Practicality * weights["practicality"]
	totalScore += dimensions.Structure * weights["structure"]

	return totalScore
}

// determineQualityLevel converts numeric score to quality level
func (qm *QualityMetrics) determineQualityLevel(score float64) string {
	if score >= 0.85 {
		return "excellent"
	} else if score >= 0.70 {
		return "good"
	} else if score >= 0.55 {
		return "fair"
	} else {
		return "poor"
	}
}

// generateInsights provides human-readable insights about the response quality
func (qm *QualityMetrics) generateInsights(dimensions *QualityDimensions, answer string) ([]string, []string, []string) {
	var strengths, weaknesses, suggestions []string

	// Analyze each dimension
	if dimensions.Clarity >= 0.8 {
		strengths = append(strengths, "Response is clear and easy to understand")
	} else if dimensions.Clarity < 0.6 {
		weaknesses = append(weaknesses, "Response could be clearer and more concise")
		suggestions = append(suggestions, "Use shorter sentences and explain technical terms")
	}

	if dimensions.Completeness >= 0.8 {
		strengths = append(strengths, "Response thoroughly addresses the question")
	} else if dimensions.Completeness < 0.6 {
		weaknesses = append(weaknesses, "Response doesn't fully address all aspects of the question")
		suggestions = append(suggestions, "Provide more comprehensive coverage of the topic")
	}

	if dimensions.Accuracy >= 0.8 {
		strengths = append(strengths, "Response appears technically accurate")
	} else if dimensions.Accuracy < 0.6 {
		weaknesses = append(weaknesses, "Response may contain technical inaccuracies")
		suggestions = append(suggestions, "Verify syntax and technical details")
	}

	if dimensions.Practicality >= 0.8 {
		strengths = append(strengths, "Response provides actionable guidance")
	} else if dimensions.Practicality < 0.6 {
		weaknesses = append(weaknesses, "Response lacks practical, actionable steps")
		suggestions = append(suggestions, "Include specific commands and step-by-step instructions")
	}

	if dimensions.Structure >= 0.8 {
		strengths = append(strengths, "Response is well-organized and structured")
	} else if dimensions.Structure < 0.6 {
		weaknesses = append(weaknesses, "Response structure could be improved")
		suggestions = append(suggestions, "Use headings, lists, and code blocks for better organization")
	}

	return strengths, weaknesses, suggestions
}

// ExportAssessment exports the quality assessment as JSON for external use
func (qm *QualityMetrics) ExportAssessment(assessment *QualityAssessment) (string, error) {
	jsonData, err := json.MarshalIndent(assessment, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
