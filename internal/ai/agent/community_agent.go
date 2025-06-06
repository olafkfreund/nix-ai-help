package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// CommunityAgent handles community-related queries and assistance.
type CommunityAgent struct {
	provider    ai.Provider
	role        string
	contextData interface{}
}

// CommunityContext provides context for community-related operations.
type CommunityContext struct {
	UserLevel         string            // beginner, intermediate, advanced
	InterestAreas     []string          // packaging, development, documentation, etc.
	CurrentProjects   []string          // projects user is working on
	CommunityGoals    []string          // what user wants to achieve in community
	ExperienceLevel   string            // experience with open source
	PreferredChannels []string          // preferred communication channels
	ContributionType  string            // code, documentation, testing, etc.
	Metadata          map[string]string // additional community context
}

// NewCommunityAgent creates a new community agent.
func NewCommunityAgent(provider ai.Provider) *CommunityAgent {
	return &CommunityAgent{
		provider: provider,
		role:     string(roles.RoleCommunity),
	}
}

// Query processes community-related queries.
func (a *CommunityAgent) Query(ctx context.Context, input string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	// Build community-specific prompt
	prompt := a.buildCommunityPrompt(input)

	// Use provider to generate response
	return a.provider.Query(ctx, prompt)
}

// GenerateResponse generates a response for community assistance.
func (a *CommunityAgent) GenerateResponse(ctx context.Context, input string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	// Enhance prompt with role and context
	prompt := a.enhancePromptWithRole(input)

	// Generate response using provider
	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate community response: %w", err)
	}

	return a.formatCommunityResponse(response), nil
}

// SetRole sets the role for the agent.
func (a *CommunityAgent) SetRole(role string) {
	a.role = role
}

// SetContext sets the context for community operations.
func (a *CommunityAgent) SetContext(context interface{}) error {
	if context == nil {
		a.contextData = nil
		return nil
	}

	if communityCtx, ok := context.(*CommunityContext); ok {
		a.contextData = communityCtx
		return nil
	}

	return fmt.Errorf("invalid context type for CommunityAgent")
}

// FindCommunityResources helps users find relevant community resources.
func (a *CommunityAgent) FindCommunityResources(resourceType string, topic string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As a NixOS Community Specialist, help find community resources for:
Resource Type: %s
Topic: %s

%s

Provide specific recommendations with links and descriptions.`,
		resourceType, topic, a.formatCommunityContext())

	ctx := context.Background()
	return a.provider.Query(ctx, prompt)
}

// GuideContribution provides guidance for contributing to NixOS projects.
func (a *CommunityAgent) GuideContribution(contributionType string, projectArea string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As a NixOS Community Specialist, provide contribution guidance for:
Contribution Type: %s
Project Area: %s

%s

Include step-by-step instructions and best practices.`,
		contributionType, projectArea, a.formatCommunityContext())

	ctx := context.Background()
	return a.provider.Query(ctx, prompt)
}

// RecommendProjects suggests relevant community projects.
func (a *CommunityAgent) RecommendProjects(interests []string, skillLevel string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As a NixOS Community Specialist, recommend projects for:
Interests: %s
Skill Level: %s

%s

Suggest specific projects with rationale for recommendations.`,
		strings.Join(interests, ", "), skillLevel, a.formatCommunityContext())

	ctx := context.Background()
	return a.provider.Query(ctx, prompt)
}

// ExplainCommunityChannels explains different community communication channels.
func (a *CommunityAgent) ExplainCommunityChannels(purpose string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As a NixOS Community Specialist, explain community channels for:
Purpose: %s

%s

Describe the channels, their culture, and when to use each one.`,
		purpose, a.formatCommunityContext())

	ctx := context.Background()
	return a.provider.Query(ctx, prompt)
}

// PlanCommunityInvolvement helps plan community involvement strategy.
func (a *CommunityAgent) PlanCommunityInvolvement(goals []string, timeCommitment string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not available")
	}

	prompt := fmt.Sprintf(`As a NixOS Community Specialist, create involvement plan for:
Goals: %s
Time Commitment: %s

%s

Provide a structured plan with actionable steps.`,
		strings.Join(goals, ", "), timeCommitment, a.formatCommunityContext())

	ctx := context.Background()
	return a.provider.Query(ctx, prompt)
}

// buildCommunityPrompt constructs a community-specific prompt.
func (a *CommunityAgent) buildCommunityPrompt(input string) string {
	var prompt strings.Builder

	// Add role context
	if template, exists := roles.RolePromptTemplate[roles.RoleCommunity]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	// Add community context
	if a.contextData != nil {
		prompt.WriteString("Community Context:\n")
		prompt.WriteString(a.formatCommunityContext())
		prompt.WriteString("\n\n")
	}

	// Add user input
	prompt.WriteString("User Query: ")
	prompt.WriteString(input)

	return prompt.String()
}

// formatCommunityContext formats the community context for inclusion in prompts.
func (a *CommunityAgent) formatCommunityContext() string {
	if a.contextData == nil {
		return "No specific community context provided."
	}

	ctx, ok := a.contextData.(*CommunityContext)
	if !ok {
		return "Invalid community context."
	}

	var context strings.Builder
	context.WriteString("Community Profile:\n")

	if ctx.UserLevel != "" {
		context.WriteString(fmt.Sprintf("- User Level: %s\n", ctx.UserLevel))
	}

	if len(ctx.InterestAreas) > 0 {
		context.WriteString(fmt.Sprintf("- Interest Areas: %s\n", strings.Join(ctx.InterestAreas, ", ")))
	}

	if len(ctx.CurrentProjects) > 0 {
		context.WriteString(fmt.Sprintf("- Current Projects: %s\n", strings.Join(ctx.CurrentProjects, ", ")))
	}

	if len(ctx.CommunityGoals) > 0 {
		context.WriteString(fmt.Sprintf("- Community Goals: %s\n", strings.Join(ctx.CommunityGoals, ", ")))
	}

	if ctx.ExperienceLevel != "" {
		context.WriteString(fmt.Sprintf("- Open Source Experience: %s\n", ctx.ExperienceLevel))
	}

	if len(ctx.PreferredChannels) > 0 {
		context.WriteString(fmt.Sprintf("- Preferred Channels: %s\n", strings.Join(ctx.PreferredChannels, ", ")))
	}

	if ctx.ContributionType != "" {
		context.WriteString(fmt.Sprintf("- Contribution Type: %s\n", ctx.ContributionType))
	}

	return context.String()
}

// enhancePromptWithRole enhances the prompt with role-specific information.
func (a *CommunityAgent) enhancePromptWithRole(input string) string {
	if template, exists := roles.RolePromptTemplate[roles.RoleCommunity]; exists {
		return fmt.Sprintf("%s\n\nUser Request: %s", template, input)
	}
	return input
}

// formatCommunityResponse formats the response for better readability.
func (a *CommunityAgent) formatCommunityResponse(response string) string {
	// Add any community-specific response formatting here
	return response
}
