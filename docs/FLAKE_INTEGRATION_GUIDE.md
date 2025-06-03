# 🚀 nixai Flake Integration Guide

This comprehensive guide shows you how to integrate **nixai** into your NixOS and Home Manager configurations using flakes, with all available features and options.

## Table of Contents

- [Quick Start](#quick-start)
- [NixOS System Integration](#nixos-system-integration)
- [Home Manager Integration](#home-manager-integration)
- [Combined NixOS + Home Manager Setup](#combined-nixos--home-manager-setup)
- [Configuration Options](#configuration-options)
- [Advanced Features](#advanced-features)
- [Example: Nixvim + Home Manager + nixai Neovim Integration](#example-nixvim--home-manager--nixai-neovim-integration)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### 1. Add nixai to your flake inputs

Add nixai to your `flake.nix` inputs section:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };
  
  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    # Your configuration here
  };
}
```

### 2. Basic Installation Options

You have several options for installing nixai:

#### Option A: Just the Package (Minimal)
```bash
# Run directly without installation
nix run github:olafkfreund/nix-ai-help -- "how do I enable SSH?"

# Install to user profile
nix profile install github:olafkfreund/nix-ai-help
```

#### Option B: NixOS System-wide Integration
```nix
nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
  modules = [
    nixai.nixosModules.default
    {
      services.nixai.enable = true;
    }
  ];
};
```

#### Option C: Home Manager Integration
```nix
homeConfigurations.yourusername = home-manager.lib.homeManagerConfiguration {
  modules = [
    nixai.homeManagerModules.default
    {
      services.nixai.enable = true;
    }
  ];
};
```

---

## NixOS System Integration

### Basic NixOS Configuration

Add nixai to your NixOS system configuration:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, nixai, ... }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./configuration.nix
        nixai.nixosModules.default
        {
          # Basic nixai configuration
          services.nixai = {
            enable = true;
            
            # Enable MCP server for advanced features
            mcp = {
              enable = true;
              port = 8080;
              aiProvider = "ollama";  # or "gemini", "openai"
              aiModel = "llama3";
            };
          };
        }
      ];
    };
  };
}
```

### Advanced NixOS Configuration

For a complete setup with all features enabled:

```nix
{
  services.nixai = {
    enable = true;
    
    mcp = {
      enable = true;
      
      # Package configuration
      package = nixai.packages.${system}.nixai;
      
      # Network configuration
      host = "localhost";
      port = 8080;
      socketPath = "/run/nixai/mcp.sock";
      
      # AI Provider settings
      aiProvider = "ollama";  # Options: ollama, gemini, openai
      aiModel = "llama3";     # Model depends on provider
      
      # Documentation sources for MCP server
      documentationSources = [
        "https://wiki.nixos.org/wiki/NixOS_Wiki"
        "https://nix.dev/manual/nix"
        "https://nixos.org/manual/nixpkgs/stable/"
        "https://nix.dev/manual/nix/2.28/language/"
        "https://nix-community.github.io/home-manager/"
      ];
    };
    
    # Additional configuration
    config = {
      # Add any additional YAML configuration here
      debug_mode = false;
      cache_dir = "/var/cache/nixai";
    };
  };
}
```

---

## Home Manager Integration

### Basic Home Manager Configuration

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    homeConfigurations.yourusername = home-manager.lib.homeManagerConfiguration {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      modules = [
        nixai.homeManagerModules.default
        {
          # Basic nixai configuration
          services.nixai = {
            enable = true;
            
            mcp = {
              enable = true;
              port = 8081;  # Different port to avoid conflicts
            };
          };
        }
      ];
    };
  };
}
```

### Advanced Home Manager Configuration

Complete setup with editor integrations:

```nix
{
  services.nixai = {
    enable = true;
    
    # MCP Server configuration
    mcp = {
      enable = true;
      package = nixai.packages.${pkgs.system}.nixai;
      
      # User-specific paths
      socketPath = "$HOME/.local/share/nixai/mcp.sock";
      host = "localhost";
      port = 8081;
      
      # AI settings
      aiProvider = "ollama";
      aiModel = "llama3";
      
      documentationSources = [
        "https://wiki.nixos.org/wiki/NixOS_Wiki"
        "https://nix.dev/manual/nix"
        "https://nixos.org/manual/nixpkgs/stable/"
        "https://nix.dev/manual/nix/2.28/language/"
        "https://nix-community.github.io/home-manager/"
      ];
    };
    
    # VS Code integration
    vscodeIntegration = true;
    
    # Neovim integration
    neovimIntegration = {
      enable = true;
      useNixVim = true;
      autoStartMcp = true;
      
      keybindings = {
        askNixai = "<leader>na";
        askNixaiVisual = "<leader>na";
        startMcpServer = "<leader>ns";
      };
    };
  };
}
```

---

## Combined NixOS + Home Manager Setup

For the most comprehensive setup, use both NixOS and Home Manager modules:

### flake.nix
```nix
{
  description = "My NixOS configuration with nixai";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./configuration.nix
        nixai.nixosModules.default
        home-manager.nixosModules.home-manager
        {
          home-manager.useGlobalPkgs = true;
          home-manager.useUserPackages = true;
          home-manager.users.yourusername = {
            imports = [ nixai.homeManagerModules.default ];
            
            # Home Manager nixai configuration
            services.nixai = {
              enable = true;
              mcp.enable = true;
              neovimIntegration.enable = true;
              vscodeIntegration = true;
            };
          };
          
          # NixOS nixai configuration
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

---

## Configuration Options

### AI Provider Options

#### Ollama (Local, Private)
```nix
services.nixai.mcp = {
  aiProvider = "ollama";
  aiModel = "llama3";        # or "llama3.1", "codellama", "mistral", etc.
};
```

Set up Ollama:
```bash
# Install Ollama
nix-shell -p ollama

# Pull a model
ollama pull llama3

# Start Ollama service (if not using NixOS service)
ollama serve
```

#### OpenAI
```nix
services.nixai.mcp = {
  aiProvider = "openai";
  aiModel = "gpt-4";         # or "gpt-3.5-turbo", "gpt-4-turbo", etc.
};
```

Set environment variable:
```bash
export OPENAI_API_KEY="your-api-key-here"
```

#### Google Gemini
```nix
services.nixai.mcp = {
  aiProvider = "gemini";
  aiModel = "gemini-pro";    # or "gemini-pro-vision"
};
```

Set environment variable:
```bash
export GEMINI_API_KEY="your-api-key-here"
```

### MCP Server Options

```nix
services.nixai.mcp = {
  enable = true;
  
  # Network configuration
  host = "localhost";        # Host to bind to
  port = 8080;              # HTTP port (NixOS: 8080, Home Manager: 8081)
  socketPath = "/path/to/socket";  # Unix socket path
  
  # Documentation sources
  documentationSources = [
    "https://wiki.nixos.org/wiki/NixOS_Wiki"
    "https://nix.dev/manual/nix"
    "https://nixos.org/manual/nixpkgs/stable/"
    # Add custom sources here
  ];
};
```

### Editor Integration Options

#### VS Code Integration
```nix
services.nixai.vscodeIntegration = true;
```

This automatically configures:
- MCP extension settings
- Socket path configuration
- Auto-enable MCP features

#### Neovim Integration
```nix
services.nixai.neovimIntegration = {
  enable = true;
  useNixVim = true;          # Use NixVim configuration
  autoStartMcp = true;       # Auto-start MCP server
  
  keybindings = {
    askNixai = "<leader>na";           # Ask nixai in normal mode
    askNixaiVisual = "<leader>na";     # Ask about selection in visual mode
    startMcpServer = "<leader>ns";     # Start MCP server
  };
};
```

---

## Advanced Features

### 1. Custom Configuration

Add custom YAML configuration:

```nix
services.nixai.config = {
  debug_mode = true;
  cache_dir = "/custom/cache/path";
  timeout = 30;
  max_retries = 3;
  
  # Custom AI settings
  ai_settings = {
    temperature = 0.7;
    max_tokens = 2048;
  };
};
```

### 2. Multiple AI Providers

You can set up multiple configurations for different use cases:

```nix
# System-wide with Ollama for privacy
services.nixai.mcp = {
  aiProvider = "ollama";
  aiModel = "llama3";
};

# User-specific with OpenAI for advanced features
home-manager.users.yourusername.services.nixai.mcp = {
  aiProvider = "openai";
  aiModel = "gpt-4";
  port = 8082;  # Different port
};
```

### 3. Custom Documentation Sources

Add your own documentation sources:

```nix
services.nixai.mcp.documentationSources = [
  # Standard sources
  "https://wiki.nixos.org/wiki/NixOS_Wiki"
  "https://nix.dev/manual/nix"
  
  # Custom sources
  "https://your-company.com/nixos-docs"
  "file:///path/to/local/docs"
];
```

### 4. Security Hardening

For production environments:

```nix
services.nixai = {
  enable = true;
  mcp = {
    enable = true;
    host = "127.0.0.1";  # Bind only to localhost
    port = 8080;
    
    # Use secure socket path
    socketPath = "/var/lib/nixai/secure.sock";
  };
  
  config = {
    # Enable audit logging
    audit_log = true;
    log_level = "info";
    
    # Security settings
    max_request_size = "1MB";
    rate_limit = {
      requests_per_minute = 60;
      burst = 10;
    };
  };
};
```

---

## Example: Nixvim + Home Manager + nixai Neovim Integration

This example shows how to use `nixvim` with Home Manager to install `nixai` and enable the Neovim integration, including keybindings and MCP server autostart.

Add this to your `home.nix`:

```nix
{ config, pkgs, ... }:
let
  nixai-flake = builtins.getFlake "github:olafkfreund/nix-ai-help";
in {
  imports = [ nixai-flake.homeManagerModules.default ];

  # Enable nixai and Neovim integration
  services.nixai = {
    enable = true;
    mcp.enable = true;
    neovimIntegration = {
      enable = true;
      useNixVim = true;
      keybindings = {
        askNixai = "<leader>na";
        askNixaiVisual = "<leader>na";
        startMcpServer = "<leader>ns";
      };
      autoStartMcp = true;
    };
  };

  # Nixvim configuration (minimal example)
  programs.nixvim = {
    enable = true;
    extraConfigVim = ''
      set number
      set relativenumber
      set expandtab
      set shiftwidth=2
      set tabstop=2
    '';
    # Optionally add plugins, LSP, etc.
  };
}
```

**Notes:**

- This will install `nixai`, enable the MCP server, and set up Neovim keybindings for asking questions and starting the server.
- You can customize the keybindings and other options as needed.
- Make sure you are using a recent version of Home Manager and Nixvim.
- If you encounter issues, check the logs with `journalctl --user -u nixai-mcp` and ensure the MCP server is running.

---

## Troubleshooting

### Common Issues

#### 1. Package Not Found
```bash
error: attribute 'nixai' missing
```

**Solution**: Update your flake lock:
```bash
nix flake lock --update-input nixai
```

#### 2. MCP Server Won't Start
Check the service status:
```bash
# NixOS
sudo systemctl status nixai-mcp

# Home Manager
systemctl --user status nixai-mcp
```

Check logs:
```bash
# NixOS
sudo journalctl -u nixai-mcp -f

# Home Manager
journalctl --user -u nixai-mcp -f
```

#### 3. Port Conflicts
If you get port binding errors, change the port:
```nix
services.nixai.mcp.port = 8082;  # Use different port
```

#### 4. AI Provider Issues

**Ollama not responding**:
```bash
# Check if Ollama is running
systemctl status ollama

# Test connection
curl http://localhost:11434/api/tags
```

**API Key issues** (OpenAI/Gemini):
```bash
# Check environment variables
echo $OPENAI_API_KEY
echo $GEMINI_API_KEY
```

#### 5. Permission Issues
```bash
# Fix socket permissions
sudo chown -R yourusername:yourusername ~/.local/share/nixai/
```

### Getting Help

1. **Check logs**: Always check systemd service logs first
2. **Test manually**: Try running nixai commands directly
3. **Verify configuration**: Use `nixai health` to check system status
4. **Community support**: Join NixOS community channels for help

### Debug Mode

Enable debug mode for verbose logging:
```nix
services.nixai.config.debug_mode = true;
```

Or run with debug flag:
```bash
nixai --debug "your question"
```

---

## Example Complete Configuration

Here's a complete example combining all features:

### flake.nix
```nix
{
  description = "Complete nixai integration example";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./configuration.nix
        nixai.nixosModules.default
        home-manager.nixosModules.home-manager
        {
          # NixOS system-wide configuration
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              port = 8080;
              aiProvider = "ollama";
              aiModel = "llama3";
              documentationSources = [
                "https://wiki.nixos.org/wiki/NixOS_Wiki"
                "https://nix.dev/manual/nix"
                "https://nixos.org/manual/nixpkgs/stable/"
                "https://my-company.com/nixos-docs"  # Custom docs
              ];
            };
            config = {
              debug_mode = false;
              log_level = "info";
              cache_dir = "/var/cache/nixai";
            };
          };

          # Home Manager configuration
          home-manager.useGlobalPkgs = true;
          home-manager.useUserPackages = true;
          home-manager.users.myuser = {
            imports = [ nixai.homeManagerModules.default ];
            
            services.nixai = {
              enable = true;
              mcp = {
                enable = true;
                port = 8081;  # Different port for user
                aiProvider = "openai";  # Different provider for user
                aiModel = "gpt-4";
              };
              
              # Enable all integrations
              vscodeIntegration = true;
              neovimIntegration = {
                enable = true;
                useNixVim = true;
                autoStartMcp = true;
                keybindings = {
                  askNixai = "<leader>na";
                  askNixaiVisual = "<leader>nv";
                  startMcpServer = "<leader>ns";
                };
              };
            };
          };
        }
      ];
    };
  };
}
```

This configuration provides:
- ✅ System-wide nixai with Ollama (private)
- ✅ User-specific nixai with OpenAI (advanced features)
- ✅ VS Code integration
- ✅ Neovim integration with custom keybindings
- ✅ Custom documentation sources
- ✅ Comprehensive logging and debugging
- ✅ MCP server on both system and user level

---

## Next Steps

After setting up nixai:

1. **Test the installation**: Run `nixai health` to verify everything works
2. **Configure AI providers**: Set up your preferred AI provider (Ollama, OpenAI, or Gemini)
3. **Try the features**: Explore different nixai commands and capabilities
4. **Customize configuration**: Adjust settings for your specific needs
5. **Integrate with editors**: Set up VS Code or Neovim integration
6. **Join the community**: Get involved with nixai development and support

Happy NixOS configuring with AI assistance! 🚀
