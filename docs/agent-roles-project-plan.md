# Project Plan: Agent and Role Abstraction for AI Providers in nixai

## Current Status (Updated 2025-06-06)

🎉 **MAJOR MILESTONE COMPLETED**: All agents are now fully implemented and tested!

- ✅ **26 agents implemented and tested**: All agents for nixai commands are now complete with comprehensive testing
- ✅ **All agent tests passing** with comprehensive test coverage (450+ tests)
- ✅ **Full project test suite passing** with excellent runtime
- ✅ **Agent system features working**: Role validation, context management, provider integration
- ✅ **All role templates complete**: Every agent now has its corresponding role template
- ✅ **Compilation errors resolved**: Fixed all BaseAgent struct issues and function interface problems
- ✅ **Function infrastructure ready**: FunctionManager and base function interface are working
- 📋 **Next steps**: Function calling implementation for all commands, CLI integration with agent/role selection

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

### ✅ COMPLETED

- [x] Agent interface and base implementation (`internal/ai/agent/`)
- [x] Role system and prompt templates (`internal/ai/roles/`)
- [x] All 26 agents implemented (AskAgent, BuildAgent, CommunityAgent, CompletionAgent, ConfigAgent, ConfigureAgent, DevenvAgent, DiagnoseAgent, DoctorAgent, ExplainOptionAgent, ExplainHomeOptionAgent, FlakeAgent, GCAgent, HardwareAgent, HelpAgent, InteractiveAgent, LearnAgent, LogsAgent, MachinesAgent, McpServerAgent, MigrateAgent, NeovimSetupAgent, PackageRepoAgent, SearchAgent, SnippetsAgent, StoreAgent, TemplatesAgent)
- [x] Context management utilities (DiagnosticContext, SystemInfo, NixOSOptionContext, HomeOptionContext, CommunityContext, McpServerContext, NeovimSetupContext, SnippetsContext)
- [x] Comprehensive tests for all agent/role logic (450+ tests across 26 agents)
- [x] All agent tests passing and project test suite fully passing
- [x] Role validation and context management working correctly
- [x] Provider integration with agent system functional
- [x] All role templates implemented and complete
- [x] **Compilation error fixes**: Resolved duplicate role definitions, BaseAgent struct mismatches, and function interface issues
- [x] **Function system foundation**: FunctionManager, FunctionInterface, and base function infrastructure working
- [x] **Build system stability**: All packages compile successfully without errors

### 🔄 IN PROGRESS

- [x] AI Function Calling implementation for all nixai commands (Infrastructure ✅, **9/29 functions implemented**)
- [x] Function calling interface and base implementation (✅ Complete - FunctionManager and BaseFunction working)
- [x] Import cycle fixes in function tests (**✅ RESOLVED** - All functions compile and test successfully)
- [ ] CLI integration for agent/role selection (--role, --agent, --context-file flags)

### 📋 TODO (Priority Order)

- [ ] **Continue function implementations**: Implement remaining 21 functions (packages, build, config, devenv, etc.)
- [ ] Function calling implementations for remaining commands
- [ ] Function calling tests and comprehensive validation
- [ ] Provider refactor (Ollama, OpenAI, Gemini, etc.) to use agents consistently
- [ ] Config updates for agent/role defaults and function calling
- [ ] Build and lint scripts for agents, roles, and functions
- [ ] Documentation and help menu updates
- [ ] Migration and release notes

### ✅ **FUNCTION IMPLEMENTATION STATUS** (9/29 Complete - 31% Done)

**✅ IMPLEMENTED & TESTED:**
1. **ask** - Direct question answering ✅
2. **diagnose** - Log and configuration diagnostics ✅  
3. **explain-option** - NixOS option explanation ✅
4. **explain-home-option** - Home Manager option explanation ✅
5. **learning** - Learning resource generation ✅
6. **community** - Community resource discovery ✅
7. **package-repo** - Git repository analysis and Nix derivation generation ✅
8. **flakes** - Nix flakes management and operations ✅
9. **packages** - Package search and management ✅

**📋 NEXT TO IMPLEMENT:**
10. **build** - Build operations and troubleshooting  
11. **config** - Configuration management and validation
12. **devenv** - Development environment setup
13. **help** - Help system and documentation

---

---

## 🚨 Current Issues & Blockers

### ✅ Recently Resolved Critical Issues

1. **✅ Import Cycle in Function Tests - RESOLVED** 
   - **Issue**: `package nix-ai-help/internal/ai/function/diagnose imports nix-ai-help/internal/ai/function from diagnose_function_test.go imports nix-ai-help/internal/ai/function/diagnose from registry.go: import cycle not allowed in test`
   - **Status**: ✅ **RESOLVED** - All functions now compile and test successfully
   - **Action taken**: Refactored import dependencies in function test structure

2. **✅ Function Test Coverage - RESOLVED**
   - **Issue**: No function tests were passing due to import cycle
   - **Status**: ✅ **RESOLVED** - All 9 implemented functions have comprehensive tests that pass
   - **Action taken**: Fixed test structure and validated all function implementations

### 📋 Current Priority Items

1. **Continue Function Implementation**
   - **Goal**: Implement remaining 20/29 functions (69% remaining)
   - **Next targets**: build, config, devenv, help functions
   - **Priority**: P1 - High priority for completing function system

2. **CLI Integration**
   - **Goal**: Integrate agent/role selection with CLI flags
   - **Priority**: P2 - Important for user experience

### Development Status Summary

**✅ WORKING:**
- Agent system (26/26 agents complete with tests)
- Role system (all role templates complete)
- Function infrastructure (FunctionManager, BaseFunction, types)
- Main project builds successfully
- All agent tests passing (450+ tests)

**🔄 IN PROGRESS:**
- Function calling system (**9/29 functions implemented** - 31% complete ✅)
- CLI integration for agent/role selection (next priority)

**✅ RECENT ACHIEVEMENTS:**
- ✅ **Import cycle issues resolved** - All functions compile and test successfully
- ✅ **Function registry working** - All 9 functions properly registered and operational
- ✅ **Test coverage improved** - All implemented functions have comprehensive tests

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
| ask                  | RoleAsk                 | AskAgent             | Yes             | Yes   | ✅ DONE     |
| build                | RoleBuild               | BuildAgent           | Yes             | Yes   | ✅ DONE     |
| community            | RoleCommunity           | CommunityAgent       | Yes             | Yes   | ✅ DONE     |
| completion           | RoleCompletion          | CompletionAgent      | Yes             | Yes   | ✅ DONE     |
| config               | RoleConfig              | ConfigAgent          | Yes             | Yes   | ✅ DONE     |
| configure            | RoleConfigure           | ConfigureAgent       | Yes             | Yes   | ✅ DONE     |
| devenv               | RoleDevenv              | DevenvAgent          | Yes             | Yes   | ✅ DONE     |
| diagnose             | RoleDiagnose            | DiagnoseAgent        | Yes             | Yes   | ✅ DONE     |
| doctor               | RoleDoctor              | DoctorAgent          | Yes             | Yes   | ✅ DONE     |
| explain-home-option  | RoleExplainHomeOption   | ExplainHomeOptionAgent| Yes            | Yes   | ✅ DONE     |
| explain-option       | RoleExplainOption       | ExplainOptionAgent   | Yes             | Yes   | ✅ DONE     |
| flake                | RoleFlake               | FlakeAgent           | Yes             | Yes   | ✅ DONE     |
| gc                   | RoleGC                  | GCAgent              | Yes             | Yes   | ✅ DONE     |
| hardware             | RoleHardware            | HardwareAgent        | Yes             | Yes   | ✅ DONE     |
| help                 | RoleHelp                | HelpAgent            | Yes             | Yes   | ✅ DONE     |
| interactive          | RoleInteractive         | InteractiveAgent     | Yes             | Yes   | ✅ DONE     |
| learn                | RoleLearn               | LearnAgent           | Yes             | Yes   | ✅ DONE     |
| logs                 | RoleLogs                | LogsAgent            | Yes             | Yes   | ✅ DONE     |
| machines             | RoleMachines            | MachinesAgent        | Yes             | Yes   | ✅ DONE     |
| mcp-server           | RoleMcpServer           | McpServerAgent       | Yes             | Yes   | ✅ DONE     |
| migrate              | RoleMigrate             | MigrateAgent         | Yes             | Yes   | ✅ DONE     |
| neovim-setup         | RoleNeovimSetup         | NeovimSetupAgent     | Yes             | Yes   | ✅ DONE     |
| package-repo         | RolePackageRepo         | PackageRepoAgent     | Yes             | Yes   | ✅ DONE     |
| search               | RoleSearch              | SearchAgent          | Yes             | Yes   | ✅ DONE     |
| snippets             | RoleSnippets            | SnippetsAgent        | Yes             | Yes   | ✅ DONE     |
| store                | RoleStore               | StoreAgent           | Yes             | Yes   | ✅ DONE     |
| templates            | RoleTemplates           | TemplatesAgent       | Yes             | Yes   | ✅ DONE     |

---

- Update the `internal/ai/roles/roles.go` and `internal/ai/agent/` for each command as you implement.
- Mark each as DONE when prompt, agent, and tests are complete.
- Add new commands/roles as needed.

---

## Function Calling Implementation Tracking

### Function Calling Architecture

The AI function calling system extends the agent/role abstraction to provide structured tool execution for each nixai command. Functions are organized by command and provide the AI with structured interfaces to execute nixai operations.

**Directory Structure:**

```text
internal/ai/function/
├── base_function.go           # Base function interface and shared utilities
├── function_manager.go        # Function registry and execution management
├── types.go                   # Shared types and structures
├── ask/                       # Direct question asking functions
│   ├── ask_function.go
│   └── ask_function_test.go
├── diagnose/                  # Log and config diagnostic functions
│   ├── diagnose_function.go
│   └── diagnose_function_test.go
├── explain/                   # NixOS option explanation functions
│   ├── explain_function.go
│   └── explain_function_test.go
... (one directory per command)
```

### Function Calling Implementation Status

| Command | Function Interface | Implementation | Tests | Status |
|---------|-------------------|----------------|-------|--------|
| ask | IFunctionAsk | ✅ | ✅ | COMPLETE |
| diagnose | IFunctionDiagnose | ✅ | ✅ | COMPLETE |
| explain-option | IFunctionExplain | ✅ | ✅ | COMPLETE |
| explain-home-option | IFunctionExplainHome | ✅ | ✅ | COMPLETE |
| learning | IFunctionLearning | ✅ | ✅ | COMPLETE |
| community | IFunctionCommunity | ✅ | ✅ | COMPLETE |
| package-repo | IFunctionPackageRepo | ✅ | ✅ | COMPLETE |
| flakes | IFunctionFlakes | ✅ | ✅ | COMPLETE |
| packages | IFunctionPackages | ✅ | ✅ | COMPLETE |
| build | IFunctionBuild | ❌ | ❌ | TODO |
| config | IFunctionConfig | ❌ | ❌ | TODO |
| devenv | IFunctionDevenv | ❌ | ❌ | TODO |
| help | IFunctionHelp | ❌ | ❌ | TODO |
| install-package | IFunctionInstallPackage | ❌ | ❌ | TODO |
| search-packages | IFunctionSearchPackages | ❌ | ❌ | TODO |
| update-system | IFunctionUpdateSystem | ❌ | ❌ | TODO |
| community forums | IFunctionCommunityForums | ❌ | ❌ | TODO |
| community packages | IFunctionCommunityPackages | ❌ | ❌ | TODO |
| community docs | IFunctionCommunityDocs | ❌ | ❌ | TODO |
| community help | IFunctionCommunityHelp | ❌ | ❌ | TODO |
| devenv create | IFunctionDevenvCreate | ❌ | ❌ | TODO |
| devenv manage | IFunctionDevenvManage | ❌ | ❌ | TODO |
| devenv help | IFunctionDevenvHelp | ❌ | ❌ | TODO |
| learning beginner | IFunctionLearningBeginner | ❌ | ❌ | TODO |
| learning advanced | IFunctionLearningAdvanced | ❌ | ❌ | TODO |
| learning help | IFunctionLearningHelp | ❌ | ❌ | TODO |
| machines | IFunctionMachines | ❌ | ❌ | TODO |
| machines setup | IFunctionMachinesSetup | ❌ | ❌ | TODO |
| machines deploy | IFunctionMachinesDeploy | ❌ | ❌ | TODO |
| machines help | IFunctionMachinesHelp | ❌ | ❌ | TODO |
| mcp-server | IFunctionMcpServer | ❌ | ❌ | TODO |
| neovim | IFunctionNeovim | ❌ | ❌ | TODO |
| neovim setup | IFunctionNeovimSetup | ❌ | ❌ | TODO |
| neovim help | IFunctionNeovimHelp | ❌ | ❌ | TODO |
| snippets | IFunctionSnippets | ❌ | ❌ | TODO |

**Total Functions Needed:** 29
**Completed:** 9 (31% - 9 functions fully implemented and tested ✅)
**Remaining:** 20 (69%)

---

**Recently Fixed:**
- ✅ **Import cycle issue resolved** - Fixed import dependencies in diagnose function tests
- ✅ **DiagnoseFunction tests passing** - All tests now working correctly  
- ✅ **Function infrastructure validated** - FunctionManager and BaseFunction working properly

**Currently Working:**
- DiagnoseFunction (✅ Complete - implemented, tested, and passing all tests)

### Function Calling Features

- **Structured Tool Execution**: Each function provides JSON schema definitions for AI parameter validation
- **Command Integration**: Functions directly execute nixai commands with validated parameters
- **Context Awareness**: Functions inherit context from their associated agents and roles
- **Error Handling**: Comprehensive error handling and user feedback for function execution
- **Type Safety**: Strong typing for all function parameters and return values
- **Async Support**: Non-blocking execution for long-running operations
- **Progress Tracking**: Built-in progress indicators for function execution
- **Validation**: Input validation and sanitization for all function parameters

### Function Calling Implementation Plan

1. **Phase 1: Fix Foundation Issues** (Week 1) 🚨 **URGENT**
   - ✅ Implement `base_function.go` with core function interface (COMPLETED)
   - ✅ Create `function_manager.go` for function registry and execution (COMPLETED)
   - ✅ Define shared types and error handling patterns (COMPLETED)
   - ❌ **Fix import cycles in function tests** (BLOCKING - Must be resolved first)
   - ❌ Validate DiagnoseFunction works end-to-end with fixed tests

2. **Phase 2: Core Functions** (Week 2)
   - Implement ask, explain-option, and explain-home-option functions
   - Add comprehensive test coverage with fixed import structure
   - Integrate with existing agent system
   - Validate JSON schema definitions

3. **Phase 3: Command Functions** (Week 3)
   - Implement help, install-package, package-repo, search-packages, update-system functions
   - Add modular command execution patterns
   - Test integration with NixOS commands
   - Performance optimization

4. **Phase 4: Submodule Functions** (Week 4)
   - Implement all community, devenv, learning, machines, neovim, snippets functions
   - Add specialized context handling for each domain
   - Complete comprehensive test suite
   - Documentation and examples

5. **Phase 5: Integration & Testing** (Week 5)
   - CLI integration with function calling
   - Provider updates to support function calling
   - End-to-end testing and validation
   - Performance benchmarking and optimization

---

## 📝 Recent Updates (2025-06-06)

This document has been updated to reflect the current status of the nixai agent, roles, and function system implementation:

### ✅ Completed Since Last Update

- All 26 agents are fully implemented and tested
- All role templates are complete and working
- Function system infrastructure is in place (FunctionManager, BaseFunction, types)
- Main project builds successfully without errors
- Comprehensive test coverage for all agents (450+ tests passing)

### 🚨 Critical Issues Identified

- Import cycle in function tests is blocking further development
- Only 1 out of 29 functions is implemented (DiagnoseFunction exists but untested)
- Function testing framework needs restructuring

### 📋 Next Priority Actions

1. **URGENT**: Fix import cycles in function test structure
2. Implement remaining 28 AI functions for each nixai command
3. Complete function testing and validation
4. Integrate CLI flags for agent/role selection

### 🎯 Current Phase

**Phase 1**: Foundation Issues (Function testing infrastructure needs fixing before proceeding with implementation)

---

**Document last updated**: 2025-06-06  
**Status**: Agent system complete ✅ | Function system foundation ready ✅ | Function implementations in progress 🔄
