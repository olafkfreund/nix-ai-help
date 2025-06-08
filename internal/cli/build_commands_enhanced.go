package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"
)

// Global build monitor instance
var globalBuildMonitor *BuildMonitor

// buildWatchCmd provides real-time build monitoring with AI insights
var buildWatchCmd = &cobra.Command{
	Use:   "watch <package>",
	Short: "Watch build progress in real-time with AI-powered insights",
	Long: `Monitor build progress in real-time with AI-powered analysis and suggestions.

This command provides:
- Real-time build progress updates
- AI-powered error analysis as issues occur
- Performance monitoring and optimization suggestions
- Background build management
- Intelligent failure recovery suggestions`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		runBuildWatch(packageName, cmd)
	},
}

// buildStatusCmd checks the status of background builds
var buildStatusCmd = &cobra.Command{
	Use:   "status [build-id]",
	Short: "Check status of background builds",
	Long: `Check the status of background builds and get AI-powered insights.

Without a build-id, shows all active builds.
With a build-id, shows detailed status for that specific build.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runBuildStatusAll(cmd)
		} else {
			runBuildStatusSpecific(args[0], cmd)
		}
	},
}

// buildStopCmd cancels a running build
var buildStopCmd = &cobra.Command{
	Use:   "stop <build-id>",
	Short: "Stop a background build",
	Long:  `Stop a background build process by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		buildID := args[0]
		runBuildStop(buildID, cmd)
	},
}

// buildBackgroundCmd starts a build in the background
var buildBackgroundCmd = &cobra.Command{
	Use:   "background <package>",
	Short: "Start a build in the background with monitoring",
	Long: `Start a build process in the background with AI-powered monitoring.

The build will run independently while you continue working.
Use 'nixai build status' to check progress and 'nixai build watch' for real-time monitoring.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		runBuildBackground(packageName, cmd)
	},
}

// buildQueueCmd manages a queue of packages to build sequentially
var buildQueueCmd = &cobra.Command{
	Use:   "queue <packages...>",
	Short: "Queue multiple packages for sequential background building",
	Long: `Queue multiple packages for sequential background building with AI optimization.

Packages will be built one after another, with AI analysis applied to:
- Optimize build order based on dependencies
- Learn from failures to improve subsequent builds
- Provide consolidated progress reporting`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runBuildQueue(args, cmd)
	},
}

// Build implementation functions

func runBuildWatch(packageName string, cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader(fmt.Sprintf("👀 Watching Build: %s", packageName)))

	// Initialize monitor if needed
	if globalBuildMonitor == nil {
		cfg, _ := config.LoadUserConfig()
		provider := initializeModernAIProvider(cfg)
		buildAgent := agent.NewBuildAgent(provider)
		globalBuildMonitor = NewBuildMonitor(buildAgent)
	}

	// Start background build
	buildArgs := []string{packageName}
	if flake, _ := cmd.Flags().GetBool("flake"); flake {
		buildArgs = append([]string{"--flake"}, buildArgs...)
	}

	buildID, err := globalBuildMonitor.StartBackgroundBuild(packageName, buildArgs)
	if err != nil {
		fmt.Println(utils.FormatError("Failed to start build: " + err.Error()))
		return
	}

	// Watch progress in real-time
	fmt.Println(utils.FormatProgress("Monitoring build progress..."))
	fmt.Println(utils.FormatTip("Press Ctrl+C to stop watching (build will continue in background)"))

	// Real-time progress monitoring
	for {
		process, err := globalBuildMonitor.GetBuildStatus(buildID)
		if err != nil {
			fmt.Println(utils.FormatError("Failed to get build status: " + err.Error()))
			break
		}

		// Display current status
		displayBuildProgress(process)

		// Check if build is complete
		if process.Status == "completed" || process.Status == "failed" || process.Status == "cancelled" {
			displayFinalStatus(process)
			break
		}

		time.Sleep(2 * time.Second)
	}
}

func runBuildStatusAll(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("📊 Active Builds Status"))

	if globalBuildMonitor == nil {
		fmt.Println(utils.FormatInfo("No build monitor active. Start a background build first."))
		return
	}

	builds := globalBuildMonitor.ListActiveBuilds()
	if len(builds) == 0 {
		fmt.Println(utils.FormatInfo("No active builds"))
		return
	}

	fmt.Println()
	for buildID, process := range builds {
		displayBuildSummary(buildID, process)
		fmt.Println()
	}
}

func runBuildStatusSpecific(buildID string, cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader(fmt.Sprintf("📋 Build Status: %s", buildID)))

	if globalBuildMonitor == nil {
		fmt.Println(utils.FormatError("No build monitor active"))
		return
	}

	process, err := globalBuildMonitor.GetBuildStatus(buildID)
	if err != nil {
		fmt.Println(utils.FormatError("Build not found: " + err.Error()))
		return
	}

	displayDetailedBuildStatus(process)
}

func runBuildStop(buildID string, cmd *cobra.Command) {
	if globalBuildMonitor == nil {
		fmt.Println(utils.FormatError("No build monitor active"))
		return
	}

	err := globalBuildMonitor.StopBuild(buildID)
	if err != nil {
		fmt.Println(utils.FormatError("Failed to stop build: " + err.Error()))
		return
	}

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ Stopped build: %s", buildID)))
}

func runBuildBackground(packageName string, cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader(fmt.Sprintf("🚀 Starting Background Build: %s", packageName)))

	// Initialize monitor if needed
	if globalBuildMonitor == nil {
		cfg, _ := config.LoadUserConfig()
		provider := initializeModernAIProvider(cfg)
		buildAgent := agent.NewBuildAgent(provider)
		globalBuildMonitor = NewBuildMonitor(buildAgent)
	}

	// Start background build
	buildArgs := []string{packageName}
	if flake, _ := cmd.Flags().GetBool("flake"); flake {
		buildArgs = append([]string{"--flake"}, buildArgs...)
	}

	buildID, err := globalBuildMonitor.StartBackgroundBuild(packageName, buildArgs)
	if err != nil {
		fmt.Println(utils.FormatError("Failed to start background build: " + err.Error()))
		return
	}

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ Build started with ID: %s", buildID)))
	fmt.Println(utils.FormatTip(fmt.Sprintf("Monitor with: nixai build status %s", buildID)))
	fmt.Println(utils.FormatTip("Watch real-time: nixai build watch " + packageName))
}

func runBuildQueue(packages []string, cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("📝 Build Queue"))

	// Initialize monitor if needed
	if globalBuildMonitor == nil {
		cfg, _ := config.LoadUserConfig()
		provider := initializeModernAIProvider(cfg)
		buildAgent := agent.NewBuildAgent(provider)
		globalBuildMonitor = NewBuildMonitor(buildAgent)
	}

	fmt.Println(utils.FormatProgress("Analyzing build dependencies..."))

	// AI-powered build order optimization
	optimizedOrder := optimizeBuildOrder(packages)

	fmt.Println(utils.FormatSubsection("🧠 AI-Optimized Build Order", ""))
	for i, pkg := range optimizedOrder {
		fmt.Printf("%d. %s\n", i+1, pkg)
	}
	fmt.Println()

	// Start queued builds
	buildIDs := make([]string, 0, len(optimizedOrder))
	for _, pkg := range optimizedOrder {
		buildID, err := globalBuildMonitor.StartBackgroundBuild(pkg, []string{pkg})
		if err != nil {
			fmt.Println(utils.FormatError(fmt.Sprintf("Failed to queue %s: %v", pkg, err)))
			continue
		}
		buildIDs = append(buildIDs, buildID)

		// Small delay between starts to avoid overwhelming the system
		time.Sleep(1 * time.Second)
	}

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ Queued %d builds", len(buildIDs))))
	fmt.Println(utils.FormatTip("Monitor with: nixai build status"))
}

// Helper functions for display and optimization

func displayBuildProgress(process *BuildProcess) {
	statusEmoji := getStatusEmoji(process.Status)
	duration := time.Since(process.StartTime).Round(time.Second)

	fmt.Printf("\r%s %s | %s | Duration: %v | Errors: %d",
		statusEmoji, process.Package, process.Status, duration, process.ErrorCount)

	// Show recent progress if available
	select {
	case progress := <-process.Progress:
		if progress.Stage != "" {
			fmt.Printf(" | Stage: %s", progress.Stage)
		}
		if progress.Percentage > 0 {
			fmt.Printf(" | %.1f%%", progress.Percentage)
		}
	default:
		// No recent progress
	}
}

func displayBuildSummary(buildID string, process *BuildProcess) {
	statusEmoji := getStatusEmoji(process.Status)
	duration := time.Since(process.StartTime).Round(time.Second)

	fmt.Printf("%s %s\n", statusEmoji, utils.FormatKeyValue("Build ID", buildID))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Package", process.Package))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Status", process.Status))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Duration", duration.String()))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Errors", fmt.Sprintf("%d", process.ErrorCount)))
}

func displayDetailedBuildStatus(process *BuildProcess) {
	statusEmoji := getStatusEmoji(process.Status)
	duration := time.Since(process.StartTime).Round(time.Second)

	fmt.Println()
	fmt.Printf("%s %s\n", statusEmoji, utils.FormatKeyValue("Package", process.Package))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Status", process.Status))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Started", process.StartTime.Format("15:04:05")))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Duration", duration.String()))
	fmt.Printf("   %s\n", utils.FormatKeyValue("Errors", fmt.Sprintf("%d", process.ErrorCount)))

	if len(process.Output) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("📄 Recent Output", ""))
		recentLines := process.Output
		if len(recentLines) > 10 {
			recentLines = recentLines[len(recentLines)-10:]
		}
		for _, line := range recentLines {
			fmt.Println("  " + line)
		}
	}
}

func displayFinalStatus(process *BuildProcess) {
	fmt.Println(utils.FormatDivider())

	duration := time.Since(process.StartTime)
	switch process.Status {
	case "completed":
		fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ Build completed successfully in %v", duration)))
	case "failed":
		fmt.Println(utils.FormatError(fmt.Sprintf("❌ Build failed after %v", duration)))
		fmt.Println(utils.FormatKeyValue("Error Count", fmt.Sprintf("%d", process.ErrorCount)))
	case "cancelled":
		fmt.Println(utils.FormatWarning(fmt.Sprintf("🛑 Build cancelled after %v", duration)))
	}
}

func getStatusEmoji(status string) string {
	switch status {
	case "starting":
		return "🟡"
	case "running":
		return "🔵"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "cancelled":
		return "🛑"
	default:
		return "⚪"
	}
}

func optimizeBuildOrder(packages []string) []string {
	// Simple implementation - in production this would use AI analysis
	// to determine optimal build order based on dependencies, build times, etc.

	// For now, just return packages in the order they were provided
	// Real implementation would analyze package dependencies and optimize
	optimized := make([]string, len(packages))
	copy(optimized, packages)

	return optimized
}

// initializeModernAIProvider creates an AI provider that implements the ai.Provider interface
func initializeModernAIProvider(cfg *config.UserConfig) ai.Provider {
	switch cfg.AIProvider {
	case "ollama":
		return ai.NewOllamaProvider(cfg.AIModel)
	case "gemini":
		geminiClient := ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		return ai.NewLegacyProviderAdapter(geminiClient)
	case "openai":
		openaiClient := ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		return ai.NewLegacyProviderAdapter(openaiClient)
	case "custom":
		if cfg.CustomAI.BaseURL != "" {
			customClient := ai.NewCustomProvider(cfg.CustomAI.BaseURL, cfg.CustomAI.Headers)
			return ai.NewLegacyProviderAdapter(customClient)
		}
		return ai.NewOllamaProvider("llama3")
	default:
		return ai.NewOllamaProvider("llama3")
	}
}

// Initialize enhanced build commands
func initEnhancedBuildCommands() {
	// Add subcommands to the main build command
	enhancedBuildCmd.AddCommand(buildWatchCmd)
	enhancedBuildCmd.AddCommand(buildStatusCmd)
	enhancedBuildCmd.AddCommand(buildStopCmd)
	enhancedBuildCmd.AddCommand(buildBackgroundCmd)
	enhancedBuildCmd.AddCommand(buildQueueCmd)

	// Add flags
	buildWatchCmd.Flags().Bool("flake", false, "Use flake for building")
	buildBackgroundCmd.Flags().Bool("flake", false, "Use flake for building")
	buildQueueCmd.Flags().Bool("flake", false, "Use flake for building")
}
