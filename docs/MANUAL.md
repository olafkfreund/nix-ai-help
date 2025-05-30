# nixai User Manual

Welcome to **nixai** – your AI-powered NixOS assistant for diagnostics, documentation, and automation from the command line. This manual covers all major features, with real-world usage examples for both beginners and advanced users.

---

## Table of Contents
- [Getting Started](#getting-started)
- [Diagnosing NixOS Issues](#diagnosing-nixos-issues)
- [Explaining NixOS and Home Manager Options](#explaining-nixos-and-home-manager-options)
- [Searching for Packages and Services](#searching-for-packages-and-services)
- [AI-Powered Package Repository Analysis](#ai-powered-package-repository-analysis)
- [System Health Checks](#system-health-checks)
- [Interactive Mode](#interactive-mode)
- [Advanced Usage](#advanced-usage)
- [Configuration](#configuration)
- [Tips & Troubleshooting](#tips--troubleshooting)

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
# Enter the Nix development environment
nix develop

# Build the application
just build

# Run help
./nixai --help
```

---

## Diagnosing NixOS Issues

nixai can analyze logs, config snippets, or `nix log` output to help you diagnose problems.

### Basic Example: Diagnose a Log File
```sh
nixai diagnose --log-file /var/log/nixos/nixos-rebuild.log
```

### Diagnose a Nix Log
```sh
nixai diagnose --nix-log /nix/store/xxxx.drv
```

### Diagnose a Config Snippet
```sh
echo 'services.nginx.enable = true;' | nixai diagnose
```

---

## Explaining NixOS and Home Manager Options

Get detailed, AI-powered explanations for any option, including type, default, description, best practices, and usage examples.

### NixOS Option
```sh
nixai explain-option services.nginx.enable
```

### Home Manager Option
```sh
nixai explain-home-option programs.git.enable
```

### Natural Language Query
```sh
nixai explain-option "how to enable SSH access"
```

---

## Searching for Packages and Services

Search for Nix packages or services and get clean, numbered results.

### Search for a Package
```sh
nixai search pkg nginx
```

### Search for a Service
```sh
nixai search service postgresql
```

---

## AI-Powered Package Repository Analysis

Automatically analyze a project and generate a Nix derivation using AI.

### Analyze Local Project
```sh
nixai package-repo . --local
```

### Analyze Remote Repository
```sh
nixai package-repo https://github.com/user/project
```

### Analyze Only (No Derivation Generation)
```sh
nixai package-repo . --analyze-only
```

### Advanced: Custom Output and Name
```sh
nixai package-repo https://github.com/user/rust-app --output ./derivations --name my-package
```

---

## System Health Checks

Run a comprehensive health check on your NixOS system.

```sh
nixai health
```

---

## Interactive Mode

Launch an interactive shell for all features:

```sh
nixai interactive
```

- Use `set-nixos-path` to specify your config folder interactively.
- Run any command in a guided, conversational way.

---

## Advanced Usage

### Specify NixOS Config Path
```sh
nixai search --nixos-path /etc/nixos pkg nginx
```

### Use a Different AI Provider
```sh
nixai diagnose --provider openai --log-file /var/log/nixos/nixos-rebuild.log
```

### Get Examples for a Service
```sh
nixai service-examples nginx
```

### Flake Input Analysis & AI Explanations
```sh
nixai flake explain --flake /etc/nixos/flake.nix
```

---

## Configuration

nixai uses a YAML config file (usually at `~/.config/nixai/config.yaml`). You can set:
- Preferred AI provider/model
- NixOS config folder
- Log level
- Documentation sources

Example config:
```yaml
ai_provider: ollama
ai_model: llama3
nixos_folder: ~/nixos-config
log_level: info
```

---

## Tips & Troubleshooting

- If you see errors about missing config path, set it with `--nixos-path` or in your config file.
- For best privacy, use Ollama as your AI provider (local inference).
- Use `just lint` and `just test` for code quality and reliability.
- All features are available in both CLI and interactive modes.

---

> _nixai: Your AI-powered NixOS assistant, right in your terminal._
