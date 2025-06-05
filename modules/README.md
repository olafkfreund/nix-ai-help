# nixai NixOS and Home Manager Modules

This directory contains NixOS and Home Manager modules for integrating nixai into your configuration.

---

## 🚀 Quick Start: Flake-based Installation

**1. Add nixai as an input to your flake:**

```nix
inputs.nixai.url = "github:olafkfreund/nix-ai-help";
```

**2. Use the module in your NixOS or Home Manager configuration:**

### NixOS Example

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

### Home Manager Example

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

---

## 🛠️ Troubleshooting: "attribute 'nixai' missing"

If you see an error like:

```
error: attribute 'nixai' missing
```

**Solution:**
- Always import the nixai module from the flake (as shown above), not as a raw .nix file.
- The nixai package is only available as `pkgs.nixai` when you use the module from the flake.
- If you use overlays or custom pkgs, ensure you pass the nixai package from the flake outputs.

---

## 📚 More Guides

- [Flake Quick Reference](../docs/FLAKE_QUICK_REFERENCE.md) – Copy-paste snippets
- [Complete Flake Integration Guide](../docs/FLAKE_INTEGRATION_GUIDE.md) – All options explained

---

## ⚙️ Advanced: Manual Import (Not Recommended)

If you must use the module without flakes, you need to provide the nixai package yourself:

```nix
imports = [ ./path/to/nixai/modules/nixos.nix ];
services.nixai = {
  enable = true;
  mcp = {
    enable = true;
    package = <your-nixai-package>;
  };
};
```

But for best results, always use the flake-based approach above.

---

# Full Manual and Advanced Configuration

## NixOS Module

The NixOS module allows you to integrate nixai system-wide with proper service management.

### Basic Usage

Add the module to your NixOS configuration:

```nix
{ config, pkgs, ... }:

{
  imports = [ 
    # Path to the nixai module
    ./path/to/nixai/modules/nixos.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # All other settings are optional and have sensible defaults
    };
  };
}
```

### Advanced Configuration

Full configuration with all available options:

```nix
{ config, pkgs, ... }:

{
  imports = [ 
    ./path/to/nixai/modules/nixos.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      socketPath = "/run/nixai/mcp.sock";
      host = "localhost";
      port = 8080;
      documentationSources = [
        "https://wiki.nixos.org/wiki/NixOS_Wiki"
        "https://nix.dev/manual/nix"
        "https://nixos.org/manual/nixpkgs/stable/"
        "https://nix.dev/manual/nix/2.28/language/"
        "https://nix-community.github.io/home-manager/"
      ];
      aiProvider = "ollama";  # Options: "ollama", "gemini", "openai"
      aiModel = "llama3";
    };
    config = {
      # Additional configuration to merge into config.yaml
      # This is optional
    };
  };
}
```

## Home Manager Module

The Home Manager module allows you to integrate nixai at the user level.

### Basic Usage

Add the module to your Home Manager configuration:

```nix
{ config, pkgs, ... }:

{
  imports = [ 
    ./path/to/nixai/modules/home-manager.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # All other settings are optional and have sensible defaults
    };
  };
}
```

### Advanced Configuration

Full configuration with all available options:

```nix
{ config, pkgs, ... }:

{
  imports = [ 
    ./path/to/nixai/modules/home-manager.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      socketPath = "$HOME/.local/share/nixai/mcp.sock";
      host = "localhost";
      port = 8080;
      documentationSources = [
        "https://wiki.nixos.org/wiki/NixOS_Wiki"
        "https://nix.dev/manual/nix"
        "https://nixos.org/manual/nixpkgs/stable/"
        "https://nix.dev/manual/nix/2.28/language/"
        "https://nix-community.github.io/home-manager/"
      ];
      aiProvider = "ollama";  # Options: "ollama", "gemini", "openai"
      aiModel = "llama3";
    };
    vscodeIntegration = true;  # Enable VS Code integration
  };
}
```

## VS Code Integration

The Home Manager module includes VS Code integration that can be enabled with `vscodeIntegration = true`. This will:

1. Install the nixai VS Code extension (when available)
2. Configure the extension to use the specified socket path
3. Enable MCP protocol handlers for AI assistants

Note: This requires Home Manager's VS Code module to be enabled with `programs.vscode.enable = true`.

## Multi-Endpoint MCP Server Support

You can define multiple MCP endpoints for advanced workflows (e.g., dev/prod/test, remote/local):

```nix
services.nixai = {
  enable = true;
  mcp.enable = true;
  mcp.endpoints = [
    {
      name = "default";
      socketPath = "/run/nixai/mcp.sock";
      host = "localhost";
      port = 8080;
    }
    {
      name = "test";
      socketPath = "/tmp/nixai-test.sock";
      host = "localhost";
      port = 8082;
    }
  ];
};
```

- These endpoints are written to `config.yaml` and available to the CLI and Neovim integration.
- In Neovim, you get per-endpoint keymaps (e.g., `<leader>nd` for default, `<leader>nt` for test).
- The CLI and integrations will use the correct socket/host/port for each endpoint.

See also: [docs/neovim-integration.md](../docs/neovim-integration.md) for Neovim multi-endpoint usage.

---

## 📝 Usage Examples

- Run `nixai "your question"` or `nixai --ask "your question"` for direct AI help.
- See the [manual](../docs/MANUAL.md) for all commands and options.

---

## 🧩 Multi-Endpoint & VS Code/Neovim Integration

- See [docs/neovim-integration.md](../docs/neovim-integration.md) and [docs/MCP_VSCODE_INTEGRATION.md](../docs/MCP_VSCODE_INTEGRATION.md) for advanced workflows.

---

For any issues, see the troubleshooting section above or open an issue on GitHub.
