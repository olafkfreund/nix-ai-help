package packages

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// PackagesFunction implements AI function calling for package management
type PackagesFunction struct {
	*functionbase.BaseFunction
	searchAgent *agent.SearchAgent
	logger      *logger.Logger
}

// PackagesRequest represents the input parameters for the packages function
type PackagesRequest struct {
	Operation     string            `json:"operation"`
	PackageName   string            `json:"package_name,omitempty"`
	SearchQuery   string            `json:"search_query,omitempty"`
	Channel       string            `json:"channel,omitempty"`
	SystemArch    string            `json:"system_arch,omitempty"`
	IncludeUnfree bool              `json:"include_unfree,omitempty"`
	ShowDetails   bool              `json:"show_details,omitempty"`
	Limit         int               `json:"limit,omitempty"`
	SortBy        string            `json:"sort_by,omitempty"`
	Filters       map[string]string `json:"filters,omitempty"`
	InstallMethod string            `json:"install_method,omitempty"`
	ConfigType    string            `json:"config_type,omitempty"`
	ShowVersions  bool              `json:"show_versions,omitempty"`
	Dependencies  bool              `json:"dependencies,omitempty"`
	Sources       []string          `json:"sources,omitempty"`
}

// PackagesResponse represents the output of the packages function
type PackagesResponse struct {
	Success          bool                     `json:"success"`
	Message          string                   `json:"message"`
	Output           string                   `json:"output,omitempty"`
	Error            string                   `json:"error,omitempty"`
	Packages         []map[string]interface{} `json:"packages,omitempty"`
	InstallCommands  []string                 `json:"install_commands,omitempty"`
	ConfigSnippets   []string                 `json:"config_snippets,omitempty"`
	Suggestions      []string                 `json:"suggestions,omitempty"`
	NextSteps        []string                 `json:"next_steps,omitempty"`
	Documentation    []string                 `json:"documentation,omitempty"`
	AlternativeNames []string                 `json:"alternative_names,omitempty"`
}

// NewPackagesFunction creates a new packages function
func NewPackagesFunction() *PackagesFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("operation", "Package operation to perform", true),
		functionbase.StringParam("package_name", "Name of the package to search, install, or query", false),
		functionbase.StringParam("search_query", "Search query for finding packages", false),
		functionbase.StringParam("channel", "NixOS channel to search (unstable, 23.11, 24.05, etc.)", false),
		functionbase.StringParam("system_arch", "System architecture (x86_64-linux, aarch64-linux, etc.)", false),
		functionbase.BoolParam("include_unfree", "Include unfree packages in search results", false),
		functionbase.BoolParam("show_details", "Show detailed package information", false),
		functionbase.IntParam("limit", "Maximum number of results to return", false),
		functionbase.StringParam("sort_by", "Sort results by (relevance, name, popularity, updated)", false),
		functionbase.ObjectParam("filters", "Additional filters for package search", false),
		functionbase.StringParam("install_method", "Installation method preference (declarative, imperative, flakes)", false),
		functionbase.StringParam("config_type", "Configuration type (system, user, flake)", false),
		functionbase.BoolParam("show_versions", "Show available package versions", false),
		functionbase.BoolParam("dependencies", "Show package dependencies", false),
		functionbase.ArrayParam("sources", "Package sources to search (nixpkgs, nur, flakes)", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"packages",
		"Provides AI-powered assistance for NixOS package management including search, installation, and configuration guidance",
		parameters,
	)

	return &PackagesFunction{
		BaseFunction: baseFunc,
		searchAgent:  agent.NewSearchAgent(nil),
		logger:       logger.NewLogger(),
	}
}

// Execute runs the packages function with the given parameters
func (f *PackagesFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	// Parse parameters into request struct
	request, err := f.parseRequest(params)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse request parameters: %v", err),
		}, nil
	}

	// Validate the request
	if err := f.validateRequest(request); err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("request validation failed: %v", err),
		}, nil
	}

	// Execute the package operation
	response, err := f.executePackageOperation(ctx, request, options)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to execute package operation: %v", err),
		}, nil
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts the parameters map into a PackagesRequest struct
func (f *PackagesFunction) parseRequest(params map[string]interface{}) (*PackagesRequest, error) {
	request := &PackagesRequest{
		Limit:  10,          // Default limit
		SortBy: "relevance", // Default sort
	}

	// Parse operation (required)
	if op, ok := params["operation"].(string); ok {
		request.Operation = op
	} else {
		return nil, fmt.Errorf("operation parameter is required")
	}

	// Parse optional parameters
	if name, ok := params["package_name"].(string); ok {
		request.PackageName = name
	}

	if query, ok := params["search_query"].(string); ok {
		request.SearchQuery = query
	}

	if channel, ok := params["channel"].(string); ok {
		request.Channel = channel
	}

	if arch, ok := params["system_arch"].(string); ok {
		request.SystemArch = arch
	}

	if unfree, ok := params["include_unfree"].(bool); ok {
		request.IncludeUnfree = unfree
	}

	if details, ok := params["show_details"].(bool); ok {
		request.ShowDetails = details
	}

	if limit, ok := params["limit"].(int); ok {
		request.Limit = limit
	}

	if sortBy, ok := params["sort_by"].(string); ok {
		request.SortBy = sortBy
	}

	if filters, ok := params["filters"].(map[string]interface{}); ok {
		request.Filters = make(map[string]string)
		for k, v := range filters {
			if str, ok := v.(string); ok {
				request.Filters[k] = str
			}
		}
	}

	if method, ok := params["install_method"].(string); ok {
		request.InstallMethod = method
	}

	if configType, ok := params["config_type"].(string); ok {
		request.ConfigType = configType
	}

	if versions, ok := params["show_versions"].(bool); ok {
		request.ShowVersions = versions
	}

	if deps, ok := params["dependencies"].(bool); ok {
		request.Dependencies = deps
	}

	if sources, ok := params["sources"].([]interface{}); ok {
		request.Sources = make([]string, len(sources))
		for i, src := range sources {
			if str, ok := src.(string); ok {
				request.Sources[i] = str
			}
		}
	}

	return request, nil
}

// validateRequest validates the PackagesRequest
func (f *PackagesFunction) validateRequest(request *PackagesRequest) error {
	// Validate operation
	validOps := map[string]bool{
		"search": true, "install": true, "info": true, "list": true,
		"update": true, "remove": true, "versions": true, "dependencies": true,
		"alternatives": true, "compare": true, "help": true,
	}

	if !validOps[request.Operation] {
		return fmt.Errorf("invalid operation: %s", request.Operation)
	}

	// Operation-specific validation
	switch request.Operation {
	case "install", "info", "versions", "dependencies", "remove":
		if request.PackageName == "" {
			return fmt.Errorf("package_name is required for %s operation", request.Operation)
		}
	case "search":
		if request.SearchQuery == "" && request.PackageName == "" {
			return fmt.Errorf("either search_query or package_name is required for search operation")
		}
	case "compare":
		if request.PackageName == "" {
			return fmt.Errorf("package_name is required for compare operation")
		}
	}

	// Validate limit
	if request.Limit < 1 || request.Limit > 100 {
		request.Limit = 10 // Reset to default
	}

	// Validate sort_by
	validSorts := map[string]bool{
		"relevance": true, "name": true, "popularity": true, "updated": true,
	}
	if !validSorts[request.SortBy] {
		request.SortBy = "relevance" // Reset to default
	}

	return nil
}

// executePackageOperation executes the specified package operation
func (f *PackagesFunction) executePackageOperation(ctx context.Context, request *PackagesRequest, options *functionbase.FunctionOptions) (*PackagesResponse, error) {
	// Build the prompt for the search agent
	prompt := f.buildOperationPrompt(request)

	// Query the search agent
	result, err := f.searchAgent.Query(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query search agent: %w", err)
	}

	// Build response based on operation
	response := &PackagesResponse{
		Success: true,
		Message: fmt.Sprintf("Package %s operation completed", request.Operation),
		Output:  result,
		Suggestions: []string{
			"Verify package availability in your channel",
			"Check system compatibility before installation",
			"Consider declarative configuration for system packages",
		},
		NextSteps: []string{
			"Review the provided package information",
			"Test installation in a safe environment first",
			"Update your configuration as needed",
		},
		Documentation: []string{
			"https://nixos.wiki/wiki/Nix_package_manager",
			"https://search.nixos.org/packages",
			"https://nix.dev/manual/nix/2.28/command-ref/nix-env",
		},
	}

	// Add operation-specific guidance
	switch request.Operation {
	case "install":
		response.InstallCommands = []string{
			fmt.Sprintf("nix-env -iA nixpkgs.%s", request.PackageName),
			fmt.Sprintf("nix profile install nixpkgs#%s", request.PackageName),
		}
		if request.ConfigType == "system" {
			response.ConfigSnippets = []string{
				fmt.Sprintf("environment.systemPackages = with pkgs; [ %s ];", request.PackageName),
			}
		} else if request.ConfigType == "user" {
			response.ConfigSnippets = []string{
				fmt.Sprintf("home.packages = with pkgs; [ %s ];", request.PackageName),
			}
		}
	case "search":
		response.AlternativeNames = []string{
			"Check package name variations",
			"Look for similar packages",
			"Consider meta-packages",
		}
	}

	return response, nil
}

// buildOperationPrompt builds a prompt for the search agent based on the operation
func (f *PackagesFunction) buildOperationPrompt(request *PackagesRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Help with NixOS package %s operation.\n\n", request.Operation))

	prompt.WriteString("Operation Details:\n")
	prompt.WriteString(fmt.Sprintf("- Operation: %s\n", request.Operation))

	if request.PackageName != "" {
		prompt.WriteString(fmt.Sprintf("- Package Name: %s\n", request.PackageName))
	}

	if request.SearchQuery != "" {
		prompt.WriteString(fmt.Sprintf("- Search Query: %s\n", request.SearchQuery))
	}

	if request.Channel != "" {
		prompt.WriteString(fmt.Sprintf("- Channel: %s\n", request.Channel))
	}

	if request.SystemArch != "" {
		prompt.WriteString(fmt.Sprintf("- System Architecture: %s\n", request.SystemArch))
	}

	if request.IncludeUnfree {
		prompt.WriteString("- Include unfree packages: Yes\n")
	}

	if request.ShowDetails {
		prompt.WriteString("- Show detailed information: Yes\n")
	}

	if request.Limit > 0 {
		prompt.WriteString(fmt.Sprintf("- Result limit: %d\n", request.Limit))
	}

	if request.SortBy != "" {
		prompt.WriteString(fmt.Sprintf("- Sort by: %s\n", request.SortBy))
	}

	if request.InstallMethod != "" {
		prompt.WriteString(fmt.Sprintf("- Preferred installation method: %s\n", request.InstallMethod))
	}

	if request.ConfigType != "" {
		prompt.WriteString(fmt.Sprintf("- Configuration type: %s\n", request.ConfigType))
	}

	if request.ShowVersions {
		prompt.WriteString("- Show versions: Yes\n")
	}

	if request.Dependencies {
		prompt.WriteString("- Show dependencies: Yes\n")
	}

	if len(request.Sources) > 0 {
		prompt.WriteString(fmt.Sprintf("- Package sources: %s\n", strings.Join(request.Sources, ", ")))
	}

	if len(request.Filters) > 0 {
		prompt.WriteString("- Additional filters:\n")
		for k, v := range request.Filters {
			prompt.WriteString(fmt.Sprintf("  - %s: %s\n", k, v))
		}
	}

	prompt.WriteString("\nPlease provide:\n")
	prompt.WriteString("1. Relevant package information and search results\n")
	prompt.WriteString("2. Installation instructions and commands\n")
	prompt.WriteString("3. Configuration examples (system/user/flakes)\n")
	prompt.WriteString("4. Version and compatibility information\n")
	prompt.WriteString("5. Dependencies and alternatives\n")
	prompt.WriteString("6. Troubleshooting tips and best practices\n")

	return prompt.String()
}
