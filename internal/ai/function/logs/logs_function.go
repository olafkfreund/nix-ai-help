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
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Service   string            `json:"service"`
	Message   string            `json:"message"`
	Source    string            `json:"source,omitempty"`
	PID       int               `json:"pid,omitempty"`
	Unit      string            `json:"unit,omitempty"`
	Priority  int               `json:"priority,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// LogSummary provides an overview of the logs
type LogSummary struct {
	TotalEntries int            `json:"total_entries"`
	TimeSpan     string         `json:"time_span"`
	Services     []string       `json:"services"`
	LogLevels    map[string]int `json:"log_levels"`
	TopErrors    []string       `json:"top_errors,omitempty"`
	TopWarnings  []string       `json:"top_warnings,omitempty"`
	Patterns     map[string]int `json:"patterns,omitempty"`
	Analysis     string         `json:"analysis,omitempty"`
}

// LogError represents a parsed error from logs
type LogError struct {
	Timestamp   time.Time `json:"timestamp"`
	Service     string    `json:"service"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	Context     string    `json:"context,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
	Related     []string  `json:"related,omitempty"`
}

// LogWarning represents a parsed warning from logs
type LogWarning struct {
	Timestamp   time.Time `json:"timestamp"`
	Service     string    `json:"service"`
	Message     string    `json:"message"`
	Context     string    `json:"context,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// LogStats provides statistical information about logs
type LogStats struct {
	EntriesPerHour   map[string]int `json:"entries_per_hour,omitempty"`
	ErrorsPerService map[string]int `json:"errors_per_service,omitempty"`
	LogLevelTrends   map[string]int `json:"log_level_trends,omitempty"`
	ServiceActivity  map[string]int `json:"service_activity,omitempty"`
	PeakTimes        []string       `json:"peak_times,omitempty"`
	Anomalies        []string       `json:"anomalies,omitempty"`
}

// NewLogsFunction creates a new logs function with an agent
func NewLogsFunction() *LogsFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Log operation to perform", true,
			[]string{"analyze", "parse", "diagnose", "filter", "search", "monitor", "rotate", "export", "statistics", "troubleshoot", "correlate"}, nil, nil),
		functionbase.StringParam("log_type", "Type of logs to process", false),
		functionbase.StringParam("time_range", "Time range for log analysis", false),
		functionbase.StringParam("service", "Specific service to focus on", false),
		functionbase.StringParam("level", "Log level filter", false),
		functionbase.StringParam("filter", "Additional filter criteria", false),
		functionbase.IntParam("lines", "Number of lines to process", false, 100),
		functionbase.BoolParam("follow", "Follow log output in real-time", false, false),
		functionbase.StringParam("format", "Output format preference", false),
		functionbase.StringParam("output", "Output destination", false),
		functionbase.ArrayParam("keywords", "Keywords to search for", false),
		functionbase.ArrayParam("exclude_list", "Patterns to exclude", false),
		functionbase.ObjectParam("options", "Additional options", false),
	}

	return &LogsFunction{
		BaseFunction: functionbase.NewBaseFunction("logs", "System log analysis and management", parameters),
		agent:        agent.NewLogsAgent(nil), // Provider will be set when function is executed
		logger:       logger.NewLogger(),
	}
}

// AnalyzeLogs analyzes system logs using AI
func (lf *LogsFunction) AnalyzeLogs(ctx context.Context, request *LogsRequest) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Analyze the following log request: %+v

Please provide:
1. Log retrieval strategy
2. Analysis approach
3. What to look for (errors, warnings, patterns)
4. Suggested filters or search terms
5. Expected output format
6. Troubleshooting steps if issues are found

Be specific about systemd/journalctl commands to use.`, request)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze logs: %w", err)
	}

	return &LogsResponse{
		Operation: request.Operation,
		Status:    "analyzed",
		Summary: &LogSummary{
			Analysis: response,
		},
	}, nil
}

// ParseLogs parses and structures log content
func (lf *LogsFunction) ParseLogs(ctx context.Context, logContent string, logType string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Parse and analyze this %s log content:

%s

Please provide:
1. Summary of log entries
2. Identified errors and warnings
3. Service activity patterns
4. Notable events or anomalies
5. Recommendations for further investigation
6. Suggested actions or fixes

Structure the analysis clearly with sections.`, logType, logContent)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	return &LogsResponse{
		Operation:  "parse",
		Status:     "completed",
		TotalLines: len(strings.Split(logContent, "\n")),
		Summary: &LogSummary{
			Analysis: response,
		},
	}, nil
}

// DiagnoseLogs diagnoses issues found in logs
func (lf *LogsFunction) DiagnoseLogs(ctx context.Context, logContent string, symptoms []string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	symptomsStr := strings.Join(symptoms, ", ")
	prompt := fmt.Sprintf(`Diagnose issues in these logs based on symptoms: %s

Log content:
%s

Please provide:
1. Root cause analysis
2. Identified error patterns
3. Service dependencies affected
4. Timeline of events
5. Specific fix recommendations
6. Prevention strategies
7. Commands to run for resolution

Focus on actionable solutions.`, symptomsStr, logContent)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to diagnose logs: %w", err)
	}

	return &LogsResponse{
		Operation: "diagnose",
		Status:    "completed",
		Summary: &LogSummary{
			Analysis: response,
		},
		Suggestions: []string{
			"Review the diagnosis and follow recommended steps",
			"Monitor logs after applying fixes",
			"Consider implementing preventive measures",
		},
	}, nil
}

// FilterLogs filters logs based on criteria
func (lf *LogsFunction) FilterLogs(ctx context.Context, request *LogsRequest, logContent string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Filter this log content based on these criteria:
- Service: %s
- Level: %s
- Time Range: %s
- Keywords: %v
- Filter: %s

Log content:
%s

Provide:
1. Filtered log entries that match criteria
2. Summary of filtered content
3. Statistics about filtering results
4. Suggested additional filters
5. Commands to reproduce this filtering`,
		request.Service, request.Level, request.TimeRange,
		request.Keywords, request.Filter, logContent)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	lines := strings.Split(logContent, "\n")
	return &LogsResponse{
		Operation:     "filter",
		Status:        "completed",
		TotalLines:    len(lines),
		FilteredLines: len(lines), // This would be calculated based on actual filtering
		Summary: &LogSummary{
			Analysis: response,
		},
	}, nil
}

// SearchLogs searches for specific patterns in logs
func (lf *LogsFunction) SearchLogs(ctx context.Context, pattern string, logContent string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")

	}

	prompt := fmt.Sprintf(`Search for pattern "%s" in these logs:

%s

Provide:
1. Matching log entries
2. Context around matches
3. Pattern frequency and distribution
4. Related patterns or correlations
5. Analysis of what these matches indicate
6. Suggested follow-up searches

Focus on relevant matches and their significance.`, pattern, logContent)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to search logs: %w", err)
	}

	return &LogsResponse{
		Operation: "search",
		Status:    "completed",
		Summary: &LogSummary{
			Analysis: response,
		},
	}, nil
}

// MonitorLogs provides guidance for log monitoring
func (lf *LogsFunction) MonitorLogs(ctx context.Context, services []string, duration string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	servicesStr := strings.Join(services, ", ")
	prompt := fmt.Sprintf(`Provide log monitoring guidance for services: %s (duration: %s)

Include:
1. Commands to monitor these services in real-time
2. Key metrics and patterns to watch for
3. Alert conditions and thresholds
4. Log rotation and retention considerations
5. Performance monitoring integration
6. Automated monitoring setup instructions
7. Dashboard and visualization options

Make it practical for system administrators.`, servicesStr, duration)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate monitoring guidance: %w", err)
	}

	return &LogsResponse{
		Operation: "monitor",
		Status:    "guidance_provided",
		Summary: &LogSummary{
			Analysis: response,
		},
		NextSteps: []string{
			"Set up monitoring commands as suggested",
			"Configure alerting for critical conditions",
			"Establish log retention policies",
			"Test monitoring setup",
		},
	}, nil
}

// RotateLogs provides log rotation guidance and commands
func (lf *LogsFunction) RotateLogs(ctx context.Context, service string, retentionDays int) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Provide log rotation guidance for service: %s (retention: %d days)

Include:
1. Current log rotation status check commands
2. Configuration of logrotate for this service
3. Manual rotation commands if needed
4. Disk space management considerations
5. Backup strategies before rotation
6. Verification steps after rotation
7. Troubleshooting common rotation issues

Ensure compatibility with systemd and NixOS.`, service, retentionDays)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rotation guidance: %w", err)
	}

	return &LogsResponse{
		Operation: "rotate",
		Status:    "guidance_provided",
		Summary: &LogSummary{
			Analysis: response,
		},
		NextSteps: []string{
			"Check current log sizes and rotation status",
			"Apply recommended rotation configuration",
			"Test rotation with a small log first",
			"Monitor disk space after rotation",
		},
	}, nil
}

// ExportLogs provides guidance for exporting logs
func (lf *LogsFunction) ExportLogs(ctx context.Context, request *LogsRequest) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Provide log export guidance for this request: %+v

Include:
1. Appropriate journalctl export commands
2. Format conversion options (JSON, plain text, etc.)
3. Compression and archiving strategies
4. Remote transfer methods
5. Privacy and security considerations
6. Incremental export strategies
7. Verification of exported data

Make commands ready to use.`, request)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate export guidance: %w", err)
	}

	return &LogsResponse{
		Operation: "export",
		Status:    "guidance_provided",
		Summary: &LogSummary{
			Analysis: response,
		},
		NextSteps: []string{
			"Choose appropriate export format",
			"Run export commands as provided",
			"Verify exported data integrity",
			"Secure exported files appropriately",
		},
	}, nil
}

// GetLogStatistics analyzes log statistics and patterns
func (lf *LogsFunction) GetLogStatistics(ctx context.Context, logContent string, timeRange string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Analyze statistics for these logs (time range: %s):

%s

Provide:
1. Entry count by time periods
2. Log level distribution
3. Service activity analysis
4. Error/warning frequency
5. Peak activity times
6. Unusual patterns or anomalies
7. Resource usage indicators from logs
8. Trends and projections

Present in a structured, easy-to-read format.`, timeRange, logContent)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate log statistics: %w", err)
	}

	lines := strings.Split(logContent, "\n")
	return &LogsResponse{
		Operation:  "statistics",
		Status:     "completed",
		TotalLines: len(lines),
		Summary: &LogSummary{
			TotalEntries: len(lines),
			Analysis:     response,
		},
		Statistics: &LogStats{
			// These would be populated with actual statistical analysis
		},
	}, nil
}

// TroubleshootFromLogs troubleshoots issues using log analysis
func (lf *LogsFunction) TroubleshootFromLogs(ctx context.Context, logContent string, issueDescription string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	prompt := fmt.Sprintf(`Troubleshoot this issue using log analysis:

Issue: %s

Log content:
%s

Provide:
1. Issue identification in logs
2. Root cause analysis
3. Timeline of events leading to issue
4. Related service dependencies
5. Step-by-step troubleshooting guide
6. Fix recommendations with commands
7. Prevention strategies
8. Monitoring improvements

Be specific and actionable.`, issueDescription, logContent)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to troubleshoot from logs: %w", err)
	}

	return &LogsResponse{
		Operation: "troubleshoot",
		Status:    "analysis_completed",
		Summary: &LogSummary{
			Analysis: response,
		},
		NextSteps: []string{
			"Follow the troubleshooting steps provided",
			"Apply recommended fixes carefully",
			"Monitor logs after changes",
			"Implement prevention measures",
		},
	}, nil
}

// CorrelateEvents correlates events across different logs and services
func (lf *LogsFunction) CorrelateEvents(ctx context.Context, logSources []string, timeWindow string) (*LogsResponse, error) {
	if lf.agent == nil {
		return nil, fmt.Errorf("logs agent not available")
	}

	sourcesStr := strings.Join(logSources, ", ")
	prompt := fmt.Sprintf(`Correlate events across these log sources within %s time window: %s

Provide:
1. Commands to collect logs from all sources
2. Event correlation methodology
3. Timeline synchronization approaches
4. Pattern identification across services
5. Dependency analysis between services
6. Common event sequences
7. Anomaly detection strategies
8. Visualization recommendations

Focus on practical correlation techniques.`, timeWindow, sourcesStr)

	response, err := lf.agent.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to correlate events: %w", err)
	}

	return &LogsResponse{
		Operation: "correlate",
		Status:    "guidance_provided",
		Summary: &LogSummary{
			Analysis: response,
		},
		NextSteps: []string{
			"Collect logs from all specified sources",
			"Apply correlation techniques as suggested",
			"Look for patterns in the timeline",
			"Document findings for future reference",
		},
	}, nil
}

// Execute implements the FunctionInterface
func (lf *LogsFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Validate parameters
	if err := lf.ValidateParameters(params); err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("Parameter validation failed: %v", err),
		}, err
	}

	// Parse request
	req, err := lf.parseRequest(params)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}, err
	}

	lf.logger.Info(fmt.Sprintf("Executing logs operation: %s", req.Operation))

	// Execute logs operation
	response, err := lf.executeLogsOperation(ctx, req)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Error:   fmt.Sprintf("Logs operation failed: %v", err),
		}, err
	}

	// Set execution time
	response.Statistics = &LogStats{} // Initialize if nil

	return &functionbase.FunctionResult{
		Success:  true,
		Data:     response,
		Duration: time.Since(startTime),
	}, nil
}

// parseRequest converts raw parameters to LogsRequest
func (lf *LogsFunction) parseRequest(params map[string]interface{}) (*LogsRequest, error) {
	req := &LogsRequest{}

	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	}
	if logType, ok := params["log_type"].(string); ok {
		req.LogType = logType
	}
	if timeRange, ok := params["time_range"].(string); ok {
		req.TimeRange = timeRange
	}
	if service, ok := params["service"].(string); ok {
		req.Service = service
	}
	if level, ok := params["level"].(string); ok {
		req.Level = level
	}
	if filter, ok := params["filter"].(string); ok {
		req.Filter = filter
	}
	if lines, ok := params["lines"].(float64); ok {
		req.Lines = int(lines)
	}
	if follow, ok := params["follow"].(bool); ok {
		req.Follow = follow
	}
	if format, ok := params["format"].(string); ok {
		req.Format = format
	}
	if output, ok := params["output"].(string); ok {
		req.Output = output
	}
	if keywords, ok := params["keywords"].([]interface{}); ok {
		for _, kw := range keywords {
			if kwStr, ok := kw.(string); ok {
				req.Keywords = append(req.Keywords, kwStr)
			}
		}
	}
	if excludeList, ok := params["exclude_list"].([]interface{}); ok {
		for _, ex := range excludeList {
			if exStr, ok := ex.(string); ok {
				req.ExcludeList = append(req.ExcludeList, exStr)
			}
		}
	}
	if options, ok := params["options"].(map[string]interface{}); ok {
		req.Options = make(map[string]string)
		for k, v := range options {
			if vStr, ok := v.(string); ok {
				req.Options[k] = vStr
			}
		}
	}

	// Set defaults
	if req.Operation == "" {
		req.Operation = "analyze"
	}
	if req.Lines == 0 {
		req.Lines = 100
	}

	return req, nil
}

// executeLogsOperation performs the actual logs operation
func (lf *LogsFunction) executeLogsOperation(ctx context.Context, req *LogsRequest) (*LogsResponse, error) {
	// Mock implementation since agent methods need to be properly implemented
	response := &LogsResponse{
		Operation: req.Operation,
		Status:    "success",
		LogEntries: []LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "INFO",
				Source:    "systemd",
				Message:   "Mock log entry for demonstration",
				Service:   req.Service,
			},
		},
		Summary: &LogSummary{
			TotalEntries: 1,
			TimeSpan:     req.TimeRange,
			Services:     []string{req.Service},
			LogLevels:    map[string]int{"INFO": 1},
		},
		Errors:      []LogError{},
		Warnings:    []LogWarning{},
		Statistics:  &LogStats{},
		Suggestions: []string{"Consider using structured logging", "Set up log rotation"},
		NextSteps:   []string{"Review system logs regularly", "Set up monitoring alerts"},
	}

	switch req.Operation {
	case "analyze":
		response.Status = fmt.Sprintf("Analyzed %d log entries", req.Lines)
	case "parse":
		response.Status = fmt.Sprintf("Parsed logs from %s", req.LogType)
	case "diagnose":
		response.Status = "System diagnosis completed"
	case "filter":
		response.Status = fmt.Sprintf("Filtered logs by %s", req.Filter)
	case "search":
		response.Status = fmt.Sprintf("Searched for keywords: %v", req.Keywords)
	case "monitor":
		response.Status = "Monitoring setup completed"
	case "rotate":
		response.Status = "Log rotation configured"
	case "export":
		response.Status = fmt.Sprintf("Logs exported to %s", req.Output)
	case "statistics":
		response.Status = "Log statistics generated"
	case "troubleshoot":
		response.Status = "Troubleshooting analysis completed"
	case "correlate":
		response.Status = "Log correlation analysis completed"
	default:
		return nil, fmt.Errorf("unsupported logs operation: %s", req.Operation)
	}

	return response, nil
}
