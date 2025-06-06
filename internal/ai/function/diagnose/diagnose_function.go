package diagnose

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
)

// DiagnoseFunction implements AI function calling for NixOS diagnostics
type DiagnoseFunction struct {
	*functionbase.BaseFunction
	diagnoseAgent *agent.DiagnoseAgent
	logger        *logger.Logger
}

// DiagnoseRequest represents the input parameters for the diagnose function
type DiagnoseRequest struct {
	LogData           string             `json:"log_data,omitempty"`
	ConfigSnippet     string             `json:"config_snippet,omitempty"`
	ErrorMessage      string             `json:"error_message,omitempty"`
	UserDescription   string             `json:"user_description,omitempty"`
	CommandOutput     string             `json:"command_output,omitempty"`
	SystemInfo        *SystemInfoRequest `json:"system_info,omitempty"`
	AnalysisType      string             `json:"analysis_type,omitempty"`
	Severity          string             `json:"severity,omitempty"`
	IncludeSteps      bool               `json:"include_steps"`
	IncludePrevention bool               `json:"include_prevention"`
}

// SystemInfoRequest represents system information for diagnosis
type SystemInfoRequest struct {
	NixVersion    string `json:"nix_version,omitempty"`
	NixOSVersion  string `json:"nixos_version,omitempty"`
	Channel       string `json:"channel,omitempty"`
	Generation    string `json:"generation,omitempty"`
	Architecture  string `json:"architecture,omitempty"`
	IsFlakeSystem bool   `json:"is_flake_system"`
}

// DiagnoseResponse represents the output of the diagnose function
type DiagnoseResponse struct {
	Summary         string                     `json:"summary"`
	RootCause       string                     `json:"root_cause,omitempty"`
	FixSteps        []string                   `json:"fix_steps,omitempty"`
	Verification    string                     `json:"verification,omitempty"`
	Prevention      string                     `json:"prevention,omitempty"`
	Resources       []string                   `json:"resources,omitempty"`
	Severity        string                     `json:"severity"`
	Diagnostics     []nixos.Diagnostic         `json:"diagnostics,omitempty"`
	Recommendations []DiagnosticRecommendation `json:"recommendations,omitempty"`
}

// DiagnosticRecommendation represents a specific recommendation
type DiagnosticRecommendation struct {
	Type        string `json:"type"`   // config, command, package, etc.
	Action      string `json:"action"` // add, remove, modify, run
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	ConfigPath  string `json:"config_path,omitempty"`
	Priority    string `json:"priority"` // high, medium, low
}

// NewDiagnoseFunction creates a new diagnose function
func NewDiagnoseFunction() *DiagnoseFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("user_description", "Description of the problem or issue", false),
		functionbase.StringParam("log_data", "Log output or error logs to analyze", false),
		functionbase.StringParam("config_snippet", "NixOS configuration snippet related to the issue", false),
		functionbase.StringParam("error_message", "Specific error message encountered", false),
		functionbase.StringParam("command_output", "Output from NixOS commands", false),
		functionbase.StringParamWithOptions("analysis_type", "Type of analysis to perform", false,
			[]string{"general", "build", "service", "configuration", "boot", "network", "storage"}, nil, nil),
		functionbase.StringParamWithOptions("severity", "Expected severity level", false,
			[]string{"low", "medium", "high", "critical"}, nil, nil),
		functionbase.BoolParam("include_steps", "Include step-by-step fix instructions", false),
		functionbase.BoolParam("include_prevention", "Include prevention recommendations", false),
		functionbase.ObjectParam("system_info", "System information for context", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"diagnose",
		"Diagnose NixOS issues and provide comprehensive solutions with step-by-step fixes",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Diagnose a service startup issue",
			Parameters: map[string]interface{}{
				"user_description": "My nginx service won't start after rebuilding",
				"log_data":         "systemctl status nginx shows failed to bind to port 80",
				"analysis_type":    "service",
				"include_steps":    true,
			},
			Expected: "Comprehensive diagnosis with service-specific troubleshooting steps",
		},
		{
			Description: "Analyze build error",
			Parameters: map[string]interface{}{
				"error_message":      "builder for '/nix/store/...' failed with exit code 1",
				"command_output":     "nixos-rebuild switch output",
				"analysis_type":      "build",
				"include_prevention": true,
			},
			Expected: "Build error analysis with fix steps and prevention tips",
		},
	}
	baseFunc.SetSchema(schema)

	return &DiagnoseFunction{
		BaseFunction:  baseFunc,
		diagnoseAgent: agent.NewDiagnoseAgent(),
		logger:        logger.NewLogger(),
	}
}

// Execute runs the diagnose function
func (df *DiagnoseFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	df.logger.Debug("Starting diagnose function execution")

	// Parse parameters into structured request
	request, err := df.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Report progress if callback is available
	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    1,
			Total:      5,
			Percentage: 20,
			Message:    "Analyzing input parameters",
			Stage:      "preparation",
		})
	}

	// Build diagnostic context from request
	diagnosticContext := df.buildDiagnosticContext(request)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    2,
			Total:      5,
			Percentage: 40,
			Message:    "Building diagnostic context",
			Stage:      "analysis",
		})
	}

	// Run automated diagnostics if we have log data
	var existingDiagnostics []nixos.Diagnostic
	if request.LogData != "" || request.ErrorMessage != "" {
		existingDiagnostics = df.runAutomatedDiagnostics(request)
		diagnosticContext.ExistingDiagnostics = existingDiagnostics
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    3,
			Total:      5,
			Percentage: 60,
			Message:    "Running automated diagnostics",
			Stage:      "diagnosis",
		})
	}

	// Generate AI-powered diagnosis using the agent
	userInput := df.buildUserInput(request)
	diagnosis, err := df.diagnoseAgent.Query(ctx, userInput, string(roles.RoleDiagnose), diagnosticContext)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to generate diagnosis"), nil
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    4,
			Total:      5,
			Percentage: 80,
			Message:    "Generating AI diagnosis",
			Stage:      "generation",
		})
	}

	// Parse and structure the diagnosis response
	response := df.parseResponse(diagnosis, request, existingDiagnostics)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    5,
			Total:      5,
			Percentage: 100,
			Message:    "Diagnosis complete",
			Stage:      "complete",
		})
	}

	df.logger.Debug("Diagnose function execution completed successfully")
	return functionbase.CreateSuccessResult(response, "Diagnosis completed successfully"), nil
}

// parseRequest converts parameters map to structured request
func (df *DiagnoseFunction) parseRequest(params map[string]interface{}) (*DiagnoseRequest, error) {
	request := &DiagnoseRequest{
		IncludeSteps:      true,      // Default to true
		IncludePrevention: true,      // Default to true
		AnalysisType:      "general", // Default analysis type
	}

	// Parse string parameters
	if val, ok := params["user_description"].(string); ok {
		request.UserDescription = val
	}
	if val, ok := params["log_data"].(string); ok {
		request.LogData = val
	}
	if val, ok := params["config_snippet"].(string); ok {
		request.ConfigSnippet = val
	}
	if val, ok := params["error_message"].(string); ok {
		request.ErrorMessage = val
	}
	if val, ok := params["command_output"].(string); ok {
		request.CommandOutput = val
	}
	if val, ok := params["analysis_type"].(string); ok {
		request.AnalysisType = val
	}
	if val, ok := params["severity"].(string); ok {
		request.Severity = val
	}

	// Parse boolean parameters
	if val, ok := params["include_steps"].(bool); ok {
		request.IncludeSteps = val
	}
	if val, ok := params["include_prevention"].(bool); ok {
		request.IncludePrevention = val
	}

	// Parse system_info object
	if sysInfo, ok := params["system_info"].(map[string]interface{}); ok {
		request.SystemInfo = &SystemInfoRequest{}

		if val, ok := sysInfo["nix_version"].(string); ok {
			request.SystemInfo.NixVersion = val
		}
		if val, ok := sysInfo["nixos_version"].(string); ok {
			request.SystemInfo.NixOSVersion = val
		}
		if val, ok := sysInfo["channel"].(string); ok {
			request.SystemInfo.Channel = val
		}
		if val, ok := sysInfo["generation"].(string); ok {
			request.SystemInfo.Generation = val
		}
		if val, ok := sysInfo["architecture"].(string); ok {
			request.SystemInfo.Architecture = val
		}
		if val, ok := sysInfo["is_flake_system"].(bool); ok {
			request.SystemInfo.IsFlakeSystem = val
		}
	}

	// Validate that at least one input is provided
	if request.UserDescription == "" && request.LogData == "" &&
		request.ConfigSnippet == "" && request.ErrorMessage == "" &&
		request.CommandOutput == "" {
		return nil, fmt.Errorf("at least one input parameter must be provided (user_description, log_data, config_snippet, error_message, or command_output)")
	}

	return request, nil
}

// buildDiagnosticContext creates agent.DiagnosticContext from request
func (df *DiagnoseFunction) buildDiagnosticContext(request *DiagnoseRequest) *agent.DiagnosticContext {
	context := &agent.DiagnosticContext{
		LogData:         request.LogData,
		ConfigSnippet:   request.ConfigSnippet,
		ErrorMessage:    request.ErrorMessage,
		CommandOutput:   request.CommandOutput,
		UserDescription: request.UserDescription,
	}

	// Convert system info if provided
	if request.SystemInfo != nil {
		context.SystemInfo = &agent.SystemInfo{
			NixVersion:    request.SystemInfo.NixVersion,
			NixOSVersion:  request.SystemInfo.NixOSVersion,
			Channel:       request.SystemInfo.Channel,
			Generation:    request.SystemInfo.Generation,
			Architecture:  request.SystemInfo.Architecture,
			IsFlakeSystem: request.SystemInfo.IsFlakeSystem,
		}
	}

	return context
}

// buildUserInput creates a user input string from the request
func (df *DiagnoseFunction) buildUserInput(request *DiagnoseRequest) string {
	var parts []string

	if request.UserDescription != "" {
		parts = append(parts, fmt.Sprintf("Problem: %s", request.UserDescription))
	}

	if request.AnalysisType != "general" {
		parts = append(parts, fmt.Sprintf("Analysis type requested: %s", request.AnalysisType))
	}

	if request.Severity != "" {
		parts = append(parts, fmt.Sprintf("Expected severity: %s", request.Severity))
	}

	var requirements []string
	if request.IncludeSteps {
		requirements = append(requirements, "step-by-step fix instructions")
	}
	if request.IncludePrevention {
		requirements = append(requirements, "prevention recommendations")
	}

	if len(requirements) > 0 {
		parts = append(parts, fmt.Sprintf("Please include: %s", strings.Join(requirements, ", ")))
	}

	if len(parts) == 0 {
		return "Please analyze the provided NixOS diagnostic information and provide a comprehensive diagnosis."
	}

	return strings.Join(parts, ". ")
}

// runAutomatedDiagnostics performs basic automated analysis
func (df *DiagnoseFunction) runAutomatedDiagnostics(request *DiagnoseRequest) []nixos.Diagnostic {
	var diagnostics []nixos.Diagnostic

	// Analyze error patterns
	if request.ErrorMessage != "" {
		if diag := df.analyzeErrorMessage(request.ErrorMessage); diag != nil {
			diagnostics = append(diagnostics, *diag)
		}
	}

	// Analyze log data
	if request.LogData != "" {
		logDiagnostics := df.analyzeLogData(request.LogData)
		diagnostics = append(diagnostics, logDiagnostics...)
	}

	return diagnostics
}

// analyzeErrorMessage performs basic error message analysis
func (df *DiagnoseFunction) analyzeErrorMessage(errorMsg string) *nixos.Diagnostic {
	lowerError := strings.ToLower(errorMsg)

	// Common error patterns
	if strings.Contains(lowerError, "permission denied") {
		return &nixos.Diagnostic{
			ErrorType: "permission",
			Issue:     "Permission denied error detected",
			Severity:  "medium",
			Details:   "This typically indicates insufficient permissions or ownership issues",
		}
	}

	if strings.Contains(lowerError, "no space left") {
		return &nixos.Diagnostic{
			ErrorType: "storage",
			Issue:     "Storage space issue detected",
			Severity:  "high",
			Details:   "Disk space is full or nearly full",
		}
	}

	if strings.Contains(lowerError, "connection refused") || strings.Contains(lowerError, "network") {
		return &nixos.Diagnostic{
			ErrorType: "network",
			Issue:     "Network connectivity issue detected",
			Severity:  "medium",
			Details:   "Network connection or service availability problem",
		}
	}

	if strings.Contains(lowerError, "builder") && strings.Contains(lowerError, "failed") {
		return &nixos.Diagnostic{
			ErrorType: "build",
			Issue:     "Build failure detected",
			Severity:  "high",
			Details:   "Package or system build process failed",
		}
	}

	return nil
}

// analyzeLogData performs basic log analysis
func (df *DiagnoseFunction) analyzeLogData(logData string) []nixos.Diagnostic {
	var diagnostics []nixos.Diagnostic
	lines := strings.Split(logData, "\n")

	for _, line := range lines {
		lowerLine := strings.ToLower(line)

		if strings.Contains(lowerLine, "error") && !strings.Contains(lowerLine, "error:") {
			diagnostics = append(diagnostics, nixos.Diagnostic{
				ErrorType: "error",
				Issue:     "Error found in logs",
				Severity:  "medium",
				Details:   strings.TrimSpace(line),
			})
		}

		if strings.Contains(lowerLine, "failed") {
			diagnostics = append(diagnostics, nixos.Diagnostic{
				ErrorType: "failure",
				Issue:     "Failure found in logs",
				Severity:  "medium",
				Details:   strings.TrimSpace(line),
			})
		}
	}

	return diagnostics
}

// parseResponse structures the AI diagnosis response
func (df *DiagnoseFunction) parseResponse(diagnosis string, request *DiagnoseRequest, diagnostics []nixos.Diagnostic) *DiagnoseResponse {
	response := &DiagnoseResponse{
		Summary:     diagnosis,
		Diagnostics: diagnostics,
		Severity:    request.Severity,
	}

	if response.Severity == "" {
		response.Severity = "medium" // Default severity
	}

	// Try to extract structured information from the diagnosis
	// This is a simple implementation - in practice, you might want more sophisticated parsing
	lines := strings.Split(diagnosis, "\n")
	var currentSection string
	var stepCount int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect sections
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "root cause") || strings.Contains(lowerLine, "cause analysis") {
			currentSection = "root_cause"
			continue
		} else if strings.Contains(lowerLine, "step") && strings.Contains(lowerLine, "fix") {
			currentSection = "fix_steps"
			stepCount = 0
			continue
		} else if strings.Contains(lowerLine, "verification") || strings.Contains(lowerLine, "confirm") {
			currentSection = "verification"
			continue
		} else if strings.Contains(lowerLine, "prevention") || strings.Contains(lowerLine, "avoid") {
			currentSection = "prevention"
			continue
		}

		// Extract content based on current section
		switch currentSection {
		case "root_cause":
			if response.RootCause == "" {
				response.RootCause = line
			} else {
				response.RootCause += " " + line
			}
		case "fix_steps":
			if strings.HasPrefix(line, fmt.Sprintf("%d.", stepCount+1)) ||
				strings.Contains(lowerLine, "step") {
				stepCount++
				response.FixSteps = append(response.FixSteps, line)
			}
		case "verification":
			if response.Verification == "" {
				response.Verification = line
			} else {
				response.Verification += " " + line
			}
		case "prevention":
			if response.Prevention == "" {
				response.Prevention = line
			} else {
				response.Prevention += " " + line
			}
		}
	}

	// Generate recommendations based on analysis type
	response.Recommendations = df.generateRecommendations(request, diagnostics)

	return response
}

// generateRecommendations creates specific recommendations based on the analysis
func (df *DiagnoseFunction) generateRecommendations(request *DiagnoseRequest, diagnostics []nixos.Diagnostic) []DiagnosticRecommendation {
	var recommendations []DiagnosticRecommendation

	// Analysis type specific recommendations
	switch request.AnalysisType {
	case "service":
		recommendations = append(recommendations, DiagnosticRecommendation{
			Type:        "command",
			Action:      "run",
			Description: "Check service status and logs",
			Command:     "systemctl status <service-name> && journalctl -u <service-name>",
			Priority:    "high",
		})
	case "build":
		recommendations = append(recommendations, DiagnosticRecommendation{
			Type:        "command",
			Action:      "run",
			Description: "Clear build cache and retry",
			Command:     "nix-collect-garbage -d && nixos-rebuild switch",
			Priority:    "medium",
		})
	case "configuration":
		recommendations = append(recommendations, DiagnosticRecommendation{
			Type:        "config",
			Action:      "modify",
			Description: "Validate configuration syntax",
			Command:     "nixos-rebuild dry-build",
			Priority:    "high",
		})
	}

	// Add diagnostic-specific recommendations
	for _, diag := range diagnostics {
		switch diag.ErrorType {
		case "permission":
			recommendations = append(recommendations, DiagnosticRecommendation{
				Type:        "command",
				Action:      "run",
				Description: "Check file permissions and ownership",
				Command:     "ls -la <affected-file>",
				Priority:    "high",
			})
		case "storage":
			recommendations = append(recommendations, DiagnosticRecommendation{
				Type:        "command",
				Action:      "run",
				Description: "Check disk space and clean up",
				Command:     "df -h && nix-collect-garbage -d",
				Priority:    "critical",
			})
		}
	}

	return recommendations
}
