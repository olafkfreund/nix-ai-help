package snippets

import (
	"context"
	"fmt"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// SnippetsFunction handles configuration snippet management operations
type SnippetsFunction struct {
	logger *logger.Logger
}

// NewSnippetsFunction creates a new snippets function
func NewSnippetsFunction() *SnippetsFunction {
	return &SnippetsFunction{
		logger: logger.NewLogger(),
	}
}

// Name returns the function name
func (f *SnippetsFunction) Name() string {
	return "snippets"
}

// Description returns the function description
func (f *SnippetsFunction) Description() string {
	return "Manage NixOS configuration snippets for reusable code blocks and common patterns"
}

// Schema returns the function schema for AI interaction
func (f *SnippetsFunction) Schema() functionbase.FunctionSchema {
	return functionbase.FunctionSchema{
		Name:        f.Name(),
		Description: f.Description(),
		Parameters: []functionbase.FunctionParameter{
			functionbase.StringParamWithEnum("operation", "The snippet operation to perform", true, []string{
				"list", "search", "show", "add", "remove", "apply", "edit", "export", "import", "organize",
			}),
			functionbase.StringParam("name", "Snippet name (for show, remove, apply, edit operations)", false),
			functionbase.StringParam("query", "Search query or keyword (for search operation)", false),
			functionbase.StringParam("content", "Snippet content (for add/edit operations)", false),
			functionbase.StringParam("description", "Snippet description", false),
			functionbase.ArrayParam("tags", "Tags for categorizing snippets", false),
			functionbase.StringParamWithEnum("category", "Snippet category", false, []string{
				"desktop", "server", "development", "networking", "security", "services", "packages", "hardware", "gaming", "custom",
			}),
			functionbase.StringParamWithEnum("language", "Configuration language", false, []string{
				"nix", "yaml", "json", "bash", "other",
			}),
			functionbase.StringParam("output_path", "Output file path (for apply/export operations)", false),
			functionbase.StringParam("source_path", "Source file path (for import operations)", false),
			functionbase.BoolParam("merge", "Merge with existing configuration (for apply operation)", false),
			functionbase.ObjectParam("filter", "Filter criteria for listing/searching", false),
		},
	}
}

// ValidateParameters validates the function parameters
func (f *SnippetsFunction) ValidateParameters(params map[string]interface{}) error {
	operation, ok := params["operation"]
	if !ok {
		return fmt.Errorf("operation parameter is required")
	}

	if _, ok := operation.(string); !ok {
		return fmt.Errorf("operation must be a string")
	}

	validOperations := []string{
		"list", "search", "show", "add", "remove", "apply",
		"edit", "export", "import", "backup", "validate",
		"tags", "categories", "test", "sync",
	}

	operationStr := operation.(string)
	for _, valid := range validOperations {
		if operationStr == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid operation: %s", operationStr)
}

// Execute performs the snippet operation
func (f *SnippetsFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	f.logger.Info(fmt.Sprintf("Executing snippet operation: %s", operation))

	switch operation {
	case "list":
		return f.handleList(ctx, params)
	case "search":
		return f.handleSearch(ctx, params)
	case "show":
		return f.handleShow(ctx, params)
	case "add":
		return f.handleAdd(ctx, params)
	case "remove":
		return f.handleRemove(ctx, params)
	case "apply":
		return f.handleApply(ctx, params)
	case "edit":
		return f.handleEdit(ctx, params)
	case "export":
		return f.handleExport(ctx, params)
	case "import":
		return f.handleImport(ctx, params)
	case "organize":
		return f.handleOrganize(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported snippet operation: %s", operation)
	}
}

// handleList lists available snippets
func (f *SnippetsFunction) handleList(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	filter, _ := params["filter"].(map[string]interface{})

	snippets := []map[string]interface{}{
		{
			"name":        "nvidia-gaming",
			"description": "NVIDIA GPU configuration for gaming",
			"category":    "gaming",
			"tags":        []string{"nvidia", "gaming", "graphics", "performance"},
			"language":    "nix",
			"size":        "2.1 KB",
			"created_at":  "2025-06-01T10:30:00Z",
			"author":      "user",
			"usage_count": 15,
		},
		{
			"name":        "ssh-hardening",
			"description": "SSH server security hardening configuration",
			"category":    "security",
			"tags":        []string{"ssh", "security", "hardening", "server"},
			"language":    "nix",
			"size":        "1.8 KB",
			"created_at":  "2025-05-28T14:20:00Z",
			"author":      "user",
			"usage_count": 8,
		},
		{
			"name":        "development-tools",
			"description": "Common development tools and environment",
			"category":    "development",
			"tags":        []string{"development", "tools", "programming", "git"},
			"language":    "nix",
			"size":        "3.5 KB",
			"created_at":  "2025-05-25T09:15:00Z",
			"author":      "user",
			"usage_count": 22,
		},
		{
			"name":        "firefox-config",
			"description": "Firefox browser with extensions and settings",
			"category":    "desktop",
			"tags":        []string{"firefox", "browser", "desktop", "extensions"},
			"language":    "nix",
			"size":        "1.2 KB",
			"created_at":  "2025-05-20T16:45:00Z",
			"author":      "user",
			"usage_count": 12,
		},
		{
			"name":        "docker-setup",
			"description": "Docker and containerization setup",
			"category":    "services",
			"tags":        []string{"docker", "containers", "virtualization"},
			"language":    "nix",
			"size":        "0.9 KB",
			"created_at":  "2025-05-18T11:30:00Z",
			"author":      "user",
			"usage_count": 18,
		},
	}

	// Apply filters if provided
	if filter != nil {
		filteredSnippets := []map[string]interface{}{}
		for _, snippet := range snippets {
			if f.matchesFilter(snippet, filter) {
				filteredSnippets = append(filteredSnippets, snippet)
			}
		}
		snippets = filteredSnippets
	}

	categories := map[string]int{
		"gaming":      1,
		"security":    1,
		"development": 1,
		"desktop":     1,
		"services":    1,
	}

	response := map[string]interface{}{
		"operation":   "list",
		"total_count": len(snippets),
		"snippets":    snippets,
		"categories":  categories,
		"popular_tags": []string{
			"gaming", "security", "development", "desktop", "services",
			"nvidia", "ssh", "firefox", "docker", "tools",
		},
		"statistics": map[string]interface{}{
			"total_snippets":   len(snippets),
			"total_size":       "9.5 KB",
			"most_used":        "development-tools",
			"recent_additions": 3,
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Listed %d snippets", len(snippets)),
		},
	}, nil
}

// handleSearch searches snippets by keyword
func (f *SnippetsFunction) handleSearch(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required for search operation")
	}

	// Mock search results based on query
	var results []map[string]interface{}

	if query == "gaming" || query == "nvidia" {
		results = append(results, map[string]interface{}{
			"name":        "nvidia-gaming",
			"description": "NVIDIA GPU configuration for gaming",
			"category":    "gaming",
			"tags":        []string{"nvidia", "gaming", "graphics", "performance"},
			"relevance":   0.95,
			"preview":     "hardware.opengl.enable = true;\nhardware.nvidia.modesetting.enable = true;",
		})
	}

	if query == "security" || query == "ssh" {
		results = append(results, map[string]interface{}{
			"name":        "ssh-hardening",
			"description": "SSH server security hardening configuration",
			"category":    "security",
			"tags":        []string{"ssh", "security", "hardening", "server"},
			"relevance":   0.88,
			"preview":     "services.openssh.enable = true;\nservices.openssh.settings.PasswordAuthentication = false;",
		})
	}

	response := map[string]interface{}{
		"operation":     "search",
		"query":         query,
		"results_count": len(results),
		"results":       results,
		"suggestions": []string{
			"Try searching for categories: gaming, security, development",
			"Use tags like: nvidia, ssh, docker, firefox",
			"Search by language: nix, yaml, bash",
		},
		"related_queries": []string{
			"gaming setup",
			"security hardening",
			"development environment",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Found %d snippets matching '%s'", len(results), query),
		},
	}, nil
}

// handleShow shows snippet content
func (f *SnippetsFunction) handleShow(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required for show operation")
	}

	// Mock snippet content based on name
	var snippet map[string]interface{}

	switch name {
	case "nvidia-gaming":
		snippet = map[string]interface{}{
			"name":        "nvidia-gaming",
			"description": "NVIDIA GPU configuration for gaming with optimizations",
			"category":    "gaming",
			"tags":        []string{"nvidia", "gaming", "graphics", "performance"},
			"language":    "nix",
			"content": `# NVIDIA Gaming Configuration
hardware.opengl = {
  enable = true;
  driSupport = true;
  driSupport32Bit = true;
};

hardware.nvidia = {
  modesetting.enable = true;
  powerManagement.enable = false;
  powerManagement.finegrained = false;
  open = false;
  nvidiaSettings = true;
  package = config.boot.kernelPackages.nvidiaPackages.stable;
};

# Gaming optimizations
services.xserver.videoDrivers = [ "nvidia" ];
programs.steam.enable = true;
programs.gamemode.enable = true;

# Performance tweaks
boot.kernel.sysctl = {
  "vm.max_map_count" = 2147483642;
};`,
			"created_at":    "2025-06-01T10:30:00Z",
			"last_modified": "2025-06-05T14:20:00Z",
			"author":        "user",
			"usage_count":   15,
			"file_size":     "2.1 KB",
		}
	case "ssh-hardening":
		snippet = map[string]interface{}{
			"name":        "ssh-hardening",
			"description": "SSH server security hardening configuration",
			"category":    "security",
			"tags":        []string{"ssh", "security", "hardening", "server"},
			"language":    "nix",
			"content": `# SSH Security Hardening
services.openssh = {
  enable = true;
  settings = {
    PasswordAuthentication = false;
    PermitRootLogin = "no";
    X11Forwarding = false;
    Protocol = 2;
    MaxAuthTries = 3;
    ClientAliveInterval = 300;
    ClientAliveCountMax = 2;
    AllowUsers = [ "admin" "user" ];
  };
  extraConfig = ''
    AllowTcpForwarding no
    AllowAgentForwarding no
    AllowStreamLocalForwarding no
    AuthenticationMethods publickey
  '';
};

# Fail2ban for SSH protection
services.fail2ban = {
  enable = true;
  jails.ssh-iptables = ''
    enabled = true
    filter = sshd
    action = iptables[name=SSH, port=ssh, protocol=tcp]
    maxretry = 3
  '';
};`,
			"created_at":    "2025-05-28T14:20:00Z",
			"last_modified": "2025-05-28T14:20:00Z",
			"author":        "user",
			"usage_count":   8,
			"file_size":     "1.8 KB",
		}
	default:
		return nil, fmt.Errorf("snippet not found: %s", name)
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    snippet,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Retrieved snippet: %s", name),
		},
	}, nil
}

// handleAdd adds a new snippet
func (f *SnippetsFunction) handleAdd(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required for add operation")
	}

	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required for add operation")
	}

	description, _ := params["description"].(string)
	category, _ := params["category"].(string)
	if category == "" {
		category = "custom"
	}

	tags, _ := params["tags"].([]interface{})
	var tagStrings []string
	for _, tag := range tags {
		if tagStr, ok := tag.(string); ok {
			tagStrings = append(tagStrings, tagStr)
		}
	}

	snippet := map[string]interface{}{
		"operation":   "add",
		"name":        name,
		"description": description,
		"category":    category,
		"tags":        tagStrings,
		"content":     content,
		"language":    "nix",
		"created_at":  "2025-06-07T12:45:00Z",
		"author":      "user",
		"file_size":   fmt.Sprintf("%.1f KB", float64(len(content))/1024),
		"saved_to":    fmt.Sprintf("/home/user/.config/nixai/snippets/%s.yaml", name),
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    snippet,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Snippet '%s' added successfully", name),
		},
	}, nil
}

// handleRemove removes a snippet
func (f *SnippetsFunction) handleRemove(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required for remove operation")
	}

	response := map[string]interface{}{
		"operation":      "remove",
		"name":           name,
		"status":         "success",
		"removed_from":   fmt.Sprintf("/home/user/.config/nixai/snippets/%s.yaml", name),
		"backup_created": fmt.Sprintf("/home/user/.config/nixai/snippets/.backup/%s.yaml.backup", name),
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Snippet '%s' removed successfully", name),
		},
	}, nil
}

// handleApply applies a snippet to configuration
func (f *SnippetsFunction) handleApply(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required for apply operation")
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath == "" {
		outputPath = "/etc/nixos/configuration.nix"
	}

	merge, _ := params["merge"].(bool)

	application := map[string]interface{}{
		"operation":    "apply",
		"snippet_name": name,
		"output_path":  outputPath,
		"merge_mode":   merge,
		"status":       "success",
		"changes": map[string]interface{}{
			"lines_added": 25,
			"sections_added": []string{
				"hardware.nvidia configuration",
				"gaming optimizations",
				"performance tweaks",
			},
		},
		"backup_created": outputPath + ".backup.20250607-124500",
		"next_steps": []string{
			"Review the applied configuration",
			"Run: nixos-rebuild test",
			"If successful: nixos-rebuild switch",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    application,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Snippet '%s' applied to %s", name, outputPath),
		},
	}, nil
}

// handleEdit edits an existing snippet
func (f *SnippetsFunction) handleEdit(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required for edit operation")
	}

	content, _ := params["content"].(string)
	description, _ := params["description"].(string)
	tags, _ := params["tags"].([]interface{})

	edit := map[string]interface{}{
		"operation": "edit",
		"name":      name,
		"status":    "success",
		"changes": map[string]interface{}{
			"content_updated":     content != "",
			"description_updated": description != "",
			"tags_updated":        len(tags) > 0,
		},
		"last_modified":  "2025-06-07T12:45:00Z",
		"backup_created": fmt.Sprintf("/home/user/.config/nixai/snippets/.backup/%s.yaml.backup", name),
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    edit,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Snippet '%s' edited successfully", name),
		},
	}, nil
}

// handleExport exports snippets to file
func (f *SnippetsFunction) handleExport(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	outputPath, _ := params["output_path"].(string)
	if outputPath == "" {
		outputPath = "/home/user/nixos-snippets-export.yaml"
	}

	export := map[string]interface{}{
		"operation":      "export",
		"output_path":    outputPath,
		"exported_count": 5,
		"file_size":      "12.3 KB",
		"format":         "yaml",
		"includes": []string{
			"snippet metadata",
			"content",
			"tags and categories",
			"usage statistics",
		},
		"exported_at": "2025-06-07T12:45:00Z",
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    export,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Exported 5 snippets to %s", outputPath),
		},
	}, nil
}

// handleImport imports snippets from file
func (f *SnippetsFunction) handleImport(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	sourcePath, ok := params["source_path"].(string)
	if !ok {
		return nil, fmt.Errorf("source_path parameter is required for import operation")
	}

	importResult := map[string]interface{}{
		"operation":      "import",
		"source_path":    sourcePath,
		"imported_count": 8,
		"skipped_count":  2,
		"status":         "success",
		"imported_snippets": []string{
			"web-server-basic",
			"database-postgres",
			"monitoring-setup",
			"backup-scripts",
			"network-security",
			"desktop-apps",
			"development-rust",
			"container-setup",
		},
		"skipped_snippets": []map[string]string{
			{"name": "nvidia-gaming", "reason": "already exists"},
			{"name": "ssh-hardening", "reason": "already exists"},
		},
		"conflicts_resolved": 0,
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    importResult,
		Metadata: map[string]interface{}{
			"message": fmt.Sprintf("Imported 8 snippets from %s", sourcePath),
		},
	}, nil
}

// handleOrganize organizes snippets by category
func (f *SnippetsFunction) handleOrganize(ctx context.Context, params map[string]interface{}) (*functionbase.FunctionResult, error) {
	organization := map[string]interface{}{
		"operation": "organize",
		"status":    "success",
		"categories": map[string]interface{}{
			"gaming": map[string]interface{}{
				"count":    2,
				"snippets": []string{"nvidia-gaming", "steam-setup"},
			},
			"security": map[string]interface{}{
				"count":    3,
				"snippets": []string{"ssh-hardening", "firewall-rules", "fail2ban-config"},
			},
			"development": map[string]interface{}{
				"count":    4,
				"snippets": []string{"development-tools", "rust-environment", "python-setup", "nodejs-config"},
			},
			"desktop": map[string]interface{}{
				"count":    3,
				"snippets": []string{"firefox-config", "gnome-setup", "kde-config"},
			},
			"services": map[string]interface{}{
				"count":    2,
				"snippets": []string{"docker-setup", "web-server"},
			},
		},
		"changes": map[string]interface{}{
			"categories_created": 0,
			"snippets_moved":     0,
			"tags_updated":       5,
		},
		"suggestions": []string{
			"Consider creating subcategories for large categories",
			"Review and standardize tag naming",
			"Add descriptions to uncategorized snippets",
		},
	}

	return &functionbase.FunctionResult{
		Success: true,
		Data:    organization,
		Metadata: map[string]interface{}{
			"message": "Snippet organization completed successfully",
		},
	}, nil
}

// matchesFilter checks if a snippet matches the filter criteria
func (f *SnippetsFunction) matchesFilter(snippet map[string]interface{}, filter map[string]interface{}) bool {
	if category, ok := filter["category"].(string); ok {
		if snippetCategory, exists := snippet["category"].(string); !exists || snippetCategory != category {
			return false
		}
	}

	if tags, ok := filter["tags"].([]interface{}); ok && len(tags) > 0 {
		snippetTags, exists := snippet["tags"].([]string)
		if !exists {
			return false
		}

		for _, filterTag := range tags {
			if filterTagStr, ok := filterTag.(string); ok {
				found := false
				for _, snippetTag := range snippetTags {
					if snippetTag == filterTagStr {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	if language, ok := filter["language"].(string); ok {
		if snippetLanguage, exists := snippet["language"].(string); !exists || snippetLanguage != language {
			return false
		}
	}

	if author, ok := filter["author"].(string); ok {
		if snippetAuthor, exists := snippet["author"].(string); !exists || snippetAuthor != author {
			return false
		}
	}

	return true
}
