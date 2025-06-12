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

// GroqClient implements the AIProvider interface for Groq's API.
type GroqClient struct {
	APIKey     string
	APIURL     string
	Model      string
	HTTPClient *http.Client
}

// NewGroqClient creates a new Groq client with the provided API key.
func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		APIKey:     apiKey,
		APIURL:     "https://api.groq.com/openai/v1/chat/completions",
		Model:      "llama-3.3-70b-versatile", // default model
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewGroqClientWithModel creates a Groq client with a specific model.
func NewGroqClientWithModel(apiKey, model string) *GroqClient {
	client := NewGroqClient(apiKey)
	if model != "" {
		client.Model = model
	}
	return client
}

// NewGroqProviderWithModel creates a new GroqClient with a specific model configuration.
func NewGroqProviderWithModel(providerConfig *config.AIProviderConfig, modelName string) (*GroqClient, error) {
	// Validate that the model exists in the provider configuration
	model, exists := providerConfig.Models[modelName]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in Groq provider configuration", modelName)
	}

	apiKey := os.Getenv(providerConfig.EnvVar)
	if apiKey == "" && providerConfig.RequiresAPIKey {
		return nil, fmt.Errorf("Groq API key not found in environment variable %s", providerConfig.EnvVar)
	}

	baseURL := providerConfig.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1/chat/completions"
	}

	return &GroqClient{
		APIKey:     apiKey,
		APIURL:     baseURL,
		Model:      model.Name,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GroqRequest represents a request to the Groq API (OpenAI-compatible format).
type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// GroqMessage represents a message in the Groq API.
type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqResponse represents a response from the Groq API.
type GroqResponse struct {
	Choices []GroqChoice `json:"choices"`
	Error   *GroqError   `json:"error,omitempty"`
}

// GroqChoice represents a choice in the Groq response.
type GroqChoice struct {
	Message GroqMessage `json:"message"`
}

// GroqError represents an error response from the Groq API.
type GroqError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Query implements the Provider interface for GroqClient.
func (client *GroqClient) Query(ctx context.Context, prompt string) (string, error) {
	request := GroqRequest{
		Model: client.Model,
		Messages: []GroqMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
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

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Groq API returned status %d", resp.StatusCode)
	}

	var response GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("Groq API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateResponse implements the Provider interface for GroqClient.
func (client *GroqClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return client.Query(ctx, prompt)
}

// CheckHealth checks if the Groq API is accessible and responding.
func (client *GroqClient) CheckHealth() error {
	// Simple health check by making a minimal request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Query(ctx, "Hello")
	if err != nil {
		return fmt.Errorf("Groq API health check failed: %w", err)
	}

	return nil
}

// GetSelectedModel returns the currently selected model.
func (client *GroqClient) GetSelectedModel() string {
	return client.Model
}

// SetModel updates the selected model.
func (client *GroqClient) SetModel(model string) {
	client.Model = model
}

// SetTimeout updates the HTTP client timeout for Groq requests.
func (client *GroqClient) SetTimeout(timeout time.Duration) {
	client.HTTPClient.Timeout = timeout
}

// GetTimeout returns the current HTTP client timeout.
func (client *GroqClient) GetTimeout() time.Duration {
	return client.HTTPClient.Timeout
}

// Legacy Provider Wrapper for backward compatibility

// GroqLegacyProvider wraps GroqClient to provide legacy AIProvider interface.
type GroqLegacyProvider struct {
	*GroqClient
}

// NewGroqLegacyProvider creates a legacy provider wrapper.
func NewGroqLegacyProvider(apiKey, model string) *GroqLegacyProvider {
	return &GroqLegacyProvider{
		GroqClient: NewGroqClientWithModel(apiKey, model),
	}
}

// Query implements the legacy AIProvider interface.
func (g *GroqLegacyProvider) Query(prompt string) (string, error) {
	return g.GroqClient.Query(context.Background(), prompt)
}

// GenerateResponse implements the legacy AIProvider interface.
func (g *GroqLegacyProvider) GenerateResponse(prompt string) (string, error) {
	return g.GroqClient.GenerateResponse(context.Background(), prompt)
}
