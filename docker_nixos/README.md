# Docker NixOS Test Environment for nixai

This directory contains a complete Docker-based NixOS 25.05 test environment for the nixai project. It creates an isolated testing environment with nixai cloned and built inside the container, eliminating permission issues and providing a clean NixOS-like environment.

## Quick Start

```zsh
# Build and run the Docker container with a single command
./docker_nixos/nixai-docker.sh

# Run the interactive demo to see all features
./docker_nixos/nixai-docker.sh demo

# Inside the container, nixai is pre-installed and ready to use
nixai --help
nixai "How do I configure git in NixOS?"
```

Additional options:
```zsh
# Only build the Docker image
./docker_nixos/nixai-docker.sh build

# Run without rebuilding
./docker_nixos/nixai-docker.sh --no-build run

# Run the test suite
./docker_nixos/nixai-docker.sh test

# Run the interactive demo
./docker_nixos/nixai-docker.sh demo
./docker_nixos/nixai-docker.sh --demo

# Show help
./docker_nixos/nixai-docker.sh --help
```

## Testing Ollama Connectivity

The Docker container automatically connects to Ollama running on your host machine:

```zsh
# Inside the container, check Ollama connectivity
./test_host_connectivity.sh
```

## Interactive Demo

Run the comprehensive nixai features demo to see all functionality in action:

```zsh
# Run the full interactive demo
./docker_nixos/nixai-docker.sh demo

# Alternative syntax
./docker_nixos/nixai-docker.sh --demo
```

The demo includes:
- Basic nixai functionality and help
- NixOS option explanations
- Home Manager integration
- Direct question answering
- AI provider testing
- Development environment showcase
- Building from source
- Testing framework
- MCP server integration
- Real-world usage examples

## What's Included

The Docker image provides:

- **NixOS 25.05** environment with Nix package manager
- **Isolated nixai repository** cloned inside the container
- **Pre-installed nixai** binary 
- **Development tools**: Go, Just, Neovim, Git, Curl
- **Nix flakes** support enabled
- **Ollama integration** configured for `host.docker.internal:11434` with auto-detection
- **AI provider keys** loaded from `.ai_keys` file into environment variables
- **Cross-platform support** for Linux, macOS with host.docker.internal resolution

## File Structure

```text
docker_nixos/
├── Dockerfile              # Main Docker image definition
├── nixai-docker.sh        # All-in-one build and run script
├── test_host_connectivity.sh  # Diagnoses Ollama connection issues
└── README.md              # This documentation
```

## Ollama Integration

The Docker container automatically connects to Ollama running on your host machine:

1. Before starting, ensure Ollama is running on your host:
   ```zsh
   ollama serve
   ```

2. Inside the container, test Ollama connectivity:
   ```zsh
   ./test_host_connectivity.sh
   ```

3. Use nixai with Ollama:
   ```zsh
   nixai --provider ollama "How do I configure SSH in NixOS?"
   ```

## AI Provider Setup

AI provider keys are automatically loaded from `.ai_keys` in the Docker context:

```bash
# .ai_keys format (shell export syntax)
export OPENAI_API_KEY=sk-...
export GEMINI_API_KEY=...
export OLLAMA_HOST=http://host.docker.internal:11434
```

The script creates a sample `.ai_keys` file if one doesn't exist.

## Installation Methods Inside Container

The Docker environment provides multiple ways to build and install nixai:

### 1. Pre-installed Binary (Default)

nixai is automatically built and installed during Docker image creation:

```zsh
# Already available globally
nixai --help
nixai "explain git configuration in NixOS"
nixai explain-option programs.git
nixai explain-home-option programs.neovim
```

### 2. Development Build with Justfile

For development and testing:

```zsh
cd /root/nixai
nix develop .#docker  # Enter Nix dev shell

# Build to /tmp (writable location)
just build-docker
just run-docker
just install-docker  # Install globally
```

## Testing Inside Docker

Test basic functionality:

```zsh
# Test basic commands
nixai --help

# Test direct questions
nixai "How do I install Neovim in NixOS?"

# Test option explanations
nixai explain-option programs.git
nixai explain-home-option programs.neovim
```

## Troubleshooting

### Ollama Connection Issues

```zsh
# Check if Ollama is running on host
# On host system:
ollama serve

# In container, test connection:
curl http://host.docker.internal:11434/api/version

# Run diagnostic script
./test_host_connectivity.sh
```

### Nix Issues

```zsh
# If flakes don't work, check experimental features
nix --version
nix flake check  # Should work without --experimental-features
```

This Docker environment provides a complete, reproducible testing setup for nixai development and ensures all features work correctly in a NixOS-like environment.
