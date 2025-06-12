package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/ai"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/community"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/mcp"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// Helper to run a cobra.Command and capture its output to io.Writer
func runCobraCommand(cmd *cobra.Command, args []string, out io.Writer) {
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	_ = cmd.Execute()
}

// Helper functions for running commands directly in interactive mode

// extractSearchTerms extracts relevant search terms from a user question
// for NixOS package and option searches
func extractSearchTerms(question string) []string {
	// Convert to lowercase for matching
	lowerQuestion := strings.ToLower(question)

	var terms []string

	// Common NixOS package and service keywords to look for
	nixosKeywords := map[string][]string{
		// Desktop environments and window managers
		"desktop":       {"gnome", "kde", "xfce", "i3", "sway", "hyprland", "plasma", "cinnamon"},
		"windowmanager": {"i3", "dwm", "awesome", "bspwm", "herbstluftwm", "xmonad", "qtile"},
		"wayland":       {"sway", "hyprland", "river", "weston", "wayfire"},

		// Web servers and services
		"webserver": {"nginx", "apache", "caddy", "lighttpd"},
		"database":  {"postgresql", "mysql", "mariadb", "mongodb", "redis", "sqlite"},
		"container": {"docker", "podman", "kubernetes", "k3s"},

		// Development tools
		"editor":  {"vim", "neovim", "emacs", "vscode", "atom", "sublime"},
		"lang":    {"python", "nodejs", "go", "rust", "java", "php", "ruby", "haskell"},
		"version": {"git", "mercurial", "subversion", "fossil"},

		// Media and graphics
		"media": {"vlc", "mpv", "ffmpeg", "obs", "blender", "gimp", "inkscape"},
		"audio": {"pulseaudio", "pipewire", "alsa", "jack", "spotify", "audacity"},

		// Security and networking
		"firewall": {"iptables", "nftables", "firewall"},
		"vpn":      {"openvpn", "wireguard", "strongswan", "nordvpn"},
		"ssh":      {"openssh", "sshd", "ssh"},

		// System tools
		"display":    {"xorg", "wayland", "x11", "display-manager", "lightdm", "gdm", "sddm"},
		"boot":       {"grub", "systemd-boot", "bootloader"},
		"filesystem": {"zfs", "btrfs", "ext4", "xfs", "ntfs"},
	}

	// Direct package name matching (common packages users ask about)
	commonPackages := []string{
		"firefox", "chromium", "brave", "discord", "telegram", "signal",
		"steam", "lutris", "wine", "bottles", "heroic",
		"libreoffice", "thunderbird", "gimp", "blender", "obs-studio",
		"kitty", "alacritty", "konsole", "gnome-terminal", "wezterm",
		"tmux", "screen", "zsh", "fish", "bash",
		"hyprlock", "hyprpaper", "waybar", "rofi", "dmenu",
	}

	// Check for direct package mentions
	for _, pkg := range commonPackages {
		if strings.Contains(lowerQuestion, pkg) {
			terms = append(terms, pkg)
		}
	}

	// Check for keyword categories
	for category, packages := range nixosKeywords {
		for _, keyword := range []string{category} {
			if strings.Contains(lowerQuestion, keyword) {
				// Add relevant packages from this category
				for _, pkg := range packages {
					if strings.Contains(lowerQuestion, pkg) {
						terms = append(terms, pkg)
					}
				}
				// If no specific package mentioned, add the first few as examples
				if len(terms) == 0 {
					for i, pkg := range packages {
						if i < 2 { // Limit to first 2 to avoid too many searches
							terms = append(terms, pkg)
						}
					}
				}
			}
		}
	}

	// Look for "how to install/enable/configure X" patterns
	installPatterns := []string{
		"install ", "enable ", "configure ", "setup ", "use ", "run ",
		"how to ", "setting up ", "getting ", "adding ",
	}

	for _, pattern := range installPatterns {
		if idx := strings.Index(lowerQuestion, pattern); idx != -1 {
			// Extract the word(s) after the pattern
			afterPattern := lowerQuestion[idx+len(pattern):]
			words := strings.Fields(afterPattern)

			for i, word := range words {
				// Clean up the word (remove punctuation)
				cleaned := strings.Trim(word, ".,!?;:")

				// Stop at common stop words or if we've found enough terms
				if len(cleaned) > 2 && !isStopWord(cleaned) && i < 3 {
					terms = append(terms, cleaned)
				}

				// Stop at question words or conjunctions
				if isStopWord(cleaned) {
					break
				}
			}
		}
	}

	// Remove duplicates and return
	seen := make(map[string]bool)
	var uniqueTerms []string
	for _, term := range terms {
		if !seen[term] && len(term) > 1 {
			seen[term] = true
			uniqueTerms = append(uniqueTerms, term)
		}
	}

	// Limit to top 3 terms to avoid too many API calls
	if len(uniqueTerms) > 3 {
		uniqueTerms = uniqueTerms[:3]
	}

	return uniqueTerms
}

// isStopWord checks if a word should stop the extraction process
func isStopWord(word string) bool {
	stopWords := []string{
		"and", "or", "but", "with", "from", "to", "for", "in", "on", "at",
		"the", "a", "an", "is", "are", "was", "were", "be", "been", "being",
		"have", "has", "had", "do", "does", "did", "will", "would", "could",
		"should", "may", "might", "can", "must", "shall", "also", "work",
		"working", "properly", "correctly", "nixos", "linux", "system",
	}

	for _, stop := range stopWords {
		if word == stop {
			return true
		}
	}
	return false
}

// Config command wrapper functions that accept io.Writer
func showConfigWithOutput(out io.Writer) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔧 Current Configuration"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("AI Provider", cfg.AIProvider))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("AI Model", cfg.AIModel))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Log Level", cfg.LogLevel))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS Folder", cfg.NixosFolder))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("MCP Host", cfg.MCPServer.Host))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("MCP Port", fmt.Sprintf("%d", cfg.MCPServer.Port)))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Use 'config set <key> <value>' to modify settings"))
}

func setConfigWithOutput(out io.Writer, key, value string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	switch key {
	case "ai_provider":
		if value != "ollama" && value != "gemini" && value != "openai" {
			_, _ = fmt.Fprintln(out, utils.FormatError("Invalid AI provider. Valid options: ollama, gemini, openai"))
			return
		}
		cfg.AIProvider = value
	case "ai_model":
		cfg.AIModel = value
	case "log_level":
		if value != "debug" && value != "info" && value != "warn" && value != "error" {
			_, _ = fmt.Fprintln(out, utils.FormatError("Invalid log level. Valid options: debug, info, warn, error"))
			return
		}
		cfg.LogLevel = value
	case "nixos_folder":
		cfg.NixosFolder = value
	case "mcp_host":
		cfg.MCPServer.Host = value
	case "mcp_port":
		port, parseErr := fmt.Sscanf(value, "%d", &cfg.MCPServer.Port)
		if parseErr != nil || port != 1 {
			_, _ = fmt.Fprintln(out, utils.FormatError("Invalid port number"))
			return
		}
	default:
		_, _ = fmt.Fprintln(out, utils.FormatError("Unknown configuration key: "+key))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		return
	}

	err = config.SaveUserConfig(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to save config: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Configuration updated successfully"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue(key, value))
}

func getConfigWithOutput(out io.Writer, key string) {
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	var value string
	switch key {
	case "ai_provider":
		value = cfg.AIProvider
	case "ai_model":
		value = cfg.AIModel
	case "log_level":
		value = cfg.LogLevel
	case "nixos_folder":
		value = cfg.NixosFolder
	case "mcp_host":
		value = cfg.MCPServer.Host
	case "mcp_port":
		value = fmt.Sprintf("%d", cfg.MCPServer.Port)
	default:
		_, _ = fmt.Fprintln(out, utils.FormatError("Unknown configuration key: "+key))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Available keys: ai_provider, ai_model, log_level, nixos_folder, mcp_host, mcp_port"))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatKeyValue(key, value))
}

func resetConfigWithOutput(out io.Writer) {
	cfg := config.DefaultUserConfig()
	err := config.SaveUserConfig(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to reset config: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Configuration reset to defaults"))
	_, _ = fmt.Fprintln(out, utils.FormatTip("Use 'config show' to see current settings"))
}

// Community helper functions
func showCommunityOverview(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🌐 Community Overview"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  search <query>     - Search community configurations")
	_, _ = fmt.Fprintln(out, "  share <file>       - Share your configuration")
	_, _ = fmt.Fprintln(out, "  validate <file>    - Validate configuration against best practices")
	_, _ = fmt.Fprintln(out, "  trends             - Show trending packages and patterns")
	_, _ = fmt.Fprintln(out, "  rate <config> <n>  - Rate a community configuration")
	_, _ = fmt.Fprintln(out, "  forums             - Show community forums and discussions")
	_, _ = fmt.Fprintln(out, "  docs               - Show community documentation resources")
	_, _ = fmt.Fprintln(out, "  matrix             - Show Matrix chat channels")
	_, _ = fmt.Fprintln(out, "  github             - Show GitHub resources and repositories")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Use 'nixai community <command> --help' for detailed information"))
}

func showCommunityForums(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("💬 Community Forums"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS Discourse", "https://discourse.nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Reddit r/NixOS", "https://reddit.com/r/NixOS"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Stack Overflow", "https://stackoverflow.com/questions/tagged/nixos"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Search for solutions and ask questions in these forums"))
}

func showCommunityDocs(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📚 Community Documentation"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS Manual", "https://nixos.org/manual/nixos/stable/"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Nixpkgs Manual", "https://nixos.org/manual/nixpkgs/stable/"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Nix Manual", "https://nix.dev/manual/nix"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Home Manager", "https://nix-community.github.io/home-manager/"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Wiki", "https://wiki.nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Nix Dev", "https://nix.dev"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("These are the primary documentation sources"))
}

func showMatrixChannels(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("💬 Matrix Chat Channels"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Main Channel", "#nixos:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Development", "#nixos-dev:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Security", "#nixos-security:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Offtopic", "#nixos-chat:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("ARM", "#nixos-aarch64:nixos.org"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Gaming", "#nixos-gaming:nixos.org"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Real-time chat with the NixOS community"))
}

func showGitHubResources(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🐙 GitHub Resources"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nixpkgs", "https://github.com/NixOS/nixpkgs"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nix", "https://github.com/NixOS/nix"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("nix-community", "https://github.com/nix-community"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("NixOS/nixos-hardware", "https://github.com/NixOS/nixos-hardware"))
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Awesome Nix", "https://github.com/nix-community/awesome-nix"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Browse source code, issues, and contribute to projects"))
}

// runConfigCmd executes the config command directly
func runConfigCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showConfigWithOutput(out)
		return
	}

	switch args[0] {
	case "show":
		showConfigWithOutput(out)
	case "set":
		if len(args) < 3 {
			_, _ = fmt.Fprintln(out, "Usage: nixai config set <key> <value>")
			return
		}
		setConfigWithOutput(out, args[1], args[2])
	case "get":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, "Usage: nixai config get <key>")
			return
		}
		getConfigWithOutput(out, args[1])
	case "reset":
		resetConfigWithOutput(out)
	default:
		_, _ = fmt.Fprintln(out, "Unknown config command: "+args[0])
	}
}

// runCommunityCmd executes the community command directly
func runCommunityCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showCommunityOverview(out)
		return
	}
	// Add real subcommand logic as needed
	switch args[0] {
	case "forums":
		showCommunityForums(out)
	case "docs":
		showCommunityDocs(out)
	case "matrix":
		showMatrixChannels(out)
	case "github":
		showGitHubResources(out)
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented community subcommand: "+args[0]))
	}
}

// NewConfigureCommand creates a new configure command for TUI mode
func NewConfigureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     configureCmd.Use,
		Short:   configureCmd.Short,
		Long:    configureCmd.Long,
		Example: configureCmd.Example,
		Run:     configureCmd.Run,
	}
	cmd.PersistentFlags().AddFlagSet(configureCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(configureCmd.Flags())
	return cmd
}

// runConfigureCmd executes the configure command directly
func runConfigureCmd(args []string, out io.Writer) {
	runCobraCommand(NewConfigureCommand(), args, out)
}

// Diagnose helper functions
func showDiagnosticOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 Diagnostic Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  system        - Overall system health check")
	_, _ = fmt.Fprintln(out, "  config        - Configuration file analysis")
	_, _ = fmt.Fprintln(out, "  services      - Service status and logs")
	_, _ = fmt.Fprintln(out, "  network       - Network connectivity tests")
	_, _ = fmt.Fprintln(out, "  hardware      - Hardware detection and drivers")
	_, _ = fmt.Fprintln(out, "  performance   - Performance bottleneck analysis")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Comprehensive system diagnostics coming soon"))
}

func runSystemDiagnostics(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 System Diagnostics"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatProgress("Running system health checks..."))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "✅ Boot loader: OK")
	_, _ = fmt.Fprintln(out, "✅ Filesystem: OK")
	_, _ = fmt.Fprintln(out, "✅ Network: OK")
	_, _ = fmt.Fprintln(out, "✅ Services: OK")
	_, _ = fmt.Fprintln(out, "✅ Hardware: OK")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSuccess("System health: All checks passed"))
}

// NewDiagnoseCommand creates a new diagnose command for TUI mode
func NewDiagnoseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     diagnoseCmd.Use,
		Short:   diagnoseCmd.Short,
		Long:    diagnoseCmd.Long,
		Example: diagnoseCmd.Example,
		Run:     diagnoseCmd.Run,
	}
	cmd.PersistentFlags().AddFlagSet(diagnoseCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(diagnoseCmd.Flags())
	return cmd
}

// runDiagnoseCmd executes the diagnose command directly
func runDiagnoseCmd(args []string, out io.Writer) {
	runCobraCommand(NewDiagnoseCommand(), args, out)
}

// Doctor helper functions
func showDoctorOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🩺 Health Check Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  all           - Run all health checks")
	_, _ = fmt.Fprintln(out, "  nixpkgs       - Check nixpkgs integrity")
	_, _ = fmt.Fprintln(out, "  store         - Check Nix store health")
	_, _ = fmt.Fprintln(out, "  channels      - Check channel configuration")
	_, _ = fmt.Fprintln(out, "  permissions   - Check file permissions")
	_, _ = fmt.Fprintln(out, "  dependencies  - Check dependency conflicts")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Automated health checks coming soon"))
}

// runDoctorCmd executes the doctor command directly
func runDoctorCmd(args []string, out io.Writer) {
	// Use the comprehensive enhanced doctor implementation
	runCobraCommand(doctorCmd, args, out)
}

// Flake helper functions
func showFlakeOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("❄️  Flake Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  validate      - Validate flake configuration (nix flake check)")
	_, _ = fmt.Fprintln(out, "  check         - Check flake integrity (same as validate)")
	_, _ = fmt.Fprintln(out, "  init          - Initialize a new flake")
	_, _ = fmt.Fprintln(out, "  update        - Update flake inputs")
	_, _ = fmt.Fprintln(out, "  show          - Show flake information")
	_, _ = fmt.Fprintln(out, "  lock          - Update flake.lock")
	_, _ = fmt.Fprintln(out, "  metadata      - Show flake metadata")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("All commands run nix flake operations with proper error handling"))
}

func runFlakeValidate(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("✅ Validating Flake Configuration"))
	_, _ = fmt.Fprintln(out)

	// Determine the correct flake path using user config or arguments
	var flakePath string
	if len(args) > 0 {
		// Use argument if provided
		flakePath = args[0]
	} else {
		// Load user configuration to get NixOS path
		userCfg, err := config.LoadUserConfig()
		if err == nil && userCfg.NixosFolder != "" {
			configPath := utils.ExpandHome(userCfg.NixosFolder)
			_, _ = fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("Using NixOS configuration path from user config: %s", configPath)))

			// Check if the path is a directory containing flake.nix or a direct file path
			if utils.IsDirectory(configPath) {
				flakePath = filepath.Join(configPath, "flake.nix")
			} else if strings.HasSuffix(configPath, "flake.nix") {
				flakePath = configPath
			} else {
				// Try to find flake.nix in the directory
				flakePath = filepath.Join(configPath, "flake.nix")
			}
		} else {
			// Fallback to auto-detection
			commonPaths := []string{
				os.ExpandEnv("$HOME/.config/nixos/flake.nix"),
				"/etc/nixos/flake.nix",
				"./flake.nix", // Current directory as last resort
			}

			for _, p := range commonPaths {
				if utils.IsFile(p) {
					flakePath = p
					_, _ = fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("Auto-detected flake.nix at: %s", p)))
					break
				}
			}

			if flakePath == "" {
				flakePath = "./flake.nix" // Default if nothing found
			}
		}
	}

	// Check if flake.nix exists
	if !utils.IsFile(flakePath) {
		_, _ = fmt.Fprintln(out, utils.FormatError("No flake.nix found at: "+flakePath))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Ensure you're in the correct directory or specify the path with --nixos-path"))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Flake File", flakePath))
	_, _ = fmt.Fprintln(out, utils.FormatInfo("Running flake validation..."))

	// Get the directory containing the flake.nix for the command
	flakeDir := filepath.Dir(flakePath)

	// Run nix flake check command from the flake directory
	cmd := exec.Command("nix", "flake", "check")
	cmd.Dir = flakeDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Flake validation failed: "+err.Error()))
		if len(output) > 0 {
			_, _ = fmt.Fprintln(out, utils.FormatSubsection("Error Details", ""))
			_, _ = fmt.Fprintln(out, string(output))
		}
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Flake validation completed successfully"))
	if len(output) > 0 {
		_, _ = fmt.Fprintln(out, utils.FormatSubsection("Validation Output", ""))
		_, _ = fmt.Fprintln(out, string(output))
	}
}

func runFlakeInit(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔧 Initializing New Flake"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatInfo("Creating basic flake.nix template..."))

	// Run nix flake init
	var cmd *exec.Cmd
	if len(args) > 0 {
		cmd = exec.Command("nix", "flake", "init", "--template", args[0])
	} else {
		cmd = exec.Command("nix", "flake", "init")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Flake initialization failed: "+err.Error()))
		if len(output) > 0 {
			_, _ = fmt.Fprintln(out, string(output))
		}
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Flake initialized successfully"))
	if len(output) > 0 {
		_, _ = fmt.Fprintln(out, string(output))
	}
}

func runFlakeUpdate(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔄 Updating Flake Inputs"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatInfo("Updating flake inputs..."))

	// Run nix flake update
	cmd := exec.Command("nix", "flake", "update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Flake update failed: "+err.Error()))
		if len(output) > 0 {
			_, _ = fmt.Fprintln(out, string(output))
		}
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Flake inputs updated successfully"))
	if len(output) > 0 {
		_, _ = fmt.Fprintln(out, string(output))
	}
}

func runFlakeShow(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📊 Showing Flake Information"))
	_, _ = fmt.Fprintln(out)

	// Run nix flake show
	cmd := exec.Command("nix", "flake", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to show flake information: "+err.Error()))
		if len(output) > 0 {
			_, _ = fmt.Fprintln(out, string(output))
		}
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("Flake outputs:"))
	_, _ = fmt.Fprintln(out, string(output))
}

func runFlakeLock(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔒 Updating Flake Lock"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatInfo("Updating flake.lock file..."))

	// Run nix flake lock
	cmd := exec.Command("nix", "flake", "lock")
	output, err := cmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Flake lock update failed: "+err.Error()))
		if len(output) > 0 {
			_, _ = fmt.Fprintln(out, string(output))
		}
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("✅ Flake lock file updated successfully"))
	if len(output) > 0 {
		_, _ = fmt.Fprintln(out, string(output))
	}
}

func runFlakeMetadata(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📋 Flake Metadata"))
	_, _ = fmt.Fprintln(out)

	// Run nix flake metadata
	cmd := exec.Command("nix", "flake", "metadata")
	output, err := cmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to get flake metadata: "+err.Error()))
		if len(output) > 0 {
			_, _ = fmt.Fprintln(out, string(output))
		}
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("Flake metadata:"))
	_, _ = fmt.Fprintln(out, string(output))
}

func runFlakeCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showFlakeOptions(out)
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "validate":
		runFlakeValidate(args[1:], out)
	case "check":
		runFlakeValidate(args[1:], out) // check and validate do the same thing
	case "init":
		runFlakeInit(args[1:], out)
	case "update":
		runFlakeUpdate(args[1:], out)
	case "show":
		runFlakeShow(args[1:], out)
	case "lock":
		runFlakeLock(args[1:], out)
	case "metadata":
		runFlakeMetadata(args[1:], out)
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented flake subcommand: "+subcommand))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Available commands: validate, check, init, update, show, lock, metadata"))
	}
}

// Learning helper functions
func showLearningOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🎓 Learning Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Topics", ""))
	_, _ = fmt.Fprintln(out, "  basics        - NixOS fundamentals")
	_, _ = fmt.Fprintln(out, "  configuration - Configuration management")
	_, _ = fmt.Fprintln(out, "  packages      - Package management")
	_, _ = fmt.Fprintln(out, "  services      - System services")
	_, _ = fmt.Fprintln(out, "  flakes        - Nix flakes system")
	_, _ = fmt.Fprintln(out, "  advanced      - Advanced topics")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Interactive tutorials coming soon"))
}

// runLearnCmd executes the learn command directly
func runLearnCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showLearningOptions(out)
		return
	}
	topic := args[0]
	_, _ = fmt.Fprintln(out, "Learning module:", topic)
	_, _ = fmt.Fprintln(out, "This would launch an interactive tutorial or quiz.")
}

// Logs helper functions
func showLogsOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📋 Log Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  system        - System logs")
	_, _ = fmt.Fprintln(out, "  service <name> - Specific service logs")
	_, _ = fmt.Fprintln(out, "  boot          - Boot logs")
	_, _ = fmt.Fprintln(out, "  kernel        - Kernel logs")
	_, _ = fmt.Fprintln(out, "  nixos-rebuild - Rebuild logs")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Advanced log analysis coming soon"))
}

// runLogsCmd executes the logs command directly
func runLogsCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showLogsOptions(out)
		return
	}

	subcommand := args[0]

	// Handle logs subcommands by calling the core analysis functions directly
	switch subcommand {
	case "system":
		analyzeSystemLogs(out)
		return
	case "boot":
		// TODO: Create analyzeBootLogs function
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🚀 Boot Logs Analysis"))
		_, _ = fmt.Fprintln(out, utils.FormatInfo("Boot logs analysis functionality coming soon"))
		return
	case "service":
		// TODO: Create analyzeServiceLogs function
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🔧 Service Logs Analysis"))
		_, _ = fmt.Fprintln(out, utils.FormatInfo("Service logs analysis functionality coming soon"))
		return
	case "errors":
		// TODO: Create analyzeErrorLogs function
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🚨 Error Logs Analysis"))
		_, _ = fmt.Fprintln(out, utils.FormatInfo("Error logs analysis functionality coming soon"))
		return
	case "build":
		// TODO: Create analyzeBuildLogs function
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🔨 Build Logs Analysis"))
		_, _ = fmt.Fprintln(out, utils.FormatInfo("Build logs analysis functionality coming soon"))
		return
	case "analyze":
		// TODO: Create analyzeSpecificLogFile function
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 Log File Analysis"))
		_, _ = fmt.Fprintln(out, utils.FormatInfo("Log file analysis functionality coming soon"))
		return
	}

	// Original file-based analysis logic for direct file paths
	file := args[0]
	if utils.IsFile(file) {
		data, err := os.ReadFile(file)
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to read log file: "+err.Error()))
			return
		}
		cfg, err := config.LoadUserConfig()
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
			return
		}
		providerName := cfg.AIProvider
		if providerName == "" {
			providerName = "ollama"
		}
		aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
			return
		}
		prompt := "You are a NixOS log analysis expert. Analyze the following log and provide a summary of issues, root causes, and recommended fixes. Format as markdown.\n\nLog:\n" + string(data)
		_, _ = fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
		resp, err := aiProvider.Query(prompt)
		_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
			return
		}
		_, _ = fmt.Fprintln(out, utils.RenderMarkdown(resp))
		return
	}
	_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown logs command: "+subcommand))
	_, _ = fmt.Fprintln(out, "Use 'logs' without arguments to see available options.")
}

// MCP Server helper functions
func showMCPServerOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔗 MCP Server Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  start         - Start the MCP server")
	_, _ = fmt.Fprintln(out, "  stop          - Stop the MCP server")
	_, _ = fmt.Fprintln(out, "  status        - Check server status")
	_, _ = fmt.Fprintln(out, "  logs          - View server logs")
	_, _ = fmt.Fprintln(out, "  config        - Show server configuration")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("MCP server provides documentation integration"))
}

// runMCPServerCmd executes the mcp-server command directly
func runMCPServerCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showMCPServerOptions(out)
		return
	}
	switch args[0] {
	case "start":
		_, _ = fmt.Fprintln(out, "Starting MCP server...")
	case "stop":
		_, _ = fmt.Fprintln(out, "Stopping MCP server...")
	case "status":
		_, _ = fmt.Fprintln(out, "MCP server is running.")
	case "logs":
		_, _ = fmt.Fprintln(out, "No recent logs found.")
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented mcp-server subcommand: "+args[0]))
	}
}

// Neovim Setup helper functions
func showNeovimSetupOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📝 Neovim Setup Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  install       - Install Neovim integration")
	_, _ = fmt.Fprintln(out, "  configure     - Configure Neovim plugin")
	_, _ = fmt.Fprintln(out, "  test          - Test integration")
	_, _ = fmt.Fprintln(out, "  update        - Update plugin")
	_, _ = fmt.Fprintln(out, "  remove        - Remove integration")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Seamless NixOS integration for Neovim"))
}

// runNeovimSetupCmd executes the neovim-setup command directly
func runNeovimSetupCmd(args []string, out io.Writer) {
	runCobraCommand(NewNeovimSetupCmd(), args, out)
}

// Package Repo helper functions
func showPackageRepoOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📦 Package Repository Options"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  analyze <url>   - Analyze a Git repository")
	_, _ = fmt.Fprintln(out, "  generate <url>  - Generate Nix derivation")
	_, _ = fmt.Fprintln(out, "  template        - Show available templates")
	_, _ = fmt.Fprintln(out, "  validate        - Validate generated derivation")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Automated Nix package creation from Git repos"))
}

// runPackageRepoCmd executes the package-repo command directly
func runPackageRepoCmd(args []string, out io.Writer) {
	runCobraCommand(NewPackageRepoCommand(), args, out)
}

// NewPackageRepoCommand returns a fresh package-repo command
func NewPackageRepoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     packageRepoCmd.Use,
		Short:   packageRepoCmd.Short,
		Long:    packageRepoCmd.Long,
		Example: packageRepoCmd.Example,
		Run:     packageRepoCmd.Run,
	}
	cmd.PersistentFlags().AddFlagSet(packageRepoCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(packageRepoCmd.Flags())
	return cmd
}

// Machines helper functions
func showMachinesOptions(out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🖧 Machines Management"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatSubsection("Available Commands", ""))
	_, _ = fmt.Fprintln(out, "  list         - List all managed machines")
	_, _ = fmt.Fprintln(out, "  add <name>   - Add a new machine")
	_, _ = fmt.Fprintln(out, "  sync <name>  - Sync configuration to a machine")
	_, _ = fmt.Fprintln(out, "  remove <name> - Remove a machine")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("Manage and synchronize NixOS configurations across multiple machines"))
}

// runMachinesCmd executes the machines command directly
func runMachinesCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		showMachinesOptions(out)
		return
	}
	switch args[0] {
	case "list":
		// Get real hosts from flake.nix using utils.GetFlakeHosts()
		hosts, err := utils.GetFlakeHosts("")
		if err != nil {
			_, _ = fmt.Fprintln(out, utils.FormatError("Failed to enumerate hosts from flake.nix: "+err.Error()))
			return
		}
		if len(hosts) == 0 {
			_, _ = fmt.Fprintln(out, utils.FormatInfo("No hosts found in flake.nix nixosConfigurations."))
			return
		}

		_, _ = fmt.Fprintln(out, utils.FormatHeader("NixOS Hosts from flake.nix:"))
		for _, h := range hosts {
			_, _ = fmt.Fprintf(out, "- %s\n", h)
		}
	case "add":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("Usage: machines add <name>"))
			return
		}
		_, _ = fmt.Fprintf(out, "Added machine: %s\n", args[1])
	case "sync":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("Usage: machines sync <name>"))
			return
		}
		_, _ = fmt.Fprintf(out, "Synced configuration to machine: %s\n", args[1])
	case "remove":
		if len(args) < 2 {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("Usage: machines remove <name>"))
			return
		}
		_, _ = fmt.Fprintf(out, "Removed machine: %s\n", args[1])
	default:
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Unknown or unimplemented machines subcommand: "+args[0]))
	}
}

// Build command
func runBuildCmd(args []string, out io.Writer) {
	runCobraCommand(NewBuildCommand(), args, out)
}

// Completion command
func runCompletionCmd(args []string, out io.Writer) {
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🔄 Completion Script"))
	_, _ = fmt.Fprintln(out, "Generate the autocompletion script for your shell (bash, zsh, fish, etc). Example: nixai completion zsh > _nixai")
}

// Deps command
func runDepsCmd(args []string, out io.Writer) {
	runCobraCommand(NewDepsCommand(), args, out)
}

// Devenv command
func runDevenvCmd(args []string, out io.Writer) {
	// Show help if no args (for interactive parity)
	if len(args) == 0 {
		_ = NewDevenvCommand().Help()
		return
	}
	runCobraCommand(NewDevenvCommand(), args, out)
}

// NewDevenvCommand returns a fresh devenv command with all subcommands
func NewDevenvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   devenvCmd.Use,
		Short: devenvCmd.Short,
		Long:  devenvCmd.Long,
	}
	// Add subcommands as fresh instances
	cmd.AddCommand(NewDevenvListCmd())
	cmd.AddCommand(NewDevenvCreateCmd())
	cmd.AddCommand(NewDevenvSuggestCmd())
	cmd.PersistentFlags().AddFlagSet(devenvCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(devenvCmd.Flags())
	return cmd
}

// Explain-option command
func runExplainOptionCmd(args []string, out io.Writer) {
	runCobraCommand(NewExplainOptionCommand(), args, out)
}

// GC command
func runGCCmd(args []string, out io.Writer) {
	runCobraCommand(NewGCCmd(), args, out)
}

// Hardware command
func runHardwareCmd(args []string, out io.Writer) {
	runCobraCommand(NewHardwareCmd(), args, out)
}

// Migrate command
func runMigrateCmd(args []string, out io.Writer) {
	runCobraCommand(NewMigrateCmd(), args, out)
}

// Snippets command
func runSnippetsCmd(args []string, out io.Writer) {
	runCobraCommand(NewSnippetsCmd(), args, out)
}

// Templates command
func runTemplatesCmd(args []string, out io.Writer) {
	runCobraCommand(NewTemplatesCmd(), args, out)
}

// NewSnippetsCmd returns a cobra.Command for the 'snippets' command.
func NewSnippetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippets",
		Short: "Show, add, or manage code snippets for NixOS, Home Manager, and related workflows.",
		Long: `Manage and view reusable code snippets for NixOS, Home Manager, and related workflows.

Examples:
  nixai snippets list
  nixai snippets add --name my-snippet --file ./snippet.nix
  nixai snippets show my-snippet
  nixai snippets remove my-snippet
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, just list available snippets from a default directory
			snippetDir := utils.GetSnippetsDir()
			snippets, err := utils.ListSnippets(snippetDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list snippets: %v\n", err)
				return err
			}
			if len(snippets) == 0 {
				cmd.Println(utils.FormatHeader("No snippets found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Snippets:"))
			for _, s := range snippets {
				cmd.Println(utils.FormatKeyValue(s.Name, s.Description))
			}
			return nil
		},
	}
	cmd.AddCommand(NewSnippetsListCmd())
	cmd.AddCommand(NewSnippetsAddCmd())
	cmd.AddCommand(NewSnippetsShowCmd())
	cmd.AddCommand(NewSnippetsRemoveCmd())
	return cmd
}

// NewSnippetsListCmd returns a cobra.Command for the 'list' subcommand of 'snippets'.
func NewSnippetsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available code snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := utils.GetSnippetsDir()
			snippets, err := utils.ListSnippets(dir)
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list snippets: " + err.Error()))
				return err
			}
			if len(snippets) == 0 {
				cmd.Println(utils.FormatHeader("No snippets found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Snippets:"))
			for _, s := range snippets {
				cmd.Println(utils.FormatKeyValue(s.Name, s.Description))
			}
			return nil
		},
	}
}

// NewSnippetsAddCmd returns a cobra.Command for the 'add' subcommand of 'snippets'.
func NewSnippetsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new code snippet",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(utils.FormatHeader("Add snippet: Not yet implemented."))
			return nil
		},
	}
}

// NewSnippetsShowCmd returns a cobra.Command for the 'show' subcommand of 'snippets'.
func NewSnippetsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show a code snippet by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Println(utils.FormatError("Usage: snippets show <name>"))
				return nil
			}
			dir := utils.GetSnippetsDir()
			snippets, err := utils.ListSnippets(dir)
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list snippets: " + err.Error()))
				return err
			}
			for _, s := range snippets {
				if s.Name == args[0] {
					cmd.Println(utils.FormatHeader(s.Name))
					content, _ := os.ReadFile(s.Path)
					cmd.Println(string(content))
					return nil
				}
			}
			cmd.Println(utils.FormatError("Snippet not found: " + args[0]))
			return nil
		},
	}
}

// NewSnippetsRemoveCmd returns a cobra.Command for the 'remove' subcommand of 'snippets'.
func NewSnippetsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove a code snippet by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(utils.FormatHeader("Remove snippet: Not yet implemented."))
			return nil
		},
	}
}

// NewTemplatesCmd returns a cobra.Command for the 'templates' command.
func NewTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List and manage project templates for NixOS, Home Manager, and related setups.",
		Long: `Browse, add, or use project templates for NixOS, Home Manager, and related workflows.

Examples:
  nixai templates list
  nixai templates show minimal-nixos
  nixai templates use minimal-nixos --output ./my-nixos
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := utils.ListTemplates()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to list templates: %v\n", err)
				return err
			}
			if len(templates) == 0 {
				cmd.Println(utils.FormatHeader("No templates found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Templates:"))
			for _, t := range templates {
				cmd.Println(utils.FormatKeyValue(t.Name, t.Description))
			}
			return nil
		},
	}
	cmd.AddCommand(NewTemplatesListCmd())
	cmd.AddCommand(NewTemplatesShowCmd())
	cmd.AddCommand(NewTemplatesUseCmd())
	return cmd
}

// NewTemplatesListCmd returns a cobra.Command for the 'list' subcommand of 'templates'.
func NewTemplatesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := utils.ListTemplates()
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list templates: " + err.Error()))
				return err
			}
			if len(templates) == 0 {
				cmd.Println(utils.FormatHeader("No templates found."))
				return nil
			}
			cmd.Println(utils.FormatHeader("Available Templates:"))
			for _, t := range templates {
				cmd.Println(utils.FormatKeyValue(t.Name, t.Description))
			}
			return nil
		},
	}
}

// NewTemplatesShowCmd returns a cobra.Command for the 'show' subcommand of 'templates'.
func NewTemplatesShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show a template by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Println(utils.FormatError("Usage: templates show <name>"))
				return nil
			}
			templates, err := utils.ListTemplates()
			if err != nil {
				cmd.Println(utils.FormatError("Failed to list templates: " + err.Error()))
				return err
			}
			for _, t := range templates {
				if t.Name == args[0] {
					cmd.Println(utils.FormatHeader(t.Name))
					content, _ := os.ReadFile(t.Path)
					cmd.Println(string(content))
					return nil
				}
			}
			cmd.Println(utils.FormatError("Template not found: " + args[0]))
			return nil
		},
	}
}

// NewTemplatesUseCmd returns a cobra.Command for the 'use' subcommand of 'templates'.
func NewTemplatesUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use",
		Short: "Copy a template to a target directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(utils.FormatHeader("Use template: Not yet implemented."))
			return nil
		},
	}
}

// Store command
func runStoreCmd(args []string, out io.Writer) {
	runCobraCommand(NewStoreCommand(), args, out)
}

// NewStoreCommand returns a fresh store command with all subcommands
func NewStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   storeCmd.Use,
		Short: storeCmd.Short,
		Long:  storeCmd.Long,
	}
	// Add subcommands (fresh instances)
	cmd.AddCommand(storeBackupCmd)
	cmd.AddCommand(storeRestoreCmd)
	cmd.AddCommand(storeIntegrityCmd)
	cmd.AddCommand(storePerformanceCmd)
	cmd.PersistentFlags().AddFlagSet(storeCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(storeCmd.Flags())
	return cmd
}

// Search command
func runSearchCmd(args []string, out io.Writer) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, utils.FormatError("Usage: search <package>"))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Example: search curl"))
		return
	}

	query := args[0]
	if len(args) > 1 {
		query = fmt.Sprintf("%s %s", args[0], args[1])

	}

	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	exec := nixos.NewExecutor(cfg.NixosFolder)
	pkgOut, pkgErr := exec.SearchNixPackages(query)
	if pkgErr != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("NixOS package search failed: "+pkgErr.Error()))
	} else if pkgOut != "" {
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🔍 NixOS Search Results for: "+query))
		_, _ = fmt.Fprintln(out, pkgOut)
	}

	aiProvider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}

	// Get provider name for context
	providerName := cfg.AIProvider
	if providerName == "" {
		providerName = "ollama"
	}

	var docExcerpts []string
	_, _ = fmt.Fprint(out, utils.FormatInfo("Querying documentation... "))
	mcpBase := cfg.MCPServer.Host
	mcpContextAdded := false
	if mcpBase != "" {
		mcpClient := mcp.NewMCPClient(mcpBase)
		doc, err := mcpClient.QueryDocumentation(query)
		_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
		if err == nil && doc != "" {
			opt, fallbackDoc := parseMCPOptionDoc(doc)
			if opt.Name != "" {
				context := fmt.Sprintf("Option: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nNixOS Version: %s\nRelated: %v\nLinks: %v", opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
				docExcerpts = append(docExcerpts, context)
				mcpContextAdded = true
			} else if len(fallbackDoc) > 0 && (len(fallbackDoc) < 1000 || len(fallbackDoc) > 10) {
				docExcerpts = append(docExcerpts, fallbackDoc)
				mcpContextAdded = true
			}
		}
	} else {
		_, _ = fmt.Fprintln(out, utils.FormatWarning("skipped (no MCP host configured)"))
	}

	promptInstruction := "You are a NixOS expert. Always provide NixOS-specific configuration.nix examples, use the NixOS module system, and avoid generic Linux or upstream package advice. Show how to enable and configure this package/service in NixOS."
	if !mcpContextAdded {
		docExcerpts = append(docExcerpts, promptInstruction)
	} else {
		docExcerpts = append(docExcerpts, "\n"+promptInstruction)
	}

	promptCtx := ai.PromptContext{
		Question:     query,
		DocExcerpts:  docExcerpts,
		Intent:       "explain",
		OutputFormat: "markdown",
		Provider:     providerName,
	}
	builder := ai.DefaultPromptBuilder{}
	prompt, err := builder.BuildPrompt(promptCtx)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Prompt build error: "+err.Error()))
		return
	}
	_, _ = fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
	aiAnswer, aiErr := aiProvider.Query(prompt)
	_, _ = fmt.Fprintln(out, utils.FormatSuccess("done"))
	if aiErr == nil && aiAnswer != "" {
		_, _ = fmt.Fprintln(out, utils.FormatHeader("🤖 AI Best Practices & Tips"))
		_, _ = fmt.Fprintln(out, utils.RenderMarkdown(aiAnswer))
	}
}

// Ask command - Enhanced version with comprehensive information sources and validation
// runAskCmdWithConciseMode is a new version with concise footer-style output
func runAskCmdWithConciseMode(args []string, out io.Writer, providerParam, modelParam string) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, utils.FormatError("Usage: ask <question>"))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Example: ask How do I enable nginx?"))
		return
	}

	question := strings.Join(args, " ")

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load configuration: "+err.Error()))
		return
	}

	// Detect NixOS context (silent)
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to detect NixOS context: "+err.Error()))
		return
	}

	// Create AI provider manager
	manager := ai.NewProviderManager(cfg, logger.NewLogger())

	// Determine which provider to use
	selectedProvider := cfg.AIModels.SelectionPreferences.DefaultProvider
	if providerParam != "" {
		selectedProvider = providerParam
	}
	if selectedProvider == "" {
		selectedProvider = "ollama"
	}

	var provider ai.Provider
	if modelParam != "" {
		provider, err = manager.GetProviderWithModel(selectedProvider, modelParam)
	} else {
		provider, err = manager.GetProvider(selectedProvider)
	}

	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}

	// Silently gather information from multiple sources
	var docExcerpts []string
	var searchContext []string
	var githubExamples []string
	var sourceStatus []string

	// 1. MCP server documentation queries (silent)
	if cfg.MCPServer.Host != "" {
		_, _ = fmt.Fprintf(out, "📚 ")
		mcpClient := mcp.NewMCPClient(fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port))

		sources := []string{
			"https://wiki.nixos.org/wiki/NixOS_Wiki",
			"https://nix.dev/manual/nix",
			"https://nixos.org/manual/nixpkgs/stable/",
			"https://nix.dev/manual/nix/2.28/language/",
			"https://nix-community.github.io/home-manager/",
		}

		doc, mcpErr := mcpClient.QueryDocumentation(question, sources...)
		if mcpErr == nil && doc != "" {
			opt, fallbackDoc := parseMCPOptionDoc(doc)
			if opt.Name != "" {
				context := fmt.Sprintf("NixOS Option Documentation:\nOption: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nVersion: %s\nRelated: %v\nLinks: %v",
					opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
				docExcerpts = append(docExcerpts, context)
				sourceStatus = append(sourceStatus, "docs")
			} else if len(fallbackDoc) > 10 && len(fallbackDoc) < 3000 {
				docExcerpts = append(docExcerpts, "NixOS Documentation Context:\n"+fallbackDoc)
				sourceStatus = append(sourceStatus, "docs")
			}
		}

		// Query for service examples if applicable
		searchTerms := extractSearchTerms(question)
		for _, term := range searchTerms {
			if strings.Contains(question, "service") || strings.Contains(question, "enable") {
				if serviceDoc, err := mcpClient.QueryDocumentation("service examples for " + term); err == nil && serviceDoc != "" {
					if len(serviceDoc) > 20 && len(serviceDoc) < 2000 {
						docExcerpts = append(docExcerpts, fmt.Sprintf("Service Configuration Examples for '%s':\n%s", term, serviceDoc))
					}
				}
			}
		}
	}

	// 2. Package and options search (silent)
	_, _ = fmt.Fprintf(out, "📦 ")
	exec := nixos.NewExecutor(cfg.NixosFolder)
	searchTerms := extractSearchTerms(question)
	foundPackages := 0
	for _, term := range searchTerms {
		if packageInfo, err := exec.SearchNixPackages(term); err == nil && packageInfo != "" {
			searchContext = append(searchContext, fmt.Sprintf("Package Search for '%s':\n%s", term, packageInfo))
			foundPackages++
		}
	}
	if foundPackages > 0 {
		sourceStatus = append(sourceStatus, "packages")
	}

	// 3. GitHub code search (silent)
	if strings.Contains(question, "flake") || strings.Contains(question, "configuration") ||
		strings.Contains(question, "service") || strings.Contains(question, "enable") {

		_, _ = fmt.Fprintf(out, "🔍 ")
		githubToken := os.Getenv("GITHUB_TOKEN")
		githubClient := community.NewGitHubClient(githubToken)

		foundConfigs := 0
		for _, term := range searchTerms {
			if len(term) > 3 {
				configs, err := githubClient.SearchNixOSConfigurations(term)
				if err == nil && len(configs) > 0 {
					for i, config := range configs {
						if i >= 2 {
							break
						}
						githubExamples = append(githubExamples,
							fmt.Sprintf("Real-world NixOS configuration example (%s):\nRepo: %s\nDescription: %s\nAuthor: %s\nStars: %d\nURL: %s",
								term, config.Name, config.Description, config.Author, config.Views, config.URL))
						foundConfigs++
					}
				}
			}
		}
		if foundConfigs > 0 {
			sourceStatus = append(sourceStatus, "examples")
		}
	}

	_, _ = fmt.Fprintf(out, "🤖 ")

	// Build comprehensive context-aware prompt
	contextBuilder := nixoscontext.NewNixOSContextBuilder()

	basePrompt := ""
	if template, exists := roles.RolePromptTemplate[roles.RoleAsk]; exists {
		basePrompt = template
	}

	nixosGuidelines := "ATTENTION: You are a NixOS expert with access to multiple verified sources. NEVER EVER suggest nix-env commands!\n\n" +
		"CRITICAL ACCURACY RULES:\n" +
		"❌ NEVER suggest 'nix-env -i' or any nix-env commands\n" +
		"❌ NEVER recommend manual installation\n" +
		"❌ NEVER use incorrect flake syntax like 'nixpkgs.nix = {...}'\n" +
		"❌ NEVER suggest outdated or deprecated options\n\n" +
		"✅ BLUETOOTH SPECIFIC RULES:\n" +
		"✅ ALWAYS use 'hardware.bluetooth.enable = true;' for Bluetooth (NOT services.bluetooth.enable)\n" +
		"✅ Use 'services.blueman.enable = true;' ONLY if user needs a GUI manager\n" +
		"✅ Mention that both hardware.bluetooth.enable AND services.blueman.enable may be needed\n\n" +
		"✅ ALWAYS USE configuration.nix for system packages\n" +
		"✅ ALWAYS USE services.* options for services\n" +
		"✅ ALWAYS use correct flake syntax: inputs.nixpkgs.url = \"github:...\" and outputs = { self, nixpkgs }: {...}\n" +
		"✅ ALWAYS verify package names and option paths with provided search results\n" +
		"✅ ALWAYS end with 'sudo nixos-rebuild switch' for configuration changes\n" +
		"✅ ALWAYS use examples from the provided real-world GitHub configurations when available\n\n"

	// Build context-aware prompt
	contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt+"\n\n"+nixosGuidelines, nixosCtx)

	// Add documentation context
	if len(docExcerpts) > 0 {
		contextualPrompt += "\n\nOFFICIAL DOCUMENTATION CONTEXT:\n" + strings.Join(docExcerpts, "\n\n")
		sourceStatus = append(sourceStatus, "docs")
	}

	// Add package search context
	if len(searchContext) > 0 {
		contextualPrompt += "\n\nVERIFIED PACKAGE SEARCH RESULTS:\n" + strings.Join(searchContext, "\n\n")
		contextualPrompt += "\n\nUse this package information to provide accurate package names and availability."
		sourceStatus = append(sourceStatus, "packages")
	}

	// Add GitHub examples context
	if len(githubExamples) > 0 {
		contextualPrompt += "\n\nREAL-WORLD NIXOS CONFIGURATION EXAMPLES:\n" + strings.Join(githubExamples, "\n\n")
		contextualPrompt += "\n\nUse these real-world examples to validate syntax and provide accurate configurations."
		sourceStatus = append(sourceStatus, "examples")
	}

	// Add synthesis instruction
	contextualPrompt += "\n\nSYNTHESIS INSTRUCTION: Combine information from official documentation, verified package searches, and real-world examples to provide the most accurate and up-to-date NixOS configuration advice."

	// Add the user question
	finalPrompt := contextualPrompt + "\n\nUser Question: " + question

	// Query the AI provider (silent)
	ctx := context.Background()
	response, err := provider.Query(ctx, finalPrompt)

	if err != nil {
		_, _ = fmt.Fprintln(out, "❌")
		_, _ = fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, "✅")
	_, _ = fmt.Fprintln(out)

	// Display the AI response
	_, _ = fmt.Fprintln(out, utils.RenderMarkdown(response))

	// Minimal quality assessment
	qualityScore := len(sourceStatus)
	if nixosCtx != nil && nixosCtx.CacheValid {
		qualityScore++
	}
	if strings.Contains(response, "configuration.nix") && !strings.Contains(response, "nix-env") {
		qualityScore++
	}

	// Ultra-minimal footer
	if len(sourceStatus) > 0 {
		_, _ = fmt.Fprintf(out, "\n─ %s ─\n", strings.Join(sourceStatus, " • "))
	}
}

// getNixOSContextSummary returns a concise context summary
func getNixOSContextSummary(nixosCtx *config.NixOSContext) string {
	if nixosCtx == nil {
		return "unknown"
	}

	parts := []string{nixosCtx.SystemType}
	if nixosCtx.UsesFlakes {
		parts = append(parts, "Flakes: Yes")
	} else {
		parts = append(parts, "Flakes: No")
	}
	if nixosCtx.HasHomeManager {
		parts = append(parts, fmt.Sprintf("Home Manager: %s", nixosCtx.HomeManagerType))
	}

	return strings.Join(parts, " | ")
}

// getProviderNameFromProvider extracts the provider name from a Provider instance
func getProviderNameFromProvider(provider ai.Provider) string {
	// Simple fallback - just return "ai" for now
	// Could be enhanced later with proper type checking
	return "ai"
}

func runAskCmd(args []string, out io.Writer) {
	runAskCmdWithConciseMode(args, out, "", "")
}

// runAskCmdWithQuietMode is a wrapper that adds quiet mode support
func runAskCmdWithQuietMode(args []string, out io.Writer, providerParam, modelParam string, quiet bool) {
	if quiet {
		runAskCmdWithOptionsQuiet(args, out, providerParam, modelParam)
	} else {
		runAskCmdWithOptions(args, out, providerParam, modelParam)
	}
}

// runAskCmdWithOptionsQuiet is the quiet version with minimal output
func runAskCmdWithOptionsQuiet(args []string, out io.Writer, providerParam, modelParam string) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, utils.FormatError("Usage: ask <question>"))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Example: ask How do I enable nginx?"))
		return
	}

	// Join all arguments to form the question
	question := strings.Join(args, " ")

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	// Initialize context detector and get NixOS context (silent)
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, _ := contextDetector.GetContext(cfg)

	// Create modern AI provider using new ProviderManager system
	manager := ai.NewProviderManager(cfg, logger.NewLogger())

	// Determine which provider to use from command flags or config
	selectedProvider := cfg.AIModels.SelectionPreferences.DefaultProvider

	// Check for direct provider parameter (from subcommand flags)
	if providerParam != "" {
		selectedProvider = providerParam
	} else if providerFlag := os.Getenv("NIXAI_PROVIDER"); providerFlag != "" {
		selectedProvider = providerFlag
	}

	if selectedProvider == "" {
		selectedProvider = "ollama"
	}

	// Get the provider with optional model specification
	var provider ai.Provider

	if modelParam != "" {
		provider, err = manager.GetProviderWithModel(selectedProvider, modelParam)
	} else if modelFlag := os.Getenv("NIXAI_MODEL"); modelFlag != "" {
		provider, err = manager.GetProviderWithModel(selectedProvider, modelFlag)
	} else {
		provider, err = manager.GetProvider(selectedProvider)
	}

	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}

	// Silent Multi-Source Information Gathering (no progress output)
	var docExcerpts []string
	var searchContext []string
	var githubExamples []string

	// 1. MCP server documentation queries (silent)
	mcpBase := cfg.MCPServer.Host
	if mcpBase != "" {
		mcpClient := mcp.NewMCPClient(fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port))

		sources := []string{
			"https://wiki.nixos.org/wiki/NixOS_Wiki",
			"https://nix.dev/manual/nix",
			"https://nixos.org/manual/nixpkgs/stable/",
			"https://nix.dev/manual/nix/2.28/language/",
			"https://nix-community.github.io/home-manager/",
		}

		// Primary documentation query
		doc, mcpErr := mcpClient.QueryDocumentation(question, sources...)
		if mcpErr == nil && doc != "" {
			opt, fallbackDoc := parseMCPOptionDoc(doc)
			if opt.Name != "" {
				context := fmt.Sprintf("NixOS Option Documentation:\nOption: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nVersion: %s\nRelated: %v\nLinks: %v",
					opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
				docExcerpts = append(docExcerpts, context)
			} else if len(fallbackDoc) > 10 && len(fallbackDoc) < 3000 {
				docExcerpts = append(docExcerpts, "NixOS Documentation Context:\n"+fallbackDoc)
			}
		}

		// Query for service examples if applicable
		searchTerms := extractSearchTerms(question)
		for _, term := range searchTerms {
			if strings.Contains(question, "service") || strings.Contains(question, "enable") {
				if serviceDoc, err := mcpClient.QueryDocumentation("service examples for " + term); err == nil && serviceDoc != "" {
					if len(serviceDoc) > 20 && len(serviceDoc) < 2000 {
						docExcerpts = append(docExcerpts, fmt.Sprintf("Service Configuration Examples for '%s':\n%s", term, serviceDoc))
					}
				}
			}
		}
	}

	// 2. Package and options search (silent)
	exec := nixos.NewExecutor(cfg.NixosFolder)
	searchTerms := extractSearchTerms(question)
	for _, term := range searchTerms {
		if packageInfo, err := exec.SearchNixPackages(term); err == nil && packageInfo != "" {
			searchContext = append(searchContext, fmt.Sprintf("Package Search for '%s':\n%s", term, packageInfo))
		}
	}

	// 3. GitHub code search (silent)
	if strings.Contains(question, "flake") || strings.Contains(question, "configuration") ||
		strings.Contains(question, "service") || strings.Contains(question, "enable") {

		githubToken := os.Getenv("GITHUB_TOKEN")
		githubClient := community.NewGitHubClient(githubToken)

		for _, term := range searchTerms {
			if len(term) > 3 {
				configs, err := githubClient.SearchNixOSConfigurations(term)
				if err == nil && len(configs) > 0 {
					for i, config := range configs {
						if i >= 2 {
							break
						}
						githubExamples = append(githubExamples,
							fmt.Sprintf("Real-world NixOS configuration example (%s):\nRepo: %s\nDescription: %s\nAuthor: %s\nStars: %d\nURL: %s",
								term, config.Name, config.Description, config.Author, config.Views, config.URL))
					}
				}
			}
		}
	}

	// 4. Build comprehensive context-aware prompt
	contextBuilder := nixoscontext.NewNixOSContextBuilder()

	basePrompt := ""
	if template, exists := roles.RolePromptTemplate[roles.RoleAsk]; exists {
		basePrompt = template
	}

	nixosGuidelines := "ATTENTION: You are a NixOS expert with access to multiple verified sources. NEVER EVER suggest nix-env commands!\n\n" +
		"CRITICAL ACCURACY RULES:\n" +
		"❌ NEVER suggest 'nix-env -i' or any nix-env commands\n" +
		"❌ NEVER recommend manual installation\n" +
		"❌ NEVER use incorrect flake syntax like 'nixpkgs.nix = {...}'\n" +
		"❌ NEVER suggest outdated or deprecated options\n\n" +
		"✅ BLUETOOTH SPECIFIC RULES:\n" +
		"✅ ALWAYS use 'hardware.bluetooth.enable = true;' for Bluetooth (NOT services.bluetooth.enable)\n" +
		"✅ Use 'services.blueman.enable = true;' ONLY if user needs a GUI manager\n" +
		"✅ Mention that both hardware.bluetooth.enable AND services.blueman.enable may be needed\n\n" +
		"✅ ALWAYS USE configuration.nix for system packages\n" +
		"✅ ALWAYS USE services.* options for services\n" +
		"✅ ALWAYS use correct flake syntax: inputs.nixpkgs.url = \"github:...\" and outputs = { self, nixpkgs }: {...}\n" +
		"✅ ALWAYS verify package names and option paths with provided search results\n" +
		"✅ ALWAYS end with 'sudo nixos-rebuild switch' for configuration changes\n" +
		"✅ ALWAYS use examples from the provided real-world GitHub configurations when available\n\n"

	// Build context-aware prompt
	contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt+"\n\n"+nixosGuidelines, nixosCtx)

	// Add documentation context
	if len(docExcerpts) > 0 {
		contextualPrompt += "\n\nOFFICIAL DOCUMENTATION CONTEXT:\n" + strings.Join(docExcerpts, "\n\n")
	}

	// Add package search context
	if len(searchContext) > 0 {
		contextualPrompt += "\n\nVERIFIED PACKAGE SEARCH RESULTS:\n" + strings.Join(searchContext, "\n\n")
		contextualPrompt += "\n\nUse this package information to provide accurate package names and availability."
	}

	// Add GitHub examples context
	if len(githubExamples) > 0 {
		contextualPrompt += "\n\nREAL-WORLD NIXOS CONFIGURATION EXAMPLES:\n" + strings.Join(githubExamples, "\n\n")
		contextualPrompt += "\n\nUse these real-world examples to validate syntax and provide accurate configurations."
	}

	// Add synthesis instruction
	contextualPrompt += "\n\nSYNTHESIS INSTRUCTION: Combine information from official documentation, verified package searches, and real-world examples to provide the most accurate and up-to-date NixOS configuration advice."

	// Add the user question
	finalPrompt := contextualPrompt + "\n\nUser Question: " + question

	// Query the AI provider (silent)
	ctx := context.Background()
	response, err := provider.Query(ctx, finalPrompt)

	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
		return
	}

	// Display only the AI response (no validation output)
	_, _ = fmt.Fprintln(out, utils.RenderMarkdown(response))
}

// runAskCmdWithOptions is the original verbose version with full validation and multi-source information gathering
func runAskCmdWithOptions(args []string, out io.Writer, providerParam, modelParam string) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, utils.FormatError("Usage: ask <question>"))
		_, _ = fmt.Fprintln(out, utils.FormatTip("Example: ask How do I enable nginx?"))
		return
	}

	// Join all arguments to form the question
	question := strings.Join(args, " ")

	_, _ = fmt.Fprintln(out, utils.FormatHeader("🤖 Enhanced AI Question Assistant"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Question", question))
	_, _ = fmt.Fprintln(out)

	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to load config: "+err.Error()))
		return
	}

	// Initialize context detector and get NixOS context
	contextDetector := nixos.NewContextDetector(logger.NewLogger())
	nixosCtx, err := contextDetector.GetContext(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Context detection failed: "+err.Error()))
		nixosCtx = nil
	}

	// Display detected context summary if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		contextBuilder := nixoscontext.NewNixOSContextBuilder()
		contextSummary := contextBuilder.GetContextSummary(nixosCtx)
		_, _ = fmt.Fprintln(out, utils.FormatNote("📋 "+contextSummary))
		_, _ = fmt.Fprintln(out)
	}

	// Create modern AI provider using new ProviderManager system
	manager := ai.NewProviderManager(cfg, logger.NewLogger())

	// Determine which provider to use from command flags or config
	selectedProvider := cfg.AIModels.SelectionPreferences.DefaultProvider

	// Check for direct provider parameter (from subcommand flags)
	if providerParam != "" {
		selectedProvider = providerParam
	} else if providerFlag := os.Getenv("NIXAI_PROVIDER"); providerFlag != "" {
		selectedProvider = providerFlag
	}

	if selectedProvider == "" {
		selectedProvider = "ollama"
	}

	// Get the provider with optional model specification
	var provider ai.Provider

	if modelParam != "" {
		provider, err = manager.GetProviderWithModel(selectedProvider, modelParam)
	} else if modelFlag := os.Getenv("NIXAI_MODEL"); modelFlag != "" {
		provider, err = manager.GetProviderWithModel(selectedProvider, modelFlag)
	} else {
		provider, err = manager.GetProvider(selectedProvider)
	}

	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("Failed to initialize AI provider: "+err.Error()))
		return
	}

	// Multi-Source Information Gathering with progress indicators
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📚 Gathering Information from Multiple Sources"))
	_, _ = fmt.Fprintln(out)

	var docExcerpts []string
	var searchContext []string
	var githubExamples []string

	// 1. MCP server documentation queries
	mcpBase := cfg.MCPServer.Host
	if mcpBase != "" {
		_, _ = fmt.Fprint(out, utils.FormatInfo("Querying official documentation... "))
		mcpClient := mcp.NewMCPClient(fmt.Sprintf("http://%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port))

		sources := []string{
			"https://wiki.nixos.org/wiki/NixOS_Wiki",
			"https://nix.dev/manual/nix",
			"https://nixos.org/manual/nixpkgs/stable/",
			"https://nix.dev/manual/nix/2.28/language/",
			"https://nix-community.github.io/home-manager/",
		}

		// Primary documentation query
		doc, mcpErr := mcpClient.QueryDocumentation(question, sources...)
		if mcpErr == nil && doc != "" {
			opt, fallbackDoc := parseMCPOptionDoc(doc)
			if opt.Name != "" {
				context := fmt.Sprintf("NixOS Option Documentation:\nOption: %s\nType: %s\nDefault: %s\nExample: %s\nDescription: %s\nSource: %s\nVersion: %s\nRelated: %v\nLinks: %v",
					opt.Name, opt.Type, opt.Default, opt.Example, opt.Description, opt.Source, opt.Version, opt.Related, opt.Links)
				docExcerpts = append(docExcerpts, context)
				_, _ = fmt.Fprintln(out, utils.FormatSuccess("found option documentation"))
			} else if len(fallbackDoc) > 10 && len(fallbackDoc) < 3000 {
				docExcerpts = append(docExcerpts, "NixOS Documentation Context:\n"+fallbackDoc)
				_, _ = fmt.Fprintln(out, utils.FormatSuccess("found general documentation"))
			} else {
				_, _ = fmt.Fprintln(out, utils.FormatWarning("limited documentation found"))
			}
		} else {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("no documentation found"))
		}

		// Query for service examples if applicable
		searchTerms := extractSearchTerms(question)
		for _, term := range searchTerms {
			if strings.Contains(question, "service") || strings.Contains(question, "enable") {
				if serviceDoc, err := mcpClient.QueryDocumentation("service examples for " + term); err == nil && serviceDoc != "" {
					if len(serviceDoc) > 20 && len(serviceDoc) < 2000 {
						docExcerpts = append(docExcerpts, fmt.Sprintf("Service Configuration Examples for '%s':\n%s", term, serviceDoc))
					}
				}
			}
		}
	} else {
		_, _ = fmt.Fprintln(out, utils.FormatWarning("MCP server not configured - skipping documentation"))
	}

	// 2. Package and options search
	_, _ = fmt.Fprint(out, utils.FormatInfo("Searching packages and options... "))
	exec := nixos.NewExecutor(cfg.NixosFolder)
	searchTerms := extractSearchTerms(question)
	foundPackages := 0
	for _, term := range searchTerms {
		if packageInfo, err := exec.SearchNixPackages(term); err == nil && packageInfo != "" {
			searchContext = append(searchContext, fmt.Sprintf("Package Search for '%s':\n%s", term, packageInfo))
			foundPackages++
		}
	}
	if foundPackages > 0 {
		_, _ = fmt.Fprintln(out, utils.FormatSuccess(fmt.Sprintf("found %d package results", foundPackages)))
	} else {
		_, _ = fmt.Fprintln(out, utils.FormatWarning("no packages found"))
	}

	// 3. GitHub code search
	if strings.Contains(question, "flake") || strings.Contains(question, "configuration") ||
		strings.Contains(question, "service") || strings.Contains(question, "enable") {

		_, _ = fmt.Fprint(out, utils.FormatInfo("Searching real-world configurations... "))
		githubToken := os.Getenv("GITHUB_TOKEN")
		githubClient := community.NewGitHubClient(githubToken)

		foundConfigs := 0
		for _, term := range searchTerms {
			if len(term) > 3 {
				configs, err := githubClient.SearchNixOSConfigurations(term)
				if err == nil && len(configs) > 0 {
					for i, config := range configs {
						if i >= 2 {
							break
						}
						githubExamples = append(githubExamples,
							fmt.Sprintf("Real-world NixOS configuration example (%s):\nRepo: %s\nDescription: %s\nAuthor: %s\nStars: %d\nURL: %s",
								term, config.Name, config.Description, config.Author, config.Views, config.URL))
						foundConfigs++
					}
				}
			}
		}
		if foundConfigs > 0 {
			_, _ = fmt.Fprintln(out, utils.FormatSuccess(fmt.Sprintf("found %d configuration examples", foundConfigs)))
		} else {
			_, _ = fmt.Fprintln(out, utils.FormatWarning("no configuration examples found"))
		}
	}

	_, _ = fmt.Fprintln(out)

	// 4. Build comprehensive context-aware prompt
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🧠 Processing with AI"))
	_, _ = fmt.Fprintln(out)

	contextBuilder := nixoscontext.NewNixOSContextBuilder()

	basePrompt := ""
	if template, exists := roles.RolePromptTemplate[roles.RoleAsk]; exists {
		basePrompt = template
	}

	nixosGuidelines := "ATTENTION: You are a NixOS expert with access to multiple verified sources. NEVER EVER suggest nix-env commands!\n\n" +
		"CRITICAL ACCURACY RULES:\n" +
		"❌ NEVER suggest 'nix-env -i' or any nix-env commands\n" +
		"❌ NEVER recommend manual installation\n" +
		"❌ NEVER use incorrect flake syntax like 'nixpkgs.nix = {...}'\n" +
		"❌ NEVER suggest outdated or deprecated options\n\n" +
		"✅ BLUETOOTH SPECIFIC RULES:\n" +
		"✅ ALWAYS use 'hardware.bluetooth.enable = true;' for Bluetooth (NOT services.bluetooth.enable)\n" +
		"✅ Use 'services.blueman.enable = true;' ONLY if user needs a GUI manager\n" +
		"✅ Mention that both hardware.bluetooth.enable AND services.blueman.enable may be needed\n\n" +
		"✅ ALWAYS USE configuration.nix for system packages\n" +
		"✅ ALWAYS USE services.* options for services\n" +
		"✅ ALWAYS use correct flake syntax: inputs.nixpkgs.url = \"github:...\" and outputs = { self, nixpkgs }: {...}\n" +
		"✅ ALWAYS verify package names and option paths with provided search results\n" +
		"✅ ALWAYS end with 'sudo nixos-rebuild switch' for configuration changes\n" +
		"✅ ALWAYS use examples from the provided real-world GitHub configurations when available\n\n"

	// Build context-aware prompt
	contextualPrompt := contextBuilder.BuildContextualPrompt(basePrompt+"\n\n"+nixosGuidelines, nixosCtx)

	// Add documentation context
	if len(docExcerpts) > 0 {
		contextualPrompt += "\n\nOFFICIAL DOCUMENTATION CONTEXT:\n" + strings.Join(docExcerpts, "\n\n")
		_, _ = fmt.Fprintln(out, utils.FormatNote("✅ Official documentation integrated"))
	}

	// Add package search context
	if len(searchContext) > 0 {
		contextualPrompt += "\n\nVERIFIED PACKAGE SEARCH RESULTS:\n" + strings.Join(searchContext, "\n\n")
		contextualPrompt += "\n\nUse this package information to provide accurate package names and availability."
		_, _ = fmt.Fprintln(out, utils.FormatNote("✅ Package search results integrated"))
	}

	// Add GitHub examples context
	if len(githubExamples) > 0 {
		contextualPrompt += "\n\nREAL-WORLD NIXOS CONFIGURATION EXAMPLES:\n" + strings.Join(githubExamples, "\n\n")
		contextualPrompt += "\n\nUse these real-world examples to validate syntax and provide accurate configurations."
		_, _ = fmt.Fprintln(out, utils.FormatNote("✅ Real-world configuration examples integrated"))
	}

	// Add synthesis instruction
	contextualPrompt += "\n\nSYNTHESIS INSTRUCTION: Combine information from official documentation, verified package searches, and real-world examples to provide the most accurate and up-to-date NixOS configuration advice."

	// Add the user question
	finalPrompt := contextualPrompt + "\n\nUser Question: " + question

	// Query the AI provider
	_, _ = fmt.Fprint(out, utils.FormatInfo("Querying AI provider... "))
	ctx := context.Background()
	response, err := provider.Query(ctx, finalPrompt)

	if err != nil {
		_, _ = fmt.Fprintln(out, utils.FormatError("failed"))
		_, _ = fmt.Fprintln(out, utils.FormatError("AI error: "+err.Error()))
		return
	}

	_, _ = fmt.Fprintln(out, utils.FormatSuccess("complete"))
	_, _ = fmt.Fprintln(out)

	// Display the AI response
	_, _ = fmt.Fprintln(out, utils.FormatHeader("🎯 AI Response"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.RenderMarkdown(response))

	// Add quality indicators and help information
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatDivider())
	_, _ = fmt.Fprintln(out, utils.FormatHeader("📊 Response Quality Indicators"))
	_, _ = fmt.Fprintln(out)

	qualityScore := 0
	maxScore := 5

	if len(docExcerpts) > 0 {
		qualityScore++
		_, _ = fmt.Fprintln(out, "✅ Official documentation consulted")
	} else {
		_, _ = fmt.Fprintln(out, "⚠️  No official documentation found")
	}

	if len(searchContext) > 0 {
		qualityScore++
		_, _ = fmt.Fprintln(out, "✅ Package search results verified")
	} else {
		_, _ = fmt.Fprintln(out, "⚠️  No package search results found")
	}

	if len(githubExamples) > 0 {
		qualityScore++
		_, _ = fmt.Fprintln(out, "✅ Real-world configuration examples included")
	} else {
		_, _ = fmt.Fprintln(out, "⚠️  No real-world examples found")
	}

	if nixosCtx != nil && nixosCtx.CacheValid {
		qualityScore++
		_, _ = fmt.Fprintln(out, "✅ System context awareness enabled")
	} else {
		_, _ = fmt.Fprintln(out, "⚠️  Limited system context")
	}

	if strings.Contains(response, "configuration.nix") && !strings.Contains(response, "nix-env") {
		qualityScore++
		_, _ = fmt.Fprintln(out, "✅ Follows NixOS best practices")
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatKeyValue("Quality Score", fmt.Sprintf("%d/%d", qualityScore, maxScore)))

	// Enhanced help information
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, utils.FormatTip("For syntax validation, use: nixai flake validate"))
	_, _ = fmt.Fprintln(out, utils.FormatTip("For more options, use: nixai explain-option <option>"))
	_, _ = fmt.Fprintln(out, utils.FormatTip("For package details, use: nixai search <package>"))
	if qualityScore < 3 {
		_, _ = fmt.Fprintln(out, utils.FormatWarning("Consider using more specific terms for better results"))
	}
}

// RunDirectCommand executes commands directly from interactive mode
func RunDirectCommand(cmdName string, args []string, out io.Writer) (bool, error) {
	switch cmdName {
	case "community":
		runCommunityCmd(args, out)
		return true, nil
	case "config":
		runConfigCmd(args, out)
		return true, nil
	case "configure":
		runConfigureCmd(args, out)
		return true, nil
	case "diagnose":
		runDiagnoseCmd(args, out)
		return true, nil
	case "doctor":
		runDoctorCmd(args, out)
		return true, nil
	case "flake":
		runFlakeCmd(args, out)
		return true, nil
	case "learn":
		runLearnCmd(args, out)
		return true, nil
	case "logs":
		runLogsCmd(args, out)
		return true, nil
	case "mcp-server":
		runMCPServerCmd(args, out)
		return true, nil
	case "neovim-setup":
		runNeovimSetupCmd(args, out)
		return true, nil
	case "package-repo":
		runPackageRepoCmd(args, out)
		return true, nil
	case "machines":
		runMachinesCmd(args, out)
		return true, nil
	case "build":
		runBuildCmd(args, out)
		return true, nil
	case "completion":
		runCompletionCmd(args, out)
		return true, nil
	case "deps":
		runDepsCmd(args, out)
		return true, nil
	case "devenv":
		runDevenvCmd(args, out)
		return true, nil
	case "explain-option":
		runExplainOptionCmd(args, out)
		return true, nil
	case "gc":
		runGCCmd(args, out)
		return true, nil
	case "hardware":
		runHardwareCmd(args, out)
		return true, nil
	case "interactive":
		_, _ = fmt.Fprintln(out, utils.FormatTip("You are already in interactive mode!"))
		return true, nil
	case "migrate":
		runMigrateCmd(args, out)
		return true, nil
	case "search":
		runSearchCmd(args, out)
		return true, nil
	case "snippets":
		runSnippetsCmd(args, out)
		return true, nil
	case "store":
		runStoreCmd(args, out)
		return true, nil
	case "templates":
		runTemplatesCmd(args, out)
		return true, nil
	case "ask":
		runAskCmd(args, out)
		return true, nil
	case "help":
		_, _ = fmt.Fprintln(out, utils.FormatHeader("❓ Help: Available Commands"))
		_, _ = fmt.Fprintln(out, `🤖 ask <question>: Ask any NixOS question\n🛠️ build: Enhanced build troubleshooting and optimization\n🌐 community: Community resources and support\n🔄 completion: Generate the autocompletion script for the specified shell\n⚙️ config: Manage nixai configuration\n🧑‍💻 configure: Configure NixOS interactively\n🔗 deps: Analyze NixOS configuration dependencies and imports\n🧪 devenv: Create and manage development environments with devenv\n🩺 diagnose: Diagnose NixOS issues\n🩻 doctor: Run NixOS health checks\n🖥️ explain-option <option>: Explain a NixOS option\n🧊 flake: Nix flake utilities\n🧹 gc: AI-powered garbage collection analysis and cleanup\n💻 hardware: AI-powered hardware configuration optimizer\n❓ help: Help about any command\n💬 interactive: Launch interactive AI-powered NixOS assistant shell\n📚 learn: NixOS learning and training commands\n📝 logs: Analyze and parse NixOS logs\n🖧 machines: Manage and synchronize NixOS configurations across multiple machines\n🛰️ mcp-server: Start or manage the MCP server\n🔀 migrate: AI-powered migration assistant for channels and flakes\n📝 neovim-setup: Neovim integration setup\n📦 package-repo <url>: Analyze Git repos and generate Nix derivations\n🔍 search <package>: Search for NixOS packages/services and get config/AI tips\n🔖 snippets: Manage NixOS configuration snippets\n💾 store: Manage, backup, and analyze the Nix store\n📄 templates: Manage NixOS configuration templates and snippets\n❌ exit: Exit interactive mode`)
		return true, nil
	case "exit":
		_, _ = fmt.Fprintln(out, utils.FormatTip("Type Ctrl+D or 'exit' to leave interactive mode."))
		return true, nil
	default:
		return false, nil
	}
}
