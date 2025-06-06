package flakes

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// FlakesFunction implements AI function calling for Nix flakes management
type FlakesFunction struct {
	*functionbase.BaseFunction
	flakeAgent *agent.FlakeAgent
	logger     *logger.Logger
}

// FlakesRequest represents the input parameters for the flakes function
type FlakesRequest struct {
	Operation   string            `json:"operation"`
	FlakeRef    string            `json:"flake_ref,omitempty"`
	Path        string            `json:"path,omitempty"`
	Package     string            `json:"package,omitempty"`
	Attribute   string            `json:"attribute,omitempty"`
	Template    string            `json:"template,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	Update      []string          `json:"update,omitempty"`
	Interactive bool              `json:"interactive,omitempty"`
	ShowOutput  bool              `json:"show_output,omitempty"`
}

// FlakesResponse represents the output of the flakes function
type FlakesResponse struct {
	Success       bool                   `json:"success"`
	Message       string                 `json:"message"`
	Output        string                 `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
	FlakeInfo     map[string]interface{} `json:"flake_info,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	NextSteps     []string               `json:"next_steps,omitempty"`
	Documentation []string               `json:"documentation,omitempty"`
}

// NewFlakesFunction creates a new flakes function
func NewFlakesFunction() *FlakesFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Flakes operation to perform", true,
			[]string{"init", "update", "check", "show", "build", "develop", "run", "shell",
				"search", "info", "lock", "unlock", "list-inputs", "template", "metadata", "outputs", "help"}, nil, nil),
		functionbase.StringParam("flake_ref", "Flake reference (e.g., github:owner/repo, ., /path/to/flake)", false),
		functionbase.StringParam("path", "Working directory path for flake operations", false),
		functionbase.StringParam("package", "Package or app name to build/run", false),
		functionbase.StringParam("attribute", "Specific flake attribute to target", false),
		functionbase.StringParam("template", "Template name or reference for init operation", false),
		functionbase.ArrayParam("args", "Additional arguments to pass to the command", false),
		functionbase.ObjectParam("options", "Additional options as key-value pairs", false),
		functionbase.ArrayParam("update", "Specific inputs to update (for update operation)", false),
		functionbase.BoolParam("interactive", "Enable interactive mode if available", false),
		functionbase.BoolParam("show_output", "Show detailed command output", false),
	}

	base := functionbase.NewBaseFunction(
		"flakes",
		"Manage Nix flakes - modern package and development environment management",
		parameters,
	)

	return &FlakesFunction{
		BaseFunction: base,
		flakeAgent:   agent.NewFlakeAgent(nil), // Provider will be set when function is executed
		logger:       logger.NewLogger(),
	}
}

// Execute executes the flakes function
func (f *FlakesFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	f.logger.Info("Starting flakes function execution")

	// Parse and validate input
	request, err := f.parseRequest(params)
	if err != nil {
		f.logger.Error("Failed to parse flakes request")
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate request
	if err := f.validateRequest(request); err != nil {
		f.logger.Error("Invalid flakes request")
		return functionbase.CreateErrorResult(err, "Invalid request parameters"), nil
	}

	// Execute the flakes operation
	response, err := f.executeFlakesOperation(ctx, request, options)
	if err != nil {
		f.logger.Error("Failed to execute flakes operation")
		return functionbase.CreateErrorResult(err, "Flakes operation failed"), nil
	}

	f.logger.Info("Flakes function completed successfully")
	return functionbase.CreateSuccessResult(response, "Flakes operation completed"), nil
}

// parseRequest parses the input arguments into a FlakesRequest
func (f *FlakesFunction) parseRequest(params map[string]interface{}) (*FlakesRequest, error) {
	request := &FlakesRequest{}

	// Required operation
	if op, ok := params["operation"].(string); ok {
		request.Operation = op
	} else {
		return nil, fmt.Errorf("operation is required and must be a string")
	}

	// Optional fields
	if flakeRef, ok := params["flake_ref"].(string); ok {
		request.FlakeRef = flakeRef
	}

	if path, ok := params["path"].(string); ok {
		request.Path = path
	}

	if pkg, ok := params["package"].(string); ok {
		request.Package = pkg
	}

	if attr, ok := params["attribute"].(string); ok {
		request.Attribute = attr
	}

	if template, ok := params["template"].(string); ok {
		request.Template = template
	}

	if args, ok := params["args"].([]interface{}); ok {
		request.Args = make([]string, len(args))
		for i, arg := range args {
			if s, ok := arg.(string); ok {
				request.Args[i] = s
			}
		}
	}

	if options, ok := params["options"].(map[string]interface{}); ok {
		request.Options = make(map[string]string)
		for k, v := range options {
			if s, ok := v.(string); ok {
				request.Options[k] = s
			}
		}
	}

	if update, ok := params["update"].([]interface{}); ok {
		request.Update = make([]string, len(update))
		for i, u := range update {
			if s, ok := u.(string); ok {
				request.Update[i] = s
			}
		}
	}

	if interactive, ok := params["interactive"].(bool); ok {
		request.Interactive = interactive
	}

	if showOutput, ok := params["show_output"].(bool); ok {
		request.ShowOutput = showOutput
	}

	return request, nil
}

// validateRequest validates the FlakesRequest
func (f *FlakesFunction) validateRequest(request *FlakesRequest) error {
	// Validate operation
	validOps := map[string]bool{
		"init": true, "update": true, "check": true, "show": true,
		"build": true, "develop": true, "run": true, "shell": true,
		"search": true, "info": true, "lock": true, "unlock": true,
		"list-inputs": true, "template": true, "metadata": true, "outputs": true, "help": true,
	}

	if !validOps[request.Operation] {
		return fmt.Errorf("invalid operation: %s", request.Operation)
	}

	// Operation-specific validation
	switch request.Operation {
	case "init":
		if request.Path == "" {
			request.Path = "." // Default to current directory
		}
	case "run", "build":
		if request.FlakeRef == "" && request.Package == "" {
			return fmt.Errorf("flake_ref or package is required for %s operation", request.Operation)
		}
	case "template":
		if request.Template == "" {
			return fmt.Errorf("template is required for template operation")
		}
	}

	return nil
}

// executeFlakesOperation executes the specified flakes operation
func (f *FlakesFunction) executeFlakesOperation(ctx context.Context, request *FlakesRequest, options *functionbase.FunctionOptions) (*FlakesResponse, error) {
	// Build the prompt for the flake agent
	prompt := f.buildOperationPrompt(request)

	// Query the flake agent
	result, err := f.flakeAgent.Query(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query flake agent: %w", err)
	}

	// Build response based on operation
	response := &FlakesResponse{
		Success: true,
		Message: fmt.Sprintf("Flake %s operation guidance", request.Operation),
		Output:  result,
		Suggestions: []string{
			"Review the provided guidance carefully",
			"Test commands in a safe environment first",
			"Check flake.nix syntax before committing",
		},
		NextSteps: []string{
			"Follow the provided steps in order",
			"Verify results after each step",
			"Check for any error messages",
		},
		Documentation: []string{
			"https://nixos.wiki/wiki/Flakes",
			"https://nix.dev/concepts/flakes",
		},
	}

	return response, nil
}

// buildOperationPrompt builds a prompt for the flake agent based on the operation
func (f *FlakesFunction) buildOperationPrompt(request *FlakesRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Help with Nix flakes %s operation.\n\n", request.Operation))

	prompt.WriteString("Operation Details:\n")
	prompt.WriteString(fmt.Sprintf("- Operation: %s\n", request.Operation))

	if request.FlakeRef != "" {
		prompt.WriteString(fmt.Sprintf("- Flake Reference: %s\n", request.FlakeRef))
	}

	if request.Path != "" {
		prompt.WriteString(fmt.Sprintf("- Working Directory: %s\n", request.Path))
	}

	if request.Package != "" {
		prompt.WriteString(fmt.Sprintf("- Package: %s\n", request.Package))
	}

	if request.Attribute != "" {
		prompt.WriteString(fmt.Sprintf("- Attribute: %s\n", request.Attribute))
	}

	if request.Template != "" {
		prompt.WriteString(fmt.Sprintf("- Template: %s\n", request.Template))
	}

	if len(request.Args) > 0 {
		prompt.WriteString(fmt.Sprintf("- Additional Arguments: %s\n", strings.Join(request.Args, " ")))
	}

	if len(request.Update) > 0 {
		prompt.WriteString(fmt.Sprintf("- Inputs to Update: %s\n", strings.Join(request.Update, ", ")))
	}

	if request.Interactive {
		prompt.WriteString("- Interactive mode preferred\n")
	}

	if request.ShowOutput {
		prompt.WriteString("- Show detailed output\n")
	}

	prompt.WriteString("\nPlease provide:\n")
	prompt.WriteString("1. Exact command(s) to run\n")
	prompt.WriteString("2. Expected behavior and output\n")
	prompt.WriteString("3. Common troubleshooting tips\n")
	prompt.WriteString("4. Best practices for this operation\n")
	prompt.WriteString("5. Any prerequisites or setup needed\n")

	return prompt.String()
}

// validateFlakeRef validates a flake reference format
func (f *FlakesFunction) validateFlakeRef(flakeRef string) error {
	if flakeRef == "" {
		return nil // Empty is valid (defaults to current directory)
	}

	// Common patterns for flake references
	patterns := []string{
		`^github:[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`, // github:owner/repo
		`^gitlab:[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`, // gitlab:owner/repo
		`^git\+https://.*\.git$`,                   // git+https://...
		`^https://.*$`,                             // https://...
		`^file://.*$`,                              // file://...
		`^\./?.*$`,                                 // ./path or /path
		`^/.*$`,                                    // absolute path
		`^[a-zA-Z0-9_.-]+$`,                        // simple name
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, flakeRef)
		if err != nil {
			continue
		}
		if matched {
			return nil
		}
	}

	return fmt.Errorf("invalid flake reference format: %s", flakeRef)
}

// validateRequest validates the FlakesRequest
func (f *FlakesFunction) validateRequest(request *FlakesRequest) error {
	// Validate operation
	validOps := map[string]bool{
		"init": true, "update": true, "check": true, "show": true,
		"build": true, "develop": true, "run": true, "shell": true,
		"search": true, "info": true, "lock": true, "unlock": true,
		"list-inputs": true, "template": true, "metadata": true, "outputs": true, "help": true,
	}

	if !validOps[request.Operation] {
		return fmt.Errorf("invalid operation: %s", request.Operation)
	}

	// Operation-specific validation
	switch request.Operation {
	case "init":
		if request.Path == "" {
			request.Path = "." // Default to current directory
		}
	case "run", "build":
		if request.FlakeRef == "" && request.Package == "" {
			return fmt.Errorf("flake_ref or package is required for %s operation", request.Operation)
		}
	case "template":
		if request.Template == "" {
			return fmt.Errorf("template is required for template operation")
		}
	}

	return nil
}

// executeFlakesOperation executes the specified flakes operation
func (f *FlakesFunction) executeFlakesOperation(ctx context.Context, request *FlakesRequest, options *functionbase.FunctionOptions) (*FlakesResponse, error) {
	// Build the prompt for the flake agent
	prompt := f.buildOperationPrompt(request)

	// Query the flake agent
	result, err := f.flakeAgent.Query(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query flake agent: %w", err)
	}

	// Build response based on operation
	response := &FlakesResponse{
		Success: true,
		Message: fmt.Sprintf("Flake %s operation guidance", request.Operation),
		Output:  result,
		Suggestions: []string{
			"Review the provided guidance carefully",
			"Test commands in a safe environment first",
			"Check flake.nix syntax before committing",
		},
		NextSteps: []string{
			"Follow the provided steps in order",
			"Verify results after each step",
			"Check for any error messages",
		},
		Documentation: []string{
			"https://nixos.wiki/wiki/Flakes",
			"https://nix.dev/concepts/flakes",
		},
	}

	return response, nil
}

// buildOperationPrompt builds a prompt for the flake agent based on the operation
func (f *FlakesFunction) buildOperationPrompt(request *FlakesRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Help with Nix flakes %s operation.\n\n", request.Operation))

	prompt.WriteString("Operation Details:\n")
	prompt.WriteString(fmt.Sprintf("- Operation: %s\n", request.Operation))

	if request.FlakeRef != "" {
		prompt.WriteString(fmt.Sprintf("- Flake Reference: %s\n", request.FlakeRef))
	}

	if request.Path != "" {
		prompt.WriteString(fmt.Sprintf("- Working Directory: %s\n", request.Path))
	}

	if request.Package != "" {
		prompt.WriteString(fmt.Sprintf("- Package: %s\n", request.Package))
	}

	if request.Attribute != "" {
		prompt.WriteString(fmt.Sprintf("- Attribute: %s\n", request.Attribute))
	}

	if request.Template != "" {
		prompt.WriteString(fmt.Sprintf("- Template: %s\n", request.Template))
	}

	if len(request.Args) > 0 {
		prompt.WriteString(fmt.Sprintf("- Additional Arguments: %s\n", strings.Join(request.Args, " ")))
	}

	if len(request.Update) > 0 {
		prompt.WriteString(fmt.Sprintf("- Inputs to Update: %s\n", strings.Join(request.Update, ", ")))
	}

	if request.Interactive {
		prompt.WriteString("- Interactive mode preferred\n")
	}

	if request.ShowOutput {
		prompt.WriteString("- Show detailed output\n")
	}

	prompt.WriteString("\nPlease provide:\n")
	prompt.WriteString("1. Exact command(s) to run\n")
	prompt.WriteString("2. Expected behavior and output\n")
	prompt.WriteString("3. Common troubleshooting tips\n")
	prompt.WriteString("4. Best practices for this operation\n")
	prompt.WriteString("5. Any prerequisites or setup needed\n")

	return prompt.String()
}

// validateFlakeRef validates a flake reference format
func (f *FlakesFunction) validateFlakeRef(flakeRef string) error {
	if flakeRef == "" {
		return nil // Empty is valid (defaults to current directory)
	}

	// Common patterns for flake references
	patterns := []string{
		`^github:[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`, // github:owner/repo
		`^gitlab:[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`, // gitlab:owner/repo
		`^git\+https://.*\.git$`,                   // git+https://...
		`^https://.*$`,                             // https://...
		`^file://.*$`,                              // file://...
		`^\./?.*$`,                                 // ./path or /path
		`^/.*$`,                                    // absolute path
		`^[a-zA-Z0-9_.-]+$`,                        // simple name
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, flakeRef)
		if err != nil {
			continue
		}
		if matched {
			return nil
		}
	}

	return fmt.Errorf("invalid flake reference format: %s", flakeRef)
}
