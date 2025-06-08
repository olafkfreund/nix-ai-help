# nixai: NixOS AI Assistant

![Build Status](https://github.com/olafkfreund/nix-ai-help/actions/workflows/ci.yaml/badge.svg?branch=main)

---

## 🌟 Slogan

**nixai: Your AI-powered, privacy-first NixOS assistant with 24+ specialized commands — automate, troubleshoot, and master NixOS from your terminal with intelligent agents and role-based expertise.**

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

### 🤖 AI-Powered Command System

- **24+ Specialized Commands**: Complete command-line toolkit for all NixOS tasks and operations
- **Intelligent Agent Architecture**: Role-based AI behavior with specialized expertise domains
- **Direct Question Interface**: `nixai -a "your question"` for instant AI-powered assistance
- **Context-Aware Responses**: Commands adapt behavior based on role, context, and system state
- **Multi-Provider AI Support**: Local Ollama (privacy-first), OpenAI, Gemini with intelligent fallback

### 🩺 System Management & Diagnostics

- **Comprehensive Health Checks**: `nixai doctor` for full system diagnostics and health monitoring
- **Advanced Log Analysis**: AI-powered parsing of systemd logs, build failures, and error messages
- **Configuration Validation**: Detect, analyze, and fix NixOS configuration issues automatically
- **Hardware Detection & Optimization**: `nixai hardware` with 6 specialized subcommands for system analysis
- **Dependency Analysis**: `nixai deps` for configuration dependencies and import chain analysis

### 🔧 Enhanced Hardware Management

- **Comprehensive Hardware Detection**: `nixai hardware detect` for detailed system analysis
- **Intelligent Optimization**: `nixai hardware optimize` with AI-powered configuration recommendations
- **Driver Management**: `nixai hardware drivers` for automatic driver and firmware configuration
- **Laptop Support**: `nixai hardware laptop` with power management and mobile-specific optimizations
- **Hardware Comparison**: `nixai hardware compare` for current vs optimal settings analysis
- **Function Interface**: `nixai hardware function` for advanced hardware function calling

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

- **Community Resources**: `nixai community` for NixOS community links and support channels
- **MCP Server**: `nixai mcp-server` for Model Context Protocol integration with editors
- **Neovim Integration**: `nixai neovim-setup` for seamless editor integration
- **Interactive Shell**: `nixai interactive` for guided assistance and exploration

### 🎨 User Experience

- **Beautiful Terminal Output**: Colorized, formatted output with syntax highlighting via glamour
- **Interactive & CLI Modes**: Use interactively or via CLI, with piped input support
- **Progress Indicators**: Real-time feedback during API calls and long-running operations
- **Role & Agent Selection**: `--role` and `--agent` flags for specialized behavior and expertise
- **TUI Mode**: `--tui` flag for terminal user interface on supported commands

### 🔒 Privacy & Performance

- **Privacy-First Design**: Defaults to local LLMs (Ollama), with fallback to cloud providers
- **Multiple AI Providers**: Support for Ollama, OpenAI, Gemini, and other LLM providers
- **Modular Architecture**: Clean separation of concerns with testable, maintainable components
- **Production Ready**: Comprehensive error handling, validation, and robust operation

---

## 📝 Common Usage Examples

> For all commands, options, and real-world examples, see the [User Manual](docs/MANUAL.md).

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
nixai interactive                               # Launch interactive AI shell
nixai gc analyze                               # AI-powered garbage collection
nixai store backup                             # Nix store management
nixai community                                # Access community resources
nixai learn nix-language                       # Interactive learning modules
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

### Testing & Quality

```zsh
just test                    # Run all tests
just test-coverage          # Generate coverage report
just lint                   # Run linters
just format                 # Format code
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
- [Hardware Guide](docs/hardware.md) - Comprehensive hardware detection and optimization
- [Agent Architecture](docs/agents.md) - AI agent system and role-based behavior
- [Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md) - Advanced flake setup and integration

### 🚀 Integration Guides

- [VS Code Integration](docs/MCP_VSCODE_INTEGRATION.md) - Model Context Protocol integration
- [Neovim Integration](docs/neovim-integration.md) - Editor integration and MCP setup
- [Community Resources](docs/community.md) - Community support and contribution guides

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
```

---

**For full command help, advanced usage, and troubleshooting, see the [User Manual](docs/MANUAL.md).**
