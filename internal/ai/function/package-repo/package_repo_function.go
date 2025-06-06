package packagerepo

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// PackageRepoFunction analyzes Git repositories and generates Nix derivations
type PackageRepoFunction struct {
	*functionbase.BaseFunction
	logger logger.Logger
}

// PackageRepoRequest represents the input parameters for package-repo analysis
type PackageRepoRequest struct {
	RepoURL      string            `json:"repo_url"`
	Language     string            `json:"language,omitempty"`
	BuildSystem  string            `json:"build_system,omitempty"`
	OutputFormat string            `json:"output_format,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Detailed     bool              `json:"detailed,omitempty"`
}

// PackageRepoResponse represents the output of the package-repo function
type PackageRepoResponse struct {
	RepoURL           string            `json:"repo_url"`
	DetectedLanguage  string            `json:"detected_language"`
	DetectedBuild     string            `json:"detected_build_system"`
	NixDerivation     string            `json:"nix_derivation"`
	Dependencies      []string          `json:"dependencies"`
	BuildInstructions string            `json:"build_instructions"`
	Metadata          map[string]string `json:"metadata"`
	Warnings          []string          `json:"warnings,omitempty"`
}

// NewPackageRepoFunction creates a new package-repo function
func NewPackageRepoFunction() *PackageRepoFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("repo_url", "Git repository URL to analyze (e.g., 'https://github.com/user/repo')", true),
		functionbase.StringParam("language", "Programming language hint (auto-detected if not specified)", false),
		functionbase.StringParam("build_system", "Build system hint (auto-detected if not specified)", false),
		functionbase.StringParam("output_format", "Output format: 'derivation' (default), 'flake', or 'overlay'", false),
		functionbase.ObjectParam("dependencies", "Override detected dependencies (key-value pairs)", false),
		functionbase.BoolParam("detailed", "Generate detailed derivation with comments and examples", false, false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"package-repo",
		"Analyze Git repositories and generate Nix derivations for packaging",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Analyze a Python project repository",
			Parameters: map[string]interface{}{
				"repo_url":      "https://github.com/user/python-project",
				"language":      "python",
				"output_format": "derivation",
				"detailed":      true,
			},
			Expected: "Generated Nix derivation for the Python project with detailed comments",
		},
		{
			Description: "Generate flake for a Rust project",
			Parameters: map[string]interface{}{
				"repo_url":      "https://github.com/user/rust-project",
				"output_format": "flake",
			},
			Expected: "Generated flake.nix for the Rust project with cargo build system",
		},
	}
	baseFunc.SetSchema(schema)

	return &PackageRepoFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
	}
}

// ValidateParameters validates the function parameters with custom checks
func (prf *PackageRepoFunction) ValidateParameters(params map[string]interface{}) error {
	// First run base validation
	if err := prf.BaseFunction.ValidateParameters(params); err != nil {
		return err
	}

	// Custom validation for repo_url parameter
	if repoURL, ok := params["repo_url"].(string); ok {
		if strings.TrimSpace(repoURL) == "" {
			return fmt.Errorf("repo_url parameter cannot be empty")
		}
		if !prf.isValidRepoURL(repoURL) {
			return fmt.Errorf("repo_url must be a valid Git repository URL")
		}
	}

	// Validate output_format if provided
	if outputFormat, ok := params["output_format"].(string); ok {
		validFormats := []string{"derivation", "flake", "overlay"}
		if !prf.contains(validFormats, outputFormat) {
			return fmt.Errorf("output_format must be one of: %s", strings.Join(validFormats, ", "))
		}
	}

	return nil
}

// Execute runs the package-repo function
func (prf *PackageRepoFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	prf.logger.Debug("Starting package-repo function execution")

	// Parse parameters into structured request
	request, err := prf.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate that we have a repo URL
	if request.RepoURL == "" {
		return functionbase.CreateErrorResult(
			fmt.Errorf("repo_url parameter is required and cannot be empty"),
			"Missing required parameter",
		), nil
	}

	// Analyze the repository
	analysis, err := prf.analyzeRepository(request)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to analyze repository"), nil
	}

	// Generate the Nix derivation
	derivation, err := prf.generateNixDerivation(request, analysis)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to generate Nix derivation"), nil
	}

	// Build the response
	response := &PackageRepoResponse{
		RepoURL:           request.RepoURL,
		DetectedLanguage:  analysis.Language,
		DetectedBuild:     analysis.BuildSystem,
		NixDerivation:     derivation,
		Dependencies:      analysis.Dependencies,
		BuildInstructions: analysis.BuildInstructions,
		Metadata:          analysis.Metadata,
		Warnings:          analysis.Warnings,
	}

	prf.logger.Debug("Package-repo function execution completed successfully")

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts raw parameters to structured PackageRepoRequest
func (prf *PackageRepoFunction) parseRequest(params map[string]interface{}) (*PackageRepoRequest, error) {
	request := &PackageRepoRequest{}

	// Extract repo_url (required)
	if repoURL, ok := params["repo_url"].(string); ok {
		request.RepoURL = strings.TrimSpace(repoURL)
	}

	// Extract language (optional)
	if language, ok := params["language"].(string); ok {
		request.Language = strings.TrimSpace(language)
	}

	// Extract build_system (optional)
	if buildSystem, ok := params["build_system"].(string); ok {
		request.BuildSystem = strings.TrimSpace(buildSystem)
	}

	// Extract output_format (optional, default to "derivation")
	if outputFormat, ok := params["output_format"].(string); ok {
		request.OutputFormat = strings.TrimSpace(outputFormat)
	} else {
		request.OutputFormat = "derivation"
	}

	// Extract dependencies (optional)
	if deps, ok := params["dependencies"].(map[string]interface{}); ok {
		request.Dependencies = make(map[string]string)
		for k, v := range deps {
			if strVal, ok := v.(string); ok {
				request.Dependencies[k] = strVal
			}
		}
	}

	// Extract detailed (optional, default false)
	if detailed, ok := params["detailed"].(bool); ok {
		request.Detailed = detailed
	}

	return request, nil
}

// RepositoryAnalysis represents the results of repository analysis
type RepositoryAnalysis struct {
	Language          string
	BuildSystem       string
	Dependencies      []string
	BuildInstructions string
	Metadata          map[string]string
	Warnings          []string
}

// analyzeRepository performs static analysis of the repository
func (prf *PackageRepoFunction) analyzeRepository(request *PackageRepoRequest) (*RepositoryAnalysis, error) {
	analysis := &RepositoryAnalysis{
		Metadata: make(map[string]string),
		Warnings: []string{},
	}

	// Extract repository information from URL
	repoInfo := prf.parseRepoURL(request.RepoURL)
	analysis.Metadata["owner"] = repoInfo.Owner
	analysis.Metadata["repo"] = repoInfo.Name
	analysis.Metadata["host"] = repoInfo.Host

	// Detect language (use hint if provided, otherwise detect)
	if request.Language != "" {
		analysis.Language = request.Language
	} else {
		analysis.Language = prf.detectLanguageFromURL(request.RepoURL)
	}

	// Detect build system (use hint if provided, otherwise detect)
	if request.BuildSystem != "" {
		analysis.BuildSystem = request.BuildSystem
	} else {
		analysis.BuildSystem = prf.detectBuildSystem(analysis.Language)
	}

	// Generate dependencies based on language and build system
	analysis.Dependencies = prf.generateDependencies(analysis.Language, analysis.BuildSystem)

	// Generate build instructions
	analysis.BuildInstructions = prf.generateBuildInstructions(analysis.Language, analysis.BuildSystem)

	// Add warnings for common issues
	if analysis.Language == "unknown" {
		analysis.Warnings = append(analysis.Warnings, "Could not detect programming language - defaulting to generic derivation")
	}

	return analysis, nil
}

// RepoInfo contains parsed repository information
type RepoInfo struct {
	Host  string
	Owner string
	Name  string
}

// parseRepoURL extracts owner and repository name from URL
func (prf *PackageRepoFunction) parseRepoURL(repoURL string) RepoInfo {
	info := RepoInfo{}

	// Parse URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return info
	}

	info.Host = parsedURL.Host

	// Extract owner and repo from path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) >= 2 {
		info.Owner = pathParts[0]
		info.Name = strings.TrimSuffix(pathParts[1], ".git")
	}

	return info
}

// detectLanguageFromURL attempts to detect language from repository URL patterns
func (prf *PackageRepoFunction) detectLanguageFromURL(repoURL string) string {
	// Common language indicators in repository names
	repoName := strings.ToLower(filepath.Base(repoURL))
	repoName = strings.TrimSuffix(repoName, ".git")

	languagePatterns := map[string][]string{
		"python":  {"python", "py", "flask", "django", "fastapi"},
		"rust":    {"rust", "rs", "cargo"},
		"go":      {"go", "golang"},
		"node":    {"node", "js", "npm", "yarn", "react", "vue", "angular"},
		"java":    {"java", "spring", "maven", "gradle"},
		"cpp":     {"cpp", "c++", "cmake"},
		"c":       {"c", "make"},
		"haskell": {"haskell", "hs", "cabal", "stack"},
	}

	for lang, patterns := range languagePatterns {
		for _, pattern := range patterns {
			if strings.Contains(repoName, pattern) {
				return lang
			}
		}
	}

	return "unknown"
}

// detectBuildSystem determines the appropriate build system for a language
func (prf *PackageRepoFunction) detectBuildSystem(language string) string {
	buildSystems := map[string]string{
		"python":  "setuptools",
		"rust":    "cargo",
		"go":      "go",
		"node":    "npm",
		"java":    "maven",
		"cpp":     "cmake",
		"c":       "make",
		"haskell": "cabal",
	}

	if buildSystem, exists := buildSystems[language]; exists {
		return buildSystem
	}

	return "stdenv"
}

// generateDependencies creates a list of common dependencies for the language
func (prf *PackageRepoFunction) generateDependencies(language, buildSystem string) []string {
	dependencyMap := map[string][]string{
		"python":  {"python3", "python3Packages.pip", "python3Packages.setuptools"},
		"rust":    {"rustc", "cargo"},
		"go":      {"go"},
		"node":    {"nodejs", "npm"},
		"java":    {"jdk", "maven"},
		"cpp":     {"gcc", "cmake", "pkg-config"},
		"c":       {"gcc", "make", "pkg-config"},
		"haskell": {"ghc", "cabal-install"},
	}

	if deps, exists := dependencyMap[language]; exists {
		return deps
	}

	return []string{"stdenv.cc"}
}

// generateBuildInstructions creates build instructions for the derivation
func (prf *PackageRepoFunction) generateBuildInstructions(language, buildSystem string) string {
	instructions := map[string]string{
		"python": `
    buildPhase = ''
      python setup.py build
    '';
    
    installPhase = ''
      python setup.py install --prefix=$out
    '';`,
		"rust": `
    buildPhase = ''
      cargo build --release
    '';
    
    installPhase = ''
      cargo install --path . --root $out
    '';`,
		"go": `
    buildPhase = ''
      go build -o $pname ./...
    '';
    
    installPhase = ''
      mkdir -p $out/bin
      mv $pname $out/bin/
    '';`,
		"node": `
    buildPhase = ''
      npm ci
      npm run build
    '';
    
    installPhase = ''
      mkdir -p $out/lib/node_modules/$pname
      cp -r . $out/lib/node_modules/$pname
      mkdir -p $out/bin
      ln -s $out/lib/node_modules/$pname/bin/* $out/bin/
    '';`,
	}

	if instruction, exists := instructions[language]; exists {
		return instruction
	}

	return `
    buildPhase = ''
      make
    '';
    
    installPhase = ''
      make install PREFIX=$out
    '';`
}

// generateNixDerivation creates the Nix derivation based on analysis
func (prf *PackageRepoFunction) generateNixDerivation(request *PackageRepoRequest, analysis *RepositoryAnalysis) (string, error) {
	switch request.OutputFormat {
	case "flake":
		return prf.generateFlake(request, analysis), nil
	case "overlay":
		return prf.generateOverlay(request, analysis), nil
	default:
		return prf.generateDerivation(request, analysis), nil
	}
}

// generateDerivation creates a standard Nix derivation
func (prf *PackageRepoFunction) generateDerivation(request *PackageRepoRequest, analysis *RepositoryAnalysis) string {
	repoInfo := prf.parseRepoURL(request.RepoURL)

	template := `{ lib, stdenv%s }:

stdenv.mkDerivation rec {
  pname = "%s";
  version = "unstable";

  src = fetchFromGitHub {
    owner = "%s";
    repo = "%s";
    rev = "main";  # or specific commit/tag
    sha256 = "0000000000000000000000000000000000000000000000000000";  # Use nix-prefetch-url or similar
  };

  nativeBuildInputs = [ %s ];
  buildInputs = [ %s ];
%s
  meta = with lib; {
    description = "Description for %s";
    homepage = "%s";
    license = licenses.unfree;  # Update as appropriate
    maintainers = with maintainers; [ ];
    platforms = platforms.all;
  };
}`

	// Generate dependency strings
	var extraDeps []string
	var nativeBuildInputs []string
	var buildInputs []string

	for _, dep := range analysis.Dependencies {
		if prf.isNativeBuildInput(dep) {
			nativeBuildInputs = append(nativeBuildInputs, dep)
		} else {
			buildInputs = append(buildInputs, dep)
		}

		if dep != "stdenv.cc" && dep != "lib" {
			extraDeps = append(extraDeps, dep)
		}
	}

	extraDepsStr := ""
	if len(extraDeps) > 0 {
		extraDepsStr = ", " + strings.Join(extraDeps, ", ")
	}

	nativeBuildStr := strings.Join(nativeBuildInputs, " ")
	buildInputStr := strings.Join(buildInputs, " ")

	buildPhase := analysis.BuildInstructions
	if request.Detailed {
		buildPhase = "\n  # Build and install instructions" + buildPhase + "\n"
	}

	return fmt.Sprintf(template,
		extraDepsStr,
		repoInfo.Name,
		repoInfo.Owner,
		repoInfo.Name,
		nativeBuildStr,
		buildInputStr,
		buildPhase,
		repoInfo.Name,
		request.RepoURL,
	)
}

// generateFlake creates a flake.nix for the repository
func (prf *PackageRepoFunction) generateFlake(request *PackageRepoRequest, analysis *RepositoryAnalysis) string {
	repoInfo := prf.parseRepoURL(request.RepoURL)

	template := `{
  description = "%s - Generated by nixai package-repo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        %s = pkgs.stdenv.mkDerivation rec {
          pname = "%s";
          version = "unstable";

          src = ./.;

          nativeBuildInputs = with pkgs; [ %s ];
          buildInputs = with pkgs; [ %s ];
%s
          meta = with pkgs.lib; {
            description = "Description for %s";
            homepage = "%s";
            license = licenses.unfree;  # Update as appropriate
            maintainers = with maintainers; [ ];
            platforms = platforms.all;
          };
        };
      in
      {
        packages.default = %s;
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ %s ];
        };
      });
}`

	// Generate dependency strings
	var nativeBuildInputs []string
	var buildInputs []string

	for _, dep := range analysis.Dependencies {
		if prf.isNativeBuildInput(dep) {
			nativeBuildInputs = append(nativeBuildInputs, dep)
		} else {
			buildInputs = append(buildInputs, dep)
		}
	}

	nativeBuildStr := strings.Join(nativeBuildInputs, " ")
	buildInputStr := strings.Join(buildInputs, " ")
	devShellInputs := strings.Join(analysis.Dependencies, " ")

	return fmt.Sprintf(template,
		repoInfo.Name,
		repoInfo.Name,
		repoInfo.Name,
		nativeBuildStr,
		buildInputStr,
		analysis.BuildInstructions,
		repoInfo.Name,
		request.RepoURL,
		repoInfo.Name,
		devShellInputs,
	)
}

// generateOverlay creates a Nix overlay for the repository
func (prf *PackageRepoFunction) generateOverlay(request *PackageRepoRequest, analysis *RepositoryAnalysis) string {
	repoInfo := prf.parseRepoURL(request.RepoURL)

	template := `final: prev: {
  %s = prev.stdenv.mkDerivation rec {
    pname = "%s";
    version = "unstable";

    src = prev.fetchFromGitHub {
      owner = "%s";
      repo = "%s";
      rev = "main";  # or specific commit/tag
      sha256 = "0000000000000000000000000000000000000000000000000000";
    };

    nativeBuildInputs = with prev; [ %s ];
    buildInputs = with prev; [ %s ];
%s
    meta = with prev.lib; {
      description = "Description for %s";
      homepage = "%s";
      license = licenses.unfree;  # Update as appropriate
      maintainers = with maintainers; [ ];
      platforms = platforms.all;
    };
  };
}`

	// Generate dependency strings
	var nativeBuildInputs []string
	var buildInputs []string

	for _, dep := range analysis.Dependencies {
		if prf.isNativeBuildInput(dep) {
			nativeBuildInputs = append(nativeBuildInputs, dep)
		} else {
			buildInputs = append(buildInputs, dep)
		}
	}

	nativeBuildStr := strings.Join(nativeBuildInputs, " ")
	buildInputStr := strings.Join(buildInputs, " ")

	return fmt.Sprintf(template,
		repoInfo.Name,
		repoInfo.Name,
		repoInfo.Owner,
		repoInfo.Name,
		nativeBuildStr,
		buildInputStr,
		analysis.BuildInstructions,
		repoInfo.Name,
		request.RepoURL,
	)
}

// isValidRepoURL validates if the provided URL is a valid Git repository URL
func (prf *PackageRepoFunction) isValidRepoURL(repoURL string) bool {
	// Regex pattern for common Git repository URLs
	patterns := []string{
		`^https://github\.com/[\w\-\.]+/[\w\-\.]+/?$`,
		`^https://gitlab\.com/[\w\-\.]+/[\w\-\.]+/?$`,
		`^https://bitbucket\.org/[\w\-\.]+/[\w\-\.]+/?$`,
		`^git@github\.com:[\w\-\.]+/[\w\-\.]+\.git$`,
		`^git@gitlab\.com:[\w\-\.]+/[\w\-\.]+\.git$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, repoURL); matched {
			return true
		}
	}

	return false
}

// isNativeBuildInput determines if a dependency should be a nativeBuildInput
func (prf *PackageRepoFunction) isNativeBuildInput(dep string) bool {
	nativeDeps := []string{
		"cmake", "make", "pkg-config", "cargo", "go", "npm", "maven",
		"python3Packages.setuptools", "python3Packages.pip", "cabal-install",
	}

	return prf.contains(nativeDeps, dep)
}

// contains checks if a slice contains a specific string
func (prf *PackageRepoFunction) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
