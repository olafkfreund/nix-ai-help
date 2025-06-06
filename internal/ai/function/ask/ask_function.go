package ask

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// AskFunction implements AI function calling for direct question handling
type AskFunction struct {
	*functionbase.BaseFunction
	askAgent *agent.AskAgent
	logger   *logger.Logger
}

// AskRequest represents the input parameters for the ask function
type AskRequest struct {
	Question      string   `json:"question"`
	Category      string   `json:"category,omitempty"`
	Context       string   `json:"context,omitempty"`
	Urgency       string   `json:"urgency,omitempty"`
	RelatedTopics []string `json:"related_topics,omitempty"`
}

// AskResponse represents the output of the ask function
type AskResponse struct {
	Answer            string   `json:"answer"`
	RelatedTopics     []string `json:"related_topics,omitempty"`
	SuggestedActions  []string `json:"suggested_actions,omitempty"`
	DocumentationRefs []string `json:"documentation_refs,omitempty"`
	Category          string   `json:"category,omitempty"`
	Confidence        string   `json:"confidence,omitempty"`
}

// NewAskFunction creates a new ask function
func NewAskFunction() *AskFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("question", "The question to ask about NixOS, Nix, or Home Manager", true),
		functionbase.StringParamWithOptions("category", "Category of the question", false,
			[]string{"nixos", "nix", "home-manager", "general", "troubleshooting", "configuration"}, nil, nil),
		functionbase.StringParam("context", "Additional context for the question", false),
		functionbase.StringParamWithOptions("urgency", "Urgency level of the question", false,
			[]string{"low", "normal", "high", "urgent"}, nil, nil),
		{
			Name:        "related_topics",
			Type:        "array",
			Description: "Related topics to consider",
			Required:    false,
		},
	}

	baseFunc := functionbase.NewBaseFunction(
		"ask",
		"Ask questions about NixOS, Nix, Home Manager, or general system configuration",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Ask a general NixOS question",
			Parameters: map[string]interface{}{
				"question": "How do I enable SSH on NixOS?",
				"category": "nixos",
				"urgency":  "normal",
			},
			Expected: "Detailed answer with configuration examples and explanations",
		},
		{
			Description: "Ask about Home Manager configuration",
			Parameters: map[string]interface{}{
				"question":       "How do I configure Git with Home Manager?",
				"category":       "home-manager",
				"context":        "I want to set up my development environment",
				"related_topics": []string{"git", "development", "dotfiles"},
			},
			Expected: "Home Manager configuration examples and best practices",
		},
	}
	baseFunc.SetSchema(schema)

	return &AskFunction{
		BaseFunction: baseFunc,
		askAgent:     agent.NewAskAgent(nil), // Provider will be set when function is executed
		logger:       logger.NewLogger(),
	}
}

// Execute runs the ask function
func (af *AskFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	af.logger.Debug("Starting ask function execution")

	// Parse parameters into structured request
	request, err := af.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate that we have a question
	if request.Question == "" {
		return functionbase.CreateErrorResult(
			fmt.Errorf("question parameter is required"),
			"Missing required parameter",
		), nil
	}

	// Report progress if callback is available
	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    1,
			Total:      4,
			Percentage: 25,
			Message:    "Processing question",
			Stage:      "preparation",
		})
	}

	// Build question context
	questionContext := af.buildQuestionContext(request)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    2,
			Total:      4,
			Percentage: 50,
			Message:    "Querying AI provider",
			Stage:      "processing",
		})
	}

	// Query the ask agent
	answer, err := af.askAgent.Query(ctx, questionContext)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to get answer from AI provider"), nil
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    3,
			Total:      4,
			Percentage: 75,
			Message:    "Processing response",
			Stage:      "formatting",
		})
	}

	// Build the response
	response := af.buildResponse(request, answer)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    4,
			Total:      4,
			Percentage: 100,
			Message:    "Completed successfully",
			Stage:      "complete",
		})
	}

	af.logger.Debug("Ask function execution completed successfully")

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts raw parameters to structured AskRequest
func (af *AskFunction) parseRequest(params map[string]interface{}) (*AskRequest, error) {
	request := &AskRequest{}

	// Extract question (required)
	if question, ok := params["question"].(string); ok {
		request.Question = strings.TrimSpace(question)
	}

	// Extract optional parameters
	if category, ok := params["category"].(string); ok {
		request.Category = strings.TrimSpace(category)
	}

	if context, ok := params["context"].(string); ok {
		request.Context = strings.TrimSpace(context)
	}

	if urgency, ok := params["urgency"].(string); ok {
		request.Urgency = strings.TrimSpace(urgency)
	}

	// Extract related topics array
	if relatedTopics, ok := params["related_topics"].([]interface{}); ok {
		for _, topic := range relatedTopics {
			if topicStr, ok := topic.(string); ok {
				request.RelatedTopics = append(request.RelatedTopics, strings.TrimSpace(topicStr))
			}
		}
	}

	return request, nil
}

// buildQuestionContext creates a formatted context string for the AI
func (af *AskFunction) buildQuestionContext(request *AskRequest) string {
	var contextParts []string

	// Add the main question
	contextParts = append(contextParts, fmt.Sprintf("Question: %s", request.Question))

	// Add category context
	if request.Category != "" {
		contextParts = append(contextParts, fmt.Sprintf("Category: %s", request.Category))
	}

	// Add additional context
	if request.Context != "" {
		contextParts = append(contextParts, fmt.Sprintf("Context: %s", request.Context))
	}

	// Add urgency
	if request.Urgency != "" {
		contextParts = append(contextParts, fmt.Sprintf("Urgency: %s", request.Urgency))
	}

	// Add related topics
	if len(request.RelatedTopics) > 0 {
		contextParts = append(contextParts, fmt.Sprintf("Related topics: %s", strings.Join(request.RelatedTopics, ", ")))
	}

	return strings.Join(contextParts, "\n")
}

// buildResponse creates the structured response
func (af *AskFunction) buildResponse(request *AskRequest, answer string) *AskResponse {
	response := &AskResponse{
		Answer:   answer,
		Category: request.Category,
	}

	// Determine confidence based on question complexity and available information
	response.Confidence = af.determineConfidence(request, answer)

	// Extract related topics from the request and potentially from the answer
	response.RelatedTopics = request.RelatedTopics

	// Add suggested actions based on the question category
	response.SuggestedActions = af.generateSuggestedActions(request)

	// Add documentation references based on the category
	response.DocumentationRefs = af.generateDocumentationRefs(request)

	return response
}

// determineConfidence analyzes the request and response to determine confidence level
func (af *AskFunction) determineConfidence(request *AskRequest, answer string) string {
	lowerQuestion := strings.ToLower(request.Question)

	// High confidence for well-defined questions (highest priority)
	if strings.Contains(lowerQuestion, "how to") ||
		strings.Contains(lowerQuestion, "enable") ||
		strings.Contains(lowerQuestion, "configure") {
		return "high"
	}

	// Medium confidence for recommendation/opinion questions
	if strings.Contains(lowerQuestion, "best") ||
		strings.Contains(lowerQuestion, "recommend") ||
		strings.Contains(lowerQuestion, "should i") {
		return "medium"
	}

	// Very broad or vague questions (check for simple patterns)
	if len(strings.Fields(request.Question)) < 4 ||
		strings.Contains(lowerQuestion, "help") ||
		(strings.Contains(lowerQuestion, "what is") && len(strings.Fields(request.Question)) <= 4) {
		return "low"
	}

	// Default to medium confidence for other questions
	return "medium"
}

// generateSuggestedActions provides actionable next steps based on the question
func (af *AskFunction) generateSuggestedActions(request *AskRequest) []string {
	var actions []string

	switch request.Category {
	case "nixos":
		actions = append(actions, "Check your NixOS configuration.nix")
		actions = append(actions, "Run 'nixos-rebuild switch' to apply changes")
		actions = append(actions, "Review the NixOS manual for detailed documentation")
	case "home-manager":
		actions = append(actions, "Check your Home Manager configuration")
		actions = append(actions, "Run 'home-manager switch' to apply changes")
		actions = append(actions, "Review Home Manager documentation")
	case "troubleshooting":
		actions = append(actions, "Check system logs with 'journalctl'")
		actions = append(actions, "Verify your configuration syntax")
		actions = append(actions, "Try a system rollback if needed")
	default:
		actions = append(actions, "Test the suggested solution in a safe environment")
		actions = append(actions, "Backup your configuration before making changes")
		actions = append(actions, "Consult official documentation for more details")
	}

	return actions
}

// generateDocumentationRefs provides relevant documentation links
func (af *AskFunction) generateDocumentationRefs(request *AskRequest) []string {
	var refs []string

	switch request.Category {
	case "nixos":
		refs = append(refs, "https://nixos.org/manual/nixos/stable/")
		refs = append(refs, "https://wiki.nixos.org/")
	case "nix":
		refs = append(refs, "https://nix.dev/manual/nix")
		refs = append(refs, "https://nixos.org/manual/nixpkgs/stable/")
	case "home-manager":
		refs = append(refs, "https://nix-community.github.io/home-manager/")
		refs = append(refs, "https://wiki.nixos.org/wiki/Home_Manager")
	default:
		refs = append(refs, "https://nixos.org/learn.html")
		refs = append(refs, "https://nix.dev/")
	}

	return refs
}
