# nixai User Manual

Welcome to **nixai** – your AI-powered NixOS assistant for diagnostics, documentation, and automation from the command line. This manual covers all major features, with real-world usage examples for both beginners and advanced users.

> **Latest Update (May 2025)**: Major improvements to documentation display with HTML filtering, enhanced terminal formatting, and comprehensive editor integration. The `explain-option` and `explain-home-option` commands now provide clean, beautifully formatted output with all HTML artifacts removed. Direct question functionality has been enhanced with better error handling and documentation context retrieval. All three AI providers (Ollama, Gemini, OpenAI) have been comprehensively tested and verified working.

---

## 🆕 Recent Improvements & Features

### Documentation Display Enhancements (May 2025)

- **🧹 HTML Filtering**: Complete removal of HTML tags, DOCTYPE declarations, wiki navigation elements, and raw content from all documentation output
- **🎨 Enhanced Formatting**: Consistent use of headers, dividers, key-value pairs, and glamour markdown rendering for improved readability
- **🏠 Smart Option Detection**: Automatic visual distinction between NixOS options (`🖥️ NixOS Option`) and Home Manager options (`🏠 Home Manager Option`)
- **🔧 Robust Error Handling**: Better error messages, graceful fallbacks when MCP server is unavailable, and clear feedback for configuration issues
- **🧪 Comprehensive Testing**: All improvements are backed by extensive test coverage to ensure reliability

### Core Capabilities

- **🤖 Direct Question Assistant**: Ask questions directly with `nixai "your question"` for instant AI-powered help
- **📖 Documentation Integration**: Enhanced MCP server integration for official NixOS documentation retrieval
- **🔌 Editor Integration**: Full support for Neovim and VS Code with automatic setup and configuration
- **📦 Package Analysis**: AI-powered repository analysis with Nix derivation generation
- **🔍 Option Explanation**: Comprehensive explanations for NixOS and Home Manager options with examples and best practices

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
- [Configuration Templates & Snippets](#configuration-templates--snippets)
- [Interactive Mode](#interactive-mode)
- [Editor Integration](#editor-integration)
  - [Neovim Integration](#neovim-integration)
- [Advanced Usage](#advanced-usage)
  - [Enhanced Build Troubleshooter](#enhanced-build-troubleshooter)
  - [Dependency & Import Graph Analyzer](#dependency--import-graph-analyzer)
- [Configuration](#configuration)
- [Tips & Troubleshooting](#tips--troubleshooting)
- [Development Environment (devenv) Feature](#development-environment-devenv-feature)
- [Neovim + Home Manager Integration](#neovim--home-manager-integration)
- [Migration Assistant](#migration-assistant)

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

Get detailed, AI-powered explanations for any option, including type, default, description, best practices, and usage examples. The explanation output now features enhanced HTML filtering and beautiful terminal formatting for improved readability.

### Enhanced Documentation Display

As of May 2025, the `explain-option` and `explain-home-option` commands feature significant improvements:

- **🧹 Complete HTML Filtering:** Removes all HTML tags, DOCTYPE declarations, wiki navigation elements, and raw content
- **🎨 Beautiful Formatting:** Consistent headers, dividers, key-value pairs, and glamour markdown rendering
- **🏠 Smart Detection:** Automatic visual distinction between NixOS options (`🖥️ NixOS Option`) and Home Manager options (`🏠 Home Manager Option`)
- **📖 Clean Documentation:** Official documentation is filtered and formatted for optimal terminal display
- **🔧 Robust Error Handling:** Graceful fallbacks when documentation sources are unavailable

### NixOS Option

```sh
nixai explain-option services.nginx.enable
```

**Example Output:**

```
🖥️ NixOS Option: services.nginx.enable

📋 Option Information
├─ Type: boolean
├─ Default: false
└─ Source: /nix/store/.../nixos/modules/services/web-servers/nginx.nix

📖 Documentation
Whether to enable the nginx Web Server.

🤖 AI Explanation
[Detailed AI-generated explanation with examples and best practices]

💡 Usage Examples
[Basic, common, and advanced configuration examples]
```

### Home Manager Option

```sh
nixai explain-home-option programs.git.enable
```

**Example Output:**

```
🏠 Home Manager Option: programs.git.enable

📋 Option Information  
├─ Type: boolean
├─ Default: false
└─ Module: programs.git

📖 Documentation
Whether to enable Git, a distributed version control system.

🤖 AI Explanation
[Detailed explanation specific to Home Manager context]

💡 Usage Examples
[Home Manager-specific configuration examples]
```

### Natural Language Query

```sh
nixai explain-option "how to enable SSH access"
```

The system intelligently maps natural language queries to appropriate NixOS or Home Manager options and provides comprehensive explanations.

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

The `nixai health` command provides comprehensive health checks for your NixOS system, with AI-powered analysis and recommendations.

### What does `nixai health` check?

- NixOS configuration validity
- Failed system services
- Disk space usage
- Nix channel status
- Boot integrity
- Network connectivity
- Nix store health

### Real Life Examples

```sh
# Run comprehensive health check
nixai health

# Output includes:
# ✓ Configuration validation passed
# ⚠ Low disk space on /nix/store (85% full)
# ✗ Service postgresql.service failed
# ✓ Network connectivity OK
# ... plus AI recommendations
```

**Example: Fixing a failed service**

```
⚠ Failed Services:
   postgresql.service - PostgreSQL database server

💡 AI Analysis:
   PostgreSQL service failure is commonly caused by:
   1. Incorrect data directory permissions
   2. Port conflicts with existing services
   3. Invalid configuration syntax
   
   Recommended actions:
   1. Check service logs: systemctl status postgresql.service
   2. Verify data directory: ls -la /var/lib/postgresql/
   3. Review configuration: nixos-option services.postgresql
```

**Example: Security recommendations**

```
🔒 Security Analysis:
   - SSH root login is enabled (consider disabling)
   - Firewall has open ports: 22, 80, 443, 8080
   - Automatic updates are disabled
   
   Recommendations:
   1. Disable SSH root login: services.openssh.permitRootLogin = "no";
   2. Review open ports and close unnecessary ones
   3. Enable automatic security updates
```

---

## Configuration Templates & Snippets

nixai provides a powerful template and snippet management system to help you quickly set up and reuse NixOS configurations. This feature includes curated templates for common setups and personal snippet management for your custom configurations.

### Templates

Templates are pre-built NixOS configurations for common use cases. nixai includes built-in templates and can search GitHub for real-world configurations.

#### Browsing Templates

```sh
# List all available templates
nixai templates list

# Show template categories
nixai templates categories

# Search templates by keyword
nixai templates search gaming
nixai templates search desktop kde
nixai templates search server nginx
```

#### Viewing Template Details

```sh
# Show complete template details and content
nixai templates show desktop-minimal
nixai templates show gaming-config
nixai templates show server-basic
```

#### Applying Templates

```sh
# Apply template to default location (/etc/nixos/configuration.nix)
nixai templates apply desktop-minimal

# Apply template to specific file
nixai templates apply gaming-config --output ./gaming.nix

# Merge template with existing configuration
nixai templates apply server-basic --merge --output /etc/nixos/server.nix
```

#### GitHub Integration

Search GitHub for real-world NixOS configurations:

```sh
# Search GitHub for configurations
nixai templates github "gaming nixos configuration"
nixai templates github "kde plasma nixos"
nixai templates github "thinkpad nixos hardware"
nixai templates github "server nginx configuration.nix"
```

#### Saving Custom Templates

```sh
# Save local configuration file as template
nixai templates save my-desktop /etc/nixos/configuration.nix --category Desktop --description "My custom desktop setup"

# Save from GitHub URL
nixai templates save nvidia-gaming https://github.com/user/repo/blob/main/nixos/gaming.nix --category Gaming

# Add tags for better organization
nixai templates save my-template config.nix --tags "nvidia,gaming,performance"
```

### Snippets

Snippets allow you to save and reuse small configuration fragments. Perfect for commonly used service configurations, hardware settings, or package lists.

#### Managing Snippets

```sh
# List all saved snippets
nixai snippets list

# Search snippets by name or tag
nixai snippets search nvidia
nixai snippets search gaming
nixai snippets search development
```

#### Adding Snippets

```sh
# Save configuration file as snippet
nixai snippets add nvidia-drivers --file ./hardware.nix --description "NVIDIA driver configuration" --tags "nvidia,graphics"

# Save from stdin (pipe content)
cat hardware-config.nix | nixai snippets add my-hardware --description "Custom hardware config"

# Add snippet with multiple tags
nixai snippets add gaming-packages --file packages.nix --tags "gaming,steam,lutris"
```

#### Using Snippets

```sh
# Show snippet content
nixai snippets show nvidia-drivers

# Apply snippet to file
nixai snippets apply gaming-setup --output ./gaming.nix

# Apply snippet to stdout (for copying)
nixai snippets apply my-config
```

#### Organizing Snippets

```sh
# Remove old or unused snippets
nixai snippets remove old-config

# Search by multiple criteria
nixai snippets search "nvidia AND gaming"
```

### Built-in Templates

nixai includes curated templates for common NixOS configurations:

#### Desktop Environments

- **desktop-minimal**: Minimal GNOME desktop with essential applications
- **desktop-kde**: Full KDE Plasma desktop environment
- **desktop-xfce**: Lightweight XFCE desktop setup

#### Gaming Configurations

- **gaming-config**: Gaming-optimized configuration with Steam, drivers, and performance tweaks
- **gaming-nvidia**: NVIDIA-specific gaming setup with proper drivers and settings

#### Server Configurations

- **server-basic**: Basic server with SSH, firewall, and essential tools
- **server-web**: Web server with nginx, SSL, and security hardening
- **server-database**: Database server with PostgreSQL or MySQL

#### Development Environments

- **development-env**: Development setup with common programming tools, git, and editors
- **development-nix**: Nix development environment for nixpkgs contributions

### Real-World Examples

#### Setting up a Gaming System

```sh
# Browse gaming templates
nixai templates search gaming

# View gaming template details
nixai templates show gaming-config

# Apply gaming template
nixai templates apply gaming-config --output /etc/nixos/gaming.nix

# Save your custom gaming tweaks as snippet
nixai snippets add my-gaming-tweaks --file ./performance.nix --tags "gaming,performance"
```

#### Creating a Development Environment

```sh
# Search for development templates
nixai templates search development

# Look for specific language setups on GitHub
nixai templates github "rust development nixos"

# Apply development template
nixai templates apply development-env --merge

# Save your editor configuration as snippet
nixai snippets add neovim-config --file ./editor.nix --tags "neovim,development"
```

#### Server Setup Workflow

```sh
# Browse server templates
nixai templates categories
nixai templates search server

# Apply basic server template
nixai templates apply server-basic --output /etc/nixos/server.nix

# Add specific service snippets
nixai snippets apply nginx-config --output ./services.nix
nixai snippets apply postgresql-setup --output ./database.nix
```

### Tips and Best Practices

- **Use Categories**: Templates are organized by category (Desktop, Gaming, Server, Development) for easy browsing
- **Tag Everything**: Use descriptive tags when saving templates and snippets for better searchability
- **GitHub Discovery**: Use the GitHub search to find real-world configurations for inspiration
- **Incremental Building**: Start with a base template and add snippets to customize your configuration
- **Backup First**: Always backup your existing configuration before applying templates
- **Test Configurations**: Use `nixos-rebuild test` to test configurations before committing them

### Integration with Other Features

The template and snippet system integrates seamlessly with other nixai features:

- **Health Checks**: `nixai health` can validate templates before applying them
- **Option Explanation**: Use `nixai explain-option` to understand options used in templates
- **Direct Questions**: Ask questions about template configurations with `nixai "how does this gaming template work?"`

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

## Editor Integration

nixai provides seamless integration with popular editors through the MCP (Model Context Protocol) server, enabling you to access NixOS documentation and AI assistance directly within your development environment. The integration supports both Neovim and VS Code with automatic setup and configuration.

### VS Code Integration

Complete VS Code integration with Copilot, Claude Dev, and other MCP-compatible extensions for in-editor NixOS assistance.

#### Quick Setup

1. **Start the MCP server:**

```sh
# Start the server in background mode
nixai mcp-server start -d

# Check server status
nixai mcp-server status
```

2. **Install required VS Code extensions:**
   - `automatalabs.copilot-mcp` - Copilot MCP extension
   - `zebradev.mcp-server-runner` - MCP Server Runner  
   - `saoudrizwan.claude-dev` - Claude Dev (Cline)

3. **Configure VS Code settings:**

Add to your `.vscode/settings.json`:

```json
{
  "mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "copilot.mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "claude-dev.mcpServers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "mcp.enableDebug": true,
  "claude-dev.enableMcp": true
}
```

#### Using nixai in VS Code

Once configured, you can:

- **Ask Copilot about NixOS:** Chat with GitHub Copilot and ask NixOS-specific questions - it will automatically query nixai's documentation
- **Use Claude Dev:** Access comprehensive NixOS help through the Claude Dev extension
- **Get contextual suggestions:** Receive NixOS-specific completions and explanations while editing configuration files

For detailed VS Code setup instructions, see [VS Code Integration Guide](MCP_VSCODE_INTEGRATION.md).

### Neovim Integration

nixai provides comprehensive Neovim integration with lua configuration, custom commands, and keybindings for seamless NixOS assistance.

#### Automatic Setup

Use the built-in command to automatically configure Neovim integration:

```sh
# Basic setup with default socket path
nixai neovim-setup

# With custom socket path
nixai neovim-setup --socket-path=$HOME/.local/share/nixai/mcp.sock

# With custom Neovim config directory  
nixai neovim-setup --config-dir=$HOME/.config/nvim
```

This command:

1. Creates a `nixai.lua` module in your Neovim configuration
2. Provides instructions for adding it to your `init.lua` or `init.vim`
3. Sets up keymaps for NixOS documentation lookup and option explanations

#### Manual Setup

Add to your `init.lua`:

```lua
-- nixai integration
local ok, nixai = pcall(require, "nixai")
if ok then
  nixai.setup({
    socket_path = "/tmp/nixai-mcp.sock",
    keybindings = true, -- Enable default keybindings
  })
else
  vim.notify("nixai module not found", vim.log.levels.WARN)
end
```

#### Available Commands

- `:NixaiExplainOption <option>` - Explain NixOS options
- `:NixaiExplainHomeOption <option>` - Explain Home Manager options  
- `:NixaiSearch <query>` - Search packages and services
- `:NixaiDiagnose` - Diagnose current buffer or selection
- `:NixaiAsk <question>` - Ask direct questions

#### Default Keybindings

- `<leader>ne` - Explain option under cursor
- `<leader>ns` - Search packages/services
- `<leader>nd` - Diagnose current buffer
- `<leader>na` - Ask nixai a question

For comprehensive Neovim setup instructions, see [Neovim Integration Guide](neovim-integration.md).

### Home Manager Integration

Both editors can be automatically configured through Home Manager:

```nix
# home.nix
{ config, pkgs, ... }:
{
  imports = [
    # Import the nixai Home Manager module
    (builtins.fetchTarball "https://github.com/olafkfreund/nix-ai-help/archive/main.tar.gz")/modules/home-manager.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      socketPath = "$HOME/.local/share/nixai/mcp.sock";
    };
    # Automatically configure VS Code
    vscodeIntegration = true;
    # Automatically configure Neovim
    neovimIntegration = true;
  };
}
```

    prompt_title = 'NixOS Query',
    finder = require('telescope.finders').new_dynamic({
      fn = function(prompt)
        if #prompt > 0 then
          local result = require('nixai').query_docs(prompt)
          if result and result.content and result.content[1] then
            return {{value = result.content[1].text, display = prompt}}
          end
        end
        return {}
      end,
      entry_maker = function(entry)
        return {
          value = entry,
          display = entry.display,
          ordinal = entry.display,
        }
      end,
    }),
    sorter = require('telescope.config').values.generic_sorter({}),
    attach_mappings = function(prompt_bufnr)
      require('telescope.actions').select_default:replace(function()
        require('telescope.actions').close(prompt_bufnr)
        local selection = require('telescope.actions.state').get_selected_entry()
        require('nixai').show_in_float({
          content = {{text = selection.value.value}}
        }, "NixOS: " .. selection.value.display)
      end)
      return true
    end,
  }):find()
end

vim.keymap.set('n', '<leader>nt', nixai_picker, {desc = 'Telescope NixOS Query'})

```

#### Benefits of Neovim Integration

- Seamless workflow for NixOS users who prefer working in Neovim
- Context-aware suggestions based on your current file and cursor position
- Quick access to NixOS and Home Manager documentation and options
- Floating windows with properly formatted markdown display
- Works with your existing Neovim configuration

#### Requirements

- Running nixai MCP server (`nixai mcp-server start --background`)
- socat installed (`nix-env -iA nixos.socat` or add to your system packages)

For more details and advanced usage, see the [Neovim Integration](neovim-integration.md) documentation.

---

## 📝 Neovim + Home Manager Integration

For a bulletproof, copy-pasteable Neovim setup with Home Manager, see the [Neovim Integration Guide](neovim-integration.md). It covers:
- Minimal working config for `home-manager.nix`
- LSP and plugin setup
- Troubleshooting and health checks
- Best practices for reproducible Neovim environments

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

### Enhanced Build Troubleshooter

The Enhanced Build Troubleshooter is a comprehensive tool for analyzing build failures, optimizing build performance, and resolving common Nix build issues. It provides AI-powered analysis and actionable recommendations through a set of specialized subcommands.

#### Basic Build with AI Assistance

```sh
# Build a package with AI assistance for any failures
nixai build .#mypackage

# Build the current flake with AI assistance
nixai build
```

When using the basic `build` command, nixai will:
1. Run the standard `nix build` command
2. Capture any build failures
3. Provide an AI-generated summary of the problem
4. Suggest fixes based on the error patterns detected

#### Deep Build Analysis

```sh
nixai build debug firefox
```

The `debug` subcommand performs comprehensive analysis of build failures:

- 🔍 **Error Pattern Recognition**: Identifies common error types (dependency issues, compiler errors, missing inputs)
- 📊 **Detailed Analysis**: Provides step-by-step explanation of the error chain
- 🛠️ **Actionable Recommendations**: Suggests specific fixes for each identified issue
- 📚 **Documentation Links**: References relevant NixOS/Nixpkgs documentation
- 📋 **Comprehensive Report**: Generates a detailed failure analysis report

**Example Output:**

```
🔍 Deep Build Analysis: firefox

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 Build Environment
├─ Nixpkgs Version: 23.11
├─ System: x86_64-linux
├─ Cores Available: 8
└─ Memory Available: 16.0 GB

📊 Error Analysis
├─ Error Type: Missing Dependency
├─ Phase: Configure
├─ Component: firefox-112.0.2
└─ Root Cause: Required dependency 'libwebp' not found

🛠️ Recommended Fixes
1. Add missing dependency to buildInputs:
   buildInputs = old.buildInputs ++ [ pkgs.libwebp ];

2. Verify package is available in your nixpkgs version:
   nix-env -qA nixpkgs.libwebp

3. Apply this patch to your firefox derivation:
   [Detailed patch instructions]

💡 Additional Context
The error occurred because the build system expects libwebp for image processing,
but it wasn't included in the build environment. Firefox recently made this a
required dependency rather than optional.
```

#### Intelligent Retry with Automatic Fixes

```sh
nixai build retry
```

The `retry` subcommand:

- Analyzes the last build failure
- Automatically identifies common issues that can be fixed
- Applies recommended fixes and retries the build
- Provides detailed progress updates during the retry process
- Shows a comparison of before/after states

This command is particularly useful for common failure patterns like missing dependencies, permission issues, or simple configuration problems that have standard solutions.

#### Cache Miss Analysis

```sh
nixai build cache-miss
```

The `cache-miss` subcommand analyzes why builds aren't using the binary cache:

- 📊 **Cache Statistics**: Hit/miss rates and patterns
- 🔍 **Miss Reasons**: Identifies why specific builds aren't found in the cache
- 🌐 **Cache Configuration**: Analyzes substituter settings and connectivity
- 🔑 **Key Verification**: Checks for trusted keys and signing issues
- 📈 **Optimization Recommendations**: Suggests settings to improve cache utilization

**Example Output:**

```
📊 Cache Analysis Results

Cache Performance
├─ Hit Rate: 75%
├─ Miss Rate: 25%
├─ Cache Size: 2.5GB
├─ Recent Hits: 42
└─ Recent Misses: 14

Miss Reasons
├─ 8 misses due to missing trusted keys
├─ 4 misses due to custom package overrides
├─ 2 misses due to network connectivity issues

🛠️ Recommended Optimizations
1. Add missing trusted keys:
   nix-env --option trusted-public-keys 'cache.nixos.org-1:...'

2. Configure additional binary caches:
   nix.settings.substituters = [
     "https://cache.nixos.org"
     "https://nixpkgs-wayland.cachix.org"
   ];

3. Verify network connectivity to cache.nixos.org
```

#### Sandbox Debugging

```sh
nixai build sandbox-debug
```

The `sandbox-debug` subcommand helps resolve sandbox-related build issues:

- 🔒 **Sandbox Configuration**: Analyzes current sandbox settings
- 🔍 **Permission Analysis**: Identifies permission and access issues
- 🌐 **Network Access**: Diagnoses network-related sandbox problems
- 📁 **Path Access**: Identifies missing or inaccessible paths
- 🛠️ **Fix Recommendations**: Suggests sandbox configuration changes

This command is particularly useful for builds that fail with permission errors, network access issues, or path-related problems.

#### Build Performance Profiling

```sh
nixai build profile --package vim
```

The `profile` subcommand analyzes build performance and identifies optimization opportunities:

- ⏱️ **Time Analysis**: Breaks down where build time is spent
- 🧮 **Resource Usage**: CPU, memory, and I/O utilization 
- 🔍 **Bottleneck Detection**: Identifies performance bottlenecks
- 📊 **Comparison**: Benchmarks against typical build times
- 🚀 **Optimization Suggestions**: Recommendations to improve build speed

**Example Output:**

```
⚡ Build Performance Profile: vim

Build Time Breakdown
├─ Total Time: 4m 32s
├─ CPU Usage: 85%
├─ Memory Peak: 2.1GB
├─ Network Time: 45s  
├─ Compilation Time: 3m 20s
└─ Download Time: 27s

🔍 Bottlenecks Identified
1. Single-threaded compilation phase (3m 20s)
2. Limited parallelization in test phase
3. High memory usage during linking

🚀 Optimization Recommendations
1. Increase parallelization:
   nix.settings.max-jobs = 8;
   
2. Allocate more memory to builds:
   nix.settings.cores = 0;  # Use all cores
   
3. Consider using ccache:
   nix.settings.extra-sandbox-paths = [ "/var/cache/ccache" ];
```

#### Integration with Other nixai Features

The Enhanced Build Troubleshooter integrates seamlessly with other nixai features:

- **Documentation Integration**: Links to relevant NixOS docs via the MCP server
- **AI-Powered Analysis**: Uses the configured AI provider for intelligent analysis
- **System Health Context**: Incorporates system health data for better recommendations
- **Configuration Awareness**: Respects your NixOS config path settings
- **Terminal Formatting**: Beautiful, colorized terminal output with progress indicators

---

### Dependency & Import Graph Analyzer

The Dependency & Import Graph Analyzer helps you understand, visualize, and optimize the relationships between packages and configuration files in your NixOS system. This powerful tool provides AI-powered insights into your dependency tree and suggests optimizations to improve your system's performance and maintainability.

#### Analyzing Dependency Trees

```sh
nixai deps analyze
```

The `analyze` subcommand provides a comprehensive view of your system's package dependencies:

- 🔍 **Full System Analysis**: Maps all package relationships in your current configuration
- 📊 **Hierarchy Visualization**: Shows parent-child relationships between packages
- 🔎 **Circular Dependency Detection**: Identifies potential circular dependencies
- 📝 **AI-Powered Summary**: Provides an overview of your dependency structure with insights
- 🚩 **Issue Flagging**: Highlights potential problems like outdated packages or uncommon version constraints

**Example Output:**

```
📊 Dependency Analysis

System Overview
├─ Total Packages: 1,248
├─ Direct Dependencies: 142
├─ Indirect Dependencies: 1,106
├─ Deepest Chain: 15 levels
└─ Potential Issues: 3 found

Key Dependencies
├─ gcc [10.3.0] - Used by 428 packages
├─ glibc [2.35] - Used by 1,052 packages
├─ python3 [3.10.9] - Used by 89 packages
└─ openssl [3.0.8] - Used by 124 packages

🚩 Issues Detected
1. Circular dependency: python3 → pip → setuptools → python3
2. Multiple python versions: python 3.9 and 3.10
3. Outdated dependency: openssl 3.0.8 (3.0.9 available)

🤖 AI Analysis
Your system has a moderate-sized dependency tree with some outdated packages
and a circular dependency that may cause build issues. Consider updating
openssl and standardizing on a single Python version.
```

#### Understanding Package Inclusion

```sh
nixai deps why firefox
```

The `why` subcommand explains why a specific package is installed on your system:

- 🔍 **Origin Tracing**: Identifies the source of package inclusion
- 📋 **Full Path**: Shows the complete dependency chain leading to the package
- 🔍 **Alternative Paths**: Identifies multiple inclusion paths if they exist
- 🔄 **Version Resolution**: Explains version selection logic
- 🗑️ **Removal Impact**: Analysis of what would happen if the package were removed

**Example Output:**

```
❓ Why is firefox installed?

📋 Primary Inclusion Path:
configuration.nix
└─ environment.systemPackages
   └─ firefox [114.0.2]

📋 Alternative Paths:
home-manager
└─ home.packages
   └─ firefox [114.0.2]

💪 Direct Dependency: Yes
   This package was explicitly requested in your configuration.

🔄 Version Selection:
   Version 114.0.2 was selected from nixpkgs (override in /etc/nixos/overlays/firefox.nix)
   Default version would have been 113.0.1

🗑️ Removal Impact:
   Removing firefox would not break any other packages.
   2 user configurations reference this package.
```

#### Finding and Resolving Conflicts

```sh
nixai deps conflicts
```

The `conflicts` subcommand detects and helps resolve package conflicts:

- 🔍 **Conflict Detection**: Identifies conflicting package versions or flags
- 📋 **Comprehensive Report**: Details all conflicts with their sources
- 🛠️ **Resolution Suggestions**: Provides specific fix recommendations for each conflict
- 📈 **Priority Analysis**: Determines which conflicts are most critical to resolve
- 📊 **Before/After Comparison**: Shows the impact of proposed resolutions

**Example Output:**

```
🚫 Dependency Conflicts

Found 3 package conflicts in your configuration:

1. 🔴 Critical: openssl version conflict
   ├─ Path 1: nixpkgs.openssl [3.0.8] via environment.systemPackages
   ├─ Path 2: nixpkgs.openssl [1.1.1t] via letsencrypt
   └─ Resolution: Add the following to your configuration.nix:
      nixpkgs.config.packageOverrides = pkgs: {
        letsencrypt = pkgs.letsencrypt.override {
          openssl = pkgs.openssl;
        };
      };

2. 🟠 Important: python package conflict
   ├─ Path 1: python39
   ├─ Path 2: python310
   └─ Resolution: Standardize on one Python version:
      environment.systemPackages = with pkgs; [
        (python310.withPackages (ps: with ps; [ 
          # your Python packages here
        ]))
      ];

3. 🟡 Minor: gtk theme conflict
   ├─ Path 1: gnome.adwaita-icon-theme
   ├─ Path 2: custom-icon-theme
   └─ Resolution: Set GTK_THEME environment variable:
      environment.variables.GTK_THEME = "Adwaita";
```

#### Optimizing Dependencies

```sh
nixai deps optimize
```

The `optimize` subcommand analyzes your dependency structure and suggests optimizations:

- 🔍 **Inefficiency Detection**: Identifies redundant or unnecessary dependencies
- 📊 **Size Impact Analysis**: Shows the impact of each dependency on system size
- 🚀 **Performance Suggestions**: Recommends changes to improve build/runtime performance
- 💾 **Disk Usage Optimization**: Identifies opportunities to reduce system size
- 📝 **Configuration Recommendations**: Suggests specific configuration changes

**Example Output:**

```
⚡ Dependency Optimization

System Analysis
├─ Current Closure Size: 8.2 GB
├─ Redundant Packages: 14 found
├─ Unnecessary Dev Deps: 8 found
└─ Optimization Potential: ~1.1 GB (~13%)

🔍 Optimization Opportunities

1. 💾 Remove unnecessary development dependencies (~650 MB)
   ├─ Current: python310Full [includes dev tools, docs, tests]
   ├─ Suggested: python310 [minimal runtime only]
   └─ Configuration Change:
      - environment.systemPackages = with pkgs; [ python310Full ];
      + environment.systemPackages = with pkgs; [ python310 ];

2. 🚀 Consolidate duplicate libraries (~250 MB)
   ├─ Issue: Multiple versions of openssl, glib, and gtk
   └─ Resolution: Add overlay to standardize versions
     
3. 🧹 Clean up unused dependencies (~200 MB)
   ├─ kde-frameworks [only kdeconnect is used]
   └─ texlive-full [only basic LaTeX commands used]

📈 Expected Impact
├─ Storage Saved: ~1.1 GB
├─ Build Time Reduction: ~15%
└─ Boot Time Improvement: ~8%
```

#### Generating Dependency Graphs

```sh
nixai deps graph
```

The `graph` subcommand generates visual representations of your dependency structure:

- 📊 **Visualization**: Creates DOT or SVG graph of package relationships
- 🔍 **Interactive Exploration**: Optional output for interactive graph viewers
- 🎯 **Focused Views**: Generate graphs for specific packages or subsystems
- 🎨 **Customizable Display**: Options for detail level and graph layout
- 📁 **Import Maps**: Visualizes relationships between your configuration files

**Example Output:**

The command generates a dependency graph visualization and outputs:

```
📊 Dependency Graph Generated

Generated Files:
├─ nixos-deps.dot - DOT format graph (for processing)
└─ nixos-deps.svg - SVG visualization (for viewing)

Graph Statistics:
├─ Nodes: 248 packages
├─ Edges: 1,047 relationships
└─ Clusters: 12 major dependency groups

To view the interactive graph:
xdg-open nixos-deps.svg

To generate a focused graph for a specific package:
nixai deps graph --focus firefox
```

#### Integration with Other nixai Features

The Dependency & Import Graph Analyzer integrates with other nixai features:

- **Build Troubleshooter**: Provides dependency context for build failure analysis
- **Package Repository Analysis**: Leverages dependency information for better Nix derivations
- **System Health**: Incorporates dependency data in health reports
- **Configuration Management**: Shows the impact of configuration changes on dependencies

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

  - **Current Model**: gemini-2.5-flash-preview-05-20 (updated from deprecated gemini-pro)
  - **Tested**: ✅ Working with updated API endpoints and model

#### Example config for OpenAI or Gemini

```yaml
ai_provider: openai   # or 'gemini' or 'ollama'
ai_model: gpt-4       # or 'llama3', 'gemini-2.5-flash-preview-05-20', etc.
# ...other config options...
```

You can also override the provider and model at runtime:

```sh
nixai diagnose --provider openai --model gpt-4 --log-file /var/log/nixos/nixos-rebuild.log
nixai explain-option --provider gemini --model gemini-2.5-flash-preview-05-20 networking.firewall.enable
```

**Note:**

- If using OpenAI or Gemini, the API key must be set in your environment or in the config file under `openai_api_key` or `gemini_api_key` (environment variable is preferred for security).
- If no provider is set, Ollama is used by default for privacy.

### Example config with API keys (not recommended, prefer env vars)

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
- **Gemini API Updates**: Updated from deprecated `gemini-pro` to `gemini-2.5-flash-preview-05-20` model
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

## 🆕 Development Environment (devenv) Feature

nixai now supports rapid creation of reproducible development environments for Python, Rust, Node.js, and Go using the `devenv` command. This feature leverages [devenv.sh](https://devenv.sh/) and Nix to provide language-specific templates, framework/tool options, and service/database integration.

### Practical Usage

#### List Available Templates

    nixai devenv list

#### Create a New Project

    nixai devenv create python myproject --framework fastapi --with-poetry --services postgres,redis
    nixai devenv create golang my-go-app --framework gin --with-grpc
    nixai devenv create nodejs my-node-app --with-typescript --services mongodb
    nixai devenv create rust my-rust-app --with-wasm

#### Get AI-Powered Suggestions

    nixai devenv suggest "web app with database and REST API"

### How to Add a New Language or Framework

1. Edit `internal/devenv/builtin_templates.go` and implement the `Template` interface (see existing templates for examples).
2. Register your template in `registerBuiltinTemplates()` in `service.go`.
3. Add or update tests in `service_test.go`.
4. Document your new template in the main README and this manual.

### Example: Minimal Template Implementation

```go
// ExampleTemplate implements the Template interface
 type ExampleTemplate struct{}

 func (e *ExampleTemplate) Name() string { return "example" }
 func (e *ExampleTemplate) Description() string { return "Example language environment" }
 func (e *ExampleTemplate) RequiredInputs() []devenv.InputField { return nil }
 func (e *ExampleTemplate) SupportedServices() []string { return nil }
 func (e *ExampleTemplate) Validate(config devenv.TemplateConfig) error { return nil }
 func (e *ExampleTemplate) Generate(config devenv.TemplateConfig) (*devenv.DevenvConfig, error) {
     // ... generate config ...
     return &devenv.DevenvConfig{/* ... */}, nil
 }
```

### Testing

- Run all tests: `go test ./internal/devenv/...`
- Try creating projects with various options and check the generated `devenv.nix`

---

## 🔄 Migration Assistant

nixai provides a robust migration assistant to help you safely convert your NixOS configuration between legacy channels and modern flakes. The migration assistant includes:

**Features:**

- Migration Analysis: Detects your current setup and analyzes migration complexity
- Step-by-Step Guide: AI-generated migration steps with safety checks
- Backup & Rollback: Automatic backup and rollback procedures
- Validation: Comprehensive validation of migration success
- Best Practices: Integration of flake best practices and optimizations

### Usage

**Analyze your current setup:**

```sh
nixai migrate analyze --nixos-path /etc/nixos
```

**Convert from channels to flakes:**

```sh
nixai migrate to-flakes --nixos-path /etc/nixos
```

- The assistant will walk you through the migration, create a backup, and validate the result.
- All output is formatted with glamour for easy reading.
- If anything goes wrong, you can roll back to your previous configuration.

### Best Practices & Safety

- Always review the migration analysis before proceeding
- Backups are created automatically and can be restored if needed
- All changes are validated before finalizing the migration

### Example Workflow

1. Analyze:

   ```sh
   nixai migrate analyze --nixos-path /etc/nixos
   ```

2. Migrate:

   ```sh
   nixai migrate to-flakes --nixos-path /etc/nixos
   ```

3. Rollback (if needed):
   - Follow the instructions provided by nixai to restore from backup

### Planned Features

- `nixai migrate from-flakes` (convert back to channels)
- `nixai migrate channel-upgrade` (safe channel upgrades)
- `nixai migrate flake-inputs` (update/explain flake inputs)

### Troubleshooting

- If migration fails, check the backup directory for your previous configuration
- Review AI explanations for manual steps or caveats
- For complex setups, consult the official NixOS documentation or ask direct questions with `nixai --ask`

---
