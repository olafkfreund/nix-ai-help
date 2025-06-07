# NixAI Role System Documentation

## Overview

The nixai role system is a sophisticated behavioral framework that defines how AI agents respond to user queries and tasks. Each role provides specialized expertise, response patterns, and context-specific guidance for different aspects of NixOS configuration, troubleshooting, and system management.

## Architecture

### Role Type System

Roles are defined as typed constants in `internal/ai/roles/roles.go`:

```go
type RoleType string
```

Each role type maps to a comprehensive prompt template that defines:
- **Expertise Areas**: Specific knowledge domains and capabilities
- **Response Structure**: How to format and organize responses
- **Context Awareness**: What information to consider when analyzing requests
- **Best Practices**: Recommended approaches and safety considerations
- **Resource Integration**: Links to documentation and community resources

### Role Integration

Roles integrate with the agent system through:
- **Agent Interface**: All agents implement role-based behavior via `SetRole()`
- **Prompt Templates**: Role-specific prompts enhance AI responses
- **Context Management**: Roles define what context data is relevant
- **CLI Integration**: Users can specify roles via `--role` flag

## Complete Role Catalog

### Core Assistance Roles

#### `ask` - Direct Question Assistant
- **Purpose**: Answer direct NixOS configuration questions clearly and concisely
- **Usage**: `nixai "question" --role ask`
- **Expertise**: General NixOS knowledge, quick problem solving
- **Response Style**: Clear, direct answers with practical guidance

#### `diagnose` - System Diagnostic Specialist
- **Purpose**: Analyze NixOS problems and provide structured, actionable solutions
- **Usage**: `nixai diagnose --role diagnose`
- **Expertise**: 
  - Log analysis (systemd, build logs, journalctl)
  - Configuration issues (syntax errors, missing options)
  - Build failures (derivation errors, hash mismatches)
  - System services and hardware problems
- **Response Structure**:
  - Problem Summary
  - Root Cause Analysis
  - Fix Steps (numbered, specific)
  - Prevention Guidelines
  - Documentation Links

#### `help` - Command and Feature Guide
- **Purpose**: Help users navigate nixai toolkit and NixOS ecosystem
- **Usage**: `nixai help --role help`
- **Expertise**: Command discovery, workflow optimization, user onboarding
- **Features**:
  - Needs assessment and command recommendations
  - Workflow guidance and feature explanation
  - Learning support and troubleshooting assistance

#### `interactive` - Conversational Assistant
- **Purpose**: Provide step-by-step conversational guidance
- **Usage**: `nixai interactive --role interactive`
- **Expertise**: Session management, adaptive communication, problem-solving workflows
- **Features**:
  - Context continuity throughout conversations
  - Adaptive explanations based on user expertise
  - Breaking complex tasks into manageable steps

### Configuration Management Roles

#### `explain-option` - NixOS Option Explainer
- **Purpose**: Explain NixOS system configuration options in detail
- **Usage**: `nixai explain-option services.nginx.enable --role explain-option`
- **Expertise**: NixOS options, system services, package configuration
- **Response Format**:
  - Option overview (path, type, defaults)
  - Practical context and use cases
  - Configuration examples (basic and advanced)
  - Related options and dependencies
  - Best practices and security considerations

#### `explain-home-option` - Home Manager Option Explainer
- **Purpose**: Explain Home Manager user-level configuration options
- **Usage**: `nixai explain-home-option programs.git.enable --role explain-home-option`
- **Expertise**: User programs, dotfiles, development environments, desktop configuration
- **Focus Areas**:
  - User vs system configuration decisions
  - Development workflow integration
  - Portable configuration management

#### `config` - Configuration Management Specialist
- **Purpose**: Guide configuration architecture and organization
- **Usage**: `nixai config --role config`
- **Expertise**: Modular configurations, file organization, version control
- **Services**:
  - Configuration analysis and improvement recommendations
  - Modular organization patterns
  - Security hardening and optimization

#### `configure` - System Configuration Assistant
- **Purpose**: Guide initial NixOS setup and configuration
- **Usage**: `nixai configure --role configure`
- **Expertise**: Initial setup, hardware configuration, service setup
- **Process**:
  - Environment assessment
  - Configuration planning
  - Step-by-step guidance with validation

### Development and Build Roles

#### `build` - Build System Specialist
- **Purpose**: Handle NixOS build system troubleshooting and optimization
- **Usage**: `nixai build --role build`
- **Expertise**: nixos-rebuild, nix-build, derivations, build failures
- **Capabilities**:
  - Build analysis and error diagnosis
  - Performance optimization and caching strategies
  - Cross-compilation and platform support

#### `package-repo` - Package Repository Analyst
- **Purpose**: Convert external projects into Nix packages
- **Usage**: `nixai package-repo <repo-url> --role package-repo`
- **Expertise**: Repository analysis, derivation creation, build system integration
- **Process**:
  - Project assessment and dependency analysis
  - Nix derivation generation
  - Quality assurance and validation

#### `flake` - Nix Flakes Specialist
- **Purpose**: Modern Nix development workflows and flake management
- **Usage**: `nixai flake --role flake`
- **Expertise**: Flake structure, development workflows, reproducibility
- **Focus Areas**:
  - Flake architecture and best practices
  - Development environment setup
  - Migration from legacy Nix expressions

#### `devenv` - Development Environment Expert
- **Purpose**: Nix-based development environment setup and optimization
- **Usage**: `nixai devenv --role devenv`
- **Expertise**: devenv.sh, language ecosystems, service integration
- **Services**:
  - Environment assessment and configuration generation
  - Service integration (databases, development tools)
  - Team collaboration and reproducible setups

#### `templates` - Template and Scaffolding Expert
- **Purpose**: Template design, generation, and management
- **Usage**: `nixai templates --role templates`
- **Expertise**: Template architecture, configuration scaffolding, best practices
- **Capabilities**:
  - Template generation and customization
  - Quality validation and testing
  - Integration with existing configurations

### System Management Roles

#### `doctor` - System Health Diagnostician
- **Purpose**: Comprehensive system health monitoring and diagnostics
- **Usage**: `nixai doctor --role doctor`
- **Expertise**: System monitoring, preventive maintenance, security assessment
- **Assessment Areas**:
  - System services and performance metrics
  - NixOS-specific health (store integrity, channels)
  - Security vulnerabilities and optimization opportunities

#### `gc` - Garbage Collection Specialist
- **Purpose**: Nix store management and cleanup optimization
- **Usage**: `nixai gc --role gc`
- **Expertise**: Store management, generation cleanup, performance optimization
- **Services**:
  - Storage analysis and cleanup strategy
  - Safe cleanup procedures with validation
  - Automation and scheduling guidance

#### `hardware` - Hardware Configuration Specialist
- **Purpose**: Hardware detection, driver configuration, and optimization
- **Usage**: `nixai hardware --role hardware`
- **Expertise**: Hardware detection, driver management, performance tuning
- **Capabilities**:
  - Hardware assessment and driver configuration
  - Performance optimization for specific hardware
  - Compatibility troubleshooting

#### `store` - Nix Store Management Expert
- **Purpose**: Advanced Nix store operations and optimization
- **Usage**: `nixai store --role store`
- **Expertise**: Store architecture, operations, security, performance
- **Advanced Features**:
  - Store structure analysis and integrity verification
  - Remote store configuration and binary caches
  - Security management and access control

#### `machines` - Multi-Machine Management Expert
- **Purpose**: Distributed NixOS systems and infrastructure automation
- **Usage**: `nixai machines --role machines`
- **Expertise**: Fleet management, deployment strategies, automation
- **Capabilities**:
  - Multi-machine architecture design
  - Deployment orchestration and scaling
  - Distributed system monitoring and troubleshooting

### Specialized Roles

#### `search` - Search and Discovery Assistant
- **Purpose**: Find packages, options, and documentation efficiently
- **Usage**: `nixai search <query> --role search`
- **Expertise**: Package discovery, option search, documentation retrieval
- **Features**:
  - Advanced search strategies and filtering
  - Context enhancement and result presentation
  - Resource integration across multiple sources

#### `learn` - Learning Guide and Educational Assistant
- **Purpose**: Structured NixOS education and skill development
- **Usage**: `nixai learn --role learn`
- **Expertise**: Educational design, concept explanation, hands-on learning
- **Approach**:
  - Learning assessment and path recommendation
  - Progressive skill building with practical exercises
  - Resource curation and progress tracking

#### `logs` - Log Analysis Specialist
- **Purpose**: System log interpretation and issue diagnosis
- **Usage**: `nixai logs --role logs`
- **Expertise**: Log analysis, monitoring setup, performance analysis
- **Capabilities**:
  - Multi-source log correlation and pattern identification
  - Root cause analysis through log chronology
  - Monitoring configuration and alerting setup

#### `migrate` - Migration Assistant
- **Purpose**: System migrations, upgrades, and configuration transfers
- **Usage**: `nixai migrate --role migrate`
- **Expertise**: Version migrations, machine transfers, configuration modernization
- **Process**:
  - Pre-migration assessment and planning
  - Risk management and rollback strategies
  - Post-migration verification and optimization

#### `community` - Community Guide and Resource Coordinator
- **Purpose**: Connect users with NixOS community and resources
- **Usage**: `nixai community --role community`
- **Expertise**: Community resources, contribution guidelines, event information
- **Services**:
  - Resource navigation and expert connections
  - Contribution support and process guidance
  - Community engagement and networking

#### `neovim-setup` - Neovim Configuration Specialist
- **Purpose**: Neovim setup and configuration in NixOS environment
- **Usage**: `nixai neovim-setup --role neovim-setup`
- **Expertise**: Neovim configuration, plugin management, development integration
- **Focus**: NixOS-specific Neovim configuration patterns and best practices

#### `snippets` - Code Snippet Assistant
- **Purpose**: Common NixOS configuration snippets and examples
- **Usage**: `nixai snippets --role snippets`
- **Expertise**: Configuration patterns, common use cases, code examples
- **Features**: Curated snippet library with context and explanations

#### `completion` - Shell Completion Expert
- **Purpose**: Command-line completion systems and optimization
- **Usage**: `nixai completion --role completion`
- **Expertise**: Cross-shell completion, performance optimization, user experience
- **Capabilities**: Completion script generation, installation, and troubleshooting

#### `mcp-server` - MCP Server Management
- **Purpose**: Model Context Protocol server configuration and management
- **Usage**: `nixai mcp-server --role mcp-server`
- **Expertise**: MCP server setup, documentation integration, API management
- **Focus**: Integration with VS Code and documentation sources

## Using Roles

### CLI Usage

Specify roles using the `--role` flag:

```bash
# Use specific role for targeted assistance
nixai "How do I configure nginx?" --role explain-option

# Diagnose system issues
nixai diagnose /var/log/nixos-rebuild.log --role diagnose

# Get help with commands
nixai help --role help

# Interactive troubleshooting
nixai interactive --role interactive
```

### Role Selection Guidelines

Choose roles based on your specific needs:

**For Quick Questions**: Use `ask` for direct, concise answers
**For System Issues**: Use `diagnose` for structured problem-solving
**For Configuration**: Use `explain-option` or `explain-home-option` for detailed explanations
**For Learning**: Use `learn` for educational guidance and skill development
**For Setup**: Use `configure` for initial system configuration
**For Development**: Use `devenv`, `flake`, or `build` for development-related tasks

### Context Integration

Roles work with context data:

```bash
# Provide context file for enhanced analysis
nixai diagnose --role diagnose --context-file error.log

# Specify nixos path for configuration-aware responses
nixai explain-option services.nginx --role explain-option --nixos-path /etc/nixos
```

## Role Validation

The system validates roles before use:

```go
func ValidateRole(role string) bool {
    // Checks if role is supported and properly defined
}
```

Invalid roles will result in clear error messages guiding users to valid options.

## Best Practices

### Role Selection
- **Match Role to Task**: Choose the most specific role for your use case
- **Consider Context**: Some roles work better with additional context data
- **Experiment**: Try different roles to find the most helpful approach

### Context Provision
- **Log Files**: Provide log files for diagnostic roles
- **Configuration Path**: Specify nixos path for configuration-related queries
- **Error Messages**: Include full error messages for build and system issues

### Response Interpretation
- **Follow Structure**: Role responses follow consistent patterns for easy parsing
- **Check Resources**: Review provided documentation links and references
- **Validate Steps**: Test recommended solutions in safe environments first

## Role Development

### Adding New Roles

1. **Define Role Type**: Add to `RoleType` constants
2. **Create Prompt Template**: Add detailed template to `RolePromptTemplate`
3. **Update Validation**: Include in `ValidateRole()` function
4. **Document Usage**: Add to this documentation
5. **Test Integration**: Ensure CLI and agent integration works

### Role Template Guidelines

Effective role templates should:
- **Define Clear Expertise**: Specify exact knowledge domains
- **Structure Responses**: Provide consistent formatting guidelines
- **Include Context**: Specify what information to consider
- **Provide Examples**: Show expected response patterns
- **Reference Resources**: Link to relevant documentation

## Advanced Features

### Role Inheritance
Some roles share common patterns and can inherit base behaviors while adding specialized capabilities.

### Dynamic Role Selection
Future enhancements may include automatic role selection based on query analysis and context.

### Role Composition
Complex tasks may benefit from combining multiple role perspectives for comprehensive solutions.

## Troubleshooting

### Common Issues

**Invalid Role Error**: Ensure role name matches exactly (case-sensitive)
**Missing Context**: Some roles require additional context for optimal performance
**Provider Issues**: Verify AI provider is configured and accessible

### Debug Mode
Use verbose logging to see role selection and prompt construction:

```bash
nixai diagnose --role diagnose --log-level debug
```

## Future Enhancements

- **Role Analytics**: Track role effectiveness and usage patterns
- **Custom Roles**: Allow users to define custom role templates
- **Role Chaining**: Enable sequential role application for complex workflows
- **Smart Selection**: Automatic role recommendation based on query analysis

---

*This documentation reflects the current state of the nixai role system. For the latest updates and implementation details, refer to the source code in `internal/ai/roles/` and the broader agent system documentation.*
