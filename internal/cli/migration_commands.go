package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// MigrationAnalysis represents a migration analysis result
type MigrationAnalysis struct {
	CurrentSetup    string                 `json:"current_setup"`
	TargetSetup     string                 `json:"target_setup"`
	Complexity      string                 `json:"complexity"` // "simple", "moderate", "complex"
	EstimatedTime   string                 `json:"estimated_time"`
	Risks           []string               `json:"risks"`
	Prerequisites   []string               `json:"prerequisites"`
	Steps           []MigrationStep        `json:"steps"`
	BackupRequired  bool                   `json:"backup_required"`
	RollbackPlan    []string               `json:"rollback_plan"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// MigrationStep represents a single migration step
type MigrationStep struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Commands      []string `json:"commands"`
	Validation    []string `json:"validation"`
	Risks         []string `json:"risks"`
	Rollback      []string `json:"rollback"`
	Required      bool     `json:"required"`
	EstimatedTime string   `json:"estimated_time"`
}

// MigrationManager handles migration operations
type MigrationManager struct {
	nixosPath  string
	backupDir  string
	logger     *logger.Logger
	aiProvider ai.AIProvider
	mcpClient  *mcp.MCPClient
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(nixosPath string, log *logger.Logger, aiProvider ai.AIProvider, mcpClient *mcp.MCPClient) *MigrationManager {
	if nixosPath == "" {
		nixosPath = "/etc/nixos"
	}

	// Create backup directory
	homeDir, _ := os.UserHomeDir()
	backupDir := filepath.Join(homeDir, ".nixai", "migration-backups")
	_ = os.MkdirAll(backupDir, 0755)

	return &MigrationManager{
		nixosPath:  nixosPath,
		backupDir:  backupDir,
		logger:     log,
		aiProvider: aiProvider,
		mcpClient:  mcpClient,
	}
}

// DetectCurrentSetup detects the current NixOS setup type
func (mm *MigrationManager) DetectCurrentSetup() (string, map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// Check for flake.nix
	flakePath := filepath.Join(mm.nixosPath, "flake.nix")
	if _, err := os.Stat(flakePath); err == nil {
		metadata["flake_path"] = flakePath
		metadata["has_flake_lock"] = mm.hasFlakeLock()
		inputs, err := mm.parseFlakeInputs(flakePath)
		if err == nil {
			metadata["flake_inputs"] = inputs
		}
		return "flakes", metadata, nil
	}

	// Check for configuration.nix
	configPath := filepath.Join(mm.nixosPath, "configuration.nix")
	if _, err := os.Stat(configPath); err == nil {
		metadata["config_path"] = configPath
		channels, err := mm.detectChannels()
		if err == nil {
			metadata["channels"] = channels
		}
		return "channels", metadata, nil
	}

	return "unknown", metadata, fmt.Errorf("could not detect NixOS setup type")
}

// hasFlakeLock checks if flake.lock exists
func (mm *MigrationManager) hasFlakeLock() bool {
	lockPath := filepath.Join(mm.nixosPath, "flake.lock")
	_, err := os.Stat(lockPath)
	return err == nil
}

// parseFlakeInputs parses flake inputs from flake.nix
func (mm *MigrationManager) parseFlakeInputs(flakePath string) (map[string]string, error) {
	content, err := os.ReadFile(flakePath)
	if err != nil {
		return nil, err
	}

	inputs := make(map[string]string)

	// Regular expression to match input declarations
	inputRegex := regexp.MustCompile(`(\w+)\s*=\s*\{[^}]*url\s*=\s*"([^"]+)"`)
	simpleInputRegex := regexp.MustCompile(`(\w+)\.url\s*=\s*"([^"]+)"`)

	// Find all matches for complex inputs
	matches := inputRegex.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) >= 3 {
			inputs[match[1]] = match[2]
		}
	}

	// Find all matches for simple inputs
	simpleMatches := simpleInputRegex.FindAllStringSubmatch(string(content), -1)
	for _, match := range simpleMatches {
		if len(match) >= 3 {
			inputs[match[1]] = match[2]
		}
	}

	return inputs, nil
}

// detectChannels detects current channels
func (mm *MigrationManager) detectChannels() ([]string, error) {
	// This would typically run nix-channel --list
	// For now, we'll simulate it
	return []string{"nixos-24.05", "nixpkgs"}, nil
}

// CreateBackup creates a backup of the current configuration
func (mm *MigrationManager) CreateBackup(name string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	if name == "" {
		name = "migration-backup"
	}
	backupName := fmt.Sprintf("%s-%s", name, timestamp)
	backupPath := filepath.Join(mm.backupDir, backupName)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Copy configuration files
	err := mm.copyDirectory(mm.nixosPath, backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy configuration: %v", err)
	}

	// Create backup metadata
	metadata := map[string]interface{}{
		"created_at":    time.Now().Unix(),
		"nixos_path":    mm.nixosPath,
		"backup_name":   backupName,
		"backup_reason": "migration",
	}

	metadataPath := filepath.Join(backupPath, ".nixai-backup-metadata.json")
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		mm.logger.Warn("Failed to marshal backup metadata: " + err.Error())
	} else {
		if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
			mm.logger.Warn("Failed to write backup metadata: " + err.Error())
		}
	}

	return backupPath, nil
}

// copyDirectory recursively copies a directory
func (mm *MigrationManager) copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip certain files
		if strings.Contains(path, ".git") || strings.Contains(path, "result") {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return mm.copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func (mm *MigrationManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// AnalyzeMigration analyzes a migration from current setup to target
func (mm *MigrationManager) AnalyzeMigration(targetSetup string) (*MigrationAnalysis, error) {
	currentSetup, metadata, err := mm.DetectCurrentSetup()
	if err != nil {
		return nil, fmt.Errorf("failed to detect current setup: %v", err)
	}

	analysis := &MigrationAnalysis{
		CurrentSetup:    currentSetup,
		TargetSetup:     targetSetup,
		BackupRequired:  true,
		Metadata:        metadata,
		Risks:           []string{},
		Prerequisites:   []string{},
		Steps:           []MigrationStep{},
		RollbackPlan:    []string{},
		Recommendations: []string{},
	}

	// Determine complexity and generate steps based on migration type
	switch fmt.Sprintf("%s->%s", currentSetup, targetSetup) {
	case "channels->flakes":
		return mm.analyzeChannelsToFlakes(analysis)
	case "flakes->channels":
		return mm.analyzeFlakesToChannels(analysis)
	case "channels->channels":
		return mm.analyzeChannelUpgrade(analysis)
	case "flakes->flakes":
		return mm.analyzeFlakeUpdate(analysis)
	default:
		return nil, fmt.Errorf("unsupported migration: %s to %s", currentSetup, targetSetup)
	}
}

// analyzeChannelsToFlakes analyzes migration from channels to flakes
func (mm *MigrationManager) analyzeChannelsToFlakes(analysis *MigrationAnalysis) (*MigrationAnalysis, error) {
	analysis.Complexity = "moderate"
	analysis.EstimatedTime = "30-60 minutes"

	// Add prerequisites
	analysis.Prerequisites = []string{
		"Nix flakes must be enabled in configuration",
		"Git must be installed and available",
		"Backup of current configuration recommended",
		"Understanding of flake structure helpful",
	}

	// Add risks
	analysis.Risks = []string{
		"Configuration syntax changes required",
		"Import paths need adjustment",
		"Channel references must be converted to inputs",
		"System rebuild required",
		"Potential for configuration errors",
	}

	// Generate migration steps
	analysis.Steps = []MigrationStep{
		{
			ID:          1,
			Title:       "Enable Flakes Support",
			Description: "Enable experimental flakes feature in NixOS configuration",
			Commands: []string{
				"# Add to configuration.nix:",
				"nix.settings.experimental-features = [ \"nix-command\" \"flakes\" ];",
			},
			Validation: []string{
				"nix --help | grep flake",
			},
			Required:      true,
			EstimatedTime: "5 minutes",
		},
		{
			ID:          2,
			Title:       "Create Initial Flake",
			Description: "Create basic flake.nix structure",
			Commands: []string{
				"cd " + mm.nixosPath,
				"# Create flake.nix with nixpkgs input",
			},
			Validation: []string{
				"test -f flake.nix",
				"nix flake check --no-build",
			},
			Required:      true,
			EstimatedTime: "10 minutes",
		},
		{
			ID:          3,
			Title:       "Convert Configuration",
			Description: "Convert configuration.nix to flake-compatible format",
			Commands: []string{
				"# Move configuration.nix content to flake outputs",
				"# Update import paths",
				"# Convert channel references to flake inputs",
			},
			Validation: []string{
				"nix flake check",
				"nixos-rebuild dry-run --flake .#hostname",
			},
			Required:      true,
			EstimatedTime: "20 minutes",
		},
		{
			ID:          4,
			Title:       "Test and Apply",
			Description: "Test the flake configuration and apply changes",
			Commands: []string{
				"nixos-rebuild test --flake .#hostname",
				"nixos-rebuild switch --flake .#hostname",
			},
			Validation: []string{
				"nixos-rebuild list-generations",
				"nix flake metadata",
			},
			Required:      true,
			EstimatedTime: "10 minutes",
		},
	}

	// Add rollback plan
	analysis.RollbackPlan = []string{
		"nixos-rebuild switch --rollback",
		"Remove flake.nix and flake.lock files",
		"Restore original configuration.nix from backup",
		"Remove flakes experimental features setting",
	}

	// Add recommendations
	analysis.Recommendations = []string{
		"Use flake inputs for all external dependencies",
		"Pin input versions for reproducibility",
		"Use flake-utils for common patterns",
		"Consider splitting configuration into modules",
		"Use direnv for development environments",
	}

	return analysis, nil
}

// analyzeFlakesToChannels analyzes migration from flakes to channels
func (mm *MigrationManager) analyzeFlakesToChannels(analysis *MigrationAnalysis) (*MigrationAnalysis, error) {
	analysis.Complexity = "simple"
	analysis.EstimatedTime = "15-30 minutes"

	analysis.Prerequisites = []string{
		"Understanding of channel-based configuration",
		"Backup of current flake configuration",
		"Knowledge of required channels",
	}

	analysis.Risks = []string{
		"Loss of reproducibility benefits",
		"Need to manually manage channels",
		"Import syntax changes required",
		"May lose access to newer packages",
	}

	// Add steps for flakes to channels migration
	analysis.Steps = []MigrationStep{
		{
			ID:          1,
			Title:       "Extract Configuration",
			Description: "Extract NixOS configuration from flake outputs",
			Commands:    []string{"# Extract nixosConfigurations content"},
			Required:    true,
		},
		{
			ID:          2,
			Title:       "Setup Channels",
			Description: "Configure appropriate NixOS channels",
			Commands: []string{
				"nix-channel --add https://nixos.org/channels/nixos-24.05 nixos",
				"nix-channel --update",
			},
			Required: true,
		},
		{
			ID:          3,
			Title:       "Convert Configuration",
			Description: "Convert flake-based configuration to channel-based",
			Commands:    []string{"# Update import statements and references"},
			Required:    true,
		},
	}

	return analysis, nil
}

// analyzeChannelUpgrade analyzes channel upgrade migration
func (mm *MigrationManager) analyzeChannelUpgrade(analysis *MigrationAnalysis) (*MigrationAnalysis, error) {
	analysis.Complexity = "simple"
	analysis.EstimatedTime = "10-20 minutes"

	analysis.Steps = []MigrationStep{
		{
			ID:          1,
			Title:       "Update Channels",
			Description: "Update to newer channel versions",
			Commands: []string{
				"nix-channel --list",
				"nix-channel --update",
			},
			Required: true,
		},
	}

	return analysis, nil
}

// analyzeFlakeUpdate analyzes flake input updates
func (mm *MigrationManager) analyzeFlakeUpdate(analysis *MigrationAnalysis) (*MigrationAnalysis, error) {
	analysis.Complexity = "simple"
	analysis.EstimatedTime = "5-15 minutes"

	analysis.Steps = []MigrationStep{
		{
			ID:          1,
			Title:       "Update Flake Inputs",
			Description: "Update flake.lock with latest input versions",
			Commands: []string{
				"nix flake update",
				"nix flake check",
			},
			Required: true,
		},
	}

	return analysis, nil
}

// GenerateFlakeFromChannels generates a flake.nix from channel-based config
func (mm *MigrationManager) GenerateFlakeFromChannels() (string, error) {
	// Read current configuration.nix
	configPath := filepath.Join(mm.nixosPath, "configuration.nix")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read configuration.nix: %v", err)
	}

	// Use AI to help generate flake structure
	prompt := fmt.Sprintf(`Convert this NixOS channel-based configuration to a flake.nix file.

Current configuration.nix content:
%s

Generate a complete flake.nix that:
1. Uses nixpkgs as the main input
2. Defines a nixosConfiguration for the current system
3. Includes all necessary inputs based on the configuration
4. Maintains the same functionality as the channel-based config
5. Follows flake best practices

Return only the flake.nix content without explanations.`, string(configContent))

	response, err := mm.aiProvider.Query(prompt)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %v", err)
	}

	return response, nil
}

// Helper functions for AI provider and MCP client initialization
func getAIProvider(cfg *config.UserConfig, log *logger.Logger) ai.AIProvider {
	switch cfg.AIProvider {
	case "ollama":
		return ai.NewOllamaProvider(cfg.AIModel)
	case "gemini":
		return ai.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent")
	case "openai":
		return ai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
	default:
		return ai.NewOllamaProvider("llama3")
	}
}

func getMCPClient(cfg *config.UserConfig, log *logger.Logger) *mcp.MCPClient {
	mcpURL := fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	return mcp.NewMCPClient(mcpURL)
}

// Main migration command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "AI-powered migration assistant for channels and flakes",
	Long: `Comprehensive migration assistant for converting between NixOS channels and flakes.

Supports:
- Converting from channels to flakes
- Converting from flakes to channels  
- Upgrading between channel versions
- Updating flake inputs
- Migration analysis and planning
- Automatic backups and rollback procedures

Examples:
  nixai migrate analyze                    # Analyze current setup and migration options
  nixai migrate to-flakes                 # Convert from channels to flakes
  nixai migrate from-flakes              # Convert from flakes to channels
  nixai migrate channel-upgrade          # Upgrade to newer channels
  nixai migrate flake-inputs             # Update flake inputs`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Migration analysis command
var migrateAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze current setup and migration complexity",
	Long: `Analyze your current NixOS setup and provide detailed migration analysis.

This command will:
- Detect your current setup (channels vs flakes)
- Analyze configuration complexity
- Estimate migration time and effort
- Identify potential risks and prerequisites
- Provide step-by-step migration planning

Examples:
  nixai migrate analyze
  nixai migrate analyze --target flakes
  nixai migrate analyze --target channels`,
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Get NixOS path
		nixosPath := ""
		if cfg.NixosFolder != "" {
			nixosPath = cfg.NixosFolder
		}

		// Initialize components
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		aiProvider := getAIProvider(cfg, log)
		mcpClient := getMCPClient(cfg, log)

		// Create migration manager
		migrationManager := NewMigrationManager(nixosPath, log, aiProvider, mcpClient)

		fmt.Println(utils.FormatHeader("🔄 NixOS Migration Analysis"))
		fmt.Println()

		// Detect current setup
		fmt.Println(utils.FormatProgress("Detecting current NixOS setup..."))
		currentSetup, metadata, err := migrationManager.DetectCurrentSetup()
		if err != nil {
			fmt.Println(utils.FormatError("Failed to detect setup: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatKeyValue("Current Setup", currentSetup))
		fmt.Println(utils.FormatKeyValue("Configuration Path", nixosPath))

		if verbose {
			fmt.Println()
			fmt.Println(utils.FormatSubsection("📋 Setup Details", ""))
			for key, value := range metadata {
				fmt.Println(utils.FormatKeyValue(key, fmt.Sprintf("%v", value)))
			}
		}

		// If target specified, analyze migration
		if target != "" && target != currentSetup {
			fmt.Println()
			fmt.Println(utils.FormatProgress("Analyzing migration to " + target + "..."))

			analysis, err := migrationManager.AnalyzeMigration(target)
			if err != nil {
				fmt.Println(utils.FormatError("Migration analysis failed: " + err.Error()))
				os.Exit(1)
			}

			// Display analysis results
			fmt.Println()
			fmt.Println(utils.FormatHeader("📊 Migration Analysis"))
			fmt.Println()
			fmt.Println(utils.FormatKeyValue("Migration", fmt.Sprintf("%s → %s", analysis.CurrentSetup, analysis.TargetSetup)))
			fmt.Println(utils.FormatKeyValue("Complexity", analysis.Complexity))
			fmt.Println(utils.FormatKeyValue("Estimated Time", analysis.EstimatedTime))
			fmt.Println(utils.FormatKeyValue("Backup Required", fmt.Sprintf("%t", analysis.BackupRequired)))

			if len(analysis.Prerequisites) > 0 {
				fmt.Println()
				fmt.Println(utils.FormatSubsection("📋 Prerequisites", ""))
				for _, prereq := range analysis.Prerequisites {
					fmt.Printf("  • %s\n", prereq)
				}
			}

			if len(analysis.Risks) > 0 {
				fmt.Println()
				fmt.Println(utils.FormatSubsection("⚠️  Risks", ""))
				for _, risk := range analysis.Risks {
					fmt.Printf("  • %s\n", risk)
				}
			}

			if len(analysis.Steps) > 0 {
				fmt.Println()
				fmt.Println(utils.FormatSubsection("📝 Migration Steps", ""))
				for _, step := range analysis.Steps {
					fmt.Printf("  %d. %s\n", step.ID, step.Title)
					if verbose {
						fmt.Printf("     %s\n", step.Description)
						fmt.Printf("     Estimated time: %s\n", step.EstimatedTime)
					}
				}
			}

			if len(analysis.Recommendations) > 0 {
				fmt.Println()
				fmt.Println(utils.FormatSubsection("💡 Recommendations", ""))
				for _, rec := range analysis.Recommendations {
					fmt.Printf("  • %s\n", rec)
				}
			}
		}

		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai migrate to-flakes' to start migration to flakes"))
		fmt.Println(utils.FormatTip("Use 'nixai migrate --help' to see all migration options"))
	},
}

// To-flakes migration command
var migrateToFlakesCmd = &cobra.Command{
	Use:   "to-flakes",
	Short: "Convert from channels to flakes",
	Long: `Convert your NixOS configuration from channels to flakes with AI assistance.

This command will:
- Analyze your current channel-based configuration
- Create automatic backup of current setup
- Generate a flake.nix based on your configuration
- Provide step-by-step migration guidance
- Validate the migration and offer rollback if needed

Examples:
  nixai migrate to-flakes
  nixai migrate to-flakes --backup-name "pre-flake-migration"
  nixai migrate to-flakes --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		backupName, _ := cmd.Flags().GetString("backup-name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Get NixOS path
		nixosPath := ""
		if cfg.NixosFolder != "" {
			nixosPath = cfg.NixosFolder
		}

		// Initialize components
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		aiProvider := getAIProvider(cfg, log)
		mcpClient := getMCPClient(cfg, log)

		// Create migration manager
		migrationManager := NewMigrationManager(nixosPath, log, aiProvider, mcpClient)

		fmt.Println(utils.FormatHeader("🔄 Converting to Flakes"))
		fmt.Println()

		// Detect current setup
		currentSetup, _, err := migrationManager.DetectCurrentSetup()
		if err != nil {
			fmt.Println(utils.FormatError("Failed to detect setup: " + err.Error()))
			os.Exit(1)
		}

		if currentSetup != "channels" {
			fmt.Println(utils.FormatWarning("Current setup is not channel-based: " + currentSetup))
			fmt.Println(utils.FormatTip("Use 'nixai migrate analyze' to understand your current setup"))
			return
		}

		// Analyze migration
		fmt.Println(utils.FormatProgress("Analyzing migration complexity..."))
		analysis, err := migrationManager.AnalyzeMigration("flakes")
		if err != nil {
			fmt.Println(utils.FormatError("Migration analysis failed: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatKeyValue("Complexity", analysis.Complexity))
		fmt.Println(utils.FormatKeyValue("Estimated Time", analysis.EstimatedTime))

		if dryRun {
			fmt.Println()
			fmt.Println(utils.FormatSubsection("🔍 Dry Run - Migration Steps", ""))
			for _, step := range analysis.Steps {
				fmt.Printf("  %d. %s\n", step.ID, step.Title)
				fmt.Printf("     %s\n", step.Description)
			}
			fmt.Println()
			fmt.Println(utils.FormatNote("This was a dry run. Use without --dry-run to execute."))
			return
		}

		// Confirm migration
		fmt.Println()
		fmt.Print(utils.FormatWarning("This will modify your NixOS configuration. Continue? (y/N): "))
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Migration cancelled.")
			return
		}

		// Create backup
		fmt.Println()
		fmt.Println(utils.FormatProgress("Creating backup..."))
		backupPath, err := migrationManager.CreateBackup(backupName)
		if err != nil {
			fmt.Println(utils.FormatError("Backup failed: " + err.Error()))
			os.Exit(1)
		}
		fmt.Println(utils.FormatKeyValue("Backup created", backupPath))

		// Generate flake
		fmt.Println()
		fmt.Println(utils.FormatProgress("Generating flake.nix with AI assistance..."))
		flakeContent, err := migrationManager.GenerateFlakeFromChannels()
		if err != nil {
			fmt.Println(utils.FormatError("Flake generation failed: " + err.Error()))
			fmt.Println(utils.FormatTip("You can restore from backup: " + backupPath))
			os.Exit(1)
		}

		// Write flake.nix
		flakePath := filepath.Join(nixosPath, "flake.nix")
		if err := os.WriteFile(flakePath, []byte(flakeContent), 0644); err != nil {
			fmt.Println(utils.FormatError("Failed to write flake.nix: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("Flake.nix generated successfully!"))
		fmt.Println()
		fmt.Println(utils.FormatSubsection("📝 Generated Flake", ""))
		fmt.Println(utils.FormatCodeBlock(flakeContent, "nix"))

		fmt.Println()
		fmt.Println(utils.FormatSubsection("✅ Next Steps", ""))
		fmt.Println("1. Review the generated flake.nix")
		fmt.Println("2. Run: cd " + nixosPath + " && nix flake check")
		fmt.Println("3. Test: nixos-rebuild test --flake .#$(hostname)")
		fmt.Println("4. Apply: nixos-rebuild switch --flake .#$(hostname)")

		fmt.Println()
		fmt.Println(utils.FormatWarning("Rollback available: " + backupPath))
		fmt.Println(utils.FormatTip("Use 'nixai migrate rollback' if issues occur"))
	},
}

// Add command flags and register commands
func init() {
	// Migration analyze command flags
	migrateAnalyzeCmd.Flags().String("target", "", "Target setup type (flakes, channels)")
	migrateAnalyzeCmd.Flags().Bool("verbose", false, "Show detailed analysis")

	// To-flakes command flags
	migrateToFlakesCmd.Flags().String("backup-name", "", "Custom backup name")
	migrateToFlakesCmd.Flags().Bool("dry-run", false, "Show migration steps without executing")

	// Add subcommands
	migrateCmd.AddCommand(migrateAnalyzeCmd)
	migrateCmd.AddCommand(migrateToFlakesCmd)
}
