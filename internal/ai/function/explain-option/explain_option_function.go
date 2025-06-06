package explainoption

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// ExplainOptionFunction implements AI function calling for NixOS option explanation
type ExplainOptionFunction struct {
	*functionbase.BaseFunction
	logger *logger.Logger
}

// ExplainOptionRequest represents the input parameters for the explain-option function
type ExplainOptionRequest struct {
	Option       string `json:"option"`
	Module       string `json:"module,omitempty"`
	ShowExamples bool   `json:"show_examples,omitempty"`
	Detailed     bool   `json:"detailed,omitempty"`
}

// ExplainOptionResponse represents the output of the explain-option function
type ExplainOptionResponse struct {
	Option         string   `json:"option"`
	Description    string   `json:"description"`
	Type           string   `json:"type,omitempty"`
	Default        string   `json:"default,omitempty"`
	Examples       []string `json:"examples,omitempty"`
	RelatedOptions []string `json:"related_options,omitempty"`
	Documentation  string   `json:"documentation,omitempty"`
	Module         string   `json:"module,omitempty"`
}

// NewExplainOptionFunction creates a new explain-option function
func NewExplainOptionFunction() *ExplainOptionFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("option", "The NixOS option to explain (e.g., 'services.openssh.enable')", true),
		functionbase.StringParam("module", "Optional module name to narrow the search", false),
		functionbase.BoolParam("show_examples", "Whether to include usage examples", false, false),
		functionbase.BoolParam("detailed", "Whether to provide detailed explanation", false, false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"explain-option",
		"Explain NixOS configuration options with detailed information and examples",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Explain a basic service option",
			Parameters: map[string]interface{}{
				"option":        "services.openssh.enable",
				"show_examples": true,
			},
			Expected: "Detailed explanation of the SSH service enable option with examples",
		},
	}
	baseFunc.SetSchema(schema)

	return &ExplainOptionFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
	}
}

// ValidateParameters validates the function parameters with custom checks
func (eof *ExplainOptionFunction) ValidateParameters(params map[string]interface{}) error {
	// First run base validation
	if err := eof.BaseFunction.ValidateParameters(params); err != nil {
		return err
	}

	// Custom validation for option parameter
	if option, ok := params["option"].(string); ok {
		if strings.TrimSpace(option) == "" {
			return fmt.Errorf("option parameter cannot be empty")
		}
	}

	return nil
}

// Execute runs the explain-option function
func (eof *ExplainOptionFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	eof.logger.Debug("Starting explain-option function execution")

	// Parse parameters into structured request
	request, err := eof.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate that we have an option
	if request.Option == "" {
		return functionbase.CreateErrorResult(
			fmt.Errorf("option parameter is required and cannot be empty"),
			"Missing required parameter",
		), nil
	}

	// Build the response
	response := &ExplainOptionResponse{
		Option:      request.Option,
		Module:      request.Module,
		Description: eof.generateBasicDescription(request.Option),
	}

	// Enhance response with examples and related options
	if request.ShowExamples {
		response.Examples = eof.generateExamples(request.Option)
	}
	response.RelatedOptions = eof.findRelatedOptions(request.Option)

	eof.logger.Debug("Explain-option function execution completed successfully")

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts raw parameters to structured ExplainOptionRequest
func (eof *ExplainOptionFunction) parseRequest(params map[string]interface{}) (*ExplainOptionRequest, error) {
	request := &ExplainOptionRequest{}

	// Extract option (required)
	if option, ok := params["option"].(string); ok {
		request.Option = strings.TrimSpace(option)
	}

	// Extract optional parameters
	if module, ok := params["module"].(string); ok {
		request.Module = strings.TrimSpace(module)
	}

	if showExamples, ok := params["show_examples"].(bool); ok {
		request.ShowExamples = showExamples
	}

	if detailed, ok := params["detailed"].(bool); ok {
		request.Detailed = detailed
	}

	return request, nil
}

// generateBasicDescription provides a basic description when documentation is not available
func (eof *ExplainOptionFunction) generateBasicDescription(option string) string {
	parts := strings.Split(option, ".")
	if len(parts) < 2 {
		return fmt.Sprintf("NixOS configuration option: %s", option)
	}

	// Generate description based on option structure
	// Check for boot options first, before generic .enable check
	if strings.Contains(option, "boot.") {
		return fmt.Sprintf("Boot-related configuration option: %s", strings.Join(parts[1:], "."))
	}

	if strings.HasSuffix(option, ".enable") {
		service := strings.Join(parts[:len(parts)-1], ".")
		return fmt.Sprintf("Enables the %s service/feature", service)
	}

	if strings.Contains(option, "services.") {
		return fmt.Sprintf("Configuration option for the %s service", strings.Join(parts[1:], "."))
	}

	if strings.Contains(option, "networking.") {
		return fmt.Sprintf("Network configuration option: %s", strings.Join(parts[1:], "."))
	}

	return fmt.Sprintf("NixOS configuration option: %s", option)
}

// generateExamples generates usage examples for the option
func (eof *ExplainOptionFunction) generateExamples(option string) []string {
	var examples []string

	if strings.HasSuffix(option, ".enable") {
		examples = append(examples, fmt.Sprintf("%s = true;", option))
		examples = append(examples, fmt.Sprintf("%s = false;", option))
	} else if strings.Contains(option, ".port") {
		examples = append(examples, fmt.Sprintf("%s = 22;", option))
		examples = append(examples, fmt.Sprintf("%s = 8080;", option))
	} else if strings.Contains(option, ".extraConfig") {
		examples = append(examples, fmt.Sprintf(`%s = ''
  # Additional configuration here
'';`, option))
	} else {
		examples = append(examples, fmt.Sprintf("%s = \"value\";", option))
	}

	return examples
}

// findRelatedOptions finds options related to the given option
func (eof *ExplainOptionFunction) findRelatedOptions(option string) []string {
	var related []string
	parts := strings.Split(option, ".")

	if len(parts) < 2 {
		return related
	}

	// For service options, add common related options
	if strings.Contains(option, "services.") && len(parts) >= 2 {
		service := parts[1]
		if !strings.HasSuffix(option, ".enable") {
			related = append(related, fmt.Sprintf("services.%s.enable", service))
		}
		related = append(related, fmt.Sprintf("services.%s.package", service))
		related = append(related, fmt.Sprintf("services.%s.extraConfig", service))
	}

	// For boot options, add related boot options
	if strings.Contains(option, "boot.") {
		related = append(related, "boot.loader.systemd-boot.enable")
		related = append(related, "boot.loader.grub.enable")
		related = append(related, "boot.kernelPackages")
	}

	// For networking options, add related networking options
	if strings.Contains(option, "networking.") {
		related = append(related, "networking.hostName")
		related = append(related, "networking.firewall.enable")
		related = append(related, "networking.networkmanager.enable")
	}

	// Remove the original option from related options
	filtered := []string{}
	for _, rel := range related {
		if rel != option {
			filtered = append(filtered, rel)
		}
	}

	return filtered
}
