# 🚀 nixai Project Plan

> **NixAI**: Your AI-powered, privacy-first, terminal-based NixOS assistant for troubleshooting, configuration, and learning.

---

## 🎯 Purpose

A console-based Linux application to help solve NixOS configuration problems and assist in creating and configuring NixOS from the command line. Uses LLMs (Ollama, Gemini, OpenAI, etc.) and integrates with an MCP server for documentation queries.

---

## ✨ Key Features

- 🩺 Diagnose NixOS configuration and log issues using LLMs
- 📚 Query NixOS documentation from multiple sources
- 🖥️ Execute and parse local NixOS commands
- 📥 Accept log input via pipe, file, or `nix log`
- 🔍 Search for Nix packages and services with clean, numbered results
- ⚙️ Show config/test options and available `nixos-option` settings for selected package/service
- 📂 Specify NixOS config folder with `--nixos-path`/`-n` (CLI) or `set-nixos-path` (interactive)
- 💬 Interactive and CLI modes
- 🤖 User-selectable AI provider (Ollama preferred for privacy)
- 🆕 **Robust flake input parser** (supports both `name.url = ...;` and `name = { url = ...; ... };` forms)
- 🆕 **AI-powered flake input explanation** (`nixai flake explain-inputs` and `nixai flake explain <input>`) with upstream README/flake.nix summarization
- 🆕 **Beautiful terminal output**: colorized, Markdown/HTML rendered with ANSI colors

---

## 📝 Recent Changes (May 2025)

- ➕ Added `--nix-log` (`-g`) flag to `nixai diagnose` to analyze output from `nix log` (optionally with a derivation/path)
- 🧹 Improved search: clean output, numbered results, config option lookup, and config path awareness
- 🔄 All features available in both CLI and interactive modes
- 🏗️ Flake input parser now supports all real-world input forms (attribute sets, comments, whitespace)
- 🤖 `nixai flake explain` and `nixai flake explain-inputs` now provide AI-powered, colorized, terminal-friendly explanations for all flake inputs
- 📖 README and help text updated for new features

---

## ⚙️ Configuration

- All config loaded from YAML (`configs/default.yaml`)
- AI provider, documentation sources, and more are user-configurable

---

## 🛠️ Build & Test

- Use `justfile` for build/test/lint/run
- Use `flake.nix` for reproducible dev environments

---

## 🗺️ Roadmap / TODO

- [x] Add robust, user-friendly Nix package/service search (CLI & interactive)
- [x] Integrate `nixos-option` for config lookup
- [x] Add `--nixos-path`/`-n` and `set-nixos-path` for config folder selection
- [x] Add `--nix-log`/`-g` to diagnose from `nix log`
- [x] Robust flake input parser for all input forms
- [x] AI-powered flake input explanation and upstream summarization
- [x] Terminal markdown/HTML formatting for explain output
- [x] (Optional) Use config path for context-aware features everywhere
- [ ] (Optional) Automate service option lookup further
- [ ] (Optional) Enhance user guidance and error handling for config path
- [ ] (Optional) Add more tests for new features

---

## 🧠 Planned: AI-Assisted Nix Configuration Management

- Add a `nixai config` command for AI-powered Nix configuration help:
  - Explain and suggest usage of `nix config` commands (show, set, unset, edit)
  - Interactive config editing: guide users through setting/unsetting options
  - Explain config options and best practices
  - Summarize current config and suggest improvements
  - Parse and review nix.conf or nix.conf.d, with AI-powered suggestions
  - Generate and explain `nix config` commands from natural language
  - Reverse lookup: explain and undo config commands
  - Show config sources and precedence
- Enhance question answering to recognize config-related queries and trigger the above logic
- Integrate with NixOS options and workflows for a seamless experience

---

## 🧩 Planned: AI-Powered Flake Input Analysis and Explanation

- Add a `nixai flake explain-inputs` and `nixai flake explain <input>` subcommand:
  - Parse the `inputs` section of the user's `flake.nix` (now robust to all forms)
  - For each input, fetch the referenced repo's `README.md` and/or `flake.nix` (if GitHub or similar)
  - Use the AI provider to summarize and explain the purpose of each input, how it's used, and best practices
  - Output a clean, numbered summary for each input, with explanations and actionable suggestions (now colorized/markdown in terminal)
  - Optionally, allow users to select an input for more details (full README, flake.nix, usage examples)
- **Benefits:** Users get instant, AI-powered insight into their flake inputs, best practices, and potential improvements for reproducibility and maintainability
- **Implementation:** Local flake.nix parsing, remote README.md/flake.nix fetching, AI summarization, and terminal rendering are all complete

---

## 🤝 Contributing

- Follow Go idioms and best practices
- Keep code modular and well-documented
- Add/update tests for all new features and bugfixes
- Update README and this file for any new features or changes

---

> See **README.md** for usage and configuration details.
