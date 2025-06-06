package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// LSPPosition represents a position in a text document
type LSPPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// LSPRange represents a range in a text document
type LSPRange struct {
	Start LSPPosition `json:"start"`
	End   LSPPosition `json:"end"`
}

// LSPLocation represents a location in a text document
type LSPLocation struct {
	URI   string   `json:"uri"`
	Range LSPRange `json:"range"`
}

// LSPDiagnostic represents a diagnostic message
type LSPDiagnostic struct {
	Range    LSPRange `json:"range"`
	Severity int      `json:"severity"` // 1=Error, 2=Warning, 3=Information, 4=Hint
	Code     string   `json:"code,omitempty"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

// LSPCompletionItem represents a completion suggestion
type LSPCompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"`          // 1=Text, 2=Method, 3=Function, etc.
	Detail        string `json:"detail"`        // Brief description
	Documentation string `json:"documentation"` // Full documentation
	InsertText    string `json:"insertText"`    // Text to insert
}

// LSPHover represents hover information
type LSPHover struct {
	Contents []string `json:"contents"`
	Range    LSPRange `json:"range,omitempty"`
}

// NixLSPProvider provides LSP-like features for Nix files
type NixLSPProvider struct {
	logger           logger.Logger
	nixosOptions     map[string]NixOSOption
	homeManagerOpts  map[string]string
	diagnosticsCache map[string][]LSPDiagnostic
	completionCache  map[string][]LSPCompletionItem
}

// NewNixLSPProvider creates a new Nix LSP provider
func NewNixLSPProvider(logger logger.Logger) *NixLSPProvider {
	return &NixLSPProvider{
		logger:           logger,
		nixosOptions:     make(map[string]NixOSOption),
		homeManagerOpts:  make(map[string]string),
		diagnosticsCache: make(map[string][]LSPDiagnostic),
		completionCache:  make(map[string][]LSPCompletionItem),
	}
}

// LoadNixOSOptions loads NixOS options for completion and hover
func (nlsp *NixLSPProvider) LoadNixOSOptions() error {
	nlsp.logger.Debug("Loading NixOS options for LSP features")

	// Load from ElasticSearch or cache
	// For now, use static options - in production, load from ES
	staticOptions := map[string]NixOSOption{
		"services.nginx.enable": {
			Name:        "services.nginx.enable",
			Description: "Whether to enable the Nginx web server.",
			OptionType:  "boolean",
			Default:     "false",
			Example:     "true",
		},
		"networking.firewall.enable": {
			Name:        "networking.firewall.enable",
			Description: "Whether to enable the firewall.",
			OptionType:  "boolean",
			Default:     "true",
			Example:     "false",
		},
		"programs.zsh.enable": {
			Name:        "programs.zsh.enable",
			Description: "Whether to configure zsh as an interactive shell.",
			OptionType:  "boolean",
			Default:     "false",
			Example:     "true",
		},
		"users.users": {
			Name:        "users.users",
			Description: "Additional user accounts to be created automatically by the system.",
			OptionType:  "attribute set of (submodule)",
			Default:     "{}",
			Example:     `{ alice = { isNormalUser = true; extraGroups = [ "wheel" ]; }; }`,
		},
		"environment.systemPackages": {
			Name:        "environment.systemPackages",
			Description: "The set of packages that appear in /run/current-system/sw.",
			OptionType:  "list of package",
			Default:     "[]",
			Example:     "[ pkgs.firefox pkgs.git ]",
		},
	}

	nlsp.nixosOptions = staticOptions
	nlsp.logger.Info(fmt.Sprintf("Loaded %d NixOS options for LSP features", len(staticOptions)))
	return nil
}

// ProvideCompletion provides completion suggestions for Nix files
func (nlsp *NixLSPProvider) ProvideCompletion(fileContent string, position LSPPosition) ([]LSPCompletionItem, error) {
	nlsp.logger.Debug(fmt.Sprintf("Providing completions at line %d, char %d", position.Line, position.Character))

	// Parse the current line to determine context
	lines := strings.Split(fileContent, "\n")
	if position.Line >= len(lines) {
		return nil, fmt.Errorf("position out of bounds")
	}

	currentLine := lines[position.Line]
	beforeCursor := currentLine[:position.Character]

	// Check if we're completing an option
	if nlsp.isOptionContext(beforeCursor) {
		return nlsp.completeOptions(beforeCursor)
	}

	// Check if we're in a package context
	if nlsp.isPackageContext(beforeCursor) {
		return nlsp.completePackages(beforeCursor)
	}

	// Check if we're in a string context for system values
	if nlsp.isSystemValueContext(beforeCursor) {
		return nlsp.completeSystemValues(beforeCursor)
	}

	// Default Nix language completions
	return nlsp.completeNixKeywords(beforeCursor)
}

// ProvideDiagnostics provides real-time diagnostics for Nix files
func (nlsp *NixLSPProvider) ProvideDiagnostics(filePath, fileContent string) ([]LSPDiagnostic, error) {
	nlsp.logger.Debug(fmt.Sprintf("Providing diagnostics for %s", filePath))

	var diagnostics []LSPDiagnostic

	// Check syntax with nix-instantiate
	syntaxDiagnostics, err := nlsp.checkSyntax(filePath, fileContent)
	if err != nil {
		nlsp.logger.Error(fmt.Sprintf("Syntax check failed: %v", err))
	} else {
		diagnostics = append(diagnostics, syntaxDiagnostics...)
	}

	// Check for common configuration issues
	configDiagnostics := nlsp.checkConfigurationIssues(fileContent)
	diagnostics = append(diagnostics, configDiagnostics...)

	// Check for deprecated options
	deprecatedDiagnostics := nlsp.checkDeprecatedOptions(fileContent)
	diagnostics = append(diagnostics, deprecatedDiagnostics...)

	// Cache results
	nlsp.diagnosticsCache[filePath] = diagnostics

	return diagnostics, nil
}

// ProvideHover provides hover information for symbols
func (nlsp *NixLSPProvider) ProvideHover(fileContent string, position LSPPosition) (*LSPHover, error) {
	nlsp.logger.Debug(fmt.Sprintf("Providing hover at line %d, char %d", position.Line, position.Character))

	// Get the word at the cursor position
	word, wordRange := nlsp.getWordAtPosition(fileContent, position)
	if word == "" {
		return nil, nil
	}

	// Check if it's a NixOS option
	if option, exists := nlsp.nixosOptions[word]; exists {
		return &LSPHover{
			Contents: []string{
				fmt.Sprintf("**%s**", option.Name),
				fmt.Sprintf("Type: `%s`", option.OptionType),
				fmt.Sprintf("Default: `%s`", option.Default),
				"",
				option.Description,
				"",
				fmt.Sprintf("Example: `%s`", option.Example),
			},
			Range: wordRange,
		}, nil
	}

	// Check if it's a Nix built-in function
	if builtinDoc := nlsp.getNixBuiltinDocumentation(word); builtinDoc != "" {
		return &LSPHover{
			Contents: []string{builtinDoc},
			Range:    wordRange,
		}, nil
	}

	return nil, nil
}

// ProvideDefinition provides go-to-definition functionality
func (nlsp *NixLSPProvider) ProvideDefinition(fileContent string, position LSPPosition) ([]LSPLocation, error) {
	nlsp.logger.Debug(fmt.Sprintf("Providing definition at line %d, char %d", position.Line, position.Character))

	word, _ := nlsp.getWordAtPosition(fileContent, position)
	if word == "" {
		return nil, nil
	}

	// For NixOS options, try to find the definition in nixpkgs
	if option, exists := nlsp.nixosOptions[word]; exists {
		if option.Source != "" {
			// Parse the source location
			return []LSPLocation{
				{
					URI: fmt.Sprintf("https://github.com/NixOS/nixpkgs/blob/master/%s", option.Source),
					Range: LSPRange{
						Start: LSPPosition{Line: 0, Character: 0},
						End:   LSPPosition{Line: 0, Character: 0},
					},
				},
			}, nil
		}
	}

	return nil, nil
}

// Helper methods

func (nlsp *NixLSPProvider) isOptionContext(text string) bool {
	// Check if we're in a context where NixOS options are expected
	patterns := []string{
		`\s+([a-zA-Z][a-zA-Z0-9_]*\.)*[a-zA-Z][a-zA-Z0-9_]*$`,
		`=\s*([a-zA-Z][a-zA-Z0-9_]*\.)*[a-zA-Z][a-zA-Z0-9_]*$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

func (nlsp *NixLSPProvider) isPackageContext(text string) bool {
	// Check if we're in a package list context
	patterns := []string{
		`systemPackages\s*=\s*\[.*pkgs\.`,
		`packages\s*=\s*\[.*pkgs\.`,
		`with\s+pkgs;\s*\[.*`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

func (nlsp *NixLSPProvider) isSystemValueContext(text string) bool {
	// Check if we're completing system-specific values
	patterns := []string{
		`system\s*=\s*"`,
		`hostPlatform\s*=\s*"`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

func (nlsp *NixLSPProvider) completeOptions(text string) ([]LSPCompletionItem, error) {
	// Extract the partial option path
	re := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9_]*\.)*([a-zA-Z][a-zA-Z0-9_]*)$`)
	matches := re.FindStringSubmatch(text)

	var prefix string
	if len(matches) > 0 {
		prefix = matches[0]
	}

	var completions []LSPCompletionItem
	for optionName, option := range nlsp.nixosOptions {
		if strings.HasPrefix(optionName, prefix) {
			completions = append(completions, LSPCompletionItem{
				Label:         optionName,
				Kind:          3, // Field
				Detail:        option.OptionType,
				Documentation: option.Description,
				InsertText:    optionName,
			})
		}
	}

	return completions, nil
}

func (nlsp *NixLSPProvider) completePackages(text string) ([]LSPCompletionItem, error) {
	// Common packages for completion
	commonPackages := []string{
		"firefox", "chromium", "git", "vim", "neovim", "vscode", "docker",
		"nodejs", "python3", "gcc", "rustc", "go", "java", "wget", "curl",
		"htop", "tmux", "zsh", "bash", "fish", "tree", "ripgrep", "fd",
	}

	var completions []LSPCompletionItem
	for _, pkg := range commonPackages {
		completions = append(completions, LSPCompletionItem{
			Label:      pkg,
			Kind:       7, // Module
			Detail:     "package",
			InsertText: pkg,
		})
	}

	return completions, nil
}

func (nlsp *NixLSPProvider) completeSystemValues(text string) ([]LSPCompletionItem, error) {
	systems := []string{
		"x86_64-linux",
		"aarch64-linux",
		"x86_64-darwin",
		"aarch64-darwin",
	}

	var completions []LSPCompletionItem
	for _, system := range systems {
		completions = append(completions, LSPCompletionItem{
			Label:      system,
			Kind:       21, // Constant
			Detail:     "system architecture",
			InsertText: system,
		})
	}

	return completions, nil
}

func (nlsp *NixLSPProvider) completeNixKeywords(text string) ([]LSPCompletionItem, error) {
	keywords := []string{
		"let", "in", "with", "import", "inherit", "rec", "if", "then", "else",
		"assert", "or", "builtins", "derivation", "pkgs", "lib", "config",
	}

	var completions []LSPCompletionItem
	for _, keyword := range keywords {
		completions = append(completions, LSPCompletionItem{
			Label:      keyword,
			Kind:       14, // Keyword
			Detail:     "Nix keyword",
			InsertText: keyword,
		})
	}

	return completions, nil
}

func (nlsp *NixLSPProvider) checkSyntax(filePath, content string) ([]LSPDiagnostic, error) {
	// Write content to a temporary file
	tmpFile, err := os.CreateTemp("", "nixlsp-*.nix")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		return nil, err
	}
	tmpFile.Close()

	// Run nix-instantiate to check syntax
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nix-instantiate", "--parse", "--strict", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	var diagnostics []LSPDiagnostic
	if err != nil {
		// Parse nix error output
		diagnostic := nlsp.parseNixError(string(output))
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
	}

	return diagnostics, nil
}

func (nlsp *NixLSPProvider) parseNixError(output string) *LSPDiagnostic {
	// Parse nix error messages to extract line/column information
	// Example: "error: syntax error, unexpected '}' at /tmp/file.nix:10:5"
	re := regexp.MustCompile(`error: (.+) at .+:(\d+):(\d+)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) >= 4 {
		line, _ := strconv.Atoi(matches[2])
		character, _ := strconv.Atoi(matches[3])

		return &LSPDiagnostic{
			Range: LSPRange{
				Start: LSPPosition{Line: line - 1, Character: character - 1},
				End:   LSPPosition{Line: line - 1, Character: character},
			},
			Severity: 1, // Error
			Source:   "nix-instantiate",
			Message:  matches[1],
		}
	}

	return nil
}

func (nlsp *NixLSPProvider) checkConfigurationIssues(content string) []LSPDiagnostic {
	var diagnostics []LSPDiagnostic
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Check for common issues
		trimmed := strings.TrimSpace(line)

		// Check for deprecated syntax
		if strings.Contains(trimmed, "config.") && !strings.Contains(trimmed, "# config.") {
			diagnostics = append(diagnostics, LSPDiagnostic{
				Range: LSPRange{
					Start: LSPPosition{Line: i, Character: 0},
					End:   LSPPosition{Line: i, Character: len(line)},
				},
				Severity: 2, // Warning
				Source:   "nixai-lsp",
				Message:  "Consider using direct option reference instead of config.*",
			})
		}

		// Check for missing semicolons in attribute sets
		if strings.Contains(trimmed, "=") && !strings.HasSuffix(trimmed, ";") && !strings.HasSuffix(trimmed, "{") {
			diagnostics = append(diagnostics, LSPDiagnostic{
				Range: LSPRange{
					Start: LSPPosition{Line: i, Character: len(line) - 1},
					End:   LSPPosition{Line: i, Character: len(line)},
				},
				Severity: 3, // Information
				Source:   "nixai-lsp",
				Message:  "Consider adding semicolon for clarity",
			})
		}
	}

	return diagnostics
}

func (nlsp *NixLSPProvider) checkDeprecatedOptions(content string) []LSPDiagnostic {
	var diagnostics []LSPDiagnostic
	lines := strings.Split(content, "\n")

	// List of known deprecated options
	deprecatedOptions := map[string]string{
		"services.xserver.displayManager.lightdm": "Use services.xserver.displayManager.lightdm.enable instead",
		"networking.networkmanager":               "Use networking.networkmanager.enable instead",
	}

	for i, line := range lines {
		for deprecated, replacement := range deprecatedOptions {
			if strings.Contains(line, deprecated) {
				start := strings.Index(line, deprecated)
				diagnostics = append(diagnostics, LSPDiagnostic{
					Range: LSPRange{
						Start: LSPPosition{Line: i, Character: start},
						End:   LSPPosition{Line: i, Character: start + len(deprecated)},
					},
					Severity: 2, // Warning
					Source:   "nixai-lsp",
					Message:  fmt.Sprintf("Deprecated option. %s", replacement),
				})
			}
		}
	}

	return diagnostics
}

func (nlsp *NixLSPProvider) getWordAtPosition(content string, position LSPPosition) (string, LSPRange) {
	lines := strings.Split(content, "\n")
	if position.Line >= len(lines) {
		return "", LSPRange{}
	}

	line := lines[position.Line]
	if position.Character >= len(line) {
		return "", LSPRange{}
	}

	// Find word boundaries
	start := position.Character
	end := position.Character

	// Move start backwards
	for start > 0 && (isAlphaNumeric(rune(line[start-1])) || line[start-1] == '.' || line[start-1] == '_') {
		start--
	}

	// Move end forwards
	for end < len(line) && (isAlphaNumeric(rune(line[end])) || line[end] == '.' || line[end] == '_') {
		end++
	}

	word := line[start:end]
	wordRange := LSPRange{
		Start: LSPPosition{Line: position.Line, Character: start},
		End:   LSPPosition{Line: position.Line, Character: end},
	}

	return word, wordRange
}

func (nlsp *NixLSPProvider) getNixBuiltinDocumentation(builtinName string) string {
	builtins := map[string]string{
		"builtins":   "Set containing all built-in functions",
		"import":     "Import and evaluate a Nix expression from a file",
		"derivation": "Create a derivation",
		"map":        "Apply a function to each element in a list",
		"filter":     "Filter a list using a predicate function",
		"length":     "Return the length of a list",
		"head":       "Return the first element of a list",
		"tail":       "Return all but the first element of a list",
		"toString":   "Convert a value to a string",
		"attrNames":  "Return the names of all attributes in an attribute set",
		"hasAttr":    "Check if an attribute exists in an attribute set",
		"getAttr":    "Get the value of an attribute from an attribute set",
	}

	if doc, exists := builtins[builtinName]; exists {
		return fmt.Sprintf("**%s** (built-in)\n\n%s", builtinName, doc)
	}

	return ""
}

func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// FormatDiagnostics formats diagnostics for display
func (nlsp *NixLSPProvider) FormatDiagnostics(diagnostics []LSPDiagnostic) string {
	if len(diagnostics) == 0 {
		return "✅ No issues found"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d issue(s):\n\n", len(diagnostics)))

	for i, diag := range diagnostics {
		severity := "INFO"
		emoji := "ℹ️"
		switch diag.Severity {
		case 1:
			severity = "ERROR"
			emoji = "❌"
		case 2:
			severity = "WARNING"
			emoji = "⚠️"
		case 3:
			severity = "INFO"
			emoji = "ℹ️"
		case 4:
			severity = "HINT"
			emoji = "💡"
		}

		result.WriteString(fmt.Sprintf("%d. %s **%s** (Line %d:%d)\n",
			i+1, emoji, severity, diag.Range.Start.Line+1, diag.Range.Start.Character+1))
		result.WriteString(fmt.Sprintf("   %s\n", diag.Message))
		if diag.Source != "" {
			result.WriteString(fmt.Sprintf("   Source: %s\n", diag.Source))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// FormatCompletions formats completion items for display
func (nlsp *NixLSPProvider) FormatCompletions(completions []LSPCompletionItem) string {
	if len(completions) == 0 {
		return "No completions available"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d completion(s):\n\n", len(completions)))

	for i, comp := range completions {
		result.WriteString(fmt.Sprintf("%d. **%s**", i+1, comp.Label))
		if comp.Detail != "" {
			result.WriteString(fmt.Sprintf(" (%s)", comp.Detail))
		}
		result.WriteString("\n")
		if comp.Documentation != "" {
			result.WriteString(fmt.Sprintf("   %s\n", comp.Documentation))
		}
		result.WriteString("\n")
	}

	return result.String()
}
