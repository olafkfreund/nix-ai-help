package nixos

import "testing"

func TestDiagnose(t *testing.T) {
	diagnostics := Diagnose("error: something failed", "user input")
	if len(diagnostics) < 2 {
		t.Errorf("expected at least 2 diagnostics, got %d", len(diagnostics))
	}
}

func TestSuggestFix(t *testing.T) {
	d := Diagnostic{Issue: "Error detected in log output"}
	s := SuggestFix(d)
	if s == "" {
		t.Error("expected a suggestion, got empty string")
	}
}
