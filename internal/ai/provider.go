package ai

import "context"

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	Content      string
	Done         bool
	Error        error
	TokensUsed   int
	PartialSaved bool // Indicates if partial response was saved for recovery
}

// Provider defines the interface that all AI providers must implement.
// This interface supports both simple queries and context-aware operations.
type Provider interface {
	// Legacy Query method for backward compatibility
	Query(prompt string) (string, error)
	// Context-aware methods
	GenerateResponse(ctx context.Context, prompt string) (string, error)
	// New streaming method
	StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error)
	// Method to get partial response on token limit or other failures
	GetPartialResponse() string
}

// AIProvider is the legacy interface for backward compatibility.
// New code should use Provider interface instead.
type AIProvider interface {
	Query(prompt string) (string, error)
}

// LegacyProviderAdapter wraps an AIProvider to implement the new Provider interface
type LegacyProviderAdapter struct {
	legacy      AIProvider
	lastPartial string
}

// NewLegacyProviderAdapter creates an adapter that wraps a legacy AIProvider
func NewLegacyProviderAdapter(legacy AIProvider) Provider {
	return &LegacyProviderAdapter{
		legacy: legacy,
	}
}

// Query implements the Provider interface by delegating to the legacy provider
func (a *LegacyProviderAdapter) Query(prompt string) (string, error) {
	result, err := a.legacy.Query(prompt)
	if err != nil {
		// Save partial result if any
		a.lastPartial = result
	}
	return result, err
}

// GenerateResponse implements the Provider interface by delegating to the legacy provider
func (a *LegacyProviderAdapter) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Most legacy providers don't have a separate GenerateResponse method, so use Query
	if gr, ok := a.legacy.(interface{ GenerateResponse(string) (string, error) }); ok {
		result, err := gr.GenerateResponse(prompt)
		if err != nil {
			a.lastPartial = result
		}
		return result, err
	}
	return a.Query(prompt)
}

// StreamResponse implements basic streaming by chunking the full response
func (a *LegacyProviderAdapter) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 1)

	go func() {
		defer close(responseChan)

		result, err := a.legacy.Query(prompt)
		if err != nil {
			a.lastPartial = result
			responseChan <- StreamResponse{
				Content:      result,
				Error:        err,
				Done:         true,
				PartialSaved: result != "",
			}
			return
		}

		// Send the complete response as a single chunk for legacy providers
		responseChan <- StreamResponse{
			Content: result,
			Done:    true,
		}
	}()

	return responseChan, nil
}

// GetPartialResponse returns the last partial response saved during errors
func (a *LegacyProviderAdapter) GetPartialResponse() string {
	return a.lastPartial
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
