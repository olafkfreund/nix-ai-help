package packaging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/packaging/detection"
	"nix-ai-help/pkg/logger"
)

// BuildSystem represents different build systems
type BuildSystem string

const (
	BuildSystemCMake     BuildSystem = "cmake"
	BuildSystemMeson     BuildSystem = "meson"
	BuildSystemAutotools BuildSystem = "autotools"
	BuildSystemMake      BuildSystem = "make"
	BuildSystemCargoRust BuildSystem = "cargo"
	BuildSystemNpm       BuildSystem = "npm"
	BuildSystemYarn      BuildSystem = "yarn"
	BuildSystemPython    BuildSystem = "python"
	BuildSystemGo        BuildSystem = "go"
	BuildSystemGradle    BuildSystem = "gradle"
	BuildSystemMaven     BuildSystem = "maven"
	BuildSystemUnknown   BuildSystem = "unknown"
)

// Dependency represents a project dependency
type Dependency struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // "build", "runtime", "dev"
	Version string `json:"version,omitempty"`
	System  bool   `json:"system"` // true for system libraries
}

// RepoAnalysis contains the analysis results of a repository
type RepoAnalysis struct {
	RepoURL      string       `json:"repo_url"`
	LocalPath    string       `json:"local_path"`
	ProjectName  string       `json:"project_name"`
	BuildSystem  BuildSystem  `json:"build_system"`
	Language     string       `json:"language"`
	Dependencies []Dependency `json:"dependencies"`
	BuildFiles   []string     `json:"build_files"`
	HasTests     bool         `json:"has_tests"`
	License      string       `json:"license,omitempty"`
	Description  string       `json:"description,omitempty"`
}

// RepositoryAnalyzer analyzes Git repositories for packaging
type RepositoryAnalyzer struct {
	detector *detection.EnhancedDetector
	logger   *logger.Logger
}

// NewRepositoryAnalyzer creates a new repository analyzer
func NewRepositoryAnalyzer(log *logger.Logger) *RepositoryAnalyzer {
	return &RepositoryAnalyzer{
		detector: detection.NewEnhancedDetector(log),
		logger:   log,
	}
}

// AnalyzeRepository analyzes a repository for packaging information
func (ra *RepositoryAnalyzer) AnalyzeRepository(repoPath string) (*RepoAnalysis, error) {
	analysis := &RepoAnalysis{
		LocalPath:    repoPath,
		Dependencies: []Dependency{},
		BuildFiles:   []string{},
	}

	// Extract project name from path
	analysis.ProjectName = filepath.Base(repoPath)

	// Detect build system and collect build files
	buildSystem, buildFiles, err := ra.detectBuildSystem(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect build system: %w", err)
	}
	analysis.BuildSystem = buildSystem
	analysis.BuildFiles = buildFiles

	// Detect primary language
	language, err := ra.detectLanguage(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect language: %w", err)
	}
	analysis.Language = language

	// Analyze dependencies based on build system
	dependencies, err := ra.analyzeDependencies(repoPath, buildSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}
	analysis.Dependencies = dependencies

	// Check for tests
	analysis.HasTests = ra.hasTests(repoPath)

	// Try to find license
	analysis.License = ra.findLicense(repoPath)

	// Try to find description
	analysis.Description = ra.findDescription(repoPath)

	return analysis, nil
}

// detectBuildSystem detects the build system used by the project
func (ra *RepositoryAnalyzer) detectBuildSystem(repoPath string) (BuildSystem, []string, error) {
	var buildFiles []string

	// Check for various build system files
	checks := map[string]BuildSystem{
		"CMakeLists.txt": BuildSystemCMake,
		"meson.build":    BuildSystemMeson,
		"configure.ac":   BuildSystemAutotools,
		"configure.in":   BuildSystemAutotools,
		"Makefile":       BuildSystemMake,
		"makefile":       BuildSystemMake,
		"Cargo.toml":     BuildSystemCargoRust,
		"package.json":   BuildSystemNpm,
		"yarn.lock":      BuildSystemYarn,
		"setup.py":       BuildSystemPython,
		"pyproject.toml": BuildSystemPython,
		"go.mod":         BuildSystemGo,
		"build.gradle":   BuildSystemGradle,
		"pom.xml":        BuildSystemMaven,
	}

	var detectedSystem = BuildSystemUnknown
	priority := 0

	// Priority order for build systems (higher number = higher priority)
	priorities := map[BuildSystem]int{
		BuildSystemCMake:     8,
		BuildSystemMeson:     9,
		BuildSystemAutotools: 7,
		BuildSystemCargoRust: 10,
		BuildSystemGo:        10,
		BuildSystemNpm:       6,
		BuildSystemYarn:      5,
		BuildSystemPython:    4,
		BuildSystemGradle:    3,
		BuildSystemMaven:     2,
		BuildSystemMake:      1,
	}

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "target" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		fileName := info.Name()
		if buildSystem, exists := checks[fileName]; exists {
			buildFiles = append(buildFiles, path)
			if systemPriority := priorities[buildSystem]; systemPriority > priority {
				detectedSystem = buildSystem
				priority = systemPriority
			}
		}

		return nil
	})

	if err != nil {
		return BuildSystemUnknown, nil, err
	}

	return detectedSystem, buildFiles, nil
}

// detectLanguage attempts to detect the primary programming language
func (ra *RepositoryAnalyzer) detectLanguage(repoPath string) (string, error) {
	// Use enhanced detection system
	ctx := context.Background()
	opts := detection.DefaultAnalysisOptions()

	results, err := ra.detector.DetectLanguages(ctx, repoPath, opts)
	if err != nil {
		ra.logger.Warn(fmt.Sprintf("Enhanced language detection failed, falling back to basic detection: %v", err))
		return ra.detectLanguageBasic(repoPath)
	}

	// Return the highest confidence language
	if len(results) > 0 {
		primary := results[0] // Results are sorted by confidence
		ra.logger.Debug(fmt.Sprintf("Enhanced language detection completed: %s (confidence: %.2f, evidence: %d)",
			primary.Language, primary.Confidence, len(primary.Evidence)))
		return primary.Language, nil
	}

	// Fallback to basic detection if no languages detected
	return ra.detectLanguageBasic(repoPath)
}

// detectLanguageBasic provides basic language detection as fallback
func (ra *RepositoryAnalyzer) detectLanguageBasic(repoPath string) (string, error) {
	languageCount := make(map[string]int)

	extensions := map[string]string{
		".c":    "C",
		".cpp":  "C++",
		".cxx":  "C++",
		".cc":   "C++",
		".h":    "C/C++",
		".hpp":  "C++",
		".rs":   "Rust",
		".go":   "Go",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".py":   "Python",
		".java": "Java",
		".kt":   "Kotlin",
		".sh":   "Shell",
		".rb":   "Ruby",
		".php":  "PHP",
		".cs":   "C#",
		".fs":   "F#",
		".ml":   "OCaml",
		".hs":   "Haskell",
		".nim":  "Nim",
		".zig":  "Zig",
	}

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "target" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if lang, exists := extensions[ext]; exists {
			languageCount[lang]++
		}

		return nil
	})

	if err != nil {
		return "Unknown", err
	}

	// Find the most common language
	maxCount := 0
	primaryLanguage := "Unknown"
	for lang, count := range languageCount {
		if count > maxCount {
			maxCount = count
			primaryLanguage = lang
		}
	}

	return primaryLanguage, nil
}

// analyzeDependencies analyzes project dependencies based on build system
func (ra *RepositoryAnalyzer) analyzeDependencies(repoPath string, buildSystem BuildSystem) ([]Dependency, error) {
	var dependencies []Dependency

	switch buildSystem {
	case BuildSystemNpm, BuildSystemYarn:
		deps, err := ra.analyzeNpmDependencies(repoPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, deps...)

	case BuildSystemCargoRust:
		deps, err := ra.analyzeCargoDependencies(repoPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, deps...)

	case BuildSystemGo:
		deps, err := ra.analyzeGoDependencies(repoPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, deps...)

	case BuildSystemPython:
		deps, err := ra.analyzePythonDependencies(repoPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, deps...)

	case BuildSystemCMake:
		deps, err := ra.analyzeCMakeDependencies(repoPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, deps...)

	case BuildSystemMeson:
		deps, err := ra.analyzeMesonDependencies(repoPath)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, deps...)
	}

	return dependencies, nil
}

// hasTests checks if the project has test files
func (ra *RepositoryAnalyzer) hasTests(repoPath string) bool {
	testIndicators := []string{
		"test", "tests", "spec", "specs", "__tests__",
		"*_test.go", "*_test.py", "test_*.py", "*.test.js",
		"*.spec.js", "*.test.ts", "*.spec.ts",
	}

	hasTests := false
	_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || hasTests {
			return err
		}

		name := strings.ToLower(info.Name())
		for _, indicator := range testIndicators {
			if strings.Contains(name, indicator) {
				hasTests = true
				return filepath.SkipDir
			}
		}
		return nil
	})

	return hasTests
}

// findLicense attempts to find the project license
func (ra *RepositoryAnalyzer) findLicense(repoPath string) string {
	licenseFiles := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "COPYING", "COPYRIGHT"}

	for _, filename := range licenseFiles {
		licensePath := filepath.Join(repoPath, filename)
		if _, err := os.Stat(licensePath); err == nil {
			content, err := os.ReadFile(licensePath)
			if err == nil {
				// Simple license detection
				contentStr := strings.ToUpper(string(content))
				if strings.Contains(contentStr, "MIT") {
					return "MIT"
				} else if strings.Contains(contentStr, "APACHE") {
					return "Apache-2.0"
				} else if strings.Contains(contentStr, "GPL") {
					if strings.Contains(contentStr, "VERSION 3") {
						return "GPL-3.0"
					} else if strings.Contains(contentStr, "VERSION 2") {
						return "GPL-2.0"
					}
					return "GPL"
				} else if strings.Contains(contentStr, "BSD") {
					return "BSD"
				}
			}
		}
	}

	return ""
}

// findDescription attempts to find project description
func (ra *RepositoryAnalyzer) findDescription(repoPath string) string {
	readmeFiles := []string{"README.md", "README.txt", "README", "README.rst"}

	for _, filename := range readmeFiles {
		readmePath := filepath.Join(repoPath, filename)
		if content, err := os.ReadFile(readmePath); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines[:min(10, len(lines))] {
				line = strings.TrimSpace(line)
				if len(line) > 20 && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "!") {
					return line
				}
			}
		}
	}

	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
