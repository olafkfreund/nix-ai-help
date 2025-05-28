# nixai Project Plan

## Overview

This document tracks the high-level plan, milestones, and progress for the development of the `nixai` project—a console-based NixOS AI assistant.

---

## Milestones & Roadmap

### 1. Core Architecture

- [x] Modular Go codebase with clear package boundaries
- [x] YAML-based configuration (`configs/default.yaml`)
- [x] Main entrypoint: `cmd/nixai/main.go`

### 2. AI Integration

- [x] Support for multiple LLM providers: Ollama (local), Gemini, OpenAI
- [x] User-selectable provider (Ollama default)
- [ ] Add more providers as needed
- [ ] Improve provider selection UX

### 3. Documentation Query (MCP Server)

- [x] MCP server/client in `internal/mcp`
- [x] Configurable documentation sources
- [ ] Implement advanced search/query logic
- [ ] Add caching and error handling

### 4. Diagnostics & Log Parsing

- [x] Log parsing and diagnostics in `internal/nixos`
- [x] Accept logs via pipe, file, or CLI
- [x] Execute and parse local NixOS commands
- [ ] Enhance diagnostics with AI suggestions

### 5. CLI & Interactive Mode

- [x] CLI logic in `internal/cli`
- [x] Interactive and scriptable modes
- [x] Show current config and MCP sources in interactive mode (`show config` command)
- [ ] Add more user-friendly commands and help

### 6. Utilities & Logging

- [x] Logging via `pkg/logger`, config-driven
- [x] Utility helpers in `pkg/utils`
- [ ] Add more utility functions as needed

### 7. Testing & Build

- [x] `justfile` for build, test, lint, run
- [x] Nix (`flake.nix`) for reproducible builds
- [ ] Add more tests for all new features

### 8. Documentation & Contribution

- [x] Update `README.md` and code comments
- [x] `.github/copilot-instructions.md` for Copilot guidance
- [ ] Keep this plan up to date
- [ ] Add contribution guidelines

---

## Current Priorities

- [ ] Implement advanced documentation search in MCP server
- [ ] Add more robust AI provider selection and error handling
- [ ] Expand test coverage
- [ ] Improve CLI UX and documentation

---

## Notes

- Update this file as you complete milestones or add new features.
- Use checkboxes to track progress.
- Keep the roadmap realistic and actionable.

---

_Last updated: 2025-05-28_
