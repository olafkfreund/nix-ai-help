package agent

import (
	"context"
	"errors"
	"testing"
)

func TestNewCommunityAgent(t *testing.T) {
	provider := &MockProvider{}
	agent := NewCommunityAgent(provider)

	if agent == nil {
		t.Fatal("Expected agent to be created")
	}

	if agent.provider != provider {
		t.Error("Expected provider to be set")
	}
}

func TestCommunityAgent_SetContext(t *testing.T) {
	agent := NewCommunityAgent(&MockProvider{})

	// Test valid context
	ctx := &CommunityContext{
		UserLevel:         "intermediate",
		InterestAreas:     []string{"packaging", "documentation"},
		CurrentProjects:   []string{"nixos-config"},
		CommunityGoals:    []string{"contribute to nixpkgs"},
		ExperienceLevel:   "beginner",
		PreferredChannels: []string{"discord", "forum"},
		ContributionType:  "code",
	}

	err := agent.SetContext(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test invalid context
	err = agent.SetContext("invalid")
	if err == nil {
		t.Error("Expected error for invalid context type")
	}

	// Test nil context
	err = agent.SetContext(nil)
	if err != nil {
		t.Errorf("Expected no error for nil context, got: %v", err)
	}
}

func TestCommunityAgent_Query(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name:     "successful community query",
			input:    "How do I contribute to nixpkgs?",
			provider: &MockProvider{response: "Here's how to contribute to nixpkgs..."},
			wantErr:  false,
		},
		{
			name:     "query about community resources",
			input:    "Where can I find NixOS community help?",
			provider: &MockProvider{response: "You can find help in these places..."},
			wantErr:  false,
		},
		{
			name:     "provider error",
			input:    "How do I join the community?",
			provider: &MockProvider{err: errors.New("provider error")},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.Query(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_GenerateResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name:     "generate community guidance",
			input:    "I want to contribute but don't know where to start",
			provider: &MockProvider{response: "Here's how to get started..."},
			wantErr:  false,
		},
		{
			name:     "generate project recommendations",
			input:    "Suggest some projects for a beginner",
			provider: &MockProvider{response: "Here are some beginner projects..."},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.GenerateResponse(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_FindCommunityResources(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		topic        string
		provider     *MockProvider
		wantErr      bool
	}{
		{
			name:         "find documentation resources",
			resourceType: "documentation",
			topic:        "packaging",
			provider:     &MockProvider{response: "Documentation resources for packaging..."},
			wantErr:      false,
		},
		{
			name:         "find communication channels",
			resourceType: "communication",
			topic:        "general help",
			provider:     &MockProvider{response: "Communication channels for help..."},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.FindCommunityResources(tt.resourceType, tt.topic)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindCommunityResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_GuideContribution(t *testing.T) {
	tests := []struct {
		name             string
		contributionType string
		projectArea      string
		provider         *MockProvider
		wantErr          bool
	}{
		{
			name:             "guide code contribution",
			contributionType: "code",
			projectArea:      "nixpkgs",
			provider:         &MockProvider{response: "Guide for code contribution to nixpkgs..."},
			wantErr:          false,
		},
		{
			name:             "guide documentation contribution",
			contributionType: "documentation",
			projectArea:      "nixos-wiki",
			provider:         &MockProvider{response: "Guide for documentation contribution..."},
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.GuideContribution(tt.contributionType, tt.projectArea)

			if (err != nil) != tt.wantErr {
				t.Errorf("GuideContribution() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_RecommendProjects(t *testing.T) {
	tests := []struct {
		name       string
		interests  []string
		skillLevel string
		provider   *MockProvider
		wantErr    bool
	}{
		{
			name:       "recommend for beginner",
			interests:  []string{"packaging", "development"},
			skillLevel: "beginner",
			provider:   &MockProvider{response: "Beginner project recommendations..."},
			wantErr:    false,
		},
		{
			name:       "recommend for advanced user",
			interests:  []string{"kernel", "security"},
			skillLevel: "advanced",
			provider:   &MockProvider{response: "Advanced project recommendations..."},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.RecommendProjects(tt.interests, tt.skillLevel)

			if (err != nil) != tt.wantErr {
				t.Errorf("RecommendProjects() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_ExplainCommunityChannels(t *testing.T) {
	tests := []struct {
		name     string
		purpose  string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name:     "explain channels for getting help",
			purpose:  "getting help",
			provider: &MockProvider{response: "Channels for getting help..."},
			wantErr:  false,
		},
		{
			name:     "explain channels for contributing",
			purpose:  "contributing",
			provider: &MockProvider{response: "Channels for contributing..."},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.ExplainCommunityChannels(tt.purpose)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExplainCommunityChannels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_PlanCommunityInvolvement(t *testing.T) {
	tests := []struct {
		name           string
		goals          []string
		timeCommitment string
		provider       *MockProvider
		wantErr        bool
	}{
		{
			name:           "plan involvement for contributor",
			goals:          []string{"contribute packages", "help newcomers"},
			timeCommitment: "5 hours per week",
			provider:       &MockProvider{response: "Plan for contributor involvement..."},
			wantErr:        false,
		},
		{
			name:           "plan minimal involvement",
			goals:          []string{"stay updated"},
			timeCommitment: "1 hour per week",
			provider:       &MockProvider{response: "Plan for minimal involvement..."},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(tt.provider)

			_, err := agent.PlanCommunityInvolvement(tt.goals, tt.timeCommitment)

			if (err != nil) != tt.wantErr {
				t.Errorf("PlanCommunityInvolvement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommunityAgent_formatCommunityContext(t *testing.T) {
	tests := []struct {
		name    string
		context interface{}
		want    string
	}{
		{
			name:    "nil context",
			context: nil,
			want:    "No specific community context provided.",
		},
		{
			name: "comprehensive community context",
			context: &CommunityContext{
				UserLevel:         "intermediate",
				InterestAreas:     []string{"packaging", "development"},
				CurrentProjects:   []string{"personal-config", "team-setup"},
				CommunityGoals:    []string{"contribute to nixpkgs", "mentor newcomers"},
				ExperienceLevel:   "experienced",
				PreferredChannels: []string{"discord", "github"},
				ContributionType:  "code",
			},
			want: "Community Profile:",
		},
		{
			name: "minimal context",
			context: &CommunityContext{
				UserLevel: "beginner",
			},
			want: "Community Profile:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewCommunityAgent(&MockProvider{})
			agent.SetContext(tt.context)

			result := agent.formatCommunityContext()

			if !contains(result, tt.want) {
				t.Errorf("formatCommunityContext() = %v, want to contain %v", result, tt.want)
			}
		})
	}
}

func TestCommunityAgent_SetRole(t *testing.T) {
	agent := NewCommunityAgent(&MockProvider{})

	agent.SetRole("custom-role")

	if agent.role != "custom-role" {
		t.Errorf("Expected role to be 'custom-role', got %s", agent.role)
	}
}

func TestCommunityAgent_ValidationErrors(t *testing.T) {
	// Test with nil provider
	agent := NewCommunityAgent(nil)

	_, err := agent.Query(context.Background(), "test")
	if err == nil {
		t.Error("Expected error with nil provider")
	}

	_, err = agent.GenerateResponse(context.Background(), "test")
	if err == nil {
		t.Error("Expected error with nil provider")
	}

	_, err = agent.FindCommunityResources("docs", "packaging")
	if err == nil {
		t.Error("Expected error with nil provider")
	}
}
