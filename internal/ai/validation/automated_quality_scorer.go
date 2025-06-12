package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// AutomatedQualityScorer uses local Nix commands to provide comprehensive quality scoring
type AutomatedQualityScorer struct {
	logger *logger.Logger
}

// AutomatedQualityScore represents a comprehensive quality assessment
type AutomatedQualityScore struct {
	OverallScore      int                     `json:"overall_score"` // 0-100
	BreakdownScores   ScoreBreakdown          `json:"breakdown_scores"`
	ValidationResults ValidationResults       `json:"validation_results"`
	ExecutionTime     time.Duration           `json:"execution_time"`
	CommandsRun       []string                `json:"commands_run"`
	Issues            []AutomatedQualityIssue `json:"issues"`
	Recommendations   []string                `json:"recommendations"`
	Metadata          map[string]interface{}  `json:"metadata"`
}

// ScoreBreakdown shows points for each validation category
type ScoreBreakdown struct {
	SyntaxScore    int `json:"syntax_score"`    // 0-30 points
	PackageScore   int `json:"package_score"`   // 0-25 points
	OptionScore    int `json:"option_score"`    // 0-25 points
	CommandScore   int `json:"command_score"`   // 0-10 points
	StructureScore int `json:"structure_score"` // 0-10 points
}

// ValidationResults contains detailed validation outcomes
type ValidationResults struct {
	SyntaxValid        bool            `json:"syntax_valid"`
	PackagesValid      []PackageResult `json:"packages_valid"`
	OptionsValid       []OptionResult  `json:"options_valid"`
	CommandsValid      []CommandResult `json:"commands_valid"`
	FlakeValid         bool            `json:"flake_valid"`
	ConfigurationValid bool            `json:"configuration_valid"`
}

// PackageResult represents package validation outcome
type PackageResult struct {
	Name        string `json:"name"`
	Exists      bool   `json:"exists"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	AttrPath    string `json:"attr_path,omitempty"`
}

// OptionResult represents option validation outcome
type OptionResult struct {
	Name    string `json:"name"`
	Valid   bool   `json:"valid"`
	Type    string `json:"type,omitempty"`
	Default string `json:"default,omitempty"`
}

// CommandResult represents command validation outcome
type CommandResult struct {
	Command   string `json:"command"`
	Available bool   `json:"available"`
	Valid     bool   `json:"valid"`
}

// AutomatedQualityIssue represents a specific quality concern from automated scoring
type AutomatedQualityIssue struct {
	Category   string `json:"category"` // "syntax", "package", "option", "command", "structure"
	Severity   string `json:"severity"` // "low", "medium", "high", "critical"
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	LineNumber int    `json:"line_number,omitempty"`
}

// NewAutomatedQualityScorer creates a new automated quality scorer
func NewAutomatedQualityScorer() *AutomatedQualityScorer {
	return &AutomatedQualityScorer{
		logger: logger.NewLogger(),
	}
}

// ScoreAnswer performs comprehensive quality scoring using local Nix commands
func (aqs *AutomatedQualityScorer) ScoreAnswer(ctx context.Context, question, answer string) (*AutomatedQualityScore, error) {
	startTime := time.Now()

	score := &AutomatedQualityScore{
		BreakdownScores: ScoreBreakdown{},
		ValidationResults: ValidationResults{
			PackagesValid: []PackageResult{},
			OptionsValid:  []OptionResult{},
			CommandsValid: []CommandResult{},
		},
		CommandsRun:     []string{},
		Issues:          []AutomatedQualityIssue{},
		Recommendations: []string{},
		Metadata:        make(map[string]interface{}),
	}

	// Step 1: Syntax Validation (30 points max)
	syntaxScore := aqs.validateSyntax(ctx, answer, score)
	score.BreakdownScores.SyntaxScore = syntaxScore

	// Step 2: Package Verification (25 points max)
	packageScore := aqs.validatePackages(ctx, answer, score)
	score.BreakdownScores.PackageScore = packageScore

	// Step 3: Option Verification (25 points max)
	optionScore := aqs.validateOptions(ctx, answer, score)
	score.BreakdownScores.OptionScore = optionScore

	// Step 4: Command Availability (10 points max)
	commandScore := aqs.validateCommands(ctx, answer, score)
	score.BreakdownScores.CommandScore = commandScore

	// Step 5: Structure Quality (10 points max)
	structureScore := aqs.validateStructure(ctx, answer, score)
	score.BreakdownScores.StructureScore = structureScore

	// Calculate overall score
	score.OverallScore = syntaxScore + packageScore + optionScore + commandScore + structureScore
	score.ExecutionTime = time.Since(startTime)

	// Generate recommendations based on scoring
	aqs.generateRecommendations(score)

	aqs.logger.Printf("Automated quality scoring completed: %d/100 points in %v",
		score.OverallScore, score.ExecutionTime)

	return score, nil
}

// validateSyntax checks Nix syntax using local commands (30 points max)
func (aqs *AutomatedQualityScorer) validateSyntax(ctx context.Context, answer string, score *AutomatedQualityScore) int {
	points := 0

	// Extract Nix expressions from the answer
	expressions := aqs.extractNixExpressions(answer)

	if len(expressions) == 0 {
		// No Nix code to validate, give partial points if it's descriptive
		if len(answer) > 50 {
			return 15 // Partial points for descriptive answers
		}
		return 5
	}

	validExpressions := 0

	for _, expr := range expressions {
		// Test syntax with nix-instantiate --parse
		cmd := exec.CommandContext(ctx, "nix-instantiate", "--parse", "-")
		cmd.Stdin = strings.NewReader(expr)
		score.CommandsRun = append(score.CommandsRun, "nix-instantiate --parse")

		output, err := cmd.CombinedOutput()
		if err == nil {
			validExpressions++
			score.ValidationResults.SyntaxValid = true
		} else {
			score.Issues = append(score.Issues, AutomatedQualityIssue{
				Category:   "syntax",
				Severity:   "high",
				Message:    fmt.Sprintf("Syntax error in Nix expression: %s", string(output)),
				Suggestion: "Check Nix syntax - missing semicolons, braces, or invalid expressions",
			})
		}
	}

	// Calculate syntax points (up to 30)
	if len(expressions) > 0 {
		syntaxRatio := float64(validExpressions) / float64(len(expressions))
		points = int(syntaxRatio * 30)
	}

	// Bonus points for flake validation
	if aqs.containsFlakeCode(answer) {
		flakeValid := aqs.validateFlakeStructure(ctx, answer, score)
		if flakeValid {
			points += 5 // Bonus for valid flake
			score.ValidationResults.FlakeValid = true
		}
	}

	return min(points, 30)
}

// validatePackages verifies package existence using nix commands (25 points max)
func (aqs *AutomatedQualityScorer) validatePackages(ctx context.Context, answer string, score *AutomatedQualityScore) int {
	packages := aqs.extractPackageNames(answer)

	if len(packages) == 0 {
		return 15 // Partial points if no packages mentioned
	}

	validPackages := 0

	for _, pkg := range packages {
		result := PackageResult{Name: pkg}

		// Method 1: nix search
		cmd := exec.CommandContext(ctx, "nix", "search", "nixpkgs", pkg, "--json")
		score.CommandsRun = append(score.CommandsRun, fmt.Sprintf("nix search nixpkgs %s --json", pkg))

		output, err := cmd.Output()
		if err == nil && len(output) > 2 {
			result.Exists = true

			// Try to extract metadata from JSON
			var searchResult map[string]interface{}
			if json.Unmarshal(output, &searchResult) == nil {
				for _, data := range searchResult {
					if pkgData, ok := data.(map[string]interface{}); ok {
						if desc, ok := pkgData["description"].(string); ok {
							result.Description = desc
						}
						if version, ok := pkgData["version"].(string); ok {
							result.Version = version
						}
					}
					break // Just take first match
				}
			}

			validPackages++
		} else {
			// Method 2: nix-env -qaP as fallback
			cmd = exec.CommandContext(ctx, "nix-env", "-qaP", pkg)
			score.CommandsRun = append(score.CommandsRun, fmt.Sprintf("nix-env -qaP %s", pkg))

			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				result.Exists = true
				result.AttrPath = strings.TrimSpace(string(output))
				validPackages++
			} else {
				score.Issues = append(score.Issues, AutomatedQualityIssue{
					Category:   "package",
					Severity:   "medium",
					Message:    fmt.Sprintf("Package '%s' not found in nixpkgs", pkg),
					Suggestion: fmt.Sprintf("Verify package name or check if '%s' is available in current nixpkgs channel", pkg),
				})
			}
		}

		score.ValidationResults.PackagesValid = append(score.ValidationResults.PackagesValid, result)
	}

	// Calculate package points (up to 25)
	if len(packages) > 0 {
		packageRatio := float64(validPackages) / float64(len(packages))
		return int(packageRatio * 25)
	}

	return 15 // Default for answers without packages
}

// validateOptions verifies NixOS options using local commands (25 points max)
func (aqs *AutomatedQualityScorer) validateOptions(ctx context.Context, answer string, score *AutomatedQualityScore) int {
	options := aqs.extractOptionNames(answer)

	if len(options) == 0 {
		return 15 // Partial points if no options mentioned
	}

	validOptions := 0

	for _, opt := range options {
		result := OptionResult{Name: opt}

		// Use nixos-option to validate
		cmd := exec.CommandContext(ctx, "nixos-option", opt)
		score.CommandsRun = append(score.CommandsRun, fmt.Sprintf("nixos-option %s", opt))

		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			result.Valid = true
			validOptions++

			// Parse output for additional info
			outputStr := string(output)
			if typeMatch := regexp.MustCompile(`Type:\s*(.+)`).FindStringSubmatch(outputStr); len(typeMatch) > 1 {
				result.Type = strings.TrimSpace(typeMatch[1])
			}
			if defaultMatch := regexp.MustCompile(`Default:\s*(.+)`).FindStringSubmatch(outputStr); len(defaultMatch) > 1 {
				result.Default = strings.TrimSpace(defaultMatch[1])
			}
		} else {
			score.Issues = append(score.Issues, AutomatedQualityIssue{
				Category:   "option",
				Severity:   "medium",
				Message:    fmt.Sprintf("NixOS option '%s' not found or invalid", opt),
				Suggestion: fmt.Sprintf("Check option name '%s' in NixOS manual or options search", opt),
			})
		}

		score.ValidationResults.OptionsValid = append(score.ValidationResults.OptionsValid, result)
	}

	// Calculate option points (up to 25)
	if len(options) > 0 {
		optionRatio := float64(validOptions) / float64(len(options))
		return int(optionRatio * 25)
	}

	return 15 // Default for answers without options
}

// validateCommands checks command availability (10 points max)
func (aqs *AutomatedQualityScorer) validateCommands(ctx context.Context, answer string, score *AutomatedQualityScore) int {
	commands := aqs.extractCommands(answer)

	if len(commands) == 0 {
		return 5 // Partial points if no commands mentioned
	}

	validCommands := 0

	for _, cmd := range commands {
		result := CommandResult{Command: cmd}

		// Extract binary name
		parts := strings.Fields(cmd)
		if len(parts) > 0 {
			binary := parts[0]

			// Check with 'which'
			whichCmd := exec.CommandContext(ctx, "which", binary)
			score.CommandsRun = append(score.CommandsRun, fmt.Sprintf("which %s", binary))

			if err := whichCmd.Run(); err == nil {
				result.Available = true
				result.Valid = true
				validCommands++
			} else {
				score.Issues = append(score.Issues, AutomatedQualityIssue{
					Category:   "command",
					Severity:   "low",
					Message:    fmt.Sprintf("Command '%s' not available", binary),
					Suggestion: fmt.Sprintf("Install package containing '%s' or verify command name", binary),
				})
			}
		}

		score.ValidationResults.CommandsValid = append(score.ValidationResults.CommandsValid, result)
	}

	// Calculate command points (up to 10)
	if len(commands) > 0 {
		commandRatio := float64(validCommands) / float64(len(commands))
		return int(commandRatio * 10)
	}

	return 5 // Default for answers without commands
}

// validateStructure assesses code structure and best practices (10 points max)
func (aqs *AutomatedQualityScorer) validateStructure(ctx context.Context, answer string, score *AutomatedQualityScore) int {
	points := 0

	// Check for good practices
	if strings.Contains(answer, "with pkgs;") && strings.Contains(answer, "[") && strings.Contains(answer, "]") {
		points += 2 // Good package list structure
	}

	if strings.Contains(answer, "configuration.nix") || strings.Contains(answer, "home.nix") {
		points += 2 // References proper config files
	}

	if strings.Contains(answer, "nixos-rebuild") || strings.Contains(answer, "home-manager") {
		points += 2 // Mentions proper rebuild commands
	}

	if aqs.containsCodeBlocks(answer) {
		points += 2 // Well-formatted with code blocks
	}

	// Check for flake best practices
	if aqs.containsFlakeCode(answer) {
		if strings.Contains(answer, "inputs") && strings.Contains(answer, "outputs") {
			points += 2 // Proper flake structure
		}
	}

	return min(points, 10)
}

// validateFlakeStructure validates flake using nix flake commands
func (aqs *AutomatedQualityScorer) validateFlakeStructure(ctx context.Context, answer string, score *AutomatedQualityScore) bool {
	flakeCode := aqs.extractFlakeCode(answer)
	if flakeCode == "" {
		return false
	}

	// Create temporary flake for testing
	tmpDir, err := os.MkdirTemp("", "flake-validation-")
	if err != nil {
		return false
	}
	defer os.RemoveAll(tmpDir)

	flakePath := tmpDir + "/flake.nix"
	if err := os.WriteFile(flakePath, []byte(flakeCode), 0644); err != nil {
		return false
	}

	// Test with nix flake check
	cmd := exec.CommandContext(ctx, "nix", "flake", "check", tmpDir, "--no-build")
	score.CommandsRun = append(score.CommandsRun, "nix flake check --no-build")

	return cmd.Run() == nil
}

// Helper methods for extraction

func (aqs *AutomatedQualityScorer) extractNixExpressions(answer string) []string {
	var expressions []string

	// Extract code blocks
	codeBlockPattern := regexp.MustCompile("```(?:nix)?\n([^`]+)```")
	matches := codeBlockPattern.FindAllStringSubmatch(answer, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expr := strings.TrimSpace(match[1])
			if aqs.looksLikeNixExpression(expr) {
				expressions = append(expressions, expr)
			}
		}
	}

	return expressions
}

func (aqs *AutomatedQualityScorer) extractPackageNames(answer string) []string {
	var packages []string

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`pkgs\.([a-zA-Z0-9\-_]+)`),
		regexp.MustCompile(`with pkgs; \[([^\]]+)\]`),
		regexp.MustCompile(`nix-shell -p ([a-zA-Z0-9\-_\s]+)`),
		regexp.MustCompile(`nix-env -iA nixos\.([a-zA-Z0-9\-_]+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(answer, -1)
		for _, match := range matches {
			if len(match) > 1 {
				pkgList := strings.Fields(match[1])
				for _, pkg := range pkgList {
					cleaned := strings.Trim(pkg, "[]{}()\"',")
					if len(cleaned) > 1 {
						packages = append(packages, cleaned)
					}
				}
			}
		}
	}

	return aqs.removeDuplicates(packages)
}

func (aqs *AutomatedQualityScorer) extractOptionNames(answer string) []string {
	var options []string

	optionPattern := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9]*(?:\.[a-zA-Z][a-zA-Z0-9]*){1,})\s*=`)
	matches := optionPattern.FindAllStringSubmatch(answer, -1)

	for _, match := range matches {
		if len(match) > 1 {
			options = append(options, match[1])
		}
	}

	return aqs.removeDuplicates(options)
}

func (aqs *AutomatedQualityScorer) extractCommands(answer string) []string {
	var commands []string

	commandPatterns := []*regexp.Regexp{
		regexp.MustCompile(`nixos-rebuild\s+[a-z-]+`),
		regexp.MustCompile(`nix-env\s+[^;\n]+`),
		regexp.MustCompile(`nix\s+[a-z-]+\s+[^;\n]+`),
		regexp.MustCompile(`systemctl\s+[a-z]+\s+[^;\n]+`),
		regexp.MustCompile(`home-manager\s+[a-z-]+`),
	}

	for _, pattern := range commandPatterns {
		matches := pattern.FindAllString(answer, -1)
		commands = append(commands, matches...)
	}

	return aqs.removeDuplicates(commands)
}

func (aqs *AutomatedQualityScorer) looksLikeNixExpression(text string) bool {
	nixIndicators := []string{"{", "}", "=", ";", "with", "let", "in", "import", "pkgs", "lib"}

	for _, indicator := range nixIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}

	return false
}

func (aqs *AutomatedQualityScorer) containsFlakeCode(answer string) bool {
	flakeIndicators := []string{"inputs", "outputs", "flake.nix", "description ="}

	for _, indicator := range flakeIndicators {
		if strings.Contains(answer, indicator) {
			return true
		}
	}

	return false
}

func (aqs *AutomatedQualityScorer) containsCodeBlocks(answer string) bool {
	return strings.Contains(answer, "```")
}

func (aqs *AutomatedQualityScorer) extractFlakeCode(answer string) string {
	// Extract the first flake.nix content found
	codeBlockPattern := regexp.MustCompile("```(?:nix)?\n([^`]+)```")
	matches := codeBlockPattern.FindAllStringSubmatch(answer, -1)

	for _, match := range matches {
		if len(match) > 1 {
			code := strings.TrimSpace(match[1])
			if strings.Contains(code, "inputs") && strings.Contains(code, "outputs") {
				return code
			}
		}
	}

	return ""
}

func (aqs *AutomatedQualityScorer) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (aqs *AutomatedQualityScorer) generateRecommendations(score *AutomatedQualityScore) {
	if score.BreakdownScores.SyntaxScore < 20 {
		score.Recommendations = append(score.Recommendations,
			"Improve Nix syntax - check for missing semicolons, braces, or syntax errors")
	}

	if score.BreakdownScores.PackageScore < 15 {
		score.Recommendations = append(score.Recommendations,
			"Verify package names exist in nixpkgs - use 'nix search' to confirm availability")
	}

	if score.BreakdownScores.OptionScore < 15 {
		score.Recommendations = append(score.Recommendations,
			"Check NixOS option names - use 'nixos-option' or NixOS manual for valid options")
	}

	if score.BreakdownScores.CommandScore < 5 {
		score.Recommendations = append(score.Recommendations,
			"Ensure referenced commands are available on target system")
	}

	if score.BreakdownScores.StructureScore < 5 {
		score.Recommendations = append(score.Recommendations,
			"Improve code structure - use proper formatting and follow NixOS best practices")
	}

	if score.OverallScore < 50 {
		score.Recommendations = append(score.Recommendations,
			"Consider providing more detailed examples and ensure technical accuracy")
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
