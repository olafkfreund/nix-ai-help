package agent

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// GCAgent handles garbage collection and storage management operations
type GCAgent struct {
	BaseAgent
}

// GCContext contains garbage collection and storage management context information
type GCContext struct {
	StoreSize        string   `json:"store_size,omitempty"`        // Current store size
	StoreUsage       string   `json:"store_usage,omitempty"`       // Store usage percentage
	GenerationCount  int      `json:"generation_count,omitempty"`  // Number of generations
	OldestGeneration string   `json:"oldest_generation,omitempty"` // Oldest generation date
	LastCleanup      string   `json:"last_cleanup,omitempty"`      // Last GC run date
	LargeItems       []string `json:"large_items,omitempty"`       // Large store items
	UnusedItems      []string `json:"unused_items,omitempty"`      // Unused store items
	RootsCount       int      `json:"roots_count,omitempty"`       // Number of GC roots
	GCRoots          []string `json:"gc_roots,omitempty"`          // GC root paths
	DiskSpace        string   `json:"disk_space,omitempty"`        // Available disk space
	CleanupOptions   []string `json:"cleanup_options,omitempty"`   // Available cleanup options
	AutoGCEnabled    bool     `json:"auto_gc_enabled,omitempty"`   // Auto GC status
	GCSchedule       string   `json:"gc_schedule,omitempty"`       // GC schedule configuration
	StorePaths       []string `json:"store_paths,omitempty"`       // Specific store paths to analyze
	DryRun           bool     `json:"dry_run,omitempty"`           // Dry run mode
	KeepOutputs      bool     `json:"keep_outputs,omitempty"`      // Keep build outputs
	KeepDerivations  bool     `json:"keep_derivations,omitempty"`  // Keep derivations
	MaxFreed         string   `json:"max_freed,omitempty"`         // Maximum space to free
	MinAge           string   `json:"min_age,omitempty"`           // Minimum age for cleanup
	SystemProfile    string   `json:"system_profile,omitempty"`    // System profile path
	UserProfiles     []string `json:"user_profiles,omitempty"`     // User profile paths
}

// NewGCAgent creates a new GC agent with the specified provider.
func NewGCAgent(provider ai.Provider) *GCAgent {
	agent := &GCAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleGC,
		},
	}
	return agent
}

// Query handles GC-related queries using the provider.
func (a *GCAgent) Query(ctx context.Context, prompt string) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("AI provider not configured")
	}

	if err := a.validateRole(); err != nil {
		return "", err
	}

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		return p.QueryWithContext(ctx, prompt)
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		return p.Query(prompt)
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateResponse handles GC-specific response generation.
func (a *GCAgent) GenerateResponse(ctx context.Context, input string) (string, error) {
	if a.role == "" {
		return "", fmt.Errorf("role not set for GCAgent")
	}

	// Build enhanced prompt with GC context and role
	prompt := a.buildContextualPrompt(input)
	enhancedPrompt := a.enhancePromptWithRole(prompt)

	// Use provider to generate response
	response, err := a.provider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", fmt.Errorf("GC agent response generation failed: %w", err)
	}

	return a.enhanceResponseWithGCGuidance(response), nil
}

// buildContextualPrompt creates a comprehensive prompt with GC context.
func (a *GCAgent) buildContextualPrompt(input string) string {
	prompt := fmt.Sprintf("GC Query: %s\n\n", input)

	// Add GC context if available
	if a.contextData != nil {
		if gcCtx, ok := a.contextData.(*GCContext); ok {
			prompt += a.buildGCContextSection(gcCtx)
		}
	}

	return prompt
}

// buildGCContextSection creates a formatted context section for GC operations.
func (a *GCAgent) buildGCContextSection(ctx *GCContext) string {
	var contextStr string

	if ctx.StoreSize != "" || ctx.StoreUsage != "" {
		contextStr += "## Store Status\n"
		if ctx.StoreSize != "" {
			contextStr += fmt.Sprintf("- Store Size: %s\n", ctx.StoreSize)
		}
		if ctx.StoreUsage != "" {
			contextStr += fmt.Sprintf("- Store Usage: %s\n", ctx.StoreUsage)
		}
		if ctx.DiskSpace != "" {
			contextStr += fmt.Sprintf("- Available Disk Space: %s\n", ctx.DiskSpace)
		}
		contextStr += "\n"
	}

	if ctx.GenerationCount > 0 || ctx.OldestGeneration != "" {
		contextStr += "## Generation Management\n"
		if ctx.GenerationCount > 0 {
			contextStr += fmt.Sprintf("- Generation Count: %d\n", ctx.GenerationCount)
		}
		if ctx.OldestGeneration != "" {
			contextStr += fmt.Sprintf("- Oldest Generation: %s\n", ctx.OldestGeneration)
		}
		if ctx.LastCleanup != "" {
			contextStr += fmt.Sprintf("- Last Cleanup: %s\n", ctx.LastCleanup)
		}
		contextStr += "\n"
	}

	if len(ctx.LargeItems) > 0 {
		contextStr += "## Large Store Items\n"
		for _, item := range ctx.LargeItems {
			contextStr += fmt.Sprintf("- %s\n", item)
		}
		contextStr += "\n"
	}

	if len(ctx.UnusedItems) > 0 {
		contextStr += "## Unused Store Items\n"
		for _, item := range ctx.UnusedItems {
			contextStr += fmt.Sprintf("- %s\n", item)
		}
		contextStr += "\n"
	}

	if ctx.RootsCount > 0 || len(ctx.GCRoots) > 0 {
		contextStr += "## GC Roots\n"
		if ctx.RootsCount > 0 {
			contextStr += fmt.Sprintf("- Root Count: %d\n", ctx.RootsCount)
		}
		if len(ctx.GCRoots) > 0 {
			contextStr += "- Root Paths:\n"
			for _, root := range ctx.GCRoots {
				contextStr += fmt.Sprintf("  - %s\n", root)
			}
		}
		contextStr += "\n"
	}

	if len(ctx.CleanupOptions) > 0 {
		contextStr += "## Cleanup Options\n"
		for _, option := range ctx.CleanupOptions {
			contextStr += fmt.Sprintf("- %s\n", option)
		}
		contextStr += "\n"
	}

	if ctx.AutoGCEnabled || ctx.GCSchedule != "" {
		contextStr += "## GC Configuration\n"
		contextStr += fmt.Sprintf("- Auto GC Enabled: %t\n", ctx.AutoGCEnabled)
		if ctx.GCSchedule != "" {
			contextStr += fmt.Sprintf("- GC Schedule: %s\n", ctx.GCSchedule)
		}
		contextStr += "\n"
	}

	if len(ctx.StorePaths) > 0 {
		contextStr += "## Store Paths to Analyze\n"
		for _, path := range ctx.StorePaths {
			contextStr += fmt.Sprintf("- %s\n", path)
		}
		contextStr += "\n"
	}

	// Add cleanup preferences
	if ctx.DryRun || ctx.KeepOutputs || ctx.KeepDerivations {
		contextStr += "## Cleanup Preferences\n"
		if ctx.DryRun {
			contextStr += "- Dry Run Mode: Enabled\n"
		}
		if ctx.KeepOutputs {
			contextStr += "- Keep Build Outputs: Yes\n"
		}
		if ctx.KeepDerivations {
			contextStr += "- Keep Derivations: Yes\n"
		}
		if ctx.MaxFreed != "" {
			contextStr += fmt.Sprintf("- Max Space to Free: %s\n", ctx.MaxFreed)
		}
		if ctx.MinAge != "" {
			contextStr += fmt.Sprintf("- Minimum Age for Cleanup: %s\n", ctx.MinAge)
		}
		contextStr += "\n"
	}

	return contextStr
}

// enhancePromptWithRole adds role-specific instructions to a generic prompt.
func (a *GCAgent) enhancePromptWithRole(prompt string) string {
	rolePrompt := roles.RolePromptTemplate[a.role]
	return fmt.Sprintf("%s\n\n%s", rolePrompt, prompt)
}

// enhanceResponseWithGCGuidance adds GC-specific guidance to responses.
func (a *GCAgent) enhanceResponseWithGCGuidance(response string) string {
	guidance := "\n\n---\n**GC Safety Tips:**\n"
	guidance += "- Always test GC commands with `--dry-run` first\n"
	guidance += "- Ensure no important builds are running before cleanup\n"
	guidance += "- Consider keeping recent generations as rollback points\n"
	guidance += "- Monitor disk space after cleanup operations\n"

	return response + guidance
}
