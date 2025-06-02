# 🚀 nixai Flake Quick Reference

This is a quick reference for integrating nixai into your Nix flakes. For comprehensive documentation, see the [complete Flake Integration Guide](FLAKE_INTEGRATION_GUIDE.md).

## Quick Start

### 1. Add to flake inputs

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };
}
```

### 2. Use directly (no configuration)

```bash
# Run directly
nix run github:olafkfreund/nix-ai-help -- "how do I enable SSH?"

# Install to profile
nix profile install github:olafkfreund/nix-ai-help
```

## NixOS Integration

```nix
# In your NixOS configuration
nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
  modules = [
    nixai.nixosModules.default
    {
      services.nixai = {
        enable = true;
        mcp = {
          enable = true;
          aiProvider = "ollama";  # or "openai", "gemini"
          aiModel = "llama3";
        };
      };
    }
  ];
};
```

## Home Manager Integration

```nix
# In your Home Manager configuration
homeConfigurations.yourusername = home-manager.lib.homeManagerConfiguration {
  modules = [
    nixai.homeManagerModules.default
    {
      services.nixai = {
        enable = true;
        mcp.enable = true;
        vscodeIntegration = true;      # Auto-configure VS Code
        neovimIntegration.enable = true; # Auto-configure Neovim
      };
    }
  ];
};
```

## Combined Setup (Both NixOS + Home Manager)

```nix
{
  nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
    modules = [
      nixai.nixosModules.default
      home-manager.nixosModules.home-manager
      {
        # System-wide nixai
        services.nixai.enable = true;
        
        # Home Manager per-user
        home-manager.users.myuser = {
          imports = [ nixai.homeManagerModules.default ];
          services.nixai = {
            enable = true;
            neovimIntegration.enable = true;
          };
        };
      }
    ];
  };
}
```

## AI Provider Setup

### Ollama (Local/Private - Recommended)
```nix
services.nixai.mcp = {
  aiProvider = "ollama";
  aiModel = "llama3";
};
```

Set up Ollama:
```bash
# Install and pull model
nix-shell -p ollama
ollama pull llama3
ollama serve
```

### OpenAI
```nix
services.nixai.mcp = {
  aiProvider = "openai";
  aiModel = "gpt-4";
};
```

```bash
export OPENAI_API_KEY="your-key-here"
```

### Google Gemini
```nix
services.nixai.mcp = {
  aiProvider = "gemini";
  aiModel = "gemini-pro";
};
```

```bash
export GEMINI_API_KEY="your-key-here"
```

## Common Configuration Options

```nix
services.nixai = {
  enable = true;
  
  mcp = {
    enable = true;
    host = "localhost";
    port = 8080;  # 8081 for Home Manager
    aiProvider = "ollama";
    aiModel = "llama3";
    
    # Custom documentation sources
    documentationSources = [
      "https://wiki.nixos.org/wiki/NixOS_Wiki"
      "https://nix.dev/manual/nix"
      "https://your-company.com/docs"  # Add custom sources
    ];
  };
  
  # Editor integrations (Home Manager only)
  vscodeIntegration = true;
  neovimIntegration = {
    enable = true;
    keybindings = {
      askNixai = "<leader>na";
      askNixaiVisual = "<leader>na";
    };
  };
  
  # Additional config
  config = {
    debug_mode = false;
    log_level = "info";
  };
};
```

## Usage After Installation

```bash
# Direct questions
nixai "how do I enable SSH?"
nixai --ask "configure NVIDIA drivers"

# Interactive mode
nixai

# Health check
nixai health

# Specific commands
nixai service-examples nginx
nixai find-option "enable firewall"
nixai hardware detect
nixai gc analyze
nixai templates list
```

## Troubleshooting

### Package not found
```bash
nix flake lock --update-input nixai
```

### MCP server issues
```bash
# Check status
sudo systemctl status nixai-mcp        # NixOS
systemctl --user status nixai-mcp      # Home Manager

# Check logs
sudo journalctl -u nixai-mcp -f        # NixOS
journalctl --user -u nixai-mcp -f      # Home Manager
```

### Ollama not responding
```bash
# Check Ollama
systemctl status ollama
curl http://localhost:11434/api/tags
```

## Need More Help?

- 📚 [Complete Flake Integration Guide](FLAKE_INTEGRATION_GUIDE.md) - Comprehensive setup and configuration
- 📖 [User Manual](MANUAL.md) - Full feature documentation
- 🔧 [VS Code Integration](MCP_VSCODE_INTEGRATION.md) - VS Code setup guide
- 🟢 [Neovim Integration](neovim-integration.md) - Neovim setup guide
- 🛠️ [Main README](../README.md) - Project overview and features

---

Happy NixOS configuring with AI assistance! 🚀
