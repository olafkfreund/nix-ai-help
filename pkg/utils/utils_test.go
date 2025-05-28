package utils

import (
	"os"
	"testing"
)

func TestIsFileAndIsDirectory(t *testing.T) {
	f, err := os.CreateTemp("", "testfile-*.tmp")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	if !IsFile(f.Name()) {
		t.Errorf("expected %s to be a file", f.Name())
	}
}

func TestSplitLines(t *testing.T) {
	input := "a\nb\nc"
	lines := SplitLines(input)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !Contains(slice, "b") {
		t.Errorf("expected slice to contain 'b'")
	}
	if Contains(slice, "d") {
		t.Errorf("expected slice to not contain 'd'")
	}
}
