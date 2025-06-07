package agent

import (
	"context"
	"fmt"
	"regexp"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// Agent defines the interface for all AI agents.
type Agent interface {
	Query(ctx context.Context, question string) (string, error)
	GenerateResponse(ctx context.Context, prompt string) (string, error)
	SetRole(role roles.RoleType) error
	SetContext(contextData interface{})
	SetProvider(provider ai.Provider)
}

// BaseAgent provides common functionality for all agents.
type BaseAgent struct {
	provider    ai.Provider
	role        roles.RoleType
	contextData interface{}
}

// SetRole sets the role for the agent.
func (a *BaseAgent) SetRole(role roles.RoleType) error {
	if !roles.ValidateRole(string(role)) {
		return fmt.Errorf("unsupported role: %s", role)
	}
	a.role = role
	return nil
}

// SetContext sets the context data for the agent.
func (a *BaseAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}

// SetProvider sets the AI provider for the agent.
func (a *BaseAgent) SetProvider(provider ai.Provider) {
	a.provider = provider
}

// validateRole validates that the agent has a proper role set.
func (a *BaseAgent) validateRole() error {
	if a.role == "" {
		return fmt.Errorf("agent role not set")
	}
	if !roles.ValidateRole(string(a.role)) {
		return fmt.Errorf("invalid role: %s", a.role)
	}
	return nil
}

// findFirstMatch finds the first regex match in a string.
func findFirstMatch(text, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	match := re.FindString(text)
	return match
}

// OllamaAgent is a basic implementation of the Agent interface using the Ollama provider.
type OllamaAgent struct {
	BaseAgent
}

func NewOllamaAgent(provider ai.Provider) *OllamaAgent {
	return &OllamaAgent{
		BaseAgent: BaseAgent{
			provider: provider,
		},
	}
}

func (a *OllamaAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt, ok := roles.RolePromptTemplate[a.role]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", a.role)
	}

	// Combine role prompt with question
	fullPrompt := fmt.Sprintf("%s\n\n%s", prompt, question)

	return a.provider.Query(ctx, fullPrompt)
}

func (a *OllamaAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	return a.provider.GenerateResponse(ctx, prompt)
}
