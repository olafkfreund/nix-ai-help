package hardware

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// HardwareFunction handles hardware detection and configuration
type HardwareFunction struct {
	*functionbase.BaseFunction
	agent  *agent.HardwareAgent
	logger *logger.Logger
}

// HardwareRequest represents the input parameters for the hardware function
type HardwareRequest struct {
	Context        string            `json:"context"`
	Operation      string            `json:"operation,omitempty"`
	ComponentType  string            `json:"component_type,omitempty"`
	DetectAll      bool              `json:"detect_all,omitempty"`
	Generate       bool              `json:"generate,omitempty"`
	Format         string            `json:"format,omitempty"`
	IncludeDrivers bool              `json:"include_drivers,omitempty"`
	Options        map[string]string `json:"options,omitempty"`
}

// HardwareResponse represents the output of the hardware function
type HardwareResponse struct {
	Context         string              `json:"context"`
	Status          string              `json:"status"`
	Operation       string              `json:"operation"`
	Hardware        []HardwareComponent `json:"hardware,omitempty"`
	Configuration   string              `json:"configuration,omitempty"`
	Recommendations []string            `json:"recommendations,omitempty"`
	Issues          []HardwareIssue     `json:"issues,omitempty"`
	ErrorMessage    string              `json:"error_message,omitempty"`
	ExecutionTime   time.Duration       `json:"execution_time,omitempty"`
}

// HardwareComponent represents a detected hardware component
type HardwareComponent struct {
	Type          string            `json:"type"`
	Name          string            `json:"name"`
	Vendor        string            `json:"vendor,omitempty"`
	Model         string            `json:"model,omitempty"`
	Driver        string            `json:"driver,omitempty"`
	Supported     bool              `json:"supported"`
	Status        string            `json:"status"`
	Configuration map[string]string `json:"configuration,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// HardwareIssue represents a hardware-related issue
type HardwareIssue struct {
	Component   string   `json:"component"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Solution    string   `json:"solution,omitempty"`
	Resources   []string `json:"resources,omitempty"`
}

// NewHardwareFunction creates a new hardware function instance
func NewHardwareFunction() *HardwareFunction {
	return &HardwareFunction{
		BaseFunction: &functionbase.BaseFunction{
			FuncName:    "hardware",
			FuncDesc:    "Detect and configure hardware components for NixOS",
			FuncVersion: "1.0.0",
		},
		agent:  agent.NewHardwareAgent(),
		logger: logger.NewLogger(),
	}
}

// Name returns the function name
func (f *HardwareFunction) Name() string {
	return f.FuncName
}

// Description returns the function description
func (f *HardwareFunction) Description() string {
	return f.FuncDesc
}

// Version returns the function version
func (f *HardwareFunction) Version() string {
	return f.FuncVersion
}

// Parameters returns the function parameter schema
func (f *HardwareFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"context": map[string]interface{}{
				"type":        "string",
				"description": "The context or reason for the hardware operation",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The hardware operation to perform",
				"enum":        []string{"detect", "generate-config", "scan", "test", "diagnose", "list-drivers"},
				"default":     "detect",
			},
			"component_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of hardware component to focus on",
				"enum":        []string{"cpu", "gpu", "network", "audio", "storage", "input", "display", "all"},
				"default":     "all",
			},
			"detect_all": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to detect all hardware components",
				"default":     true,
			},
			"generate": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to generate NixOS configuration",
				"default":     false,
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "The output format for configuration",
				"enum":        []string{"nix", "json", "yaml"},
				"default":     "nix",
			},
			"include_drivers": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include driver information",
				"default":     true,
			},
			"options": map[string]interface{}{
				"type":        "object",
				"description": "Additional hardware detection options",
			},
		},
		"required": []string{"context"},
	}
}

// Execute runs the hardware function with the given parameters
func (f *HardwareFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Parse the request
	var req HardwareRequest
	if err := f.ParseParams(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Set defaults
	if req.Operation == "" {
		req.Operation = "detect"
	}
	if req.ComponentType == "" {
		req.ComponentType = "all"
	}
	if req.Format == "" {
		req.Format = "nix"
	}

	f.logger.Info(fmt.Sprintf("Executing hardware operation: %s for %s", req.Operation, req.ComponentType))

	// Execute the hardware operation
	response, err := f.executeHardwareOperation(ctx, &req)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Data: HardwareResponse{
				Context:       req.Context,
				Operation:     req.Operation,
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

// executeHardwareOperation performs the actual hardware operation
func (f *HardwareFunction) executeHardwareOperation(ctx context.Context, req *HardwareRequest) (*HardwareResponse, error) {
	response := &HardwareResponse{
		Context:   req.Context,
		Operation: req.Operation,
		Status:    "success",
		Hardware:  []HardwareComponent{},
		Issues:    []HardwareIssue{},
	}

	switch req.Operation {
	case "detect":
		return f.detectHardware(ctx, req, response)
	case "generate-config":
		return f.generateConfig(ctx, req, response)
	case "scan":
		return f.scanHardware(ctx, req, response)
	case "test":
		return f.testHardware(ctx, req, response)
	case "diagnose":
		return f.diagnoseHardware(ctx, req, response)
	case "list-drivers":
		return f.listDrivers(ctx, req, response)
	default:
		return nil, fmt.Errorf("unsupported hardware operation: %s", req.Operation)
	}
}

// detectHardware detects hardware components
func (f *HardwareFunction) detectHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Detecting hardware components")

	// Detect hardware based on component type
	components, err := f.agent.DetectHardware(ctx, req.ComponentType, req.IncludeDrivers)
	if err != nil {
		return nil, fmt.Errorf("hardware detection failed: %w", err)
	}

	// Convert to response format
	for _, comp := range components {
		hardware := HardwareComponent{
			Type:          comp.Type,
			Name:          comp.Name,
			Vendor:        comp.Vendor,
			Model:         comp.Model,
			Driver:        comp.Driver,
			Supported:     comp.Supported,
			Status:        comp.Status,
			Configuration: comp.Configuration,
			Metadata:      comp.Metadata,
		}
		response.Hardware = append(response.Hardware, hardware)
	}

	// Generate recommendations
	response.Recommendations = f.generateHardwareRecommendations(response.Hardware)

	// Check for issues
	response.Issues = f.detectHardwareIssues(response.Hardware)

	// Generate configuration if requested
	if req.Generate {
		config, err := f.agent.GenerateConfiguration(ctx, components, req.Format)
		if err != nil {
			f.logger.Error(fmt.Sprintf("Failed to generate configuration: %v", err))
		} else {
			response.Configuration = config
		}
	}

	f.logger.Info(fmt.Sprintf("Detected %d hardware components", len(response.Hardware)))

	return response, nil
}

// generateConfig generates NixOS hardware configuration
func (f *HardwareFunction) generateConfig(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Generating hardware configuration")

	// First detect hardware
	components, err := f.agent.DetectHardware(ctx, req.ComponentType, req.IncludeDrivers)
	if err != nil {
		return nil, fmt.Errorf("hardware detection failed during config generation: %w", err)
	}

	// Generate configuration
	config, err := f.agent.GenerateConfiguration(ctx, components, req.Format)
	if err != nil {
		return nil, fmt.Errorf("configuration generation failed: %w", err)
	}

	response.Configuration = config

	// Also include hardware details
	for _, comp := range components {
		hardware := HardwareComponent{
			Type:          comp.Type,
			Name:          comp.Name,
			Vendor:        comp.Vendor,
			Model:         comp.Model,
			Driver:        comp.Driver,
			Supported:     comp.Supported,
			Status:        comp.Status,
			Configuration: comp.Configuration,
		}
		response.Hardware = append(response.Hardware, hardware)
	}

	// Generate recommendations for configuration
	response.Recommendations = f.generateConfigRecommendations(components, req.Format)

	f.logger.Info("Hardware configuration generated successfully")

	return response, nil
}

// scanHardware performs comprehensive hardware scan
func (f *HardwareFunction) scanHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Performing comprehensive hardware scan")

	// Perform deep scan
	scanResult, err := f.agent.ScanHardware(ctx, req.ComponentType)
	if err != nil {
		return nil, fmt.Errorf("hardware scan failed: %w", err)
	}

	// Convert scan results
	for _, comp := range scanResult.Components {
		hardware := HardwareComponent{
			Type:          comp.Type,
			Name:          comp.Name,
			Vendor:        comp.Vendor,
			Model:         comp.Model,
			Driver:        comp.Driver,
			Supported:     comp.Supported,
			Status:        comp.Status,
			Configuration: comp.Configuration,
			Metadata:      comp.Metadata,
		}
		response.Hardware = append(response.Hardware, hardware)
	}

	// Include scan issues
	for _, issue := range scanResult.Issues {
		hwIssue := HardwareIssue{
			Component:   issue.Component,
			Severity:    issue.Severity,
			Description: issue.Description,
			Solution:    issue.Solution,
			Resources:   issue.Resources,
		}
		response.Issues = append(response.Issues, hwIssue)
	}

	// Generate comprehensive recommendations
	response.Recommendations = f.generateScanRecommendations(scanResult)

	f.logger.Info(fmt.Sprintf("Hardware scan completed: %d components, %d issues", len(response.Hardware), len(response.Issues)))

	return response, nil
}

// testHardware tests hardware functionality
func (f *HardwareFunction) testHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Testing hardware functionality")

	// Run hardware tests
	testResults, err := f.agent.TestHardware(ctx, req.ComponentType)
	if err != nil {
		return nil, fmt.Errorf("hardware testing failed: %w", err)
	}

	// Convert test results
	for _, result := range testResults {
		hardware := HardwareComponent{
			Type:          result.Component.Type,
			Name:          result.Component.Name,
			Vendor:        result.Component.Vendor,
			Model:         result.Component.Model,
			Driver:        result.Component.Driver,
			Supported:     result.Component.Supported,
			Status:        result.TestStatus,
			Configuration: result.Component.Configuration,
			Metadata: map[string]string{
				"test_result": result.TestResult,
				"test_time":   result.TestTime.String(),
			},
		}
		response.Hardware = append(response.Hardware, hardware)

		// Add issues for failed tests
		if result.TestStatus == "failed" {
			issue := HardwareIssue{
				Component:   result.Component.Name,
				Severity:    "warning",
				Description: fmt.Sprintf("Hardware test failed: %s", result.TestResult),
				Solution:    result.Solution,
			}
			response.Issues = append(response.Issues, issue)
		}
	}

	response.Recommendations = f.generateTestRecommendations(testResults)

	f.logger.Info(fmt.Sprintf("Hardware testing completed: %d components tested", len(testResults)))

	return response, nil
}

// diagnoseHardware diagnoses hardware issues
func (f *HardwareFunction) diagnoseHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Diagnosing hardware issues")

	// Run hardware diagnostics
	diagnosis, err := f.agent.DiagnoseHardware(ctx, req.ComponentType)
	if err != nil {
		return nil, fmt.Errorf("hardware diagnosis failed: %w", err)
	}

	// Convert diagnosis results
	for _, comp := range diagnosis.Components {
		hardware := HardwareComponent{
			Type:          comp.Type,
			Name:          comp.Name,
			Vendor:        comp.Vendor,
			Model:         comp.Model,
			Driver:        comp.Driver,
			Supported:     comp.Supported,
			Status:        comp.DiagnosisStatus,
			Configuration: comp.Configuration,
			Metadata: map[string]string{
				"diagnosis_result": comp.DiagnosisResult,
				"confidence":       fmt.Sprintf("%.2f", comp.Confidence),
			},
		}
		response.Hardware = append(response.Hardware, hardware)
	}

	// Convert issues
	for _, issue := range diagnosis.Issues {
		hwIssue := HardwareIssue{
			Component:   issue.Component,
			Severity:    issue.Severity,
			Description: issue.Description,
			Solution:    issue.Solution,
			Resources:   issue.Resources,
		}
		response.Issues = append(response.Issues, hwIssue)
	}

	response.Recommendations = f.generateDiagnosisRecommendations(diagnosis)

	f.logger.Info(fmt.Sprintf("Hardware diagnosis completed: %d issues found", len(response.Issues)))

	return response, nil
}

// listDrivers lists available drivers for hardware
func (f *HardwareFunction) listDrivers(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Listing available hardware drivers")

	// Get available drivers
	drivers, err := f.agent.ListDrivers(ctx, req.ComponentType)
	if err != nil {
		return nil, fmt.Errorf("failed to list drivers: %w", err)
	}

	// Convert to hardware components format
	for _, driver := range drivers {
		hardware := HardwareComponent{
			Type:      driver.ComponentType,
			Name:      driver.Name,
			Driver:    driver.Name,
			Supported: driver.Supported,
			Status:    driver.Status,
			Metadata: map[string]string{
				"version":     driver.Version,
				"description": driver.Description,
				"package":     driver.Package,
			},
		}
		response.Hardware = append(response.Hardware, hardware)
	}

	response.Recommendations = f.generateDriverRecommendations(drivers)

	f.logger.Info(fmt.Sprintf("Listed %d available drivers", len(drivers)))

	return response, nil
}

// generateHardwareRecommendations generates recommendations based on detected hardware
func (f *HardwareFunction) generateHardwareRecommendations(hardware []HardwareComponent) []string {
	recommendations := []string{}

	unsupportedCount := 0
	for _, hw := range hardware {
		if !hw.Supported {
			unsupportedCount++
		}
	}

	if unsupportedCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d hardware components may need additional configuration", unsupportedCount))
	}

	// Check for common hardware types
	hasGPU := false
	hasAudio := false
	hasNetwork := false

	for _, hw := range hardware {
		switch hw.Type {
		case "gpu":
			hasGPU = true
		case "audio":
			hasAudio = true
		case "network":
			hasNetwork = true
		}
	}

	if hasGPU {
		recommendations = append(recommendations, "GPU detected. Consider enabling appropriate graphics drivers.")
	}
	if hasAudio {
		recommendations = append(recommendations, "Audio hardware detected. Ensure PulseAudio or PipeWire is configured.")
	}
	if hasNetwork {
		recommendations = append(recommendations, "Network hardware detected. Consider configuring NetworkManager.")
	}

	recommendations = append(recommendations, "Use 'nixai hardware --operation=generate-config' to create hardware configuration.")

	return recommendations
}

// generateConfigRecommendations generates recommendations for configuration
func (f *HardwareFunction) generateConfigRecommendations(components []agent.HardwareComponent, format string) []string {
	recommendations := []string{}

	recommendations = append(recommendations, fmt.Sprintf("Generated %s configuration for %d hardware components", format, len(components)))
	recommendations = append(recommendations, "Review the configuration before applying to your system")
	recommendations = append(recommendations, "Consider backing up your current configuration first")
	recommendations = append(recommendations, "Test the configuration in a virtual machine if possible")

	return recommendations
}

// generateScanRecommendations generates recommendations based on scan results
func (f *HardwareFunction) generateScanRecommendations(scanResult *agent.ScanResult) []string {
	recommendations := []string{}

	if len(scanResult.Issues) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d hardware issues that need attention", len(scanResult.Issues)))
	}

	recommendations = append(recommendations, "Use 'nixai hardware --operation=test' to verify hardware functionality")
	recommendations = append(recommendations, "Use 'nixai hardware --operation=diagnose' for detailed issue analysis")

	return recommendations
}

// generateTestRecommendations generates recommendations based on test results
func (f *HardwareFunction) generateTestRecommendations(testResults []agent.TestResult) []string {
	recommendations := []string{}

	failedTests := 0
	for _, result := range testResults {
		if result.TestStatus == "failed" {
			failedTests++
		}
	}

	if failedTests > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d hardware tests failed. Review issues and solutions.", failedTests))
	} else {
		recommendations = append(recommendations, "All hardware tests passed successfully.")
	}

	return recommendations
}

// generateDiagnosisRecommendations generates recommendations based on diagnosis
func (f *HardwareFunction) generateDiagnosisRecommendations(diagnosis *agent.DiagnosisResult) []string {
	recommendations := []string{}

	criticalIssues := 0
	for _, issue := range diagnosis.Issues {
		if issue.Severity == "critical" {
			criticalIssues++
		}
	}

	if criticalIssues > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d critical hardware issues require immediate attention", criticalIssues))
	}

	recommendations = append(recommendations, "Follow the provided solutions to resolve hardware issues")
	recommendations = append(recommendations, "Check NixOS hardware compatibility list for additional information")

	return recommendations
}

// generateDriverRecommendations generates recommendations for drivers
func (f *HardwareFunction) generateDriverRecommendations(drivers []agent.Driver) []string {
	recommendations := []string{}

	unsupportedDrivers := 0
	for _, driver := range drivers {
		if !driver.Supported {
			unsupportedDrivers++
		}
	}

	if unsupportedDrivers > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d drivers may not be fully supported", unsupportedDrivers))
	}

	recommendations = append(recommendations, "Enable required drivers in your NixOS configuration")
	recommendations = append(recommendations, "Consider using nixos-hardware for automatic driver configuration")

	return recommendations
}

// detectHardwareIssues detects common hardware issues
func (f *HardwareFunction) detectHardwareIssues(hardware []HardwareComponent) []HardwareIssue {
	issues := []HardwareIssue{}

	for _, hw := range hardware {
		if !hw.Supported {
			issue := HardwareIssue{
				Component:   hw.Name,
				Severity:    "warning",
				Description: fmt.Sprintf("%s (%s) may not be fully supported", hw.Name, hw.Type),
				Solution:    "Check NixOS hardware database for compatibility information",
				Resources:   []string{"https://github.com/NixOS/nixos-hardware"},
			}
			issues = append(issues, issue)
		}

		if hw.Status == "error" || hw.Status == "failed" {
			issue := HardwareIssue{
				Component:   hw.Name,
				Severity:    "error",
				Description: fmt.Sprintf("%s has hardware errors", hw.Name),
				Solution:    "Check hardware connections and driver configuration",
			}
			issues = append(issues, issue)
		}
	}

	return issues
}
