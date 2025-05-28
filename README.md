# nixai: NixOS AI Assistant

![nixai logo](https://nixos.org/logo/nixos-logo-only-hires.png)

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

- All configuration is loaded from YAML (see `configs/default.yaml`).

- You can set the AI provider, documentation sources, and more.

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
