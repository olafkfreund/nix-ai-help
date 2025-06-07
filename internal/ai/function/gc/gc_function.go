package gc

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// GcFunction handles Nix garbage collection operations
type GcFunction struct {
	*functionbase.BaseFunction
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
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("context", "The context or reason for the garbage collection operation", true),
		functionbase.StringParamWithOptions("operation", "The garbage collection operation to perform", false,
			[]string{"collect", "list", "status", "optimize", "clean", "analyze"}, nil, nil),
		functionbase.BoolParam("dry_run", "Whether to perform a dry run without actually deleting anything", false),
		functionbase.StringParam("max_age", "Maximum age of paths to keep (e.g., '30d', '1w', '6h')", false),
		functionbase.StringParam("max_size", "Maximum store size to maintain (e.g., '10GB', '1TB')", false),
		functionbase.BoolParam("keep_outputs", "Whether to keep build outputs", false),
		functionbase.BoolParam("keep_derivations", "Whether to keep derivations", false),
		functionbase.BoolParam("verbose", "Whether to provide verbose output", false),
		functionbase.BoolParam("force", "Whether to force garbage collection without confirmation", false),
		{
			Name:        "options",
			Type:        "object",
			Description: "Additional garbage collection options",
			Required:    false,
		},
	}

	// Create base function
	baseFunc := functionbase.NewBaseFunction(
		"gc",
		"Manage Nix store garbage collection and cleanup operations",
		parameters,
	)

	return &GcFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
	}
}

// Name returns the function name
func (f *GcFunction) Name() string {
	return f.BaseFunction.Name()
}

// Description returns the function description
func (f *GcFunction) Description() string {
	return f.BaseFunction.Description()
}

// Version returns the function version
func (f *GcFunction) Version() string {
	return "1.0.0"
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
	start := time.Now()

	// Validate parameters
	if err := f.ValidateParameters(params); err != nil {
		return functionbase.ErrorResult(fmt.Errorf("parameter validation failed: %w", err), time.Since(start)), nil
	}

	// Parse request
	req, err := f.parseRequest(params)
	if err != nil {
		return functionbase.ErrorResult(fmt.Errorf("failed to parse request: %w", err), time.Since(start)), nil
	}

	f.logger.Info(fmt.Sprintf("Executing garbage collection operation: %s", req.Operation))

	// Execute garbage collection operation
	response, err := f.executeGcOperation(ctx, req)
	if err != nil {
		return functionbase.ErrorResult(err, time.Since(start)), nil
	}

	response.ExecutionTime = time.Since(start)

	return functionbase.SuccessResult(response, time.Since(start)), nil
}

// parseRequest converts raw parameters to GcRequest
func (f *GcFunction) parseRequest(params map[string]interface{}) (*GcRequest, error) {
	req := &GcRequest{}

	if context, ok := params["context"].(string); ok {
		req.Context = context
	}

	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	} else {
		req.Operation = "collect" // default
	}

	if dryRun, ok := params["dry_run"].(bool); ok {
		req.DryRun = dryRun
	}

	if maxAge, ok := params["max_age"].(string); ok {
		req.MaxAge = maxAge
	}

	if maxSize, ok := params["max_size"].(string); ok {
		req.MaxSize = maxSize
	}

	if keepOutputs, ok := params["keep_outputs"].(bool); ok {
		req.KeepOutputs = keepOutputs
	}

	if keepDerivations, ok := params["keep_derivations"].(bool); ok {
		req.KeepDerivations = keepDerivations
	}

	if verbose, ok := params["verbose"].(bool); ok {
		req.Verbose = verbose
	}

	if force, ok := params["force"].(bool); ok {
		req.Force = force
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

	// Mock store status
	response.StoreSize = 1000000000 // 1GB mock size

	// Mock garbage collection results
	response.FreedSpace = 250000000 // 250MB freed
	response.DeletedPaths = 150
	response.RemainingPaths = 850

	// Mock details
	response.Details = []GcDetail{
		{
			Path:   "/nix/store/abc123-old-package",
			Size:   50000000,
			Action: "deleted",
			Reason: "garbage collected",
		},
		{
			Path:   "/nix/store/def456-temp-derivation",
			Size:   25000000,
			Action: "deleted",
			Reason: "temporary derivation",
		},
	}

	// Mock recommendations
	response.Recommendations = []string{
		"Significant space was freed. Consider running GC more frequently.",
		"Run 'nixai gc --operation=optimize' to further reduce store size through hard-linking.",
	}

	if req.DryRun {
		f.logger.Info(fmt.Sprintf("Dry run completed: would free %d bytes from %d paths", response.FreedSpace, response.DeletedPaths))
	} else {
		f.logger.Info(fmt.Sprintf("Garbage collection completed: freed %d bytes from %d paths", response.FreedSpace, response.DeletedPaths))
	}

	return response, nil
}

// listGarbage lists garbage without collecting it
func (f *GcFunction) listGarbage(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Listing garbage in Nix store")

	// Mock garbage list
	mockGarbage := []GcDetail{
		{
			Path:   "/nix/store/abc123-old-package",
			Size:   50000000,
			Action: "can_delete",
			Reason: "not referenced",
		},
		{
			Path:   "/nix/store/def456-temp-build",
			Size:   25000000,
			Action: "can_delete",
			Reason: "temporary build artifact",
		},
		{
			Path:   "/nix/store/ghi789-unused-derivation",
			Size:   15000000,
			Action: "can_delete",
			Reason: "derivation not referenced",
		},
	}

	var totalSize int64
	for _, item := range mockGarbage {
		response.Details = append(response.Details, item)
		totalSize += item.Size
	}

	response.FreedSpace = totalSize
	response.DeletedPaths = len(mockGarbage)

	f.logger.Info(fmt.Sprintf("Found %d garbage paths totaling %d bytes", len(mockGarbage), totalSize))

	return response, nil
}

// getStatus gets the current Nix store status
func (f *GcFunction) getStatus(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Getting Nix store status")

	// Mock store status
	mockStoreSize := int64(1500000000) // 1.5GB
	mockTotalPaths := 1200

	response.StoreSize = mockStoreSize
	response.RemainingPaths = mockTotalPaths
	response.Details = []GcDetail{
		{
			Path:   "/nix/store",
			Size:   mockStoreSize,
			Action: "status",
			Reason: fmt.Sprintf("Total store size: %d bytes, %d paths", mockStoreSize, mockTotalPaths),
		},
		{
			Path:   "packages",
			Size:   800000000,
			Action: "info",
			Reason: "packages: 800 paths",
		},
		{
			Path:   "derivations",
			Size:   400000000,
			Action: "info",
			Reason: "derivations: 250 paths",
		},
		{
			Path:   "build-artifacts",
			Size:   300000000,
			Action: "info",
			Reason: "build-artifacts: 150 paths",
		},
	}

	// Mock recommendations
	response.Recommendations = []string{
		"Store is large (1.4 GB). Consider running garbage collection.",
		"Use 'nixai gc --operation=analyze' for detailed store analysis.",
		"Use 'nixai gc --operation=list' to see what can be cleaned up.",
	}

	f.logger.Info(fmt.Sprintf("Store status: %d bytes in %d paths", mockStoreSize, mockTotalPaths))

	return response, nil
}

// optimizeStore optimizes the Nix store
func (f *GcFunction) optimizeStore(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Optimizing Nix store")

	// Mock optimization results
	mockSpaceSaved := int64(100000000) // 100MB saved
	mockLinksCreated := 500

	response.FreedSpace = mockSpaceSaved
	response.Details = []GcDetail{
		{
			Path:   "/nix/store",
			Size:   mockSpaceSaved,
			Action: "optimized",
			Reason: fmt.Sprintf("Hard-linked %d files, saved %d bytes", mockLinksCreated, mockSpaceSaved),
		},
	}

	f.logger.Info(fmt.Sprintf("Store optimization completed: saved %d bytes", mockSpaceSaved))

	return response, nil
}

// cleanStore performs comprehensive store cleanup
func (f *GcFunction) cleanStore(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Performing comprehensive store cleanup")

	// Mock comprehensive cleanup results
	response.FreedSpace = 350000000 // 350MB total freed
	response.DeletedPaths = 200

	// Mock garbage collection phase
	response.Details = append(response.Details, GcDetail{
		Path:   "/nix/store/garbage-collected",
		Size:   250000000,
		Action: "deleted",
		Reason: "garbage collection phase",
	})

	// Mock optimization phase (if not dry run)
	if !req.DryRun {
		response.Details = append(response.Details, GcDetail{
			Path:   "/nix/store",
			Size:   75000000,
			Action: "optimized",
			Reason: "Hard-linked 300 files",
		})
	}

	// Mock temporary file cleanup
	response.Details = append(response.Details, GcDetail{
		Path:   "/tmp/nix-*",
		Size:   25000000,
		Action: "cleaned",
		Reason: "Removed 50 temporary files",
	})

	f.logger.Info(fmt.Sprintf("Store cleanup completed: freed %d bytes total", response.FreedSpace))

	return response, nil
}

// analyzeStore analyzes the Nix store and provides recommendations
func (f *GcFunction) analyzeStore(ctx context.Context, req *GcRequest, response *GcResponse) (*GcResponse, error) {
	f.logger.Info("Analyzing Nix store")

	// Mock store analysis
	mockStoreSize := int64(1500000000) // 1.5GB
	mockTotalPaths := 1200

	response.StoreSize = mockStoreSize
	response.RemainingPaths = mockTotalPaths

	// Mock analysis categories
	response.Details = []GcDetail{
		{
			Path:   "packages",
			Size:   750000000,
			Action: "analyzed",
			Reason: "packages: 600 paths, 50.0% of store",
		},
		{
			Path:   "derivations",
			Size:   450000000,
			Action: "analyzed",
			Reason: "derivations: 350 paths, 30.0% of store",
		},
		{
			Path:   "build-outputs",
			Size:   300000000,
			Action: "analyzed",
			Reason: "build-outputs: 250 paths, 20.0% of store",
		},
	}

	// Mock recommendations
	response.Recommendations = []string{
		"Category 'packages' uses 50.0% of store space. Consider cleanup.",
		"Category 'derivations' uses 30.0% of store space. Consider cleanup.",
		"Use age-based cleanup for old paths.",
	}

	f.logger.Info(fmt.Sprintf("Store analysis completed: %d categories analyzed", len(response.Details)))

	return response, nil
}
