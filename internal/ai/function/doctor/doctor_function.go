package doctor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// DoctorFunction handles system health checks and diagnostic operations
type DoctorFunction struct {
	*functionbase.BaseFunction
	agent  *agent.DoctorAgent
	logger *logger.Logger
}

// DoctorRequest represents the input parameters for the doctor function
type DoctorRequest struct {
	Operation    string            `json:"operation"`
	CheckType    string            `json:"check_type,omitempty"`
	Severity     string            `json:"severity,omitempty"`
	Component    string            `json:"component,omitempty"`
	Category     string            `json:"category,omitempty"`
	AutoFix      bool              `json:"auto_fix,omitempty"`
	Verbose      bool              `json:"verbose,omitempty"`
	OutputFormat string            `json:"output_format,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
}

// DoctorResponse represents the output of the doctor function
type DoctorResponse struct {
	Operation       string         `json:"operation"`
	Status          string         `json:"status"`
	OverallHealth   string         `json:"overall_health"`
	Checks          []HealthCheck  `json:"checks,omitempty"`
	Issues          []Issue        `json:"issues,omitempty"`
	Fixes           []Fix          `json:"fixes,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`
	Summary         *HealthSummary `json:"summary,omitempty"`
	ExecutionTime   time.Duration  `json:"execution_time,omitempty"`
}

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Result      string `json:"result,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Category    string `json:"category,omitempty"`
	Details     string `json:"details,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

// Issue represents a detected system issue
type Issue struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Category    string   `json:"category"`
	Component   string   `json:"component,omitempty"`
	Impact      string   `json:"impact,omitempty"`
	Solutions   []string `json:"solutions,omitempty"`
	References  []string `json:"references,omitempty"`
}

// Fix represents an applied or suggested fix
type Fix struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Command     string `json:"command,omitempty"`
	Result      string `json:"result,omitempty"`
	Applied     bool   `json:"applied"`
}

// HealthSummary provides an overall health summary
type HealthSummary struct {
	TotalChecks    int `json:"total_checks"`
	PassedChecks   int `json:"passed_checks"`
	FailedChecks   int `json:"failed_checks"`
	WarningChecks  int `json:"warning_checks"`
	CriticalIssues int `json:"critical_issues"`
	Warnings       int `json:"warnings"`
	FixesApplied   int `json:"fixes_applied"`
}

// NewDoctorFunction creates a new doctor function
func NewDoctorFunction() *DoctorFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Doctor operation to perform", true,
			[]string{"check", "diagnose", "fix", "status", "summary", "full-scan"}, nil, nil),
		functionbase.StringParamWithOptions("check_type", "Type of health check", false,
			[]string{"system", "nix", "nixos", "packages", "services", "storage", "network", "security"}, nil, nil),
		functionbase.StringParamWithOptions("severity", "Minimum severity level to report", false,
			[]string{"info", "warning", "error", "critical"}, nil, nil),
		functionbase.StringParam("component", "Specific component to check", false),
		functionbase.StringParam("category", "Category of checks to run", false),
		functionbase.BoolParam("auto_fix", "Automatically apply safe fixes", false),
		functionbase.BoolParam("verbose", "Enable verbose output", false),
		functionbase.StringParamWithOptions("output_format", "Output format", false,
			[]string{"text", "json", "summary", "detailed"}, nil, nil),
		{
			Name:        "options",
			Type:        "object",
			Description: "Additional options for doctor operations",
			Required:    false,
		},
	}

	// Create base function
	baseFunc := functionbase.NewBaseFunction(
		"doctor",
		"Perform comprehensive system health checks and diagnostics for NixOS installations",
		parameters,
	)

	// Create doctor function
	doctorFunc := &DoctorFunction{
		BaseFunction: baseFunc,
		agent:        agent.NewDoctorAgent(),
		logger:       logger.NewLogger(),
	}

	return doctorFunc
}

// Execute implements the FunctionInterface
func (f *DoctorFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Validate parameters
	if err := f.ValidateParameters(params); err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("Parameter validation failed: %v", err),
		}, err
	}

	// Parse request
	req, err := f.parseRequest(params)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}, err
	}

	f.logger.Info(fmt.Sprintf("Executing doctor operation: %s", req.Operation))

	// Execute doctor operation
	response, err := f.executeDoctor(ctx, req)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("Doctor operation failed: %v", err),
		}, err
	}

	// Set execution time
	response.ExecutionTime = time.Since(startTime)

	return &functionbase.FunctionResult{
		Success:  true,
		Data:     response,
		Duration: time.Since(startTime),
	}, nil
}

// parseRequest converts raw parameters to DoctorRequest
func (f *DoctorFunction) parseRequest(params map[string]interface{}) (*DoctorRequest, error) {
	req := &DoctorRequest{}

	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	}

	if checkType, ok := params["check_type"].(string); ok {
		req.CheckType = checkType
	}

	if severity, ok := params["severity"].(string); ok {
		req.Severity = severity
	} else {
		req.Severity = "warning" // default
	}

	if component, ok := params["component"].(string); ok {
		req.Component = component
	}

	if category, ok := params["category"].(string); ok {
		req.Category = category
	}

	if autoFix, ok := params["auto_fix"].(bool); ok {
		req.AutoFix = autoFix
	}

	if verbose, ok := params["verbose"].(bool); ok {
		req.Verbose = verbose
	}

	if outputFormat, ok := params["output_format"].(string); ok {
		req.OutputFormat = outputFormat
	} else {
		req.OutputFormat = "detailed" // default
	}

	if options, ok := params["options"].(map[string]interface{}); ok {
		req.Options = make(map[string]string)
		for k, v := range options {
			if str, ok := v.(string); ok {
				req.Options[k] = str
			}
		}
	}

	return req, nil
}

// executeDoctor performs the doctor operation using the agent
func (f *DoctorFunction) executeDoctor(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	// Execute different doctor operations
	switch req.Operation {
	case "check":
		return f.executeHealthCheck(ctx, req)
	case "diagnose":
		return f.executeDiagnosis(ctx, req)
	case "fix":
		return f.executeFixes(ctx, req)
	case "status":
		return f.executeStatus(ctx, req)
	case "summary":
		return f.executeSummary(ctx, req)
	case "full-scan":
		return f.executeFullScan(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported doctor operation: %s", req.Operation)
	}
}

// executeHealthCheck performs specific health checks
func (f *DoctorFunction) executeHealthCheck(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	checks, err := f.agent.RunHealthChecks(req.CheckType, req.Component)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	response := &DoctorResponse{
		Operation:     req.Operation,
		Status:        "success",
		OverallHealth: f.calculateOverallHealth(checks),
		Checks:        f.parseHealthChecks(checks),
	}

	// Generate summary
	response.Summary = f.generateHealthSummary(response.Checks)

	return response, nil
}

// executeDiagnosis performs system diagnosis
func (f *DoctorFunction) executeDiagnosis(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	issues, err := f.agent.DiagnoseSystem(req.CheckType, req.Severity)
	if err != nil {
		return nil, fmt.Errorf("system diagnosis failed: %w", err)
	}

	response := &DoctorResponse{
		Operation:     req.Operation,
		Status:        "success",
		Issues:        f.parseIssues(issues),
		OverallHealth: f.calculateHealthFromIssues(issues),
	}

	// Generate recommendations
	response.Recommendations = f.generateRecommendations(response.Issues)

	return response, nil
}

// executeFixes applies fixes to detected issues
func (f *DoctorFunction) executeFixes(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	fixes, err := f.agent.ApplyFixes(req.Component, req.AutoFix)
	if err != nil {
		return nil, fmt.Errorf("applying fixes failed: %w", err)
	}

	response := &DoctorResponse{
		Operation: req.Operation,
		Status:    "success",
		Fixes:     f.parseFixes(fixes),
	}

	return response, nil
}

// executeStatus returns current system status
func (f *DoctorFunction) executeStatus(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	status, err := f.agent.GetSystemStatus()
	if err != nil {
		return nil, fmt.Errorf("getting system status failed: %w", err)
	}

	response := &DoctorResponse{
		Operation:     req.Operation,
		Status:        "success",
		OverallHealth: status,
	}

	return response, nil
}

// executeSummary generates a health summary
func (f *DoctorFunction) executeSummary(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	summary, err := f.agent.GenerateSummary()
	if err != nil {
		return nil, fmt.Errorf("generating summary failed: %w", err)
	}

	response := &DoctorResponse{
		Operation: req.Operation,
		Status:    "success",
		Summary:   f.parseHealthSummaryFromAgent(summary),
	}

	return response, nil
}

// executeFullScan performs a comprehensive system scan
func (f *DoctorFunction) executeFullScan(ctx context.Context, req *DoctorRequest) (*DoctorResponse, error) {
	scanResult, err := f.agent.RunFullScan()
	if err != nil {
		return nil, fmt.Errorf("full scan failed: %w", err)
	}

	response := &DoctorResponse{
		Operation:     req.Operation,
		Status:        "success",
		OverallHealth: f.extractOverallHealth(scanResult),
		Checks:        f.extractHealthChecks(scanResult),
		Issues:        f.extractIssues(scanResult),
		Summary:       f.extractSummary(scanResult),
	}

	return response, nil
}

// Helper methods for parsing agent responses
func (f *DoctorFunction) parseHealthChecks(checks []string) []HealthCheck {
	var result []HealthCheck
	for _, check := range checks {
		// Parse check format: "name:status:description"
		parts := strings.Split(check, ":")
		if len(parts) >= 2 {
			result = append(result, HealthCheck{
				Name:   parts[0],
				Status: parts[1],
				Description: func() string {
					if len(parts) > 2 {
						return parts[2]
					}
					return ""
				}(),
			})
		}
	}
	return result
}

func (f *DoctorFunction) parseIssues(issues []string) []Issue {
	var result []Issue
	for i, issue := range issues {
		// Parse basic issue format
		result = append(result, Issue{
			ID:          fmt.Sprintf("issue_%d", i+1),
			Title:       f.extractTitle(issue),
			Description: issue,
			Severity:    f.extractSeverity(issue),
			Category:    f.extractCategory(issue),
		})
	}
	return result
}

func (f *DoctorFunction) parseFixes(fixes []string) []Fix {
	var result []Fix
	for i, fix := range fixes {
		result = append(result, Fix{
			ID:          fmt.Sprintf("fix_%d", i+1),
			Description: fix,
			Status:      "available",
			Applied:     false,
		})
	}
	return result
}

func (f *DoctorFunction) calculateOverallHealth(checks []string) string {
	if len(checks) == 0 {
		return "unknown"
	}

	failCount := 0
	for _, check := range checks {
		if strings.Contains(strings.ToLower(check), "fail") || strings.Contains(strings.ToLower(check), "error") {
			failCount++
		}
	}

	if failCount == 0 {
		return "healthy"
	} else if failCount < len(checks)/3 {
		return "warning"
	} else {
		return "critical"
	}
}

func (f *DoctorFunction) calculateHealthFromIssues(issues []string) string {
	if len(issues) == 0 {
		return "healthy"
	}

	criticalCount := 0
	for _, issue := range issues {
		if strings.Contains(strings.ToLower(issue), "critical") {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		return "critical"
	} else if len(issues) > 3 {
		return "warning"
	} else {
		return "fair"
	}
}

func (f *DoctorFunction) generateHealthSummary(checks []HealthCheck) *HealthSummary {
	summary := &HealthSummary{
		TotalChecks: len(checks),
	}

	for _, check := range checks {
		switch strings.ToLower(check.Status) {
		case "pass", "ok", "healthy":
			summary.PassedChecks++
		case "fail", "error", "critical":
			summary.FailedChecks++
		case "warning", "warn":
			summary.WarningChecks++
		}
	}

	return summary
}

func (f *DoctorFunction) generateRecommendations(issues []Issue) []string {
	var recommendations []string

	if len(issues) == 0 {
		recommendations = append(recommendations, "System appears healthy - continue regular maintenance")
	}

	for _, issue := range issues {
		if issue.Severity == "critical" {
			recommendations = append(recommendations, fmt.Sprintf("Urgently address: %s", issue.Title))
		}
	}

	recommendations = append(recommendations, "Run regular system updates")
	recommendations = append(recommendations, "Monitor system logs for anomalies")

	return recommendations
}

func (f *DoctorFunction) parseHealthSummaryFromAgent(summary string) *HealthSummary {
	// Parse summary string from agent - implement based on agent response format
	return &HealthSummary{
		TotalChecks: 1,
	}
}

func (f *DoctorFunction) extractOverallHealth(scanResult string) string {
	if strings.Contains(strings.ToLower(scanResult), "healthy") {
		return "healthy"
	} else if strings.Contains(strings.ToLower(scanResult), "warning") {
		return "warning"
	} else if strings.Contains(strings.ToLower(scanResult), "critical") {
		return "critical"
	}
	return "unknown"
}

func (f *DoctorFunction) extractHealthChecks(scanResult string) []HealthCheck {
	// Extract health checks from scan result
	return []HealthCheck{}
}

func (f *DoctorFunction) extractIssues(scanResult string) []Issue {
	// Extract issues from scan result
	return []Issue{}
}

func (f *DoctorFunction) extractSummary(scanResult string) *HealthSummary {
	// Extract summary from scan result
	return &HealthSummary{}
}

func (f *DoctorFunction) extractTitle(issue string) string {
	// Extract title from issue description
	lines := strings.Split(issue, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return "Unknown Issue"
}

func (f *DoctorFunction) extractSeverity(issue string) string {
	lower := strings.ToLower(issue)
	if strings.Contains(lower, "critical") {
		return "critical"
	} else if strings.Contains(lower, "error") {
		return "error"
	} else if strings.Contains(lower, "warning") {
		return "warning"
	}
	return "info"
}

func (f *DoctorFunction) extractCategory(issue string) string {
	lower := strings.ToLower(issue)
	if strings.Contains(lower, "network") {
		return "network"
	} else if strings.Contains(lower, "storage") {
		return "storage"
	} else if strings.Contains(lower, "service") {
		return "services"
	} else if strings.Contains(lower, "package") {
		return "packages"
	}
	return "system"
}
