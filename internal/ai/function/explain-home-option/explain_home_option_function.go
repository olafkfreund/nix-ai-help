package explainhomeoption

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// ExplainHomeOptionFunction provides explanations for Home Manager options
type ExplainHomeOptionFunction struct {
	*functionbase.BaseFunction
	logger *logger.Logger
}

// ExplainHomeOptionRequest represents the input parameters for explain-home-option
type ExplainHomeOptionRequest struct {
	Option       string `json:"option"`
	Module       string `json:"module,omitempty"`
	ShowExamples bool   `json:"show_examples,omitempty"`
	Detailed     bool   `json:"detailed,omitempty"`
}

// ExplainHomeOptionResponse represents the output of the explain-home-option function
type ExplainHomeOptionResponse struct {
	Option         string   `json:"option"`
	Description    string   `json:"description"`
	Type           string   `json:"type,omitempty"`
	Default        string   `json:"default,omitempty"`
	Examples       []string `json:"examples,omitempty"`
	RelatedOptions []string `json:"related_options,omitempty"`
	Documentation  string   `json:"documentation,omitempty"`
	Module         string   `json:"module,omitempty"`
}

// NewExplainHomeOptionFunction creates a new explain-home-option function
func NewExplainHomeOptionFunction() *ExplainHomeOptionFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("option", "The Home Manager option to explain (e.g., 'programs.git.enable')", true),
		functionbase.StringParam("module", "Optional module name to narrow the search", false),
		functionbase.BoolParam("show_examples", "Whether to include usage examples", false, false),
		functionbase.BoolParam("detailed", "Whether to provide detailed explanation", false, false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"explain-home-option",
		"Explain Home Manager configuration options with detailed information and examples",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Explain a basic program option",
			Parameters: map[string]interface{}{
				"option":        "programs.git.enable",
				"show_examples": true,
			},
			Expected: "Detailed explanation of the Git program enable option with examples",
		},
	}
	baseFunc.SetSchema(schema)

	return &ExplainHomeOptionFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
	}
}

// ValidateParameters validates the function parameters with custom checks
func (ehof *ExplainHomeOptionFunction) ValidateParameters(params map[string]interface{}) error {
	// First run base validation
	if err := ehof.BaseFunction.ValidateParameters(params); err != nil {
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

// Execute runs the explain-home-option function
func (ehof *ExplainHomeOptionFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	ehof.logger.Debug("Starting explain-home-option function execution")

	// Parse parameters into structured request
	request, err := ehof.parseRequest(params)
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
	response := &ExplainHomeOptionResponse{
		Option:      request.Option,
		Module:      request.Module,
		Description: ehof.generateBasicDescription(request.Option),
	}

	// Enhance response with examples and related options
	if request.ShowExamples {
		response.Examples = ehof.generateExamples(request.Option)
	}
	response.RelatedOptions = ehof.findRelatedOptions(request.Option)

	ehof.logger.Debug("Explain-home-option function execution completed successfully")

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts raw parameters to structured ExplainHomeOptionRequest
func (ehof *ExplainHomeOptionFunction) parseRequest(params map[string]interface{}) (*ExplainHomeOptionRequest, error) {
	request := &ExplainHomeOptionRequest{}

	// Extract option (required)
	if option, ok := params["option"].(string); ok {
		request.Option = strings.TrimSpace(option)
	}

	// Extract module (optional)
	if module, ok := params["module"].(string); ok {
		request.Module = strings.TrimSpace(module)
	}

	// Extract show_examples (optional, default false)
	if showExamples, ok := params["show_examples"].(bool); ok {
		request.ShowExamples = showExamples
	}

	// Extract detailed (optional, default false)
	if detailed, ok := params["detailed"].(bool); ok {
		request.Detailed = detailed
	}

	return request, nil
}

// generateBasicDescription creates a basic description for the Home Manager option
func (ehof *ExplainHomeOptionFunction) generateBasicDescription(option string) string {
	// Handle common Home Manager option patterns
	parts := strings.Split(option, ".")
	if len(parts) < 2 {
		return fmt.Sprintf("Home Manager option: %s", option)
	}

	category := parts[0]
	subcategory := parts[1]

	// Handle specific Home Manager patterns
	switch {
	case category == "programs" && strings.HasSuffix(option, ".enable"):
		program := strings.Join(parts[1:len(parts)-1], ".")
		return fmt.Sprintf("Enable the %s program in Home Manager. When enabled, Home Manager will configure and manage the %s program settings.", program, program)

	case category == "services" && strings.HasSuffix(option, ".enable"):
		service := strings.Join(parts[1:len(parts)-1], ".")
		return fmt.Sprintf("Enable the %s service in Home Manager. When enabled, Home Manager will start and manage the %s service for the user.", service, service)

	case category == "home" && subcategory == "sessionVariables":
		return "Define environment variables that will be set in your shell session. These variables are available to all programs started from your shell."

	case category == "home" && subcategory == "packages":
		return "List of packages to install in the user environment via Home Manager. These packages will be available to the user."

	case category == "xdg" && subcategory == "enable":
		return "Enable XDG Base Directory specification support in Home Manager. This helps organize configuration files according to XDG standards."

	case category == "programs" && subcategory == "bash":
		return fmt.Sprintf("Configure bash shell options: %s", strings.Join(parts[2:], "."))

	case category == "programs" && subcategory == "zsh":
		return fmt.Sprintf("Configure zsh shell options: %s", strings.Join(parts[2:], "."))

	case category == "programs" && subcategory == "git":
		return fmt.Sprintf("Configure Git version control options: %s", strings.Join(parts[2:], "."))

	case category == "programs" && subcategory == "vim" || subcategory == "neovim":
		return fmt.Sprintf("Configure %s editor options: %s", subcategory, strings.Join(parts[2:], "."))

	default:
		return fmt.Sprintf("Home Manager %s configuration option: %s", category, option)
	}
}

// generateExamples creates usage examples for the Home Manager option
func (ehof *ExplainHomeOptionFunction) generateExamples(option string) []string {
	var examples []string

	// Handle common Home Manager option patterns
	parts := strings.Split(option, ".")
	if len(parts) < 2 {
		return []string{fmt.Sprintf("%s = true;", option)}
	}

	category := parts[0]
	subcategory := parts[1]

	switch {
	case category == "programs" && strings.HasSuffix(option, ".enable"):
		program := strings.Join(parts[1:len(parts)-1], ".")
		examples = append(examples, fmt.Sprintf("programs.%s.enable = true;", program))

		// Add common configuration examples
		if program == "git" {
			examples = append(examples, `programs.git = {
  enable = true;
  userName = "Your Name";
  userEmail = "your.email@example.com";
};`)
		} else if program == "bash" || program == "zsh" {
			examples = append(examples, fmt.Sprintf(`programs.%s = {
  enable = true;
  enableCompletion = true;
  shellAliases = {
    ll = "ls -l";
    la = "ls -la";
  };
};`, program))
		}

	case category == "services" && strings.HasSuffix(option, ".enable"):
		service := strings.Join(parts[1:len(parts)-1], ".")
		examples = append(examples, fmt.Sprintf("services.%s.enable = true;", service))

	case category == "home" && subcategory == "packages":
		examples = append(examples, `home.packages = with pkgs; [
  firefox
  git
  neovim
];`)

	case category == "home" && subcategory == "sessionVariables":
		examples = append(examples, `home.sessionVariables = {
  EDITOR = "nvim";
  BROWSER = "firefox";
};`)

	default:
		examples = append(examples, fmt.Sprintf("%s = true;", option))
	}

	return examples
}

// findRelatedOptions finds options related to the given Home Manager option
func (ehof *ExplainHomeOptionFunction) findRelatedOptions(option string) []string {
	var related []string

	parts := strings.Split(option, ".")
	if len(parts) < 2 {
		return related
	}

	category := parts[0]
	subcategory := parts[1]

	// Add related options based on the option category
	switch {
	case category == "programs" && subcategory == "git":
		related = append(related, []string{
			"programs.git.enable",
			"programs.git.userName",
			"programs.git.userEmail",
			"programs.git.aliases",
			"programs.git.extraConfig",
		}...)

	case category == "programs" && (subcategory == "bash" || subcategory == "zsh"):
		related = append(related, []string{
			fmt.Sprintf("programs.%s.enable", subcategory),
			fmt.Sprintf("programs.%s.enableCompletion", subcategory),
			fmt.Sprintf("programs.%s.shellAliases", subcategory),
			fmt.Sprintf("programs.%s.initExtra", subcategory),
		}...)

	case category == "programs" && (subcategory == "vim" || subcategory == "neovim"):
		related = append(related, []string{
			fmt.Sprintf("programs.%s.enable", subcategory),
			fmt.Sprintf("programs.%s.defaultEditor", subcategory),
			fmt.Sprintf("programs.%s.extraConfig", subcategory),
		}...)

	case category == "home":
		related = append(related, []string{
			"home.packages",
			"home.sessionVariables",
			"home.file",
			"home.stateVersion",
		}...)

	case category == "xdg":
		related = append(related, []string{
			"xdg.enable",
			"xdg.configFile",
			"xdg.dataFile",
		}...)
	}

	// Filter out the current option from related options
	filtered := make([]string, 0, len(related))
	for _, rel := range related {
		if rel != option {
			filtered = append(filtered, rel)
		}
	}

	return filtered
}
