package mcp

import (
	"fmt"
	"strings"
)

// Phase 3: Community & Learning Tools Handlers

// handleGetCommunityResources - Get NixOS community resources, forums, and channels
func (m *MCPServer) handleGetCommunityResources(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling get-community-resources request | args=%v", args))

	// Extract parameters
	resourceType, _ := args["type"].(string)
	if resourceType == "" {
		resourceType = "all"
	}

	category, _ := args["category"].(string)
	if category == "" {
		category = "general"
	}

	// Mock community resources data
	resources := map[string]interface{}{
		"forums": []map[string]interface{}{
			{
				"name":        "NixOS Discourse",
				"url":         "https://discourse.nixos.org/",
				"description": "Official NixOS community forum for discussions, announcements, and help",
				"type":        "forum",
				"activity":    "Very Active",
			},
			{
				"name":        "Reddit r/NixOS",
				"url":         "https://www.reddit.com/r/NixOS/",
				"description": "NixOS subreddit for community discussions and sharing",
				"type":        "forum",
				"activity":    "Active",
			},
		},
		"chat": []map[string]interface{}{
			{
				"name":        "NixOS Matrix",
				"url":         "https://matrix.to/#/#nixos:nixos.org",
				"description": "Official NixOS Matrix channel for real-time chat and support",
				"type":        "chat",
				"activity":    "Very Active",
			},
			{
				"name":        "NixOS IRC",
				"url":         "irc://irc.libera.chat/#nixos",
				"description": "Traditional IRC channel for NixOS discussions",
				"type":        "chat",
				"activity":    "Active",
			},
		},
	}

	// Filter by resource type if specified
	var filteredResources interface{}
	if resourceType != "all" {
		if resourceData, exists := resources[resourceType]; exists {
			filteredResources = resourceData
		} else {
			filteredResources = []interface{}{}
		}
	} else {
		filteredResources = resources
	}

	response := fmt.Sprintf(`🌟 **NixOS Community Resources**

📍 **Resource Type**: %s
📂 **Category**: %s

%s

💡 **Getting Started Tips**:
• Join the Discourse forum for structured discussions
• Use Matrix/IRC for real-time help and chat
• Follow awesome-nix for curated resources
• Start with Nix Pills if you're new to Nix

🔗 **Quick Links**:
• Official Website: https://nixos.org/
• Documentation: https://nixos.org/manual/
• Package Search: https://search.nixos.org/
• Options Search: https://search.nixos.org/options

For more specific help, use: nixai community --help`,
		resourceType, category, formatResourceData(filteredResources))

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleGetLearningResources - Get structured learning paths and tutorials
func (m *MCPServer) handleGetLearningResources(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling get-learning-resources request | args=%v", args))

	// Extract parameters
	level, _ := args["level"].(string)
	if level == "" {
		level = "beginner"
	}

	topic, _ := args["topic"].(string)
	if topic == "" {
		topic = "general"
	}

	// Mock learning resources data
	resources := map[string]interface{}{
		"beginner": []map[string]interface{}{
			{
				"title":       "NixOS Basics",
				"url":         "https://nixos.org/learn.html",
				"description": "Introduction to NixOS concepts and installation",
				"duration":    "2-3 hours",
				"type":        "tutorial",
			},
		},
	}

	// Get resources for the specified level
	levelResources, exists := resources[level]
	if !exists {
		levelResources = resources["beginner"]
	}

	response := fmt.Sprintf(`📚 **NixOS Learning Resources**

🎯 **Level**: %s
📋 **Topic**: %s

%s`,
		level, topic, formatLearningData(levelResources))

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleGetConfigurationTemplates - Get pre-built configuration templates
func (m *MCPServer) handleGetConfigurationTemplates(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling get-configuration-templates request | args=%v", args))

	templateType, _ := args["type"].(string)
	if templateType == "" {
		templateType = "desktop"
	}

	response := fmt.Sprintf(`📋 **NixOS Configuration Templates**

🏷️ **Type**: %s

Templates available for download and customization.`,
		templateType)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleGetConfigurationSnippets - Get reusable configuration code snippets
func (m *MCPServer) handleGetConfigurationSnippets(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling get-configuration-snippets request | args=%v", args))

	category, _ := args["category"].(string)
	if category == "" {
		category = "common"
	}

	response := fmt.Sprintf(`🧩 **NixOS Configuration Snippets**

📂 **Category**: %s

Code snippets for common configurations.`,
		category)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleManageMachines - Manage multiple NixOS machines and configurations
func (m *MCPServer) handleManageMachines(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling manage-machines request | args=%v", args))

	action, _ := args["action"].(string)
	if action == "" {
		action = "list"
	}

	response := fmt.Sprintf(`🖥️ **NixOS Machine Management**

📋 **Action**: %s

Machine management functionality.`,
		action)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleCompareConfigurations - Compare configurations between machines or versions
func (m *MCPServer) handleCompareConfigurations(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling compare-configurations request | args=%v", args))

	source, _ := args["source"].(string)
	target, _ := args["target"].(string)

	response := fmt.Sprintf(`⚖️ **Configuration Comparison**

🔍 **Source**: %s
📋 **Target**: %s

Configuration comparison results.`,
		source, target)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleGetDeploymentStatus - Get deployment status and history
func (m *MCPServer) handleGetDeploymentStatus(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling get-deployment-status request | args=%v", args))

	deploymentId, _ := args["deployment_id"].(string)

	response := fmt.Sprintf(`🚀 **Deployment Status**

🆔 **Deployment ID**: %s

Deployment status and history.`,
		deploymentId)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// handleInteractiveAssistance - Provide interactive help and guidance
func (m *MCPServer) handleInteractiveAssistance(args map[string]interface{}) (interface{}, error) {
	m.logger.Debug(fmt.Sprintf("Handling interactive-assistance request | args=%v", args))

	topic, _ := args["topic"].(string)
	if topic == "" {
		topic = "general"
	}

	response := fmt.Sprintf(`🤖 **Interactive NixOS Assistant**

📚 **Topic**: %s

Interactive assistance and guidance.`,
		topic)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}, nil
}

// Helper formatting functions

func formatResourceData(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		var result strings.Builder
		for category, resources := range v {
			result.WriteString(fmt.Sprintf("### %s\n", strings.Title(category)))
			if resourceList, ok := resources.([]map[string]interface{}); ok {
				for _, resource := range resourceList {
					result.WriteString(fmt.Sprintf("• **%s**: %s\n  URL: %s\n\n",
						resource["name"], resource["description"], resource["url"]))
				}
			}
		}
		return result.String()
	case []map[string]interface{}:
		var result strings.Builder
		for _, resource := range v {
			result.WriteString(fmt.Sprintf("• **%s**: %s\n  URL: %s\n\n",
				resource["name"], resource["description"], resource["url"]))
		}
		return result.String()
	}
	return "No resources available"
}

func formatLearningData(data interface{}) string {
	if resources, ok := data.([]map[string]interface{}); ok {
		var result strings.Builder
		for i, resource := range resources {
			result.WriteString(fmt.Sprintf("%d. **%s**\n   %s\n   Duration: %s\n   URL: %s\n\n",
				i+1, resource["title"], resource["description"],
				resource["duration"], resource["url"]))
		}
		return result.String()
	}
	return "No learning resources available"
}
