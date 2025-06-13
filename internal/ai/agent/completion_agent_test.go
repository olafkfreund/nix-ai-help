package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestCompletionAgent_NewCompletionAgent(t *testing.T) {
	mockProvider := &MockProvider{response: "test"}
	agent := NewCompletionAgent(mockProvider)

	if agent == nil {
		t.Fatal("NewCompletionAgent returned nil")
	}
	if agent.role != roles.RoleCompletion {
		t.Errorf("Expected role %v, got %v", roles.RoleCompletion, agent.role)
	}
}

func TestCompletionAgent_SetContext(t *testing.T) {
	mockProvider := &MockProvider{response: "test"}
	agent := NewCompletionAgent(mockProvider)

	ctx := &CompletionContext{
		ShellType:   "bash",
		CommandName: "nixai",
		Subcommands: []string{"ask", "build", "config"},
		Flags:       []string{"--help", "--verbose"},
	}

	err := agent.SetContext(ctx)
	if err != nil {
		t.Errorf("SetContext failed: %v", err)
	}

	if agent.contextData != ctx {
		t.Error("Context not set correctly")
	}
}

func TestCompletionAgent_SetContextInvalidType(t *testing.T) {
	mockProvider := &MockProvider{response: "test"}
	agent := NewCompletionAgent(mockProvider)

	err := agent.SetContext("invalid context")
	if err == nil {
		t.Error("Expected error for invalid context type")
	}
}

func TestCompletionAgent_GenerateCompletionScript(t *testing.T) {
	mockProvider := &MockProvider{response: "# Bash completion for nixai\n_nixai() { ... }"}
	agent := NewCompletionAgent(mockProvider)

	ctx := &CompletionContext{
		ShellType:        "bash",
		CommandName:      "nixai",
		Subcommands:      []string{"ask", "build"},
		Flags:            []string{"--help", "--verbose"},
		FlagDescriptions: map[string]string{"--help": "Show help", "--verbose": "Verbose output"},
	}
	agent.SetContext(ctx)

	response, err := agent.GenerateCompletionScript(context.Background(), "nixai", "bash")
	if err != nil {
		t.Errorf("GenerateCompletionScript failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestCompletionAgent_GenerateCompletionScriptNoContext(t *testing.T) {
	mockProvider := &MockProvider{response: "test"}
	agent := NewCompletionAgent(mockProvider)

	_, err := agent.GenerateCompletionScript(context.Background(), "nixai", "bash")
	if err == nil {
		t.Error("Expected error when context not set")
	}
}

func TestCompletionAgent_InstallCompletions(t *testing.T) {
	mockProvider := &MockProvider{response: "1. Install bash-completion\n2. Copy script to completion directory"}
	agent := NewCompletionAgent(mockProvider)

	ctx := &CompletionContext{
		ShellType:      "bash",
		PackageManager: "nix",
		InstallMethod:  "global",
		CompletionPath: "/etc/bash_completion.d",
	}
	agent.SetContext(ctx)

	response, err := agent.InstallCompletions(context.Background(), "bash", "/etc/bash_completion.d")
	if err != nil {
		t.Errorf("InstallCompletions failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestCompletionAgent_DiagnoseCompletionIssues(t *testing.T) {
	mockProvider := &MockProvider{response: "Issue: Completion not working\nSolution: Check shell configuration"}
	agent := NewCompletionAgent(mockProvider)

	ctx := &CompletionContext{
		ShellType:        "bash",
		CompletionErrors: []string{"command not found", "completion not loaded"},
		SystemInfo:       "NixOS 23.11",
	}
	agent.SetContext(ctx)

	response, err := agent.DiagnoseCompletionIssues(context.Background(), "Completions not working")
	if err != nil {
		t.Errorf("DiagnoseCompletionIssues failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestCompletionAgent_OptimizeCompletions(t *testing.T) {
	mockProvider := &MockProvider{response: "Recommendations:\n1. Enable caching\n2. Reduce suggestion count"}
	agent := NewCompletionAgent(mockProvider)

	ctx := &CompletionContext{
		CompletionSpeed: "fast",
		CacheEnabled:    false,
		MaxSuggestions:  100,
	}
	agent.SetContext(ctx)

	response, err := agent.OptimizeCompletions(context.Background(), "Current setup details")
	if err != nil {
		t.Errorf("OptimizeCompletions failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}
