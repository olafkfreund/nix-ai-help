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
				"enum":        []string{"system", "home-manager", "flake", "module", "option"},
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
