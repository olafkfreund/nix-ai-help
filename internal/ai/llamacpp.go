package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
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
	defer resp.Body.Close()

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
