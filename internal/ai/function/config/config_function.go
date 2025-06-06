package config

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// ConfigFunction implements AI function calling for configuration management
type ConfigFunction struct {
	*functionbase.BaseFunction
	configAgent *agent.ConfigAgent
	logger      *logger.Logger
}

// ConfigRequest represents the input parameters for the config function
type ConfigRequest struct {
	Operation     string            `json:"operation"`
	ConfigType    string            `json:"config_type,omitempty"`
	ConfigPath    string            `json:"config_path,omitempty"`
	ConfigContent string            `json:"config_content,omitempty"`
	Options       map[string]string `json:"options,omitempty"`
	System        string            `json:"system,omitempty"`
	Hardware      string            `json:"hardware,omitempty"`
	Services      []string          `json:"services,omitempty"`
	Packages      []string          `json:"packages,omitempty"`
	HomeManager   bool              `json:"home_manager,omitempty"`
	Flakes        bool              `json:"flakes,omitempty"`
	Validate      bool              `json:"validate,omitempty"`
	Backup        bool              `json:"backup,omitempty"`
	DryRun        bool              `json:"dry_run,omitempty"`
}

// ConfigResponse represents the output of the config function
type ConfigResponse struct {
	Status            string            `json:"status"`
	ConfigContent     string            `json:"config_content,omitempty"`
	ValidationResult  string            `json:"validation_result,omitempty"`
	Recommendations   []string          `json:"recommendations,omitempty"`
	SuggestedCommands []string          `json:"suggested_commands,omitempty"`
	WarningMessages   []string          `json:"warning_messages,omitempty"`
	OptimizationTips  []string          `json:"optimization_tips,omitempty"`
	DocumentationRefs []string          `json:"documentation_refs,omitempty"`
	ConfigDiff        string            `json:"config_diff,omitempty"`
	BackupPath        string            `json:"backup_path,omitempty"`
	AppliedOptions    map[string]string `json:"applied_options,omitempty"`
}

// NewConfigFunction creates a new config function
func NewConfigFunction() *ConfigFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Type of configuration operation to perform", true,
			[]string{"generate", "validate", "optimize", "analyze", "update", "backup", "restore", "migrate", "diff"}, nil, nil),
		functionbase.StringParamWithOptions("config_type", "Type of configuration to work with", false,
			[]string{"nixos", "home-manager", "flake", "shell", "service", "hardware", "desktop"}, nil, nil),
		functionbase.StringParam("config_path", "Path to the configuration file or directory", false),
		functionbase.StringParam("config_content", "Configuration content to process or validate", false),
		{
			Name:        "options",
			Type:        "object",
			Description: "Configuration options as key-value pairs",
			Required:    false,
			Properties: map[string]functionbase.FunctionParameter{
				"hostname": functionbase.StringParam("hostname", "System hostname", false),
				"timezone": functionbase.StringParam("timezone", "System timezone", false),
				"locale":   functionbase.StringParam("locale", "System locale", false),
				"desktop":  functionbase.StringParam("desktop", "Desktop environment", false),
			},
		},
		functionbase.StringParamWithOptions("system", "Target system architecture", false,
			[]string{"x86_64-linux", "aarch64-linux", "x86_64-darwin", "aarch64-darwin"}, nil, nil),
		functionbase.StringParam("hardware", "Hardware description or configuration", false),
		{
			Name:        "services",
			Type:        "array",
			Description: "System services to configure",
			Required:    false,
			Items: &functionbase.FunctionParameter{
				Type: "string",
			},
		},
		{
			Name:        "packages",
			Type:        "array",
			Description: "Packages to include in configuration",
			Required:    false,
			Items: &functionbase.FunctionParameter{
				Type: "string",
			},
		},
		functionbase.BoolParam("home_manager", "Whether to include Home Manager configuration", false),
		functionbase.BoolParam("flakes", "Whether to use flakes for configuration", false),
		functionbase.BoolParam("validate", "Whether to validate the configuration", false),
		functionbase.BoolParam("backup", "Whether to create a backup before changes", false),
		functionbase.BoolParam("dry_run", "Whether to perform a dry run without applying changes", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"config",
		"Handle NixOS and Home Manager configuration management including generation, validation, optimization, and migration of configuration files",
		parameters,
	)

	return &ConfigFunction{
		BaseFunction: baseFunc,
		configAgent:  agent.NewConfigAgent(nil), // Will be set when executing
		logger:       logger.NewLogger(),
	}
}

// Execute performs the configuration operation
func (cf *ConfigFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	// Validate parameters
	if err := cf.ValidateParameters(params); err != nil {
		return functionbase.ErrorResult(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Parse parameters into request struct
	req, err := cf.parseRequest(params)
	if err != nil {
		return functionbase.ErrorResult(fmt.Sprintf("Failed to parse request: %v", err)), nil
	}

	// Validate operation
	if err := cf.validateOperation(req); err != nil {
		return functionbase.ErrorResult(err.Error()), nil
	}

	cf.logger.Info(fmt.Sprintf("Executing config function with operation: %s", req.Operation))

	// Initialize config agent with provider if available
	if options != nil && options.Provider != nil {
		cf.configAgent = agent.NewConfigAgent(options.Provider)
	}

	// Execute the configuration operation
	response, err := cf.executeConfigOperation(ctx, req)
	if err != nil {
		return functionbase.ErrorResult(fmt.Sprintf("Configuration operation failed: %v", err)), nil
	}

	return functionbase.SuccessResult(response), nil
}

// parseRequest converts the parameters map to a structured request
func (cf *ConfigFunction) parseRequest(params map[string]interface{}) (*ConfigRequest, error) {
	req := &ConfigRequest{}

	// Required parameters
	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	} else {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	// Optional parameters
	if configType, ok := params["config_type"].(string); ok {
		req.ConfigType = configType
	}

	if configPath, ok := params["config_path"].(string); ok {
		req.ConfigPath = configPath
	}

	if configContent, ok := params["config_content"].(string); ok {
		req.ConfigContent = configContent
	}

	if options, ok := params["options"].(map[string]interface{}); ok {
		req.Options = make(map[string]string)
		for k, v := range options {
			if strVal, ok := v.(string); ok {
				req.Options[k] = strVal
			}
		}
	}

	if system, ok := params["system"].(string); ok {
		req.System = system
	}

	if hardware, ok := params["hardware"].(string); ok {
		req.Hardware = hardware
	}

	if services, ok := params["services"].([]interface{}); ok {
		for _, service := range services {
			if serviceStr, ok := service.(string); ok {
				req.Services = append(req.Services, serviceStr)
			}
		}
	}

	if packages, ok := params["packages"].([]interface{}); ok {
		for _, pkg := range packages {
			if pkgStr, ok := pkg.(string); ok {
				req.Packages = append(req.Packages, pkgStr)
			}
		}
	}

	if homeManager, ok := params["home_manager"].(bool); ok {
		req.HomeManager = homeManager
	}

	if flakes, ok := params["flakes"].(bool); ok {
		req.Flakes = flakes
	}

	if validate, ok := params["validate"].(bool); ok {
		req.Validate = validate
	}

	if backup, ok := params["backup"].(bool); ok {
		req.Backup = backup
	}

	if dryRun, ok := params["dry_run"].(bool); ok {
		req.DryRun = dryRun
	}

	return req, nil
}

// validateOperation validates the configuration operation and required parameters
func (cf *ConfigFunction) validateOperation(req *ConfigRequest) error {
	switch req.Operation {
	case "generate":
		if req.ConfigType == "" {
			return fmt.Errorf("config_type must be specified for generate operations")
		}
	case "validate", "analyze", "optimize":
		if req.ConfigPath == "" && req.ConfigContent == "" {
			return fmt.Errorf("config_path or config_content must be specified for %s operations", req.Operation)
		}
	case "update", "backup", "restore":
		if req.ConfigPath == "" {
			return fmt.Errorf("config_path must be specified for %s operations", req.Operation)
		}
	case "diff":
		if req.ConfigPath == "" && req.ConfigContent == "" {
			return fmt.Errorf("config_path or config_content must be specified for diff operations")
		}
	case "migrate":
		if req.ConfigPath == "" {
			return fmt.Errorf("config_path must be specified for migration operations")
		}
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

// executeConfigOperation performs the actual configuration operation using the config agent
func (cf *ConfigFunction) executeConfigOperation(ctx context.Context, req *ConfigRequest) (*ConfigResponse, error) {
	// Prepare config context for the agent
	configContext := &agent.ConfigContext{
		Operation:     req.Operation,
		ConfigType:    req.ConfigType,
		ConfigPath:    req.ConfigPath,
		ConfigContent: req.ConfigContent,
		Options:       req.Options,
		System:        req.System,
		Hardware:      req.Hardware,
		Services:      req.Services,
		Packages:      req.Packages,
		HomeManager:   req.HomeManager,
		Flakes:        req.Flakes,
		Validate:      req.Validate,
		Backup:        req.Backup,
		DryRun:        req.DryRun,
	}

	cf.configAgent.SetConfigContext(configContext)

	var result string
	var err error

	switch req.Operation {
	case "generate":
		result, err = cf.configAgent.GenerateConfiguration(ctx, req.ConfigType, configContext)
	case "validate":
		result, err = cf.configAgent.ValidateConfiguration(ctx, configContext)
	case "optimize":
		result, err = cf.configAgent.OptimizeConfiguration(ctx, configContext)
	case "analyze":
		result, err = cf.configAgent.AnalyzeConfiguration(ctx, configContext)
	case "update":
		result, err = cf.configAgent.UpdateConfiguration(ctx, configContext)
	case "backup":
		result, err = cf.configAgent.BackupConfiguration(ctx, configContext)
	case "restore":
		result, err = cf.configAgent.RestoreConfiguration(ctx, configContext)
	case "migrate":
		result, err = cf.configAgent.MigrateConfiguration(ctx, configContext)
	case "diff":
		result, err = cf.configAgent.DiffConfiguration(ctx, configContext)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	if err != nil {
		return nil, err
	}

	// Parse the agent response into structured output
	return cf.parseAgentResponse(result, req), nil
}

// parseAgentResponse converts the agent's text response into structured ConfigResponse
func (cf *ConfigFunction) parseAgentResponse(response string, req *ConfigRequest) *ConfigResponse {
	configResp := &ConfigResponse{
		Status: "success",
	}

	// Extract structured information from the response
	lines := strings.Split(response, "\n")
	var currentSection string
	var commands []string
	var recommendations []string
	var warnings []string
	var optimizations []string
	var docs []string
	var configContent strings.Builder
	var inConfigBlock bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Detect configuration content blocks
		if strings.Contains(line, "```") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "let") {
			inConfigBlock = !inConfigBlock
			if inConfigBlock && (strings.Contains(line, "nix") || strings.Contains(line, "nixos")) {
				continue
			}
			if !inConfigBlock {
				continue
			}
		}

		if inConfigBlock {
			configContent.WriteString(line + "\n")
			continue
		}

		// Detect sections
		lowerLine := strings.ToLower(trimmed)
		if strings.Contains(lowerLine, "command") || strings.Contains(lowerLine, "run") {
			currentSection = "commands"
			continue
		} else if strings.Contains(lowerLine, "recommend") || strings.Contains(lowerLine, "suggest") {
			currentSection = "recommendations"
			continue
		} else if strings.Contains(lowerLine, "warn") || strings.Contains(lowerLine, "caution") {
			currentSection = "warnings"
			continue
		} else if strings.Contains(lowerLine, "optim") || strings.Contains(lowerLine, "performance") {
			currentSection = "optimization"
			continue
		} else if strings.Contains(lowerLine, "doc") || strings.Contains(lowerLine, "reference") {
			currentSection = "docs"
			continue
		}

		// Extract commands (lines that start with nix, nixos-rebuild, etc.)
		if strings.HasPrefix(trimmed, "nix ") || strings.HasPrefix(trimmed, "nixos-rebuild") ||
			strings.HasPrefix(trimmed, "home-manager") || strings.HasPrefix(trimmed, "sudo nixos-rebuild") {
			commands = append(commands, trimmed)
		}

		// Extract content based on current section
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			switch currentSection {
			case "recommendations":
				recommendations = append(recommendations, content)
			case "warnings":
				warnings = append(warnings, content)
			case "optimization":
				optimizations = append(optimizations, content)
			case "docs":
				docs = append(docs, content)
			}
		}
	}

	// Set the main response content
	if req.Operation == "generate" {
		configResp.ConfigContent = configContent.String()
		if configResp.ConfigContent == "" {
			configResp.ConfigContent = response
		}
	} else if req.Operation == "validate" {
		configResp.ValidationResult = cf.extractValidationResult(response)
	} else if req.Operation == "diff" {
		configResp.ConfigDiff = cf.extractDiffResult(response)
	} else if req.Operation == "backup" {
		configResp.BackupPath = cf.extractBackupPath(response)
	}

	// Set extracted information
	if len(commands) > 0 {
		configResp.SuggestedCommands = commands
	}
	if len(recommendations) > 0 {
		configResp.Recommendations = recommendations
	}
	if len(warnings) > 0 {
		configResp.WarningMessages = warnings
	}
	if len(optimizations) > 0 {
		configResp.OptimizationTips = optimizations
	}
	if len(docs) > 0 {
		configResp.DocumentationRefs = docs
	}

	// Set applied options
	if req.Options != nil && len(req.Options) > 0 {
		configResp.AppliedOptions = req.Options
	}

	return configResp
}

// extractValidationResult extracts validation information from the response
func (cf *ConfigFunction) extractValidationResult(response string) string {
	lines := strings.Split(response, "\n")
	var validationLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "valid") || strings.Contains(lowerLine, "error") ||
			strings.Contains(lowerLine, "warning") || strings.Contains(lowerLine, "issue") {
			validationLines = append(validationLines, line)
		}
	}

	if len(validationLines) > 0 {
		return strings.Join(validationLines, "\n")
	}

	return "Configuration appears to be valid"
}

// extractDiffResult extracts diff information from the response
func (cf *ConfigFunction) extractDiffResult(response string) string {
	lines := strings.Split(response, "\n")
	var diffLines []string
	inDiffBlock := false

	for _, line := range lines {
		if strings.Contains(line, "diff") || strings.Contains(line, "---") || strings.Contains(line, "+++") {
			inDiffBlock = true
		}
		if inDiffBlock && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "@")) {
			diffLines = append(diffLines, line)
		}
		if inDiffBlock && strings.TrimSpace(line) == "" && len(diffLines) > 0 {
			break
		}
	}

	if len(diffLines) > 0 {
		return strings.Join(diffLines, "\n")
	}

	return response
}

// extractBackupPath extracts backup path information from the response
func (cf *ConfigFunction) extractBackupPath(response string) string {
	// Look for file paths in the response
	pathRegex := regexp.MustCompile(`(/[^/\s]+)+\.bak|backup.*?(/[^/\s]+)+|saved.*?(/[^/\s]+)+`)
	matches := pathRegex.FindStringSubmatch(response)
	if len(matches) > 0 {
		for _, match := range matches {
			if strings.Contains(match, "/") {
				return match
			}
		}
	}
	return ""
}
