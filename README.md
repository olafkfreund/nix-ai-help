# nixai: NixOS AI Assistant

![Build Status](https://github.com/olafkfreund/nix-ai-help/actions/workflows/ci.yaml/badge.svg?branch=main)

---

## 🌟 Slogan

**nixai: Your AI-powered, privacy-first NixOS assistant — automate, troubleshoot, and master NixOS from your terminal.**

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

**Ask a question instantly:**

```zsh
nixai -a "How do I enable SSH in NixOS?"
```

---

## ✨ Key Features

- 🤖 **Direct Question Assistant**: `nixai -a "your question"` for instant AI-powered NixOS help
- 🩺 **System Diagnostics**: `nixai doctor` for full health checks and troubleshooting
- 🔍 **Search**: `nixai search <query>` for packages, options, and docs
- 📝 **Explain Options**: `nixai explain-option <option>` and `nixai explain-home-option <option>`
- 🧩 **Flake Management**: `nixai flake` for info, update, check, and init
- 🏠 **Home Manager Support**: Dedicated commands for user-level configs
- 📦 **Package Repo Analysis**: `nixai package-repo <repo>` for AI-powered derivation generation
- 📝 **Templates & Snippets**: `nixai templates`, `nixai snippets` for reusable configs
- 🖥️ **Multi-Machine Management**: `nixai machines` for flake-based host management
- 🎨 **Beautiful Terminal Output**: All output is colorized and formatted with [glamour](https://github.com/charmbracelet/glamour)
- 🛠️ **Interactive & CLI Modes**: Use interactively or via CLI, with piped input support
- 🔒 **Privacy-First**: Defaults to local LLMs (Ollama), with fallback to cloud providers if configured

---

## 📝 Common Usage Examples

> For all commands, options, and real-world examples, see the [User Manual](docs/MANUAL.md).

**Ask a question:**

```zsh
nixai "How do I enable Bluetooth?"
nixai --ask "What is a Nix flake?"
```

**System diagnostics:**

```zsh
nixai doctor
```

**Search for a package or option:**

```zsh
nixai search nginx
nixai search networking.firewall.enable --type option
```

**Explain a NixOS or Home Manager option:**

```zsh
nixai explain-option services.nginx.enable
nixai explain-home-option programs.neovim.enable
```

**Build a system or flake target:**

```zsh
nixai build system
nixai build .#my-machine
```

**Manage machines (flake-based):**

```zsh
nixai machines list
nixai machines show my-machine
```

**Generate a Nix derivation for a repo:**

```zsh
nixai package-repo https://github.com/user/project
```

---

## 🛠️ Development & Contribution

- Use `just` for build, test, lint, and run tasks
- All features are covered by tests; see the [User Manual](docs/MANUAL.md) for details
- See `docs/FLAKE_INTEGRATION_GUIDE.md` for flake integration and advanced setup

---

## 📚 More Resources

- [User Manual & Command Reference](docs/MANUAL.md)
- [Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md)
- [VS Code Integration](docs/MCP_VSCODE_INTEGRATION.md)
- [Neovim Integration](docs/neovim-integration.md)

---

**For full command help, advanced usage, and troubleshooting, see the [User Manual](docs/MANUAL.md).**
