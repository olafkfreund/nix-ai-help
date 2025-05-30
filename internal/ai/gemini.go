package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GeminiClient is a struct that holds the configuration for the Gemini AI provider.
type GeminiClient struct {
	APIKey  string
	BaseURL string
}

// NewGeminiClient initializes a new GeminiClient with the provided API key and base URL.
func NewGeminiClient(apiKey, baseURL string) *GeminiClient {
	return &GeminiClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

// GeminiRequest represents a request to the Gemini AI model (Google API format)
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// GeminiResponse represents a response from the Gemini AI model (Google API format)
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Query sends a request to the official Google Gemini API and returns the response.
func (c *GeminiClient) Query(prompt string) (string, error) {
	apiURL := c.BaseURL
	if apiURL == "" {
		apiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
	}
	apiKey := c.APIKey
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}
	// Build request body
	requestBody := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	urlWithKey := apiURL + "?key=" + apiKey
	// DEBUG: Print prompt and API URL for troubleshooting
	fmt.Printf("[Gemini Debug] API URL: %s\n", urlWithKey)
	fmt.Printf("[Gemini Debug] Prompt (truncated): %s\n", prompt[:min(500, len(prompt))])
	req, err := http.NewRequest("POST", urlWithKey, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("[Gemini Debug] Non-200 response: %s\n%s\n", resp.Status, string(b))
		return "", fmt.Errorf("received non-200 response: %s\n%s", resp.Status, string(b))
	}
	var responseBody GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		fmt.Printf("[Gemini Debug] Failed to decode response: %v\n", err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if len(responseBody.Candidates) == 0 || len(responseBody.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini API")
	}
	return responseBody.Candidates[0].Content.Parts[0].Text, nil
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
