package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// WikiClient handles interactions with the NixOS Wiki
type WikiClient struct {
	client  *http.Client
	baseURL string
	logger  *logger.Logger
}

// WikiSearchResult represents a search result from the wiki
type WikiSearchResult struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Snippet     string  `json:"snippet"`
	Relevance   float64 `json:"relevance"`
	LastUpdated string  `json:"last_updated"`
}

// WikiPageContent represents the content of a wiki page
type WikiPageContent struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	URL         string   `json:"url"`
	LastUpdated string   `json:"last_updated"`
	Categories  []string `json:"categories"`
	BackLinks   []string `json:"back_links"`
}

// WikiValidationResult represents the result of wiki validation
type WikiValidationResult struct {
	MatchingPages   []WikiSearchResult `json:"matching_pages"`
	RelevantContent []string           `json:"relevant_content"`
	Contradictions  []string           `json:"contradictions"`
	Confidence      float64            `json:"confidence"`
	LastChecked     time.Time          `json:"last_checked"`
}

// MediaWikiAPIResponse represents the MediaWiki API response
type MediaWikiAPIResponse struct {
	Query struct {
		Search []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
			Size    int    `json:"size"`
		} `json:"search"`
		Pages map[string]struct {
			PageID  int    `json:"pageid"`
			Title   string `json:"title"`
			Extract string `json:"extract"`
			Touched string `json:"touched"`
		} `json:"pages"`
	} `json:"query"`
}

// NewWikiClient creates a new wiki client
func NewWikiClient() *WikiClient {
	return &WikiClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://wiki.nixos.org",
		logger:  logger.NewLogger(),
	}
}

// SearchWiki searches the NixOS wiki for relevant content
func (wc *WikiClient) SearchWiki(ctx context.Context, query string) ([]WikiSearchResult, error) {
	wc.logger.Printf("Searching wiki for query: %s", query)

	// Use MediaWiki API for search
	apiURL := fmt.Sprintf("%s/api.php", wc.baseURL)

	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("list", "search")
	params.Set("srsearch", query)
	params.Set("srlimit", "10")
	params.Set("srinfo", "")
	params.Set("srprop", "size|snippet|titlesnippet")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create wiki search request: %w", err)
	}

	req.Header.Set("User-Agent", "nixai/1.0 (https://github.com/nixos/nixai)")

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search wiki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wiki search failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read wiki response: %w", err)
	}

	var apiResp MediaWikiAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse wiki response: %w", err)
	}

	var results []WikiSearchResult
	for _, searchResult := range apiResp.Query.Search {
		relevance := wc.calculateRelevance(query, searchResult.Title, searchResult.Snippet)

		result := WikiSearchResult{
			Title:     searchResult.Title,
			URL:       fmt.Sprintf("%s/wiki/%s", wc.baseURL, url.QueryEscape(searchResult.Title)),
			Snippet:   wc.cleanSnippet(searchResult.Snippet),
			Relevance: relevance,
		}

		results = append(results, result)
	}

	wc.logger.Printf("Wiki search completed with %d results", len(results))
	return results, nil
}

// GetPageContent retrieves the full content of a wiki page
func (wc *WikiClient) GetPageContent(ctx context.Context, pageTitle string) (*WikiPageContent, error) {
	wc.logger.Printf("Fetching page content for: %s", pageTitle)

	apiURL := fmt.Sprintf("%s/api.php", wc.baseURL)

	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("prop", "extracts|info")
	params.Set("titles", pageTitle)
	params.Set("exintro", "")
	params.Set("explaintext", "")
	params.Set("exsectionformat", "plain")
	params.Set("inprop", "url")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create page content request: %w", err)
	}

	req.Header.Set("User-Agent", "nixai/1.0 (https://github.com/nixos/nixai)")

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("page content fetch failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read page response: %w", err)
	}

	var apiResp MediaWikiAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	for _, page := range apiResp.Query.Pages {
		content := &WikiPageContent{
			Title:       page.Title,
			Content:     page.Extract,
			URL:         fmt.Sprintf("%s/wiki/%s", wc.baseURL, url.QueryEscape(page.Title)),
			LastUpdated: page.Touched,
			Categories:  []string{}, // Would need additional API call
			BackLinks:   []string{}, // Would need additional API call
		}

		wc.logger.Printf("Page content retrieved - title: %s, length: %d", content.Title, len(content.Content))
		return content, nil
	}

	return nil, fmt.Errorf("page not found: %s", pageTitle)
}

// ValidateAgainstWiki validates an answer against wiki content
func (wc *WikiClient) ValidateAgainstWiki(ctx context.Context, question, answer string) (*WikiValidationResult, error) {
	wc.logger.Printf("Validating against wiki for question: %s", question)

	result := &WikiValidationResult{
		MatchingPages:   []WikiSearchResult{},
		RelevantContent: []string{},
		Contradictions:  []string{},
		Confidence:      0.5,
		LastChecked:     time.Now(),
	}

	// Search for relevant wiki pages
	searchQuery := wc.extractSearchTerms(question, answer)
	searchResults, err := wc.SearchWiki(ctx, searchQuery)
	if err != nil {
		wc.logger.Printf("Failed to search wiki: %v", err)
		return result, err
	}

	result.MatchingPages = searchResults

	// Analyze the most relevant pages
	for i, searchResult := range searchResults {
		if i >= 3 { // Limit to top 3 results
			break
		}

		if searchResult.Relevance > 0.6 {
			pageContent, err := wc.GetPageContent(ctx, searchResult.Title)
			if err != nil {
				wc.logger.Printf("Failed to get page content for %s: %v", searchResult.Title, err)
				continue
			}

			// Check for relevant content
			relevantSections := wc.findRelevantContent(answer, pageContent.Content)
			result.RelevantContent = append(result.RelevantContent, relevantSections...)

			// Check for contradictions
			contradictions := wc.findContradictions(answer, pageContent.Content)
			result.Contradictions = append(result.Contradictions, contradictions...)
		}
	}

	// Calculate confidence based on findings
	result.Confidence = wc.calculateWikiConfidence(result, searchResults)

	wc.logger.Printf("Wiki validation completed - matching_pages: %d, relevant_content: %d, contradictions: %d, confidence: %.2f",
		len(result.MatchingPages),
		len(result.RelevantContent),
		len(result.Contradictions),
		result.Confidence,
	)

	return result, nil
}

// extractSearchTerms extracts relevant search terms from question and answer
func (wc *WikiClient) extractSearchTerms(question, answer string) string {
	// Extract technical terms that are likely to be in the wiki
	technicalTerms := []string{}

	// Common NixOS/Nix patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\b(services\.[a-zA-Z0-9_-]+)\b`),
		regexp.MustCompile(`\b(boot\.[a-zA-Z0-9_-]+)\b`),
		regexp.MustCompile(`\b(networking\.[a-zA-Z0-9_-]+)\b`),
		regexp.MustCompile(`\b(hardware\.[a-zA-Z0-9_-]+)\b`),
		regexp.MustCompile(`\b(nix[a-zA-Z0-9_-]*)\b`),
		regexp.MustCompile(`\b(flake[s]?)\b`),
		regexp.MustCompile(`\b(derivation[s]?)\b`),
		regexp.MustCompile(`\b(channel[s]?)\b`),
		regexp.MustCompile(`\b(overlay[s]?)\b`),
	}

	combinedText := question + " " + answer
	for _, pattern := range patterns {
		matches := pattern.FindAllString(combinedText, -1)
		technicalTerms = append(technicalTerms, matches...)
	}

	// Also include specific package names mentioned
	packagePattern := regexp.MustCompile(`pkgs\.([a-zA-Z0-9_-]+)`)
	packageMatches := packagePattern.FindAllStringSubmatch(combinedText, -1)
	for _, match := range packageMatches {
		if len(match) > 1 {
			technicalTerms = append(technicalTerms, match[1])
		}
	}

	// Remove duplicates and join
	uniqueTerms := wc.removeDuplicates(technicalTerms)
	searchQuery := strings.Join(uniqueTerms, " ")

	// Fallback to extracting key words from question if no technical terms found
	if len(uniqueTerms) == 0 {
		words := strings.Fields(question)
		importantWords := []string{}
		for _, word := range words {
			if len(word) > 4 && !wc.isCommonWord(word) {
				importantWords = append(importantWords, word)
			}
		}
		searchQuery = strings.Join(importantWords, " ")
	}

	wc.logger.Printf("Extracted search terms: %s", searchQuery)
	return searchQuery
}

// calculateRelevance calculates how relevant a search result is
func (wc *WikiClient) calculateRelevance(query, title, snippet string) float64 {
	relevance := 0.0

	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(title)
	snippetLower := strings.ToLower(snippet)

	// Title match is most important
	queryWords := strings.Fields(queryLower)
	titleWords := strings.Fields(titleLower)

	titleMatches := 0
	for _, queryWord := range queryWords {
		for _, titleWord := range titleWords {
			if strings.Contains(titleWord, queryWord) || strings.Contains(queryWord, titleWord) {
				titleMatches++
				break
			}
		}
	}

	if len(queryWords) > 0 {
		relevance += (float64(titleMatches) / float64(len(queryWords))) * 0.6
	}

	// Snippet match is secondary
	snippetMatches := 0
	for _, queryWord := range queryWords {
		if strings.Contains(snippetLower, queryWord) {
			snippetMatches++
		}
	}

	if len(queryWords) > 0 {
		relevance += (float64(snippetMatches) / float64(len(queryWords))) * 0.4
	}

	return relevance
}

// findRelevantContent finds content in the wiki page that's relevant to the answer
func (wc *WikiClient) findRelevantContent(answer, pageContent string) []string {
	var relevantSections []string

	answerLower := strings.ToLower(answer)
	_ = answerLower // Used in the loop below

	// Split page content into paragraphs
	paragraphs := strings.Split(pageContent, "\n\n")

	for _, paragraph := range paragraphs {
		if len(paragraph) < 50 { // Skip very short paragraphs
			continue
		}

		paragraphLower := strings.ToLower(paragraph)

		// Check if paragraph contains similar concepts
		similarity := wc.calculateTextSimilarity(answerLower, paragraphLower)
		if similarity > 0.3 {
			relevantSections = append(relevantSections, paragraph)
		}
	}

	return relevantSections
}

// findContradictions finds potential contradictions between answer and wiki content
func (wc *WikiClient) findContradictions(answer, pageContent string) []string {
	var contradictions []string

	// This is a simplified contradiction detection
	// In practice, this would need more sophisticated NLP

	answerLower := strings.ToLower(answer)
	pageContentLower := strings.ToLower(pageContent)

	// Look for explicit contradictory phrases
	contradictoryPairs := map[string]string{
		"deprecated":    "recommended",
		"not supported": "supported",
		"doesn't work":  "works",
		"impossible":    "possible",
		"never":         "always",
		"can't":         "can",
	}

	for negative, positive := range contradictoryPairs {
		if strings.Contains(answerLower, negative) && strings.Contains(pageContentLower, positive) {
			contradictions = append(contradictions, fmt.Sprintf("Answer suggests '%s' but wiki indicates '%s'", negative, positive))
		}
		if strings.Contains(answerLower, positive) && strings.Contains(pageContentLower, negative) {
			contradictions = append(contradictions, fmt.Sprintf("Answer suggests '%s' but wiki indicates '%s'", positive, negative))
		}
	}

	return contradictions
}

// calculateWikiConfidence calculates confidence based on wiki validation results
func (wc *WikiClient) calculateWikiConfidence(result *WikiValidationResult, searchResults []WikiSearchResult) float64 {
	confidence := 0.5 // Base confidence

	// Boost confidence if we found relevant pages
	if len(searchResults) > 0 {
		avgRelevance := 0.0
		for _, result := range searchResults {
			avgRelevance += result.Relevance
		}
		avgRelevance /= float64(len(searchResults))
		confidence += avgRelevance * 0.3
	}

	// Boost confidence if we found relevant content
	if len(result.RelevantContent) > 0 {
		confidence += 0.2
	}

	// Reduce confidence if we found contradictions
	if len(result.Contradictions) > 0 {
		confidence -= float64(len(result.Contradictions)) * 0.2
	}

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// Helper methods

func (wc *WikiClient) cleanSnippet(snippet string) string {
	// Remove HTML tags and clean up snippet
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	cleaned := htmlTagRegex.ReplaceAllString(snippet, "")

	// Replace HTML entities
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")

	return strings.TrimSpace(cleaned)
}

func (wc *WikiClient) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var unique []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			unique = append(unique, item)
		}
	}

	return unique
}

func (wc *WikiClient) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true, "but": true,
		"not": true, "you": true, "all": true, "can": true, "had": true,
		"her": true, "was": true, "one": true, "our": true, "out": true,
		"day": true, "get": true, "has": true, "him": true, "his": true,
		"how": true, "its": true, "may": true, "new": true, "now": true,
		"old": true, "see": true, "two": true, "way": true, "who": true,
		"boy": true, "did": true, "man": true, "men": true, "put": true,
		"say": true, "she": true, "too": true, "use": true, "with": true,
		"that": true, "have": true, "will": true, "your": true, "from": true,
		"they": true, "know": true, "want": true, "been": true, "good": true,
		"much": true, "some": true, "time": true, "very": true, "when": true,
		"come": true, "here": true, "just": true, "like": true, "long": true,
		"make": true, "many": true, "over": true, "such": true, "take": true,
		"than": true, "them": true, "well": true, "were": true, "what": true,
		"where": true,
	}

	return commonWords[strings.ToLower(word)]
}

func (wc *WikiClient) calculateTextSimilarity(text1, text2 string) float64 {
	// Simple word overlap similarity
	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Create word frequency maps
	freq1 := make(map[string]int)
	freq2 := make(map[string]int)

	for _, word := range words1 {
		if len(word) > 3 && !wc.isCommonWord(word) {
			freq1[word]++
		}
	}

	for _, word := range words2 {
		if len(word) > 3 && !wc.isCommonWord(word) {
			freq2[word]++
		}
	}

	// Calculate overlap
	overlap := 0
	for word := range freq1 {
		if freq2[word] > 0 {
			overlap++
		}
	}

	// Similarity is overlap divided by union size
	union := len(freq1) + len(freq2) - overlap
	if union == 0 {
		return 0.0
	}

	return float64(overlap) / float64(union)
}
