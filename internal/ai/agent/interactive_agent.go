package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// InteractiveAgent is specialized for managing interactive troubleshooting sessions.
type InteractiveAgent struct {
	BaseAgent
	sessionHistory []string // Track conversation history
}

// InteractiveContext contains structured information for interactive sessions.
type InteractiveContext struct {
	SessionID       string            // Unique session identifier
	UserLevel       string            // Beginner, Intermediate, Advanced
	CurrentTask     string            // What the user is trying to accomplish
	SystemState     string            // Current system state information
	PrevCommands    []string          // Previously executed commands
	SessionHistory  []string          // Previous interactions in session
	ErrorContext    string            // Any current errors or issues
	StepNumber      int               // Current step in troubleshooting
	Preferences     map[string]string // User preferences for interaction
	Metadata        map[string]string // Additional session metadata
}

// NewInteractiveAgent creates a new InteractiveAgent.
func NewInteractiveAgent(provider ai.Provider) *InteractiveAgent {
	agent := &InteractiveAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleInteractive,
		},
		sessionHistory: make([]string, 0),
	}
	return agent
}

// Query handles interactive session queries with session context.
func (a *InteractiveAgent) Query(ctx context.Context, input string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build enhanced context for the interactive session
	interactiveCtx := a.buildInteractiveContext(input)

	// Build the enhanced prompt
	prompt := a.buildInteractivePrompt(input, interactiveCtx)

	// Query the AI provider
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to query provider: %w", err)
	}

	// Add to session history and enhance response
	a.addToHistory(input, response)
	return a.enhanceResponseWithInteractiveGuidance(response), nil
}

// GenerateResponse generates a response using the provider's GenerateResponse method.
func (a *InteractiveAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Enhance the prompt with role-specific instructions
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", err
	}

	return a.enhanceResponseWithInteractiveGuidance(response), nil
}

// QueryWithContext queries with additional structured context.
func (a *InteractiveAgent) QueryWithContext(ctx context.Context, input string, interactiveCtx *InteractiveContext) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt := a.buildInteractivePrompt(input, interactiveCtx)
	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", err
	}

	a.addToHistory(input, response)
	return a.enhanceResponseWithInteractiveGuidance(response), nil
}

// StartSession initializes a new interactive session.
func (a *InteractiveAgent) StartSession(ctx context.Context, userLevel string) (string, error) {
	sessionCtx := &InteractiveContext{
		SessionID:      generateSessionID(),
		UserLevel:      userLevel,
		CurrentTask:    "starting session",
		SessionHistory: make([]string, 0),
		StepNumber:     1,
		Preferences:    make(map[string]string),
		Metadata:       make(map[string]string),
	}

	welcomeMsg := "Welcome to nixai interactive mode! How can I help you with your NixOS configuration today?"
	return a.QueryWithContext(ctx, welcomeMsg, sessionCtx)
}

// buildInteractiveContext builds comprehensive context for an interactive session.
func (a *InteractiveAgent) buildInteractiveContext(input string) *InteractiveContext {
	interactiveCtx := &InteractiveContext{
		SessionID:      generateSessionID(),
		UserLevel:      a.determineUserLevel(input),
		CurrentTask:    a.extractCurrentTask(input),
		SessionHistory: a.sessionHistory,
		StepNumber:     len(a.sessionHistory) + 1,
		Preferences:    make(map[string]string),
		Metadata:       make(map[string]string),
	}

	// Add context based on input content
	if strings.Contains(strings.ToLower(input), "error") {
		interactiveCtx.ErrorContext = a.extractErrorContext(input)
	}

	return interactiveCtx
}

// buildInteractivePrompt constructs an enhanced prompt for interactive sessions.
func (a *InteractiveAgent) buildInteractivePrompt(input string, interactiveCtx *InteractiveContext) string {
	var prompt strings.Builder

	// Start with role-specific prompt
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("## Interactive Session\n\n")
	prompt.WriteString(fmt.Sprintf("**User Input**: %s\n\n", input))

	if interactiveCtx != nil {
		prompt.WriteString("### Session Context:\n")

		if interactiveCtx.SessionID != "" {
			prompt.WriteString(fmt.Sprintf("- **Session ID**: %s\n", interactiveCtx.SessionID))
		}

		if interactiveCtx.UserLevel != "" {
			prompt.WriteString(fmt.Sprintf("- **User Level**: %s\n", interactiveCtx.UserLevel))
		}

		if interactiveCtx.CurrentTask != "" {
			prompt.WriteString(fmt.Sprintf("- **Current Task**: %s\n", interactiveCtx.CurrentTask))
		}

		if interactiveCtx.StepNumber > 0 {
			prompt.WriteString(fmt.Sprintf("- **Step Number**: %d\n", interactiveCtx.StepNumber))
		}

		if interactiveCtx.ErrorContext != "" {
			prompt.WriteString(fmt.Sprintf("- **Error Context**: %s\n", interactiveCtx.ErrorContext))
		}

		if len(interactiveCtx.SessionHistory) > 0 {
			prompt.WriteString(fmt.Sprintf("- **Previous Interactions**: %d exchanges\n", len(interactiveCtx.SessionHistory)))
			// Include last few interactions for context
			if len(interactiveCtx.SessionHistory) > 0 {
				prompt.WriteString("- **Recent History**:\n")
				start := len(interactiveCtx.SessionHistory) - 3
				if start < 0 {
					start = 0
				}
				for i := start; i < len(interactiveCtx.SessionHistory); i++ {
					prompt.WriteString(fmt.Sprintf("  - %s\n", interactiveCtx.SessionHistory[i]))
				}
			}
		}

		prompt.WriteString("\n")
	}

	prompt.WriteString("### Instructions:\n")
	prompt.WriteString("Provide an interactive, conversational response that:\n")
	prompt.WriteString("1. Acknowledges the user's input and current context\n")
	prompt.WriteString("2. Provides clear, actionable next steps\n")
	prompt.WriteString("3. Explains concepts at the appropriate user level\n")
	prompt.WriteString("4. Offers multiple options when relevant\n")
	prompt.WriteString("5. Maintains conversation flow and session continuity\n")
	prompt.WriteString("6. Suggests follow-up questions or next steps\n\n")

	return prompt.String()
}

// determineUserLevel attempts to determine user expertise level from input.
func (a *InteractiveAgent) determineUserLevel(input string) string {
	lowerInput := strings.ToLower(input)
	
	// Advanced indicators
	if strings.Contains(lowerInput, "derivation") || 
	   strings.Contains(lowerInput, "nix expression") ||
	   strings.Contains(lowerInput, "overlay") ||
	   strings.Contains(lowerInput, "flake.lock") {
		return "Advanced"
	}
	
	// Intermediate indicators
	if strings.Contains(lowerInput, "configuration.nix") ||
	   strings.Contains(lowerInput, "home-manager") ||
	   strings.Contains(lowerInput, "channel") {
		return "Intermediate"
	}
	
	// Default to beginner for safety
	return "Beginner"
}

// extractCurrentTask tries to understand what the user is trying to accomplish.
func (a *InteractiveAgent) extractCurrentTask(input string) string {
	lowerInput := strings.ToLower(input)
	
	if strings.Contains(lowerInput, "install") {
		return "Package Installation"
	} else if strings.Contains(lowerInput, "config") {
		return "Configuration Management"
	} else if strings.Contains(lowerInput, "build") || strings.Contains(lowerInput, "rebuild") {
		return "System Building"
	} else if strings.Contains(lowerInput, "error") || strings.Contains(lowerInput, "fail") {
		return "Troubleshooting"
	} else if strings.Contains(lowerInput, "setup") {
		return "System Setup"
	}
	
	return "General Assistance"
}

// extractErrorContext extracts error information from user input.
func (a *InteractiveAgent) extractErrorContext(input string) string {
	// Look for common error patterns
	if strings.Contains(strings.ToLower(input), "error:") {
		// Try to extract the actual error message
		parts := strings.Split(input, "error:")
		if len(parts) > 1 {
			return "Error: " + strings.TrimSpace(parts[1])
		}
	}
	return "User reported an error condition"
}

// addToHistory adds an interaction to the session history.
func (a *InteractiveAgent) addToHistory(input, response string) {
	a.sessionHistory = append(a.sessionHistory, fmt.Sprintf("User: %s", input))
	a.sessionHistory = append(a.sessionHistory, fmt.Sprintf("Assistant: %s", response))
	
	// Keep history manageable (last 20 interactions)
	if len(a.sessionHistory) > 20 {
		a.sessionHistory = a.sessionHistory[len(a.sessionHistory)-20:]
	}
}

// enhanceResponseWithInteractiveGuidance adds interactive-specific guidance to responses.
func (a *InteractiveAgent) enhanceResponseWithInteractiveGuidance(response string) string {
	guidance := "\n\n💡 **Interactive Tips**:\n" +
		"- Type 'help' for available commands\n" +
		"- Use 'exit' or 'quit' to leave interactive mode\n" +
		"- Ask follow-up questions for clarification\n" +
		"- Request step-by-step guidance for complex tasks"
	
	return response + guidance
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *InteractiveAgent) enhancePromptWithRole(prompt string) string {
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		return fmt.Sprintf("%s\n\n%s", template, prompt)
	}
	return prompt
}

// generateSessionID creates a simple session identifier.
func generateSessionID() string {
	return fmt.Sprintf("interactive-%d", len("session"))
}

// GetSessionHistory returns the current session history.
func (a *InteractiveAgent) GetSessionHistory() []string {
	return a.sessionHistory
}

// ClearSessionHistory clears the session history.
func (a *InteractiveAgent) ClearSessionHistory() {
	a.sessionHistory = make([]string, 0)
}
