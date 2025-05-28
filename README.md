# nixai: NixOS AI Assistant

![nixai logo](https://nixos.org/logo/nixos-logo-only-hires.png)

---

## 🚀 Project Overview

**nixai** is a powerful, console-based Linux application designed to help you solve NixOS configuration problems, create and configure NixOS systems, and diagnose issues—all from the command line. It leverages advanced Large Language Models (LLMs) like Gemini, OpenAI, and Ollama, with a strong preference for local Ollama models to ensure your privacy. nixai integrates an MCP server to query NixOS documentation from multiple official and community sources, and provides interactive and scriptable diagnostics, log parsing, and command execution.

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

## ✨ Features

- Diagnose NixOS configuration and log issues using LLMs.
- Query NixOS documentation from multiple official and community sources.
- Execute and parse local NixOS commands.
- Accept log input via pipe or file.
- User-selectable AI provider (Ollama, Gemini, OpenAI, etc.).
- Modular, testable, and well-documented Go codebase.

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

## 🤝 Contributing

- Please submit a pull request or open an issue for enhancements or bug fixes.
- Follow the guidelines in `.github/copilot-instructions.md` for consistency.

---

## 📄 License

This project is licensed under the MIT License. See the LICENSE file for details.

---

## 🙏 Acknowledgments

- Thanks to the NixOS community and documentation authors.
- Special thanks to the creators of the AI models used in this project.

---

> _nixai: Your AI-powered NixOS assistant, right in your terminal._
