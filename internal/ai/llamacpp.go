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

// LlamaCppProvider implements the AIProvider interface for llamacpp.
type LlamaCppProvider struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

// NewLlamaCppProvider creates a new LlamaCppProvider.
func NewLlamaCppProvider(model string) *LlamaCppProvider {
	endpoint := os.Getenv("LLAMACPP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080/completion" // adjust to your llamacpp server endpoint
	}
	return &LlamaCppProvider{
		Endpoint: endpoint,
		Model:    model,
		Client:   &http.Client{Timeout: 60 * time.Second},
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

	return &LlamaCppProvider{
		Endpoint: endpoint,
		Model:    model.Name,
		Client:   &http.Client{Timeout: 60 * time.Second},
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

// GenerateResponse is an alias for Query.
func (l *LlamaCppProvider) GenerateResponse(prompt string) (string, error) {
	return l.Query(prompt)
}
