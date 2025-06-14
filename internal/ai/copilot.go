package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CopilotClient represents a client for interacting with GitHub Copilot's OpenAI-compatible API.
type CopilotClient struct {
	APIKey     string
	APIURL     string
	Model      string
	HTTPClient *http.Client
}

// NewCopilotClient creates a new GitHub Copilot client with the provided API key.
func NewCopilotClient(apiKey string) *CopilotClient {
	return &CopilotClient{
		APIKey:     apiKey,
		APIURL:     "https://api.githubcopilot.com/chat/completions",
		Model:      "gpt-4", // Default model for GitHub Copilot
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewCopilotClientWithModel creates a Copilot client with a specific model.
func NewCopilotClientWithModel(apiKey, model string) *CopilotClient {
	client := NewCopilotClient(apiKey)
	if model != "" {
		client.Model = model
	}
	return client
}

// CopilotRequest represents a request to the GitHub Copilot API (OpenAI-compatible).
type CopilotRequest struct {
	Model    string           `json:"model"`
	Messages []CopilotMessage `json:"messages"`
	Stream   bool             `json:"stream,omitempty"`
}

// CopilotMessage represents a message in the chat.
type CopilotMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CopilotResponse represents a response from the GitHub Copilot API.
type CopilotResponse struct {
	Choices []CopilotChoice `json:"choices"`
	Usage   CopilotUsage    `json:"usage,omitempty"`
}

// CopilotChoice represents a choice in the response.
type CopilotChoice struct {
	Message CopilotMessage `json:"message"`
	Index   int            `json:"index"`
}

// CopilotUsage represents token usage information.
type CopilotUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CopilotStreamResponse represents a streaming response from GitHub Copilot API.
type CopilotStreamResponse struct {
	Choices []CopilotStreamChoice `json:"choices"`
}

// CopilotStreamChoice represents a choice in the streaming response.
type CopilotStreamChoice struct {
	Delta CopilotStreamDelta `json:"delta"`
}

// CopilotStreamDelta represents the delta content in streaming.
type CopilotStreamDelta struct {
	Content string `json:"content"`
}

// GenerateResponseFromMessages generates a response from the GitHub Copilot API based on the provided messages.
func (client *CopilotClient) GenerateResponseFromMessages(messages []CopilotMessage) (string, error) {
	request := CopilotRequest{
		Model:    client.Model,
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
	req.Header.Set("User-Agent", "nixai/1.0")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateResponseFromMessagesContext generates a response with context support.
func (client *CopilotClient) GenerateResponseFromMessagesContext(ctx context.Context, messages []CopilotMessage) (string, error) {
	request := CopilotRequest{
		Model:    client.Model,
		Messages: messages,
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
	req.Header.Set("User-Agent", "nixai/1.0")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// StreamResponseFromMessages generates a streaming response from the GitHub Copilot API.
func (client *CopilotClient) StreamResponseFromMessages(ctx context.Context, messages []CopilotMessage) (<-chan StreamResponse, error) {
	request := CopilotRequest{
		Model:    client.Model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", client.APIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "nixai/1.0")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	responseChan := make(chan StreamResponse, 10)

	go func() {
		defer close(responseChan)
		defer func() { _ = resp.Body.Close() }()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and non-data lines
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Remove "data: " prefix
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				responseChan <- StreamResponse{
					Done: true,
				}
				return
			}

			var streamResp CopilotStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				responseChan <- StreamResponse{
					Error: fmt.Errorf("failed to decode stream response: %w", err),
					Done:  true,
				}
				return
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				responseChan <- StreamResponse{
					Content: streamResp.Choices[0].Delta.Content,
					Done:    false,
				}
			}
		}

		if err := scanner.Err(); err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("error reading stream: %w", err),
				Done:  true,
			}
		}
	}()

	return responseChan, nil
}

// Query implements the AIProvider interface (legacy signature for compatibility).
func (client *CopilotClient) Query(prompt string) (string, error) {
	messages := []CopilotMessage{{Role: "user", Content: prompt}}
	return client.GenerateResponseFromMessages(messages)
}

// QueryWithContext implements the Provider interface with context support for CopilotClient.
func (client *CopilotClient) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	messages := []CopilotMessage{{Role: "user", Content: prompt}}
	return client.GenerateResponseFromMessagesContext(ctx, messages)
}

// GenerateResponse implements the Provider interface with context support for CopilotClient.
func (client *CopilotClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return client.QueryWithContext(ctx, prompt)
}

// StreamResponse implements the Provider interface for streaming responses.
func (client *CopilotClient) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	messages := []CopilotMessage{{Role: "user", Content: prompt}}
	return client.StreamResponseFromMessages(ctx, messages)
}

// GetPartialResponse returns empty string as this implementation doesn't track partial responses.
func (client *CopilotClient) GetPartialResponse() string {
	return ""
}

// SetTimeout sets the HTTP client timeout.
func (client *CopilotClient) SetTimeout(timeout time.Duration) {
	client.HTTPClient.Timeout = timeout
}
