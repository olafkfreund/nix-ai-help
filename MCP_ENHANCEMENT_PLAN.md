# 🚀 MCP Server Enhancement Plan for NixOS Users

**Goal**: Transform the nixai MCP server into a comprehensive NixOS assistance platform for VS Code, Neovim, and other editors.

**Current State**: 12 tools (4 docs, 4 LSP, 4 context)  
**Target State**: 40+ tools covering the full nixai command suite

---

## 📋 Implementation Phases

### **Phase 1: Core NixOS Operations (8 New Tools) ⭐ HIGH PRIORITY**

#### 1. **Build & Diagnostics Tools**
```go
// build_system_analyze - Analyze build issues and suggest fixes
{
  "name": "build_system_analyze",
  "description": "Analyze build issues and suggest fixes with AI",
  "inputSchema": {
    "type": "object", 
    "properties": {
      "buildLog": {"type": "string"},
      "project": {"type": "string"},
      "depth": {"type": "string", "enum": ["basic", "detailed"], "default": "basic"}
    }
  }
}

// diagnose_system - Diagnose NixOS configuration and system issues
{
  "name": "diagnose_system", 
  "description": "Diagnose NixOS system issues from logs or config files",
  "inputSchema": {
    "type": "object",
    "properties": {
      "logContent": {"type": "string"},
      "logType": {"type": "string", "enum": ["system", "build", "service"], "default": "system"},
      "context": {"type": "string"}
    }
  }
}
```

#### 2. **Configuration Management Tools**
```go
// generate_configuration - Interactive NixOS configuration generation
{
  "name": "generate_configuration",
  "description": "Generate NixOS configuration based on requirements",
  "inputSchema": {
    "type": "object",
    "properties": {
      "configType": {"type": "string", "enum": ["desktop", "server", "minimal"], "default": "desktop"},
      "services": {"type": "array", "items": {"type": "string"}},
      "features": {"type": "array", "items": {"type": "string"}}
    }
  }
}

// validate_configuration - Validate NixOS configuration syntax and logic
{
  "name": "validate_configuration",
  "description": "Validate NixOS configuration files for syntax and logic errors",
  "inputSchema": {
    "type": "object", 
    "properties": {
      "configContent": {"type": "string"},
      "configPath": {"type": "string"},
      "checkLevel": {"type": "string", "enum": ["syntax", "logic", "full"], "default": "full"}
    }
  }
}
```

#### 3. **Package & Service Tools**
```go
// analyze_package_repo - Analyze Git repos and generate Nix derivations  
{
  "name": "analyze_package_repo",
  "description": "Analyze Git repositories and generate Nix derivations",
  "inputSchema": {
    "type": "object",
    "properties": {
      "repoUrl": {"type": "string"},
      "packageName": {"type": "string"},
      "outputFormat": {"type": "string", "enum": ["derivation", "flake", "shell"], "default": "derivation"}
    }
  }
}

// get_service_examples - Get NixOS service configuration examples
{
  "name": "get_service_examples", 
  "description": "Get practical configuration examples for NixOS services",
  "inputSchema": {
    "type": "object",
    "properties": {
      "serviceName": {"type": "string"},
      "useCase": {"type": "string"},
      "detailed": {"type": "boolean", "default": false}
    }
  }
}
```

#### 4. **Health & Maintenance Tools**
```go
// check_system_health - Comprehensive NixOS system health check
{
  "name": "check_system_health",
  "description": "Perform comprehensive NixOS system health checks",
  "inputSchema": {
    "type": "object",
    "properties": {
      "checkType": {"type": "string", "enum": ["quick", "full"], "default": "quick"},
      "includeRecommendations": {"type": "boolean", "default": true}
    }
  }
}

// analyze_garbage_collection - Analyze and suggest garbage collection
{
  "name": "analyze_garbage_collection",
  "description": "Analyze Nix store and suggest safe garbage collection",
  "inputSchema": {
    "type": "object", 
    "properties": {
      "analysisType": {"type": "string", "enum": ["safe", "aggressive"], "default": "safe"},
      "dryRun": {"type": "boolean", "default": true}
    }
  }
}

// get_hardware_info - Get hardware-specific NixOS configuration suggestions
{
  "name": "get_hardware_info",
  "description": "Get hardware detection and optimization suggestions",
  "inputSchema": {
    "type": "object",
    "properties": {
      "detectionType": {"type": "string", "enum": ["auto", "manual"], "default": "auto"},
      "includeOptimizations": {"type": "boolean", "default": true}
    }
  }
}
```

---

### **Phase 2: Development & Workflow Tools (10 New Tools) ⭐ MEDIUM PRIORITY**

#### 1. **Development Environment Tools**
```go
// create_devenv - Create development environments with devenv
{
  "name": "create_devenv",
  "description": "Create development environment using devenv templates",
  "inputSchema": {
    "type": "object",
    "properties": {
      "language": {"type": "string"},
      "framework": {"type": "string"},
      "services": {"type": "array", "items": {"type": "string"}},
      "projectName": {"type": "string"}
    }
  }
}

// suggest_devenv_template - AI-powered development template suggestions
{
  "name": "suggest_devenv_template", 
  "description": "Get AI-powered development environment template suggestions",
  "inputSchema": {
    "type": "object",
    "properties": {
      "description": {"type": "string"},
      "requirements": {"type": "array", "items": {"type": "string"}}
    }
  }
}

// setup_neovim_integration - Setup Neovim integration with nixai
{
  "name": "setup_neovim_integration",
  "description": "Setup and configure Neovim integration with nixai MCP",
  "inputSchema": {
    "type": "object",
    "properties": {
      "configType": {"type": "string", "enum": ["minimal", "full"], "default": "full"},
      "socketPath": {"type": "string", "default": "/tmp/nixai-mcp.sock"}
    }
  }
}
```

#### 2. **Flake Management Tools**
```go
// flake_operations - Perform flake operations (init, update, show, etc.)
{
  "name": "flake_operations",
  "description": "Perform NixOS flake operations and management",
  "inputSchema": {
    "type": "object",
    "properties": {
      "operation": {"type": "string", "enum": ["init", "update", "show", "check"], "default": "show"},
      "flakePath": {"type": "string"},
      "options": {"type": "array", "items": {"type": "string"}}
    }
  }
}

// migrate_to_flakes - Migrate from channels to flakes
{
  "name": "migrate_to_flakes",
  "description": "Migrate NixOS configuration from channels to flakes",
  "inputSchema": {
    "type": "object",
    "properties": {
      "backupName": {"type": "string"},
      "dryRun": {"type": "boolean", "default": true},
      "includeHomeManager": {"type": "boolean", "default": true}
    }
  }
}
```

#### 3. **Dependency & Analysis Tools**
```go
// analyze_dependencies - Analyze NixOS configuration dependencies
{
  "name": "analyze_dependencies",
  "description": "Analyze NixOS configuration dependencies and relationships",
  "inputSchema": {
    "type": "object",
    "properties": {
      "analysisType": {"type": "string", "enum": ["packages", "services", "all"], "default": "all"},
      "maxDepth": {"type": "integer", "default": 3},
      "outputFormat": {"type": "string", "enum": ["text", "graph"], "default": "text"}
    }
  }
}

// explain_dependency_chain - Explain why a package is included
{
  "name": "explain_dependency_chain",
  "description": "Explain why a specific package is included in the system",
  "inputSchema": {
    "type": "object",
    "properties": {
      "packageName": {"type": "string"},
      "showPath": {"type": "boolean", "default": true}
    }
  }
}
```

#### 4. **Store & Performance Tools**
```go
// store_operations - Perform Nix store operations
{
  "name": "store_operations",
  "description": "Perform Nix store backup, restore, and analysis operations",
  "inputSchema": {
    "type": "object",
    "properties": {
      "operation": {"type": "string", "enum": ["backup", "restore", "analyze", "optimize"], "default": "analyze"},
      "path": {"type": "string"},
      "options": {"type": "array", "items": {"type": "string"}}
    }
  }
}

// performance_analysis - Analyze system performance and optimization
{
  "name": "performance_analysis",
  "description": "Analyze NixOS system performance and suggest optimizations",
  "inputSchema": {
    "type": "object",
    "properties": {
      "analysisType": {"type": "string", "enum": ["boot", "build", "runtime"], "default": "runtime"},
      "includeRecommendations": {"type": "boolean", "default": true}
    }
  }
}

// search_advanced - Advanced package and option search
{
  "name": "search_advanced",
  "description": "Advanced multi-source search for packages, options, and configurations",
  "inputSchema": {
    "type": "object", 
    "properties": {
      "query": {"type": "string"},
      "searchType": {"type": "string", "enum": ["packages", "options", "services", "all"], "default": "all"},
      "sources": {"type": "array", "items": {"type": "string"}}
    }
  }
}
```

---

### **Phase 3: Community & Learning Tools (8 New Tools) ⭐ LOW PRIORITY**

#### 1. **Community & Learning**
```go
// get_community_resources - Access NixOS community resources
{
  "name": "get_community_resources",
  "description": "Get NixOS community resources, forums, and support channels",
  "inputSchema": {
    "type": "object",
    "properties": {
      "resourceType": {"type": "string", "enum": ["forums", "documentation", "tutorials", "all"], "default": "all"},
      "topic": {"type": "string"}
    }
  }
}

// get_learning_resources - Get NixOS learning materials
{
  "name": "get_learning_resources", 
  "description": "Get curated NixOS learning materials and tutorials",
  "inputSchema": {
    "type": "object",
    "properties": {
      "skillLevel": {"type": "string", "enum": ["beginner", "intermediate", "advanced"], "default": "beginner"},
      "topic": {"type": "string"}
    }
  }
}

// get_configuration_templates - Get pre-built configuration templates
{
  "name": "get_configuration_templates",
  "description": "Get pre-built NixOS configuration templates",
  "inputSchema": {
    "type": "object",
    "properties": {
      "templateType": {"type": "string", "enum": ["desktop", "server", "development"], "default": "desktop"},
      "features": {"type": "array", "items": {"type": "string"}}
    }
  }
}

// get_configuration_snippets - Get configuration code snippets
{
  "name": "get_configuration_snippets",
  "description": "Get specific NixOS configuration code snippets",
  "inputSchema": {
    "type": "object",
    "properties": {
      "category": {"type": "string"},
      "searchTerm": {"type": "string"},
      "includeExplanation": {"type": "boolean", "default": true}
    }
  }
}
```

#### 2. **Multi-Machine & Deployment**
```go
// manage_machines - Manage multi-machine NixOS deployments
{
  "name": "manage_machines",
  "description": "Manage multi-machine NixOS configurations and deployments",
  "inputSchema": {
    "type": "object", 
    "properties": {
      "operation": {"type": "string", "enum": ["list", "deploy", "sync", "status"], "default": "list"},
      "machine": {"type": "string"},
      "options": {"type": "array", "items": {"type": "string"}}
    }
  }
}

// compare_configurations - Compare NixOS configurations
{
  "name": "compare_configurations",
  "description": "Compare different NixOS configurations and show differences",
  "inputSchema": {
    "type": "object",
    "properties": {
      "source": {"type": "string"},
      "target": {"type": "string"},
      "compareType": {"type": "string", "enum": ["packages", "services", "all"], "default": "all"}
    }
  }
}

// get_deployment_status - Get deployment status for machines
{
  "name": "get_deployment_status",
  "description": "Get deployment status and health for managed machines",
  "inputSchema": {
    "type": "object",
    "properties": {
      "machine": {"type": "string"},
      "detailed": {"type": "boolean", "default": false}
    }
  }
}

// interactive_assistance - Launch interactive TUI assistance
{
  "name": "interactive_assistance",
  "description": "Launch interactive TUI assistance for guided NixOS help",
  "inputSchema": {
    "type": "object",
    "properties": {
      "mode": {"type": "string", "enum": ["guided", "explorer"], "default": "guided"},
      "startingTopic": {"type": "string"}
    }
  }
}
```

---

## 🎯 Implementation Strategy

### **Priority Order**
1. **Phase 1** (8 tools): Core operations that directly enhance daily NixOS workflows
2. **Phase 2** (10 tools): Development-focused tools for programmers using NixOS  
3. **Phase 3** (8 tools): Community and advanced features

### **Implementation Pattern**
Each tool will follow this pattern:
1. **Handler Function**: Implement in `/internal/mcp/enhanced_handlers.go`
2. **Tool Registration**: Add to tool list in `server.go`
3. **Tool Call**: Add case to tools/call switch in `server.go`
4. **Reuse Existing Logic**: Leverage existing CLI command implementations
5. **Response Formatting**: Use existing formatters with MCP-specific adaptations

### **Code Reuse Strategy**
- **Build Tools**: Reuse `internal/build` package logic
- **Config Tools**: Reuse `internal/cli/configure_commands.go` logic  
- **Package Tools**: Reuse `internal/cli/package_repo.go` logic
- **Health Tools**: Reuse `internal/health` package logic
- **Devenv Tools**: Reuse `internal/devenv` package logic
- **Context Integration**: All tools will be context-aware using existing context system

---

## 📊 Expected Benefits

### **For VS Code Users**
- **Complete NixOS workflow** without leaving the editor
- **Context-aware suggestions** based on actual system configuration
- **Real-time diagnostics** and configuration validation
- **AI-powered assistance** for complex NixOS tasks

### **For Neovim Users**  
- **Seamless integration** with existing Neovim workflows
- **MCP-based AI assistance** with full NixOS context
- **In-editor configuration generation** and validation
- **Direct access to nixai's full command suite**

### **For Development Teams**
- **Consistent development environments** via devenv integration
- **Multi-machine management** from any editor
- **Collaborative configuration sharing** and templates
- **Automated deployment and health monitoring**

---

## 🚀 Next Steps

1. **Implement Phase 1** (8 core tools) - Target: 2-3 days
2. **Test integration** with VS Code and Neovim  
3. **Gather user feedback** and prioritize Phase 2 tools
4. **Implement Phase 2** (10 development tools) - Target: 4-5 days
5. **Implement Phase 3** (8 community tools) - Target: 2-3 days

**Total Implementation Time**: ~10 days for 26 new tools (40+ total tools)

This enhancement will make the nixai MCP server the most comprehensive NixOS assistance platform available for any editor!
