# Docker NixOS Test Environment for nixai

This directory contains a complete Docker-based NixOS 25.05 test environment for the nixai project. It creates an isolated testing environment with nixai cloned and built inside the container, eliminating permission issues and providing a clean NixOS-like environment.

## Quick Start

```zsh
# Build and start the Docker container (with isolated nixai repo)
./docker_nixos/build_and_run_docker.sh

# Inside the container, nixai is pre-installed and ready to use
nixai --help
nixai "How do I configure git in NixOS?"
```

## What's Included

The Docker image provides:
- **NixOS 25.05** environment with Nix package manager
- **Isolated nixai repository** cloned inside the container at `/workspace`
- **Pre-installed nixai** binary (`/usr/local/bin/nixai`)
- **Development tools**: Go, Just, Neovim, Git, Curl, Python3, Node.js, Alejandra
- **Nix flakes** support enabled
- **Ollama integration** configured for `host.docker.internal:11434`
- **AI provider keys** placeholder configuration
- **Complete isolation** - no permission or mounting issues

## File Structure

```text
docker_nixos/
├── Dockerfile              # Main Docker image definition
├── build_and_run_docker.sh # Build and run script
└── README.md               # This documentation
```

## Building and Running

### Option 1: Automated Script (Recommended)

```zsh
# From the repository root
./docker_nixos/build_and_run_docker.sh
```

This script:

1. Builds the Docker image with tag `nixai-nixos-test`
2. Clones the nixai repository inside the container (isolated environment)
3. Adds `host.docker.internal` for Ollama access
4. Starts an interactive shell with all tools ready

### Option 2: Manual Docker Commands

```zsh
# Build the image
cd docker_nixos
docker build -t nixai-nixos-test .

# Run the container
docker run -it --rm \
  --add-host=host.docker.internal:host-gateway \
  -v "$(pwd)/..":/workspace \
  nixai-nixos-test
```

## Installation Methods Inside Container

The Docker environment provides multiple ways to build and install nixai:

### Method 1: Pre-installed Binary (Default)

nixai is automatically built and installed during Docker image creation:

```zsh
# Already available globally
nixai --help
nixai "explain git configuration in NixOS"
nixai explain-option programs.git
nixai explain-home-option programs.neovim
```

### Method 2: Development Build with Justfile

For development and testing with the mounted repository:

```zsh
cd /workspace
nix develop .#docker  # Enter Nix dev shell

# Build to /tmp (writable location)
just build-docker
just run-docker
just run-docker-args "explain-option services.nginx"

# Install globally in container
just install-docker
```

### Method 3: Standard Build (with permission fix)

```zsh
cd /workspace
nix develop .#docker

# Regular build commands work in /tmp
just build     # Builds to ./nixai (may fail due to permissions)
just build-docker  # Builds to /tmp/nixai (recommended)
```

### Method 4: Nix Build

```zsh
cd /workspace
nix develop .#docker

# Build with Nix
nix build
./result/bin/nixai --help
```

## Testing Inside Docker

### Basic Functionality Tests

```zsh
# Test basic commands
nixai --help
nixai --version

# Test direct questions
nixai "How do I install Neovim in NixOS?"
nixai "What is the difference between nix-env and nix profile?"

# Test option explanations
nixai explain-option programs.git
nixai explain-option services.openssh
nixai explain-home-option programs.neovim
```

### Advanced Testing

```zsh
cd /workspace
nix develop .#docker

# Run the test suite
just test
just test-all
just test-mcp
just test-providers

# Test with coverage
just test-coverage

# Test specific functionality
just test-nixos-parse /var/log/nixos-rebuild.log
just test-ai ollama
```

### MCP Server Testing

```zsh
# Start MCP server in background
just mcp-start

# Check status
just mcp-status

# View logs
just mcp-logs

# Stop server
just mcp-stop
```

### Interactive Mode Testing

```zsh
# Test interactive mode
nixai --interactive

# Test with debug logging
nixai --log-level debug "Configure git in NixOS"
```

## Development Workflow

### 1. Start Development Environment

```zsh
./docker_nixos/build_and_run_docker.sh
cd /workspace
nix develop .#docker
```

### 2. Make Code Changes

Edit files in your host editor - changes are immediately available in `/workspace`.

### 3. Build and Test

```zsh
# Quick build and test
just build-docker
/tmp/nixai --help

# Full development workflow
just dev  # deps, fmt, lint, test, build

# Install for global testing
just install-docker
nixai "test question"
```

### 4. Run Comprehensive Tests

```zsh
# All tests
just test-all

# Specific test suites
just test-mcp
just test-providers
just test-vscode
```

## Environment Configuration

### AI Provider Setup

AI provider keys are automatically loaded from `.ai_keys` in the repository root:

```bash
# .ai_keys format (shell export syntax)
export OPENAI_API_KEY=sk-...
export GEMINI_API_KEY=...
export ANTHROPIC_API_KEY=...
```

### Ollama Integration

The container is pre-configured to connect to your host Ollama instance:

```bash
# Environment variable (pre-set)
OLLAMA_HOST=http://host.docker.internal:11434

# Test Ollama connection
nixai --provider ollama "test question"
```

**Note**: Make sure Ollama is running on your host system and accessible on port 11434.

### Nix Configuration

Experimental features are enabled by default:

```
# /etc/nix/nix.conf
experimental-features = nix-command flakes
```

Git is configured to trust the `/workspace` directory for repository operations.

## Troubleshooting

### Permission Issues

If you encounter permission errors when building to `/workspace`:

```zsh
# Use Docker-specific build commands
just build-docker    # Builds to /tmp
just install-docker  # Installs globally
```

### Ollama Connection Issues

```zsh
# Check if Ollama is running on host
# On host system:
ollama serve

# In container, test connection:
curl http://host.docker.internal:11434/api/version
```

### Nix Issues

```zsh
# If flakes don't work, check experimental features
nix --version
nix flake check  # Should work without --experimental-features

# Force rebuild Nix environment
nix develop .#docker --rebuild
```

### Build Issues

```zsh
# Clean and rebuild
just clean
just build-docker

# Check Go environment
go version
go env GOPATH
go env GOROOT
```

## Available Justfile Commands

### Docker-Specific Commands

- `just build-docker` - Build to `/tmp/nixai`
- `just run-docker` - Run Docker-built binary
- `just run-docker-args "args"` - Run with arguments
- `just install-docker` - Install globally in container

### Development Commands

- `just build` - Standard build
- `just test` - Run tests
- `just test-all` - Run all test suites
- `just dev` - Full development workflow
- `just lint` - Code linting
- `just fmt` - Code formatting

### Testing Commands

- `just test-mcp` - MCP server tests
- `just test-providers` - AI provider tests
- `just test-vscode` - VS Code integration tests
- `just test-coverage` - Coverage analysis

### Nix Commands

- `just nix-build` - Build with Nix
- `just nix-develop` - Enter Nix shell

See `just help` for a complete list of available commands.

## Performance Tips

1. **Use Docker-specific commands** (`just build-docker`) for faster builds
2. **Pre-install dependencies** by rebuilding the Docker image when Go modules change
3. **Use Nix build** for production-like testing
4. **Mount only necessary directories** to improve I/O performance

## Integration Examples

### Testing NixOS Configuration Generation

```zsh
# Test configuration for a specific service
nixai "Generate NixOS config for Nginx with SSL"

# Test Home Manager configuration
nixai "Configure Neovim with LSP in Home Manager"
```

### Testing Documentation Queries

```zsh
# Test MCP documentation integration
nixai "How to configure SSH keys in NixOS?"
nixai "What are the best practices for NixOS modules?"
```

### Testing Package Analysis

```zsh
# Test package repository analysis
nixai package-repo analyze /path/to/some/repo
```

This Docker environment provides a complete, reproducible testing setup for nixai development and ensures all features work correctly in a NixOS-like environment.
