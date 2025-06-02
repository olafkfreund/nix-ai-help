# Adding nixai to Your flake.nix - Copy & Paste Examples

This document provides ready-to-use examples for adding nixai to your flake.nix file.

## Basic flake.nix Structure

Here's what your flake.nix should look like with nixai added:

```nix
{
  description = "My NixOS configuration with nixai";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    # Add nixai input
    nixai.url = "github:olafkfreund/nix-ai-help";
  };

  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    # Your configuration here - see examples below
  };
}
```

## Example 1: NixOS Only (System-wide)

```nix
{
  description = "NixOS with nixai system-wide";

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
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              aiProvider = "ollama";  # Change to "openai" or "gemini" if preferred
              aiModel = "llama3";
            };
          };
        }
      ];
    };
  };
}
```

## Example 2: Home Manager Only (Per-user)

```nix
{
  description = "Home Manager with nixai per-user";

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
          services.nixai = {
            enable = true;
            mcp.enable = true;
            vscodeIntegration = true;      # Auto-configure VS Code
            neovimIntegration.enable = true; # Auto-configure Neovim
          };
        }
      ];
    };
  };
}
```

## Example 3: Combined NixOS + Home Manager

```nix
{
  description = "Complete setup with NixOS and Home Manager";

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
          # System-wide nixai
          services.nixai = {
            enable = true;
            mcp.enable = true;
          };

          # Home Manager configuration
          home-manager.useGlobalPkgs = true;
          home-manager.useUserPackages = true;
          home-manager.users.yourusername = {
            imports = [ nixai.homeManagerModules.default ];
            
            services.nixai = {
              enable = true;
              mcp.enable = true;
              vscodeIntegration = true;
              neovimIntegration.enable = true;
            };
          };
        }
      ];
    };
  };
}
```

## Step-by-Step Instructions

### 1. Add nixai Input
In your `flake.nix`, add to the `inputs` section:
```nix
nixai.url = "github:olafkfreund/nix-ai-help";
```

### 2. Add to Function Parameters
Update your `outputs` function to include `nixai`:
```nix
outputs = { self, nixpkgs, home-manager, nixai, ... }: {
```

### 3. Import Module
Choose one:
- **NixOS**: Add `nixai.nixosModules.default` to your modules list
- **Home Manager**: Add `nixai.homeManagerModules.default` to your modules list

### 4. Enable Service
Add the configuration:
```nix
services.nixai = {
  enable = true;
  mcp.enable = true;  # Enables advanced features
};
```

### 5. Rebuild
```bash
# For NixOS
sudo nixos-rebuild switch --flake .

# For Home Manager
home-manager switch --flake .
```

## AI Provider Setup

### Ollama (Recommended - Local/Private)
```bash
# Install Ollama
nix-shell -p ollama

# Pull model
ollama pull llama3

# Start service (if not using NixOS module)
ollama serve
```

### OpenAI
Set environment variable:
```bash
export OPENAI_API_KEY="your-api-key-here"
```

### Google Gemini
Set environment variable:
```bash
export GEMINI_API_KEY="your-api-key-here"
```

## Common Configuration Options

```nix
services.nixai = {
  enable = true;
  
  mcp = {
    enable = true;
    host = "localhost";
    port = 8080;  # Use 8081 for Home Manager to avoid conflicts
    aiProvider = "ollama";  # "ollama", "openai", or "gemini"
    aiModel = "llama3";     # Model varies by provider
    
    # Add custom documentation sources
    documentationSources = [
      "https://wiki.nixos.org/wiki/NixOS_Wiki"
      "https://nix.dev/manual/nix"
      "https://nixos.org/manual/nixpkgs/stable/"
      "https://your-company.com/nixos-docs"  # Custom source
    ];
  };
  
  # Home Manager only - editor integrations
  vscodeIntegration = true;
  neovimIntegration = {
    enable = true;
    keybindings = {
      askNixai = "<leader>na";
      askNixaiVisual = "<leader>na";
    };
  };
};
```

## Usage After Installation

```bash
# Direct questions
nixai "how do I enable SSH?"

# Interactive mode
nixai

# Specific commands
nixai health
nixai service-examples nginx
nixai find-option "enable firewall"
```

## Troubleshooting

If you get "package not found" errors:
```bash
nix flake lock --update-input nixai
```

## Need More Help?

- 📚 [Complete Flake Integration Guide](FLAKE_INTEGRATION_GUIDE.md)
- 📖 [User Manual](MANUAL.md)
- 🔧 [VS Code Integration](MCP_VSCODE_INTEGRATION.md)

---

Copy and paste the example that matches your setup, replace `yourhostname` and `yourusername` with your actual values, and you're ready to go! 🚀
