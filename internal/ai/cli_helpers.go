package ai

import (
	"context"
	"fmt"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// CLIProviderManager is a helper struct that simplifies AI provider usage in CLI commands
type CLIProviderManager struct {
	manager *ProviderManager
	logger  *logger.Logger
}

// NewCLIProviderManager creates a new CLI provider manager
func NewCLIProviderManager(cfg *config.UserConfig, log *logger.Logger) *CLIProviderManager {
	if log == nil {
		log = logger.NewLogger()
	}

	return &CLIProviderManager{
		manager: NewProviderManager(cfg, log),
		logger:  log,
	}
}

// GetProviderForCLI retrieves the best provider for CLI usage with automatic fallback
func (cpm *CLIProviderManager) GetProviderForCLI() (Provider, error) {
	// Try to get healthy default provider
	defaultProvider, err := cpm.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get default provider: %w", err)
	}

	// Check if it's healthy
	if healthyProvider, err := cpm.manager.GetHealthyProvider(getProviderName(defaultProvider)); err == nil {
		return healthyProvider, nil
	}

	// Fall back to any available provider
	availableProviders := cpm.manager.GetAvailableProviders()
	for _, providerName := range availableProviders {
		if provider, err := cpm.manager.GetHealthyProvider(providerName); err == nil {
			cpm.logger.Info(fmt.Sprintf("Fell back to provider: %s", providerName))
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no healthy providers available")
}

// GetProviderForCLITask retrieves the best provider for a specific CLI task
func (cpm *CLIProviderManager) GetProviderForCLITask(task string) (Provider, error) {
	provider, _, err := cpm.manager.GetProviderForTask(task)
	return provider, err
}

// GetLegacyProviderForCLI retrieves a legacy AIProvider for CLI commands that haven't been updated yet
func (cpm *CLIProviderManager) GetLegacyProviderForCLI() (AIProvider, error) {
	provider, err := cpm.GetProviderForCLI()
	if err != nil {
		return nil, err
	}

	// If it's a wrapper, extract the legacy provider
	if wrapper, ok := provider.(*ProviderWrapper); ok {
		return wrapper.legacy, nil
	}

	// If it's already a legacy adapter, extract the legacy provider
	if adapter, ok := provider.(*LegacyProviderAdapter); ok {
		return adapter.legacy, nil
	}

	// Otherwise, create a reverse adapter
	return &ProviderToLegacyAdapter{provider: provider}, nil
}

// QueryWithProgress performs a query with progress indication
func (cpm *CLIProviderManager) QueryWithProgress(ctx context.Context, prompt string, progressCallback func(string)) (string, error) {
	if progressCallback != nil {
		progressCallback("Getting AI provider...")
	}

	provider, err := cpm.GetProviderForCLI()
	if err != nil {
		return "", err
	}

	if progressCallback != nil {
		progressCallback("Querying AI...")
	}

	response, err := provider.Query(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI query failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback("Done")
	}

	return response, nil
}

// LegacyQueryWithProgress performs a legacy query with progress indication
func (cpm *CLIProviderManager) LegacyQueryWithProgress(prompt string, progressCallback func(string)) (string, error) {
	if progressCallback != nil {
		progressCallback("Getting AI provider...")
	}

	provider, err := cpm.GetLegacyProviderForCLI()
	if err != nil {
		return "", err
	}

	if progressCallback != nil {
		progressCallback("Querying AI...")
	}

	response, err := provider.Query(prompt)
	if err != nil {
		return "", fmt.Errorf("AI query failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback("Done")
	}

	return response, nil
}

// getProviderName attempts to determine the provider name from a Provider instance
func getProviderName(provider Provider) string {
	// Try to determine provider type by checking the underlying type
	switch provider.(type) {
	case *OllamaProvider:
		return "ollama"
	case *ProviderWrapper:
		// Try to check the wrapped provider
		if wrapper, ok := provider.(*ProviderWrapper); ok {
			switch wrapper.legacy.(type) {
			case *GeminiClient:
				return "gemini"
			case *OpenAIClient:
				return "openai"
			case *LlamaCppProvider:
				return "llamacpp"
			case *CustomProvider:
				return "custom"
			case *ClaudeLegacyProvider:
				return "claude"
			}
		}
	}
	return "unknown"
}

// Global instance for easy CLI usage
var GlobalCLIManager *CLIProviderManager

// InitGlobalCLIManager initializes the global CLI manager instance
func InitGlobalCLIManager(cfg *config.UserConfig, log *logger.Logger) {
	GlobalCLIManager = NewCLIProviderManager(cfg, log)
}

// GetGlobalCLIManager returns the global CLI manager instance
func GetGlobalCLIManager() *CLIProviderManager {
	return GlobalCLIManager
}

// Convenience functions for direct CLI usage

// QuickQuery performs a quick query using the global CLI manager
func QuickQuery(ctx context.Context, prompt string) (string, error) {
	if GlobalCLIManager == nil {
		return "", fmt.Errorf("global CLI manager not initialized")
	}
	return GlobalCLIManager.QueryWithProgress(ctx, prompt, nil)
}

// QuickLegacyQuery performs a quick legacy query using the global CLI manager
func QuickLegacyQuery(prompt string) (string, error) {
	if GlobalCLIManager == nil {
		return "", fmt.Errorf("global CLI manager not initialized")
	}
	return GlobalCLIManager.LegacyQueryWithProgress(prompt, nil)
}

// QuickProvider returns a provider for quick CLI usage
func QuickProvider() (Provider, error) {
	if GlobalCLIManager == nil {
		return nil, fmt.Errorf("global CLI manager not initialized")
	}
	return GlobalCLIManager.GetProviderForCLI()
}

// QuickLegacyProvider returns a legacy provider for quick CLI usage
func QuickLegacyProvider() (AIProvider, error) {
	if GlobalCLIManager == nil {
		return nil, fmt.Errorf("global CLI manager not initialized")
	}
	return GlobalCLIManager.GetLegacyProviderForCLI()
}
