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
	Model   string // Added model support
}

// NewGeminiClient initializes a new GeminiClient with the provided API key and base URL.
func NewGeminiClient(apiKey, baseURL string) *GeminiClient {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent"
	}
	return &GeminiClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   "gemini-pro", // default model
	}
}

// NewGeminiClientWithModel creates a GeminiClient with a specific model.
func NewGeminiClientWithModel(apiKey, baseURL, model string) *GeminiClient {
	client := NewGeminiClient(apiKey, baseURL)
	if model != "" {
		client.Model = model
		// Update URL to use the specified model
		if baseURL == "" {
			client.BaseURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
		}
	}
	return client
}

// CheckHealth checks if the Gemini API is accessible and responding.
func (g *GeminiClient) CheckHealth() error {
	// For Gemini, we can check by making a simple request to the models list endpoint
	listURL := "https://generativelanguage.googleapis.com/v1beta/models"
	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("X-Goog-Api-Key", g.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gemini API not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gemini API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// GetSelectedModel returns the currently selected model.
func (g *GeminiClient) GetSelectedModel() string {
	return g.Model
}

// SetModel updates the selected model.
func (g *GeminiClient) SetModel(modelName string) {
	g.Model = modelName
	// Update the base URL to use the new model
	g.BaseURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", modelName)
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
		apiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent"
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("received non-200 response: %s\n%s", resp.Status, string(b))
	}

	var responseBody GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
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
