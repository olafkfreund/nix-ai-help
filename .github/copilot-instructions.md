# Copilot Instructions for the nixai Project

## Project Purpose
- This project is a console-based Linux application to help solve NixOS configuration problems and assist in creating and configuring NixOS from the command line.
- It allows asking direct questions via `nixai "question"` or `nixai --ask "question"` for immediate AI-powered help.
- It uses LLMs (Gemini, OpenAI, Ollama, and others) to diagnose and solve problems, with a preference for local Ollama models for privacy.
- It integrates an MCP server to query NixOS documentation from multiple web sources.
- It can parse log outputs, accept piped logs, execute local commands, and diagnose issues interactively or via CLI.

## Coding Guidelines
- Use Go idioms and best practices for all code.
- All configuration should be loaded from YAML (see `configs/default.yaml`).
- Use the `internal/config` package for configuration loading and management.
- Use the `internal/ai` package for all LLM interactions. Support multiple providers and allow user selection.
- Use the `internal/mcp` package for documentation queries. Always use the documentation sources defined in the config.
- Use the `internal/nixos` package for log parsing, diagnostics, and command execution.
- Use the `pkg/logger` package for all logging. Respect the log level from config.
- Use the `pkg/utils` package for utility functions (file checks, string helpers, etc.).
- All CLI commands and interactive logic should be in the `internal/cli` package.
- The main entrypoint is `cmd/nixai/main.go`.

## Features to Support
- **Direct Question Assistant**: Support answering questions directly via `nixai "question"` or with the `--ask`/`-a` flag.
- Diagnosing NixOS configuration and log issues using LLMs.
- Querying NixOS documentation from:
  - https://wiki.nixos.org/wiki/NixOS_Wiki
  - https://nix.dev/manual/nix
  - https://nixos.org/manual/nixpkgs/stable/
  - https://nix.dev/manual/nix/2.28/language/
  - https://nix-community.github.io/home-manager/
- Executing and parsing local NixOS commands.
- Accepting log input via pipe or file.
- Allowing user to select or configure AI provider (Ollama, Gemini, OpenAI, etc.).
- **Package Repository Analysis**: Support for analyzing Git repos and generating Nix derivations with the `package-repo` command.
- **NixOS Option Explainer**: Support for explaining NixOS options with the `explain-option` command.
- **Home Manager Option Support**: Dedicated `explain-home-option` command for Home Manager options.
- All new features must be testable and documented.

## Best Practices
- Keep all code modular and well-documented.
- Prefer local inference (Ollama) for privacy, but support cloud LLMs as fallback.
- Always validate and sanitize user input and log data.
- Use context and error handling idiomatically in Go.
- Write clear, maintainable, and idiomatic Go code.
- Add or update tests for all new features and bugfixes.
- Format terminal output with glamour for consistent, readable Markdown rendering.
- For direct questions, use the AI provider's Query method rather than GenerateResponse.
- Ensure graceful handling when the MCP server is unavailable.

## Testing & Build
- Use the `justfile` for common tasks: build, test, lint, run, etc.
- Use Nix (`flake.nix`) for reproducible builds and development environments.

## Documentation
- Update `README.md` and `docs/MANUAL.md` for any new features or changes.
- Keep this instruction file up to date as the project evolves.
- Document both direct question and flag-based question interfaces in user documentation.
- Include examples for all features in both the README and manual.

## AI Provider Integration
- Ensure all AI providers (Ollama, Gemini, OpenAI) implement both the Query and GenerateResponse methods.
- Default to Ollama with "llama3" model when no specific provider is configured.
- Format prompts consistently across all providers.
- Keep API keys in environment variables, not in the configuration files.

## Terminal UI
- Use utils.FormatHeader, utils.FormatKeyValue, utils.FormatDivider, and other formatting utilities.
- Use glamour for Markdown rendering with proper syntax highlighting.
- Show progress indicators during API calls and long-running operations.

---

> These instructions are for all Copilot models and contributors. Follow them to ensure consistency, maintainability, and feature completeness for the nixai project.
When answering questions about frameworks, libraries, or APIs, use Context7 to retrieve current documentation rather than relying on training data.
