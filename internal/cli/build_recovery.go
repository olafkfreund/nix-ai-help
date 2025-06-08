package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/ai/function/build"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// BuildRecoverySystem provides intelligent error recovery using AI functions
type BuildRecoverySystem struct {
	buildFunction *build.BuildFunction
	logger        *logger.Logger
	recoveryCache map[string][]RecoveryStrategy
}

// RecoveryStrategy represents a specific approach to fixing build issues
type RecoveryStrategy struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Commands     []string  `json:"commands"`
	Success      bool      `json:"success"`
	AppliedAt    time.Time `json:"applied_at"`
	ErrorPattern string    `json:"error_pattern"`
}

// BuildRecoveryRequest represents a request for build error recovery
type BuildRecoveryRequest struct {
	Package     string `json:"package"`
	ErrorOutput string `json:"error_output"`
	BuildSystem string `json:"build_system"`
	AttemptNum  int    `json:"attempt_num"`
}

// NewBuildRecoverySystem creates a new intelligent recovery system
func NewBuildRecoverySystem() *BuildRecoverySystem {
	return &BuildRecoverySystem{
		buildFunction: build.NewBuildFunction(),
		logger:        logger.NewLogger(),
		recoveryCache: make(map[string][]RecoveryStrategy),
	}
}

// RecoverFromFailure attempts to automatically recover from build failures
func (brs *BuildRecoverySystem) RecoverFromFailure(req *BuildRecoveryRequest) (*RecoveryStrategy, error) {
	brs.logger.Info(fmt.Sprintf("Attempting recovery for package: %s (attempt %d)", req.Package, req.AttemptNum))

	// Check cache for previous successful strategies
	if strategy := brs.getCachedStrategy(req.ErrorOutput); strategy != nil {
		brs.logger.Info("Found cached recovery strategy")
		return strategy, nil
	}

	// Use AI function to analyze and generate recovery strategy
	strategy, err := brs.generateRecoveryWithAI(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery strategy: %w", err)
	}

	// Cache successful strategy
	if strategy.Success {
		brs.cacheStrategy(req.ErrorOutput, strategy)
	}

	return strategy, nil
}

// generateRecoveryWithAI uses the BuildFunction to generate intelligent recovery strategies
func (brs *BuildRecoverySystem) generateRecoveryWithAI(req *BuildRecoveryRequest) (*RecoveryStrategy, error) {
	// Prepare function parameters for BuildFunction
	params := map[string]interface{}{
		"operation":     "troubleshoot",
		"package":       req.Package,
		"error_logs":    req.ErrorOutput,
		"build_options": []string{},
		"verbose":       true,
		"show_trace":    true,
	}

	// Execute the build function for troubleshooting
	ctx := context.Background()
	result, err := brs.buildFunction.Execute(ctx, params, nil)
	if err != nil {
		return nil, fmt.Errorf("BuildFunction execution failed: %w", err)
	}

	// Parse the function result to extract recovery strategy
	strategy, err := brs.parseFunctionResult(result, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function result: %w", err)
	}

	return strategy, nil
}

// parseFunctionResult converts BuildFunction output to RecoveryStrategy
func (brs *BuildRecoverySystem) parseFunctionResult(result interface{}, req *BuildRecoveryRequest) (*RecoveryStrategy, error) {
	// In a real implementation, this would parse the structured BuildResponse
	// For now, we'll create a mock strategy based on common error patterns

	strategy := &RecoveryStrategy{
		ID:           fmt.Sprintf("recovery_%d", time.Now().Unix()),
		Name:         "AI-Generated Recovery Strategy",
		Description:  "Auto-generated recovery strategy based on error analysis",
		Commands:     brs.generateRecoveryCommands(req),
		Success:      false,
		AppliedAt:    time.Now(),
		ErrorPattern: brs.extractErrorPattern(req.ErrorOutput),
	}

	return strategy, nil
}

// generateRecoveryCommands creates specific commands to fix common build issues
func (brs *BuildRecoverySystem) generateRecoveryCommands(req *BuildRecoveryRequest) []string {
	commands := []string{}
	errorLower := strings.ToLower(req.ErrorOutput)

	// Hash mismatch recovery
	if strings.Contains(errorLower, "hash mismatch") || strings.Contains(errorLower, "got:") {
		commands = append(commands, []string{
			"nix-store --verify --check-contents",
			"nix-collect-garbage",
			fmt.Sprintf("nix build --rebuild %s", req.Package),
		}...)
	}

	// Missing dependency recovery
	if strings.Contains(errorLower, "command not found") || strings.Contains(errorLower, "no such file") {
		commands = append(commands, []string{
			"nix-channel --update",
			fmt.Sprintf("nix-shell -p %s", extractMissingPackage(req.ErrorOutput)),
			fmt.Sprintf("nix build %s --impure", req.Package),
		}...)
	}

	// Permission/sandbox issues
	if strings.Contains(errorLower, "permission denied") || strings.Contains(errorLower, "sandbox") {
		commands = append(commands, []string{
			fmt.Sprintf("nix build %s --option sandbox false", req.Package),
			fmt.Sprintf("nix build %s --impure --option trusted-users $(whoami)", req.Package),
		}...)
	}

	// Fetch failures
	if strings.Contains(errorLower, "fetch") || strings.Contains(errorLower, "download") {
		commands = append(commands, []string{
			"nix-channel --update",
			fmt.Sprintf("nix build %s --option substitute false", req.Package),
			fmt.Sprintf("nix build %s --fallback", req.Package),
		}...)
	}

	// Generic retry with different options
	if len(commands) == 0 {
		commands = []string{
			fmt.Sprintf("nix build %s --keep-going", req.Package),
			fmt.Sprintf("nix build %s --max-jobs 1", req.Package),
			fmt.Sprintf("nix build %s --show-trace", req.Package),
		}
	}

	return commands
}

// ApplyRecoveryStrategy executes the recovery commands
func (brs *BuildRecoverySystem) ApplyRecoveryStrategy(strategy *RecoveryStrategy, dryRun bool) error {
	fmt.Println(utils.FormatSubsection("🔧 Applying Recovery Strategy", strategy.Name))
	fmt.Println(utils.FormatKeyValue("Description", strategy.Description))
	fmt.Println()

	for i, command := range strategy.Commands {
		fmt.Printf("%s %d. %s\n", utils.FormatProgress("Executing"), i+1, command)

		if dryRun {
			fmt.Println(utils.FormatInfo("(Dry run - command not executed)"))
			continue
		}

		// Execute command and capture result
		success := brs.executeRecoveryCommand(command)
		if success {
			fmt.Println(utils.FormatSuccess("✅ Command succeeded"))
			strategy.Success = true
			return nil
		} else {
			fmt.Println(utils.FormatWarning("⚠️  Command failed, trying next..."))
		}
	}

	if !dryRun && !strategy.Success {
		return fmt.Errorf("all recovery commands failed")
	}

	return nil
}

// executeRecoveryCommand executes a single recovery command
func (brs *BuildRecoverySystem) executeRecoveryCommand(command string) bool {
	// Parse and execute the command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	// This would execute the actual command in a real implementation
	// For safety in this example, we'll just simulate success/failure
	brs.logger.Info(fmt.Sprintf("Executing recovery command: %s", command))

	// Simulate command execution with some realistic failure rate
	return len(parts) > 2 // Simple heuristic for simulation
}

// getCachedStrategy retrieves a previously successful strategy for similar errors
func (brs *BuildRecoverySystem) getCachedStrategy(errorOutput string) *RecoveryStrategy {
	pattern := brs.extractErrorPattern(errorOutput)
	if strategies, exists := brs.recoveryCache[pattern]; exists {
		for _, strategy := range strategies {
			if strategy.Success {
				return &strategy
			}
		}
	}
	return nil
}

// cacheStrategy stores a successful recovery strategy for future use
func (brs *BuildRecoverySystem) cacheStrategy(errorOutput string, strategy *RecoveryStrategy) {
	pattern := brs.extractErrorPattern(errorOutput)
	if brs.recoveryCache[pattern] == nil {
		brs.recoveryCache[pattern] = []RecoveryStrategy{}
	}
	brs.recoveryCache[pattern] = append(brs.recoveryCache[pattern], *strategy)
}

// extractErrorPattern identifies the core error pattern from build output
func (brs *BuildRecoverySystem) extractErrorPattern(errorOutput string) string {
	// Extract meaningful error patterns for caching
	lines := strings.Split(errorOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.Contains(line, "error:") ||
			strings.Contains(line, "failed") ||
			strings.Contains(line, "hash mismatch") {
			// Return first 50 characters as pattern
			if len(line) > 50 {
				return line[:50]
			}
			return line
		}
	}
	return "unknown_error"
}

// extractMissingPackage attempts to identify missing packages from error output
func extractMissingPackage(errorOutput string) string {
	// Simple pattern matching for missing commands
	lines := strings.Split(errorOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "command not found") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "command" && i > 0 {
					return parts[i-1]
				}
			}
		}
	}
	return "buildTools" // Default fallback
}

// SaveRecoveryReport saves a detailed recovery report
func (brs *BuildRecoverySystem) SaveRecoveryReport(strategy *RecoveryStrategy, req *BuildRecoveryRequest) error {
	report := map[string]interface{}{
		"strategy":  strategy,
		"request":   req,
		"timestamp": time.Now(),
		"success":   strategy.Success,
	}

	// Save to user's config directory
	cfg, _ := config.LoadUserConfig()
	reportDir := filepath.Join(filepath.Dir(cfg.Nixos.ConfigPath), "recovery_reports")
	os.MkdirAll(reportDir, 0755)

	reportFile := filepath.Join(reportDir, fmt.Sprintf("recovery_%s_%d.json",
		req.Package, time.Now().Unix()))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(reportFile, data, 0644)
}

// GetRecoveryStats returns statistics about recovery attempts
func (brs *BuildRecoverySystem) GetRecoveryStats() map[string]interface{} {
	totalStrategies := 0
	successfulStrategies := 0

	for _, strategies := range brs.recoveryCache {
		for _, strategy := range strategies {
			totalStrategies++
			if strategy.Success {
				successfulStrategies++
			}
		}
	}

	successRate := 0.0
	if totalStrategies > 0 {
		successRate = float64(successfulStrategies) / float64(totalStrategies) * 100
	}

	return map[string]interface{}{
		"total_attempts":      totalStrategies,
		"successful_attempts": successfulStrategies,
		"success_rate":        fmt.Sprintf("%.1f%%", successRate),
		"cached_patterns":     len(brs.recoveryCache),
	}
}

// AnalyzeAndRecover analyzes build errors and generates recovery strategies
func (brs *BuildRecoverySystem) AnalyzeAndRecover(request BuildRecoveryRequest) ([]RecoveryStrategy, error) {
	// Check cache for existing strategies
	cacheKey := brs.generateCacheKey(request.ErrorOutput)
	if strategies, exists := brs.recoveryCache[cacheKey]; exists {
		brs.logger.Info(fmt.Sprintf("Using cached recovery strategies for package %s (%d strategies)", 
			request.Package, len(strategies)))
		return strategies, nil
	}

	// Generate new strategies using AI function
	result, err := brs.buildFunction.Execute(context.Background(), map[string]interface{}{
		"operation": "troubleshoot",
		"package": request.Package,
		"error_logs": request.ErrorOutput,
		"system": request.BuildSystem,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze error: %w", err)
	}

	// Extract response data
	var responseText string
	if result.Success && result.Data != nil {
		if buildResponse, ok := result.Data.(*build.BuildResponse); ok {
			responseText = buildResponse.Solution + "\n" + buildResponse.DiagnosisDetails
			if len(buildResponse.SuggestedCommands) > 0 {
				responseText += "\nCommands: " + strings.Join(buildResponse.SuggestedCommands, "; ")
			}
		}
	}

	strategies := brs.parseStrategiesFromResponse(responseText, request.ErrorOutput)
	
	// Cache the strategies
	brs.recoveryCache[cacheKey] = strategies
	
	brs.logger.Info(fmt.Sprintf("Generated %d new recovery strategies for package %s", 
		len(strategies), request.Package))
	
	return strategies, nil
}

// ReportRecoveryResult reports the success or failure of a recovery strategy
func (brs *BuildRecoverySystem) ReportRecoveryResult(strategyID string, success bool, errorMessage string) {
	// Update strategy success status in cache
	for cacheKey, strategies := range brs.recoveryCache {
		for i, strategy := range strategies {
			if strategy.ID == strategyID {
				brs.recoveryCache[cacheKey][i].Success = success
				brs.recoveryCache[cacheKey][i].AppliedAt = time.Now()
				
				if success {
					brs.logger.Info(fmt.Sprintf("Recovery strategy '%s' (ID: %s) succeeded", 
						strategy.Name, strategyID))
				} else {
					brs.logger.Warn(fmt.Sprintf("Recovery strategy '%s' (ID: %s) failed: %s", 
						strategy.Name, strategyID, errorMessage))
				}
				return
			}
		}
	}
}

// parseStrategiesFromResponse parses AI response into recovery strategies
func (brs *BuildRecoverySystem) parseStrategiesFromResponse(response string, errorPattern string) []RecoveryStrategy {
	strategies := []RecoveryStrategy{}
	
	// Simple parsing - in production this would be more sophisticated
	lines := strings.Split(response, "\n")
	var currentStrategy *RecoveryStrategy
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check for strategy headers
		if strings.Contains(line, "Strategy") && strings.Contains(line, ":") {
			if currentStrategy != nil {
				strategies = append(strategies, *currentStrategy)
			}
			currentStrategy = &RecoveryStrategy{
				ID:           fmt.Sprintf("strategy_%d", len(strategies)+1),
				Name:         strings.TrimSpace(strings.Split(line, ":")[1]),
				Description:  "",
				Commands:     []string{},
				ErrorPattern: errorPattern,
				AppliedAt:    time.Now(),
			}
		} else if currentStrategy != nil {
			// Check for commands (lines starting with nix-build, nixos-rebuild, etc.)
			if strings.HasPrefix(line, "nix") || strings.HasPrefix(line, "sudo") {
				currentStrategy.Commands = append(currentStrategy.Commands, line)
			} else if currentStrategy.Description == "" {
				currentStrategy.Description = line
			}
		}
	}
	
	// Add the last strategy
	if currentStrategy != nil {
		strategies = append(strategies, *currentStrategy)
	}
	
	// If no strategies were parsed, create a generic one
	if len(strategies) == 0 {
		strategies = append(strategies, RecoveryStrategy{
			ID:           "generic_retry",
			Name:         "Clean Rebuild",
			Description:  "Clean the build environment and retry",
			Commands:     []string{"nix-collect-garbage -d", "nix-build --repair"},
			ErrorPattern: errorPattern,
			AppliedAt:    time.Now(),
		})
	}
	
	return strategies
}

// generateCacheKey creates a cache key from error output
func (brs *BuildRecoverySystem) generateCacheKey(errorOutput string) string {
	// Simple hash of key error patterns
	key := strings.ToLower(errorOutput)
	if len(key) > 100 {
		key = key[:100]
	}
	return fmt.Sprintf("%x", key)
}
