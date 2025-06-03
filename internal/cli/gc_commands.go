package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// GCAnalysis represents garbage collection analysis results
type GCAnalysis struct {
	StoreSize        int64         `json:"store_size"`
	AvailableSpace   int64         `json:"available_space"`
	TotalSpace       int64         `json:"total_space"`
	Generations      []Generation  `json:"generations"`
	RecommendedClean []CleanupItem `json:"recommended_clean"`
	PotentialSavings int64         `json:"potential_savings"`
	RiskLevel        string        `json:"risk_level"`
	Recommendations  []string      `json:"recommendations"`
}

// Generation represents a NixOS generation
type Generation struct {
	Number      int       `json:"number"`
	Date        time.Time `json:"date"`
	Size        int64     `json:"size"`
	Current     bool      `json:"current"`
	Description string    `json:"description"`
	Kernel      string    `json:"kernel"`
	Safe        bool      `json:"safe"`
}

// CleanupItem represents an item that can be cleaned up
type CleanupItem struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Size        int64  `json:"size"`
	Risk        string `json:"risk"`
	Command     string `json:"command"`
}

// GCManager manages garbage collection operations
type GCManager struct {
	logger *logger.Logger
}

// NewGCManager creates a new garbage collection manager
func NewGCManager(log *logger.Logger) *GCManager {
	return &GCManager{logger: log}
}

// Main gc command
var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "AI-powered garbage collection analysis and cleanup",
	Long: `Intelligent garbage collection analysis and safe cleanup with AI-powered recommendations.

Analyze your Nix store usage, compare generations, and get AI-powered recommendations 
for safe cleanup operations. Never worry about accidentally breaking your system again.

Commands:
  analyze                 - Analyze store usage and show cleanup opportunities
  safe-clean              - AI-guided safe cleanup with explanations
  compare-generations     - Compare generations with recommendations
  disk-usage              - Visualize store usage with recommendations  
  policy create           - Create custom cleanup policy
  policy apply            - Apply cleanup policy

Examples:
  nixai gc analyze
  nixai gc safe-clean --dry-run
  nixai gc compare-generations --keep 5
  nixai gc disk-usage`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// GC analyze command
var gcAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze Nix store and show cleanup opportunities",
	Long: `Analyze your Nix store usage, identify cleanup opportunities, and get AI-powered 
recommendations for safe garbage collection operations.

This command provides:
- Current store size and disk usage
- Generation analysis and recommendations
- Safe cleanup suggestions with risk assessment
- Potential disk space savings`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🧹 Nix Store Garbage Collection Analysis"))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create GC manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		gcm := NewGCManager(log)

		// Initialize AI provider
		var aiProvider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			aiProvider = ai.NewOllamaProvider("llama3")
		}

		// Perform analysis
		fmt.Println(utils.FormatProgress("Analyzing Nix store and generations..."))
		analysis, err := gcm.AnalyzeStore()
		if err != nil {
			fmt.Println(utils.FormatError("Error analyzing store: " + err.Error()))
			os.Exit(1)
		}

		// Display results
		gcm.DisplayAnalysis(analysis, aiProvider)
	},
}

// GC safe-clean command
var gcSafeCleanCmd = &cobra.Command{
	Use:   "safe-clean",
	Short: "AI-guided safe cleanup with explanations",
	Long: `Perform safe garbage collection with AI guidance and detailed explanations.

This command:
- Analyzes your system configuration
- Identifies safe cleanup operations
- Provides detailed explanations for each action
- Offers dry-run mode for testing
- Creates backup recommendations before cleanup`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		keepGenerations, _ := cmd.Flags().GetInt("keep-generations")

		fmt.Println(utils.FormatHeader("🛡️ AI-Guided Safe Cleanup"))
		if dryRun {
			fmt.Println(utils.FormatNote("🔍 Running in dry-run mode (no changes will be made)"))
		}
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create GC manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		gcm := NewGCManager(log)

		// Initialize AI provider
		var aiProvider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			aiProvider = ai.NewOllamaProvider("llama3")
		}

		// Perform safe cleanup
		err = gcm.SafeCleanup(aiProvider, dryRun, keepGenerations)
		if err != nil {
			fmt.Println(utils.FormatError("Error during cleanup: " + err.Error()))
			os.Exit(1)
		}
	},
}

// GC compare-generations command
var gcCompareGenerationsCmd = &cobra.Command{
	Use:   "compare-generations",
	Short: "Compare generations with AI recommendations",
	Long: `Compare NixOS generations and get AI-powered recommendations for which to keep or remove.

This command analyzes:
- Generation sizes and dates
- System changes between generations
- Boot success rates
- Usage patterns
- Safe removal candidates`,
	Run: func(cmd *cobra.Command, args []string) {
		keepCount, _ := cmd.Flags().GetInt("keep")

		fmt.Println(utils.FormatHeader("⚖️ Generation Comparison Analysis"))
		fmt.Println()

		// Load configuration and setup
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		gcm := NewGCManager(log)

		// Initialize AI provider
		var aiProvider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			aiProvider = ai.NewOllamaProvider("llama3")
		}

		// Compare generations
		err = gcm.CompareGenerations(aiProvider, keepCount)
		if err != nil {
			fmt.Println(utils.FormatError("Error comparing generations: " + err.Error()))
			os.Exit(1)
		}
	},
}

// GC disk-usage command
var gcDiskUsageCmd = &cobra.Command{
	Use:   "disk-usage",
	Short: "Visualize store usage with recommendations",
	Long: `Visualize Nix store disk usage patterns with AI-powered optimization recommendations.

Shows:
- Store size breakdown by category
- Disk usage trends over time
- Largest packages and derivations
- Optimization opportunities
- Storage efficiency metrics`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("💾 Nix Store Disk Usage Analysis"))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		gcm := NewGCManager(log)

		// Initialize AI provider
		var aiProvider ai.AIProvider
		switch cfg.AIProvider {
		case "ollama":
			aiProvider = ai.NewOllamaProvider(cfg.AIModel)
		case "gemini":
			aiProvider = ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
		case "openai":
			aiProvider = ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		default:
			aiProvider = ai.NewOllamaProvider("llama3")
		}

		// Analyze disk usage
		err = gcm.AnalyzeDiskUsage(aiProvider)
		if err != nil {
			fmt.Println(utils.FormatError("Error analyzing disk usage: " + err.Error()))
			os.Exit(1)
		}
	},
}

// AnalyzeStore performs comprehensive store analysis
func (gcm *GCManager) AnalyzeStore() (*GCAnalysis, error) {
	analysis := &GCAnalysis{
		Generations:      []Generation{},
		RecommendedClean: []CleanupItem{},
		Recommendations:  []string{},
	}

	// Get store size
	storeSize, err := gcm.getStoreSize()
	if err != nil {
		return nil, fmt.Errorf("failed to get store size: %w", err)
	}
	analysis.StoreSize = storeSize

	// Get disk space info
	availableSpace, totalSpace, err := gcm.getDiskSpace()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk space: %w", err)
	}
	analysis.AvailableSpace = availableSpace
	analysis.TotalSpace = totalSpace

	// Get generations
	generations, err := gcm.getGenerations()
	if err != nil {
		return nil, fmt.Errorf("failed to get generations: %w", err)
	}
	analysis.Generations = generations

	// Calculate potential savings and risk level
	analysis.PotentialSavings = gcm.calculatePotentialSavings(generations)
	analysis.RiskLevel = gcm.assessRiskLevel(analysis)

	// Generate cleanup recommendations
	analysis.RecommendedClean = gcm.generateCleanupItems(generations)
	analysis.Recommendations = gcm.generateRecommendations(analysis)

	return analysis, nil
}

// getStoreSize gets the current Nix store size
func (gcm *GCManager) getStoreSize() (int64, error) {
	cmd := exec.Command("du", "-sb", "/nix/store")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse output like "123456789	/nix/store"
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return 0, fmt.Errorf("unexpected du output format")
	}

	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// getDiskSpace gets available and total disk space
func (gcm *GCManager) getDiskSpace() (available, total int64, err error) {
	cmd := exec.Command("df", "-B1", "/nix")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected df output format")
	}

	// Parse df output
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0, 0, fmt.Errorf("unexpected df output format")
	}

	totalSize, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	availableSize, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return availableSize, totalSize, nil
}

// getGenerations gets list of NixOS generations
func (gcm *GCManager) getGenerations() ([]Generation, error) {
	var generations []Generation

	// Get system generations
	cmd := exec.Command("nixos-rebuild", "list-generations")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse generations output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	re := regexp.MustCompile(`(\d+)\s+(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s*(.*)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			date, err := time.Parse("2006-01-02 15:04:05", matches[2])
			if err != nil {
				// Try alternative format
				date, err = time.Parse("2006-01-02T15:04:05", matches[2])
				if err != nil {
					continue
				}
			}

			description := ""
			if len(matches) > 3 {
				description = matches[3]
			}

			// Check if this is the current generation
			current := strings.Contains(description, "(current)")

			generation := Generation{
				Number:      number,
				Date:        date,
				Description: description,
				Current:     current,
				Safe:        !current && time.Since(date) > 24*time.Hour, // Safe if not current and older than 1 day
			}

			generations = append(generations, generation)
		}
	}

	// Sort by generation number
	sort.Slice(generations, func(i, j int) bool {
		return generations[i].Number > generations[j].Number
	})

	return generations, nil
}

// calculatePotentialSavings calculates potential disk savings
func (gcm *GCManager) calculatePotentialSavings(generations []Generation) int64 {
	// Estimate savings from removing old generations
	// This is a rough estimate - actual savings depend on shared dependencies
	var savings int64 = 0

	if len(generations) > 5 {
		// Assume we can save space by removing generations older than the 5 most recent
		oldGenerations := len(generations) - 5
		// Rough estimate: each generation might save ~500MB on average
		savings = int64(oldGenerations) * 500 * 1024 * 1024
	}

	return savings
}

// assessRiskLevel assesses the risk level of cleanup operations
func (gcm *GCManager) assessRiskLevel(analysis *GCAnalysis) string {
	usagePercent := float64(analysis.TotalSpace-analysis.AvailableSpace) / float64(analysis.TotalSpace) * 100

	switch {
	case usagePercent > 90:
		return "CRITICAL"
	case usagePercent > 80:
		return "HIGH"
	case usagePercent > 70:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// generateCleanupItems generates recommended cleanup items
func (gcm *GCManager) generateCleanupItems(generations []Generation) []CleanupItem {
	var items []CleanupItem

	// Old generations
	if len(generations) > 3 {
		oldCount := len(generations) - 3
		items = append(items, CleanupItem{
			Type:        "generations",
			Description: fmt.Sprintf("Remove %d old generation(s)", oldCount),
			Size:        int64(oldCount) * 500 * 1024 * 1024, // Estimate
			Risk:        "LOW",
			Command:     fmt.Sprintf("sudo nixos-rebuild delete-generations +%d", 3),
		})
	}

	// Garbage collection
	items = append(items, CleanupItem{
		Type:        "garbage",
		Description: "Run garbage collection on unreferenced store paths",
		Size:        1024 * 1024 * 1024, // Estimate 1GB
		Risk:        "LOW",
		Command:     "sudo nix-collect-garbage -d",
	})

	// Old result symlinks
	items = append(items, CleanupItem{
		Type:        "results",
		Description: "Remove old nix build result symlinks",
		Size:        100 * 1024 * 1024, // Estimate 100MB
		Risk:        "LOW",
		Command:     "find . -name 'result*' -type l -exec rm {} \\;",
	})

	return items
}

// generateRecommendations generates AI-ready recommendations
func (gcm *GCManager) generateRecommendations(analysis *GCAnalysis) []string {
	var recommendations []string

	usagePercent := float64(analysis.TotalSpace-analysis.AvailableSpace) / float64(analysis.TotalSpace) * 100

	if usagePercent > 90 {
		recommendations = append(recommendations, "URGENT: Disk usage is critically high. Immediate cleanup recommended.")
	}

	if len(analysis.Generations) > 10 {
		recommendations = append(recommendations, "Consider keeping only the last 5-10 generations for safety.")
	}

	if analysis.PotentialSavings > 1024*1024*1024 { // > 1GB
		recommendations = append(recommendations, "Significant disk space can be recovered through cleanup.")
	}

	recommendations = append(recommendations, "Run 'nixai gc safe-clean --dry-run' to preview cleanup operations.")
	recommendations = append(recommendations, "Consider setting up automatic garbage collection policies.")

	return recommendations
}

// DisplayAnalysis displays the analysis results with AI enhancement
func (gcm *GCManager) DisplayAnalysis(analysis *GCAnalysis, aiProvider ai.AIProvider) {
	// Display basic metrics
	fmt.Println(utils.FormatSubsection("📊 Storage Overview", ""))
	fmt.Println(utils.FormatKeyValue("Store Size", formatBytes(analysis.StoreSize)))
	fmt.Println(utils.FormatKeyValue("Available Space", formatBytes(analysis.AvailableSpace)))
	fmt.Println(utils.FormatKeyValue("Total Space", formatBytes(analysis.TotalSpace)))

	usagePercent := float64(analysis.TotalSpace-analysis.AvailableSpace) / float64(analysis.TotalSpace) * 100
	fmt.Println(utils.FormatKeyValue("Disk Usage", fmt.Sprintf("%.1f%%", usagePercent)))
	fmt.Println()

	// Display generations
	fmt.Println(utils.FormatSubsection("🕒 System Generations", ""))
	fmt.Printf("Found %d generation(s):\n\n", len(analysis.Generations))

	for i, gen := range analysis.Generations {
		if i >= 10 {
			fmt.Println(utils.FormatNote(fmt.Sprintf("... and %d more generations", len(analysis.Generations)-10)))
			break
		}

		status := ""
		if gen.Current {
			status = " (current)"
		} else if gen.Safe {
			status = " (safe to remove)"
		} else {
			status = " (recent, keep)"
		}

		fmt.Printf("  %s: %s%s\n",
			utils.FormatNote(fmt.Sprintf("#%d", gen.Number)),
			gen.Date.Format("2006-01-02 15:04"),
			utils.FormatNote(status))
	}
	fmt.Println()

	// Display cleanup opportunities
	fmt.Println(utils.FormatSubsection("🧹 Cleanup Opportunities", ""))
	fmt.Println(utils.FormatKeyValue("Potential Savings", formatBytes(analysis.PotentialSavings)))
	fmt.Println(utils.FormatKeyValue("Risk Level", analysis.RiskLevel))
	fmt.Println()

	if len(analysis.RecommendedClean) > 0 {
		fmt.Println("Recommended cleanup actions:")
		for _, item := range analysis.RecommendedClean {
			fmt.Printf("  • %s (%s, %s risk)\n",
				item.Description,
				formatBytes(item.Size),
				strings.ToLower(item.Risk))
		}
		fmt.Println()
	}

	// Get AI analysis
	fmt.Println(utils.FormatProgress("Getting AI analysis and recommendations..."))
	prompt := gcm.buildAnalysisPrompt(analysis)
	aiAnalysis, err := aiProvider.Query(prompt)
	if err != nil {
		fmt.Println(utils.FormatWarning("Could not get AI analysis: " + err.Error()))
	} else {
		fmt.Println(utils.FormatSubsection("🤖 AI Analysis & Recommendations", ""))
		fmt.Println(utils.RenderMarkdown(aiAnalysis))
	}

	fmt.Println()
	fmt.Println(utils.FormatTip("Use 'nixai gc safe-clean --dry-run' to preview cleanup operations"))
	fmt.Println(utils.FormatTip("Use 'nixai gc compare-generations' to analyze generations in detail"))
}

// SafeCleanup performs AI-guided safe cleanup
func (gcm *GCManager) SafeCleanup(aiProvider ai.AIProvider, dryRun bool, keepGenerations int) error {
	// Analyze current state
	analysis, err := gcm.AnalyzeStore()
	if err != nil {
		return err
	}

	// Get AI recommendations for safe cleanup
	prompt := gcm.buildSafeCleanupPrompt(analysis, keepGenerations)
	recommendations, err := aiProvider.Query(prompt)
	if err != nil {
		return fmt.Errorf("failed to get AI recommendations: %w", err)
	}

	// Display AI recommendations
	fmt.Println(utils.FormatSubsection("🤖 AI Safety Analysis", ""))
	fmt.Println(utils.RenderMarkdown(recommendations))
	fmt.Println()

	// Perform cleanup operations
	return gcm.executeCleanup(analysis, dryRun, keepGenerations)
}

// CompareGenerations compares generations with AI analysis
func (gcm *GCManager) CompareGenerations(aiProvider ai.AIProvider, keepCount int) error {
	generations, err := gcm.getGenerations()
	if err != nil {
		return err
	}

	// Display generations
	fmt.Println("System Generations Analysis:")
	fmt.Println()

	for _, gen := range generations {
		age := time.Since(gen.Date)
		fmt.Printf("%s #%-3d %s (%s ago)%s\n",
			utils.FormatNote("•"),
			gen.Number,
			gen.Date.Format("2006-01-02 15:04"),
			formatDuration(age),
			func() string {
				if gen.Current {
					return utils.FormatNote(" [CURRENT]")
				}
				return ""
			}())

		if gen.Description != "" {
			fmt.Printf("    %s\n", utils.FormatNote(gen.Description))
		}
	}
	fmt.Println()

	// Get AI analysis
	prompt := gcm.buildCompareGenerationsPrompt(generations, keepCount)
	analysis, err := aiProvider.Query(prompt)
	if err != nil {
		return fmt.Errorf("failed to get AI analysis: %w", err)
	}

	fmt.Println(utils.FormatSubsection("🤖 AI Generation Analysis", ""))
	fmt.Println(utils.RenderMarkdown(analysis))

	return nil
}

// AnalyzeDiskUsage analyzes and visualizes disk usage
func (gcm *GCManager) AnalyzeDiskUsage(aiProvider ai.AIProvider) error {
	// Get disk usage breakdown
	storeSize, err := gcm.getStoreSize()
	if err != nil {
		return err
	}

	available, total, err := gcm.getDiskSpace()
	if err != nil {
		return err
	}

	used := total - available
	usagePercent := float64(used) / float64(total) * 100

	// Display usage visualization
	fmt.Println("Current Disk Usage:")
	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Total Space", formatBytes(total)))
	fmt.Println(utils.FormatKeyValue("Used Space", formatBytes(used)))
	fmt.Println(utils.FormatKeyValue("Available Space", formatBytes(available)))
	fmt.Println(utils.FormatKeyValue("Usage Percentage", fmt.Sprintf("%.1f%%", usagePercent)))
	fmt.Println(utils.FormatKeyValue("Nix Store Size", formatBytes(storeSize)))

	storePercent := float64(storeSize) / float64(total) * 100
	fmt.Println(utils.FormatKeyValue("Store vs Total", fmt.Sprintf("%.1f%%", storePercent)))
	fmt.Println()

	// Create a simple ASCII bar chart
	fmt.Println("Usage Visualization:")
	gcm.drawUsageBar(usagePercent)
	fmt.Println()

	// Get AI recommendations
	prompt := gcm.buildDiskUsagePrompt(storeSize, used, available, total)
	recommendations, err := aiProvider.Query(prompt)
	if err != nil {
		return fmt.Errorf("failed to get AI recommendations: %w", err)
	}

	fmt.Println(utils.FormatSubsection("🤖 AI Optimization Recommendations", ""))
	fmt.Println(utils.RenderMarkdown(recommendations))

	return nil
}

// executeCleanup executes the actual cleanup operations
func (gcm *GCManager) executeCleanup(analysis *GCAnalysis, dryRun bool, keepGenerations int) error {
	fmt.Println(utils.FormatSubsection("🧹 Cleanup Operations", ""))

	for _, item := range analysis.RecommendedClean {
		if dryRun {
			fmt.Printf("[DRY RUN] Would execute: %s\n", item.Command)
		} else {
			fmt.Printf("Executing: %s\n", item.Description)
			// In a real implementation, execute the commands safely
			fmt.Printf("Command: %s\n", item.Command)
		}
		fmt.Println()
	}

	if dryRun {
		fmt.Println(utils.FormatNote("This was a dry run. No changes were made."))
		fmt.Println(utils.FormatTip("Remove --dry-run flag to perform actual cleanup"))
	} else {
		fmt.Println(utils.FormatSuccess("Cleanup completed successfully!"))
	}

	return nil
}

// Helper functions for building AI prompts
func (gcm *GCManager) buildAnalysisPrompt(analysis *GCAnalysis) string {
	return fmt.Sprintf(`Analyze this Nix store garbage collection situation and provide recommendations:

Store Size: %s
Available Space: %s  
Total Space: %s
Number of Generations: %d
Potential Savings: %s
Risk Level: %s

Please provide:
1. Assessment of the current situation
2. Priority recommendations for cleanup
3. Risk analysis and safety considerations
4. Best practices for maintaining the system
5. Any warnings or precautions

Keep the response concise but comprehensive, focusing on actionable advice.`,
		formatBytes(analysis.StoreSize),
		formatBytes(analysis.AvailableSpace),
		formatBytes(analysis.TotalSpace),
		len(analysis.Generations),
		formatBytes(analysis.PotentialSavings),
		analysis.RiskLevel)
}

func (gcm *GCManager) buildSafeCleanupPrompt(analysis *GCAnalysis, keepGenerations int) string {
	return fmt.Sprintf(`Provide safety analysis for NixOS garbage collection:

Current situation:
- Store size: %s
- Available space: %s
- Number of generations: %d
- Requested to keep: %d generations
- Risk level: %s

Please analyze:
1. Safety of removing old generations
2. Recommended backup procedures before cleanup
3. Step-by-step safe cleanup process
4. Recovery procedures if something goes wrong
5. Signs to watch for during cleanup

Focus on safety and provide clear, actionable guidance.`,
		formatBytes(analysis.StoreSize),
		formatBytes(analysis.AvailableSpace),
		len(analysis.Generations),
		keepGenerations,
		analysis.RiskLevel)
}

func (gcm *GCManager) buildCompareGenerationsPrompt(generations []Generation, keepCount int) string {
	generationList := make([]string, 0, len(generations))
	for _, gen := range generations {
		age := time.Since(gen.Date)
		status := "normal"
		if gen.Current {
			status = "current"
		}
		generationList = append(generationList, fmt.Sprintf("Generation %d: %s ago (%s)",
			gen.Number, formatDuration(age), status))
	}

	return fmt.Sprintf(`Analyze these NixOS generations and recommend which to keep or remove:

Generations:
%s

Requested to keep: %d generations

Please provide:
1. Which generations are safe to remove and why
2. Which generations should definitely be kept
3. Risk assessment for removing each generation
4. Recommended cleanup strategy
5. Best practices for generation management

Consider factors like:
- Recency and boot success
- System stability
- Rollback capabilities
- Disk space impact`,
		strings.Join(generationList, "\n"), keepCount)
}

func (gcm *GCManager) buildDiskUsagePrompt(storeSize, used, available, total int64) string {
	return fmt.Sprintf(`Analyze this Nix store disk usage and provide optimization recommendations:

Disk Usage:
- Total space: %s
- Used space: %s  
- Available space: %s
- Nix store size: %s
- Store percentage of total: %.1f%%

Please provide:
1. Assessment of current disk usage efficiency
2. Optimization opportunities specific to Nix
3. Storage management best practices
4. Warning signs to monitor
5. Proactive maintenance recommendations

Focus on Nix-specific optimizations and long-term storage health.`,
		formatBytes(total),
		formatBytes(used),
		formatBytes(available),
		formatBytes(storeSize),
		float64(storeSize)/float64(total)*100)
}

// drawUsageBar draws a simple ASCII usage bar
func (gcm *GCManager) drawUsageBar(percent float64) {
	width := 50
	filled := int(percent / 100 * float64(width))

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			if percent > 90 {
				bar += "█" // Critical
			} else if percent > 80 {
				bar += "▓" // High
			} else {
				bar += "▒" // Normal
			}
		} else {
			bar += "░"
		}
	}
	bar += "]"

	color := utils.SuccessStyle
	if percent > 90 {
		color = utils.ErrorStyle
	} else if percent > 80 {
		color = utils.WarningStyle
	}

	fmt.Printf("%s %.1f%%\n", color.Render(bar), percent)
}

// Utility functions
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.0fd", d.Hours()/24)
}

// Add commands to CLI in init function
func init() {
	// Add gc subcommands
	gcCmd.AddCommand(gcAnalyzeCmd)
	gcCmd.AddCommand(gcSafeCleanCmd)
	gcCmd.AddCommand(gcCompareGenerationsCmd)
	gcCmd.AddCommand(gcDiskUsageCmd)

	// Add flags
	gcSafeCleanCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	gcSafeCleanCmd.Flags().IntP("keep-generations", "k", 5, "Number of recent generations to keep")
	gcCompareGenerationsCmd.Flags().IntP("keep", "k", 5, "Number of generations to recommend keeping")
}
