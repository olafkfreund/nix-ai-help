package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// OllamaProvider implements the AIProvider interface for the Ollama model.
type OllamaProvider struct {
	modelPath string
}

// NewOllamaProvider creates a new instance of OllamaProvider.
func NewOllamaProvider(modelPath string) *OllamaProvider {
	return &OllamaProvider{modelPath: modelPath}
}

// Query sends a query to the Ollama model and returns the response.
func (o *OllamaProvider) Query(prompt string) (string, error) {
	// Use stdin to pass the prompt to avoid command line length limitations
	cmd := exec.Command("ollama", "run", o.modelPath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewBufferString(prompt)

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing command: %s, stderr: %s", err.Error(), stderr.String())
	}

	response := out.String()

	// Clean up the response by removing the prompt echo if present
	// Ollama sometimes echoes the prompt back
	if len(response) > len(prompt) && response[:len(prompt)] == prompt {
		response = response[len(prompt):]
	}

	return response, nil
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
