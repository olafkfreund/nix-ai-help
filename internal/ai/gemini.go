package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// GeminiRequest represents a request to the Gemini AI model.
type GeminiRequest struct {
	Prompt string `json:"prompt"`
}

// GeminiResponse represents a response from the Gemini AI model.
type GeminiResponse struct {
	Text string `json:"text"`
}

// Query sends a request to the Gemini AI model and returns the response.
func (c *GeminiClient) Query(prompt string) (string, error) {
	reqBody := GeminiRequest{Prompt: prompt}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/query", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	var responseBody GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return responseBody.Text, nil
}
