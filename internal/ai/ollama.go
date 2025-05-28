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
	cmd := exec.Command("ollama", "run", o.modelPath, prompt)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing command: %s, stderr: %s", err.Error(), stderr.String())
	}

	return out.String(), nil
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
