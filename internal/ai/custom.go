package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CustomProvider implements the Provider interface for user-defined HTTP API AI providers.
type CustomProvider struct {
	BaseURL string            // Base URL of the API endpoint
	Headers map[string]string // Custom headers (e.g., Authorization, Content-Type)
	client  *http.Client
}

func NewCustomProvider(baseURL string, headers map[string]string) *CustomProvider {
	if headers == nil {
		headers = make(map[string]string)
	}
	// Set default Content-Type if not provided
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	return &CustomProvider{
		BaseURL: baseURL,
		Headers: headers,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Query sends a prompt to the custom HTTP API provider and returns its response.
func (c *CustomProvider) Query(prompt string) (string, error) {
	// Create request payload - adapt this structure based on your API's expected format
	payload := map[string]interface{}{
		"prompt":     prompt,
		"max_tokens": 2048,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("custom provider request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("custom provider returned status %d", resp.StatusCode)
	}

	var response struct {
		Response string `json:"response"`
		Text     string `json:"text"`
		Content  string `json:"content"`
		Answer   string `json:"answer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Try different common response field names
	if response.Response != "" {
		return response.Response, nil
	}
	if response.Text != "" {
		return response.Text, nil
	}
	if response.Content != "" {
		return response.Content, nil
	}
	if response.Answer != "" {
		return response.Answer, nil
	}

	return "", fmt.Errorf("no recognized response field found in API response")
}

// GenerateResponse is an alias for Query for compatibility.
func (c *CustomProvider) GenerateResponse(prompt string) (string, error) {
	return c.Query(prompt)
}
