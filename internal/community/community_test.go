package community

import (
	"nix-ai-help/internal/config"
	"os"
	"testing"
)

func TestManager_SearchConfigurations(t *testing.T) {
	mgr, _ := setupTestManager(t)
	results, err := mgr.SearchConfigurations("gaming", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result for 'gaming'")
	}
}

func TestManager_SearchByCategory(t *testing.T) {
	mgr, _ := setupTestManager(t)
	results, err := mgr.SearchByCategory("guides", "nixos", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result for category 'guides'")
	}
}

func TestManager_ShareConfiguration(t *testing.T) {
	mgr, _ := setupTestManager(t)
	file, err := os.CreateTemp("", "nixos-config-*.nix")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	_, _ = file.WriteString("# NixOS config example\n")
	_ = file.Close()
	conf := &Configuration{
		Name:     "Test Config",
		Author:   "tester",
		FilePath: file.Name(),
	}
	err = mgr.ShareConfiguration(conf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.ID == "" {
		t.Error("expected configuration to have an ID after sharing")
	}
}

func TestManager_ValidateConfiguration(t *testing.T) {
	mgr, _ := setupTestManager(t)
	file, err := os.CreateTemp("", "nixos-config-*.nix")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	_, _ = file.WriteString("system.stateVersion = \"23.11\";\n")
	_ = file.Close()
	result, err := mgr.ValidateConfiguration(file.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid {
		t.Error("expected configuration to be valid")
	}
}

func TestManager_GetTrends(t *testing.T) {
	mgr, _ := setupTestManager(t)
	trends, err := mgr.GetTrends()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trends.TotalConfigurations == 0 {
		t.Error("expected some trending configurations")
	}
}

// Minimal test setup utility
func setupTestManager(t *testing.T) (*Manager, func()) {
	tempDir := t.TempDir()
	cfg := &config.UserConfig{}
	manager := NewManager(cfg)
	manager.cache = NewCacheManager(tempDir)
	cleanup := func() {}
	return manager, cleanup
}
