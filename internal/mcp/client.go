package mcp

import (
	"bytes"
	"encoding/json"
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
	requestBody, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/query", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	var responseBody struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	return responseBody.Result, nil
}
