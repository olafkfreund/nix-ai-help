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
- [Configuration](#configuration)
- [Tips & Troubleshooting](#tips--troubleshooting)
- [Development Environment (devenv) Feature](#development-environment-devenv-feature)
- [Neovim + Home Manager Integration](#neovim--home-manager-integration)

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
