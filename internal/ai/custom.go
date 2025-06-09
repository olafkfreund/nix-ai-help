package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"nix-ai-help/internal/config"
)

// CustomProvider implements the Provider interface for user-defined HTTP API AI providers.
type CustomProvider struct {
	BaseURL string            // Base URL of the API endpoint
	Headers map[string]string // Custom headers (e.g., Authorization, Content-Type)
	Model   string            // Selected model name
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
		Model:   "",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewCustomProviderWithModel creates a new CustomProvider with a specific model configuration.
func NewCustomProviderWithModel(providerConfig *config.AIProviderConfig, modelName string) (*CustomProvider, error) {
	// Validate that the model exists in the provider configuration
	model, exists := providerConfig.Models[modelName]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in Custom provider configuration", modelName)
	}

	// Extract headers from configuration (if any)
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	// Check for API key requirement
	if providerConfig.RequiresAPIKey && providerConfig.EnvVar != "" {
		apiKey := os.Getenv(providerConfig.EnvVar)
		if apiKey != "" {
			headers["Authorization"] = "Bearer " + apiKey
		}
	}

	return &CustomProvider{
		BaseURL: providerConfig.BaseURL,
		Headers: headers,
		Model:   model.Name,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// CheckHealth checks if the custom provider server is accessible and responding.
func (c *CustomProvider) CheckHealth() error {
	// Try to make a simple request to check if the server is running
	req, err := http.NewRequest("GET", c.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Add custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("custom provider server not accessible: %w", err)
	}
	defer resp.Body.Close()

	// Accept both successful responses and method not allowed (since we're doing GET on POST endpoint)
	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusMethodNotAllowed {
		return fmt.Errorf("custom provider server returned error status: %d", resp.StatusCode)
	}

	return nil
}

// GetSelectedModel returns the currently selected model.
func (c *CustomProvider) GetSelectedModel() string {
	return c.Model
}

// SetModel updates the selected model.
func (c *CustomProvider) SetModel(modelName string) {
	c.Model = modelName
}

// Query sends a prompt to the custom HTTP API provider and returns its response.
func (c *CustomProvider) Query(prompt string) (string, error) {
	// Create request payload - adapt this structure based on your API's expected format
	payload := map[string]interface{}{
		"prompt":     prompt,
		"max_tokens": 2048,
	}

	if c.Model != "" {
		payload["model"] = c.Model
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
