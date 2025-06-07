# nixai: NixOS AI Assistant

![Build Status](https://github.com/olafkfreund/nix-ai-help/actions/workflows/ci.yaml/badge.svg?branch=main)

---

## 🌟 Slogan

**nixai: Your AI-powered, privacy-first NixOS assistant with 29 specialized functions — automate, troubleshoot, and master NixOS from your terminal with intelligent agents and role-based expertise.**

---

## 📖 User Manual & Command Reference

See the full [nixai User Manual](docs/MANUAL.md) for comprehensive documentation, advanced usage, and real-world examples for every command.

---

## 🚀 Quick Start

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
just build
./nixai --help
```

**Ask a question instantly with intelligent agents:**

```zsh
nixai -a "How do I enable SSH in NixOS?"
nixai -a "Debug my failing build" --agent diagnose --role troubleshooter
```

### 🎯 Advanced Features at a Glance

- **29 Specialized Functions**: Complete AI-powered toolkit for all NixOS operations
- **26+ Intelligent Agents**: Each specialized for specific domains (build, hardware, community, etc.)
- **Role-Based AI**: Agents adapt behavior based on context and user-selected roles
- **Function Calling System**: Type-safe, validated, async execution with progress indicators
- **Multi-Provider AI**: Local Ollama, OpenAI, Gemini, with intelligent fallback

---

## ✨ Key Features

### 🤖 AI-Powered Function System

- **29 Specialized Functions**: Complete function calling system for all NixOS tasks
- **Intelligent Agent Architecture**: 26+ specialized agents with role-based behavior
- **Advanced Function Registry**: Type-safe parameter validation and async execution
- **Direct Questions**: `nixai -a "your question"` for instant AI-powered help
- **Context-Aware Responses**: Agents adapt behavior based on role and context

### 🩺 System Management & Diagnostics

- **Comprehensive Health Checks**: `nixai doctor` for full system diagnostics
- **Log Analysis**: AI-powered parsing of systemd logs, build failures, and error messages
- **Configuration Validation**: Detect and fix NixOS configuration issues
- **Hardware Detection**: `nixai hardware` for system info and hardware-specific configs

### 🔍 Search & Discovery

- **Multi-Source Search**: `nixai search <query>` across packages, options, and documentation
- **NixOS Options**: `nixai explain-option <option>` with detailed explanations
- **Home Manager Options**: `nixai explain-home-option <option>` for user-level configs
- **Documentation Integration**: Query official NixOS docs, wiki, and community resources

### 🧩 Development & Package Management

- **Flake Management**: Complete flake lifecycle with `nixai flake` commands
- **Package Repository Analysis**: `nixai package-repo <repo>` for AI-generated derivations
- **Development Environments**: `nixai devenv` for project-specific dev shells
- **Build System Integration**: Smart build commands with error analysis

### 🏠 Configuration & Templates

- **Home Manager Support**: Dedicated commands for user-level configurations
- **Templates & Snippets**: `nixai templates`, `nixai snippets` for reusable configs
- **Configuration Migration**: `nixai migrate` for system upgrades and transitions
- **Multi-Machine Management**: `nixai machines` for flake-based host management

### 🎨 User Experience

- **Beautiful Terminal Output**: Colorized, formatted output with [glamour](https://github.com/charmbracelet/glamour)
- **Interactive & CLI Modes**: Use interactively or via CLI, with piped input support
- **Progress Indicators**: Real-time feedback during API calls and operations
- **Role & Agent Selection**: `--role` and `--agent` flags for specialized behavior

### 🔒 Privacy & Performance

- **Privacy-First**: Defaults to local LLMs (Ollama), with fallback to cloud providers
- **Multiple AI Providers**: Support for Ollama, OpenAI, Gemini, and other LLM providers
- **Modular Architecture**: Clean separation of concerns with testable components
- **Production Ready**: Comprehensive error handling and validation

---

## 📝 Common Usage Examples

> For all commands, options, and real-world examples, see the [User Manual](docs/MANUAL.md).

**Ask a question with role-based behavior:**

```zsh
nixai "How do I enable Bluetooth?"
nixai --ask "What is a Nix flake?" --role system-architect
nixai -a "Debug my failing build" --agent diagnose
```

**System diagnostics and troubleshooting:**

```zsh
nixai doctor
nixai diagnose --context-file /etc/nixos/configuration.nix
nixai logs --role troubleshooter
```

**Search with multiple providers:**

```zsh
nixai search nginx
nixai search networking.firewall.enable --type option
nixai search "gpu drivers" --agent hardware
```

**Function-based operations:**

```zsh
nixai explain-option services.nginx.enable
nixai explain-home-option programs.neovim.enable
nixai hardware --role hardware-specialist
```

**Advanced agent and role usage:**

```zsh
nixai --agent build-specialist "Why is my derivation failing?"
nixai flake --role flake-expert
nixai community --agent community-guide
```

**Build and package management:**

```zsh
nixai build system
nixai build .#my-machine --agent build-specialist
nixai package-repo https://github.com/user/project --role packager
```

**Multi-machine and template management:**

```zsh
nixai machines list
nixai machines show my-machine --role system-architect
nixai templates --agent configuration-guide
nixai snippets search "graphics"
```

---

## 🛠️ Development & Contribution

- Use `just` for build, test, lint, and run tasks
- All features are covered by tests; see the [User Manual](docs/MANUAL.md) for details
- See `docs/FLAKE_INTEGRATION_GUIDE.md` for flake integration and advanced setup

---

## 📚 More Resources

### 📖 Documentation

- [User Manual & Command Reference](docs/MANUAL.md)
- [Function System Guide](docs/functions.md) - Complete guide to nixai's 29 specialized functions
- [Agent Architecture](docs/agents.md) - AI agent system with 26+ specialized agents
- [Role System](docs/roles.md) - Role-based AI behavior and prompt templates

### 🚀 Integration Guides

- [Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md)
- [VS Code Integration](docs/MCP_VSCODE_INTEGRATION.md)
- [Neovim Integration](docs/neovim-integration.md)

### 📋 Examples & References

- [Copy-Paste Examples](docs/COPY_PASTE_EXAMPLES.md)
- [Flake Quick Reference](docs/FLAKE_QUICK_REFERENCE.md)

---

**For full command help, advanced usage, and troubleshooting, see the [User Manual](docs/MANUAL.md).**
