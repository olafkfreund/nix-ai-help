package validation

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/pkg/logger"
)

func TestEnhancedValidator_ValidateAnswer(t *testing.T) {
	// Create validator with minimal dependencies for testing
	logger := logger.NewLogger()
	validator := NewEnhancedValidator("", 0, "", logger)

	tests := []struct {
		name           string
		question       string
		answer         string
		expectAccurate bool
		expectQuality  string
		minConfidence  float64
	}{
		{
			name:           "Simple NixOS package installation",
			question:       "How do I install a package in NixOS?",
			answer:         "Add the package to environment.systemPackages in configuration.nix and run sudo nixos-rebuild switch.",
			expectAccurate: true,
			expectQuality:  "fair", // Minimal answer, should be fair quality
			minConfidence:  0.5,
		},
		{
			name:           "Complex flake configuration",
			question:       "How do I create a development environment with Python?",
			answer:         `{ inputs.nixpkgs.url = "github:nixos/nixpkgs"; outputs = { nixpkgs, ... }: { devShells.default = nixpkgs.legacyPackages.x86_64-linux.mkShell { buildInputs = [ nixpkgs.legacyPackages.x86_64-linux.python3 ]; }; }; }`,
			expectAccurate: true,
			expectQuality:  "fair",
			minConfidence:  0.4, // Lower due to complexity
		},
		{
			name:           "Invalid syntax answer",
			question:       "How do I configure NixOS?",
			answer:         "Use { config pkgs } environment.systemPackages = firefox;", // Invalid syntax
			expectAccurate: true,                                                        // Still accurate advice, just poor syntax
			expectQuality:  "poor",
			minConfidence:  0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := validator.ValidateAnswer(ctx, tt.question, tt.answer)
			if err != nil {
				t.Fatalf("ValidateAnswer() error = %v", err)
			}

			if result == nil {
				t.Fatal("ValidateAnswer() returned nil result")
			}

			// Check basic result structure
			if result.IsAccurate != tt.expectAccurate {
				t.Errorf("IsAccurate = %v, want %v", result.IsAccurate, tt.expectAccurate)
			}

			if result.QualityLevel != tt.expectQuality {
				t.Errorf("QualityLevel = %v, want %v", result.QualityLevel, tt.expectQuality)
			}

			if result.ConfidenceScore == nil {
				t.Error("ConfidenceScore is nil")
			} else if result.ConfidenceScore.Overall < tt.minConfidence {
				t.Errorf("Overall confidence = %v, want >= %v", result.ConfidenceScore.Overall, tt.minConfidence)
			}

			// Check that validation time is reasonable (should be less than 30 seconds)
			if result.ValidationTime > 30*time.Second {
				t.Errorf("ValidationTime = %v, too long", result.ValidationTime)
			}

			// Check that sources were consulted
			if len(result.SourcesConsulted) == 0 {
				t.Error("No sources were consulted")
			}

			// Verify confidence score components are within valid range [0,1]
			conf := result.ConfidenceScore
			if conf.Overall < 0 || conf.Overall > 1 {
				t.Errorf("Overall confidence %v outside valid range [0,1]", conf.Overall)
			}
			if conf.SourceVerification < 0 || conf.SourceVerification > 1 {
				t.Errorf("SourceVerification %v outside valid range [0,1]", conf.SourceVerification)
			}
			if conf.ToolVerification < 0 || conf.ToolVerification > 1 {
				t.Errorf("ToolVerification %v outside valid range [0,1]", conf.ToolVerification)
			}
			if conf.SyntaxValidity < 0 || conf.SyntaxValidity > 1 {
				t.Errorf("SyntaxValidity %v outside valid range [0,1]", conf.SyntaxValidity)
			}
		})
	}
}

func TestEnhancedValidator_ComponentIntegration(t *testing.T) {
	logger := logger.NewLogger()
	validator := NewEnhancedValidator("", 0, "", logger)

	// Test that all components are properly initialized
	if validator == nil {
		t.Fatal("NewEnhancedValidator returned nil")
	}

	// Test basic validation with minimal input
	ctx := context.Background()
	result, err := validator.ValidateAnswer(ctx, "test question", "test answer")
	if err != nil {
		t.Fatalf("ValidateAnswer failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check that essential fields are populated
	expectedSources := []string{
		"pre-answer-validation",
		"nixos-validator",
		"flake-validator",
		"nix-tools",
		"community-sources",
		"search-nixos-org",
		"cross-reference",
	}

	if len(result.SourcesConsulted) != len(expectedSources) {
		t.Errorf("Expected %d sources, got %d", len(expectedSources), len(result.SourcesConsulted))
	}

	// Verify all expected sources are present
	sourceMap := make(map[string]bool)
	for _, source := range result.SourcesConsulted {
		sourceMap[source] = true
	}

	for _, expected := range expectedSources {
		if !sourceMap[expected] {
			t.Errorf("Expected source %s not found in result", expected)
		}
	}
}

func TestConfidenceScoring(t *testing.T) {
	logger := logger.NewLogger()
	validator := NewEnhancedValidator("", 0, "", logger)

	ctx := context.Background()

	// Test with a high-quality answer
	goodAnswer := "To install a package in NixOS, add it to environment.systemPackages = with pkgs; [ packageName ]; in your configuration.nix file, then run sudo nixos-rebuild switch to apply the changes."

	result, err := validator.ValidateAnswer(ctx, "How to install packages?", goodAnswer)
	if err != nil {
		t.Fatalf("ValidateAnswer failed: %v", err)
	}

	// High quality answer should have decent confidence
	if result.ConfidenceScore.Overall < 0.4 {
		t.Errorf("Expected confidence >= 0.4 for good answer, got %v", result.ConfidenceScore.Overall)
	}

	// Test with a poor quality answer
	poorAnswer := "install package"

	result2, err := validator.ValidateAnswer(ctx, "How to install packages?", poorAnswer)
	if err != nil {
		t.Fatalf("ValidateAnswer failed: %v", err)
	}

	// Poor answer should have lower confidence than good answer
	if result2.ConfidenceScore.Overall >= result.ConfidenceScore.Overall {
		t.Errorf("Poor answer confidence (%v) should be lower than good answer (%v)",
			result2.ConfidenceScore.Overall, result.ConfidenceScore.Overall)
	}
}

func TestEnhancedValidatorWithAutomatedScoring(t *testing.T) {
	// Initialize logger
	testLogger := logger.NewLogger()

	// Create enhanced validator
	validator := NewEnhancedValidator("localhost", 8080, "", testLogger)

	// Test question and answer
	question := "How do I install Firefox on NixOS?"
	answer := `To install Firefox on NixOS, you can add it to your configuration.nix:

{
  environment.systemPackages = with pkgs; [
    firefox
  ];
}

Then rebuild your system:
sudo nixos-rebuild switch`

	// Test validation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := validator.ValidateAnswer(ctx, question, answer)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check that automated quality score was computed
	if result.AutomatedQualityScore == nil {
		t.Error("AutomatedQualityScore should not be nil")
	} else {
		t.Logf("Automated Quality Score: %d/100", result.AutomatedQualityScore.OverallScore)
		t.Logf("Breakdown: Syntax=%d, Package=%d, Option=%d, Command=%d, Structure=%d",
			result.AutomatedQualityScore.BreakdownScores.SyntaxScore,
			result.AutomatedQualityScore.BreakdownScores.PackageScore,
			result.AutomatedQualityScore.BreakdownScores.OptionScore,
			result.AutomatedQualityScore.BreakdownScores.CommandScore,
			result.AutomatedQualityScore.BreakdownScores.StructureScore,
		)

		// Verify score is within valid range
		if result.AutomatedQualityScore.OverallScore < 0 || result.AutomatedQualityScore.OverallScore > 100 {
			t.Errorf("Overall score %d is out of valid range [0-100]", result.AutomatedQualityScore.OverallScore)
		}
	}

	// Check basic validation result fields
	t.Logf("Overall Quality Level: %s", result.QualityLevel)
	t.Logf("Is Accurate: %v", result.IsAccurate)
	t.Logf("Sources Consulted: %d", len(result.SourcesConsulted))
	t.Logf("Quality Issues: %d", len(result.QualityIssues))
	t.Logf("Recommendations: %d", len(result.Recommendations))
	t.Logf("Validation Time: %v", result.ValidationTime)

	// Verify that automated quality scorer was consulted
	found := false
	for _, source := range result.SourcesConsulted {
		if source == "automated-quality-scorer" {
			found = true
			break
		}
	}
	if !found {
		t.Error("automated-quality-scorer should be in sources consulted")
	}

	// Log any issues found
	for i, issue := range result.QualityIssues {
		t.Logf("Issue %d: [%s] %s - %s", i+1, issue.Severity, issue.Type, issue.Message)
	}

	// Log recommendations
	for i, rec := range result.Recommendations {
		t.Logf("Recommendation %d: %s", i+1, rec)
	}
}

func TestAutomatedQualityScorerStandalone(t *testing.T) {
	scorer := NewAutomatedQualityScorer()

	question := "How do I enable SSH on NixOS?"
	answer := `To enable SSH on NixOS, add this to your configuration.nix:

{
  services.openssh.enable = true;
  services.openssh.settings.PermitRootLogin = "no";
}

Then rebuild:
sudo nixos-rebuild switch`

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	score, err := scorer.ScoreAnswer(ctx, question, answer)
	if err != nil {
		t.Fatalf("Scoring failed: %v", err)
	}

	if score == nil {
		t.Fatal("Score is nil")
	}

	t.Logf("Overall Score: %d/100", score.OverallScore)
	t.Logf("Execution Time: %v", score.ExecutionTime)
	t.Logf("Commands Run: %d", len(score.CommandsRun))

	// Verify breakdown scores
	breakdown := score.BreakdownScores
	t.Logf("Breakdown: Syntax=%d, Package=%d, Option=%d, Command=%d, Structure=%d",
		breakdown.SyntaxScore, breakdown.PackageScore, breakdown.OptionScore,
		breakdown.CommandScore, breakdown.StructureScore)

	// Log validation results
	results := score.ValidationResults
	t.Logf("Validation Results: Syntax=%v, Flake=%v, Config=%v",
		results.SyntaxValid, results.FlakeValid, results.ConfigurationValid)
	t.Logf("Packages validated: %d", len(results.PackagesValid))
	t.Logf("Options validated: %d", len(results.OptionsValid))
	t.Logf("Commands validated: %d", len(results.CommandsValid))

	// Log issues and recommendations
	for i, issue := range score.Issues {
		t.Logf("Issue %d: [%s] %s - %s", i+1, issue.Severity, issue.Category, issue.Message)
	}

	for i, rec := range score.Recommendations {
		t.Logf("Recommendation %d: %s", i+1, rec)
	}
}

func TestQualityLevelDeterminationWithAutomatedScore(t *testing.T) {
	testLogger := logger.NewLogger()
	validator := NewEnhancedValidator("localhost", 8080, "", testLogger)

	tests := []struct {
		name           string
		automatedScore int
		confidence     float64
		highIssues     int
		criticalIssues int
		expected       string
	}{
		{
			name:           "Excellent quality",
			automatedScore: 95,
			confidence:     0.95,
			highIssues:     0,
			criticalIssues: 0,
			expected:       "excellent",
		},
		{
			name:           "Good quality",
			automatedScore: 80,
			confidence:     0.80,
			highIssues:     0,
			criticalIssues: 0,
			expected:       "good",
		},
		{
			name:           "Fair quality",
			automatedScore: 65,
			confidence:     0.60,
			highIssues:     1,
			criticalIssues: 0,
			expected:       "fair",
		},
		{
			name:           "Poor quality - low score",
			automatedScore: 35,
			confidence:     0.40,
			highIssues:     0,
			criticalIssues: 0,
			expected:       "poor",
		},
		{
			name:           "Poor quality - critical issue",
			automatedScore: 90,
			confidence:     0.90,
			highIssues:     0,
			criticalIssues: 1,
			expected:       "poor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EnhancedValidationResult{
				IsAccurate: true,
				ConfidenceScore: &AnswerConfidence{
					Overall: tt.confidence,
				},
				AutomatedQualityScore: &AutomatedQualityScore{
					OverallScore: tt.automatedScore,
				},
				QualityIssues: make([]QualityIssue, tt.highIssues+tt.criticalIssues),
			}

			// Add high severity issues
			for i := 0; i < tt.highIssues; i++ {
				result.QualityIssues[i] = QualityIssue{Severity: "high"}
			}

			// Add critical severity issues
			for i := tt.highIssues; i < tt.highIssues+tt.criticalIssues; i++ {
				result.QualityIssues[i] = QualityIssue{Severity: "critical"}
			}

			qualityLevel := validator.determineQualityLevel(result)
			if qualityLevel != tt.expected {
				t.Errorf("Expected quality level %s, got %s", tt.expected, qualityLevel)
			}
		})
	}
}
