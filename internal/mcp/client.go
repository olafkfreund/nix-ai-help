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

func (c *MCPClient) QueryDocumentation(query string) (string, error) {
	fmt.Printf("DEBUG: MCP client querying for: %s\n", query)
	requestBody, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: Making POST request to: %s/query\n", c.baseURL)
	resp, err := c.httpClient.Post(c.baseURL+"/query", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("DEBUG: HTTP request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: HTTP response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// Read the response body for debugging
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("DEBUG: Non-200 response body: %s\n", string(body))
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	var responseBody struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		fmt.Printf("DEBUG: JSON decode error: %v\n", err)
		return "", err
	}

	fmt.Printf("DEBUG: MCP client received %d characters: %s...\n", len(responseBody.Result), responseBody.Result[:minInt(200, len(responseBody.Result))])
	return responseBody.Result, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
