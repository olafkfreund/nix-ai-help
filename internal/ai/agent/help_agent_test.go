package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestNewHelpAgent(t *testing.T) {
	provider := &MockProvider{}
	agent := NewHelpAgent(provider)

	if agent == nil {
		t.Fatal("NewHelpAgent returned nil")
	}

	if agent.provider != provider {
		t.Error("Provider not set correctly")
	}

	if agent.role != roles.RoleHelp {
		t.Errorf("Expected role %s, got %s", roles.RoleHelp, agent.role)
	}
}

func TestHelpAgent_Query(t *testing.T) {
	tests := []struct {
		name         string
		question     string
		context      *HelpContext
		providerResp string
		expectError  bool
		expectInResp []string
	}{
		{
			name:         "basic help question",
			question:     "How do I rebuild my NixOS system?",
			providerResp: "Use nixos-rebuild switch",
			expectInResp: []string{"nixos-rebuild", "💡 **Additional Help Resources**"},
		},
		{
			name:     "help with context",
			question: "I'm having build errors",
			context: &HelpContext{
				UserLevel:    "Beginner",
				CurrentTask:  "System rebuild",
				ErrorContext: "hash mismatch error",
			},
			providerResp: "Try nixai diagnose for build errors",
			expectInResp: []string{"diagnose", "💡 **Additional Help Resources**"},
		},
		{
			name:         "command recommendation",
			question:     "What command helps with package search?",
			providerResp: "Use nixai search",
			expectInResp: []string{"search", "💡 **Additional Help Resources**"},
		},
		{
			name:     "workflow guidance",
			question: "How do I set up a development environment?",
			context: &HelpContext{
				UserGoal:    "Setup development environment",
				UserLevel:   "Intermediate",
				Environment: "NixOS 23.11",
			},
			providerResp: "Use nixai devenv for development setup",
			expectInResp: []string{"devenv", "💡 **Additional Help Resources**"},
		},
		{
			name:     "troubleshooting help",
			question: "My system won't boot after update",
			context: &HelpContext{
				UserLevel:    "Advanced",
				CurrentTask:  "System recovery",
				ErrorContext: "boot failure after nixos-rebuild",
			},
			providerResp: "Use nixai doctor to diagnose boot issues",
			expectInResp: []string{"doctor", "💡 **Additional Help Resources**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{
				response: tt.providerResp,
			}
			agent := NewHelpAgent(provider)

			if tt.context != nil {
				agent.SetHelpContext(tt.context)
			}

			response, err := agent.Query(context.Background(), tt.question)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			for _, expected := range tt.expectInResp {
				if !strings.Contains(response, expected) {
					t.Errorf("Expected response to contain %q, got: %s", expected, response)
				}
			}
		})
	}
}

func TestHelpAgent_GenerateResponse(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		providerResp string
		expectError  bool
		expectInResp []string
	}{
		{
			name:         "generate help response",
			prompt:       "Explain nixai commands",
			providerResp: "nixai provides various commands",
			expectInResp: []string{"nixai provides", "💡 **Additional Help Resources**"},
		},
		{
			name:         "generate workflow response",
			prompt:       "Explain development workflow",
			providerResp: "Setup devenv with nixai devenv",
			expectInResp: []string{"devenv", "💡 **Additional Help Resources**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{
				response: tt.providerResp,
			}
			agent := NewHelpAgent(provider)

			response, err := agent.GenerateResponse(context.Background(), tt.prompt)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			for _, expected := range tt.expectInResp {
				if !strings.Contains(response, expected) {
					t.Errorf("Expected response to contain %q, got: %s", expected, response)
				}
			}
		})
	}
}

func TestHelpAgent_BuildHelpPrompt(t *testing.T) {
	provider := &MockProvider{}
	agent := NewHelpAgent(provider)

	tests := []struct {
		name           string
		question       string
		context        *HelpContext
		expectInPrompt []string
	}{
		{
			name:     "basic prompt",
			question: "How do I use nixai?",
			expectInPrompt: []string{
				"Help Request",
				"How do I use nixai?",
				"Available nixai Commands",
				"nixai ask",
			},
		},
		{
			name:     "prompt with context",
			question: "Help with configuration",
			context: &HelpContext{
				UserGoal:     "Setup system",
				UserLevel:    "Beginner",
				CurrentTask:  "Initial setup",
				AvailableCmd: []string{"config", "doctor"},
				ErrorContext: "permission denied",
				Environment:  "NixOS 23.11",
			},
			expectInPrompt: []string{
				"Help Request",
				"User Goal**: Setup system",
				"Experience Level**: Beginner",
				"Current Task**: Initial setup",
				"Available Commands**: config, doctor",
				"Error Context**: permission denied",
				"Environment**: NixOS 23.11",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := agent.buildHelpPrompt(tt.question, tt.context)

			for _, expected := range tt.expectInPrompt {
				if !strings.Contains(prompt, expected) {
					t.Errorf("Expected prompt to contain %q, got: %s", expected, prompt)
				}
			}
		})
	}
}

func TestHelpAgent_BuildCommandReference(t *testing.T) {
	provider := &MockProvider{}
	agent := NewHelpAgent(provider)

	reference := agent.buildCommandReference()

	expectedCommands := []string{
		"nixai ask",
		"nixai diagnose",
		"nixai doctor",
		"nixai search",
		"nixai build",
		"nixai flake",
		"nixai interactive",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(reference, cmd) {
			t.Errorf("Expected command reference to contain %q", cmd)
		}
	}
}

func TestHelpAgent_GetAvailableCommands(t *testing.T) {
	provider := &MockProvider{}
	agent := NewHelpAgent(provider)

	commands := agent.GetAvailableCommands()

	expectedCommands := []string{
		"ask", "diagnose", "doctor", "search", "build", "flake",
		"interactive", "help", "config", "devenv",
	}

	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %q not found in available commands", expected)
		}
	}
}

func TestHelpAgent_SuggestCommand(t *testing.T) {
	tests := []struct {
		name         string
		userInput    string
		providerResp string
		expectError  bool
	}{
		{
			name:         "suggest for build issue",
			userInput:    "I have build errors",
			providerResp: "diagnose",
		},
		{
			name:         "suggest for package search",
			userInput:    "find a package",
			providerResp: "search",
		},
		{
			name:         "suggest for system health",
			userInput:    "check system health",
			providerResp: "doctor",
		},
		{
			name:         "suggest for learning",
			userInput:    "I want to learn NixOS",
			providerResp: "learn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{
				response: tt.providerResp,
			}
			agent := NewHelpAgent(provider)

			suggestion, err := agent.SuggestCommand(context.Background(), tt.userInput)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && suggestion != tt.providerResp {
				t.Errorf("Expected suggestion %q, got %q", tt.providerResp, suggestion)
			}
		})
	}
}

func TestHelpAgent_ExplainWorkflow(t *testing.T) {
	tests := []struct {
		name         string
		goal         string
		providerResp string
		expectError  bool
		expectInResp []string
	}{
		{
			name:         "development setup workflow",
			goal:         "Setup development environment",
			providerResp: "1. Use nixai devenv 2. Configure packages",
			expectInResp: []string{"devenv", "💡 **Additional Help Resources**"},
		},
		{
			name:         "system troubleshooting workflow",
			goal:         "Fix system issues",
			providerResp: "1. Use nixai doctor 2. Use nixai diagnose",
			expectInResp: []string{"doctor", "diagnose", "💡 **Additional Help Resources**"},
		},
		{
			name:         "package management workflow",
			goal:         "Find and install packages",
			providerResp: "1. Use nixai search 2. Configure in configuration.nix",
			expectInResp: []string{"search", "💡 **Additional Help Resources**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{
				response: tt.providerResp,
			}
			agent := NewHelpAgent(provider)

			response, err := agent.ExplainWorkflow(context.Background(), tt.goal)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			for _, expected := range tt.expectInResp {
				if !strings.Contains(response, expected) {
					t.Errorf("Expected response to contain %q, got: %s", expected, response)
				}
			}
		})
	}
}

func TestHelpAgent_SetRole(t *testing.T) {
	provider := &MockProvider{}
	agent := NewHelpAgent(provider)

	// Test setting valid role
	err := agent.SetRole(roles.RoleHelp)
	if err != nil {
		t.Errorf("Unexpected error setting valid role: %v", err)
	}

	// Test setting invalid role
	err = agent.SetRole(roles.RoleType("invalid"))
	if err == nil {
		t.Error("Expected error setting invalid role")
	}
}

func TestHelpAgent_SetContext(t *testing.T) {
	provider := &MockProvider{}
	agent := NewHelpAgent(provider)

	ctx := &HelpContext{
		UserGoal:  "Test goal",
		UserLevel: "Beginner",
	}

	agent.SetHelpContext(ctx)

	if agent.contextData != ctx {
		t.Error("Context not set correctly")
	}

	retrievedCtx := agent.getHelpContextFromData()
	if retrievedCtx != ctx {
		t.Error("Retrieved context does not match set context")
	}
}

func TestHelpAgent_ValidationErrors(t *testing.T) {
	provider := &MockProvider{}
	agent := &HelpAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			// Intentionally not setting role to test validation
		},
	}

	_, err := agent.Query(context.Background(), "test")
	if err == nil {
		t.Error("Expected validation error for missing role")
	}

	_, err = agent.GenerateResponse(context.Background(), "test")
	if err == nil {
		t.Error("Expected validation error for missing role")
	}

	_, err = agent.SuggestCommand(context.Background(), "test")
	if err == nil {
		t.Error("Expected validation error for missing role")
	}

	_, err = agent.ExplainWorkflow(context.Background(), "test")
	if err == nil {
		t.Error("Expected validation error for missing role")
	}
}
