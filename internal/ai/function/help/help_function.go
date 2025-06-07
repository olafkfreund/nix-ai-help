package help

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
)

// HelpFunction implements AI-powered help and guidance operations for NixOS
type HelpFunction struct {
	*functionbase.BaseFunction
	agent *agent.HelpAgent
}

// HelpResponse represents the structured response from help operations
type HelpResponse struct {
	// Core response fields
	HelpContent   string             `json:"help_content"`
	Documentation []DocumentationRef `json:"documentation,omitempty"`
	Examples      []HelpExample      `json:"examples,omitempty"`
	Guides        []Guide            `json:"guides,omitempty"`

	// Context-specific information
	QuickStart    []string      `json:"quick_start,omitempty"`
	CommonIssues  []CommonIssue `json:"common_issues,omitempty"`
	BestPractices []string      `json:"best_practices,omitempty"`
	RelatedTopics []string      `json:"related_topics,omitempty"`

	// Commands and scripts
	Commands []string `json:"commands,omitempty"`
	Scripts  []Script `json:"scripts,omitempty"`

	// Learning resources
	Tutorials  []Tutorial  `json:"tutorials,omitempty"`
	References []Reference `json:"references,omitempty"`
	Videos     []Video     `json:"videos,omitempty"`

	// Interactive assistance
	NextSteps       []string          `json:"next_steps,omitempty"`
	Prerequisites   []string          `json:"prerequisites,omitempty"`
	Troubleshooting []Troubleshooting `json:"troubleshooting,omitempty"`
}

// DocumentationRef represents a documentation reference
type DocumentationRef struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Section     string `json:"section,omitempty"`
	Description string `json:"description,omitempty"`
	Relevance   string `json:"relevance,omitempty"`
}

// HelpExample represents a practical example
type HelpExample struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code,omitempty"`
	Command     string `json:"command,omitempty"`
	Output      string `json:"output,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

// Guide represents a step-by-step guide
type Guide struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Steps        []string `json:"steps"`
	Difficulty   string   `json:"difficulty,omitempty"`
	TimeEstimate string   `json:"time_estimate,omitempty"`
}

// CommonIssue represents a common problem and its solution
type CommonIssue struct {
	Problem    string   `json:"problem"`
	Symptoms   []string `json:"symptoms,omitempty"`
	Solution   string   `json:"solution"`
	Commands   []string `json:"commands,omitempty"`
	Prevention string   `json:"prevention,omitempty"`
}

// Script represents a helpful script
type Script struct {
	Name         string   `json:"name"`
	Purpose      string   `json:"purpose"`
	Language     string   `json:"language"`
	Content      string   `json:"content"`
	Usage        string   `json:"usage,omitempty"`
	Requirements []string `json:"requirements,omitempty"`
}

// Tutorial represents a learning tutorial
type Tutorial struct {
	Title       string   `json:"title"`
	URL         string   `json:"url,omitempty"`
	Description string   `json:"description"`
	Level       string   `json:"level,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	Topics      []string `json:"topics,omitempty"`
}

// Reference represents a reference resource
type Reference struct {
	Title       string `json:"title"`
	URL         string `json:"url,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description"`
	Relevance   string `json:"relevance,omitempty"`
}

// Video represents a video resource
type Video struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Channel     string `json:"channel,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Description string `json:"description"`
	Level       string `json:"level,omitempty"`
}

// Troubleshooting represents troubleshooting guidance
type Troubleshooting struct {
	Issue      string   `json:"issue"`
	Symptoms   []string `json:"symptoms,omitempty"`
	Diagnosis  []string `json:"diagnosis"`
	Solutions  []string `json:"solutions"`
	Commands   []string `json:"commands,omitempty"`
	Prevention string   `json:"prevention,omitempty"`
}

// NewHelpFunction creates a new help function with agent
func NewHelpFunction(agent *agent.HelpAgent) *HelpFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("query", "Help query or question", true),
		functionbase.StringParam("topic", "Specific topic to get help on", false),
		functionbase.StringParam("level", "Difficulty level: beginner, intermediate, advanced", false),
		functionbase.StringParam("format", "Response format: tutorial, guide, reference, example", false),
		functionbase.BoolParam("interactive", "Enable interactive help", false, false),
		functionbase.BoolParam("show_examples", "Include practical examples", false, true),
		functionbase.BoolParam("detailed", "Provide detailed explanations", false, false),
		functionbase.ArrayParam("related_topics", "Related topics to consider", false),
		functionbase.StringParam("context", "Additional context for the help request", false),
	}

	return &HelpFunction{
		BaseFunction: functionbase.NewBaseFunction("help", "AI-powered help and guidance for NixOS", parameters),
		agent:        agent,
	}
}

// GetHelp provides general help and guidance
func (hf *HelpFunction) GetHelp(ctx context.Context, query string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Provide comprehensive help for this NixOS question: %s

Please provide:
1. Clear, actionable answer
2. Relevant documentation links
3. Practical examples with code
4. Step-by-step guides if applicable
5. Common issues and solutions
6. Best practices
7. Next steps for the user

Format as detailed guidance with examples.`, query)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate help response: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// GetQuickStart provides quick start guidance for a topic
func (hf *HelpFunction) GetQuickStart(ctx context.Context, topic string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Provide a quick start guide for: %s

Include:
1. Prerequisites
2. Essential first steps
3. Basic commands to get started
4. Common gotchas to avoid
5. Where to find more detailed documentation

Keep it concise but complete for beginners.`, topic)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quick start: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// GetTutorials finds relevant tutorials for a topic
func (hf *HelpFunction) GetTutorials(ctx context.Context, topic string, level string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Find and recommend tutorials for: %s (level: %s)

Please suggest:
1. Official NixOS tutorials
2. Community tutorials and blog posts
3. Video tutorials
4. Interactive learning resources
5. Hands-on exercises
6. Practice projects

Organize by difficulty level and include descriptions.`, topic, level)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to find tutorials: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// GetBestPractices provides best practices for a topic
func (hf *HelpFunction) GetBestPractices(ctx context.Context, topic string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Provide best practices for: %s

Cover:
1. Recommended approaches
2. Common patterns to follow
3. Patterns to avoid
4. Performance considerations
5. Security best practices
6. Maintainability guidelines
7. Community conventions

Include concrete examples where helpful.`, topic)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get best practices: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// GetTroubleshooting provides troubleshooting guidance
func (hf *HelpFunction) GetTroubleshooting(ctx context.Context, problem string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Help troubleshoot this NixOS problem: %s

Provide:
1. Common causes of this issue
2. Diagnostic steps to identify the root cause
3. Step-by-step solutions
4. Commands to run for diagnosis and fixes
5. How to prevent this issue in the future
6. When to seek additional help

Be systematic and thorough.`, problem)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate troubleshooting guide: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// GetExamples provides practical examples for a topic
func (hf *HelpFunction) GetExamples(ctx context.Context, topic string, useCase string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Provide practical examples for: %s (use case: %s)

Include:
1. Basic examples for common scenarios
2. Advanced examples for complex use cases
3. Complete configuration examples
4. Command-line examples with expected output
5. Real-world use case implementations
6. Examples with explanations

Make examples copy-pastable and well-commented.`, topic, useCase)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate examples: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// SearchDocumentation searches for relevant documentation
func (hf *HelpFunction) SearchDocumentation(ctx context.Context, query string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Search for documentation related to: %s

Find and suggest:
1. Official NixOS manual sections
2. Nixpkgs manual references
3. Home Manager documentation
4. Wiki articles
5. API documentation
6. Community guides and resources

Provide direct links and brief descriptions of relevance.`, query)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to search documentation: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// parseHelpResponse parses the AI response into structured help response
func (hf *HelpFunction) parseHelpResponse(response string) *HelpResponse {
	helpResponse := &HelpResponse{
		HelpContent: response,
	}

	// Extract documentation URLs
	urlRegex := regexp.MustCompile(`https?://[^\s\)]+`)
	urls := urlRegex.FindAllString(response, -1)
	for _, url := range urls {
		helpResponse.Documentation = append(helpResponse.Documentation, DocumentationRef{
			URL: url,
		})
	}

	// Extract commands (lines starting with $ or containing sudo/nix commands)
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "$") ||
			strings.Contains(line, "nix ") ||
			strings.Contains(line, "nixos-") ||
			strings.Contains(line, "sudo") {
			helpResponse.Commands = append(helpResponse.Commands, line)
		}
	}

	return helpResponse
}

// Helper methods for different types of help requests

// AnswerQuestion provides a direct answer to a specific question
func (hf *HelpFunction) AnswerQuestion(ctx context.Context, question string) (string, error) {
	if hf.agent == nil {
		return "", fmt.Errorf("help agent not available")
	}

	return hf.agent.Query(ctx, question)
}

// ExplainConcept explains a NixOS concept in detail
func (hf *HelpFunction) ExplainConcept(ctx context.Context, concept string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	prompt := fmt.Sprintf(`Explain the NixOS concept: %s

Please provide:
1. Clear definition and explanation
2. Why it's important in NixOS
3. How it relates to other concepts
4. Practical examples of usage
5. Common misconceptions
6. Further reading suggestions

Make it accessible for different skill levels.`, concept)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to explain concept: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}

// CompareOptions compares different approaches or options
func (hf *HelpFunction) CompareOptions(ctx context.Context, options []string, context string) (*HelpResponse, error) {
	if hf.agent == nil {
		return nil, fmt.Errorf("help agent not available")
	}

	optionsList := strings.Join(options, ", ")
	prompt := fmt.Sprintf(`Compare these NixOS options: %s

Context: %s

For each option, provide:
1. Description and use cases
2. Pros and cons
3. Performance implications
4. Complexity considerations
5. Recommendations for when to use each
6. Migration considerations between options

End with a clear recommendation based on the context.`, optionsList, context)

	response, err := hf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to compare options: %w", err)
	}

	return hf.parseHelpResponse(response), nil
}
