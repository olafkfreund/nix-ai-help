package nixos

import (
	"fmt"
	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
	"regexp"
	"strings"
)

// Diagnostic represents a diagnostic report for NixOS configuration issues.
type Diagnostic struct {
	Issue     string
	Details   string
	ErrorType string   // Category of error (syntax, package, service, etc.)
	Severity  string   // low, medium, high, critical
	Steps     []string // Step-by-step fix instructions
	DocsLinks []string // Related documentation links
}

// ErrorPattern represents a pattern for matching specific types of errors
type ErrorPattern struct {
	Pattern     *regexp.Regexp
	ErrorType   string
	Severity    string
	Description string
}

// UserErrorPattern represents a user-defined error pattern from YAML config.
type UserErrorPattern struct {
	Name        string `yaml:"name"`
	Pattern     string `yaml:"pattern"`
	ErrorType   string `yaml:"error_type"`
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
}

// getUserErrorPatterns loads user-defined error patterns from config.
func getUserErrorPatterns() ([]UserErrorPattern, error) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		return nil, err
	}
	var patterns []UserErrorPattern
	for _, ep := range cfg.Diagnostics.ErrorPatterns {
		patterns = append(patterns, UserErrorPattern{
			Name:        ep.Name,
			Pattern:     ep.Pattern,
			ErrorType:   ep.ErrorType,
			Severity:    ep.Severity,
			Description: ep.Description,
		})
	}
	return patterns, nil
}

// Diagnose analyzes the provided log output and user input to identify potential NixOS configuration issues.
func Diagnose(logOutput string, userInput string, aiProvider ai.AIProvider) []Diagnostic {
	var diagnostics []Diagnostic
	log := logger.NewLogger()

	// Enhanced error pattern recognition
	errorPatterns := map[string]ErrorPattern{
		"syntax_error": {
			Pattern:     regexp.MustCompile(`(?i)(syntax error|unexpected token|parse error|expecting|unexpected '\)')`),
			ErrorType:   "syntax",
			Severity:    "high",
			Description: "NixOS configuration syntax error",
		},
		"missing_package": {
			Pattern:     regexp.MustCompile(`(?i)(cannot find package|not found:|missing package|attribute '.*' missing|The option .* does not exist)`),
			ErrorType:   "package",
			Severity:    "medium",
			Description: "Missing package or attribute",
		},
		"failed_service": {
			Pattern:     regexp.MustCompile(`(?i)(failed to start|service failed|unit failed|systemctl.*failed|ExecStart.*failed)`),
			ErrorType:   "service",
			Severity:    "high",
			Description: "Service failed to start",
		},
		"permission_error": {
			Pattern:     regexp.MustCompile(`(?i)(permission denied|access denied|insufficient permissions|Operation not permitted)`),
			ErrorType:   "permission",
			Severity:    "medium",
			Description: "Permission or access error",
		},
		"build_failure": {
			Pattern:     regexp.MustCompile(`(?i)(build failed|compilation failed|make.*error|gcc.*error|cargo.*error|npm.*error)`),
			ErrorType:   "build",
			Severity:    "high",
			Description: "Package build failure",
		},
		"dependency_error": {
			Pattern:     regexp.MustCompile(`(?i)(dependency.*failed|circular dependency|conflict.*dependency|unsatisfied dependency)`),
			ErrorType:   "dependency",
			Severity:    "high",
			Description: "Dependency resolution error",
		},
		"channel_error": {
			Pattern:     regexp.MustCompile(`(?i)(channel.*error|nix-channel.*failed|unstable.*|stable.*channel)`),
			ErrorType:   "channel",
			Severity:    "medium",
			Description: "Nix channel issue",
		},
		"disk_space": {
			Pattern:     regexp.MustCompile(`(?i)(no space left|disk full|insufficient space|storage.*full)`),
			ErrorType:   "system",
			Severity:    "critical",
			Description: "Insufficient disk space",
		},
		"network_error": {
			Pattern:     regexp.MustCompile(`(?i)(network.*error|connection.*failed|timeout|DNS.*failed|unable to fetch)`),
			ErrorType:   "network",
			Severity:    "medium",
			Description: "Network connectivity issue",
		},
		"derivation_error": {
			Pattern:     regexp.MustCompile(`(?i)(derivation.*failed|hash mismatch|fixed-output derivation|builder.*failed)`),
			ErrorType:   "derivation",
			Severity:    "high",
			Description: "Nix derivation build error",
		},
	}

	// Merge user-defined patterns from config
	userPatterns, err := getUserErrorPatterns()
	if err == nil {
		for _, up := range userPatterns {
			if up.Pattern == "" || up.Name == "" {
				continue
			}
			compiled, err := regexp.Compile(up.Pattern)
			if err != nil {
				log.Warn("Invalid user error pattern: " + up.Name + ", skipping: " + err.Error())
				continue
			}
			errorPatterns[up.Name] = ErrorPattern{
				Pattern:     compiled,
				ErrorType:   up.ErrorType,
				Severity:    up.Severity,
				Description: up.Description,
			}
		}
	}

	// Check for each error pattern
	for _, pattern := range errorPatterns {
		if pattern.Pattern.MatchString(logOutput) {
			diagnostic := Diagnostic{
				Issue:     pattern.Description,
				Details:   extractErrorContext(logOutput, pattern.Pattern),
				ErrorType: pattern.ErrorType,
				Severity:  pattern.Severity,
				Steps:     generateFixSteps(pattern.ErrorType),
				DocsLinks: getRelevantDocsLinks(pattern.ErrorType),
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	// If no specific patterns matched but there's an error, create a generic diagnostic
	if strings.Contains(logOutput, "error") && len(diagnostics) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Issue:     "Unclassified error detected",
			Details:   extractGenericError(logOutput),
			ErrorType: "generic",
			Severity:  "medium",
			Steps:     []string{"Review the error message carefully", "Check your configuration syntax", "Consult the NixOS manual"},
			DocsLinks: []string{"https://nixos.org/manual/nixos/stable/"},
		})
	}

	// User input analysis: parse config snippets
	if userInput != "" {
		if strings.Contains(userInput, "=") {
			// Looks like a config snippet, try parsing
			_, err := ParseNixConfig(userInput)
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{
					Issue:     "Invalid Nix config snippet",
					Details:   err.Error(),
					ErrorType: "syntax",
					Severity:  "high",
					Steps:     generateFixSteps("syntax"),
					DocsLinks: getRelevantDocsLinks("syntax"),
				})
			} else {
				diagnostics = append(diagnostics, Diagnostic{
					Issue:     "User config snippet provided",
					Details:   "Parsed config: OK",
					ErrorType: "info",
					Severity:  "low",
					Steps:     []string{"Configuration appears valid"},
					DocsLinks: []string{},
				})
			}
		} else {
			diagnostics = append(diagnostics, Diagnostic{
				Issue:     "User input provided",
				Details:   fmt.Sprintf("Received user input: %s", userInput),
				ErrorType: "info",
				Severity:  "low",
				Steps:     []string{"Review the provided input for any issues"},
				DocsLinks: []string{},
			})
		}
	}

	// Enhanced AI integration: ask LLM for structured diagnosis and fixes
	if (len(logOutput) > 0 || len(userInput) > 0) && aiProvider != nil {
		prompt := buildEnhancedDiagnosticPrompt(logOutput, userInput, diagnostics)
		aiResp, err := aiProvider.Query(prompt)
		if err == nil && aiResp != "" {
			diagnostics = append(diagnostics, Diagnostic{
				Issue:     "AI-Enhanced Analysis",
				Details:   aiResp,
				ErrorType: "ai_analysis",
				Severity:  "medium",
				Steps:     extractAISteps(aiResp),
				DocsLinks: extractAILinks(aiResp),
			})
		} else if err != nil {
			log.Warn("AI provider error: " + err.Error())
		}
	}

	return diagnostics
}

// Helper functions for enhanced error processing

// extractErrorContext extracts relevant context around an error match
func extractErrorContext(logOutput string, pattern *regexp.Regexp) string {
	lines := strings.Split(logOutput, "\n")
	var context []string

	for i, line := range lines {
		if pattern.MatchString(line) {
			// Include 2 lines before and after the match for context
			start := max(0, i-2)
			end := min(len(lines), i+3)

			for j := start; j < end; j++ {
				if j == i {
					context = append(context, ">>> "+strings.TrimSpace(lines[j]))
				} else {
					context = append(context, "    "+strings.TrimSpace(lines[j]))
				}
			}
			break
		}
	}

	if len(context) == 0 {
		// If no specific context found, return the matched portion
		matches := pattern.FindStringSubmatch(logOutput)
		if len(matches) > 0 {
			return "Error found: " + matches[0]
		}
	}

	return strings.Join(context, "\n")
}

// extractGenericError extracts error information for unclassified errors
func extractGenericError(logOutput string) string {
	lines := strings.Split(logOutput, "\n")
	var errorLines []string

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "error") {
			errorLines = append(errorLines, strings.TrimSpace(line))
			if len(errorLines) >= 3 { // Limit to first 3 error lines
				break
			}
		}
	}

	if len(errorLines) > 0 {
		return strings.Join(errorLines, "\n")
	}

	return "Error detected but details unclear from log output"
}

// generateFixSteps provides step-by-step fix instructions based on error type
func generateFixSteps(errorType string) []string {
	switch errorType {
	case "syntax":
		return []string{
			"Check your configuration.nix for syntax errors",
			"Look for missing semicolons, brackets, or quotes",
			"Run 'nixos-rebuild dry-run' to validate syntax",
			"Use a Nix-aware editor with syntax highlighting",
		}
	case "package":
		return []string{
			"Verify the package name is correct",
			"Check if the package exists in your channel",
			"Update your channels with 'nix-channel --update'",
			"Search for the package with 'nix-env -qaP | grep <package>'",
		}
	case "service":
		return []string{
			"Check service status with 'systemctl status <service>'",
			"View service logs with 'journalctl -u <service>'",
			"Verify service configuration in configuration.nix",
			"Check if required dependencies are installed",
		}
	case "permission":
		return []string{
			"Check file/directory permissions",
			"Ensure the user has necessary privileges",
			"Consider if the operation requires root access",
			"Review systemd service user/group settings",
		}
	case "build":
		return []string{
			"Check for missing build dependencies",
			"Verify the package source is accessible",
			"Consider using a different version or channel",
			"Check for available binary cache",
		}
	case "dependency":
		return []string{
			"Review package dependencies for conflicts",
			"Try updating all packages to resolve conflicts",
			"Consider using package overrides",
			"Check for circular dependencies in configuration",
		}
	case "channel":
		return []string{
			"Update channels with 'nix-channel --update'",
			"List current channels with 'nix-channel --list'",
			"Consider switching to a different channel",
			"Verify channel URLs are correct",
		}
	case "system":
		return []string{
			"Free up disk space",
			"Clean Nix store with 'nix-collect-garbage -d'",
			"Check available space with 'df -h'",
			"Consider moving to a larger disk",
		}
	case "network":
		return []string{
			"Check internet connectivity",
			"Verify DNS resolution",
			"Check proxy settings if applicable",
			"Try using a different mirror or channel",
		}
	case "derivation":
		return []string{
			"Check if the derivation builds locally",
			"Verify source integrity and availability",
			"Consider using '--option substitute false'",
			"Check for updated package versions",
		}
	default:
		return []string{
			"Review the error message carefully",
			"Check the NixOS manual for guidance",
			"Search for similar issues in forums",
			"Consider asking for help in NixOS community",
		}
	}
}

// getRelevantDocsLinks returns documentation links relevant to the error type
func getRelevantDocsLinks(errorType string) []string {
	switch errorType {
	case "syntax":
		return []string{
			"https://nixos.org/manual/nixos/stable/#sec-configuration",
			"https://nix.dev/manual/nix/2.28/language/",
			"https://nix.dev/",
		}
	case "package":
		return []string{
			"https://nixos.org/manual/nixpkgs/stable/",
			"https://search.nixos.org/packages",
		}
	case "service":
		return []string{
			"https://nixos.org/manual/nixos/stable/#sec-systemd",
			"https://wiki.nixos.org/wiki/Systemd",
		}
	case "channel":
		return []string{
			"https://nixos.org/manual/nix/stable/command-ref/nix-channel.html",
			"https://wiki.nixos.org/wiki/Nix_channels",
		}
	case "build":
		return []string{
			"https://nixos.org/manual/nixpkgs/stable/#chap-quick-start",
			"https://wiki.nixos.org/wiki/Packaging",
			"https://nix.dev",
		}
	default:
		return []string{
			"https://nixos.org/manual/nixos/stable/",
			"https://wiki.nixos.org/wiki/NixOS_Wiki",
		}
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildEnhancedDiagnosticPrompt creates a structured prompt for AI analysis
func buildEnhancedDiagnosticPrompt(logOutput, userInput string, existingDiagnostics []Diagnostic) string {
	var prompt strings.Builder

	prompt.WriteString("You are a NixOS expert. Analyze the following error information and provide a structured diagnosis.\n\n")

	if len(existingDiagnostics) > 0 {
		prompt.WriteString("Initial diagnosis found these issues:\n")
		for _, diag := range existingDiagnostics {
			if diag.ErrorType != "info" && diag.ErrorType != "ai_analysis" {
				prompt.WriteString(fmt.Sprintf("- %s (%s): %s\n", diag.Issue, diag.Severity, diag.Details))
			}
		}
		prompt.WriteString("\n")
	}

	if logOutput != "" {
		prompt.WriteString("Error log:\n```\n")
		prompt.WriteString(logOutput)
		prompt.WriteString("\n```\n\n")
	}

	if userInput != "" {
		prompt.WriteString("User configuration or input:\n```\n")
		prompt.WriteString(userInput)
		prompt.WriteString("\n```\n\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. Root cause analysis of the error\n")
	prompt.WriteString("2. Step-by-step fix instructions\n")
	prompt.WriteString("3. Prevention tips for future\n")
	prompt.WriteString("4. Any relevant NixOS documentation links\n\n")
	prompt.WriteString("Format your response clearly with numbered steps and be specific about NixOS commands and configurations.")

	return prompt.String()
}

// extractAISteps extracts step-by-step instructions from AI response
func extractAISteps(aiResponse string) []string {
	var steps []string
	lines := strings.Split(aiResponse, "\n")

	// Compile regex once for better performance
	numberedStepPattern := regexp.MustCompile(`^\d+\.`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for numbered steps or bullet points
		if numberedStepPattern.MatchString(line) {
			steps = append(steps, line)
		} else if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			steps = append(steps, line)
		}
	}

	// If no structured steps found, return a generic message
	if len(steps) == 0 {
		steps = append(steps, "Follow the AI analysis provided above")
	}

	return steps
}

// extractAILinks extracts documentation links from AI response
func extractAILinks(aiResponse string) []string {
	var links []string

	// Look for http/https URLs
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlPattern.FindAllString(aiResponse, -1)

	for _, match := range matches {
		// Clean up any trailing punctuation
		link := strings.TrimRight(match, ".,;:!?)")
		links = append(links, link)
	}

	return links
}

// SuggestFix provides actionable and contextual suggestions for fixing identified issues.
func SuggestFix(diagnostic Diagnostic) string {
	// Use the structured steps if available
	if len(diagnostic.Steps) > 0 {
		var suggestions strings.Builder
		suggestions.WriteString("Suggested steps:\n")
		for i, step := range diagnostic.Steps {
			suggestions.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}

		if len(diagnostic.DocsLinks) > 0 {
			suggestions.WriteString("\nRelevant documentation:\n")
			for _, link := range diagnostic.DocsLinks {
				suggestions.WriteString(fmt.Sprintf("- %s\n", link))
			}
		}

		return suggestions.String()
	}

	// Fallback to legacy suggestions based on issue type
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
	case "AI Suggestion", "AI-Enhanced Analysis":
		return "Review the AI analysis above for actionable steps."
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
	sb.WriteString(utils.FormatHeader("🔍 NixOS Error Analysis Report"))
	sb.WriteString("\n\n")

	// Group diagnostics by severity
	critical := filterBySeverity(diags, "critical")
	high := filterBySeverity(diags, "high")
	medium := filterBySeverity(diags, "medium")
	low := filterBySeverity(diags, "low")

	// Display in order of severity
	severityGroups := []struct {
		name  string
		items []Diagnostic
		icon  string
	}{
		{"🚨 Critical Issues", critical, "🚨"},
		{"⚠️  High Priority", high, "⚠️"},
		{"📋 Medium Priority", medium, "📋"},
		{"ℹ️  Information", low, "ℹ️"},
	}

	issueCount := 1
	for _, group := range severityGroups {
		if len(group.items) == 0 {
			continue
		}

		sb.WriteString(utils.FormatSubsection(group.name, ""))
		sb.WriteString("\n")

		for _, d := range group.items {
			// Create a formatted diagnostic entry with enhanced information
			title := fmt.Sprintf("Issue %d: %s", issueCount, d.Issue)

			var content strings.Builder
			content.WriteString(utils.FormatKeyValue("Type", d.ErrorType))
			content.WriteString(utils.FormatKeyValue("Severity", strings.ToUpper(d.Severity)))
			content.WriteString(utils.FormatKeyValue("Details", d.Details))

			// For AI-enhanced diagnostics, only show the AI's Markdown block (Details), skip extracted steps/docs
			if d.ErrorType == "ai_analysis" {
				sb.WriteString(utils.FormatBox(title, content.String()))
				sb.WriteString("\n")
				issueCount++
				continue
			}

			// Add fix steps if available
			if len(d.Steps) > 0 {
				content.WriteString(utils.FormatSubsection("🔧 Fix Steps", ""))
				for i, step := range d.Steps {
					content.WriteString(fmt.Sprintf("   %d. %s\n", i+1, step))
				}
			}

			// Add documentation links if available
			if len(d.DocsLinks) > 0 {
				content.WriteString(utils.FormatSubsection("📚 Documentation", ""))
				for _, link := range d.DocsLinks {
					content.WriteString(fmt.Sprintf("   • %s\n", link))
				}
			}

			sb.WriteString(utils.FormatBox(title, content.String()))
			sb.WriteString("\n")
			issueCount++
		}
	}

	// Add a helpful footer
	sb.WriteString(utils.FormatDivider())
	sb.WriteString("\n")
	sb.WriteString(utils.FormatTip("Run 'nixai decode-error' for specialized error analysis"))
	sb.WriteString(utils.FormatNote("Use 'nixai interactive' for step-by-step troubleshooting"))

	return sb.String()
}

// filterBySeverity filters diagnostics by severity level
func filterBySeverity(diags []Diagnostic, severity string) []Diagnostic {
	var filtered []Diagnostic
	for _, diag := range diags {
		if diag.Severity == severity {
			filtered = append(filtered, diag)
		}
	}
	return filtered
}
