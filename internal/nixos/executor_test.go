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
