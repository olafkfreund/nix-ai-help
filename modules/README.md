# nixai NixOS and Home Manager Modules

This directory contains NixOS and Home Manager modules for integrating nixai into your configuration.

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
