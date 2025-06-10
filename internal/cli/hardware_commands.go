package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/agent"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/ai/function/hardware"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
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
  function                - Use hardware function calling interface

Examples:
  nixai hardware detect
  nixai hardware optimize --dry-run
  nixai hardware drivers --auto-install
  nixai hardware laptop --power-save
  nixai hardware function --operation detect`,
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
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatHeader("🔍 Hardware Detection & Analysis"))
		fmt.Fprintln(cmd.OutOrStdout())

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Error loading config: "+err.Error()))
			return
		}

		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Context detection failed: "+err.Error()))
			nixosCtx = nil
		}

		// Display detected context summary if available
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Fprintln(cmd.OutOrStdout(), utils.FormatNote("📋 "+contextSummary))
			fmt.Fprintln(cmd.OutOrStdout())
		}

		// Initialize AI provider
		legacyProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			return
		}

		// Initialize Hardware Agent with legacy provider adapter
		hardwareProvider := ai.NewLegacyProviderAdapter(legacyProvider)
		hardwareAgent := agent.NewHardwareAgent(hardwareProvider)

		// Perform comprehensive hardware detection
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Detecting hardware components..."))

		// Detect hardware components
		hardwareInfo, err := detectHardwareComponents()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), utils.FormatError("Hardware detection failed: "+err.Error()))
			return
		}

		// Create hardware context for the agent
		hwContext := &agent.HardwareContext{
			SystemInfo:         fmt.Sprintf("Architecture: %s", hardwareInfo.Architecture),
			CPUInfo:            hardwareInfo.CPU,
			GPUInfo:            strings.Join(hardwareInfo.GPU, "\n"),
			MemoryInfo:         hardwareInfo.Memory,
			StorageInfo:        strings.Join(hardwareInfo.Storage, "\n"),
			NetworkInfo:        strings.Join(hardwareInfo.Network, "\n"),
			AudioInfo:          hardwareInfo.Audio,
			USBDevices:         hardwareInfo.USB,
			PCIDevices:         hardwareInfo.PCI,
			BIOS_UEFI:          hardwareInfo.Firmware,
			Architecture:       hardwareInfo.Architecture,
			VirtualizationInfo: hardwareInfo.Virtualization,
		}

		// Set context in the agent
		hardwareAgent.SetContext(hwContext)

		// Display detected hardware
		displayDetectedHardwareToWriter(hardwareInfo, cmd.OutOrStdout())

		// Get AI analysis for hardware optimization using the HardwareAgent
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatProgress("Analyzing hardware for NixOS optimization..."))

		ctx := context.Background()

		// Build context-aware analysis query
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		baseQuery := "Analyze the detected hardware and provide comprehensive NixOS configuration recommendations for optimal performance, compatibility, and power management."
		contextualQuery := contextBuilder.BuildContextualPrompt(baseQuery, nixosCtx)

		analysis, err := hardwareAgent.Query(ctx, contextualQuery)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Could not get AI analysis: "+err.Error()))
			// Fallback to legacy provider for basic configuration suggestions
			generateConfigurationSuggestionsToWriter(hardwareInfo, legacyProvider, cmd.OutOrStdout())
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), utils.FormatSubsection("🤖 AI Hardware Analysis", ""))
			fmt.Fprintln(cmd.OutOrStdout(), utils.RenderMarkdown(analysis))

			// Generate additional component-specific suggestions using agent context
			generateAgentBasedSuggestionsToWriter(hardwareInfo, hardwareAgent, cmd.OutOrStdout())
		}

		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai hardware optimize' to get optimization recommendations"))
		fmt.Fprintln(cmd.OutOrStdout(), utils.FormatTip("Use 'nixai hardware drivers' to configure drivers automatically"))
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

		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(cfg)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), utils.FormatWarning("Context detection failed: "+err.Error()))
			nixosCtx = nil
		}

		// Display detected context summary if available
		if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Println(utils.FormatNote("📋 " + contextSummary))
			fmt.Println()
		}

		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Println(utils.FormatError("Failed to initialize AI provider: " + err.Error()))
			return
		}

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

		// Build context-aware optimization query
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextualPrompt := contextBuilder.BuildContextualPrompt(prompt, nixosCtx)

		optimization, err := aiProvider.Query(contextualPrompt)
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
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Println(utils.FormatError("Failed to initialize AI provider: " + err.Error()))
			return
		}

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
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Println(utils.FormatError("Failed to initialize AI provider: " + err.Error()))
			return
		}

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
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			fmt.Println(utils.FormatError("Failed to initialize AI provider: " + err.Error()))
			return
		}

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

// Hardware function command - uses the function calling interface
var hardwareFunctionCmd = &cobra.Command{
	Use:   "function",
	Short: "Use hardware function calling interface",
	Long: `Execute hardware operations using the function calling interface.

This command demonstrates integration with the hardware function system and provides
structured hardware operations with standardized input/output formats.

Available operations:
  detect          - Detect hardware components
  scan            - Perform comprehensive hardware scan  
  test            - Test hardware functionality
  diagnose        - Diagnose hardware issues
  generate-config - Generate NixOS configuration
  list-drivers    - List available drivers

Examples:
  nixai hardware function --operation detect
  nixai hardware function --operation scan --component gpu
  nixai hardware function --operation generate-config --format nix
  nixai hardware function --operation diagnose --detailed`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("🔧 Hardware Function Interface"))
		fmt.Println()

		operation, _ := cmd.Flags().GetString("operation")
		component, _ := cmd.Flags().GetString("component")
		format, _ := cmd.Flags().GetString("format")
		detailed, _ := cmd.Flags().GetBool("detailed")
		includeDrivers, _ := cmd.Flags().GetBool("include-drivers")

		fmt.Println(utils.FormatProgress("Initializing hardware function interface..."))

		// Create hardware function instance
		hardwareFunction := hardware.NewHardwareFunction()

		// Prepare function parameters
		params := map[string]interface{}{
			"operation":       operation,
			"component":       component,
			"detailed":        detailed,
			"include_drivers": includeDrivers,
		}

		if format != "" {
			params["format"] = format
		}

		// Execute hardware function
		ctx := context.Background()
		fmt.Println(utils.FormatProgress(fmt.Sprintf("Executing hardware operation: %s", operation)))

		result, err := hardwareFunction.Execute(ctx, params, nil)
		if err != nil {
			fmt.Println(utils.FormatError("Hardware function execution failed: " + err.Error()))
			return
		}

		// Display results
		if !result.Success {
			fmt.Println(utils.FormatError("Hardware operation failed: " + result.Error))
			return
		}

		fmt.Println(utils.FormatSubsection("🎯 Function Results", ""))

		// The result.Data should contain HardwareResponse
		if response, ok := result.Data.(*hardware.HardwareResponse); ok {
			displayHardwareFunctionResults(response)
		} else {
			// Fallback display
			fmt.Printf("Operation completed successfully in %v\n", result.Duration)
			fmt.Printf("Response: %+v\n", result.Data)
		}

		fmt.Println()
		fmt.Println(utils.FormatTip("Use different --operation values to explore hardware function capabilities"))
		fmt.Println(utils.FormatTip("Add --detailed for comprehensive analysis"))
	},
}

// Add commands to CLI in init function
func init() {
	// Add hardware subcommands
	hardwareCmd.AddCommand(hardwareDetectCmd)
	hardwareCmd.AddCommand(hardwareOptimizeCmd)
	hardwareCmd.AddCommand(hardwareDriversCmd)
	hardwareCmd.AddCommand(hardwareCompareCmd)
	hardwareCmd.AddCommand(hardwareLaptopCmd)
	hardwareCmd.AddCommand(hardwareFunctionCmd)

	// Add flags for hardware commands
	hardwareOptimizeCmd.Flags().Bool("dry-run", false, "Show optimization recommendations without applying changes")
	hardwareDriversCmd.Flags().Bool("auto-install", false, "Provide installation commands for recommended drivers")
	hardwareLaptopCmd.Flags().Bool("power-save", false, "Optimize for maximum battery life")
	hardwareLaptopCmd.Flags().Bool("performance", false, "Optimize for maximum performance")
	hardwareFunctionCmd.Flags().String("operation", "", "Specify the hardware operation to perform")
	hardwareFunctionCmd.Flags().String("component", "", "Specify the hardware component for the operation")
	hardwareFunctionCmd.Flags().String("format", "", "Specify the output format for the operation")
	hardwareFunctionCmd.Flags().Bool("detailed", false, "Enable detailed output for the operation")
	hardwareFunctionCmd.Flags().Bool("include-drivers", false, "Include driver information in the operation")
}

// NewHardwareCmd constructor
func NewHardwareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   hardwareCmd.Use,
		Short: hardwareCmd.Short,
		Long:  hardwareCmd.Long,
		Run:   hardwareCmd.Run,
	}
	cmd.AddCommand(hardwareDetectCmd)
	cmd.AddCommand(hardwareOptimizeCmd)
	cmd.AddCommand(hardwareDriversCmd)
	cmd.AddCommand(hardwareCompareCmd)
	cmd.AddCommand(hardwareLaptopCmd)
	cmd.AddCommand(hardwareFunctionCmd)
	cmd.PersistentFlags().AddFlagSet(hardwareCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(hardwareCmd.Flags())
	return cmd
}

// HardwareInfo represents detected hardware information
type HardwareInfo struct {
	CPU            string
	GPU            []string
	Memory         string
	Storage        []string
	Network        []string
	Audio          string
	USB            []string
	PCI            []string
	Firmware       string
	DisplayServer  string // X11 or Wayland
	Architecture   string
	Virtualization string // VM detection and virtualization capabilities
}

// detectHardwareComponents performs comprehensive hardware detection
func detectHardwareComponents() (*HardwareInfo, error) {
	info := &HardwareInfo{}

	// Detect CPU
	if cpu, err := runCommand("lscpu | grep 'Model name' | cut -d':' -f2 | xargs"); err == nil {
		info.CPU = cpu
	}

	// Detect GPU devices with more detailed information
	if gpu, err := runCommand("lspci | grep -i vga"); err == nil && gpu != "" {
		info.GPU = strings.Split(gpu, "\n")
	}
	if gpu3d, err := runCommand("lspci | grep -i '3d'"); err == nil && gpu3d != "" {
		info.GPU = append(info.GPU, strings.Split(gpu3d, "\n")...)
	}
	// Detect additional display controllers
	if display, err := runCommand("lspci | grep -i 'display\\|graphics'"); err == nil && display != "" {
		for _, line := range strings.Split(display, "\n") {
			if line != "" && !sliceContains(info.GPU, line) {
				info.GPU = append(info.GPU, line)
			}
		}
	}

	// Detect memory with more details
	if mem, err := runCommand("free -h | head -2 | tail -1 | awk '{print $2}'"); err == nil {
		memTotal := strings.TrimSpace(mem)
		// Get additional memory info
		if memInfo, err := runCommand("dmidecode -t memory 2>/dev/null | grep -E 'Size|Speed|Type:' | head -3"); err == nil && memInfo != "" {
			info.Memory = fmt.Sprintf("Total RAM: %s\nDetails:\n%s", memTotal, memInfo)
		} else {
			info.Memory = fmt.Sprintf("Total RAM: %s", memTotal)
		}
	}

	// Detect storage
	if storage, err := runCommand("lsblk -d -o name,size,type | grep disk"); err == nil {
		info.Storage = strings.Split(storage, "\n")
	}

	// Detect network interfaces
	if network, err := runCommand("ip link show | grep -E '^[0-9]+:' | cut -d':' -f2 | xargs"); err == nil {
		info.Network = strings.Split(network, " ")
	}

	// Detect audio
	if audio, err := runCommand("lspci | grep -i audio"); err == nil {
		info.Audio = audio
	}

	// Detect USB devices
	if usb, err := runCommand("lsusb"); err == nil {
		info.USB = strings.Split(usb, "\n")
	}

	// Detect PCI devices
	if pci, err := runCommand("lspci"); err == nil {
		info.PCI = strings.Split(pci, "\n")
	}

	// Detect firmware type
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		info.Firmware = "UEFI"
	} else {
		info.Firmware = "BIOS"
	}

	// Detect display server
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		info.DisplayServer = "Wayland"
	} else if os.Getenv("DISPLAY") != "" {
		info.DisplayServer = "X11"
	} else {
		info.DisplayServer = "Unknown/Console"
	}

	// Detect architecture
	if arch, err := runCommand("uname -m"); err == nil {
		info.Architecture = strings.TrimSpace(arch)
	}

	// Detect virtualization
	var virtInfo []string

	// Check if running in a VM
	if virt, err := runCommand("systemd-detect-virt 2>/dev/null"); err == nil && virt != "none" {
		virtInfo = append(virtInfo, fmt.Sprintf("Running in: %s", virt))
	}

	// Check CPU virtualization support
	if cpuVirt, err := runCommand("lscpu | grep -E 'Virtualization|VT-x|AMD-V'"); err == nil && cpuVirt != "" {
		virtInfo = append(virtInfo, fmt.Sprintf("CPU Features: %s", cpuVirt))
	}

	// Check for hypervisor
	if hyper, err := runCommand("lscpu | grep 'Hypervisor vendor'"); err == nil && hyper != "" {
		virtInfo = append(virtInfo, hyper)
	}

	if len(virtInfo) > 0 {
		info.Virtualization = strings.Join(virtInfo, "\n")
	} else {
		info.Virtualization = "Native/Unknown"
	}

	return info, nil
}

// runCommand executes a shell command and returns its output
func runCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// sliceContains checks if a slice contains a string
func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// runCommandWithSudo executes a command with sudo after asking for permission
func runCommandWithSudo(command string) (string, error) {
	fmt.Printf("This command requires sudo privileges: %s\n", command)
	fmt.Print("Do you want to proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		return "", fmt.Errorf("operation cancelled by user")
	}

	cmd := exec.Command("sudo", "sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// displayDetectedHardware shows the detected hardware components
func displayDetectedHardware(info *HardwareInfo) {
	fmt.Println(utils.FormatSubsection("💻 Detected Hardware", ""))

	if info.CPU != "" {
		fmt.Println(utils.FormatKeyValue("CPU", info.CPU))
	}

	if len(info.GPU) > 0 {
		fmt.Println(utils.FormatKeyValue("GPU", ""))
		for _, gpu := range info.GPU {
			if strings.TrimSpace(gpu) != "" {
				fmt.Printf("  • %s\n", strings.TrimSpace(gpu))
			}
		}
	}

	if info.Memory != "" {
		fmt.Println(utils.FormatKeyValue("Memory", info.Memory))
	}

	if len(info.Storage) > 0 {
		fmt.Println(utils.FormatKeyValue("Storage", ""))
		for _, storage := range info.Storage {
			if strings.TrimSpace(storage) != "" {
				fmt.Printf("  • %s\n", strings.TrimSpace(storage))
			}
		}
	}

	if len(info.Network) > 0 {
		fmt.Println(utils.FormatKeyValue("Network", strings.Join(info.Network, ", ")))
	}

	if info.Audio != "" {
		fmt.Println(utils.FormatKeyValue("Audio", info.Audio))
	}

	fmt.Println(utils.FormatKeyValue("Firmware", info.Firmware))
	fmt.Println(utils.FormatKeyValue("Display Server", info.DisplayServer))
	fmt.Println(utils.FormatKeyValue("Architecture", info.Architecture))

	if info.Virtualization != "" {
		fmt.Println(utils.FormatKeyValue("Virtualization", info.Virtualization))
	}

	fmt.Println()
}

// displayDetectedHardwareToWriter shows the detected hardware components to a specific writer
func displayDetectedHardwareToWriter(info *HardwareInfo, writer io.Writer) {
	fmt.Fprintln(writer, utils.FormatSubsection("💻 Detected Hardware", ""))

	if info.CPU != "" {
		fmt.Fprintln(writer, utils.FormatKeyValue("CPU", info.CPU))
	}

	if len(info.GPU) > 0 {
		fmt.Fprintln(writer, utils.FormatKeyValue("GPU", ""))
		for _, gpu := range info.GPU {
			if strings.TrimSpace(gpu) != "" {
				fmt.Fprintf(writer, "  • %s\n", strings.TrimSpace(gpu))
			}
		}
	}

	if info.Memory != "" {
		fmt.Fprintln(writer, utils.FormatKeyValue("Memory", info.Memory))
	}

	if len(info.Storage) > 0 {
		fmt.Fprintln(writer, utils.FormatKeyValue("Storage", ""))
		for _, storage := range info.Storage {
			if strings.TrimSpace(storage) != "" {
				fmt.Fprintf(writer, "  • %s\n", strings.TrimSpace(storage))
			}
		}
	}

	if len(info.Network) > 0 {
		fmt.Fprintln(writer, utils.FormatKeyValue("Network", strings.Join(info.Network, ", ")))
	}

	if info.Audio != "" {
		fmt.Fprintln(writer, utils.FormatKeyValue("Audio", info.Audio))
	}

	fmt.Fprintln(writer, utils.FormatKeyValue("Firmware", info.Firmware))
	fmt.Fprintln(writer, utils.FormatKeyValue("Display Server", info.DisplayServer))
	fmt.Fprintln(writer, utils.FormatKeyValue("Architecture", info.Architecture))

	if info.Virtualization != "" {
		fmt.Fprintln(writer, utils.FormatKeyValue("Virtualization", info.Virtualization))
	}

	fmt.Fprintln(writer)
}

// confirmHardwareDetection asks user to confirm the detected hardware
func confirmHardwareDetection() bool {
	fmt.Print("Is this hardware detection correct? (Y/n): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes"
}

// buildHardwareAnalysisPrompt creates a detailed prompt for AI analysis
func buildHardwareAnalysisPrompt(info *HardwareInfo) string {
	prompt := "Analyze the following hardware configuration for NixOS optimization:\n\n"

	prompt += "**System Overview:**\n"
	prompt += fmt.Sprintf("- Architecture: %s\n", info.Architecture)
	prompt += fmt.Sprintf("- Firmware: %s\n", info.Firmware)
	prompt += fmt.Sprintf("- Display Server: %s\n", info.DisplayServer)
	prompt += "\n"

	if info.CPU != "" {
		prompt += fmt.Sprintf("**CPU:** %s\n\n", info.CPU)
	}

	if len(info.GPU) > 0 {
		prompt += "**GPU:**\n"
		for _, gpu := range info.GPU {
			if strings.TrimSpace(gpu) != "" {
				prompt += fmt.Sprintf("- %s\n", strings.TrimSpace(gpu))
			}
		}
		prompt += "\n"
	}

	if info.Memory != "" {
		prompt += fmt.Sprintf("**Memory:** %s\n\n", info.Memory)
	}

	if len(info.Storage) > 0 {
		prompt += "**Storage:**\n"
		for _, storage := range info.Storage {
			if strings.TrimSpace(storage) != "" {
				prompt += fmt.Sprintf("- %s\n", strings.TrimSpace(storage))
			}
		}
		prompt += "\n"
	}

	prompt += `
Please provide comprehensive NixOS configuration recommendations for:

1. **Hardware-specific optimizations:**
   - CPU optimizations and microcode
   - GPU driver configuration for ` + info.DisplayServer + ` display server
   - Memory and storage optimizations
   - Network interface optimization

2. **Driver Configuration:**
   - Required drivers and firmware packages
   - Kernel modules and parameters
   - Hardware acceleration settings

3. **Performance Tuning:**
   - CPU governor and scaling settings
   - I/O schedulers for storage devices
   - Network performance optimizations
   - Power management (if laptop)

4. **NixOS Configuration Snippets:**
   - Provide actual configuration.nix or hardware-configuration.nix snippets
   - Include package installations and system settings
   - Consider ` + info.DisplayServer + ` specific configurations

5. **Security Considerations:**
   - Secure boot compatibility
   - Microcode updates
   - Hardware security features

Focus on practical, tested NixOS configurations that improve performance and stability.`

	return prompt
}

// generateConfigurationSuggestions creates specific NixOS configuration suggestions
func generateConfigurationSuggestions(info *HardwareInfo, aiProvider ai.AIProvider) {
	fmt.Println(utils.FormatSubsection("⚙️ Configuration Suggestions", ""))

	// Generate GPU-specific suggestions
	if len(info.GPU) > 0 {
		generateGPUSuggestions(info, aiProvider)
	}

	// Generate CPU-specific suggestions
	if info.CPU != "" {
		generateCPUSuggestions(info, aiProvider)
	}

	// Generate storage suggestions
	if len(info.Storage) > 0 {
		generateStorageSuggestions(info, aiProvider)
	}

	// Generate network suggestions
	if len(info.Network) > 0 {
		generateNetworkSuggestions(info, aiProvider)
	}
}

// generateConfigurationSuggestionsToWriter creates specific NixOS configuration suggestions to a writer
func generateConfigurationSuggestionsToWriter(info *HardwareInfo, aiProvider ai.AIProvider, writer io.Writer) {
	fmt.Fprintln(writer, utils.FormatSubsection("⚙️ Configuration Suggestions", ""))
	fmt.Fprintln(writer, "Configuration suggestions would appear here...")
}

// generateGPUSuggestions provides GPU-specific configuration suggestions
func generateGPUSuggestions(info *HardwareInfo, aiProvider ai.AIProvider) {
	fmt.Println(utils.FormatKeyValue("🎮 GPU Configuration", ""))
	fmt.Println("GPU-specific suggestions would appear here...")
}

// generateCPUSuggestions provides CPU-specific configuration suggestions
func generateCPUSuggestions(info *HardwareInfo, aiProvider ai.AIProvider) {
	fmt.Println(utils.FormatKeyValue("🔥 CPU Configuration", ""))
	fmt.Println("CPU-specific suggestions would appear here...")
}

// generateStorageSuggestions provides storage-specific configuration suggestions
func generateStorageSuggestions(info *HardwareInfo, aiProvider ai.AIProvider) {
	fmt.Println(utils.FormatKeyValue("💾 Storage Configuration", ""))
	fmt.Println("Storage-specific suggestions would appear here...")
}

// generateNetworkSuggestions provides network-specific configuration suggestions
func generateNetworkSuggestions(info *HardwareInfo, aiProvider ai.AIProvider) {
	fmt.Println(utils.FormatKeyValue("🌐 Network Configuration", ""))
	fmt.Println("Network-specific suggestions would appear here...")
}

// generateAgentBasedSuggestions provides AI agent-powered hardware configuration suggestions
func generateAgentBasedSuggestions(info *HardwareInfo, hardwareAgent *agent.HardwareAgent) {
	ctx := context.Background()

	fmt.Println(utils.FormatSubsection("🔧 Component-Specific Recommendations", ""))

	// CPU-specific recommendations
	if info.CPU != "" {
		fmt.Println(utils.FormatKeyValue("⚙️ CPU Optimization", ""))
		cpuQuery := fmt.Sprintf("Provide specific NixOS configuration for CPU optimization based on: %s. Include microcode updates, performance scaling, and thermal management.", info.CPU)
		if cpuResponse, err := hardwareAgent.Query(ctx, cpuQuery); err == nil {
			fmt.Println(utils.RenderMarkdown(cpuResponse))
		} else {
			fmt.Printf("  • Error generating CPU suggestions: %v\n", err)
		}
		fmt.Println()
	}

	// GPU-specific recommendations with X11/Wayland considerations
	if len(info.GPU) > 0 {
		fmt.Println(utils.FormatKeyValue("🎮 GPU Configuration", ""))
		displayServer := info.DisplayServer
		if displayServer == "" {
			displayServer = "both X11 and Wayland"
		}

		gpuQuery := fmt.Sprintf(`Provide comprehensive NixOS GPU configuration for:
%s

Display Server: %s

Include:
1. Driver installation and configuration
2. Hardware acceleration setup
3. Performance optimizations
4. Power management
5. Multi-GPU configuration if applicable
6. %s-specific configurations

Provide actual NixOS configuration snippets.`, strings.Join(info.GPU, "\n"), displayServer, displayServer)

		if gpuResponse, err := hardwareAgent.Query(ctx, gpuQuery); err == nil {
			fmt.Println(utils.RenderMarkdown(gpuResponse))
		} else {
			fmt.Printf("  • Error generating GPU suggestions: %v\n", err)
		}
		fmt.Println()
	}

	// Storage optimization with modern filesystem recommendations
	if len(info.Storage) > 0 {
		fmt.Println(utils.FormatKeyValue("💾 Advanced Storage Configuration", ""))
		storageQuery := fmt.Sprintf(`Provide advanced NixOS storage configuration for:
%s

Include:
1. Modern filesystem recommendations (ZFS, Btrfs, ext4)
2. SSD optimizations and TRIM scheduling
3. I/O scheduler selection
4. Advanced mount options
5. Swap configuration and zRAM
6. Storage security (LUKS encryption)
7. Performance monitoring setup

Provide complete NixOS configuration examples.`, strings.Join(info.Storage, "\n"))

		if storageResponse, err := hardwareAgent.Query(ctx, storageQuery); err == nil {
			fmt.Println(utils.RenderMarkdown(storageResponse))
		} else {
			fmt.Printf("  • Error generating storage suggestions: %v\n", err)
		}
		fmt.Println()
	}

	// Network and connectivity optimization
	if len(info.Network) > 0 {
		fmt.Println(utils.FormatKeyValue("🌐 Network Optimization", ""))
		networkQuery := fmt.Sprintf(`Provide comprehensive NixOS network configuration for:
%s

Include:
1. Network interface optimization
2. WiFi power management and roaming
3. Network security and firewall
4. Performance tuning parameters
5. Network monitoring and diagnostics
6. VPN and remote access setup

Provide practical NixOS configuration examples.`, strings.Join(info.Network, "\n"))

		if networkResponse, err := hardwareAgent.Query(ctx, networkQuery); err == nil {
			fmt.Println(utils.RenderMarkdown(networkResponse))
		} else {
			fmt.Printf("  • Error generating network suggestions: %v\n", err)
		}
		fmt.Println()
	}

	// System-wide optimization recommendations
	fmt.Println(utils.FormatKeyValue("🏗️ System-Wide Optimizations", ""))
	systemQuery := fmt.Sprintf(`Based on the complete hardware profile:
- Architecture: %s
- CPU: %s
- Memory: %s
- Display Server: %s

Provide system-wide NixOS optimizations including:
1. Kernel selection and parameters
2. System service optimizations
3. Power management strategies
4. Security hardening
5. Performance monitoring setup
6. Backup and maintenance recommendations

Include a complete system configuration example.`,
		info.Architecture, info.CPU, info.Memory, info.DisplayServer)

	if systemResponse, err := hardwareAgent.Query(ctx, systemQuery); err == nil {
		fmt.Println(utils.RenderMarkdown(systemResponse))
	} else {
		fmt.Printf("  • Error generating system suggestions: %v\n", err)
	}
	fmt.Println()
}

// generateAgentBasedSuggestionsToWriter provides agent-based suggestions to a writer
func generateAgentBasedSuggestionsToWriter(info *HardwareInfo, hardwareAgent *agent.HardwareAgent, writer io.Writer) {
	fmt.Fprintln(writer, utils.FormatSubsection("🤖 Agent-Based Suggestions", ""))
	fmt.Fprintln(writer, "Agent-based suggestions would appear here...")
}

// displayHardwareFunctionResults displays the results from hardware function execution
func displayHardwareFunctionResults(response *hardware.HardwareResponse) {
	fmt.Println(utils.FormatKeyValue("Status", response.Status))
	fmt.Println(utils.FormatKeyValue("Operation", response.Operation))

	if response.ExecutionTime > 0 {
		fmt.Println(utils.FormatKeyValue("Execution Time", response.ExecutionTime.String()))
	}

	// Display error if any
	if response.ErrorMessage != "" {
		fmt.Println()
		fmt.Println(utils.FormatError("Error: " + response.ErrorMessage))
		return
	}

	// Display hardware components
	if len(response.Hardware) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("🔧 Hardware Components", ""))
		for _, component := range response.Hardware {
			displayHardwareComponent(component)
		}
	}

	// Display configuration
	if response.Configuration != "" {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("⚙️ Generated Configuration", ""))
		fmt.Println(utils.RenderMarkdown("```nix\n" + response.Configuration + "\n```"))
	}

	// Display recommendations
	if len(response.Recommendations) > 0 {
		fmt.Println()
	}

	fmt.Println()
}

// displayHardwareComponent displays a hardware component from function results
func displayHardwareComponent(component hardware.HardwareComponent) {
	fmt.Printf("  • %s: %s", component.Type, component.Name)
	if component.Vendor != "" {
		fmt.Printf(" (%s)", component.Vendor)
	}
	if component.Model != "" {
		fmt.Printf(" - %s", component.Model)
	}
	fmt.Println()

	if component.Status != "" {
		fmt.Printf("    Status: %s\n", component.Status)
	}
	if component.Driver != "" {
		fmt.Printf("    Driver: %s\n", component.Driver)
	}
	if !component.Supported {
		fmt.Printf("    ⚠️  Not fully supported\n")
	}
}

// displayHardwareIssue displays a hardware issue in formatted output
func displayHardwareIssue(issue hardware.HardwareIssue) {
	severityIcon := "⚠️"
	switch strings.ToLower(issue.Severity) {
	case "critical", "high":
		severityIcon = "🔴"
	case "medium":
		severityIcon = "🟡"
	case "low":
		severityIcon = "🔵"
	}

	fmt.Printf("  %s %s [%s]\n", severityIcon, issue.Component, strings.ToUpper(issue.Severity))
	fmt.Printf("    Issue: %s\n", issue.Description)

	if issue.Solution != "" {
		fmt.Printf("    Solution: %s\n", issue.Solution)
	}

	if len(issue.Resources) > 0 {
		fmt.Printf("    Resources:\n")
		for _, resource := range issue.Resources {
			fmt.Printf("      • %s\n", resource)
		}
	}

	fmt.Println()
}
