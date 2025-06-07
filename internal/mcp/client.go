package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MCPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMCPClient(baseURL string) *MCPClient {
	return &MCPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *MCPClient) QueryDocumentation(query string, sources ...string) (string, error) {
	var requestBody interface{}

	if len(sources) > 0 {
		requestBody = map[string]interface{}{
			"query":   query,
			"sources": sources,
		}
	} else {
		requestBody = map[string]string{"query": query}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/query", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Read the response body for debugging
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP status %d: %s", resp.StatusCode, string(body))
	}

	var responseBody struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	return responseBody.Result, nil
}

// OptionCompletion queries the MCP server for option completions given a prefix.
func (c *MCPClient) OptionCompletion(prefix string) ([]string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name":      "complete_nixos_option",
			"arguments": map[string]interface{}{"prefix": prefix},
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Result struct {
			Options []string `json:"options"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Result.Options, nil
}
