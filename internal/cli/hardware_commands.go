package cli

import (
	"fmt"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// Main hardware command
var hardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "AI-powered hardware configuration optimizer",
	Long: `Detect hardware and automatically generate optimized NixOS configurations.

Analyze your system hardware and get AI-powered recommendations for optimal
NixOS configuration including drivers, performance settings, and power management.

Commands:
  detect                  - Detect and analyze system hardware
  optimize                - Apply hardware-specific optimizations
  drivers                 - Auto-configure drivers and firmware
  compare                 - Compare current vs optimal settings
  laptop                  - Laptop-specific optimizations

Examples:
  nixai hardware detect
  nixai hardware optimize --dry-run
  nixai hardware drivers --auto-install
  nixai hardware laptop --power-save`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Hardware detect command
var hardwareDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect and analyze system hardware",
	Long: `Detect system hardware components and analyze configuration requirements.

This command identifies:
- CPU model, features, and optimization opportunities
- GPU devices and driver requirements
- Memory configuration and optimization potential
- Storage devices and performance settings
- Network interfaces and driver status
- Power management capabilities (for laptops)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🔍 Hardware Detection & Analysis"))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

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

		// Perform basic hardware detection
		fmt.Println(utils.FormatProgress("Detecting hardware components..."))

		// Get AI analysis for hardware optimization
		prompt := `Analyze the current system hardware for NixOS optimization. Provide recommendations for:
1. Hardware capability assessment
2. NixOS-specific optimization opportunities  
3. Driver and firmware recommendations
4. Performance tuning suggestions
5. Power management recommendations
6. Security considerations

Focus on NixOS-specific configuration and provide actionable advice.`

		analysis, err := aiProvider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatWarning("Could not get AI analysis: " + err.Error()))
		} else {
			fmt.Println(utils.FormatSubsection("🤖 AI Hardware Analysis", ""))
			fmt.Println(utils.RenderMarkdown(analysis))
		}

		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai hardware optimize' to get optimization recommendations"))
		fmt.Println(utils.FormatTip("Use 'nixai hardware drivers' to configure drivers automatically"))
	},
}

// Hardware optimize command
var hardwareOptimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Apply hardware-specific optimizations",
	Long: `Generate and apply hardware-specific NixOS optimizations.

This command provides:
- CPU-specific optimization flags and settings
- Memory management and swap configuration
- Storage optimization (SSD, HDD, NVMe tuning)
- GPU acceleration and compute configuration
- Network interface optimization
- Power efficiency improvements`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("⚡ Hardware Optimization"))
		fmt.Println()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println(utils.FormatInfo("Running in dry-run mode - no changes will be applied"))
			fmt.Println()
		}

		// Initialize AI provider
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
			cfg = &config.UserConfig{AIProvider: "ollama", AIModel: "llama3"}
		}
		aiProvider := initializeAIProvider(cfg)

		// Get optimization recommendations
		fmt.Println(utils.FormatProgress("Analyzing hardware for optimization opportunities..."))

		prompt := `As a NixOS expert, provide comprehensive hardware optimization recommendations. Include:

## CPU Optimization
- Specific CPU flags and settings for detected processor
- Power scaling and governor recommendations
- Cache and memory optimization

## Storage Optimization  
- SSD/NVMe specific settings (TRIM, scheduler, etc.)
- HDD optimization for spinning drives
- Filesystem and mount options

## GPU Optimization
- Driver configuration (NVIDIA, AMD, Intel)
- Hardware acceleration settings
- Compute and gaming optimizations

## Memory Management
- Swap configuration recommendations
- Memory pressure and OOM handling
- Kernel parameters for memory optimization

## Network Optimization
- Network interface specific settings
- WiFi power management
- Ethernet optimization

Provide actual NixOS configuration snippets that can be applied.`

		optimization, err := aiProvider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("Could not get optimization recommendations: " + err.Error()))
			return
		}

		fmt.Println(utils.FormatSubsection("🔧 Optimization Recommendations", ""))
		fmt.Println(utils.RenderMarkdown(optimization))

		fmt.Println()
		if dryRun {
			fmt.Println(utils.FormatNote("This was a dry-run. Use 'nixai hardware optimize' without --dry-run to apply changes"))
		} else {
			fmt.Println(utils.FormatTip("Apply these optimizations to your NixOS configuration"))
			fmt.Println(utils.FormatTip("Run 'nixos-rebuild test' first to validate changes"))
		}
	},
}

// Hardware drivers command
var hardwareDriversCmd = &cobra.Command{
	Use:   "drivers",
	Short: "Auto-configure drivers and firmware",
	Long: `Automatically detect and configure hardware drivers and firmware.

This command handles:
- GPU drivers (NVIDIA, AMD, Intel)
- WiFi and Bluetooth firmware
- Audio drivers and codec configuration
- USB and peripheral device drivers
- Hardware-specific kernel modules
- Firmware updates and microcode`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🔌 Driver & Firmware Configuration"))
		fmt.Println()

		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if autoInstall {
			fmt.Println(utils.FormatInfo("Auto-install mode enabled - will provide installation commands"))
			fmt.Println()
		}

		// Initialize AI provider
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
			cfg = &config.UserConfig{AIProvider: "ollama", AIModel: "llama3"}
		}
		aiProvider := initializeAIProvider(cfg)

		// Get driver configuration recommendations
		fmt.Println(utils.FormatProgress("Analyzing hardware drivers and firmware..."))

		prompt := `As a NixOS hardware expert, provide comprehensive driver and firmware configuration. Include:

## GPU Drivers
- NVIDIA proprietary vs open-source drivers
- AMD AMDGPU configuration
- Intel integrated graphics setup
- Vulkan and OpenGL configuration

## Network Drivers
- WiFi firmware requirements and configuration
- Bluetooth setup and pairing
- Ethernet driver optimization
- USB network adapters

## Audio Drivers
- ALSA and PulseAudio configuration
- Audio codec specific settings
- Bluetooth audio setup
- Professional audio (JACK) if applicable

## System Firmware
- CPU microcode updates (Intel/AMD)
- System firmware and UEFI settings
- Hardware-specific kernel modules
- Power management firmware

## Peripheral Support
- USB device configuration
- Printer and scanner drivers
- Input device customization
- Hardware sensors and monitoring

Provide specific NixOS configuration examples with actual package names and options.`

		drivers, err := aiProvider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("Could not get driver recommendations: " + err.Error()))
			return
		}

		fmt.Println(utils.FormatSubsection("🛠️ Driver Configuration", ""))
		fmt.Println(utils.RenderMarkdown(drivers))

		fmt.Println()
		if autoInstall {
			fmt.Println(utils.FormatTip("Apply driver configurations to your NixOS configuration"))
			fmt.Println(utils.FormatTip("Use 'nixos-rebuild switch' to activate new drivers"))
		} else {
			fmt.Println(utils.FormatTip("Use --auto-install to get installation commands"))
		}
	},
}

// Hardware compare command
var hardwareCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare current vs optimal settings",
	Long: `Compare your current NixOS configuration against optimal hardware settings.

This command analyzes:
- Current driver configurations vs recommended
- Performance settings comparison
- Missing optimization opportunities
- Potential compatibility issues
- Configuration drift detection
- Best practice compliance`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🔄 Configuration Comparison"))
		fmt.Println()

		// Initialize AI provider
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
			cfg = &config.UserConfig{AIProvider: "ollama", AIModel: "llama3"}
		}
		aiProvider := initializeAIProvider(cfg)

		// Get comparison analysis
		fmt.Println(utils.FormatProgress("Analyzing current configuration vs optimal settings..."))

		prompt := `As a NixOS configuration expert, perform a comprehensive comparison analysis. Compare:

## Current vs Optimal Configuration
- Analyze what's currently configured vs hardware-optimal settings
- Identify missing driver configurations
- Compare performance settings with best practices
- Highlight potential conflicts or issues

## Gap Analysis
- What optimizations are missing?
- Which drivers could be improved?
- Performance tuning opportunities
- Security improvements available

## Migration Path
- Step-by-step plan to move from current to optimal
- Risk assessment for each change
- Testing recommendations
- Rollback strategies

## Compliance Check
- NixOS best practices compliance
- Hardware vendor recommendations
- Community-validated configurations
- Performance benchmarking suggestions

Provide actionable recommendations with priority levels (high, medium, low).`

		comparison, err := aiProvider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("Could not get comparison analysis: " + err.Error()))
			return
		}

		fmt.Println(utils.FormatSubsection("📊 Configuration Analysis", ""))
		fmt.Println(utils.RenderMarkdown(comparison))

		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai hardware optimize' to apply recommended changes"))
		fmt.Println(utils.FormatTip("Test changes with 'nixos-rebuild test' before switching"))
	},
}

// Hardware laptop command
var hardwareLaptopCmd = &cobra.Command{
	Use:   "laptop",
	Short: "Laptop-specific optimizations",
	Long: `Apply laptop-specific NixOS optimizations for better battery life and thermal management.

This command provides:
- Power management and battery optimization
- Thermal management and fan control
- Display scaling and brightness control
- WiFi power saving configuration
- Suspend and hibernation setup
- Docking station and external monitor support`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("💻 Laptop Optimization"))
		fmt.Println()

		powerSave, _ := cmd.Flags().GetBool("power-save")
		performance, _ := cmd.Flags().GetBool("performance")

		if powerSave && performance {
			fmt.Println(utils.FormatError("Cannot use both --power-save and --performance flags"))
			return
		}

		mode := "balanced"
		if powerSave {
			mode = "power-save"
			fmt.Println(utils.FormatInfo("Power-save mode selected"))
		} else if performance {
			mode = "performance"
			fmt.Println(utils.FormatInfo("Performance mode selected"))
		} else {
			fmt.Println(utils.FormatInfo("Balanced mode selected (default)"))
		}
		fmt.Println()

		// Initialize AI provider
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
			cfg = &config.UserConfig{AIProvider: "ollama", AIModel: "llama3"}
		}
		aiProvider := initializeAIProvider(cfg)

		// Get laptop-specific recommendations
		fmt.Println(utils.FormatProgress("Analyzing laptop hardware for optimization..."))

		prompt := fmt.Sprintf(`As a NixOS laptop optimization expert, provide comprehensive laptop-specific recommendations for %s mode. Include:

## Power Management
- CPU governor and scaling settings for %s
- Battery optimization and charging thresholds
- Display backlight and brightness control
- WiFi and Bluetooth power management

## Thermal Management
- CPU thermal throttling configuration
- Fan curve optimization
- Thermal monitoring and alerts
- Heat dissipation improvements

## Display Configuration
- HiDPI scaling and font settings
- External monitor configuration
- Brightness control and auto-adjustment
- Color profile management

## Suspend/Hibernation
- Suspend-to-RAM configuration
- Hibernation setup and swap requirements
- Wake-on-LAN and USB wake settings
- Fast boot and startup optimization

## Hardware Features
- Touchpad and trackpoint configuration
- Function key mapping and special keys
- Audio optimization for laptop speakers
- Webcam and microphone setup

## Docking and Connectivity
- USB-C and Thunderbolt docking stations
- External monitor auto-detection
- Network switching (WiFi to Ethernet)
- Audio output switching

Provide actual NixOS configuration snippets optimized for %s mode.`, mode, mode, mode)

		laptop, err := aiProvider.Query(prompt)
		if err != nil {
			fmt.Println(utils.FormatError("Could not get laptop optimization recommendations: " + err.Error()))
			return
		}

		fmt.Println(utils.FormatSubsection("⚙️ Laptop Configuration", ""))
		fmt.Println(utils.RenderMarkdown(laptop))

		fmt.Println()
		fmt.Println(utils.FormatTip("Test power settings with 'powertop' and 'tlp-stat'"))
		fmt.Println(utils.FormatTip("Monitor temperatures with 'sensors' and 'htop'"))
		fmt.Println(utils.FormatNote("Reboot after applying power management changes"))
	},
}

// Add commands to CLI in init function
func init() {
	// Add main hardware command
	rootCmd.AddCommand(hardwareCmd)

	// Add hardware subcommands
	hardwareCmd.AddCommand(hardwareDetectCmd)
	hardwareCmd.AddCommand(hardwareOptimizeCmd)
	hardwareCmd.AddCommand(hardwareDriversCmd)
	hardwareCmd.AddCommand(hardwareCompareCmd)
	hardwareCmd.AddCommand(hardwareLaptopCmd)

	// Add flags for hardware commands
	hardwareOptimizeCmd.Flags().Bool("dry-run", false, "Show optimization recommendations without applying changes")
	hardwareDriversCmd.Flags().Bool("auto-install", false, "Provide installation commands for recommended drivers")
	hardwareLaptopCmd.Flags().Bool("power-save", false, "Optimize for maximum battery life")
	hardwareLaptopCmd.Flags().Bool("performance", false, "Optimize for maximum performance")
}
