package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai/roles"
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
	if !roles.ValidateRole(role) {
		return "", fmt.Errorf("unsupported role: %s", role)
	}
	prompt, ok := roles.RolePromptTemplate[roles.RoleType(role)]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", role)
	}
	// For now, just return the formatted prompt and input (simulate LLM call)
	return fmt.Sprintf("%s\n\n%s", prompt, input), nil
}

func (a *OllamaAgent) GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	// For now, just call Query (simulate different logic if needed)
	return a.Query(ctx, input, role, contextData)
}

func (a *OllamaAgent) SetRole(role string) {
	a.role = role
}

func (a *OllamaAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}
