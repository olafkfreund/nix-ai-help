package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

// GitHub API structures for searching code
type GitHubSearchResponse struct {
	TotalCount int                  `json:"total_count"`
	Items      []GitHubSearchResult `json:"items"`
}

type GitHubSearchResult struct {
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	Sha        string               `json:"sha"`
	URL        string               `json:"url"`
	GitURL     string               `json:"git_url"`
	HTMLURL    string               `json:"html_url"`
	Repository GitHubRepositoryInfo `json:"repository"`
	Score      float64              `json:"score"`
}

type GitHubRepositoryInfo struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	FullName    string      `json:"full_name"`
	Owner       GitHubOwner `json:"owner"`
	HTMLURL     string      `json:"html_url"`
	Description string      `json:"description"`
	Language    string      `json:"language"`
	StarCount   int         `json:"stargazers_count"`
	ForksCount  int         `json:"forks_count"`
	UpdatedAt   string      `json:"updated_at"`
}

type GitHubOwner struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

// GitHubContentResponse represents GitHub file content API response
type GitHubContentResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

// Template represents a configuration template
type Template struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Category    string            `yaml:"category"`
	Tags        []string          `yaml:"tags"`
	Source      string            `yaml:"source"`      // "builtin", "github", "custom"
	GitHubRepo  string            `yaml:"github_repo"` // For GitHub templates
	FilePath    string            `yaml:"file_path"`   // Path within repo
	Content     string            `yaml:"content"`     // Template content
	Metadata    map[string]string `yaml:"metadata"`
}

// Snippet represents a saved configuration snippet
type Snippet struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Content     string            `yaml:"content"`
	CreatedAt   time.Time         `yaml:"created_at"`
	Source      string            `yaml:"source"` // "user", "template", "github"
	Metadata    map[string]string `yaml:"metadata"`
}

// TemplateManager manages templates and snippets
type TemplateManager struct {
	configDir string
	logger    *logger.Logger
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(configDir string, log *logger.Logger) *TemplateManager {
	if configDir == "" {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config", "nixai")
	}

	// Ensure config directory exists
	_ = os.MkdirAll(configDir, 0755)
	_ = os.MkdirAll(filepath.Join(configDir, "templates"), 0755)
	_ = os.MkdirAll(filepath.Join(configDir, "snippets"), 0755)

	return &TemplateManager{
		configDir: configDir,
		logger:    log,
	}
}

// Methods for TemplateManager

// GetTemplate retrieves a specific template by name
func (tm *TemplateManager) GetTemplate(name string) (*Template, error) {
	// Check builtin templates first
	builtinTemplates := tm.LoadBuiltinTemplates()
	for _, template := range builtinTemplates {
		if template.Name == name {
			return &template, nil
		}
	}

	// Check saved custom templates
	customTemplates, err := tm.LoadCustomTemplates()
	if err == nil {
		for _, template := range customTemplates {
			if template.Name == name {
				return &template, nil
			}
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

// LoadCustomTemplates loads templates saved by the user
func (tm *TemplateManager) LoadCustomTemplates() ([]Template, error) {
	templatesDir := filepath.Join(tm.configDir, "templates")

	var templates []Template

	// Read all YAML files in templates directory
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
	if err != nil {
		return templates, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			tm.logger.Warn("Failed to read template file: " + file)
			continue
		}

		var template Template
		if err := yaml.Unmarshal(data, &template); err != nil {
			tm.logger.Warn("Failed to parse template file: " + file)
			continue
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// ApplyTemplate applies a template to the configuration
func (tm *TemplateManager) ApplyTemplate(template *Template, outputPath string, merge bool) error {
	content := template.Content

	// If no output path specified, use default
	if outputPath == "" {
		outputPath = "/etc/nixos/configuration.nix"

		// Check if we have permission to write to /etc/nixos
		if _, err := os.Stat("/etc/nixos"); os.IsNotExist(err) {
			// Fallback to current directory
			outputPath = "./configuration.nix"
		}
	}

	if merge {
		// Merge with existing configuration
		if existingContent, err := os.ReadFile(outputPath); err == nil {
			// Simple merge - in a real implementation, this would be more sophisticated
			content = string(existingContent) + "\n\n# Added from template: " + template.Name + "\n" + content
		}
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write configuration
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

// SaveTemplate saves a new template from a source
func (tm *TemplateManager) SaveTemplate(name, source, category, description string, tags []string) error {
	var content string
	var gitHubRepo, filePath string

	// Determine source type and read content
	if strings.HasPrefix(source, "http") {
		// GitHub URL
		var err error
		content, gitHubRepo, filePath, err = tm.fetchGitHubContent(source)
		if err != nil {
			return fmt.Errorf("failed to fetch GitHub content: %v", err)
		}
	} else {
		// Local file
		data, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", source, err)
		}
		content = string(data)
	}

	// Create template
	template := Template{
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        tags,
		Source:      "custom",
		GitHubRepo:  gitHubRepo,
		FilePath:    filePath,
		Content:     content,
		Metadata:    make(map[string]string),
	}

	// Set default description if empty
	if template.Description == "" {
		template.Description = "Custom template saved from " + source
	}

	// Set default category if empty
	if template.Category == "" {
		template.Category = "Custom"
	}

	// Save to file
	templatePath := filepath.Join(tm.configDir, "templates", name+".yaml")
	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %v", err)
	}

	if err := os.WriteFile(templatePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save template: %v", err)
	}

	return nil
}

// fetchGitHubContent fetches content from a GitHub URL
func (tm *TemplateManager) fetchGitHubContent(url string) (content, repo, path string, err error) {
	// Parse GitHub URL to extract repo and file path
	// Example: https://github.com/user/repo/blob/main/config.nix
	// Convert to raw URL: https://raw.githubusercontent.com/user/repo/main/config.nix

	if strings.Contains(url, "github.com") && strings.Contains(url, "/blob/") {
		// Parse the URL to extract parts
		parts := strings.Split(url, "/")
		if len(parts) >= 7 {
			user := parts[3]
			repoName := parts[4]
			branch := parts[6]
			filePath := strings.Join(parts[7:], "/")

			repo = fmt.Sprintf("%s/%s", user, repoName)
			path = filePath

			// Construct raw URL
			rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", user, repoName, branch, filePath)

			// Fetch content
			resp, err := http.Get(rawURL)
			if err != nil {
				return "", "", "", err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				return "", "", "", fmt.Errorf("failed to fetch content: HTTP %d", resp.StatusCode)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", "", "", err
			}

			content = string(data)
			return content, repo, path, nil
		}
	}

	return "", "", "", fmt.Errorf("unsupported URL format")
}

// GetCategories returns template categories with counts
func (tm *TemplateManager) GetCategories() map[string]int {
	categories := make(map[string]int)

	// Count builtin templates
	builtinTemplates := tm.LoadBuiltinTemplates()
	for _, template := range builtinTemplates {
		category := template.Category
		if category == "" {
			category = "General"
		}
		categories[category]++
	}

	// Count custom templates
	customTemplates, err := tm.LoadCustomTemplates()
	if err == nil {
		for _, template := range customTemplates {
			category := template.Category
			if category == "" {
				category = "General"
			}
			categories[category]++
		}
	}

	return categories
}

// LoadSnippets loads all saved snippets
func (tm *TemplateManager) LoadSnippets() ([]Snippet, error) {
	snippetsDir := filepath.Join(tm.configDir, "snippets")

	var snippets []Snippet

	// Read all YAML files in snippets directory
	files, err := filepath.Glob(filepath.Join(snippetsDir, "*.yaml"))
	if err != nil {
		return snippets, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			tm.logger.Warn("Failed to read snippet file: " + file)
			continue
		}

		var snippet Snippet
		if err := yaml.Unmarshal(data, &snippet); err != nil {
			tm.logger.Warn("Failed to parse snippet file: " + file)
			continue
		}

		snippets = append(snippets, snippet)
	}

	// Sort by creation time (newest first)
	sort.Slice(snippets, func(i, j int) bool {
		return snippets[i].CreatedAt.After(snippets[j].CreatedAt)
	})

	return snippets, nil
}

// SearchSnippets searches snippets by query
func (tm *TemplateManager) SearchSnippets(query string) ([]Snippet, error) {
	allSnippets, err := tm.LoadSnippets()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matches []Snippet

	for _, snippet := range allSnippets {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(snippet.Name), query) ||
			strings.Contains(strings.ToLower(snippet.Description), query) {
			matches = append(matches, snippet)
			continue
		}

		// Search in tags
		for _, tag := range snippet.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, snippet)
				break
			}
		}
	}

	return matches, nil
}

// SaveSnippet saves a new snippet
func (tm *TemplateManager) SaveSnippet(name, filePath, description string, tags []string) error {
	var content string

	// Read content from file or stdin
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", filePath, err)
		}
		content = string(data)
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %v", err)
			}
			content = string(data)
		} else {
			return fmt.Errorf("no content provided - specify --file or pipe content")
		}
	}

	// Create snippet
	snippet := Snippet{
		Name:        name,
		Description: description,
		Tags:        tags,
		Content:     content,
		CreatedAt:   time.Now(),
		Source:      "user",
		Metadata:    make(map[string]string),
	}

	// Set default description if empty
	if snippet.Description == "" {
		snippet.Description = "User-created snippet"
	}

	// Save to file
	snippetPath := filepath.Join(tm.configDir, "snippets", name+".yaml")
	data, err := yaml.Marshal(snippet)
	if err != nil {
		return fmt.Errorf("failed to marshal snippet: %v", err)
	}

	if err := os.WriteFile(snippetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save snippet: %v", err)
	}

	return nil
}

// GetSnippet retrieves a specific snippet by name
func (tm *TemplateManager) GetSnippet(name string) (*Snippet, error) {
	snippets, err := tm.LoadSnippets()
	if err != nil {
		return nil, err
	}

	for _, snippet := range snippets {
		if snippet.Name == name {
			return &snippet, nil
		}
	}

	return nil, fmt.Errorf("snippet not found: %s", name)
}

// ApplySnippet applies a snippet to configuration
func (tm *TemplateManager) ApplySnippet(name, outputPath string) error {
	snippet, err := tm.GetSnippet(name)
	if err != nil {
		return err
	}

	if outputPath == "" {
		// Output to stdout if no file specified
		fmt.Print(snippet.Content)
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write content
	if err := os.WriteFile(outputPath, []byte(snippet.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// RemoveSnippet removes a snippet
func (tm *TemplateManager) RemoveSnippet(name string) error {
	snippetPath := filepath.Join(tm.configDir, "snippets", name+".yaml")

	if _, err := os.Stat(snippetPath); os.IsNotExist(err) {
		return fmt.Errorf("snippet not found: %s", name)
	}

	if err := os.Remove(snippetPath); err != nil {
		return fmt.Errorf("failed to remove snippet: %v", err)
	}

	return nil
}

// getCategoryDescription provides descriptions for template categories
func getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"Desktop":     "Desktop environment configurations",
		"Gaming":      "Gaming-optimized configurations",
		"Server":      "Server and headless configurations",
		"Development": "Development environment setups",
		"Minimal":     "Minimal and lightweight configurations",
		"Hardware":    "Hardware-specific configurations",
		"Security":    "Security-hardened configurations",
		"Custom":      "User-created templates",
		"General":     "General purpose configurations",
	}

	if desc, exists := descriptions[category]; exists {
		return desc
	}
	return "Configuration templates"
}

// Main templates command
var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage NixOS configuration templates and snippets",
	Long: `Manage curated NixOS configuration templates with GitHub code search integration.

Browse and apply templates for common NixOS configurations including desktop environments, 
servers, development setups, and more. Templates are sourced from curated collections 
and real-world GitHub repositories.

Commands:
  list                    - Browse available templates
  search <query>          - Search templates by keyword or category  
  github <query>          - Search GitHub for NixOS configurations
  apply <name>            - Apply template to current configuration
  show <name>             - Show template details and content
  save <name> <file>      - Save configuration as template
  categories              - Show template categories

Examples:
  nixai templates list
  nixai templates search gaming
  nixai templates search desktop kde  
  nixai templates github "gaming nixos configuration"
  nixai templates apply desktop-minimal
  nixai templates show server-basic`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Templates list command
var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available NixOS configuration templates",
	Long:  "Browse all available curated NixOS configuration templates organized by category.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.FormatHeader("📚 Available NixOS Configuration Templates"))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Load builtin templates
		templates := tm.LoadBuiltinTemplates()
		if len(templates) == 0 {
			fmt.Println(utils.FormatWarning("No templates available"))
			return
		}

		// Group templates by category
		categories := make(map[string][]Template)
		for _, template := range templates {
			category := template.Category
			if category == "" {
				category = "General"
			}
			categories[category] = append(categories[category], template)
		}

		// Sort categories
		var sortedCategories []string
		for category := range categories {
			sortedCategories = append(sortedCategories, category)
		}
		sort.Strings(sortedCategories)

		// Display templates by category
		for _, category := range sortedCategories {
			fmt.Println(utils.FormatSubsection("🏷️ "+category, ""))
			templates := categories[category]

			// Sort templates by name
			sort.Slice(templates, func(i, j int) bool {
				return templates[i].Name < templates[j].Name
			})

			for _, template := range templates {
				tagsStr := ""
				if len(template.Tags) > 0 {
					tagsStr = " (" + strings.Join(template.Tags, ", ") + ")"
				}
				fmt.Printf("  %s%s\n",
					utils.FormatKeyValue(template.Name, template.Description),
					utils.FormatNote(tagsStr))
			}
			fmt.Println()
		}

		fmt.Println(utils.FormatTip("Use 'nixai templates show <name>' to view template details"))
		fmt.Println(utils.FormatTip("Use 'nixai templates search <query>' to find specific templates"))
		fmt.Println(utils.FormatTip("Use 'nixai templates github <query>' to search GitHub for more configurations"))
	},
}

// Templates search command
var templatesSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search NixOS templates by keyword or category",
	Long: `Search available NixOS configuration templates by keyword, tag, or category.

Examples:
  nixai templates search gaming
  nixai templates search desktop kde
  nixai templates search development
  nixai templates search server nginx`,
	Args: conditionalArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		fmt.Println(utils.FormatHeader("🔍 Searching Templates: " + query))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Search templates
		templates := tm.SearchTemplates(query)
		if len(templates) == 0 {
			fmt.Println(utils.FormatWarning("No templates found matching: " + query))
			fmt.Println()
			fmt.Println(utils.FormatTip("Try 'nixai templates github \"" + query + "\"' to search GitHub"))
			return
		}

		// Display results
		fmt.Printf("Found %d template(s):\n\n", len(templates))
		for i, template := range templates {
			fmt.Printf("%s. %s\n",
				utils.FormatNote(fmt.Sprintf("%d", i+1)),
				utils.FormatKeyValue(template.Name, template.Description))

			if template.Category != "" {
				fmt.Printf("   %s\n", utils.FormatNote("Category: "+template.Category))
			}

			if len(template.Tags) > 0 {
				fmt.Printf("   %s\n", utils.FormatNote("Tags: "+strings.Join(template.Tags, ", ")))
			}
			fmt.Println()
		}

		fmt.Println(utils.FormatTip("Use 'nixai templates show <name>' to view template details"))
		fmt.Println(utils.FormatTip("Use 'nixai templates apply <name>' to apply a template"))
	},
}

// Templates GitHub search command
var templatesGithubCmd = &cobra.Command{
	Use:   "github <query>",
	Short: "Search GitHub for NixOS configurations",
	Long: `Search GitHub repositories for real-world NixOS configurations using GitHub's code search API.

This command finds working NixOS configurations from the community, including:
- Desktop environment configurations
- Server setups and services
- Gaming optimizations
- Development environments
- Hardware-specific configurations

Examples:
  nixai templates github "gaming nixos configuration"
  nixai templates github "kde plasma nixos"
  nixai templates github "server nginx configuration.nix"
  nixai templates github "thinkpad nixos hardware"`,
	Args: conditionalArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		fmt.Println(utils.FormatHeader("🔍 Searching GitHub: " + query))
		fmt.Println()

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Search GitHub
		fmt.Println(utils.FormatProgress("Searching GitHub repositories..."))
		results, err := tm.SearchGitHub(query)
		if err != nil {
			fmt.Println(utils.FormatError("Error searching GitHub: " + err.Error()))
			fmt.Println(utils.FormatTip("Check your internet connection and try again"))
			return
		}

		if len(results.Items) == 0 {
			fmt.Println(utils.FormatWarning("No GitHub results found for: " + query))
			return
		}

		fmt.Printf("Found %d result(s) from GitHub:\n\n", len(results.Items))

		// Display first 10 results
		maxResults := 10
		if len(results.Items) < maxResults {
			maxResults = len(results.Items)
		}

		for i := 0; i < maxResults; i++ {
			result := results.Items[i]
			fmt.Printf("%s. %s\n",
				utils.FormatNote(fmt.Sprintf("%d", i+1)),
				utils.FormatKeyValue(result.Repository.FullName, result.Repository.Description))

			fmt.Printf("   %s\n", utils.FormatNote("File: "+result.Path))
			fmt.Printf("   %s\n", utils.FormatNote("⭐ "+fmt.Sprintf("%d", result.Repository.StarCount)))
			fmt.Printf("   %s\n", utils.FormatNote("🔗 "+result.HTMLURL))
			fmt.Println()
		}

		if len(results.Items) > maxResults {
			fmt.Printf(utils.FormatNote("... and %d more results\n"), len(results.Items)-maxResults)
			fmt.Println()
		}

		fmt.Println(utils.FormatTip("Click the URLs to view the configurations in your browser"))
		fmt.Println(utils.FormatTip("Use 'nixai templates save <name> <url>' to save promising configs as templates"))
	},
}

// Templates show command
var templatesShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template details and content",
	Long: `Display detailed information about a specific template including its content,
metadata, category, tags, and usage examples.

Examples:
  nixai templates show desktop-minimal
  nixai templates show gaming-config
  nixai templates show server-basic`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Find template
		template, err := tm.GetTemplate(templateName)
		if err != nil {
			fmt.Println(utils.FormatError("Template not found: " + templateName))
			fmt.Println(utils.FormatTip("Use 'nixai templates list' to see available templates"))
			os.Exit(1)
		}

		// Display template details
		fmt.Println(utils.FormatHeader("📄 Template: " + template.Name))
		fmt.Println()
		fmt.Println(utils.FormatKeyValue("Name", template.Name))
		fmt.Println(utils.FormatKeyValue("Description", template.Description))
		fmt.Println(utils.FormatKeyValue("Category", template.Category))
		fmt.Println(utils.FormatKeyValue("Source", template.Source))

		if len(template.Tags) > 0 {
			fmt.Println(utils.FormatKeyValue("Tags", strings.Join(template.Tags, ", ")))
		}

		if template.GitHubRepo != "" {
			fmt.Println(utils.FormatKeyValue("GitHub Repo", template.GitHubRepo))
		}

		if template.FilePath != "" {
			fmt.Println(utils.FormatKeyValue("File Path", template.FilePath))
		}

		fmt.Println()
		fmt.Println(utils.FormatSection("Template Content", ""))
		fmt.Println(utils.FormatCodeBlock(template.Content, "nix"))

		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai templates apply " + templateName + "' to apply this template"))
	},
}

// Templates apply command
var templatesApplyCmd = &cobra.Command{
	Use:   "apply <name>",
	Short: "Apply template to current configuration",
	Long: `Apply a template to the current NixOS configuration. This will either create
a new configuration file or help merge the template with your existing configuration.

Examples:
  nixai templates apply desktop-minimal
  nixai templates apply gaming-config --merge
  nixai templates apply server-basic --output /etc/nixos/server.nix`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		merge, _ := cmd.Flags().GetBool("merge")
		output, _ := cmd.Flags().GetString("output")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Find template
		template, err := tm.GetTemplate(templateName)
		if err != nil {
			fmt.Println(utils.FormatError("Template not found: " + templateName))
			os.Exit(1)
		}

		fmt.Println(utils.FormatHeader("🔧 Applying Template: " + template.Name))
		fmt.Println()

		// Apply template
		err = tm.ApplyTemplate(template, output, merge)
		if err != nil {
			fmt.Println(utils.FormatError("Error applying template: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Template applied successfully!"))
		if output != "" {
			fmt.Println(utils.FormatKeyValue("Output file", output))
		}
		fmt.Println()
		fmt.Println(utils.FormatTip("Review the generated configuration before rebuilding"))
		fmt.Println(utils.FormatTip("Use 'sudo nixos-rebuild switch' to apply changes"))
	},
}

// Templates save command
var templatesSaveCmd = &cobra.Command{
	Use:   "save <name> <source>",
	Short: "Save configuration as template",
	Long: `Save a NixOS configuration file or URL as a reusable template.

The source can be:
- Local file path (e.g., /etc/nixos/configuration.nix)
- GitHub URL (e.g., https://github.com/user/repo/blob/main/configuration.nix)
- GitHub raw URL

Examples:
  nixai templates save my-config /etc/nixos/configuration.nix
  nixai templates save gaming-setup https://github.com/user/nixos-configs/blob/main/gaming.nix
  nixai templates save server-config ./server-configuration.nix --category Server`,
	Args: conditionalExactArgsValidator(2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		source := args[1]
		category, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		fmt.Println(utils.FormatHeader("💾 Saving Template: " + templateName))
		fmt.Println()

		// Save template
		err = tm.SaveTemplate(templateName, source, category, description, tags)
		if err != nil {
			fmt.Println(utils.FormatError("Error saving template: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Template saved successfully!"))
		fmt.Println(utils.FormatKeyValue("Name", templateName))
		fmt.Println(utils.FormatKeyValue("Source", source))
		if category != "" {
			fmt.Println(utils.FormatKeyValue("Category", category))
		}
		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai templates show " + templateName + "' to view the saved template"))
	},
}

// Templates categories command
var templatesCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List template categories",
	Long:  "Show all available template categories with counts and descriptions.",
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		fmt.Println(utils.FormatHeader("📚 Template Categories"))
		fmt.Println()

		categories := tm.GetCategories()
		if len(categories) == 0 {
			fmt.Println(utils.FormatWarning("No categories found"))
			return
		}

		for category, count := range categories {
			description := getCategoryDescription(category)
			fmt.Printf("  %s (%d template%s)\n",
				utils.FormatKeyValue(category, description),
				count,
				func() string {
					if count == 1 {
						return ""
					} else {
						return "s"
					}
				}())
		}

		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai templates search <category>' to find templates in a specific category"))
	},
}

// Snippets command
var snippetsCmd = &cobra.Command{
	Use:   "snippets",
	Short: "Manage NixOS configuration snippets",
	Long: `Save, organize, and reuse NixOS configuration snippets.

Commands:
  list                    - List saved snippets
  search <query>          - Search snippets by name or tag
  add <name>              - Save current config as snippet
  apply <name>            - Apply snippet to configuration
  remove <name>           - Remove saved snippet
  show <name>             - Show snippet content

Examples:
  nixai snippets list
  nixai snippets search nvidia
  nixai snippets add my-nvidia-config
  nixai snippets apply gaming-setup`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Snippet subcommands
var snippetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved snippets",
	Long:  "Display all saved configuration snippets organized by category.",
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		fmt.Println(utils.FormatHeader("📝 Saved Snippets"))
		fmt.Println()

		snippets, err := tm.LoadSnippets()
		if err != nil {
			fmt.Println(utils.FormatError("Error loading snippets: " + err.Error()))
			os.Exit(1)
		}

		if len(snippets) == 0 {
			fmt.Println(utils.FormatWarning("No snippets saved"))
			fmt.Println(utils.FormatTip("Use 'nixai snippets add <name>' to save configuration snippets"))
			return
		}

		// Group by tags or display chronologically
		for i, snippet := range snippets {
			fmt.Printf("%s. %s\n",
				utils.FormatNote(fmt.Sprintf("%d", i+1)),
				utils.FormatKeyValue(snippet.Name, snippet.Description))

			if len(snippet.Tags) > 0 {
				fmt.Printf("   %s\n", utils.FormatNote("Tags: "+strings.Join(snippet.Tags, ", ")))
			}

			fmt.Printf("   %s\n", utils.FormatNote("Created: "+snippet.CreatedAt.Format("2006-01-02 15:04")))
			fmt.Println()
		}

		fmt.Println(utils.FormatTip("Use 'nixai snippets show <name>' to view snippet content"))
		fmt.Println(utils.FormatTip("Use 'nixai snippets apply <name>' to apply a snippet"))
	},
}

var snippetsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search snippets by name or tag",
	Long:  "Search saved configuration snippets by name, description, or tags.",
	Args:  conditionalArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		fmt.Println(utils.FormatHeader("🔍 Searching Snippets: " + query))
		fmt.Println()

		snippets, err := tm.SearchSnippets(query)
		if err != nil {
			fmt.Println(utils.FormatError("Error searching snippets: " + err.Error()))
			os.Exit(1)
		}

		if len(snippets) == 0 {
			fmt.Println(utils.FormatWarning("No snippets found matching: " + query))
			return
		}

		fmt.Printf("Found %d snippet(s):\n\n", len(snippets))
		for i, snippet := range snippets {
			fmt.Printf("%s. %s\n",
				utils.FormatNote(fmt.Sprintf("%d", i+1)),
				utils.FormatKeyValue(snippet.Name, snippet.Description))

			if len(snippet.Tags) > 0 {
				fmt.Printf("   %s\n", utils.FormatNote("Tags: "+strings.Join(snippet.Tags, ", ")))
			}
			fmt.Println()
		}
	},
}

var snippetsAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Save configuration as snippet",
	Long: `Save a configuration file or text as a reusable snippet.

Examples:
  nixai snippets add my-nvidia-config --file /etc/nixos/nvidia.nix
  nixai snippets add gaming-setup --file ./gaming.nix --tags gaming,performance
  echo "services.nginx.enable = true;" | nixai snippets add nginx-basic`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		snippetName := args[0]
		file, _ := cmd.Flags().GetString("file")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		fmt.Println(utils.FormatHeader("💾 Saving Snippet: " + snippetName))
		fmt.Println()

		// Save snippet
		err = tm.SaveSnippet(snippetName, file, description, tags)
		if err != nil {
			fmt.Println(utils.FormatError("Error saving snippet: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Snippet saved successfully!"))
		fmt.Println(utils.FormatKeyValue("Name", snippetName))
		if description != "" {
			fmt.Println(utils.FormatKeyValue("Description", description))
		}
		if len(tags) > 0 {
			fmt.Println(utils.FormatKeyValue("Tags", strings.Join(tags, ", ")))
		}
	},
}

var snippetsApplyCmd = &cobra.Command{
	Use:   "apply <name>",
	Short: "Apply snippet to configuration",
	Long:  "Apply a saved snippet to the current configuration or output it to a file.",
	Args:  conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		snippetName := args[0]
		output, _ := cmd.Flags().GetString("output")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Apply snippet
		err = tm.ApplySnippet(snippetName, output)
		if err != nil {
			fmt.Println(utils.FormatError("Error applying snippet: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Snippet applied successfully!"))
		if output != "" {
			fmt.Println(utils.FormatKeyValue("Output", output))
		}
	},
}

var snippetsShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show snippet content",
	Long:  "Display the content and metadata of a saved snippet.",
	Args:  conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		snippetName := args[0]

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Get snippet
		snippet, err := tm.GetSnippet(snippetName)
		if err != nil {
			fmt.Println(utils.FormatError("Snippet not found: " + snippetName))
			os.Exit(1)
		}

		// Display snippet
		fmt.Println(utils.FormatHeader("📄 Snippet: " + snippet.Name))
		fmt.Println()
		fmt.Println(utils.FormatKeyValue("Name", snippet.Name))
		fmt.Println(utils.FormatKeyValue("Description", snippet.Description))
		fmt.Println(utils.FormatKeyValue("Created", snippet.CreatedAt.Format("2006-01-02 15:04:05")))
		fmt.Println(utils.FormatKeyValue("Source", snippet.Source))

		if len(snippet.Tags) > 0 {
			fmt.Println(utils.FormatKeyValue("Tags", strings.Join(snippet.Tags, ", ")))
		}

		fmt.Println()
		fmt.Println(utils.FormatSection("Content", ""))
		fmt.Println(utils.FormatCodeBlock(snippet.Content, "nix"))
	},
}

var snippetsRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove saved snippet",
	Long:  "Delete a saved configuration snippet.",
	Args:  conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		snippetName := args[0]

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create template manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		tm := NewTemplateManager("", log)

		// Remove snippet
		err = tm.RemoveSnippet(snippetName)
		if err != nil {
			fmt.Println(utils.FormatError("Error removing snippet: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Snippet removed successfully!"))
		fmt.Println(utils.FormatKeyValue("Removed", snippetName))
	},
}

// Load builtin templates
func (tm *TemplateManager) LoadBuiltinTemplates() []Template {
	// For now, return some example builtin templates
	// In a full implementation, these would be loaded from embedded files or config
	return []Template{
		{
			Name:        "desktop-minimal",
			Description: "Minimal desktop environment with essential applications",
			Category:    "Desktop",
			Tags:        []string{"desktop", "minimal", "gnome"},
			Source:      "builtin",
			Content:     getMinimalDesktopTemplate(),
		},
		{
			Name:        "gaming-config",
			Description: "Gaming-optimized NixOS configuration with Steam and drivers",
			Category:    "Gaming",
			Tags:        []string{"gaming", "steam", "nvidia", "performance"},
			Source:      "builtin",
			Content:     getGamingTemplate(),
		},
		{
			Name:        "server-basic",
			Description: "Basic server configuration with SSH and firewall",
			Category:    "Server",
			Tags:        []string{"server", "ssh", "firewall", "minimal"},
			Source:      "builtin",
			Content:     getServerTemplate(),
		},
		{
			Name:        "development-env",
			Description: "Development environment with common programming tools",
			Category:    "Development",
			Tags:        []string{"development", "programming", "tools", "git"},
			Source:      "builtin",
			Content:     getDevelopmentTemplate(),
		},
	}
}

// Search templates by query
func (tm *TemplateManager) SearchTemplates(query string) []Template {
	templates := tm.LoadBuiltinTemplates()
	var matches []Template

	query = strings.ToLower(query)

	for _, template := range templates {
		// Search in name, description, category, and tags
		if strings.Contains(strings.ToLower(template.Name), query) ||
			strings.Contains(strings.ToLower(template.Description), query) ||
			strings.Contains(strings.ToLower(template.Category), query) {
			matches = append(matches, template)
			continue
		}

		// Search in tags
		for _, tag := range template.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, template)
				break
			}
		}
	}

	return matches
}

// Search GitHub for NixOS configurations
func (tm *TemplateManager) SearchGitHub(query string) (*GitHubSearchResponse, error) {
	// Enhance the query to focus on NixOS configurations
	enhancedQuery := fmt.Sprintf("%s configuration.nix OR flake.nix language:nix", query)

	// URL encode the query
	encodedQuery := url.QueryEscape(enhancedQuery)

	// GitHub API URL
	apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s&sort=stars&order=desc&per_page=20", encodedQuery)

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixai/1.0")

	// Check for GitHub token
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResponse GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}

	return &searchResponse, nil
}

// Template content generators (these would be improved with actual curated content)
func getMinimalDesktopTemplate() string {
	return `# Minimal Desktop Configuration
{ config, pkgs, ... }:

{
  # Enable the X11 windowing system
  services.xserver.enable = true;
  
  # Enable GNOME Desktop Environment
  services.xserver.displayManager.gdm.enable = true;
  services.xserver.desktopManager.gnome.enable = true;
  
  # Enable sound
  sound.enable = true;
  hardware.pulseaudio.enable = true;
  
  # Enable NetworkManager
  networking.networkmanager.enable = true;
  
  # Define user account
  users.users.USERNAME = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" ];
  };
  
  # Essential packages
  environment.systemPackages = with pkgs; [
    firefox
    gnome.gnome-terminal
    gnome.nautilus
    git
    vim
  ];
  
  # Enable automatic login (optional)
  # services.xserver.displayManager.autoLogin.enable = true;
  # services.xserver.displayManager.autoLogin.user = "USERNAME";
}`
}

func getGamingTemplate() string {
	return `# Gaming Configuration
{ config, pkgs, ... }:

{
  # Enable Steam and gaming packages
  programs.steam = {
    enable = true;
    remotePlay.openFirewall = true;
    dedicatedServer.openFirewall = true;
  };
  
  # Enable 32-bit libraries for gaming
  hardware.opengl.driSupport32Bit = true;
  hardware.pulseaudio.support32Bit = true;
  
  # NVIDIA drivers (if applicable)
  services.xserver.videoDrivers = [ "nvidia" ];
  hardware.nvidia = {
    modesetting.enable = true;
    powerManagement.enable = false;
    powerManagement.finegrained = false;
    open = false;
    nvidiaSettings = true;
    package = config.boot.kernelPackages.nvidiaPackages.stable;
  };
  
  # Gaming packages
  environment.systemPackages = with pkgs; [
    steam
    lutris
    wine
    winetricks
    gamemode
    mangohud
    discord
  ];
  
  # Performance optimizations
  programs.gamemode.enable = true;
  
  # Enable Xbox controller support
  hardware.xone.enable = true;
}`
}

func getServerTemplate() string {
	return `# Basic Server Configuration  
{ config, pkgs, ... }:

{
  # Enable SSH
  services.openssh = {
    enable = true;
    settings = {
      PasswordAuthentication = false;
      KbdInteractiveAuthentication = false;
      PermitRootLogin = "no";
    };
  };
  
  # Firewall configuration
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 ]; # SSH
  };
  
  # Automatic security updates
  system.autoUpgrade = {
    enable = true;
    allowReboot = false;
  };
  
  # Essential server packages
  environment.systemPackages = with pkgs; [
    htop
    git
    curl
    wget
    vim
    tmux
  ];
  
  # User configuration
  users.users.admin = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    openssh.authorizedKeys.keys = [
      # Add your SSH public key here
      # "ssh-rsa AAAAB3Nz... your-key-here"
    ];
  };
  
  # Disable unnecessary services
  services.xserver.enable = false;
  sound.enable = false;
}`
}

func getDevelopmentTemplate() string {
	return `# Development Environment Configuration
{ config, pkgs, ... }:

{
  # Development packages
  environment.systemPackages = with pkgs; [
    # Version control
    git
    gh
    
    # Editors
    vim
    neovim
    vscode
    
    # Languages
    nodejs
    python3
    rustc
    cargo
    go
    
    # Tools
    docker
    docker-compose
    kubernetes
    minikube
    
    # Database tools
    postgresql
    redis
    
    # Network tools
    curl
    wget
    jq
    
    # Build tools
    gnumake
    gcc
    pkg-config
  ];
  
  # Enable Docker
  virtualisation.docker.enable = true;
  
  # Enable development services
  services.postgresql = {
    enable = true;
    package = pkgs.postgresql_14;
  };
  
  services.redis.servers."" = {
    enable = true;
    port = 6379;
  };
  
  # User configuration for development
  users.users.dev = {
    isNormalUser = true;
    extraGroups = [ "wheel" "docker" ];
  };
  
  # Enable Nix development features
  nix.settings.experimental-features = [ "nix-command" "flakes" ];
}`
}

// Add commands to CLI in init function
func init() {
	// Add template and snippet subcommands
	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesSearchCmd)
	templatesCmd.AddCommand(templatesGithubCmd)
	templatesCmd.AddCommand(templatesShowCmd)
	templatesCmd.AddCommand(templatesApplyCmd)
	templatesCmd.AddCommand(templatesSaveCmd)
	templatesCmd.AddCommand(templatesCategoriesCmd)

	snippetsCmd.AddCommand(snippetsListCmd)
	snippetsCmd.AddCommand(snippetsSearchCmd)
	snippetsCmd.AddCommand(snippetsAddCmd)
	snippetsCmd.AddCommand(snippetsApplyCmd)
	snippetsCmd.AddCommand(snippetsShowCmd)
	snippetsCmd.AddCommand(snippetsRemoveCmd)

	// Add flags to GitHub search command
	templatesGithubCmd.Flags().IntP("limit", "l", 10, "Maximum number of results to show")
	templatesGithubCmd.Flags().StringP("language", "", "nix", "Programming language to filter by")
	templatesGithubCmd.Flags().StringP("sort", "s", "stars", "Sort results by: stars, updated, created")

	// Add flags to apply command
	templatesApplyCmd.Flags().BoolP("merge", "m", false, "Merge template with existing configuration")
	templatesApplyCmd.Flags().StringP("output", "o", "", "Output file for applied template")

	// Add flags to save command
	templatesSaveCmd.Flags().StringP("category", "c", "", "Category for the template")
	templatesSaveCmd.Flags().StringP("description", "d", "", "Description for the template")
	templatesSaveCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags for the template")

	// Add flags to add snippet command
	snippetsAddCmd.Flags().StringP("file", "f", "", "File path for the snippet")
	snippetsAddCmd.Flags().StringP("description", "d", "", "Description for the snippet")
	snippetsAddCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags for the snippet")

	// Add flags to apply snippet command
	snippetsApplyCmd.Flags().StringP("output", "o", "", "Output file for applied snippet")
}
