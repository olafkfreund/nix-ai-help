package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ModelRegistry provides functions for managing AI models and providers
type ModelRegistry struct {
	config *UserConfig
}

// NewModelRegistry creates a new model registry instance
func NewModelRegistry(config *UserConfig) *ModelRegistry {
	return &ModelRegistry{config: config}
}

// GetProvider returns the configuration for a specific provider
func (mr *ModelRegistry) GetProvider(providerName string) (*AIProviderConfig, error) {
	provider, exists := mr.config.AIModels.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", providerName)
	}
	return &provider, nil
}

// GetModel returns the configuration for a specific model within a provider
func (mr *ModelRegistry) GetModel(providerName, modelName string) (*AIModelConfig, error) {
	provider, err := mr.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	model, exists := provider.Models[modelName]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in provider '%s'", modelName, providerName)
	}
	return &model, nil
}

// GetAvailableProviders returns a list of all configured providers
func (mr *ModelRegistry) GetAvailableProviders() []string {
	providers := make([]string, 0, len(mr.config.AIModels.Providers))
	for name := range mr.config.AIModels.Providers {
		providers = append(providers, name)
	}
	return providers
}

// GetAvailableModels returns a list of all models for a specific provider
func (mr *ModelRegistry) GetAvailableModels(providerName string) ([]string, error) {
	provider, err := mr.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	models := make([]string, 0, len(provider.Models))
	for name := range provider.Models {
		models = append(models, name)
	}
	return models, nil
}

// IsProviderAvailable checks if a provider is available and properly configured
func (mr *ModelRegistry) IsProviderAvailable(providerName string) bool {
	provider, err := mr.GetProvider(providerName)
	if err != nil {
		return false
	}

	// Check if provider is available
	if !provider.Available {
		return false
	}

	// Check API key if required
	if provider.RequiresAPIKey {
		apiKey := os.Getenv(provider.EnvVar)
		if apiKey == "" {
			return false
		}
	}

	return true
}

// GetRecommendedModelForTask returns the recommended model for a specific task
func (mr *ModelRegistry) GetRecommendedModelForTask(task string) (string, string, error) {
	preferences := mr.config.AIModels.SelectionPreferences

	var modelSpec string
	switch task {
	case "question_answering":
		// For backward compatibility, use the general task model
		if taskPref, exists := preferences.TaskModels["general_help"]; exists && len(taskPref.Primary) > 0 {
			modelSpec = taskPref.Primary[0]
		} else {
			modelSpec = preferences.DefaultModels[preferences.DefaultProvider]
		}
	case "log_analysis":
		if taskPref, exists := preferences.TaskModels["nixos_config"]; exists && len(taskPref.Primary) > 0 {
			modelSpec = taskPref.Primary[0]
		} else {
			modelSpec = preferences.DefaultModels[preferences.DefaultProvider]
		}
	case "code_generation":
		if taskPref, exists := preferences.TaskModels["code_generation"]; exists && len(taskPref.Primary) > 0 {
			modelSpec = taskPref.Primary[0]
		} else {
			modelSpec = preferences.DefaultModels[preferences.DefaultProvider]
		}
	case "documentation":
		if taskPref, exists := preferences.TaskModels["general_help"]; exists && len(taskPref.Primary) > 0 {
			modelSpec = taskPref.Primary[0]
		} else {
			modelSpec = preferences.DefaultModels[preferences.DefaultProvider]
		}
	default:
		modelSpec = preferences.DefaultModels[preferences.DefaultProvider]
	}

	// Parse model specification (format: "provider:model")
	parts := strings.Split(modelSpec, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid model specification format: %s", modelSpec)
	}

	providerName, modelName := parts[0], parts[1]

	// Validate that the provider and model exist
	_, err := mr.GetModel(providerName, modelName)
	if err != nil {
		return "", "", fmt.Errorf("recommended model not available: %w", err)
	}

	return providerName, modelName, nil
}

// GetModelSpecs returns detailed specifications for a model
func (mr *ModelRegistry) GetModelSpecs(providerName, modelName string) (*AIModelConfig, error) {
	return mr.GetModel(providerName, modelName)
}

// ValidateConfiguration validates the AI models configuration
func (mr *ModelRegistry) ValidateConfiguration() error {
	// Check that at least one provider is configured
	if len(mr.config.AIModels.Providers) == 0 {
		return fmt.Errorf("no AI providers configured")
	}

	// Validate each provider
	for providerName, provider := range mr.config.AIModels.Providers {
		if err := mr.validateProvider(providerName, &provider); err != nil {
			return fmt.Errorf("provider '%s' validation failed: %w", providerName, err)
		}
	}

	// Validate selection preferences
	if err := mr.validateSelectionPreferences(); err != nil {
		return fmt.Errorf("selection preferences validation failed: %w", err)
	}

	return nil
}

// validateProvider validates a single provider configuration
func (mr *ModelRegistry) validateProvider(providerName string, provider *AIProviderConfig) error {
	// Check that at least one model is configured
	if len(provider.Models) == 0 {
		return fmt.Errorf("no models configured for provider")
	}

	// Validate each model
	for modelName, model := range provider.Models {
		if err := mr.validateModel(modelName, &model); err != nil {
			return fmt.Errorf("model '%s' validation failed: %w", modelName, err)
		}
	}

	// Check API key environment variable if required
	if provider.RequiresAPIKey && provider.EnvVar == "" {
		return fmt.Errorf("API key environment variable not specified for provider requiring authentication")
	}

	return nil
}

// validateModel validates a single model configuration
func (mr *ModelRegistry) validateModel(modelName string, model *AIModelConfig) error {
	// Check required fields
	if model.Name == "" {
		return fmt.Errorf("model name is required")
	}

	if model.ContextWindow <= 0 {
		return fmt.Errorf("context window must be positive")
	}

	if model.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}

	return nil
}

// validateSelectionPreferences validates the selection preferences
func (mr *ModelRegistry) validateSelectionPreferences() error {
	preferences := mr.config.AIModels.SelectionPreferences

	// Validate default provider
	if preferences.DefaultProvider != "" {
		_, err := mr.GetProvider(preferences.DefaultProvider)
		if err != nil {
			return fmt.Errorf("default provider not available: %w", err)
		}
	}

	// Validate default models
	for provider, model := range preferences.DefaultModels {
		if model != "" {
			_, err := mr.GetModel(provider, model)
			if err != nil {
				return fmt.Errorf("default model %s/%s not available: %w", provider, model, err)
			}
		}
	}

	// Validate task models
	for taskName, taskPref := range preferences.TaskModels {
		for _, spec := range taskPref.Primary {
			if spec != "" {
				parts := strings.Split(spec, ":")
				if len(parts) != 2 {
					return fmt.Errorf("invalid model specification format for task %s: %s", taskName, spec)
				}

				providerName, modelName := parts[0], parts[1]
				_, err := mr.GetModel(providerName, modelName)
				if err != nil {
					return fmt.Errorf("task model not available for %s: %w", taskName, err)
				}
			}
		}

		for _, spec := range taskPref.Fallback {
			if spec != "" {
				parts := strings.Split(spec, ":")
				if len(parts) != 2 {
					return fmt.Errorf("invalid model specification format for task %s fallback: %s", taskName, spec)
				}

				providerName, modelName := parts[0], parts[1]
				_, err := mr.GetModel(providerName, modelName)
				if err != nil {
					return fmt.Errorf("fallback model not available for %s: %w", taskName, err)
				}
			}
		}
	}

	return nil
}

// GetProviderInfo returns detailed information about a provider
func (mr *ModelRegistry) GetProviderInfo(providerName string) (map[string]interface{}, error) {
	provider, err := mr.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":             providerName,
		"available":        provider.Available,
		"base_url":         provider.BaseURL,
		"requires_api_key": provider.RequiresAPIKey,
		"env_var":          provider.EnvVar,
		"model_count":      len(provider.Models),
		"is_available":     mr.IsProviderAvailable(providerName),
	}

	// Add model list
	models := make([]string, 0, len(provider.Models))
	for name := range provider.Models {
		models = append(models, name)
	}
	info["models"] = models

	return info, nil
}

// GetModelInfo returns detailed information about a specific model
func (mr *ModelRegistry) GetModelInfo(providerName, modelName string) (map[string]interface{}, error) {
	model, err := mr.GetModel(providerName, modelName)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":            model.Name,
		"description":     model.Description,
		"context_window":  model.ContextWindow,
		"max_tokens":      model.MaxTokens,
		"recommended_for": model.RecommendedFor,
		"available":       mr.IsProviderAvailable(providerName),
	}

	return info, nil
}

// CheckProviderStatus checks the status of a provider
func (mr *ModelRegistry) CheckProviderStatus(providerName string) map[string]interface{} {
	status := map[string]interface{}{
		"provider":     providerName,
		"configured":   false,
		"enabled":      false,
		"api_key_set":  false,
		"available":    false,
		"last_checked": time.Now().Format(time.RFC3339),
	}

	provider, err := mr.GetProvider(providerName)
	if err != nil {
		status["error"] = err.Error()
		return status
	}

	status["configured"] = true
	status["enabled"] = provider.Available

	if provider.RequiresAPIKey {
		apiKey := os.Getenv(provider.EnvVar)
		status["api_key_set"] = apiKey != ""
		status["api_key_env_var"] = provider.EnvVar
	} else {
		status["api_key_set"] = true // Not required
	}

	status["available"] = mr.IsProviderAvailable(providerName)

	return status
}

// GetDiscoveryConfig returns the discovery configuration
func (mr *ModelRegistry) GetDiscoveryConfig() *AIDiscoveryConfig {
	return &mr.config.AIModels.Discovery
}

// ListAllModels returns a comprehensive list of all configured models
func (mr *ModelRegistry) ListAllModels() map[string][]string {
	result := make(map[string][]string)

	for providerName, provider := range mr.config.AIModels.Providers {
		models := make([]string, 0, len(provider.Models))
		for modelName := range provider.Models {
			models = append(models, modelName)
		}
		result[providerName] = models
	}

	return result
}
