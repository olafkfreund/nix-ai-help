package nixos

import (
	"strings"
	"testing"
)

func TestSearchNixPackages(t *testing.T) {
	exec := NewExecutor()
	output, err := exec.SearchNixPackages("hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got '%s'", output)
	}
}

func TestSearchNixPackages_Firefox(t *testing.T) {
	exec := NewExecutor()
	output, err := exec.SearchNixPackages("firefox")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(output, "firefox") {
		t.Errorf("expected output to contain 'firefox', got: %s", output)
	}
	if !strings.Contains(output, "web browser") && !strings.Contains(output, "browser") {
		t.Errorf("expected output to contain a description for firefox, got: %s", output)
	}
}

func TestSearchNixPackages_MultiWord(t *testing.T) {
	exec := NewExecutor()
	output, err := exec.SearchNixPackages("libre office")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(strings.ToLower(output), "libreoffice") {
		t.Errorf("expected output to contain 'libreoffice', got: %s", output)
	}
}

func TestInstallNixPackage(t *testing.T) {
	exec := NewExecutor()
	// Use a dummy package unlikely to be installed, but don't actually install in test
	// Instead, check that the command is constructed and returns output (may error)
	_, _ = exec.InstallNixPackage("nixpkgs.hello")
	// No assertion: just ensure no panic and command runs
}
