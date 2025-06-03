# Copilot Instructions for the nixai Project

## Project Purpose
- nixai is a modular, console-based Linux application for solving NixOS configuration problems and assisting with NixOS setup and troubleshooting from the command line.
- Supports direct AI-powered help via `nixai "question"` or `nixai --ask "question"`.
- Integrates multiple LLMs (Ollama, Gemini, OpenAI, etc.), defaulting to local Ollama for privacy, with user-selectable providers.
- Leverages an MCP server to query NixOS documentation from multiple official and community sources.
- Parses log outputs, accepts piped logs, executes and diagnoses local NixOS commands, and supports both interactive and CLI-driven workflows.
- Modular submodules for community, packaging, learning, devenv, neovim, machines, and more.

## Coding Guidelines
- Use idiomatic Go and modular design for all code.
- All configuration is loaded from YAML (see `configs/default.yaml`) via the `internal/config` package.
- Use `internal/ai` for all LLM interactions. All providers must implement both `Query` and `GenerateResponse` methods. Allow user/provider selection and fallback.
- Use `internal/mcp` for documentation queries, always using sources from config.
- Use `internal/nixos` for log parsing, diagnostics, and NixOS command execution.
- Use `pkg/logger` for all logging, respecting log level from config.
- Use `pkg/utils` for utility functions (file checks, string helpers, formatting, etc.).
- All CLI commands and interactive logic must be in `internal/cli`. Each submodule may have its own Copilot instructions.
- Main entrypoint: `cmd/nixai/main.go`.
- All new features must be testable and documented.

## Features to Support
- **Direct Question Assistant**: Answer questions via `nixai "question"` or `--ask`/`-a` flag, using the provider's `Query` method.
- **Log & Config Diagnostics**: Diagnose NixOS issues from logs, configs, or piped input using LLMs.
- **Documentation Query**: Query NixOS docs from:
  - https://wiki.nixos.org/wiki/NixOS_Wiki
  - https://nix.dev/manual/nix
  - https://nix.dev/ 
  - https://nixos.org/manual/nixpkgs/stable/
  - https://nix.dev/manual/nix/2.28/language/
  - https://nix-community.github.io/home-manager/
- **NixOS Command Execution**: Run and parse local NixOS commands.
- **AI Provider Selection**: User can select/configure AI provider (Ollama, Gemini, OpenAI, etc.).
- **Package Repository Analysis**: Analyze Git repos and generate Nix derivations with `package-repo` command.
- **NixOS Option Explainer**: Explain NixOS options with `explain-option` command.
- **Home Manager Option Support**: Explain Home Manager options with `explain-home-option` command.
- **Community, Learning, Devenv, Machines, Neovim**: Modular commands for each area, each with its own help menu and Copilot instructions.
- **CLI Help Menus**: All commands must provide clear, formatted, actionable help menus with examples.
- **Interactive & Piped Input**: Accept logs/configs via pipe or file for analysis.
- **Progress Indicators**: Show progress during API calls and long operations.

## Best Practices
- Keep code modular, well-documented, and testable.
- Prefer local inference (Ollama) for privacy, fallback to cloud LLMs if needed.
- Validate and sanitize all user/log input.
- Use context and idiomatic error handling throughout.
- Format terminal output with `glamour` and `utils` formatting helpers (`FormatHeader`, `FormatKeyValue`, `FormatDivider`, etc.).
- For direct questions, always use the provider's `Query` method.
- Gracefully handle MCP server unavailability.
- Add or update tests for all new features and bugfixes.
- All new commands and features must be reflected in help menus and documentation.

## Testing & Build
- Use the `justfile` for build, test, lint, and run tasks.
- Use Nix (`flake.nix`) for reproducible builds and dev environments.
- All features must be covered by tests; update or add tests as needed.

## Documentation
- Update `README.md` and `docs/MANUAL.md` for all new features, commands, and changes.
- Keep this instruction file and all submodule `.instructions.md` files up to date.
- Document both direct question and flag-based interfaces, with examples in README and manual.
- Each submodule should have its own Copilot instructions reflecting its responsibilities and integration points.

## AI Provider Integration
- All providers (Ollama, Gemini, OpenAI, etc.) must implement both `Query` and `GenerateResponse` methods.
- Default to Ollama with `llama3` model if not specified.
- Format prompts consistently across providers.
- API keys must be kept in environment variables, never in config files.

## Terminal UI
- Use `utils.FormatHeader`, `utils.FormatKeyValue`, `utils.FormatDivider`, and other formatting utilities for all output.
- Use `glamour` for Markdown rendering with syntax highlighting.
- Show progress indicators for API calls and long-running operations.
- All CLI commands must provide clear, actionable help menus with usage examples.

---

> These instructions are for all Copilot models and contributors. Follow them to ensure consistency, maintainability, and feature completeness for the nixai project. All submodules must have their own up-to-date `.instructions.md` files. When answering questions about frameworks, libraries, or APIs, use Context7 to retrieve current documentation rather than relying on training data.
When answering questions about frameworks, libraries, or APIs, use Context7 to retrieve current documentation rather than relying on training data.
