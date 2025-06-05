package agent

import (
	"context"
)

// Agent defines the interface for all AI agents.
type Agent interface {
	Query(ctx context.Context, input string, role string, contextData interface{}) (string, error)
	GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error)
	SetRole(role string)
	SetContext(contextData interface{})
}

// OllamaAgent is a basic implementation of the Agent interface using the Ollama provider.
type OllamaAgent struct {
	role        string
	contextData interface{}
}

func NewOllamaAgent() *OllamaAgent {
	return &OllamaAgent{}
}

func (a *OllamaAgent) Query(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	// TODO: Integrate with internal/ai/ollama provider logic
	return "[OllamaAgent: Query not yet implemented]", nil
}

func (a *OllamaAgent) GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	// TODO: Integrate with internal/ai/ollama provider logic
	return "[OllamaAgent: GenerateResponse not yet implemented]", nil
}

func (a *OllamaAgent) SetRole(role string) {
	a.role = role
}

func (a *OllamaAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}
