package explain

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/pkg/logger"
)

// ExplainOptionFunction implements AI function calling for NixOS option explanations
type ExplainOptionFunction struct {
	*functionbase.BaseFunction
	explainAgent *agent.ExplainOptionAgent
	mcpClient    *mcp.MCPClient
	logger       *logger.Logger
}

// ExplainOptionRequest represents the input parameters for the explain-option function
type ExplainOptionRequest struct {
	OptionPath    string `json:"option_path"`
	ShowExamples  bool   `json:"show_examples"`
	ShowRelated   bool   `json:"show_related"`
	ShowDefault   bool   `json:"show_default"`
	IncludeUsage  bool   `json:"include_usage"`
	DetailLevel   string `json:"detail_level,omitempty"`
	ContextFilter string `json:"context_filter,omitempty"`
}

// ExplainOptionResponse represents the output of the explain-option function
type ExplainOptionResponse struct {
	OptionPath      string            `json:"option_path"`
	OptionType      string            `json:"option_type,omitempty"`
	Description     string            `json:"description"`
	DefaultValue    string            `json:"default_value,omitempty"`
	Examples        []OptionExample   `json:"examples,omitempty"`
	RelatedOptions  []string          `json:"related_options,omitempty"`
	UsageGuidelines string            `json:"usage_guidelines,omitempty"`
	Category        string            `json:"category,omitempty"`
	PackageName     string            `json:"package_name,omitempty"`
	ServiceName     string            `json:"service_name,omitempty"`
	Documentation   []string          `json:"documentation,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// OptionExample represents a configuration example
type OptionExample struct {
	Description string `json:"description"`
	Code        string `json:"code"`
	Context     string `json:"context,omitempty"`
}

// NewExplainOptionFunction creates a new explain-option function
func NewExplainOptionFunction() *ExplainOptionFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("option_path", "NixOS option path to explain (e.g., 'services.nginx.enable')", true),
		functionbase.BoolParam("show_examples", "Include configuration examples", false),
		functionbase.BoolParam("show_related", "Include related options", false),
		functionbase.BoolParam("show_default", "Show default value", false),
		functionbase.BoolParam("include_usage", "Include usage guidelines", false),
		functionbase.StringParamWithOptions("detail_level", "Level of detail in explanation", false,
			[]string{"basic", "detailed", "comprehensive"}, nil, nil),
		functionbase.StringParam("context_filter", "Filter explanations by context (e.g., 'server', 'desktop')", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"explain-option",
		"Explain NixOS configuration options with examples, related options, and usage guidelines",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Explain a service option",
			Parameters: map[string]interface{}{
				"option_path":   "services.nginx.enable",
				"show_examples": true,
				"show_related":  true,
				"detail_level":  "detailed",
			},
			Expected: "Comprehensive explanation of nginx service enablement with examples",
		},
		{
			Description: "Explain a system option",
			Parameters: map[string]interface{}{
				"option_path":    "boot.loader.systemd-boot.enable",
				"show_examples":  true,
				"show_default":   true,
				"include_usage":  true,
				"context_filter": "desktop",
			},
			Expected: "Detailed explanation of systemd-boot with desktop-specific guidance",
		},
	}
	baseFunc.SetSchema(schema)

	return &ExplainOptionFunction{
		BaseFunction: baseFunc,
		explainAgent: agent.NewExplainOptionAgent(nil, nil), // Provider and MCP client set later
		mcpClient:    nil,                                   // Will be initialized when needed
		logger:       logger.NewLogger(),
	}
}

// Execute runs the explain-option function
func (eof *ExplainOptionFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	eof.logger.Debug("Starting explain-option function execution")

	// Parse parameters into structured request
	request, err := eof.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate that we have an option path
	if request.OptionPath == "" {
		return functionbase.CreateErrorResult(
			fmt.Errorf("option_path parameter is required"),
			"Missing required parameter",
		), nil
	}

	// Report progress if callback is available
	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    1,
			Total:      5,
			Percentage: 20,
			Message:    "Analyzing option path",
			Stage:      "preparation",
		})
	}

	// Build option context
	optionContext := eof.buildOptionContext(request)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    2,
			Total:      5,
			Percentage: 40,
			Message:    "Querying documentation",
			Stage:      "research",
		})
	}

	// Query documentation if MCP client is available
	docInfo := eof.queryDocumentation(ctx, request.OptionPath)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    3,
			Total:      5,
			Percentage: 60,
			Message:    "Generating explanation",
			Stage:      "processing",
		})
	}

	// Query the explain agent
	explanation, err := eof.explainAgent.Query(ctx, optionContext)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to get explanation from AI provider"), nil
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    4,
			Total:      5,
			Percentage: 80,
			Message:    "Building response",
			Stage:      "formatting",
		})
	}

	// Build the response
	response := eof.buildResponse(request, explanation, docInfo)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    5,
			Total:      5,
			Percentage: 100,
			Message:    "Completed successfully",
			Stage:      "complete",
		})
	}

	eof.logger.Debug("Explain-option function execution completed successfully")

	result := &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": "Option explained successfully",
		},
	}
	return result, nil
}

// parseRequest converts raw parameters to structured ExplainOptionRequest
func (eof *ExplainOptionFunction) parseRequest(params map[string]interface{}) (*ExplainOptionRequest, error) {
	request := &ExplainOptionRequest{}

	// Extract option_path (required)
	if optionPath, ok := params["option_path"].(string); ok {
		request.OptionPath = strings.TrimSpace(optionPath)
	}

	// Extract boolean parameters
	if showExamples, ok := params["show_examples"].(bool); ok {
		request.ShowExamples = showExamples
	}

	if showRelated, ok := params["show_related"].(bool); ok {
		request.ShowRelated = showRelated
	}

	if showDefault, ok := params["show_default"].(bool); ok {
		request.ShowDefault = showDefault
	}

	if includeUsage, ok := params["include_usage"].(bool); ok {
		request.IncludeUsage = includeUsage
	}

	// Extract optional string parameters
	if detailLevel, ok := params["detail_level"].(string); ok {
		request.DetailLevel = strings.TrimSpace(detailLevel)
	}

	if contextFilter, ok := params["context_filter"].(string); ok {
		request.ContextFilter = strings.TrimSpace(contextFilter)
	}

	return request, nil
}

// buildOptionContext creates a formatted context string for the AI
func (eof *ExplainOptionFunction) buildOptionContext(request *ExplainOptionRequest) string {
	var contextParts []string

	// Add the main option path
	contextParts = append(contextParts, fmt.Sprintf("Explain NixOS option: %s", request.OptionPath))

	// Add requested details
	var requestedDetails []string
	if request.ShowExamples {
		requestedDetails = append(requestedDetails, "examples")
	}
	if request.ShowRelated {
		requestedDetails = append(requestedDetails, "related options")
	}
	if request.ShowDefault {
		requestedDetails = append(requestedDetails, "default value")
	}
	if request.IncludeUsage {
		requestedDetails = append(requestedDetails, "usage guidelines")
	}

	if len(requestedDetails) > 0 {
		contextParts = append(contextParts, fmt.Sprintf("Include: %s", strings.Join(requestedDetails, ", ")))
	}

	// Add detail level
	if request.DetailLevel != "" {
		contextParts = append(contextParts, fmt.Sprintf("Detail level: %s", request.DetailLevel))
	}

	// Add context filter
	if request.ContextFilter != "" {
		contextParts = append(contextParts, fmt.Sprintf("Context filter: %s", request.ContextFilter))
	}

	return strings.Join(contextParts, "\n")
}

// queryDocumentation queries MCP server for documentation information
func (eof *ExplainOptionFunction) queryDocumentation(ctx context.Context, optionPath string) map[string]interface{} {
	docInfo := make(map[string]interface{})

	// If MCP client is not available, return empty info
	if eof.mcpClient == nil {
		eof.logger.Debug("MCP client not available, skipping documentation query")
		return docInfo
	}

	// Query documentation sources for the option
	query := fmt.Sprintf("NixOS option %s", optionPath)

	// Try to get documentation from various sources
	// This would be implemented based on the actual MCP client interface
	eof.logger.Debug(fmt.Sprintf("Querying documentation for option %s with query: %s", optionPath, query))
	eof.logger.Debug(fmt.Sprintf("Querying documentation for option: %s", optionPath))

	return docInfo
}

// buildResponse creates the structured response
func (eof *ExplainOptionFunction) buildResponse(request *ExplainOptionRequest, explanation string, docInfo map[string]interface{}) *ExplainOptionResponse {
	response := &ExplainOptionResponse{
		OptionPath:  request.OptionPath,
		Description: explanation,
	}

	// Parse option path to extract information
	eof.parseOptionPath(request.OptionPath, response)

	// Add examples if requested
	if request.ShowExamples {
		response.Examples = eof.generateExamples(request.OptionPath, request.ContextFilter)
	}

	// Add related options if requested
	if request.ShowRelated {
		response.RelatedOptions = eof.findRelatedOptions(request.OptionPath)
	}

	// Add usage guidelines if requested
	if request.IncludeUsage {
		response.UsageGuidelines = eof.generateUsageGuidelines(request.OptionPath, request.ContextFilter)
	}

	// Add documentation references
	response.Documentation = eof.generateDocumentationRefs(request.OptionPath)

	// Add metadata
	response.Metadata = map[string]string{
		"detail_level": request.DetailLevel,
	}
	if request.ContextFilter != "" {
		response.Metadata["context_filter"] = request.ContextFilter
	}

	return response
}

// parseOptionPath extracts information from the option path
func (eof *ExplainOptionFunction) parseOptionPath(optionPath string, response *ExplainOptionResponse) {
	parts := strings.Split(optionPath, ".")

	if len(parts) > 0 {
		response.Category = parts[0]
	}

	// Extract service name if it's a service option
	if len(parts) >= 2 && parts[0] == "services" {
		response.ServiceName = parts[1]
		response.PackageName = parts[1] // Often the same for services
	}

	// Extract program name if it's a programs option
	if len(parts) >= 2 && parts[0] == "programs" {
		response.PackageName = parts[1]
	}

	// Determine option type based on the last part
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		switch lastPart {
		case "enable":
			response.OptionType = "boolean"
			response.DefaultValue = "false"
		case "package":
			response.OptionType = "package"
		case "config", "extraConfig":
			response.OptionType = "string"
		default:
			response.OptionType = "unknown"
		}
	}
}

// generateExamples creates configuration examples for the option
func (eof *ExplainOptionFunction) generateExamples(optionPath, contextFilter string) []OptionExample {
	var examples []OptionExample

	// Generate basic example
	basicExample := OptionExample{
		Description: fmt.Sprintf("Basic usage of %s", optionPath),
		Code:        fmt.Sprintf("%s = true;", optionPath),
		Context:     "Basic configuration",
	}

	// Customize based on option type
	if strings.HasSuffix(optionPath, ".enable") {
		examples = append(examples, basicExample)
	} else if strings.Contains(optionPath, "config") {
		examples = append(examples, OptionExample{
			Description: fmt.Sprintf("Configuration example for %s", optionPath),
			Code:        fmt.Sprintf("%s = ''\n  # Your configuration here\n'';", optionPath),
			Context:     "Configuration file content",
		})
	}

	// Add context-specific examples
	if contextFilter == "server" {
		examples = append(examples, OptionExample{
			Description: "Server-optimized configuration",
			Code:        fmt.Sprintf("%s = true; # Server setup", optionPath),
			Context:     "Server environment",
		})
	} else if contextFilter == "desktop" {
		examples = append(examples, OptionExample{
			Description: "Desktop-friendly configuration",
			Code:        fmt.Sprintf("%s = true; # Desktop setup", optionPath),
			Context:     "Desktop environment",
		})
	}

	return examples
}

// findRelatedOptions finds options related to the given option path
func (eof *ExplainOptionFunction) findRelatedOptions(optionPath string) []string {
	var related []string

	parts := strings.Split(optionPath, ".")

	if len(parts) >= 2 {
		// Find options in the same category
		prefix := strings.Join(parts[:2], ".")

		// Add prefix-related options
		related = append(related, prefix+".enable")

		// Common related patterns
		if strings.HasSuffix(optionPath, ".enable") {
			basePrefix := strings.TrimSuffix(optionPath, ".enable")
			related = append(related,
				basePrefix+".package",
				basePrefix+".config",
				basePrefix+".extraConfig",
			)
		}

		// Service-specific relations
		if parts[0] == "services" && len(parts) >= 2 {
			serviceName := parts[1]
			related = append(related,
				fmt.Sprintf("networking.firewall.allowedTCPPorts"), // Common for services
				fmt.Sprintf("services.%s.user", serviceName),
				fmt.Sprintf("services.%s.group", serviceName),
			)
		}

		// Boot-related options
		if parts[0] == "boot" {
			related = append(related,
				"boot.loader.grub.enable",
				"boot.loader.systemd-boot.enable",
				"boot.kernelPackages",
			)
		}
	}

	return related
}

// generateUsageGuidelines creates usage guidelines for the option
func (eof *ExplainOptionFunction) generateUsageGuidelines(optionPath, contextFilter string) string {
	var guidelines []string

	// General guidelines
	guidelines = append(guidelines, "1. Add this option to your NixOS configuration.nix")
	guidelines = append(guidelines, "2. Run 'nixos-rebuild switch' to apply changes")

	// Option-specific guidelines
	if strings.HasSuffix(optionPath, ".enable") {
		guidelines = append(guidelines, "3. This is a boolean option (true/false)")
		guidelines = append(guidelines, "4. Setting to 'true' enables the feature/service")
	}

	// Context-specific guidelines
	if contextFilter == "server" {
		guidelines = append(guidelines, "5. Consider security implications for server deployment")
		guidelines = append(guidelines, "6. Review firewall settings if network services are involved")
	} else if contextFilter == "desktop" {
		guidelines = append(guidelines, "5. Consider user experience and desktop integration")
		guidelines = append(guidelines, "6. Check compatibility with your desktop environment")
	}

	// Service-specific guidelines
	if strings.Contains(optionPath, "services.") {
		guidelines = append(guidelines, "7. Check service status with 'systemctl status <service>'")
		guidelines = append(guidelines, "8. Review logs with 'journalctl -u <service>'")
	}

	return strings.Join(guidelines, "\n")
}

// generateDocumentationRefs provides relevant documentation links
func (eof *ExplainOptionFunction) generateDocumentationRefs(optionPath string) []string {
	var refs []string

	// Always include general NixOS documentation
	refs = append(refs, "https://nixos.org/manual/nixos/stable/")
	refs = append(refs, "https://wiki.nixos.org/")

	// Add specific documentation based on option category
	parts := strings.Split(optionPath, ".")
	if len(parts) > 0 {
		switch parts[0] {
		case "services":
			refs = append(refs, "https://wiki.nixos.org/wiki/Category:Services")
		case "boot":
			refs = append(refs, "https://wiki.nixos.org/wiki/Bootloader")
		case "networking":
			refs = append(refs, "https://wiki.nixos.org/wiki/Networking")
		case "security":
			refs = append(refs, "https://wiki.nixos.org/wiki/Security")
		case "programs":
			refs = append(refs, "https://wiki.nixos.org/wiki/Category:Applications")
		}
	}

	// Add package-specific documentation if available
	if len(parts) >= 2 && (parts[0] == "services" || parts[0] == "programs") {
		packageName := parts[1]
		refs = append(refs, fmt.Sprintf("https://search.nixos.org/packages?query=%s", packageName))
	}

	return refs
}
