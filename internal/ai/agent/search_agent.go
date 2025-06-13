package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// SearchAgent handles documentation and package search operations
type SearchAgent struct {
	BaseAgent
}

// SearchContext contains search-specific context information
type SearchContext struct {
	SearchQuery    string   `json:"search_query,omitempty"`
	SearchType     string   `json:"search_type,omitempty"` // packages, options, docs, etc.
	SearchResults  []string `json:"search_results,omitempty"`
	SearchSources  []string `json:"search_sources,omitempty"` // nixpkgs, wiki, manual, etc.
	ChannelVersion string   `json:"channel_version,omitempty"`
	SystemArch     string   `json:"system_arch,omitempty"`
	PackageFilters []string `json:"package_filters,omitempty"` // license, maintainer, etc.
	SearchLimit    int      `json:"search_limit,omitempty"`
	SortBy         string   `json:"sort_by,omitempty"` // relevance, name, popularity
	IncludeUnfree  bool     `json:"include_unfree,omitempty"`
	SearchHistory  []string `json:"search_history,omitempty"`
	RelatedQueries []string `json:"related_queries,omitempty"`
	MCPResults     string   `json:"mcp_results,omitempty"` // MCP server search results
	DocSections    []string `json:"doc_sections,omitempty"`
}

// NewSearchAgent creates a new SearchAgent with the Search role
func NewSearchAgent(provider ai.Provider) *SearchAgent {
	agent := &SearchAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleSearch,
		},
	}
	return agent
}

// Query handles search-related questions and operations
func (a *SearchAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	prompt, ok := roles.RolePromptTemplate[a.role]
	if !ok {
		return "", fmt.Errorf("no prompt template for role: %s", a.role)
	}

	// Build context-aware prompt
	fullPrompt := a.buildContextualPrompt(prompt, question)

	if p, ok := a.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		return p.QueryWithContext(ctx, fullPrompt)
	}
	if p, ok := a.provider.(interface{ Query(string) (string, error) }); ok {
		return p.Query(fullPrompt)
	}
	return "", fmt.Errorf("provider does not implement QueryWithContext or Query")
}

// GenerateResponse handles search response generation
func (a *SearchAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Add search-specific context to the prompt
	contextualPrompt := a.buildContextualPrompt("", prompt)

	return a.provider.GenerateResponse(ctx, contextualPrompt)
}

// buildContextualPrompt constructs a context-aware prompt for search operations
func (a *SearchAgent) buildContextualPrompt(rolePrompt, userInput string) string {
	var promptParts []string

	if rolePrompt != "" {
		promptParts = append(promptParts, rolePrompt)
	}

	// Add search context if available
	if a.contextData != nil {
		if searchCtx, ok := a.contextData.(*SearchContext); ok {
			contextStr := a.formatSearchContext(searchCtx)
			if contextStr != "" {
				promptParts = append(promptParts, "Search Context:")
				promptParts = append(promptParts, contextStr)
			}
		}
	}

	// Add user input
	promptParts = append(promptParts, "Search Request:")
	promptParts = append(promptParts, userInput)

	return strings.Join(promptParts, "\n\n")
}

// formatSearchContext formats SearchContext into a readable string
func (a *SearchAgent) formatSearchContext(ctx *SearchContext) string {
	var parts []string

	if ctx.SearchQuery != "" {
		parts = append(parts, fmt.Sprintf("Query: %s", ctx.SearchQuery))
	}

	if ctx.SearchType != "" {
		parts = append(parts, fmt.Sprintf("Search Type: %s", ctx.SearchType))
	}

	if len(ctx.SearchSources) > 0 {
		parts = append(parts, fmt.Sprintf("Sources: %s", strings.Join(ctx.SearchSources, ", ")))
	}

	if ctx.ChannelVersion != "" {
		parts = append(parts, fmt.Sprintf("Channel: %s", ctx.ChannelVersion))
	}

	if ctx.SystemArch != "" {
		parts = append(parts, fmt.Sprintf("Architecture: %s", ctx.SystemArch))
	}

	if len(ctx.PackageFilters) > 0 {
		parts = append(parts, fmt.Sprintf("Filters: %s", strings.Join(ctx.PackageFilters, ", ")))
	}

	if ctx.SearchLimit > 0 {
		parts = append(parts, fmt.Sprintf("Limit: %d", ctx.SearchLimit))
	}

	if ctx.SortBy != "" {
		parts = append(parts, fmt.Sprintf("Sort By: %s", ctx.SortBy))
	}

	if ctx.IncludeUnfree {
		parts = append(parts, "Include Unfree: true")
	}

	if len(ctx.SearchResults) > 0 {
		resultsStr := strings.Join(ctx.SearchResults, "\n")
		parts = append(parts, fmt.Sprintf("Current Results:\n%s", resultsStr))
	}

	if len(ctx.RelatedQueries) > 0 {
		parts = append(parts, fmt.Sprintf("Related Queries: %s", strings.Join(ctx.RelatedQueries, ", ")))
	}

	if ctx.MCPResults != "" {
		parts = append(parts, fmt.Sprintf("Documentation Results:\n%s", ctx.MCPResults))
	}

	if len(ctx.DocSections) > 0 {
		parts = append(parts, fmt.Sprintf("Documentation Sections: %s", strings.Join(ctx.DocSections, ", ")))
	}

	if len(ctx.SearchHistory) > 0 {
		parts = append(parts, fmt.Sprintf("Search History: %s", strings.Join(ctx.SearchHistory, ", ")))
	}

	return strings.Join(parts, "\n")
}

// SetSearchContext is a convenience method to set SearchContext
func (a *SearchAgent) SetSearchContext(ctx *SearchContext) {
	a.SetContext(ctx)
}
