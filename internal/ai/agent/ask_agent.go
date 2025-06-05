package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// AskAgent is specialized for handling direct question-answer interactions.
type AskAgent struct {
	BaseAgent
}

// AskContext contains structured information for question handling.
type AskContext struct {
	Question      string            // The user's question
	Category      string            // Question category (NixOS, Home Manager, etc.)
	Urgency       string            // Question urgency level
	Context       string            // Additional context provided
	RelatedTopics []string          // Related topics to consider
	Metadata      map[string]string // Additional question metadata
}

// NewAskAgent creates a new AskAgent.
func NewAskAgent(provider ai.Provider) *AskAgent {
	agent := &AskAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleAsk,
		},
	}
	return agent
}

// Query handles direct questions with enhanced context.
func (a *AskAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build enhanced context for the question
	askCtx := a.buildAskContext(question)

	// Build the enhanced prompt
	prompt := a.buildAskPrompt(question, askCtx)

	// Query the AI provider
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to query provider: %w", err)
	}

	// Enhance response with helpful guidance
	return a.enhanceResponseWithAskGuidance(response), nil
}

// GenerateResponse generates a response using the provider's GenerateResponse method.
func (a *AskAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Enhance the prompt with role-specific instructions
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", err
	}

	return a.enhanceResponseWithAskGuidance(response), nil
}

// QueryWithContext queries with additional structured context.
func (a *AskAgent) QueryWithContext(ctx context.Context, question string, askCtx *AskContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildAskPrompt(question, askCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", err
	}

	return a.enhanceResponseWithAskGuidance(response), nil
}

// buildAskContext builds context information for the question.
func (a *AskAgent) buildAskContext(question string) *AskContext {
	askCtx := &AskContext{
		Question:      question,
		Category:      a.categorizeQuestion(question),
		Urgency:       a.determineUrgency(question),
		RelatedTopics: a.findRelatedTopics(question),
		Metadata:      make(map[string]string),
	}

	return askCtx
}

// buildAskPrompt constructs an enhanced prompt for direct questions.
func (a *AskAgent) buildAskPrompt(question string, askCtx *AskContext) string {
	var prompt strings.Builder

	// Start with role-specific prompt
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("## Direct Question Request\n\n")
	prompt.WriteString(fmt.Sprintf("**User Question**: %s\n\n", question))

	if askCtx != nil {
		prompt.WriteString("### Context Information:\n")

		if askCtx.Category != "" {
			prompt.WriteString(fmt.Sprintf("- **Question Category**: %s\n", askCtx.Category))
		}

		if askCtx.Urgency != "" {
			prompt.WriteString(fmt.Sprintf("- **Urgency Level**: %s\n", askCtx.Urgency))
		}

		if len(askCtx.RelatedTopics) > 0 {
			prompt.WriteString(fmt.Sprintf("- **Related Topics**: %s\n", strings.Join(askCtx.RelatedTopics, ", ")))
		}

		if askCtx.Context != "" {
			prompt.WriteString(fmt.Sprintf("- **Additional Context**: %s\n", askCtx.Context))
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("### Instructions:\n")
	prompt.WriteString("Please provide a comprehensive answer focusing on:\n")
	prompt.WriteString("1. Direct, actionable solution to the question\n")
	prompt.WriteString("2. Step-by-step instructions where applicable\n") 
	prompt.WriteString("3. Code examples and configuration snippets\n")
	prompt.WriteString("4. Common pitfalls and troubleshooting tips\n")
	prompt.WriteString("5. Related topics and further reading\n\n")

	return prompt.String()
}

// categorizeQuestion determines the category of the question.
func (a *AskAgent) categorizeQuestion(question string) string {
	question = strings.ToLower(question)

	if strings.Contains(question, "nixos") || strings.Contains(question, "system") {
		return "NixOS System Configuration"
	} else if strings.Contains(question, "home-manager") || strings.Contains(question, "home manager") {
		return "Home Manager Configuration"
	} else if strings.Contains(question, "flake") || strings.Contains(question, "flakes") {
		return "Nix Flakes"
	} else if strings.Contains(question, "package") || strings.Contains(question, "derivation") {
		return "Package Management"
	} else if strings.Contains(question, "build") || strings.Contains(question, "compile") {
		return "Build Configuration"
	} else if strings.Contains(question, "service") || strings.Contains(question, "systemd") {
		return "Service Management"
	}

	return "General Nix/NixOS"
}

// determineUrgency estimates the urgency level of the question.
func (a *AskAgent) determineUrgency(question string) string {
	question = strings.ToLower(question)

	if strings.Contains(question, "broken") || strings.Contains(question, "error") || 
	   strings.Contains(question, "fail") || strings.Contains(question, "urgent") {
		return "High - System Issue"
	} else if strings.Contains(question, "how") || strings.Contains(question, "setup") ||
	          strings.Contains(question, "configure") {
		return "Medium - Configuration"
	}

	return "Low - General Question"
}

// findRelatedTopics suggests related topics based on the question.
func (a *AskAgent) findRelatedTopics(question string) []string {
	var topics []string
	question = strings.ToLower(question)

	if strings.Contains(question, "flake") {
		topics = append(topics, "flake.nix", "flake inputs", "flake outputs")
	}
	if strings.Contains(question, "package") {
		topics = append(topics, "nixpkgs", "overlays", "package derivations")
	}
	if strings.Contains(question, "service") {
		topics = append(topics, "systemd", "configuration.nix", "service options")
	}
	if strings.Contains(question, "home") {
		topics = append(topics, "home.nix", "user programs", "dotfiles")
	}

	return topics
}

// enhanceResponseWithAskGuidance adds helpful tips to the response.
func (a *AskAgent) enhanceResponseWithAskGuidance(response string) string {
	guidance := "\n\n💡 **Additional Tips:**\n"
	guidance += "- Use `nixai doctor` to diagnose system issues\n"
	guidance += "- Use `nixai search <term>` to find packages or options\n"
	guidance += "- Use `nixai explain-option <option>` for detailed option information\n"
	guidance += "- Check the NixOS manual: https://nixos.org/manual/\n"

	return response + guidance
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *AskAgent) enhancePromptWithRole(prompt string) string {
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		return fmt.Sprintf("%s\n\n%s", template, prompt)
	}
	return prompt
}
