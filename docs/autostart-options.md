# Auto-Start Options for nixai MCP Server on NixOS

This document outlines various methods to configure the nixai MCP server to automatically start on NixOS systems.

## 1. Using systemd User Services

A systemd user service runs in the context of your user account and doesn't require root privileges.

### Configuration

Create a systemd user service file:

```nix
# In your home-manager or NixOS configuration
systemd.user.services.nixai-mcp = {
  description = "nixai Model Context Protocol Server";
  wantedBy = [ "default.target" ];
  serviceConfig = {
    ExecStart = "${pkgs.nixai}/bin/nixai mcp";
    Restart = "on-failure";
    RestartSec = 5;
  };
  environment = {
    # Configure environment variables if needed
    NIXAI_CONFIG = "$HOME/.config/nixai/config.yaml";
    # AI provider API keys if necessary
    # OPENAI_API_KEY = ""; # Use secrets management instead of hardcoding
  };
};
```

## 2. Using NixOS System Services

For system-wide availability, you can create a system service:

```nix
# In your NixOS configuration
systemd.services.nixai-mcp = {
  description = "nixai Model Context Protocol Server";
  after = [ "network.target" ];
  wantedBy = [ "multi-user.target" ];
  serviceConfig = {
    ExecStart = "${pkgs.nixai}/bin/nixai mcp";
    Restart = "on-failure";
    RestartSec = 5;
    User = "yourusername"; # Replace with appropriate user
    Group = "users";
    # Security hardening options
    CapabilityBoundingSet = "";
    LockPersonality = true;
    MemoryDenyWriteExecute = true;
    NoNewPrivileges = true;
    PrivateDevices = true;
    PrivateTmp = true;
    ProtectClock = true;
    ProtectControlGroups = true;
    ProtectHome = true;
    ProtectHostname = true;
    ProtectKernelLogs = true;
    ProtectKernelModules = true;
    ProtectKernelTunables = true;
    ProtectSystem = "strict";
    RemoveIPC = true;
    RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
    RestrictNamespaces = true;
    RestrictRealtime = true;
    RestrictSUIDSGID = true;
    SystemCallArchitectures = "native";
    SystemCallFilter = [ "@system-service" "~@resources" "~@privileged" ];
  };
  environment = {
    # Configure environment variables if needed
    NIXAI_CONFIG = "/etc/nixai/config.yaml";
  };
};
```

## 3. Using Home Manager's `launchd` or `systemd` Modules

For Home Manager users:

```nix
# In your Home Manager configuration
home.file.".config/nixai/config.yaml".source = ./path/to/your/config.yaml;

systemd.user.services.nixai-mcp = {
  Unit = {
    Description = "nixai Model Context Protocol Server";
    After = [ "network.target" ];
  };
  Service = {
    ExecStart = "${pkgs.nixai}/bin/nixai mcp";
    Restart = "on-failure";
    RestartSec = 5;
  };
  Install = {
    WantedBy = [ "default.target" ];
  };
};
```

## 4. Using Session Auto-Start (Desktop Environment)

### GNOME

Create a desktop entry file:

```nix
# In your home-manager configuration
home.file.".config/autostart/nixai-mcp.desktop".text = ''
[Desktop Entry]
Type=Application
Name=nixai MCP Server
Exec=${pkgs.nixai}/bin/nixai mcp
StartupNotify=false
Terminal=false
'';
```

### KDE Plasma

Similar approach using a desktop entry file:

```nix
# In your home-manager configuration
home.file.".config/autostart/nixai-mcp.desktop".text = ''
[Desktop Entry]
Type=Application
Name=nixai MCP Server
Exec=${pkgs.nixai}/bin/nixai mcp
StartupNotify=false
Terminal=false
X-KDE-AutostartScript=true
'';
```

## 5. Using a NixOS Module

Create a dedicated NixOS module for nixai:

```nix
# nixai.nix module
{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.nixai;
in {
  options.services.nixai = {
    enable = mkEnableOption "nixai service";
    
    user = mkOption {
      type = types.str;
      default = "nobody";
      description = "User to run nixai as";
    };
    
    group = mkOption {
      type = types.str;
      default = "nogroup";
      description = "Group to run nixai as";
    };
    
    configFile = mkOption {
      type = types.path;
      description = "Path to nixai configuration file";
    };
    
    port = mkOption {
      type = types.port;
      default = 3080;
      description = "Port for MCP server to listen on";
    };
  };

  config = mkIf cfg.enable {
    systemd.services.nixai-mcp = {
      description = "nixai Model Context Protocol Server";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];
      
      serviceConfig = {
        User = cfg.user;
        Group = cfg.group;
        ExecStart = "${pkgs.nixai}/bin/nixai mcp --config ${cfg.configFile} --port ${toString cfg.port}";
        Restart = "on-failure";
        
        # Security hardening
        CapabilityBoundingSet = "";
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectHome = true;
        ProtectSystem = "strict";
      };
    };
  };
}
```

Then in your configuration.nix:

```nix
{
  imports = [ ./path/to/nixai.nix ];
  
  services.nixai = {
    enable = true;
    user = "yourusername";
    group = "users";
    configFile = "/path/to/config.yaml";
    port = 3080; # Default port
  };
}
```

## Best Practices

1. **Security**: Avoid hardcoding API keys in NixOS configurations. Use environment variables or secret management solutions.
2. **Idempotency**: Ensure your service definitions are idempotent and work correctly with NixOS's activation model.
3. **Resource Limits**: Consider adding resource limits to your service to prevent excessive resource usage:
   ```nix
   serviceConfig = {
     # Add to your service config
     MemoryMax = "200M";
     CPUQuota = "20%";
   };
   ```
4. **Logging**: Configure logging appropriately:
   ```nix
   serviceConfig = {
     # Add to your service config
     StandardOutput = "journal";
     StandardError = "journal";
   };
   ```

## Troubleshooting

- Check service status: `systemctl --user status nixai-mcp` (for user services)
- View logs: `journalctl --user -u nixai-mcp` (for user services)
- For system services: `sudo systemctl status nixai-mcp` and `sudo journalctl -u nixai-mcp`
