package detection

import (
	"regexp"
)

// LanguagePatterns defines detection patterns for various programming languages
var LanguagePatterns = map[string][]DetectionRule{
	"javascript": {
		{
			Name:        "JavaScript files",
			FilePattern: "*.js",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "JavaScript modules",
			FilePattern: "*.mjs",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "ES6 modules",
			FilePattern: "*.es6",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Node.js shebang",
			Content:    regexp.MustCompile(`^#!/usr/bin/env node`),
			Confidence: 0.8,
			Priority:   2,
		},
		{
			Name:        "Package.json",
			FilePattern: "package.json",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:       "CommonJS require",
			Content:    regexp.MustCompile(`require\s*\(['"]`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "ES6 import",
			Content:    regexp.MustCompile(`import\s+.*\s+from\s+['"]`),
			Confidence: 0.8,
			Priority:   2,
		},
	},
	"typescript": {
		{
			Name:        "TypeScript files",
			FilePattern: "*.ts",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "TypeScript JSX",
			FilePattern: "*.tsx",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "TypeScript config",
			FilePattern: "tsconfig.json",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "TypeScript declarations",
			FilePattern: "*.d.ts",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Type annotations",
			Content:    regexp.MustCompile(`:\s*(string|number|boolean|object)\s*[=;]`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "Interface declarations",
			Content:    regexp.MustCompile(`interface\s+\w+\s*{`),
			Confidence: 0.8,
			Priority:   2,
		},
	},
	"python": {
		{
			Name:        "Python files",
			FilePattern: "*.py",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Python wheels",
			FilePattern: "*.pyw",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Python shebang",
			Content:    regexp.MustCompile(`^#!/usr/bin/env python`),
			Confidence: 0.8,
			Priority:   2,
		},
		{
			Name:       "Python3 shebang",
			Content:    regexp.MustCompile(`^#!/usr/bin/env python3`),
			Confidence: 0.85,
			Priority:   2,
		},
		{
			Name:        "Requirements file",
			FilePattern: "requirements.txt",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "Setup file",
			FilePattern: "setup.py",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Pipfile",
			FilePattern: "Pipfile",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "Pyproject.toml",
			FilePattern: "pyproject.toml",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Python imports",
			Content:    regexp.MustCompile(`^import\s+\w+|^from\s+\w+\s+import`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "Python def",
			Content:    regexp.MustCompile(`^def\s+\w+\s*\(`),
			Confidence: 0.6,
			Priority:   3,
		},
	},
	"rust": {
		{
			Name:        "Rust files",
			FilePattern: "*.rs",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Cargo.toml",
			FilePattern: "Cargo.toml",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Cargo.lock",
			FilePattern: "Cargo.lock",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Rust macros",
			Content:    regexp.MustCompile(`#\[\w+\]`),
			Confidence: 0.8,
			Priority:   2,
		},
		{
			Name:       "Rust use statements",
			Content:    regexp.MustCompile(`^use\s+\w+`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "Rust fn",
			Content:    regexp.MustCompile(`fn\s+\w+\s*\(`),
			Confidence: 0.6,
			Priority:   3,
		},
	},
	"go": {
		{
			Name:        "Go files",
			FilePattern: "*.go",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Go module",
			FilePattern: "go.mod",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Go sum",
			FilePattern: "go.sum",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Go package",
			Content:    regexp.MustCompile(`^package\s+\w+`),
			Confidence: 0.8,
			Priority:   2,
		},
		{
			Name:       "Go import",
			Content:    regexp.MustCompile(`import\s+\(`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "Go func",
			Content:    regexp.MustCompile(`func\s+\w+\s*\(`),
			Confidence: 0.6,
			Priority:   3,
		},
	},
	"java": {
		{
			Name:        "Java files",
			FilePattern: "*.java",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Java class files",
			FilePattern: "*.class",
			Confidence:  0.8,
			Priority:    2,
		},
		{
			Name:        "Maven POM",
			FilePattern: "pom.xml",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "Gradle build",
			FilePattern: "build.gradle",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "Gradle Kotlin build",
			FilePattern: "build.gradle.kts",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Java package",
			Content:    regexp.MustCompile(`^package\s+[\w.]+;`),
			Confidence: 0.8,
			Priority:   2,
		},
		{
			Name:       "Java import",
			Content:    regexp.MustCompile(`^import\s+[\w.]+;`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "Java class",
			Content:    regexp.MustCompile(`public\s+class\s+\w+`),
			Confidence: 0.8,
			Priority:   2,
		},
	},
	"csharp": {
		{
			Name:        "C# files",
			FilePattern: "*.cs",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "C# project",
			FilePattern: "*.csproj",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "C# solution",
			FilePattern: "*.sln",
			Confidence:  0.85,
			Priority:    1,
		},
		{
			Name:       "C# using",
			Content:    regexp.MustCompile(`^using\s+\w+`),
			Confidence: 0.7,
			Priority:   3,
		},
		{
			Name:       "C# namespace",
			Content:    regexp.MustCompile(`namespace\s+[\w.]+`),
			Confidence: 0.8,
			Priority:   2,
		},
	},
	"php": {
		{
			Name:        "PHP files",
			FilePattern: "*.php",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "PHP composer",
			FilePattern: "composer.json",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "PHP tags",
			Content:    regexp.MustCompile(`<\?php`),
			Confidence: 0.9,
			Priority:   2,
		},
	},
	"ruby": {
		{
			Name:        "Ruby files",
			FilePattern: "*.rb",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Gemfile",
			FilePattern: "Gemfile",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:       "Ruby shebang",
			Content:    regexp.MustCompile(`^#!/usr/bin/env ruby`),
			Confidence: 0.8,
			Priority:   2,
		},
		{
			Name:       "Ruby require",
			Content:    regexp.MustCompile(`require\s+['"]`),
			Confidence: 0.7,
			Priority:   3,
		},
	},
	"cpp": {
		{
			Name:        "C++ files",
			FilePattern: "*.cpp",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "C++ headers",
			FilePattern: "*.hpp",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "C++ alternate extension",
			FilePattern: "*.cxx",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "CMake",
			FilePattern: "CMakeLists.txt",
			Confidence:  0.8,
			Priority:    1,
		},
		{
			Name:       "C++ includes",
			Content:    regexp.MustCompile(`#include\s*<\w+>`),
			Confidence: 0.7,
			Priority:   3,
		},
	},
	"c": {
		{
			Name:        "C files",
			FilePattern: "*.c",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "C headers",
			FilePattern: "*.h",
			Confidence:  0.8,
			Priority:    2,
		},
		{
			Name:        "Makefile",
			FilePattern: "Makefile",
			Confidence:  0.7,
			Priority:    2,
		},
	},
}

// BuildSystemPatterns defines patterns for detecting build systems
var BuildSystemPatterns = map[string][]DetectionRule{
	"npm": {
		{
			Name:        "package.json",
			FilePattern: "package.json",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "package-lock.json",
			FilePattern: "package-lock.json",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "node_modules",
			FilePattern: "node_modules",
			Confidence:  0.8,
			Priority:    2,
		},
	},
	"yarn": {
		{
			Name:        "yarn.lock",
			FilePattern: "yarn.lock",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        ".yarnrc",
			FilePattern: ".yarnrc",
			Confidence:  0.8,
			Priority:    2,
		},
	},
	"pnpm": {
		{
			Name:        "pnpm-lock.yaml",
			FilePattern: "pnpm-lock.yaml",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "pnpm-workspace.yaml",
			FilePattern: "pnpm-workspace.yaml",
			Confidence:  0.9,
			Priority:    1,
		},
	},
	"cargo": {
		{
			Name:        "Cargo.toml",
			FilePattern: "Cargo.toml",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Cargo.lock",
			FilePattern: "Cargo.lock",
			Confidence:  0.9,
			Priority:    1,
		},
	},
	"pip": {
		{
			Name:        "requirements.txt",
			FilePattern: "requirements.txt",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "setup.py",
			FilePattern: "setup.py",
			Confidence:  0.85,
			Priority:    1,
		},
	},
	"pipenv": {
		{
			Name:        "Pipfile",
			FilePattern: "Pipfile",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "Pipfile.lock",
			FilePattern: "Pipfile.lock",
			Confidence:  0.9,
			Priority:    1,
		},
	},
	"poetry": {
		{
			Name:        "pyproject.toml",
			FilePattern: "pyproject.toml",
			Confidence:  0.9,
			Priority:    1,
		},
	},
	"go": {
		{
			Name:        "go.mod",
			FilePattern: "go.mod",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "go.sum",
			FilePattern: "go.sum",
			Confidence:  0.9,
			Priority:    1,
		},
	},
	"maven": {
		{
			Name:        "pom.xml",
			FilePattern: "pom.xml",
			Confidence:  0.95,
			Priority:    1,
		},
	},
	"gradle": {
		{
			Name:        "build.gradle",
			FilePattern: "build.gradle",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "build.gradle.kts",
			FilePattern: "build.gradle.kts",
			Confidence:  0.95,
			Priority:    1,
		},
		{
			Name:        "gradle.properties",
			FilePattern: "gradle.properties",
			Confidence:  0.8,
			Priority:    2,
		},
	},
	"cmake": {
		{
			Name:        "CMakeLists.txt",
			FilePattern: "CMakeLists.txt",
			Confidence:  0.95,
			Priority:    1,
		},
	},
	"make": {
		{
			Name:        "Makefile",
			FilePattern: "Makefile",
			Confidence:  0.9,
			Priority:    1,
		},
		{
			Name:        "makefile",
			FilePattern: "makefile",
			Confidence:  0.9,
			Priority:    1,
		},
	},
}

// GetLanguagePatterns returns detection patterns for a specific language
func GetLanguagePatterns(language string) ([]DetectionRule, bool) {
	patterns, exists := LanguagePatterns[language]
	return patterns, exists
}

// GetBuildSystemPatterns returns detection patterns for a specific build system
func GetBuildSystemPatterns(buildSystem string) ([]DetectionRule, bool) {
	patterns, exists := BuildSystemPatterns[buildSystem]
	return patterns, exists
}

// GetAllLanguages returns a list of all supported languages
func GetAllLanguages() []string {
	languages := make([]string, 0, len(LanguagePatterns))
	for lang := range LanguagePatterns {
		languages = append(languages, lang)
	}
	return languages
}

// GetAllBuildSystems returns a list of all supported build systems
func GetAllBuildSystems() []string {
	systems := make([]string, 0, len(BuildSystemPatterns))
	for system := range BuildSystemPatterns {
		systems = append(systems, system)
	}
	return systems
}
