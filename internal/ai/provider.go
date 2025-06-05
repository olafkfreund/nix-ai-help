package ai

import "context"

// Provider defines the interface that all AI providers must implement.
// This interface supports both simple queries and context-aware operations.
type Provider interface {
	Query(ctx context.Context, prompt string) (string, error)
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

// AIProvider is the legacy interface for backward compatibility.
// New code should use Provider interface instead.
type AIProvider interface {
	Query(prompt string) (string, error)
}

// LegacyProviderAdapter wraps an AIProvider to implement the new Provider interface
type LegacyProviderAdapter struct {
	legacy AIProvider
}

// NewLegacyProviderAdapter creates an adapter that wraps a legacy AIProvider
func NewLegacyProviderAdapter(legacy AIProvider) Provider {
	return &LegacyProviderAdapter{legacy: legacy}
}

// Query implements the Provider interface by delegating to the legacy provider
func (a *LegacyProviderAdapter) Query(ctx context.Context, prompt string) (string, error) {
	return a.legacy.Query(prompt)
}

// GenerateResponse implements the Provider interface by delegating to the legacy provider
func (a *LegacyProviderAdapter) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Most legacy providers don't have a separate GenerateResponse method, so use Query
	if gr, ok := a.legacy.(interface{ GenerateResponse(string) (string, error) }); ok {
		return gr.GenerateResponse(prompt)
	}
	return a.legacy.Query(prompt)
}

// ProviderFactory manages registration and retrieval of AI providers.
type ProviderFactory struct {
	providers map[string]Provider
	legacy    map[string]AIProvider
}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]Provider),
		legacy:    make(map[string]AIProvider),
	}
}

// RegisterProvider registers a new Provider implementation.
func (f *ProviderFactory) RegisterProvider(name string, provider Provider) {
	f.providers[name] = provider
}

// RegisterLegacyProvider registers a legacy AIProvider implementation.
func (f *ProviderFactory) RegisterLegacyProvider(name string, provider AIProvider) {
	f.legacy[name] = provider
	// Also register it as a new Provider using the adapter
	f.providers[name] = NewLegacyProviderAdapter(provider)
}

// GetProvider retrieves a Provider by name.
func (f *ProviderFactory) GetProvider(name string) (Provider, bool) {
	provider, exists := f.providers[name]
	return provider, exists
}

// GetLegacyProvider retrieves a legacy AIProvider by name.
func (f *ProviderFactory) GetLegacyProvider(name string) (AIProvider, bool) {
	provider, exists := f.legacy[name]
	return provider, exists
}
