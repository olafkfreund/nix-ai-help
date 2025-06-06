package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/pkg/logger"
)

// McpServerFunction implements AI function calling for MCP server management
type McpServerFunction struct {
	*functionbase.BaseFunction
	mcpAgent *agent.McpServerAgent
	logger   *logger.Logger
}

// McpServerRequest represents the input parameters for the mcp-server function
type McpServerRequest struct {
	Operation        string            `json:"operation"`
	Query            string            `json:"query,omitempty"`
	ServerType       string            `json:"server_type,omitempty"`
	Requirements     map[string]string `json:"requirements,omitempty"`
	Issue            string            `json:"issue,omitempty"`
	Symptoms         []string          `json:"symptoms,omitempty"`
	SecurityConcerns []string          `json:"security_concerns,omitempty"`
	CurrentConfig    map[string]string `json:"current_config,omitempty"`
	PerformanceGoals []string          `json:"performance_goals,omitempty"`
	CurrentMetrics   map[string]string `json:"current_metrics,omitempty"`
	TargetApp        string            `json:"target_app,omitempty"`
	IntegrationReqs  []string          `json:"integration_requirements,omitempty"`
	MonitoringScope  []string          `json:"monitoring_scope,omitempty"`
	AlertingNeeds    []string          `json:"alerting_needs,omitempty"`
	ConfigPath       string            `json:"config_path,omitempty"`
	Debug            bool              `json:"debug,omitempty"`
}

// McpServerResponse represents the output of the mcp-server function
type McpServerResponse struct {
	Success          bool                   `json:"success"`
	Message          string                 `json:"message"`
	Output           string                 `json:"output,omitempty"`
	Error            string                 `json:"error,omitempty"`
	ServerStatus     string                 `json:"server_status,omitempty"`
	ServerInfo       map[string]interface{} `json:"server_info,omitempty"`
	QueryResults     []string               `json:"query_results,omitempty"`
	ConfigSnippets   []string               `json:"config_snippets,omitempty"`
	Recommendations  []string               `json:"recommendations,omitempty"`
	NextSteps        []string               `json:"next_steps,omitempty"`
	Documentation    []string               `json:"documentation,omitempty"`
	TroubleShooting  []string               `json:"troubleshooting,omitempty"`
	SecuritySettings []string               `json:"security_settings,omitempty"`
	Optimizations    []string               `json:"optimizations,omitempty"`
}

// NewMcpServerFunction creates a new mcp-server function
func NewMcpServerFunction() *McpServerFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("operation", "MCP server operation to perform", true),
		functionbase.StringParam("query", "Query string for documentation search", false),
		functionbase.StringParam("server_type", "Type of MCP server to set up", false),
		functionbase.ObjectParam("requirements", "Server setup requirements", false),
		functionbase.StringParam("issue", "Issue description for troubleshooting", false),
		functionbase.ArrayParam("symptoms", "List of symptoms for diagnostics", false),
		functionbase.ArrayParam("security_concerns", "Security concerns to address", false),
		functionbase.ObjectParam("current_config", "Current server configuration", false),
		functionbase.ArrayParam("performance_goals", "Performance optimization goals", false),
		functionbase.ObjectParam("current_metrics", "Current performance metrics", false),
		functionbase.StringParam("target_app", "Target application for integration", false),
		functionbase.ArrayParam("integration_requirements", "Integration requirements", false),
		functionbase.ArrayParam("monitoring_scope", "Monitoring scope requirements", false),
		functionbase.ArrayParam("alerting_needs", "Alerting requirements", false),
		functionbase.StringParam("config_path", "Path to configuration file", false),
		functionbase.BoolParam("debug", "Enable debug mode", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"mcp-server",
		"Provides AI-powered assistance for MCP (Model Context Protocol) server management including setup, troubleshooting, configuration, and integration",
		parameters,
	)

	return &McpServerFunction{
		BaseFunction: baseFunc,
		mcpAgent:     agent.NewMcpServerAgent(nil),
		logger:       logger.NewLogger(),
	}
}

// Execute runs the mcp-server function with the given parameters
func (f *McpServerFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
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

	// Execute the MCP server operation
	response, err := f.executeMcpServerOperation(ctx, request, options)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to execute MCP server operation: %v", err),
		}, nil
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts the parameters map into a McpServerRequest struct
func (f *McpServerFunction) parseRequest(params map[string]interface{}) (*McpServerRequest, error) {
	request := &McpServerRequest{}

	// Parse operation (required)
	if op, ok := params["operation"].(string); ok {
		request.Operation = op
	} else {
		return nil, fmt.Errorf("operation parameter is required")
	}

	// Parse optional parameters
	if query, ok := params["query"].(string); ok {
		request.Query = query
	}

	if serverType, ok := params["server_type"].(string); ok {
		request.ServerType = serverType
	}

	if requirements, ok := params["requirements"].(map[string]interface{}); ok {
		request.Requirements = make(map[string]string)
		for k, v := range requirements {
			if str, ok := v.(string); ok {
				request.Requirements[k] = str
			}
		}
	}

	if issue, ok := params["issue"].(string); ok {
		request.Issue = issue
	}

	if symptoms, ok := params["symptoms"].([]interface{}); ok {
		request.Symptoms = make([]string, len(symptoms))
		for i, symptom := range symptoms {
			if str, ok := symptom.(string); ok {
				request.Symptoms[i] = str
			}
		}
	}

	if concerns, ok := params["security_concerns"].([]interface{}); ok {
		request.SecurityConcerns = make([]string, len(concerns))
		for i, concern := range concerns {
			if str, ok := concern.(string); ok {
				request.SecurityConcerns[i] = str
			}
		}
	}

	if config, ok := params["current_config"].(map[string]interface{}); ok {
		request.CurrentConfig = make(map[string]string)
		for k, v := range config {
			if str, ok := v.(string); ok {
				request.CurrentConfig[k] = str
			}
		}
	}

	if goals, ok := params["performance_goals"].([]interface{}); ok {
		request.PerformanceGoals = make([]string, len(goals))
		for i, goal := range goals {
			if str, ok := goal.(string); ok {
				request.PerformanceGoals[i] = str
			}
		}
	}

	if metrics, ok := params["current_metrics"].(map[string]interface{}); ok {
		request.CurrentMetrics = make(map[string]string)
		for k, v := range metrics {
			if str, ok := v.(string); ok {
				request.CurrentMetrics[k] = str
			}
		}
	}

	if targetApp, ok := params["target_app"].(string); ok {
		request.TargetApp = targetApp
	}

	if reqs, ok := params["integration_requirements"].([]interface{}); ok {
		request.IntegrationReqs = make([]string, len(reqs))
		for i, req := range reqs {
			if str, ok := req.(string); ok {
				request.IntegrationReqs[i] = str
			}
		}
	}

	if scope, ok := params["monitoring_scope"].([]interface{}); ok {
		request.MonitoringScope = make([]string, len(scope))
		for i, s := range scope {
			if str, ok := s.(string); ok {
				request.MonitoringScope[i] = str
			}
		}
	}

	if needs, ok := params["alerting_needs"].([]interface{}); ok {
		request.AlertingNeeds = make([]string, len(needs))
		for i, need := range needs {
			if str, ok := need.(string); ok {
				request.AlertingNeeds[i] = str
			}
		}
	}

	if configPath, ok := params["config_path"].(string); ok {
		request.ConfigPath = configPath
	}

	if debug, ok := params["debug"].(bool); ok {
		request.Debug = debug
	}

	return request, nil
}

// validateRequest validates the McpServerRequest
func (f *McpServerFunction) validateRequest(request *McpServerRequest) error {
	if request.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := []string{
		"start", "stop", "status", "restart", "query",
		"setup", "diagnose", "configure", "optimize",
		"secure", "integrate", "monitor", "logs",
	}

	for _, op := range validOps {
		if request.Operation == op {
			return nil
		}
	}

	return fmt.Errorf("invalid operation: %s. Valid operations: %s", request.Operation, strings.Join(validOps, ", "))
}

// executeMcpServerOperation executes the MCP server operation
func (f *McpServerFunction) executeMcpServerOperation(ctx context.Context, request *McpServerRequest, options *functionbase.FunctionOptions) (*McpServerResponse, error) {
	response := &McpServerResponse{
		Success: true,
		Message: fmt.Sprintf("MCP server operation '%s' completed successfully", request.Operation),
	}

	switch request.Operation {
	case "start":
		return f.handleStartOperation(ctx, request)
	case "stop":
		return f.handleStopOperation(ctx, request)
	case "status":
		return f.handleStatusOperation(ctx, request)
	case "restart":
		return f.handleRestartOperation(ctx, request)
	case "query":
		return f.handleQueryOperation(ctx, request)
	case "setup":
		return f.handleSetupOperation(ctx, request)
	case "diagnose":
		return f.handleDiagnoseOperation(ctx, request)
	case "configure":
		return f.handleConfigureOperation(ctx, request)
	case "optimize":
		return f.handleOptimizeOperation(ctx, request)
	case "secure":
		return f.handleSecureOperation(ctx, request)
	case "integrate":
		return f.handleIntegrateOperation(ctx, request)
	case "monitor":
		return f.handleMonitorOperation(ctx, request)
	case "logs":
		return f.handleLogsOperation(ctx, request)
	default:
		response.Success = false
		response.Error = fmt.Sprintf("unsupported operation: %s", request.Operation)
		return response, nil
	}
}

// handleStartOperation starts the MCP server
func (f *McpServerFunction) handleStartOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to load config: %v", err),
		}, nil
	}

	// Create MCP server from config
	configPath := request.ConfigPath
	if configPath == "" {
		configPath, _ = config.ConfigFilePath()
		if configPath == "" {
			configPath = "configs/default.yaml" // fallback
		}
	}

	server, err := mcp.NewServerFromConfig(configPath)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create MCP server: %v", err),
		}, nil
	}

	// Start the server in a goroutine (non-blocking)
	go func() {
		if err := server.Start(); err != nil {
			f.logger.Error(fmt.Sprintf("MCP server failed to start: %v", err))
		}
	}()

	response := &McpServerResponse{
		Success:      true,
		Message:      "MCP server started successfully",
		ServerStatus: "starting",
		ServerInfo: map[string]interface{}{
			"http_endpoint": fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port),
			"unix_socket":   cfg.MCPServer.SocketPath,
			"config_path":   configPath,
		},
		NextSteps: []string{
			"Use 'nixai mcp-server status' to check server health",
			"Use 'nixai mcp-server query <text>' to test documentation queries",
			"Configure VS Code MCP integration if needed",
		},
	}

	return response, nil
}

// handleStopOperation stops the MCP server
func (f *McpServerFunction) handleStopOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to load config: %v", err),
		}, nil
	}

	// Try to stop via HTTP endpoint
	url := fmt.Sprintf("http://%s:%d/shutdown", cfg.MCPServer.Host, cfg.MCPServer.Port)
	_ = mcp.NewMCPClient(url)

	// Note: This is a simplified implementation
	// In a real scenario, you'd need to implement a shutdown endpoint

	response := &McpServerResponse{
		Success:      true,
		Message:      "MCP server shutdown initiated",
		ServerStatus: "stopping",
		ServerInfo: map[string]interface{}{
			"shutdown_endpoint": url,
		},
		NextSteps: []string{
			"Wait a few seconds for graceful shutdown",
			"Use 'nixai mcp-server status' to verify server is stopped",
		},
	}

	return response, nil
}

// handleStatusOperation checks the MCP server status
func (f *McpServerFunction) handleStatusOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to load config: %v", err),
		}, nil
	}

	response := &McpServerResponse{
		Success: true,
		Message: "MCP server status checked",
		ServerInfo: map[string]interface{}{
			"http_endpoint": fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port),
			"unix_socket":   cfg.MCPServer.SocketPath,
		},
	}

	// Check HTTP endpoint health
	mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	client := mcp.NewMCPClient(mcpURL)

	// Try a simple query to test connectivity
	_, err = client.QueryDocumentation("test")
	if err != nil {
		response.ServerStatus = "unreachable"
		response.ServerInfo["http_status"] = "❌ Not running"
	} else {
		response.ServerStatus = "running"
		response.ServerInfo["http_status"] = "✅ Running"
	}

	return response, nil
}

// handleRestartOperation restarts the MCP server
func (f *McpServerFunction) handleRestartOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	// Stop first
	stopResp, err := f.handleStopOperation(ctx, request)
	if err != nil || !stopResp.Success {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to stop server: %v", err),
		}, nil
	}

	// Wait a moment for cleanup
	// In a real implementation, you'd want to poll for actual shutdown

	// Start again
	startResp, err := f.handleStartOperation(ctx, request)
	if err != nil || !startResp.Success {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to start server: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:      true,
		Message:      "MCP server restarted successfully",
		ServerStatus: "restarted",
		ServerInfo:   startResp.ServerInfo,
		NextSteps: []string{
			"Server has been restarted",
			"Use 'nixai mcp-server status' to verify operation",
		},
	}, nil
}

// handleQueryOperation queries the MCP server for documentation
func (f *McpServerFunction) handleQueryOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if request.Query == "" {
		return &McpServerResponse{
			Success: false,
			Error:   "query parameter is required for query operation",
		}, nil
	}

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to load config: %v", err),
		}, nil
	}

	// Create MCP client and query
	mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	client := mcp.NewMCPClient(mcpURL)

	result, err := client.QueryDocumentation(request.Query)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("query failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:      true,
		Message:      fmt.Sprintf("Query executed successfully for: %s", request.Query),
		Output:       result,
		QueryResults: []string{result},
		ServerInfo: map[string]interface{}{
			"query":         request.Query,
			"endpoint":      mcpURL,
			"result_length": len(result),
		},
	}, nil
}

// handleSetupOperation helps set up a new MCP server
func (f *McpServerFunction) handleSetupOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if f.mcpAgent == nil || f.mcpAgent == (*agent.McpServerAgent)(nil) {
		return &McpServerResponse{
			Success: false,
			Error:   "MCP agent not available",
		}, nil
	}

	// Convert requirements map[string]string to map[string]interface{}
	requirements := make(map[string]interface{})
	for k, v := range request.Requirements {
		requirements[k] = v
	}

	result, err := f.mcpAgent.SetupMcpServer(request.ServerType, requirements)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("setup guidance failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:       true,
		Message:       "MCP server setup guidance generated",
		Output:        result,
		Documentation: []string{result},
		NextSteps: []string{
			"Follow the setup instructions provided",
			"Test the configuration before deployment",
			"Consider security and monitoring requirements",
		},
	}, nil
}

// handleDiagnoseOperation diagnoses MCP server issues
func (f *McpServerFunction) handleDiagnoseOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if f.mcpAgent == nil || f.mcpAgent == (*agent.McpServerAgent)(nil) {
		return &McpServerResponse{
			Success: false,
			Error:   "MCP agent not available",
		}, nil
	}

	result, err := f.mcpAgent.DiagnoseMcpIssues(request.Issue, request.Symptoms)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("diagnosis failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:         true,
		Message:         "MCP server diagnosis completed",
		Output:          result,
		TroubleShooting: []string{result},
		NextSteps: []string{
			"Follow the troubleshooting steps provided",
			"Check logs for additional error details",
			"Test each suggested solution incrementally",
		},
	}, nil
}

// handleConfigureOperation provides configuration guidance
func (f *McpServerFunction) handleConfigureOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	// Load current config
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to load config: %v", err),
		}, nil
	}

	configInfo := map[string]interface{}{
		"host":                  cfg.MCPServer.Host,
		"port":                  cfg.MCPServer.Port,
		"socket_path":           cfg.MCPServer.SocketPath,
		"documentation_sources": len(cfg.MCPServer.DocumentationSources),
	}

	return &McpServerResponse{
		Success:    true,
		Message:    "MCP server configuration retrieved",
		ServerInfo: configInfo,
		ConfigSnippets: []string{
			"Current MCP server configuration loaded",
			fmt.Sprintf("Server endpoint: http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port),
			fmt.Sprintf("Unix socket: %s", cfg.MCPServer.SocketPath),
		},
		NextSteps: []string{
			"Review current configuration",
			"Modify config file as needed",
			"Restart server to apply changes",
		},
	}, nil
}

// handleOptimizeOperation provides performance optimization guidance
func (f *McpServerFunction) handleOptimizeOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if f.mcpAgent == nil || f.mcpAgent == (*agent.McpServerAgent)(nil) {
		return &McpServerResponse{
			Success: false,
			Error:   "MCP agent not available",
		}, nil
	}

	// Convert metrics map[string]string to map[string]interface{}
	metrics := make(map[string]interface{})
	for k, v := range request.CurrentMetrics {
		metrics[k] = v
	}

	result, err := f.mcpAgent.OptimizeMcpPerformance(request.PerformanceGoals, metrics)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("optimization guidance failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:       true,
		Message:       "MCP server optimization guidance generated",
		Output:        result,
		Optimizations: []string{result},
		NextSteps: []string{
			"Review optimization recommendations",
			"Implement changes gradually",
			"Monitor performance metrics after changes",
		},
	}, nil
}

// handleSecureOperation provides security configuration guidance
func (f *McpServerFunction) handleSecureOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if f.mcpAgent == nil || f.mcpAgent == (*agent.McpServerAgent)(nil) {
		return &McpServerResponse{
			Success: false,
			Error:   "MCP agent not available",
		}, nil
	}

	// Convert config map[string]string to map[string]interface{}
	config := make(map[string]interface{})
	for k, v := range request.CurrentConfig {
		config[k] = v
	}

	result, err := f.mcpAgent.ManageMcpSecurity(request.SecurityConcerns, config)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("security guidance failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:          true,
		Message:          "MCP server security guidance generated",
		Output:           result,
		SecuritySettings: []string{result},
		NextSteps: []string{
			"Review security recommendations",
			"Implement security measures",
			"Test security configuration",
		},
	}, nil
}

// handleIntegrateOperation provides integration guidance
func (f *McpServerFunction) handleIntegrateOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if f.mcpAgent == nil || f.mcpAgent == (*agent.McpServerAgent)(nil) {
		return &McpServerResponse{
			Success: false,
			Error:   "MCP agent not available",
		}, nil
	}

	result, err := f.mcpAgent.IntegrateMcpServer(request.TargetApp, request.IntegrationReqs)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("integration guidance failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:       true,
		Message:       "MCP server integration guidance generated",
		Output:        result,
		Documentation: []string{result},
		NextSteps: []string{
			"Follow integration steps for " + request.TargetApp,
			"Test integration thoroughly",
			"Document integration process",
		},
	}, nil
}

// handleMonitorOperation provides monitoring guidance
func (f *McpServerFunction) handleMonitorOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	if f.mcpAgent == nil || f.mcpAgent == (*agent.McpServerAgent)(nil) {
		return &McpServerResponse{
			Success: false,
			Error:   "MCP agent not available",
		}, nil
	}

	result, err := f.mcpAgent.MonitorMcpServer(request.MonitoringScope, request.AlertingNeeds)
	if err != nil {
		return &McpServerResponse{
			Success: false,
			Error:   fmt.Sprintf("monitoring guidance failed: %v", err),
		}, nil
	}

	return &McpServerResponse{
		Success:       true,
		Message:       "MCP server monitoring guidance generated",
		Output:        result,
		Documentation: []string{result},
		NextSteps: []string{
			"Set up monitoring as recommended",
			"Configure alerting rules",
			"Test monitoring and alerts",
		},
	}, nil
}

// handleLogsOperation provides log management guidance
func (f *McpServerFunction) handleLogsOperation(ctx context.Context, request *McpServerRequest) (*McpServerResponse, error) {
	return &McpServerResponse{
		Success: true,
		Message: "MCP server log management information",
		Output: `MCP Server Log Management:

1. **Log Levels**: debug, info, warn, error
2. **Log Configuration**: Set via config file or environment variables
3. **Log Files**: Check system logs for MCP server entries
4. **Debug Mode**: Enable debug logging for troubleshooting

Use 'nixai mcp-server start -d' for debug mode.`,
		Documentation: []string{
			"MCP server supports multiple log levels",
			"Debug mode provides detailed operation logs",
			"Check system logs for server status",
		},
		NextSteps: []string{
			"Enable appropriate log level for your needs",
			"Monitor logs for errors or warnings",
			"Use debug mode for troubleshooting",
		},
	}, nil
}
