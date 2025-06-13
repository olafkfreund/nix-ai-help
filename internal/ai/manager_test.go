package ai

import (
	"testing"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

func TestProviderManagerBasicFunctionality(t *testing.T) {
	// Create test configuration
	testConfig := &config.UserConfig{
		AIProvider: "ollama",
		AIModel:    "llama3",
		AIModels: config.AIModelsConfig{
			Providers: map[string]config.AIProviderConfig{
				"ollama": {
					Available:      true,
					BaseURL:        "http://localhost:11434",
					RequiresAPIKey: false,
					EnvVar:         "",
					Models: map[string]config.AIModelConfig{
						"llama3": {
							Name:           "llama3",
							Description:    "Test model",
							ContextWindow:  8192,
							MaxTokens:      4096,
							RecommendedFor: []string{"general"},
						},
					},
				},
				"gemini": {
					Available:      true,
					BaseURL:        "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent",
					RequiresAPIKey: true,
					EnvVar:         "GEMINI_API_KEY",
					Models: map[string]config.AIModelConfig{
						"gemini-pro": {
							Name:           "gemini-pro",
							Description:    "Gemini Pro model",
							ContextWindow:  30720,
							MaxTokens:      8192,
							RecommendedFor: []string{"analysis"},
						},
					},
				},
			},
			SelectionPreferences: config.AISelectionPreferences{
				DefaultProvider: "ollama",
				DefaultModels: map[string]string{
					"ollama": "llama3",
					"gemini": "gemini-pro",
				},
				TaskModels: map[string]config.TaskModelPreferences{
					"general_help": {
						Primary:  []string{"ollama:llama3"},
						Fallback: []string{"gemini:gemini-pro"},
					},
				},
			},
			Discovery: config.AIDiscoveryConfig{
				AutoDiscover:  true,
				CacheDuration: 300,
				CheckTimeout:  30,
				MaxRetries:    3,
			},
		},
	}

	// Create provider manager
	log := logger.NewLogger()
	pm := NewProviderManager(testConfig, log)

	// Test basic provider creation
	t.Run("GetDefaultProvider", func(t *testing.T) {
		provider, err := pm.GetDefaultProvider()
		if err != nil {
			t.Fatalf("Failed to get default provider: %v", err)
		}
		if provider == nil {
			t.Fatal("Default provider is nil")
		}
	})

	t.Run("GetProvider", func(t *testing.T) {
		provider, err := pm.GetProvider("ollama")
		if err != nil {
			t.Fatalf("Failed to get ollama provider: %v", err)
		}
		if provider == nil {
			t.Fatal("Ollama provider is nil")
		}
	})

	t.Run("GetAvailableProviders", func(t *testing.T) {
		providers := pm.GetAvailableProviders()
		if len(providers) == 0 {
			t.Fatal("No available providers found")
		}

		// Should include ollama and gemini
		found := make(map[string]bool)
		for _, p := range providers {
			found[p] = true
		}

		if !found["ollama"] {
			t.Error("Ollama provider not found in available providers")
		}
		if !found["gemini"] {
			t.Error("Gemini provider not found in available providers")
		}
	})

	t.Run("GetAvailableModels", func(t *testing.T) {
		models, err := pm.GetAvailableModels("ollama")
		if err != nil {
			t.Fatalf("Failed to get available models: %v", err)
		}
		if len(models) == 0 {
			t.Fatal("No available models found for ollama")
		}

		found := false
		for _, m := range models {
			if m == "llama3" {
				found = true
				break
			}
		}
		if !found {
			t.Error("llama3 model not found in available models for ollama")
		}
	})

	t.Run("GetProviderInfo", func(t *testing.T) {
		info, err := pm.GetProviderInfo("ollama")
		if err != nil {
			t.Fatalf("Failed to get provider info: %v", err)
		}
		if info == nil {
			t.Fatal("Provider info is nil")
		}
		if info.BaseURL != "http://localhost:11434" {
			t.Errorf("Expected BaseURL 'http://localhost:11434', got '%s'", info.BaseURL)
		}
	})

	t.Run("GetModelInfo", func(t *testing.T) {
		info, err := pm.GetModelInfo("ollama", "llama3")
		if err != nil {
			t.Fatalf("Failed to get model info: %v", err)
		}
		if info == nil {
			t.Fatal("Model info is nil")
		}
		if info.Name != "llama3" {
			t.Errorf("Expected model name 'llama3', got '%s'", info.Name)
		}
	})

	t.Run("ValidateConfiguration", func(t *testing.T) {
		err := pm.ValidateConfiguration()
		if err != nil {
			t.Fatalf("Configuration validation failed: %v", err)
		}
	})
}

func TestProviderManagerErrors(t *testing.T) {
	// Create test configuration with minimal setup
	testConfig := &config.UserConfig{
		AIModels: config.AIModelsConfig{
			Providers: map[string]config.AIProviderConfig{
				"test": {
					Available:      false,
					BaseURL:        "",
					RequiresAPIKey: false,
					Models:         map[string]config.AIModelConfig{},
				},
			},
			SelectionPreferences: config.AISelectionPreferences{
				DefaultProvider: "nonexistent",
			},
		},
	}

	log := logger.NewLogger()
	pm := NewProviderManager(testConfig, log)

	t.Run("GetProviderUnavailable", func(t *testing.T) {
		_, err := pm.GetProvider("test")
		if err == nil {
			t.Error("Expected error for unavailable provider")
		}
	})

	t.Run("GetProviderNonexistent", func(t *testing.T) {
		_, err := pm.GetProvider("nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent provider")
		}
	})

	t.Run("GetModelNonexistent", func(t *testing.T) {
		_, err := pm.GetModelInfo("nonexistent", "nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent model")
		}
	})
}

func TestProviderManagerCaching(t *testing.T) {
	// Create test configuration
	testConfig := &config.UserConfig{
		AIModels: config.AIModelsConfig{
			Providers: map[string]config.AIProviderConfig{
				"ollama": {
					Available:      true,
					BaseURL:        "http://localhost:11434",
					RequiresAPIKey: false,
					Models: map[string]config.AIModelConfig{
						"llama3": {
							Name:          "llama3",
							Description:   "Test model",
							ContextWindow: 8192,
						},
					},
				},
			},
			SelectionPreferences: config.AISelectionPreferences{
				DefaultProvider: "ollama",
				DefaultModels: map[string]string{
					"ollama": "llama3",
				},
			},
		},
	}

	log := logger.NewLogger()
	pm := NewProviderManager(testConfig, log)

	// Get provider twice to test caching
	provider1, err := pm.GetProvider("ollama")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}

	provider2, err := pm.GetProvider("ollama")
	if err != nil {
		t.Fatalf("Failed to get provider second time: %v", err)
	}

	// Should be the same instance due to caching
	if provider1 != provider2 {
		t.Error("Provider instances should be the same due to caching")
	}

	// Test cache refresh
	pm.RefreshProviders()

	provider3, err := pm.GetProvider("ollama")
	if err != nil {
		t.Fatalf("Failed to get provider after refresh: %v", err)
	}

	// Should be a different instance after refresh
	if provider1 == provider3 {
		t.Error("Provider instances should be different after cache refresh")
	}
}

func TestLegacyCompatibility(t *testing.T) {
	// Create test configuration
	testConfig := &config.UserConfig{
		AIModels: config.AIModelsConfig{
			Providers: map[string]config.AIProviderConfig{
				"ollama": {
					Available:      true,
					BaseURL:        "http://localhost:11434",
					RequiresAPIKey: false,
					Models: map[string]config.AIModelConfig{
						"llama3": {
							Name:        "llama3",
							Description: "Test model",
						},
					},
				},
			},
			SelectionPreferences: config.AISelectionPreferences{
				DefaultProvider: "ollama",
				DefaultModels: map[string]string{
					"ollama": "llama3",
				},
			},
		},
	}

	log := logger.NewLogger()
	pm := NewProviderManager(testConfig, log)

	t.Run("CreateLegacyProvider", func(t *testing.T) {
		legacyProvider, err := pm.CreateLegacyProvider("ollama")
		if err != nil {
			t.Fatalf("Failed to create legacy provider: %v", err)
		}
		if legacyProvider == nil {
			t.Fatal("Legacy provider is nil")
		}

		// Test that it implements the legacy interface
		_, err = legacyProvider.Query("test prompt")
		// Note: This will fail in tests since we don't have actual Ollama running,
		// but it should not panic or have interface issues
		if err == nil {
			t.Log("Legacy provider query succeeded (unexpected in test environment)")
		}
	})
}
