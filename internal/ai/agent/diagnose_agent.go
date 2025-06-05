package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai/roles"
)

// DiagnoseAgent is a specialized agent for the 'diagnose' command.
type DiagnoseAgent struct {
	role        string
	contextData interface{}
}

func NewDiagnoseAgent() *DiagnoseAgent {
	return &DiagnoseAgent{role: string(roles.RoleDiagnose)}
}

func (a *DiagnoseAgent) Query(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	if !roles.ValidateRole(role) {
		return "", fmt.Errorf("unsupported role: %s", role)
	}
	prompt, ok := roles.RolePromptTemplate[roles.RoleType(role)]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", role)
	}
	return fmt.Sprintf("%s\n\n%s", prompt, input), nil
}

func (a *DiagnoseAgent) GenerateResponse(ctx context.Context, input string, role string, contextData interface{}) (string, error) {
	return a.Query(ctx, input, role, contextData)
}

func (a *DiagnoseAgent) SetRole(role string) {
	a.role = role
}

func (a *DiagnoseAgent) SetContext(contextData interface{}) {
	a.contextData = contextData
}
