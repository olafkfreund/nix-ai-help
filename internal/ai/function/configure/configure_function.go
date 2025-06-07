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
	return &ConfigureFunction{
		BaseFunction: &functionbase.BaseFunction{
			FuncName:    "configure",
			FuncDesc:    "Manage NixOS configuration files and options",
			FuncVersion: "1.0.0",
		},
		agent:  agent.NewConfigAgent(),
		logger: logger.NewLogger(),
	}
}

// Name returns the function name
func (f *ConfigureFunction) Name() string {
	return f.FuncName
}

// Description returns the function description
func (f *ConfigureFunction) Description() string {
	return f.FuncDesc
}

// Version returns the function version
func (f *ConfigureFunction) Version() string {
	return f.FuncVersion
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

	// Parse the request
	var req ConfigureRequest
	if err := f.ParseParams(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
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
		return &functionbase.FunctionResult{
			Success: false,
			Data: ConfigureResponse{
				Context:       req.Context,
				Status:        "error",
				ErrorMessage:  err.Error(),
				ExecutionTime: time.Since(startTime),
			},
			Error:         err,
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	response.ExecutionTime = time.Since(startTime)

	return &functionbase.FunctionResult{
		Success:       true,
		Data:          *response,
		ExecutionTime: time.Since(startTime),
	}, nil
}

// executeConfigureOperation performs the actual configuration operation
func (f *ConfigureFunction) executeConfigureOperation(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	switch strings.ToLower(req.Operation) {
	case "get":
		return f.getConfiguration(ctx, req)
	case "set":
		return f.setConfiguration(ctx, req)
	case "add":
		return f.addConfiguration(ctx, req)
	case "remove":
		return f.removeConfiguration(ctx, req)
	case "validate":
		return f.validateConfiguration(ctx, req)
	case "backup":
		return f.backupConfiguration(ctx, req)
	case "restore":
		return f.restoreConfiguration(ctx, req)
	case "preview":
		return f.previewConfiguration(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported configure operation: %s", req.Operation)
	}
}

// getConfiguration retrieves configuration values
func (f *ConfigureFunction) getConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to get configuration
	result, err := f.agent.GetConfiguration(ctx, req.ConfigType, req.ConfigPath, req.Module, req.Option)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	response.PreviewContent = result
	f.logger.Info("Configuration retrieved successfully")

	return response, nil
}

// setConfiguration sets configuration values
func (f *ConfigureFunction) setConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	if req.Option == "" || req.Value == nil {
		return nil, fmt.Errorf("option and value are required for set operation")
	}

	// Create backup if requested
	var backupPath string
	if req.Backup {
		backup, err := f.agent.CreateBackup(ctx, req.ConfigPath)
		if err != nil {
			f.logger.Error(fmt.Sprintf("Failed to create backup: %v", err))
		} else {
			backupPath = backup
			response.BackupPath = backupPath
		}
	}

	// Validate before changes if requested
	if req.Validate {
		validation, err := f.agent.ValidateConfiguration(ctx, req.ConfigPath)
		if err != nil {
			f.logger.Error(fmt.Sprintf("Validation failed: %v", err))
		} else {
			response.ValidationResult = &ValidationResult{
				Valid:   validation.Valid,
				Summary: validation.Summary,
			}
		}
	}

	// Perform the configuration change
	if !req.DryRun {
		err := f.agent.SetConfiguration(ctx, req.ConfigType, req.ConfigPath, req.Module, req.Option, req.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to set configuration: %w", err)
		}
	}

	// Record the change
	change := ConfigChange{
		Path:      req.ConfigPath,
		Option:    req.Option,
		NewValue:  req.Value,
		Operation: "set",
		Status:    "success",
	}
	response.Changes = append(response.Changes, change)

	f.logger.Info(fmt.Sprintf("Configuration set successfully: %s = %v", req.Option, req.Value))

	return response, nil
}

// addConfiguration adds new configuration entries
func (f *ConfigureFunction) addConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to add configuration
	err := f.agent.AddConfiguration(ctx, req.ConfigType, req.ConfigPath, req.Module, req.Option, req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to add configuration: %w", err)
	}

	// Record the change
	change := ConfigChange{
		Path:      req.ConfigPath,
		Option:    req.Option,
		NewValue:  req.Value,
		Operation: "add",
		Status:    "success",
	}
	response.Changes = append(response.Changes, change)

	f.logger.Info(fmt.Sprintf("Configuration added successfully: %s", req.Option))

	return response, nil
}

// removeConfiguration removes configuration entries
func (f *ConfigureFunction) removeConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to remove configuration
	err := f.agent.RemoveConfiguration(ctx, req.ConfigType, req.ConfigPath, req.Module, req.Option)
	if err != nil {
		return nil, fmt.Errorf("failed to remove configuration: %w", err)
	}

	// Record the change
	change := ConfigChange{
		Path:      req.ConfigPath,
		Option:    req.Option,
		Operation: "remove",
		Status:    "success",
	}
	response.Changes = append(response.Changes, change)

	f.logger.Info(fmt.Sprintf("Configuration removed successfully: %s", req.Option))

	return response, nil
}

// validateConfiguration validates configuration files
func (f *ConfigureFunction) validateConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to validate configuration
	validation, err := f.agent.ValidateConfiguration(ctx, req.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	response.ValidationResult = &ValidationResult{
		Valid:   validation.Valid,
		Summary: validation.Summary,
	}

	f.logger.Info("Configuration validation completed")

	return response, nil
}

// backupConfiguration creates configuration backups
func (f *ConfigureFunction) backupConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to create backup
	backupPath, err := f.agent.CreateBackup(ctx, req.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	response.BackupPath = backupPath
	f.logger.Info(fmt.Sprintf("Configuration backup created: %s", backupPath))

	return response, nil
}

// restoreConfiguration restores configuration from backup
func (f *ConfigureFunction) restoreConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to restore configuration
	err := f.agent.RestoreBackup(ctx, req.ConfigPath, req.Options["backup_path"])
	if err != nil {
		return nil, fmt.Errorf("failed to restore configuration: %w", err)
	}

	f.logger.Info("Configuration restored successfully")

	return response, nil
}

// previewConfiguration shows what changes would be made
func (f *ConfigureFunction) previewConfiguration(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	response := &ConfigureResponse{
		Context: req.Context,
		Status:  "success",
		Changes: []ConfigChange{},
	}

	// Use agent to preview configuration
	preview, err := f.agent.PreviewChanges(ctx, req.ConfigType, req.ConfigPath, req.Module, req.Option, req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to preview configuration: %w", err)
	}

	response.PreviewContent = preview
	f.logger.Info("Configuration preview generated")

	return response, nil
}
