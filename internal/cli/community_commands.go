package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/community"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// communitySearchCmd searches for community configurations
var communitySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search community configurations and packages",
	Long: `Search the NixOS community for configurations, packages, and solutions.

This command searches across multiple community sources including:
- Community-shared configurations from Discourse forums
- Popular package configurations and best practices
- GitHub repositories with NixOS configurations
- Curated configuration examples

Examples:
  nixai community search "gaming setup"
  nixai community search "docker configuration"
  nixai community search "kde plasma"
  nixai community search "server nginx"`,
	Args: conditionalArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		limit, _ := cmd.Flags().GetInt("limit")
		category, _ := cmd.Flags().GetString("category")

		runCommunitySearch(query, limit, category, cmd)
	},
}

// communityShareCmd shares a configuration with the community
var communityShareCmd = &cobra.Command{
	Use:   "share <config-file>",
	Short: "Share your configuration with the community",
	Long: `Share your NixOS configuration with the community to help others and get feedback.

This command:
- Validates your configuration before sharing
- Anonymizes sensitive information
- Adds metadata and description
- Publishes to community platforms
- Provides sharing statistics

Examples:
  nixai community share ./configuration.nix
  nixai community share ./flake.nix --description "Gaming setup with Steam"
  nixai community share ./home.nix --category desktop --tags gaming,multimedia`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		description, _ := cmd.Flags().GetString("description")
		category, _ := cmd.Flags().GetString("category")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		runCommunityShare(configFile, description, category, tags, cmd)
	},
}

// communityValidateCmd validates configuration against best practices
var communityValidateCmd = &cobra.Command{
	Use:   "validate <config-file>",
	Short: "Validate configuration against community best practices",
	Long: `Validate your NixOS configuration against community best practices and standards.

This command analyzes your configuration for:
- Security vulnerabilities and misconfigurations
- Performance optimization opportunities
- Maintainability and code quality issues
- Compliance with NixOS conventions
- Compatibility with different system configurations

The validation includes:
- Static analysis of Nix expressions
- Security pattern detection
- Performance impact assessment
- Best practice recommendations
- Community feedback integration

Examples:
  nixai community validate ./configuration.nix
  nixai community validate ./flake.nix --detailed
  nixai community validate ./home.nix --fix-suggestions`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		detailed, _ := cmd.Flags().GetBool("detailed")
		fixSuggestions, _ := cmd.Flags().GetBool("fix-suggestions")

		runCommunityValidate(configFile, detailed, fixSuggestions, cmd)
	},
}

// communityTrendsCmd shows trending packages and configurations
var communityTrendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Show trending packages and configuration patterns",
	Long: `Display trending packages, configurations, and patterns in the NixOS community.

This command shows:
- Most popular packages by download/usage
- Trending configuration patterns
- Recent community activity
- Emerging tools and frameworks
- Configuration quality trends

Data sources include:
- Hydra build statistics
- GitHub repository activity
- Community forum discussions
- Package manager telemetry (anonymized)

Examples:
  nixai community trends
  nixai community trends --timeframe weekly
  nixai community trends --category desktop
  nixai community trends --detailed`,
	Run: func(cmd *cobra.Command, args []string) {
		timeframe, _ := cmd.Flags().GetString("timeframe")
		category, _ := cmd.Flags().GetString("category")
		detailed, _ := cmd.Flags().GetBool("detailed")

		runCommunityTrends(timeframe, category, detailed, cmd)
	},
}

// communityRateCmd rate and review community configurations
var communityRateCmd = &cobra.Command{
	Use:   "rate <config-name> <rating>",
	Short: "Rate and review community configurations",
	Long: `Rate and provide feedback on community-shared configurations.

This command allows you to:
- Rate configurations on a 1-5 scale
- Provide detailed reviews and feedback
- Report issues or suggestions
- View existing ratings and reviews
- Help maintain configuration quality

Rating criteria:
- Functionality (does it work as expected?)
- Code quality (clean, readable, maintainable?)
- Documentation (well-documented and explained?)
- Security (follows security best practices?)
- Performance (efficient resource usage?)

Examples:
  nixai community rate "gaming-setup-v2" 5 --comment "Excellent configuration, works perfectly"
  nixai community rate "server-config" 4 --comment "Good but needs better documentation"
  nixai community rate "kde-plasma-setup" 3`,
	Args: conditionalExactArgsValidator(2),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]
		ratingStr := args[1]
		comment, _ := cmd.Flags().GetString("comment")

		rating, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil || rating < 1 || rating > 5 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Rating must be a number between 1 and 5"))
			return
		}

		runCommunityRate(configName, rating, comment, cmd)
	},
}

// Implementation functions

func runCommunitySearch(query string, limit int, category string, cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("🔍 Community Search: "+query))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Error loading config: "+err.Error()))
		return
	}

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Context detection failed: "+err.Error()))
		nixosCtx = nil
	}

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatNote("📋 "+contextSummary))
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// Create community manager
	manager := community.NewManager(cfg)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Searching community configurations..."))

	// Enhance search with context if available
	var contextualQuery string
	if nixosCtx != nil && nixosCtx.CacheValid {
		// Build context-aware search query
		contextInfo := fmt.Sprintf("NixOS %s", nixosCtx.NixOSVersion)
		if nixosCtx.UsesFlakes {
			contextInfo += " (flakes)"
		}
		if nixosCtx.HasHomeManager {
			contextInfo += " with Home Manager"
		}
		contextualQuery = query + " (compatible with " + contextInfo + ")"
	} else {
		contextualQuery = query
	}

	var results []community.Configuration
	if category != "" {
		results, err = manager.SearchByCategory(category, contextualQuery, limit)
	} else {
		results, err = manager.SearchConfigurations(contextualQuery, limit)
	}

	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Search failed: "+err.Error()))
		return
	}

	if len(results) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("No configurations found matching: "+query))
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Try broader search terms or different category"))
		return
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess(fmt.Sprintf("Found %d configuration(s):", len(results))))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	for i, config := range results {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s. %s\n",
			utils.FormatNote(fmt.Sprintf("%d", i+1)),
			utils.FormatKeyValue(config.Name, config.Description))

		if config.Author != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", utils.FormatNote("Author: "+config.Author))
		}

		if config.Rating > 0 {
			stars := strings.Repeat("⭐", int(config.Rating))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", utils.FormatNote(fmt.Sprintf("Rating: %s (%.1f/5)", stars, config.Rating)))
		}

		if len(config.Tags) > 0 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", utils.FormatNote("Tags: "+strings.Join(config.Tags, ", ")))
		}

		if config.URL != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", utils.FormatNote("🔗 "+config.URL))
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai community validate <file>' to check your configuration"))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai community share <file>' to contribute your configuration"))
}

func runCommunityShare(configFile, description, category string, tags []string, cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("📤 Sharing Configuration: "+filepath.Base(configFile)))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// Validate file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Configuration file not found: "+configFile))
		return
	}

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Error loading config: "+err.Error()))
		return
	}

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Context detection failed: "+err.Error()))
		nixosCtx = nil
	}

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatNote("📋 "+contextSummary))
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// Create community manager
	manager := community.NewManager(cfg)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Validating configuration before sharing..."))

	// Validate configuration first
	validation, err := manager.ValidateConfiguration(configFile)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Validation failed: "+err.Error()))
		return
	}

	if !validation.IsValid {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Configuration has validation issues:"))
		for _, issue := range validation.Issues {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  • "+issue)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai community validate "+configFile+"' for detailed analysis"))
		return
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess("✅ Configuration validation passed"))

	// Create configuration object
	config := &community.Configuration{
		Name:        filepath.Base(configFile),
		Description: description,
		Tags:        tags,
		Author:      "anonymous", // Could be enhanced to get from git config
		Rating:      0,
		URL:         "", // Will be populated after sharing
	}

	// Enhance metadata with context if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		// Add context-specific tags
		contextTags := []string{}
		if nixosCtx.UsesFlakes {
			contextTags = append(contextTags, "flakes")
		}
		if nixosCtx.UsesChannels {
			contextTags = append(contextTags, "channels")
		}
		if nixosCtx.HasHomeManager {
			contextTags = append(contextTags, "home-manager")
		}
		if nixosCtx.NixOSVersion != "" {
			contextTags = append(contextTags, "nixos-"+nixosCtx.NixOSVersion)
		}

		// Merge context tags with user-provided tags
		config.Tags = append(config.Tags, contextTags...)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Sharing configuration with community..."))

	// Share configuration
	err = manager.ShareConfiguration(config)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Sharing failed: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess("🎉 Configuration shared successfully!"))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Name", config.Name))
	if description != "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Description", description))
	}
	if category != "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Category", category))
	}
	if len(tags) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Tags", strings.Join(tags, ", ")))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Your configuration is now available for others to discover and use"))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Check 'nixai community trends' to see how it's performing"))
}

func runCommunityValidate(configFile string, detailed, fixSuggestions bool, cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("🔍 Validating Configuration: "+filepath.Base(configFile)))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// Validate file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Configuration file not found: "+configFile))
		return
	}

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Error loading config: "+err.Error()))
		return
	}

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Context detection failed: "+err.Error()))
		nixosCtx = nil
	}

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatNote("📋 "+contextSummary))
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// Create community manager
	manager := community.NewManager(cfg)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Analyzing configuration against best practices..."))

	// Validate configuration
	result, err := manager.ValidateConfiguration(configFile)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Validation failed: "+err.Error()))
		return
	}

	// Display results
	if result.IsValid {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess("✅ Configuration validation passed"))
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("⚠️  Configuration has issues"))
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Overall Score", fmt.Sprintf("%.1f/10", result.Score)))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	if len(result.Issues) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("❌ Issues Found", ""))
		for _, issue := range result.Issues {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  • "+issue)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	if len(result.Suggestions) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("💡 Suggestions", ""))
		for _, suggestion := range result.Suggestions {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  • "+suggestion)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	if detailed && len(result.BestPractices) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("📋 Best Practices Applied", ""))
		for _, practice := range result.BestPractices {
			status := "📋" // Show all practices without Applied field
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s %s - %s\n", status, practice.Title, practice.Description)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	if fixSuggestions {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("🔧 AI-Powered Fix Suggestions", ""))

		// Get AI suggestions for fixes
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		} else {
			prompt := buildValidationFixPrompt(configFile, result, nixosCtx)

			// Build context-aware prompt if context is available
			if nixosCtx != nil && nixosCtx.CacheValid {
				contextBuilder := nixoscontext.NewNixOSContextBuilder()
				prompt = contextBuilder.BuildContextualPrompt(prompt, nixosCtx)
			}

			suggestions, aiErr := aiProvider.Query(prompt)
			if aiErr != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Failed to get AI suggestions: "+aiErr.Error()))
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.RenderMarkdown(suggestions))
			}
		}
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai community share "+configFile+"' to contribute after fixing issues"))
}

func runCommunityTrends(timeframe, category string, detailed bool, cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("📊 Community Trends"))
	if timeframe != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), " (%s)\n", timeframe)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Error loading config: "+err.Error()))
		return
	}

	// Create community manager
	manager := community.NewManager(cfg)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Fetching community trends data..."))

	// Get trends
	trends, err := manager.GetTrends()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Failed to fetch trends: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("🔥 Popular Packages", ""))
	for i, pkg := range trends.PopularPackages {
		if i >= 10 { // Show top 10
			break
		}
		stars := strings.Repeat("⭐", int(pkg.Rating))
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1,
			utils.FormatKeyValue(pkg.Name, pkg.Description),
		)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s | %s downloads\n",
			stars, utils.FormatNote(fmt.Sprintf("%d", pkg.Downloads)))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("🚀 Trending Configurations", ""))
	for i, config := range trends.TrendingConfigs {
		if i >= 8 { // Show top 8
			break
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1,
			utils.FormatKeyValue(config.Name, config.Description))
		if config.Author != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", utils.FormatNote("by "+config.Author))
		}
		if len(config.Tags) > 0 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", utils.FormatNote("Tags: "+strings.Join(config.Tags, ", ")))
		}
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	if detailed {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("📈 Community Statistics", ""))
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Total Configurations", fmt.Sprintf("%d", trends.TotalConfigurations)))
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Active Contributors", fmt.Sprintf("%d", trends.ActiveContributors)))
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Packages Tracked", fmt.Sprintf("%d", trends.PackagesTracked)))
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Last Updated", trends.LastUpdated.Format("2006-01-02 15:04:05")))
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai community search <package>' to find configurations using trending packages"))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai community validate <file>' to check your configuration against trends"))
}

func runCommunityRate(configName string, rating float64, comment string, cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("⭐ Rating Configuration: "+configName))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Error loading config: "+err.Error()))
		return
	}

	// Create community manager
	manager := community.NewManager(cfg)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Submitting rating..."))

	// Submit rating
	err = manager.RateConfiguration(configName, rating, comment)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatError("Failed to submit rating: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSuccess("✅ Rating submitted successfully!"))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Configuration", configName))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Rating", fmt.Sprintf("%.1f/5 %s", rating, strings.Repeat("⭐", int(rating)))))
	if comment != "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatKeyValue("Comment", comment))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Thank you for contributing to the community!"))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Your feedback helps others find quality configurations"))
}

// Helper functions

func buildValidationFixPrompt(configFile string, result *community.ValidationResult, nixosCtx *config.NixOSContext) string {
	basePrompt := fmt.Sprintf(`Please analyze this NixOS configuration validation result and provide specific, actionable fix suggestions:

Configuration File: %s
Validation Score: %.1f/10

Issues Found:
%s

Current Suggestions:
%s

Please provide:
1. Specific code fixes for each issue
2. Example configuration snippets
3. Best practice explanations
4. Security improvements if applicable
5. Performance optimization suggestions

Focus on practical, copy-pasteable solutions that follow NixOS conventions.`,
		configFile,
		result.Score,
		strings.Join(result.Issues, "\n"),
		strings.Join(result.Suggestions, "\n"))

	// Add context information if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextInfo := fmt.Sprintf(`

Current System Context:
- NixOS Version: %s
- Uses Flakes: %t
- Uses Channels: %t
- Has Home Manager: %t (%s)
- System Type: %s

Please tailor your suggestions to this specific system configuration.`,
			nixosCtx.NixOSVersion,
			nixosCtx.UsesFlakes,
			nixosCtx.UsesChannels,
			nixosCtx.HasHomeManager,
			nixosCtx.HomeManagerType,
			nixosCtx.SystemType)

		basePrompt += contextInfo
	}

	return basePrompt
}

// Add community commands to the main community command
func init() {
	// Add community subcommands
	communityCmd.AddCommand(communitySearchCmd)
	communityCmd.AddCommand(communityShareCmd)
	communityCmd.AddCommand(communityValidateCmd)
	communityCmd.AddCommand(communityTrendsCmd)
	communityCmd.AddCommand(communityRateCmd)

	// Add flags to search command
	communitySearchCmd.Flags().IntP("limit", "l", 20, "Maximum number of results to show")
	communitySearchCmd.Flags().StringP("category", "c", "", "Filter by category (desktop, server, development, etc.)")

	// Add flags to share command
	communityShareCmd.Flags().StringP("description", "d", "", "Description of the configuration")
	communityShareCmd.Flags().StringP("category", "c", "", "Category for the configuration")
	communityShareCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags for the configuration")

	// Add flags to validate command
	communityValidateCmd.Flags().BoolP("detailed", "d", false, "Show detailed validation report")
	communityValidateCmd.Flags().BoolP("fix-suggestions", "f", false, "Get AI-powered fix suggestions")

	// Add flags to trends command
	communityTrendsCmd.Flags().StringP("timeframe", "t", "weekly", "Timeframe for trends (daily, weekly, monthly)")
	communityTrendsCmd.Flags().StringP("category", "c", "", "Filter trends by category")
	communityTrendsCmd.Flags().BoolP("detailed", "d", false, "Show detailed statistics")

	// Add flags to rate command
	communityRateCmd.Flags().StringP("comment", "c", "", "Optional comment with your rating")
}
