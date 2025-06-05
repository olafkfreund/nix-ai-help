package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// StoreContext represents context for Nix store management operations.
type StoreContext struct {
	// Store information
	StorePath    string `json:"store_path,omitempty"`
	StoreSize    string `json:"store_size,omitempty"`
	FreeSpace    string `json:"free_space,omitempty"`
	StoreVersion string `json:"store_version,omitempty"`

	// Operation context
	TargetPaths   []string `json:"target_paths,omitempty"`
	QueryType     string   `json:"query_type,omitempty"`
	OperationType string   `json:"operation_type,omitempty"`

	// Garbage collection context
	Generations    []string `json:"generations,omitempty"`
	Roots          []string `json:"roots,omitempty"`
	OldGenerations int      `json:"old_generations,omitempty"`
	DryRun         bool     `json:"dry_run,omitempty"`

	// Store health
	CorruptedPaths []string `json:"corrupted_paths,omitempty"`
	OrphanedPaths  []string `json:"orphaned_paths,omitempty"`
	Issues         []string `json:"issues,omitempty"`

	// Performance context
	CacheHitRate     string   `json:"cache_hit_rate,omitempty"`
	BuildPerformance string   `json:"build_performance,omitempty"`
	NetworkStores    []string `json:"network_stores,omitempty"`
}

// StoreAgent represents an agent specialized in Nix store management.
type StoreAgent struct {
	BaseAgent
	context *StoreContext
}

// NewStoreAgent creates a new StoreAgent instance.
func NewStoreAgent(provider ai.Provider) *StoreAgent {
	agent := &StoreAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleStore,
		},
		context: &StoreContext{},
	}
	return agent
}

// SetContext sets the store context for operations.
func (a *StoreAgent) SetContext(ctx *StoreContext) {
	a.context = ctx
}

// GetContext returns the current store context.
func (a *StoreAgent) GetContext() *StoreContext {
	return a.context
}

// AnalyzeStoreHealth analyzes the current state and health of the Nix store.
func (a *StoreAgent) AnalyzeStoreHealth(ctx context.Context, storePath string) (string, error) {
	a.context.StorePath = storePath
	a.context.OperationType = "health_analysis"

	prompt := a.buildPrompt("Analyze the health and current state of the Nix store", map[string]interface{}{
		"store_path": storePath,
		"operation":  "comprehensive health check",
		"include":    "integrity, size analysis, corruption detection, performance metrics",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to analyze store health: %w", err)
	}

	return a.formatStoreResponse(response, "Store Health Analysis"), nil
}

// QueryStorePaths performs various queries on store paths and dependencies.
func (a *StoreAgent) QueryStorePaths(ctx context.Context, paths []string, queryType string) (string, error) {
	a.context.TargetPaths = paths
	a.context.QueryType = queryType
	a.context.OperationType = "path_query"

	prompt := a.buildPrompt("Query and analyze Nix store paths and their relationships", map[string]interface{}{
		"target_paths": paths,
		"query_type":   queryType,
		"operation":    "store path analysis",
		"include":      "dependencies, reverse dependencies, closures, sizes",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to query store paths: %w", err)
	}

	return a.formatStoreResponse(response, "Store Path Query Results"), nil
}

// OptimizeGarbageCollection provides intelligent garbage collection strategies.
func (a *StoreAgent) OptimizeGarbageCollection(ctx context.Context, generations []string, dryRun bool) (string, error) {
	a.context.Generations = generations
	a.context.DryRun = dryRun
	a.context.OperationType = "garbage_collection"

	prompt := a.buildPrompt("Optimize garbage collection strategy for the Nix store", map[string]interface{}{
		"generations": generations,
		"dry_run":     dryRun,
		"operation":   "intelligent garbage collection",
		"include":     "safety checks, space estimation, root preservation, cleanup strategy",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to optimize garbage collection: %w", err)
	}

	return a.formatStoreResponse(response, "Garbage Collection Strategy"), nil
}

// RepairStoreIntegrity diagnoses and repairs store integrity issues.
func (a *StoreAgent) RepairStoreIntegrity(ctx context.Context, corruptedPaths []string) (string, error) {
	a.context.CorruptedPaths = corruptedPaths
	a.context.OperationType = "integrity_repair"

	prompt := a.buildPrompt("Diagnose and repair Nix store integrity issues", map[string]interface{}{
		"corrupted_paths": corruptedPaths,
		"operation":       "store integrity repair",
		"include":         "corruption detection, repair strategies, data recovery, prevention",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to repair store integrity: %w", err)
	}

	return a.formatStoreResponse(response, "Store Integrity Repair"), nil
}

// OptimizeStorePerformance provides performance optimization recommendations.
func (a *StoreAgent) OptimizeStorePerformance(ctx context.Context, performanceIssues []string) (string, error) {
	a.context.Issues = performanceIssues
	a.context.OperationType = "performance_optimization"

	prompt := a.buildPrompt("Optimize Nix store performance and access patterns", map[string]interface{}{
		"performance_issues": performanceIssues,
		"operation":          "store performance optimization",
		"include":            "caching strategies, network optimization, access patterns, deduplication",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to optimize store performance: %w", err)
	}

	return a.formatStoreResponse(response, "Store Performance Optimization"), nil
}

// ManageStoreCopying provides guidance for store copying and migration operations.
func (a *StoreAgent) ManageStoreCopying(ctx context.Context, sourceStore, targetStore string, paths []string) (string, error) {
	a.context.TargetPaths = paths
	a.context.OperationType = "store_copying"

	prompt := a.buildPrompt("Manage Nix store copying and migration operations", map[string]interface{}{
		"source_store": sourceStore,
		"target_store": targetStore,
		"paths":        paths,
		"operation":    "store copying and migration",
		"include":      "copy strategies, network optimization, integrity verification, rollback plans",
	})

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to manage store copying: %w", err)
	}

	return a.formatStoreResponse(response, "Store Copying Management"), nil
}

// buildPrompt creates a specialized prompt for store operations.
func (a *StoreAgent) buildPrompt(task string, details map[string]interface{}) string {
	var prompt strings.Builder

	// Add role-specific context
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	// Add task description
	prompt.WriteString(fmt.Sprintf("**Task**: %s\n\n", task))

	// Add store context
	prompt.WriteString("**Store Context**:\n")
	if a.context.StorePath != "" {
		prompt.WriteString(fmt.Sprintf("- Store Path: %s\n", a.context.StorePath))
	}
	if a.context.StoreSize != "" {
		prompt.WriteString(fmt.Sprintf("- Store Size: %s\n", a.context.StoreSize))
	}
	if a.context.OperationType != "" {
		prompt.WriteString(fmt.Sprintf("- Operation Type: %s\n", a.context.OperationType))
	}
	if len(a.context.TargetPaths) > 0 {
		prompt.WriteString(fmt.Sprintf("- Target Paths: %v\n", a.context.TargetPaths))
	}
	if len(a.context.Issues) > 0 {
		prompt.WriteString(fmt.Sprintf("- Known Issues: %v\n", a.context.Issues))
	}

	// Add specific task details
	if len(details) > 0 {
		prompt.WriteString("\n**Operation Details**:\n")
		for key, value := range details {
			prompt.WriteString(fmt.Sprintf("- %s: %v\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value))
		}
	}

	// Add safety and best practices reminder
	prompt.WriteString("\n**Requirements**:\n")
	prompt.WriteString("- Provide specific nix-store commands with detailed explanations\n")
	prompt.WriteString("- Include safety warnings for potentially destructive operations\n")
	prompt.WriteString("- Estimate space savings and performance impacts\n")
	prompt.WriteString("- Suggest preventive maintenance practices\n")
	prompt.WriteString("- Explain the rationale behind recommended operations\n")
	prompt.WriteString("- Consider system stability and dependency preservation\n")

	return prompt.String()
}

// formatStoreResponse formats the AI response for store management operations.
func (a *StoreAgent) formatStoreResponse(response, operation string) string {
	var formatted strings.Builder

	formatted.WriteString(fmt.Sprintf("# %s\n\n", operation))
	formatted.WriteString(response)

	// Add context-specific footer
	formatted.WriteString("\n\n---\n")
	formatted.WriteString("**⚠️  Store Operation Safety Reminders**:\n")
	formatted.WriteString("- Always backup important data before major store operations\n")
	formatted.WriteString("- Test operations with --dry-run when available\n")
	formatted.WriteString("- Verify store integrity after significant changes\n")
	formatted.WriteString("- Monitor disk space during large operations\n")
	formatted.WriteString("- Keep track of active roots and generations\n")

	return formatted.String()
}
