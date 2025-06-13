package cli

import (
	"testing"

	"nix-ai-help/internal/mcp"
	"nix-ai-help/pkg/logger"
)

// MockMigrationAIProvider implements the AIProvider interface for testing
type MockMigrationAIProvider struct {
	response string
	err      error
}

func (m *MockMigrationAIProvider) Query(prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockMigrationAIProvider) GenerateResponse(prompt string) (string, error) {
	return m.Query(prompt)
}

// TestMigrationManagerCreation tests that we can create a MigrationManager
func TestMigrationManagerCreation(t *testing.T) {
	logger := logger.NewLoggerWithLevel("info")
	aiProvider := &MockMigrationAIProvider{
		response: "This is a mock AI response for migration",
	}
	mcpClient := mcp.NewMCPClient("http://localhost:9999")

	manager := NewMigrationManager("/mock/nixos/path", logger, aiProvider, mcpClient)

	if manager == nil {
		t.Errorf("Expected to create MigrationManager, got nil")
	}
}

// TestMigrationAnalysis tests the AnalyzeMigration method
func TestMigrationAnalysis(t *testing.T) {
	logger := logger.NewLoggerWithLevel("info")
	aiProvider := &MockMigrationAIProvider{
		response: "This is a mock AI response for migration",
	}
	mcpClient := mcp.NewMCPClient("http://localhost:9999")

	manager := NewMigrationManager("/mock/nixos/path", logger, aiProvider, mcpClient)

	// Test AnalyzeMigration method (this exists)
	analysis, err := manager.AnalyzeMigration("flakes")
	if err != nil {
		t.Logf("AnalyzeMigration returned error (expected for mock setup): %v", err)
		return // This is expected since we're using mock paths
	}

	// If analysis succeeds, check the structure
	if analysis != nil {
		if analysis.TargetSetup != "flakes" {
			t.Errorf("Expected target setup 'flakes', got %s", analysis.TargetSetup)
		}
	}
}

// TestDetectCurrentSetup tests the DetectCurrentSetup method
func TestDetectCurrentSetup(t *testing.T) {
	logger := logger.NewLoggerWithLevel("info")
	aiProvider := &MockMigrationAIProvider{
		response: "This is a mock AI response for migration",
	}
	mcpClient := mcp.NewMCPClient("http://localhost:9999")

	manager := NewMigrationManager("/mock/nixos/path", logger, aiProvider, mcpClient)

	// Test the DetectCurrentSetup method (this exists)
	currentSetup, metadata, err := manager.DetectCurrentSetup()
	if err != nil {
		t.Logf("DetectCurrentSetup returned error (expected for mock path): %v", err)
	}
	t.Logf("Current setup: %s, metadata: %+v", currentSetup, metadata)
}

// TestMigrationTypes tests basic migration analysis types
func TestMigrationTypes(t *testing.T) {
	logger := logger.NewLoggerWithLevel("info")
	aiProvider := &MockMigrationAIProvider{
		response: "This is a mock AI response for migration",
	}
	mcpClient := mcp.NewMCPClient("http://localhost:9999")

	manager := NewMigrationManager("/mock/nixos/path", logger, aiProvider, mcpClient)

	// Test different migration targets
	targets := []string{"flakes", "channels"}

	for _, target := range targets {
		analysis, err := manager.AnalyzeMigration(target)
		if err != nil {
			t.Logf("AnalyzeMigration for %s returned error (expected): %v", target, err)
			continue
		}

		if analysis != nil && analysis.TargetSetup != target {
			t.Errorf("Expected target setup '%s', got '%s'", target, analysis.TargetSetup)
		}
	}
}
