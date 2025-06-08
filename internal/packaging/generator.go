package packaging

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/mcp"
)

// DerivationGenerator generates Nix derivations using AI assistance
type DerivationGenerator struct {
	aiProvider ai.AIProvider
	mcpClient  *mcp.MCPClient
}

// NewDerivationGenerator creates a new derivation generator
func NewDerivationGenerator(aiProvider ai.AIProvider, mcpClient *mcp.MCPClient) *DerivationGenerator {
	return &DerivationGenerator{
		aiProvider: aiProvider,
		mcpClient:  mcpClient,
	}
}

// GenerateDerivation generates a Nix derivation for the analyzed repository
func (dg *DerivationGenerator) GenerateDerivation(ctx context.Context, analysis *RepoAnalysis) (string, error) {
	// Get relevant nixpkgs documentation and examples
	nixpkgsContext, err := dg.GetNixpkgsContext(ctx, analysis.BuildSystem, analysis.Language)
	if err != nil {
		return "", fmt.Errorf("failed to get nixpkgs context: %w", err)
	}

	// Create a comprehensive prompt for AI
	prompt := dg.createDerivationPrompt(analysis, nixpkgsContext)

	// Generate derivation using AI
	response, err := dg.aiProvider.Query(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate derivation: %w", err)
	}

	// Extract and clean the derivation from the response
	derivation := dg.ExtractDerivation(response)

	return derivation, nil
}

// GetNixpkgsContext retrieves relevant nixpkgs documentation and examples
func (dg *DerivationGenerator) GetNixpkgsContext(ctx context.Context, buildSystem BuildSystem, language string) (string, error) {
	var contextQueries []string

	// Build system specific queries
	switch buildSystem {
	case BuildSystemCMake:
		contextQueries = append(contextQueries, "cmake derivation examples", "mkDerivation cmake")
	case BuildSystemMeson:
		contextQueries = append(contextQueries, "meson derivation examples", "mesonFlags buildInputs")
	case BuildSystemCargoRust:
		contextQueries = append(contextQueries, "rust cargo derivation", "buildRustPackage cargoSha256")
	case BuildSystemGo:
		contextQueries = append(contextQueries, "go derivation examples", "buildGoModule vendorSha256")
	case BuildSystemNpm:
		contextQueries = append(contextQueries, "nodejs npm derivation", "buildNpmPackage npmDepsHash")
	case BuildSystemPython:
		contextQueries = append(contextQueries, "python derivation examples", "buildPythonPackage")
	case BuildSystemAutotools:
		contextQueries = append(contextQueries, "autotools derivation", "autoreconfHook configureFlags")
	default:
		contextQueries = append(contextQueries, "mkDerivation standard build")
	}

	// Language specific queries
	switch language {
	case "C", "C++":
		contextQueries = append(contextQueries, "c cpp build inputs stdenv.cc")
	case "Rust":
		contextQueries = append(contextQueries, "rust buildRustPackage examples")
	case "Go":
		contextQueries = append(contextQueries, "golang buildGoModule examples")
	case "Python":
		contextQueries = append(contextQueries, "python buildPythonPackage setuptools")
	case "JavaScript", "TypeScript":
		contextQueries = append(contextQueries, "nodejs javascript typescript derivation")
	}

	var allContext strings.Builder

	// Query MCP server for each context
	for _, query := range contextQueries {
		if dg.mcpClient != nil {
			context, err := dg.mcpClient.QueryDocumentation(query)
			if err == nil && context != "" {
				allContext.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", query, context))
			}
		}
	}

	return allContext.String(), nil
}

// createDerivationPrompt creates a comprehensive prompt for AI derivation generation
func (dg *DerivationGenerator) createDerivationPrompt(analysis *RepoAnalysis, nixpkgsContext string) string {
	var prompt strings.Builder

	prompt.WriteString(`You are an expert Nix package maintainer. Generate a Nix derivation for the following project.

PROJECT ANALYSIS:
`)

	prompt.WriteString(fmt.Sprintf("- Project Name: %s\n", analysis.ProjectName))
	prompt.WriteString(fmt.Sprintf("- Build System: %s\n", analysis.BuildSystem))
	prompt.WriteString(fmt.Sprintf("- Primary Language: %s\n", analysis.Language))
	if analysis.License != "" {
		prompt.WriteString(fmt.Sprintf("- License: %s\n", analysis.License))
	}
	if analysis.Description != "" {
		prompt.WriteString(fmt.Sprintf("- Description: %s\n", analysis.Description))
	}
	prompt.WriteString(fmt.Sprintf("- Has Tests: %t\n", analysis.HasTests))

	if len(analysis.BuildFiles) > 0 {
		prompt.WriteString("\nBuild Files Found:\n")
		for _, file := range analysis.BuildFiles {
			prompt.WriteString(fmt.Sprintf("- %s\n", file))
		}
	}

	if len(analysis.Dependencies) > 0 {
		prompt.WriteString("\nDependencies:\n")

		buildDeps := []Dependency{}
		runtimeDeps := []Dependency{}
		devDeps := []Dependency{}

		for _, dep := range analysis.Dependencies {
			switch dep.Type {
			case "build":
				buildDeps = append(buildDeps, dep)
			case "runtime":
				runtimeDeps = append(runtimeDeps, dep)
			case "dev":
				devDeps = append(devDeps, dep)
			}
		}

		if len(buildDeps) > 0 {
			prompt.WriteString("\nBuild Dependencies:\n")
			for _, dep := range buildDeps {
				prompt.WriteString(fmt.Sprintf("- %s", dep.Name))
				if dep.Version != "" {
					prompt.WriteString(fmt.Sprintf(" (%s)", dep.Version))
				}
				if dep.System {
					prompt.WriteString(" [system library]")
				}
				prompt.WriteString("\n")
			}
		}

		if len(runtimeDeps) > 0 {
			prompt.WriteString("\nRuntime Dependencies:\n")
			for _, dep := range runtimeDeps {
				prompt.WriteString(fmt.Sprintf("- %s", dep.Name))
				if dep.Version != "" {
					prompt.WriteString(fmt.Sprintf(" (%s)", dep.Version))
				}
				prompt.WriteString("\n")
			}
		}

		if len(devDeps) > 0 {
			prompt.WriteString("\nDevelopment Dependencies:\n")
			for _, dep := range devDeps {
				prompt.WriteString(fmt.Sprintf("- %s", dep.Name))
				if dep.Version != "" {
					prompt.WriteString(fmt.Sprintf(" (%s)", dep.Version))
				}
				prompt.WriteString("\n")
			}
		}
	}

	if nixpkgsContext != "" {
		prompt.WriteString("\nRELEVANT NIXPKGS DOCUMENTATION AND EXAMPLES:\n")
		prompt.WriteString(nixpkgsContext)
	}

	prompt.WriteString(`
INSTRUCTIONS:
1. Generate a complete Nix derivation that follows nixpkgs conventions
2. Use the appropriate build function for the detected build system
3. Map the detected dependencies to nixpkgs packages when possible
4. Include proper meta attributes (description, license, maintainers, platforms)
5. Add comments explaining any complex build steps
6. Use modern Nix syntax and best practices
7. Include buildInputs, nativeBuildInputs appropriately
8. For unknown dependencies, add comments suggesting manual mapping
9. Include doCheck = true if tests were detected
10. Ensure the derivation is formatted properly with proper indentation

EXAMPLE STRUCTURE FOR GO PROJECTS:
{ stdenv, lib, buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname = "project-name";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "owner";
    repo = "repo";
    rev = "v${version}";
    sha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
  };

  vendorHash = "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=";

  nativeBuildInputs = [ /* build tools */ ];
  buildInputs = [ /* runtime dependencies */ ];

  ldflags = [ "-s" "-w" ];

  doCheck = true;

  meta = with lib; {
    description = "Brief description";
    homepage = "https://github.com/owner/repo";
    license = licenses.mit;
    maintainers = with maintainers; [ ];
    platforms = platforms.unix;
  };
}

OUTPUT FORMAT:
Provide ONLY the Nix derivation code without any explanation or markdown formatting.
The derivation should be a complete, valid Nix expression that can be built.
Start with the function signature { ... }: and end with the closing brace.

DERIVATION:
`)

	return prompt.String()
}

// ExtractDerivation extracts the Nix derivation from AI response
func (dg *DerivationGenerator) ExtractDerivation(response string) string {
	// Remove common markdown formatting if present
	response = strings.ReplaceAll(response, "```nix", "")
	response = strings.ReplaceAll(response, "```", "")

	// Find the derivation start - look for the first { that starts a line or follows certain patterns
	lines := strings.Split(response, "\n")
	var derivationLines []string
	inDerivation := false
	braceCount := 0
	startFound := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines before derivation starts
		if !inDerivation && trimmed == "" {
			continue
		}

		// Look for derivation start patterns:
		// 1. Line that starts with { (function signature)
		// 2. Line that contains a derivation call like "buildGoModule {" or "stdenv.mkDerivation {"
		if !inDerivation {
			// Check if this line starts the derivation
			if strings.HasPrefix(trimmed, "{") ||
				strings.Contains(trimmed, "buildGoModule {") ||
				strings.Contains(trimmed, "mkDerivation {") ||
				strings.Contains(trimmed, "buildRustPackage {") ||
				strings.Contains(trimmed, "buildPythonPackage {") {

				// If this looks like a function signature, start from here
				if strings.HasPrefix(trimmed, "{") && strings.Contains(trimmed, ",") && strings.HasSuffix(trimmed, ":") {
					inDerivation = true
					startFound = true
				} else if i > 0 {
					// Look backwards for the function signature
					for j := i - 1; j >= 0; j-- {
						prevLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(prevLine, "{") && strings.Contains(prevLine, ",") {
							// Found function signature, start from there
							derivationLines = append(derivationLines, lines[j:][:i-j+1]...)
							inDerivation = true
							startFound = true
							break
						}
						if prevLine == "" {
							continue
						}
						// Stop if we hit non-whitespace that doesn't look like a function signature
						break
					}
					if !startFound {
						// No function signature found, start from current line
						inDerivation = true
						startFound = true
					}
				}
			}
		}

		if inDerivation && !startFound {
			derivationLines = append(derivationLines, line)
		}

		if inDerivation {
			// Count braces to know when derivation ends
			for _, char := range line {
				switch char {
				case '{':
					braceCount++
				case '}':
					braceCount--
				}
			}

			// End when braces are balanced and we have at least some content
			if braceCount == 0 && len(derivationLines) > 2 {
				break
			}
		}
	}

	// If we didn't find a proper derivation, try a simpler approach
	if len(derivationLines) == 0 || braceCount != 0 {
		// Look for anything that resembles a derivation block
		response = strings.TrimSpace(response)

		// Try to find the first line that looks like a function signature
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "{") && (strings.Contains(trimmed, "stdenv") || strings.Contains(trimmed, "buildGo") || strings.Contains(trimmed, "fetchFrom")) {
				// Found likely start, take everything from here to the end
				derivationLines = lines[i:]
				break
			}
		}

		// If still nothing found, just clean up what we have
		if len(derivationLines) == 0 {
			return strings.TrimSpace(response)
		}
	}

	derivation := strings.Join(derivationLines, "\n")

	// Clean up the derivation
	derivation = strings.TrimSpace(derivation)

	return derivation
}

// SuggestNixpkgsMappings suggests nixpkgs packages for dependencies
func (dg *DerivationGenerator) SuggestNixpkgsMappings(ctx context.Context, dependencies []Dependency) (map[string]string, error) {
	mappings := make(map[string]string)

	// Common dependency mappings
	commonMappings := map[string]string{
		// System libraries
		"openssl":    "openssl",
		"zlib":       "zlib",
		"libpng":     "libpng",
		"libjpeg":    "libjpeg",
		"sqlite":     "sqlite",
		"curl":       "curl",
		"git":        "git",
		"cmake":      "cmake",
		"pkg-config": "pkg-config",
		"pkgconfig":  "pkg-config",

		// Build tools
		"make":     "gnumake",
		"autoconf": "autoconf",
		"automake": "automake",
		"libtool":  "libtool",
		"meson":    "meson",
		"ninja":    "ninja",

		// Language specific
		"python3": "python3",
		"nodejs":  "nodejs",
		"npm":     "nodejs", // npm comes with nodejs
		"cargo":   "cargo",
		"rustc":   "rustc",
		"go":      "go",
		"gcc":     "gcc",
		"clang":   "clang",
	}

	for _, dep := range dependencies {
		depName := strings.ToLower(dep.Name)

		// Check common mappings first
		if nixpkg, exists := commonMappings[depName]; exists {
			mappings[dep.Name] = nixpkg
			continue
		}

		// For system libraries, try querying nixpkgs
		if dep.System && dg.mcpClient != nil {
			query := fmt.Sprintf("nixpkgs package %s library", dep.Name)
			result, err := dg.mcpClient.QueryDocumentation(query)
			if err == nil && result != "" {
				// Simple heuristic: if the result contains the dependency name,
				// suggest it as a potential mapping
				if strings.Contains(strings.ToLower(result), depName) {
					mappings[dep.Name] = dep.Name // Suggest the same name
				}
			}
		}
	}

	return mappings, nil
}

// ValidateDerivation performs basic validation on the generated derivation
func (dg *DerivationGenerator) ValidateDerivation(derivation string) []string {
	var issues []string

	// Check for basic structure
	if !strings.HasPrefix(strings.TrimSpace(derivation), "{") {
		issues = append(issues, "Derivation should start with opening brace {")
	}

	if !strings.HasSuffix(strings.TrimSpace(derivation), "}") {
		issues = append(issues, "Derivation should end with closing brace }")
	}

	// Check for required attributes
	requiredAttrs := []string{"pname", "version", "src"}
	for _, attr := range requiredAttrs {
		if !strings.Contains(derivation, attr) {
			issues = append(issues, fmt.Sprintf("Missing required attribute: %s", attr))
		}
	}

	// Check for meta section
	if !strings.Contains(derivation, "meta") {
		issues = append(issues, "Missing meta section (recommended)")
	}

	// Check for balanced braces
	braceCount := 0
	for _, char := range derivation {
		switch char {
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
	}
	if braceCount != 0 {
		issues = append(issues, "Unbalanced braces in derivation")
	}

	return issues
}
