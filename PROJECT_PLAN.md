# nixai Project Plan

## Purpose

A console-based Linux application to help solve NixOS configuration problems and assist in creating and configuring NixOS from the command line. Uses LLMs (Ollama, Gemini, OpenAI, etc.) and integrates with an MCP server for documentation queries.

## Key Features

- Diagnose NixOS configuration and log issues using LLMs
- Query NixOS documentation from multiple sources
- Execute and parse local NixOS commands
- Accept log input via pipe, file, or `nix log`
- Search for Nix packages and services with clean, numbered results
- Show config/test options and available `nixos-option` settings for selected package/service
- Specify NixOS config folder with `--nixos-path`/`-n` (CLI) or `set-nixos-path` (interactive)
- Interactive and CLI modes
- User-selectable AI provider (Ollama preferred for privacy)

## Recent Changes (May 2025)

- Added `--nix-log` (`-g`) flag to `nixai diagnose` to analyze output from `nix log` (optionally with a derivation/path)
- Improved search: clean output, numbered results, config option lookup, and config path awareness
- All features available in both CLI and interactive modes
- README and help text updated for new features

## Configuration

- All config loaded from YAML (`configs/default.yaml`)
- AI provider, documentation sources, and more are user-configurable

## Build & Test

- Use `justfile` for build/test/lint/run
- Use `flake.nix` for reproducible dev environments

## Roadmap / TODO

- [x] Add robust, user-friendly Nix package/service search (CLI & interactive)
- [x] Integrate `nixos-option` for config lookup
- [x] Add `--nixos-path`/`-n` and `set-nixos-path` for config folder selection
- [x] Add `--nix-log`/`-g` to diagnose from `nix log`
- [ ] (Optional) Use config path for context-aware features
- [ ] (Optional) Automate service option lookup further
- [ ] (Optional) Enhance user guidance and error handling for config path
- [ ] (Optional) Add more tests for new features

## Contributing

- Follow Go idioms and best practices
- Keep code modular and well-documented
- Add/update tests for all new features and bugfixes
- Update README and this file for any new features or changes

---

See README.md for usage and configuration details.
