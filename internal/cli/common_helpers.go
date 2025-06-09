// Package cli provides the command-line interface for nixai
package cli

import (
	"context"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// GetAIProviderManager creates and returns a provider manager using the configuration system
func GetAIProviderManager(cfg *config.UserConfig, log *logger.Logger) *ai.ProviderManager {
	return ai.NewProviderManager(cfg, log)
}

// GetLegacyAIProvider gets a legacy AIProvider using the new ProviderManager system
func GetLegacyAIProvider(cfg *config.UserConfig, log *logger.Logger) (ai.AIProvider, error) {
	manager := ai.NewProviderManager(cfg, log)

	// Get the configured default provider or fall back to ollama
	defaultProvider := cfg.AIModels.SelectionPreferences.DefaultProvider
	if defaultProvider == "" {
		defaultProvider = "ollama"
	}

	provider, err := manager.GetProvider(defaultProvider)
	if err != nil {
		// Fall back to ollama legacy provider on error
		return ai.NewOllamaLegacyProvider("llama3"), nil
	}

	// Use NewProviderWrapper to convert Provider to AIProvider
	return &ProviderToLegacyAdapter{provider: provider}, nil
}

// ProviderToLegacyAdapter adapts a Provider to the legacy AIProvider interface
type ProviderToLegacyAdapter struct {
	provider ai.Provider
}

// Query implements the legacy AIProvider interface
func (p *ProviderToLegacyAdapter) Query(prompt string) (string, error) {
	return p.provider.Query(context.Background(), prompt)
}

// InitializeAIProvider creates the appropriate AI provider based on configuration
// Deprecated: Use GetLegacyAIProvider() for new code
func InitializeAIProvider(cfg *config.UserConfig) ai.AIProvider {
	provider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		// Fall back to ollama legacy provider on error
		return ai.NewOllamaLegacyProvider("llama3")
	}
	return provider
}

// SummarizeBuildOutput extracts error messages from build output
func SummarizeBuildOutput(output string) string {
	lines := strings.Split(output, "\n")
	var summary []string
	for _, line := range lines {
		if strings.Contains(line, "error:") ||
			strings.Contains(line, "failed") ||
			strings.Contains(line, "cannot") {
			summary = append(summary, line)
		}
	}
	return strings.Join(summary, "\n")
}
