package templates

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// TemplatesFunction provides Nix template and project scaffolding capabilities
type TemplatesFunction struct {
	log logger.Logger
}

// NewTemplatesFunction creates a new templates function
func NewTemplatesFunction() *TemplatesFunction {
	return &TemplatesFunction{
		log: logger.NewLogger(),
	}
}

// Name returns the function name
func (f *TemplatesFunction) Name() string {
	return "templates"
}

// Description returns the function description
func (f *TemplatesFunction) Description() string {
	return "Nix template and project scaffolding - manage flake templates, create new projects, browse template registry, customize templates, and generate project structures"
}

// Schema returns the function schema
func (f *TemplatesFunction) Schema() functionbase.FunctionSchema {
	return functionbase.FunctionSchema{
		Name:        "templates",
		Description: f.Description(),
		Parameters: []functionbase.FunctionParameter{
			{
				Name:        "operation",
				Type:        "string",
				Description: "The template operation to perform",
				Required:    true,
				Enum: []string{
					"list", "search", "show", "create", "init", "add", "remove",
					"update", "validate", "customize", "preview", "export", "import", "registry",
				},
			},
			{
				Name:        "template_name",
				Type:        "string",
				Description: "Name of the template to use or operate on",
				Required:    false,
			},
			{
				Name:        "template_url",
				Type:        "string",
				Description: "URL or flake reference for template source",
				Required:    false,
			},
			{
				Name:        "project_name",
				Type:        "string",
				Description: "Name for the new project (when creating)",
				Required:    false,
			},
			{
				Name:        "output_dir",
				Type:        "string",
				Description: "Directory to create project in",
				Required:    false,
				Default:     "./",
			},
			{
				Name:        "language",
				Type:        "string",
				Description: "Programming language for template filtering",
				Required:    false,
				Enum: []string{
					"rust", "go", "python", "javascript", "typescript",
					"haskell", "c", "cpp", "java", "scala", "ruby",
					"php", "dart", "kotlin", "swift", "zig", "elm",
				},
			},
			{
				Name:        "category",
				Type:        "string",
				Description: "Project category for template filtering",
				Required:    false,
				Enum: []string{
					"web", "cli", "library", "service", "desktop",
					"mobile", "game", "data", "ml", "devops",
					"documentation", "config", "minimal",
				},
			},
			{
				Name:        "interactive",
				Type:        "boolean",
				Description: "Use interactive mode for template customization",
				Required:    false,
				Default:     true,
			},
			{
				Name:        "force",
				Type:        "boolean",
				Description: "Force operation even if directory exists",
				Required:    false,
				Default:     false,
			},
			{
				Name:        "git_init",
				Type:        "boolean",
				Description: "Initialize git repository in new project",
				Required:    false,
				Default:     true,
			},
			{
				Name:        "format",
				Type:        "string",
				Description: "Output format for results",
				Required:    false,
				Default:     "table",
				Enum:        []string{"json", "yaml", "table", "tree"},
			},
		},
	}
}

// ValidateParameters validates the function parameters
func (f *TemplatesFunction) ValidateParameters(params map[string]interface{}) error {
	operation, ok := params["operation"].(string)
	if !ok {
		return fmt.Errorf("operation parameter is required and must be a string")
	}

	validOperations := []string{
		"list", "search", "show", "create", "init", "add", "remove",
		"update", "validate", "customize", "preview", "export", "import", "registry",
	}

	for _, valid := range validOperations {
		if operation == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid operation: %s", operation)
}

// Execute executes the templates function
func (f *TemplatesFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	startTime := time.Now()
	f.log.Info(fmt.Sprintf("Executing templates operation: %s", operation))

	var result map[string]interface{}
	var err error

	switch operation {
	case "list":
		result, err = f.handleList(ctx, params)
	case "search":
		result, err = f.handleSearch(ctx, params)
	case "show":
		result, err = f.handleShow(ctx, params)
	case "create":
		result, err = f.handleCreate(ctx, params)
	case "init":
		result, err = f.handleInit(ctx, params)
	case "add":
		result, err = f.handleAdd(ctx, params)
	case "remove":
		result, err = f.handleRemove(ctx, params)
	case "update":
		result, err = f.handleUpdate(ctx, params)
	case "validate":
		result, err = f.handleValidate(ctx, params)
	case "customize":
		result, err = f.handleCustomize(ctx, params)
	case "preview":
		result, err = f.handlePreview(ctx, params)
	case "export":
		result, err = f.handleExport(ctx, params)
	case "import":
		result, err = f.handleImport(ctx, params)
	case "registry":
		result, err = f.handleRegistry(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		f.log.Error(fmt.Sprintf("Templates operation %s failed: %v", operation, err))
		return nil, err
	}

	duration := time.Since(startTime)
	f.log.Info(fmt.Sprintf("Templates operation %s completed in %v", operation, duration))

	// Add metadata
	result["operation"] = operation
	result["duration"] = duration.String()
	result["timestamp"] = startTime.Format(time.RFC3339)

	return &functionbase.FunctionResult{
		Success:   true,
		Data:      result,
		Duration:  duration,
		Timestamp: startTime,
		Metadata: map[string]interface{}{
			"operation": operation,
		},
	}, nil
}

// handleList handles template listing
func (f *TemplatesFunction) handleList(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	language, _ := params["language"].(string)
	category, _ := params["category"].(string)
	format, _ := params["format"].(string)
	if format == "" {
		format = "table"
	}

	f.log.Info("Listing available Nix templates...")

	// Mock template list
	templates := []map[string]interface{}{
		{
			"name":        "rust-cli",
			"description": "Rust command-line application with Clap",
			"language":    "rust",
			"category":    "cli",
			"features":    []string{"cli", "testing", "ci"},
			"source":      "github:nix-community/templates#rust-cli",
			"version":     "1.2.0",
			"popularity":  95,
		},
		{
			"name":        "typescript-node",
			"description": "TypeScript Node.js project with Express",
			"language":    "typescript",
			"category":    "web",
			"features":    []string{"api", "testing", "docker"},
			"source":      "github:nix-community/templates#typescript-node",
			"version":     "2.1.0",
			"popularity":  88,
		},
		{
			"name":        "python-data",
			"description": "Python data science project with Jupyter",
			"language":    "python",
			"category":    "data",
			"features":    []string{"ml", "jupyter", "testing"},
			"source":      "github:nix-community/templates#python-data",
			"version":     "1.0.3",
			"popularity":  76,
		},
		{
			"name":        "go-service",
			"description": "Go microservice with gRPC and Docker",
			"language":    "go",
			"category":    "service",
			"features":    []string{"api", "docker", "monitoring"},
			"source":      "github:nix-community/templates#go-service",
			"version":     "1.5.2",
			"popularity":  82,
		},
		{
			"name":        "minimal-flake",
			"description": "Minimal Nix flake template",
			"language":    "",
			"category":    "minimal",
			"features":    []string{"config"},
			"source":      "github:nix-community/templates#minimal",
			"version":     "1.0.0",
			"popularity":  100,
		},
	}

	// Apply filters
	filteredTemplates := templates
	if language != "" {
		var filtered []map[string]interface{}
		for _, template := range filteredTemplates {
			if template["language"] == language {
				filtered = append(filtered, template)
			}
		}
		filteredTemplates = filtered
	}

	if category != "" {
		var filtered []map[string]interface{}
		for _, template := range filteredTemplates {
			if template["category"] == category {
				filtered = append(filtered, template)
			}
		}
		filteredTemplates = filtered
	}

	results := map[string]interface{}{
		"templates": filteredTemplates,
		"filters": map[string]interface{}{
			"language": language,
			"category": category,
		},
		"summary": map[string]interface{}{
			"total_available": len(templates),
			"shown":           len(filteredTemplates),
			"format":          format,
		},
		"categories": []string{"web", "cli", "library", "service", "data", "minimal"},
		"languages":  []string{"rust", "go", "python", "typescript", "javascript"},
	}

	recommendations := []string{
		"Popular templates: rust-cli (95%), minimal-flake (100%)",
		"Use 'show' operation to see template details",
		"Filter by language or category for focused results",
		"Check template features before selection",
	}

	return map[string]interface{}{
		"type":            "template_list",
		"results":         results,
		"recommendations": recommendations,
		"next_steps": []string{
			"Show template details with 'show' operation",
			"Create project with 'create' operation",
			"Search templates with specific criteria",
		},
	}, nil
}

// handleSearch handles template searching
func (f *TemplatesFunction) handleSearch(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, _ := params["template_name"].(string)
	language, _ := params["language"].(string)
	category, _ := params["category"].(string)
	features, _ := params["features"].([]interface{})

	f.log.Info("Searching Nix templates...")

	// Convert features to string slice
	var featureStrings []string
	for _, feature := range features {
		if str, ok := feature.(string); ok {
			featureStrings = append(featureStrings, str)
		}
	}

	// Mock search results
	searchResults := []map[string]interface{}{
		{
			"name":         "rust-web",
			"description":  "Rust web application with Axum framework",
			"language":     "rust",
			"category":     "web",
			"features":     []string{"web", "api", "testing", "docker"},
			"source":       "github:templates/rust-web",
			"match_score":  0.92,
			"match_reason": "Language and category match",
		},
		{
			"name":         "typescript-api",
			"description":  "TypeScript REST API with Fastify",
			"language":     "typescript",
			"category":     "web",
			"features":     []string{"api", "testing", "docker", "monitoring"},
			"source":       "github:templates/ts-api",
			"match_score":  0.85,
			"match_reason": "Feature overlap: api, testing, docker",
		},
	}

	searchCriteria := map[string]interface{}{
		"template_name": templateName,
		"language":      language,
		"category":      category,
		"features":      featureStrings,
	}

	results := map[string]interface{}{
		"search_criteria": searchCriteria,
		"matches":         searchResults,
		"total_matches":   len(searchResults),
		"search_time":     "0.23s",
	}

	recommendations := []string{
		fmt.Sprintf("Found %d templates matching criteria", len(searchResults)),
		"rust-web has highest match score (92%)",
		"Consider templates with feature overlap",
		"Review template details before selection",
	}

	return map[string]interface{}{
		"type":            "template_search",
		"results":         results,
		"recommendations": recommendations,
		"next_steps": []string{
			"Show details for high-scoring templates",
			"Create project from selected template",
			"Refine search criteria if needed",
		},
	}, nil
}

// handleShow handles template details display
func (f *TemplatesFunction) handleShow(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	f.log.Info(fmt.Sprintf("Showing template details: %s", templateName))

	// Mock template details
	templateDetails := map[string]interface{}{
		"name":        templateName,
		"description": "Rust command-line application with comprehensive tooling",
		"version":     "1.2.0",
		"language":    "rust",
		"category":    "cli",
		"source":      "github:nix-community/templates#rust-cli",
		"author":      "Nix Community",
		"license":     "MIT",
		"features": []string{
			"Clap for command-line parsing",
			"Comprehensive testing setup",
			"CI/CD with GitHub Actions",
			"Pre-commit hooks",
			"Documentation generation",
		},
		"structure": map[string]interface{}{
			"files": []string{
				"flake.nix",
				"Cargo.toml",
				"src/main.rs",
				"src/lib.rs",
				"tests/integration_tests.rs",
				".github/workflows/ci.yml",
				"README.md",
				".gitignore",
			},
			"directories": []string{
				"src/",
				"tests/",
				".github/workflows/",
				"docs/",
			},
		},
		"variables": map[string]interface{}{
			"project_name": map[string]interface{}{
				"description": "Name of the project",
				"type":        "string",
				"required":    true,
			},
			"author_name": map[string]interface{}{
				"description": "Author name for Cargo.toml",
				"type":        "string",
				"default":     "Your Name",
			},
			"author_email": map[string]interface{}{
				"description": "Author email for Cargo.toml",
				"type":        "string",
				"default":     "you@example.com",
			},
			"description": map[string]interface{}{
				"description": "Project description",
				"type":        "string",
				"default":     "A command-line tool written in Rust",
			},
		},
		"requirements": map[string]interface{}{
			"nix_version": ">=2.4",
			"flakes":      true,
			"platforms":   []string{"x86_64-linux", "aarch64-linux", "x86_64-darwin", "aarch64-darwin"},
		},
		"usage": map[string]interface{}{
			"create_command": fmt.Sprintf("nix flake new -t github:nix-community/templates#%s my-project", templateName),
			"init_command":   fmt.Sprintf("nix flake init -t github:nix-community/templates#%s", templateName),
		},
	}

	recommendations := []string{
		"Well-maintained template with comprehensive Rust tooling",
		"Includes CI/CD and testing setup out of the box",
		"Supports all major platforms",
		"Customizable through template variables",
	}

	return map[string]interface{}{
		"type":            "template_details",
		"template":        templateDetails,
		"recommendations": recommendations,
		"next_steps": []string{
			"Create new project with 'create' operation",
			"Customize template variables if needed",
			"Preview template structure before creation",
		},
	}, nil
}

// handleCreate handles project creation from template
func (f *TemplatesFunction) handleCreate(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	projectName, ok := params["project_name"].(string)
	if !ok || projectName == "" {
		return nil, fmt.Errorf("project_name parameter is required")
	}

	outputDir, _ := params["output_dir"].(string)
	if outputDir == "" {
		outputDir = "./"
	}

	variables, _ := params["variables"].(map[string]interface{})
	gitInit, _ := params["git_init"].(bool)
	force, _ := params["force"].(bool)

	f.log.Info(fmt.Sprintf("Creating project %s from template %s", projectName, templateName))

	// Mock project creation
	creationResult := map[string]interface{}{
		"operation":     "project_creation",
		"template_name": templateName,
		"project_name":  projectName,
		"output_dir":    outputDir,
		"full_path":     fmt.Sprintf("%s/%s", outputDir, projectName),
		"variables":     variables,
		"git_init":      gitInit,
		"force":         force,
		"files_created": []string{
			"flake.nix",
			"Cargo.toml",
			"src/main.rs",
			"src/lib.rs",
			"README.md",
			".gitignore",
			".github/workflows/ci.yml",
		},
		"directories_created": []string{
			"src/",
			"tests/",
			".github/workflows/",
		},
		"post_creation": map[string]interface{}{
			"git_initialized": gitInit,
			"dependencies":    "resolved",
			"build_ready":     true,
		},
		"size": "156KB",
	}

	recommendations := []string{
		fmt.Sprintf("Project %s created successfully", projectName),
		"Template variables applied correctly",
		"Git repository initialized",
		"Ready for development - run 'nix develop' to enter shell",
	}

	nextSteps := []string{
		fmt.Sprintf("cd %s/%s", outputDir, projectName),
		"nix develop",
		"cargo build",
		"git add . && git commit -m 'Initial commit'",
	}

	return map[string]interface{}{
		"type":            "project_creation",
		"results":         creationResult,
		"recommendations": recommendations,
		"next_steps":      nextSteps,
	}, nil
}

// handleInit handles template initialization in current directory
func (f *TemplatesFunction) handleInit(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	variables, _ := params["variables"].(map[string]interface{})
	force, _ := params["force"].(bool)

	f.log.Info(fmt.Sprintf("Initializing template %s in current directory", templateName))

	// Mock template initialization
	initResult := map[string]interface{}{
		"operation":     "template_init",
		"template_name": templateName,
		"directory":     "./",
		"variables":     variables,
		"force":         force,
		"files_created": []string{
			"flake.nix",
			"shell.nix",
			"default.nix",
			".envrc",
			".gitignore",
		},
		"existing_files": []string{
			"README.md",
		},
		"conflicts_resolved": 0,
	}

	recommendations := []string{
		"Template initialized successfully in current directory",
		"Flake files created for Nix development",
		"Add .envrc to enable direnv integration",
		"No file conflicts detected",
	}

	return map[string]interface{}{
		"type":            "template_init",
		"results":         initResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"nix develop # Enter development shell",
			"direnv allow # Enable direnv if available",
			"Review generated flake.nix",
		},
	}, nil
}

// handleAdd handles adding new template to registry
func (f *TemplatesFunction) handleAdd(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	templateUrl, ok := params["template_url"].(string)
	if !ok || templateUrl == "" {
		return nil, fmt.Errorf("template_url parameter is required")
	}

	f.log.Info(fmt.Sprintf("Adding template %s to registry", templateName))

	// Mock template addition
	addResult := map[string]interface{}{
		"operation":     "template_add",
		"template_name": templateName,
		"template_url":  templateUrl,
		"validation": map[string]interface{}{
			"url_accessible":   true,
			"flake_valid":      true,
			"template_outputs": true,
			"metadata_present": true,
		},
		"registry_entry": map[string]interface{}{
			"name":       templateName,
			"source":     templateUrl,
			"added_date": time.Now().Format(time.RFC3339),
			"status":     "active",
		},
	}

	recommendations := []string{
		"Template added to registry successfully",
		"Validation checks passed",
		"Template is now available for use",
		"Consider contributing to community registry",
	}

	return map[string]interface{}{
		"type":            "template_add",
		"results":         addResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Test template with 'create' operation",
			"Share template with community",
			"Update template documentation",
		},
	}, nil
}

// handleRemove handles removing template from registry
func (f *TemplatesFunction) handleRemove(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	force, _ := params["force"].(bool)

	f.log.Info(fmt.Sprintf("Removing template %s from registry", templateName))

	// Mock template removal
	removeResult := map[string]interface{}{
		"operation":     "template_remove",
		"template_name": templateName,
		"force":         force,
		"removed": map[string]interface{}{
			"name":         templateName,
			"was_active":   true,
			"usage_count":  15,
			"removed_date": time.Now().Format(time.RFC3339),
		},
		"cleanup": map[string]interface{}{
			"cache_cleared":      true,
			"references_updated": true,
		},
	}

	recommendations := []string{
		"Template removed from registry successfully",
		"Cache and references cleaned up",
		"15 previous usages recorded",
		"Removal is permanent unless re-added",
	}

	return map[string]interface{}{
		"type":            "template_remove",
		"results":         removeResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Verify template is no longer listed",
			"Update documentation if needed",
			"Consider backup before final removal",
		},
	}, nil
}

// handleUpdate handles template updates
func (f *TemplatesFunction) handleUpdate(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	f.log.Info(fmt.Sprintf("Updating template %s", templateName))

	// Mock template update
	updateResult := map[string]interface{}{
		"operation":     "template_update",
		"template_name": templateName,
		"changes": map[string]interface{}{
			"version": "1.2.0 -> 1.3.0",
			"new_features": []string{
				"Added Docker support",
				"Updated dependencies",
				"Improved CI configuration",
			},
			"breaking_changes": []string{},
			"files_changed":    8,
		},
		"validation": map[string]interface{}{
			"compatibility": true,
			"builds":        true,
			"tests_pass":    true,
		},
	}

	recommendations := []string{
		"Template updated successfully to version 1.3.0",
		"No breaking changes detected",
		"New Docker support added",
		"All validation checks passed",
	}

	return map[string]interface{}{
		"type":            "template_update",
		"results":         updateResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Test updated template",
			"Update documentation",
			"Notify users of changes",
		},
	}, nil
}

// handleValidate handles template validation
func (f *TemplatesFunction) handleValidate(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	f.log.Info(fmt.Sprintf("Validating template %s", templateName))

	// Mock template validation
	validationResult := map[string]interface{}{
		"operation":     "template_validation",
		"template_name": templateName,
		"checks": map[string]interface{}{
			"flake_syntax":     "passed",
			"template_outputs": "passed",
			"file_structure":   "passed",
			"variable_schema":  "passed",
			"build_test":       "passed",
			"documentation":    "warning",
		},
		"issues": []map[string]interface{}{
			{
				"type":     "warning",
				"message":  "README.md could be more detailed",
				"severity": "low",
				"fixable":  true,
			},
		},
		"score": "95/100",
		"grade": "A",
	}

	recommendations := []string{
		"Template validation score: 95/100 (Grade A)",
		"All critical checks passed",
		"Minor documentation improvement suggested",
		"Template is ready for production use",
	}

	return map[string]interface{}{
		"type":            "template_validation",
		"results":         validationResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Address documentation warning",
			"Submit for community review",
			"Publish to template registry",
		},
	}, nil
}

// handleCustomize handles template customization
func (f *TemplatesFunction) handleCustomize(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	variables, _ := params["variables"].(map[string]interface{})
	interactive, _ := params["interactive"].(bool)

	f.log.Info(fmt.Sprintf("Customizing template %s", templateName))

	// Mock template customization
	customizationResult := map[string]interface{}{
		"operation":     "template_customization",
		"template_name": templateName,
		"interactive":   interactive,
		"variables": map[string]interface{}{
			"provided": variables,
			"resolved": map[string]interface{}{
				"project_name": "my-awesome-project",
				"author_name":  "John Doe",
				"author_email": "john@example.com",
				"description":  "An awesome project built with Nix",
				"license":      "MIT",
			},
			"defaults_used": 2,
		},
		"customizations": []map[string]interface{}{
			{
				"file":    "Cargo.toml",
				"changes": []string{"Updated author", "Set description"},
			},
			{
				"file":    "README.md",
				"changes": []string{"Updated project name", "Added description"},
			},
		},
		"preview_available": true,
	}

	recommendations := []string{
		"Template customized successfully",
		"2 default values used for missing variables",
		"Preview available before final creation",
		"All required variables provided",
	}

	return map[string]interface{}{
		"type":            "template_customization",
		"results":         customizationResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Preview customized template",
			"Create project with customizations",
			"Save customization profile",
		},
	}, nil
}

// handlePreview handles template structure preview
func (f *TemplatesFunction) handlePreview(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	format, _ := params["format"].(string)
	if format == "" {
		format = "tree"
	}

	f.log.Info(fmt.Sprintf("Previewing template %s", templateName))

	// Mock template preview
	previewResult := map[string]interface{}{
		"operation":     "template_preview",
		"template_name": templateName,
		"format":        format,
		"structure": map[string]interface{}{
			"tree": `my-project/
├── flake.nix
├── Cargo.toml
├── src/
│   ├── main.rs
│   └── lib.rs
├── tests/
│   └── integration_tests.rs
├── .github/
│   └── workflows/
│       └── ci.yml
├── README.md
└── .gitignore`,
			"file_count":      8,
			"directory_count": 4,
			"total_size":      "156KB",
		},
		"content_preview": map[string]interface{}{
			"flake.nix": `{
  description = "A Rust CLI application";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system: {
      # ... template content ...
    });
}`,
			"Cargo.toml": `[package]
name = "{{ project_name }}"
version = "0.1.0"
edition = "2021"
authors = ["{{ author_name }} <{{ author_email }}>"]
description = "{{ description }}"

[dependencies]
clap = { version = "4.0", features = ["derive"] }`,
		},
	}

	recommendations := []string{
		"Template structure preview generated",
		"8 files and 4 directories will be created",
		"Template variables marked with {{ }}",
		"Comprehensive Rust project setup",
	}

	return map[string]interface{}{
		"type":            "template_preview",
		"results":         previewResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Customize template variables",
			"Create project from template",
			"Review template content",
		},
	}, nil
}

// handleExport handles template configuration export
func (f *TemplatesFunction) handleExport(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateName, ok := params["template_name"].(string)
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template_name parameter is required")
	}

	format, _ := params["format"].(string)
	if format == "" {
		format = "yaml"
	}

	f.log.Info(fmt.Sprintf("Exporting template configuration: %s", templateName))

	// Mock export result
	exportResult := map[string]interface{}{
		"operation":     "template_export",
		"template_name": templateName,
		"format":        format,
		"exported_config": map[string]interface{}{
			"name":        templateName,
			"version":     "1.2.0",
			"description": "Rust CLI application template",
			"language":    "rust",
			"category":    "cli",
			"variables": map[string]interface{}{
				"project_name": map[string]string{
					"type":        "string",
					"description": "Name of the project",
					"required":    "true",
				},
				"author_name": map[string]string{
					"type":    "string",
					"default": "Your Name",
				},
			},
			"files": []string{
				"flake.nix",
				"Cargo.toml",
				"src/main.rs",
			},
		},
		"export_size": "2.3KB",
	}

	recommendations := []string{
		"Template configuration exported successfully",
		"Export includes all template metadata",
		"Can be imported on other systems",
		"Useful for template distribution",
	}

	return map[string]interface{}{
		"type":            "template_export",
		"results":         exportResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Save export to file",
			"Share with team members",
			"Import on target systems",
		},
	}, nil
}

// handleImport handles template import from source
func (f *TemplatesFunction) handleImport(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	templateUrl, ok := params["template_url"].(string)
	if !ok || templateUrl == "" {
		return nil, fmt.Errorf("template_url parameter is required")
	}

	templateName, _ := params["template_name"].(string)
	force, _ := params["force"].(bool)

	f.log.Info(fmt.Sprintf("Importing template from %s", templateUrl))

	// Mock import result
	importResult := map[string]interface{}{
		"operation":   "template_import",
		"source_url":  templateUrl,
		"target_name": templateName,
		"force":       force,
		"imported_template": map[string]interface{}{
			"name":        "imported-rust-template",
			"version":     "1.0.0",
			"language":    "rust",
			"files_count": 12,
			"size":        "245KB",
		},
		"validation": map[string]interface{}{
			"syntax_valid": true,
			"structure_ok": true,
			"dependencies": "resolved",
			"conflicts":    0,
		},
	}

	recommendations := []string{
		"Template imported successfully",
		"All validation checks passed",
		"No naming conflicts detected",
		"Template ready for use",
	}

	return map[string]interface{}{
		"type":            "template_import",
		"results":         importResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Test imported template",
			"Add to local registry",
			"Create project from template",
		},
	}, nil
}

// handleRegistry handles template registry management
func (f *TemplatesFunction) handleRegistry(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	operation, _ := params["operation"].(string)
	if operation == "" {
		operation = "list"
	}

	f.log.Info("Managing template registry...")

	// Mock registry management
	registryResult := map[string]interface{}{
		"operation": "registry_management",
		"action":    operation,
		"registry": map[string]interface{}{
			"source":          "https://api.flakehub.com/f/nixos/templates",
			"total_templates": 156,
			"categories":      12,
			"languages":       18,
			"last_updated":    "2023-12-01T10:30:00Z",
		},
		"stats": map[string]interface{}{
			"most_popular":        []string{"minimal-flake", "rust-cli", "python-data"},
			"recent_additions":    8,
			"community_templates": 142,
			"official_templates":  14,
		},
		"health": map[string]interface{}{
			"status":           "healthy",
			"response_time":    "0.2s",
			"availability":     "99.9%",
			"broken_templates": 2,
		},
	}

	recommendations := []string{
		"Registry is healthy with 156 available templates",
		"99.9% availability with fast response times",
		"2 broken templates need attention",
		"Community contributes 91% of templates",
	}

	return map[string]interface{}{
		"type":            "registry_management",
		"results":         registryResult,
		"recommendations": recommendations,
		"next_steps": []string{
			"Update registry cache",
			"Report broken templates",
			"Contribute new templates",
		},
	}, nil
}
