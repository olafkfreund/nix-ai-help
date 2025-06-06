package build

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// BuildFunction implements AI function calling for build operations and troubleshooting
type BuildFunction struct {
	*functionbase.BaseFunction
	buildAgent *agent.BuildAgent
	logger     *logger.Logger
}

// BuildRequest represents the input parameters for the build function
type BuildRequest struct {
	Operation     string   `json:"operation"`
	Package       string   `json:"package,omitempty"`
	Configuration string   `json:"configuration,omitempty"`
	ErrorLogs     string   `json:"error_logs,omitempty"`
	BuildOptions  []string `json:"build_options,omitempty"`
	System        string   `json:"system,omitempty"`
	Flake         bool     `json:"flake,omitempty"`
	Impure        bool     `json:"impure,omitempty"`
	KeepGoing     bool     `json:"keep_going,omitempty"`
	ShowTrace     bool     `json:"show_trace,omitempty"`
	Verbose       bool     `json:"verbose,omitempty"`
	MaxJobs       int      `json:"max_jobs,omitempty"`
	Cores         int      `json:"cores,omitempty"`
}

// BuildResponse represents the output of the build function
type BuildResponse struct {
	Status              string   `json:"status"`
	Solution            string   `json:"solution"`
	DiagnosisDetails    string   `json:"diagnosis_details,omitempty"`
	SuggestedCommands   []string `json:"suggested_commands,omitempty"`
	TroubleshootingTips []string `json:"troubleshooting_tips,omitempty"`
	OptimizationTips    []string `json:"optimization_tips,omitempty"`
	DocumentationRefs   []string `json:"documentation_refs,omitempty"`
	EstimatedTime       string   `json:"estimated_time,omitempty"`
	Dependencies        []string `json:"dependencies,omitempty"`
}

// NewBuildFunction creates a new build function
func NewBuildFunction() *BuildFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Type of build operation to perform", true,
			[]string{"build", "troubleshoot", "optimize", "analyze", "clean", "rebuild", "check"}, nil, nil),
		functionbase.StringParam("package", "Package name or derivation path to build", false),
		functionbase.StringParam("configuration", "Build configuration or flake reference", false),
		functionbase.StringParam("error_logs", "Build error logs or output for troubleshooting", false),
		{
			Name:        "build_options",
			Type:        "array",
			Description: "Additional build options and flags",
			Required:    false,
			Items: &functionbase.FunctionParameter{
				Type: "string",
			},
		},
		functionbase.StringParamWithOptions("system", "Target system architecture", false,
			[]string{"x86_64-linux", "aarch64-linux", "x86_64-darwin", "aarch64-darwin"}, nil, nil),
		functionbase.BoolParam("flake", "Whether this is a flake-based build", false),
		functionbase.BoolParam("impure", "Allow impure builds", false),
		functionbase.BoolParam("keep_going", "Keep building other derivations on failure", false),
		functionbase.BoolParam("show_trace", "Show detailed error traces", false),
		functionbase.BoolParam("verbose", "Enable verbose output", false),
		{
			Name:        "max_jobs",
			Type:        "integer",
			Description: "Maximum number of build jobs",
			Required:    false,
			Minimum:     func() *float64 { v := 1.0; return &v }(),
			Maximum:     func() *float64 { v := 128.0; return &v }(),
		},
		{
			Name:        "cores",
			Type:        "integer",
			Description: "Number of CPU cores to use per job",
			Required:    false,
			Minimum:     func() *float64 { v := 1.0; return &v }(),
			Maximum:     func() *float64 { v := 64.0; return &v }(),
		},
	}

	baseFunc := functionbase.NewBaseFunction(
		"build",
		"Handle NixOS and Nix build operations including building packages, troubleshooting build failures, optimizing build performance, and analyzing build issues",
		parameters,
	)

	return &BuildFunction{
		BaseFunction: baseFunc,
		buildAgent:   agent.NewBuildAgent(nil), // Will be set when executing
		logger:       logger.NewLogger(),
	}
}

// Execute performs the build operation
func (bf *BuildFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	// Validate parameters
	if err := bf.ValidateParameters(params); err != nil {
		return functionbase.ErrorResult(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Parse parameters into request struct
	req, err := bf.parseRequest(params)
	if err != nil {
		return functionbase.ErrorResult(fmt.Sprintf("Failed to parse request: %v", err)), nil
	}

	// Validate operation
	if err := bf.validateOperation(req); err != nil {
		return functionbase.ErrorResult(err.Error()), nil
	}

	bf.logger.Info(fmt.Sprintf("Executing build function with operation: %s", req.Operation))

	// Initialize build agent with provider if available
	if options != nil && options.Provider != nil {
		bf.buildAgent = agent.NewBuildAgent(options.Provider)
	}

	// Execute the build operation
	response, err := bf.executeBuildOperation(ctx, req)
	if err != nil {
		return functionbase.ErrorResult(fmt.Sprintf("Build operation failed: %v", err)), nil
	}

	return functionbase.SuccessResult(response), nil
}

// parseRequest converts the parameters map to a structured request
func (bf *BuildFunction) parseRequest(params map[string]interface{}) (*BuildRequest, error) {
	req := &BuildRequest{}

	// Required parameters
	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	} else {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	// Optional parameters
	if package_name, ok := params["package"].(string); ok {
		req.Package = package_name
	}

	if configuration, ok := params["configuration"].(string); ok {
		req.Configuration = configuration
	}

	if errorLogs, ok := params["error_logs"].(string); ok {
		req.ErrorLogs = errorLogs
	}

	if buildOptions, ok := params["build_options"].([]interface{}); ok {
		for _, opt := range buildOptions {
			if optStr, ok := opt.(string); ok {
				req.BuildOptions = append(req.BuildOptions, optStr)
			}
		}
	}

	if system, ok := params["system"].(string); ok {
		req.System = system
	}

	if flake, ok := params["flake"].(bool); ok {
		req.Flake = flake
	}

	if impure, ok := params["impure"].(bool); ok {
		req.Impure = impure
	}

	if keepGoing, ok := params["keep_going"].(bool); ok {
		req.KeepGoing = keepGoing
	}

	if showTrace, ok := params["show_trace"].(bool); ok {
		req.ShowTrace = showTrace
	}

	if verbose, ok := params["verbose"].(bool); ok {
		req.Verbose = verbose
	}

	if maxJobs, ok := params["max_jobs"].(float64); ok {
		req.MaxJobs = int(maxJobs)
	}

	if cores, ok := params["cores"].(float64); ok {
		req.Cores = int(cores)
	}

	return req, nil
}

// validateOperation validates the build operation and required parameters
func (bf *BuildFunction) validateOperation(req *BuildRequest) error {
	switch req.Operation {
	case "build", "rebuild":
		if req.Package == "" && req.Configuration == "" {
			return fmt.Errorf("package or configuration must be specified for build operations")
		}
	case "troubleshoot":
		if req.ErrorLogs == "" && req.Package == "" {
			return fmt.Errorf("error_logs or package must be specified for troubleshooting")
		}
	case "optimize", "analyze", "check":
		// These operations can work with or without specific packages
	case "clean":
		// Clean operation doesn't require additional parameters
	default:
		return fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	// Validate system architecture if specified
	if req.System != "" {
		validSystems := []string{"x86_64-linux", "aarch64-linux", "x86_64-darwin", "aarch64-darwin"}
		isValid := false
		for _, sys := range validSystems {
			if req.System == sys {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid system architecture: %s", req.System)
		}
	}

	return nil
}

// executeBuildOperation performs the actual build operation using the build agent
func (bf *BuildFunction) executeBuildOperation(ctx context.Context, req *BuildRequest) (*BuildResponse, error) {
	// Prepare build context for the agent
	buildContext := &agent.BuildContext{
		Operation:     req.Operation,
		Package:       req.Package,
		Configuration: req.Configuration,
		ErrorLogs:     req.ErrorLogs,
		BuildOptions:  req.BuildOptions,
		System:        req.System,
		Flake:         req.Flake,
		Impure:        req.Impure,
		KeepGoing:     req.KeepGoing,
		ShowTrace:     req.ShowTrace,
		Verbose:       req.Verbose,
		MaxJobs:       req.MaxJobs,
		Cores:         req.Cores,
	}

	bf.buildAgent.SetBuildContext(buildContext)

	var result string
	var err error

	switch req.Operation {
	case "build", "rebuild":
		result, err = bf.buildAgent.BuildPackage(ctx, req.Package, buildContext)
	case "troubleshoot":
		result, err = bf.buildAgent.TroubleshootBuild(ctx, req.ErrorLogs, buildContext)
	case "optimize":
		result, err = bf.buildAgent.OptimizeBuild(ctx, buildContext)
	case "analyze":
		result, err = bf.buildAgent.AnalyzeBuild(ctx, req.Package, buildContext)
	case "check":
		result, err = bf.buildAgent.CheckBuild(ctx, req.Package, buildContext)
	case "clean":
		result, err = bf.buildAgent.CleanBuild(ctx, buildContext)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	if err != nil {
		return nil, err
	}

	// Parse the agent response into structured output
	return bf.parseAgentResponse(result, req), nil
}

// parseAgentResponse converts the agent's text response into structured BuildResponse
func (bf *BuildFunction) parseAgentResponse(response string, req *BuildRequest) *BuildResponse {
	buildResp := &BuildResponse{
		Status:   "success",
		Solution: response,
	}

	// Extract structured information from the response
	lines := strings.Split(response, "\n")
	var currentSection string
	var commands []string
	var tips []string
	var optimizations []string
	var docs []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect sections
		if strings.Contains(strings.ToLower(line), "command") || strings.Contains(strings.ToLower(line), "run") {
			currentSection = "commands"
			continue
		} else if strings.Contains(strings.ToLower(line), "tip") || strings.Contains(strings.ToLower(line), "suggestion") {
			currentSection = "tips"
			continue
		} else if strings.Contains(strings.ToLower(line), "optim") || strings.Contains(strings.ToLower(line), "performance") {
			currentSection = "optimization"
			continue
		} else if strings.Contains(strings.ToLower(line), "doc") || strings.Contains(strings.ToLower(line), "reference") {
			currentSection = "docs"
			continue
		}

		// Extract commands (lines that start with nix, nixos-rebuild, etc.)
		if strings.HasPrefix(line, "nix ") || strings.HasPrefix(line, "nixos-rebuild") ||
			strings.HasPrefix(line, "nix-build") || strings.HasPrefix(line, "nix-shell") {
			commands = append(commands, line)
		}

		// Extract tips based on current section
		if currentSection == "tips" && (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")) {
			tips = append(tips, strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		} else if currentSection == "optimization" && (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")) {
			optimizations = append(optimizations, strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		} else if currentSection == "docs" && (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")) {
			docs = append(docs, strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		}
	}

	// Extract estimated time if mentioned
	timeRegex := regexp.MustCompile(`(\d+)\s*(minute|hour|second)s?`)
	if matches := timeRegex.FindStringSubmatch(response); len(matches) > 0 {
		buildResp.EstimatedTime = matches[0]
	}

	// Set extracted information
	if len(commands) > 0 {
		buildResp.SuggestedCommands = commands
	}
	if len(tips) > 0 {
		buildResp.TroubleshootingTips = tips
	}
	if len(optimizations) > 0 {
		buildResp.OptimizationTips = optimizations
	}
	if len(docs) > 0 {
		buildResp.DocumentationRefs = docs
	}

	// Add diagnosis details for troubleshooting operations
	if req.Operation == "troubleshoot" || req.Operation == "analyze" {
		buildResp.DiagnosisDetails = bf.extractDiagnosisDetails(response)
	}

	return buildResp
}

// extractDiagnosisDetails extracts diagnostic information from the response
func (bf *BuildFunction) extractDiagnosisDetails(response string) string {
	lines := strings.Split(response, "\n")
	var diagnosis []string

	inDiagnosisSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "diagnos") ||
			strings.Contains(strings.ToLower(line), "analysis") ||
			strings.Contains(strings.ToLower(line), "issue") {
			inDiagnosisSection = true
			continue
		}

		if inDiagnosisSection && line != "" {
			if strings.Contains(strings.ToLower(line), "solution") ||
				strings.Contains(strings.ToLower(line), "command") {
				break
			}
			diagnosis = append(diagnosis, line)
		}
	}

	if len(diagnosis) > 0 {
		return strings.Join(diagnosis, "\n")
	}

	// Fallback: extract first few sentences
	sentences := strings.Split(response, ". ")
	if len(sentences) > 3 {
		return strings.Join(sentences[:3], ". ") + "."
	}

	return response
}
