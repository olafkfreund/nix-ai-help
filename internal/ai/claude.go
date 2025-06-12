package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"nix-ai-help/internal/config"
)

// ClaudeClient implements the AIProvider interface for Anthropic's Claude API.
type ClaudeClient struct {
	APIKey     string
	APIURL     string
	Model      string
	HTTPClient *http.Client
}

// NewClaudeClient creates a new Claude client with the provided API key.
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		APIKey:     apiKey,
		APIURL:     "https://api.anthropic.com/v1/messages",
		Model:      "claude-sonnet-4-20250514", // default model
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewClaudeClientWithModel creates a Claude client with a specific model.
func NewClaudeClientWithModel(apiKey, model string) *ClaudeClient {
	client := NewClaudeClient(apiKey)
	if model != "" {
		client.Model = model
	}
	return client
}

// NewClaudeProviderWithModel creates a new ClaudeClient with a specific model configuration.
func NewClaudeProviderWithModel(providerConfig *config.AIProviderConfig, modelName string) (*ClaudeClient, error) {
	// Validate that the model exists in the provider configuration
	model, exists := providerConfig.Models[modelName]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in Claude provider configuration", modelName)
	}

	apiKey := os.Getenv(providerConfig.EnvVar)
	if apiKey == "" && providerConfig.RequiresAPIKey {
		return nil, fmt.Errorf("Claude API key not found in environment variable %s", providerConfig.EnvVar)
	}

	baseURL := providerConfig.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/messages"
	}

	return &ClaudeClient{
		APIKey:     apiKey,
		APIURL:     baseURL,
		Model:      model.Name,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ClaudeRequest represents a request to the Claude API.
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeMessage represents a message in the Claude API.
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents a response from the Claude API.
type ClaudeResponse struct {
	Content []ClaudeContent `json:"content"`
	Error   *ClaudeError    `json:"error,omitempty"`
}

// ClaudeContent represents content in a Claude response.
type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ClaudeError represents an error response from the Claude API.
type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Query implements the Provider interface for ClaudeClient.
func (client *ClaudeClient) Query(ctx context.Context, prompt string) (string, error) {
	request := ClaudeRequest{
		Model:     client.Model,
		MaxTokens: 4096,
		Messages: []ClaudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", client.APIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned status %d", resp.StatusCode)
	}

	var response ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("Claude API error: %s", response.Error.Message)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}

// GenerateResponse implements the Provider interface for ClaudeClient.
func (client *ClaudeClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return client.Query(ctx, prompt)
}

// CheckHealth checks if the Claude API is accessible and responding.
func (client *ClaudeClient) CheckHealth() error {
	// Simple health check by making a minimal request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Query(ctx, "Hello")
	if err != nil {
		return fmt.Errorf("Claude API health check failed: %w", err)
	}

	return nil
}

// GetSelectedModel returns the currently selected model.
func (client *ClaudeClient) GetSelectedModel() string {
	return client.Model
}

// SetModel updates the selected model.
func (client *ClaudeClient) SetModel(model string) {
	client.Model = model
}

// SetTimeout updates the HTTP client timeout for Claude requests.
func (client *ClaudeClient) SetTimeout(timeout time.Duration) {
	client.HTTPClient.Timeout = timeout
}

// GetTimeout returns the current HTTP client timeout.
func (client *ClaudeClient) GetTimeout() time.Duration {
	return client.HTTPClient.Timeout
}

// Legacy Provider Wrapper for backward compatibility

// ClaudeLegacyProvider wraps ClaudeClient to provide legacy AIProvider interface.
type ClaudeLegacyProvider struct {
	*ClaudeClient
}

// NewClaudeLegacyProvider creates a legacy provider wrapper.
func NewClaudeLegacyProvider(apiKey, model string) *ClaudeLegacyProvider {
	return &ClaudeLegacyProvider{
		ClaudeClient: NewClaudeClientWithModel(apiKey, model),
	}
}

// Query implements the legacy AIProvider interface.
func (c *ClaudeLegacyProvider) Query(prompt string) (string, error) {
	return c.ClaudeClient.Query(context.Background(), prompt)
}

// GenerateResponse implements the legacy AIProvider interface.
func (c *ClaudeLegacyProvider) GenerateResponse(prompt string) (string, error) {
	return c.ClaudeClient.GenerateResponse(context.Background(), prompt)
}
