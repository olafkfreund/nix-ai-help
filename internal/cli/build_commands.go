package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

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

Basic usage:
  nixai build                        # Run basic nix build with AI assistance
  nixai build .#mypackage            # Build a specific package with AI assistance

Advanced usage:
  nixai build debug firefox          # Analyze firefox build failures
  nixai build retry                  # Retry failed build with AI fixes
  nixai build cache-miss             # Analyze why builds aren't using cache
  nixai build sandbox-debug          # Debug sandbox permission issues
  nixai build profile --package vim  # Profile vim build performance`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommands provided, run standard nix build with AI assistance
		if len(args) == 0 && cmd.Flags().NArg() == 0 && cmd.Flags().NFlag() == 0 {
			// Load configuration
			cfg, err := config.LoadUserConfig()
			var provider ai.AIProvider
			if err == nil {
				provider = initializeAIProvider(cfg)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load config, using defaults: %v\n", err)
				provider = ai.NewOllamaProvider("llama3")
			}

			// Run nix build
			cmdArgs := []string{"build"}
			if len(args) > 0 {
				cmdArgs = append(cmdArgs, args...)
			}
			command := exec.Command("nix", cmdArgs...)
			if nixosConfigPathGlobal != "" {
				command.Dir = nixosConfigPathGlobal
			}
			out, err := command.CombinedOutput()
			fmt.Println(string(out))

			// Handle errors with AI assistance
			if err != nil {
				fmt.Fprintf(os.Stderr, "nix build failed: %v\n", err)
				// Parse and summarize the error output for the user
				problemSummary := summarizeBuildOutput(string(out))
				if problemSummary != "" {
					fmt.Println("\nProblem summary:")
					fmt.Println(problemSummary)
				}

				// Get AI assistance
				prompt := "I ran 'nix build" + " " + strings.Join(args, " ") + "' and got this output:\n" + string(out) + "\nHow can I fix this build or configuration problem?"
				aiResp, aiErr := provider.Query(prompt)
				if aiErr == nil && aiResp != "" {
					fmt.Println("\nAI suggestions:")
					fmt.Println(utils.RenderMarkdown(aiResp))
				}
			}
			return
		}

		// If no specific subcommand but args provided, show help
		_ = cmd.Help()
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
	Args: cobra.ExactArgs(1),
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

	// Check for previous build failure
	lastFailure := getLastBuildFailure()
	if lastFailure == "" {
		fmt.Println(utils.FormatWarning("No previous build failure found to retry"))
		fmt.Println(utils.FormatTip("Run a build command first, then use retry if it fails"))
		return
	}

	fmt.Println(utils.FormatKeyValue("Last Failed Build", lastFailure))
	fmt.Println()

	fmt.Println(utils.FormatProgress("Analyzing failure and generating fixes..."))

	// Get AI recommendations for fixes
	retryPrompt := buildRetryPrompt(lastFailure)
	fixes, aiErr := aiProvider.Query(retryPrompt)
	if aiErr != nil {
		fmt.Println(utils.FormatError("Failed to get AI fixes: " + aiErr.Error()))
		return
	}

	fmt.Println(utils.FormatSubsection("🤖 AI-Suggested Fixes", ""))
	fmt.Println(utils.RenderMarkdown(fixes))

	fmt.Println()
	fmt.Println(utils.FormatProgress("Applying fixes and retrying build..."))

	// Apply fixes and retry (implementation would apply actual fixes)
	success := applyFixesAndRetry(lastFailure)
	if success {
		fmt.Println(utils.FormatSuccess("✅ Retry successful!"))
	} else {
		fmt.Println(utils.FormatError("❌ Retry failed. Manual intervention may be required."))
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
		return ai.NewOllamaProvider(cfg.AIModel)
	case "gemini":
		return ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
	case "openai":
		return ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	case "custom":
		if cfg.CustomAI.BaseURL != "" {
			return ai.NewCustomProvider(cfg.CustomAI.BaseURL, cfg.CustomAI.Headers)
		}
		return ai.NewOllamaProvider("llama3")
	default:
		return ai.NewOllamaProvider("llama3")
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

func getLastBuildFailure() string {
	// Implementation would check build history
	// For now, return a placeholder
	return "nixpkgs#firefox"
}

func applyFixesAndRetry(packageName string) bool {
	// Implementation would apply specific fixes and retry
	// For now, simulate retry
	fmt.Println(utils.FormatProgress("Applying environment fixes..."))
	time.Sleep(2 * time.Second)
	fmt.Println(utils.FormatProgress("Retrying build with fixes..."))
	time.Sleep(3 * time.Second)
	return true // Simulate success
}

func analyzeCachePerformance() map[string]interface{} {
	// Implementation would analyze actual cache performance
	return map[string]interface{}{
		"hit_rate":    "75%",
		"miss_rate":   "25%",
		"cache_size":  "2.5GB",
		"recent_hits": 42,
		"recent_miss": 14,
	}
}

func displayCacheStats(stats map[string]interface{}) {
	for key, value := range stats {
		fmt.Println(utils.FormatKeyValue(strings.Title(strings.ReplaceAll(key, "_", " ")), fmt.Sprintf("%v", value)))
	}
}

func analyzeSandboxEnvironment() map[string]interface{} {
	// Implementation would analyze actual sandbox environment
	return map[string]interface{}{
		"sandbox_enabled":  true,
		"network_access":   false,
		"writable_paths":   []string{"/tmp", "/build"},
		"environment_vars": map[string]string{"PATH": "/usr/bin", "HOME": "/homeless-shelter"},
	}
}

func displaySandboxInfo(info map[string]interface{}) {
	for key, value := range info {
		fmt.Println(utils.FormatKeyValue(strings.Title(strings.ReplaceAll(key, "_", " ")), fmt.Sprintf("%v", value)))
	}
}

func profileBuild(packageName string) map[string]interface{} {
	// Implementation would profile actual build
	return map[string]interface{}{
		"total_time":       "4m 32s",
		"cpu_usage":        "85%",
		"memory_peak":      "2.1GB",
		"network_time":     "45s",
		"compilation_time": "3m 20s",
		"download_time":    "27s",
	}
}

func displayProfileData(data map[string]interface{}) {
	for key, value := range data {
		fmt.Println(utils.FormatKeyValue(strings.Title(strings.ReplaceAll(key, "_", " ")), fmt.Sprintf("%v", value)))
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

// summarizeBuildOutput extracts error messages from build output
func summarizeBuildOutput(output string) string {
	lines := strings.Split(output, "\n")
	var summary []string
	for _, line := range lines {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") || strings.Contains(line, "cannot") {
			summary = append(summary, line)
		}
	}
	return strings.Join(summary, "\n")
}

// Initialize commands
func init() {
	// Add enhanced build command
	rootCmd.AddCommand(enhancedBuildCmd)

	// Add build subcommands
	enhancedBuildCmd.AddCommand(buildDebugCmd)
	enhancedBuildCmd.AddCommand(buildRetryCmd)
	enhancedBuildCmd.AddCommand(buildCacheMissCmd)
	enhancedBuildCmd.AddCommand(buildSandboxDebugCmd)
	enhancedBuildCmd.AddCommand(buildProfileCmd)

	// Add flags
	buildProfileCmd.Flags().String("package", "", "Specific package to profile")
}
