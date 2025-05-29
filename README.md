# nixai: NixOS AI Assistant

![nixai logo](https://nixos.org/logo/nixos-logo-only-hires.png)

---

## 📚 Table of Contents

- [🚀 Project Overview](#-project-overview)
- [✨ Features](#-features)
- [🧩 Flake Input Analysis & AI Explanations](#-flake-input-analysis--ai-explanations)
- [🎨 Terminal Output Formatting](#-terminal-output-formatting)
- [🛠️ Installation & Usage](#%EF%B8%8F-installation--usage)
- [📝 Commands & Usage](#-commands--usage)
- [🗺️ Project Plan](#%EF%B8%8F-project-plan)
- [Configuration](#configuration)
- [Build & Test](#build--test)
- [Where to Find NixOS Build Logs](#where-to-find-nixos-build-logs)
- [Example: Diagnosing a Build Failure](#example-diagnosing-a-build-failure)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [🙏 Acknowledgments](#-acknowledgments)

---

## 🚀 Project Overview

**nixai** is a powerful, console-based Linux application designed to help you solve NixOS configuration problems, create and configure NixOS systems, and diagnose issues—all from the command line. It leverages advanced Large Language Models (LLMs) like Gemini, OpenAI, and Ollama, with a strong preference for local Ollama models to ensure your privacy. nixai integrates an MCP server to query NixOS documentation from multiple official and community sources, and provides interactive and scriptable diagnostics, log parsing, and command execution.

---

## ✨ Features

- Diagnose NixOS issues from logs, config snippets, or `nix log` output.

- Query NixOS documentation from multiple official and community sources.

- Search for Nix packages and services with clean, numbered results.

- Show configuration options for packages/services (integrates with `nixos-option`).

- Specify your NixOS config folder with `--nixos-path`/`-n`.

- Execute and parse local NixOS commands.

- Accept log input via pipe or file.

- User-selectable AI provider (Ollama, Gemini, OpenAI, etc.).

- Interactive and CLI modes.

- Modular, testable, and well-documented Go codebase.

- Privacy-first: prefers local LLMs (Ollama) by default.

- **NEW:** 🧩 **Flake Input Analysis & AI Explanations** — Analyze and explain flake inputs using AI, with upstream README/flake.nix summaries.

- **NEW:** 🎨 **Beautiful Terminal Output** — All Markdown/HTML output is colorized and formatted for readability using [glamour](https://github.com/charmbracelet/glamour) and [termenv](https://github.com/muesli/termenv).

---

## 🚀 What’s New (May 2025)

- **Config Path Awareness Everywhere:** All features now respect the NixOS config path, settable via `--nixos-path`, config file, or interactively. If unset or invalid, you’ll get clear guidance on how to fix it.
- **Automated Service Option Lookup:** When searching for services, nixai now lists all available options for a service using `nixos-option --find services.<name>`, not just the top-level enable flag.
- **Enhanced Error Handling:** If your config path is missing or invalid, nixai will print actionable instructions for setting it (CLI flag, config, or interactive command).
- **More Tests:** New tests cover service option lookup, diagnostics, and error handling for robust reliability.

---

## 🧩 Flake Input Analysis & AI Explanations

Easily analyze your `flake.nix` inputs and get AI-powered explanations for each input, including upstream README and flake.nix summaries. nixai supports both `name.url = ...;` and `name = { url = ...; ... };` forms, robustly handling comments and whitespace.

**Example:**

```sh
nixai flake explain --flake /path/to/flake.nix
```

---

## 🎨 Terminal Output Formatting

All Markdown and HTML output from nixai is rendered as beautiful, colorized terminal output using [glamour](https://github.com/charmbracelet/glamour) and [termenv](https://github.com/muesli/termenv). This makes AI explanations, documentation, and search results easy to read and visually appealing.

- Works in all modern terminals.
- Respects your terminal theme (light/dark).
- Makes complex output (tables, code, links) readable at a glance.

---

## 🛠️ Installation & Usage

### Using Nix (Recommended)

```sh
nix build .#nixai

./result/bin/nixai
```

### Using Go

```sh
go build -o nixai ./cmd/nixai/main.go

./nixai
```

### Common Tasks (with just)

```sh
just build   # Build the application

just run     # Run the application

just test    # Run tests

just lint    # Lint the code

just fmt     # Format the code

just all     # Clean, build, test, and run
```

---

## 📝 Commands & Usage

### Diagnose NixOS Issues

```sh
nixai diagnose --log-file /path/to/logfile

nixai diagnose --nix-log [drv-or-path]   # Analyze the output of `nix log` (optionally with a derivation/path)

nixai diagnose --config-snippet '...'

echo "...log..." | nixai diagnose
```

### Search for Packages or Services

```sh
nixai search pkg <query>

nixai search service <query>
```

- Clean, numbered results.

- Select a result to see config/test options and available `nixos-option` settings.

### Interactive Mode

```sh
nixai interactive
```

- Supports all search and diagnose features.

- Use `set-nixos-path` to specify your config folder interactively.

---

## 📝 How to Use the Latest Features

### Set or Fix Your NixOS Config Path

- **CLI:**

  ```sh
  nixai search --nixos-path /etc/nixos pkg <query>
  ```

- **Config File:**

  Edit `~/.config/nixai/config.yaml` and set `nixos_folder: /etc/nixos`.

- **Interactive:**

  Start with `nixai interactive` and use `set-nixos-path`.

If the path is missing or invalid, nixai will show you exactly how to fix it.

### Automated Service Option Lookup

- When you search for a service (e.g., `nixai search service nginx`), nixai will now display all available options for that service, not just the enable flag. This uses `nixos-option --find` for comprehensive results.

### Error Guidance

- If you run a command and the config path is not set or is invalid, nixai will print a clear error and suggest how to set it.

### Testing

- All new features are covered by tests. Run them with:

  ```sh
  just test
  # or
  go test ./...
  ```

---

## 🗺️ Project Plan

### 1. **Core Architecture**

- Modular Go codebase with clear package boundaries.

- YAML-based configuration (`configs/default.yaml`) for all settings.

- Main entrypoint: `cmd/nixai/main.go`.

### 2. **AI Integration**

- Support for multiple LLM providers: Ollama (local), Gemini, OpenAI, and more.

- User-selectable provider with Ollama as default for privacy.

- All LLM logic in `internal/ai`.

### 3. **Documentation Query (MCP Server)**

- MCP server/client in `internal/mcp`.

- Query NixOS docs from:

  - [NixOS Wiki](https://wiki.nixos.org/wiki/NixOS_Wiki)

  - [nix.dev manual](https://nix.dev/manual/nix)

  - [Nixpkgs Manual](https://nixos.org/manual/nixpkgs/stable/)

  - [Nix Language Manual](https://nix.dev/manual/nix/2.28/language/)

  - [Home Manager Manual](https://nix-community.github.io/home-manager/)

- Always use sources from config.

### 4. **Diagnostics & Log Parsing**

- Log parsing and diagnostics in `internal/nixos`.

- Accept logs via pipe, file, or CLI.

- Execute and parse local NixOS commands.

### 5. **CLI & Interactive Mode**

- All CLI logic in `internal/cli`.

- Interactive and scriptable modes.

- User-friendly command structure.

### 6. **Utilities & Logging**

- Logging via `pkg/logger`, respecting config log level.

- Utility helpers in `pkg/utils`.

### 7. **Testing & Build**

- Use `justfile` for build, test, lint, and run tasks.

- Nix (`flake.nix`) for reproducible builds and dev environments.

- All new features must be testable and documented.

### 8. **Documentation & Contribution**

- Update this `README.md` and code comments for all changes.

- Keep `.github/copilot-instructions.md` up to date.

- Contributions welcome via PR or issue.

---

## Configuration

nixai supports persistent, user-editable configuration for all users. On first run, a config file is created at:

You can edit this file to set your preferred NixOS config folder, AI provider, model, log level, documentation sources, and more. Example contents:

```yaml
ai_provider: ollama
ai_model: llama3
nixos_folder: ~/nixos-config
log_level: info
mcp_server:
  host: localhost
  port: 8080
  documentation_sources:
    - https://wiki.nixos.org/wiki/NixOS_Wiki
    - https://nix.dev/manual/nix
    - https://nixos.org/manual/nixpkgs/stable/
    - https://nix.dev/manual/nix/2.28/language/
    - https://nix-community.github.io/home-manager/
nixos:
  config_path: ~/nixos-config/configuration.nix
  log_path: /var/log/nixos/nixos-rebuild.log
diagnostics:
  enabled: true
  threshold: 1
commands:
  timeout: 30
  retries: 2
```

---

## Build & Test

- Use the `justfile` for common tasks: `just build`, `just test`, etc.

- Nix (`flake.nix`) provides a reproducible dev environment.

---

## Where to Find NixOS Build Logs

- Latest build logs: `/var/log/nixos/nixos-rebuild.log` (system), `/var/log/nix/drvs/` (per-derivation).

- Use `nix log` for recent build failures.

---

## Example: Diagnosing a Build Failure

```sh
sudo nixos-rebuild switch

nixai diagnose --nix-log
```

---

## 🤝 Contributing

- All code must be idiomatic Go, modular, and well-documented.

- Add or update tests for all new features and bugfixes.

- See PROJECT_PLAN.md for roadmap and tasks.

---

## 📄 License

This project is licensed under the MIT License. See the LICENSE file for details.

---

## 🙏 Acknowledgments

- Thanks to the NixOS community and documentation authors.

- Special thanks to the creators of the AI models used in this project.

---

> _nixai: Your AI-powered NixOS assistant, right in your terminal._
