package gc

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// GcFunction handles Nix garbage collection operations
type GcFunction struct {
	*functionbase.BaseFunction
	agent  *agent.GcAgent
	logger *logger.Logger
}

// GcRequest represents the input parameters for the gc function
type GcRequest struct {
	Context         string            `json:"context"`
	Operation       string            `json:"operation,omitempty"`
	DryRun          bool              `json:"dry_run,omitempty"`
	MaxAge          string            `json:"max_age,omitempty"`
	MaxSize         string            `json:"max_size,omitempty"`
	KeepOutputs     bool              `json:"keep_outputs,omitempty"`
	KeepDerivations bool              `json:"keep_derivations,omitempty"`
	Verbose         bool              `json:"verbose,omitempty"`
	Force           bool              `json:"force,omitempty"`
	Options         map[string]string `json:"options,omitempty"`
}

// GcResponse represents the output of the gc function
type GcResponse struct {
	Context         string        `json:"context"`
	Status          string        `json:"status"`
	Operation       string        `json:"operation"`
	FreedSpace      int64         `json:"freed_space,omitempty"`
	DeletedPaths    int           `json:"deleted_paths,omitempty"`
	RemainingPaths  int           `json:"remaining_paths,omitempty"`
	StoreSize       int64         `json:"store_size,omitempty"`
	Details         []GcDetail    `json:"details,omitempty"`
	Recommendations []string      `json:"recommendations,omitempty"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	ExecutionTime   time.Duration `json:"execution_time,omitempty"`
}

// GcDetail represents details about garbage collection results
type GcDetail struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size,omitempty"`
	LastAccessed time.Time `json:"last_accessed,omitempty"`
	Action       string    `json:"action"` // deleted, kept, skipped
	Reason       string    `json:"reason,omitempty"`
}

// NewGcFunction creates a new gc function instance
func NewGcFunction() *GcFunction {
	return &GcFunction{
		BaseFunction: &functionbase.BaseFunction{
			FuncName:    "gc",
			FuncDesc:    "Manage Nix store garbage collection and cleanup operations",
			FuncVersion: "1.0.0",
		},
		agent:  agent.NewGcAgent(),
		logger: logger.NewLogger(),
	}
}

// Name returns the function name
func (f *GcFunction) Name() string {
	return f.FuncName
}

// Description returns the function description
func (f *GcFunction) Description() string {
	return f.FuncDesc
}

// Version returns the function version
func (f *GcFunction) Version() string {
	return f.FuncVersion
}

// Parameters returns the function parameter schema
func (f *GcFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"context": map[string]interface{}{
				"type":        "string",
				"description": "The context or reason for the garbage collection operation",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The garbage collection operation to perform",
				"enum":        []string{"collect", "list", "status", "optimize", "clean", "analyze"},
				"default":     "collect",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to perform a dry run without actually deleting anything",
				"default":     false,
			},
			"max_age": map[string]interface{}{
				"type":        "string",
				"description": "Maximum age of paths to keep (e.g., '30d', '1w', '6h')",
			},
			"max_size": map[string]interface{}{
				"type":        "string",
				"description": "Maximum store size to maintain (e.g., '10GB', '1TB')",
			},
			"keep_outputs": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to keep build outputs",
				"default":     false,
			},
			"keep_derivations": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to keep derivations",
				"default":     false,
			},
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to provide verbose output",
				"default":     false,
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to force garbage collection without confirmation",
				"default":     false,
			},
			"options": map[string]interface{}{
				"type":        "object",
				"description": "Additional garbage collection options",
			},
		},
		"required": []string{"context"},
	}
}

// Execute runs the gc function with the given parameters
func (f *GcFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Parse the request
	var req GcRequest
	if err := f.ParseParams(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Set defaults
	if req.Operation == "" {
		req.Operation = "collect"
	}

	f.logger.Info(fmt.Sprintf("Executing garbage collection operation: %s", req.Operation))

	// Execute the garbage collection operation
	response, err := f.executeGcOperation(ctx, &req)
	if err != nil {
		return &functionbase.FunctionResult{
			Success: false,
			Data: GcResponse{
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

// executeGcOperation performs the actual garbage collection operation
func (f *GcFunction) executeGcOperation(ctx context.Context, req *GcRequest) (*GcResponse, error) {
	response := &GcResponse{
		Context:   req.Context,
		Operation: req.Operation,
		Status:    "success",
		Details:   []GcDetail{},
	}

	switch req.Operation {
	case "collect":
		return f.performCollection(ctx, req, response)
	case "list":
		return f.listGarbage(ctx, req, response)
	case "status":
		return f.getStatus(ctx, req, response)
	case "optimize":
		return f.optimizeStore(ctx, req, response)
	case "clean":
		return f.cleanStore(ctx, req, response)
	case "analyze":
		return f.analyzeStore(ctx, req, response)
	default:
		return nil, fmt.Errorf("unsupported gc operation: %s", req.Operation)
	}
}

// performCollection performs garbage collection
func (f *GcFunction) performCollection(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Starting garbage collection")

	// Get current store status before collection
	beforeStatus, err := f.agent.GetStoreStatus(ctx)
	if err != nil {
		f.logger.Error(fmt.Sprintf("Failed to get store status: %v", err))
	} else {
		response.StoreSize = beforeStatus.TotalSize
	}

	// Perform garbage collection
	result, err := f.agent.Collect(ctx, &agent.GcOptions{
		DryRun:          req.DryRun,
		MaxAge:          req.MaxAge,
		MaxSize:         req.MaxSize,
		KeepOutputs:     req.KeepOutputs,
		KeepDerivations: req.KeepDerivations,
		Verbose:         req.Verbose,
		Force:           req.Force,
	})
	if err != nil {
		return nil, fmt.Errorf("garbage collection failed: %w", err)
	}

	// Update response with results
	response.FreedSpace = result.FreedSpace
	response.DeletedPaths = result.DeletedPaths
	response.RemainingPaths = result.RemainingPaths

	// Convert details
	for _, detail := range result.Details {
		response.Details = append(response.Details, GcDetail{
			Path:         detail.Path,
			Size:         detail.Size,
			LastAccessed: detail.LastAccessed,
			Action:       detail.Action,
			Reason:       detail.Reason,
		})
	}

	// Generate recommendations
	response.Recommendations = f.generateRecommendations(result)

	if req.DryRun {
		f.logger.Info(fmt.Sprintf("Dry run completed: would free %d bytes from %d paths", result.FreedSpace, result.DeletedPaths))
	} else {
		f.logger.Info(fmt.Sprintf("Garbage collection completed: freed %d bytes from %d paths", result.FreedSpace, result.DeletedPaths))
	}

	return response, nil
}

// listGarbage lists garbage without collecting it
func (f *GcFunction) listGarbage(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Listing garbage in Nix store")

	garbage, err := f.agent.ListGarbage(ctx, &agent.GcOptions{
		MaxAge:          req.MaxAge,
		MaxSize:         req.MaxSize,
		KeepOutputs:     req.KeepOutputs,
		KeepDerivations: req.KeepDerivations,
		Verbose:         req.Verbose,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list garbage: %w", err)
	}

	// Convert garbage list to details
	var totalSize int64
	for _, item := range garbage {
		response.Details = append(response.Details, GcDetail{
			Path:         item.Path,
			Size:         item.Size,
			LastAccessed: item.LastAccessed,
			Action:       "can_delete",
			Reason:       item.Reason,
		})
		totalSize += item.Size
	}

	response.FreedSpace = totalSize
	response.DeletedPaths = len(garbage)

	f.logger.Info(fmt.Sprintf("Found %d garbage paths totaling %d bytes", len(garbage), totalSize))

	return response, nil
}

// getStatus gets the current Nix store status
func (f *GcFunction) getStatus(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Getting Nix store status")

	status, err := f.agent.GetStoreStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get store status: %w", err)
	}

	response.StoreSize = status.TotalSize
	response.RemainingPaths = status.TotalPaths
	response.Details = []GcDetail{
		{
			Path:   "/nix/store",
			Size:   status.TotalSize,
			Action: "status",
			Reason: fmt.Sprintf("Total store size: %d bytes, %d paths", status.TotalSize, status.TotalPaths),
		},
	}

	// Add breakdown by category if available
	if status.Categories != nil {
		for category, info := range status.Categories {
			response.Details = append(response.Details, GcDetail{
				Path:   category,
				Size:   info.Size,
				Action: "info",
				Reason: fmt.Sprintf("%s: %d paths", category, info.Count),
			})
		}
	}

	// Generate recommendations based on status
	response.Recommendations = f.generateStatusRecommendations(status)

	f.logger.Info(fmt.Sprintf("Store status: %d bytes in %d paths", status.TotalSize, status.TotalPaths))

	return response, nil
}

// optimizeStore optimizes the Nix store
func (f *GcFunction) optimizeStore(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Optimizing Nix store")

	result, err := f.agent.OptimizeStore(ctx, req.Force)
	if err != nil {
		return nil, fmt.Errorf("store optimization failed: %w", err)
	}

	response.FreedSpace = result.SpaceSaved
	response.Details = []GcDetail{
		{
			Path:   "/nix/store",
			Size:   result.SpaceSaved,
			Action: "optimized",
			Reason: fmt.Sprintf("Hard-linked %d files, saved %d bytes", result.LinksCreated, result.SpaceSaved),
		},
	}

	f.logger.Info(fmt.Sprintf("Store optimization completed: saved %d bytes", result.SpaceSaved))

	return response, nil
}

// cleanStore performs comprehensive store cleanup
func (f *GcFunction) cleanStore(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Performing comprehensive store cleanup")

	// First collect garbage
	gcResult, err := f.agent.Collect(ctx, &agent.GcOptions{
		DryRun:          req.DryRun,
		MaxAge:          req.MaxAge,
		MaxSize:         req.MaxSize,
		KeepOutputs:     req.KeepOutputs,
		KeepDerivations: req.KeepDerivations,
		Force:           req.Force,
	})
	if err != nil {
		return nil, fmt.Errorf("garbage collection failed during cleanup: %w", err)
	}

	response.FreedSpace += gcResult.FreedSpace
	response.DeletedPaths += gcResult.DeletedPaths

	// Then optimize if not dry run
	if !req.DryRun {
		optimizeResult, err := f.agent.OptimizeStore(ctx, req.Force)
		if err != nil {
			f.logger.Error(fmt.Sprintf("Store optimization failed during cleanup: %v", err))
		} else {
			response.FreedSpace += optimizeResult.SpaceSaved
			response.Details = append(response.Details, GcDetail{
				Path:   "/nix/store",
				Size:   optimizeResult.SpaceSaved,
				Action: "optimized",
				Reason: fmt.Sprintf("Hard-linked %d files", optimizeResult.LinksCreated),
			})
		}
	}

	// Clean up temporary files
	tempCleanup, err := f.agent.CleanTempFiles(ctx, req.DryRun)
	if err != nil {
		f.logger.Error(fmt.Sprintf("Temporary file cleanup failed: %v", err))
	} else {
		response.FreedSpace += tempCleanup.SpaceFreed
		response.Details = append(response.Details, GcDetail{
			Path:   "/tmp/nix-*",
			Size:   tempCleanup.SpaceFreed,
			Action: "cleaned",
			Reason: fmt.Sprintf("Removed %d temporary files", tempCleanup.FilesRemoved),
		})
	}

	f.logger.Info(fmt.Sprintf("Store cleanup completed: freed %d bytes total", response.FreedSpace))

	return response, nil
}

// analyzeStore analyzes the Nix store and provides recommendations
func (f *GcFunction) analyzeStore(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Analyzing Nix store")

	analysis, err := f.agent.AnalyzeStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("store analysis failed: %w", err)
	}

	response.StoreSize = analysis.TotalSize
	response.RemainingPaths = analysis.TotalPaths

	// Add analysis details
	for category, info := range analysis.Categories {
		response.Details = append(response.Details, GcDetail{
			Path:   category,
			Size:   info.Size,
			Action: "analyzed",
			Reason: fmt.Sprintf("%s: %d paths, %.1f%% of store", category, info.Count, info.Percentage),
		})
	}

	// Generate detailed recommendations
	response.Recommendations = f.generateAnalysisRecommendations(analysis)

	f.logger.Info(fmt.Sprintf("Store analysis completed: %d categories analyzed", len(analysis.Categories)))

	return response, nil
}

// generateRecommendations generates recommendations based on GC results
func (f *GcFunction) generateRecommendations(result *agent.GcResult) []string {
	recommendations := []string{}

	if result.FreedSpace > 1<<30 { // > 1GB
		recommendations = append(recommendations, "Significant space was freed. Consider running GC more frequently.")
	}

	if result.DeletedPaths > 1000 {
		recommendations = append(recommendations, "Many paths were deleted. Consider adjusting your garbage collection policy.")
	}

	if result.RemainingPaths > 10000 {
		recommendations = append(recommendations, "Large number of paths remaining. Consider using more aggressive GC settings.")
	}

	recommendations = append(recommendations, "Run 'nixai gc --operation=optimize' to further reduce store size through hard-linking.")

	return recommendations
}

// generateStatusRecommendations generates recommendations based on store status
func (f *GcFunction) generateStatusRecommendations(status *agent.StoreStatus) []string {
	recommendations := []string{}

	sizeGB := float64(status.TotalSize) / (1 << 30)
	if sizeGB > 50 {
		recommendations = append(recommendations, fmt.Sprintf("Store is large (%.1f GB). Consider running garbage collection.", sizeGB))
	}

	if status.TotalPaths > 50000 {
		recommendations = append(recommendations, "Many paths in store. Consider more frequent garbage collection.")
	}

	recommendations = append(recommendations, "Use 'nixai gc --operation=analyze' for detailed store analysis.")
	recommendations = append(recommendations, "Use 'nixai gc --operation=list' to see what can be cleaned up.")

	return recommendations
}

// generateAnalysisRecommendations generates recommendations based on store analysis
func (f *GcFunction) generateAnalysisRecommendations(analysis *agent.StoreAnalysis) []string {
	recommendations := []string{}

	// Find largest categories
	type categoryInfo struct {
		name       string
		size       int64
		percentage float64
	}

	var categories []categoryInfo
	for name, info := range analysis.Categories {
		categories = append(categories, categoryInfo{
			name:       name,
			size:       info.Size,
			percentage: info.Percentage,
		})
	}

	// Sort by size and recommend cleanup for largest categories
	for _, cat := range categories {
		if cat.percentage > 20 {
			recommendations = append(recommendations, fmt.Sprintf("Category '%s' uses %.1f%% of store space. Consider cleanup.", cat.name, cat.percentage))
		}
	}

	if analysis.OldestPath != nil {
		recommendations = append(recommendations, fmt.Sprintf("Oldest path is from %v. Consider age-based cleanup.", analysis.OldestPath.LastAccessed))
	}

	return recommendations
}
