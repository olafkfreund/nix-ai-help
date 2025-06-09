package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpenAIClient represents a client for interacting with the OpenAI API.
type OpenAIClient struct {
	APIKey     string
	APIURL     string
	Model      string // Added model support
	HTTPClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client with the provided API key.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		APIKey:     apiKey,
		APIURL:     "https://api.openai.com/v1/chat/completions",
		Model:      "gpt-3.5-turbo", // default model
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewOpenAIClientWithModel creates an OpenAI client with a specific model.
func NewOpenAIClientWithModel(apiKey, model string) *OpenAIClient {
	client := NewOpenAIClient(apiKey)
	if model != "" {
		client.Model = model
	}
	return client
}

// Request represents a request to the OpenAI API.
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a message in the chat.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents a response from the OpenAI API.
type Response struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a choice in the response.
type Choice struct {
	Message Message `json:"message"`
}

// GenerateResponseFromMessages generates a response from the OpenAI API based on the provided messages.
func (client *OpenAIClient) GenerateResponseFromMessages(messages []Message) (string, error) {
	request := Request{
		Model:    client.Model, // Use the configured model
		Messages: messages,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", client.APIURL, bytes.NewBuffer(body))
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
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateResponse implements the legacy AIProvider interface for simple prompts.
func (client *OpenAIClient) GenerateResponse(prompt string) (string, error) {
	return client.Query(prompt)
}

// Query implements the AIProvider interface for OpenAIClient.
func (client *OpenAIClient) Query(prompt string) (string, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return client.GenerateResponseFromMessages(messages)
}

// CheckHealth checks if the OpenAI API is accessible and responding.
func (client *OpenAIClient) CheckHealth() error {
	// For OpenAI, we can check by making a simple request to the models endpoint
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openAI API not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("openAI API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// GetSelectedModel returns the currently selected model.
func (client *OpenAIClient) GetSelectedModel() string {
	return client.Model
}

// SetModel updates the selected model.
func (client *OpenAIClient) SetModel(modelName string) {
	client.Model = modelName
}
