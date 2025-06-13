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

// LlamaCppProvider implements the AIProvider interface for llamacpp.
type LlamaCppProvider struct {
	Endpoint    string
	Model       string
	Client      *http.Client
	lastPartial string // Store partial response for token limit cases
}

// NewLlamaCppProvider creates a new LlamaCppProvider.
func NewLlamaCppProvider(model string) *LlamaCppProvider {
	endpoint := os.Getenv("LLAMACPP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080/completion" // adjust to your llamacpp server endpoint
	}

	// Default timeout, will be updated if config is available
	timeout := 60 * time.Second

	return &LlamaCppProvider{
		Endpoint: endpoint,
		Model:    model,
		Client:   &http.Client{Timeout: timeout},
	}
}

// NewLlamaCppProviderWithModel creates a new LlamaCppProvider with a specific model configuration.
func NewLlamaCppProviderWithModel(providerConfig *config.AIProviderConfig, modelName string) (*LlamaCppProvider, error) {
	// Validate that the model exists in the provider configuration
	model, exists := providerConfig.Models[modelName]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in LlamaCpp provider configuration", modelName)
	}

	endpoint := providerConfig.BaseURL
	if endpoint == "" {
		endpoint = os.Getenv("LLAMACPP_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://localhost:8080/completion"
		}
	}

	// Default timeout of 60 seconds, will be configurable
	timeout := 60 * time.Second

	return &LlamaCppProvider{
		Endpoint: endpoint,
		Model:    model.Name,
		Client:   &http.Client{Timeout: timeout},
	}, nil
}

// CheckHealth checks if the LlamaCpp server is accessible and responding.
func (l *LlamaCppProvider) CheckHealth() error {
	// Try to make a simple request to check if the server is running
	req, err := http.NewRequest("GET", l.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("llamacpp server not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("llamacpp server returned error status: %d", resp.StatusCode)
	}

	return nil
}

// GetSelectedModel returns the currently selected model.
func (l *LlamaCppProvider) GetSelectedModel() string {
	return l.Model
}

// SetModel updates the selected model.
func (l *LlamaCppProvider) SetModel(modelName string) {
	l.Model = modelName
}

// SetTimeout updates the HTTP client timeout for llamacpp requests.
func (l *LlamaCppProvider) SetTimeout(timeout time.Duration) {
	l.Client.Timeout = timeout
}

// GetTimeout returns the current HTTP client timeout.
func (l *LlamaCppProvider) GetTimeout() time.Duration {
	return l.Client.Timeout
}

// llamacppRequest is the request format for llamacpp's API.
type llamacppRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
}

// llamacppResponse is the response format from llamacpp's API.
type llamacppResponse struct {
	Content string `json:"content"`
}

// Query sends a prompt to llamacpp and returns the response.
func (l *LlamaCppProvider) Query(prompt string) (string, error) {
	result, err := l.queryLlamaCpp(prompt, false)
	if err != nil {
		// Save partial result for recovery
		l.lastPartial = result
	}
	return result, err
}

// Context-aware Query method to implement new Provider interface
func (l *LlamaCppProvider) QueryContext(ctx context.Context, prompt string) (string, error) {
	result, err := l.queryLlamaCppWithContext(ctx, prompt, false)
	if err != nil {
		l.lastPartial = result
	}
	return result, err
}

// GenerateResponse is an alias for Query.
func (l *LlamaCppProvider) GenerateResponse(prompt string) (string, error) {
	return l.Query(prompt)
}

// GenerateResponseContext is the context-aware version
func (l *LlamaCppProvider) GenerateResponseContext(ctx context.Context, prompt string) (string, error) {
	return l.QueryContext(ctx, prompt)
}

// StreamResponse implements streaming for LlamaCpp API
func (l *LlamaCppProvider) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 100)

	go func() {
		defer close(responseChan)

		// LlamaCpp typically doesn't support native streaming, so we simulate it
		// by making the request and sending the response in chunks
		result, err := l.queryLlamaCppWithContext(ctx, prompt, true)

		if err != nil {
			l.lastPartial = result
			responseChan <- StreamResponse{
				Content:      result,
				Error:        err,
				Done:         true,
				PartialSaved: result != "",
			}
			return
		}

		// Simulate streaming by sending chunks of the response
		chunkSize := 50 // Send 50 characters at a time for smooth streaming effect
		for i := 0; i < len(result); i += chunkSize {
			end := i + chunkSize
			if end > len(result) {
				end = len(result)
			}

			chunk := result[i:end]
			isDone := end >= len(result)

			responseChan <- StreamResponse{
				Content: chunk,
				Done:    isDone,
			}

			// Small delay to simulate streaming
			if !isDone {
				select {
				case <-ctx.Done():
					l.lastPartial = result[:end]
					responseChan <- StreamResponse{
						Content:      result[:end],
						Error:        ctx.Err(),
						Done:         true,
						PartialSaved: true,
					}
					return
				case <-time.After(10 * time.Millisecond):
					// Continue
				}
			}
		}

		l.lastPartial = "" // Clear on successful completion
	}()

	return responseChan, nil
}

// GetPartialResponse returns the last partial response saved during errors
func (l *LlamaCppProvider) GetPartialResponse() string {
	return l.lastPartial
}

// queryLlamaCpp is the legacy implementation
func (l *LlamaCppProvider) queryLlamaCpp(prompt string, streaming bool) (string, error) {
	reqBody, _ := json.Marshal(llamacppRequest{Prompt: prompt, Model: l.Model})
	resp, err := l.Client.Post(l.Endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("llamacpp request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result llamacppResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("llamacpp decode failed: %w", err)
	}
	return result.Content, nil
}

// queryLlamaCppWithContext is the context-aware implementation
func (l *LlamaCppProvider) queryLlamaCppWithContext(ctx context.Context, prompt string, streaming bool) (string, error) {
	reqBody, _ := json.Marshal(llamacppRequest{Prompt: prompt, Model: l.Model})

	req, err := http.NewRequestWithContext(ctx, "POST", l.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llamacpp request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llamacpp returned status %d", resp.StatusCode)
	}

	var result llamacppResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("llamacpp decode failed: %w", err)
	}

	return result.Content, nil
}
