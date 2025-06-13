package config

import (
	"testing"
)

func TestModelRegistryBasicFunctionality(t *testing.T) {
	// Create a test configuration
	testConfig := &UserConfig{
		AIModels: AIModelsConfig{
			Providers: map[string]AIProviderConfig{
				"ollama": {
					Available:      true,
					BaseURL:        "http://localhost:11434",
					RequiresAPIKey: false,
					EnvVar:         "",
					Models: map[string]AIModelConfig{
						"llama3": {
							Name:           "llama3",
							Description:    "Test model",
							ContextWindow:  8192,
							MaxTokens:      4096,
							RecommendedFor: []string{"general"},
						},
					},
				},
			},
			SelectionPreferences: AISelectionPreferences{
				DefaultProvider: "ollama",
				DefaultModels: map[string]string{
					"ollama": "llama3",
				},
				TaskModels: map[string]TaskModelPreferences{
					"general_help": {
						Primary:  []string{"ollama:llama3"},
						Fallback: []string{},
					},
				},
			},
			Discovery: AIDiscoveryConfig{
				AutoDiscover:  true,
				CacheDuration: 300,
				CheckTimeout:  30,
				MaxRetries:    3,
			},
		},
	}

	// Create model registry
	registry := NewModelRegistry(testConfig)

	// Test provider retrieval
	provider, err := registry.GetProvider("ollama")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected BaseURL 'http://localhost:11434', got '%s'", provider.BaseURL)
	}

	// Test model retrieval
	model, err := registry.GetModel("ollama", "llama3")
	if err != nil {
		t.Fatalf("Failed to get model: %v", err)
	}
	if model.Name != "llama3" {
		t.Errorf("Expected model name 'llama3', got '%s'", model.Name)
	}

	// Test provider availability
	if !registry.IsProviderAvailable("ollama") {
		t.Error("Expected ollama provider to be available")
	}

	// Test getting available providers
	providers := registry.GetAvailableProviders()
	if len(providers) != 1 || providers[0] != "ollama" {
		t.Errorf("Expected providers ['ollama'], got %v", providers)
	}

	// Test getting available models
	models, err := registry.GetAvailableModels("ollama")
	if err != nil {
		t.Fatalf("Failed to get available models: %v", err)
	}
	if len(models) != 1 || models[0] != "llama3" {
		t.Errorf("Expected models ['llama3'], got %v", models)
	}

	// Test recommended model for task
	providerName, modelName, err := registry.GetRecommendedModelForTask("question_answering")
	if err != nil {
		t.Fatalf("Failed to get recommended model: %v", err)
	}
	if providerName != "ollama" || modelName != "llama3" {
		t.Errorf("Expected ollama:llama3, got %s:%s", providerName, modelName)
	}

	// Test configuration validation
	if err := registry.ValidateConfiguration(); err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}

	// Test provider info
	info, err := registry.GetProviderInfo("ollama")
	if err != nil {
		t.Fatalf("Failed to get provider info: %v", err)
	}
	if info["name"] != "ollama" {
		t.Errorf("Expected provider name 'ollama', got '%v'", info["name"])
	}

	// Test model info
	modelInfo, err := registry.GetModelInfo("ollama", "llama3")
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}
	if modelInfo["name"] != "llama3" {
		t.Errorf("Expected model name 'llama3', got '%v'", modelInfo["name"])
	}

	// Test provider status check
	status := registry.CheckProviderStatus("ollama")
	if status["provider"] != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%v'", status["provider"])
	}
	if status["configured"] != true {
		t.Error("Expected provider to be configured")
	}

	// Test discovery config
	discoveryConfig := registry.GetDiscoveryConfig()
	if !discoveryConfig.AutoDiscover {
		t.Error("Expected AutoDiscover to be true")
	}

	// Test list all models
	allModels := registry.ListAllModels()
	if len(allModels) != 1 {
		t.Errorf("Expected 1 provider in all models, got %d", len(allModels))
	}
	if len(allModels["ollama"]) != 1 {
		t.Errorf("Expected 1 model for ollama, got %d", len(allModels["ollama"]))
	}
}

func TestModelRegistryErrors(t *testing.T) {
	// Create a minimal test configuration
	testConfig := &UserConfig{
		AIModels: AIModelsConfig{
			Providers: map[string]AIProviderConfig{},
			SelectionPreferences: AISelectionPreferences{
				DefaultProvider: "",
				DefaultModels:   map[string]string{},
				TaskModels:      map[string]TaskModelPreferences{},
			},
		},
	}

	registry := NewModelRegistry(testConfig)

	// Test non-existent provider
	_, err := registry.GetProvider("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent provider")
	}

	// Test non-existent model
	_, err = registry.GetModel("nonexistent", "model")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}

	// Test validation with no providers
	err = registry.ValidateConfiguration()
	if err == nil {
		t.Error("Expected validation error with no providers")
	}
}
