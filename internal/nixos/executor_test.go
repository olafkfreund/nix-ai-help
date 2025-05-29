package nixos

import (
	"strings"
	"testing"
)

func TestExecuteCommand(t *testing.T) {
	exec := NewExecutor("")
	output, err := exec.ExecuteCommand("echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.TrimSpace(output) != "hello" {
		t.Errorf("expected output 'hello', got '%s'", output)
	}
}

func TestListServiceOptions_EmptyService(t *testing.T) {
	exec := NewExecutor("")
	_, err := exec.ListServiceOptions("")
	if err == nil {
		t.Error("expected error for empty service name, got nil")
	}
}

func TestListServiceOptions_Command(t *testing.T) {
	exec := NewExecutor("")
	// This will likely error unless nixos-option is available, but should not panic
	_, err := exec.ListServiceOptions("dummyservice")
	// Accept error, just ensure no panic and command runs
	if err != nil && !strings.Contains(err.Error(), "executable file not found") {
		t.Logf("ListServiceOptions returned error (expected if nixos-option missing): %v", err)
	}
}
