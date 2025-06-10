package cli

import (
	"encoding/json"
	"fmt"
	"time"

	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage NixOS system context detection and caching",
	Long: `Manage the NixOS system context detection and caching system.

The context system automatically detects your NixOS configuration details including:
- Flakes vs channels usage
- Home Manager configuration
- NixOS version and system type
- Enabled services and installed packages
- Configuration file locations

This context information is used throughout nixai to provide more relevant
and targeted assistance.

Examples:
  nixai context detect    # Force re-detect system context
  nixai context show     # Display current context information
  nixai context reset    # Clear cached context and force refresh
  nixai context status   # Show context system status`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// contextDetectCmd forces context re-detection
var contextDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Force re-detection of NixOS system context",
	Long: `Force re-detection of your NixOS system context, ignoring any cached information.

This will scan your system to detect:
- Configuration type (flakes vs channels)
- Home Manager setup
- NixOS version and system information
- Enabled services and packages
- File locations and paths

The detected context will be cached for faster access in future commands.

Examples:
  nixai context detect                    # Re-detect context
  nixai context detect --format json     # Output in JSON format
  nixai context detect --verbose         # Show detailed detection process`,
	Run: func(cmd *cobra.Command, args []string) {
		runContextDetect(cmd)
	},
}

// contextShowCmd displays current context
var contextShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current NixOS system context",
	Long: `Display the current NixOS system context information.

Shows both cached and live context data including:
- System configuration type and version
- File paths and locations
- Enabled services and packages
- Cache validity and last detection time

Examples:
  nixai context show                     # Show context summary
  nixai context show --format json      # Output in JSON format
  nixai context show --detailed         # Show all context details`,
	Run: func(cmd *cobra.Command, args []string) {
		runContextShow(cmd)
	},
}

// contextResetCmd clears context cache
var contextResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear cached context and force refresh",
	Long: `Clear the cached NixOS system context and force a fresh detection.

This will:
- Remove all cached context information
- Force re-detection on next context access
- Reset any context-related errors or invalid states

Useful when your system configuration has changed significantly.

Examples:
  nixai context reset                    # Clear cache and refresh
  nixai context reset --confirm          # Skip confirmation prompt`,
	Run: func(cmd *cobra.Command, args []string) {
		runContextReset(cmd)
	},
}

// contextStatusCmd shows context system status
var contextStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show context detection system status",
	Long: `Show the status of the context detection system including:
- Cache validity and age
- Last detection results
- Any detection errors or warnings
- System compatibility information

Examples:
  nixai context status                   # Show status summary
  nixai context status --format json    # Output in JSON format`,
	Run: func(cmd *cobra.Command, args []string) {
		runContextStatus(cmd)
	},
}

// Implementation functions

func runContextDetect(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("🔍 NixOS Context Detection"))
	fmt.Println()

	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println(utils.FormatError("Error loading config: " + err.Error()))
		return
	}

	// Create context detector
	log := logger.NewLogger()
	if verbose {
		log = logger.NewLoggerWithLevel("debug")
	}
	contextDetector := nixos.NewContextDetector(log)

	if verbose {
		fmt.Println(utils.FormatProgress("Starting context detection process..."))
	}

	// Force fresh detection by clearing cache first
	contextDetector.ClearCache()

	// Detect context
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		fmt.Println(utils.FormatError("Context detection failed: " + err.Error()))
		return
	}

	if nixosCtx == nil {
		fmt.Println(utils.FormatWarning("No context detected"))
		return
	}

	// Output results
	switch format {
	case "json":
		data, err := json.MarshalIndent(nixosCtx, "", "  ")
		if err != nil {
			fmt.Println(utils.FormatError("Failed to marshal context: " + err.Error()))
			return
		}
		fmt.Println(string(data))
	default:
		displayContextInfo(nixosCtx, true)
	}

	if verbose {
		fmt.Println()
		fmt.Println(utils.FormatSuccess("✅ Context detection completed"))
		fmt.Printf("Cache location: %s\n", utils.FormatNote(contextDetector.GetCacheLocation()))
	}
}

func runContextShow(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("📋 NixOS System Context"))
	fmt.Println()

	format, _ := cmd.Flags().GetString("format")
	detailed, _ := cmd.Flags().GetBool("detailed")

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println(utils.FormatError("Error loading config: " + err.Error()))
		return
	}

	// Create context detector
	contextDetector := nixos.NewContextDetector(logger.NewLogger())

	// Get context (will use cache if valid)
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		fmt.Println(utils.FormatError("Failed to get context: " + err.Error()))
		return
	}

	if nixosCtx == nil {
		fmt.Println(utils.FormatWarning("No context available"))
		fmt.Println(utils.FormatTip("Run 'nixai context detect' to detect system context"))
		return
	}

	// Output results
	switch format {
	case "json":
		data, err := json.MarshalIndent(nixosCtx, "", "  ")
		if err != nil {
			fmt.Println(utils.FormatError("Failed to marshal context: " + err.Error()))
			return
		}
		fmt.Println(string(data))
	default:
		displayContextInfo(nixosCtx, detailed)
	}
}

func runContextReset(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("🔄 Reset NixOS Context"))
	fmt.Println()

	confirm, _ := cmd.Flags().GetBool("confirm")

	if !confirm {
		fmt.Println("This will clear all cached context information and force re-detection.")
		fmt.Println()

		if !utils.PromptYesNo("Continue with context reset?") {
			fmt.Println(utils.FormatInfo("Context reset cancelled"))
			return
		}
	}

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println(utils.FormatError("Error loading config: " + err.Error()))
		return
	}

	// Create context detector and clear cache
	contextDetector := nixos.NewContextDetector(logger.NewLogger())

	fmt.Println(utils.FormatProgress("Clearing context cache..."))
	contextDetector.ClearCache()

	fmt.Println(utils.FormatProgress("Re-detecting system context..."))

	// Force fresh detection
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		fmt.Println(utils.FormatError("Context re-detection failed: " + err.Error()))
		return
	}

	fmt.Println(utils.FormatSuccess("✅ Context reset completed"))
	fmt.Println()

	if nixosCtx != nil {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		fmt.Println(utils.FormatNote("📋 " + contextSummary))
	}
}

func runContextStatus(cmd *cobra.Command) {
	fmt.Println(utils.FormatHeader("📊 Context System Status"))
	fmt.Println()

	format, _ := cmd.Flags().GetString("format")

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println(utils.FormatError("Error loading config: " + err.Error()))
		return
	}

	// Create context detector
	contextDetector := nixos.NewContextDetector(logger.NewLogger())

	// Get context status
	nixosCtx, err := contextDetector.GetContext(cfg)

	status := map[string]interface{}{
		"cache_location": contextDetector.GetCacheLocation(),
		"has_context":    nixosCtx != nil,
		"cache_valid":    false,
		"last_detected":  nil,
		"errors":         []string{},
	}

	if err != nil {
		status["errors"] = []string{err.Error()}
	}

	if nixosCtx != nil {
		status["cache_valid"] = nixosCtx.CacheValid
		status["last_detected"] = nixosCtx.LastDetected
		if len(nixosCtx.DetectionErrors) > 0 {
			status["errors"] = nixosCtx.DetectionErrors
		}
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			fmt.Println(utils.FormatError("Failed to marshal status: " + err.Error()))
			return
		}
		fmt.Println(string(data))
	default:
		displayContextStatus(status, nixosCtx)
	}
}

// Helper functions

func displayContextInfo(nixosCtx *config.NixOSContext, detailed bool) {
	// Basic context summary
	contextBuilder := nixoscontext.NewNixOSContextBuilder()
	contextSummary := contextBuilder.GetContextSummary(nixosCtx)
	fmt.Println(utils.FormatKeyValue("System Summary", contextSummary))
	fmt.Println()

	// System Information
	fmt.Println(utils.FormatSubsection("System Information", ""))
	fmt.Println(utils.FormatKeyValue("System Type", nixosCtx.SystemType))
	if nixosCtx.NixOSVersion != "" {
		fmt.Println(utils.FormatKeyValue("NixOS Version", nixosCtx.NixOSVersion))
	}
	if nixosCtx.NixVersion != "" {
		fmt.Println(utils.FormatKeyValue("Nix Version", nixosCtx.NixVersion))
	}
	fmt.Println()

	// Configuration
	fmt.Println(utils.FormatSubsection("Configuration", ""))
	fmt.Println(utils.FormatKeyValue("Uses Flakes", formatBool(nixosCtx.UsesFlakes)))
	fmt.Println(utils.FormatKeyValue("Uses Channels", formatBool(nixosCtx.UsesChannels)))
	fmt.Println(utils.FormatKeyValue("Has Home Manager", formatBool(nixosCtx.HasHomeManager)))
	if nixosCtx.HasHomeManager {
		fmt.Println(utils.FormatKeyValue("Home Manager Type", nixosCtx.HomeManagerType))
	}
	fmt.Println()

	// File Paths
	if detailed {
		fmt.Println(utils.FormatSubsection("File Paths", ""))
		if nixosCtx.NixOSConfigPath != "" {
			fmt.Println(utils.FormatKeyValue("NixOS Config", nixosCtx.NixOSConfigPath))
		}
		if nixosCtx.HomeManagerConfigPath != "" {
			fmt.Println(utils.FormatKeyValue("Home Manager Config", nixosCtx.HomeManagerConfigPath))
		}
		if nixosCtx.FlakeFile != "" {
			fmt.Println(utils.FormatKeyValue("Flake File", nixosCtx.FlakeFile))
		}
		if nixosCtx.ConfigurationNix != "" {
			fmt.Println(utils.FormatKeyValue("Configuration.nix", nixosCtx.ConfigurationNix))
		}
		if nixosCtx.HardwareConfigNix != "" {
			fmt.Println(utils.FormatKeyValue("Hardware Config", nixosCtx.HardwareConfigNix))
		}
		fmt.Println()

		// Services and Packages (if available)
		if len(nixosCtx.EnabledServices) > 0 {
			fmt.Println(utils.FormatSubsection("Enabled Services", fmt.Sprintf("(%d total)", len(nixosCtx.EnabledServices))))
			for i, service := range nixosCtx.EnabledServices {
				if i >= 10 { // Limit display to first 10
					fmt.Println(utils.FormatNote(fmt.Sprintf("... and %d more", len(nixosCtx.EnabledServices)-10)))
					break
				}
				fmt.Println("  • " + service)
			}
			fmt.Println()
		}

		if len(nixosCtx.InstalledPackages) > 0 {
			fmt.Println(utils.FormatSubsection("Installed Packages", fmt.Sprintf("(%d total)", len(nixosCtx.InstalledPackages))))
			for i, pkg := range nixosCtx.InstalledPackages {
				if i >= 10 { // Limit display to first 10
					fmt.Println(utils.FormatNote(fmt.Sprintf("... and %d more", len(nixosCtx.InstalledPackages)-10)))
					break
				}
				fmt.Println("  • " + pkg)
			}
			fmt.Println()
		}
	}

	// Cache Information
	fmt.Println(utils.FormatSubsection("Cache Information", ""))
	fmt.Println(utils.FormatKeyValue("Cache Valid", formatBool(nixosCtx.CacheValid)))
	if !nixosCtx.LastDetected.IsZero() {
		fmt.Println(utils.FormatKeyValue("Last Detected", nixosCtx.LastDetected.Format("2006-01-02 15:04:05")))
		fmt.Println(utils.FormatKeyValue("Cache Age", time.Since(nixosCtx.LastDetected).Round(time.Second).String()))
	}

	if len(nixosCtx.DetectionErrors) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Detection Errors", ""))
		for _, errMsg := range nixosCtx.DetectionErrors {
			fmt.Println("  • " + utils.FormatError(errMsg))
		}
	}
}

func displayContextStatus(status map[string]interface{}, nixosCtx *config.NixOSContext) {
	fmt.Println(utils.FormatKeyValue("Cache Location", status["cache_location"].(string)))
	fmt.Println(utils.FormatKeyValue("Has Context", formatBool(status["has_context"].(bool))))
	fmt.Println(utils.FormatKeyValue("Cache Valid", formatBool(status["cache_valid"].(bool))))

	if lastDetected := status["last_detected"]; lastDetected != nil {
		if ts, ok := lastDetected.(time.Time); ok {
			fmt.Println(utils.FormatKeyValue("Last Detected", ts.Format("2006-01-02 15:04:05")))
			fmt.Println(utils.FormatKeyValue("Cache Age", time.Since(ts).Round(time.Second).String()))
		}
	}

	if errors, ok := status["errors"].([]string); ok && len(errors) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Errors", ""))
		for _, errMsg := range errors {
			fmt.Println("  • " + utils.FormatError(errMsg))
		}
	}

	fmt.Println()

	// System health check
	if nixosCtx != nil && nixosCtx.CacheValid {
		fmt.Println(utils.FormatSuccess("✅ Context system is healthy"))

		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		fmt.Println(utils.FormatNote("📋 " + contextSummary))
	} else {
		fmt.Println(utils.FormatWarning("⚠️  Context system needs attention"))
		fmt.Println(utils.FormatTip("Run 'nixai context detect' to refresh context"))
	}
}

func formatBool(b bool) string {
	if b {
		return "✅ Yes"
	}
	return "❌ No"
}

// Initialize context commands
func init() {
	// Add subcommands to context command
	contextCmd.AddCommand(contextDetectCmd)
	contextCmd.AddCommand(contextShowCmd)
	contextCmd.AddCommand(contextResetCmd)
	contextCmd.AddCommand(contextStatusCmd)

	// Add flags to detect command
	contextDetectCmd.Flags().StringP("format", "f", "", "Output format (json)")
	contextDetectCmd.Flags().BoolP("verbose", "v", false, "Show detailed detection process")

	// Add flags to show command
	contextShowCmd.Flags().StringP("format", "f", "", "Output format (json)")
	contextShowCmd.Flags().BoolP("detailed", "d", false, "Show detailed context information")

	// Add flags to reset command
	contextResetCmd.Flags().BoolP("confirm", "y", false, "Skip confirmation prompt")

	// Add flags to status command
	contextStatusCmd.Flags().StringP("format", "f", "", "Output format (json)")
}
