# Project Plan: Agent and Role Abstraction for AI Providers in nixai

## Overview

This project introduces an "agent" abstraction layer for AI providers in nixai, enabling advanced context management, role-based prompt engineering, and improved orchestration across multiple LLMs. The agent system will allow for specialized behaviors (e.g., diagnoser, explainer, searcher) and more intelligent context handling, making AI interactions more powerful and modular.

---

## Motivation & Goals

- **Contextual Intelligence**: Agents can manage and inject relevant context (logs, configs, docs) into prompts, improving answer quality.
- **Role-based Behavior**: Define roles (diagnoser, explainer, searcher, etc.) to tailor LLM responses to user intent.
- **Provider Orchestration**: Agents can select/fallback between providers based on role, context, or user preference.
- **Extensibility**: New roles and agent types can be added without major refactoring.
- **Testability**: Agents encapsulate logic, making it easier to test and mock behaviors.

---

## Design

### 1. Agent Interface

- Define an `Agent` interface in `internal/ai/agent/`:
  - `Query(ctx, input, role, context) (string, error)`
  - `GenerateResponse(ctx, input, role, context) (string, error)`
  - `SetRole(role string)`
  - `SetContext(context interface{})`
- Each provider implements the Agent interface, supporting roles and context injection.

### 2. Role Definitions

- Roles are enums/strings: `diagnoser`, `explainer`, `searcher`, `summarizer`, etc.
- Each role has prompt templates and context requirements.
- Role logic lives in `internal/ai/roles/`.

### 3. Context Management

- Agents accept structured context (logs, configs, docs, etc.).
- Context is validated and sanitized before prompt construction.
- Use `pkg/utils` for formatting and context helpers.

### 4. Integration Points

- **internal/ai/agent/**: Core agent logic, provider wrappers.
- **internal/ai/roles/**: Role templates, prompt logic, and context requirements.
- **internal/cli**: CLI flags for role/agent selection, context passing.
- **configs/**: YAML schema update to support agent/role defaults.
- **pkg/logger**: Log agent/role selection and context usage.

### 5. CLI & User Experience

- New flags: `--role`, `--agent`, `--context-file`.
- Help menus updated with agent/role usage examples.
- Progress indicators for agent operations.

---

## Step-by-Step Implementation Plan

1. **Directory Structure**
   - Create `internal/ai/agent/` for all agent implementations and logic.
   - Create `internal/ai/roles/` for all role definitions, templates, and logic.
   - Add `.instructions.md` Copilot instruction files in both folders to track process and best practices.

2. **Agent Interface & Base Implementation**
   - Define an `Agent` interface in `internal/ai/agent/agent.go`.
   - Implement a base agent and at least one provider-backed agent (e.g., OllamaAgent).
   - Add build, lint, and test scripts for agents.

3. **Role System**
   - Define roles as enums/strings: `diagnoser`, `explainer`, `searcher`, `summarizer`, etc.
   - Implement role logic, prompt templates, and context requirements in `internal/ai/roles/`.
   - Add build, lint, and test scripts for roles.
   - Document each role in a Copilot instruction file in `internal/ai/roles/.instructions.md`.

4. **Integration Points**
   - Update provider logic to use the Agent interface.
   - Ensure agents can select and use roles dynamically.
   - Add context validation and formatting using `pkg/utils`.
   - Log agent/role selection and context usage with `pkg/logger`.

5. **CLI & Config Updates**
   - Add CLI flags: `--role`, `--agent`, `--context-file`.
   - Update help menus with agent/role usage examples.
   - Extend YAML config to support agent/role defaults and options.

6. **Testing & Quality**
   - Add unit and integration tests for each agent and role.
   - Ensure all new code passes linting and build checks (update `justfile` as needed).
   - Add progress indicators for agent operations.

7. **Documentation**
   - Update `README.md`, `docs/MANUAL.md`, and help menus with agent/role usage and examples.
   - Maintain `.instructions.md` in both `agent/` and `roles/` to document process, best practices, and progress.

8. **Migration & Backward Compatibility**
   - Ensure backward compatibility with existing provider logic.
   - Provide migration notes and update documentation as needed.

---

## Milestones & Deliverables

- [ ] Agent interface and base implementation (`internal/ai/agent/`)
- [ ] Role system and prompt templates (`internal/ai/roles/`)
- [ ] Provider refactor (Ollama, OpenAI, Gemini, etc.) to use agents and roles
- [ ] Context management utilities
- [ ] Config and CLI updates for agent/role selection
- [ ] Tests for agent/role logic and provider compliance
- [ ] Build and lint scripts for agents and roles
- [ ] Documentation and help menu updates
- [ ] Copilot instruction files in both `agent/` and `roles/` folders
- [ ] Migration and release notes

---

## Risks & Mitigations

- **Complexity**: Keep agent/role logic modular and well-documented. Separate agent and role logic into their respective folders for clarity.
- **Provider Compatibility**: Test all providers for compliance with new agent and role interfaces.
- **User Experience**: Provide clear help, examples, and migration guidance for both agent and role usage.

---

## Progress Tracking

- Use this document to check off milestones and add notes as work progresses.
- Update `.instructions.md` files in both `internal/ai/agent/` and `internal/ai/roles/` to document implementation details, best practices, and lessons learned for agents and roles.
- Update related documentation as features are implemented.

---

## Contributors

- Please add your name and date when you contribute to this project plan or implementation.

---
## Command-to-Role/Agent Mapping Progress

Below is the tracking table for agent/role implementation for each nixai command. Update this as you implement and test each one.

| Command              | RoleType                | Agent Implementation | Prompt Template | Tests | Status      |
|----------------------|-------------------------|----------------------|-----------------|-------|-------------|
| ask                  | RoleAsk                 | AskAgent             | Yes             | No    | TODO        |
| build                | RoleBuild               | BuildAgent           | Yes             | No    | TODO        |
| community            | RoleCommunity           | CommunityAgent       | Yes             | No    | TODO        |
| completion           | RoleCompletion          | CompletionAgent      | Yes             | No    | TODO        |
| config               | RoleConfig              | ConfigAgent          | Yes             | No    | TODO        |
| configure            | RoleConfigure           | ConfigureAgent       | Yes             | No    | TODO        |
| devenv               | RoleDevenv              | DevenvAgent          | Yes             | No    | TODO        |
| diagnose             | RoleDiagnose            | DiagnoseAgent        | Yes             | Yes   | DONE        |
| doctor               | RoleDoctor              | DoctorAgent          | Yes             | No    | TODO        |
| explain-home-option  | RoleExplainHomeOption   | ExplainHomeOptionAgent| Yes            | No    | TODO        |
| explain-option       | RoleExplainOption       | ExplainOptionAgent   | Yes             | No    | TODO        |
| flake                | RoleFlake               | FlakeAgent           | Yes             | No    | TODO        |
| gc                   | RoleGC                  | GCAgent              | Yes             | No    | TODO        |
| hardware             | RoleHardware            | HardwareAgent        | Yes             | No    | TODO        |
| help                 | RoleHelp                | HelpAgent            | Yes             | No    | TODO        |
| interactive          | RoleInteractive         | InteractiveAgent     | Yes             | No    | TODO        |
| learn                | RoleLearn               | LearnAgent           | Yes             | No    | TODO        |
| logs                 | RoleLogs                | LogsAgent            | Yes             | No    | TODO        |
| machines             | RoleMachines            | MachinesAgent        | Yes             | No    | TODO        |
| mcp-server           | RoleMcpServer           | McpServerAgent       | Yes             | No    | TODO        |
| migrate              | RoleMigrate             | MigrateAgent         | Yes             | No    | TODO        |
| neovim-setup         | RoleNeovimSetup         | NeovimSetupAgent     | Yes             | No    | TODO        |
| package-repo         | RolePackageRepo         | PackageRepoAgent     | Yes             | No    | TODO        |
| search               | RoleSearch              | SearchAgent          | Yes             | No    | TODO        |
| snippets             | RoleSnippets            | SnippetsAgent        | Yes             | No    | TODO        |
| store                | RoleStore               | StoreAgent           | Yes             | No    | TODO        |
| templates            | RoleTemplates           | TemplatesAgent       | Yes             | No    | TODO        |

---

- Update the `internal/ai/roles/roles.go` and `internal/ai/agent/` for each command as you implement.
- Mark each as DONE when prompt, agent, and tests are complete.
- Add new commands/roles as needed.

*Last updated: 2025-06-05*
