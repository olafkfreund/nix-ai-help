package ai

import "context"

// ProviderWrapper wraps a legacy AIProvider to implement the new Provider interface
type ProviderWrapper struct {
	legacy AIProvider
}

// NewProviderWrapper creates a new provider wrapper
func NewProviderWrapper(legacy AIProvider) Provider {
	return &ProviderWrapper{legacy: legacy}
}

// Query implements the Provider interface by delegating to the legacy provider
func (w *ProviderWrapper) Query(ctx context.Context, prompt string) (string, error) {
	// TODO: Add context cancellation support if needed
	return w.legacy.Query(prompt)
}

// GenerateResponse implements the Provider interface by delegating to the legacy provider
func (w *ProviderWrapper) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Check if the legacy provider has a GenerateResponse method
	if gr, ok := w.legacy.(interface{ GenerateResponse(string) (string, error) }); ok {
		return gr.GenerateResponse(prompt)
	}
	// Otherwise use Query
	return w.legacy.Query(prompt)
}
