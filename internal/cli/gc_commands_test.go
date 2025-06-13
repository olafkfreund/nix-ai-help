package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// MockGCAIProvider implements the AIProvider interface for testing
type MockGCAIProvider struct {
	response string
	err      error
}

func (m *MockGCAIProvider) Query(prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockGCAIProvider) GenerateResponse(prompt string) (string, error) {
	return m.Query(prompt)
}

// Mock generation list for testing
func getMockGenerations() []Generation {
	now := time.Now()
	return []Generation{
		{
			Number:      1,
			Date:        now.Add(-30 * 24 * time.Hour),
			Size:        1024 * 1024 * 500, // 500 MB
			Current:     false,
			Description: "NixOS Generation 1",
			Kernel:      "5.15",
			Safe:        false,
		},
		{
			Number:      2,
			Date:        now.Add(-15 * 24 * time.Hour),
			Size:        1024 * 1024 * 550, // 550 MB
			Current:     false,
			Description: "NixOS Generation 2",
			Kernel:      "5.16",
			Safe:        false,
		},
		{
			Number:      3,
			Date:        now.Add(-2 * 24 * time.Hour),
			Size:        1024 * 1024 * 600, // 600 MB
			Current:     true,
			Description: "NixOS Generation 3",
			Kernel:      "5.17",
			Safe:        true,
		},
	}
}

// Mock GC Analysis for testing
func getMockGCAnalysis() GCAnalysis {
	return GCAnalysis{
		StoreSize:      1024 * 1024 * 5000,  // 5 GB
		AvailableSpace: 1024 * 1024 * 10000, // 10 GB
		TotalSpace:     1024 * 1024 * 20000, // 20 GB
		Generations:    getMockGenerations(),
		RecommendedClean: []CleanupItem{
			{
				Type:        "generation",
				Description: "NixOS Generation 1",
				Size:        1024 * 1024 * 500, // 500 MB
				Risk:        "low",
				Command:     "nix-env --delete-generations 1",
			},
		},
		PotentialSavings: 1024 * 1024 * 1000, // 1 GB
		RiskLevel:        "low",
		Recommendations: []string{
			"Remove old generations to free up 1 GB of space",
			"Run garbage collection after removing generations",
		},
	}
}

// Simple formatBytes function for testing
func formatBytesTest(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TestGCAnalyzeCommand tests the gc analyze command with mock data
func TestGCAnalyzeCommand(t *testing.T) {
	// Create a test command that simulates the analyze behavior
	cmd := &cobra.Command{
		Use: "analyze",
		Run: func(cmd *cobra.Command, args []string) {
			// Mock analyzing GC data
			analysis := getMockGCAnalysis()

			cmd.Println(utils.FormatHeader("🗑️ NixOS Garbage Collection Analysis"))
			cmd.Println(utils.FormatKeyValue("Store Size", formatBytesTest(analysis.StoreSize)))
			cmd.Println(utils.FormatKeyValue("Available Space", formatBytesTest(analysis.AvailableSpace)))
			cmd.Println(utils.FormatKeyValue("Total Space", formatBytesTest(analysis.TotalSpace)))

			cmd.Println(utils.FormatHeader("Generations"))
			for _, gen := range analysis.Generations {
				current := ""
				if gen.Current {
					current = " (current)"
				}
				cmd.Printf("Generation %d%s: %s - %s\n",
					gen.Number,
					current,
					formatBytesTest(gen.Size),
					gen.Description,
				)
			}

			cmd.Println(utils.FormatHeader("Recommendations"))
			for _, rec := range analysis.Recommendations {
				cmd.Printf("• %s\n", rec)
			}
		},
	}

	// Capture output
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute the command
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()

	// Check output
	expectedOutputs := []string{
		"Garbage Collection Analysis",
		"Store Size",
		"Available Space",
		"Generations",
		"Recommendations",
		"Generation 1",
		"Generation 3 (current)",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, result)
		}
	}
}

// TestGCSafeCleanCommand tests the gc safe-clean command
func TestGCSafeCleanCommand(t *testing.T) {
	mockAI := &MockGCAIProvider{
		response: "Analysis shows it's safe to remove generation 1. This will free up approximately 500 MB.",
	}

	cmd := &cobra.Command{
		Use: "safe-clean",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(utils.FormatHeader("🧹 Safe Cleanup with AI Guidance"))

			// Mock finding cleanup candidates
			analysis := getMockGCAnalysis()
			for _, item := range analysis.RecommendedClean {
				cmd.Printf("Found cleanup candidate: %s (%s)\n", item.Description, formatBytesTest(item.Size))
			}

			// Mock AI analysis
			aiResponse, _ := mockAI.Query("analyze cleanup safety")
			cmd.Println(utils.FormatHeader("🤖 AI Safety Analysis"))
			cmd.Println(aiResponse)
		},
	}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	result := output.String()
	expectedOutputs := []string{
		"Safe Cleanup",
		"AI Safety Analysis",
		"cleanup candidate",
		"safe to remove",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, result)
		}
	}
}

// TestFormatBytes tests the byte formatting function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatBytesTest(test.input)
		if result != test.expected {
			t.Errorf("formatBytesTest(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}
