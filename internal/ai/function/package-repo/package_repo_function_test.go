package packagerepo

import (
	"context"
	"strings"
	"testing"
)

func TestNewPackageRepoFunction(t *testing.T) {
	prf := NewPackageRepoFunction()

	if prf == nil {
		t.Fatal("NewPackageRepoFunction returned nil")
	}

	if prf.Name() != "package-repo" {
		t.Errorf("Expected function name 'package-repo', got '%s'", prf.Name())
	}

	if prf.Description() == "" {
		t.Error("Function description should not be empty")
	}
}

func TestPackageRepoFunction_ValidateParameters(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid GitHub URL",
			params: map[string]interface{}{
				"repo_url": "https://github.com/user/repo",
			},
			expectError: false,
		},
		{
			name: "Valid GitLab URL",
			params: map[string]interface{}{
				"repo_url": "https://gitlab.com/user/repo",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"repo_url":      "https://github.com/user/python-project",
				"language":      "python",
				"build_system":  "setuptools",
				"output_format": "flake",
				"detailed":      true,
			},
			expectError: false,
		},
		{
			name: "Missing repo_url parameter",
			params: map[string]interface{}{
				"language": "python",
			},
			expectError: true,
		},
		{
			name: "Empty repo_url parameter",
			params: map[string]interface{}{
				"repo_url": "",
			},
			expectError: true,
		},
		{
			name: "Invalid repo_url",
			params: map[string]interface{}{
				"repo_url": "not-a-valid-url",
			},
			expectError: true,
		},
		{
			name: "Invalid output_format",
			params: map[string]interface{}{
				"repo_url":      "https://github.com/user/repo",
				"output_format": "invalid",
			},
			expectError: true,
		},
		{
			name: "Valid SSH Git URL",
			params: map[string]interface{}{
				"repo_url": "git@github.com:user/repo.git",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := prf.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestPackageRepoFunction_ParseRequest(t *testing.T) {
	prf := NewPackageRepoFunction()

	params := map[string]interface{}{
		"repo_url":      "https://github.com/user/python-project",
		"language":      "python",
		"build_system":  "setuptools",
		"output_format": "flake",
		"dependencies": map[string]interface{}{
			"requests": "2.28.0",
			"flask":    "2.0.0",
		},
		"detailed": true,
	}

	request, err := prf.parseRequest(params)
	if err != nil {
		t.Fatalf("Unexpected error parsing request: %v", err)
	}

	if request.RepoURL != "https://github.com/user/python-project" {
		t.Errorf("Expected repo_url 'https://github.com/user/python-project', got '%s'", request.RepoURL)
	}

	if request.Language != "python" {
		t.Errorf("Expected language 'python', got '%s'", request.Language)
	}

	if request.BuildSystem != "setuptools" {
		t.Errorf("Expected build_system 'setuptools', got '%s'", request.BuildSystem)
	}

	if request.OutputFormat != "flake" {
		t.Errorf("Expected output_format 'flake', got '%s'", request.OutputFormat)
	}

	if !request.Detailed {
		t.Error("Expected Detailed to be true")
	}

	if len(request.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(request.Dependencies))
	}
}

func TestPackageRepoFunction_ParseRepoURL(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name          string
		repoURL       string
		expectedHost  string
		expectedOwner string
		expectedRepo  string
	}{
		{
			name:          "GitHub HTTPS URL",
			repoURL:       "https://github.com/nixos/nixpkgs",
			expectedHost:  "github.com",
			expectedOwner: "nixos",
			expectedRepo:  "nixpkgs",
		},
		{
			name:          "GitHub HTTPS URL with .git",
			repoURL:       "https://github.com/user/repo.git",
			expectedHost:  "github.com",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "GitLab URL",
			repoURL:       "https://gitlab.com/gitlab-org/gitlab",
			expectedHost:  "gitlab.com",
			expectedOwner: "gitlab-org",
			expectedRepo:  "gitlab",
		},
		{
			name:          "BitBucket URL",
			repoURL:       "https://bitbucket.org/atlassian/stash",
			expectedHost:  "bitbucket.org",
			expectedOwner: "atlassian",
			expectedRepo:  "stash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoInfo := prf.parseRepoURL(tt.repoURL)

			if repoInfo.Host != tt.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tt.expectedHost, repoInfo.Host)
			}

			if repoInfo.Owner != tt.expectedOwner {
				t.Errorf("Expected owner '%s', got '%s'", tt.expectedOwner, repoInfo.Owner)
			}

			if repoInfo.Name != tt.expectedRepo {
				t.Errorf("Expected repo name '%s', got '%s'", tt.expectedRepo, repoInfo.Name)
			}
		})
	}
}

func TestPackageRepoFunction_DetectLanguageFromURL(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name             string
		repoURL          string
		expectedLanguage string
	}{
		{
			name:             "Python project",
			repoURL:          "https://github.com/user/python-app",
			expectedLanguage: "python",
		},
		{
			name:             "Rust project",
			repoURL:          "https://github.com/user/rust-cli",
			expectedLanguage: "rust",
		},
		{
			name:             "Go project",
			repoURL:          "https://github.com/user/go-service",
			expectedLanguage: "go",
		},
		{
			name:             "Node.js project",
			repoURL:          "https://github.com/user/node-app",
			expectedLanguage: "node",
		},
		{
			name:             "React project",
			repoURL:          "https://github.com/user/react-frontend",
			expectedLanguage: "node",
		},
		{
			name:             "Unknown project",
			repoURL:          "https://github.com/user/mystery-project",
			expectedLanguage: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			language := prf.detectLanguageFromURL(tt.repoURL)

			if language != tt.expectedLanguage {
				t.Errorf("Expected language '%s', got '%s'", tt.expectedLanguage, language)
			}
		})
	}
}

func TestPackageRepoFunction_DetectBuildSystem(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name                string
		language            string
		expectedBuildSystem string
	}{
		{
			name:                "Python build system",
			language:            "python",
			expectedBuildSystem: "setuptools",
		},
		{
			name:                "Rust build system",
			language:            "rust",
			expectedBuildSystem: "cargo",
		},
		{
			name:                "Go build system",
			language:            "go",
			expectedBuildSystem: "go",
		},
		{
			name:                "Node.js build system",
			language:            "node",
			expectedBuildSystem: "npm",
		},
		{
			name:                "Unknown language",
			language:            "unknown",
			expectedBuildSystem: "stdenv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildSystem := prf.detectBuildSystem(tt.language)

			if buildSystem != tt.expectedBuildSystem {
				t.Errorf("Expected build system '%s', got '%s'", tt.expectedBuildSystem, buildSystem)
			}
		})
	}
}

func TestPackageRepoFunction_GenerateDependencies(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name                 string
		language             string
		buildSystem          string
		expectedDependencies []string
	}{
		{
			name:                 "Python dependencies",
			language:             "python",
			buildSystem:          "setuptools",
			expectedDependencies: []string{"python3", "python3Packages.pip", "python3Packages.setuptools"},
		},
		{
			name:                 "Rust dependencies",
			language:             "rust",
			buildSystem:          "cargo",
			expectedDependencies: []string{"rustc", "cargo"},
		},
		{
			name:                 "Go dependencies",
			language:             "go",
			buildSystem:          "go",
			expectedDependencies: []string{"go"},
		},
		{
			name:                 "Unknown language",
			language:             "unknown",
			buildSystem:          "stdenv",
			expectedDependencies: []string{"stdenv.cc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := prf.generateDependencies(tt.language, tt.buildSystem)

			if len(deps) != len(tt.expectedDependencies) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDependencies), len(deps))
			}

			for i, expected := range tt.expectedDependencies {
				if i >= len(deps) || deps[i] != expected {
					t.Errorf("Expected dependency '%s', got '%s'", expected, deps[i])
				}
			}
		})
	}
}

func TestPackageRepoFunction_GenerateBuildInstructions(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name        string
		language    string
		buildSystem string
		contains    []string
	}{
		{
			name:        "Python build instructions",
			language:    "python",
			buildSystem: "setuptools",
			contains:    []string{"python setup.py build", "python setup.py install"},
		},
		{
			name:        "Rust build instructions",
			language:    "rust",
			buildSystem: "cargo",
			contains:    []string{"cargo build --release", "cargo install"},
		},
		{
			name:        "Go build instructions",
			language:    "go",
			buildSystem: "go",
			contains:    []string{"go build", "mkdir -p $out/bin"},
		},
		{
			name:        "Unknown language instructions",
			language:    "unknown",
			buildSystem: "stdenv",
			contains:    []string{"make", "make install"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instructions := prf.generateBuildInstructions(tt.language, tt.buildSystem)

			if instructions == "" {
				t.Error("Build instructions should not be empty")
			}

			for _, required := range tt.contains {
				if !strings.Contains(instructions, required) {
					t.Errorf("Build instructions should contain '%s': %s", required, instructions)
				}
			}
		})
	}
}

func TestPackageRepoFunction_GenerateDerivation(t *testing.T) {
	prf := NewPackageRepoFunction()

	request := &PackageRepoRequest{
		RepoURL:      "https://github.com/user/test-repo",
		Language:     "python",
		OutputFormat: "derivation",
		Detailed:     false,
	}

	analysis := &RepositoryAnalysis{
		Language:          "python",
		BuildSystem:       "setuptools",
		Dependencies:      []string{"python3", "python3Packages.setuptools"},
		BuildInstructions: "buildPhase = ''python setup.py build'';",
		Metadata:          map[string]string{},
	}

	derivation := prf.generateDerivation(request, analysis)

	if derivation == "" {
		t.Error("Derivation should not be empty")
	}

	// Check for key components
	expectedComponents := []string{
		"stdenv.mkDerivation",
		"pname = \"test-repo\"",
		"fetchFromGitHub",
		"owner = \"user\"",
		"repo = \"test-repo\"",
		"nativeBuildInputs",
		"buildInputs",
		"meta = with lib",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(derivation, component) {
			t.Errorf("Derivation should contain '%s': %s", component, derivation)
		}
	}
}

func TestPackageRepoFunction_GenerateFlake(t *testing.T) {
	prf := NewPackageRepoFunction()

	request := &PackageRepoRequest{
		RepoURL:      "https://github.com/user/test-repo",
		Language:     "rust",
		OutputFormat: "flake",
		Detailed:     false,
	}

	analysis := &RepositoryAnalysis{
		Language:          "rust",
		BuildSystem:       "cargo",
		Dependencies:      []string{"rustc", "cargo"},
		BuildInstructions: "buildPhase = ''cargo build --release'';",
		Metadata:          map[string]string{},
	}

	flake := prf.generateFlake(request, analysis)

	if flake == "" {
		t.Error("Flake should not be empty")
	}

	// Check for key components
	expectedComponents := []string{
		"description = \"test-repo",
		"inputs = {",
		"nixpkgs.url",
		"flake-utils.url",
		"outputs = {",
		"packages.default",
		"devShells.default",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(flake, component) {
			t.Errorf("Flake should contain '%s': %s", component, flake)
		}
	}
}

func TestPackageRepoFunction_IsValidRepoURL(t *testing.T) {
	prf := NewPackageRepoFunction()

	tests := []struct {
		name    string
		url     string
		isValid bool
	}{
		{
			name:    "Valid GitHub HTTPS URL",
			url:     "https://github.com/user/repo",
			isValid: true,
		},
		{
			name:    "Valid GitLab HTTPS URL",
			url:     "https://gitlab.com/user/repo",
			isValid: true,
		},
		{
			name:    "Valid GitHub SSH URL",
			url:     "git@github.com:user/repo.git",
			isValid: true,
		},
		{
			name:    "Invalid URL",
			url:     "not-a-url",
			isValid: false,
		},
		{
			name:    "Invalid scheme",
			url:     "ftp://github.com/user/repo",
			isValid: false,
		},
		{
			name:    "Empty URL",
			url:     "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := prf.isValidRepoURL(tt.url)

			if isValid != tt.isValid {
				t.Errorf("Expected URL validity %v, got %v for URL: %s", tt.isValid, isValid, tt.url)
			}
		})
	}
}

func TestPackageRepoFunction_Execute(t *testing.T) {
	prf := NewPackageRepoFunction()
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Basic execution",
			params: map[string]interface{}{
				"repo_url": "https://github.com/user/test-repo",
			},
			expectError: false,
		},
		{
			name: "With language hint",
			params: map[string]interface{}{
				"repo_url": "https://github.com/user/python-project",
				"language": "python",
			},
			expectError: false,
		},
		{
			name: "Generate flake",
			params: map[string]interface{}{
				"repo_url":      "https://github.com/user/rust-project",
				"output_format": "flake",
				"detailed":      true,
			},
			expectError: false,
		},
		{
			name: "Missing repo_url",
			params: map[string]interface{}{
				"language": "python",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prf.Execute(ctx, tt.params, nil)

			if tt.expectError {
				if err != nil || (result != nil && result.Success) {
					t.Error("Expected execution error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected execution error: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if !result.Success {
				t.Error("Result should be successful")
			}

			if result.Data == nil {
				t.Error("Result data should not be nil")
			}

			// Verify the response structure
			response, ok := result.Data.(*PackageRepoResponse)
			if !ok {
				t.Errorf("Expected *PackageRepoResponse, got %T", result.Data)
				return
			}

			if response.RepoURL == "" {
				t.Error("Response repo_url should not be empty")
			}

			if response.DetectedLanguage == "" {
				t.Error("Response detected_language should not be empty")
			}

			if response.NixDerivation == "" {
				t.Error("Response nix_derivation should not be empty")
			}

			if len(response.Dependencies) == 0 {
				t.Error("Response should include dependencies")
			}
		})
	}
}
