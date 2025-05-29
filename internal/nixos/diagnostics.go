package nixos

import (
	"fmt"
	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
	"regexp"
	"strings"
)

// Diagnostic represents a diagnostic report for NixOS configuration issues.
type Diagnostic struct {
	Issue   string
	Details string
}

// Diagnose analyzes the provided log output and user input to identify potential NixOS configuration issues.
func Diagnose(logOutput string, userInput string, aiProvider ai.AIProvider) []Diagnostic {
	var diagnostics []Diagnostic
	log := logger.NewLogger()

	// Smarter log parsing: detect common NixOS/Nix errors
	syntaxErr := regexp.MustCompile(`(?i)syntax error|unexpected token|parse error`)
	missingPkg := regexp.MustCompile(`(?i)cannot find package|not found:|missing package`)
	failedService := regexp.MustCompile(`(?i)failed to start|service failed|unit failed`)

	if syntaxErr.MatchString(logOutput) {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:   "NixOS syntax error",
			Details: "Check your configuration.nix for syntax mistakes.",
		})
	}
	if missingPkg.MatchString(logOutput) {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:   "Missing package",
			Details: "A required package could not be found. Check your configuration and channels.",
		})
	}
	if failedService.MatchString(logOutput) {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:   "Service failed to start",
			Details: "A systemd service failed. Check the service status and logs.",
		})
	}
	if strings.Contains(logOutput, "error") && len(diagnostics) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:   "Error detected in log output",
			Details: "Please check the configuration for errors.",
		})
	}

	// User input analysis: parse config snippets
	if userInput != "" {
		if strings.Contains(userInput, "=") {
			// Looks like a config snippet, try parsing
			_, err := ParseNixConfig(userInput)
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{
					Issue:   "Invalid Nix config snippet",
					Details: err.Error(),
				})
			} else {
				diagnostics = append(diagnostics, Diagnostic{
					Issue:   "User config snippet provided",
					Details: "Parsed config: OK",
				})
			}
		} else {
			diagnostics = append(diagnostics, Diagnostic{
				Issue:   "User input provided",
				Details: fmt.Sprintf("Received user input: %s", userInput),
			})
		}
	}

	// AI integration: ask LLM for additional suggestions if log or input is non-trivial
	if (len(logOutput) > 0 || len(userInput) > 0) && aiProvider != nil {
		prompt := "You are a NixOS expert. Analyze the following log and user input. Provide a diagnosis and actionable suggestions.\n\nLog:\n" + logOutput + "\n\nUser input:\n" + userInput
		aiResp, err := aiProvider.Query(prompt)
		if err == nil && aiResp != "" {
			diagnostics = append(diagnostics, Diagnostic{
				Issue:   "AI Suggestion",
				Details: aiResp,
			})
		} else if err != nil {
			log.Warn("AI provider error: " + err.Error())
		}
	}

	return diagnostics
}

// SuggestFix provides actionable and contextual suggestions for fixing identified issues.
func SuggestFix(diagnostic Diagnostic) string {
	switch diagnostic.Issue {
	case "NixOS syntax error":
		return "Check your configuration.nix for syntax errors. Run `nixos-rebuild switch` and review the error output. See: https://nixos.org/manual/nixos/stable/#sec-configuration"
	case "Missing package":
		return "Ensure the package name is correct and your channels are up to date. See: https://nixos.org/manual/nixpkgs/stable/"
	case "Service failed to start":
		return "Check the service status with `systemctl status <service>`. See logs with `journalctl -u <service>`."
	case "Invalid Nix config snippet":
		return "Check your Nix config syntax. See: https://nix.dev/manual/nix/2.28/language/"
	case "User config snippet provided":
		return "Your config snippet appears valid. If you have issues, check for typos or missing options."
	case "Error detected in log output":
		return "Check the NixOS configuration files for syntax errors or misconfigurations."
	case "User input provided":
		return "Consider validating the input against expected configuration parameters."
	case "AI Suggestion":
		return "Review the AI suggestion above for actionable steps."
	default:
		return "No suggestions available. Try running `nixos-rebuild switch` or consult the NixOS manual."
	}
}

// FormatDiagnostics returns a formatted string for displaying diagnostics.
func FormatDiagnostics(diags []Diagnostic) string {
	if len(diags) == 0 {
		return utils.FormatInfo("No diagnostics found.")
	}

	var sb strings.Builder
	
	// Add a header for the diagnostics
	sb.WriteString(utils.FormatHeader("🔍 NixOS Diagnostics Report"))
	sb.WriteString("\n\n")
	
	for i, d := range diags {
		// Create a formatted diagnostic entry
		title := fmt.Sprintf("Issue %d: %s", i+1, d.Issue)
		
		// Format the diagnostic as a box with details and suggestions
		content := fmt.Sprintf("%s\n\n%s\n%s", 
			utils.FormatKeyValue("Details", d.Details),
			utils.FormatSubsection("Suggested Fix", SuggestFix(d)),
			"")
		
		sb.WriteString(utils.FormatBox(title, content))
		sb.WriteString("\n")
	}
	
	// Add a helpful footer
	sb.WriteString(utils.FormatDivider())
	sb.WriteString("\n")
	sb.WriteString(utils.FormatTip("Run 'nixai interactive' for more detailed troubleshooting"))
	
	return sb.String()
}
