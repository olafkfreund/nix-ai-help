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
- [🖥️ Multi-Machine Configuration Manager](#multi-machine-configuration-manager)
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
- [Flake Creation & Correction (`nixai flake create`)](#flake-creation--correction-nixai-flake-create)
- [🩺 Doctor Command: System Diagnostics & Troubleshooting](#doctor-command-system-diagnostics--troubleshooting)

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

The simplest and most direct way to use nixai is by asking questions about NixOS directly from the command line. This feature makes nixai incredibly accessible for both beginners and experts.

#### Basic Usage

```sh
# Ask questions directly by providing them as arguments
./nixai "how do I enable SSH in NixOS?"
./nixai "what is a Nix flake?"
./nixai "how to configure services.postgresql in NixOS?"

# Alternative: use the --ask or -a flag
./nixai --ask "how do I update packages in NixOS?"
./nixai -a "what are NixOS generations?"
```

#### Real-World Scenarios

**Scenario 1: New NixOS User Setting Up SSH**
```sh
./nixai "I just installed NixOS and need to enable SSH access for remote work. How do I do this securely?"
```

*Expected output includes:*
- Complete configuration snippet for `/etc/nixos/configuration.nix`
- Security best practices (disable root login, key-only auth)
- How to apply changes with `nixos-rebuild switch`
- Firewall configuration recommendations

**Scenario 2: Developer Setting Up Development Environment**
```sh
./nixai "I'm a Python developer who needs Docker, PostgreSQL, and VS Code on NixOS. What's the best way to set this up?"
```

*Expected output includes:*
- System packages vs. user packages recommendations
- Home Manager vs. system configuration
- Development environment best practices
- Service configuration examples

**Scenario 3: Troubleshooting a Failed Service**
```sh
./nixai "My nginx service keeps failing to start after I added SSL configuration. How do I debug this?"
```

*Expected output includes:*
- Debugging commands (`systemctl status`, `journalctl`)
- Common SSL configuration mistakes
- Certificate path requirements in NixOS
- Testing and validation steps

**Scenario 4: Understanding Advanced NixOS Concepts**
```sh
./nixai "I keep hearing about overlays and want to customize Firefox. Can you explain overlays and show me how?"
```

*Expected output includes:*
- Conceptual explanation of overlays
- Complete working example for Firefox customization
- Best practices for overlay organization
- Alternative approaches comparison

#### How It Works

Both methods are equivalent and provide the same functionality:

1. **Question Processing**: The question is sent to your configured AI provider (Ollama, Gemini, or OpenAI)
2. **Documentation Context**: If the MCP server is running, it queries relevant NixOS documentation to provide context
3. **AI Analysis**: The AI generates a comprehensive response with practical examples and best practices
4. **Formatted Output**: The response is formatted with proper Markdown rendering in your terminal

#### Pro Tips for Better Results

**🎯 Be Specific and Context-Rich**
```sh
# Good: Specific with context
./nixai "I'm running NixOS 23.11 on a laptop with NVIDIA graphics. How do I enable hardware acceleration for video playback?"

# Less helpful: Too vague
./nixai "graphics not working"
```

**🔧 Include Your Current Setup**
```sh
# Good: Shows current state
./nixai "I have services.xserver.enable = true but want to switch to Wayland with GNOME. What changes do I need?"

# Less helpful: Assumes context
./nixai "how to enable Wayland"
```

**📊 Ask for Comparisons**
```sh
./nixai "What are the pros and cons of using nix-shell vs nix develop vs devenv for Python development?"
```

**🛠 Request Complete Workflows**
```sh
./nixai "Show me the complete process to set up a Rust development environment with cross-compilation support"
```

#### Advanced Question Patterns

**Configuration Validation**
```sh
./nixai "Here's my current nginx config: [paste config]. Is this secure and following NixOS best practices?"
```

**Migration Assistance**
```sh
./nixai "I'm migrating from Ubuntu where I used docker-compose. How should I handle containerized services in NixOS?"
```

**Performance Optimization**
```sh
./nixai "My NixOS system feels slow during builds. How can I optimize nix builds and system performance?"
```

**Security Hardening**
```sh
./nixai "I need to harden my NixOS server for production use. What security configurations should I implement?"
```

#### Expected Output Features

When using the direct question functionality, nixai will:

- **📊 Progress Indicators**: Show a progress indicator while retrieving documentation and generating responses
- **🎨 Rich Formatting**: Format output as readable, colorized Markdown in your terminal
- **💻 Syntax Highlighting**: Include proper code syntax highlighting for configuration snippets
- **🔗 Documentation Links**: Provide links to official documentation when relevant
- **⚡ Quick Actions**: Suggest follow-up commands or next steps
- **🧪 Test Commands**: Include verification commands to test configurations

---

## Diagnosing NixOS Issues

nixai can analyze logs, config snippets, or `nix log` output to help you diagnose problems. This is one of the most powerful features for troubleshooting NixOS configurations and build failures.

### Basic Examples

#### Diagnose a Log File

```sh
nixai diagnose --log-file /var/log/nixos/nixos-rebuild.log
```

#### Diagnose a Nix Log

```sh
nixai diagnose --nix-log /nix/store/xxxx.drv
```

#### Diagnose a Config Snippet

```sh
echo 'services.nginx.enable = true;' | nixai diagnose
```

### Real-World Troubleshooting Scenarios

#### Scenario 1: Failed NixOS Rebuild

**Problem**: Your `nixos-rebuild switch` command failed with a cryptic error.

```sh
# First, capture the full log
sudo nixos-rebuild switch 2>&1 | tee rebuild.log

# If it fails, diagnose the log
nixai diagnose --log-file rebuild.log
```

**Expected AI Analysis**:
- Identifies the specific point of failure
- Explains common causes (syntax errors, missing dependencies, conflicting options)
- Provides step-by-step resolution instructions
- Suggests verification commands

#### Scenario 2: Service Configuration Issues

**Problem**: PostgreSQL service won't start after configuration changes.

```sh
# Get service logs and diagnose
journalctl -u postgresql.service --no-pager | nixai diagnose
```

**Example Configuration Issue**:
```nix
# This might be in your configuration.nix
services.postgresql = {
  enable = true;
  dataDir = "/var/lib/postgresql/data";  # ❌ Wrong path for NixOS
  port = 5432;
};
```

**AI Diagnosis Output**:
- Identifies that `dataDir` is incorrectly set (NixOS manages this automatically)
- Explains NixOS PostgreSQL service options
- Provides corrected configuration
- Shows how to check service status and logs

#### Scenario 3: Build Failure for Custom Package

**Problem**: Your custom Nix derivation fails to build.

```sh
# Build and diagnose if it fails
nix build .#mypackage 2>&1 | nixai diagnose
```

**Common Issues Identified**:
- Missing build dependencies (`nativeBuildInputs` vs `buildInputs`)
- Incorrect source hash (`outputHashMismatch`)
- Build environment issues (missing environment variables)
- Installation phase problems

#### Scenario 4: Home Manager Integration Issues

**Problem**: Home Manager fails to activate with NixOS configuration.

```sh
# Capture Home Manager logs
home-manager switch 2>&1 | tee hm-switch.log
nixai diagnose --log-file hm-switch.log
```

**Common Diagnosis Results**:
- Conflicting package versions between system and user
- Missing Home Manager modules
- Incorrect user configuration syntax
- Permission issues with dotfiles

#### Scenario 5: Flake Input Resolution Problems

**Problem**: Flake inputs won't update or resolve correctly.

```sh
# Diagnose flake lock issues
nix flake update 2>&1 | nixai diagnose
```

**Expected Analysis**:
- Network connectivity issues
- Repository access problems
- Version conflicts between inputs
- Suggestions for manual input overrides

### Advanced Diagnostic Features

#### Interactive Diagnosis Mode

```sh
# Start interactive mode for complex issues
nixai diagnose --interactive
```

This mode allows you to:
- Paste multiple log segments
- Provide context about recent changes
- Get follow-up questions from the AI
- Receive step-by-step debugging guidance

#### Continuous Log Monitoring

```sh
# Monitor and diagnose logs in real-time
journalctl -f | nixai diagnose --stream
```

**Use cases**:
- Monitoring service startup issues
- Debugging real-time application problems
- Watching for configuration reload issues

#### Configuration Validation

```sh
# Validate configuration before applying
nixai diagnose --config-only /etc/nixos/configuration.nix
```

**What it checks**:
- Syntax validation
- Option compatibility
- Best practice recommendations
- Security considerations

### Pro Tips for Better Diagnosis

#### Provide Context

```sh
# Include system information for better analysis
uname -a > system-info.txt
nixos-version >> system-info.txt
nix --version >> system-info.txt

# Combine with error logs
cat system-info.txt error.log | nixai diagnose
```

#### Use Structured Output

```sh
# Get structured diagnosis for automation
nixai diagnose --format json error.log
```

#### Focus on Specific Components

```sh
# Diagnose only networking issues
nixai diagnose --filter networking error.log

# Focus on systemd services
nixai diagnose --filter systemd error.log
```

### Common Issue Patterns

#### Build Environment Problems

**Symptoms**: Packages build locally but fail in CI/CD
**Common Diagnosis**: Missing `nativeBuildInputs`, impure build environment
**AI Solution**: Provides clean build environment setup

#### Version Conflicts

**Symptoms**: Different package versions causing conflicts
**Common Diagnosis**: Multiple nixpkgs inputs, version pinning issues
**AI Solution**: Shows how to unify package versions across inputs

#### Permission and Security Issues

**Symptoms**: Services can't access files, socket connection failures
**Common Diagnosis**: SELinux/AppArmor conflicts, incorrect file permissions
**AI Solution**: Provides security context fixes and permission corrections

### Integration with Other nixai Features

#### Combine with Option Explanation

```sh
# Diagnose and then explain problematic options
nixai diagnose error.log
nixai explain-option services.nginx  # Based on diagnosis results
```

#### Use with Package Search

```sh
# Find alternative packages when builds fail
nixai diagnose build-error.log
nixai search pkg alternative-package-name  # Based on diagnosis suggestions
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

### Real-World Package Analysis Examples

#### Example 1: Local Go Project with Complex Dependencies

**Scenario**: You have a Go microservice with gRPC, database connections, and custom build requirements.

```sh
# Analyze the current directory
nixai package-repo . --local
```

**Project Structure**:
```text
my-go-service/
├── go.mod (requires gRPC, protobuf, database drivers)
├── go.sum
├── proto/ (protobuf definitions)
├── migrations/ (database migrations)
├── scripts/generate.sh (code generation)
└── Dockerfile (for reference)
```

**Expected AI Analysis and Output**:
```nix
# Generated: my-go-service.nix
{ lib, buildGoModule, fetchFromGitHub, protobuf, protoc-gen-go, protoc-gen-go-grpc }:

buildGoModule rec {
  pname = "my-go-service";
  version = "1.0.0";

  src = ./.;

  vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

  nativeBuildInputs = [ protobuf protoc-gen-go protoc-gen-go-grpc ];

  preBuild = ''
    # Generate protobuf code
    make generate
  '';

  ldflags = [ "-s" "-w" "-X main.version=${version}" ];

  meta = with lib; {
    description = "A Go microservice with gRPC support";
    license = licenses.mit;
    maintainers = [ maintainers.yourname ];
  };
}
```

**AI Explanation Includes**:
- Why `nativeBuildInputs` was chosen for build tools
- How `preBuild` handles code generation
- Explanation of `vendorHash` and how to compute it
- Recommendations for packaging the protobuf files separately if shared

#### Example 2: Python Project with Multiple Dependencies

**Scenario**: A Flask web application with scientific computing dependencies.

```sh
nixai package-repo https://github.com/myorg/data-analysis-api
```

**Project Analysis**:
- **Detected**: `pyproject.toml`, `requirements.txt`, `setup.py`
- **Dependencies**: Flask, NumPy, Pandas, SciPy, custom C extensions
- **Special requirements**: CUDA support optional, testing with pytest

**Generated Derivation**:
```nix
{ lib, python3Packages, fetchFromGitHub }:

python3Packages.buildPythonApplication rec {
  pname = "data-analysis-api";
  version = "2.1.0";

  src = fetchFromGitHub {
    owner = "myorg";
    repo = "data-analysis-api";
    rev = "v${version}";
    sha256 = "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=";
  };

  propagatedBuildInputs = with python3Packages; [
    flask
    numpy
    pandas
    scipy
    requests
    pytest
  ];

  checkInputs = with python3Packages; [
    pytestCheckHook
  ];

  # Disable tests that require GPU
  disabledTests = [
    "test_cuda_acceleration"
  ];

  meta = with lib; {
    description = "Data analysis API with scientific computing";
    homepage = "https://github.com/myorg/data-analysis-api";
    license = licenses.apache2;
  };
}
```

**AI Insights**:
- Explains why `buildPythonApplication` vs `buildPythonPackage`
- Shows how to handle optional CUDA dependencies
- Provides guidance on test exclusion patterns
- Suggests overlay patterns for scientific Python packages

#### Example 3: Node.js Project with Monorepo Structure

**Scenario**: A Next.js application in a monorepo with shared packages.

```sh
nixai package-repo https://github.com/company/frontend-monorepo --output ./nix-packages
```

**Monorepo Structure**:
```text
frontend-monorepo/
├── package.json (workspaces configuration)
├── apps/
│   ├── web/ (Next.js app)
│   └── admin/ (React admin panel)
└── packages/
    ├── ui-components/
    └── shared-utils/
```

**Generated Output**:
```sh
Generated multiple derivations:
- ./nix-packages/web-app.nix
- ./nix-packages/admin-app.nix
- ./nix-packages/ui-components.nix
- ./nix-packages/shared-utils.nix
- ./nix-packages/default.nix (combines all)
```

**Example default.nix**:
```nix
{ pkgs ? import <nixpkgs> {} }:

let
  sharedUtils = pkgs.callPackage ./shared-utils.nix {};
  uiComponents = pkgs.callPackage ./ui-components.nix { inherit sharedUtils; };
in {
  webApp = pkgs.callPackage ./web-app.nix { inherit sharedUtils uiComponents; };
  adminApp = pkgs.callPackage ./admin-app.nix { inherit sharedUtils uiComponents; };
  
  # Development shell with all dependencies
  devShell = pkgs.mkShell {
    buildInputs = [ sharedUtils uiComponents ];
    shellHook = ''
      echo "Frontend monorepo development environment"
    '';
  };
}
```

#### Example 4: Rust Project with Custom Build Requirements

**Scenario**: A Rust CLI tool that needs specific build flags and system dependencies.

```sh
nixai package-repo . --analyze-only
```

**Analysis Output**:
```text
🦀 Rust Project Analysis
├─ Package: my-rust-cli v0.3.0
├─ Edition: 2021
├─ Dependencies: 
│  ├─ Runtime: clap, serde, tokio, reqwest
│  ├─ Build: cc (C bindings)
│  └─ System: openssl, pkg-config
├─ Features: default, tls, async
└─ Special Requirements: 
   ├─ Links to system OpenSSL
   ├─ Custom build script (build.rs)
   └─ Requires pkg-config

🤖 AI Recommendations:
1. Use buildRustPackage with explicit system dependencies
2. Set OPENSSL_DIR and PKG_CONFIG_PATH for C bindings
3. Consider vendoring dependencies for reproducibility
4. Add feature flags to control optional dependencies

📦 Suggested Nix Expression:
```

**Follow-up Generation**:
```sh
nixai package-repo . --name my-rust-cli --output ./packaging
```

#### Example 5: Complex Multi-Language Project

**Scenario**: A project with Rust backend, TypeScript frontend, and Python ML pipeline.

```sh
nixai package-repo https://github.com/ai-startup/full-stack-ml
```

**AI Analysis**:
```text
🔍 Multi-Language Project Detected:
├─ Backend (Rust): src/backend/Cargo.toml
├─ Frontend (TypeScript): src/frontend/package.json
├─ ML Pipeline (Python): ml/pyproject.toml
└─ Documentation (mdBook): docs/book.toml

💡 Packaging Strategy:
1. Create separate derivations for each component
2. Use a master flake.nix to orchestrate builds
3. Provide unified development environment
4. Handle cross-component dependencies
```

**Generated flake.nix**:
```nix
{
  description = "Full-stack ML application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    rust-overlay.url = "github:oxalica/rust-overlay";
  };

  outputs = { self, nixpkgs, rust-overlay }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ rust-overlay.overlays.default ];
      };

      backend = pkgs.callPackage ./nix/backend.nix {};
      frontend = pkgs.callPackage ./nix/frontend.nix {};
      mlPipeline = pkgs.callPackage ./nix/ml-pipeline.nix {};
      docs = pkgs.callPackage ./nix/docs.nix {};

    in {
      packages.${system} = {
        inherit backend frontend mlPipeline docs;
        default = pkgs.symlinkJoin {
          name = "full-stack-ml";
          paths = [ backend frontend mlPipeline docs ];
        };
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          # Rust toolchain
          rust-bin.stable.latest.default
          # Node.js for frontend
          nodejs_20
          # Python for ML
          python3
          python3Packages.pip
          # Build tools
          pkg-config
          openssl
        ];
      };
    };
}
```

### Troubleshooting Package Analysis

#### Common Issues and Solutions

**Issue**: "Unable to determine build system"
```sh
# Solution: Provide hints
nixai package-repo . --build-system makefile --language c
```

**Issue**: "Missing system dependencies"
```sh
# Solution: Analyze dependencies first
nixai package-repo . --analyze-only --verbose
# Then add missing deps to the generated derivation
```

**Issue**: "Private repository access denied"
```sh
# Solution: Use SSH key or token
nixai package-repo git@github.com:private/repo.git --ssh-key ~/.ssh/deploy_key
```

#### Advanced Usage Patterns

**Continuous Integration Integration**:
```sh
# Generate Nix expressions in CI
nixai package-repo . --output-format json > package-info.json
# Use the JSON to generate CI matrix builds
```

**Template Generation**:
```sh
# Create reusable templates
nixai package-repo template-repo --template-mode
# Apply templates to similar projects
nixai package-repo new-project --template go-microservice
```

**Dependency Analysis**:
```sh
# Analyze just dependencies without generating derivation
nixai package-repo . --deps-only
# Output: dependency tree with Nix package mappings
```

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

The `nixai health` command provides comprehensive health checks for your NixOS system, with AI-powered analysis and recommendations. This is essential for maintaining a healthy NixOS installation and catching issues before they become critical.

### What does `nixai health` check?

- ✅ **NixOS configuration validity** - Syntax and semantic validation
- 🔧 **Failed system services** - Detection and analysis of service failures  
- 💾 **Disk space usage** - Storage analysis with cleanup suggestions
- 📡 **Nix channel status** - Channel consistency and update recommendations
- 🚀 **Boot integrity** - Bootloader and generation management
- 🌐 **Network connectivity** - Internet access and Nix binary cache connectivity
- 🏪 **Nix store health** - Store corruption detection and repair suggestions

### Real-World Health Check Scenarios

#### Scenario 1: Daily System Maintenance

```sh
# Run comprehensive health check
nixai health
```

**Example Output with Issues Found**:

```text
🏥 NixOS System Health Check
============================

✅ Configuration Validation
├─ Syntax: Valid
├─ Module imports: All resolved
└─ Option conflicts: None detected

⚠️  System Services (2 issues found)
├─ 🔴 postgresql.service - Failed to start
│  └─ Last failure: 2 hours ago
├─ 🟡 nginx.service - Degraded (high memory usage)
│  └─ Memory usage: 450MB (expected: <100MB)
└─ ✅ 23 other services running normally

💾 Disk Space Analysis
├─ Root (/): 45GB used / 100GB total (45% - OK)
├─ Nix Store (/nix): 85GB used / 100GB total (85% - WARNING)
│  └─ 🧹 Garbage collection could free ~12GB
└─ Home (/home): 15GB used / 50GB total (30% - OK)

📡 Network & Connectivity
├─ ✅ Internet connectivity: OK
├─ ✅ cache.nixos.org: Reachable (15ms)
├─ ⚠️  cache.flox.dev: Timeout (may affect builds)
└─ ✅ GitHub.com: Reachable (25ms)

🚀 Boot Configuration
├─ ✅ Current generation: 127 (booted)
├─ ⚠️  Old generations: 15 (cleanup recommended)
│  └─ 🧹 Can free ~2.1GB by cleaning generations >7 days old
└─ ✅ Bootloader: systemd-boot, working correctly

🏪 Nix Store Health
├─ ✅ Store consistency: OK
├─ ✅ Database integrity: OK
└─ ⚠️  Unreferenced paths: 847 (3.2GB) - cleanup available
```

**AI-Powered Recommendations**:

```text
🤖 AI Analysis & Recommendations:

🔧 Critical Issues (Fix Soon):
1. PostgreSQL service failure - likely caused by:
   • Data directory permissions issue
   • Port 5432 already in use
   • Invalid configuration syntax
   
   Suggested actions:
   ┌─ Immediate diagnosis:
   │  sudo systemctl status postgresql.service
   │  sudo journalctl -u postgresql.service --since "2 hours ago"
   │
   └─ Common fixes:
      • Check services.postgresql.port configuration
      • Verify data directory permissions: /var/lib/postgresql/
      • Review recent configuration changes

2. Nginx memory usage is unusually high:
   • Expected: <100MB, Actual: 450MB
   • Possible causes: memory leak, excessive worker processes
   
   Suggested investigation:
   ┌─ Check configuration:
   │  nixai explain-option services.nginx.workerProcesses
   │  sudo nginx -t  # Test configuration
   │
   └─ Monitor resource usage:
      top -p $(pgrep nginx)

⚠️  Maintenance Tasks (Schedule Soon):
1. Disk space cleanup (could free ~17GB total):
   sudo nix-collect-garbage --delete-older-than 7d
   sudo nixos-rebuild switch --delete-older-than 7d

2. Review and clean old generations:
   nixos-rebuild list-generations
   # Keep only last 5 generations:
   sudo nix-env --delete-generations +5 --profile /nix/var/nix/profiles/system

💡 Optimization Opportunities:
1. Add binary cache substituters to speed up builds:
   nix.settings.substituters = [
     "https://cache.nixos.org/"
     "https://nix-community.cachix.org"
   ];

2. Consider enabling automatic garbage collection:
   nix.gc = {
     automatic = true;
     dates = "weekly";
     options = "--delete-older-than 30d";
   };
```

#### Scenario 2: Pre-Production Deployment Check

```sh
# Detailed health check for production readiness
nixai health --production-mode --nixos-path /etc/nixos
```

**Enhanced Checks for Production**:
- 🔒 **Security Configuration** - Firewall, SSH hardening, user privileges
- 🚨 **Service Monitoring** - All critical services operational
- 📈 **Performance Metrics** - System resource utilization
- 🔄 **Backup Verification** - Configuration and data backup status
- 🌐 **Network Security** - Open ports, certificate expiration

#### Scenario 3: Troubleshooting System Issues

```sh
# Focused health check with debug information
nixai health --log-level debug --format json > health-report.json
```

**Use Cases**:
- Creating detailed reports for support requests
- Automated monitoring and alerting integration
- Historical health tracking and trend analysis
- Pre-change validation in CI/CD pipelines

#### Scenario 4: Post-Update Validation

```sh
# Validate system after major updates
nixai health --post-update --compare-with-generation 125
```

**Validation Areas**:
- Service status compared to previous generation
- Configuration drift detection
- Performance regression analysis
- New issues introduced by updates

### Advanced Health Check Features

#### Custom Health Check Profiles

```sh
# Use predefined profiles for different scenarios
nixai health --profile server        # Focus on server-specific checks
nixai health --profile desktop       # Desktop environment validation
nixai health --profile developer     # Development tools and environment
```

#### Integration with Monitoring Systems

```sh
# Export health data for external monitoring
nixai health --prometheus-metrics > /var/lib/prometheus/nixos-health.prom
nixai health --nagios-format > /tmp/nixos-check-results
```

#### Automated Remediation

```sh
# Run health check with automatic safe fixes
nixai health --auto-fix --dry-run     # Preview what would be fixed
nixai health --auto-fix               # Apply safe automatic fixes
```

**Safe automatic fixes include**:
- Garbage collection for disk space
- Restarting failed non-critical services
- Updating nix channels if stale
- Clearing temporary files and caches

### Continuous Health Monitoring

#### Scheduled Health Checks

Add to your NixOS configuration for regular monitoring:

```nix
# /etc/nixos/configuration.nix
services.nixai.healthChecks = {
  enable = true;
  schedule = "daily";
  notifications = {
    email = "admin@example.com";
    critical-only = false;
  };
  autoFix = {
    enable = true;
    diskCleanup = true;
    serviceRestart = true;
  };
};
```

#### Integration with System Monitoring

```sh
# Create systemd service for regular checks
sudo systemctl enable nixai-health.timer
sudo systemctl start nixai-health.timer

# View health check history
journalctl -u nixai-health.service
```

### Health Check Exit Codes

For scripting and automation:

- **0**: All checks passed, system healthy
- **1**: Minor issues found, attention recommended  
- **2**: Major issues found, immediate action required
- **3**: Critical issues found, system stability at risk

**Example automation script**:
```bash
#!/bin/bash
nixai health --quiet
case $? in
  0) echo "✅ System healthy" ;;
  1) echo "⚠️  Minor issues detected, scheduling maintenance" ;;
  2) echo "🔧 Major issues found, notifying administrators" ;;
  3) echo "🚨 Critical issues detected, immediate intervention required" ;;
esac
```
   
   Recommendations:
   1. Disable SSH root login: services.openssh.permitRootLogin = "no";
   2. Review open ports and close unnecessary ones
   3. Enable automatic security updates
```

---

## 🖥️ Multi-Machine Configuration Manager

The Multi-Machine Configuration Manager lets you centrally manage, synchronize, and deploy NixOS configurations across multiple machines—all from the command line. This powerful feature is essential for teams managing development environments, server fleets, or hybrid cloud infrastructures where configuration consistency and automated deployment are critical.

### What It Does

- **Machine Registry**: Register and manage multiple NixOS machines in a central registry with SSH connectivity
- **Fleet Operations**: Group machines for bulk operations (e.g., deploy to all web servers, update all development machines)
- **Configuration Sync**: Synchronize configurations between local and remote machines with conflict detection
- **Deployment Management**: Deploy configuration changes with rollback support and health validation
- **Drift Detection**: Compare configurations across machines to identify configuration drift
- **Status Monitoring**: Check connectivity, health, and deployment status of all registered machines
- **Backup & Recovery**: Automatic backup of configurations before deployment with easy rollback
- **Template Deployment**: Deploy standardized configurations to new machines

### Real-World Multi-Machine Scenarios

#### Scenario 1: Managing Development Team Infrastructure

**Situation**: A software development team needs to maintain consistent development environments across multiple developer workstations and shared infrastructure.

```sh
# Set up the machine registry for the development team
nixai machines init --registry-path ~/team-nixos-configs

# Register development workstations
nixai machines add dev-alice 192.168.1.100 --description "Alice's development workstation" \
  --ssh-user alice --ssh-key ~/.ssh/team_devs \
  --nixos-path /home/alice/nixos-config

nixai machines add dev-bob 192.168.1.101 --description "Bob's development workstation" \
  --ssh-user bob --ssh-key ~/.ssh/team_devs \
  --nixos-path /home/bob/nixos-config

nixai machines add dev-carol 192.168.1.102 --description "Carol's development workstation" \
  --ssh-user carol --ssh-key ~/.ssh/team_devs \
  --nixos-path /home/carol/nixos-config

# Register shared development infrastructure
nixai machines add build-server 192.168.1.50 --description "CI/CD build server" \
  --ssh-user nixos --ssh-key ~/.ssh/infrastructure \
  --nixos-path /etc/nixos

nixai machines add dev-database 192.168.1.51 --description "Development database server" \
  --ssh-user nixos --ssh-key ~/.ssh/infrastructure \
  --nixos-path /etc/nixos

# Create logical groups for easier management
nixai machines groups add developers dev-alice dev-bob dev-carol
nixai machines groups add infrastructure build-server dev-database
nixai machines groups add all-dev developers infrastructure
```

**Managing the development environment**:

```sh
# Check status of all development machines
nixai machines status
```

**Example Output**:
```text
🖥️  Machine Status Report
=======================

👥 Group: developers (3 machines)
├─ 🟢 dev-alice (192.168.1.100)
│  ├─ Status: Online, responsive (15ms)
│  ├─ Last sync: 2 hours ago
│  ├─ Config generation: 47 (current)
│  └─ Health: ✅ All services running
├─ 🟢 dev-bob (192.168.1.101)
│  ├─ Status: Online, responsive (12ms)
│  ├─ Last sync: 3 hours ago  
│  ├─ Config generation: 46 (1 behind)
│  └─ Health: ⚠️  1 service degraded (docker)
└─ 🟢 dev-carol (192.168.1.102)
│  ├─ Status: Online, responsive (18ms)
│  ├─ Last sync: 1 hour ago
│  ├─ Config generation: 47 (current)
│  └─ Health: ✅ All services running

🏗️  Group: infrastructure (2 machines)
├─ 🟢 build-server (192.168.1.50)
│  ├─ Status: Online, high load (CPU: 85%)
│  ├─ Last sync: 6 hours ago
│  ├─ Config generation: 23 (current)
│  └─ Health: ✅ CI/CD pipelines active
└─ 🟢 dev-database (192.168.1.51)
│  ├─ Status: Online, responsive (8ms)
│  ├─ Last sync: 4 hours ago
│  ├─ Config generation: 15 (current)
│  └─ Health: ✅ PostgreSQL active

📊 Summary:
├─ Total machines: 5
├─ Online: 5 (100%)
├─ Up to date: 4 (80%)
├─ Needs attention: 1 (dev-bob: service issue)
└─ Overall health: Good
```

**Deploy updated development tools to all developers**:

```sh
# Check what would change before deploying
nixai machines diff --group developers --local-config ./configs/dev-environment.nix

# Deploy the new configuration to all developer machines
nixai machines deploy --group developers --config ./configs/dev-environment.nix \
  --confirm-each --backup --validate-health
```

**Deployment Output**:
```text
🚀 Group Deployment: developers
===============================

📋 Deployment Plan:
├─ Target machines: 3 (dev-alice, dev-bob, dev-carol)
├─ Configuration: ./configs/dev-environment.nix
├─ Backup: Enabled (rollback available)
├─ Health validation: Enabled
└─ Confirmation: Required for each machine

🤖 AI Configuration Analysis:
📦 Changes detected:
├─ Added packages: nodejs_20, python311, docker-compose
├─ Service updates: docker service configuration
├─ Environment changes: Added development environment variables
└─ Estimated deployment time: ~5 minutes per machine

⚠️  Potential Impact:
├─ Docker service restart required (may interrupt running containers)
├─ New nodejs version may require npm package rebuilds
└─ Recommendation: Deploy during low-activity periods

🎯 Deploying to dev-alice (192.168.1.100)...
├─ 💾 Creating backup... Done (generation 47 backed up)
├─ 📤 Uploading configuration... Done (2.3MB transferred)
├─ 🔨 Building configuration... Done (3m 24s)
├─ 🔄 Switching to new configuration... Done
├─ 🏥 Health validation... ✅ All services healthy
└─ ✅ Deployment successful (generation 48)

Continue with dev-bob? [y/N/s(kip)/a(ll)] y

🎯 Deploying to dev-bob (192.168.1.101)...
├─ ⚠️  Pre-deployment issue detected: Docker service degraded
├─ 🛠️  Auto-fixing docker service... Done
├─ 💾 Creating backup... Done (generation 46 backed up)
├─ 📤 Uploading configuration... Done (2.3MB transferred)
├─ 🔨 Building configuration... Done (3m 18s)
├─ 🔄 Switching to new configuration... Done
├─ 🏥 Health validation... ✅ All services healthy (docker fixed)
└─ ✅ Deployment successful (generation 47)

🎯 Deploying to dev-carol (192.168.1.102)...
├─ 💾 Creating backup... Done (generation 47 backed up)
├─ 📤 Uploading configuration... Done (2.3MB transferred)
├─ 🔨 Building configuration... Done (3m 31s)
├─ 🔄 Switching to new configuration... Done
├─ 🏥 Health validation... ✅ All services healthy
└─ ✅ Deployment successful (generation 48)

📊 Deployment Summary:
├─ ✅ Successful: 3/3 machines
├─ ⏱️  Total time: 11m 45s
├─ 🔄 Rollback available on all machines
└─ 📝 Deployment log: /tmp/nixai-deploy-20241201-1430.log
```

#### Scenario 2: Production Server Fleet Management

**Situation**: Managing a production environment with multiple web servers, database servers, and load balancers that need consistent configuration and coordinated updates.

```sh
# Register production infrastructure
nixai machines add lb1 10.0.1.10 --description "Primary load balancer" \
  --environment production --role load-balancer \
  --ssh-user nixos --ssh-key ~/.ssh/production

nixai machines add lb2 10.0.1.11 --description "Secondary load balancer" \
  --environment production --role load-balancer \
  --ssh-user nixos --ssh-key ~/.ssh/production

nixai machines add web1 10.0.2.10 --description "Web server 1" \
  --environment production --role web-server \
  --ssh-user nixos --ssh-key ~/.ssh/production

nixai machines add web2 10.0.2.11 --description "Web server 2" \
  --environment production --role web-server \
  --ssh-user nixos --ssh-key ~/.ssh/production

nixai machines add web3 10.0.2.12 --description "Web server 3" \
  --environment production --role web-server \
  --ssh-user nixos --ssh-key ~/.ssh/production

nixai machines add db1 10.0.3.10 --description "Primary database server" \
  --environment production --role database \
  --ssh-user nixos --ssh-key ~/.ssh/production

nixai machines add db2 10.0.3.11 --description "Secondary database server" \
  --environment production --role database \
  --ssh-user nixos --ssh-key ~/.ssh/production

# Create production groups
nixai machines groups add load-balancers lb1 lb2
nixai machines groups add web-servers web1 web2 web3
nixai machines groups add databases db1 db2
nixai machines groups add production-all load-balancers web-servers databases
```

**Rolling update with health checks and rollback capability**:

```sh
# Perform rolling update of web servers with zero downtime
nixai machines deploy --group web-servers --rolling-update \
  --max-unavailable 1 --health-check-url "http://localhost:8080/health" \
  --rollback-on-failure --config ./configs/web-server-v2.nix
```

**Rolling Update Output**:
```text
🔄 Rolling Update: web-servers
==============================

📋 Rolling Update Configuration:
├─ Target group: web-servers (3 machines)
├─ Max unavailable: 1 machine at a time
├─ Health check: http://localhost:8080/health
├─ Rollback on failure: Enabled
├─ Configuration: ./configs/web-server-v2.nix
└─ Load balancer integration: Automatic

🤖 AI Pre-Deployment Analysis:
📦 Configuration changes detected:
├─ Application update: v1.2.4 → v1.3.0
├─ Nginx configuration: Added new security headers
├─ Monitoring: Updated Prometheus metrics endpoint
├─ Dependencies: Updated OpenSSL security patch
└─ Estimated impact: Low risk, security improvements

⚡ Rolling Update Process:

🎯 Phase 1: Updating web1 (10.0.2.10)
├─ 🔄 Removing from load balancer rotation... Done
├─ ⏱️  Waiting for active connections to drain (30s)... Done
├─ 💾 Creating backup snapshot... Done (generation 156)
├─ 📤 Deploying new configuration... Done (4m 12s)
├─ 🔄 Starting services... Done
├─ 🏥 Health check (attempt 1/5): ✅ Healthy
├─ 🔄 Adding back to load balancer rotation... Done
├─ ⏱️  Monitoring for 2 minutes... ✅ Stable
└─ ✅ web1 update successful

🎯 Phase 2: Updating web2 (10.0.2.11)
├─ 🔄 Removing from load balancer rotation... Done
├─ ⏱️  Waiting for active connections to drain (30s)... Done  
├─ 💾 Creating backup snapshot... Done (generation 156)
├─ 📤 Deploying new configuration... Done (4m 8s)
├─ 🔄 Starting services... Done
├─ 🏥 Health check (attempt 1/5): ✅ Healthy
├─ 🔄 Adding back to load balancer rotation... Done
├─ ⏱️  Monitoring for 2 minutes... ✅ Stable
└─ ✅ web2 update successful

🎯 Phase 3: Updating web3 (10.0.2.12)
├─ 🔄 Removing from load balancer rotation... Done
├─ ⏱️  Waiting for active connections to drain (30s)... Done
├─ 💾 Creating backup snapshot... Done (generation 156)
├─ 📤 Deploying new configuration... Done (4m 15s)
├─ 🔄 Starting services... Done
├─ 🏥 Health check (attempt 1/5): ✅ Healthy
├─ 🔄 Adding back to load balancer rotation... Done
├─ ⏱️  Monitoring for 2 minutes... ✅ Stable
└─ ✅ web3 update successful

🎉 Rolling Update Complete!
├─ ✅ All machines updated successfully
├─ ⏱️  Total update time: 14m 32s
├─ 📊 Zero downtime achieved
├─ 🏥 All health checks passed
└─ 🔄 Rollback snapshots available for 7 days
```

#### Scenario 3: Configuration Drift Detection and Remediation

**Situation**: Over time, production machines may accumulate configuration drift due to manual changes or failed deployments. Regular drift detection helps maintain consistency.

```sh
# Comprehensive drift detection across all machines
nixai machines drift-check --all --detailed --fix-recommendations
```

**Drift Detection Output**:
```text
🔍 Configuration Drift Analysis
===============================

📊 Drift Summary:
├─ Machines analyzed: 7
├─ Clean configurations: 4
├─ Minor drift detected: 2
├─ Major drift detected: 1
└─ Critical issues: 0

🟡 Minor Drift: web2 (10.0.2.11)
├─ Issue: SSH configuration difference
├─ Expected: PermitRootLogin = "no"
├─ Actual: PermitRootLogin = "yes"
├─ Source: Manual modification 3 days ago
├─ Risk Level: Medium (security concern)
└─ 🛠️  Fix: Deploy standard SSH configuration

🟠 Major Drift: db1 (10.0.3.10)
├─ Issue: PostgreSQL configuration divergence
├─ Expected: max_connections = 200
├─ Actual: max_connections = 400
├─ Additional: Custom shared_preload_libraries setting
├─ Source: Manual performance tuning 1 week ago
├─ Risk Level: High (performance/stability impact)
└─ 🛠️  Recommendations:
   ┌─ Option 1: Revert to standard configuration
   ├─ Option 2: Update standard config to include custom settings
   └─ Option 3: Create exception policy for this machine

🟢 Clean Configurations:
├─ ✅ lb1, lb2: Load balancer configs match expected
├─ ✅ web1, web3: Web server configs match expected
├─ ✅ db2: Database config matches expected
└─ ✅ All machines: System packages and services aligned

🤖 AI Recommendations:

High Priority Actions:
1. Fix SSH security issue on web2:
   nixai machines deploy web2 --config ./configs/web-server-secure.nix

2. Address PostgreSQL drift on db1:
   - Review performance requirements
   - Update standard config or create exception
   - Document any approved deviations

Preventive Measures:
1. Enable automatic drift monitoring:
   nixai machines monitor --schedule daily --alert-on-drift

2. Implement configuration immutability:
   - Use read-only filesystem for /etc/nixos
   - Restrict manual configuration changes
   - Set up change approval workflow

3. Regular compliance audits:
   nixai machines audit --compliance-rules ./security-baseline.yaml
```

**Automated drift remediation**:

```sh
# Fix minor drift automatically
nixai machines drift-fix --severity minor --auto-approve --group web-servers

# Create exception for approved manual changes
nixai machines exception add db1 "postgresql.max_connections" \
  --reason "Performance optimization approved by DBA team" \
  --expiry "2024-12-31" \
  --approved-by "john.doe@company.com"
```

#### Scenario 4: Disaster Recovery and Backup Management

**Situation**: Production environment needs robust backup and disaster recovery capabilities with automated failover testing.

```sh
# Set up automated backup strategy
nixai machines backup-strategy configure \
  --schedule "0 2 * * *" \
  --retention-days 30 \
  --backup-location "s3://company-nixos-backups" \
  --encrypt-with-key ~/.ssh/backup-encryption.pub

# Create disaster recovery plan
nixai machines disaster-recovery plan create production-dr \
  --primary-group production-all \
  --backup-region us-west-2 \
  --recovery-time-objective 15m \
  --recovery-point-objective 4h
```

**Test disaster recovery procedures**:

```sh
# Simulate disaster and test recovery
nixai machines disaster-recovery test production-dr --dry-run
```

**Disaster Recovery Test Output**:
```text
🚨 Disaster Recovery Test: production-dr
========================================

📋 Test Scenario: Complete data center failure
├─ Affected machines: 7 (all production)
├─ Recovery target region: us-west-2
├─ RTO target: 15 minutes
├─ RPO target: 4 hours
└─ Test mode: Dry run (no actual changes)

🔄 Recovery Procedure Simulation:

Phase 1: Assessment and Preparation (Target: 2 minutes)
├─ ✅ Backup availability check: Latest backup 2 hours old
├─ ✅ Recovery infrastructure check: Standby machines available
├─ ✅ Network connectivity test: All recovery nodes reachable
└─ ✅ Encryption keys validation: Backup decryption keys accessible

Phase 2: Infrastructure Recovery (Target: 8 minutes)
├─ 🔄 Provisioning recovery machines...
│  ├─ lb1-recovery (10.1.1.10) - Ready in 45s
│  ├─ lb2-recovery (10.1.1.11) - Ready in 52s
│  ├─ web1-recovery (10.1.2.10) - Ready in 1m 8s
│  ├─ web2-recovery (10.1.2.11) - Ready in 1m 12s
│  ├─ web3-recovery (10.1.2.12) - Ready in 1m 5s
│  ├─ db1-recovery (10.1.3.10) - Ready in 2m 15s
│  └─ db2-recovery (10.1.3.11) - Ready in 2m 22s
└─ ✅ All recovery machines provisioned (2m 45s)

Phase 3: Configuration Restoration (Target: 4 minutes)
├─ 🔄 Restoring NixOS configurations...
│  ├─ Load balancers: 1m 25s (nginx configs, SSL certificates)
│  ├─ Web servers: 2m 8s (application code, dependencies)
│  └─ Database servers: 3m 42s (PostgreSQL, data restoration)
└─ ✅ All configurations restored (3m 42s)

Phase 4: Data Recovery (Target: 1 minute)
├─ 🔄 Database restoration from backup...
│  ├─ Primary DB: Restored from backup (2 hours old) - 45s
│  └─ Replica sync: Synchronized with primary - 18s
└─ ✅ Data recovery complete (1m 3s)

Phase 5: Service Validation (Target: 2 minutes)
├─ 🏥 Health checks...
│  ├─ Load balancers: ✅ All endpoints responding
│  ├─ Web servers: ✅ Application health checks passed
│  ├─ Databases: ✅ Connections and queries working
│  └─ End-to-end test: ✅ Complete user workflow validated
└─ ✅ All services validated (1m 34s)

📊 Recovery Test Results:
├─ Total recovery time: 9m 44s (✅ Under 15m RTO)
├─ Data loss: 2 hours (✅ Within 4h RPO)
├─ Service availability: 100% after recovery
└─ Test status: ✅ PASSED

🚨 Issues Identified:
1. Database recovery took longer than expected (3m 42s vs 2m target)
   - Recommendation: Pre-warm database recovery snapshots
   - Impact: Low (still within overall RTO)

2. SSL certificate validation delay (15s)
   - Recommendation: Pre-stage certificates in recovery region
   - Impact: Very low

🛠️  Recommended Improvements:
1. Optimize database backup/restore process
2. Implement warm standby for critical databases
3. Pre-position SSL certificates and secrets
4. Automate DNS failover procedures
5. Add monitoring for backup freshness

Next Steps:
1. Schedule regular DR tests (monthly)
2. Implement recommended improvements
3. Update DR playbooks based on test results
4. Train operations team on recovery procedures
```

### Advanced Multi-Machine Features

#### Configuration Templates and Standardization

```sh
# Create machine templates for different roles
nixai machines template create web-server-template \
  --base-config ./templates/web-server-base.nix \
  --variables "app_version,ssl_cert_path,log_level" \
  --validation-rules "./validation/web-server-rules.yaml"

# Deploy machines from template
nixai machines deploy-from-template web-server-template \
  --target web4 --variables "app_version=1.3.0,ssl_cert_path=/etc/ssl/certs/app.crt,log_level=info"
```

#### Automated Compliance and Security Scanning

```sh
# Run security compliance scan across all machines
nixai machines security-scan --baseline CIS-NixOS-v1.0 \
  --output-format report --save-to compliance-report.pdf

# Continuous compliance monitoring
nixai machines compliance monitor --schedule hourly \
  --rules ./security-policies/ \
  --alert-webhook "https://company.slack.com/webhooks/security"
```

#### Performance Monitoring and Optimization

```sh
# Monitor performance across machine groups
nixai machines performance monitor --group production-all \
  --metrics "cpu,memory,disk,network" \
  --alert-thresholds "cpu>80%,memory>90%,disk>85%" \
  --duration 24h

# Optimize configurations based on usage patterns
nixai machines optimize --group web-servers \
  --analyze-period 30d \
  --apply-recommendations --confirm-each
```

### Integration with Other nixai Features

**Health Checks Integration**:
```sh
# Run health checks across all machines
nixai machines health-check --all --detailed

# Schedule regular health monitoring
nixai machines health-monitor --schedule "*/15 * * * *" \
```

**Documentation and Option Explanation**:
```sh
# Explain configuration options across machine groups
nixai machines explain-config --group web-servers \
  --option "services.nginx.virtualHosts"

# Generate documentation for machine configurations
nixai machines document --group production-all \
  --output-format markdown --save-to docs/production-configs.md
```

**Direct Questions for Multi-Machine Management**:
```sh
nixai "How do I safely update the nginx configuration across all web servers?"
nixai "What's the best practice for managing SSH keys across multiple machines?"
nixai "How can I detect if any production machines have been manually modified?"
```

### Command Reference and Best Practices

#### Essential Commands

- `nixai machines init` — Initialize machine registry
- `nixai machines add <name> <host>` — Register a new machine
- `nixai machines groups add <group> <machines...>` — Create machine groups
- `nixai machines status` — Check all machine status
- `nixai machines deploy --group <group>` — Deploy to machine group
- `nixai machines diff` — Compare configurations across machines
- `nixai machines drift-check` — Detect configuration drift
- `nixai machines backup` — Manage machine backups
- `nixai machines disaster-recovery` — Disaster recovery operations

#### Best Practices

1. **Start Small**: Begin with a few machines and gradually expand
2. **Use Groups**: Organize machines into logical groups (environment, role, team)
3. **Regular Monitoring**: Set up automated drift detection and health monitoring
4. **Backup Strategy**: Implement comprehensive backup and disaster recovery plans
5. **Security First**: Use SSH keys, restrict access, enable audit logging
6. **Documentation**: Maintain clear documentation of machine purposes and configurations
7. **Testing**: Test all deployment procedures in staging before production
8. **Gradual Rollouts**: Use rolling updates for production deployments
9. **Monitoring**: Implement comprehensive monitoring and alerting
10. **Compliance**: Regular security and compliance scanning

#### Troubleshooting

**Common Issues and Solutions**:

- **SSH Connection Failures**: Verify SSH keys, network connectivity, and user permissions
- **Configuration Conflicts**: Use `nixai machines diff` to identify differences
- **Deployment Failures**: Check logs, validate configurations, ensure sufficient resources
- **Drift Detection Issues**: Review manual changes, update templates, create exceptions
- **Performance Problems**: Monitor resource usage, optimize configurations, scale horizontally

**Getting Help**:
```sh
# Get specific help for machine management
nixai "How do I troubleshoot SSH connection issues with my machines?"
nixai "Best practices for organizing machines into groups?"
nixai "How to recover from a failed deployment across multiple machines?"
```

---

## Configuration Templates & Snippets

nixai provides a powerful template and snippet management system to help you quickly set up and reuse NixOS configurations. This feature includes curated templates for common setups, personal snippet management for your custom configurations, and AI-powered template discovery from GitHub repositories.

### What Templates and Snippets Provide

**Templates** are complete, ready-to-use NixOS configurations for common scenarios:
- Pre-configured desktop environments (GNOME, KDE, XFCE)  
- Gaming setups with optimized drivers and performance
- Server configurations with security hardening
- Development environments for different programming languages
- Hardware-specific configurations (laptop, workstation, server)

**Snippets** are reusable configuration fragments:
- Service configurations (nginx, PostgreSQL, Docker)
- Hardware settings (graphics drivers, audio, networking)
- Package collections for specific workflows
- Security policies and firewall rules
- Performance optimizations and tweaks

### Real-World Template Usage Scenarios

#### Scenario 1: Setting Up a Gaming Workstation

**Situation**: A user wants to quickly configure a NixOS system optimized for gaming with Steam, NVIDIA drivers, and performance tweaks.

```sh
# Discover available gaming templates
nixai templates search gaming
```

**Example Output**:
```text
🎮 Gaming Templates Found:

📦 Built-in Templates:
├─ gaming-nvidia          - Gaming setup with NVIDIA drivers and Steam
├─ gaming-amd             - Gaming setup with AMD drivers and Steam  
├─ gaming-minimal         - Basic gaming setup without GPU-specific configs
└─ retro-gaming           - Retro gaming with emulators and compatibility layers

🌐 GitHub Templates (curated):
├─ nixos-gaming-rig       - Complete gaming configuration with RGB, VR support
├─ competitive-gaming     - Low-latency setup for competitive gaming
├─ streamer-setup         - Gaming + streaming configuration (OBS, etc.)
└─ lan-party-config       - Portable gaming configuration for events

💡 AI Recommendation:
For NVIDIA-based gaming, 'gaming-nvidia' template provides the best starting point
with optimized drivers, Steam, and performance tweaks already configured.
```

```sh
# View detailed template information
nixai templates show gaming-nvidia
```

**Template Details Output**:
```text
🎮 Template: gaming-nvidia
==========================

📋 Description:
Complete gaming configuration optimized for NVIDIA graphics cards, including
Steam, Lutris, performance optimizations, and essential gaming utilities.

🎯 Key Features:
├─ NVIDIA proprietary drivers with proper configuration
├─ Steam with Proton compatibility layers
├─ Lutris for GOG and Epic Games integration
├─ Hardware-accelerated video playback
├─ Gaming-optimized kernel parameters
├─ Audio configuration for gaming headsets
├─ RGB lighting support (OpenRGB)
└─ Performance monitoring tools

🛠️ Hardware Requirements:
├─ NVIDIA GPU (GTX 900 series or newer recommended)
├─ 16GB+ RAM for modern games
├─ SSD for faster game loading
└─ Decent audio setup (headset or speakers)

🔧 Configuration Preview:
```nix
{
  # Enable NVIDIA drivers
  services.xserver.videoDrivers = [ "nvidia" ];
  hardware.nvidia = {
    modesetting.enable = true;
    powerManagement.enable = true;
    open = false;  # Use proprietary driver
    nvidiaSettings = true;
  };

  # Gaming packages
  environment.systemPackages = with pkgs; [
    steam
    lutris
    discord
    obs-studio
    openrgb
    mangohud
    gamemode
  ];

  # Gaming optimizations
  programs.steam = {
    enable = true;
    remotePlay.openFirewall = true;
    dedicatedServer.openFirewall = true;
  };

  # Performance tweaks
  boot.kernel.sysctl = {
    "vm.max_map_count" = 262144;  # For some games
    "kernel.sched_autogroup_enabled" = 0;  # Gaming performance
  };
}
```

🚀 Quick Start:
1. Apply template: nixai templates apply gaming-nvidia
2. Rebuild system: sudo nixos-rebuild switch
3. Reboot to load NVIDIA drivers
4. Configure Steam and start gaming!

⚠️  Important Notes:
- Requires enabling unfree packages in nixpkgs config
- May need to configure specific GPU model settings
- Steam will need initial setup after first boot
- Some anti-cheat systems may not work on NixOS
```

```sh
# Apply the gaming template to current system
nixai templates apply gaming-nvidia --merge --backup
```

**Template Application Output**:
```text
🎮 Applying Template: gaming-nvidia
====================================

📋 Pre-Application Analysis:
├─ Current configuration: /etc/nixos/configuration.nix
├─ Backup location: /etc/nixos/backups/config-backup-20241201-1445.nix
├─ Merge mode: Enabled (preserving existing settings)
├─ Conflicts detected: 2 (will be resolved)
└─ Estimated apply time: 3-5 minutes

⚠️  Configuration Conflicts:
1. Video drivers: Current 'intel' vs template 'nvidia'
   Resolution: Will configure hybrid graphics (Intel + NVIDIA)
   
2. Audio: Current PulseAudio vs template PipeWire
   Resolution: Will upgrade to PipeWire (better gaming audio)

🤖 AI Recommendations:
├─ Enable unfree packages: Add nixpkgs.config.allowUnfree = true
├─ Hardware optimization: Consider enabling hardware.cpu.intel.updateMicrocode
├─ Gaming performance: Add boot.kernelParams for low-latency audio
└─ Security: Gaming template includes some security relaxations for compatibility

📦 Additional Packages Being Added:
├─ Essential: steam, lutris, discord, mangohud
├─ Utilities: gamemode, openrgb, obs-studio
├─ Drivers: nvidia, nvidia-settings, nvidia-prime
└─ Audio: pipewire, pipewire-pulse, wireplumber

Apply gaming template configuration? [y/N] y

🔄 Applying Configuration Changes...
├─ 💾 Creating backup... Done (/etc/nixos/backups/config-backup-20241201-1445.nix)
├─ 🔧 Merging template configuration... Done
├─ 🧹 Resolving conflicts... Done (2 conflicts resolved)
├─ 📝 Writing new configuration... Done (/etc/nixos/configuration.nix)
├─ ✅ Configuration syntax validation... Passed
└─ 🎯 Template applied successfully!

🚀 Next Steps:
1. Review merged configuration: /etc/nixos/configuration.nix
2. Rebuild system: sudo nixos-rebuild switch
3. Reboot to activate NVIDIA drivers
4. Configure Steam: Add library folders, enable Proton
5. Test gaming performance with included monitoring tools

💡 Pro Tips:
- Use 'mangohud steam' to monitor gaming performance
- Configure 'gamemode' for automatic performance optimization  
- Use 'openrgb' to control RGB lighting
- Enable Steam Remote Play for gaming on other devices

🔄 Rollback Available:
If needed, restore previous configuration:
sudo cp /etc/nixos/backups/config-backup-20241201-1445.nix /etc/nixos/configuration.nix
sudo nixos-rebuild switch
```

#### Scenario 2: Corporate Development Environment

**Situation**: A company needs to standardize development environments across their team with specific security policies, development tools, and corporate configurations.

```sh
# Search for corporate/enterprise templates
nixai templates search "corporate development enterprise"
```

```sh
# Create a custom corporate template from existing configuration
nixai templates save corporate-dev-standard ./corporate-base.nix \
  --category "Corporate" \
  --description "Standard corporate development environment with security policies" \
  --tags "corporate,development,security,compliance" \
  --variables "employee_name,department,security_level"
```

**Custom Template Creation Output**:
```text
📋 Creating Custom Template: corporate-dev-standard
===================================================

🔍 Analyzing Source Configuration:
├─ Source file: ./corporate-base.nix
├─ Configuration size: 324 lines
├─ Detected services: 12 (SSH, VPN, monitoring, etc.)
├─ Security policies: 8 (firewall, user restrictions, audit)
├─ Development tools: 15 packages
└─ Variables identified: 3 custom + 3 suggested

🎯 Template Variables Configuration:
├─ employee_name (string): Employee full name for certificates
├─ department (enum): [engineering, qa, design, management]
├─ security_level (enum): [standard, elevated, admin]
├─ 📋 Suggested additions:
│  ├─ project_access (list): Project repositories to grant access
│  ├─ vpn_profile (string): VPN configuration profile
│  └─ backup_retention (number): Backup retention days

🔒 Security Policies Detected:
├─ SSH key management with corporate CA
├─ Mandatory VPN for external access
├─ Disk encryption requirements
├─ Audit logging and monitoring
├─ Package installation restrictions
├─ Network access controls
├─ Time-based access policies
└─ Data loss prevention measures

💼 Corporate Integrations:
├─ Active Directory authentication
├─ Corporate certificate authority
├─ Centralized logging (Splunk)
├─ Endpoint protection (CrowdStrike)
├─ Software compliance scanning
└─ Hardware inventory management

✅ Template Saved Successfully!
├─ Template ID: corporate-dev-standard
├─ Location: ~/.config/nixai/templates/corporate-dev-standard.yaml
├─ Ready for deployment across organization
└─ Version control integration available

🚀 Usage Examples:
# Deploy to new employee workstation
nixai templates apply corporate-dev-standard \
  --variables "employee_name=John Doe,department=engineering,security_level=standard"

# Bulk deployment to team
nixai templates deploy-bulk corporate-dev-standard \
  --csv-file ./new-hires.csv \
  --target-machines ./workstations.yaml
```

```sh
# Deploy corporate template to new employee workstation
nixai templates apply corporate-dev-standard \
  --variables "employee_name=Jane Smith,department=engineering,security_level=elevated" \
  --target-machine dev-workstation-42 \
  --compliance-check
```

**Corporate Deployment Output**:
```text
🏢 Corporate Template Deployment
================================

👤 Employee Configuration:
├─ Name: Jane Smith
├─ Department: Engineering  
├─ Security Level: Elevated
├─ Workstation: dev-workstation-42
└─ Compliance Check: Enabled

🔍 Pre-Deployment Compliance Validation:
├─ ✅ Hardware meets corporate standards
├─ ✅ BIOS security settings verified
├─ ✅ Disk encryption capability confirmed
├─ ✅ Network access to corporate resources
├─ ⚠️  TPM 2.0 required for elevated security (detected TPM 1.2)
└─ 🔧 Auto-remediation available for TPM issue

🛡️ Security Policies Being Applied:
├─ Full disk encryption with corporate keys
├─ SSH access restricted to corporate network
├─ VPN mandatory for external access
├─ Endpoint monitoring and compliance
├─ Code signing requirements for development
├─ Data classification and handling policies
└─ Regular security policy updates

🔧 Development Environment:
├─ Programming languages: Go, Python, TypeScript, Rust
├─ IDEs: VS Code with corporate extensions
├─ Version control: Git with corporate hooks
├─ Container tools: Docker, Podman (corporate registry)
├─ Security tools: SAST, DAST, dependency scanning
├─ Monitoring: Performance and security monitoring
└─ Backup: Automated corporate backup integration

⚠️  Compliance Issues Found:
1. TPM Version Mismatch:
   - Required: TPM 2.0 for elevated security
   - Current: TPM 1.2
   - Impact: Cannot enable hardware-based encryption
   - Resolution: Hardware upgrade required OR security level downgrade

Apply corporate configuration with TPM limitation? [y/N/downgrade] downgrade

🔄 Adjusting Security Level: elevated → standard
├─ Removing TPM 2.0 dependent features
├─ Using software-based encryption instead
├─ Maintaining all other security policies
└─ Compliance team will be notified

📦 Installing Corporate Software Stack...
├─ 🔄 Base system packages... Done (2m 15s)
├─ 🔄 Development tools... Done (3m 42s)
├─ 🔄 Security agents... Done (1m 28s)
├─ 🔄 Monitoring tools... Done (45s)
├─ 🔄 Corporate certificates... Done (12s)
└─ 🔄 Policy enforcement... Done (8s)

🔐 Configuring Security Policies...
├─ Setting up disk encryption... Done
├─ Configuring corporate VPN... Done
├─ Installing endpoint protection... Done
├─ Setting up audit logging... Done
├─ Configuring access controls... Done
└─ Enabling compliance monitoring... Done

✅ Corporate Development Environment Deployed!

📋 Employee Onboarding Checklist:
├─ ✅ Workstation configured with corporate policies
├─ ✅ Development tools installed and configured
├─ ✅ Security monitoring active
├─ ✅ VPN access configured
├─ ✅ Corporate certificates installed
├─ 📧 Welcome email sent with next steps
└─ 📅 Security briefing scheduled

🔗 Important Links for Jane Smith:
├─ Employee handbook: https://intranet.company.com/handbook
├─ Development guidelines: https://dev.company.com/guidelines  
├─ Security policies: https://security.company.com/policies
├─ IT support: https://support.company.com
└─ Compliance training: https://training.company.com/security

🚨 Compliance Reporting:
├─ Workstation ID: dev-workstation-42
├─ Configuration applied: corporate-dev-standard-v2.1
├─ Security level: standard (elevated requested, downgraded due to hardware)
├─ Compliance officer notified: security@company.com
└─ Next compliance check: 30 days
```

#### Scenario 3: Home Lab Server Setup

**Situation**: A homelab enthusiast wants to set up multiple services (media server, home automation, network storage) with proper security and monitoring.

```sh
# Search for homelab and server templates
nixai templates search "homelab media server nas"
```

```sh
# Combine multiple templates for a comprehensive homelab setup
nixai templates combine \
  --base server-minimal \
  --add media-server \
  --add home-automation \
  --add network-storage \
  --output homelab-complete \
  --resolve-conflicts
```

**Template Combination Output**:
```text
🏠 Creating Homelab Template Combination
========================================

📦 Base Template: server-minimal
├─ SSH server with security hardening
├─ Basic firewall configuration
├─ System monitoring and logging
├─ Automatic security updates
└─ Essential server utilities

➕ Additional Templates:
├─ media-server: Plex, Jellyfin, media management
├─ home-automation: Home Assistant, MQTT, Zigbee
└─ network-storage: Samba, NFS, backup solutions

🔍 Conflict Resolution Analysis:
├─ Port conflicts: 3 found, resolved automatically
├─ Service conflicts: 1 found (nginx vs apache)
├─ Package conflicts: 2 found (python versions)
└─ Configuration conflicts: 4 found, merged intelligently

🛠️ Conflict Resolutions Applied:
1. Web Server Conflict (nginx vs apache):
   Resolution: Using nginx as reverse proxy for all services
   
2. Python Version Conflict (3.9 vs 3.11):
   Resolution: Standardized on Python 3.11 for all services
   
3. Port Allocation:
   ├─ Plex: 32400 (unchanged)
   ├─ Home Assistant: 8123 (unchanged)  
   ├─ Jellyfin: 8096 (changed from 8080 to avoid nginx conflict)
   └─ Samba: 445 (unchanged)

🔧 Enhanced Integrations Added:
├─ Single Sign-On: Authentik for unified authentication
├─ Reverse Proxy: nginx with automatic SSL certificates
├─ Monitoring: Prometheus + Grafana dashboards
├─ Backup: Automated backups for all service data
├─ Network: VLAN segmentation for IoT devices
└─ Security: Fail2ban, intrusion detection, log analysis

💡 AI Optimizations Suggested:
├─ Docker isolation for media services (security)
├─ ZFS storage pool for data integrity
├─ VPN access for remote management
├─ Automated certificate management (Let's Encrypt)
└─ Energy-efficient scheduling for maintenance tasks

📋 Combined Template Summary:
├─ Total services: 12 major services
├─ Package count: 89 packages
├─ Configuration size: 845 lines
├─ Estimated setup time: 25-40 minutes
└─ Memory requirement: 8GB+ recommended

✅ Homelab Template Created Successfully!
Template saved as: homelab-complete
Ready for deployment to homelab server.
```

```sh
# Deploy the comprehensive homelab template
nixai templates apply homelab-complete \
  --interactive \
  --hardware-optimize \
  --config-wizard
```

**Interactive Homelab Deployment**:
```text
🏠 Homelab Complete Setup Wizard
================================

🖥️ Hardware Configuration:
CPU: Intel i5-8400 (6 cores)
RAM: 16GB DDR4
Storage: 500GB NVMe SSD + 4TB HDD
Network: Gigabit Ethernet

📋 Service Configuration Wizard:

1️⃣ Media Server Setup:
   Media library location [/media/storage]: /mnt/nas/media
   Enable Plex? [Y/n]: y
   Enable Jellyfin as backup? [y/N]: y
   Enable automatic media organization? [Y/n]: y
   
2️⃣ Home Automation Setup:
   Home Assistant installation type [docker/native]: docker
   Enable Zigbee support? [Y/n]: y
   Zigbee coordinator device [/dev/ttyUSB0]: /dev/ttyACM0
   Enable MQTT broker? [Y/n]: y
   
3️⃣ Network Storage Setup:
   Samba share name [homelab]: nas
   Enable Time Machine support? [y/N]: n
   NFS exports for media? [Y/n]: y
   Enable automatic backups? [Y/n]: y
   
4️⃣ Security & Access:
   Enable VPN access? [Y/n]: y
   VPN type [wireguard/openvpn]: wireguard
   External domain [optional]: homelab.example.com
   Enable Let's Encrypt SSL? [Y/n]: y
   
5️⃣ Monitoring & Maintenance:
   Enable Prometheus monitoring? [Y/n]: y
   Enable Grafana dashboards? [Y/n]: y
   Log retention days [30]: 60
   Backup retention weeks [4]: 8

🔧 Generating Optimized Configuration...
├─ Hardware-specific optimizations applied
├─ Service dependencies resolved
├─ Network configuration optimized
├─ Storage layout planned
└─ Security policies configured

📦 Installation Plan:
├─ Base system setup: ~8 minutes
├─ Docker services deployment: ~12 minutes
├─ SSL certificate generation: ~3 minutes
├─ Initial data migration: ~10 minutes
├─ Service health verification: ~5 minutes
└─ Total estimated time: ~40 minutes

Continue with installation? [Y/n] y

🚀 Deploying Homelab Configuration...

Phase 1: Base System Configuration
├─ 🔄 Updating system packages... Done (2m 15s)
├─ 🔄 Configuring networking... Done (45s)
├─ 🔄 Setting up storage (ZFS pool)... Done (3m 20s)
├─ 🔄 Configuring firewall rules... Done (30s)
└─ ✅ Base system ready

Phase 2: Service Container Deployment  
├─ 🔄 Installing Docker and compose... Done (2m 5s)
├─ 🔄 Deploying Plex container... Done (3m 40s)
├─ 🔄 Deploying Jellyfin container... Done (2m 15s)
├─ 🔄 Deploying Home Assistant... Done (4m 30s)
├─ 🔄 Setting up reverse proxy... Done (1m 25s)
└─ ✅ All services deployed

Phase 3: SSL and External Access
├─ 🔄 Generating Let's Encrypt certificates... Done (2m 10s)
├─ 🔄 Configuring WireGuard VPN... Done (1m 40s)
├─ 🔄 Setting up dynamic DNS... Done (45s)
└─ ✅ External access configured

Phase 4: Monitoring and Backup Setup
├─ 🔄 Deploying Prometheus... Done (1m 50s)
├─ 🔄 Configuring Grafana dashboards... Done (2m 20s)
├─ 🔄 Setting up automated backups... Done (1m 15s)
└─ ✅ Monitoring and backup active

🎉 Homelab Setup Complete!
==========================

🌐 Access URLs:
├─ Plex: https://plex.homelab.example.com
├─ Jellyfin: https://jellyfin.homelab.example.com  
├─ Home Assistant: https://ha.homelab.example.com
├─ Grafana: https://grafana.homelab.example.com
└─ Router/Admin: https://admin.homelab.example.com

🔐 Initial Credentials:
├─ Home Assistant: Check /var/log/homeassistant/setup.log
├─ Grafana: admin / (generated password in /var/lib/grafana/admin_password)
├─ WireGuard config: /etc/wireguard/client.conf
└─ All services use SSO via Authentik

📱 Mobile Access:
├─ Download WireGuard app
├─ Import config from /etc/wireguard/client.conf  
├─ Connect to access all services securely
└─ Home Assistant app works with internal URL

📊 System Status:
├─ All services: ✅ Running
├─ SSL certificates: ✅ Valid
├─ Backup system: ✅ Active
├─ Monitoring: ✅ Collecting data
└─ Overall health: ✅ Excellent

📋 Next Steps:
1. Configure media libraries in Plex/Jellyfin
2. Set up Home Assistant devices and automations
3. Import existing data and backups
4. Configure monitoring alerts in Grafana
5. Test VPN access from mobile devices
6. Review security settings and firewall rules

💡 Pro Tips:
- Use Grafana to monitor system performance
- Set up Plex remote access for external streaming
- Configure Home Assistant automations for energy savings
- Regular backup testing with automated restore verification
- Monitor disk space - media libraries grow quickly!
```

### Advanced Template and Snippet Features

#### Template Versioning and Updates

```sh
# Check for template updates
nixai templates update-check --all

# Update specific template
nixai templates update gaming-nvidia --preview-changes

# Rollback to previous template version
nixai templates rollback gaming-nvidia --version 1.2.3
```

#### Snippet Collections and Sharing

```sh
# Export snippet collection for team sharing
nixai snippets export --tag "corporate" --format tar.gz \
  --output corporate-snippets-v1.0.tar.gz

# Import shared snippets
nixai snippets import ./team-snippets.tar.gz --verify-signatures

# Create snippet from successful configuration
nixai snippets extract-from-system "nginx-ssl-config" \
  --service nginx --auto-tag --description "Production SSL configuration"
```

### Integration with Other nixai Features

**Health Checks Integration**:
```sh
# Validate template before applying
nixai templates validate gaming-nvidia --health-check

# Check system compatibility
nixai templates compatibility-check media-server --hardware-scan
```

**AI-Powered Template Recommendations**:
```sh
# Get personalized template suggestions
nixai "What template would work best for a machine learning development setup?"

# Template optimization suggestions
nixai templates optimize homelab-complete --analyze-usage --suggest-improvements
```

**Documentation Integration**:
```sh
# Get detailed explanation of template components
nixai explain-option --template gaming-nvidia "services.xserver.videoDrivers"

# Generate template documentation
nixai templates document corporate-dev-standard --output-format markdown
```

---

## Interactive Mode

The `nixai interactive` command launches a powerful conversational shell that provides guided assistance for all nixai features. This mode is ideal for exploration, troubleshooting, and learning NixOS concepts through an AI-powered interface.

### Real-World Interactive Scenarios

#### Scenario 1: New User Learning NixOS

**Situation**: A developer new to NixOS needs help understanding and configuring their system.

```sh
nixai interactive
```

**Interactive Session**:
```text
🤖 Welcome to nixai Interactive Mode!
Type 'help' for available commands, 'quit' to exit.

nixai> help
Available commands:
  diagnose          - Diagnose system issues and logs
  explain-option    - Explain NixOS configuration options
  explain-home-option - Explain Home Manager options
  search           - Search packages and services
  health           - Run system health checks
  package-repo     - Analyze repositories for Nix packaging
  devenv           - Create development environments
  set-nixos-path   - Change NixOS configuration path
  set ai           - Switch AI provider
  show-config      - Display current configuration
  quit             - Exit interactive mode

nixai> explain-option services.openssh.enable
```

**AI Response**:
```text
🔧 NixOS Option: services.openssh.enable

Description:
Whether to enable the OpenSSH secure shell daemon, which allows secure remote logins.

Type: boolean
Default: false

🚀 Real-World Usage:
When enabled, this starts the SSH daemon (sshd) allowing remote access to your system.

Example Configuration:
```nix
services.openssh = {
  enable = true;
  settings = {
    PasswordAuthentication = false;  # Security: key-only auth
    PermitRootLogin = "no";          # Security: disable root login
    Port = 22;                       # Default SSH port
  };
};
```

🔒 Security Best Practices:
1. Always disable password authentication in production
2. Use SSH keys for authentication
3. Consider changing the default port (22) to reduce automated attacks
4. Enable fail2ban for additional protection

💡 Related Options:
- services.openssh.settings.PasswordAuthentication
- services.openssh.settings.PermitRootLogin  
- services.openssh.settings.Port
- services.fail2ban.enable (for additional security)

nixai> search package firefox
```

**Package Search Results**:
```text
🔍 Search Results for 'firefox':

📦 Packages:
1. firefox - Mozilla Firefox web browser
2. firefox-esr - Mozilla Firefox Extended Support Release
3. firefox-bin - Firefox binary package (faster install)
4. firefox-devedition-bin - Firefox Developer Edition

📋 Configuration Examples:
1. System-wide installation:
   environment.systemPackages = [ pkgs.firefox ];

2. Home Manager installation:
   home.packages = [ pkgs.firefox ];

3. Firefox with custom configuration:
   programs.firefox = {
     enable = true;
     profiles.default = {
       settings = {
         "browser.startup.homepage" = "https://nixos.org";
       };
     };
   };

nixai> 1
```

**Detailed Package Information**:
```text
🦊 firefox - Mozilla Firefox web browser

Package Details:
- Version: 121.0
- License: MPL 2.0
- Homepage: https://www.mozilla.org/firefox/
- Platforms: x86_64-linux, aarch64-linux, x86_64-darwin

Installation Options:
1. Add to system packages
2. Install via Home Manager  
3. Configure with programs.firefox

Would you like me to show configuration examples? (y/n)
```

#### Scenario 2: Troubleshooting System Issues

**Situation**: A user is experiencing boot issues and needs guided troubleshooting.

```text
nixai> diagnose
Paste or type your log, or provide a log file path: 
/var/log/nixos/nixos-rebuild.log

🔍 Analyzing log file: /var/log/nixos/nixos-rebuild.log

Log Analysis Results:
==================

🚨 Critical Issues Found:
1. Systemd service failure: postgresql.service
   - Error: Failed to start PostgreSQL database server
   - Likely cause: Permission issues with data directory
   
2. Configuration syntax error in networking configuration
   - Line 45: Invalid IP address format
   - Expected: "192.168.1.100/24"
   - Found: "192.168.1.100/255"

⚠️  Warnings:
1. Deprecated option usage: services.xserver.enable
   - Suggestion: Use services.xserver.displayManager.* instead

🔧 Recommended Actions:

Priority 1 - Fix PostgreSQL:
1. Check data directory permissions:
   sudo ls -la /var/lib/postgresql/
   
2. Reset PostgreSQL data directory:
   sudo systemctl stop postgresql
   sudo rm -rf /var/lib/postgresql/14/main
   sudo systemctl start postgresql

Priority 2 - Fix networking configuration:
1. Edit /etc/nixos/configuration.nix
2. Change line 45 from "192.168.1.100/255" to "192.168.1.100/24"

Priority 3 - Update deprecated options:
1. Replace services.xserver.enable with modern display manager options

Would you like detailed instructions for any of these fixes? (y/n)

nixai> y

Which issue would you like detailed help with?
1. PostgreSQL service failure
2. Networking configuration
3. Deprecated options
4. All of the above

nixai> 1
```

**Detailed PostgreSQL Fix Guide**:
```text
🐘 PostgreSQL Service Failure - Detailed Fix Guide

Problem Analysis:
The PostgreSQL service failed to start, likely due to data directory permissions or corruption.

Step-by-Step Resolution:

1. Check Current Status:
   ```sh
   sudo systemctl status postgresql.service
   sudo journalctl -u postgresql.service --since "1 hour ago"
   ```

2. Check Data Directory:
   ```sh
   sudo ls -la /var/lib/postgresql/
   sudo ls -la /var/lib/postgresql/14/
   ```

3. Safe Reset Process:
   ```sh
   # Stop the service
   sudo systemctl stop postgresql.service
   
   # Backup existing data (if any)
   sudo cp -r /var/lib/postgresql/14/main /var/lib/postgresql/14/main.backup
   
   # Remove corrupted data
   sudo rm -rf /var/lib/postgresql/14/main
   
   # Rebuild NixOS to recreate service
   sudo nixos-rebuild switch
   
   # Start PostgreSQL
   sudo systemctl start postgresql.service
   ```

4. Verify Fix:
   ```sh
   sudo systemctl status postgresql.service
   sudo -u postgres psql -c "SELECT version();"
   ```

5. Restore Data (if backup exists):
   ```sh
   sudo -u postgres pg_dump backup_db > backup.sql
   sudo -u postgres psql new_db < backup.sql
   ```

Prevention Tips:
- Regular database backups
- Monitor disk space in /var/lib/postgresql/
- Use proper PostgreSQL upgrade procedures

Need help with the next issue? Type 'help 2' for networking configuration.
```

#### Scenario 3: Development Environment Setup

**Situation**: A developer wants to set up a comprehensive development environment.

```text
nixai> devenv suggest "full-stack development with react frontend, node backend, and database"

🤖 AI Development Environment Analysis:

Based on your description, I recommend a multi-environment setup:

🎯 Recommended Approach:
1. Frontend: nodejs template with React and TypeScript
2. Backend: nodejs template with Express/Fastify  
3. Database: PostgreSQL service integration

📋 Suggested Commands:

Frontend Setup:
  nixai devenv create nodejs frontend --framework react --with-typescript

Backend Setup:  
  nixai devenv create nodejs backend --framework express --services postgres,redis

Full-Stack Alternative:
  nixai devenv create nodejs fullstack-app --framework nextjs --with-typescript --services postgres,redis

💡 Why This Configuration:
- Next.js provides both frontend and backend in one project
- TypeScript for better development experience
- PostgreSQL for persistent data storage
- Redis for caching and session management

Would you like me to create one of these environments? (y/n)

nixai> devenv create nodejs fullstack-app --framework nextjs --with-typescript --services postgres,redis --interactive

🚀 Creating Development Environment
Template: nodejs
Project: fullstack-app

📋 Template Configuration
Configure template options (press Enter for defaults):

Node.js version (nodejs_version) [20]: 20
Package manager (package_manager) [npm] (choices: npm, yarn, pnpm): pnpm
Web framework (framework) [nextjs]: nextjs
Include TypeScript (with_typescript) [true]: true
Additional packages (extra_packages) []: @next/mdx, prisma

✅ Development environment created successfully!

Location: /home/user/projects/fullstack-app
devenv.nix: /home/user/projects/fullstack-app/devenv.nix

Generated Configuration:
- Next.js 14+ with TypeScript
- pnpm package manager
- PostgreSQL database (accessible at localhost:5432)
- Redis cache (accessible at localhost:6379)
- ESLint, Prettier, and development tools
- Custom scripts: dev, build, start, typecheck

Next Steps:
1. cd fullstack-app
2. devenv shell  # Enter development environment
3. devenv up     # Start PostgreSQL and Redis services
4. pnpm install  # Install dependencies
5. pnpm dev      # Start development server

💡 Tips:
- Edit devenv.nix to add more packages or services
- Use environment variables: DATABASE_URL, REDIS_URL (auto-configured)
- Access services: PostgreSQL (port 5432), Redis (port 6379)
```

#### Scenario 4: System Configuration Management

**Situation**: Managing complex NixOS configuration with multiple options.

```text
nixai> show-config
Current nixai Configuration:
===========================

AI Provider: ollama
AI Model: llama3
Log Level: info
NixOS Path: /etc/nixos
MCP Server: Running (PID: 12345)

Available Commands: 18
Active Features: All enabled

nixai> set-nixos-path /home/user/nixos-config

✅ NixOS configuration path updated: /home/user/nixos-config

nixai> explain-option networking.firewall

🔒 NixOS Option: networking.firewall

The networking.firewall options control the built-in Linux firewall (iptables/nftables).

Key Options:
├─ networking.firewall.enable = true/false
├─ networking.firewall.allowedTCPPorts = [ 22 80 443 ]
├─ networking.firewall.allowedUDPPorts = [ 53 ]
├─ networking.firewall.allowPing = true/false
└─ networking.firewall.interfaces.<name>.allowedTCPPorts = [ ... ]

🏠 Real-World Configuration Examples:

Basic Web Server:
```nix
networking.firewall = {
  enable = true;
  allowedTCPPorts = [ 22 80 443 ];  # SSH, HTTP, HTTPS
  allowedUDPPorts = [ ];
  allowPing = false;  # Security: disable ping
};
```

Development Machine:
```nix
networking.firewall = {
  enable = true;
  allowedTCPPorts = [ 22 3000 8080 ];  # SSH + dev servers
  allowPing = true;
  interfaces = {
    "wlp2s0" = {  # Specific interface rules
      allowedTCPPorts = [ 8000 8001 ];
    };
  };
};
```

🔧 Interface-Specific Rules:
Use networking.firewall.interfaces.<interface> for granular control:
- "eth0" for ethernet
- "wlp2s0" for WiFi (check with 'ip link')
- "lo" for loopback

nixai> search service postgresql

🐘 PostgreSQL Service Configuration:

Service: services.postgresql
Description: PostgreSQL database server configuration

Key Options:
├─ services.postgresql.enable = true
├─ services.postgresql.package = pkgs.postgresql_15
├─ services.postgresql.dataDir = "/var/lib/postgresql/15"
├─ services.postgresql.authentication = "..."
└─ services.postgresql.settings = { ... }

📋 Common Configurations:

Basic Setup:
```nix
services.postgresql = {
  enable = true;
  package = pkgs.postgresql_15;
  ensureDatabases = [ "myapp" ];
  ensureUsers = [{
    name = "myapp";
    ensureDBOwnership = true;
  }];
};
```

Advanced Configuration:
```nix
services.postgresql = {
  enable = true;
  package = pkgs.postgresql_15;
  dataDir = "/var/lib/postgresql/15";
  authentication = ''
    local all all trust
    host all all 127.0.0.1/32 md5
    host all all ::1/128 md5
  '';
  settings = {
    max_connections = 200;
    shared_buffers = "256MB";
    effective_cache_size = "1GB";
    log_statement = "all";
  };
  ensureDatabases = [ "production" "development" "test" ];
  ensureUsers = [
    {
      name = "app_user";
      ensureDBOwnership = true;
    }
    {
      name = "read_only";
      ensurePermissions = {
        "DATABASE production" = "CONNECT";
        "SCHEMA public" = "USAGE";
        "ALL TABLES IN SCHEMA public" = "SELECT";
      };
    }
  ];
};
```

Would you like to see backup configuration or performance tuning? (y/n)
```

### Advanced Interactive Features

#### Context-Aware Assistance

```text
nixai> set ai gemini
✅ AI provider switched to: gemini

nixai> explain-option --context development services.docker

🐳 Docker Service for Development (Context: Development Environment)

services.docker is essential for containerized development workflows.

Development-Focused Configuration:
```nix
virtualisation.docker = {
  enable = true;
  enableOnBoot = true;
  autoPrune = {
    enable = true;
    dates = "weekly";
  };
};

# Add user to docker group for non-root access
users.users.youruser.extraGroups = [ "docker" ];

# Development tools
environment.systemPackages = with pkgs; [
  docker-compose
  docker-buildx
  dive  # Docker image analysis
];
```

💻 Development Best Practices:
1. Enable BuildKit: export DOCKER_BUILDKIT=1
2. Use multi-stage builds for smaller images
3. Regular cleanup: docker system prune
4. Consider Podman as rootless alternative

🔧 Common Development Patterns:
- Bind mount source code: -v $(pwd):/app
- Use docker-compose for multi-container setups
- Leverage .dockerignore for faster builds
```

#### Guided Troubleshooting Workflows

```text
nixai> health

🏥 System Health Check Results:
============================

✅ Configuration: Valid
⚠️  Services: 2 issues detected
✅ Disk Space: Sufficient
❌ Network: 1 critical issue
✅ Boot: Working correctly

Would you like guided assistance for the issues found? (y/n)

nixai> y

🔧 Guided Issue Resolution:

Issue 1: Failed service 'bluetooth.service'
Severity: Medium
Impact: Bluetooth devices not working

Diagnostic Steps:
1. Check hardware: lsusb | grep -i bluetooth
2. Check service status: systemctl status bluetooth.service
3. Review logs: journalctl -u bluetooth.service

Quick Fix Options:
a) Restart service: sudo systemctl restart bluetooth.service
b) Reset Bluetooth: sudo systemctl restart bluetooth && sudo bluetoothctl
c) Hardware reset: Remove/reinsert USB Bluetooth adapter

Which option would you like to try? (a/b/c/manual)

nixai> a

Executing: sudo systemctl restart bluetooth.service
✅ Command completed successfully

Verifying fix...
✅ bluetooth.service is now active and running
✅ Bluetooth functionality restored

Issue 2: Network connectivity problems detected
Severity: High
Impact: Internet access affected

Running network diagnostics...
- Ping test: ❌ Failed
- DNS resolution: ❌ Failed  
- Default gateway: ✅ Reachable
- Network interface: ✅ Up

This suggests a DNS configuration issue.

Would you like me to:
1. Check DNS servers configuration
2. Reset network configuration
3. Run advanced network diagnostics
4. Show manual troubleshooting steps

nixai> 1

🌐 DNS Configuration Analysis:

Current DNS servers:
- Primary: 8.8.8.8 (Google)
- Secondary: 1.1.1.1 (Cloudflare)

Testing DNS servers:
- 8.8.8.8: ❌ Timeout
- 1.1.1.1: ❌ Timeout
- Router DNS: ✅ Working

🔧 Recommended Fix:
Your ISP/router DNS is working, but external DNS servers are blocked.

Configuration Update:
```nix
networking.nameservers = [ "192.168.1.1" ];  # Use router DNS
```

Apply this fix? (y/n)
```

### Interactive Mode Power Features

#### Command Completion and Context

```text
nixai> explain-<TAB>
explain-option        explain-home-option

nixai> explain-option services.<TAB>
services.nginx        services.postgresql   services.docker
services.openssh      services.xserver      services.bluetooth

nixai> explain-option services.nginx.<TAB>
services.nginx.enable          services.nginx.virtualHosts
services.nginx.package         services.nginx.user
services.nginx.group           services.nginx.defaultListen
```

#### Multi-Command Workflows

```text
nixai> diagnose --log-file /var/log/nixos/rebuild.log && health && explain-option services.postgresql.settings

🔄 Executing multi-command workflow...

Step 1: Diagnosing log file...
[Diagnosis results]

Step 2: Running health check...
[Health check results]

Step 3: Explaining PostgreSQL settings...
[Option explanation]

✅ Workflow completed. All commands executed successfully.
```

#### Session Management

```text
nixai> save-session troubleshooting-2024-01-15
✅ Session saved with 15 commands and results

nixai> load-session troubleshooting-2024-01-15
✅ Session loaded. Previous context restored.

nixai> show-history
Recent Commands:
1. explain-option services.postgresql.enable
2. search package nginx
3. diagnose --log-file /var/log/rebuild.log
4. health
5. devenv create python api-server
[...]

nixai> repeat 3
Repeating command: diagnose --log-file /var/log/rebuild.log
[Diagnosis results]
```

### Interactive Mode Benefits

**For New Users**:
- Guided learning with contextual explanations
- Tab completion for discovering options and commands
- Safe exploration without making system changes
- Built-in help and examples

**For Experienced Users**:
- Rapid prototyping of configurations
- Multi-command workflows
- Context switching between different configuration paths
- Advanced debugging and analysis tools

**For System Administrators**:
- Automated health checks and diagnostics
- Guided troubleshooting workflows
- Configuration validation and optimization
- Integration with existing monitoring workflows

**Getting Started with Interactive Mode**:
1. Launch: `nixai interactive`
2. Get help: `help`
3. Set your context: `set-nixos-path /path/to/config`
4. Choose AI provider: `set ai ollama` (or gemini/openai)
5. Start exploring: `explain-option`, `search`, `diagnose`, etc.

The interactive mode transforms nixai from a command-line tool into an intelligent NixOS assistant that guides you through complex configuration and troubleshooting tasks.

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

## 🩺 Doctor Command: System Diagnostics & Troubleshooting

The `nixai doctor` command provides a comprehensive diagnostics report for your nixai setup, including:

- **MCP Server Health**: Checks if the documentation server is running, healthy, and accessible (port/socket, process, /healthz endpoint)
- **AI Provider Health**: Verifies connectivity and configuration for Ollama, OpenAI, and Gemini (API reachability, key presence/validity)
- **Actionable Feedback**: Glamour-formatted output with clear next steps for resolving common issues

### Usage Example

```sh
nixai doctor
```

**Example Output:**

```text
🩺 nixai Doctor: System Diagnostics
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

MCP Server Diagnostics
━━━━━━━━━━━━━━━━━━━━━━
✅ MCP server is running and healthy on http://localhost:8081.
✅ Port is open: localhost:8081
✅ MCP server process is running (pgrep matched).

AI Provider Diagnostics
━━━━━━━━━━━━━━━━━━━━━━━
✅ Ollama API reachable at http://localhost:11434
✅ OpenAI API reachable (key valid).
✅ Gemini API reachable (key valid).

ℹ️  See the README or docs/MANUAL.md for troubleshooting steps and more details.
```

If any issues are detected, nixai doctor will provide warnings, errors, and actionable suggestions (e.g., how to start the MCP server, set API keys, or check service status).

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

The `nixai devenv` command provides rapid creation of reproducible development environments for Python, Rust, Node.js, and Go using [devenv.sh](https://devenv.sh/) and Nix. This powerful system offers language-specific templates, framework/tool options, service/database integration, and AI-powered project suggestions to streamline development setup.

### Real-World Development Environment Scenarios

#### Scenario 1: Full-Stack Web Application Setup

**Situation**: Starting a new e-commerce project with Next.js frontend, Go backend, and PostgreSQL database.

```sh
# Create the frontend application
nixai devenv create nodejs ecommerce-frontend --framework nextjs --with-typescript --services postgres
```

**Generated Environment Features**:
- Next.js 14+ with TypeScript configuration
- ESLint, Prettier, and development tools
- PostgreSQL database service
- Hot reloading and development scripts
- Environment variables pre-configured

**Generated devenv.nix excerpt**:
```nix
{
  languages.javascript = {
    enable = true;
    npm.enable = true;
  };
  
  packages = [
    "nodejs_20"
    "nodePackages.typescript"
    "nodePackages.eslint"
    "nodePackages.prettier"
  ];

  services.postgres = {
    enable = true;
    initialDatabases = [{ name = "ecommerce-frontend"; }];
  };

  scripts = {
    dev = "next dev";
    build = "next build";
    start = "next start";
    typecheck = "tsc --noEmit";
  };

  env = {
    NODE_ENV = "development";
    DATABASE_URL = "postgres://postgres@localhost/ecommerce-frontend";
  };
}
```

**Next Steps**:
```sh
cd ecommerce-frontend
devenv shell    # Enter development environment
devenv up       # Start PostgreSQL service
npm run dev     # Start Next.js development server
```

#### Scenario 2: Machine Learning Pipeline Development

**Situation**: Setting up a data science environment for computer vision research.

```sh
# Get AI suggestion first
nixai devenv suggest "machine learning project with jupyter notebooks and data visualization"
```

**AI Response**:
```text
🤖 AI Analysis & Recommendations:

Recommended Template: python
Description: Python development environment with data science tools

Suggested Configuration:
- Package Manager: poetry (for dependency management)
- Additional Tools: jupyter, pandas, numpy, matplotlib
- Services: postgres (for data storage)
- Framework: Use with_jupyter=true for notebook support

Create Project:
  nixai devenv create python ml-vision-project --with-poetry --with-jupyter --services postgres
```

**Creating the Environment**:
```sh
nixai devenv create python ml-vision-project --with-poetry --with-jupyter --services postgres
```

**Enhanced Setup with Custom Options**:
```sh
cd ml-vision-project

# Add the generated devenv.nix includes:
# - Python 3.11 environment
# - Poetry for dependency management
# - Jupyter Lab and Notebook
# - PostgreSQL for data storage
# - Development tools (black, flake8, mypy)

# Enter the environment
devenv shell

# Start services
devenv up

# Install ML dependencies
poetry add torch torchvision opencv-python pandas numpy matplotlib seaborn scikit-learn

# Launch Jupyter Lab
jupyter lab
```

#### Scenario 3: Microservices Development with gRPC

**Situation**: Building a distributed system with multiple Go microservices communicating via gRPC.

```sh
# Create the first microservice
nixai devenv create golang user-service --framework gin --with-grpc --services postgres,redis

# Create the second microservice
nixai devenv create golang order-service --framework gin --with-grpc --services postgres,redis
```

**Generated Environment Features**:
- Go 1.21+ with proper CGO support
- gRPC and Protocol Buffers toolchain
- Gin web framework setup
- PostgreSQL and Redis services
- Development tools (golangci-lint, delve debugger)
- Air for live reloading

**Example Generated Scripts**:
```sh
# Available in both services
go run .              # Build and run the service
go build              # Build the service
go test ./...         # Run all tests
go fmt ./...          # Format code
golangci-lint run     # Run linter
air                   # Start with live reloading (if air enabled)
protoc                # Protocol buffer compiler
proto-gen             # Generate Go code from .proto files
```

#### Scenario 4: Rust WebAssembly Project

**Situation**: Creating a high-performance web application with Rust backend compiled to WebAssembly.

```sh
nixai devenv create rust wasm-game --with-wasm --services redis
```

**Generated Environment Includes**:
- Rust stable toolchain
- WebAssembly tools (wasm-pack, binaryen)
- Redis for session storage
- Development utilities

**Development Workflow**:
```sh
cd wasm-game
devenv shell

# Initialize Rust project
cargo init --name wasm-game

# Build for WebAssembly
wasm-pack build

# Test WebAssembly
wasm-pack test --headless --firefox

# Start Redis service
devenv up

# Development with hot reload
cargo watch -x "run"
```

#### Scenario 5: Interactive Project Creation

**Situation**: Setting up a Python web API with guided configuration.

```sh
nixai devenv create python api-server --interactive --services postgres,redis
```

**Interactive Configuration Flow**:
```text
🚀 Creating Development Environment
Template: python
Project: api-server

📋 Template Configuration
Configure template options (press Enter for defaults):

Python version (python_version) [311]: 
Package manager (package_manager) [pip] (choices: pip, poetry, pipenv): poetry
Web framework (framework) [] (choices: flask, fastapi, django): fastapi
Include Jupyter support (with_jupyter) [false]: true
Include Django (with_django) [false]: 
Include FastAPI (with_fastapi) [true]: true

✅ Development environment created successfully!

Location: /home/user/projects/api-server
devenv.nix: /home/user/projects/api-server/devenv.nix

Next Steps:
  1. cd api-server
  2. devenv shell  # Enter the development environment  
  3. devenv up     # Start services (if any)

💡 Edit devenv.nix to customize your environment
💡 Use 'devenv --help' to learn more about devenv commands
```

### Advanced Development Environment Features

#### Custom Environment Variables and Configuration

```sh
# Create environment with custom settings
nixai devenv create golang api-gateway --framework gin --services postgres

# The generated devenv.nix allows customization:
```

**Customizable devenv.nix**:
```nix
{
  languages.go = {
    enable = true;
    version = "1.21";
  };

  packages = [
    "git" "curl" "gcc"
    "golangci-lint" "delve" "gopls"
  ];

  services.postgres = {
    enable = true;
    initialDatabases = [{ name = "api-gateway"; }];
  };

  env = {
    CGO_ENABLED = "1";
    DATABASE_URL = "postgres://postgres@localhost/api-gateway";
    LOG_LEVEL = "debug";
    API_PORT = "8080";
  };

  scripts = {
    dev = "air";
    build = "go build -o bin/api-gateway";
    test = "go test ./...";
    migrate = "go run cmd/migrate/main.go";
  };

  enterShell = ''
    echo "🚀 API Gateway development environment"
    echo "Database: $DATABASE_URL"
    echo "Run 'devenv up' to start PostgreSQL"
  '';
}
```

#### Multi-Service Development Setup

```sh
# Set up environment for microservices development
mkdir microservices && cd microservices

# Create multiple related services
nixai devenv create golang auth-service --framework gin --with-grpc --services postgres
nixai devenv create golang user-service --framework gin --with-grpc --services postgres  
nixai devenv create nodejs frontend --framework nextjs --with-typescript --services redis

# Each gets its own devenv.nix optimized for its role
```

#### AI-Powered Template Suggestions

```sh
# Get suggestions for complex setups
nixai devenv suggest "real-time chat application with websockets and user authentication"
```

**Example AI Response**:
```text
🤖 AI Analysis & Recommendations:

Based on your description "real-time chat application with websockets and user authentication", 
I recommend the following setup:

Primary Recommendation: golang template
- Go excels at concurrent WebSocket handling
- Strong ecosystem for real-time applications
- Excellent performance for chat systems

Suggested Configuration:
├─ Framework: gin (lightweight, good WebSocket support)
├─ Services: postgres (user data), redis (session management, message caching)
├─ Additional Tools: WebSocket libraries, JWT authentication
└─ Database: PostgreSQL for persistent data, Redis for real-time features

Alternative Option: nodejs template
- Excellent WebSocket support with Socket.io
- Rich ecosystem for real-time applications
- Good TypeScript support for larger applications

Create Project:
  nixai devenv create golang chat-server --framework gin --services postgres,redis
  
Next Steps:
1. Add WebSocket support: go get github.com/gorilla/websocket
2. Implement JWT authentication
3. Set up Redis pub/sub for message broadcasting
4. Configure PostgreSQL for user and message storage
```

### Development Environment Best Practices

#### Template Selection Guide

- **Python**: Data science, ML, web APIs (Flask/FastAPI), automation
- **Rust**: Systems programming, WebAssembly, high-performance applications  
- **Node.js**: Frontend, full-stack web, real-time applications, APIs
- **Go**: Microservices, APIs, DevOps tools, concurrent applications

#### Service Integration Patterns

**Database Services**:
- `postgres`: Primary database for most applications
- `mysql`: Alternative relational database
- `mongodb`: Document database for flexible schemas  
- `redis`: Caching, sessions, real-time features

**Development Workflow**:
1. Use `nixai devenv suggest` for AI-powered recommendations
2. Create environment with appropriate template and services
3. Enter environment with `devenv shell`
4. Start services with `devenv up`
5. Customize `devenv.nix` as needed
6. Use provided scripts for common development tasks

#### Performance and Productivity Tips

- **Use Interactive Mode** (`--interactive`) for guided setup
- **Leverage AI Suggestions** for complex project requirements
- **Customize Scripts** in devenv.nix for project-specific workflows
- **Version Pin Dependencies** for reproducible environments
- **Document Environment Setup** for team collaboration

### Troubleshooting Development Environments

#### Common Issues and Solutions

**Issue**: "devenv command not found"
```sh
# Solution: Install devenv globally or use nix-shell
nix-env -iA nixpkgs.devenv
# Or use nix-shell
nix-shell -p devenv
```

**Issue**: "Service failed to start"
```sh
# Solution: Check service logs and configuration
devenv processes          # List running processes
devenv logs postgres      # View service logs
devenv ps                 # Check process status
```

**Issue**: "Package conflicts or missing dependencies"
```sh
# Solution: Review and update devenv.nix
# Check package availability
nix search nixpkgs package-name
# Update packages list in devenv.nix
```

**Issue**: "Environment variables not set correctly"
```sh
# Solution: Verify env section in devenv.nix and restart shell
devenv shell
echo $DATABASE_URL        # Verify variable is set
```

#### Integration with Other nixai Features

**Health Checks**:
```sh
nixai health              # Will check devenv environments
```

**Option Documentation**:
```sh
nixai explain-option "languages.python"
nixai explain-option "services.postgres"
```

**Direct Questions**:
```sh
nixai "How do I add a new package to my devenv environment?"
nixai "Best practices for PostgreSQL configuration in devenv"
```

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

## Flake Creation & Correction (`nixai flake create`)

The `nixai flake create` command helps you quickly create, correct, or upgrade a `flake.nix` for any project folder, with AI-powered build system detection and best-practice suggestions.

### Features

- Create a minimal flake from scratch (`--from-scratch`)
- Analyze a project folder and generate a flake for Go, Node, Rust, Python, or generic projects (`--analyze`)
- Correct and upgrade an existing `flake.nix` with AI assistance (`--fix`)
- Overwrite with `--force`
- Customize system and description

### Usage Examples

#### Create a minimal flake.nix in the current directory

```sh
nixai flake create --from-scratch
```

#### Analyze a project folder and generate a flake.nix

```sh
nixai flake create . --analyze
```

#### Fix and update an existing flake.nix using AI

```sh
nixai flake create . --fix --force
```

#### Specify system and description

```sh
nixai flake create myproject --from-scratch --system x86_64-linux --desc "My Project Flake"
```

#### Real-life Example: Node.js Project

Suppose you have a Node.js project with a `package.json` file. Run:

```sh
nixai flake create . --analyze
```

This will generate a `flake.nix` using `buildNpmPackage` and fill in the project name and metadata. The output will include an AI explanation and best practices for your flake.

#### Real-life Example: Rust Project

For a Rust project with a `Cargo.toml` file:

```sh
nixai flake create . --analyze
```

The tool will detect Rust, use `buildRustPackage`, and generate a suitable flake.nix.

#### Real-life Example: Fixing an Existing Flake

If you have a `flake.nix` that is incomplete or broken:

```sh
nixai flake create . --fix --force
```

The AI will review, correct, and optionally overwrite your flake.nix, providing a summary of changes and best practices.

---
