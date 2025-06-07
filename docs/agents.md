# Agent System Documentation

The nixai project implements a sophisticated agent system that provides specialized AI-powered assistance for different aspects of NixOS configuration and troubleshooting. This document explains how the agent system works, what each agent does, and how to use them effectively.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Core Agent Interface](#core-agent-interface)
- [Available Agents](#available-agents)
- [Context System](#context-system)
- [Role System](#role-system)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)

## Overview

The agent system provides a modular approach to AI-powered NixOS assistance. Each agent specializes in a specific domain (e.g., diagnostics, build issues, Home Manager configuration) and uses role-based prompting to deliver targeted expertise. All agents share a common interface while providing specialized functionality and context awareness.

### Key Features

- **Specialized Expertise**: Each agent is tailored for specific NixOS domains
- **Context Awareness**: Agents can process structured context data for better responses
- **Role-Based Prompting**: Consistent behavior through predefined role templates
- **Provider Abstraction**: Support for multiple AI providers (Ollama, OpenAI, Gemini, etc.)
- **Interactive & CLI Modes**: Works in both interactive sessions and command-line interfaces

## Architecture

The agent system is built around several key components:

```
Agent Interface
├── BaseAgent (Common functionality)
├── Provider Abstraction (AI backend)
├── Role System (Behavior templates)
└── Context System (Structured data)
```

### Core Components

1. **Agent Interface**: Defines standard methods all agents must implement
2. **BaseAgent**: Provides common functionality shared by all agents
3. **Provider System**: Abstracts different AI providers (Ollama, OpenAI, etc.)
4. **Role System**: Provides specialized prompt templates for different behaviors
5. **Context System**: Structured data types for domain-specific information

## Core Agent Interface

All agents implement the `Agent` interface defined in `internal/ai/agent/agent.go`:

```go
type Agent interface {
    Query(ctx context.Context, question string) (string, error)
    GenerateResponse(ctx context.Context, prompt string) (string, error)
    SetRole(role roles.RoleType) error
    SetContext(contextData interface{})
    SetProvider(provider ai.Provider)
}
```

### Method Descriptions

- **`Query`**: Handles domain-specific questions with role-based prompting
- **`GenerateResponse`**: Generates responses using raw prompts (less specialized)
- **`SetRole`**: Configures the agent's behavioral role and prompt template
- **`SetContext`**: Provides structured context data for enhanced responses
- **`SetProvider`**: Sets the AI provider backend (Ollama, OpenAI, etc.)

## Available Agents

### Core Agents

| Agent | Role | Purpose | Context Type |
|-------|------|---------|--------------|
| **AskAgent** | `ask` | Direct question answering | `AskContext` |
| **DiagnoseAgent** | `diagnose` | System diagnostics and troubleshooting | `DiagnosticContext` |
| **HelpAgent** | `help` | Command and feature guidance | `HelpContext` |
| **InteractiveAgent** | `interactive` | Interactive sessions with history | `InteractiveContext` |

### Configuration Agents

| Agent | Role | Purpose | Context Type |
|-------|------|---------|--------------|
| **ExplainOptionAgent** | `explain-option` | NixOS option explanations | `OptionContext` |
| **ExplainHomeOptionAgent** | `explain-home-option` | Home Manager option explanations | `HomeOptionContext` |
| **ConfigureAgent** | `configure` | System configuration assistance | `ConfigurationContext` |
| **ConfigAgent** | `config` | Configuration file management | N/A |

### Development Agents

| Agent | Role | Purpose | Context Type |
|-------|------|---------|--------------|
| **BuildAgent** | `build` | Build issues and derivations | `BuildContext` |
| **DevenvAgent** | `devenv` | Development environment setup | `DevenvContext` |
| **PackageRepoAgent** | `package-repo` | Package repository analysis | `PackageRepoContext` |
| **FlakeAgent** | `flake` | Nix flakes management | `FlakeContext` |
| **TemplatesAgent** | `templates` | Template and boilerplate generation | `TemplateContext` |

### System Management Agents

| Agent | Role | Purpose | Context Type |
|-------|------|---------|--------------|
| **DoctorAgent** | `doctor` | System health diagnostics | `DoctorContext` |
| **GCAgent** | `gc` | Garbage collection and cleanup | `GCContext` |
| **HardwareAgent** | `hardware` | Hardware configuration and drivers | `HardwareContext` |
| **MachinesAgent** | `machines` | Multi-machine management | `MachinesContext` |
| **MigrateAgent** | `migrate` | System migration assistance | `MigrationContext` |

### Specialized Agents

| Agent | Role | Purpose | Context Type |
|-------|------|---------|--------------|
| **CommunityAgent** | `community` | Community resources and contributions | `CommunityContext` |
| **LearnAgent** | `learn` | Educational content and tutorials | `LearnContext` |
| **NeovimSetupAgent** | `neovim-setup` | Neovim configuration on NixOS | `NeovimContext` |
| **SearchAgent** | `search` | Package and option searching | `SearchContext` |
| **SnippetsAgent** | `snippets` | Code snippets and examples | `SnippetsContext` |
| **StoreAgent** | `store` | Nix store management | `StoreContext` |
| **LogsAgent** | `logs` | Log analysis and interpretation | `LogsContext` |
| **CompletionAgent** | `completion` | Shell completion assistance | `CompletionContext` |
| **McpServerAgent** | `mcp-server` | MCP server setup and management | `McpServerContext` |

## Context System

Each agent can work with structured context data to provide more accurate and relevant responses. Context types provide domain-specific information that agents use to enhance their responses.

### Common Context Patterns

Most context types include:
- **Identifiers**: Names, paths, or IDs relevant to the domain
- **Configuration**: Current settings or preferences
- **Environment**: System state, versions, or capabilities
- **Metadata**: Additional contextual information

### Example Context Types

#### AskContext
```go
type AskContext struct {
    Question      string
    Category      string // "configuration", "troubleshooting", etc.
    Urgency       string // "low", "medium", "high"
    Context       string // Additional context from documentation
    RelatedTopics []string
    Metadata      map[string]string
}
```

#### DiagnosticContext
```go
type DiagnosticContext struct {
    LogEntries     []string
    ErrorMessages  []string
    SystemInfo     string
    ConfigSnippet  string
    FailedServices []string
    RecentChanges  []string
}
```

#### BuildContext
```go
type BuildContext struct {
    ProjectPath     string
    BuildCommand    string
    ErrorOutput     string
    Dependencies    []string
    BuildSystem     string // "cargo", "npm", "make", etc.
    TargetPlatform  string
}
```

## Role System

The role system provides behavioral templates that define how agents should respond. Roles are defined in `internal/ai/roles/roles.go` and include specialized prompt templates.

### Available Roles

| Role | Purpose | Behavior |
|------|---------|----------|
| `ask` | General questions | Clear, concise answers |
| `diagnose` | Problem solving | Structured diagnostic approach |
| `explain-option` | NixOS options | Detailed option explanations with examples |
| `explain-home-option` | Home Manager options | User-focused configuration guidance |
| `build` | Build issues | Technical troubleshooting for builds |
| `doctor` | System health | Comprehensive system analysis |
| `interactive` | Sessions | Conversational, context-aware responses |
| `help` | Command guidance | Clear instructions and examples |

### Role Prompts

Each role has a detailed prompt template that defines:
- **Expertise Areas**: What the agent specializes in
- **Response Format**: How to structure answers
- **Focus Areas**: Key aspects to emphasize
- **Best Practices**: Guidelines for recommendations

## Usage Examples

### Basic Agent Usage

```bash
# Direct question with default agent
nixai "How do I configure nginx?"

# Use specific agent
nixai --agent ask "How do I configure nginx?"

# Use specific role
nixai --role explain-option --agent ask "services.nginx.enable"
```

### Using Context Files

```bash
# Provide context from file
nixai --context-file system-info.json --agent diagnose "System won't boot"

# Pipe logs for analysis
journalctl -u nginx | nixai --agent logs --role diagnose
```

### Interactive Sessions

```bash
# Start interactive session with specific agent
nixai interactive --agent learn --role learn

# Session with context
nixai interactive --agent devenv --context-file project.json
```

### Advanced Usage

```bash
# Build troubleshooting with context
nixai --agent build --role build --context-file build-error.json "Fix build failure"

# Home Manager configuration
nixai --agent explain-home-option --role explain-home-option "programs.git"

# System migration assistance
nixai --agent migrate --role migrate --context-file migration-plan.json "Migration steps"
```

## Context File Examples

### Build Context
```json
{
  "projectPath": "/home/user/myproject",
  "buildCommand": "nix build",
  "errorOutput": "error: hash mismatch in fixed-output derivation",
  "dependencies": ["rustc", "cargo"],
  "buildSystem": "cargo",
  "targetPlatform": "x86_64-linux"
}
```

### Diagnostic Context
```json
{
  "logEntries": [
    "systemd[1]: nginx.service: Failed with result 'exit-code'",
    "nginx[1234]: nginx: [emerg] bind() to 0.0.0.0:80 failed"
  ],
  "errorMessages": ["Port 80 already in use"],
  "systemInfo": "NixOS 23.11, nginx 1.24.0",
  "configSnippet": "services.nginx.enable = true;",
  "failedServices": ["nginx"],
  "recentChanges": ["Added nginx configuration"]
}
```

### Home Manager Context
```json
{
  "optionName": "programs.git.enable",
  "currentValue": "false",
  "userLevel": "intermediate",
  "category": "programs",
  "dotfileLocation": "/home/user/.config/git",
  "metadata": {
    "editor": "neovim",
    "shell": "fish"
  }
}
```

## Best Practices

### Agent Selection

1. **Choose the Right Agent**: Match the agent to your specific need
   - Use `DiagnoseAgent` for troubleshooting
   - Use `ExplainOptionAgent` for understanding NixOS options
   - Use `BuildAgent` for build-related issues
   - Use `HelpAgent` for general guidance

2. **Provide Context**: Use context files when available for better responses
3. **Use Appropriate Roles**: Select roles that match your use case

### Context Usage

1. **Structure Context Properly**: Use the correct context type for each agent
2. **Include Relevant Information**: Provide logs, configs, and error messages
3. **Keep Context Focused**: Don't include irrelevant information

### Provider Configuration

1. **Choose Appropriate Providers**: 
   - Use `ollama` for privacy-focused local inference
   - Use `openai` for advanced reasoning capabilities
   - Use `gemini` for balanced performance
2. **Configure Models**: Select models appropriate for your hardware and needs

### Interactive Sessions

1. **Start with Context**: Provide initial context when starting sessions
2. **Use Session History**: Leverage conversation history for follow-up questions
3. **Switch Agents**: Use different agents for different phases of work

### Integration Patterns

1. **Pipe Integration**: Use with system commands for log analysis
2. **Scripting**: Integrate into automation workflows
3. **Development Workflow**: Use in CI/CD for build troubleshooting

## Advanced Features

### Agent Chaining

Agents can work together in complex workflows:

```bash
# First diagnose the issue
nixai --agent diagnose "System performance issues" > diagnosis.txt

# Then get build-specific help
nixai --agent build --context-file diagnosis.txt "Optimize build performance"

# Finally get learning resources
nixai --agent learn "Performance optimization techniques"
```

### Custom Context Creation

Create context files programmatically:

```bash
# Generate build context from current directory
nixai generate-context --type build --output build-context.json

# Use generated context
nixai --agent build --context-file build-context.json "Fix build issues"
```

### Provider Fallback

The system supports provider fallback for reliability:

```yaml
# In config.yaml
ai_provider: "ollama"
fallback_providers: ["openai", "gemini"]
```

## Troubleshooting

### Common Issues

1. **Agent Not Found**: Check agent name spelling and availability
2. **Context Type Mismatch**: Ensure context file matches agent requirements
3. **Provider Unavailable**: Check provider configuration and connectivity
4. **Role Invalid**: Verify role name against available roles

### Debug Mode

Enable debug logging for troubleshooting:

```bash
nixai --debug --agent ask "test question"
```

### Health Checks

Use the doctor agent for system health:

```bash
nixai --agent doctor "Check nixai configuration"
```

## Contributing

When adding new agents:

1. **Implement Agent Interface**: All methods must be implemented
2. **Define Context Type**: Create appropriate context structures
3. **Add Role Definition**: Include role in the roles system
4. **Write Tests**: Comprehensive test coverage required
5. **Update Documentation**: Add agent to this documentation

### Agent Template

```go
type MyAgent struct {
    BaseAgent
}

func NewMyAgent(provider ai.Provider) *MyAgent {
    return &MyAgent{
        BaseAgent: BaseAgent{
            provider: provider,
            role:     roles.RoleMyAgent,
        },
    }
}

func (a *MyAgent) Query(ctx context.Context, question string) (string, error) {
    // Implementation
}

func (a *MyAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
    // Implementation
}
```

This agent system provides a flexible and powerful foundation for AI-assisted NixOS configuration and troubleshooting, with each agent bringing specialized expertise to solve specific types of problems.
