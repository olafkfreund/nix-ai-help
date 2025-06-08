# nixai Installation Guide

This guide covers all installation methods for nixai, from simple one-command installs to advanced NixOS/Home Manager integration.

## 🚀 Quick Installation Methods

### Method 1: Direct Run (No Installation Required)

```bash
# Run nixai directly from GitHub
nix run github:olafkfreund/nix-ai-help -- --help
nix run github:olafkfreund/nix-ai-help -- -a "How do I configure SSH?"
```

### Method 2: System-wide Installation via Flakes

```bash
# Install to user profile
nix profile install github:olafkfreund/nix-ai-help

# Now nixai is available globally
nixai --help
```

## 📦 Non-Flake Installation Methods

### Method 3: callPackage Installation

For users who don't use flakes, add this to your NixOS `configuration.nix` or Home Manager configuration:

```nix
{ config, pkgs, ... }:

let
  nixai = pkgs.callPackage (builtins.fetchGit {
    url = "https://github.com/olafkfreund/nix-ai-help.git";
    ref = "main";
  } + "/package.nix") {};
in {
  # For NixOS system-wide installation
  environment.systemPackages = [ nixai ];
  
  # OR for Home Manager user installation
  # home.packages = [ nixai ];
}
```

### Method 4: Local Build with package.nix

```bash
# Clone the repository
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help

# Option A: Build using the standalone wrapper (recommended)
nix-build standalone.nix
./result/bin/nixai --help

# Option B: Build using callPackage manually
nix-build -E 'with import <nixpkgs> {}; callPackage ./package.nix {}'

# Install the built package
nix-env -i ./result
```

### Method 5: Manual nix-env Installation

```bash
# Install directly with nix-env
nix-env -if https://github.com/olafkfreund/nix-ai-help/archive/main.tar.gz
```

## 🏠 NixOS & Home Manager Integration

### NixOS Module Integration

Add to your `flake.nix` inputs:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, nixai, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        nixai.nixosModules.default
        {
          services.nixai = {
            enable = true;
            mcp.enable = true;
          };
        }
      ];
    };
  };
}
```

### Home Manager Integration

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    homeConfigurations.myuser = home-manager.lib.homeManagerConfiguration {
      pkgs = import nixpkgs { system = "x86_64-linux"; };
      modules = [
        nixai.homeManagerModules.default
        {
          services.nixai = {
            enable = true;
            mcp.enable = true;
          };
        }
      ];
    };
  };
}
```

## 🛠️ Development Installation

For contributors and developers:

```bash
# Clone and enter development environment
git clone https://github.com/olafkfreund/nix-ai-help.git
cd nix-ai-help

# Option 1: Use Nix flake dev shell
nix develop

# Option 2: Use justfile commands
just build
just test
./nixai --help
```

## 🔧 Prerequisites

### Required
- **Nix package manager** (with flakes enabled for flake-based methods)

### Optional but Recommended
- **Ollama** for local AI inference (privacy-first approach)
- **Git** for development
- **Just** for development tasks

### Installing Prerequisites

```bash
# Install Nix (if not already installed)
curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install

# Enable flakes (add to ~/.config/nix/nix.conf or /etc/nix/nix.conf)
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf

# Install Ollama for local AI
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3
```

## 🎯 Quick Start After Installation

```bash
# Check installation
nixai --version
nixai --help

# Ask your first question
nixai -a "How do I enable SSH in NixOS?"

# Start MCP server for documentation features
nixai mcp-server start

# Explain NixOS options
nixai explain-option services.openssh.enable

# Interactive mode
nixai --interactive
```

## 🚀 Advanced Configuration

### Custom AI Provider

```bash
# Use OpenAI instead of Ollama
export OPENAI_API_KEY="your-key-here"
nixai -a "Configure networking" --provider openai

# Use Gemini
export GEMINI_API_KEY="your-key-here" 
nixai -a "Setup firewall" --provider gemini
```

### Custom Configuration

```bash
# Generate default config
nixai --generate-config

# Edit configuration file
vim ~/.config/nixai/config.yaml

# Use custom config path
nixai --config /path/to/custom/config.yaml
```

## 📊 Comparison Table

| Method | Flakes Required | System-wide | User-level | Auto-updates | Best For |
|--------|----------------|-------------|------------|-------------|----------|
| Direct Run | ✅ | ❌ | ❌ | ✅ | Quick testing |
| Flake Profile | ✅ | ❌ | ✅ | Manual | Daily use |
| callPackage | ❌ | ✅ | ✅ | Manual | Traditional Nix |
| Local Build | ❌ | ❌ | ✅ | Manual | Development |
| nix-env | ❌ | ❌ | ✅ | Manual | Simple install |
| NixOS Module | ✅ | ✅ | ❌ | ✅ | System integration |
| Home Manager | ✅ | ❌ | ✅ | ✅ | User integration |

## 🔍 Troubleshooting

### Common Issues

**"command not found: nixai"**
- Ensure installation completed successfully
- Check your PATH includes the Nix profile directory
- Try `which nixai` to locate the binary

**"flakes not enabled"**
- Add `experimental-features = nix-command flakes` to your Nix configuration
- Restart your shell after configuration changes

**"Failed to connect to Ollama"**
- Install Ollama: `curl -fsSL https://ollama.com/install.sh | sh`
- Start Ollama service: `ollama serve`
- Pull a model: `ollama pull llama3`

### Getting Help

- View built-in help: `nixai --help`
- Check the [User Manual](docs/MANUAL.md)
- Visit the [GitHub repository](https://github.com/olafkfreund/nix-ai-help)
- Open an issue for bugs or feature requests

---

Choose the installation method that best fits your Nix setup and preferences. The flake-based methods are recommended for modern Nix users, while the non-flake methods provide compatibility with traditional Nix workflows.
