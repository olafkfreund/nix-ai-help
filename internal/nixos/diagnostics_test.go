package nixos

import "testing"

func TestDiagnose(t *testing.T) {
	diagnostics := Diagnose("error: something failed", "user input", nil)
	if len(diagnostics) < 2 {
		t.Errorf("expected at least 2 diagnostics, got %d", len(diagnostics))
	}
}

func TestDiagnose_SyntaxError(t *testing.T) {
	d := Diagnose("syntax error: unexpected token", "", nil)
	found := false
	for _, diag := range d {
		if diag.Issue == "NixOS syntax error" {
			found = true
		}
	}
	if !found {
		t.Error("expected NixOS syntax error diagnostic")
	}
}

func TestDiagnose_MissingPackage(t *testing.T) {
	d := Diagnose("cannot find package: foo", "", nil)
	found := false
	for _, diag := range d {
		if diag.Issue == "Missing package" {
			found = true
		}
	}
	if !found {
		t.Error("expected Missing package diagnostic")
	}
}

func TestDiagnose_FailedService(t *testing.T) {
	d := Diagnose("failed to start nginx.service", "", nil)
	found := false
	for _, diag := range d {
		if diag.Issue == "Service failed to start" {
			found = true
		}
	}
	if !found {
		t.Error("expected Service failed to start diagnostic")
	}
}

func TestSuggestFix(t *testing.T) {
	d := Diagnostic{Issue: "Error detected in log output"}
	s := SuggestFix(d)
	if s == "" {
		t.Error("expected a suggestion, got empty string")
	}
}
