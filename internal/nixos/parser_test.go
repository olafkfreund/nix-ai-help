package nixos

import "testing"

func TestParseLog(t *testing.T) {
	log := "INFO something happened\nERROR something failed"
	parsed, err := ParseLog(log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed["INFO"] != "something happened" {
		t.Errorf("expected 'something happened', got '%v'", parsed["INFO"])
	}
	if parsed["ERROR"] != "something failed" {
		t.Errorf("expected 'something failed', got '%v'", parsed["ERROR"])
	}
}
