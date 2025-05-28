package nixos

import (
	"fmt"
	"strings"
)

// Diagnostic represents a diagnostic report for NixOS configuration issues.
type Diagnostic struct {
	Issue   string
	Details string
}

// Diagnose analyzes the provided log output and user input to identify potential NixOS configuration issues.
func Diagnose(logOutput string, userInput string) []Diagnostic {
	var diagnostics []Diagnostic

	// Example log analysis (this should be replaced with actual parsing logic)
	if strings.Contains(logOutput, "error") {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:   "Error detected in log output",
			Details: "Please check the configuration for errors.",
		})
	}

	// Example user input analysis
	if userInput != "" {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:   "User input provided",
			Details: fmt.Sprintf("Received user input: %s", userInput),
		})
	}

	// Additional diagnostic checks can be added here

	return diagnostics
}

// SuggestFix provides suggestions for fixing identified issues based on the diagnostics.
func SuggestFix(diagnostic Diagnostic) string {
	switch diagnostic.Issue {
	case "Error detected in log output":
		return "Check the NixOS configuration files for syntax errors or misconfigurations."
	case "User input provided":
		return "Consider validating the input against expected configuration parameters."
	default:
		return "No suggestions available."
	}
}
