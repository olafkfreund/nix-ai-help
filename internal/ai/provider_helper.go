package ai

import (
"fmt"
"nix-ai-help/internal/config"
"nix-ai-help/pkg/logger"
)

// GetProvider is a convenience function to get an AI provider with the specified
// provider name and model. This creates a provider manager internally.
func GetProvider(providerName, modelName string) (Provider, error) {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// If provider name is empty, use the default from config
	if providerName == "" {
		providerName = cfg.AIProvider
	}

	// If model name is empty, use the default from config
	if modelName == "" {
		modelName = cfg.AIModel
	}

	// Create provider manager
	pm := NewProviderManager(cfg, logger.NewLogger())

	// Get provider with specified model
	if modelName != "" {
		return pm.GetProviderWithModel(providerName, modelName)
	}
	
	// Get provider with default model
	return pm.GetProvider(providerName)
}
