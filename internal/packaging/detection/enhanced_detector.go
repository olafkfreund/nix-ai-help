package detection

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// EnhancedDetector provides advanced language and build system detection
type EnhancedDetector struct {
	logger   *logger.Logger
	patterns map[string][]DetectionRule
}

// NewEnhancedDetector creates a new enhanced detector instance
func NewEnhancedDetector(log *logger.Logger) *EnhancedDetector {
	return &EnhancedDetector{
		logger:   log,
		patterns: LanguagePatterns,
	}
}

// DetectLanguages analyzes a repository and returns detected languages with confidence scores
func (d *EnhancedDetector) DetectLanguages(ctx context.Context, repoPath string, opts *AnalysisOptions) ([]LanguageResult, error) {
	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	d.logger.Debug(fmt.Sprintf("Starting enhanced language detection for path: %s", repoPath))
	start := time.Now()

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Collect file information
	files, err := d.collectFiles(timeoutCtx, repoPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	d.logger.Debug(fmt.Sprintf("Collected %d files for analysis", len(files)))

	// Analyze files for language patterns
	languageScores := make(map[string]*LanguageScore)

	for _, file := range files {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("analysis timeout exceeded")
		default:
		}

		if err := d.analyzeFile(file, languageScores); err != nil {
			d.logger.Warn(fmt.Sprintf("Failed to analyze file %s: %v", file.RelPath, err))
			continue
		}
	}

	// Convert scores to results
	results := d.calculateResults(languageScores)

	// Filter by minimum confidence
	filteredResults := make([]LanguageResult, 0)
	for _, result := range results {
		if result.Confidence >= opts.MinConfidence {
			filteredResults = append(filteredResults, result)
		}
	}

	// Sort by confidence (highest first)
	sort.Slice(filteredResults, func(i, j int) bool {
		return filteredResults[i].Confidence > filteredResults[j].Confidence
	})

	d.logger.Debug(fmt.Sprintf("Language detection completed in %v, found %d languages", time.Since(start), len(filteredResults)))

	return filteredResults, nil
}

// FileInfo represents information about a file
type FileInfo struct {
	AbsPath string
	RelPath string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// LanguageScore tracks scoring information for a language
type LanguageScore struct {
	Language   string
	TotalScore float64
	Evidence   []DetectionEvidence
	Files      []string
	Metadata   map[string]interface{}
}

// collectFiles walks the repository and collects file information
func (d *EnhancedDetector) collectFiles(ctx context.Context, repoPath string, opts *AnalysisOptions) ([]FileInfo, error) {
	var files []FileInfo
	count := 0

	err := filepath.WalkDir(repoPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check file count limit
		if count >= opts.MaxFiles {
			return filepath.SkipDir
		}

		// Get relative path
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// Skip hidden files if not included
		if !opts.IncludeHidden && strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be analyzed
		if !opts.ShouldAnalyzeFile(relPath) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			return err
		}

		files = append(files, FileInfo{
			AbsPath: path,
			RelPath: relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		})

		count++
		return nil
	})

	return files, err
}

// analyzeFile analyzes a single file and updates language scores
func (d *EnhancedDetector) analyzeFile(file FileInfo, scores map[string]*LanguageScore) error {
	// Skip directories
	if file.IsDir {
		return nil
	}

	// Analyze file extension
	d.analyzeFileExtension(file, scores)

	// Analyze file content for small text files
	if file.Size < 1024*1024 && d.isTextFile(file.AbsPath) { // 1MB limit
		if err := d.analyzeFileContent(file, scores); err != nil {
			return err
		}
	}

	return nil
}

// analyzeFileExtension checks file extension patterns
func (d *EnhancedDetector) analyzeFileExtension(file FileInfo, scores map[string]*LanguageScore) {
	fileName := filepath.Base(file.RelPath)

	for language, patterns := range LanguagePatterns {
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern.FilePattern, fileName); matched {
				d.addEvidence(scores, language, DetectionEvidence{
					Type:       EvidenceFileExtension,
					Source:     file.RelPath,
					Confidence: pattern.Confidence,
					Details:    fmt.Sprintf("File matches pattern: %s", pattern.FilePattern),
				}, pattern.Confidence)
			}
		}
	}
}

// analyzeFileContent analyzes file content for language patterns
func (d *EnhancedDetector) analyzeFileContent(file FileInfo, scores map[string]*LanguageScore) error {
	content, err := d.readFileContent(file.AbsPath, 8192) // Read first 8KB
	if err != nil {
		return err
	}

	for language, patterns := range LanguagePatterns {
		for _, pattern := range patterns {
			if pattern.Content != nil && pattern.Content.MatchString(content) {
				d.addEvidence(scores, language, DetectionEvidence{
					Type:       EvidenceContentPattern,
					Source:     file.RelPath,
					Confidence: pattern.Confidence,
					Details:    fmt.Sprintf("Content matches pattern: %s", pattern.Name),
				}, pattern.Confidence)
			}
		}
	}

	return nil
}

// readFileContent reads the beginning of a file safely
func (d *EnhancedDetector) readFileContent(path string, maxBytes int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var content strings.Builder
	bytesRead := 0

	for scanner.Scan() && bytesRead < maxBytes {
		line := scanner.Text()
		content.WriteString(line)
		content.WriteString("\n")
		bytesRead += len(line) + 1
	}

	return content.String(), scanner.Err()
}

// isTextFile determines if a file is likely to be text
func (d *EnhancedDetector) isTextFile(path string) bool {
	// Simple heuristic based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".json": true, ".yaml": true, ".yml": true,
		".toml": true, ".ini": true, ".cfg": true, ".conf": true,
		".js": true, ".ts": true, ".py": true, ".go": true, ".rs": true,
		".java": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".cs": true, ".php": true, ".rb": true, ".sh": true, ".bash": true,
		".dockerfile": true, ".gitignore": true, ".gitattributes": true,
	}

	return textExtensions[ext] || ext == ""
}

// addEvidence adds evidence for a language detection
func (d *EnhancedDetector) addEvidence(scores map[string]*LanguageScore, language string, evidence DetectionEvidence, score float64) {
	if scores[language] == nil {
		scores[language] = &LanguageScore{
			Language: language,
			Evidence: []DetectionEvidence{},
			Files:    []string{},
			Metadata: make(map[string]interface{}),
		}
	}

	langScore := scores[language]
	langScore.TotalScore += score
	langScore.Evidence = append(langScore.Evidence, evidence)

	// Add file to list if not already present
	found := false
	for _, file := range langScore.Files {
		if file == evidence.Source {
			found = true
			break
		}
	}
	if !found {
		langScore.Files = append(langScore.Files, evidence.Source)
	}
}

// calculateResults converts language scores to results with normalized confidence
func (d *EnhancedDetector) calculateResults(scores map[string]*LanguageScore) []LanguageResult {
	if len(scores) == 0 {
		return []LanguageResult{}
	}

	// Find maximum score for normalization
	maxScore := 0.0
	for _, score := range scores {
		if score.TotalScore > maxScore {
			maxScore = score.TotalScore
		}
	}

	// Convert to results
	results := make([]LanguageResult, 0, len(scores))
	for _, score := range scores {
		confidence := score.TotalScore
		if maxScore > 0 {
			confidence = score.TotalScore / maxScore
		}

		results = append(results, LanguageResult{
			Language:   score.Language,
			Confidence: confidence,
			Evidence:   score.Evidence,
			Files:      score.Files,
			Metadata:   score.Metadata,
		})
	}

	return results
}
