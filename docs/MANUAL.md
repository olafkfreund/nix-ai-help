# nixai User Manual

Welcome to **nixai** – your AI-powered NixOS assistant for diagnostics, documentation, and automation from the command line. This manual covers all major features, with real-world usage examples for both beginners and advanced users.

> **Latest Update (June 2025)**: Major improvements to documentation display, enhanced terminal formatting, new subcommands, and comprehensive editor integration. All commands now provide actionable help menus, and direct question functionality is smarter than ever. (And yes, nixai can now explain why your jokes don't compile!)

---

## Table of Contents

- Getting Started
- Direct Question Assistant
- Diagnosing NixOS Issues
- Explaining NixOS and Home Manager Options
- Searching for Packages and Services
- AI-Powered Package Repository Analysis
- System Health Checks
- Multi-Machine Management (Flake-based)
- Configuration Templates & Snippets
- Interactive Mode
- Editor Integration
- Advanced Usage & Tips
- Shell Integration
- Command Reference (with real-life examples)
- FAQ & Troubleshooting
- 🦙 llamacpp Provider (Local, Fast, Open Source)

---

## Getting Started

### Prerequisites

- Nix (with flakes enabled)
- Go (if developing outside Nix shell)
- just (for development tasks)
- Ollama (for local LLM inference, recommended)
- git

### Basic Setup

```sh
# Ask questions directly by providing them as arguments
nixai "how do I enable SSH in NixOS?"
nixai "what is a Nix flake?"
nixai "how to configure services.postgresql in NixOS?"

# Alternative: use the --ask or -a flag
nixai --ask "how do I update packages in NixOS?"
nixai -a "what are NixOS generations?"
```

#### Real-World Scenarios

- **New NixOS User:**

  ```sh
  nixai "I just installed NixOS and need to enable SSH access for remote work. How do I do this securely?"
  ```

- **Developer Environment:**

  ```sh
  nixai "I'm a Python developer who needs Docker, PostgreSQL, and VS Code on NixOS. What's the best way to set this up?"
  ```

- **Troubleshooting:**

  ```sh
  nixai "My nginx service keeps failing to start after I added SSL configuration. How do I debug this?"
  ```

#### Pro Tips

- Be specific: "How do I enable hardware acceleration for video playback on NixOS 23.11 with NVIDIA?"
- Include your setup: "I have services.xserver.enable = true but want to switch to Wayland with GNOME."
- Ask for comparisons: "nix-shell vs nix develop vs devenv for Python?"
- Request workflows: "Show me the complete process to set up a Rust dev environment with cross-compilation."

---

## Command Reference & Real-Life Examples

### Direct Question Assistant

Ask anything about NixOS, Home Manager, packaging, or troubleshooting:

```sh
nixai "How do I enable SSH in NixOS?"
nixai --ask "How do I update packages in NixOS?"
```

- **Tip:** Use quotes for multi-word questions. Both direct and --ask/-a flag work identically.

### Diagnose NixOS Issues

Analyze logs, configs, or `nix log` output:

```sh
nixai diagnose --log-file /var/log/nixos/nixos-rebuild.log
nixai diagnose --nix-log /nix/store/xxxx.drv
echo 'services.nginx.enable = true;' | nixai diagnose
```

- **Tip:** Pipe logs directly for instant analysis.

### Explain NixOS Options

Get detailed, AI-powered explanations for any option:

```sh
nixai explain-option services.nginx.enable
nixai explain-option "how to enable SSH access"
```

- **Tip:** Works for both exact option names and natural language queries.

### Explain Home Manager Options

```sh
nixai explain-home-option programs.git.enable
```

- **Tip:** Use for user-level configuration options.

### Search for Packages or Services

```sh
nixai search pkg nginx
nixai search service postgresql
```

- **Tip:** Shows all available options, config snippets, and best practices.

### AI-Powered Package Repository Analysis

```sh
nixai package-repo . --local
nixai package-repo https://github.com/user/project
nixai package-repo . --analyze-only
nixai package-repo https://github.com/user/rust-app --output ./derivations --name my-package
```

- **Tip:** Works for Go, Python, Node.js, Rust, and more. Use --analyze-only to preview.

### System Health Checks

```sh
nixai health
nixai health --nixos-path ~/.config/nixos
nixai health --log-level debug
```

- **Tip:** Use before/after upgrades or for daily maintenance.

### Multi-Machine Management (Flake-based)

nixai provides powerful flake-based machine management using your `nixosConfigurations` from `flake.nix`, with integrated deploy-rs support for streamlined deployments.

#### 🔍 Discovering and Listing Hosts

```zsh
# List all hosts from your flake.nix
nixai machines list

# Automatically discovers hosts from ~/.config/nixos/flake.nix
# Example output: ["dex5550", "p510", "p620", "razer"]
```

#### 🚀 Deployment Options

**Traditional nixos-rebuild (Default):**

```zsh
# Deploy to a specific host
nixai machines deploy --machine hostname

# Deploy with custom target
nixai machines deploy --machine hostname --target-host user@remote-host
```

**Deploy-rs Integration (Recommended):**

```zsh
# Interactive setup with prompts for SSH details
nixai machines setup-deploy-rs

# Non-interactive setup with defaults
nixai machines setup-deploy-rs --non-interactive

# Deploy specific host with deploy-rs
nixai machines deploy --method deploy-rs --machine hostname

# Deploy all hosts (parallel deployment)
nixai machines deploy --method deploy-rs

# Dry run to check configuration
nixai machines deploy --method deploy-rs --dry-run
```

#### 🛠️ Deploy-rs Configuration

The `setup-deploy-rs` command automatically:

1. **Adds deploy-rs input** to your flake.nix
2. **Discovers all hosts** from nixosConfigurations  
3. **Prompts for SSH details** (hostnames, users) per host
4. **Generates deploy configuration** in your flake outputs
5. **Creates deploy nodes** for each host with proper settings

#### 📋 Requirements

- **flake.nix** with `nixosConfigurations` defining your hosts
- **SSH access** configured for remote hosts (for deploy-rs)
- **deploy-rs** added as flake input (automated by setup command)

#### 💡 Best Practices

- **Use deploy-rs** for production deployments (more robust, parallel support)
- **Test with --dry-run** before deploying to production systems
- **Use meaningful hostnames** in nixosConfigurations that match your network
- **Keep SSH configurations** properly maintained for reliable deployments

For comprehensive setup details, see `docs/FLAKE_INTEGRATION_GUIDE.md`.

### Configuration Templates & Snippets

Browse, apply, and manage templates/snippets:

```sh
nixai templates list
nixai templates search gaming
nixai templates apply gaming-nvidia
nixai snippets list
nixai snippets add my-nvidia-config --file /etc/nixos/hardware.nix
```

- **Tip:** Use --merge and --backup when applying templates.

### Interactive Mode

Conversational shell for guided assistance:

```sh
nixai interactive
```

- **Tip:** Type 'help' for available commands. Tab completion is supported!

### Editor Integration

- **Neovim:** `nixai neovim-setup` (see docs/neovim-integration.md)
- **VS Code:** See docs/MCP_VSCODE_INTEGRATION.md

### MCP Server Management

```sh
nixai mcp-server start
nixai mcp-server status
nixai mcp-server stop
```

### MCP Server Background/Daemon Mode

You can start the MCP server in the background (daemon mode) so it continues running independently of your terminal session:

```sh
nixai mcp-server start --background
nixai mcp-server start -d
nixai mcp-server start --daemon  # alias for --background
```

- Use `nixai mcp-server status` to check if the server is running.
- Use `nixai mcp-server stop` to stop the background server.
- All background/daemon flags are equivalent.

### Development Environment (devenv)

Create and manage dev environments:

```sh
nixai devenv list
nixai devenv create python demo-py --framework fastapi --with-poetry --services postgres,redis
nixai devenv suggest "web app with database and REST API"
```

### Flake & Store Management

```sh
nixai flake explain --flake /etc/nixos/flake.nix
nixai store backup
nixai store restore my-backup.tar.gz
```

---

## 🦙 llamacpp Provider (Local, Fast, Open Source)

llamacpp is supported as a local AI provider for privacy and speed. You can use any compatible model served by llamacpp's HTTP API.

### Configuration Example

```yaml
ai_provider: llamacpp
ai_model: llama-2-7b-chat
```

Set the endpoint for llamacpp via environment variable:

```sh
export LLAMACPP_ENDPOINT="http://localhost:8080/completion"
```

If unset, the default endpoint is `http://localhost:8080/completion`.

### Home Manager Example

```nix
services.nixai = {
  enable = true;
  mcp.enable = true;
  mcp.aiProvider = "llamacpp";
  mcp.aiModel = "llama-2-7b-chat";
  mcp.documentationSources = [ "https://wiki.nixos.org/wiki/NixOS_Wiki" ];
};
```

### CLI Usage Example

```sh
nixai --provider llamacpp "How do I enable SSH in NixOS?"
```

### Troubleshooting
- Ensure your llamacpp server is running and accessible at the configured endpoint.
- You can use any model supported by your llamacpp build.
- For best results, use a chat-optimized model (e.g., llama-2-7b-chat).
- If you get connection errors, check the `LLAMACPP_ENDPOINT` value and that the server is reachable.

---

## 🐚 Shell Integration & Tips

### Quick Access Alias

Add to your `.zshrc` or `.bashrc`:

```sh
alias nxai='nixai'
```

### Automatic Error Decoding

Pipe errors or logs for instant help:

```sh
journalctl -xef | nixai diagnose
nix build .#mypkg 2>&1 | nixai diagnose
```

### Real-Time Monitoring

```sh
nixai store health --watch
```

### Tab Completion

```sh
nixai completion zsh > ~/.nixai-completion.zsh
echo "source ~/.nixai-completion.zsh" >> ~/.zshrc
```

---

## Tips, Tricks, and a Joke

- Use `nixai --help` or `nixai <command> --help` for detailed help and examples for every command.
- Combine `nixai` with pipes, files, or interactive mode for maximum flexibility.
- Integrate with your editor for in-place explanations and diagnostics.
- **Pro tip:** If you ever get a cryptic error, just ask: `nixai "What does this error mean?"` – nixai loves a good mystery!
- **Joke:** Why did the NixOS user refuse to cross the road? Because the other side had an imperative configuration!

---

## FAQ & Troubleshooting

- **Q:** nixai says "MCP server unavailable"?
  - **A:** Start it with `nixai mcp-server start`.
- **Q:** How do I change AI provider?
  - **A:** `nixai config set ai_provider openai` (or ollama/gemini/llamacpp)
- **Q:** How do I get more help?
  - **A:** `nixai help` or `nixai <command> --help` for all commands.

---

For more advanced scenarios, see the full manual and the docs directory. Happy hacking!
