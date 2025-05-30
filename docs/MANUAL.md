# nixai User Manual

Welcome to **nixai** – your AI-powered NixOS assistant for diagnostics, documentation, and automation from the command line. This manual covers all major features, with real-world usage examples for both beginners and advanced users.

> **Latest Update (May 2025)**: Direct question functionality has been added! Ask questions directly with `nixai "your question"` or `nixai --ask "question"`. All three AI providers (Ollama, Gemini, OpenAI) have been comprehensively tested and verified working. MCP server integration provides enhanced documentation retrieval from official NixOS sources.

---

## Table of Contents

- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Basic Setup](#basic-setup)
  - [MCP Server for Documentation](#mcp-server-for-documentation)
  - [Direct Question Assistant](#direct-question-assistant)
- [Diagnosing NixOS Issues](#diagnosing-nixos-issues)
- [Explaining NixOS and Home Manager Options](#explaining-nixos-and-home-manager-options)
- [Searching for Packages and Services](#searching-for-packages-and-services)
- [AI-Powered Package Repository Analysis](#ai-powered-package-repository-analysis)
- [System Health Checks](#system-health-checks)
- [Interactive Mode](#interactive-mode)
- [Advanced Usage](#advanced-usage)
- [Configuration](#configuration)
- [Tips & Troubleshooting](#tips--troubleshooting)

---

## Getting Started

### Prerequisites

- Nix (with flakes enabled)
- Go (if developing outside Nix shell)
- just (for development tasks)
- Ollama (for local LLM inference, recommended)
- git

### Basic Setup

```sh
# Enter the Nix development environment
nix develop

# Build the application
just build

# Run help
./nixai --help
```

### MCP Server for Documentation

nixai integrates with an MCP (Model Context Protocol) server to retrieve official NixOS documentation. For optimal performance, start the MCP server:

```sh
# Start MCP server in background (recommended)
./nixai mcp-server start

# Check server status
./nixai mcp-server status

# Stop server when done
./nixai mcp-server stop
```

The MCP server queries official documentation sources including:
- NixOS Wiki
- Nix Manual
- Nixpkgs Manual  
- Nix Language Reference
- Home Manager Manual

**Note**: The MCP server runs on `localhost:8081` by default and provides enhanced documentation context for all AI providers.

### Direct Question Assistant

The simplest and most direct way to use nixai is by asking questions about NixOS directly from the command line:

```sh
# Ask questions directly by providing them as arguments
./nixai "how do I enable SSH in NixOS?"
./nixai "what is a Nix flake?"
./nixai "how to configure services.postgresql in NixOS?"

# Alternative: use the --ask or -a flag
./nixai --ask "how do I update packages in NixOS?"
./nixai -a "what are NixOS generations?"
```

Both methods are equivalent and provide the same functionality:

1. The question is sent to your configured AI provider (Ollama, Gemini, or OpenAI)
2. If the MCP server is running, it queries relevant NixOS documentation to provide context
3. The AI generates a comprehensive response with practical examples and best practices
4. The response is formatted with proper Markdown rendering in your terminal

**Tips for getting the best results:**

- Be specific in your questions for more targeted responses
- Start the MCP server for documentation-enriched answers
- For complex questions, try to break them down into specific parts
- Use quotes around your question to prevent shell interpretation of special characters

When using the direct question functionality, nixai will:

- Show a progress indicator while retrieving documentation and generating a response
- Format the output as readable, colorized Markdown in your terminal
- Include proper code syntax highlighting for configuration snippets
- Provide links to official documentation when relevant

---

## Diagnosing NixOS Issues

nixai can analyze logs, config snippets, or `nix log` output to help you diagnose problems.

### Basic Example: Diagnose a Log File

```sh
nixai diagnose --log-file /var/log/nixos/nixos-rebuild.log
```

### Diagnose a Nix Log

```sh
nixai diagnose --nix-log /nix/store/xxxx.drv
```

### Diagnose a Config Snippet

```sh
echo 'services.nginx.enable = true;' | nixai diagnose
```

---

## Explaining NixOS and Home Manager Options

Get detailed, AI-powered explanations for any option, including type, default, description, best practices, and usage examples.

### NixOS Option

```sh
nixai explain-option services.nginx.enable
```

### Home Manager Option

```sh
nixai explain-home-option programs.git.enable
```

### Natural Language Query

```sh
nixai explain-option "how to enable SSH access"
```

---

## Searching for Packages and Services

Search for Nix packages or services and get clean, numbered results.

### Search for a Package

```sh
nixai search pkg nginx
```

### Search for a Service

```sh
nixai search service postgresql
```

---

## AI-Powered Package Repository Analysis

Automatically analyze a project and generate a Nix derivation using AI. This feature leverages LLMs to detect the build system, analyze dependencies, and generate a working Nix expression for packaging your project. It works for Go, Python, Node.js, Rust, and more.

### What does it do?
- Detects the language and build system (Go, Python, Node.js, Rust, etc.)
- Analyzes dependencies and project structure
- Generates a complete Nix derivation (e.g., `buildGoModule`, `buildPythonPackage`, `buildNpmPackage`, `buildRustPackage`)
- Extracts metadata (name, version, license, description)
- Suggests best practices and improvements
- Supports both local and remote repositories
- Can output to a custom directory or just analyze without generating a file
- Leverages LLMs (Ollama, OpenAI, Gemini, etc.) and official Nix documentation sources for accuracy
- Handles monorepos and multi-language projects
- Provides AI explanations for packaging challenges and caveats

### How does it work?
- nixai inspects the repository (local path or remote URL), detects the language(s), and parses manifest files (e.g., go.mod, package.json, pyproject.toml, Cargo.toml).
- It queries the selected AI provider, using context from official NixOS documentation, to generate a best-practice Nix derivation.
- The tool can explain why certain packaging choices were made, and highlight any potential issues or manual steps required.

### Real Life Examples

**1. Package a local Go project:**
```sh
nixai package-repo . --local
```
_Output:_
```
Detected Go project (go.mod found)
Analyzing dependencies...
Generated Nix derivation using buildGoModule
Saved to ./default.nix
```

**2. Package a remote Python repository:**
```sh
nixai package-repo https://github.com/psf/requests
```
_Output:_
```
Detected Python project (setup.py found)
Analyzing pip dependencies...
Generated Nix derivation using buildPythonPackage
Saved to ./requests.nix
```

**3. Analyze a Node.js project and output to a custom directory:**
```sh
nixai package-repo https://github.com/expressjs/express --output ./nixpkgs
```
_Output:_
```
Detected Node.js project (package.json found)
Analyzing npm dependencies...
Generated Nix derivation using buildNpmPackage
Saved to ./nixpkgs/express.nix
```

**4. Analyze only, no derivation generation:**
```sh
nixai package-repo . --analyze-only
```
_Output:_
```
Detected Rust project (Cargo.toml found)
Crate dependencies: serde, tokio, clap
Project is ready for packaging. No derivation generated (analyze-only mode).
```

**5. Advanced: Custom package name and output:**
```sh
nixai package-repo https://github.com/user/rust-app --output ./derivations --name my-rust-app
```
_Output:_
```
Detected Rust project (Cargo.toml found)
Analyzing dependencies...
Generated Nix derivation using buildRustPackage
Saved to ./derivations/my-rust-app.nix
```

**6. Multi-language monorepo:**
```sh
nixai package-repo https://github.com/user/monorepo
```
_Output:_
```
Detected multiple projects: Go (api/), Node.js (web/), Python (scripts/)
Generated Nix derivations for each subproject
Saved to ./monorepo-api.nix, ./monorepo-web.nix, ./monorepo-scripts.nix
```

**7. Private repository (with authentication):**
```sh
nixai package-repo git@github.com:yourorg/private-repo.git --ssh-key ~/.ssh/id_ed25519
```
_Output:_
```
Detected Go project (go.mod found)
Analyzing dependencies...
Generated Nix derivation using buildGoModule
Saved to ./private-repo.nix
```

**8. Custom build system (Makefile, CMake, etc.):**
```sh
nixai package-repo . --analyze-only
```
_Output:_
```
Detected C project (Makefile found)
AI Suggestion: Use stdenv.mkDerivation with custom buildPhase and installPhase
Manual review recommended for non-standard build systems.
```

**9. Output as JSON for CI/CD integration:**
```sh
nixai package-repo . --output-format json
```
_Output:_
```
{
  "project": "myapp",
  "language": "go",
  "derivation": "...nix expression...",
  "dependencies": ["github.com/foo/bar", "github.com/baz/qux"]
}
```

### What kind of output can I expect?
- A ready-to-use Nix derivation file (e.g., `default.nix`, `myapp.nix`)
- AI-generated explanations for any manual steps or caveats
- Dependency analysis and best-practice suggestions
- Optionally, JSON output for automation

### Best Practices & Troubleshooting
- For best results, ensure your project has a standard manifest (go.mod, package.json, pyproject.toml, etc.)
- For private repos, use `--ssh-key` or ensure your SSH agent is running
- If the generated derivation fails to build, review the AI explanation and check for missing dependencies or custom build steps
- Use `--analyze-only` to preview before generating files
- For monorepos, review each generated derivation for correctness
- If you hit LLM rate limits or errors, try switching providers or check your API key/config
- Always review the generated Nix code before using in production

### How does this help NixOS users?
- Saves hours of manual packaging work
- Handles complex dependency trees automatically
- Ensures best practices for Nix packaging
- Works for both simple and complex/multi-language projects
- Great for onboarding new projects to Nix or sharing reproducible builds
- Provides AI explanations and links to relevant NixOS documentation for further learning

---

## System Health Checks

Run a comprehensive health check on your NixOS system to detect common issues, misconfigurations, and get actionable AI-powered recommendations.

```sh
nixai health
```

### What does `nixai health` check?
- NixOS configuration validity
- Service status (e.g., failed or inactive services)
- Disk space and filesystem health
- Outdated or broken packages
- Security warnings and best practices
- AI-powered suggestions for improving system reliability and security

### Real Life Examples

**Basic health check:**
```sh
nixai health
```
_Output:_
```
[✓] NixOS configuration is valid
[✓] All critical services are running
[!] Disk space low on /home (2% free)
[!] 3 packages are outdated
[AI] Suggestion: Consider enabling automatic updates for security
```

**Check health for a custom NixOS config path:**
```sh
nixai health --nixos-path ~/my-nixos-config
```

**Get detailed output and recommendations:**
```sh
nixai health --log-level debug
```

**Example: Fixing a failed service**
```
> nixai health
[!] Service 'nginx' is failed
[AI] Suggestion: Check nginx config syntax with `nginx -t` and review recent changes in /etc/nixos/configuration.nix
```

**Example: Security recommendations**
```
> nixai health
[AI] Security: SSH root login is enabled. It is recommended to set `services.openssh.permitRootLogin = "no";`
```

---

## Interactive Mode

Launch an interactive shell for all features:

```sh
nixai interactive
```

In interactive mode, you can:
- Run any command (diagnose, explain-option, search, etc.) in a conversational, guided way.
- Use tab completion for commands and options.
- Get instant feedback and suggestions for next steps.
- Set or change your NixOS config path on the fly with:
  ```sh
  set-nixos-path /etc/nixos
  ```
- Switch AI providers interactively:
  ```sh
  set ai openai
  set ai ollama llama3
  set ai gemini
  ```
- View and update configuration settings:
  ```sh
  show-config
  set-log-level debug
  ```
- Get help for any command:
  ```sh
  help explain-option
  help diagnose
  ```

### Real Life Examples

**Diagnose a log interactively:**
```
> diagnose
Paste or type your log, or use --log-file: /var/log/nixos/nixos-rebuild.log
[AI-powered analysis and suggestions appear]
```

**Explain a NixOS option:**
```
> explain-option networking.firewall.enable
[AI-powered explanation, best practices, and examples]
```

**Search for a package and get config options:**
```
> search pkg nginx
[Numbered results]
> 1
[Shows config/test options for nginx]
```

**Switch provider and run a command:**
```
> set ai gemini
> explain-option services.openssh.enable
```

**Get help and see available commands:**
```
> help
[Lists all available interactive commands]
```

Interactive mode is ideal for:
- New users who want a guided experience
- Power users who want to chain commands and get instant feedback
- Troubleshooting and exploring NixOS options, logs, and documentation in a conversational way

---

## Advanced Usage

### Specify NixOS Config Path

```sh
nixai search --nixos-path /etc/nixos pkg nginx
```

### Use a Different AI Provider

```sh
nixai diagnose --provider openai --log-file /var/log/nixos/nixos-rebuild.log
```

### Get Examples for a Service

```sh
nixai service-examples nginx
```

### Flake Input Analysis & AI Explanations

```sh
nixai flake explain --flake /etc/nixos/flake.nix
```

### Pipe Logs or Configs Directly

You can pipe logs or configuration snippets directly into nixai for diagnosis or explanation:

```sh
journalctl -xe | nixai diagnose
cat /etc/nixos/configuration.nix | nixai explain-option
```

### Use Custom AI Model or Parameters

Override the AI model or set advanced parameters at runtime:

```sh
nixai diagnose --provider ollama --model llama3 --temperature 0.2 --log-file /var/log/nixos/nixos-rebuild.log
```

### Analyze and Package a Private Repo with SSH

```sh
nixai package-repo git@github.com:yourorg/private-repo.git --ssh-key ~/.ssh/id_ed25519
```

### Output as JSON for Automation

```sh
nixai package-repo . --output-format json
```

### Use in Scripts or CI/CD

nixai is scriptable and can be used in CI/CD pipelines for automated diagnostics, health checks, or packaging:

```sh
nixai health --output-format json > health_report.json
nixai package-repo . --analyze-only --output-format json > analysis.json
```

### Interactive Mode Power Tips

- Use `set-nixos-path` and `set ai` to change context on the fly.
- Use tab completion for commands and options.
- Use `show-config` to review current settings.
- Use `help <command>` for detailed help on any feature.

### Troubleshooting Advanced Scenarios

- If you encounter API rate limits, try switching providers or lowering request frequency.
- For complex monorepos, review each generated derivation and consult the AI explanations for manual steps.
- For custom build systems, use `--analyze-only` and follow AI suggestions for manual packaging tweaks.
- Always validate generated Nix code with `nix build` or `nix flake check` before deploying.

---

## Configuration

nixai uses a YAML config file (usually at `~/.config/nixai/config.yaml`). You can set:

- Preferred AI provider/model (Ollama, OpenAI, Gemini)
- NixOS config folder
- Log level
- Documentation sources

### Choosing and Configuring Your AI Provider

nixai supports multiple AI providers. You can select your provider in the config file or via the `--provider` CLI flag:

- **Ollama** (default, local, privacy-first)
- **OpenAI** (cloud, requires API key)
- **Gemini** (cloud, requires API key)

#### Provider Feature Comparison

Based on comprehensive testing (May 2025), all three providers are fully functional:

| Feature | Ollama | Gemini | OpenAI |
|---------|--------|--------|--------|
| Privacy | ✅ Local | ❌ Cloud | ❌ Cloud |
| API Key Required | ❌ No | ✅ Yes | ✅ Yes |
| Speed | ⚡ Fast | ⚡ Fast | ⚡ Fast |
| Quality | ✅ Good | ✅ Excellent | ✅ Excellent |
| Cost | 💚 Free | 💰 Paid | 💰 Paid |
| Setup | 🔧 Requires Ollama | 🔧 API Key | 🔧 API Key |
| **Recommended For** | Privacy & Development | Production & Quality | Production & Quality |

**Testing Status**: ✅ All providers tested and working with `explain-option`, `find-option`, and interactive mode commands.

#### Prerequisites for Each Provider

- **Ollama**: Install [Ollama](https://ollama.com/) and pull the desired model (e.g., `ollama pull llama3`). No API key required. Runs locally.
  - **Default Model**: llama3 (automatically used when no model specified)
  - **Tested**: ✅ Working with llama3 model
  
- **OpenAI**: Requires an OpenAI API key. Sign up at [OpenAI](https://platform.openai.com/). Set your API key as an environment variable:
  ```sh
  export OPENAI_API_KEY=sk-...
  ```
  - **Default Model**: Uses OpenAI's default GPT model
  - **Tested**: ✅ Working with environment variable configuration
  
- **Gemini**: Requires a Gemini API key. Sign up at [Google AI Studio](https://ai.google.dev/). Set your API key as an environment variable:
  ```sh
  export GEMINI_API_KEY=your-gemini-key
  ```
  - **Current Model**: gemini-1.5-flash (updated from deprecated gemini-pro)
  - **Tested**: ✅ Working with updated API endpoints and model

#### Example config for OpenAI or Gemini

```yaml
ai_provider: openai   # or 'gemini' or 'ollama'
ai_model: gpt-4       # or 'llama3', 'gemini-1.5-flash', etc.
# ...other config options...
```

You can also override the provider and model at runtime:

```sh
nixai diagnose --provider openai --model gpt-4 --log-file /var/log/nixos/nixos-rebuild.log
nixai explain-option --provider gemini --model gemini-1.5-flash networking.firewall.enable
```

**Note:**
- If using OpenAI or Gemini, the API key must be set in your environment or in the config file under `openai_api_key` or `gemini_api_key` (environment variable is preferred for security).
- If no provider is set, Ollama is used by default for privacy.

### Example config with API keys (not recommended, prefer env vars):
```yaml
ai_provider: openai
ai_model: gpt-4
openai_api_key: sk-...
```

---

## Recent Testing & Validation

**Last Updated**: May 2025

nixai has been comprehensively tested with all three AI providers to ensure reliability and functionality:

### ✅ Verified Working Commands

All commands tested successfully across all providers:

```sh
# Explain NixOS options
./nixai explain-option services.nginx.enable
./nixai explain-option services.openssh.enable

# Find options using natural language
./nixai find-option "enable SSH"

# Interactive mode with provider switching
./nixai interactive
> set ai ollama llama3
> set ai gemini
> set ai openai
> explain-option services.nginx.enable
```

### 🔧 Key Fixes Applied

- **Ollama Model Handling**: Fixed empty model configuration by defaulting to "llama3"
- **Gemini API Updates**: Updated from deprecated `gemini-pro` to `gemini-1.5-flash` model
- **API Endpoints**: Fixed Gemini API URL construction for proper integration
- **MCP Server**: Validated documentation retrieval from official NixOS sources

### 📊 Current Working Configuration

```yaml
ai_provider: ollama    # Default for privacy
ai_model: llama3      # Auto-selected for Ollama
nixos_folder: ~/nixos-config
log_level: debug
mcp_server:
    host: localhost
    port: 8081
```

### 🚀 Provider Switching

You can seamlessly switch between providers:

```sh
# Via interactive mode
echo "set ai gemini" | ./nixai interactive

# Via command line flags  
./nixai explain-option --provider openai services.nginx.enable
```

---

## Tips & Troubleshooting

- If you see errors about missing config path, set it with `--nixos-path` or in your config file.
- For best privacy, use Ollama as your AI provider (local inference).
- Use `just lint` and `just test` for code quality and reliability.
- All features are available in both CLI and interactive modes.

---

> _nixai: Your AI-powered NixOS assistant, right in your terminal._
