package learning

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// LearningFunction provides NixOS learning resources and tutorials
type LearningFunction struct {
	*functionbase.BaseFunction
	logger *logger.Logger
}

// LearningRequest represents the input parameters for learning function
type LearningRequest struct {
	Topic       string `json:"topic"`
	Level       string `json:"level,omitempty"`       // beginner, intermediate, advanced
	Format      string `json:"format,omitempty"`      // tutorial, guide, reference, example
	Interactive bool   `json:"interactive,omitempty"` // whether to provide interactive learning
	Language    string `json:"language,omitempty"`    // preferred language for resources
}

// LearningResponse represents the output of the learning function
type LearningResponse struct {
	Topic         string             `json:"topic"`
	Level         string             `json:"level"`
	Resources     []LearningResource `json:"resources"`
	NextSteps     []string           `json:"next_steps,omitempty"`
	Prerequisites []string           `json:"prerequisites,omitempty"`
	EstimatedTime string             `json:"estimated_time,omitempty"`
	Interactive   bool               `json:"interactive"`
}

// LearningResource represents a single learning resource
type LearningResource struct {
	Title       string   `json:"title"`
	Type        string   `json:"type"` // tutorial, guide, documentation, video, example
	URL         string   `json:"url,omitempty"`
	Description string   `json:"description"`
	Difficulty  string   `json:"difficulty"` // beginner, intermediate, advanced
	Duration    string   `json:"duration,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Content     string   `json:"content,omitempty"` // inline content for tutorials
}

// NewLearningFunction creates a new learning function
func NewLearningFunction() *LearningFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("topic", "The NixOS topic to learn about (e.g., 'flakes', 'configuration', 'packages')", true),
		functionbase.StringParam("level", "Learning level: beginner, intermediate, or advanced", false),
		functionbase.StringParam("format", "Preferred format: tutorial, guide, reference, or example", false),
		functionbase.BoolParam("interactive", "Whether to provide interactive learning experience", false, false),
		functionbase.StringParam("language", "Preferred language for resources", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"learning",
		"Provide NixOS learning resources, tutorials, and educational content",
		parameters,
	)

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Learn about NixOS flakes for beginners",
			Parameters: map[string]interface{}{
				"topic":       "flakes",
				"level":       "beginner",
				"format":      "tutorial",
				"interactive": true,
			},
			Expected: "Beginner-friendly flakes tutorial with step-by-step guidance",
		},
		{
			Description: "Advanced NixOS configuration patterns",
			Parameters: map[string]interface{}{
				"topic":  "configuration",
				"level":  "advanced",
				"format": "guide",
			},
			Expected: "Advanced configuration patterns and best practices",
		},
	}
	baseFunc.SetSchema(schema)

	return &LearningFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
	}
}

// ValidateParameters validates the function parameters with custom checks
func (lf *LearningFunction) ValidateParameters(params map[string]interface{}) error {
	// First run base validation
	if err := lf.BaseFunction.ValidateParameters(params); err != nil {
		return err
	}

	// Custom validation for topic parameter
	if topic, ok := params["topic"].(string); ok {
		if strings.TrimSpace(topic) == "" {
			return fmt.Errorf("topic parameter cannot be empty")
		}
	}

	// Validate level if provided
	if level, ok := params["level"].(string); ok && level != "" {
		validLevels := []string{"beginner", "intermediate", "advanced"}
		if !contains(validLevels, strings.ToLower(level)) {
			return fmt.Errorf("level must be one of: %s", strings.Join(validLevels, ", "))
		}
	}

	// Validate format if provided
	if format, ok := params["format"].(string); ok && format != "" {
		validFormats := []string{"tutorial", "guide", "reference", "example"}
		if !contains(validFormats, strings.ToLower(format)) {
			return fmt.Errorf("format must be one of: %s", strings.Join(validFormats, ", "))
		}
	}

	return nil
}

// Execute runs the learning function
func (lf *LearningFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	lf.logger.Debug("Starting learning function execution")

	// Parse parameters into structured request
	request, err := lf.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate that we have a topic
	if request.Topic == "" {
		return functionbase.CreateErrorResult(
			fmt.Errorf("topic parameter is required and cannot be empty"),
			"Missing required parameter",
		), nil
	}

	// Build the response
	response := &LearningResponse{
		Topic:       request.Topic,
		Level:       request.Level,
		Interactive: request.Interactive,
		Resources:   lf.generateLearningResources(request),
	}

	// Add learning path information
	response.NextSteps = lf.generateNextSteps(request.Topic, request.Level)
	response.Prerequisites = lf.generatePrerequisites(request.Topic, request.Level)
	response.EstimatedTime = lf.estimateLearningTime(request.Topic, request.Level)

	lf.logger.Debug("Learning function execution completed successfully")

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
	}, nil
}

// parseRequest converts raw parameters to structured LearningRequest
func (lf *LearningFunction) parseRequest(params map[string]interface{}) (*LearningRequest, error) {
	request := &LearningRequest{}

	// Extract topic (required)
	if topic, ok := params["topic"].(string); ok {
		request.Topic = strings.TrimSpace(topic)
	}

	// Extract level (optional)
	if level, ok := params["level"].(string); ok {
		request.Level = strings.ToLower(strings.TrimSpace(level))
	} else {
		request.Level = "beginner" // default
	}

	// Extract format (optional)
	if format, ok := params["format"].(string); ok {
		request.Format = strings.ToLower(strings.TrimSpace(format))
	} else {
		request.Format = "tutorial" // default
	}

	// Extract interactive (optional, default false)
	if interactive, ok := params["interactive"].(bool); ok {
		request.Interactive = interactive
	}

	// Extract language (optional)
	if language, ok := params["language"].(string); ok {
		request.Language = strings.TrimSpace(language)
	} else {
		request.Language = "english" // default
	}

	return request, nil
}

// generateLearningResources creates learning resources based on the request
func (lf *LearningFunction) generateLearningResources(request *LearningRequest) []LearningResource {
	var resources []LearningResource

	topic := strings.ToLower(request.Topic)
	level := request.Level
	format := request.Format

	// Generate resources based on topic
	switch {
	case strings.Contains(topic, "flake"):
		resources = append(resources, lf.getFlakeResources(level, format)...)
	case strings.Contains(topic, "config"):
		resources = append(resources, lf.getConfigurationResources(level, format)...)
	case strings.Contains(topic, "package"):
		resources = append(resources, lf.getPackageResources(level, format)...)
	case strings.Contains(topic, "home-manager") || strings.Contains(topic, "home manager"):
		resources = append(resources, lf.getHomeManagerResources(level, format)...)
	case strings.Contains(topic, "derivation"):
		resources = append(resources, lf.getDerivationResources(level, format)...)
	case strings.Contains(topic, "nix language") || strings.Contains(topic, "nix-lang"):
		resources = append(resources, lf.getNixLanguageResources(level, format)...)
	default:
		resources = append(resources, lf.getGeneralResources(topic, level, format)...)
	}

	// Add official documentation
	resources = append(resources, LearningResource{
		Title:       "Official NixOS Manual",
		Type:        "documentation",
		URL:         "https://nixos.org/manual/nixos/stable/",
		Description: "Comprehensive official documentation for NixOS",
		Difficulty:  "intermediate",
		Tags:        []string{"official", "manual", "comprehensive"},
	})

	return resources
}

// getFlakeResources returns flake-specific learning resources
func (lf *LearningFunction) getFlakeResources(level, format string) []LearningResource {
	resources := []LearningResource{
		{
			Title:       "Introduction to Nix Flakes",
			Type:        "tutorial",
			URL:         "https://nixos.wiki/wiki/Flakes",
			Description: "Comprehensive introduction to Nix flakes concept and usage",
			Difficulty:  "beginner",
			Duration:    "30 minutes",
			Tags:        []string{"flakes", "basics", "introduction"},
		},
	}

	if level == "beginner" {
		resources = append(resources, LearningResource{
			Title:       "Your First Flake",
			Type:        "tutorial",
			Description: "Step-by-step guide to creating your first Nix flake",
			Difficulty:  "beginner",
			Duration:    "45 minutes",
			Tags:        []string{"flakes", "hands-on", "first-steps"},
			Content: `# Your First Nix Flake

## Step 1: Create a flake.nix file
Create a new directory and add a flake.nix file:

` + "```nix" + `
{
  description = "My first flake";
  
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  
  outputs = { self, nixpkgs }: {
    # Your outputs here
  };
}
` + "```" + `

## Step 2: Initialize the flake
Run: nix flake init

## Step 3: Add outputs
Learn to add packages, devShells, and more...`,
		})
	}

	if level == "advanced" {
		resources = append(resources, LearningResource{
			Title:       "Advanced Flake Patterns",
			Type:        "guide",
			Description: "Advanced patterns for structuring and organizing flakes",
			Difficulty:  "advanced",
			Duration:    "2 hours",
			Tags:        []string{"flakes", "patterns", "advanced", "architecture"},
		})
	}

	return resources
}

// getConfigurationResources returns configuration-specific learning resources
func (lf *LearningFunction) getConfigurationResources(level, format string) []LearningResource {
	resources := []LearningResource{
		{
			Title:       "NixOS Configuration Basics",
			Type:        "tutorial",
			URL:         "https://nixos.org/manual/nixos/stable/#sec-configuration-syntax",
			Description: "Learn the basics of NixOS configuration syntax and structure",
			Difficulty:  "beginner",
			Duration:    "1 hour",
			Tags:        []string{"configuration", "basics", "syntax"},
		},
	}

	if format == "example" || level == "beginner" {
		resources = append(resources, LearningResource{
			Title:       "Common Configuration Examples",
			Type:        "example",
			Description: "Real-world NixOS configuration examples",
			Difficulty:  level,
			Tags:        []string{"configuration", "examples", "practical"},
			Content: `# Common NixOS Configuration Examples

## Enable SSH
` + "```nix" + `
services.openssh = {
  enable = true;
  settings.PasswordAuthentication = false;
};
` + "```" + `

## Install packages
` + "```nix" + `
environment.systemPackages = with pkgs; [
  git
  vim
  firefox
];
` + "```" + `

## Configure users
` + "```nix" + `
users.users.myuser = {
  isNormalUser = true;
  extraGroups = [ "wheel" "networkmanager" ];
};
` + "```",
		})
	}

	return resources
}

// getPackageResources returns package management learning resources
func (lf *LearningFunction) getPackageResources(level, format string) []LearningResource {
	return []LearningResource{
		{
			Title:       "Package Management in NixOS",
			Type:        "guide",
			URL:         "https://nixos.org/manual/nixos/stable/#sec-package-management",
			Description: "Learn how to install, update, and manage packages in NixOS",
			Difficulty:  "beginner",
			Duration:    "45 minutes",
			Tags:        []string{"packages", "management", "nix-env"},
		},
		{
			Title:       "Creating Custom Packages",
			Type:        "tutorial",
			Description: "Learn to create and package your own software for Nix",
			Difficulty:  "intermediate",
			Duration:    "2 hours",
			Tags:        []string{"packages", "derivations", "custom"},
		},
	}
}

// getHomeManagerResources returns Home Manager learning resources
func (lf *LearningFunction) getHomeManagerResources(level, format string) []LearningResource {
	return []LearningResource{
		{
			Title:       "Home Manager Introduction",
			Type:        "tutorial",
			URL:         "https://nix-community.github.io/home-manager/",
			Description: "Getting started with Home Manager for user environment management",
			Difficulty:  "beginner",
			Duration:    "1 hour",
			Tags:        []string{"home-manager", "user-environment", "dotfiles"},
		},
		{
			Title:       "Home Manager Configuration Examples",
			Type:        "example",
			Description: "Common Home Manager configuration patterns",
			Difficulty:  level,
			Tags:        []string{"home-manager", "examples", "configuration"},
		},
	}
}

// getDerivationResources returns derivation learning resources
func (lf *LearningFunction) getDerivationResources(level, format string) []LearningResource {
	return []LearningResource{
		{
			Title:       "Understanding Nix Derivations",
			Type:        "guide",
			URL:         "https://nixos.org/manual/nix/stable/#ssec-derivation",
			Description: "Deep dive into Nix derivations and how they work",
			Difficulty:  "intermediate",
			Duration:    "1.5 hours",
			Tags:        []string{"derivations", "nix-theory", "internals"},
		},
	}
}

// getNixLanguageResources returns Nix language learning resources
func (lf *LearningFunction) getNixLanguageResources(level, format string) []LearningResource {
	return []LearningResource{
		{
			Title:       "Nix Language Basics",
			Type:        "tutorial",
			URL:         "https://nixos.org/manual/nix/stable/#sec-language-syntax",
			Description: "Learn the Nix expression language syntax and concepts",
			Difficulty:  "beginner",
			Duration:    "2 hours",
			Tags:        []string{"nix-language", "syntax", "expressions"},
		},
		{
			Title:       "Advanced Nix Language Features",
			Type:        "guide",
			Description: "Functions, sets, lists, and advanced language constructs",
			Difficulty:  "advanced",
			Duration:    "3 hours",
			Tags:        []string{"nix-language", "advanced", "functions"},
		},
	}
}

// getGeneralResources returns general learning resources for any topic
func (lf *LearningFunction) getGeneralResources(topic, level, format string) []LearningResource {
	return []LearningResource{
		{
			Title:       fmt.Sprintf("NixOS %s Guide", strings.Title(topic)),
			Type:        format,
			Description: fmt.Sprintf("Learn about %s in NixOS", topic),
			Difficulty:  level,
			Tags:        []string{topic, "nixos", level},
		},
		{
			Title:       "NixOS Wiki",
			Type:        "reference",
			URL:         "https://nixos.wiki/",
			Description: "Community-maintained wiki with extensive NixOS information",
			Difficulty:  "intermediate",
			Tags:        []string{"wiki", "community", "reference"},
		},
	}
}

// generateNextSteps creates suggested next learning steps
func (lf *LearningFunction) generateNextSteps(topic, level string) []string {
	var nextSteps []string

	switch level {
	case "beginner":
		nextSteps = []string{
			"Practice basic NixOS configuration",
			"Learn about the Nix package manager",
			"Explore the NixOS options reference",
			"Set up a simple development environment",
		}
	case "intermediate":
		nextSteps = []string{
			"Learn about Nix flakes",
			"Create custom packages",
			"Explore Home Manager",
			"Study advanced configuration patterns",
		}
	case "advanced":
		nextSteps = []string{
			"Contribute to nixpkgs",
			"Create NixOS modules",
			"Build custom NixOS systems",
			"Learn Nix internals and theory",
		}
	}

	// Add topic-specific next steps
	switch strings.ToLower(topic) {
	case "flakes":
		nextSteps = append(nextSteps, "Convert existing configuration to flakes")
	case "packages":
		nextSteps = append(nextSteps, "Package your own software")
	case "configuration":
		nextSteps = append(nextSteps, "Modularize your configuration")
	}

	return nextSteps
}

// generatePrerequisites creates a list of prerequisites for learning
func (lf *LearningFunction) generatePrerequisites(topic, level string) []string {
	var prereqs []string

	switch level {
	case "beginner":
		prereqs = []string{
			"Basic Linux command line knowledge",
			"Text editor familiarity",
		}
	case "intermediate":
		prereqs = []string{
			"NixOS installation and basic usage",
			"Understanding of configuration.nix",
			"Basic Nix expression language",
		}
	case "advanced":
		prereqs = []string{
			"Solid understanding of Nix language",
			"Experience with NixOS configuration",
			"Knowledge of software packaging concepts",
		}
	}

	// Add topic-specific prerequisites
	switch strings.ToLower(topic) {
	case "flakes":
		if level != "beginner" {
			prereqs = append(prereqs, "Understanding of traditional Nix workflows")
		}
	case "derivations":
		prereqs = append(prereqs, "Nix language proficiency")
	}

	return prereqs
}

// estimateLearningTime estimates how long it will take to learn the topic
func (lf *LearningFunction) estimateLearningTime(topic, level string) string {
	timeMap := map[string]map[string]string{
		"beginner": {
			"flakes":        "2-3 hours",
			"configuration": "3-4 hours",
			"packages":      "2-3 hours",
			"default":       "2-4 hours",
		},
		"intermediate": {
			"flakes":        "4-6 hours",
			"configuration": "6-8 hours",
			"packages":      "8-10 hours",
			"default":       "4-8 hours",
		},
		"advanced": {
			"flakes":        "8-12 hours",
			"configuration": "12-16 hours",
			"packages":      "16-20 hours",
			"default":       "8-16 hours",
		},
	}

	if levelMap, ok := timeMap[level]; ok {
		if time, ok := levelMap[strings.ToLower(topic)]; ok {
			return time
		}
		return levelMap["default"]
	}

	return "2-4 hours"
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
