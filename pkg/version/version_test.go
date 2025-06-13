package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	// Test that we get valid info
	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}

	if info.BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// Test platform format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Expected platform %s, got %s", expectedPlatform, info.Platform)
	}

	// Test Go version format
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("Expected GoVersion to start with 'go', got %s", info.GoVersion)
	}
}

func TestString(t *testing.T) {
	info := Get()
	str := info.String()

	// Test that string contains expected components
	if !strings.Contains(str, "version") {
		t.Error("String should contain 'version'")
	}

	if !strings.Contains(str, info.Version) {
		t.Error("String should contain version")
	}

	if !strings.Contains(str, info.GitCommit) {
		t.Error("String should contain git commit")
	}

	if !strings.Contains(str, info.BuildDate) {
		t.Error("String should contain build date")
	}

	// Test specific format
	expected := fmt.Sprintf("version %s (commit: %s, built: %s)",
		info.Version, info.GitCommit, info.BuildDate)
	if str != expected {
		t.Errorf("Expected %q, got %q", expected, str)
	}
}

func TestShort(t *testing.T) {
	info := Get()
	short := info.Short()

	// Test that short returns just the version
	if short != info.Version {
		t.Errorf("Expected short to return %s, got %s", info.Version, short)
	}
}

func TestDefaultValues(t *testing.T) {
	// Test default values when not set via ldflags
	if Version == "" {
		t.Error("Version should have default value")
	}

	if GitCommit == "" {
		t.Error("GitCommit should have default value")
	}

	if BuildDate == "" {
		t.Error("BuildDate should have default value")
	}

	// Test that GoVersion is set from runtime
	if GoVersion != runtime.Version() {
		t.Errorf("Expected GoVersion to be %s, got %s", runtime.Version(), GoVersion)
	}
}

func TestJSON(t *testing.T) {
	info := Get()

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal Info to JSON: %v", err)
	}

	// Test that JSON contains expected fields
	jsonStr := string(jsonBytes)
	expectedFields := []string{
		`"version":`,
		`"git_commit":`,
		`"build_date":`,
		`"go_version":`,
		`"platform":`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output should contain %q, got: %s", field, jsonStr)
		}
	}

	// Test JSON unmarshaling
	var unmarshaled Info
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to Info: %v", err)
	}

	// Verify unmarshaled data matches original
	if unmarshaled.Version != info.Version {
		t.Errorf("Version mismatch: expected %s, got %s", info.Version, unmarshaled.Version)
	}
	if unmarshaled.GitCommit != info.GitCommit {
		t.Errorf("GitCommit mismatch: expected %s, got %s", info.GitCommit, unmarshaled.GitCommit)
	}
}

func TestVersionStringFormat(t *testing.T) {
	// Test that String() doesn't contain duplicate "version" prefixes
	info := Get()
	versionStr := info.String()

	// Should start with "version " but not contain "version version"
	if !strings.HasPrefix(versionStr, "version ") {
		t.Errorf("Version string should start with 'version ', got: %s", versionStr)
	}

	// Should not contain duplicate "version" text
	if strings.Contains(versionStr, "version version") {
		t.Errorf("Version string should not contain duplicate 'version' text, got: %s", versionStr)
	}

	// Should contain expected format elements
	if !strings.Contains(versionStr, "commit:") {
		t.Errorf("Version string should contain 'commit:', got: %s", versionStr)
	}
	if !strings.Contains(versionStr, "built:") {
		t.Errorf("Version string should contain 'built:', got: %s", versionStr)
	}
}

// Benchmark tests
func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get()
	}
}

func BenchmarkInfo_String(b *testing.B) {
	info := Get()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = info.String()
	}
}

func BenchmarkInfo_Short(b *testing.B) {
	info := Get()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = info.Short()
	}
}
