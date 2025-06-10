package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/ai"
	nixoscontext "nix-ai-help/internal/ai/context"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

// Configuration variables
var (
	depNixosConfigPath string // Path to NixOS configuration
	depNixosHostname   string // NixOS hostname for flakes
	depMaxDepth        int    // Maximum recursion depth
	depDotOutputPath   string // Output path for DOT file
	depCurrentSystem   bool   // Analyze current running system instead of configuration
)

// NewDepsCommand creates and returns the deps command and all subcommands
func NewDepsCommand() *cobra.Command {
	// Main deps command
	depsCommand := &cobra.Command{
		Use:   "deps",
		Short: "Analyze NixOS configuration dependencies and imports",
		Long: `Provides tools to visualize and analyze NixOS configuration dependencies,
helping to understand relationships, detect conflicts, and optimize configurations.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Analyze subcommand
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Show dependency tree of the NixOS configuration with AI insights",
		Long: `Analyzes your NixOS configuration (flake.nix or configuration.nix)
to build a dependency tree up to a specified depth. It then uses AI to provide insights into
potential issues, optimization opportunities, or other relevant information
based on the detected dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDepsAnalyze()
		},
	}

	// Why subcommand
	whyCmd := &cobra.Command{
		Use:   "why [package-name]",
		Short: "Explain why a package is installed by showing dependency paths",
		Long: `Traces dependency paths to show all reasons why a specific package is included
in your NixOS configuration. This helps understand complex dependency relationships
and identify the root causes of package inclusion.`,
		Args: conditionalExactArgsValidator(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDepsWhy(args[0])
		},
	}

	// Conflicts subcommand
	conflictsCmd := &cobra.Command{
		Use:   "conflicts",
		Short: "Find and resolve dependency conflicts",
		Long: `Analyzes your NixOS configuration for dependency conflicts, such as multiple
versions of the same package being pulled in by different dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDepsConflicts()
		},
	}

	// Optimize subcommand
	optimizeCmd := &cobra.Command{
		Use:   "optimize",
		Short: "Suggest dependency optimizations",
		Long: `Analyzes your NixOS configuration for dependency optimization opportunities,
such as redundant dependencies, circular references, or inefficient dependency chains.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDepsOptimize()
		},
	}

	// Graph subcommand
	graphCmd := &cobra.Command{
		Use:   "graph",
		Short: "Generate visual dependency graph",
		Long: `Generates a visual representation of your NixOS configuration's dependency graph
in DOT format, which can be used with tools like Graphviz to create visual diagrams.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDepsGraph()
		},
	}

	// Add all subcommands
	depsCommand.AddCommand(analyzeCmd)
	depsCommand.AddCommand(whyCmd)
	depsCommand.AddCommand(conflictsCmd)
	depsCommand.AddCommand(optimizeCmd)
	depsCommand.AddCommand(graphCmd)

	// Add common flags to all commands
	depsCommand.PersistentFlags().StringVarP(&depNixosConfigPath, "nixos-path", "p", "",
		"Path to your NixOS configuration directory (if flake) or file (flake.nix or configuration.nix)")
	depsCommand.PersistentFlags().StringVar(&depNixosHostname, "hostname", "",
		"NixOS configuration hostname (for flakes, e.g., from flake.nix#nixosConfigurations.<hostname>). Auto-detected if not provided.")
	depsCommand.PersistentFlags().IntVarP(&depMaxDepth, "depth", "d", 5,
		"Maximum recursion depth for dependency analysis.")
	depsCommand.PersistentFlags().BoolVarP(&depCurrentSystem, "current", "c", false,
		"Analyze current running system (/run/current-system) instead of configuration files")

	// Add graph-specific flags
	graphCmd.Flags().StringVarP(&depDotOutputPath, "output", "o", "",
		"Output path for the DOT file. If not provided, a preview will be shown.")

	return depsCommand
}

// Helper function implementations
func runDepsAnalyze() {
	fmt.Println(utils.FormatHeader("🛠️ NixOS Dependency Analyzer"))

	// Load user configuration for context detection and AI provider
	userCfg, configErr := config.LoadUserConfig()
	if configErr != nil {
		fmt.Println(utils.FormatWarning("Could not load user config: " + configErr.Error()))
	} else {
		// Initialize context detector and get NixOS context
		contextDetector := nixos.NewContextDetector(logger.NewLogger())
		nixosCtx, err := contextDetector.GetContext(userCfg)
		if err != nil {
			fmt.Println(utils.FormatWarning("Context detection failed: " + err.Error()))
		} else if nixosCtx != nil && nixosCtx.CacheValid {
			contextBuilder := nixoscontext.NewNixOSContextBuilder()
			contextSummary := contextBuilder.GetContextSummary(nixosCtx)
			fmt.Println(utils.FormatNote("📋 " + contextSummary))
			fmt.Println()
		}
	}

	depGraph, err := getDependencyGraph()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to analyze dependencies: %v", err)))
		os.Exit(1)
	}

	if depGraph != nil {
		fmt.Println(utils.FormatSuccess("Dependency analysis complete."))
		fmt.Println(utils.FormatHeader("🌳 Dependency Graph:"))
		displayDependencyGraph(depGraph.RootNodes, "")

		// Get AI insights on the dependency graph
		fmt.Println(utils.FormatDivider())
		fmt.Println(utils.FormatHeader("🤖 AI Analysis & Insights"))

		if userCfg != nil {
			// Get AI insights
			aiInsights, aiErr := getAIInsightsForDependencies(depGraph, userCfg)
			if aiErr != nil {
				fmt.Println(utils.FormatError(fmt.Sprintf("Failed to get AI insights: %v", aiErr)))
			} else {
				// Render AI insights using glamour
				if aiInsights != "" {
					renderer, _ := glamour.NewTermRenderer(
						glamour.WithAutoStyle(),
						glamour.WithWordWrap(120),
					)
					rendered, renderErr := renderer.Render(aiInsights)
					if renderErr != nil {
						fmt.Println(utils.FormatError(fmt.Sprintf("Failed to render AI insights: %v", renderErr)))
						fmt.Println(aiInsights) // Fallback to plain text
					} else {
						fmt.Print(rendered)
					}
				}
			}
		} else {
			fmt.Println(utils.FormatWarning("Could not load user config for AI insights, skipping AI analysis"))
		}
	} else {
		fmt.Println(utils.FormatWarning("Dependency analysis did not produce a graph."))
	}
}

func runDepsWhy(packageName string) {
	fmt.Println(utils.FormatHeader(fmt.Sprintf("🔍 Why is '%s' installed?", packageName)))

	// Generate the dependency graph
	depGraph, err := getDependencyGraph()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to generate dependency graph: %v", err)))
		os.Exit(1)
	}

	// Find dependency paths for the package
	fmt.Println(utils.FormatProgress(fmt.Sprintf("Finding all dependency paths for '%s'...", packageName)))
	paths := depGraph.FindWhyPackageInstalled(packageName)

	// Display results
	if len(paths) == 0 {
		fmt.Println(utils.FormatWarning(fmt.Sprintf("No packages matching '%s' were found in the dependency tree.", packageName)))
		fmt.Println(utils.FormatInfo("Suggestions:"))
		fmt.Println("  - Check for typos in the package name")
		fmt.Println("  - Try a more general search term (substring match)")
		fmt.Println("  - Increase the analysis depth with --depth flag")
		os.Exit(0)
	}

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("Found %d dependency paths leading to '%s':", len(paths), packageName)))
	fmt.Println(utils.FormatDivider())

	// Display each dependency path
	for i, path := range paths {
		if len(path) > 0 {
			// Find the target package (last in path)
			targetPackage := path[len(path)-1]
			fmt.Println(utils.FormatHeader(fmt.Sprintf("Path %d to '%s':", i+1, targetPackage)))

			// Display the full path with indentation to show tree structure
			for j, pkg := range path {
				indent := strings.Repeat("  ", j)
				if j == len(path)-1 {
					fmt.Printf("%s└─ %s\n", indent, utils.FormatSuccess(pkg))
				} else {
					fmt.Printf("%s├─ %s\n", indent, pkg)
				}
			}
			fmt.Println()
		}
	}

	// Get AI insights on the dependencies
	userCfg, configErr := config.LoadUserConfig()
	if configErr == nil {
		fmt.Println(utils.FormatDivider())
		fmt.Println(utils.FormatHeader("🤖 AI Analysis"))

		// Create specific prompt for explaining dependencies
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("The following dependency paths explain why '%s' is included in a NixOS system:\n\n", packageName))

		for i, path := range paths {
			if i < 10 { // Limit to avoid token overflow
				sb.WriteString(fmt.Sprintf("Path %d: %s\n", i+1, strings.Join(path, " -> ")))
			}
		}

		if len(paths) > 10 {
			sb.WriteString(fmt.Sprintf("\n...and %d more paths\n", len(paths)-10))
		}

		sb.WriteString("\nPlease explain:\n")
		sb.WriteString("1. Why this package is included in the system\n")
		sb.WriteString("2. What core dependencies or system features require it\n")
		sb.WriteString("3. If it appears to be directly requested or pulled in as a dependency\n")
		sb.WriteString("4. Any insights or recommendations about this dependency situation\n")

		insights, aiErr := getAIInsights(sb.String(), userCfg)
		if aiErr != nil {
			fmt.Println(utils.FormatError(fmt.Sprintf("Failed to get AI insights: %v", aiErr)))
		} else {
			renderMarkdown(insights)
		}
	}
}

func runDepsConflicts() {
	fmt.Println(utils.FormatHeader("🛠️ Dependency Conflict Analysis"))

	// Generate the dependency graph
	depGraph, err := getDependencyGraph()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to generate dependency graph: %v", err)))
		os.Exit(1)
	}

	// Find dependency conflicts
	fmt.Println(utils.FormatProgress("Analyzing dependency conflicts..."))
	conflicts := depGraph.FindDependencyConflicts()

	// Display results
	if len(conflicts) == 0 {
		fmt.Println(utils.FormatSuccess("No dependency conflicts detected! 🎉"))
		fmt.Println(utils.FormatInfo("This means packages with the same name have consistent versions throughout your system."))
		os.Exit(0)
	}

	fmt.Println(utils.FormatWarning(fmt.Sprintf("Found %d potential dependency conflicts:", len(conflicts))))
	fmt.Println(utils.FormatDivider())

	// Display each conflict
	i := 0
	for pkgName, conflictingNodes := range conflicts {
		i++
		fmt.Println(utils.FormatHeader(fmt.Sprintf("Conflict %d: %s", i, pkgName)))
		fmt.Println("Different versions found:")

		for j, node := range conflictingNodes {
			versionStr := node.Version
			if versionStr == "" {
				versionStr = "unknown version"
			}

			// Format arrow depending on position in list
			arrow := "├─ "
			if j == len(conflictingNodes)-1 {
				arrow = "└─ "
			}

			fmt.Printf("  %s%s (%s)\n", arrow, node.Name, versionStr)
		}
		fmt.Println()
	}

	// Get AI recommendations for conflict resolution
	userCfg, configErr := config.LoadUserConfig()
	if configErr == nil {
		fmt.Println(utils.FormatDivider())
		fmt.Println(utils.FormatHeader("🤖 AI Analysis & Recommendations"))

		// Create specific prompt for conflict resolution
		var sb strings.Builder
		sb.WriteString("The following dependency conflicts were found in a NixOS system:\n\n")

		for pkgName, conflictingNodes := range conflicts {
			sb.WriteString(fmt.Sprintf("Package: %s\n", pkgName))
			sb.WriteString("Conflicting versions:\n")

			for _, node := range conflictingNodes {
				versionStr := node.Version
				if versionStr == "" {
					versionStr = "unknown version"
				}
				sb.WriteString(fmt.Sprintf("- %s (%s)\n", node.Name, versionStr))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("Please provide:\n")
		sb.WriteString("1. Analysis of potential problems these conflicts might cause\n")
		sb.WriteString("2. Specific recommendations for resolving each conflict\n")
		sb.WriteString("3. Best practices for avoiding similar conflicts in the future\n")
		sb.WriteString("4. NixOS-specific techniques for dependency conflict resolution\n")

		insights, aiErr := getAIInsights(sb.String(), userCfg)
		if aiErr != nil {
			fmt.Println(utils.FormatError(fmt.Sprintf("Failed to get AI insights: %v", aiErr)))
		} else {
			renderMarkdown(insights)
		}
	}
}

func runDepsOptimize() {
	fmt.Println(utils.FormatHeader("🚀 Dependency Optimization Analysis"))

	// Generate the dependency graph
	depGraph, err := getDependencyGraph()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to generate dependency graph: %v", err)))
		os.Exit(1)
	}

	// Analyze for optimization opportunities
	fmt.Println(utils.FormatProgress("Analyzing dependency graph for optimization opportunities..."))
	optimizations := depGraph.AnalyzeDependencyOptimizations()

	// Display results
	fmt.Println(utils.FormatSuccess(fmt.Sprintf("Found %d optimization opportunities:", len(optimizations))))
	fmt.Println(utils.FormatDivider())

	if len(optimizations) == 0 {
		fmt.Println(utils.FormatInfo("No immediate optimization opportunities detected. Your dependency graph appears well-structured! 🎉"))
	} else {
		for i, opt := range optimizations {
			fmt.Printf("%s %d: %s\n", utils.FormatKeyValue("Optimization", ""), i+1, opt)
		}
	}

	// Get AI recommendations for optimization
	userCfg, configErr := config.LoadUserConfig()
	if configErr == nil {
		fmt.Println(utils.FormatDivider())
		fmt.Println(utils.FormatHeader("🤖 AI Optimization Recommendations"))

		// Create specific prompt for optimization
		var sb strings.Builder
		sb.WriteString("Based on analysis of a NixOS system dependency graph, the following optimization opportunities were identified:\n\n")

		if len(optimizations) == 0 {
			sb.WriteString("No immediate optimization opportunities detected. The dependency graph appears well-structured.\n\n")
		} else {
			for i, opt := range optimizations {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, opt))
			}
			sb.WriteString("\n")
		}

		// Additional dependency stats for AI context
		sb.WriteString(fmt.Sprintf("Total unique packages: %d\n", len(depGraph.AllNodes)))
		sb.WriteString(fmt.Sprintf("Root-level dependencies: %d\n\n", len(depGraph.RootNodes)))

		sb.WriteString("Please provide:\n")
		sb.WriteString("1. Detailed analysis of the optimization opportunities\n")
		sb.WriteString("2. Specific NixOS techniques to implement these optimizations\n")
		sb.WriteString("3. General best practices for dependency management in Nix\n")
		sb.WriteString("4. Potential impact of these optimizations (build time, closure size, etc.)\n")
		sb.WriteString("5. Any potential trade-offs or considerations when implementing these optimizations\n")

		insights, aiErr := getAIInsights(sb.String(), userCfg)
		if aiErr != nil {
			fmt.Println(utils.FormatError(fmt.Sprintf("Failed to get AI insights: %v", aiErr)))
		} else {
			renderMarkdown(insights)
		}
	}
}

func runDepsGraph() {
	fmt.Println(utils.FormatHeader("📊 NixOS Dependency Graph Generator"))

	// Generate the dependency graph
	depGraph, err := getDependencyGraph()
	if err != nil {
		fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to generate dependency graph: %v", err)))
		os.Exit(1)
	}

	// Generate DOT representation
	fmt.Println(utils.FormatProgress("Generating DOT representation of dependency graph..."))
	dotContent := depGraph.GenerateDependencyGraphDOT()

	// Save to file or display instructions
	if depDotOutputPath != "" {
		// Ensure the file has a .dot extension
		if !strings.HasSuffix(depDotOutputPath, ".dot") {
			depDotOutputPath += ".dot"
		}

		// Write to file
		err := os.WriteFile(depDotOutputPath, []byte(dotContent), 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to write DOT file: %v", err)))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess(fmt.Sprintf("Dependency graph DOT file saved to: %s", depDotOutputPath)))
		fmt.Println(utils.FormatInfo("To generate a visual graph, use a tool like Graphviz:"))
		fmt.Printf("  dot -Tpng %s -o dependency-graph.png\n", depDotOutputPath)

		// Check if graphviz is installed
		_, graphvizErr := exec.LookPath("dot")
		if graphvizErr == nil {
			fmt.Println(utils.FormatInfo("Graphviz appears to be installed. You can generate the visualization directly:"))
			fmt.Printf("  dot -Tpng %s -o dependency-graph.png\n", depDotOutputPath)

			// Ask if user wants to generate the visualization now
			fmt.Print(utils.FormatInfo("Would you like to generate the visualization now? (y/n): "))
			var answer string
			_, _ = fmt.Scanln(&answer)

			if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
				outputPngPath := strings.TrimSuffix(depDotOutputPath, ".dot") + ".png"
				cmd := exec.Command("dot", "-Tpng", depDotOutputPath, "-o", outputPngPath)
				cmd.Stderr = os.Stderr

				fmt.Println(utils.FormatProgress("Generating visualization..."))
				err := cmd.Run()
				if err != nil {
					fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("Failed to generate visualization: %v", err)))
				} else {
					fmt.Println(utils.FormatSuccess(fmt.Sprintf("Visualization saved to: %s", outputPngPath)))
				}
			}
		} else {
			fmt.Println(utils.FormatInfo("Graphviz doesn't appear to be installed. Install it with:"))
			fmt.Println("  nix-env -iA nixpkgs.graphviz")
			fmt.Println("  # or add it to your configuration:")
			fmt.Println("  environment.systemPackages = [ pkgs.graphviz ];")
		}
	} else {
		// No output path specified, show a preview and instructions
		fmt.Println(utils.FormatWarning("No output file specified. Use --output to save the DOT file."))
		fmt.Println(utils.FormatInfo("Preview of DOT file content (first 10 lines):"))

		// Show first few lines as preview
		lines := strings.Split(dotContent, "\n")
		previewLines := 10
		if len(lines) < previewLines {
			previewLines = len(lines)
		}

		for i := 0; i < previewLines; i++ {
			fmt.Println("  " + lines[i])
		}

		if len(lines) > previewLines {
			fmt.Println("  ...")
			fmt.Printf("  (%d more lines)\n", len(lines)-previewLines)
		}

		fmt.Println(utils.FormatInfo("To save the DOT file, run:"))
		fmt.Printf("  nixai deps graph --output dependency-graph.dot\n")
	}
}

// Helper function to determine the NixOS configuration path
func determineConfigPath() (string, bool) {
	cfgPath := depNixosConfigPath // From flag
	if cfgPath == "" {
		userCfg, err := config.LoadUserConfig()
		if err == nil && userCfg.NixosFolder != "" {
			cfgPath = userCfg.NixosFolder
			fmt.Println(utils.FormatInfo(fmt.Sprintf("Using NixOS configuration path from user config: %s", cfgPath)))
		}
	}

	if cfgPath == "" {
		// Try to auto-detect common paths
		commonPaths := []string{
			"/etc/nixos/flake.nix",
			"/etc/nixos/configuration.nix",
			os.ExpandEnv("$HOME/.config/nixos/flake.nix"),
			os.ExpandEnv("$HOME/.config/nixos/configuration.nix"),
			"./flake.nix", // Check current directory
		}
		for _, p := range commonPaths {
			if utils.IsFile(p) {
				// If flake.nix is found, try to get its directory
				if strings.HasSuffix(p, "flake.nix") {
					cfgPath = filepath.Dir(p)
				} else {
					cfgPath = p
				}
				fmt.Println(utils.FormatInfo(fmt.Sprintf("Auto-detected NixOS configuration entrypoint: %s", p)))
				fmt.Println(utils.FormatInfo(fmt.Sprintf("Using configuration directory/file: %s", cfgPath)))
				break
			}
		}
	}

	if cfgPath == "" {
		fmt.Fprintln(os.Stderr, utils.FormatError("NixOS configuration path not found."))
		fmt.Fprintln(os.Stderr, utils.FormatInfo("Please specify the path using the --nixos-path flag or set 'nixos_folder' in your nixai config."))
		os.Exit(1)
	}

	if !utils.DirExists(cfgPath) && !utils.IsFile(cfgPath) {
		fmt.Fprintln(os.Stderr, utils.FormatError(fmt.Sprintf("NixOS configuration file does not exist: %s", cfgPath)))
		os.Exit(1)
	}

	fmt.Println(utils.FormatKeyValue("Configuration Path", cfgPath))
	fmt.Println(utils.FormatDivider())

	// Determine if it's a flake or legacy config
	isFlake := false
	if utils.IsDirectory(cfgPath) {
		if utils.IsFile(filepath.Join(cfgPath, "flake.nix")) {
			isFlake = true
		}
	} else { // cfgPath is a file
		if strings.HasSuffix(cfgPath, "flake.nix") {
			isFlake = true
			cfgPath = filepath.Dir(cfgPath) // for flake analysis, we need the directory
		}
	}

	return cfgPath, isFlake
}

// Helper function to generate dependency graph
func generateDependencyGraph(cfgPath string, isFlake bool) (*nixos.DependencyGraph, error) {
	var depGraph *nixos.DependencyGraph
	var err error

	if isFlake {
		fmt.Println(utils.FormatInfo("Analyzing as Flake-based configuration..."))
		if depNixosHostname == "" {
			// Attempt to get hostname from system
			hn, herr := os.Hostname()
			if herr == nil {
				depNixosHostname = strings.Split(hn, ".")[0] // Use short hostname
				fmt.Println(utils.FormatInfo(fmt.Sprintf("Auto-detected hostname: %s (use --hostname to override)", depNixosHostname)))
			} else {
				fmt.Fprintln(os.Stderr, utils.FormatError("Failed to auto-detect hostname for flake analysis."))
				fmt.Fprintln(os.Stderr, utils.FormatInfo("Please specify the NixOS configuration hostname using the --hostname flag (e.g., from flake.nix#nixosConfigurations.<hostname>)"))
				os.Exit(1)
			}
		}
		fmt.Println(utils.FormatInfo(fmt.Sprintf("Using max analysis depth: %d", depMaxDepth)))
		depGraph, err = nixos.AnalyzeFlakeDependencies(cfgPath, depNixosHostname)
	} else {
		fmt.Println(utils.FormatInfo("Analyzing as legacy (configuration.nix)-based configuration..."))
		fmt.Println(utils.FormatInfo(fmt.Sprintf("Using max analysis depth: %d", depMaxDepth)))
		depGraph, err = nixos.AnalyzeLegacyDependencies(cfgPath)
	}

	return depGraph, err
}

// Helper function for getting AI insights
func getAIInsights(prompt string, userCfg *config.UserConfig) (string, error) {
	fmt.Println(utils.FormatProgress("Getting AI insights..."))

	// Use the new ProviderManager system
	provider, err := GetLegacyAIProvider(userCfg, logger.NewLogger())
	if err != nil {
		// Fall back to ollama legacy provider on error
		provider = ai.NewOllamaLegacyProvider("llama3")
	}

	// Query AI provider
	return provider.Query(prompt)
}

// Helper function to render markdown with glamour
func renderMarkdown(markdown string) {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)
	rendered, renderErr := renderer.Render(markdown)
	if renderErr != nil {
		fmt.Println(utils.FormatError(fmt.Sprintf("Failed to render AI insights: %v", renderErr)))
		fmt.Println(markdown) // Fallback to plain text
	} else {
		fmt.Print(rendered)
	}
}

// formatDependencyGraphForAI converts the dependency graph to a string format suitable for AI analysis
func formatDependencyGraphForAI(depGraph *nixos.DependencyGraph) string {
	if depGraph == nil || len(depGraph.RootNodes) == 0 {
		return "No dependencies found."
	}

	var sb strings.Builder
	sb.WriteString("NixOS Dependency Graph Analysis:\n\n")
	sb.WriteString(fmt.Sprintf("Total nodes analyzed: %d\n", len(depGraph.AllNodes)))
	sb.WriteString(fmt.Sprintf("Root-level dependencies: %d\n\n", len(depGraph.RootNodes)))

	sb.WriteString("Dependency Tree:\n")
	formatNodeForAI(depGraph.RootNodes, &sb, "")

	return sb.String()
}

// formatNodeForAI recursively formats dependency nodes for AI analysis
func formatNodeForAI(nodes []*nixos.DependencyNode, sb *strings.Builder, prefix string) {
	for i, node := range nodes {
		connector := "├── "
		childPrefix := prefix + "│   "
		if i == len(nodes)-1 {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		versionStr := ""
		if node.Version != "" {
			versionStr = fmt.Sprintf(" (v%s)", node.Version)
		}

		_, _ = fmt.Fprintf(sb, "%s%s%s%s\n", prefix, connector, node.Name, versionStr)

		if len(node.Dependencies) > 0 {
			formatNodeForAI(node.Dependencies, sb, childPrefix)
		}
	}
}

// getAIInsightsForDependencies uses AI to analyze the dependency graph and provide insights
func getAIInsightsForDependencies(depGraph *nixos.DependencyGraph, userCfg *config.UserConfig) (string, error) {
	fmt.Println(utils.FormatProgress("Analyzing dependencies with AI..."))

	// Format dependency graph for AI analysis
	depGraphText := formatDependencyGraphForAI(depGraph)

	// Create AI prompt for dependency analysis
	prompt := fmt.Sprintf(`Analyze this NixOS dependency graph and provide insights on potential issues, optimization opportunities, security considerations, and best practices.

%s

Please provide:
1. **Overview**: Brief summary of the dependency structure
2. **Potential Issues**: Any concerning patterns or dependencies
3. **Security Considerations**: Security-related observations
4. **Optimization Opportunities**: Ways to improve the configuration
5. **Best Practices**: Recommendations for better dependency management
6. **Notable Dependencies**: Highlight any interesting or important packages

Format your response in Markdown for better readability.`, depGraphText)

	// Use the shared function to get AI insights
	return getAIInsights(prompt, userCfg)
}

func displayDependencyGraph(nodes []*nixos.DependencyNode, prefix string) {
	for i, node := range nodes {
		connector := "├── "
		childPrefix := prefix + "│   "
		if i == len(nodes)-1 {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		versionStr := ""
		if node.Version != "" {
			versionStr = fmt.Sprintf(" (v%s)", node.Version)
		}
		fmt.Printf("%s%s%s%s\n", prefix, connector, node.Name, versionStr)

		if len(node.Dependencies) > 0 {
			displayDependencyGraph(node.Dependencies, childPrefix)
		}
	}
}

// getDependencyGraph is a helper function that handles getting the dependency graph
// for all deps subcommands. It respects the --current flag and configuration paths.
func getDependencyGraph() (*nixos.DependencyGraph, error) {
	// Check if user wants to analyze current system
	if depCurrentSystem {
		return nixos.AnalyzeCurrentSystemDependencies()
	}

	// Original configuration-based analysis
	// 1. Determine NixOS configuration path
	cfgPath := depNixosConfigPath // From flag
	if cfgPath == "" {
		userCfg, err := config.LoadUserConfig()
		if err == nil && userCfg.NixosFolder != "" {
			cfgPath = userCfg.NixosFolder
		}
	}

	if cfgPath == "" {
		// Try to auto-detect common paths
		commonPaths := []string{
			"/etc/nixos/flake.nix",
			"/etc/nixos/configuration.nix",
			os.ExpandEnv("$HOME/.config/nixos/flake.nix"),
			os.ExpandEnv("$HOME/.config/nixos/configuration.nix"),
		}
		for _, p := range commonPaths {
			if utils.IsFile(p) {
				// If flake.nix is found, try to get its directory
				if strings.HasSuffix(p, "flake.nix") {
					cfgPath = filepath.Dir(p)
				} else {
					cfgPath = p
				}
				break
			}
		}
	}

	if cfgPath == "" {
		return nil, fmt.Errorf("NixOS configuration path not found. Please specify the path using the --nixos-path flag, set 'nixos_folder' in your nixai config, or use --current to analyze the running system")
	}

	if !utils.DirExists(cfgPath) && !utils.IsFile(cfgPath) {
		return nil, fmt.Errorf("NixOS configuration file does not exist: %s", cfgPath)
	}

	// Determine if it's a flake or legacy config
	isFlake := false
	actualConfigFile := cfgPath
	if utils.IsDirectory(cfgPath) {
		if utils.IsFile(filepath.Join(cfgPath, "flake.nix")) {
			isFlake = true
			actualConfigFile = filepath.Join(cfgPath, "flake.nix")
		} else if utils.IsFile(filepath.Join(cfgPath, "configuration.nix")) {
			actualConfigFile = filepath.Join(cfgPath, "configuration.nix")
		} else {
			return nil, fmt.Errorf("no flake.nix or configuration.nix found in directory: %s", cfgPath)
		}
	} else { // cfgPath is a file
		if strings.HasSuffix(cfgPath, "flake.nix") {
			isFlake = true
			cfgPath = filepath.Dir(cfgPath) // for flake analysis, we need the directory
		}
	}

	if isFlake {
		if depNixosHostname == "" {
			// Attempt to get hostname from system
			hn, herr := os.Hostname()
			if herr == nil {
				depNixosHostname = strings.Split(hn, ".")[0] // Use short hostname
			} else {
				return nil, fmt.Errorf("failed to auto-detect hostname for flake analysis. Please specify the NixOS configuration hostname using the --hostname flag")
			}
		}
		return nixos.AnalyzeFlakeDependencies(cfgPath, depNixosHostname)
	} else {
		return nixos.AnalyzeLegacyDependencies(actualConfigFile)
	}
}
