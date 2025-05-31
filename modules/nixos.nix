# NixOS module for nixai
{
  config,
  lib,
  pkgs,
  ...
}:
with lib; let
  cfg = config.services.nixai;
in {
  options.services.nixai = {
    enable = mkEnableOption "nixai service";

    mcp = {
      enable = mkEnableOption "nixai MCP server";

      package = mkOption {
        type = types.package;
        default = pkgs.nixai;
        description = "The nixai package to use";
      };

      socketPath = mkOption {
        type = types.str;
        default = "/run/nixai/mcp.sock";
        description = "Path to the MCP server Unix socket";
      };

      host = mkOption {
        type = types.str;
        default = "localhost";
        description = "Host for the MCP HTTP server to listen on";
      };

      port = mkOption {
        type = types.port;
        default = 8080;
        description = "Port for the MCP HTTP server to listen on";
      };

      documentationSources = mkOption {
        type = types.listOf types.str;
        default = [
          "https://wiki.nixos.org/wiki/NixOS_Wiki"
          "https://nix.dev/manual/nix"
          "https://nixos.org/manual/nixpkgs/stable/"
          "https://nix.dev/manual/nix/2.28/language/"
          "https://nix-community.github.io/home-manager/"
        ];
        description = "Documentation sources for the MCP server to query";
      };

      aiProvider = mkOption {
        type = types.str;
        default = "ollama";
        description = "Default AI provider to use (ollama, gemini, openai)";
      };

      aiModel = mkOption {
        type = types.str;
        default = "llama3";
        description = "Default AI model to use for the specified provider";
      };
    };

    config = mkOption {
      type = types.attrs;
      default = {};
      description = "Additional configuration options for nixai";
    };
  };

  config = mkIf cfg.enable {
    # Common configuration for nixai
    environment.systemPackages = [cfg.mcp.package];

    # Configuration for the MCP server
    systemd.services.nixai-mcp = mkIf cfg.mcp.enable {
      description = "nixai MCP Server";
      wantedBy = ["multi-user.target"];
      after = ["network.target"];
      serviceConfig = {
        ExecStart = "${cfg.mcp.package}/bin/nixai mcp-server start --socket-path=${cfg.mcp.socketPath}";
        Restart = "on-failure";
        RestartSec = "5s";

        # Security hardening
        DynamicUser = true;
        RuntimeDirectory = "nixai";
        RuntimeDirectoryMode = "0755";
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        NoNewPrivileges = true;
      };
      environment = {
        NIXAI_SOCKET_PATH = cfg.mcp.socketPath;
      };
    };

    # Create default configuration file
    environment.etc."nixai/config.yaml" = {
      text =
        builtins.toJSON {
          ai_provider = cfg.mcp.aiProvider;
          ai_model = cfg.mcp.aiModel;
          log_level = "info";
          mcp_server = {
            host = cfg.mcp.host;
            port = cfg.mcp.port;
            socket_path = cfg.mcp.socketPath;
            auto_start = cfg.mcp.enable;
            documentation_sources = cfg.mcp.documentationSources;
          };
        }
        // cfg.config;
    };
  };
}
