// Package cli provides the command-line interface for nixai
package cli

import (
	"os"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
)

// InitializeAIProvider creates the appropriate AI provider based on configuration
func InitializeAIProvider(cfg *config.UserConfig) ai.AIProvider {
	switch cfg.AIProvider {
	case "ollama":
		return ai.NewOllamaProvider(cfg.AIModel)
	case "gemini":
		return ai.NewGeminiClient(
			os.Getenv("GEMINI_API_KEY"),
			"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent",
		)
	case "openai":
		return ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	case "llamacpp":
		return ai.NewLlamaCppProvider(cfg.AIModel)
	case "custom":
		if cfg.CustomAI.BaseURL != "" {
			return ai.NewCustomProvider(cfg.CustomAI.BaseURL, cfg.CustomAI.Headers)
		}
		// fallback to Ollama if not configured
		return ai.NewOllamaProvider("llama3")
	default:
		return ai.NewOllamaProvider("llama3")
	}
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
