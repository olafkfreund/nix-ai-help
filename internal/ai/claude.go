package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ClaudeProvider implements the Provider interface for Anthropic Claude API
type ClaudeProvider struct {
	APIKey      string
	Model       string
	BaseURL     string
	MaxTokens   int
	Temperature float64
	Client      *http.Client
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []ClaudeMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// ClaudeMessage represents a message in the Claude conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response structure from Claude API
type ClaudeResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewClaudeProvider creates a new Claude provider with default settings
func NewClaudeProvider(model string) *ClaudeProvider {
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	baseURL := os.Getenv("CLAUDE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/messages"
	}

	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Default to latest Claude 3.5 Sonnet
	}

	return &ClaudeProvider{
		APIKey:      apiKey,
		Model:       model,
		BaseURL:     baseURL,
		MaxTokens:   4096,
		Temperature: 0.7,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Query implements the Provider interface for context-aware queries
func (c *ClaudeProvider) Query(ctx context.Context, prompt string) (string, error) {
	return c.queryWithContext(ctx, prompt)
}

// GenerateResponse implements the Provider interface for response generation
func (c *ClaudeProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return c.queryWithContext(ctx, prompt)
}

// queryWithContext performs the actual API call to Claude
func (c *ClaudeProvider) queryWithContext(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("Claude API key not set. Please set CLAUDE_API_KEY or ANTHROPIC_API_KEY environment variable")
	}

	// Build the request
	request := ClaudeRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		System:      "You are a helpful assistant specializing in NixOS configuration, troubleshooting, and best practices. Provide clear, accurate, and actionable advice.",
		Temperature: c.Temperature,
		Stream:      false,
	}

	// Marshal the request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Claude request: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create Claude request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Claude API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned status %d", resp.StatusCode)
	}

	// Parse the response
	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("failed to decode Claude response: %w", err)
	}

	// Extract the text content
	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	var response string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			response += content.Text
		}
	}

	if response == "" {
		return "", fmt.Errorf("no text content found in Claude response")
	}

	return response, nil
}

// SetModel allows changing the model after creation
func (c *ClaudeProvider) SetModel(model string) {
	c.Model = model
}

// SetMaxTokens allows changing the max tokens after creation
func (c *ClaudeProvider) SetMaxTokens(maxTokens int) {
	c.MaxTokens = maxTokens
}

// SetTemperature allows changing the temperature after creation
func (c *ClaudeProvider) SetTemperature(temperature float64) {
	c.Temperature = temperature
}

// SetTimeout updates the HTTP client timeout for Claude requests
func (c *ClaudeProvider) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
}

// GetTimeout returns the current HTTP client timeout
func (c *ClaudeProvider) GetTimeout() time.Duration {
	return c.Client.Timeout
}

// HealthCheck verifies that the Claude API is accessible
func (c *ClaudeProvider) HealthCheck() error {
	if c.APIKey == "" {
		return fmt.Errorf("Claude API key not configured")
	}

	// Simple test query to verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.Query(ctx, "Hello")
	if err != nil {
		return fmt.Errorf("Claude health check failed: %w", err)
	}

	return nil
}

// GetAvailableModels returns a list of available Claude models
func (c *ClaudeProvider) GetAvailableModels() []string {
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
}

// GetSelectedModel returns the currently selected model
func (c *ClaudeProvider) GetSelectedModel() string {
	return c.Model
}

// Legacy Provider Wrapper for backward compatibility
type ClaudeLegacyProvider struct {
	*ClaudeProvider
}

// NewClaudeLegacyProvider creates a legacy provider wrapper
func NewClaudeLegacyProvider(model string) *ClaudeLegacyProvider {
	return &ClaudeLegacyProvider{
		ClaudeProvider: NewClaudeProvider(model),
	}
}

// Query implements the legacy AIProvider interface
func (c *ClaudeLegacyProvider) Query(prompt string) (string, error) {
	return c.ClaudeProvider.Query(context.Background(), prompt)
}

// GenerateResponse implements the legacy AIProvider interface
func (c *ClaudeLegacyProvider) GenerateResponse(prompt string) (string, error) {
	return c.ClaudeProvider.GenerateResponse(context.Background(), prompt)
}
