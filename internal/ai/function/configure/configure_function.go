package configure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// ConfigureFunction handles NixOS configuration management operations
type ConfigureFunction struct {
	*functionbase.BaseFunction
	agent  *agent.ConfigAgent
	logger *logger.Logger
}

// ConfigureRequest represents the input parameters for the configure function
type ConfigureRequest struct {
	Context    string            `json:"context"`
	Operation  string            `json:"operation,omitempty"`
	ConfigType string            `json:"config_type,omitempty"`
	ConfigPath string            `json:"config_path,omitempty"`
	Module     string            `json:"module,omitempty"`
	Option     string            `json:"option,omitempty"`
	Value      interface{}       `json:"value,omitempty"`
	DryRun     bool              `json:"dry_run,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Validate   bool              `json:"validate,omitempty"`
	Format     string            `json:"format,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
}

// ConfigureResponse represents the output of the configure function
type ConfigureResponse struct {
	Context          string            `json:"context"`
	Status           string            `json:"status"`
	Changes          []ConfigChange    `json:"changes,omitempty"`
	ValidationResult *ValidationResult `json:"validation_result,omitempty"`
	BackupPath       string            `json:"backup_path,omitempty"`
	PreviewContent   string            `json:"preview_content,omitempty"`
	ErrorMessage     string            `json:"error_message,omitempty"`
	ExecutionTime    time.Duration     `json:"execution_time,omitempty"`
}

// ConfigChange represents a single configuration change
type ConfigChange struct {
	Path      string      `json:"path"`
	Option    string      `json:"option"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value"`
	Operation string      `json:"operation"` // add, update, remove
	Status    string      `json:"status"`    // success, failed, pending
}

// ValidationResult represents configuration validation results
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
	Summary  string              `json:"summary"`
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Path       string `json:"path"`
	Option     string `json:"option"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a configuration validation warning
type ValidationWarning struct {
	Path       string `json:"path"`
	Option     string `json:"option"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// NewConfigureFunction creates a new configure function instance
func NewConfigureFunction() *ConfigureFunction {
	parameters := []functionbase.FunctionParameter{
		{
			Name:        "context",
			Type:        "string",
			Description: "The context or reason for the configuration operation",
			Required:    true,
		},
		{
			Name:        "operation",
			Type:        "string",
			Description: "The operation to perform (get, set, add, remove, validate, backup, restore, preview)",
			Required:    false,
		},
		{
			Name:        "config_type",
			Type:        "string",
			Description: "Type of configuration (system, user, package)",
			Required:    false,
		},
		{
			Name:        "config_path",
			Type:        "string",
			Description: "Path to configuration file",
			Required:    false,
		},
		{
			Name:        "module",
			Type:        "string",
			Description: "Configuration module to modify",
			Required:    false,
		},
		{
			Name:        "option",
			Type:        "string",
			Description: "Configuration option to modify",
			Required:    false,
		},
		{
			Name:        "value",
			Type:        "string",
			Description: "Value to set for the configuration option",
			Required:    false,
		},
		{
			Name:        "dry_run",
			Type:        "boolean",
			Description: "Whether to perform a dry run without making changes",
			Required:    false,
		},
		{
			Name:        "backup",
			Type:        "boolean",
			Description: "Whether to create a backup before making changes",
			Required:    false,
		},
	}

	return &ConfigureFunction{
		BaseFunction: functionbase.NewBaseFunction("configure", "Manage NixOS configuration files and options", parameters),
		agent:        agent.NewConfigAgent(),
		logger:       logger.NewLogger(),
	}
}

// Name returns the function name
func (f *ConfigureFunction) Name() string {
	return f.BaseFunction.Name()
}

// Description returns the function description
func (f *ConfigureFunction) Description() string {
	return f.BaseFunction.Description()
}

// Version returns the function version
func (f *ConfigureFunction) Version() string {
	return "1.0.0"
}

// Parameters returns the function parameter schema
func (f *ConfigureFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"context": map[string]interface{}{
				"type":        "string",
				"description": "The context or reason for the configuration operation",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The configuration operation to perform",
				"enum":        []string{"get", "set", "add", "remove", "validate", "backup", "restore", "preview"},
				"default":     "get",
			},
			"config_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of configuration to manage",
				"enum":        []string{"system", "nixos", "home-manager", "flake", "module", "option"},
				"default":     "system",
			},
			"config_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the configuration file",
			},
			"module": map[string]interface{}{
				"type":        "string",
				"description": "The module name to configure",
			},
			"option": map[string]interface{}{
				"type":        "string",
				"description": "The specific option to configure",
			},
			"value": map[string]interface{}{
				"description": "The value to set for the option",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to perform a dry run without making actual changes",
				"default":     false,
			},
			"backup": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create a backup before making changes",
				"default":     true,
			},
			"validate": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to validate the configuration",
				"default":     true,
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "The output format for configuration data",
				"enum":        []string{"nix", "json", "yaml", "toml"},
				"default":     "nix",
			},
			"options": map[string]interface{}{
				"type":        "object",
				"description": "Additional configuration options",
			},
		},
		"required": []string{"context"},
	}
}

// ValidateParameters validates the function parameters
func (f *ConfigureFunction) ValidateParameters(params map[string]interface{}) error {
	// Check required parameters
	if context, ok := params["context"].(string); !ok || context == "" {
		return fmt.Errorf("parameter 'context' is required and must be a non-empty string")
	}

	// Validate operation if provided
	if operation, ok := params["operation"].(string); ok {
		validOperations := []string{"get", "set", "add", "remove", "validate", "backup", "restore", "preview"}
		valid := false
		for _, validOp := range validOperations {
			if operation == validOp {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("parameter 'operation' must be one of: %v", validOperations)
		}
	}

	// Validate config_type if provided
	if configType, ok := params["config_type"].(string); ok {
		validTypes := []string{"system", "nixos", "home-manager", "flake", "module", "option"}
		valid := false
		for _, validType := range validTypes {
			if configType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("parameter 'config_type' must be one of: %v", validTypes)
		}
	}

	// Validate format if provided
	if format, ok := params["format"].(string); ok {
		validFormats := []string{"nix", "json", "yaml", "toml"}
		valid := false
		for _, validFormat := range validFormats {
			if format == validFormat {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("parameter 'format' must be one of: %v", validFormats)
		}
	}

	return nil
}

// Execute runs the configure function with the given parameters
func (f *ConfigureFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Parse the request manually
	var req ConfigureRequest
	if context, ok := params["context"].(string); ok {
		req.Context = context
	}
	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	}
	if configType, ok := params["config_type"].(string); ok {
		req.ConfigType = configType
	}
	if configPath, ok := params["config_path"].(string); ok {
		req.ConfigPath = configPath
	}
	if module, ok := params["module"].(string); ok {
		req.Module = module
	}
	if option, ok := params["option"].(string); ok {
		req.Option = option
	}
	if value, ok := params["value"]; ok {
		req.Value = value
	}
	if dryRun, ok := params["dry_run"].(bool); ok {
		req.DryRun = dryRun
	}
	if backup, ok := params["backup"].(bool); ok {
		req.Backup = backup
	}

	// Set defaults
	if req.Operation == "" {
		req.Operation = "get"
	}
	if req.ConfigType == "" {
		req.ConfigType = "system"
	}
	if req.Format == "" {
		req.Format = "nix"
	}

	f.logger.Info(fmt.Sprintf("Executing configure operation: %s for %s", req.Operation, req.ConfigType))

	// Execute the configuration operation
	response, err := f.executeConfigureOperation(ctx, &req)
	if err != nil {
		return functionbase.ErrorResult(err, time.Since(startTime)), nil
	}

	response.ExecutionTime = time.Since(startTime)

	return functionbase.SuccessResult(*response, time.Since(startTime)), nil
}

// executeConfigureOperation performs the actual configuration operation
func (f *ConfigureFunction) executeConfigureOperation(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	// Mock implementation since agent methods don't exist yet
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	switch strings.ToLower(req.Operation) {
	case "get":
		if req.Option != "" {
			response.Status = fmt.Sprintf("Retrieved value for option %s", req.Option)
		} else {
			response.Status = fmt.Sprintf("Retrieved configuration for %s", req.ConfigType)
		}
	case "set":
		if req.Option == "" || req.Value == nil {
			return nil, fmt.Errorf("option and value are required for set operation")
		}
		response.Status = fmt.Sprintf("Set %s to %v", req.Option, req.Value)
		response.Changes = []ConfigChange{{
			Option:    req.Option,
			NewValue:  req.Value,
			Operation: "update",
			Status:    "success",
		}}
	case "add":
		response.Status = fmt.Sprintf("Added configuration for %s", req.Option)
		response.Changes = []ConfigChange{{
			Option:    req.Option,
			NewValue:  req.Value,
			Operation: "add",
			Status:    "success",
		}}
	case "remove":
		response.Status = fmt.Sprintf("Removed configuration for %s", req.Option)
		response.Changes = []ConfigChange{{
			Option:    req.Option,
			Operation: "remove",
			Status:    "success",
		}}
	case "validate":
		response.Status = "Configuration validation completed successfully"
		response.ValidationResult = &ValidationResult{
			Valid:   true,
			Errors:  []ValidationError{},
			Summary: "All configurations are valid",
		}
	case "backup":
		response.Status = "Configuration backup created successfully"
		response.BackupPath = "/tmp/nixos-config-backup-" + time.Now().Format("20060102-150405")
	case "restore":
		response.Status = "Configuration restored successfully"
	case "preview":
		response.Status = "Configuration preview generated"
		response.PreviewContent = "# Preview of configuration changes\n# No actual changes would be made"
	default:
		return nil, fmt.Errorf("unsupported configure operation: %s", req.Operation)
	}

	return response, nil
}

// parseRequest parses the raw parameters into a ConfigureRequest
func (f *ConfigureFunction) parseRequest(params map[string]interface{}) (*ConfigureRequest, error) {
	req := &ConfigureRequest{}

	if context, ok := params["context"].(string); ok {
		req.Context = context
	}

	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	} else {
		req.Operation = "get" // default operation
	}

	if configType, ok := params["config_type"].(string); ok {
		req.ConfigType = configType
	}

	if configPath, ok := params["config_path"].(string); ok {
		req.ConfigPath = configPath
	}

	if module, ok := params["module"].(string); ok {
		req.Module = module
	}

	if option, ok := params["option"].(string); ok {
		req.Option = option
	}

	if value, ok := params["value"]; ok {
		req.Value = value
	}

	if dryRun, ok := params["dry_run"].(bool); ok {
		req.DryRun = dryRun
	}

	if backup, ok := params["backup"].(bool); ok {
		req.Backup = backup
	}

	if validate, ok := params["validate"].(bool); ok {
		req.Validate = validate
	}

	if format, ok := params["format"].(string); ok {
		req.Format = format
	}

	if options, ok := params["options"].(map[string]interface{}); ok {
		req.Options = make(map[string]string)
		for k, v := range options {
			if strVal, ok := v.(string); ok {
				req.Options[k] = strVal
			}
		}
	}

	return req, nil
}

// determineOperation determines the appropriate operation based on context
func (f *ConfigureFunction) determineOperation(req *ConfigureRequest) string {
	context := strings.ToLower(req.Context)

	if strings.Contains(context, "configure") || strings.Contains(context, "update") || strings.Contains(context, "set") {
		return "update"
	}
	if strings.Contains(context, "add") || strings.Contains(context, "enable") {
		return "add"
	}
	if strings.Contains(context, "remove") || strings.Contains(context, "disable") {
		return "remove"
	}
	if strings.Contains(context, "validate") || strings.Contains(context, "check") {
		return "validate"
	}
	if strings.Contains(context, "backup") {
		return "backup"
	}
	if strings.Contains(context, "restore") {
		return "restore"
	}
	if strings.Contains(context, "preview") || strings.Contains(context, "show") {
		return "preview"
	}

	return "get" // default
}

// validateConfiguration validates the configuration request
func (f *ConfigureFunction) validateConfiguration(req *ConfigureRequest, configContent string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Only validate context if we're doing operation-specific validation
	if req.Operation != "" {
		// Operation-specific validation
		switch req.Operation {
		case "set", "add":
			if req.Option == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Path:     "option",
					Option:   req.Option,
					Message:  "Option is required for set/add operations",
					Severity: "error",
				})
			}
			if req.Value == nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Path:     "value",
					Option:   req.Option,
					Message:  "Value is required for set/add operations",
					Severity: "error",
				})
			}
		case "remove":
			if req.Option == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Path:     "option",
					Option:   req.Option,
					Message:  "Option is required for remove operations",
					Severity: "error",
				})
			}
		}
	}

	// Validate configuration content if provided (including empty content)
	trimmed := strings.TrimSpace(configContent)

	// Check for empty configuration
	if trimmed == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     "config",
			Message:  "Configuration content is empty",
			Severity: "error",
		})
	} else {
		// Basic syntax validation for Nix configurations
		if strings.Contains(configContent, "syntax error") {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Path:     "config",
				Message:  "Configuration contains syntax errors",
				Severity: "error",
			})
		}

		// Check for common Nix syntax issues
		if !f.isValidNixSyntax(configContent) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Path:     "config",
				Message:  "Configuration contains syntax errors",
				Severity: "error",
			})
		}
	}

	if result.Valid {
		result.Summary = "Configuration validation passed"
	} else {
		result.Summary = fmt.Sprintf("Configuration validation failed with %d errors", len(result.Errors))
	}

	return result
}

// isValidNixSyntax performs basic syntax validation for Nix configuration content
func (f *ConfigureFunction) isValidNixSyntax(content string) bool {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Must start and end with braces for attribute sets
	if !strings.HasPrefix(content, "{") || !strings.HasSuffix(content, "}") {
		return false
	}

	// Check for balanced braces
	braceCount := 0
	for _, char := range content {
		switch char {
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
	}
	if braceCount != 0 {
		return false
	}

	// Check for lines that should end with semicolon
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || line == "{" || line == "}" {
			continue
		}

		// Lines with assignments should end with semicolon
		if strings.Contains(line, "=") && !strings.HasSuffix(line, ";") && !strings.HasSuffix(line, "{") {
			return false
		}
	}

	return true
}

// generateBackupPath generates a backup path for configuration files
func (f *ConfigureFunction) generateBackupPath(req *ConfigureRequest) string {
	configPath := req.ConfigPath
	if configPath == "" {
		return ""
	}

	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s.backup-%s", configPath, timestamp)
}

// applyConfigurationChange applies a configuration change
func (f *ConfigureFunction) applyConfigurationChange(req *ConfigureRequest) (*ConfigChange, error) {
	// Validate operation
	validOperations := []string{"add", "update", "set", "remove"}
	validOp := false
	for _, op := range validOperations {
		if req.Operation == op {
			validOp = true
			break
		}
	}
	if !validOp {
		return nil, fmt.Errorf("invalid operation: %s", req.Operation)
	}

	// Validate option for operations that require it
	if req.Operation == "add" && req.Option == "" {
		return nil, fmt.Errorf("option is required for %s operation", req.Operation)
	}

	change := &ConfigChange{
		Path:      req.ConfigPath,
		Option:    req.Option,
		NewValue:  req.Value,
		Operation: req.Operation,
		Status:    "success",
	}

	if req.DryRun {
		change.Status = "dry-run"
		f.logger.Info("Dry run - no actual changes made")
		return change, nil
	}

	// Simulate applying the change
	f.logger.Info(fmt.Sprintf("Applying configuration change: %s %s", req.Operation, req.Option))

	return change, nil
}
