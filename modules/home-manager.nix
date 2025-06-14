# nixai Home Manager Module
# Provides user-level nixai configuration and services for Home Manager.
# This module enables per-user nixai installation with optional editor integrations.
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
        default =
          if (pkgs ? nixai)
          then pkgs.nixai
          else pkgs.callPackage ../package.nix {inherit (pkgs) lib buildGoModule;};
        description = "The nixai package to use";
      };

      socketPath = mkOption {
        type = types.str;
        default = "$HOME/.local/share/nixai/mcp.sock";
        description = "Path to the MCP server Unix socket";
        example = "$HOME/.local/share/nixai/mcp.sock";
      };

      host = mkOption {
        type = types.str;
        default = "localhost";
        description = "Host for the MCP HTTP server to listen on";
        example = "localhost";
      };

      port = mkOption {
        type = types.port;
        default = 8081;
        description = "Port for the MCP HTTP server to listen on";
        example = 8081;
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
        example = ["https://wiki.nixos.org/wiki/NixOS_Wiki"];
      };

      aiProvider = mkOption {
        type = types.str;
        default = "ollama";
        description = "Default AI provider to use (ollama, claude, groq, gemini, openai, llamacpp, custom)";
        example = "ollama";
      };

      aiModel = mkOption {
        type = types.str;
        default = "llama3";
        description = "Default AI model to use for the specified provider";
        example = "llama3";
      };

      extraFlags = mkOption {
        type = types.listOf types.str;
        default = [];
        description = "Extra flags to pass to the MCP server";
        example = ["--log-level=debug"];
      };

      environment = mkOption {
        type = types.attrsOf types.str;
        default = {};
        description = "Extra environment variables for the MCP server";
        example = {NIXAI_LOG_LEVEL = "debug";};
      };

      endpoints = mkOption {
        type = types.listOf (types.submodule ({...}: {
          options = {
            name = mkOption {
              type = types.str;
              description = "Name for this MCP server endpoint (e.g. 'default', 'prod', 'test')";
            };
            socketPath = mkOption {
              type = types.str;
              description = "Path to the MCP server Unix socket for this endpoint";
              example = "$HOME/.local/share/nixai/mcp.sock";
            };
            host = mkOption {
              type = types.str;
              default = "localhost";
              description = "Host for the MCP HTTP server to listen on for this endpoint";
            };
            port = mkOption {
              type = types.port;
              default = 8081;
              description = "Port for the MCP HTTP server to listen on for this endpoint";
            };
          };
        }));
        default = [];
        description = "List of additional/custom MCP server endpoints (for multi-server or custom setups).";
        example = [
          {
            name = "default";
            socketPath = "$HOME/.local/share/nixai/mcp.sock";
            host = "localhost";
            port = 8081;
          }
          {
            name = "test";
            socketPath = "/tmp/nixai-test.sock";
            host = "localhost";
            port = 8082;
          }
        ];
      };
    };

    vscodeIntegration = {
      enable = mkEnableOption "Enable VS Code MCP integration";

      contextAware = mkOption {
        type = types.bool;
        default = true;
        description = "Enable context-aware AI assistance in VS Code";
      };

      autoRefreshContext = mkOption {
        type = types.bool;
        default = true;
        description = "Automatically refresh NixOS context when files change";
      };

      contextTimeout = mkOption {
        type = types.int;
        default = 5000;
        description = "Timeout for context operations in milliseconds";
      };

      enabledExtensions = mkOption {
        type = types.listOf types.str;
        default = [
          "automatalabs.copilot-mcp"
          "zebradev.mcp-server-runner"
          "saoudrizwan.claude-dev"
        ];
        description = "List of VS Code extensions to configure for MCP integration";
      };
    };

    neovimIntegration = {
      enable = mkEnableOption "Enable Neovim integration with nixai";

      useNixVim = mkOption {
        type = types.bool;
        default = true;
        description = "Use NixVim for Neovim configuration with nixai integration";
      };

      keybindings = mkOption {
        type = types.attrsOf types.str;
        default = {
          askNixai = "<leader>na";
          askNixaiVisual = "<leader>na";
          startMcpServer = "<leader>ns";
        };
        description = "Keybindings for nixai integration in Neovim";
      };

      autoStartMcp = mkOption {
        type = types.bool;
        default = true;
        description = "Automatically start MCP server when Neovim loads nixai integration";
      };
    };
  };

  config = mkMerge [
    (mkIf cfg.enable {
      home.packages = [cfg.mcp.package];

      xdg.configFile."nixai/config.yaml".text = builtins.toJSON {
        ai_provider = cfg.mcp.aiProvider;
        ai_model = cfg.mcp.aiModel;
        log_level = "info";
        mcp_server = {
          host = cfg.mcp.host;
          port = cfg.mcp.port;
          socket_path = cfg.mcp.socketPath;
          auto_start = cfg.mcp.enable;
          documentation_sources = cfg.mcp.documentationSources;
          extra_flags = cfg.mcp.extraFlags;
          environment = cfg.mcp.environment;
          endpoints = cfg.mcp.endpoints;
        };
      };
    })

    (mkIf cfg.mcp.enable {
      systemd.user.services.nixai-mcp = {
        Unit = {
          Description = "nixai MCP Server";
          After = "network.target";
          PartOf = "graphical-session.target";
        };

        Service = let
          envList = lib.attrsets.mapAttrsToList (k: v: "${k}=${v}") cfg.mcp.environment;
        in
          lib.mkMerge [
            {
              ExecStart = "${cfg.mcp.package}/bin/nixai mcp-server start --socket-path=${cfg.mcp.socketPath} ${concatStringsSep " " cfg.mcp.extraFlags}";
              Restart = "on-failure";
              RestartSec = "5s";
            }
            (lib.mkIf (envList != []) {
              Environment = envList;
            })
          ];

        Install = {
          WantedBy = ["graphical-session.target"];
        };
      };
    })

    (mkIf cfg.vscodeIntegration.enable {
      # VS Code settings for MCP integration with context-aware features
      programs.vscode.userSettings = mkIf config.programs.vscode.enable {
        "mcp.servers" = {
          "nixai" = {
            "command" = "bash";
            "args" = ["-c" "socat STDIO UNIX-CONNECT:${cfg.mcp.socketPath}"];
            "env" = {};
            "capabilities" = {
              "context" = cfg.vscodeIntegration.contextAware;
              "system_detection" = true;
            };
          };
        };

        "copilot.mcp.servers" = {
          "nixai" = {
            "command" = "bash";
            "args" = ["-c" "socat STDIO UNIX-CONNECT:${cfg.mcp.socketPath}"];
            "env" = {};
            "contextAware" = cfg.vscodeIntegration.contextAware;
          };
        };

        "claude-dev.mcpServers" = {
          "nixai" = {
            "command" = "bash";
            "args" = ["-c" "socat STDIO UNIX-CONNECT:${cfg.mcp.socketPath}"];
            "env" = {};
            "useContext" = cfg.vscodeIntegration.contextAware;
          };
        };

        "mcp.enableDebug" = true;
        "claude-dev.enableMcp" = true;
        "automata.mcp.enabled" = true;
        "zebradev.mcp.enabled" = true;

        # Context-aware AI settings
        "nixai.contextIntegration" = {
          "autoRefresh" = cfg.vscodeIntegration.autoRefreshContext;
          "contextTimeout" = cfg.vscodeIntegration.contextTimeout;
          "enableDetailedContext" = false;
        };
      };

      # Optional: Install recommended VS Code extensions
      programs.vscode.extensions = mkIf config.programs.vscode.enable (
        map (ext: pkgs.vscode-extensions.${ext} or null)
        (filter (ext: ext != null) cfg.vscodeIntegration.enabledExtensions)
      );
    })

    (mkIf cfg.neovimIntegration.enable {
      programs.neovim = {
        enable = true;
        defaultEditor = true;
        viAlias = true;
        vimAlias = true;

        extraConfig = ''
          " Basic Neovim configuration
          set number relativenumber
          set expandtab tabstop=2 shiftwidth=2
          set hidden
          set ignorecase smartcase
          set termguicolors

          " Set leader key
          let mapleader = " "
        '';

        extraLuaConfig = ''
          -- Load nixai-nvim.lua integration module
          vim.g.nixai_endpoints = vim.fn.json_decode([[${builtins.toJSON cfg.mcp.endpoints}]])
          vim.g.nixai_socket_path = "${cfg.mcp.socketPath}"
          dofile("${pkgs.writeTextFile {
            name = "nixai-nvim.lua";
            text = builtins.readFile ../../modules/nixai-nvim.lua;
          }}")
          require("nixai-nvim").setup_keymaps()
        '';
      };
    })
  ];
}
