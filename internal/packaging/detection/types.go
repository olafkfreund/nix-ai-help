package detection

import (
	"path/filepath"
	"regexp"
	"time"
)

// DetectionRule represents a rule for detecting a programming language
type DetectionRule struct {
	Name        string         // Human-readable name
	FilePattern string         // Glob pattern for files
	Content     *regexp.Regexp // Optional content pattern
	Confidence  float64        // Base confidence score (0.0-1.0)
	Priority    int            // Priority for conflict resolution
}

// LanguageResult represents the result of language detection
type LanguageResult struct {
	Language   string                 // Detected language
	Confidence float64                // Overall confidence score
	Evidence   []DetectionEvidence    // Supporting evidence
	Files      []string               // Files that contributed to detection
	Metadata   map[string]interface{} // Additional metadata
}

// DetectionEvidence represents evidence for a language detection
type DetectionEvidence struct {
	Type       EvidenceType // Type of evidence
	Source     string       // File or source that provided evidence
	Confidence float64      // Confidence of this evidence
	Details    string       // Human-readable details
}

// EvidenceType represents the type of detection evidence
type EvidenceType int

const (
	EvidenceFileExtension EvidenceType = iota
	EvidenceContentPattern
	EvidenceConfigFile
	EvidencePackageManager
	EvidenceShebang
	EvidenceImportStatement
	EvidenceBuildFile
	EvidenceDocumentation
)

// String returns the string representation of EvidenceType
func (e EvidenceType) String() string {
	switch e {
	case EvidenceFileExtension:
		return "file_extension"
	case EvidenceContentPattern:
		return "content_pattern"
	case EvidenceConfigFile:
		return "config_file"
	case EvidencePackageManager:
		return "package_manager"
	case EvidenceShebang:
		return "shebang"
	case EvidenceImportStatement:
		return "import_statement"
	case EvidenceBuildFile:
		return "build_file"
	case EvidenceDocumentation:
		return "documentation"
	default:
		return "unknown"
	}
}

// AnalysisOptions controls the behavior of language detection
type AnalysisOptions struct {
	MaxFiles        int           // Maximum number of files to analyze
	Timeout         time.Duration // Maximum time to spend on analysis
	IncludeHidden   bool          // Whether to include hidden files
	MinConfidence   float64       // Minimum confidence threshold
	ExcludePatterns []string      // Patterns to exclude from analysis
	IncludePatterns []string      // Patterns to include in analysis
}

// DefaultAnalysisOptions returns sensible defaults for analysis
func DefaultAnalysisOptions() *AnalysisOptions {
	return &AnalysisOptions{
		MaxFiles:      1000,
		Timeout:       30 * time.Second,
		IncludeHidden: false,
		MinConfidence: 0.1,
		ExcludePatterns: []string{
			"**/node_modules/**",
			"**/vendor/**",
			"**/target/**",
			"**/.git/**",
			"**/build/**",
			"**/dist/**",
			"**/__pycache__/**",
			"**/.next/**",
			"**/.nuxt/**",
		},
	}
}

// ShouldAnalyzeFile determines if a file should be analyzed based on options
func (opts *AnalysisOptions) ShouldAnalyzeFile(path string) bool {
	// Check exclude patterns first
	for _, pattern := range opts.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return false
		}
	}

	// If include patterns are specified, file must match at least one
	if len(opts.IncludePatterns) > 0 {
		for _, pattern := range opts.IncludePatterns {
			if matched, _ := filepath.Match(pattern, path); matched {
				return true
			}
		}
		return false
	}

	return true
}
