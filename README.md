# nixai: NixOS AI Assistant

![Build Status](https://github.com/olafkfreund/nix-ai-help/actions/workflows/ci.yaml/badge.svg?branch=main)

---

## 🌟 Slogan

**nixai: Your AI-powered, privacy-first NixOS assistant — automate, troubleshoot, and master NixOS from your terminal.**

---

## 📖 User Manual

See the full [nixai User Manual](docs/MANUAL.md) for comprehensive feature documentation, advanced usage, real-world examples, and troubleshooting tips.

---

### This is development code. Things may not work or is broken. I'm changing the code all the time. Don't expect something production ready

## 🚀 What's New (June 2025)

- **🔧 Enhanced Build Troubleshooter:** New `build` command with advanced build failure analysis, intelligent retry mechanisms, cache analysis, sandbox debugging, and performance profiling.
- **🤖 Direct Question Assistant**: Ask questions instantly with `nixai "your question"` or `nixai --ask "question"` for immediate AI-powered NixOS help with documentation context.
- **Config Path Awareness Everywhere:** All features now respect the NixOS config path, settable via `--nixos-path`, config file, or interactively. If unset or invalid, you'll get clear guidance on how to fix it.
- **Automated Service Option Lookup:** When searching for services, nixai now lists all available options for a service using `nixos-option --find services.<n>`, not just the top-level enable flag.
- **Enhanced Error Handling:** If your config path is missing or invalid, nixai will print actionable instructions for setting it (CLI flag, config, or interactive command).
- **🏠 Home Manager vs NixOS Visual Distinction:** Smart detection automatically shows `🖥️ NixOS Option` or `🏠 Home Manager Option` headers with appropriate documentation sources.
- **🆕 Dedicated Home Manager Command:** New `explain-home-option` command specifically for Home Manager configuration options.
- **🆕 AI-Powered Package Repository Analysis:** New `package-repo` command automatically analyzes Git repositories and generates complete Nix derivations using AI-powered build system detection and dependency analysis.
- **📝 Configuration Templates & Snippets:** Browse, apply, and manage curated NixOS configuration templates with `nixai templates` and save/reuse configuration snippets with `nixai snippets`. Includes GitHub search integration for discovering real-world configurations.
- **🖥️ Multi-Machine Management (Flake-based):** Centrally manage, synchronize, and deploy NixOS configurations across multiple machines directly from your `flake.nix`. See below for details and usage examples.
- **More Tests:** New tests cover service option lookup, diagnostics, error handling, and packaging features for robust reliability.
- **🆕 Nix Store Management**: Backup, restore, verify, and analyze the Nix store directly from the CLI.
- **System State Backup & Restore**: Comprehensive backup/restore with validation and incremental support.

## Prerequisites

Before using or developing nixai, ensure you have the following installed:

- **Ollama** (for local LLM inference)
  - You must have the `llama3` model pulled and available in Ollama:

```sh
ollama pull llama3
```

- **Nix** (with flakes enabled)
- **Go** (if developing outside Nix shell)
- **just** (for development tasks)
- **git**
- (Optional) API keys for OpenAI or Gemini if you want to use cloud LLMs

All other dependencies are managed by the Nix flake and justfile.

![nixai logo](./nix-ai2.png)

---

## 📚 Table of Contents

- [🚀 Project Overview](#-project-overview)
- [✨ Features](#-features)
- [🩺 Doctor Command: System Diagnostics & Troubleshooting](#-doctor-command-system-diagnostics--troubleshooting)
- [🆕 Development Environment (devenv) Integration](#-development-environment-devenv-integration)
- [🧩 Flake Input Analysis & AI Explanations](#-flake-input-analysis--ai-explanations)
- [🔧 NixOS Option Explainer](#-nixos-option-explainer)
- [📦 AI-Powered Package Repository Analysis](#-ai-powered-package-repository-analysis)
- [🔄 MCP Server Configuration & Autostart](#-mcp-server-configuration--autostart)
- [🔄 Migration Assistant (Channels ↔ Flakes)](#-migration-assistant-channels--flakes)
- [🎨 Terminal Output Formatting](#-terminal-output-formatting)
- [🖥️ Multi-Machine Management (Flake-based)](#-multi-machine-management-flake-based)
- [🛠️ Installation & Usage](#%EF%B8%8F-installation--usage)
- [📚 Flake Integration Guide](#-flake-integration-guide)
- [📝 Commands & Usage](#-commands--usage)
- [🩺 Doctor Command: Diagnose MCP Server & AI Providers](#-doctor-command-diagnose-mcp-server--ai-providers)
- [🗺️ Project Plan](#%EF%B8%8F-project-plan)
- [Configuration](#configuration)
- [Build & Test](#build--test)
- [Where to Find NixOS Build Logs](#where-to-find-nixos-build-logs)
- [Example: Diagnosing a Build Failure](#example-diagnosing-a-build-failure)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [🙏 Acknowledgments](#-acknowledgments)
- [📸 Screenshots](#-screenshots)
- [🐚 Shell Integration: Always-On nixai Assistant](#-shell-integration-always-on-nixai-assistant)

---

## 🚀 Project Overview

**nixai** is your all-in-one, AI-powered command-line assistant for NixOS. It helps you automate, troubleshoot, and optimize every aspect of your NixOS experience — from configuration and deployment to diagnostics and learning. Powered by advanced Large Language Models (LLMs) like Ollama, Gemini, and OpenAI (with a privacy-first local default), nixai brings instant, context-aware help, beautiful documentation, and powerful automation to your terminal. 

With nixai, you can:
- Ask natural language questions and get actionable, AI-driven answers with official documentation context
- Register, manage, and deploy to fleets of NixOS machines from a single CLI
- Analyze, explain, and optimize your NixOS configs, flakes, and services
- Diagnose errors, parse logs, and get step-by-step troubleshooting
- Search, lint, and generate Nix code for packages, services, and repositories
- Use interactive or scriptable modes for both beginners and power users

nixai is designed for privacy, productivity, and beautiful terminal output. Whether you're a NixOS newcomer or a seasoned sysadmin, nixai helps you get more done, faster — with confidence.

---

## ✨ Features

- **🤖 Direct Question Assistant**: Ask questions directly with `nixai "your question"` or `nixai --ask "question"` for instant AI-powered NixOS help.

- Diagnose NixOS issues from logs, config snippets, or `nix log` output.

- Query NixOS documentation from multiple official and community sources.

- Search for Nix packages and services with clean, numbered results.

- Show configuration options for packages/services (integrates with `nixos-option`).

- **System Health Check**: Run comprehensive NixOS system health checks with AI-powered analysis.

- Specify your NixOS config folder with `--nixos-path`/`-n`.

- Execute and parse local NixOS commands.

- Accept log input via pipe or file.

- User-selectable AI provider (Ollama, Gemini, OpenAI, etc.).

- Interactive and CLI modes.

- Modular, testable, and well-documented Go codebase.

- Privacy-first: prefers local LLMs (Ollama) by default.

- **NEW:** 🧩 **Flake Input Analysis & AI Explanations** — Analyze and explain flake inputs using AI, with upstream README/flake.nix summaries.

- **NEW:** 🎨 **Beautiful Terminal Output** — All Markdown/HTML output is colorized and formatted for readability using [glamour](https://github.com/charmbracelet/glamour) and [termenv](https://github.com/muesli/termenv).

- **NEW:** ✅ **AI-Powered NixOS Option Explainer** — Get detailed, AI-generated explanations for any NixOS option with `nixai explain-option <option>`, including type, default, description, and best practices.

- **NEW:** 🏠 **Home Manager Option Support** — Dedicated `nixai explain-home-option <option>` command with visual distinction between Home Manager and NixOS options.

- **NEW:** 📦 **AI-Powered Package Repository Analysis** — Automatically analyze Git repositories and generate Nix derivations with `nixai package-repo <path>`, supporting Go, Python, Node.js, and Rust projects.

- **NEW:** 📝 **Configuration Templates & Snippets** — Browse, apply, and manage curated NixOS configuration templates with `nixai templates` and save/reuse configuration snippets with `nixai snippets`. Includes GitHub search integration for discovering real-world configurations.

- **NEW:** 🖥️ **Multi-Machine Management (Flake-based)** — Centrally manage, group, and deploy NixOS configurations to multiple machines directly from your `flake.nix`. Includes machine registry, group management, configuration sync, deployment, diff analysis, and status monitoring.

- **NEW:** 🆕 **Nix Store Management** — Backup, restore, verify, and analyze the Nix store directly from the CLI.

- **NEW:** **System State Backup & Restore** — Comprehensive backup/restore with validation and incremental support.

---

## 🩺 Doctor Command: System Diagnostics & Troubleshooting

The `nixai doctor` command provides a comprehensive diagnostics report for your nixai setup, including:
- **MCP Server Health**: Checks if the documentation server is running, healthy, and accessible (port/socket, process, /healthz endpoint)
- **AI Provider Health**: Verifies connectivity and configuration for Ollama, OpenAI, and Gemini (API reachability, key presence/validity)
- **Actionable Feedback**: Glamour-formatted output with clear next steps for resolving common issues

### Usage Example

```sh
nixai doctor
```

**Example Output:**

```
🩺 nixai Doctor: System Diagnostics
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

MCP Server Diagnostics
━━━━━━━━━━━━━━━━━━━━━━
✅ MCP server is running and healthy on http://localhost:8081.
✅ Port is open: localhost:8081
✅ MCP server process is running (pgrep matched).

AI Provider Diagnostics
━━━━━━━━━━━━━━━━━━━━━━━
✅ Ollama API reachable at http://localhost:11434
✅ OpenAI API reachable (key valid).
✅ Gemini API reachable (key valid).

ℹ️  See the README or docs/MANUAL.md for troubleshooting steps and more details.
```

If any issues are detected, nixai doctor will provide warnings, errors, and actionable suggestions (e.g., how to start the MCP server, set API keys, or check service status).

---

## 🩺 Doctor Command: Diagnose MCP Server & AI Providers

The `nixai doctor` command provides a comprehensive diagnostic checkup for your nixai environment. It helps you quickly identify and resolve common setup and runtime problems with the MCP server and all supported AI providers (Ollama, OpenAI, Gemini).

### What It Does
- Checks if the MCP server is running, healthy, and accessible (including port/socket and process status)
- Diagnoses connectivity and configuration for Ollama, OpenAI, and Gemini AI providers
- Detects missing or invalid API keys, unreachable endpoints, and common misconfigurations
- Provides actionable, glamour-formatted feedback and next steps for troubleshooting

### Usage

```zsh
nixai doctor
```

**Example Output:**

```
🩺 nixai Doctor: System Diagnostics
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

MCP Server Diagnostics
━━━━━━━━━━━━━━━━━━━━━━
✅ MCP server is running and healthy on http://localhost:8081.
✅ Port is open: localhost:8081
✅ MCP server process is running (pgrep matched).

AI Provider Diagnostics
━━━━━━━━━━━━━━━━━━━━━━━
✅ Ollama API reachable at http://localhost:11434
✅ OpenAI API reachable (key valid).
✅ Gemini API reachable (key valid).

ℹ️  See the README or docs/MANUAL.md for troubleshooting steps and more details.
```

### Troubleshooting
- If any check fails, the output will include clear error or warning messages and suggested next steps (e.g., start the MCP server, set API keys, check network connectivity).
- For MCP server issues, try restarting with `nixai mcp-server start -d` or check your configuration in `~/.config/nixai/config.yaml`.
- For AI provider issues, ensure the relevant API keys are set in your environment and the endpoints are reachable.

### When to Use
- After installation or upgrade
- When experiencing issues with documentation queries or AI-powered features
- To verify your environment before troubleshooting more complex problems

---

## 🆕 Development Environment (devenv) Integration

nixai now includes a powerful `devenv` feature for quickly scaffolding reproducible development environments for popular languages (Python, Rust, Node.js, Go) using [devenv.sh](https://devenv.sh/) and Nix. This system is:

- **Extensible**: Add new language/framework templates easily in Go
- **Flexible**: Supports framework/tool options (e.g., Gin, FastAPI, TypeScript, gRPC, etc.)
- **Service-aware**: Integrates databases/services (Postgres, Redis, MySQL, MongoDB)
- **AI-powered**: Suggests templates based on your project description

### Usage Examples

- **List templates:**

  ```sh
  nixai devenv list
  ```

- **Create a project:**

  ```sh
  nixai devenv create python myproject --framework fastapi --with-poetry --services postgres,redis
  nixai devenv create golang my-go-app --framework gin --with-grpc
  nixai devenv create nodejs my-node-app --with-typescript --services mongodb
  nixai devenv create rust my-rust-app --with-wasm
  ```

- **Get AI-powered suggestions:**

  ```sh
  nixai devenv suggest "web app with database and REST API"
  ```

### How to Add a New Language Template

1. Edit `internal/devenv/builtin_templates.go` and implement the `Template` interface (see existing templates for examples).
2. Register your template in `registerBuiltinTemplates()` in `service.go`.
3. Add or update tests in `service_test.go`.
4. Document your new template in the main README and manual.

See `internal/devenv/README.md` for a full developer guide.

---

## 🔄 MCP Server Configuration & Autostart

The nixai Model Context Protocol (MCP) server provides NixOS documentation and option queries to enhance AI responses. You can configure how the server runs and automatically start it on boot:

### Socket Path Configuration

By default, the MCP server uses `/tmp/nixai-mcp.sock` as the Unix domain socket path. You can customize this path using:

- **Command-line flag**: `nixai mcp-server start --socket-path="/path/to/socket"`
- **Environment variable**: `NIXAI_SOCKET_PATH="/path/to/socket" nixai mcp-server start`
- **NixOS/Home Manager module**: Set the `socketPath` option in your configuration

### Autostart Options

The MCP server can be configured to start automatically on boot using either system-wide or user-level services:

#### NixOS Module (System-wide)

Add the nixai NixOS module to your configuration:

```nix
# configuration.nix
{ config, pkgs, ... }:

{
  imports = [ 
    # Path to the nixai flake or local module
    (builtins.fetchTarball "https://github.com/olafkfreund/nix-ai-help/archive/main.tar.gz")/modules/nixos.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # Optional: custom socket path
      socketPath = "/run/nixai/mcp.sock";
    };
  };
}
```

#### Home Manager Module (User-level)

Add the nixai Home Manager module to your configuration:

```nix
# home.nix
{ config, pkgs, ... }:

{
  imports = [
    # Path to the nixai flake or local module
    (builtins.fetchTarball "https://github.com/olafkfreund/nix-ai-help/archive/main.tar.gz")/modules/home-manager.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # Optional: custom socket path (uses `$HOME` expansion)
      socketPath = "$HOME/.local/share/nixai/mcp.sock";
    };
    # Optional: integrate with VS Code
    vscodeIntegration = true;
  };
}
```

#### Using Flakes

> **Note:** When using flakes, nixai modules are exported per system architecture. Reference them as `nixai.nixosModules.x86_64-linux.default` (or your system, e.g. `aarch64-linux`) instead of just `nixai.nixosModules.default`.

```nix

{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      # ...
      modules = [
        nixai.nixosModules.x86_64-linux.default # <-- Use your system here
        {
          services.nixai = {
            enable = true;
            mcp.enable = true;
          };
        }
      ];
    };

    homeConfigurations.yourusername = home-manager.lib.homeManagerConfiguration {
      # ...
      modules = [
        nixai.homeManagerModules.x86_64-linux.default # <-- Use your system here
        {
          services.nixai = {
            enable = true;
            mcp.enable = true;
          };
        }
      ];
    };
  };
}
```

**Troubleshooting:**

- If you see `error: attribute 'default' missing`, reference the module as `nixai.nixosModules.<system>.default` (e.g., `nixai.nixosModules.x86_64-linux.default`).

- If you see `error: attribute 'mdDoc' missing`, update nixai to the latest version. The module now uses a plain string for the description and is compatible with all recent Nixpkgs.

See [Autostart Options Documentation](docs/autostart-options.md) for more detailed examples and troubleshooting.

---

## MCP Server Features

The MCP (Model Context Protocol) server provides advanced documentation and option explanation services for NixOS and Home Manager. Recent features include:

- **Hot-reload of configuration**: The MCP server watches its YAML config file and reloads documentation sources and settings at runtime. No restart required for config changes.
- **Structured logging**: All server logs use the `pkg/logger` structured logger and respect the configured log level (set in config YAML). Debug/info/error logs are consistent and testable.
- **/healthz endpoint**: Returns server health and uptime in JSON format.
- **/metrics endpoint**: Prometheus-compatible endpoint for metrics (uptime, request counters, etc.).
- **Graceful shutdown**: `/shutdown` endpoint and signal handling for clean server exit, with clear log output.
- **Testable logger**: Logger can be injected with a custom writer for integration tests.

### Example: Hot-reload Config

1. Edit your MCP config YAML (e.g. `~/.config/nixai/config.yaml`).
2. The server will detect changes and reload documentation sources automatically. Logs will show reload events.

### Example: Structured Logging

```text
INFO[2025-06-02T12:34:56Z] Starting MCP server | addr=127.0.0.1:8081
INFO[2025-06-02T12:35:01Z] Config file changed, reloading...
INFO[2025-06-02T12:35:01Z] Reloaded documentation sources from config.
```

### Endpoints

- `GET/POST /query` — Query NixOS/Home Manager docs (see below)
- `GET /healthz` — Health and uptime
- `GET /metrics` — Prometheus metrics
- `POST /shutdown` — Graceful shutdown

### Testing

All features are covered by integration tests, including logger output capture.

---

## 🔄 Migration Assistant (Channels ↔ Flakes)

nixai includes a comprehensive migration assistant to help you convert your NixOS configuration between legacy channels and modern flakes, with safety checks, backup/rollback, and AI-powered guidance.

**Key Features:**

- Analyze your current setup and migration complexity
- Step-by-step migration guide with best practices
- Automatic backup and rollback procedures
- Validation of migration success
- AI-powered explanations and troubleshooting

**Usage Examples:**

```sh
nixai migrate analyze --nixos-path /etc/nixos
```

```sh
nixai migrate to-flakes --nixos-path /etc/nixos
```

- The assistant will walk you through the migration, create a backup, and validate the result.
- All output is formatted with glamour for easy reading.
- If anything goes wrong, you can roll back to your previous configuration.

**Planned:**

- `nixai migrate from-flakes` (convert back to channels)
- `nixai migrate channel-upgrade` (safe channel upgrades)
- `nixai migrate flake-inputs` (update/explain flake inputs)

See the [User Manual](docs/MANUAL.md#migration-assistant) for full details and troubleshooting.

---

## ✨ Latest Features & Updates

### Recent Fixes & Improvements (May 2025)

- **✅ HTML Filtering for Clean Documentation:** The `explain-home-option` and `explain-option` commands now properly filter out HTML tags, wiki navigation elements, DOCTYPE declarations, and raw content, providing clean, formatted output with beautiful markdown rendering via glamour.
- **🎨 Enhanced Terminal Output Formatting:** All documentation output uses consistent formatting with headers, dividers, key-value pairs, and glamour markdown rendering for improved readability across all terminal environments.
- **🔧 Robust Error Handling:** Better error messages, graceful handling when MCP server is unavailable, improved timeout handling, and clear feedback for configuration issues.
- **🏠 Home Manager Option Support:** Dedicated `explain-home-option` command with smart visual distinction between Home Manager and NixOS options, including proper source detection and documentation filtering.
- **🔌 Full Editor Integration:** Complete support for Neovim and VS Code with MCP server integration for seamless in-editor NixOS assistance, including automatic setup commands and configuration generators.
- **🧪 Comprehensive Testing:** All HTML filtering, documentation display, and error handling improvements are backed by extensive test coverage to ensure reliability.

### Core Features

- **🤖 Direct Question Assistant:** Ask questions directly with `nixai "your question"` or `nixai --ask "question"` for instant AI-powered NixOS help with documentation context.
- **🎯 Config Path Awareness Everywhere:** All features respect the NixOS config path, settable via `--nixos-path`, config file, or interactively. If unset or invalid, you'll get clear guidance on how to fix it.
- **🔍 Automated Service Option Lookup:** When searching for services, nixai lists all available options for a service using `nixos-option --find services.<name>`, not just the top-level enable flag.
- **📦 AI-Powered Package Repository Analysis:** Analyze Git repositories and generate complete Nix derivations with support for Go, Python, Node.js, and Rust projects.
- **📝 Configuration Templates & Snippets:** Browse, apply, and manage curated NixOS configuration templates with `nixai templates` and save/reuse configuration snippets with `nixai snippets`. Includes GitHub search integration for discovering real-world configurations.
- **🧩 Flake Input Analysis:** Analyze and explain flake inputs using AI, with upstream README/flake.nix summaries.
- **🏥 System Health Checks:** Run comprehensive NixOS system health checks with AI-powered analysis and recommendations.
- **✅ Comprehensive Test Coverage:** Extensive test coverage for service option lookup, diagnostics, error handling, packaging features, and HTML filtering for robust reliability.
- **🆕 Nix Store Management**: Backup, restore, verify, and analyze the Nix store directly from the CLI.
- **System State Backup & Restore**: Comprehensive backup/restore with validation and incremental support.

---

## 🧩 Flake Input Analysis & AI Explanations

Easily analyze your `flake.nix` inputs and get AI-powered explanations for each input, including upstream README and flake.nix summaries. nixai supports both `name.url = ...;` and `name = { url = ...; ... };` forms, robustly handling comments and whitespace.

**Example:**

```sh
nixai flake explain --flake /path/to/flake.nix
```

---

## 🔧 NixOS Option Explainer

Get comprehensive, AI-powered explanations for any NixOS option with **usage examples**, **best practices**, and **related options**:

```sh
# Get detailed explanation with examples and best practices
nixai explain-option services.nginx.enable

# Comprehensive firewall configuration guide
nixai explain-option networking.firewall.enable

# Boot loader setup with real-world examples
nixai explain-option boot.loader.systemd-boot.enable

# Natural language queries also work
nixai explain-option "how to enable SSH access"

# Get help for complex nested options with advanced examples
nixai explain-option services.postgresql.settings.shared_preload_libraries
```

**What you get with each explanation:**

- 📖 **Purpose & Overview**: Clear explanation of what the option does
- 🔧 **Type & Default**: Data type and default value information
- 💡 **Usage Examples**: Basic, real-world, and advanced configuration examples
- ⭐ **Best Practices**: Tips, warnings, and recommendations
- 🔗 **Related Options**: Other options that work well together
- 🎨 **Beautiful Formatting**: Colorized terminal output with proper syntax highlighting

**Available in both CLI and interactive modes:**

```sh
# CLI mode - NixOS options
nixai explain-option services.nginx.enable
nixai explain-option networking.firewall.enable

# CLI mode - Home Manager options  
nixai explain-home-option programs.git.enable
nixai explain-home-option home.username

# Interactive mode
nixai interactive
> explain-option <option>
> explain-home-option <option>
```

**🎯 Smart Visual Distinction**: nixai automatically detects and displays the appropriate headers:

- `🖥️ NixOS Option` for system-level configuration options
- `🏠 Home Manager Option` for user-level configuration options

The Option Explainer provides:

- **Type**: The data type of the option (boolean, string, list, etc.)
- **Default Value**: What the option defaults to if not set
- **Description**: Official documentation from NixOS/Home Manager
- **Source**: The module file where the option is defined
- **AI Explanation**: Context, purpose, and best practices
- **Usage Examples**: Practical configuration examples (basic, common, advanced)
- **Related Options**: Other options that work well together

---

## 📦 AI-Powered Package Repository Analysis

Automatically analyze Git repositories and generate complete Nix derivations using AI-powered build system detection and dependency analysis. nixai supports multiple programming languages and automatically generates proper Nix packaging files.

**Example:**

```sh
# Analyze current directory and generate derivation
nixai package-repo . --local

# Analyze specific project
nixai package-repo /path/to/project

# Analyze only (no derivation generation)
nixai package-repo . --analyze-only

# Remote repository analysis
nixai package-repo https://github.com/user/project
```

**Supported Languages & Build Systems:**

- **Go**: Detects go.mod, analyzes dependencies, generates buildGoModule derivations
- **Python**: Detects setup.py/pyproject.toml, analyzes pip dependencies
- **Node.js**: Detects package.json, analyzes npm dependencies, generates buildNpmPackage derivations
- **Rust**: Detects Cargo.toml, analyzes crate dependencies, generates buildRustPackage derivations

**What you get with each analysis:**

- 🔍 **Build System Detection**: Automatically identifies project type and build files
- 📊 **Dependency Analysis**: Extracts and analyzes all project dependencies
- 🤖 **AI-Generated Derivations**: Complete, valid Nix derivations with proper structure
- ✅ **Validation**: Ensures generated derivations include all required attributes
- 📝 **Metadata Extraction**: Project name, version, license, and description detection
- 🔗 **Git Integration**: Automatic source URL and commit information extraction

**Key Features:**

- **Multi-Language Support**: Works with Go, Python, Node.js, and Rust projects
- **Build System Detection**: Automatically identifies build files and project structure
- **AI-Powered Generation**: Uses advanced AI to generate complete, working derivations
- **Comprehensive Validation**: Validates generated derivations for completeness and correctness
- **Path Resolution**: Supports both relative and absolute paths
- **Debug Mode**: Comprehensive logging for troubleshooting and development

---

## 🔍 Enhanced Search Command: Discover, Configure, and Master NixOS Packages & Services

The `nixai search` command now provides a powerful, interactive experience for exploring NixOS packages and services:

- **Comprehensive Option Listing:**
  - See *all* available NixOS options for a package/service, including type, default, description, and real-world examples.
- **Config Snippets for Every Setup:**
  - Instantly copy-paste configuration snippets for classic `/etc/nixos/configuration.nix`, Home Manager, and flake-based setups.
- **AI-Powered Best Practices & Pitfalls:**
  - Get advanced usage tips, best practices, and common pitfalls—sourced from both official docs and AI analysis.
- **Interactive Exploration:**
  - Use interactive prompts to view option details, copy config snippets, or ask for further explanation.
- **Beautiful Output:**
  - All results are formatted with headers, key-value tables, and Markdown rendering for easy reading and direct use in your configs.

### Example Usage

```sh
nixai search nginx
```

*What you'll see:*
- A list of all NixOS options for `nginx` (e.g., `services.nginx.enable`, `services.nginx.virtualHosts`, ...)
- For each option: type, default, description, and example usage
- Config snippets for classic, Home Manager, and flake setups (ready to copy)
- AI-powered best practices and advanced tips for configuring `nginx`
- Interactive prompt to view more details or copy a snippet

### Why Use This?
- Quickly discover all configuration options for any package or service
- Avoid common mistakes and follow best practices
- Easily adapt examples to your preferred NixOS setup style
- Learn advanced usage patterns and troubleshooting tips

See the [User Manual](docs/MANUAL.md#searching-for-packages-and-services) for full details and advanced examples.

---

## 🖥️ Multi-Machine Management (Flake-based)

nixai now manages all machines directly from your `flake.nix` using the `nixosConfigurations` attribute. There is no registry or YAML file. All commands operate on hosts defined in your flake.

### Listing Hosts

```
nixai machines list
```
Lists all hosts from `flake.nix`.

### Deploying to a Host

```
nixai machines deploy --machine <hostname>
```
Deploys to the specified host using `nixos-rebuild` (default) or `deploy-rs` if configured.

### Requirements
- Your `flake.nix` must define all hosts under `nixosConfigurations`.
- For remote deploy, use `nixos-rebuild switch --flake .#<hostname> --target-host <host>`.
- For advanced fleet deploy, configure `deploy-rs` in your flake.

See `docs/FLAKE_INTEGRATION_GUIDE.md` for migration details.

---

## 📝 Configuration Templates & Snippets

Browse, apply, and manage curated NixOS configuration templates with `nixai templates` and save/reuse configuration snippets with `nixai snippets`. Includes GitHub search integration for discovering real-world configurations.

#### Templates

```sh
# Browse available templates
nixai templates list

# Search templates by keyword or category
nixai templates search gaming
nixai templates search desktop kde

# Show template details and content
nixai templates show desktop-minimal

# Apply template to current configuration
nixai templates apply gaming-config

# Save configuration as template
nixai templates save my-template /path/to/config.nix

# List template categories
nixai templates categories

# Search GitHub for NixOS configurations
nixai templates github "gaming nixos configuration"
nixai templates github "kde plasma nixos"
```

#### Snippets

```sh
# List saved snippets
nixai snippets list

# Search snippets by name or tag
nixai snippets search nvidia

# Save current config as snippet
nixai snippets add my-nvidia-config --file /etc/nixos/hardware.nix

# Apply snippet to configuration
nixai snippets apply gaming-setup --output ./gaming.nix

# Show snippet content
nixai snippets show my-snippet

# Remove saved snippet
nixai snippets remove old-config
```

**Key Features:**

- **Built-in Templates**: Curated templates for desktop, gaming, server, and development environments
- **GitHub Integration**: Search real-world NixOS configurations from the community
- **Personal Snippets**: Save and organize your own configuration snippets with tags
- **Template Categories**: Organized by Desktop, Gaming, Server, Development, etc.
- **YAML Persistence**: Templates and snippets stored as structured YAML files
- **Merge Support**: Apply templates with merge capabilities for existing configurations

---

## 🎨 Terminal Output Formatting

All Markdown and HTML output from nixai is rendered as beautiful, colorized terminal output using [glamour](https://github.com/charmbracelet/glamour) and [termenv](https://github.com/muesli/termenv). This makes AI explanations, documentation, and search results easy to read and visually appealing.

- Works in all modern terminals.
- Respects your terminal theme (light/dark).
- Makes complex output (tables, code, links) readable at a glance.

---

## 🔧 Enhanced Build Troubleshooter

```sh
# Basic build with AI assistance
nixai build .#mypackage

# Deep analysis of build failures
nixai build debug firefox

# Intelligent retry with automated fixes
nixai build retry

# Cache miss analysis
nixai build cache-miss

# Sandbox debugging
nixai build sandbox-debug

# Build performance profiling
nixai build profile --package vim
```

The Enhanced Build Troubleshooter provides comprehensive build failure analysis and resolution capabilities:

- **🔍 Deep Analysis**: Pattern recognition for common failure types with detailed explanations
- **🤖 Intelligent Retry**: Automated application of fixes for common issues with progress feedback
- **📊 Cache Analysis**: Detailed analysis of cache hit/miss patterns with optimization recommendations
- **🔐 Sandbox Debugging**: Resolve permission issues and sandbox environment problems
- **⚡ Performance Profiling**: Identify build bottlenecks and optimization opportunities

Each subcommand provides focused analysis and actionable recommendations, helping you resolve build failures faster and optimize your build process.

---

## 🔌 Editor Integration

nixai provides seamless integration with popular editors through the MCP (Model Context Protocol) server, enabling you to access NixOS documentation and AI assistance directly within your development environment.

### 🔷 VS Code Integration

Complete VS Code integration with Copilot, Claude Dev, and other MCP-compatible extensions:

**Quick Setup:**

```sh
# Start the MCP server
nixai mcp-server start -d

# Check server status
nixai mcp-server status
```

**Required Extensions:**

- `automatalabs.copilot-mcp` - Copilot MCP extension
- `zebradev.mcp-server-runner` - MCP Server Runner
- `saoudrizwan.claude-dev` - Claude Dev (Cline)

**Configuration (.vscode/settings.json):**

```json
{
  "mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "copilot.mcp.servers": {
    "nixai": {
      "command": "bash", 
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "claude-dev.mcpServers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "mcp.enableDebug": true,
  "claude-dev.enableMcp": true
}
```

### 🟢 Neovim Integration

Comprehensive Neovim integration with lua configuration and keybindings:

**Automatic Setup:**

```sh
# Automatically configure Neovim integration
nixai neovim-setup

# With custom socket path
nixai neovim-setup --socket-path=$HOME/.local/share/nixai/mcp.sock

# With custom config directory
nixai neovim-setup --config-dir=$HOME/.config/nvim
```

**Manual Setup (init.lua):**

```lua
-- nixai integration
local ok, nixai = pcall(require, "nixai")
if ok then
  nixai.setup({
    socket_path = "/tmp/nixai-mcp.sock",
    keybindings = true, -- Enable default keybindings
  })
else
  vim.notify("nixai module not found", vim.log.levels.WARN)
end
```

**Available Commands:**

- `:NixaiExplainOption <option>` - Explain NixOS options
- `:NixaiExplainHomeOption <option>` - Explain Home Manager options
- `:NixaiSearch <query>` - Search packages and services
- `:NixaiDiagnose` - Diagnose current buffer or selection
- `:NixaiAsk <question>` - Ask direct questions

**Default Keybindings:**

- `<leader>ne` - Explain option under cursor
- `<leader>ns` - Search packages/services
- `<leader>nd` - Diagnose current buffer
- `<leader>na` - Ask nixai a question

### 🏠 Home Manager Integration

Both editors can be automatically configured through Home Manager:

```nix
# home.nix
{ config, pkgs, ... }:
{
  imports = [
    # Import the nixai Home Manager module
    (builtins.fetchTarball "https://github.com/olafkfreund/nix-ai-help/archive/main.tar.gz")/modules/home-manager.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      socketPath = "$HOME/.local/share/nixai/mcp.sock";
    };
    # Automatically configure VS Code
    vscodeIntegration = true;
    # Automatically configure Neovim  
    neovimIntegration = true;
  };
}
```

For detailed setup instructions and troubleshooting, see:

- [VS Code Integration Guide](docs/MCP_VSCODE_INTEGRATION.md)
- [Neovim Integration Guide](docs/neovim-integration.md)

---

## 🛠️ Installation & Usage

### Using Nix (Recommended)

**For Development Environment:**

```sh
# Enter development environment (includes Go, just, golangci-lint, etc.)
nix develop

# Build using just (recommended)
just build
./nixai --help
```

### System Integration with NixOS and Home Manager

nixai can be integrated into your NixOS or Home Manager configuration using the provided modules:

```nix
# NixOS configuration
{ config, pkgs, ... }:
{
  imports = [
    # Path to module (can be from flake or local)
    ./path/to/nixai/modules/nixos.nix 
  ];
  
  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # Optional configuration
    };
  };
}
```

```nix
# Home Manager configuration
{ config, pkgs, ... }:
{
  imports = [
    # Path to module (can be from flake or local)
    ./path/to/nixai/modules/home-manager.nix
  ];
  
  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # Optional configuration
    };
  };
}
```

**📚 For comprehensive flake integration setup, see our complete [Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md)** which covers:
- ✅ Adding nixai as a flake input
- ✅ NixOS and Home Manager module configuration
- ✅ All available options and features
- ✅ AI provider setup (Ollama, OpenAI, Gemini)
- ✅ Editor integrations (VS Code, Neovim)
- ✅ Troubleshooting and examples

See the [MCP Server Configuration & Autostart](#-mcp-server-configuration--autostart) section for more details.

**For Direct Nix Build:**

```sh
# Note: Nix direct build currently has some packaging issues
# Use the development environment + just build instead
nix build .#nixai  # Currently under development
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

## 🔗 Flake Integration

Want to add nixai to your flake.nix and install it system-wide or per-user? We've got you covered with comprehensive guides:

### Quick Start
- **📋 [Flake Quick Reference](docs/FLAKE_QUICK_REFERENCE.md)** - Essential copy-paste snippets for immediate use
- **📚 [Complete Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md)** - Detailed setup with all options and features

### What You Can Do
- Add nixai as a flake input to your NixOS or Home Manager configuration
- Enable system-wide or per-user installation with modules
- Configure AI providers (Ollama, OpenAI, Gemini) declaratively
- Auto-configure editor integrations (VS Code, Neovim)
- Set up MCP server for advanced documentation features
- Use in combination with other flake inputs seamlessly

Both guides include complete working examples, troubleshooting, and best practices for integrating nixai into your Nix-based system.

---

# NixAI Flake Quick Reference

## Features
- Multi-system support (x86_64-linux, aarch64-linux, darwin)
- Reproducible Go build for `nixai` CLI
- Dev shells for local and Docker development
- NixOS and Home Manager modules
- Built-in code formatter (alejandra)
- Lint check for Go code (golangci-lint)

## Usage Examples

### Build the CLI
```zsh
nix build .#nixai
```

### Run the CLI
```zsh
nix run .#nixai -- --help
```

### Enter Dev Shell
```zsh
nix develop
```

### Run Lint Check
```zsh
nix flake check
```

### Format Nix Code
```zsh
nix fmt
```

### Use NixOS Module
Add to your configuration:
```nix
imports = [
  (fetchGit { url = "https://github.com/olafkfreund/nix-ai-help"; }) + "/modules/nixos.nix"
];
```

### Use Home Manager Module
Add to your configuration:
```nix
imports = [
  (fetchGit { url = "https://github.com/olafkfreund/nix-ai-help"; }) + "/modules/home-manager.nix"
];
```

---

See `docs/FLAKE_QUICK_REFERENCE.md` and `docs/FLAKE_INTEGRATION_GUIDE.md` for more details.

---

## 🆕 Flake Creation & Correction

The `nixai flake create` command allows you to quickly create, correct, or upgrade a `flake.nix` for any project folder, with AI-powered build system detection and best-practice suggestions.

### Features
- Create a minimal flake from scratch (`--from-scratch`)
- Analyze a project folder and generate a flake for Go, Node, Rust, Python, or generic projects (`--analyze`)
- Correct and upgrade an existing `flake.nix` with AI assistance (`--fix`)
- Overwrite with `--force`
- Customize system and description

### Usage Examples

```sh
# Create a minimal flake.nix in the current directory
nixai flake create --from-scratch

# Analyze a project folder and generate a flake.nix
nixai flake create . --analyze

# Fix and update an existing flake.nix using AI
nixai flake create . --fix --force

# Specify system and description
nixai flake create myproject --from-scratch --system x86_64-linux --desc "My Project Flake"
```

The command will detect the build system (Go, Node, Rust, Python, or generic), generate a valid flake, and provide AI-powered explanations and best practices.

---

## 📚 Flake Integration Guide

For users who want to integrate **nixai** into their NixOS or Home Manager configurations using flakes, we've created a comprehensive guide that covers everything you need to know.

### 🚀 Quick Start Options

You have several ways to use nixai with flakes:

1. **Direct Usage** (no configuration needed):
   ```bash
   # Run nixai directly from the flake
   nix run github:olafkfreund/nix-ai-help -- "how do I enable SSH?"
   
   # Install to user profile
   nix profile install github:olafkfreund/nix-ai-help
   ```

2. **NixOS System Integration**:
   ```nix
   # In your flake.nix
   inputs.nixai.url = "github:olafkfreund/nix-ai-help";
   
   # In your configuration
   nixosConfigurations.hostname = nixpkgs.lib.nixosSystem {
     modules = [
       nixai.nixosModules.x86_64-linux.default # <-- Use your system here
       { services.nixai.enable = true; }
     ];
   };
   ```

3. **Home Manager Integration**:
   ```nix
   # In your flake.nix
   inputs.nixai.url = "github:olafkfreund/nix-ai-help";
   
   # In your home configuration
   homeConfigurations.username = home-manager.lib.homeManagerConfiguration {
     modules = [
       nixai.homeManagerModules.x86_64-linux.default # <-- Use your system here
       { services.nixai.enable = true; }
     ];
   };
   ```

### 📖 Complete Guide

**➡️ For detailed setup instructions, configuration options, AI provider setup, editor integrations, troubleshooting, and real-world examples, see our comprehensive [Flake Integration Guide](docs/FLAKE_INTEGRATION_GUIDE.md).**

This guide covers:

- ✅ **Complete flake setup** - Step-by-step integration into your configurations
- ✅ **All configuration options** - AI providers, MCP server, documentation sources
- ✅ **Editor integrations** - VS Code and Neovim automatic configuration
- ✅ **AI provider setup** - Ollama (local/private), OpenAI, and Gemini configuration
- ✅ **Advanced features** - Security hardening, multi-provider setups, custom sources
- ✅ **Troubleshooting** - Common issues and solutions
- ✅ **Real examples** - Complete working configurations for different use cases

---

## 🧑‍💻 Development Setup (For New Contributors)

This project uses **Nix flakes** for reproducible development environments. Here's the complete workflow for new users:

### Prerequisites

- **Nix** with flakes enabled
- **Git** for version control

### Quick Start

1. **Clone the repository:**

   ```sh
   git clone <repository-url>
   cd nix-ai-help
   ```

2. **Enter the development environment:**
   - Go 1.24.3
   - just (task runner)
   - golangci-lint
   - All required development tools

3. **Clean and install dependencies:**

   ```sh
   go clean -modcache  # Clean any cached modules
   go mod tidy         # Download and organize dependencies
   ```

4. **Build the project:**

   ```sh
   just build
   ```

5. **Test the application:**

   ```sh
   ./nixai --help      # Verify the build works
   just test           # Run the test suite
   ```

### Development Commands

```sh
# Available just commands (run `just -l` to see all)
just build          # Build nixai binary
just test           # Run all tests
just lint           # Run linter (may show minor issues)
just fmt            # Format Go code
just clean          # Remove build artifacts
just run            # Build and run nixai
just deps           # Install/update dependencies

# Manual Go commands
go build -o nixai ./cmd/nixai/main.go    # Direct Go build
go test ./...                            # Direct test execution
go mod tidy                              # Update dependencies
```

### Project Structure

```
├── cmd/nixai/           # Main application entry point
├── internal/            # Internal packages
│   ├── ai/             # LLM provider integrations (Ollama, OpenAI, Gemini)
│   ├── cli/            # CLI commands and interactive mode
│   ├── config/         # Configuration management (YAML)
│   ├── mcp/            # MCP server for documentation queries
│   └── nixos/          # NixOS-specific diagnostics and parsing
├── pkg/                # Public utility packages
│   ├── logger/         # Structured logging
│   └── utils/          # General utilities
├── configs/            # Default configuration files
├── flake.nix          # Nix flake for development environment
└── justfile           # Task automation
```

### Testing Your Changes

1. **Unit tests:**

   ```sh
   just test
   ```

2. **Integration testing:**

   ```sh
   # Test specific functionality
   ./nixai --help
   ./nixai search nginx
   ./nixai explain-option services.nginx.enable
   ```

3. **Code quality:**

   ```sh
   just lint  # Check for code quality issues
   just fmt   # Format code automatically
   ```

### Common Development Issues

- **Module permission errors**: Run `go clean -modcache` and `go mod tidy`
- **Build failures**: Ensure you're in the Nix development shell (`nix develop`)
- **Missing tools**: The Nix flake provides all required tools automatically

---

## 📝 Commands & Usage

### Ask Questions Directly

The quickest way to get help with NixOS configuration is to ask questions directly:

```sh
# Ask questions directly (most common usage)
nixai "how do I enable SSH in NixOS?"
nixai "what is a Nix flake?"
nixai "how do I configure nginx with SSL?"

# Alternative: use the --ask flag
nixai --ask "how do I update packages in NixOS?"
nixai -a "what's the difference between NixOS and other Linux distributions?"

# Both methods work identically and provide:
# - AI-powered answers with examples
# - Context from official NixOS documentation  
# - Best practices and recommendations
# - Beautiful formatted terminal output
```

**Features:**

- 🤖 **AI-Powered Responses**: Get comprehensive answers using Ollama, Gemini, or OpenAI
- 📚 **Documentation Context**: Automatic querying of official NixOS docs via MCP server
- 🎨 **Beautiful Output**: Colorized markdown with syntax highlighting
- ⚡ **Fast & Simple**: Just ask your question naturally

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

### Explain NixOS Options

```sh
nixai explain-option <option>
```

- Get AI-powered explanations for any NixOS option.
- Includes type, default value, description, and best practices.
- Uses official NixOS documentation from Elasticsearch backend.
- Automatically detects and displays `🖥️ NixOS Option` headers for system-level options.

### Explain Home Manager Options

```sh
nixai explain-home-option <option>
```

- Get AI-powered explanations for Home Manager options.
- Dedicated command for user-level configuration options.
- Shows `🏠 Home Manager Option` headers with appropriate documentation sources.
- Perfect for understanding programs, services, and home directory management.

### AI-Powered Package Repository Analysis

```sh
nixai package-repo <path>
```

- Automatically analyze Git repositories and generate Nix derivations.
- Supports Go, Python, Node.js, and Rust projects with intelligent build system detection.
- AI-powered derivation generation with nixpkgs best practices.
- Comprehensive dependency analysis and nixpkgs mapping.

**Examples:**

```sh
# Analyze current directory and generate derivation
nixai package-repo . --local

# Analyze specific project
nixai package-repo /path/to/project

# Analyze only (no derivation generation)
nixai package-repo . --analyze-only

# Remote repository analysis
nixai package-repo https://github.com/user/project

# Custom output directory and package name
nixai package-repo https://github.com/user/rust-app --output ./derivations --name my-package
```

**Key Features:**

- **Multi-Language Support**: Detects Go modules, npm packages, Python projects, Rust crates
- **Build System Detection**: Automatically identifies build files and project structure
- **AI Generation**: Creates complete, valid derivations with proper structure and metadata
- **Validation**: Ensures generated derivations include all required attributes
- **Git Integration**: Automatic source URL and commit information extraction

### Dependency & Import Graph Analyzer

```sh
# Show dependency tree with AI insights
nixai deps analyze

# Explain why a package is installed
nixai deps why firefox

# Find and resolve conflicts
nixai deps conflicts

# Suggest dependency optimizations
nixai deps optimize

# Generate visual dependency graph
nixai deps graph
```

The Dependency & Import Graph Analyzer provides comprehensive tools for visualizing and optimizing your NixOS configuration dependencies:

- **🔍 Dependency Mapping**: Build comprehensive dependency graphs with insight into package relationships
- **❓ Package Explanation**: Get clear explanations of why specific packages are installed in your system
- **🚫 Conflict Detection**: Identify and resolve package conflicts and overlapping dependencies
- **⚡ Optimization Analysis**: Find opportunities to streamline your dependency tree for better performance
- **📊 Visual Graphs**: Generate visual dependency graphs for better understanding of your system

This feature is particularly useful for debugging complex dependency issues, optimizing system performance, and understanding the relationships between packages in your NixOS configuration.

---

## 🏪 Nix Store Management

Manage, backup, restore, and analyze your Nix store directly from the CLI. These commands help ensure the integrity, recoverability, and performance of your NixOS system.

### Store Command Usage

```sh
# Create a Nix store backup (default output file)
nixai store backup

# Create a Nix store backup with a custom output file
nixai store backup --output my-backup.tar.gz

# Restore the Nix store and configuration from a backup
nixai store restore my-backup.tar.gz

# Check the integrity of the Nix store and configuration
nixai store integrity

# Analyze Nix store performance and usage
nixai store performance
```

**Example Output:**

```text
🗄️ Nix Store Backup
━━━━━━━━━━━━━━━━━━━
Creating backup...
✔️ Backup created at: my-backup.tar.gz

♻️ Nix Store Restore
━━━━━━━━━━━━━━━━━━━━
Restoring from backup: my-backup.tar.gz
✔️ Restore completed from: my-backup.tar.gz

🔍 Nix Store Integrity Check
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Checking store integrity...
✔️ Store integrity check completed (no issues found).

⚡ Nix Store Performance Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Analyzing store performance...
✔️ Store performance analysis completed.
```

See the manual for advanced options and integration with disaster recovery workflows.

---

### System State Backup (Planned/Preview)

```sh
nixai backup create               # Create comprehensive backup
nixai backup restore <backup>     # Restore from backup
nixai backup schedule             # Schedule automated backups
nixai backup verify               # Verify backup integrity
nixai backup list                 # List available backups
```

### Configuration Templates & Snippets

Manage curated NixOS configuration templates and personal configuration snippets:

#### Templates

```sh
# Browse available templates
nixai templates list

# Search templates by keyword or category
nixai templates search gaming
nixai templates search desktop kde

# Show template details and content
nixai templates show desktop-minimal

# Apply template to current configuration
nixai templates apply gaming-config

# Save configuration as template
nixai templates save my-template /path/to/config.nix

# List template categories
nixai templates categories

# Search GitHub for NixOS configurations
nixai templates github "gaming nixos configuration"
nixai templates github "kde plasma nixos"
```

#### Snippets

```sh
# List saved snippets
nixai snippets list

# Search snippets by name or tag
nixai snippets search nvidia

# Save current config as snippet
nixai snippets add my-nvidia-config --file /etc/nixos/hardware.nix

# Apply snippet to configuration
nixai snippets apply gaming-setup --output ./gaming.nix

# Show snippet content
nixai snippets show my-snippet

# Remove saved snippet
nixai snippets remove old-config
```

**Key Features:**

- **Built-in Templates**: Curated templates for desktop, gaming, server, and development environments
- **GitHub Integration**: Search real-world NixOS configurations from the community
- **Personal Snippets**: Save and organize your own configuration snippets with tags
- **Template Categories**: Organized by Desktop, Gaming, Server, Development, etc.
- **YAML Persistence**: Templates and snippets stored as structured YAML files
- **Merge Support**: Apply templates with merge capabilities for existing configurations

### System Health Check

```sh
nixai health
```

- Run comprehensive NixOS system health checks.
- Validates configuration, checks service status, disk space, and more.
- Provides AI-powered analysis and actionable recommendations.
- Beautiful, colorized terminal output with progress indicators.

### Interactive Mode

```sh
nixai interactive
```

- Supports all search, diagnose, and health check features.

- Use `set-nixos-path` to specify your config folder interactively.

### MCP Server Management

```sh
# Start the MCP server (foreground)
nixai mcp-server start

# Start in background (daemon mode)
nixai mcp-server start --background
nixai mcp-server start -d
nixai mcp-server start --daemon  # alias for --background

# With custom socket path
nixai mcp-server start --socket-path="/path/to/socket"

# Check MCP server status
nixai mcp-server status

# Stop the MCP server
nixai mcp-server stop
```

- You can use any of `--background`, `-d`, or `--daemon` to run the MCP server as a background process (daemon mode).
- MCP server provides NixOS documentation and options data to enhance AI responses
- Can be configured to use a custom socket path for communication

---

## 🐚 Shell Integration: Always-On nixai Assistant

You can integrate `nixai` into your shell for real-time help, error decoding, and system monitoring. Below are recommended setups for **zsh**, **bash**, **fish**, and **nushell**.

---

## 1. Quick Access Alias

Add to your shell config for instant access:

### zsh / bash
```sh
alias nxai='nixai'
```

### fish
```fish
alias nxai 'nixai'
```

### nushell
```nu
alias nxai = nixai $args
```

---

## 2. Automatic Error Decoding

### zsh
Add to your `.zshrc`:
```sh
function TRAPERR() {
  nixai decode-error "$BASH_COMMAND failed with exit code $?"
}
setopt ERR_EXIT
trap TRAPERR ERR
```

### bash
Add to your `.bashrc`:
```sh
nixai_on_error() {
  local status=$?
  if [[ $status -ne 0 && -n "$BASH_COMMAND" ]]; then
    nixai decode-error "$BASH_COMMAND failed with exit code $status"
  fi
}
trap nixai_on_error ERR
```

### fish
Add to your `config.fish`:
```fish
function fish_postexec --on-event fish_postexec
  if test $status -ne 0
    nixai decode-error (history --max=1 | string trim)
  end
end
```

### nushell
Add to your `config.nu`:
```nu
after-execution = { |cmd, exit_code|
  if $exit_code != 0 {
    nixai decode-error $"($cmd) failed with exit code ($exit_code)"
  }
}
```

---

## 3. Real-Time Log & Resource Monitoring

### journalctl error decoding
```sh
journalctl -xef | nixai decode-error
```

### Store health monitoring (all shells)
```sh
nixai store health --watch
```

---

## 4. Shell Completion

### zsh
```sh
nixai completion zsh > ~/.nixai-completion.zsh
echo "source ~/.nixai-completion.zsh" >> ~/.zshrc
```

### bash
```sh
nixai completion bash > ~/.nixai-completion.bash
echo "source ~/.nixai-completion.bash" >> ~/.bashrc
```

### fish
```sh
nixai completion fish > ~/.config/fish/completions/nixai.fish
```

### nushell
```sh
nixai completion nu > ~/.config/nushell/scripts/nixai-completion.nu
source ~/.config/nushell/scripts/nixai-completion.nu
```

---

## 5. Usage Examples

- **Decode last error automatically:**  
  Just run any command; if it fails, `nixai` will explain the error.
- **Pipe logs for analysis:**  
  `journalctl -xef | nixai decode-error`
- **Monitor Nix store health:**  
  `nixai store health --watch`
- **Tab-complete nixai commands:**  
  Type `nixai <TAB>` for completions.

---

> See `docs/MANUAL.md` for more advanced integration and troubleshooting tips.
