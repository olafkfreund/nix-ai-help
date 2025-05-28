# Copilot Instructions for the nixai Project

## Project Purpose
- This project is a console-based Linux application to help solve NixOS configuration problems and assist in creating and configuring NixOS from the command line.
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
- All new features must be testable and documented.

## Best Practices
- Keep all code modular and well-documented.
- Prefer local inference (Ollama) for privacy, but support cloud LLMs as fallback.
- Always validate and sanitize user input and log data.
- Use context and error handling idiomatically in Go.
- Write clear, maintainable, and idiomatic Go code.
- Add or update tests for all new features and bugfixes.

## Testing & Build
- Use the `justfile` for common tasks: build, test, lint, run, etc.
- Use Nix (`flake.nix`) for reproducible builds and development environments.

## Documentation
- Update `README.md` and code comments for any new features or changes.
- Keep this instruction file up to date as the project evolves.

---

> These instructions are for all Copilot models and contributors. Follow them to ensure consistency, maintainability, and feature completeness for the nixai project.
