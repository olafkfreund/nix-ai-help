package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// OllamaProvider implements the new Provider interface for Ollama.
type OllamaProvider struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

// NewOllamaProvider creates a new OllamaProvider.
func NewOllamaProvider(model string) *OllamaProvider {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate"
	}

	if model == "" {
		model = "llama3"
	}

	return &OllamaProvider{
		Endpoint: endpoint,
		Model:    model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ollamaRequest is the request format for Ollama's API.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaResponse is the response format from Ollama's API.
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// Query sends a prompt to Ollama with context support.
// This implements the new Provider interface.
func (o *OllamaProvider) Query(ctx context.Context, prompt string) (string, error) {
	return o.queryWithContext(ctx, prompt)
}

// GenerateResponse sends a prompt to Ollama with context support.
// This implements the new Provider interface.
func (o *OllamaProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return o.queryWithContext(ctx, prompt)
}

// queryWithContext is the internal implementation that handles the actual API call.
func (o *OllamaProvider) queryWithContext(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("ollama error: %s", result.Error)
	}

	return result.Response, nil
}

// HealthCheck checks if the Ollama server is running and accessible
func (o *OllamaProvider) HealthCheck() error {
	// Create a simple health check request
	healthURL := strings.Replace(o.Endpoint, "/api/generate", "/api/tags", 1)

	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := o.Client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama server not accessible: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama server returned status %d", resp.StatusCode)
	}

	return nil
}

// SetModel allows changing the model after creation
func (o *OllamaProvider) SetModel(model string) {
	if model != "" {
		o.Model = model
	}
}

// SetTimeout updates the HTTP client timeout for Ollama requests.
func (o *OllamaProvider) SetTimeout(timeout time.Duration) {
	o.Client.Timeout = timeout
}

// GetTimeout returns the current HTTP client timeout.
func (o *OllamaProvider) GetTimeout() time.Duration {
	return o.Client.Timeout
}

// Legacy Provider Wrapper for backward compatibility
type OllamaLegacyProvider struct {
	*OllamaProvider
}

// NewOllamaLegacyProvider creates a legacy provider wrapper.
func NewOllamaLegacyProvider(model string) *OllamaLegacyProvider {
	return &OllamaLegacyProvider{
		OllamaProvider: NewOllamaProvider(model),
	}
}

// Query implements the legacy AIProvider interface.
func (o *OllamaLegacyProvider) Query(prompt string) (string, error) {
	return o.OllamaProvider.Query(context.Background(), prompt)
}

// GenerateResponse implements the legacy AIProvider interface.
func (o *OllamaLegacyProvider) GenerateResponse(prompt string) (string, error) {
	return o.OllamaProvider.GenerateResponse(context.Background(), prompt)
}
