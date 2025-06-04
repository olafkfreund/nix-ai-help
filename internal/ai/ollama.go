package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// OllamaProvider implements the AIProvider interface for the Ollama model.
type OllamaProvider struct {
	modelPath string
	host      string
	client    *http.Client
}

// OllamaGenerateRequest represents the request structure for Ollama's generate API
type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaGenerateResponse represents the response structure for Ollama's generate API
type OllamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaProvider creates a new instance of OllamaProvider.
func NewOllamaProvider(modelPath string) *OllamaProvider {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	return &OllamaProvider{
		modelPath: modelPath,
		host:      host,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Query sends a query to the Ollama model and returns the response.
func (o *OllamaProvider) Query(prompt string) (string, error) {
	// Build the API URL
	apiURL := fmt.Sprintf("%s/api/generate", o.host)

	// Create the request payload
	requestBody := OllamaGenerateRequest{
		Model:  o.modelPath,
		Prompt: prompt,
		Stream: false, // We want a single response, not streaming
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create and send the HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama server: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama server returned non-200 response: %s\n%s", resp.Status, string(body))
	}

	// Parse the response
	var responseBody OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return responseBody.Response, nil
}

// ModelInfo returns information about the model.
func (o *OllamaProvider) ModelInfo() (string, error) {
	info := map[string]string{
		"model": "Ollama",
		"path":  o.modelPath,
	}
	infoJSON, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(infoJSON), nil
}
