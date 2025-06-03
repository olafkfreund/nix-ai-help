package nixos

import (
	"strings"
	"testing"
)

func TestParseLog(t *testing.T) {
	log := `INFO something happened
ERROR something failed
  details: failure in module
Jun  3 12:34:56 host systemd[1]: Started Nginx Service.
[2025-06-03T12:35:00Z] ERROR nginx: Failed to bind to port 80`
	entries, err := ParseLog(log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, entry := range entries {
		t.Logf("Entry %d: %+v", i, entry)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 log entries, got %d", len(entries))
	}
	if entries[0].Level != "INFO" || entries[0].Message != "something happened" {
		t.Errorf("expected INFO entry with message, got %+v", entries[0])
	}
	if entries[1].Message != "something failed\ndetails: failure in module" {
		t.Errorf("expected multi-line error message, got %+v", entries[1])
	}
	if entries[2].Unit != "systemd" || entries[2].Message != "Started Nginx Service." {
		t.Errorf("expected systemd entry, got %+v", entries[2])
	}
	if entries[3].Level != "ERROR" || entries[3].Unit != "nginx" {
		t.Errorf("expected generic format entry, got %+v", entries[3])
	}
}

func TestParseLogStream(t *testing.T) {
	logLines := []string{
		"Jun  3 12:34:56 host systemd[1]: Started Nginx Service.",
		"[2025-06-03T12:35:00Z] ERROR nginx: Failed to bind to port 80",
		"INFO something happened",
		"ERROR something failed",
		"  details: failure in module",
	}
	input := make(chan string, len(logLines))
	for _, line := range logLines {
		input <- line
	}
	close(input)
	entries := []LogEntry{}
	for entry := range ParseLogStream(input) {
		entries = append(entries, entry)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 log entries, got %d", len(entries))
	}
	if entries[0].Unit != "systemd" || !strings.Contains(entries[0].Message, "Started Nginx Service") {
		t.Errorf("expected systemd entry, got %+v", entries[0])
	}
	if entries[1].Level != "ERROR" || entries[1].Unit != "nginx" {
		t.Errorf("expected generic error entry, got %+v", entries[1])
	}
	if entries[2].Level != "INFO" {
		t.Errorf("expected INFO entry, got %+v", entries[2])
	}
	if !strings.Contains(entries[3].Message, "something failed") || !strings.Contains(entries[3].Message, "details") {
		t.Errorf("expected multi-line error message, got %+v", entries[3])
	}
}

func TestUserDefinedErrorPatternDiagnostic(t *testing.T) {
	// Simulate a log that matches a user-defined error pattern (see default.yaml)
	log := `custom permission denied: user policy block`
	entries, err := ParseLog(log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	// Now run diagnostics (requires config with the custom pattern loaded)
	diags := Diagnose(log, "", nil)
	found := false
	for _, d := range diags {
		if d.ErrorType == "permission" && d.Issue != "" && d.Severity == "medium" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("user-defined error pattern was not recognized in diagnostics: %+v", diags)
	}
}
