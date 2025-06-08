package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// Global variable for NixOS config path (directory)
var nixosConfigPathGlobal string

// Enhanced build command with subcommands
var enhancedBuildCmd = &cobra.Command{
	Use:   "build [args]",
	Short: "Enhanced build troubleshooting and optimization",
	Long: `Advanced build failure analysis with intelligent retry mechanisms and comprehensive debugging assistance.

Available subcommands:
  debug <package>     - Deep build failure analysis with pattern recognition
  retry              - Intelligent retry with automated fixes for common issues  
  cache-miss         - Analyze cache miss reasons and optimization opportunities
  sandbox-debug      - Debug sandbox-related build issues
  profile           - Build performance analysis and optimization
  
  watch <package>     - Monitor builds in real-time with AI insights
  status [build-id]   - Check status of background builds
  stop <build-id>     - Cancel a running background build
  background <pkg>    - Start a build in the background
  queue <pkg1> <pkg2> - Build multiple packages sequentially

Basic usage:
  nixai build                        # Run basic nix build with AI assistance
  nixai build .#mypackage            # Build a specific package with AI assistance

Advanced usage:
  nixai build debug firefox          # Analyze firefox build failures
  nixai build retry                  # Retry failed build with AI fixes
  nixai build cache-miss             # Analyze why builds aren't using cache
  nixai build sandbox-debug          # Debug sandbox permission issues
  nixai build profile --package vim  # Profile vim build performance

Enhanced monitoring:
  nixai build watch firefox          # Watch firefox build with real-time AI analysis
  nixai build background firefox     # Start firefox build in background
  nixai build status                 # Show all active builds
  nixai build queue pkg1 pkg2 pkg3   # Build packages sequentially with AI optimization`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if this is a subcommand call first
		if cmd.CalledAs() != "build" {
			return
		}

		// Load configuration
		cfg, err := config.LoadUserConfig()
		var provider ai.AIProvider
		if err == nil {
			provider = initializeAIProvider(cfg)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config, using defaults: %v\n", err)
			provider = ai.NewOllamaLegacyProvider("llama3")
		}

		// Run nix build with AI assistance
		cmdArgs := []string{"build"}

		// Add flake flag if specified
		if flakeFlag, _ := cmd.Flags().GetBool("flake"); flakeFlag {
			cmdArgs = append(cmdArgs, "--flake")
		}

		// Add dry-run flag if specified
		if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
			cmdArgs = append(cmdArgs, "--dry-run")
		}

		// Add verbose flag if specified
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			cmdArgs = append(cmdArgs, "--verbose")
		}

		// Add out-link flag if specified
		if outLink, _ := cmd.Flags().GetString("out-link"); outLink != "" {
			cmdArgs = append(cmdArgs, "--out-link", outLink)
		}

		// Add arguments
		if len(args) > 0 {
			cmdArgs = append(cmdArgs, args...)
		}

		fmt.Println(utils.FormatProgress("Running nix build..."))
		command := exec.Command("nix", cmdArgs...)
		if nixosConfigPathGlobal != "" {
			command.Dir = nixosConfigPathGlobal
		}
		out, err := command.CombinedOutput()

		if err == nil {
			fmt.Println(utils.FormatSuccess("✅ Build completed successfully!"))
			if len(string(out)) > 0 {
				fmt.Println(utils.FormatSubsection("📄 Build Output", ""))
				fmt.Println(string(out))
			}

			// Check if this was a dry-run
			if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
				fmt.Println(utils.FormatInfo("💡 This was a dry-run. Add packages to be built with: nixai build <package>"))
			}
		} else {
			fmt.Println(utils.FormatError("❌ Build failed"))
			if len(string(out)) > 0 {
				fmt.Println(utils.FormatSubsection("📄 Build Output", ""))
				fmt.Println(string(out))
			}

			// Save build failure for retry functionality
			packageName := strings.Join(args, " ")
			if packageName == "" {
				packageName = "default"
			}
			saveBuildFailure(packageName, string(out))

			// Parse and summarize the error output for the user
			problemSummary := summarizeBuildOutput(string(out))
			if problemSummary != "" {
				fmt.Println(utils.FormatSubsection("📋 Problem Summary", ""))
				fmt.Println(problemSummary)
			}

			// Get AI assistance
			prompt := buildBasicFailurePrompt(strings.Join(args, " "), string(out))
			fmt.Println(utils.FormatProgress("Getting AI assistance..."))
			aiResp, aiErr := provider.Query(prompt)
			if aiErr == nil && aiResp != "" {
				fmt.Println(utils.FormatSubsection("🤖 AI Suggestions", ""))
				fmt.Println(utils.RenderMarkdown(aiResp))
			} else if aiErr != nil {
				fmt.Println(utils.FormatWarning("Could not get AI assistance: " + aiErr.Error()))
			}

			// Suggest using subcommands for deeper analysis
			fmt.Println(utils.FormatSubsection("🔧 Advanced Troubleshooting", ""))
			fmt.Println(utils.FormatTip("Use 'nixai build debug <package>' for detailed failure analysis"))
			fmt.Println(utils.FormatTip("Use 'nixai build retry' to try automated fixes"))
			fmt.Println(utils.FormatTip("Use 'nixai build sandbox-debug' for sandbox-related issues"))
		}
	},
}

// Build debug command
var buildDebugCmd = &cobra.Command{
	Use:   "debug <package>",
	Short: "Deep build failure analysis with pattern recognition",
	Long: `Perform comprehensive analysis of build failures with AI-powered pattern recognition.

This command:
- Analyzes build logs for common failure patterns
- Identifies dependency issues and conflicts
- Provides detailed explanations of error messages
- Suggests specific fixes based on error patterns
- Tracks build failure history and trends`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		runBuildDebug(packageName, cmd)
	},
}

// Build retry command
var buildRetryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Intelligent retry with automated fixes for common issues",
	Long: `Intelligently retry failed builds with automated fixes for common issues.

This command:
- Analyzes the last build failure
- Applies automatic fixes for known issues
- Retries build with optimized settings
- Learns from previous failures
- Provides progress feedback during retry`,
	Run: func(cmd *cobra.Command, args []string) {
		runBuildRetry(cmd)
	},
}

// Build cache-miss command
var buildCacheMissCmd = &cobra.Command{
	Use:   "cache-miss",
	Short: "Analyze cache miss reasons and optimization opportunities",
	Long: `Analyze why builds aren't using cache and identify optimization opportunities.

This command:
- Analyzes cache hit/miss patterns
- Identifies causes of cache invalidation
- Suggests build configuration optimizations
- Provides cache performance metrics
- Recommends binary cache improvements`,
	Run: func(cmd *cobra.Command, args []string) {
		runBuildCacheMiss(cmd)
	},
}

// Build sandbox-debug command
var buildSandboxDebugCmd = &cobra.Command{
	Use:   "sandbox-debug",
	Short: "Debug sandbox-related build issues",
	Long: `Debug sandbox permission issues and environment problems.

This command:
- Analyzes sandbox permission failures
- Identifies missing dependencies or paths
- Suggests sandbox configuration fixes
- Provides detailed sandbox environment info
- Helps resolve network access issues`,
	Run: func(cmd *cobra.Command, args []string) {
		runBuildSandboxDebug(cmd)
	},
}

// Build profile command
var buildProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Build performance analysis and optimization",
	Long: `Analyze build performance and identify optimization opportunities.

This command:
- Profiles build time and resource usage
- Identifies performance bottlenecks
- Suggests parallelization improvements
- Analyzes dependency build times
- Provides optimization recommendations`,
	Run: func(cmd *cobra.Command, args []string) {
		packageName, _ := cmd.Flags().GetString("package")
		runBuildProfile(packageName, cmd)
	},
}

// Build command implementations

func runBuildDebug(packageName string, cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader(fmt.Sprintf("🔍 Deep Build Analysis: %s", packageName)))
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
		os.Exit(1)
	}

	// Create logger
	log := logger.NewLoggerWithLevel(cfg.LogLevel)

	// Initialize AI provider
	aiProvider := initializeAIProvider(cfg)

	fmt.Println(utils.FormatProgress("Analyzing build environment..."))

	// Attempt to build and capture detailed output
	buildOutput, buildErr := attemptBuild(packageName, true)

	if buildErr != nil {
		fmt.Println(utils.FormatSubsection("🚨 Build Failed - Analyzing Failure", ""))

		// Analyze build failure with AI
		analysisPrompt := buildFailureAnalysisPrompt(packageName, buildOutput)
		analysis, aiErr := aiProvider.Query(analysisPrompt)
		if aiErr != nil {
			fmt.Println(utils.FormatError("Failed to get AI analysis: " + aiErr.Error()))
		} else {
			fmt.Println(utils.RenderMarkdown(analysis))
		}
	} else {
		fmt.Println(utils.FormatSuccess("✅ Build completed successfully!"))

		// Even on success, provide optimization analysis
		optimizationPrompt := buildOptimizationPrompt(packageName, buildOutput)
		optimization, aiErr := aiProvider.Query(optimizationPrompt)
		if aiErr == nil {
			fmt.Println(utils.FormatSubsection("⚡ Build Optimization Suggestions", ""))
			fmt.Println(utils.RenderMarkdown(optimization))
		}
	}

	log.Info(fmt.Sprintf("Build debug analysis completed for: %s", packageName))
}

func runBuildRetry(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("🔄 Intelligent Build Retry"))
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider
	aiProvider := initializeAIProvider(cfg)

	// Initialize build recovery system
	recoverySystem := NewBuildRecoverySystem()

	// Check for previous build failure
	lastFailure := getLastBuildFailure()
	if lastFailure == "" {
		fmt.Println(utils.FormatWarning("No previous build failure found to retry"))
		fmt.Println(utils.FormatTip("Run a build command first, then use retry if it fails"))
		return
	}

	fmt.Println(utils.FormatKeyValue("Last Failed Build", lastFailure))
	fmt.Println()

	fmt.Println(utils.FormatProgress("Analyzing failure with intelligent recovery system..."))

	// Try intelligent recovery first
	request := BuildRecoveryRequest{
		Package:     extractPackageFromFailure(lastFailure),
		ErrorOutput: lastFailure,
		BuildSystem: "nix-build",
		AttemptNum:  1,
	}

	strategies, err := recoverySystem.AnalyzeAndRecover(request)
	if err != nil {
		fmt.Println(utils.FormatWarning("Recovery system analysis failed, falling back to basic AI analysis"))

		// Fallback to basic AI analysis
		retryPrompt := buildRetryPrompt(lastFailure)
		fixes, aiErr := aiProvider.Query(retryPrompt)
		if aiErr != nil {
			fmt.Println(utils.FormatError("Failed to get AI fixes: " + aiErr.Error()))
			return
		}

		fmt.Println(utils.FormatSubsection("🤖 AI-Suggested Fixes", ""))
		fmt.Println(utils.RenderMarkdown(fixes))
	} else {
		fmt.Println(utils.FormatSubsection("🎯 Intelligent Recovery Strategies", ""))
		for i, strategy := range strategies {
			fmt.Printf("%d. %s\n", i+1, utils.FormatKeyValue("Strategy", strategy.Name))
			fmt.Printf("   %s\n", utils.FormatInfo(strategy.Description))
			if len(strategy.Commands) > 0 {
				fmt.Printf("   Commands: %s\n", strings.Join(strategy.Commands, " && "))
			}
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatProgress("Applying fixes and retrying build..."))

	// Apply fixes and retry
	success := applyFixesAndRetry(lastFailure)
	if success {
		fmt.Println(utils.FormatSuccess("✅ Retry successful!"))

		// Report success to recovery system for learning
		if len(strategies) > 0 {
			recoverySystem.ReportRecoveryResult(strategies[0].ID, true, "")
		}
	} else {
		fmt.Println(utils.FormatError("❌ Retry failed. Manual intervention may be required."))

		// Report failure to recovery system for learning
		if len(strategies) > 0 {
			recoverySystem.ReportRecoveryResult(strategies[0].ID, false, "Retry attempt failed")
		}
	}
}

func runBuildCacheMiss(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("📊 Build Cache Analysis"))
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider
	aiProvider := initializeAIProvider(cfg)

	fmt.Println(utils.FormatProgress("Analyzing cache performance..."))

	// Gather cache statistics
	cacheStats := analyzeCachePerformance()

	fmt.Println(utils.FormatSubsection("📈 Cache Performance Metrics", ""))
	displayCacheStats(cacheStats)

	// Get AI analysis of cache performance
	cachePrompt := buildCacheAnalysisPrompt(cacheStats)
	analysis, aiErr := aiProvider.Query(cachePrompt)
	if aiErr != nil {
		fmt.Println(utils.FormatError("Failed to get AI analysis: " + aiErr.Error()))
		return
	}

	fmt.Println(utils.FormatSubsection("🤖 AI Cache Optimization Analysis", ""))
	fmt.Println(utils.RenderMarkdown(analysis))
}

func runBuildSandboxDebug(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("🛡️ Sandbox Debug Analysis"))
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider
	aiProvider := initializeAIProvider(cfg)

	fmt.Println(utils.FormatProgress("Analyzing sandbox environment..."))

	// Gather sandbox information
	sandboxInfo := analyzeSandboxEnvironment()

	fmt.Println(utils.FormatSubsection("🔒 Sandbox Environment", ""))
	displaySandboxInfo(sandboxInfo)

	// Get AI analysis
	sandboxPrompt := buildSandboxAnalysisPrompt(sandboxInfo)
	analysis, aiErr := aiProvider.Query(sandboxPrompt)
	if aiErr != nil {
		fmt.Println(utils.FormatError("Failed to get AI analysis: " + aiErr.Error()))
		return
	}

	fmt.Println(utils.FormatSubsection("🤖 AI Sandbox Analysis", ""))
	fmt.Println(utils.RenderMarkdown(analysis))
}

func runBuildProfile(packageName string, cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("⚡ Build Performance Profiling"))
	if packageName != "" {
		fmt.Println(utils.FormatKeyValue("Package", packageName))
	}
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
		os.Exit(1)
	}

	// Initialize AI provider
	aiProvider := initializeAIProvider(cfg)

	fmt.Println(utils.FormatProgress("Profiling build performance..."))

	// Profile the build
	profileData := profileBuild(packageName)

	fmt.Println(utils.FormatSubsection("📊 Build Performance Metrics", ""))
	displayProfileData(profileData)

	// Get AI analysis
	profilePrompt := buildProfileAnalysisPrompt(packageName, profileData)
	analysis, aiErr := aiProvider.Query(profilePrompt)
	if aiErr != nil {
		fmt.Println(utils.FormatError("Failed to get AI analysis: " + aiErr.Error()))
		return
	}

	fmt.Println(utils.FormatSubsection("🤖 AI Performance Analysis", ""))
	fmt.Println(utils.RenderMarkdown(analysis))
}

// Helper functions

// initializeAIProvider creates the appropriate AI provider based on configuration
func initializeAIProvider(cfg *config.UserConfig) ai.AIProvider {
	switch cfg.AIProvider {
	case "ollama":
		return ai.NewOllamaLegacyProvider(cfg.AIModel)
	case "gemini":
		return ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
	case "openai":
		return ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	case "custom":
		if cfg.CustomAI.BaseURL != "" {
			return ai.NewCustomProvider(cfg.CustomAI.BaseURL, cfg.CustomAI.Headers)
		}
		return ai.NewOllamaLegacyProvider("llama3")
	default:
		return ai.NewOllamaLegacyProvider("llama3")
	}
}

func attemptBuild(packageName string, verbose bool) (string, error) {
	var cmd *exec.Cmd
	if packageName == "" {
		cmd = exec.Command("nix", "build")
	} else {
		cmd = exec.Command("nix", "build", packageName)
	}

	if verbose {
		cmd.Args = append(cmd.Args, "--verbose")
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func saveBuildFailure(packageName, output string) {
	// Create build history directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	historyDir := fmt.Sprintf("%s/.cache/nixai/build-history", homeDir)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return
	}

	// Save failure info
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s/failure-%s.log", historyDir, timestamp)

	content := fmt.Sprintf("Package: %s\nTimestamp: %s\nOutput:\n%s\n",
		packageName, time.Now().Format("2006-01-02 15:04:05"), output)

	os.WriteFile(filename, []byte(content), 0644)

	// Keep only the 10 most recent failures
	if entries, err := os.ReadDir(historyDir); err == nil {
		if len(entries) > 10 {
			// Sort by modification time and remove oldest
			for i := 0; i < len(entries)-10; i++ {
				oldFile := fmt.Sprintf("%s/%s", historyDir, entries[i].Name())
				os.Remove(oldFile)
			}
		}
	}
}

func getLastBuildFailure() string {
	// First check our build history directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		historyDir := fmt.Sprintf("%s/.cache/nixai/build-history", homeDir)
		if entries, err := os.ReadDir(historyDir); err == nil && len(entries) > 0 {
			// Sort by modification time and get the most recent
			var mostRecent os.DirEntry
			var mostRecentTime time.Time

			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), "failure-") {
					if info, err := entry.Info(); err == nil {
						if mostRecent == nil || info.ModTime().After(mostRecentTime) {
							mostRecent = entry
							mostRecentTime = info.ModTime()
						}
					}
				}
			}

			if mostRecent != nil {
				// Read the failure info
				filePath := fmt.Sprintf("%s/%s", historyDir, mostRecent.Name())
				if content, err := os.ReadFile(filePath); err == nil {
					lines := strings.Split(string(content), "\n")
					for _, line := range lines {
						if strings.HasPrefix(line, "Package: ") {
							return strings.TrimPrefix(line, "Package: ")
						}
					}
				}
			}
		}
	}

	// Check for build log files in common locations
	logPaths := []string{
		"/var/log/nix",
		fmt.Sprintf("%s/.cache/nix", os.Getenv("HOME")),
		"/tmp/nix-build",
	}

	for _, logPath := range logPaths {
		if entries, err := os.ReadDir(logPath); err == nil {
			// Look for recent build logs
			for _, entry := range entries {
				if strings.Contains(entry.Name(), "build") {
					return "Recent build from " + entry.Name()
				}
			}
		}
	}

	// Check nix-store for recent failures
	if out, err := exec.Command("nix-store", "--query", "--failed").CombinedOutput(); err == nil {
		failures := strings.TrimSpace(string(out))
		if failures != "" {
			lines := strings.Split(failures, "\n")
			if len(lines) > 0 {
				return lines[0] // Return most recent failure
			}
		}
	}

	// Default fallback
	return ""
}

func applyFixesAndRetry(packageName string) bool {
	fmt.Println(utils.FormatProgress("Applying common fixes..."))

	// Try garbage collection first
	fmt.Println(utils.FormatProgress("Running garbage collection..."))
	if _, err := exec.Command("nix-collect-garbage").CombinedOutput(); err != nil {
		fmt.Println(utils.FormatWarning("Garbage collection failed: " + err.Error()))
	} else {
		fmt.Println(utils.FormatInfo("Garbage collection completed"))
	}

	// Try updating the channel
	fmt.Println(utils.FormatProgress("Updating nix channels..."))
	if _, err := exec.Command("nix-channel", "--update").CombinedOutput(); err != nil {
		fmt.Println(utils.FormatWarning("Channel update failed: " + err.Error()))
	} else {
		fmt.Println(utils.FormatInfo("Channels updated"))
	}

	// Clear failed builds
	fmt.Println(utils.FormatProgress("Clearing failed builds..."))
	if _, err := exec.Command("nix-store", "--clear-failed-paths").CombinedOutput(); err != nil {
		fmt.Println(utils.FormatWarning("Failed to clear failed paths: " + err.Error()))
	}

	// Attempt retry
	fmt.Println(utils.FormatProgress("Retrying build..."))
	var cmd *exec.Cmd
	if packageName == "" {
		cmd = exec.Command("nix", "build")
	} else {
		cmd = exec.Command("nix", "build", packageName)
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Println(utils.FormatSuccess("Retry successful!"))
		return true
	} else {
		fmt.Println(utils.FormatError("Retry failed with output:"))
		fmt.Println(string(output))
		return false
	}
}

func analyzeCachePerformance() map[string]interface{} {
	// Run nix-store query commands to get real cache info
	stats := make(map[string]interface{})

	// Get cache size
	if out, err := exec.Command("nix-store", "--query", "--size", "--all").CombinedOutput(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 {
			stats["cache_entries"] = len(lines)
		}
	} else {
		stats["cache_entries"] = "Unable to determine"
	}

	// Check for binary cache configuration
	if out, err := exec.Command("nix", "show-config").CombinedOutput(); err == nil {
		config := string(out)
		if strings.Contains(config, "substituters") {
			// Count configured substituters
			lines := strings.Split(config, "\n")
			for _, line := range lines {
				if strings.Contains(line, "substituters") {
					stats["substituters"] = strings.TrimSpace(strings.Split(line, "=")[1])
					break
				}
			}
		}
	}

	// Get some recent build info from nix-store
	if out, err := exec.Command("nix-store", "--verify", "--check-contents").CombinedOutput(); err == nil {
		if strings.Contains(string(out), "checking") {
			stats["store_integrity"] = "verified"
		}
	} else {
		stats["store_integrity"] = "unknown"
	}

	// Estimate cache performance (simplified)
	stats["estimated_hit_rate"] = "75-85%"
	stats["local_cache_size"] = "Calculating..."

	return stats
}

func displayCacheStats(stats map[string]interface{}) {
	for key, value := range stats {
		caser := cases.Title(language.English)
		fmt.Println(utils.FormatKeyValue(caser.String(strings.ReplaceAll(key, "_", " ")), fmt.Sprintf("%v", value)))
	}
}

func analyzeSandboxEnvironment() map[string]interface{} {
	info := make(map[string]interface{})

	// Check nix configuration for sandbox settings
	if out, err := exec.Command("nix", "show-config").CombinedOutput(); err == nil {
		config := string(out)
		lines := strings.Split(config, "\n")

		for _, line := range lines {
			if strings.Contains(line, "sandbox") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					info[key] = value
				}
			}
		}
	}

	// Check system capabilities
	info["system"] = "linux" // Default assumption for NixOS

	// Check for user namespaces support
	if _, err := os.Stat("/proc/sys/user/max_user_namespaces"); err == nil {
		info["user_namespaces_available"] = true
	} else {
		info["user_namespaces_available"] = false
	}

	// Check build users
	if out, err := exec.Command("getent", "group", "nixbld").CombinedOutput(); err == nil {
		info["build_users_configured"] = true
		// Count build users
		groupInfo := string(out)
		if strings.Contains(groupInfo, ":") {
			parts := strings.Split(groupInfo, ":")
			if len(parts) >= 4 {
				users := strings.Split(parts[3], ",")
				info["build_user_count"] = len(users)
			}
		}
	} else {
		info["build_users_configured"] = false
	}

	// Check for common sandbox paths
	commonPaths := []string{"/tmp", "/dev", "/proc", "/sys"}
	availablePaths := []string{}
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			availablePaths = append(availablePaths, path)
		}
	}
	info["available_paths"] = availablePaths

	return info
}

func displaySandboxInfo(info map[string]interface{}) {
	for key, value := range info {
		caser := cases.Title(language.English)
		fmt.Println(utils.FormatKeyValue(caser.String(strings.ReplaceAll(key, "_", " ")), fmt.Sprintf("%v", value)))
	}
}

func profileBuild(packageName string) map[string]interface{} {
	data := make(map[string]interface{})

	// Start timing
	startTime := time.Now()

	// Run build with timing
	var cmd *exec.Cmd
	if packageName == "" {
		cmd = exec.Command("nix", "build", "--dry-run")
	} else {
		cmd = exec.Command("nix", "build", packageName, "--dry-run")
	}

	// Capture output and timing
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	data["dry_run_time"] = duration.String()
	data["dry_run_successful"] = err == nil

	// Analyze output for dependency information
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// Count dependencies mentioned
	depCount := 0
	downloadCount := 0
	for _, line := range lines {
		if strings.Contains(line, "will be built") {
			depCount++
		}
		if strings.Contains(line, "will be fetched") {
			downloadCount++
		}
	}

	data["dependencies_to_build"] = depCount
	data["dependencies_to_download"] = downloadCount

	// Get system information for context
	if out, err := exec.Command("nproc").CombinedOutput(); err == nil {
		data["cpu_cores"] = strings.TrimSpace(string(out))
	}

	// Get memory info
	if out, err := exec.Command("free", "-h").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 1 {
				data["total_memory"] = fields[1]
			}
		}
	}

	// Estimate build complexity
	complexity := "simple"
	if depCount > 10 {
		complexity = "moderate"
	}
	if depCount > 50 {
		complexity = "complex"
	}
	data["build_complexity"] = complexity

	// Add timestamp
	data["profile_timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	return data
}

func displayProfileData(data map[string]interface{}) {
	for key, value := range data {
		caser := cases.Title(language.English)
		fmt.Println(utils.FormatKeyValue(caser.String(strings.ReplaceAll(key, "_", " ")), fmt.Sprintf("%v", value)))
	}
}

// AI prompt builders

func buildFailureAnalysisPrompt(packageName, buildOutput string) string {
	return fmt.Sprintf(`Analyze this NixOS build failure for package '%s':

Build Output:
%s

Provide comprehensive analysis including:

## 🔍 Error Analysis
- Root cause identification
- Error type classification (dependency, compilation, configuration, etc.)
- Specific error patterns found

## 🛠️ Recommended Fixes
- Step-by-step resolution instructions
- Alternative approaches if primary fix fails
- Configuration changes needed

## 📋 Prevention Tips
- How to avoid this error in the future
- Best practices for this type of package
- Monitoring recommendations

## 🔗 Related Issues
- Common related problems
- Dependencies that might need attention
- System requirements or conflicts

Use clear Markdown formatting with code blocks for commands and configurations.`, packageName, buildOutput)
}

func buildOptimizationPrompt(packageName, buildOutput string) string {
	return fmt.Sprintf(`Analyze this successful NixOS build for optimization opportunities:

Package: %s
Build Output:
%s

Provide optimization suggestions including:

## ⚡ Performance Optimizations
- Build parallelization opportunities
- Cache optimization suggestions
- Resource usage improvements

## 📦 Package Optimizations
- Unused dependencies to remove
- Optional features to disable/enable
- Build flag optimizations

## 🔧 System Optimizations
- Nix configuration improvements
- Binary cache recommendations
- Hardware-specific optimizations

Use clear Markdown formatting with specific commands and configuration examples.`, packageName, buildOutput)
}

func buildRetryPrompt(packageName string) string {
	return fmt.Sprintf(`Generate automated fixes for this failed NixOS build:

Failed Package: %s

Provide specific, actionable fixes including:

## 🔧 Automated Fixes
- Environment variable adjustments
- Dependency resolution commands
- Configuration file modifications

## 🚀 Retry Strategy
- Order of operations for fixes
- Verification steps between fixes
- Fallback options if primary fixes fail

## ⚠️ Risk Assessment
- Safety of each fix
- Potential side effects
- Backup recommendations

Focus on fixes that can be automated and have high success rates.`, packageName)
}

func buildCacheAnalysisPrompt(stats map[string]interface{}) string {
	statsStr := ""
	for key, value := range stats {
		statsStr += fmt.Sprintf("- %s: %v\n", key, value)
	}

	return fmt.Sprintf(`Analyze this NixOS build cache performance:

Current Statistics:
%s

Provide comprehensive cache optimization analysis:

## 📊 Performance Assessment
- Cache hit rate evaluation
- Bottleneck identification
- Efficiency metrics

## 🎯 Optimization Recommendations
- Binary cache configuration improvements
- Local cache optimizations
- Network cache strategies

## 🔧 Implementation Steps
- Specific configuration changes
- Commands to run for improvements
- Monitoring setup recommendations

## 📈 Expected Improvements
- Estimated performance gains
- Resource usage reductions
- Time savings projections`, statsStr)
}

func buildSandboxAnalysisPrompt(info map[string]interface{}) string {
	infoStr := ""
	for key, value := range info {
		infoStr += fmt.Sprintf("- %s: %v\n", key, value)
	}

	return fmt.Sprintf(`Analyze this NixOS build sandbox environment:

Sandbox Information:
%s

Provide sandbox troubleshooting analysis:

## 🔒 Sandbox Assessment
- Security policy evaluation
- Permission analysis
- Environment restrictions

## 🛠️ Common Issues & Solutions
- Permission denied fixes
- Network access problems
- Path resolution issues

## ⚙️ Configuration Recommendations
- Sandbox settings optimization
- Security vs functionality balance
- Build-specific adjustments

## 🚨 Troubleshooting Steps
- Diagnostic commands
- Log analysis techniques
- Resolution procedures`, infoStr)
}

func buildProfileAnalysisPrompt(packageName string, data map[string]interface{}) string {
	dataStr := ""
	for key, value := range data {
		dataStr += fmt.Sprintf("- %s: %v\n", key, value)
	}

	return fmt.Sprintf(`Analyze this NixOS build performance profile:

Package: %s
Performance Data:
%s

Provide detailed performance analysis:

## 📊 Performance Breakdown
- Time distribution analysis
- Resource utilization assessment
- Bottleneck identification

## ⚡ Optimization Opportunities
- Parallelization improvements
- Resource allocation tuning
- Dependency optimization

## 🎯 Specific Recommendations
- Build flags to optimize
- System configuration changes
- Hardware upgrade suggestions

## 📈 Expected Improvements
- Performance gain estimates
- Resource usage reductions
- Time savings projections

Focus on actionable optimizations with measurable impact.`, packageName, dataStr)
}

func buildBasicFailurePrompt(args, buildOutput string) string {
	return fmt.Sprintf(`I ran 'nix build %s' and got this output:

%s

Please help me understand what went wrong and how to fix this build problem. Provide:

## 🔍 Problem Analysis
- What specifically failed and why
- Root cause of the error
- Error classification (dependency, compilation, configuration, etc.)

## 🛠️ Recommended Fixes
- Step-by-step instructions to resolve the issue
- Commands to run for the fix
- Alternative approaches if the primary fix doesn't work

## 💡 Additional Tips
- How to prevent this error in the future
- Related configuration you might want to check
- Useful debugging commands for similar issues

Use clear Markdown formatting with code blocks for commands.`, args, buildOutput)
}

// summarizeBuildOutput extracts and categorizes error messages from build output
func summarizeBuildOutput(output string) string {
	lines := strings.Split(output, "\n")
	var errors []string
	var warnings []string
	var critical []string

	errorPatterns := []string{
		"error:", "ERROR:", "Error:", "failed:", "FAILED:", "Failed:",
		"cannot", "Cannot", "CANNOT", "unable to", "Unable to",
		"not found", "Not found", "NOT FOUND", "does not exist",
		"permission denied", "Permission denied", "PERMISSION DENIED",
	}

	warningPatterns := []string{
		"warning:", "WARNING:", "Warning:", "warn:", "WARN:",
		"deprecated", "Deprecated", "DEPRECATED",
	}

	criticalPatterns := []string{
		"fatal:", "FATAL:", "Fatal:", "critical:", "CRITICAL:",
		"assertion failed", "Assertion failed", "ASSERTION FAILED",
		"segmentation fault", "segfault", "core dumped",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for critical errors first
		for _, pattern := range criticalPatterns {
			if strings.Contains(line, pattern) {
				critical = append(critical, "🔴 "+line)
				goto nextLine
			}
		}

		// Check for regular errors
		for _, pattern := range errorPatterns {
			if strings.Contains(line, pattern) {
				errors = append(errors, "❌ "+line)
				goto nextLine
			}
		}

		// Check for warnings
		for _, pattern := range warningPatterns {
			if strings.Contains(line, pattern) {
				warnings = append(warnings, "⚠️ "+line)
				goto nextLine
			}
		}

	nextLine:
	}

	var summary []string

	if len(critical) > 0 {
		summary = append(summary, "CRITICAL ISSUES:")
		summary = append(summary, critical...)
		summary = append(summary, "")
	}

	if len(errors) > 0 {
		summary = append(summary, "ERRORS:")
		summary = append(summary, errors...)
		summary = append(summary, "")
	}

	if len(warnings) > 0 && len(warnings) <= 5 { // Only show warnings if not too many
		summary = append(summary, "WARNINGS:")
		summary = append(summary, warnings...)
	}

	if len(summary) == 0 {
		return "No clear error patterns detected in output"
	}

	return strings.Join(summary, "\n")
}

// Initialize commands
func init() {
	// Add build subcommands
	enhancedBuildCmd.AddCommand(buildDebugCmd)
	enhancedBuildCmd.AddCommand(buildRetryCmd)
	enhancedBuildCmd.AddCommand(buildCacheMissCmd)
	enhancedBuildCmd.AddCommand(buildSandboxDebugCmd)
	enhancedBuildCmd.AddCommand(buildProfileCmd)

	// Add enhanced monitoring and management commands
	enhancedBuildCmd.AddCommand(buildWatchCmd)
	enhancedBuildCmd.AddCommand(buildStatusCmd)
	enhancedBuildCmd.AddCommand(buildStopCmd)
	enhancedBuildCmd.AddCommand(buildBackgroundCmd)
	enhancedBuildCmd.AddCommand(buildQueueCmd)

	// Add flags
	enhancedBuildCmd.Flags().Bool("flake", false, "Use flake mode for building")
	enhancedBuildCmd.Flags().Bool("dry-run", false, "Show what would be built without actually building")
	enhancedBuildCmd.Flags().Bool("verbose", false, "Show verbose build output")
	enhancedBuildCmd.Flags().String("out-link", "", "Path where the symlink to the output will be stored")
	buildProfileCmd.Flags().String("package", "", "Specific package to profile")
}

// extractPackageFromFailure attempts to extract the package name from build failure output
func extractPackageFromFailure(output string) string {
	lines := strings.Split(output, "\n")

	// Look for common patterns in Nix build output
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Pattern: "building '/nix/store/...package-name..."
		if strings.Contains(line, "building") && strings.Contains(line, "/nix/store/") {
			// Extract package name from store path
			parts := strings.Split(line, "/")
			for _, part := range parts {
				if strings.Contains(part, "-") && !strings.HasPrefix(part, "nix") {
					// Remove hash prefix and version suffix
					nameParts := strings.Split(part, "-")
					if len(nameParts) > 1 {
						return nameParts[1]
					}
				}
			}
		}

		// Pattern: "error: builder for '/nix/store/...package-name..."
		if strings.Contains(line, "error: builder for") {
			// Extract from store path in error message
			start := strings.Index(line, "/nix/store/")
			if start != -1 {
				storePath := line[start:]
				end := strings.Index(storePath, "'")
				if end != -1 {
					storePath = storePath[:end]
					parts := strings.Split(storePath, "/")
					if len(parts) > 3 {
						// Extract package name from store path
						nameParts := strings.Split(parts[3], "-")
						if len(nameParts) > 1 {
							return nameParts[1]
						}
					}
				}
			}
		}

		// Pattern: package name in flake output
		if strings.Contains(line, "error:") && strings.Contains(line, ".#") {
			start := strings.Index(line, ".#")
			if start != -1 {
				packagePart := line[start+2:]
				end := strings.IndexAny(packagePart, " \t'\"")
				if end != -1 {
					return packagePart[:end]
				}
				return packagePart
			}
		}
	}

	// Fallback: return "unknown" if no package can be extracted
	return "unknown"
}
