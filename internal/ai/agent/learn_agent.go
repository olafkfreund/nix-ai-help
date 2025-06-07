package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// LearnAgent handles educational content and learning guidance for NixOS
type LearnAgent struct {
	BaseAgent
}

// LearnContext contains learning-specific context information
type LearnContext struct {
	Topic             string   `json:"topic,omitempty"`
	SkillLevel        string   `json:"skill_level,omitempty"` // beginner, intermediate, advanced
	LearningGoal      string   `json:"learning_goal,omitempty"`
	PreferredStyle    string   `json:"preferred_style,omitempty"` // hands-on, conceptual, reference
	TimeAvailable     string   `json:"time_available,omitempty"`  // quick, thorough, comprehensive
	CurrentKnowledge  []string `json:"current_knowledge,omitempty"`
	LearningPath      []string `json:"learning_path,omitempty"`
	Prerequisites     []string `json:"prerequisites,omitempty"`
	PracticeExercises []string `json:"practice_exercises,omitempty"`
	ResourceLinks     []string `json:"resource_links,omitempty"`
	ExampleCode       string   `json:"example_code,omitempty"`
	CommonMistakes    []string `json:"common_mistakes,omitempty"`
	NextSteps         []string `json:"next_steps,omitempty"`
	RelatedTopics     []string `json:"related_topics,omitempty"`
	ProjectIdeas      []string `json:"project_ideas,omitempty"`
}

// NewLearnAgent creates a new LearnAgent with the Learn role
func NewLearnAgent(provider ai.Provider) *LearnAgent {
	agent := &LearnAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleLearn,
		},
	}
	return agent
}

// Query handles learning and educational questions
func (a *LearnAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt, ok := roles.RolePromptTemplate[a.role]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", a.role)
	}

	// Build context-aware prompt
	fullPrompt := a.buildContextualPrompt(prompt, question)

	return a.provider.Query(ctx, fullPrompt)
}

// GenerateResponse handles learning response generation
func (a *LearnAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add learning-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// buildContextualPrompt constructs a context-aware prompt for learning operations
func (a *LearnAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add learning context if available
	if a.contextData != nil {
		if learnCtx, ok := a.contextData.(*LearnContext); ok {
			contextStr := a.formatLearnContext(learnCtx)
			if contextStr != "" {
				promptParts = append(promptParts, "Learning Context:")
				promptParts = append(promptParts, contextStr)
			}
		}
	}

	// Add user input
	promptParts = append(promptParts, "Learning Request:")
	promptParts = append(promptParts, userInput)

	return strings.Join(promptParts, "\n\n")
}

// formatLearnContext formats LearnContext into a readable string
func (a *LearnAgent) formatLearnContext(ctx *LearnContext) string {
	var parts []string

	if ctx.Topic != "" {
		parts = append(parts, fmt.Sprintf("Topic: %s", ctx.Topic))
	}

	if ctx.SkillLevel != "" {
		parts = append(parts, fmt.Sprintf("Skill Level: %s", ctx.SkillLevel))
	}

	if ctx.LearningGoal != "" {
		parts = append(parts, fmt.Sprintf("Learning Goal: %s", ctx.LearningGoal))
	}

	if ctx.PreferredStyle != "" {
		parts = append(parts, fmt.Sprintf("Learning Style: %s", ctx.PreferredStyle))
	}

	if ctx.TimeAvailable != "" {
		parts = append(parts, fmt.Sprintf("Time Available: %s", ctx.TimeAvailable))
	}

	if len(ctx.CurrentKnowledge) > 0 {
		parts = append(parts, fmt.Sprintf("Current Knowledge: %s", strings.Join(ctx.CurrentKnowledge, ", ")))
	}

	if len(ctx.Prerequisites) > 0 {
		parts = append(parts, fmt.Sprintf("Prerequisites: %s", strings.Join(ctx.Prerequisites, ", ")))
	}

	if len(ctx.LearningPath) > 0 {
		pathStr := strings.Join(ctx.LearningPath, " → ")
		parts = append(parts, fmt.Sprintf("Learning Path: %s", pathStr))
	}

	if ctx.ExampleCode != "" {
		parts = append(parts, fmt.Sprintf("Example Code:\n%s", ctx.ExampleCode))
	}

	if len(ctx.PracticeExercises) > 0 {
		exercisesStr := strings.Join(ctx.PracticeExercises, "\n")
		parts = append(parts, fmt.Sprintf("Practice Exercises:\n%s", exercisesStr))
	}

	if len(ctx.CommonMistakes) > 0 {
		parts = append(parts, fmt.Sprintf("Common Mistakes: %s", strings.Join(ctx.CommonMistakes, ", ")))
	}

	if len(ctx.RelatedTopics) > 0 {
		parts = append(parts, fmt.Sprintf("Related Topics: %s", strings.Join(ctx.RelatedTopics, ", ")))
	}

	if len(ctx.NextSteps) > 0 {
		stepsStr := strings.Join(ctx.NextSteps, "\n")
		parts = append(parts, fmt.Sprintf("Next Steps:\n%s", stepsStr))
	}

	if len(ctx.ProjectIdeas) > 0 {
		projectsStr := strings.Join(ctx.ProjectIdeas, "\n")
		parts = append(parts, fmt.Sprintf("Project Ideas:\n%s", projectsStr))
	}

	if len(ctx.ResourceLinks) > 0 {
		parts = append(parts, fmt.Sprintf("Resources: %s", strings.Join(ctx.ResourceLinks, ", ")))
	}

	return strings.Join(parts, "\n")
}

// SetLearnContext is a convenience method to set LearnContext
func (a *LearnAgent) SetLearnContext(ctx *LearnContext) {
	a.SetContext(ctx)
}
