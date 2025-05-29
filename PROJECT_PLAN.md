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
- ✅ **AI-Powered NixOS Option Explainer**: `nixai explain-option <option>` provides structured documentation with AI-generated explanations

---

## 📝 Recent Changes (May 2025)

- ➕ Added `--nix-log` (`-g`) flag to `nixai diagnose` to analyze output from `nix log` (optionally with a derivation/path)
- 🧹 Improved search: clean output, numbered results, config option lookup, and config path awareness
- 🔄 All features available in both CLI and interactive modes
- 🏗️ Flake input parser now supports all real-world input forms (attribute sets, comments, whitespace)
- 🤖 `nixai flake explain` and `nixai flake explain-inputs` now provide AI-powered, colorized, terminal-friendly explanations for all flake inputs
- 📖 README and help text updated for new features
- ✅ **NEW: AI-assisted Nix configuration management** with comprehensive `config` command (9 subcommands)
- ✅ **NEW: AI-powered service examples** with `service-examples` command for real-world configurations
- ✅ **NEW: AI-powered config linting** with `lint-config` command for comprehensive analysis
- ✅ **NEW: Enhanced error decoder** with `decode-error` command for human-friendly error explanations
- ✅ **Enhanced justfile** with 40+ comprehensive development commands and categorized help
- ✅ **Fixed interactive mode EOF handling** for proper graceful exit with piped input
- ✅ **Comprehensive testing** with MCP server integration and all features validated

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
- [x] **AI-Powered NixOS Option Explainer** with Elasticsearch backend integration
- [x] (Optional) Use config path for context-aware features everywhere
- [x] (Optional) Automate service option lookup further
- [x] (Optional) Enhance user guidance and error handling for config path
- [x] (Optional) Add more tests for new features

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

## 🚦 Planned: Advanced NixOS User Features

### 1. AI-Powered NixOS Option Explainer ✅ **COMPLETED**

- **Description:** Users can ask about any NixOS option (e.g., `services.nginx.enable`) and get a concise, AI-generated explanation, including type, default, and best practices.

- **Implementation:** ✅ **COMPLETED** & **ENHANCED**
  - ✅ Added `nixai explain-option <option>` command (CLI/interactive).
  - ✅ Integrated MCP server with Elasticsearch backend for structured NixOS option documentation.
  - ✅ AI provider integration for generating human-readable explanations.
  - ✅ Beautiful terminal output with colorized, readable formatting using glamour.
  - ✅ Robust error handling for non-existent options with graceful fallbacks.
  - ✅ Comprehensive testing and debugging completed.
  - 🆕 **Enhanced AI prompts** for comprehensive explanations including:
    - **Usage Examples**: Basic, common, and advanced configuration examples
    - **Best Practices**: Tips, warnings, and recommendations
    - **Related Options**: Other options that work well together
    - **Structured Markdown Output**: Clear headings, code blocks, and formatting
  - 🆕 **Improved User Experience**: Progress indicators, emojis, and helpful tips
  - 🆕 **Interactive Mode Support**: Full explain-option functionality in interactive mode

- **Usage:**

  ```bash
  nixai explain-option services.nginx.enable
  nixai explain-option networking.firewall.enable
  nixai explain-option boot.loader.systemd-boot.enable
  ```

### 8. AI-Assisted Nix Configuration Management ✅ **COMPLETED**

- **Description:** Comprehensive AI-powered configuration management with backup/restore, validation, optimization, and intelligent explanations for all Nix configuration operations.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai config` command with 9 subcommands (CLI/interactive):
    - ✅ `config show` - Display current configuration with AI analysis
    - ✅ `config set <key> <value>` - Set configuration options with AI guidance  
    - ✅ `config unset <key>` - Remove configuration options with safety checks
    - ✅ `config edit` - Open configuration in editor with AI tips
    - ✅ `config explain <key>` - AI-powered explanations of configuration options
    - ✅ `config analyze` - Comprehensive configuration analysis
    - ✅ `config validate` - Validate configuration and suggest improvements
    - ✅ `config optimize` - AI recommendations for performance optimization
    - ✅ `config backup` - Create timestamped configuration backups
    - ✅ `config restore <backup>` - Restore configuration from backup with validation
  - ✅ **Safety Features**: Automatic backups before changes, validation checks, and dry-run testing
  - ✅ **AI Integration**: Intelligent explanations, best practices, and optimization suggestions
  - ✅ **Multi-Config Support**: Works with user configs, system configs, and flake configurations
  - ✅ **Beautiful Output**: Progress indicators, colorized status, and markdown-rendered analysis
  - ✅ **Interactive Mode Support**: Full functionality available in interactive mode with enhanced help

- **Usage:**

  ```bash
  nixai config show                              # Show and analyze current config
  nixai config set experimental-features "nix-command flakes"
  nixai config explain substituters             # Get AI explanation
  nixai config analyze                          # Full configuration analysis
  nixai config validate                         # Validate and suggest improvements
  nixai config optimize                         # Performance optimization tips
  nixai config backup                           # Create backup
  nixai config restore backup-20250529-123456   # Restore from backup
  ```

### 9. AI-Powered Error Decoder ✅ **COMPLETED**

- **Description:** Paste or pipe in a NixOS error message and get a human-friendly explanation with actionable next steps, comprehensive troubleshooting guidance, and prevention tips.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai decode-error <error>` command (CLI/interactive).
  - ✅ **AI-Powered Analysis**: Advanced error pattern recognition and solution generation
  - ✅ **Comprehensive Troubleshooting**: Step-by-step solutions, alternative approaches, and prevention tips
  - ✅ **Documentation Integration**: Links to relevant documentation and resources via MCP server
  - ✅ **Error Classification**: Categorizes errors by type, severity, and complexity
  - ✅ **Context-Aware Solutions**: Provides solutions based on detected system configuration
  - ✅ **Beautiful Terminal Output**: Color-coded analysis with clear action items and progress indicators
  - ✅ **Interactive Mode Support**: Full functionality available in interactive mode

- **Usage:**

  ```bash
  nixai decode-error "syntax error at line 42"
  nixai decode-error "error: function 'buildNodePackage' called without required argument"
  nixai decode-error "error: infinite recursion encountered"
  ```

### 2. AI-Driven NixOS Error Decoder ✅ **COMPLETED** → **See #9 above**

### 3. Interactive NixOS Health Check ✅ **COMPLETED**

- **Description:** `nixai health` runs a series of comprehensive system checks (config validity, service status, disk space, channel status, boot integrity, network connectivity, Nix store health), summarizes findings with beautiful colorized output, and provides AI-powered analysis and recommendations.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai health` command (CLI/interactive).
  - ✅ Comprehensive system health checks including:
    - **Configuration Validation**: Checks NixOS config validity with dry-run
    - **System Services**: Monitors critical service status and failed services
    - **Disk Space Analysis**: Checks disk usage with warnings for high usage
    - **Nix Channels**: Verifies channel configuration and update status
    - **Boot System Integrity**: Validates boot configuration and generations
    - **Network Connectivity**: Tests connectivity to NixOS cache and channels
    - **Nix Store Health**: Verifies store integrity and garbage collection recommendations
  - ✅ **AI-Powered Analysis**: Provides root cause analysis, priority assessment, step-by-step solutions, and prevention tips
  - ✅ **Beautiful Terminal Output**: Progress indicators, colorized status messages, and markdown-rendered reports
  - ✅ **Error Handling**: Graceful handling of missing tools or permissions
  - ✅ **Comprehensive Reporting**: Generates detailed health reports with actionable recommendations

- **Usage:**

  ```bash
  nixai health
  # Runs comprehensive system health check with AI analysis
  ```

### 4. NixOS Upgrade Advisor ✅ **COMPLETED**

- **Description:** Guides users through upgrading NixOS with pre-upgrade checks, backup advice, step-by-step upgrade instructions, and post-upgrade validation, all enhanced with AI-powered explanations and real-time progress feedback.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai upgrade-advisor` command (CLI/interactive).
  - ✅ **Smart Configuration Detection**: Automatically detects flake-based vs traditional configurations
  - ✅ **Conditional Logic**: Different upgrade paths for flake and channel-based configs:
    - **Flake configs**: Uses `nix flake update` and `--flake` flags, skips channel checks
    - **Traditional configs**: Uses `nix-channel --update` and standard rebuild commands
  - ✅ **Comprehensive Pre-Upgrade Checks**: 
    - Configuration file validation, disk space analysis, service status
    - Config validity testing, boot loader verification, network connectivity
    - **Flake-specific**: Input validation and lock file status checks
    - **Traditional-specific**: Channel update checks and availability
  - ✅ **AI-Powered Guidance**: System compatibility analysis, upgrade recommendations, and risk assessment
  - ✅ **Beautiful Progress Feedback**: 7-step detailed progress indicators with real-time status
  - ✅ **Path-Aware Commands**: Automatically adjusts commands to use specified config paths
  - ✅ **Comprehensive Backup Advice**: Pre-upgrade backup checklist and recommendations
  - ✅ **Step-by-Step Instructions**: Detailed upgrade steps with time estimates and safety warnings
  - ✅ **Post-Upgrade Validation**: Complete checklist for verifying successful upgrades
  - ✅ **Enhanced Error Handling**: Clear validation messages and helpful guidance for setup

- **Usage:**

  ```bash
  nixai upgrade-advisor --nixos-path /etc/nixos
  # For flake-based: Uses nix flake update, nixos-rebuild --flake
  # For traditional: Uses nix-channel --update, standard nixos-rebuild
  ```

- **Key Features:**
  - **Flake Detection**: When `flake.nix` is detected, skips channel checks entirely
  - **Conditional Upgrade Steps**: Different command sequences for flake vs traditional configs
  - **Smart Path Integration**: Commands automatically use the provided configuration path
  - **Real-Time Progress**: 7-step analysis process with detailed status indicators
  - **Critical Issue Detection**: Prevents upgrades when critical problems are found

### 5. NixOS Service Usage Examples ✅ **COMPLETED**

- **Description:** For any service, show real-world config examples and explain them with AI-powered analysis and best practices.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai service-examples <service>` command (CLI/interactive).
  - ✅ Comprehensive AI integration for generating real-world configuration examples.
  - ✅ **Multi-Purpose Examples**: Basic setup, common configurations, and advanced use cases
  - ✅ **Best Practices Integration**: Security tips, performance optimizations, and common pitfalls
  - ✅ **Documentation Integration**: Uses MCP server to fetch official documentation
  - ✅ **Beautiful Terminal Output**: Markdown-rendered examples with syntax highlighting
  - ✅ **Interactive Mode Support**: Full functionality available in interactive mode

- **Usage:**

  ```bash
  nixai service-examples nginx
  nixai service-examples postgresql  
  nixai service-examples docker
  nixai service-examples openssh
  ```

### 6. Reverse Option Lookup ✅ **COMPLETED**

- **Description:** Users describe what they want in plain English (e.g., "enable SSH access") and nixai suggests relevant NixOS options and config snippets.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai find-option <description>` command (CLI/interactive).
  - ✅ AI integration to map natural language descriptions to NixOS options.
  - ✅ Comprehensive AI prompts for suggesting relevant options, examples, and best practices.
  - ✅ Beautiful terminal output with markdown rendering.
  - ✅ Interactive mode support with full functionality.

- **Usage:**

  ```bash
  nixai find-option "enable SSH access"
  nixai find-option "configure firewall"
  nixai find-option "set up automatic updates"
  nixai find-option "enable docker"
  ```

### 7. NixOS Config Linter & Formatter ✅ **COMPLETED**

- **Description:** Lint and auto-format NixOS config files, suggesting improvements and flagging anti-patterns with AI-powered analysis.

- **Implementation:** ✅ **COMPLETED**
  - ✅ Added `nixai lint-config <file>` command (CLI/interactive).
  - ✅ **Comprehensive Analysis**: Syntax validation, structure analysis, and best practices checking
  - ✅ **AI-Powered Recommendations**: Security analysis, performance suggestions, and anti-pattern detection
  - ✅ **Multi-File Support**: Works with configuration.nix, flake.nix, home.nix, and other Nix files
  - ✅ **Formatting Suggestions**: Indentation, spacing, and readability improvements
  - ✅ **Security Focus**: Identifies potential security issues and suggests mitigations
  - ✅ **Beautiful Terminal Output**: Color-coded analysis with clear action items
  - ✅ **Interactive Mode Support**: Full functionality available in interactive mode

- **Usage:**

  ```bash
  nixai lint-config /etc/nixos/configuration.nix
  nixai lint-config ./flake.nix
  nixai lint-config /home/user/.config/nixpkgs/home.nix
  ```

---

## 🤝 Contributing

- Follow Go idioms and best practices
- Keep code modular and well-documented
- Add/update tests for all new features and bugfixes
- Update README and this file for any new features or changes

---

> See **README.md** for usage and configuration details.
