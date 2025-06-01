package nixos

import (
	"encoding/json"
	"fmt"
	"nix-ai-help/pkg/utils"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// DependencyNode represents a single package or module in the dependency graph.
type DependencyNode struct {
	Name         string
	StorePath    string // The store path of the derivation
	Version      string // Extracted version, if possible
	Dependencies []*DependencyNode
	// TODO: Add more fields like Type (package, module, input), Inputs (for flake inputs specifically)
}

// DependencyGraph represents the entire dependency structure.
// For now, it can be a list of top-level nodes or a map.
type DependencyGraph struct {
	RootNodes []*DependencyNode
	AllNodes  map[string]*DependencyNode // Map store path to node for easy lookup
}

// AnalyzeFlakeDependencies attempts to parse dependencies from a flake-based NixOS configuration.
// flakeDir is the directory containing the flake.nix file.
// hostname is the target NixOS configuration hostname (e.g., from flake.nix#nixosConfigurations.<hostname>).
func AnalyzeFlakeDependencies(flakeDir string, hostname string) (*DependencyGraph, error) {
	fmt.Println(utils.FormatInfo(fmt.Sprintf("Analyzing flake dependencies in: %s for host: %s", flakeDir, hostname)))

	// Step 1: Get the store path of the top-level system derivation
	targetSystem := fmt.Sprintf(".#nixosConfigurations.%s.config.system.build.toplevel", hostname)
	buildCmd := exec.Command("nix", "build", targetSystem, "--no-link", "--print-out-paths")
	buildCmd.Dir = flakeDir

	fmt.Println(utils.FormatProgress(fmt.Sprintf("Executing: %s in %s", buildCmd.String(), flakeDir)))
	buildOut, err := buildCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nix build command failed to get system store path: %v\nStderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("nix build command failed to get system store path: %v", err)
	}

	systemStorePath := filepath.Clean(string(buildOut))
	if systemStorePath == "" {
		return nil, fmt.Errorf("failed to get system store path, nix build output was empty")
	}
	fmt.Println(utils.FormatSuccess(fmt.Sprintf("System store path: %s", systemStorePath)))

	// Step 2: Initialize graph and map for visited nodes to avoid cycles and redundant processing
	graph := &DependencyGraph{
		RootNodes: []*DependencyNode{},
		AllNodes:  make(map[string]*DependencyNode),
	}
	visited := make(map[string]bool)

	// Step 3: Recursively fetch and parse derivations
	rootNode, err := fetchAndParseDerivationRecursive(systemStorePath, graph, visited, 0, 5) // Limit depth for now
	if err != nil {
		return nil, fmt.Errorf("failed to recursively parse derivations: %w", err)
	}
	graph.RootNodes = append(graph.RootNodes, rootNode)

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("Successfully parsed %d unique derivations.", len(graph.AllNodes))))
	return graph, nil
}

// fetchAndParseDerivationRecursive is a helper to build the dependency tree.
// currentDepth and maxDepth are used to prevent infinite recursion or excessive processing.
func fetchAndParseDerivationRecursive(storePath string, graph *DependencyGraph, visited map[string]bool, currentDepth int, maxDepth int) (*DependencyNode, error) {
	if currentDepth >= maxDepth {
		fmt.Println(utils.FormatWarning(fmt.Sprintf("Reached max depth (%d) at %s, stopping further recursion for this branch.", maxDepth, storePath)))
		// Return a placeholder node if max depth is reached
		return &DependencyNode{Name: filepath.Base(storePath) + " (max depth reached)", StorePath: storePath}, nil
	}

	if visited[storePath] {
		// If already visited, return the existing node from the graph to link it correctly
		if existingNode, ok := graph.AllNodes[storePath]; ok {
			return existingNode, nil
		}
		// This case should ideally not happen if nodes are always added to AllNodes when created
		return &DependencyNode{Name: filepath.Base(storePath) + " (visited)", StorePath: storePath}, nil
	}
	visited[storePath] = true

	fmt.Println(utils.FormatProgress(fmt.Sprintf("Fetching derivation: %s (depth %d)", storePath, currentDepth)))

	// First, get the derivation path using nix path-info
	pathInfoCmd := exec.Command("nix", "path-info", "--derivation", storePath)
	drvPathBytes, err := pathInfoCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nix path-info --derivation %s failed: %v\nStderr: %s", storePath, err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("nix path-info --derivation %s failed: %v", storePath, err)
	}

	drvPath := strings.TrimSpace(string(drvPathBytes))
	if drvPath == "" {
		return nil, fmt.Errorf("could not get derivation path for %s", storePath)
	}

	// Now use the derivation path with nix derivation show
	showCmd := exec.Command("nix", "derivation", "show", drvPath)
	derivJSON, err := showCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nix derivation show %s failed: %v\nStderr: %s", drvPath, err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("nix derivation show %s failed: %v", drvPath, err)
	}

	derivInfo, err := ParseDerivationOutput(derivJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse derivation output for %s: %w", storePath, err)
	}

	currentNode := &DependencyNode{
		Name:         derivInfo.Name, // Use the name from the derivation info
		StorePath:    storePath,
		Dependencies: []*DependencyNode{},
		Version:      extractVersionFromName(derivInfo.Name), // Extract version from derivation name
	}
	graph.AllNodes[storePath] = currentNode

	// Recursively process input derivations
	for _, inputDrvPaths := range derivInfo.InputDrvs {
		for _, inputDrvPath := range inputDrvPaths {
			cleanInputPath := filepath.Clean(inputDrvPath)
			depNode, err := fetchAndParseDerivationRecursive(cleanInputPath, graph, visited, currentDepth+1, maxDepth)
			if err != nil {
				fmt.Println(utils.FormatError(fmt.Sprintf("Error processing dependency %s for %s: %v. Skipping this dependency.", cleanInputPath, storePath, err)))
				continue // Skip this dependency but continue with others
			}
			currentNode.Dependencies = append(currentNode.Dependencies, depNode)
		}
	}

	return currentNode, nil
}

// FindWhyPackageInstalled finds all paths from root nodes to the target package in the dependency graph.
// Returns a list of dependency paths, each path is a list of package names from a root node to the target.
func (graph *DependencyGraph) FindWhyPackageInstalled(packageName string) [][]string {
	var results [][]string

	// Create a map for case-insensitive matching
	packageNameLower := strings.ToLower(packageName)

	// A DFS function to find all paths to the package
	var findPaths func(node *DependencyNode, currentPath []string)
	findPaths = func(node *DependencyNode, currentPath []string) {
		// Add current node to path
		path := append(currentPath, node.Name)

		// Check if this is the target package
		nodeName := strings.ToLower(node.Name)
		if strings.Contains(nodeName, packageNameLower) {
			// Found a path, add it to results
			results = append(results, path)
			return
		}

		// Continue DFS through all dependencies
		for _, dep := range node.Dependencies {
			findPaths(dep, path)
		}
	}

	// Start DFS from all root nodes
	for _, root := range graph.RootNodes {
		findPaths(root, []string{})
	}

	return results
}

// FindDependencyConflicts identifies packages with version conflicts.
// Returns a map where keys are package names and values are lists of conflicting nodes.
func (graph *DependencyGraph) FindDependencyConflicts() map[string][]*DependencyNode {
	conflicts := make(map[string][]*DependencyNode)
	packageVersions := make(map[string]map[string]*DependencyNode) // packageName -> version -> node

	// Process all nodes to identify unique name+version combinations
	for _, node := range graph.AllNodes {
		// Extract base package name (removing version suffix)
		baseName := extractBasePackageName(node.Name)
		if baseName == "" {
			continue // Skip if we couldn't extract a meaningful base name
		}

		// Initialize map for this package if needed
		if _, exists := packageVersions[baseName]; !exists {
			packageVersions[baseName] = make(map[string]*DependencyNode)
		}

		// Add node to the appropriate version bin
		version := node.Version
		if version == "" {
			version = "unknown" // Handle unknown versions
		}

		packageVersions[baseName][version] = node
	}

	// Identify conflicts (packages with multiple versions)
	for pkgName, versions := range packageVersions {
		if len(versions) > 1 {
			// Create a list of conflicting nodes
			var conflictingNodes []*DependencyNode
			for _, node := range versions {
				conflictingNodes = append(conflictingNodes, node)
			}
			conflicts[pkgName] = conflictingNodes
		}
	}

	return conflicts
}

// extractBasePackageName tries to extract the base package name by removing version information.
func extractBasePackageName(fullName string) string {
	// This regex pattern attempts to match and remove version-like suffixes
	re := regexp.MustCompile(`^(.*?)(?:-[0-9]+(?:\.[0-9]+)*(?:[a-zA-Z0-9._+-]+)?)?$`)
	matches := re.FindStringSubmatch(fullName)
	if len(matches) > 1 {
		return matches[1]
	}
	return fullName // Return original if pattern doesn't match
}

// GenerateDependencyGraphDOT generates a DOT representation of the dependency graph for visualization.
func (graph *DependencyGraph) GenerateDependencyGraphDOT() string {
	var sb strings.Builder

	// DOT file header
	sb.WriteString("digraph dependency_graph {\n")
	sb.WriteString("  rankdir=LR;\n") // Left-to-right layout
	sb.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n\n")

	// Track processed nodes to avoid duplicates
	processed := make(map[string]bool)

	// Helper function to process nodes and edges recursively
	var processDOTNode func(node *DependencyNode)
	processDOTNode = func(node *DependencyNode) {
		if processed[node.StorePath] {
			return // Skip already processed nodes
		}
		processed[node.StorePath] = true

		// Node definition with label
		nodeName := fmt.Sprintf("\"%s\"", node.StorePath) // Use store path as unique ID
		nodeLabel := node.Name
		if node.Version != "" {
			nodeLabel = fmt.Sprintf("%s\\nv%s", nodeLabel, node.Version)
		}
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\"];\n", nodeName, nodeLabel))

		// Process edges to dependencies
		for _, dep := range node.Dependencies {
			depName := fmt.Sprintf("\"%s\"", dep.StorePath)
			sb.WriteString(fmt.Sprintf("  %s -> %s;\n", nodeName, depName))
			processDOTNode(dep)
		}
	}

	// Process all root nodes
	for _, root := range graph.RootNodes {
		processDOTNode(root)
	}

	// DOT file footer
	sb.WriteString("}\n")

	return sb.String()
}

// AnalyzeDependencyOptimizations identifies potential dependency optimizations.
// Returns a list of optimization suggestions.
func (graph *DependencyGraph) AnalyzeDependencyOptimizations() []string {
	var optimizations []string

	// Find circular dependencies
	circularDeps := graph.findCircularDependencies()
	if len(circularDeps) > 0 {
		optimizations = append(optimizations, fmt.Sprintf("Found %d circular dependency chains. Consider breaking these cycles.", len(circularDeps)))
		for i, path := range circularDeps {
			if i < 5 { // Limit to showing only a few examples
				optimizations = append(optimizations, fmt.Sprintf("  Circular path %d: %s", i+1, strings.Join(path, " -> ")))
			}
		}
		if len(circularDeps) > 5 {
			optimizations = append(optimizations, fmt.Sprintf("  ... and %d more circular dependencies", len(circularDeps)-5))
		}
	}

	// Identify duplicated functionality
	// (Note: This is a simplified implementation. Real-world detection would be more complex.)
	packageCategories := map[string][]string{
		"http-client":      {"http-client", "curl", "wget", "axios", "fetch", "request"},
		"compression":      {"zlib", "gzip", "bzip2", "xz", "lz4"},
		"image-processing": {"imagemagick", "graphicsmagick", "libpng", "libjpeg"},
		// Add more categories as needed
	}

	categoryInstances := make(map[string]map[string]bool)
	for category, keywords := range packageCategories {
		categoryInstances[category] = make(map[string]bool)
		for _, node := range graph.AllNodes {
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(node.Name), strings.ToLower(keyword)) {
					categoryInstances[category][node.Name] = true
					break
				}
			}
		}
	}

	for category, instances := range categoryInstances {
		if len(instances) > 1 {
			var pkgs []string
			for pkg := range instances {
				pkgs = append(pkgs, pkg)
			}
			optimizations = append(optimizations, fmt.Sprintf("Multiple packages providing similar %s functionality: %s", category, strings.Join(pkgs, ", ")))
		}
	}

	// Check for dependency depth
	maxDepth := 0
	var deepestPath []string
	graph.findDeepestPath(&maxDepth, &deepestPath)

	if maxDepth > 10 {
		optimizations = append(optimizations, fmt.Sprintf("Deep dependency chain detected (depth %d). Consider flattening: %s", maxDepth, strings.Join(deepestPath, " -> ")))
	}

	return optimizations
}

// findCircularDependencies detects circular dependency chains.
// Returns a list of circular paths, each represented as a list of package names.
func (graph *DependencyGraph) findCircularDependencies() [][]string {
	var circularPaths [][]string

	for storePath, node := range graph.AllNodes {
		// Start DFS from this node
		visited := make(map[string]bool)
		path := []string{}
		graph.dfsCircular(node, visited, path, &circularPaths, storePath)
	}

	return circularPaths
}

// dfsCircular performs depth-first search to find circular dependencies.
func (graph *DependencyGraph) dfsCircular(node *DependencyNode, visited map[string]bool, path []string, results *[][]string, targetPath string) {
	// Check if we've found a cycle back to the target
	if len(path) > 0 && node.StorePath == targetPath {
		cycle := append(path, node.Name)
		*results = append(*results, cycle)
		return
	}

	// Skip if already visited in this path
	if visited[node.StorePath] {
		return
	}

	// Mark as visited and add to path
	visited[node.StorePath] = true
	path = append(path, node.Name)

	// Continue DFS through dependencies
	for _, dep := range node.Dependencies {
		graph.dfsCircular(dep, visited, path, results, targetPath)
	}
}

// findDeepestPath finds the deepest dependency path in the graph.
func (graph *DependencyGraph) findDeepestPath(maxDepth *int, deepestPath *[]string) {
	for _, root := range graph.RootNodes {
		currentPath := []string{root.Name}
		graph.dfsDepth(root, currentPath, maxDepth, deepestPath, 1)
	}
}

// dfsDepth performs depth-first search to find the deepest path.
func (graph *DependencyGraph) dfsDepth(node *DependencyNode, currentPath []string, maxDepth *int, deepestPath *[]string, depth int) {
	if depth > *maxDepth {
		*maxDepth = depth
		*deepestPath = make([]string, len(currentPath))
		copy(*deepestPath, currentPath)
	}

	for _, dep := range node.Dependencies {
		path := append(currentPath, dep.Name)
		graph.dfsDepth(dep, path, maxDepth, deepestPath, depth+1)
	}
}

// AnalyzeLegacyDependencies attempts to parse dependencies from a legacy configuration.nix.
// This will be more complex and might involve direct file parsing or nix-instantiate.
func AnalyzeLegacyDependencies(configFilePath string) (*DependencyGraph, error) {
	fmt.Println(utils.FormatInfo(fmt.Sprintf("Analyzing legacy dependencies for: %s", configFilePath)))

	// Step 1: Get the store path of the top-level system derivation
	absConfigPath, err := filepath.Abs(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", configFilePath, err)
	}

	cmd := exec.Command("nix-build", "<nixpkgs/nixos>", "-A", "system", "--no-link", "--print-out-paths", "-I", fmt.Sprintf("nixos-config=%s", absConfigPath))
	fmt.Println(utils.FormatProgress(fmt.Sprintf("Executing: %s", cmd.String())))
	buildOut, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nix-build command failed to get system store path for legacy config: %v\nStderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("nix-build command failed to get system store path for legacy config: %v", err)
	}

	systemStorePath := filepath.Clean(string(buildOut))
	if systemStorePath == "" {
		return nil, fmt.Errorf("failed to get system store path for legacy config, nix-build output was empty")
	}
	fmt.Println(utils.FormatSuccess(fmt.Sprintf("Legacy system store path: %s", systemStorePath)))

	// Step 2: Initialize graph and map for visited nodes
	graph := &DependencyGraph{
		RootNodes: []*DependencyNode{},
		AllNodes:  make(map[string]*DependencyNode),
	}
	visited := make(map[string]bool)

	// Step 3: Recursively fetch and parse derivations using the same helper as flakes
	rootNodeName := filepath.Base(configFilePath)
	rootNode, err := fetchAndParseDerivationRecursive(systemStorePath, graph, visited, 0, 5) // Limit depth
	if err != nil {
		return nil, fmt.Errorf("failed to recursively parse derivations for legacy config: %w", err)
	}
	if rootNode != nil {
		rootNode.Name = fmt.Sprintf("System (from %s)", rootNodeName)
	}
	graph.RootNodes = append(graph.RootNodes, rootNode)

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("Successfully parsed %d unique derivations for legacy config.", len(graph.AllNodes))))
	return graph, nil
}

// DerivationInfo matches the structure of `nix derivation show` output.
// We only care about a subset of fields for now.
type DerivationInfo struct {
	Name      string                `json:"name"`
	Outputs   map[string]OutputInfo `json:"outputs"`
	InputDrvs map[string][]string   `json:"inputDrvs"` // Store paths of input derivations
	InputSrcs []string              `json:"inputSrcs"` // Store paths of input sources
	System    string                `json:"system"`
	Builder   string                `json:"builder"`
	Args      []string              `json:"args"`
	Env       map[string]string     `json:"env"`
}

// OutputInfo holds information about a derivation's outputs.
type OutputInfo struct {
	Path string `json:"path"`
}

// ParseDerivationOutput parses the JSON output of `nix derivation show <store-path>`.
func ParseDerivationOutput(jsonData []byte) (*DerivationInfo, error) {
	var derivationInfos map[string]DerivationInfo // The output is a map with the store path as key
	if err := json.Unmarshal(jsonData, &derivationInfos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal derivation JSON: %w", err)
	}

	// Assuming there's only one top-level key (the derivation path itself)
	for _, drvInfo := range derivationInfos {
		return &drvInfo, nil // Return the first (and should be only) derivation info
	}
	return nil, fmt.Errorf("no derivation info found in JSON output")
}

// extractVersionFromName tries to find a version string (e.g., 1.2.3, v1.2.3, 1.2.3-alpha)
// from a typical derivation name (e.g., package-name-1.2.3).
func extractVersionFromName(name string) string {
	re := regexp.MustCompile(`(?:[_-]|[^a-zA-Z0-9])(v?[0-9]+(?:\.[0-9]+)*(?:[a-zA-Z0-9._+-]+)?)$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) > 1 {
		version := matches[1]
		version = strings.Trim(version, "._-")
		if regexp.MustCompile(`^v?[0-9]`).MatchString(version) {
			return version
		}
	}
	return ""
}
