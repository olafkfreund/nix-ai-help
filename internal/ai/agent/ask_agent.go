package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai/roles"
)

// AskAgent is a specialized agent for the 'ask' command.
type AskAgent struct {
	role        string
	contextData interface{}
}

func NewAskAgent() *AskAgent {
	return &AskAgent{role: string(roles.RoleAsk)}
}

func (a *AskAgent) Query(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	if !roles.ValidateRole(role) {
		return "", fmt.Errorf("unsupported role: %s", role)
	}
	prompt, ok := roles.RolePromptTemplate[roles.RoleType(role)]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", role)
	}
	return fmt.Sprintf("%s\n\n%s", prompt, input), nil
}

func (a *AskAgent) GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	return a.Query(ctx, input, role, contextData)
}

func (a *AskAgent) SetRole(role string) {
	a.role = role
}

func (a *AskAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}
