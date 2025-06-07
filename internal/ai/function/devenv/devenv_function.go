package devenv

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// DevenvFunction handles development environment management operations
type DevenvFunction struct {
	*functionbase.BaseFunction
	agent  *agent.DevenvAgent
	logger *logger.Logger
}

// DevenvParameters defines the parameters for development environment operations
type DevenvParameters struct {
	Operation    string            `json:"operation" description:"Operation to perform: create, list, suggest, analyze, generate, setup, optimize, validate, templates"`
	ProjectType  string            `json:"project_type,omitempty" description:"Project type: web, cli, library, microservice, desktop, mobile"`
	Language     string            `json:"language,omitempty" description:"Primary programming language: go, rust, python, javascript, typescript, java, etc."`
	Frameworks   []string          `json:"frameworks,omitempty" description:"Frameworks to include: react, nextjs, gin, actix, fastapi, django, etc."`
	Services     []string          `json:"services,omitempty" description:"Services needed: postgres, redis, mysql, mongodb, elasticsearch, etc."`
	Dependencies map[string]string `json:"dependencies,omitempty" description:"Package dependencies with versions"`
	Environment  string            `json:"environment,omitempty" description:"Target environment: development, testing, staging, production"`
	Tools        []string          `json:"tools,omitempty" description:"Development tools: lsp, formatter, linter, debugger, etc."`
	Directory    string            `json:"directory,omitempty" description:"Project directory path"`
	Options      map[string]string `json:"options,omitempty" description:"Additional options and configuration"`
	Template     string            `json:"template,omitempty" description:"Template name for creation operations"`
	Query        string            `json:"query,omitempty" description:"Search query or description for suggestions"`
	NixShell     bool              `json:"nix_shell,omitempty" description:"Use nix-shell instead of devenv"`
	Flakes       bool              `json:"flakes,omitempty" description:"Use Nix flakes"`
	Direnv       bool              `json:"direnv,omitempty" description:"Enable direnv integration"`
}

// DevenvResponse represents the response from development environment operations
type DevenvResponse struct {
	Operation        string               `json:"operation"`
	Status           string               `json:"status"`
	Templates        []TemplateInfo       `json:"templates,omitempty"`
	Configuration    *DevenvConfig        `json:"configuration,omitempty"`
	SetupSteps       []string             `json:"setup_steps,omitempty"`
	GeneratedFiles   map[string]string    `json:"generated_files,omitempty"`
	Environment      *EnvironmentInfo     `json:"environment,omitempty"`
	Suggestions      []string             `json:"suggestions,omitempty"`
	ValidationIssues []ValidationIssue    `json:"validation_issues,omitempty"`
	OptimizationTips []string             `json:"optimization_tips,omitempty"`
	Documentation    []DocumentationLink  `json:"documentation,omitempty"`
	Commands         []CommandInstruction `json:"commands,omitempty"`
	Message          string               `json:"message"`
}

// TemplateInfo represents information about a devenv template
type TemplateInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Frameworks  []string `json:"frameworks"`
	Services    []string `json:"services"`
	Complexity  string   `json:"complexity"`
	Tags        []string `json:"tags"`
	Example     string   `json:"example"`
}

// DevenvConfig represents a devenv.nix configuration
type DevenvConfig struct {
	Languages   map[string]interface{} `json:"languages"`
	Packages    []string               `json:"packages"`
	Services    map[string]interface{} `json:"services"`
	Environment map[string]string      `json:"environment"`
	Scripts     map[string]interface{} `json:"scripts"`
	PreCommit   map[string]interface{} `json:"pre_commit,omitempty"`
	EnterShell  string                 `json:"enter_shell,omitempty"`
	ExitShell   string                 `json:"exit_shell,omitempty"`
	Dotenv      map[string]interface{} `json:"dotenv,omitempty"`
}

// EnvironmentInfo represents development environment information
type EnvironmentInfo struct {
	ProjectPath   string            `json:"project_path"`
	Language      string            `json:"language"`
	BuildSystem   string            `json:"build_system"`
	Frameworks    []string          `json:"frameworks"`
	Dependencies  map[string]string `json:"dependencies"`
	Services      []string          `json:"services"`
	DevTools      []string          `json:"dev_tools"`
	Configuration string            `json:"configuration"`
	Requirements  []string          `json:"requirements"`
}

// ValidationIssue represents a validation issue
type ValidationIssue struct {
	Type       string `json:"type"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Location   string `json:"location,omitempty"`
}

// DocumentationLink represents a documentation link
type DocumentationLink struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// CommandInstruction represents a command instruction
type CommandInstruction struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Required    bool   `json:"required"`
	Order       int    `json:"order"`
}

// NewDevenvFunction creates a new DevenvFunction instance
func NewDevenvFunction() *DevenvFunction {
	parameters := []functionbase.FunctionParameter{
		{
			Name:        "context",
			Type:        "string",
			Description: "The context or reason for the devenv operation",
			Required:    true,
		},
		{
			Name:        "operation",
			Type:        "string",
			Description: "The devenv operation to perform",
			Required:    true,
		},
	}

	return &DevenvFunction{
		BaseFunction: functionbase.NewBaseFunction("devenv", "Manage development environments using devenv.sh, nix-shell, and flakes", parameters),
		agent:        agent.NewDevenvAgent(nil),
		logger:       logger.NewLogger(),
	}
}

// GetSchema returns the function schema for DevenvFunction
func (f *DevenvFunction) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        f.Name,
		"description": f.Description,
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"create", "list", "suggest", "analyze", "generate", "setup", "optimize", "validate", "templates"},
					"description": "Operation to perform",
				},
				"project_type": map[string]interface{}{
					"type":        "string",
					"description": "Project type: web, cli, library, microservice, desktop, mobile",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Primary programming language",
				},
				"frameworks": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Frameworks to include",
				},
				"services": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Services needed",
				},
				"dependencies": map[string]interface{}{
					"type":        "object",
					"description": "Package dependencies with versions",
				},
				"environment": map[string]interface{}{
					"type":        "string",
					"description": "Target environment",
				},
				"tools": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Development tools",
				},
				"directory": map[string]interface{}{
					"type":        "string",
					"description": "Project directory path",
				},
				"options": map[string]interface{}{
					"type":        "object",
					"description": "Additional options and configuration",
				},
				"template": map[string]interface{}{
					"type":        "string",
					"description": "Template name for creation operations",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query or description for suggestions",
				},
				"nix_shell": map[string]interface{}{
					"type":        "boolean",
					"description": "Use nix-shell instead of devenv",
				},
				"flakes": map[string]interface{}{
					"type":        "boolean",
					"description": "Use Nix flakes",
				},
				"direnv": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable direnv integration",
				},
			},
			"required": []string{"operation"},
		},
	}
}

// Execute performs the development environment operation
func (f *DevenvFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	start := time.Now()
	f.logger.Debug(fmt.Sprintf("Executing devenv function with params: %+v", params))

	// Parse parameters
	devenvParams, err := f.parseParameters(params)
	if err != nil {
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to parse parameters: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, err
	}

	// Validate parameters
	if err := f.validateParameters(devenvParams); err != nil {
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("parameter validation failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, err
	}

	// Create devenv context
	devenvContext := f.createDevenvContext(devenvParams)
	f.agent.SetDevenvContext(devenvContext)

	// Execute operation
	var result interface{}
	switch devenvParams.Operation {
	case "create":
		result, err = f.executeCreate(ctx, devenvParams, devenvContext)
	case "list":
		result, err = f.executeList(ctx, devenvParams, devenvContext)
	case "suggest":
		result, err = f.executeSuggest(ctx, devenvParams, devenvContext)
	case "analyze":
		result, err = f.executeAnalyze(ctx, devenvParams, devenvContext)
	case "generate":
		result, err = f.executeGenerate(ctx, devenvParams, devenvContext)
	case "setup":
		result, err = f.executeSetup(ctx, devenvParams, devenvContext)
	case "optimize":
		result, err = f.executeOptimize(ctx, devenvParams, devenvContext)
	case "validate":
		result, err = f.executeValidate(ctx, devenvParams, devenvContext)
	case "templates":
		result, err = f.executeTemplates(ctx, devenvParams, devenvContext)
	default:
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("unsupported operation: %s", devenvParams.Operation),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, fmt.Errorf("unsupported operation: %s", devenvParams.Operation)
	}

	if err != nil {
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("operation failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, err
	}

	return &functionbase.FunctionResult{
		Success:   true,
		Data:      result,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// parseParameters parses the raw parameters into DevenvParameters
func (f *DevenvFunction) parseParameters(params map[string]interface{}) (*DevenvParameters, error) {
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var devenvParams DevenvParameters
	if err := json.Unmarshal(jsonBytes, &devenvParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	return &devenvParams, nil
}

// validateParameters validates the DevenvParameters
func (f *DevenvFunction) validateParameters(params *DevenvParameters) error {
	if params.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := []string{"create", "list", "suggest", "analyze", "generate", "setup", "optimize", "validate", "templates"}
	found := false
	for _, op := range validOps {
		if params.Operation == op {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid operation: %s", params.Operation)
	}

	// Operation-specific validation
	switch params.Operation {
	case "create":
		if params.Language == "" && params.Template == "" {
			return fmt.Errorf("language or template is required for create operation")
		}
	case "suggest":
		if params.Query == "" {
			return fmt.Errorf("query is required for suggest operation")
		}
	case "analyze":
		if params.Directory == "" {
			return fmt.Errorf("directory is required for analyze operation")
		}
	}

	return nil
}

// createDevenvContext creates a DevenvContext from parameters
func (f *DevenvFunction) createDevenvContext(params *DevenvParameters) *agent.DevenvContext {
	return &agent.DevenvContext{
		ProjectType:  params.ProjectType,
		Languages:    []string{params.Language},
		Tools:        params.Tools,
		Services:     params.Services,
		Frameworks:   params.Frameworks,
		Dependencies: params.Dependencies,
		Environment:  params.Environment,
		NixShell:     params.NixShell,
		Flakes:       params.Flakes,
		Direnv:       params.Direnv,
	}
}

// executeCreate handles devenv creation operations
func (f *DevenvFunction) executeCreate(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info(fmt.Sprintf("Creating development environment for %s", params.Language))

	response, err := f.agent.GenerateResponse(ctx, f.buildCreatePrompt(params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate create response: %w", err)
	}

	// Parse the AI response to extract structured information
	devenvResponse := &DevenvResponse{
		Operation: "create",
		Status:    "success",
		Message:   response,
	}

	// Parse setup steps
	devenvResponse.SetupSteps = f.parseSetupSteps(response)

	// Parse commands
	devenvResponse.Commands = f.parseCommands(response)

	// Parse generated files
	devenvResponse.GeneratedFiles = f.parseGeneratedFiles(response)

	// Parse configuration if present
	if config := f.parseDevenvConfig(response); config != nil {
		devenvResponse.Configuration = config
	}

	// Parse documentation links
	devenvResponse.Documentation = f.parseDocumentationLinks(response)

	return devenvResponse, nil
}

// executeList handles template listing operations
func (f *DevenvFunction) executeList(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info("Listing available development environment templates")

	response, err := f.agent.GenerateResponse(ctx, f.buildListPrompt(params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate list response: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation: "list",
		Status:    "success",
		Message:   response,
		Templates: f.parseTemplates(response),
	}

	return devenvResponse, nil
}

// executeSuggest handles template suggestion operations
func (f *DevenvFunction) executeSuggest(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info(fmt.Sprintf("Suggesting development environment for query: %s", params.Query))

	response, err := f.agent.GenerateResponse(ctx, f.buildSuggestPrompt(params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate suggest response: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation:   "suggest",
		Status:      "success",
		Message:     response,
		Suggestions: f.parseSuggestions(response),
		Templates:   f.parseTemplates(response),
	}

	return devenvResponse, nil
}

// executeAnalyze handles project analysis operations
func (f *DevenvFunction) executeAnalyze(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info(fmt.Sprintf("Analyzing project at %s", params.Directory))

	response, err := f.agent.AnalyzeProject(ctx, params.Directory, params.ProjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation:   "analyze",
		Status:      "success",
		Message:     response,
		Environment: f.parseEnvironmentInfo(response),
		Suggestions: f.parseSuggestions(response),
	}

	return devenvResponse, nil
}

// executeGenerate handles configuration generation operations
func (f *DevenvFunction) executeGenerate(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info("Generating development environment configuration")

	var response string
	var err error

	if params.NixShell {
		response, err = f.agent.GenerateShellNix(ctx, devenvCtx)
	} else if params.Flakes {
		response, err = f.agent.GenerateFlakeNix(ctx, devenvCtx)
	} else {
		response, err = f.agent.GenerateResponse(ctx, f.buildGeneratePrompt(params))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate configuration: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation:      "generate",
		Status:         "success",
		Message:        response,
		Configuration:  f.parseDevenvConfig(response),
		GeneratedFiles: f.parseGeneratedFiles(response),
	}

	return devenvResponse, nil
}

// executeSetup handles environment setup operations
func (f *DevenvFunction) executeSetup(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info("Setting up development environment")

	var response string
	var err error

	if params.Direnv {
		response, err = f.agent.SetupDirenv(ctx, devenvCtx)
	} else {
		response, err = f.agent.GenerateResponse(ctx, f.buildSetupPrompt(params))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to setup environment: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation:  "setup",
		Status:     "success",
		Message:    response,
		SetupSteps: f.parseSetupSteps(response),
		Commands:   f.parseCommands(response),
	}

	return devenvResponse, nil
}

// executeOptimize handles environment optimization operations
func (f *DevenvFunction) executeOptimize(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info("Optimizing development environment")

	response, err := f.agent.GenerateResponse(ctx, f.buildOptimizePrompt(params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimization response: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation:        "optimize",
		Status:           "success",
		Message:          response,
		OptimizationTips: f.parseOptimizationTips(response),
		Suggestions:      f.parseSuggestions(response),
	}

	return devenvResponse, nil
}

// executeValidate handles environment validation operations
func (f *DevenvFunction) executeValidate(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info("Validating development environment")

	response, err := f.agent.GenerateResponse(ctx, f.buildValidatePrompt(params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate validation response: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation:        "validate",
		Status:           "success",
		Message:          response,
		ValidationIssues: f.parseValidationIssues(response),
		Suggestions:      f.parseSuggestions(response),
	}

	return devenvResponse, nil
}

// executeTemplates handles template information operations
func (f *DevenvFunction) executeTemplates(ctx context.Context, params *DevenvParameters, devenvCtx *agent.DevenvContext) (*DevenvResponse, error) {
	f.logger.Info("Retrieving template information")

	response, err := f.agent.GenerateResponse(ctx, f.buildTemplatesPrompt(params))
	if err != nil {
		return nil, fmt.Errorf("failed to generate templates response: %w", err)
	}

	devenvResponse := &DevenvResponse{
		Operation: "templates",
		Status:    "success",
		Message:   response,
		Templates: f.parseTemplates(response),
	}

	return devenvResponse, nil
}

// Prompt builders

func (f *DevenvFunction) buildCreatePrompt(params *DevenvParameters) string {
	var prompt strings.Builder
	prompt.WriteString("Create a development environment configuration with the following requirements:\n\n")

	if params.Language != "" {
		prompt.WriteString(fmt.Sprintf("Language: %s\n", params.Language))
	}
	if params.ProjectType != "" {
		prompt.WriteString(fmt.Sprintf("Project Type: %s\n", params.ProjectType))
	}
	if len(params.Frameworks) > 0 {
		prompt.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(params.Frameworks, ", ")))
	}
	if len(params.Services) > 0 {
		prompt.WriteString(fmt.Sprintf("Services: %s\n", strings.Join(params.Services, ", ")))
	}
	if params.Environment != "" {
		prompt.WriteString(fmt.Sprintf("Environment: %s\n", params.Environment))
	}

	prompt.WriteString("\nProvide:\n")
	prompt.WriteString("1. Complete devenv.nix configuration\n")
	prompt.WriteString("2. Step-by-step setup instructions\n")
	prompt.WriteString("3. Required commands to run\n")
	prompt.WriteString("4. Development workflow recommendations\n")
	prompt.WriteString("5. Documentation links\n")

	return prompt.String()
}

func (f *DevenvFunction) buildListPrompt(params *DevenvParameters) string {
	return "List all available development environment templates, including:\n" +
		"1. Template name and description\n" +
		"2. Supported languages and frameworks\n" +
		"3. Included services and tools\n" +
		"4. Complexity level\n" +
		"5. Usage examples\n" +
		"Format as a structured list with clear categories."
}

func (f *DevenvFunction) buildSuggestPrompt(params *DevenvParameters) string {
	return fmt.Sprintf("Based on the description '%s', suggest the most appropriate development environment template and configuration.\n\n"+
		"Include:\n"+
		"1. Recommended template(s)\n"+
		"2. Reasoning for the suggestions\n"+
		"3. Alternative options\n"+
		"4. Configuration considerations\n"+
		"5. Next steps", params.Query)
}

func (f *DevenvFunction) buildGeneratePrompt(params *DevenvParameters) string {
	var prompt strings.Builder
	prompt.WriteString("Generate a complete development environment configuration")

	if params.NixShell {
		prompt.WriteString(" using nix-shell (shell.nix)")
	} else if params.Flakes {
		prompt.WriteString(" using Nix flakes (flake.nix)")
	} else {
		prompt.WriteString(" using devenv.sh (devenv.nix)")
	}

	prompt.WriteString(".\n\nInclude all necessary packages, tools, and environment setup.")
	return prompt.String()
}

func (f *DevenvFunction) buildSetupPrompt(params *DevenvParameters) string {
	return "Provide detailed setup instructions for the development environment, including:\n" +
		"1. Prerequisites and dependencies\n" +
		"2. Installation steps\n" +
		"3. Configuration commands\n" +
		"4. Verification procedures\n" +
		"5. Troubleshooting tips"
}

func (f *DevenvFunction) buildOptimizePrompt(params *DevenvParameters) string {
	return "Analyze and provide optimization recommendations for the development environment:\n" +
		"1. Performance improvements\n" +
		"2. Resource usage optimization\n" +
		"3. Workflow enhancements\n" +
		"4. Tool integration suggestions\n" +
		"5. Best practices"
}

func (f *DevenvFunction) buildValidatePrompt(params *DevenvParameters) string {
	return "Validate the development environment configuration and identify:\n" +
		"1. Configuration errors or warnings\n" +
		"2. Missing dependencies\n" +
		"3. Compatibility issues\n" +
		"4. Security concerns\n" +
		"5. Improvement suggestions"
}

func (f *DevenvFunction) buildTemplatesPrompt(params *DevenvParameters) string {
	return "Provide comprehensive information about available development environment templates:\n" +
		"1. Template categories and types\n" +
		"2. Language and framework support\n" +
		"3. Use cases and examples\n" +
		"4. Customization options\n" +
		"5. Getting started guides"
}

// Response parsers

func (f *DevenvFunction) parseSetupSteps(response string) []string {
	var steps []string

	// Look for numbered lists or bullet points
	stepRegex := regexp.MustCompile(`(?m)^\s*(?:\d+\.|[-*])\s*(.+)$`)
	matches := stepRegex.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) > 1 && len(strings.TrimSpace(match[1])) > 0 {
			steps = append(steps, strings.TrimSpace(match[1]))
		}
	}

	return steps
}

func (f *DevenvFunction) parseCommands(response string) []CommandInstruction {
	var commands []CommandInstruction

	// Look for command patterns
	cmdRegex := regexp.MustCompile(`(?m)^\s*(?:\$|>|#)?\s*([a-zA-Z0-9][^\n]+)$`)
	matches := cmdRegex.FindAllStringSubmatch(response, -1)

	order := 1
	for _, match := range matches {
		if len(match) > 1 {
			cmd := strings.TrimSpace(match[1])
			if len(cmd) > 0 && !strings.Contains(cmd, " ") || strings.Contains(cmd, "devenv") || strings.Contains(cmd, "nix") {
				commands = append(commands, CommandInstruction{
					Command:     cmd,
					Description: "Development environment command",
					Category:    "setup",
					Required:    true,
					Order:       order,
				})
				order++
			}
		}
	}

	return commands
}

func (f *DevenvFunction) parseGeneratedFiles(response string) map[string]string {
	files := make(map[string]string)

	// Look for file content sections
	fileRegex := regexp.MustCompile("(?s)(?:```|~~~)(?:nix|yaml|json|toml)?\\s*\\n(.*?)\\n(?:```|~~~)")
	matches := fileRegex.FindAllStringSubmatch(response, -1)

	for i, match := range matches {
		if len(match) > 1 {
			filename := fmt.Sprintf("file_%d", i+1)
			if strings.Contains(match[0], "devenv.nix") || strings.Contains(response, "devenv.nix") {
				filename = "devenv.nix"
			} else if strings.Contains(match[0], "shell.nix") {
				filename = "shell.nix"
			} else if strings.Contains(match[0], "flake.nix") {
				filename = "flake.nix"
			}
			files[filename] = strings.TrimSpace(match[1])
		}
	}

	return files
}

func (f *DevenvFunction) parseDevenvConfig(response string) *DevenvConfig {
	// Try to extract devenv configuration from the response
	configRegex := regexp.MustCompile(`(?s)devenv\.nix.*?{(.*?)}`)
	matches := configRegex.FindStringSubmatch(response)

	if len(matches) > 1 {
		// This is a simplified parser - in a real implementation,
		// you'd want more sophisticated Nix parsing
		return &DevenvConfig{
			Languages:   make(map[string]interface{}),
			Packages:    []string{},
			Services:    make(map[string]interface{}),
			Environment: make(map[string]string),
			Scripts:     make(map[string]interface{}),
		}
	}

	return nil
}

func (f *DevenvFunction) parseTemplates(response string) []TemplateInfo {
	var templates []TemplateInfo

	// This is a simplified parser - you'd want more sophisticated parsing
	// for a production implementation
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		if strings.Contains(line, "template") || strings.Contains(line, "Template") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				desc := strings.TrimSpace(parts[1])

				templates = append(templates, TemplateInfo{
					Name:        name,
					Description: desc,
					Language:    "auto-detected",
					Complexity:  "medium",
					Tags:        []string{"devenv"},
				})
			}
		}
	}

	return templates
}

func (f *DevenvFunction) parseSuggestions(response string) []string {
	var suggestions []string

	// Look for suggestion patterns
	suggRegex := regexp.MustCompile(`(?i)(?:suggest|recommend|consider).*?:?\s*(.+)`)
	matches := suggRegex.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) > 1 {
			suggestion := strings.TrimSpace(match[1])
			if len(suggestion) > 10 && len(suggestion) < 200 {
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions
}

func (f *DevenvFunction) parseEnvironmentInfo(response string) *EnvironmentInfo {
	return &EnvironmentInfo{
		ProjectPath:   ".",
		Language:      "auto-detected",
		BuildSystem:   "auto-detected",
		Frameworks:    []string{},
		Dependencies:  make(map[string]string),
		Services:      []string{},
		DevTools:      []string{},
		Configuration: "devenv.nix",
		Requirements:  []string{},
	}
}

func (f *DevenvFunction) parseOptimizationTips(response string) []string {
	var tips []string

	// Look for optimization-related content
	tipRegex := regexp.MustCompile(`(?i)(?:optimize|improve|enhance|tip).*?:?\s*(.+)`)
	matches := tipRegex.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tip := strings.TrimSpace(match[1])
			if len(tip) > 10 && len(tip) < 200 {
				tips = append(tips, tip)
			}
		}
	}

	return tips
}

func (f *DevenvFunction) parseValidationIssues(response string) []ValidationIssue {
	var issues []ValidationIssue

	// Look for warning/error patterns
	if strings.Contains(strings.ToLower(response), "error") {
		issues = append(issues, ValidationIssue{
			Type:       "error",
			Level:      "high",
			Message:    "Configuration error detected",
			Suggestion: "Review the configuration for syntax errors",
		})
	}

	if strings.Contains(strings.ToLower(response), "warning") {
		issues = append(issues, ValidationIssue{
			Type:       "warning",
			Level:      "medium",
			Message:    "Configuration warning detected",
			Suggestion: "Consider addressing the warning for better stability",
		})
	}

	return issues
}

func (f *DevenvFunction) parseDocumentationLinks(response string) []DocumentationLink {
	var links []DocumentationLink

	// Look for URL patterns
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlRegex.FindAllString(response, -1)

	for _, url := range urls {
		links = append(links, DocumentationLink{
			Title:       "Documentation",
			URL:         url,
			Category:    "reference",
			Description: "Related documentation",
		})
	}

	return links
}
