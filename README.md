# nixai: NixOS AI Assistant

![Build Status](https://github.com/olafkfreund/nix-ai-help/actions/workflows/ci.yaml/badge.svg?branch=main)

---

## 🌟 Slogan

**nixai: Your AI-powered, privacy-first NixOS assistant with 24+ specialized commands — automate, troubleshoot, and master NixOS from your terminal with intelligent agents and role-based expertise.**

## 🆕 Latest Updates (June 2025)

### ✨ **Modern TUI Interface** - **NEWLY COMPLETED!**
- **🎨 Icon-Free Design**: 100% accessible text-based interface without Unicode dependencies
- **📖 Enhanced Typography**: Larger, more readable text with improved visual hierarchy
- **📜 Smart Scrolling**: Text-based scroll indicators with Page Up/Down support
- **ℹ️ Version Display**: nixai version prominently shown in status bar
- **📰 Changelog Popup**: Press F1 to view latest features and updates
- **⌨️ Better Navigation**: Improved keyboard shortcuts and panel switching

### 🚀 **Recent Feature Additions**
- **Modern TUI Architecture**: Complete `/internal/tui/` system with Bubble Tea framework
- **Enhanced Hardware Management**: 6 specialized subcommands for comprehensive system analysis
- **AI Provider Flexibility**: Configuration-driven provider selection with fallback support
- **MCP Integration**: Model Context Protocol server for enhanced documentation queries
- **Developer Tools**: Advanced flake management, package analysis, and development environments

---

## 🎮 Modern TUI Experience

The newly modernized Terminal User Interface provides a professional, accessible experience:

### 🖥️ Interface Layout

```
┌─ Commands (24 total) ──────────────┬─ Execution Panel ─────────────────┐
│                                    │                                   │
│ ask [INPUT]                        │ Welcome to nixai TUI!             │
│   Ask any NixOS question           │ Select a command from the left    │
│ search [INPUT]                     │ panel to get started.            │
│   Search for packages/services     │                                   │
│ explain-option [INPUT]             │ Latest Updates:                   │
│   Explain a NixOS option           │ • Icon-free accessible design    │
│ hardware detect                    │ • Enhanced typography             │
│   Comprehensive hardware analysis  │ • Smart scrolling system         │
│                                    │ • Version display                │
│ (Showing 1-8 of 24)               │ • F1 changelog popup             │
└────────────────────────────────────┴───────────────────────────────────┘
Commands | F1:Changelog | Tab:Switch | ↑↓:Navigate | Enter:Select | nixai v1.2.3
```

### ✨ Key TUI Features

- **🎯 Accessibility-First**: 100% text-based design without Unicode icon dependencies
- **📖 Enhanced Readability**: Larger, bolder text with improved spacing and visual hierarchy
- **🔄 Smooth Navigation**: Arrow keys, Tab switching, Page Up/Down scrolling with indicators
- **📰 Feature Discovery**: F1 popup shows latest updates and changelog
- **⌨️ Keyboard Efficient**: Complete keyboard navigation without mouse requirement
- **🎨 Professional Design**: Clean two-panel layout with status bar and real-time feedback

---

## 📖 User Manual & Command Reference

See the full [nixai User Manual](docs/MANUAL.md) for comprehensive documentation, advanced usage, and real-world examples for every command.

---

## 🚀 Installation

### 📦 Flake-based Installation (Recommended)

**Prerequisites:**

- Nix (with flakes enabled)
- git

**1. Build and run directly:**

```zsh
nix run github:olafkfreund/nix-ai-help -- --help
```

**2. Build from source (Latest Development):**

```zsh
# Clone the repository
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help

# Build with flakes (recommended)
nix build
./result/bin/nixai --help

# Alternative: Standalone build
nix-build standalone-install.nix
./result/bin/nixai --help
```

**3. Install system-wide via flake:**

```zsh
# Clone and install
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help
nix profile install .

# Or install directly from GitHub
nix profile install github:olafkfreund/nix-ai-help
```

**4. Add to your NixOS/Home Manager configuration:**

See the [modules README](modules/README.md) for complete integration examples.

### 🏗️ Traditional Package Installation (Non-flake Users)

**Using callPackage (Most Common):**

```nix
# In your configuration.nix or home.nix
{ config, pkgs, ... }:

let
  nixai = pkgs.callPackage (builtins.fetchGit {
    url = "https://github.com/olafkfreund/nix-ai-help.git";
    ref = "main";
  } + "/package.nix") {};
in {
  environment.systemPackages = [ nixai ];  # For NixOS
  # OR
  home.packages = [ nixai ];  # For Home Manager
}
```

**Using standalone package.nix:**

```zsh
# Clone the repository
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help

# Build using package.nix
nix-build package.nix

# Install the result
nix-env -i ./result
```

**Submit to nixpkgs:**

The `package.nix` is nixpkgs-compliant and ready for submission. To add nixai to the official nixpkgs repository, you can submit a pull request to [NixOS/nixpkgs](https://github.com/NixOS/nixpkgs).

### 🛠️ Development Installation

**Prerequisites:**

- Nix (with flakes enabled)
- Go (if developing outside Nix shell)  
- just (for development tasks)
- git
- Ollama (for local LLM inference, recommended)

**Install Ollama llama3 model:**

```zsh
ollama pull llama3
```

**Alternative: Install LlamaCpp for CPU-only inference:**

```zsh
# Install llamacpp server
nix run nixpkgs#llama-cpp

# Or set environment variable for existing server
export LLAMACPP_ENDPOINT="http://localhost:8080/completion"
```

**Build and run nixai:**

```zsh
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help
just build
./nixai --help
```

**Ask a question instantly with intelligent agents:**

```zsh
nixai -a "How do I enable SSH in NixOS?"
nixai -a "Debug my failing build" --agent diagnose --role troubleshooter
```

### 🎯 Advanced Features at a Glance

- **24+ Specialized Commands**: Complete AI-powered toolkit for all NixOS operations
- **Advanced Hardware Management**: Comprehensive hardware detection, optimization, and driver management
- **Role-Based AI Agents**: Intelligent agents adapt behavior based on context and user-selected roles
- **Multi-Provider AI**: Local Ollama, OpenAI, Gemini, with intelligent fallback and privacy-first defaults
- **MCP Integration**: Model Context Protocol server for enhanced documentation queries

---

## ✨ Key Features

### 🖥️ **Modern Terminal User Interface**

- **📱 Professional TUI Experience**: Beautiful two-panel layout with commands and execution areas
- **🎯 Accessibility-First Design**: 100% text-based interface without icon dependencies
- **📖 Enhanced Typography**: Large, readable text with improved visual hierarchy and spacing
- **🔄 Smart Navigation**: Tab-based panel switching, arrow key navigation, and search functionality
- **📜 Intelligent Scrolling**: Text-based scroll indicators with smooth Page Up/Down support
- **ℹ️ Status Information**: Version display and real-time command execution feedback
- **📰 Feature Discovery**: F1 changelog popup for viewing latest updates and features
- **⌨️ Keyboard Shortcuts**: Comprehensive keyboard navigation without mouse dependency

**TUI Interface Preview:**

```text
┌─ Commands (24 total) ──────────────┬─ Execution Panel ─────────────────┐
│                                    │                                   │
│ ask [INPUT]                        │ Welcome to nixai TUI!             │
│   Ask any NixOS question           │ Select a command from the left    │
│ search [INPUT]                     │ panel to get started.            │
│   Search for packages/services     │                                   │
│ explain-option [INPUT]             │ Latest Updates:                   │
│   Explain a NixOS option           │ • Icon-free accessible design    │
│ hardware detect                    │ • Enhanced typography             │
│   Comprehensive hardware analysis  │ • Smart scrolling system         │
│                                    │ • Version display                │
│ (Showing 1-8 of 24)               │ • F1 changelog popup             │
└────────────────────────────────────┴───────────────────────────────────┘
Commands | F1:Changelog | Tab:Switch | ↑↓:Navigate | Enter:Select | nixai v1.2.3
```

### 🤖 AI-Powered Command System

- **24+ Specialized Commands**: Complete command-line toolkit for all NixOS tasks and operations
- **Intelligent Agent Architecture**: Role-based AI behavior with specialized expertise domains
- **Direct Question Interface**: `nixai -a "your question"` for instant AI-powered assistance
- **Context-Aware Responses**: Commands adapt behavior based on role, context, and system state
- **Multi-Provider AI Support**: Local Ollama (privacy-first), LlamaCpp (CPU-optimized), OpenAI, Gemini with intelligent fallback

### 🩺 System Management & Diagnostics

- **Comprehensive Health Checks**: `nixai doctor` for full system diagnostics and health monitoring
- **Advanced Log Analysis**: AI-powered parsing of systemd logs, build failures, and error messages
- **Configuration Validation**: Detect, analyze, and fix NixOS configuration issues automatically
- **Hardware Detection & Optimization**: `nixai hardware` with 6 specialized subcommands for system analysis
- **Dependency Analysis**: `nixai deps` for configuration dependencies and import chain analysis

### 🔧 Enhanced Hardware Management

- **Comprehensive Hardware Detection**: `nixai hardware detect` for detailed system analysis and component identification
- **Intelligent Optimization**: `nixai hardware optimize` with AI-powered configuration recommendations and performance tuning
- **Driver Management**: `nixai hardware drivers` for automatic driver and firmware configuration and updates
- **Laptop Support**: `nixai hardware laptop` with power management, thermal control, and mobile-specific optimizations
- **Hardware Comparison**: `nixai hardware compare` for current vs optimal settings analysis and recommendations
- **Function Interface**: `nixai hardware function` for advanced hardware function calling and direct system control
- **Performance Monitoring**: Real-time hardware metrics and optimization suggestions

### 🔍 Search & Discovery

- **Multi-Source Search**: `nixai search <query>` across packages, options, and documentation
- **NixOS Options Explorer**: `nixai explain-option <option>` with detailed explanations and examples
- **Home Manager Support**: `nixai explain-home-option <option>` for user-level configurations
- **Documentation Integration**: Query official NixOS docs, wiki, and community resources via MCP
- **Configuration Snippets**: `nixai snippets` for reusable configuration patterns

### 🧩 Development & Package Management

- **Enhanced Flake Management**: `nixai flake` with complete flake lifecycle support
- **Intelligent Package Analysis**: `nixai package-repo <repo>` with language detection and template system
- **Development Environments**: `nixai devenv` for project-specific development shells
- **Build Optimization**: `nixai build` with intelligent error analysis and troubleshooting
- **Store Management**: `nixai store` for Nix store analysis, backup, and optimization
- **Garbage Collection**: `nixai gc` with AI-powered cleanup analysis and recommendations

### 🏠 Configuration & Templates

- **Interactive Configuration**: `nixai configure` for guided NixOS setup and configuration
- **Template Management**: `nixai templates` for reusable configuration templates
- **Configuration Migration**: `nixai migrate` for system upgrades and configuration transitions
- **Multi-Machine Management**: `nixai machines` for flake-based host management and deployment
- **Learning Modules**: `nixai learn` with interactive tutorials and educational content

### 🌐 Community & Collaboration

- **Community Resources**: `nixai community` for NixOS community links, forums, and support channels
- **MCP Server Integration**: `nixai mcp-server` for Model Context Protocol integration with editors and IDEs
- **Neovim Integration**: `nixai neovim-setup` for seamless editor integration with AI-powered assistance
- **Interactive Shell**: `nixai interactive` with modern TUI for guided assistance and command exploration
- **Configuration Sharing**: Community templates and snippets for common use cases
- **Documentation Contributions**: User-contributed guides and best practices

### 🎨 User Experience

- **Beautiful Terminal Output**: Colorized, formatted output with syntax highlighting via glamour
- **Modern TUI Interface**: Professional two-panel layout with real-time command execution
- **Interactive & CLI Modes**: Use interactively via TUI or directly via CLI, with piped input support
- **Progress Indicators**: Real-time feedback during API calls and long-running operations
- **Role & Agent Selection**: `--role` and `--agent` flags for specialized behavior and expertise
- **Intelligent Help System**: Context-aware help and feature discovery with F1 changelog popup
- **Accessibility Features**: Screen reader compatible, keyboard-only navigation, high contrast themes

### 🔒 Privacy & Performance

- **Privacy-First Design**: Defaults to local LLMs (Ollama), with fallback to cloud providers
- **Multiple AI Providers**: Support for Ollama, OpenAI, Gemini, and other LLM providers
- **Modular Architecture**: Clean separation of concerns with testable, maintainable components
- **Production Ready**: Comprehensive error handling, validation, and robust operation

---

## 🧠 AI Provider Management

nixai features a **unified AI provider management system** that eliminates hardcoded model endpoints and provides flexible, configuration-driven AI provider selection.

### ✨ AI Features

- **🔧 Configuration-Driven**: All AI models and providers defined in YAML configuration
- **🔄 Dynamic Provider Selection**: Switch between providers at runtime
- **🛡️ Automatic Fallbacks**: Graceful degradation when providers are unavailable
- **🔒 Privacy-First**: Defaults to local Ollama with optional cloud provider fallbacks
- **⚡ Zero-Code Model Addition**: Add new models through configuration, not code changes

### 🎯 Supported Providers

| Provider | Models | Capabilities |
|----------|--------|-------------|
| **Ollama** (Default) | llama3, codestral, custom | Local inference, privacy-first |
| **LlamaCpp** | llama-2-7b-chat, custom models | CPU-optimized local inference |
| **Google Gemini** | gemini-2.5-pro, gemini-2.0, gemini-flash | Advanced reasoning, multimodal |
| **OpenAI** | gpt-4o, gpt-4-turbo, gpt-3.5-turbo | Industry-leading performance |
| **Custom** | User-defined | Bring your own endpoint |

### ⚙️ Configuration

All AI provider settings are managed through `configs/default.yaml`:

```yaml
ai:
  provider: "gemini"                    # Default provider
  model: "gemini-2.5-pro"              # Default model
  fallback_provider: "ollama"          # Fallback if primary fails
  
  providers:
    gemini:
      base_url: "https://generativelanguage.googleapis.com/v1beta"
      api_key_env: "GEMINI_API_KEY"
      models:
        gemini-2.5-pro:
          endpoint: "/models/gemini-2.5-pro-latest:generateContent"
          display_name: "Gemini 2.5 Pro (Latest)"
          capabilities: ["text", "code", "reasoning"]
          context_limit: 1000000
    
    ollama:
      base_url: "http://localhost:11434"
      models:
        llama3:
          model_name: "llama3"
          display_name: "Llama 3 (8B)"
          capabilities: ["text", "code"]
    
    llamacpp:
      base_url: "http://localhost:8080"
      env_var: "LLAMACPP_ENDPOINT"
      models:
        llama-2-7b-chat:
          name: "Llama 2 7B Chat"
          display_name: "CPU-optimized Llama 2"
          capabilities: ["text", "code"]
          context_limit: 4096
```

### 🚀 Usage Examples

**Using default configured provider:**

```zsh
nixai -a "How do I configure Nginx in NixOS?"
```

**Using LlamaCpp provider:**

```zsh
# Set LlamaCpp as default provider
ai_provider: llamacpp
ai_model: llama-2-7b-chat

# Configure custom endpoint via environment variable
export LLAMACPP_ENDPOINT="http://localhost:8080/completion"
nixai -a "Help me troubleshoot my NixOS build"

# Remote LlamaCpp server
export LLAMACPP_ENDPOINT="http://192.168.1.100:8080/completion"
nixai diagnose --context-file /etc/nixos/configuration.nix
```

**Provider selection (future enhancement):**

```zsh
# These commands are planned for future implementation
nixai --provider openai -a "Complex reasoning task"
nixai --provider ollama -a "Private local assistance"
nixai config set-provider gemini
```

### 🏗️ Architecture

The system includes three core components:

1. **ProviderManager**: Centralized provider instantiation and management
2. **ModelRegistry**: Configuration-driven model lookup and validation  
3. **Legacy Adapter**: Backward compatibility with existing CLI commands

This architecture eliminated 25+ hardcoded switch statements and enables adding new providers through configuration alone.

### 🖥️ LlamaCpp Setup Guide

**LlamaCpp** provides CPU-optimized local inference without requiring GPU hardware, making it perfect for privacy-focused deployments on any hardware.

#### Quick Setup

1. **Install LlamaCpp server:**

```bash
# Using Nix
nix run nixpkgs#llama-cpp

# Using package manager
# Ubuntu/Debian: apt install llama-cpp
# Arch: pacman -S llama.cpp
# macOS: brew install llama.cpp
```

1. **Download a model:**

```bash
# Example: Download Llama 2 7B Chat GGUF model
wget https://huggingface.co/microsoft/DialoGPT-medium/resolve/main/model.gguf
```

1. **Start the server:**

```bash
# Start llamacpp server on default port 8080
llama-server --model model.gguf --host 0.0.0.0 --port 8080

# Advanced options
llama-server --model model.gguf --host localhost --port 8080 \
  --ctx-size 4096 --threads 8 --n-gpu-layers 0
```

1. **Configure nixai:**

```yaml
# In configs/default.yaml
ai_provider: llamacpp
ai_model: llama-2-7b-chat

# Or via environment variable
export LLAMACPP_ENDPOINT="http://localhost:8080/completion"
```

#### Advanced Configuration

**Remote LlamaCpp Server:**

```bash
# Connect to remote llamacpp instance
export LLAMACPP_ENDPOINT="http://192.168.1.100:8080/completion"
nixai -a "Help with NixOS configuration"
```

**Multiple Models:**

```yaml
providers:
  llamacpp:
    base_url: "http://localhost:8080"
    models:
      llama-2-7b-chat:
        name: "Llama 2 7B Chat"
        context_limit: 4096
      codellama-7b:
        name: "Code Llama 7B"
        context_limit: 4096
```

**Health Check:**

```bash
# Test llamacpp connectivity
curl http://localhost:8080/health

# nixai will automatically check health and fallback if needed
nixai doctor  # Includes provider health checks
```

---

## 📝 Common Usage Examples

> For all commands, options, and real-world examples, see the [User Manual](docs/MANUAL.md).

**Launch the modern TUI interface:**

```zsh
nixai interactive                               # Modern TUI with full navigation
nixai interactive --classic                    # Classic text-based interactive mode
```

**Ask questions with intelligent AI assistance:**

```zsh
nixai "How do I enable Bluetooth?"
nixai --ask "What is a Nix flake?" --role system-architect
nixai -a "Debug my failing build" --agent diagnose
```

**System diagnostics and health monitoring:**

```zsh
nixai doctor                                      # Comprehensive system health check
nixai diagnose --context-file /etc/nixos/configuration.nix
nixai logs --role troubleshooter                 # AI-powered log analysis
nixai deps                                       # Analyze configuration dependencies
```

**Hardware detection and optimization:**

```zsh
nixai hardware detect                            # Comprehensive hardware analysis
nixai hardware optimize --dry-run               # Preview optimization recommendations
nixai hardware drivers --auto-install           # Automatic driver configuration
nixai hardware laptop --power-save              # Laptop-specific optimizations
nixai hardware compare                          # Compare current vs optimal settings
```

**Search and discovery:**

```zsh
nixai search nginx                              # Multi-source package search
nixai search networking.firewall.enable --type option
nixai explain-option services.nginx.enable      # Detailed option explanations
nixai explain-home-option programs.neovim.enable
```

**Development and package management:**

```zsh
nixai flake init                                # Initialize new flake project
nixai flake update                              # Update and optimize flake
nixai package-repo https://github.com/user/project
nixai devenv create python                     # Create development environment
nixai build system --role build-specialist     # Enhanced build troubleshooting
```

**Configuration and templates:**

```zsh
nixai configure                                 # Interactive configuration guide
nixai templates list                           # Browse configuration templates
nixai snippets search "graphics"               # Find configuration snippets
nixai migrate channels-to-flakes               # Migration assistance
```

**Multi-machine and deployment:**

```zsh
nixai machines list                             # List configured machines
nixai machines deploy my-machine               # Deploy to specific machine
nixai machines show my-machine --role system-architect
```

**Advanced features:**

```zsh
nixai interactive                               # Launch modern TUI interface
nixai interactive --classic                    # Classic interactive mode
nixai gc analyze                               # AI-powered garbage collection
nixai store backup                             # Nix store management
nixai community                                # Access community resources
nixai learn nix-language                       # Interactive learning modules
nixai mcp-server start                         # Start MCP server for editor integration
```

**TUI Navigation (Interactive Mode):**

```zsh
# In the modern TUI interface:
# ↑↓ arrows: Navigate command list
# Tab: Switch between panels  
# Enter: Select/execute commands
# /: Search commands
# F1: Show changelog and latest features
# Ctrl+C: Exit
```

---

## 🛠️ Development & Contribution

### Development Setup

**Prerequisites:**

- Nix (with flakes enabled)
- Go 1.21+ (if developing outside Nix shell)
- just (for development tasks)
- git
- Ollama (for local LLM inference, recommended)

**Quick Development Start:**

```zsh
# Clone and enter development environment
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help

# Enter development shell with all dependencies
nix develop

# Build and test
just build
just test
just lint

# Run nixai locally
./nixai --help
```

**Alternative Build Methods:**

```zsh
# Build with Nix flakes (recommended)
nix build
./result/bin/nixai --version

# Standalone build
nix-build standalone-install.nix
./result/bin/nixai --help

# Development build with Go
go build -o nixai cmd/nixai/main.go
```

### Development Workflow

- Use `just` for common development tasks (build, test, lint, run)
- All features are covered by comprehensive tests
- Follow the modular architecture patterns in `internal/`
- Use the configuration system in `configs/default.yaml`
- Maintain documentation for new features and commands
- TUI development uses Bubble Tea framework in `internal/tui/`

### Testing & Quality

```zsh
just test                    # Run all tests
just test-coverage          # Generate coverage report
just lint                   # Run linters
just format                 # Format code
just build                  # Build nixai binary
just run                    # Build and run locally
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Update documentation (README.md, docs/MANUAL.md)
5. Ensure all tests pass with `just test`
6. Submit a pull request

For detailed development guidelines, see the [User Manual](docs/MANUAL.md) and individual command documentation in `docs/`.

---

## 📚 More Resources

### 📖 Documentation

- [User Manual & Command Reference](docs/MANUAL.md) - Complete guide to all 24+ commands
- [TUI Usage Guide](docs/TUI_INPUT_COMMANDS_GUIDE.md) - Modern terminal interface guide
- [TUI Modernization Report](docs/TUI_MODERNIZATION_COMPLETION_REPORT.md) - Latest interface improvements
- [Hardware Guide](docs/hardware.md) - Comprehensive hardware detection and optimization
- [Agent Architecture](docs/agents.md) - AI agent system and role-based behavior
- [Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md) - Advanced flake setup and integration

### 🚀 Integration Guides

- [VS Code Integration](docs/MCP_VSCODE_INTEGRATION.md) - Model Context Protocol integration
- [Neovim Integration](docs/neovim-integration.md) - Editor integration and MCP setup
- [Community Resources](docs/community.md) - Community support and contribution guides
- [TUI Modernization Plan](docs/TUI_MODERNIZATION_PROJECT_PLAN.md) - Technical implementation details

### 📋 Examples & References

- [Copy-Paste Examples](docs/COPY_PASTE_EXAMPLES.md) - Ready-to-use configuration examples
- [Flake Quick Reference](docs/FLAKE_QUICK_REFERENCE.md) - Flake management cheat sheet
- [Installation Guide](docs/INSTALLATION.md) - Detailed installation instructions

### 🔧 Command Documentation

Individual command guides available in `docs/`:

- [diagnose.md](docs/diagnose.md) - System diagnostics and troubleshooting
- [hardware.md](docs/hardware.md) - Hardware detection and optimization
- [package-repo.md](docs/package-repo.md) - Repository analysis and packaging
- [machines.md](docs/machines.md) - Multi-machine management
- [learn.md](docs/learn.md) - Interactive learning system
- And many more...

---

## 🔧 Troubleshooting

### Build Issues

If you encounter build issues, try these solutions in order:

**1. Use the recommended flake installation:**

```zsh
nix build                    # Should work with current flake.nix
```

**2. Alternative build method:**

```zsh
nix-build standalone-install.nix    # Standalone build if flake fails
```

**3. Clear Nix cache and rebuild:**

```zsh
nix store gc
nix build --rebuild
```

### Common Issues

- **"go.mod file not found" errors**: Use flake installation method instead of source archives
- **Module import problems**: Ensure you're using the latest version from the main branch
- **Build failures**: Check that your Nix version supports flakes (`nix --version` should be 2.4+)
- **Vendor hash mismatches**: The current vendor hash is `sha256-pGyNwzTkHuOzEDOjmkzx0sfb1jHsqb/1FcojsCGR6CY=`
- **Hardware detection issues**: Ensure you have appropriate permissions for hardware access
- **AI provider failures**: Verify Ollama is running (`ollama list`) or check API keys for cloud providers
- **TUI display issues**: Ensure your terminal supports Unicode and has sufficient size (80x24 minimum)
- **Interactive mode problems**: Try `nixai interactive --classic` for compatibility with older terminals

### Getting Help

1. Check the [User Manual](docs/MANUAL.md) for detailed command documentation
2. Run `nixai doctor` for system diagnostics
3. See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for detailed solutions
4. Use `nixai community` for community support channels
5. Open an issue on GitHub with system details and error messages

### Verification

After installation, verify everything works:

```zsh
nixai --version              # Should show "nixai version 0.1.0"
nixai doctor                 # Run comprehensive health check
nixai hardware detect       # Test hardware detection
nixai -a "test question"     # Test AI functionality
nixai interactive           # Launch modern TUI interface
```

### Latest Features Verification

Test the newly completed TUI modernization:

```zsh
nixai interactive           # Launch modern TUI
# In TUI: Press F1 to view changelog
# In TUI: Use Tab to switch panels
# In TUI: Type / to search commands
# In TUI: Use ↑↓ arrows to navigate
```

---

**For full command help, advanced usage, and troubleshooting, see the [User Manual](docs/MANUAL.md).**
