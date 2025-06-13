package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestLearnAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "learn agent response"}
	agent := NewLearnAgent(mockProvider)

	input := "How do I learn Nix?"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "learn agent")
}

func TestLearnAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "learn agent response"}
	agent := NewLearnAgent(mockProvider)

	input := "Explain Nix concepts"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "learn agent response")
}

func TestLearnAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewLearnAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RoleLearn)
	require.NoError(t, err)
	require.Equal(t, roles.RoleLearn, agent.role)

	// Test setting context
	learnCtx := &LearnContext{SkillLevel: "beginner"}
	agent.SetContext(learnCtx)
	require.Equal(t, learnCtx, agent.contextData)
}

func TestLearnAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewLearnAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestLearnContext_Formatting(t *testing.T) {
	learnCtx := &LearnContext{
		Topic:             "nix language",
		SkillLevel:        "intermediate",
		LearningGoal:      "understand flakes",
		PreferredStyle:    "hands-on",
		TimeAvailable:     "thorough",
		CurrentKnowledge:  []string{"basic nix", "nixos configuration"},
		LearningPath:      []string{"nix language", "flakes", "packaging"},
		Prerequisites:     []string{"linux basics", "package management"},
		PracticeExercises: []string{"write derivation", "create flake"},
		ResourceLinks:     []string{"nix.dev", "nixos.org"},
		ExampleCode:       "{ pkgs, ... }: { }",
		CommonMistakes:    []string{"missing imports", "wrong syntax"},
		NextSteps:         []string{"advanced packaging", "custom nixos module"},
	}

	// Test that context can be created and has expected fields
	require.Equal(t, "nix language", learnCtx.Topic)
	require.Equal(t, "intermediate", learnCtx.SkillLevel)
	require.Equal(t, "understand flakes", learnCtx.LearningGoal)
	require.Equal(t, "hands-on", learnCtx.PreferredStyle)
	require.Equal(t, "thorough", learnCtx.TimeAvailable)
	require.Len(t, learnCtx.CurrentKnowledge, 2)
	require.Len(t, learnCtx.LearningPath, 3)
	require.Len(t, learnCtx.Prerequisites, 2)
	require.Len(t, learnCtx.PracticeExercises, 2)
	require.Len(t, learnCtx.ResourceLinks, 2)
	require.NotEmpty(t, learnCtx.ExampleCode)
	require.Len(t, learnCtx.CommonMistakes, 2)
	require.Len(t, learnCtx.NextSteps, 2)
}
