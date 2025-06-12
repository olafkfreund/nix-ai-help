package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// FactChecker validates factual accuracy of answers using multiple verification methods
type FactChecker struct {
	logger *logger.Logger
}

// FactCheckResult represents the result of fact checking
type FactCheckResult struct {
	OverallAccuracy   float64          `json:"overall_accuracy"`
	VerifiedFacts     []VerifiedFact   `json:"verified_facts"`
	UnverifiedFacts   []UnverifiedFact `json:"unverified_facts"`
	FactualErrors     []FactualError   `json:"factual_errors"`
	SourceQuality     float64          `json:"source_quality"`
	RecencyScore      float64          `json:"recency_score"`
	ConsistencyScore  float64          `json:"consistency_score"`
	ValidationMethods []string         `json:"validation_methods"`
	ValidationTime    time.Duration    `json:"validation_time"`
	Recommendations   []string         `json:"recommendations"`
}

// VerifiedFact represents a fact that has been verified
type VerifiedFact struct {
	Statement   string   `json:"statement"`
	Sources     []string `json:"sources"`
	Confidence  float64  `json:"confidence"`
	LastUpdated string   `json:"last_updated"`
	Type        string   `json:"type"` // "package", "option", "command", "concept"
}

// UnverifiedFact represents a fact that could not be verified
type UnverifiedFact struct {
	Statement string `json:"statement"`
	Reason    string `json:"reason"`
	Severity  string `json:"severity"` // "low", "medium", "high"
	Type      string `json:"type"`
}

// FactualError represents a detected factual error
type FactualError struct {
	Statement  string `json:"statement"`
	ErrorType  string `json:"error_type"` // "outdated", "incorrect", "misleading"
	Correction string `json:"correction"`
	Evidence   string `json:"evidence"`
	Severity   string `json:"severity"`
	Source     string `json:"source"`
}

// Fact represents an atomic fact extracted from text
type Fact struct {
	Statement  string
	Category   string // "package", "option", "command", "version", "concept"
	Confidence float64
	Context    string
}

// ValidationMethod represents a method used for fact checking
type ValidationMethod struct {
	Name        string
	Type        string // "static", "dynamic", "external"
	Reliability float64
	Speed       string // "fast", "medium", "slow"
	Enabled     bool
}

// NewFactChecker creates a new fact checker
func NewFactChecker() *FactChecker {
	return &FactChecker{
		logger: logger.NewLogger(),
	}
}

// CheckFacts performs comprehensive fact checking on an answer
func (fc *FactChecker) CheckFacts(ctx context.Context, question, answer string, validationResult *EnhancedValidationResult) (*FactCheckResult, error) {
	startTime := time.Now()

	result := &FactCheckResult{
		OverallAccuracy:   1.0,
		VerifiedFacts:     []VerifiedFact{},
		UnverifiedFacts:   []UnverifiedFact{},
		FactualErrors:     []FactualError{},
		ValidationMethods: []string{},
		Recommendations:   []string{},
	}

	// Extract facts from the answer
	facts := fc.extractFacts(answer)
	fc.logger.Printf("Facts extracted - count: %d", len(facts))

	// Verify each fact using multiple methods
	for _, fact := range facts {
		fc.verifyFact(ctx, fact, validationResult, result)
	}

	// Check for common factual errors
	fc.checkCommonErrors(answer, result)

	// Calculate overall scores
	result.OverallAccuracy = fc.calculateOverallAccuracy(result)
	result.SourceQuality = fc.calculateSourceQuality(validationResult)
	result.RecencyScore = fc.calculateRecencyScore(validationResult)
	result.ConsistencyScore = fc.calculateConsistencyScore(result)

	// Generate recommendations
	result.Recommendations = fc.generateFactCheckRecommendations(result)

	result.ValidationTime = time.Since(startTime)

	fc.logger.Printf("Fact checking completed - overall_accuracy: %.2f, verified_facts: %d, unverified_facts: %d, factual_errors: %d, validation_time: %v",
		result.OverallAccuracy,
		len(result.VerifiedFacts),
		len(result.UnverifiedFacts),
		len(result.FactualErrors),
		result.ValidationTime,
	)

	return result, nil
}

// extractFacts extracts factual statements from the answer
func (fc *FactChecker) extractFacts(answer string) []Fact {
	var facts []Fact

	// Extract package names
	packageFacts := fc.extractPackageFacts(answer)
	facts = append(facts, packageFacts...)

	// Extract option names
	optionFacts := fc.extractOptionFacts(answer)
	facts = append(facts, optionFacts...)

	// Extract command statements
	commandFacts := fc.extractCommandFacts(answer)
	facts = append(facts, commandFacts...)

	// Extract version statements
	versionFacts := fc.extractVersionFacts(answer)
	facts = append(facts, versionFacts...)

	// Extract conceptual facts
	conceptFacts := fc.extractConceptualFacts(answer)
	facts = append(facts, conceptFacts...)

	return facts
}

// extractPackageFacts extracts facts about packages
func (fc *FactChecker) extractPackageFacts(answer string) []Fact {
	var facts []Fact

	// Pattern for package declarations
	packagePatterns := []*regexp.Regexp{
		regexp.MustCompile(`packages\s*=\s*\[\s*([^]]+)\s*\]`),
		regexp.MustCompile(`environment\.systemPackages\s*=\s*with\s+pkgs;\s*\[\s*([^]]+)\s*\]`),
		regexp.MustCompile(`pkgs\.([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`environment\.systemPackages\s*=\s*\[\s*([^]]+)\s*\]`),
	}

	for _, pattern := range packagePatterns {
		matches := pattern.FindAllStringSubmatch(answer, -1)
		for _, match := range matches {
			if len(match) > 1 {
				packages := fc.parsePackageList(match[1])
				for _, pkg := range packages {
					facts = append(facts, Fact{
						Statement:  fmt.Sprintf("Package '%s' exists", pkg),
						Category:   "package",
						Confidence: 0.8,
						Context:    match[0],
					})
				}
			}
		}
	}

	return facts
}

// extractOptionFacts extracts facts about NixOS options
func (fc *FactChecker) extractOptionFacts(answer string) []Fact {
	var facts []Fact

	// Pattern for option declarations
	optionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`([a-zA-Z0-9_.]+)\s*=\s*[^;]+;`),
		regexp.MustCompile(`services\.([a-zA-Z0-9_-]+)\.([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`boot\.([a-zA-Z0-9_-]+)\.([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`networking\.([a-zA-Z0-9_-]+)`),
		regexp.MustCompile(`hardware\.([a-zA-Z0-9_-]+)`),
	}

	for _, pattern := range optionPatterns {
		matches := pattern.FindAllStringSubmatch(answer, -1)
		for _, match := range matches {
			if len(match) > 0 {
				option := match[1]
				if len(match) > 2 {
					option = match[1] + "." + match[2]
				}

				// Skip obvious non-options
				if fc.isLikelyOption(option) {
					facts = append(facts, Fact{
						Statement:  fmt.Sprintf("Option '%s' exists", option),
						Category:   "option",
						Confidence: 0.7,
						Context:    match[0],
					})
				}
			}
		}
	}

	return facts
}

// extractCommandFacts extracts facts about commands
func (fc *FactChecker) extractCommandFacts(answer string) []Fact {
	var facts []Fact

	// Pattern for command statements
	commandPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\$\s+(nix\s+[^;\n]+)`),
		regexp.MustCompile(`\$\s+(nixos-rebuild\s+[^;\n]+)`),
		regexp.MustCompile(`\$\s+(systemctl\s+[^;\n]+)`),
		regexp.MustCompile(`sudo\s+(nix[a-zA-Z0-9-]*\s+[^;\n]+)`),
	}

	for _, pattern := range commandPatterns {
		matches := pattern.FindAllStringSubmatch(answer, -1)
		for _, match := range matches {
			if len(match) > 1 {
				command := strings.TrimSpace(match[1])
				facts = append(facts, Fact{
					Statement:  fmt.Sprintf("Command '%s' is valid", command),
					Category:   "command",
					Confidence: 0.6,
					Context:    match[0],
				})
			}
		}
	}

	return facts
}

// extractVersionFacts extracts facts about versions
func (fc *FactChecker) extractVersionFacts(answer string) []Fact {
	var facts []Fact

	// Pattern for version statements
	versionPattern := regexp.MustCompile(`(?i)version\s+([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)
	matches := versionPattern.FindAllStringSubmatch(answer, -1)

	for _, match := range matches {
		if len(match) > 1 {
			version := match[1]
			facts = append(facts, Fact{
				Statement:  fmt.Sprintf("Version %s is referenced", version),
				Category:   "version",
				Confidence: 0.5,
				Context:    match[0],
			})
		}
	}

	return facts
}

// extractConceptualFacts extracts conceptual facts
func (fc *FactChecker) extractConceptualFacts(answer string) []Fact {
	var facts []Fact

	// Pattern for conceptual statements
	conceptPatterns := map[string]*regexp.Regexp{
		"flakes enable":    regexp.MustCompile(`flakes?\s+(?:enable|allow|provide)\s+([^.]+)`),
		"service starts":   regexp.MustCompile(`service\s+(?:starts|runs|enables)\s+([^.]+)`),
		"option controls":  regexp.MustCompile(`option\s+(?:controls|manages|sets)\s+([^.]+)`),
		"package provides": regexp.MustCompile(`package\s+(?:provides|includes|contains)\s+([^.]+)`),
	}

	for factType, pattern := range conceptPatterns {
		matches := pattern.FindAllStringSubmatch(answer, -1)
		for _, match := range matches {
			if len(match) > 1 {
				statement := strings.TrimSpace(match[1])
				facts = append(facts, Fact{
					Statement:  fmt.Sprintf("%s: %s", factType, statement),
					Category:   "concept",
					Confidence: 0.4,
					Context:    match[0],
				})
			}
		}
	}

	return facts
}

// verifyFact verifies a single fact using available validation results
func (fc *FactChecker) verifyFact(ctx context.Context, fact Fact, validationResult *EnhancedValidationResult, result *FactCheckResult) {
	switch fact.Category {
	case "package":
		fc.verifyPackageFact(fact, validationResult, result)
	case "option":
		fc.verifyOptionFact(fact, validationResult, result)
	case "command":
		fc.verifyCommandFact(fact, validationResult, result)
	default:
		// For version and conceptual facts, mark as unverified for now
		result.UnverifiedFacts = append(result.UnverifiedFacts, UnverifiedFact{
			Statement: fact.Statement,
			Reason:    "No verification method available",
			Severity:  "low",
			Type:      fact.Category,
		})
	}
}

// verifyPackageFact verifies package-related facts
func (fc *FactChecker) verifyPackageFact(fact Fact, validationResult *EnhancedValidationResult, result *FactCheckResult) {
	packageName := fc.extractPackageNameFromFact(fact.Statement)

	verified := false
	sources := []string{}

	// Check against tool validation
	if validationResult.ToolValidation != nil {
		for _, packageCheck := range validationResult.ToolValidation.PackageChecks {
			if packageCheck.PackageName == packageName {
				if packageCheck.Available {
					verified = true
					sources = append(sources, "nix-tools")
				}
				break
			}
		}
	}

	// Check against source verification
	if validationResult.SourceVerification != nil {
		for _, verifiedPkg := range validationResult.SourceVerification.PackagesVerified {
			if verifiedPkg.Name == packageName {
				verified = true
				sources = append(sources, "search-nixos-org")
				break
			}
		}
	}

	if verified {
		result.VerifiedFacts = append(result.VerifiedFacts, VerifiedFact{
			Statement:   fact.Statement,
			Sources:     sources,
			Confidence:  fact.Confidence,
			LastUpdated: time.Now().Format(time.RFC3339),
			Type:        fact.Category,
		})
	} else {
		result.UnverifiedFacts = append(result.UnverifiedFacts, UnverifiedFact{
			Statement: fact.Statement,
			Reason:    "Package not found in available sources",
			Severity:  "medium",
			Type:      fact.Category,
		})
	}
}

// verifyOptionFact verifies option-related facts
func (fc *FactChecker) verifyOptionFact(fact Fact, validationResult *EnhancedValidationResult, result *FactCheckResult) {
	optionName := fc.extractOptionNameFromFact(fact.Statement)

	verified := false
	sources := []string{}

	// Check against tool validation
	if validationResult.ToolValidation != nil {
		for _, optionCheck := range validationResult.ToolValidation.OptionChecks {
			if optionCheck.OptionName == optionName {
				if optionCheck.Valid {
					verified = true
					sources = append(sources, "nixos-option")
				}
				break
			}
		}
	}

	// Check against source verification
	if validationResult.SourceVerification != nil {
		for _, verifiedOpt := range validationResult.SourceVerification.OptionsVerified {
			if verifiedOpt.Name == optionName {
				verified = true
				sources = append(sources, "search-nixos-org")
				break
			}
		}
	}

	if verified {
		result.VerifiedFacts = append(result.VerifiedFacts, VerifiedFact{
			Statement:   fact.Statement,
			Sources:     sources,
			Confidence:  fact.Confidence,
			LastUpdated: time.Now().Format(time.RFC3339),
			Type:        fact.Category,
		})
	} else {
		result.UnverifiedFacts = append(result.UnverifiedFacts, UnverifiedFact{
			Statement: fact.Statement,
			Reason:    "Option not found in available sources",
			Severity:  "high", // Options are more critical than packages
			Type:      fact.Category,
		})
	}
}

// verifyCommandFact verifies command-related facts
func (fc *FactChecker) verifyCommandFact(fact Fact, validationResult *EnhancedValidationResult, result *FactCheckResult) {
	command := fc.extractCommandFromFact(fact.Statement)

	verified := false
	sources := []string{}

	// Check against tool validation
	if validationResult.ToolValidation != nil {
		for _, commandCheck := range validationResult.ToolValidation.CommandChecks {
			if strings.Contains(commandCheck.Command, command) {
				if commandCheck.Valid {
					verified = true
					sources = append(sources, "command-validation")
				}
				break
			}
		}
	}

	if verified {
		result.VerifiedFacts = append(result.VerifiedFacts, VerifiedFact{
			Statement:   fact.Statement,
			Sources:     sources,
			Confidence:  fact.Confidence,
			LastUpdated: time.Now().Format(time.RFC3339),
			Type:        fact.Category,
		})
	} else {
		result.UnverifiedFacts = append(result.UnverifiedFacts, UnverifiedFact{
			Statement: fact.Statement,
			Reason:    "Command not validated",
			Severity:  "medium",
			Type:      fact.Category,
		})
	}
}

// checkCommonErrors checks for common factual errors
func (fc *FactChecker) checkCommonErrors(answer string, result *FactCheckResult) {
	// Check for deprecated commands
	deprecatedCommands := map[string]string{
		"nix-env -i":              "Use 'nix profile install' or declarative configuration",
		"nix-channel --update":    "Use 'nix flake update' with flakes",
		"nixos-version":           "Use 'nixos-version' or check /etc/os-release",
		"services.xserver.enable": "Consider services.displayManager and services.desktopManager",
	}

	for deprecated, replacement := range deprecatedCommands {
		if strings.Contains(strings.ToLower(answer), strings.ToLower(deprecated)) {
			result.FactualErrors = append(result.FactualErrors, FactualError{
				Statement:  fmt.Sprintf("Uses deprecated command: %s", deprecated),
				ErrorType:  "outdated",
				Correction: replacement,
				Evidence:   "Found in answer text",
				Severity:   "medium",
				Source:     "deprecated-command-check",
			})
		}
	}

	// Check for common misconceptions
	misconceptions := map[string]string{
		"flakes are experimental": "Flakes are stable as of Nix 2.4+",
		"nix is only for nixos":   "Nix package manager works on any Linux/macOS system",
		"nix store is read-only":  "Nix store is immutable but can be modified through nix commands",
	}

	answerLower := strings.ToLower(answer)
	for misconception, correction := range misconceptions {
		if strings.Contains(answerLower, misconception) {
			result.FactualErrors = append(result.FactualErrors, FactualError{
				Statement:  fmt.Sprintf("Contains misconception: %s", misconception),
				ErrorType:  "incorrect",
				Correction: correction,
				Evidence:   "Common misconception detected",
				Severity:   "low",
				Source:     "misconception-check",
			})
		}
	}
}

// Helper methods

func (fc *FactChecker) parsePackageList(packageStr string) []string {
	// Simple parsing of package lists
	packages := strings.Split(packageStr, " ")
	var result []string

	for _, pkg := range packages {
		cleaned := strings.TrimSpace(pkg)
		cleaned = strings.Trim(cleaned, "\"'")
		if cleaned != "" && !strings.Contains(cleaned, "#") {
			result = append(result, cleaned)
		}
	}

	return result
}

func (fc *FactChecker) isLikelyOption(option string) bool {
	// Simple heuristics to determine if something looks like a NixOS option
	if len(option) < 3 {
		return false
	}

	// Common option prefixes
	validPrefixes := []string{
		"services.", "boot.", "networking.", "hardware.", "system.",
		"environment.", "users.", "security.", "programs.", "virtualisation.",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(option, prefix) {
			return true
		}
	}

	// Must contain at least one dot and be reasonable length
	return strings.Contains(option, ".") && len(option) < 100
}

func (fc *FactChecker) extractPackageNameFromFact(statement string) string {
	// Extract package name from "Package 'name' exists"
	pattern := regexp.MustCompile(`Package '([^']+)' exists`)
	matches := pattern.FindStringSubmatch(statement)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (fc *FactChecker) extractOptionNameFromFact(statement string) string {
	// Extract option name from "Option 'name' exists"
	pattern := regexp.MustCompile(`Option '([^']+)' exists`)
	matches := pattern.FindStringSubmatch(statement)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (fc *FactChecker) extractCommandFromFact(statement string) string {
	// Extract command from "Command 'name' is valid"
	pattern := regexp.MustCompile(`Command '([^']+)' is valid`)
	matches := pattern.FindStringSubmatch(statement)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Calculation methods

func (fc *FactChecker) calculateOverallAccuracy(result *FactCheckResult) float64 {
	totalFacts := len(result.VerifiedFacts) + len(result.UnverifiedFacts) + len(result.FactualErrors)

	if totalFacts == 0 {
		return 1.0 // No facts to check
	}

	verifiedScore := float64(len(result.VerifiedFacts))
	errorPenalty := float64(len(result.FactualErrors)) * 1.5 // Errors are more costly
	unverifiedPenalty := float64(len(result.UnverifiedFacts)) * 0.5

	score := (verifiedScore - errorPenalty - unverifiedPenalty) / float64(totalFacts)

	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func (fc *FactChecker) calculateSourceQuality(validationResult *EnhancedValidationResult) float64 {
	if validationResult == nil {
		return 0.5
	}

	score := 0.5
	sourceCount := len(validationResult.SourcesConsulted)

	// More sources generally mean better quality
	if sourceCount >= 4 {
		score += 0.3
	} else if sourceCount >= 2 {
		score += 0.2
	}

	// Official sources boost quality
	hasOfficialSources := false
	for _, source := range validationResult.SourcesConsulted {
		if strings.Contains(source, "nixos") || strings.Contains(source, "official") {
			hasOfficialSources = true
			break
		}
	}

	if hasOfficialSources {
		score += 0.2
	}

	if score > 1 {
		score = 1
	}

	return score
}

func (fc *FactChecker) calculateRecencyScore(validationResult *EnhancedValidationResult) float64 {
	// This would ideally check timestamps of sources
	// For now, return a moderate score
	return 0.7
}

func (fc *FactChecker) calculateConsistencyScore(result *FactCheckResult) float64 {
	if len(result.FactualErrors) == 0 {
		return 1.0
	}

	// Reduce score based on contradictions and errors
	totalIssues := len(result.FactualErrors)
	score := 1.0 - (float64(totalIssues) * 0.2)

	if score < 0 {
		return 0
	}
	return score
}

func (fc *FactChecker) generateFactCheckRecommendations(result *FactCheckResult) []string {
	var recommendations []string

	if len(result.FactualErrors) > 0 {
		recommendations = append(recommendations, "Review and correct identified factual errors")
	}

	if len(result.UnverifiedFacts) > 0 {
		recommendations = append(recommendations, "Verify unconfirmed facts against official documentation")
	}

	if result.OverallAccuracy < 0.7 {
		recommendations = append(recommendations, "Consider consulting additional authoritative sources")
	}

	if result.SourceQuality < 0.6 {
		recommendations = append(recommendations, "Use more authoritative and recent sources")
	}

	return recommendations
}
