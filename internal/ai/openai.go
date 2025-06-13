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

// StreamRequest represents a streaming request to the OpenAI API.
type StreamRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// StreamResponse represents a streaming response from OpenAI API.
type OpenAIStreamResponse struct {
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a choice in the streaming response.
type StreamChoice struct {
	Delta StreamDelta `json:"delta"`
}

// StreamDelta represents the delta content in streaming.
type StreamDelta struct {
	Content string `json:"content"`
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

// Query implements the AIProvider interface (legacy signature for compatibility).
func (client *OpenAIClient) Query(prompt string) (string, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return client.GenerateResponseFromMessages(messages)
}

// QueryWithContext implements the Provider interface with context support for OpenAIClient.
func (client *OpenAIClient) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return client.GenerateResponseFromMessagesContext(ctx, messages)
}

// GenerateResponse implements the Provider interface with context support for OpenAIClient.
func (client *OpenAIClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return client.QueryWithContext(ctx, prompt)
}

// Legacy methods for backward compatibility

// QueryLegacy implements the legacy AIProvider interface for OpenAIClient.
func (client *OpenAIClient) QueryLegacy(prompt string) (string, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return client.GenerateResponseFromMessages(messages)
}

// GenerateResponseLegacy implements the legacy AIProvider interface for simple prompts.
func (client *OpenAIClient) GenerateResponseLegacy(prompt string) (string, error) {
	return client.QueryLegacy(prompt)
}

// GenerateResponseFromMessagesContext generates a response from the OpenAI API with context support.
func (client *OpenAIClient) GenerateResponseFromMessagesContext(ctx context.Context, messages []Message) (string, error) {
	request := Request{
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

// StreamResponse implements streaming for OpenAI API
func (client *OpenAIClient) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 100)

	go func() {
		defer close(responseChan)

		messages := []Message{{Role: "user", Content: prompt}}
		request := StreamRequest{
			Model:    client.Model,
			Messages: messages,
			Stream:   true,
		}

		body, err := json.Marshal(request)
		if err != nil {
			responseChan <- StreamResponse{Error: fmt.Errorf("failed to marshal request: %w", err), Done: true}
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", client.APIURL, bytes.NewBuffer(body))
		if err != nil {
			responseChan <- StreamResponse{Error: fmt.Errorf("failed to create request: %w", err), Done: true}
			return
		}

		req.Header.Set("Authorization", "Bearer "+client.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		resp, err := client.HTTPClient.Do(req)
		if err != nil {
			responseChan <- StreamResponse{Error: fmt.Errorf("openai request failed: %w", err), Done: true}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			responseChan <- StreamResponse{Error: fmt.Errorf("openai returned status %d", resp.StatusCode), Done: true}
			return
		}

		// Read Server-Sent Events
		scanner := bufio.NewScanner(resp.Body)
		var fullResponse strings.Builder

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and non-data lines
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Extract JSON data from "data: " prefix
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				responseChan <- StreamResponse{Content: "", Done: true}
				return
			}

			var streamResp OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue // Skip malformed responses
			}

			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					fullResponse.WriteString(content)
					responseChan <- StreamResponse{
						Content: content,
						Done:    false,
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			responseChan <- StreamResponse{
				Content:      fullResponse.String(),
				Error:        err,
				Done:         true,
				PartialSaved: fullResponse.Len() > 0,
			}
		}
	}()

	return responseChan, nil
}

// GetPartialResponse returns empty for OpenAI as partial responses are handled in streaming
func (client *OpenAIClient) GetPartialResponse() string {
	return ""
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
