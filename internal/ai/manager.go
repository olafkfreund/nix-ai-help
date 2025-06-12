package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// ProviderManager manages AI providers using the configuration system.
type ProviderManager struct {
	registry  *config.ModelRegistry
	config    *config.UserConfig
	providers map[string]Provider // Cache of initialized providers
	logger    *logger.Logger
}

// NewProviderManager creates a new provider manager with the given configuration.
func NewProviderManager(cfg *config.UserConfig, log *logger.Logger) *ProviderManager {
	if log == nil {
		log = logger.NewLogger()
	}

	registry := config.NewModelRegistry(cfg)

	return &ProviderManager{
		registry:  registry,
		config:    cfg,
		providers: make(map[string]Provider),
		logger:    log,
	}
}

// GetProvider retrieves or initializes a provider by name.
func (pm *ProviderManager) GetProvider(providerName string) (Provider, error) {
	// Check cache first
	if provider, exists := pm.providers[providerName]; exists {
		return provider, nil
	}

	// Check if provider is available in configuration
	if !pm.registry.IsProviderAvailable(providerName) {
		return nil, fmt.Errorf("provider '%s' is not available or configured", providerName)
	}

	// Initialize the provider
	provider, err := pm.initializeProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider '%s': %w", providerName, err)
	}

	// Cache the provider
	pm.providers[providerName] = provider
	pm.logger.Info(fmt.Sprintf("Initialized AI provider: %s", providerName))

	return provider, nil
}

// GetProviderWithModel retrieves a provider configured for a specific model.
func (pm *ProviderManager) GetProviderWithModel(providerName, modelName string) (Provider, error) {
	// Validate that the model exists for this provider
	model, err := pm.registry.GetModel(providerName, modelName)
	if err != nil {
		return nil, fmt.Errorf("model '%s' not found for provider '%s': %w", modelName, providerName, err)
	}

	// Get the provider
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// For providers that support model selection, we'll need to wrap them
	// This will be implemented based on the specific provider interface
	pm.logger.Debug(fmt.Sprintf("Using model '%s' with provider '%s'", model.Name, providerName))

	return provider, nil
}

// GetDefaultProvider retrieves the default provider as configured.
func (pm *ProviderManager) GetDefaultProvider() (Provider, error) {
	defaultProvider := pm.config.AIModels.SelectionPreferences.DefaultProvider
	if defaultProvider == "" {
		defaultProvider = "ollama" // Final fallback
	}

	return pm.GetProvider(defaultProvider)
}

// GetRecommendedProvider retrieves the recommended provider for a specific task.
func (pm *ProviderManager) GetRecommendedProvider(task string) (Provider, string, error) {
	// Get recommended model for the task
	providerName, modelName, err := pm.registry.GetRecommendedModelForTask(task)
	if err != nil {
		// Fall back to default provider
		provider, err := pm.GetDefaultProvider()
		if err != nil {
			return nil, "", err
		}
		return provider, "", err
	}

	// Get provider with specific model
	provider, err := pm.GetProviderWithModel(providerName, modelName)
	if err != nil {
		return nil, "", err
	}

	return provider, modelName, nil
}

// GetAvailableProviders returns a list of all available providers.
func (pm *ProviderManager) GetAvailableProviders() []string {
	return pm.registry.GetAvailableProviders()
}

// GetAvailableModels returns a list of all available models for a provider.
func (pm *ProviderManager) GetAvailableModels(providerName string) ([]string, error) {
	return pm.registry.GetAvailableModels(providerName)
}

// GetProviderInfo returns information about a specific provider.
func (pm *ProviderManager) GetProviderInfo(providerName string) (*config.AIProviderConfig, error) {
	return pm.registry.GetProvider(providerName)
}

// GetModelInfo returns information about a specific model.
func (pm *ProviderManager) GetModelInfo(providerName, modelName string) (*config.AIModelConfig, error) {
	return pm.registry.GetModel(providerName, modelName)
}

// CheckProviderStatus checks the status of a provider (e.g., if it's running).
func (pm *ProviderManager) CheckProviderStatus(providerName string) (bool, error) {
	available := pm.registry.IsProviderAvailable(providerName)
	return available, nil
}

// ValidateConfiguration validates the current AI models configuration.
func (pm *ProviderManager) ValidateConfiguration() error {
	return pm.registry.ValidateConfiguration()
}

// RefreshProviders clears the provider cache, forcing reinitialization.
func (pm *ProviderManager) RefreshProviders() {
	pm.providers = make(map[string]Provider)
	pm.logger.Info("Provider cache cleared")
}

// initializeProvider creates a new provider instance based on configuration.
func (pm *ProviderManager) initializeProvider(providerName string) (Provider, error) {
	providerConfig, err := pm.registry.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	switch providerName {
	case "ollama":
		return pm.initializeOllamaProvider(providerConfig)
	case "gemini":
		return pm.initializeGeminiProvider(providerConfig)
	case "openai":
		return pm.initializeOpenAIProvider(providerConfig)
	case "llamacpp":
		return pm.initializeLlamaCppProvider(providerConfig)
	case "custom":
		return pm.initializeCustomProvider(providerConfig)
	case "claude":
		return pm.initializeClaudeProvider(providerConfig)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerName)
	}
}

// initializeOllamaProvider creates an Ollama provider instance.
func (pm *ProviderManager) initializeOllamaProvider(config *config.AIProviderConfig) (Provider, error) {
	// Get default model for Ollama
	defaultModel := pm.config.AIModels.SelectionPreferences.DefaultModels["ollama"]
	if defaultModel == "" {
		defaultModel = "llama3" // fallback
	}

	// Set custom endpoint if configured
	if config.BaseURL != "" {
		os.Setenv("OLLAMA_ENDPOINT", config.BaseURL+"/api/generate")
	}

	ollamaProvider := NewOllamaProvider(defaultModel)

	// Apply configured timeout
	timeout := pm.config.GetAITimeout("ollama")
	ollamaProvider.SetTimeout(timeout)

	pm.logger.Debug(fmt.Sprintf("Ollama provider initialized with %v timeout", timeout))

	// Create legacy wrapper and then wrap that as Provider
	legacyProvider := &OllamaLegacyProvider{OllamaProvider: ollamaProvider}
	return NewProviderWrapper(legacyProvider), nil
}

// initializeGeminiProvider creates a Gemini provider instance.
func (pm *ProviderManager) initializeGeminiProvider(config *config.AIProviderConfig) (Provider, error) {
	apiKey := os.Getenv(config.EnvVar)
	if apiKey == "" && config.RequiresAPIKey {
		return nil, fmt.Errorf("gemini API key not found in environment variable %s", config.EnvVar)
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent"
	}

	// Get default model for Gemini
	defaultModel := pm.config.AIModels.SelectionPreferences.DefaultModels["gemini"]
	if defaultModel == "" {
		defaultModel = "gemini-pro" // fallback
	}

	geminiClient := NewGeminiClientWithModel(apiKey, baseURL, defaultModel)
	return NewProviderWrapper(geminiClient), nil
}

// initializeOpenAIProvider creates an OpenAI provider instance.
func (pm *ProviderManager) initializeOpenAIProvider(config *config.AIProviderConfig) (Provider, error) {
	apiKey := os.Getenv(config.EnvVar)
	if apiKey == "" && config.RequiresAPIKey {
		return nil, fmt.Errorf("openAI API key not found in environment variable %s", config.EnvVar)
	}

	// Get default model for OpenAI
	defaultModel := pm.config.AIModels.SelectionPreferences.DefaultModels["openai"]
	if defaultModel == "" {
		defaultModel = "gpt-3.5-turbo" // fallback
	}

	openaiClient := NewOpenAIClientWithModel(apiKey, defaultModel)
	return NewProviderWrapper(openaiClient), nil
}

// initializeClaudeProvider creates a Claude provider instance.
func (pm *ProviderManager) initializeClaudeProvider(config *config.AIProviderConfig) (Provider, error) {
	apiKey := os.Getenv(config.EnvVar)
	if apiKey == "" && config.RequiresAPIKey {
		return nil, fmt.Errorf("claude API key not found in environment variable %s", config.EnvVar)
	}

	// Get default model for Claude
	defaultModel := pm.config.AIModels.SelectionPreferences.DefaultModels["claude"]
	if defaultModel == "" {
		defaultModel = "claude-3-5-sonnet-20241022" // fallback to latest Claude 3.5 Sonnet
	}

	claudeProvider := NewClaudeProvider(defaultModel)

	// Set custom base URL if configured
	if config.BaseURL != "" {
		claudeProvider.BaseURL = config.BaseURL
	}

	// Apply configured timeout
	timeout := pm.config.GetAITimeout("claude")
	claudeProvider.SetTimeout(timeout)

	pm.logger.Debug(fmt.Sprintf("Claude provider initialized with model %s and %v timeout", defaultModel, timeout))

	// Create legacy wrapper for compatibility
	legacyProvider := &ClaudeLegacyProvider{ClaudeProvider: claudeProvider}
	return NewProviderWrapper(legacyProvider), nil
}

// initializeLlamaCppProvider creates a LlamaCpp provider instance.
func (pm *ProviderManager) initializeLlamaCppProvider(config *config.AIProviderConfig) (Provider, error) {
	// Get default model for LlamaCpp
	defaultModel := pm.config.AIModels.SelectionPreferences.DefaultModels["llamacpp"]
	if defaultModel == "" {
		defaultModel = "llama3" // fallback
	}

	// Use the new model-aware constructor
	llamacppProvider, err := NewLlamaCppProviderWithModel(config, defaultModel)
	if err != nil {
		// Fall back to simple constructor if model-aware fails
		llamacppProvider = NewLlamaCppProvider(defaultModel)
	}

	// Apply configured timeout
	timeout := pm.config.GetAITimeout("llamacpp")
	llamacppProvider.SetTimeout(timeout)

	pm.logger.Debug(fmt.Sprintf("LlamaCpp provider initialized with %v timeout", timeout))

	return NewProviderWrapper(llamacppProvider), nil
}

// initializeCustomProvider creates a custom provider instance.
func (pm *ProviderManager) initializeCustomProvider(config *config.AIProviderConfig) (Provider, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("custom provider requires a base URL")
	}

	// Get default model for Custom provider
	defaultModel := pm.config.AIModels.SelectionPreferences.DefaultModels["custom"]
	if defaultModel == "" {
		// Find first available model in configuration
		for modelName := range config.Models {
			defaultModel = modelName
			break
		}
		if defaultModel == "" {
			defaultModel = "default" // fallback
		}
	}

	// Use the new model-aware constructor
	customProvider, err := NewCustomProviderWithModel(config, defaultModel)
	if err != nil {
		// Fall back to simple constructor if model-aware fails
		var headers map[string]string
		if pm.config.CustomAI.Headers != nil {
			headers = pm.config.CustomAI.Headers
		} else {
			headers = make(map[string]string)
		}
		customProvider = NewCustomProvider(config.BaseURL, headers)
	}

	// Apply configured timeout
	timeout := pm.config.GetAITimeout("custom")
	customProvider.SetTimeout(timeout)

	pm.logger.Debug(fmt.Sprintf("Custom provider initialized with %v timeout", timeout))

	return NewProviderWrapper(customProvider), nil
}

// parseModelReference parses a model reference in the format "provider:model".
func parseModelReference(modelRef string) (provider, model string, err error) {
	// Handle empty reference
	if modelRef == "" {
		return "ollama", "llama3", nil
	}

	// Check if it contains provider:model format
	parts := strings.Split(modelRef, ":")
	if len(parts) == 2 {
		provider = strings.TrimSpace(parts[0])
		model = strings.TrimSpace(parts[1])

		// Validate provider and model are not empty
		if provider == "" || model == "" {
			return "", "", fmt.Errorf("invalid model reference format: '%s'", modelRef)
		}

		return provider, model, nil
	} else if len(parts) == 1 {
		// Just a model name, use default provider
		model = strings.TrimSpace(parts[0])
		if model == "" {
			return "", "", fmt.Errorf("empty model name in reference: '%s'", modelRef)
		}
		return "ollama", model, nil
	}

	return "", "", fmt.Errorf("invalid model reference format: '%s' (expected 'provider:model' or 'model')", modelRef)
}

// Legacy compatibility methods

// CreateLegacyProvider creates a legacy AIProvider for backward compatibility.
func (pm *ProviderManager) CreateLegacyProvider(providerName string) (AIProvider, error) {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// If it's a wrapper, extract the legacy provider
	if wrapper, ok := provider.(*ProviderWrapper); ok {
		return wrapper.legacy, nil
	}

	// If it's already adapted, extract the legacy provider
	if adapter, ok := provider.(*LegacyProviderAdapter); ok {
		return adapter.legacy, nil
	}

	// Otherwise, create a reverse adapter (Provider -> AIProvider)
	return &ProviderToLegacyAdapter{provider: provider}, nil
}

// ProviderToLegacyAdapter adapts a new Provider to the legacy AIProvider interface.
type ProviderToLegacyAdapter struct {
	provider Provider
}

// Query implements the legacy AIProvider interface.
func (a *ProviderToLegacyAdapter) Query(prompt string) (string, error) {
	return a.provider.Query(context.Background(), prompt)
}

// HealthChecker interface for providers that support health checking
type HealthChecker interface {
	HealthCheck() error
}

// GetHealthyProvider retrieves a provider and ensures it's healthy, with fallback
func (pm *ProviderManager) GetHealthyProvider(providerName string) (Provider, error) {
	provider, err := pm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Check health if provider supports it
	if healthChecker, ok := provider.(HealthChecker); ok {
		if err := healthChecker.HealthCheck(); err != nil {
			pm.logger.Warn(fmt.Sprintf("Provider %s failed health check: %v", providerName, err))

			// Try fallback providers
			fallbackProviders := pm.getFallbackProviders(providerName)
			for _, fallback := range fallbackProviders {
				pm.logger.Info(fmt.Sprintf("Trying fallback provider: %s", fallback))
				if fallbackProvider, err := pm.GetProvider(fallback); err == nil {
					if fallbackChecker, ok := fallbackProvider.(HealthChecker); ok {
						if err := fallbackChecker.HealthCheck(); err == nil {
							pm.logger.Info(fmt.Sprintf("Successfully fell back to provider: %s", fallback))
							return fallbackProvider, nil
						}
					} else {
						// Assume healthy if no health check method
						pm.logger.Info(fmt.Sprintf("Successfully fell back to provider: %s", fallback))
						return fallbackProvider, nil
					}
				}
			}

			return nil, fmt.Errorf("provider %s is unhealthy and no fallback available: %w", providerName, err)
		}
	}

	return provider, nil
}

// getFallbackProviders returns a list of fallback providers for the given provider
func (pm *ProviderManager) getFallbackProviders(providerName string) []string {
	// Get fallback providers from configuration or use defaults
	var fallbacks []string

	// Check if there are task-specific fallbacks configured
	for _, taskPrefs := range pm.config.AIModels.SelectionPreferences.TaskModels {
		for _, fallback := range taskPrefs.Fallback {
			if provider, _, err := parseModelReference(fallback); err == nil && provider != providerName {
				fallbacks = append(fallbacks, provider)
			}
		}
	}

	// Add default fallbacks if none configured
	if len(fallbacks) == 0 {
		switch providerName {
		case "gemini", "openai", "claude":
			fallbacks = []string{"ollama"}
		case "ollama":
			fallbacks = []string{"claude", "gemini", "openai"}
		case "llamacpp":
			fallbacks = []string{"ollama", "claude"}
		case "custom":
			fallbacks = []string{"ollama", "claude"}
		}
	}

	// Remove duplicates and the original provider
	seen := make(map[string]bool)
	var uniqueFallbacks []string
	for _, fb := range fallbacks {
		if !seen[fb] && fb != providerName {
			seen[fb] = true
			uniqueFallbacks = append(uniqueFallbacks, fb)
		}
	}

	return uniqueFallbacks
}

// GetProviderForTask retrieves the best provider for a specific task with fallback
func (pm *ProviderManager) GetProviderForTask(task string) (Provider, string, error) {
	// First try to get recommended provider for task
	if _, model, err := pm.GetRecommendedProvider(task); err == nil {
		// Try to get healthy version
		providerName := getProviderFromModel(model)
		if healthyProvider, err := pm.GetHealthyProvider(providerName); err == nil {
			return healthyProvider, model, nil
		}
		pm.logger.Warn(fmt.Sprintf("Recommended provider for task %s is unhealthy, trying alternatives", task))
	}

	// Fall back to default provider
	defaultProvider, err := pm.GetDefaultProvider()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get default provider: %w", err)
	}

	return defaultProvider, "", nil
}

// getProviderFromModel extracts provider name from model reference
func getProviderFromModel(model string) string {
	if provider, _, err := parseModelReference(model); err == nil {
		return provider
	}
	return "ollama" // default fallback
}
