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
	HTTPClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client with the provided API key.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		APIKey:     apiKey,
		APIURL:     "https://api.openai.com/v1/chat/completions",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
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

// GenerateResponse generates a response from the OpenAI API based on the provided messages.
func (client *OpenAIClient) GenerateResponse(messages []Message) (string, error) {
	request := Request{
		Model:    "gpt-3.5-turbo",
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
	defer resp.Body.Close()

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

// Query implements the AIProvider interface for OpenAIClient.
func (client *OpenAIClient) Query(prompt string) (string, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return client.GenerateResponse(messages)
}
