package validation

import (
	"context"
	"strings"
	"time"

	"nix-ai-help/internal/community"
	"nix-ai-help/pkg/logger"
)

// AIResponse represents an AI-generated response for validation
type AIResponse struct {
	Content    string            `json:"content"`
	Question   string            `json:"question,omitempty"`
	Confidence float64           `json:"confidence,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// CommunityValidator validates answers against community knowledge and best practices
type CommunityValidator struct {
	githubClient *community.GitHubClient
	wikiClient   *WikiClient
	logger       *logger.Logger
}

// CommunityValidationResult represents the result of community validation
type CommunityValidationResult struct {
	IsValid             bool                       `json:"is_valid"`
	Confidence          float64                    `json:"confidence"`
	CommunityConsensus  float64                    `json:"community_consensus"`
	BestPractices       []string                   `json:"best_practices"`
	CommonGotchas       []string                   `json:"common_gotchas"`
	WikiValidation      *WikiValidationResult      `json:"wiki_validation,omitempty"`
	GitHubValidation    *GitHubValidationResult    `json:"github_validation,omitempty"`
	DiscourseValidation *DiscourseValidationResult `json:"discourse_validation,omitempty"`
	Issues              []ValidationIssue          `json:"issues"`
	Recommendations     []string                   `json:"recommendations"`
}

// GitHubValidationResult represents GitHub-based validation
type GitHubValidationResult struct {
	IssuesFound     []string `json:"issues_found"`
	CommitsFound    []string `json:"commits_found"`
	Confidence      float64  `json:"confidence"`
	RelevantContent []string `json:"relevant_content"`
}

// DiscourseValidationResult represents Discourse forum validation
type DiscourseValidationResult struct {
	PostsFound      []string `json:"posts_found"`
	Confidence      float64  `json:"confidence"`
	RelevantContent []string `json:"relevant_content"`
}

// NewCommunityValidator creates a new community validator
func NewCommunityValidator(githubToken string) *CommunityValidator {
	var githubClient *community.GitHubClient
	if githubToken != "" {
		githubClient = community.NewGitHubClient(githubToken)
	}

	wikiClient := NewWikiClient()

	return &CommunityValidator{
		githubClient: githubClient,
		wikiClient:   wikiClient,
		logger:       logger.NewLogger(),
	}
}

// ValidateAgainstCommunity validates answers against community sources (interface for enhanced validator)
func (cv *CommunityValidator) ValidateAgainstCommunity(ctx context.Context, question, answer string) (*CommunityValidationResult, error) {
	// Create AIResponse from question and answer
	response := &AIResponse{
		Content:  answer,
		Question: question,
	}

	// Use the existing ValidateResponse method
	return cv.ValidateResponse(ctx, response)
}

// ValidateResponse validates an AI response against community sources
func (cv *CommunityValidator) ValidateResponse(ctx context.Context, response *AIResponse) (*CommunityValidationResult, error) {
	cv.logger.Printf("Starting community validation for response")

	result := &CommunityValidationResult{
		IsValid:         true,
		Issues:          []ValidationIssue{},
		Recommendations: []string{},
	}

	// Validate against Wiki using available method
	if cv.wikiClient != nil {
		wikiResult, err := cv.validateAgainstWiki(ctx, response)
		if err != nil {
			cv.logger.Printf("Wiki validation error: %v", err)
		} else {
			result.WikiValidation = wikiResult
		}
	}

	// Calculate overall confidence
	result.Confidence = cv.calculateCommunityConfidence(result)

	// Check if validation passes threshold
	if result.Confidence < 0.6 {
		result.IsValid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "low_community_confidence",
			Severity:    "medium",
			Description: "Low confidence from community validation sources",
		})
		result.Recommendations = append(result.Recommendations, "Consider cross-referencing with official documentation")
	}

	cv.logger.Printf("Community validation completed with confidence: %.2f", result.Confidence)
	return result, nil
}

// validateAgainstWiki validates the response against NixOS Wiki content using SearchWiki method
func (cv *CommunityValidator) validateAgainstWiki(ctx context.Context, response *AIResponse) (*WikiValidationResult, error) {
	cv.logger.Printf("Validating against NixOS Wiki")

	result := &WikiValidationResult{
		MatchingPages:   []WikiSearchResult{},
		RelevantContent: []string{},
		Confidence:      0.5,
		LastChecked:     time.Now(),
	}

	// Extract key terms from the response for wiki search
	searchTerms := cv.extractSearchTerms(response.Content)

	for _, term := range searchTerms {
		// Use the available SearchWiki method
		searchResults, err := cv.wikiClient.SearchWiki(ctx, term)
		if err != nil {
			cv.logger.Printf("Wiki search error for term '%s': %v", term, err)
			continue
		}

		for _, searchResult := range searchResults {
			result.MatchingPages = append(result.MatchingPages, searchResult)

			// Check if response content aligns with wiki content
			if cv.contentAligns(response.Content, searchResult.Snippet) {
				result.RelevantContent = append(result.RelevantContent, searchResult.Title)
				result.Confidence += 0.1 * searchResult.Relevance
			}
		}
	}

	// Cap confidence at 1.0
	if result.Confidence > 1.0 {
		result.Confidence = 1.0
	}

	return result, nil
}

// extractSearchTerms extracts key terms from response content for searching
func (cv *CommunityValidator) extractSearchTerms(content string) []string {
	// Simple term extraction - could be enhanced with NLP
	terms := []string{}

	// Look for NixOS-specific terms
	nixosTerms := []string{
		"configuration.nix", "home-manager", "flake.nix", "nixpkgs",
		"systemd", "services", "packages", "overlay", "derivation",
	}

	lowerContent := strings.ToLower(content)
	for _, term := range nixosTerms {
		if strings.Contains(lowerContent, term) {
			terms = append(terms, term)
		}
	}

	// Add any quoted terms as potential search terms
	words := strings.Fields(content)
	for _, word := range words {
		if strings.HasPrefix(word, `"`) && strings.HasSuffix(word, `"`) && len(word) > 3 {
			terms = append(terms, strings.Trim(word, `"`))
		}
	}

	// Fallback: use first few words as search terms
	if len(terms) == 0 {
		words := strings.Fields(content)
		maxWords := 5
		if len(words) < maxWords {
			maxWords = len(words)
		}
		for i := 0; i < maxWords; i++ {
			if len(words[i]) > 3 { // Skip short words
				terms = append(terms, words[i])
			}
		}
	}

	return terms
}

// contentAligns checks if response content aligns with wiki content
func (cv *CommunityValidator) contentAligns(responseContent, wikiContent string) bool {
	// Simple alignment check - could be enhanced with semantic similarity
	responseLower := strings.ToLower(responseContent)
	wikiLower := strings.ToLower(wikiContent)

	// Check for common important words
	responseWords := strings.Fields(responseLower)
	wikiWords := strings.Fields(wikiLower)

	commonWords := 0
	totalWords := len(responseWords)
	if totalWords == 0 {
		return false
	}

	for _, responseWord := range responseWords {
		if len(responseWord) > 3 { // Skip short words
			for _, wikiWord := range wikiWords {
				if responseWord == wikiWord {
					commonWords++
					break
				}
			}
		}
	}

	// Consider content aligned if >20% of words match
	return float64(commonWords)/float64(totalWords) > 0.2
}

// calculateCommunityConfidence calculates overall confidence from all community sources
func (cv *CommunityValidator) calculateCommunityConfidence(result *CommunityValidationResult) float64 {
	scores := []float64{}

	if result.WikiValidation != nil {
		scores = append(scores, result.WikiValidation.Confidence)
	}

	if result.GitHubValidation != nil {
		scores = append(scores, result.GitHubValidation.Confidence)
	}

	if result.DiscourseValidation != nil {
		scores = append(scores, result.DiscourseValidation.Confidence)
	}

	if len(scores) == 0 {
		return 0.5
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}
