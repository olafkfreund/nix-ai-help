package cli

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Integration test: run the real explain-option command with a common option
func TestExplainOption_Integration(t *testing.T) {
	// Skip integration test if no AI provider API keys are available
	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("OLLAMA_HOST") == "" {
		t.Skip("Skipping integration test: no AI provider API keys set")
	}

	cmd := exec.Command("go", "run", "../../cmd/nixai/main.go", "explain-option", "services.nginx.enable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\nOutput: %s", err, string(output))
	}
	outStr := string(output)
	if !(strings.Contains(outStr, "nginx") || strings.Contains(outStr, "No relevant documentation found")) {
		t.Errorf("expected output to mention nginx or a not found message, got: %s", outStr)
	}
}
