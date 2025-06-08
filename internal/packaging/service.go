package packaging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/pkg/logger"
)

// PackagingService coordinates repository analysis and derivation generation
type PackagingService struct {
	analyzer  *RepositoryAnalyzer
	generator *DerivationGenerator
	cloner    *GitCloner
	logger    *logger.Logger
}

// PackageRequest represents a request to package a repository
type PackageRequest struct {
	RepoURL     string `json:"repo_url"`
	LocalPath   string `json:"local_path,omitempty"`
	OutputPath  string `json:"output_path,omitempty"`
	PackageName string `json:"package_name,omitempty"`
	Quiet       bool   `json:"quiet,omitempty"`
}

// PackageResult represents the result of packaging operation
type PackageResult struct {
	Analysis         *RepoAnalysis     `json:"analysis"`
	Derivation       string            `json:"derivation"`
	ValidationIssues []string          `json:"validation_issues,omitempty"`
	NixpkgsMappings  map[string]string `json:"nixpkgs_mappings,omitempty"`
	OutputFile       string            `json:"output_file,omitempty"`
}

// NewPackagingService creates a new packaging service
func NewPackagingService(aiProvider ai.AIProvider, mcpClient *mcp.MCPClient, tempDir string, logger *logger.Logger) *PackagingService {
	return &PackagingService{
		analyzer:  NewRepositoryAnalyzer(logger),
		generator: NewDerivationGenerator(aiProvider, mcpClient),
		cloner:    NewGitCloner(tempDir),
		logger:    logger,
	}
}

// PackageRepository packages a Git repository into a Nix derivation
func (ps *PackagingService) PackageRepository(ctx context.Context, req *PackageRequest) (*PackageResult, error) {
	var repoPath string
	var err error
	var shouldCleanup bool

	// Determine repository path
	if req.LocalPath != "" {
		repoPath = req.LocalPath
		ps.logger.Debug(fmt.Sprintf("Using local repository path: %s", repoPath))
	} else if req.RepoURL != "" {
		ps.logger.Info(fmt.Sprintf("Cloning repository: %s", req.RepoURL))
		repoPath, err = ps.cloneRepository(req.RepoURL, req.Quiet)
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
		shouldCleanup = true
		ps.logger.Debug(fmt.Sprintf("Repository cloned to: %s", repoPath))
	} else {
		return nil, fmt.Errorf("either repo_url or local_path must be provided")
	}

	// Ensure cleanup if we cloned the repository
	if shouldCleanup {
		defer func() {
			if err := os.RemoveAll(repoPath); err != nil {
				ps.logger.Warn(fmt.Sprintf("Failed to cleanup cloned repository %s: %v", repoPath, err))
			}
		}()
	}

	// Analyze repository
	ps.logger.Info(fmt.Sprintf("Analyzing repository: %s", repoPath))
	analysis, err := ps.analyzer.AnalyzeRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze repository: %w", err)
	}

	// Set repository URL in analysis if provided
	if req.RepoURL != "" {
		analysis.RepoURL = req.RepoURL
	}

	// Override package name if provided
	if req.PackageName != "" {
		analysis.ProjectName = req.PackageName
	}

	ps.logger.Info(fmt.Sprintf("Repository analysis complete - project: %s, build_system: %s, language: %s, dependencies: %d",
		analysis.ProjectName, analysis.BuildSystem, analysis.Language, len(analysis.Dependencies)))

	// Generate derivation
	ps.logger.Info("Generating Nix derivation")
	derivation, err := ps.generator.GenerateDerivation(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate derivation: %w", err)
	}

	// Validate derivation
	validationIssues := ps.generator.ValidateDerivation(derivation)
	if len(validationIssues) > 0 {
		ps.logger.Warn(fmt.Sprintf("Derivation validation issues found: %v", validationIssues))
	}

	// Get nixpkgs mappings for dependencies
	nixpkgsMappings, err := ps.generator.SuggestNixpkgsMappings(ctx, analysis.Dependencies)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to generate nixpkgs mappings: %v", err))
		nixpkgsMappings = make(map[string]string)
	}

	// Save derivation to file if output path specified
	var outputFile string
	if req.OutputPath != "" {
		outputFile, err = ps.saveDerivation(derivation, req.OutputPath, analysis.ProjectName)
		if err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to save derivation to file: %v", err))
		} else {
			ps.logger.Info(fmt.Sprintf("Derivation saved to: %s", outputFile))
		}
	}

	result := &PackageResult{
		Analysis:         analysis,
		Derivation:       derivation,
		ValidationIssues: validationIssues,
		NixpkgsMappings:  nixpkgsMappings,
		OutputFile:       outputFile,
	}

	return result, nil
}

// cloneRepository clones a repository and returns the local path
func (ps *PackagingService) cloneRepository(repoURL string, quiet bool) (string, error) {
	if quiet {
		return ps.cloner.CloneRepositoryQuiet(repoURL)
	}
	return ps.cloner.CloneRepository(repoURL)
}

// saveDerivation saves the derivation to a file
func (ps *PackagingService) saveDerivation(derivation, outputPath, projectName string) (string, error) {
	// #nosec G301 -- Output directory is user-controlled and not sensitive
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s.nix", projectName)
	filepath := filepath.Join(outputPath, filename)

	// #nosec G306 -- Derivation output is not sensitive, 0644 is intentional for Nix usage
	if err := os.WriteFile(filepath, []byte(derivation), 0644); err != nil {
		return "", fmt.Errorf("failed to write derivation file: %w", err)
	}

	return filepath, nil
}

// AnalyzeLocalRepository analyzes a local repository without generating a derivation
func (ps *PackagingService) AnalyzeLocalRepository(repoPath string) (*RepoAnalysis, error) {
	ps.logger.Info(fmt.Sprintf("Analyzing local repository: %s", repoPath))

	analysis, err := ps.analyzer.AnalyzeRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze repository: %w", err)
	}

	ps.logger.Info(fmt.Sprintf("Repository analysis complete - project: %s, build_system: %s, language: %s, dependencies: %d",
		analysis.ProjectName, analysis.BuildSystem, analysis.Language, len(analysis.Dependencies)))

	return analysis, nil
}

// GenerateDerivationFromAnalysis generates a derivation from existing analysis
func (ps *PackagingService) GenerateDerivationFromAnalysis(ctx context.Context, analysis *RepoAnalysis) (*PackageResult, error) {
	ps.logger.Info(fmt.Sprintf("Generating derivation from analysis for project: %s", analysis.ProjectName))

	// Generate derivation
	derivation, err := ps.generator.GenerateDerivation(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate derivation: %w", err)
	}

	// Validate derivation
	validationIssues := ps.generator.ValidateDerivation(derivation)
	if len(validationIssues) > 0 {
		ps.logger.Warn(fmt.Sprintf("Derivation validation issues found: %v", validationIssues))
	}

	// Get nixpkgs mappings
	nixpkgsMappings, err := ps.generator.SuggestNixpkgsMappings(ctx, analysis.Dependencies)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to generate nixpkgs mappings: %v", err))
		nixpkgsMappings = make(map[string]string)
	}

	result := &PackageResult{
		Analysis:         analysis,
		Derivation:       derivation,
		ValidationIssues: validationIssues,
		NixpkgsMappings:  nixpkgsMappings,
	}

	return result, nil
}

// Cleanup cleans up any temporary resources
func (ps *PackagingService) Cleanup() error {
	return ps.cloner.Cleanup()
}
