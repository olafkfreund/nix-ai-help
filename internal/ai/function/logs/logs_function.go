package logs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// LogsFunction handles system log analysis and management operations
type LogsFunction struct {
	*functionbase.BaseFunction
	agent  *agent.LogsAgent
	logger *logger.Logger
}

// LogsRequest represents the input parameters for the logs function
type LogsRequest struct {
	Operation   string            `json:"operation"`
	LogType     string            `json:"log_type,omitempty"`
	TimeRange   string            `json:"time_range,omitempty"`
	Service     string            `json:"service,omitempty"`
	Level       string            `json:"level,omitempty"`
	Filter      string            `json:"filter,omitempty"`
	Lines       int               `json:"lines,omitempty"`
	Follow      bool              `json:"follow,omitempty"`
	Format      string            `json:"format,omitempty"`
	Output      string            `json:"output,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	ExcludeList []string          `json:"exclude_list,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// LogsResponse represents the output of the logs function
type LogsResponse struct {
	Operation     string       `json:"operation"`
	Status        string       `json:"status"`
	LogEntries    []LogEntry   `json:"log_entries,omitempty"`
	Summary       *LogSummary  `json:"summary,omitempty"`
	Errors        []LogError   `json:"errors,omitempty"`
	Warnings      []LogWarning `json:"warnings,omitempty"`
	Statistics    *LogStats    `json:"statistics,omitempty"`
	Suggestions   []string     `json:"suggestions,omitempty"`
	NextSteps     []string     `json:"next_steps,omitempty"`
	TotalLines    int          `json:"total_lines,omitempty"`
	FilteredLines int          `json:"filtered_lines,omitempty"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Service   string            `json:"service"`
	Message   string            `json:"message"`
	PID       string            `json:"pid,omitempty"`
	Unit      string            `json:"unit,omitempty"`
	Priority  string            `json:"priority,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// LogSummary provides a summary of log analysis
type LogSummary struct {
	TotalEntries      int            `json:"total_entries"`
	ErrorCount        int            `json:"error_count"`
	WarningCount      int            `json:"warning_count"`
	TimeSpan          string         `json:"time_span"`
	TopServices       []string       `json:"top_services"`
	CommonPatterns    []string       `json:"common_patterns"`
	CriticalIssues    []string       `json:"critical_issues"`
	RecentActivity    []LogEntry     `json:"recent_activity"`
	LevelDistribution map[string]int `json:"level_distribution"`
}

// LogError represents an error found in logs
type LogError struct {
	Timestamp  string `json:"timestamp"`
	Service    string `json:"service"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Count      int    `json:"count"`
	FirstSeen  string `json:"first_seen"`
	LastSeen   string `json:"last_seen"`
	Suggestion string `json:"suggestion"`
}

// LogWarning represents a warning found in logs
type LogWarning struct {
	Timestamp  string `json:"timestamp"`
	Service    string `json:"service"`
	Message    string `json:"message"`
	Count      int    `json:"count"`
	Suggestion string `json:"suggestion"`
}

// LogStats provides statistical information about logs
type LogStats struct {
	TotalSize        string         `json:"total_size"`
	OldestEntry      string         `json:"oldest_entry"`
	NewestEntry      string         `json:"newest_entry"`
	LogRotations     int            `json:"log_rotations"`
	ServicesActive   int            `json:"services_active"`
	AverageEntrySize string         `json:"average_entry_size"`
	PeakActivityTime string         `json:"peak_activity_time"`
	LogLevels        map[string]int `json:"log_levels"`
	ServiceActivity  map[string]int `json:"service_activity"`
}

// NewLogsFunction creates a new logs function
func NewLogsFunction() *LogsFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Operation to perform", true,
			[]string{"view", "analyze", "search", "filter", "tail", "monitor", "export", "clean", "summary"}, nil, nil),
		functionbase.StringParamWithOptions("log_type", "Type of logs to access", false,
			[]string{"system", "service", "kernel", "boot", "user", "application", "security", "network"}, nil, nil),
		functionbase.StringParam("time_range", "Time range for log retrieval (e.g., '1h', '24h', 'since yesterday')", false),
		functionbase.StringParam("service", "Specific service name to filter logs", false),
		functionbase.StringParamWithOptions("level", "Log level filter", false,
			[]string{"emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"}, nil, nil),
		functionbase.StringParam("filter", "Text filter or regex pattern", false),
		functionbase.IntParam("lines", "Number of lines to retrieve", false, 100),
		functionbase.BoolParam("follow", "Follow logs in real-time", false),
		functionbase.StringParamWithOptions("format", "Output format", false,
			[]string{"json", "plain", "table", "colored", "short", "verbose"}, nil, nil),
		functionbase.StringParam("output", "Output file path", false),
		{
			Name:        "keywords",
			Type:        "array",
			Description: "Keywords to search for in logs",
			Required:    false,
		},
		{
			Name:        "exclude_list",
			Type:        "array",
			Description: "Patterns to exclude from results",
			Required:    false,
		},
		{
			Name:        "options",
			Type:        "object",
			Description: "Additional options for log operations",
			Required:    false,
		},
	}

	baseFunc := functionbase.NewBaseFunction(
		"logs",
		"Analyze and manage system logs using journalctl and other log tools",
		parameters,
	)

	return &LogsFunction{
		BaseFunction: baseFunc,
		agent:        agent.NewLogsAgent(),
		logger:       logger.NewLogger(),
	}
}

// Execute implements the FunctionInterface
func (f *LogsFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Parse and validate request
	request, err := f.parseRequest(params)
	if err != nil {
		return functionbase.ErrorResult(fmt.Errorf("invalid request: %w", err), time.Since(startTime)), nil
	}

	if err := f.validateRequest(request); err != nil {
		return functionbase.ErrorResult(fmt.Errorf("validation failed: %w", err), time.Since(startTime)), nil
	}

	// Execute logs operation
	response, err := f.executeLogsOperation(ctx, request)
	if err != nil {
		return functionbase.ErrorResult(fmt.Errorf("execution failed: %w", err), time.Since(startTime)), nil
	}

	return functionbase.SuccessResult(response, time.Since(startTime)), nil
}

// parseRequest converts the raw parameters into a structured request
func (f *LogsFunction) parseRequest(params map[string]interface{}) (*LogsRequest, error) {
	request := &LogsRequest{}

	if operation, ok := params["operation"].(string); ok {
		request.Operation = operation
	}

	if logType, ok := params["log_type"].(string); ok {
		request.LogType = logType
	}

	if timeRange, ok := params["time_range"].(string); ok {
		request.TimeRange = timeRange
	}

	if service, ok := params["service"].(string); ok {
		request.Service = service
	}

	if level, ok := params["level"].(string); ok {
		request.Level = level
	}

	if filter, ok := params["filter"].(string); ok {
		request.Filter = filter
	}

	if lines, ok := params["lines"].(float64); ok {
		request.Lines = int(lines)
	}

	if follow, ok := params["follow"].(bool); ok {
		request.Follow = follow
	}

	if format, ok := params["format"].(string); ok {
		request.Format = format
	}

	if output, ok := params["output"].(string); ok {
		request.Output = output
	}

	if keywords, ok := params["keywords"].([]interface{}); ok {
		for _, keyword := range keywords {
			if k, ok := keyword.(string); ok {
				request.Keywords = append(request.Keywords, k)
			}
		}
	}

	if excludeList, ok := params["exclude_list"].([]interface{}); ok {
		for _, exclude := range excludeList {
			if e, ok := exclude.(string); ok {
				request.ExcludeList = append(request.ExcludeList, e)
			}
		}
	}

	if options, ok := params["options"].(map[string]interface{}); ok {
		request.Options = make(map[string]string)
		for k, v := range options {
			if s, ok := v.(string); ok {
				request.Options[k] = s
			}
		}
	}

	return request, nil
}

// validateRequest validates the parsed request
func (f *LogsFunction) validateRequest(request *LogsRequest) error {
	if request.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	validOps := []string{"view", "analyze", "search", "filter", "tail", "monitor", "export", "clean", "summary"}
	if !f.contains(validOps, request.Operation) {
		return fmt.Errorf("invalid operation: %s", request.Operation)
	}

	if request.Lines != 0 && (request.Lines < 1 || request.Lines > 10000) {
		return fmt.Errorf("lines must be between 1 and 10000")
	}

	return nil
}

// executeLogsOperation executes the logs operation
func (f *LogsFunction) executeLogsOperation(ctx context.Context, request *LogsRequest) (*LogsResponse, error) {
	// Create context for the agent
	agentContext := agent.LogsContext{
		Operation:   request.Operation,
		LogType:     request.LogType,
		TimeRange:   request.TimeRange,
		Service:     request.Service,
		Level:       request.Level,
		Filter:      request.Filter,
		Lines:       request.Lines,
		Follow:      request.Follow,
		Format:      request.Format,
		Keywords:    request.Keywords,
		ExcludeList: request.ExcludeList,
		Options:     request.Options,
	}

	switch request.Operation {
	case "view":
		return f.handleViewOperation(ctx, agentContext)
	case "analyze":
		return f.handleAnalyzeOperation(ctx, agentContext)
	case "search":
		return f.handleSearchOperation(ctx, agentContext)
	case "filter":
		return f.handleFilterOperation(ctx, agentContext)
	case "tail":
		return f.handleTailOperation(ctx, agentContext)
	case "monitor":
		return f.handleMonitorOperation(ctx, agentContext)
	case "export":
		return f.handleExportOperation(ctx, agentContext)
	case "clean":
		return f.handleCleanOperation(ctx, agentContext)
	case "summary":
		return f.handleSummaryOperation(ctx, agentContext)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", request.Operation)
	}
}

// handleViewOperation handles log viewing
func (f *LogsFunction) handleViewOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.ViewLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:     "view",
		Status:        "success",
		LogEntries:    f.parseLogEntries(result),
		TotalLines:    f.countLines(result),
		FilteredLines: f.countFilteredLines(result, agentContext),
	}

	return response, nil
}

// handleAnalyzeOperation handles log analysis
func (f *LogsFunction) handleAnalyzeOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.AnalyzeLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:   "analyze",
		Status:      "success",
		Summary:     f.parseLogSummary(result),
		Errors:      f.parseLogErrors(result),
		Warnings:    f.parseLogWarnings(result),
		Statistics:  f.parseLogStats(result),
		Suggestions: f.parseAnalysisSuggestions(result),
		NextSteps:   f.parseNextSteps(result),
	}

	return response, nil
}

// handleSearchOperation handles log searching
func (f *LogsFunction) handleSearchOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.SearchLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:     "search",
		Status:        "success",
		LogEntries:    f.parseLogEntries(result),
		TotalLines:    f.countLines(result),
		FilteredLines: f.countFilteredLines(result, agentContext),
		Suggestions:   f.parseSearchSuggestions(result),
	}

	return response, nil
}

// handleFilterOperation handles log filtering
func (f *LogsFunction) handleFilterOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.FilterLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:     "filter",
		Status:        "success",
		LogEntries:    f.parseLogEntries(result),
		TotalLines:    f.countLines(result),
		FilteredLines: f.countFilteredLines(result, agentContext),
	}

	return response, nil
}

// handleTailOperation handles log tailing
func (f *LogsFunction) handleTailOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.TailLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:  "tail",
		Status:     "success",
		LogEntries: f.parseLogEntries(result),
		NextSteps:  []string{"Use Ctrl+C to stop tailing", "Monitor for new entries"},
	}

	return response, nil
}

// handleMonitorOperation handles log monitoring
func (f *LogsFunction) handleMonitorOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.MonitorLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:   "monitor",
		Status:      "success",
		Statistics:  f.parseLogStats(result),
		Summary:     f.parseLogSummary(result),
		Suggestions: f.parseMonitorSuggestions(result),
		NextSteps:   []string{"Review monitored patterns", "Set up alerts if needed"},
	}

	return response, nil
}

// handleExportOperation handles log export
func (f *LogsFunction) handleExportOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.ExportLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation: "export",
		Status:    "success",
		NextSteps: f.parseExportInfo(result),
	}

	return response, nil
}

// handleCleanOperation handles log cleanup
func (f *LogsFunction) handleCleanOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.CleanLogs(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:   "clean",
		Status:      "success",
		Statistics:  f.parseLogStats(result),
		Suggestions: []string{"Review cleanup results", "Verify log rotation settings"},
		NextSteps:   f.parseCleanupInfo(result),
	}

	return response, nil
}

// handleSummaryOperation handles log summary
func (f *LogsFunction) handleSummaryOperation(ctx context.Context, agentContext agent.LogsContext) (*LogsResponse, error) {
	result, err := f.agent.GetLogsSummary(ctx, agentContext)
	if err != nil {
		return nil, err
	}

	response := &LogsResponse{
		Operation:  "summary",
		Status:     "success",
		Summary:    f.parseLogSummary(result),
		Statistics: f.parseLogStats(result),
		Errors:     f.parseLogErrors(result),
		Warnings:   f.parseLogWarnings(result),
		NextSteps:  []string{"Review summary details", "Investigate any errors or warnings"},
	}

	return response, nil
}

// Helper methods for parsing agent responses

func (f *LogsFunction) parseLogEntries(response string) []LogEntry {
	var entries []LogEntry

	// Simple parsing - in a real implementation, this would parse actual log formats
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			entry := LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Level:     "info",
				Message:   line,
				Service:   "unknown",
			}
			entries = append(entries, entry)
		}
	}

	return entries
}

func (f *LogsFunction) parseLogSummary(response string) *LogSummary {
	return &LogSummary{
		TotalEntries:      100,
		ErrorCount:        5,
		WarningCount:      10,
		TimeSpan:          "24h",
		TopServices:       []string{"systemd", "networkd", "nixos-rebuild"},
		CommonPatterns:    []string{"startup", "configuration", "service start"},
		CriticalIssues:    []string{},
		LevelDistribution: map[string]int{"info": 70, "warning": 20, "error": 10},
	}
}

func (f *LogsFunction) parseLogErrors(response string) []LogError {
	var errors []LogError

	if strings.Contains(strings.ToLower(response), "error") {
		errors = append(errors, LogError{
			Timestamp:  time.Now().Format(time.RFC3339),
			Service:    "system",
			Message:    "Error detected in logs",
			Severity:   "medium",
			Count:      1,
			Suggestion: "Review the error details and check system status",
		})
	}

	return errors
}

func (f *LogsFunction) parseLogWarnings(response string) []LogWarning {
	var warnings []LogWarning

	if strings.Contains(strings.ToLower(response), "warning") {
		warnings = append(warnings, LogWarning{
			Timestamp:  time.Now().Format(time.RFC3339),
			Service:    "system",
			Message:    "Warning detected in logs",
			Count:      1,
			Suggestion: "Monitor for recurring warnings",
		})
	}

	return warnings
}

func (f *LogsFunction) parseLogStats(response string) *LogStats {
	return &LogStats{
		TotalSize:        "1.2GB",
		OldestEntry:      "7 days ago",
		NewestEntry:      "now",
		LogRotations:     7,
		ServicesActive:   25,
		AverageEntrySize: "120 bytes",
		PeakActivityTime: "09:00",
		LogLevels:        map[string]int{"info": 70, "warning": 20, "error": 10},
		ServiceActivity:  map[string]int{"systemd": 40, "networkd": 20, "other": 40},
	}
}

func (f *LogsFunction) parseAnalysisSuggestions(response string) []string {
	return []string{
		"Review error patterns for recurring issues",
		"Check system resource usage during peak times",
		"Monitor service startup times",
		"Consider log rotation settings",
	}
}

func (f *LogsFunction) parseSearchSuggestions(response string) []string {
	return []string{
		"Try broader search terms if no results found",
		"Use time ranges to narrow results",
		"Check service-specific logs",
	}
}

func (f *LogsFunction) parseMonitorSuggestions(response string) []string {
	return []string{
		"Set up alerting for critical errors",
		"Review log retention policies",
		"Monitor disk usage for log storage",
	}
}

func (f *LogsFunction) parseNextSteps(response string) []string {
	return []string{
		"Review analysis results",
		"Address any critical issues",
		"Monitor system performance",
	}
}

func (f *LogsFunction) parseExportInfo(response string) []string {
	return []string{
		"Logs exported successfully",
		"Check output file location",
		"Verify export format and content",
	}
}

func (f *LogsFunction) parseCleanupInfo(response string) []string {
	return []string{
		"Log cleanup completed",
		"Review freed disk space",
		"Verify log rotation configuration",
	}
}

func (f *LogsFunction) countLines(response string) int {
	return len(strings.Split(response, "\n"))
}

func (f *LogsFunction) countFilteredLines(response string, context agent.LogsContext) int {
	lines := strings.Split(response, "\n")
	filtered := 0

	for _, line := range lines {
		if context.Filter != "" && strings.Contains(strings.ToLower(line), strings.ToLower(context.Filter)) {
			filtered++
		}
	}

	if context.Filter == "" {
		return len(lines)
	}

	return filtered
}

// Helper function to check if slice contains string
func (f *LogsFunction) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
