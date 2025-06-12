package nixos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// ToolsExecutor executes NixOS tools to validate answers in real-time
type ToolsExecutor struct {
	logger *logger.Logger
}

// ToolValidationResult represents the result of validating an answer using NixOS tools
type ToolValidationResult struct {
	PackageChecks    []PackageCheckResult `json:"package_checks"`
	OptionChecks     []OptionCheckResult  `json:"option_checks"`
	SyntaxChecks     []SyntaxCheckResult  `json:"syntax_checks"`
	CommandChecks    []CommandCheckResult `json:"command_checks"`
	FailedChecks     []string             `json:"failed_checks"`
	SuccessfulChecks []string             `json:"successful_checks"`
	Confidence       float64              `json:"confidence"`
	ExecutionTime    time.Duration        `json:"execution_time"`
}

// PackageCheckResult represents the result of checking a package
type PackageCheckResult struct {
	PackageName string `json:"package_name"`
	Available   bool   `json:"available"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Error       string `json:"error,omitempty"`
}

// OptionCheckResult represents the result of checking a NixOS option
type OptionCheckResult struct {
	OptionName  string `json:"option_name"`
	Valid       bool   `json:"valid"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Error       string `json:"error,omitempty"`
}

// SyntaxCheckResult represents the result of checking Nix syntax
type SyntaxCheckResult struct {
	Content string `json:"content"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// CommandCheckResult represents the result of checking a command
type CommandCheckResult struct {
	Command   string `json:"command"`
	Valid     bool   `json:"valid"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// NewToolsExecutor creates a new NixOS tools executor
func NewToolsExecutor() *ToolsExecutor {
	return &ToolsExecutor{
		logger: logger.NewLogger(),
	}
}

// ValidateAnswer validates an AI-generated answer using various NixOS tools
func (te *ToolsExecutor) ValidateAnswer(ctx context.Context, answer string) (*ToolValidationResult, error) {
	startTime := time.Now()

	result := &ToolValidationResult{
		PackageChecks:    []PackageCheckResult{},
		OptionChecks:     []OptionCheckResult{},
		SyntaxChecks:     []SyntaxCheckResult{},
		CommandChecks:    []CommandCheckResult{},
		FailedChecks:     []string{},
		SuccessfulChecks: []string{},
		Confidence:       1.0,
	}

	// Check packages mentioned in the answer
	packages := te.extractPackageReferences(answer)
	for _, pkg := range packages {
		checkResult := te.checkPackageAvailability(ctx, pkg)
		result.PackageChecks = append(result.PackageChecks, checkResult)

		if checkResult.Available {
			result.SuccessfulChecks = append(result.SuccessfulChecks, fmt.Sprintf("package:%s", pkg))
		} else {
			result.FailedChecks = append(result.FailedChecks, fmt.Sprintf("package:%s", pkg))
		}
	}

	// Check NixOS options mentioned in the answer
	options := te.extractOptionReferences(answer)
	for _, opt := range options {
		checkResult := te.checkOptionValidity(ctx, opt)
		result.OptionChecks = append(result.OptionChecks, checkResult)

		if checkResult.Valid {
			result.SuccessfulChecks = append(result.SuccessfulChecks, fmt.Sprintf("option:%s", opt))
		} else {
			result.FailedChecks = append(result.FailedChecks, fmt.Sprintf("option:%s", opt))
		}
	}

	// Check Nix expressions for syntax validity
	nixExpressions := te.extractNixExpressions(answer)
	for _, expr := range nixExpressions {
		checkResult := te.checkNixSyntax(ctx, expr)
		result.SyntaxChecks = append(result.SyntaxChecks, checkResult)

		if checkResult.Valid {
			result.SuccessfulChecks = append(result.SuccessfulChecks, "syntax:valid")
		} else {
			result.FailedChecks = append(result.FailedChecks, "syntax:invalid")
		}
	}

	// Check commands mentioned in the answer
	commands := te.extractCommandReferences(answer)
	for _, cmd := range commands {
		checkResult := te.checkCommandAvailability(ctx, cmd)
		result.CommandChecks = append(result.CommandChecks, checkResult)

		if checkResult.Available {
			result.SuccessfulChecks = append(result.SuccessfulChecks, fmt.Sprintf("command:%s", cmd))
		} else {
			result.FailedChecks = append(result.FailedChecks, fmt.Sprintf("command:%s", cmd))
		}
	}

	// Calculate confidence based on validation results
	result.Confidence = te.calculateToolConfidence(result)
	result.ExecutionTime = time.Since(startTime)

	te.logger.Printf("Tool validation completed: successful_checks=%d, failed_checks=%d, confidence=%.2f, execution_time=%v",
		len(result.SuccessfulChecks), len(result.FailedChecks), result.Confidence, result.ExecutionTime)

	return result, nil
}

// checkPackageAvailability checks if a package is available using nix search
func (te *ToolsExecutor) checkPackageAvailability(ctx context.Context, packageName string) PackageCheckResult {
	result := PackageCheckResult{
		PackageName: packageName,
		Available:   false,
	}

	// Use nix search to check package availability
	cmd := exec.CommandContext(ctx, "nix", "search", "nixpkgs", packageName, "--json")
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("nix search failed: %v", err)
		return result
	}

	// Parse the JSON output to determine availability
	if len(output) > 2 { // More than just "{}"
		result.Available = true
		// TODO: Parse JSON to extract version and description
	}

	return result
}

// checkOptionValidity checks if a NixOS option is valid using nixos-option
func (te *ToolsExecutor) checkOptionValidity(ctx context.Context, optionName string) OptionCheckResult {
	result := OptionCheckResult{
		OptionName: optionName,
		Valid:      false,
	}

	// Use nixos-option to check option validity
	cmd := exec.CommandContext(ctx, "nixos-option", optionName)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("nixos-option failed: %v", err)
		return result
	}

	if len(output) > 0 {
		result.Valid = true
		// Parse output to extract type, description, and default
		te.parseNixosOptionOutput(string(output), &result)
	}

	return result
}

// checkNixSyntax checks if a Nix expression has valid syntax
func (te *ToolsExecutor) checkNixSyntax(ctx context.Context, expression string) SyntaxCheckResult {
	result := SyntaxCheckResult{
		Content: expression,
		Valid:   false,
	}

	// Use nix-instantiate to check syntax
	cmd := exec.CommandContext(ctx, "nix-instantiate", "--parse", "-")
	cmd.Stdin = strings.NewReader(expression)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Error = string(output)
		// Try to extract line and column information from error
		te.parseNixSyntaxError(string(output), &result)
	} else {
		result.Valid = true
	}

	return result
}

// checkCommandAvailability checks if a command is available in the system
func (te *ToolsExecutor) checkCommandAvailability(ctx context.Context, command string) CommandCheckResult {
	result := CommandCheckResult{
		Command:   command,
		Valid:     false,
		Available: false,
	}

	// Split command to get the binary name
	parts := strings.Fields(command)
	if len(parts) == 0 {
		result.Error = "empty command"
		return result
	}

	binary := parts[0]

	// Check if command is available
	cmd := exec.CommandContext(ctx, "which", binary)
	err := cmd.Run()
	if err == nil {
		result.Available = true
		result.Valid = true
	} else {
		result.Error = fmt.Sprintf("command not found: %s", binary)
	}

	return result
}

// extractPackageReferences extracts package references from answer text
func (te *ToolsExecutor) extractPackageReferences(answer string) []string {
	var packages []string

	// Patterns for package references
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

	return removeDuplicates(packages)
}

// extractOptionReferences extracts NixOS option references from answer text
func (te *ToolsExecutor) extractOptionReferences(answer string) []string {
	var options []string

	// Pattern for NixOS options
	optionPattern := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9]*(?:\.[a-zA-Z][a-zA-Z0-9]*){1,})\s*=`)

	matches := optionPattern.FindAllStringSubmatch(answer, -1)
	for _, match := range matches {
		if len(match) > 1 {
			options = append(options, match[1])
		}
	}

	return removeDuplicates(options)
}

// extractNixExpressions extracts Nix expressions from code blocks in the answer
func (te *ToolsExecutor) extractNixExpressions(answer string) []string {
	var expressions []string

	// Look for code blocks that might contain Nix expressions
	codeBlockPattern := regexp.MustCompile("```(?:nix)?\n([^`]+)```")
	matches := codeBlockPattern.FindAllStringSubmatch(answer, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expr := strings.TrimSpace(match[1])
			if len(expr) > 0 && te.looksLikeNixExpression(expr) {
				expressions = append(expressions, expr)
			}
		}
	}

	return expressions
}

// extractCommandReferences extracts command references from answer text
func (te *ToolsExecutor) extractCommandReferences(answer string) []string {
	var commands []string

	// Patterns for common NixOS commands
	commandPatterns := []*regexp.Regexp{
		regexp.MustCompile(`nixos-rebuild\s+[a-z-]+`),
		regexp.MustCompile(`nix-env\s+[^;\n]+`),
		regexp.MustCompile(`nix\s+[a-z-]+\s+[^;\n]+`),
		regexp.MustCompile(`systemctl\s+[a-z]+\s+[^;\n]+`),
	}

	for _, pattern := range commandPatterns {
		matches := pattern.FindAllString(answer, -1)
		commands = append(commands, matches...)
	}

	return removeDuplicates(commands)
}

// looksLikeNixExpression determines if text looks like a Nix expression
func (te *ToolsExecutor) looksLikeNixExpression(text string) bool {
	nixIndicators := []string{
		"{", "}", "=", ";", "with", "let", "in", "import", "pkgs", "lib",
	}

	for _, indicator := range nixIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}

	return false
}

// parseNixosOptionOutput parses the output from nixos-option command
func (te *ToolsExecutor) parseNixosOptionOutput(output string, result *OptionCheckResult) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Type:") {
			result.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
		} else if strings.HasPrefix(line, "Default:") {
			result.Default = strings.TrimSpace(strings.TrimPrefix(line, "Default:"))
		} else if strings.HasPrefix(line, "Description:") {
			result.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		}
	}
}

// parseNixSyntaxError parses Nix syntax errors to extract line and column information
func (te *ToolsExecutor) parseNixSyntaxError(output string, result *SyntaxCheckResult) {
	// Try to extract line number from error message
	linePattern := regexp.MustCompile(`line (\d+)`)
	if matches := linePattern.FindStringSubmatch(output); len(matches) > 1 {
		if line := parseInt(matches[1]); line > 0 {
			result.Line = line
		}
	}

	// Try to extract column number from error message
	columnPattern := regexp.MustCompile(`column (\d+)`)
	if matches := columnPattern.FindStringSubmatch(output); len(matches) > 1 {
		if column := parseInt(matches[1]); column > 0 {
			result.Column = column
		}
	}
}

// calculateToolConfidence calculates confidence based on tool validation results
func (te *ToolsExecutor) calculateToolConfidence(result *ToolValidationResult) float64 {
	totalChecks := len(result.SuccessfulChecks) + len(result.FailedChecks)
	if totalChecks == 0 {
		return 1.0
	}

	return float64(len(result.SuccessfulChecks)) / float64(totalChecks)
}

// Helper function to parse integers safely
func parseInt(s string) int {
	// Simple integer parsing without importing strconv
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}

// Enhanced verification methods using additional local Nix commands

// verifyPackageMetadata gets detailed package information using multiple commands
func (te *ToolsExecutor) verifyPackageMetadata(ctx context.Context, packageName string) PackageMetadata {
	metadata := PackageMetadata{
		Name:      packageName,
		Available: false,
	}

	// Method 1: nix-env -qaP (attribute path query)
	cmd := exec.CommandContext(ctx, "nix-env", "-qaP", packageName)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		metadata.Available = true
		metadata.AttributePath = strings.TrimSpace(string(output))
	}

	// Method 2: nix eval for metadata (if available)
	if metadata.Available {
		attrPath := fmt.Sprintf("nixpkgs#%s.meta.description", packageName)
		cmd = exec.CommandContext(ctx, "nix", "eval", attrPath, "--raw")
		if output, err := cmd.Output(); err == nil {
			metadata.Description = strings.TrimSpace(string(output))
		}

		// Get version information
		versionPath := fmt.Sprintf("nixpkgs#%s.version", packageName)
		cmd = exec.CommandContext(ctx, "nix", "eval", versionPath, "--raw")
		if output, err := cmd.Output(); err == nil {
			metadata.Version = strings.TrimSpace(string(output))
		}
	}

	return metadata
}

// verifyOptionDetails gets comprehensive option information
func (te *ToolsExecutor) verifyOptionDetails(ctx context.Context, optionName string) OptionCheckResult {
	details := OptionCheckResult{
		OptionName: optionName,
		Valid:      false,
	}

	// Method 1: nixos-option (current method)
	cmd := exec.CommandContext(ctx, "nixos-option", optionName)
	if output, err := cmd.Output(); err == nil {
		details.Valid = true
		te.parseNixosOptionOutput(string(output), &details)
	}

	if !details.Valid {
		return details
	}

	// Method 2: nix-instantiate for type information (if not already populated)
	if details.Type == "" {
		typeExpr := fmt.Sprintf("(import <nixpkgs/nixos> {}).options.%s.type", optionName)
		cmd = exec.CommandContext(ctx, "nix-instantiate", "--eval", "-E", typeExpr)
		if output, err := cmd.Output(); err == nil {
			details.Type = strings.Trim(strings.TrimSpace(string(output)), "\"")
		}
	}

	// Method 3: Default value (if not already populated)
	if details.Default == "" {
		defaultExpr := fmt.Sprintf("(import <nixpkgs/nixos> {}).options.%s.default or null", optionName)
		cmd = exec.CommandContext(ctx, "nix-instantiate", "--eval", "-E", defaultExpr)
		if output, err := cmd.Output(); err == nil {
			defaultVal := strings.TrimSpace(string(output))
			if defaultVal != "null" {
				details.Default = defaultVal
			}
		}
	}

	return details
}

// validateConfigurationSyntax performs comprehensive configuration validation
func (te *ToolsExecutor) validateConfigurationSyntax(ctx context.Context, configContent string) ConfigValidationResult {
	result := ConfigValidationResult{
		Valid:   false,
		Content: configContent,
	}

	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "nixos-config-test-*.nix")
	if err != nil {
		result.Error = fmt.Sprintf("failed to create temp file: %v", err)
		return result
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		result.Error = fmt.Sprintf("failed to write config: %v", err)
		return result
	}
	tmpFile.Close()

	// Method 1: Basic syntax check
	cmd := exec.CommandContext(ctx, "nix-instantiate", "--parse", tmpFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Error = string(output)
		return result
	}

	// Method 2: Try to instantiate as NixOS configuration
	nixosExpr := fmt.Sprintf("(import <nixpkgs/nixos> { configuration = %s; }).config.system.build.toplevel", tmpFile.Name())
	cmd = exec.CommandContext(ctx, "nix-instantiate", "--eval", "-E", nixosExpr, "--show-trace")
	if output, err := cmd.CombinedOutput(); err == nil {
		result.Valid = true
		result.BuildPath = strings.TrimSpace(string(output))
	} else {
		// Still syntactically valid, but may have semantic issues
		result.Valid = true
		result.Warnings = append(result.Warnings, string(output))
	}

	return result
}

// validateFlakeExpression validates flake syntax and structure
func (te *ToolsExecutor) validateFlakeExpression(ctx context.Context, flakeContent string) FlakeValidationResult {
	result := FlakeValidationResult{
		Valid:   false,
		Content: flakeContent,
	}

	// Create temporary directory for flake
	tmpDir, err := os.MkdirTemp("", "flake-test-")
	if err != nil {
		result.Error = fmt.Sprintf("failed to create temp dir: %v", err)
		return result
	}
	defer os.RemoveAll(tmpDir)

	flakeFile := filepath.Join(tmpDir, "flake.nix")
	if err := os.WriteFile(flakeFile, []byte(flakeContent), 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write flake: %v", err)
		return result
	}

	// Method 1: nix flake check (no build)
	cmd := exec.CommandContext(ctx, "nix", "flake", "check", tmpDir, "--no-build")
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Error = string(output)
		return result
	}

	result.Valid = true

	// Method 2: nix flake show (get outputs)
	cmd = exec.CommandContext(ctx, "nix", "flake", "show", tmpDir, "--json")
	if output, err := cmd.Output(); err == nil {
		result.Outputs = string(output)
	}

	// Method 3: nix flake metadata
	cmd = exec.CommandContext(ctx, "nix", "flake", "metadata", tmpDir, "--json")
	if output, err := cmd.Output(); err == nil {
		result.Metadata = string(output)
	}

	return result
}

// verifyWithNixRepl uses nix repl for interactive verification
func (te *ToolsExecutor) verifyWithNixRepl(ctx context.Context, expression string) ReplResult {
	result := ReplResult{
		Expression: expression,
		Valid:      false,
	}

	// Use nix repl with expression piped to stdin
	cmd := exec.CommandContext(ctx, "nix", "repl", "<nixpkgs>")
	cmd.Stdin = strings.NewReader(expression + "\n:q\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = string(output)
		return result
	}

	result.Valid = true
	result.Output = string(output)

	return result
}

// performDryRunValidation performs dry-run validation without building
func (te *ToolsExecutor) performDryRunValidation(ctx context.Context, configPath string) DryRunResult {
	result := DryRunResult{
		ConfigPath: configPath,
		Success:    false,
	}

	// Method 1: nixos-rebuild dry-build (if on NixOS)
	if te.isNixOS() {
		cmd := exec.CommandContext(ctx, "nixos-rebuild", "dry-build", "--fast", "--show-trace")
		if configPath != "" {
			cmd.Args = append(cmd.Args, "-I", fmt.Sprintf("nixos-config=%s", configPath))
		}

		output, err := cmd.CombinedOutput()
		result.Output = string(output)

		if err == nil {
			result.Success = true
		} else {
			result.Error = string(output)
		}
	}

	return result
}

// isNixOS checks if running on NixOS
func (te *ToolsExecutor) isNixOS() bool {
	_, err := os.Stat("/etc/NIXOS")
	return err == nil
}

// New data structures for enhanced validation

type PackageMetadata struct {
	Name          string `json:"name"`
	Available     bool   `json:"available"`
	AttributePath string `json:"attribute_path"`
	Description   string `json:"description"`
	Version       string `json:"version"`
	Error         string `json:"error,omitempty"`
}

type ConfigValidationResult struct {
	Valid     bool     `json:"valid"`
	Content   string   `json:"content"`
	BuildPath string   `json:"build_path,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type FlakeValidationResult struct {
	Valid    bool   `json:"valid"`
	Content  string `json:"content"`
	Outputs  string `json:"outputs,omitempty"`
	Metadata string `json:"metadata,omitempty"`
	Error    string `json:"error,omitempty"`
}

type ReplResult struct {
	Expression string `json:"expression"`
	Valid      bool   `json:"valid"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
}

type DryRunResult struct {
	ConfigPath string `json:"config_path"`
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
}
