package help

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// HelpFunction implements AI-powered help and guidance operations for NixOS
type HelpFunction struct {
	*functionbase.BaseFunction
	agent agent.HelpAgent
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
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"` // manual, wiki, guide, api
}

// HelpExample represents a code or usage example
type HelpExample struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Code        string `json:"code"`
	Language    string `json:"language,omitempty"`
	Output      string `json:"output,omitempty"`
}

// Guide represents a step-by-step guide
type Guide struct {
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	Steps         []string `json:"steps"`
	Difficulty    string   `json:"difficulty,omitempty"` // beginner, intermediate, advanced
	EstimatedTime string   `json:"estimated_time,omitempty"`
}

// CommonIssue represents a common problem and its solution
type CommonIssue struct {
	Problem    string   `json:"problem"`
	Symptoms   []string `json:"symptoms,omitempty"`
	Solution   string   `json:"solution"`
	Commands   []string `json:"commands,omitempty"`
	References []string `json:"references,omitempty"`
}

// Script represents a helpful script or command sequence
type Script struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Usage       string   `json:"usage,omitempty"`
}

// Tutorial represents a learning tutorial
type Tutorial struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Level       string `json:"level,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

// Reference represents a reference document
type Reference struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Type        string `json:"type,omitempty"` // manual, cheatsheet, quickref
}

// Video represents a video resource
type Video struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Duration    string `json:"duration,omitempty"`
	Channel     string `json:"channel,omitempty"`
}

// Troubleshooting represents troubleshooting information
type Troubleshooting struct {
	Issue      string   `json:"issue"`
	Diagnosis  []string `json:"diagnosis"`
	Solutions  []string `json:"solutions"`
	Prevention []string `json:"prevention,omitempty"`
}

// NewHelpFunction creates a new HelpFunction instance
func NewHelpFunction(cfg *config.Config, log *logger.Logger) *HelpFunction {
	return &HelpFunction{
		BaseFunction: &functionbase.BaseFunction{
			Config: cfg,
			Logger: log,
		},
		agent: agent.NewHelpAgent(cfg, log),
	}
}

// GetName returns the function name
func (f *HelpFunction) GetName() string {
	return "help"
}

// GetDescription returns the function description
func (f *HelpFunction) GetDescription() string {
	return "Provides comprehensive help, guidance, documentation, and learning resources for NixOS. " +
		"Offers interactive assistance, troubleshooting support, tutorials, examples, and best practices " +
		"for all aspects of NixOS usage, configuration, and development."
}

// GetParameters returns the function parameters schema
func (f *HelpFunction) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The help operation to perform",
				"enum":        []string{"search", "guide", "examples", "troubleshoot", "quickstart", "reference", "tutorial", "explain", "best-practices"},
			},
			"topic": map[string]interface{}{
				"type":        "string",
				"description": "The topic or subject to get help about",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query or specific question for help",
			},
			"level": map[string]interface{}{
				"type":        "string",
				"description": "Experience level for tailored help",
				"enum":        []string{"beginner", "intermediate", "advanced"},
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Help category to focus on",
				"enum":        []string{"installation", "configuration", "packages", "development", "troubleshooting", "security", "performance", "networking"},
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Preferred help format",
				"enum":        []string{"guide", "examples", "reference", "interactive", "summary"},
			},
			"context": map[string]interface{}{
				"type":        "object",
				"description": "Additional context for help request",
				"properties": map[string]interface{}{
					"system_info": map[string]interface{}{
						"type":        "string",
						"description": "System information (architecture, NixOS version, etc.)",
					},
					"error_message": map[string]interface{}{
						"type":        "string",
						"description": "Error message or issue description",
					},
					"current_config": map[string]interface{}{
						"type":        "string",
						"description": "Current configuration snippet",
					},
					"goal": map[string]interface{}{
						"type":        "string",
						"description": "What the user is trying to achieve",
					},
				},
			},
			"interactive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to provide interactive help with follow-up questions",
				"default":     false,
			},
			"include_examples": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include code examples",
				"default":     true,
			},
			"include_links": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include documentation links",
				"default":     true,
			},
		},
		"required": []string{"operation"},
	}
}

// ValidateParameters validates the provided parameters
func (f *HelpFunction) ValidateParameters(params map[string]interface{}) error {
	// Check required operation parameter
	operation, ok := params["operation"]
	if !ok {
		return fmt.Errorf("operation parameter is required")
	}

	operationStr, ok := operation.(string)
	if !ok {
		return fmt.Errorf("operation must be a string")
	}

	// Validate operation type
	validOperations := []string{"search", "guide", "examples", "troubleshoot", "quickstart", "reference", "tutorial", "explain", "best-practices"}
	if !f.isValidEnum(operationStr, validOperations) {
		return fmt.Errorf("invalid operation: %s. Must be one of: %v", operationStr, validOperations)
	}

	// Validate optional parameters
	if level, ok := params["level"]; ok {
		if levelStr, ok := level.(string); ok {
			validLevels := []string{"beginner", "intermediate", "advanced"}
			if !f.isValidEnum(levelStr, validLevels) {
				return fmt.Errorf("invalid level: %s. Must be one of: %v", levelStr, validLevels)
			}
		} else {
			return fmt.Errorf("level must be a string")
		}
	}

	if category, ok := params["category"]; ok {
		if categoryStr, ok := category.(string); ok {
			validCategories := []string{"installation", "configuration", "packages", "development", "troubleshooting", "security", "performance", "networking"}
			if !f.isValidEnum(categoryStr, validCategories) {
				return fmt.Errorf("invalid category: %s. Must be one of: %v", categoryStr, validCategories)
			}
		} else {
			return fmt.Errorf("category must be a string")
		}
	}

	if format, ok := params["format"]; ok {
		if formatStr, ok := format.(string); ok {
			validFormats := []string{"guide", "examples", "reference", "interactive", "summary"}
			if !f.isValidEnum(formatStr, validFormats) {
				return fmt.Errorf("invalid format: %s. Must be one of: %v", formatStr, validFormats)
			}
		} else {
			return fmt.Errorf("format must be a string")
		}
	}

	// Validate context object if provided
	if context, ok := params["context"]; ok {
		if contextMap, ok := context.(map[string]interface{}); ok {
			for key, value := range contextMap {
				if _, ok := value.(string); !ok {
					return fmt.Errorf("context.%s must be a string", key)
				}
			}
		} else {
			return fmt.Errorf("context must be an object")
		}
	}

	return nil
}

// Execute performs the help operation
func (f *HelpFunction) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Create help context from parameters
	helpCtx := f.createHelpContext(params)

	f.Logger.Info("Executing help operation", "operation", helpCtx.Operation, "topic", helpCtx.Topic)

	// Execute the appropriate help operation
	var response *agent.HelpResponse
	var err error

	switch helpCtx.Operation {
	case "search":
		response, err = f.agent.SearchHelp(ctx, helpCtx)
	case "guide":
		response, err = f.agent.ProvideGuide(ctx, helpCtx)
	case "examples":
		response, err = f.agent.ProvideExamples(ctx, helpCtx)
	case "troubleshoot":
		response, err = f.agent.ProvideTroubleshooting(ctx, helpCtx)
	case "quickstart":
		response, err = f.agent.ProvideQuickStart(ctx, helpCtx)
	case "reference":
		response, err = f.agent.ProvideReference(ctx, helpCtx)
	case "tutorial":
		response, err = f.agent.ProvideTutorials(ctx, helpCtx)
	case "explain":
		response, err = f.agent.ExplainConcept(ctx, helpCtx)
	case "best-practices":
		response, err = f.agent.ProvideBestPractices(ctx, helpCtx)
	default:
		return nil, fmt.Errorf("unsupported help operation: %s", helpCtx.Operation)
	}

	if err != nil {
		f.Logger.Error("Failed to execute help operation", "error", err, "operation", helpCtx.Operation)
		return nil, fmt.Errorf("failed to execute help operation '%s': %w", helpCtx.Operation, err)
	}

	// Convert agent response to function response
	return f.convertAgentResponse(response), nil
}

// createHelpContext creates a help context from parameters
func (f *HelpFunction) createHelpContext(params map[string]interface{}) *agent.HelpContext {
	ctx := &agent.HelpContext{
		Operation:       f.getStringParam(params, "operation", ""),
		Topic:           f.getStringParam(params, "topic", ""),
		Query:           f.getStringParam(params, "query", ""),
		Level:           f.getStringParam(params, "level", "beginner"),
		Category:        f.getStringParam(params, "category", ""),
		Format:          f.getStringParam(params, "format", "guide"),
		Interactive:     f.getBoolParam(params, "interactive", false),
		IncludeExamples: f.getBoolParam(params, "include_examples", true),
		IncludeLinks:    f.getBoolParam(params, "include_links", true),
	}

	// Handle context object
	if contextObj, ok := params["context"].(map[string]interface{}); ok {
		ctx.Context = &agent.HelpContextInfo{
			SystemInfo:    f.getStringParam(contextObj, "system_info", ""),
			ErrorMessage:  f.getStringParam(contextObj, "error_message", ""),
			CurrentConfig: f.getStringParam(contextObj, "current_config", ""),
			Goal:          f.getStringParam(contextObj, "goal", ""),
		}
	}

	return ctx
}

// convertAgentResponse converts agent response to function response
func (f *HelpFunction) convertAgentResponse(agentResp *agent.HelpResponse) *HelpResponse {
	resp := &HelpResponse{
		HelpContent:   agentResp.Content,
		QuickStart:    agentResp.QuickStart,
		BestPractices: agentResp.BestPractices,
		RelatedTopics: agentResp.RelatedTopics,
		Commands:      agentResp.Commands,
		NextSteps:     agentResp.NextSteps,
		Prerequisites: agentResp.Prerequisites,
	}

	// Convert documentation references
	for _, doc := range agentResp.Documentation {
		resp.Documentation = append(resp.Documentation, DocumentationRef{
			Title:       doc.Title,
			URL:         doc.URL,
			Description: doc.Description,
			Type:        doc.Type,
		})
	}

	// Convert examples
	for _, ex := range agentResp.Examples {
		resp.Examples = append(resp.Examples, HelpExample{
			Title:       ex.Title,
			Description: ex.Description,
			Code:        ex.Code,
			Language:    ex.Language,
			Output:      ex.Output,
		})
	}

	// Convert guides
	for _, guide := range agentResp.Guides {
		resp.Guides = append(resp.Guides, Guide{
			Title:         guide.Title,
			Description:   guide.Description,
			Steps:         guide.Steps,
			Difficulty:    guide.Difficulty,
			EstimatedTime: guide.EstimatedTime,
		})
	}

	// Convert common issues
	for _, issue := range agentResp.CommonIssues {
		resp.CommonIssues = append(resp.CommonIssues, CommonIssue{
			Problem:    issue.Problem,
			Symptoms:   issue.Symptoms,
			Solution:   issue.Solution,
			Commands:   issue.Commands,
			References: issue.References,
		})
	}

	// Convert scripts
	for _, script := range agentResp.Scripts {
		resp.Scripts = append(resp.Scripts, Script{
			Name:        script.Name,
			Description: script.Description,
			Commands:    script.Commands,
			Usage:       script.Usage,
		})
	}

	// Convert tutorials
	for _, tutorial := range agentResp.Tutorials {
		resp.Tutorials = append(resp.Tutorials, Tutorial{
			Title:       tutorial.Title,
			Description: tutorial.Description,
			URL:         tutorial.URL,
			Level:       tutorial.Level,
			Duration:    tutorial.Duration,
		})
	}

	// Convert references
	for _, ref := range agentResp.References {
		resp.References = append(resp.References, Reference{
			Title:       ref.Title,
			Description: ref.Description,
			URL:         ref.URL,
			Type:        ref.Type,
		})
	}

	// Convert videos
	for _, video := range agentResp.Videos {
		resp.Videos = append(resp.Videos, Video{
			Title:       video.Title,
			Description: video.Description,
			URL:         video.URL,
			Duration:    video.Duration,
			Channel:     video.Channel,
		})
	}

	// Convert troubleshooting
	for _, ts := range agentResp.Troubleshooting {
		resp.Troubleshooting = append(resp.Troubleshooting, Troubleshooting{
			Issue:      ts.Issue,
			Diagnosis:  ts.Diagnosis,
			Solutions:  ts.Solutions,
			Prevention: ts.Prevention,
		})
	}

	return resp
}

// parseAgentResponse parses the agent's text response to extract structured data
func (f *HelpFunction) parseAgentResponse(response string) *HelpResponse {
	result := &HelpResponse{
		HelpContent: response,
	}

	// Extract documentation references
	result.Documentation = f.extractDocumentation(response)

	// Extract examples
	result.Examples = f.extractExamples(response)

	// Extract guides
	result.Guides = f.extractGuides(response)

	// Extract quick start steps
	result.QuickStart = f.extractQuickStart(response)

	// Extract common issues
	result.CommonIssues = f.extractCommonIssues(response)

	// Extract best practices
	result.BestPractices = f.extractBestPractices(response)

	// Extract related topics
	result.RelatedTopics = f.extractRelatedTopics(response)

	// Extract commands
	result.Commands = f.extractCommands(response)

	// Extract scripts
	result.Scripts = f.extractScripts(response)

	// Extract tutorials
	result.Tutorials = f.extractTutorials(response)

	// Extract references
	result.References = f.extractReferences(response)

	// Extract next steps
	result.NextSteps = f.extractNextSteps(response)

	// Extract prerequisites
	result.Prerequisites = f.extractPrerequisites(response)

	// Extract troubleshooting
	result.Troubleshooting = f.extractTroubleshooting(response)

	return result
}

// extractDocumentation extracts documentation references from text
func (f *HelpFunction) extractDocumentation(text string) []DocumentationRef {
	var docs []DocumentationRef

	// Look for documentation sections
	docPattern := regexp.MustCompile(`(?i)\*\*(?:documentation|docs|references?):\*\*\s*\n((?:[-*]\s*[^\n]+\n?)+)`)
	matches := docPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					line = strings.TrimSpace(line[1:])

					// Try to extract URL and title
					urlPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
					urlMatch := urlPattern.FindStringSubmatch(line)

					if len(urlMatch) == 3 {
						docs = append(docs, DocumentationRef{
							Title: urlMatch[1],
							URL:   urlMatch[2],
							Type:  "reference",
						})
					} else if strings.Contains(line, "http") {
						// Simple URL extraction
						parts := strings.Fields(line)
						for _, part := range parts {
							if strings.HasPrefix(part, "http") {
								docs = append(docs, DocumentationRef{
									Title: strings.TrimSpace(strings.Replace(line, part, "", 1)),
									URL:   part,
									Type:  "reference",
								})
								break
							}
						}
					}
				}
			}
		}
	}

	return docs
}

// extractExamples extracts code examples from text
func (f *HelpFunction) extractExamples(text string) []HelpExample {
	var examples []HelpExample

	// Look for code blocks
	codePattern := regexp.MustCompile("```(\\w+)?\\n([^`]+)```")
	matches := codePattern.FindAllStringSubmatch(text, -1)

	for i, match := range matches {
		if len(match) >= 3 {
			example := HelpExample{
				Title:    fmt.Sprintf("Example %d", i+1),
				Code:     strings.TrimSpace(match[2]),
				Language: match[1],
			}

			// Try to find context around the code block
			codeStart := strings.Index(text, match[0])
			if codeStart > 50 {
				contextStart := codeStart - 50
				context := text[contextStart:codeStart]
				lines := strings.Split(context, "\n")
				if len(lines) > 0 {
					lastLine := strings.TrimSpace(lines[len(lines)-1])
					if lastLine != "" && !strings.HasPrefix(lastLine, "```") {
						example.Description = lastLine
					}
				}
			}

			examples = append(examples, example)
		}
	}

	return examples
}

// extractGuides extracts step-by-step guides from text
func (f *HelpFunction) extractGuides(text string) []Guide {
	var guides []Guide

	// Look for guide sections
	guidePattern := regexp.MustCompile(`(?i)\*\*(?:guide|steps|tutorial):\*\*\s*([^\n]*)\n((?:(?:\d+\.|[-*])\s*[^\n]+\n?)+)`)
	matches := guidePattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			guide := Guide{
				Title:       strings.TrimSpace(match[1]),
				Description: "",
			}

			if guide.Title == "" {
				guide.Title = "Step-by-step Guide"
			}

			// Extract steps
			stepsText := match[2]
			lines := strings.Split(strings.TrimSpace(stepsText), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					// Remove numbering or bullets
					step := regexp.MustCompile(`^(?:\d+\.|\*|-)\s*`).ReplaceAllString(line, "")
					if step != "" {
						guide.Steps = append(guide.Steps, step)
					}
				}
			}

			if len(guide.Steps) > 0 {
				guides = append(guides, guide)
			}
		}
	}

	return guides
}

// extractQuickStart extracts quick start information
func (f *HelpFunction) extractQuickStart(text string) []string {
	return f.extractListItems(text, `(?i)\*\*(?:quick.?start|getting.started):\*\*`)
}

// extractBestPractices extracts best practices
func (f *HelpFunction) extractBestPractices(text string) []string {
	return f.extractListItems(text, `(?i)\*\*(?:best.practices|recommendations):\*\*`)
}

// extractRelatedTopics extracts related topics
func (f *HelpFunction) extractRelatedTopics(text string) []string {
	return f.extractListItems(text, `(?i)\*\*(?:related.topics|see.also):\*\*`)
}

// extractCommands extracts commands
func (f *HelpFunction) extractCommands(text string) []string {
	return f.extractListItems(text, `(?i)\*\*(?:commands?|run):\*\*`)
}

// extractNextSteps extracts next steps
func (f *HelpFunction) extractNextSteps(text string) []string {
	return f.extractListItems(text, `(?i)\*\*(?:next.steps|what.next):\*\*`)
}

// extractPrerequisites extracts prerequisites
func (f *HelpFunction) extractPrerequisites(text string) []string {
	return f.extractListItems(text, `(?i)\*\*(?:prerequisites|requirements):\*\*`)
}

// extractListItems is a helper function to extract list items from sections
func (f *HelpFunction) extractListItems(text, sectionPattern string) []string {
	var items []string

	pattern := regexp.MustCompile(sectionPattern + `\s*\n((?:[-*]\s*[^\n]+\n?)+)`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					item := strings.TrimSpace(line[1:])
					if item != "" {
						items = append(items, item)
					}
				}
			}
		}
	}

	return items
}

// extractCommonIssues extracts common issues and solutions
func (f *HelpFunction) extractCommonIssues(text string) []CommonIssue {
	var issues []CommonIssue

	// Look for common issues sections
	issuePattern := regexp.MustCompile(`(?i)\*\*(?:common.issues|troubleshooting|problems):\*\*\s*\n((?:[-*]\s*[^\n]+(?:\n\s*[^\n*-]+)*\n?)+)`)
	matches := issuePattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			var currentIssue *CommonIssue

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					// New issue
					if currentIssue != nil {
						issues = append(issues, *currentIssue)
					}
					problem := strings.TrimSpace(line[1:])
					currentIssue = &CommonIssue{
						Problem: problem,
					}
				} else if currentIssue != nil && line != "" {
					// Additional solution text
					if currentIssue.Solution == "" {
						currentIssue.Solution = line
					} else {
						currentIssue.Solution += " " + line
					}
				}
			}

			if currentIssue != nil {
				issues = append(issues, *currentIssue)
			}
		}
	}

	return issues
}

// extractScripts extracts script information
func (f *HelpFunction) extractScripts(text string) []Script {
	var scripts []Script

	// Look for script sections
	scriptPattern := regexp.MustCompile(`(?i)\*\*(?:scripts?|automation):\*\*\s*\n((?:[-*]\s*[^\n]+(?:\n\s*[^\n*-]+)*\n?)+)`)
	matches := scriptPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					scriptName := strings.TrimSpace(line[1:])
					scripts = append(scripts, Script{
						Name:        scriptName,
						Description: "Helpful script for NixOS",
						Commands:    []string{scriptName},
					})
				}
			}
		}
	}

	return scripts
}

// extractTutorials extracts tutorial references
func (f *HelpFunction) extractTutorials(text string) []Tutorial {
	var tutorials []Tutorial

	// Look for tutorial sections
	tutorialPattern := regexp.MustCompile(`(?i)\*\*(?:tutorials?):\*\*\s*\n((?:[-*]\s*[^\n]+\n?)+)`)
	matches := tutorialPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					tutorialText := strings.TrimSpace(line[1:])

					// Try to extract URL and title
					urlPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
					urlMatch := urlPattern.FindStringSubmatch(tutorialText)

					if len(urlMatch) == 3 {
						tutorials = append(tutorials, Tutorial{
							Title: urlMatch[1],
							URL:   urlMatch[2],
						})
					} else {
						tutorials = append(tutorials, Tutorial{
							Title: tutorialText,
						})
					}
				}
			}
		}
	}

	return tutorials
}

// extractReferences extracts reference materials
func (f *HelpFunction) extractReferences(text string) []Reference {
	var references []Reference

	// Look for reference sections
	refPattern := regexp.MustCompile(`(?i)\*\*(?:references?|docs|manuals?):\*\*\s*\n((?:[-*]\s*[^\n]+\n?)+)`)
	matches := refPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					refText := strings.TrimSpace(line[1:])

					// Try to extract URL and title
					urlPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
					urlMatch := urlPattern.FindStringSubmatch(refText)

					if len(urlMatch) == 3 {
						references = append(references, Reference{
							Title: urlMatch[1],
							URL:   urlMatch[2],
							Type:  "reference",
						})
					} else if strings.Contains(refText, "http") {
						references = append(references, Reference{
							Title: refText,
							Type:  "reference",
						})
					}
				}
			}
		}
	}

	return references
}

// extractTroubleshooting extracts troubleshooting information
func (f *HelpFunction) extractTroubleshooting(text string) []Troubleshooting {
	var troubleshooting []Troubleshooting

	// Look for troubleshooting sections
	tsPattern := regexp.MustCompile(`(?i)\*\*(?:troubleshooting|debug):\*\*\s*\n((?:[-*]\s*[^\n]+(?:\n\s*[^\n*-]+)*\n?)+)`)
	matches := tsPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			var currentTS *Troubleshooting

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
					// New troubleshooting item
					if currentTS != nil {
						troubleshooting = append(troubleshooting, *currentTS)
					}
					issue := strings.TrimSpace(line[1:])
					currentTS = &Troubleshooting{
						Issue: issue,
					}
				} else if currentTS != nil && line != "" {
					// Additional diagnosis or solution
					currentTS.Diagnosis = append(currentTS.Diagnosis, line)
				}
			}

			if currentTS != nil {
				troubleshooting = append(troubleshooting, *currentTS)
			}
		}
	}

	return troubleshooting
}

// Helper methods

func (f *HelpFunction) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (f *HelpFunction) getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (f *HelpFunction) isValidEnum(value string, validValues []string) bool {
	for _, valid := range validValues {
		if value == valid {
			return true
		}
	}
	return false
}
