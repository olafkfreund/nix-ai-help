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
func (w *ProviderWrapper) Query(prompt string) (string, error) {
	return w.legacy.Query(prompt)
}

// QueryWithContext provides context-aware querying for ProviderWrapper
func (w *ProviderWrapper) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	// Context is ignored for legacy providers
	return w.legacy.Query(prompt)
}

// GenerateResponse implements the Provider interface by delegating to the legacy provider
func (w *ProviderWrapper) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Check if the legacy provider has a GenerateResponse method
	if gr, ok := w.legacy.(interface{ GenerateResponse(string) (string, error) }); ok {
		return gr.GenerateResponse(prompt)
	}
	// Otherwise use QueryWithContext
	return w.QueryWithContext(ctx, prompt)
}

// StreamResponse implements streaming by checking if the legacy provider supports it
func (w *ProviderWrapper) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	// Check if the legacy provider supports streaming
	if streamer, ok := w.legacy.(interface {
		StreamResponse(context.Context, string) (<-chan StreamResponse, error)
	}); ok {
		return streamer.StreamResponse(ctx, prompt)
	}

	// Fall back to basic streaming by chunking the full response
	responseChan := make(chan StreamResponse, 1)

	go func() {
		defer close(responseChan)

		result, err := w.legacy.Query(prompt)
		if err != nil {
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

// GetPartialResponse returns partial response if the legacy provider supports it
func (w *ProviderWrapper) GetPartialResponse() string {
	if partialProvider, ok := w.legacy.(interface{ GetPartialResponse() string }); ok {
		return partialProvider.GetPartialResponse()
	}
	return ""
}
