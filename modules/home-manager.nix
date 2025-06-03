# Home Manager module for nixai
{nixaiPackage ? null}:
# Accept optional nixai package parameter
{
  config,
  lib,
  pkgs,
  ...
}:
with lib; let
  cfg = config.services.nixai;

  # Use provided package or try to find nixai in pkgs, fallback to placeholder
  defaultNixaiPackage =
    if nixaiPackage != null
    then nixaiPackage
    else if pkgs ? nixai
    then pkgs.nixai
    else
      pkgs.stdenv.mkDerivation {
        pname = "nixai-placeholder";
        version = "0.0.0";
        src = pkgs.writeText "placeholder" "";
        dontUnpack = true;
        installPhase = ''
                  mkdir -p $out/bin
                  cat > $out/bin/nixai << 'EOF'
          #!/bin/sh
          echo "nixai placeholder: Please install nixai package or build from flake"
          echo "Try: nix run github:username/nixai -- \"\$@\""
          exit 1
          EOF
                  chmod +x $out/bin/nixai
        '';
        meta.description = "Placeholder for nixai when package is not available";
      };
in {
  options.services.nixai = {
    enable = mkEnableOption "nixai service";

    mcp = {
      enable = mkEnableOption "nixai MCP server";

      package = mkOption {
        type = types.package;
        default = defaultNixaiPackage;
        defaultText = literalExpression "pkgs.nixai";
        description = "The nixai package to use. Defaults to a placeholder when nixai package is not available.";
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
        description = "Default AI provider to use (ollama, gemini, openai)";
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
    };

    vscodeIntegration = mkEnableOption "Enable VS Code MCP integration";

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

        Service = {
          ExecStart = "${cfg.mcp.package}/bin/nixai mcp-server start --socket-path=${cfg.mcp.socketPath} ${concatStringsSep " " cfg.mcp.extraFlags}";
          Environment = cfg.mcp.environment;
          Restart = "on-failure";
          RestartSec = "5s";
        };

        Install = {
          WantedBy = ["graphical-session.target"];
        };
      };
    })

    (mkIf cfg.vscodeIntegration {
      programs.vscode.extensions = mkIf config.programs.vscode.enable [
        {
          name = "vscode-nixai";
          publisher = "nixos";
          # This is a placeholder - extension details would need to be filled in
          # once the VS Code extension is published
          version = "latest";
        }
      ];

      programs.vscode.userSettings = mkIf config.programs.vscode.enable {
        "automata.mcp.enabled" = true;
        "zebradev.mcp.enabled" = true;
        "nixai.socket-path" = cfg.mcp.socketPath;
      };
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
          -- nixai integration for regular Neovim
          local function nixai_query(question)
            if not question or question == "" then
              question = vim.fn.input("Ask nixai: ")
            end

            if question == "" then
              return
            end

            local cmd = string.format("${cfg.mcp.package}/bin/nixai --ask \"%s\"", question:gsub('"', '\\"'))
            local output = vim.fn.system(cmd)

            -- Create response buffer
            local buf = vim.api.nvim_create_buf(false, true)
            local lines = vim.split(output, "\n")
            vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
            vim.api.nvim_buf_set_option(buf, "filetype", "markdown")
            vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
            vim.api.nvim_buf_set_name(buf, "nixai-response")

            -- Open in split
            vim.cmd("split")
            vim.api.nvim_set_current_buf(buf)

            -- Add quit mapping
            vim.keymap.set("n", "q", ":close<CR>", { buffer = buf, silent = true })
          end

          -- Set up keymaps
          vim.keymap.set("n", "${cfg.neovimIntegration.keybindings.askNixai}", nixai_query, { desc = "Ask nixai" })
          vim.keymap.set("v", "${cfg.neovimIntegration.keybindings.askNixaiVisual}", function()
            local start_pos = vim.fn.getpos("'<")
            local end_pos = vim.fn.getpos("'>")
            local lines = vim.fn.getline(start_pos[2], end_pos[2])
            local text = table.concat(lines, "\n")
            nixai_query("Explain this code: " .. text)
          end, { desc = "Ask nixai about selection" })

          print("nixai integration loaded! Use ${cfg.neovimIntegration.keybindings.askNixai} to ask questions")
        '';
      };
    })
  ];

  meta = {
    maintainers = [lib.maintainers.olf];
    description = lib.mdDoc ''
      NixAI Home Manager module. Provides options to enable the NixAI MCP server, configure AI provider/model, and set documentation sources.

      Example usage:
      ```nix
      services.nixai = {
        enable = true;
        mcp.enable = true;
        mcp.aiProvider = "ollama";
        mcp.aiModel = "llama3";
        mcp.documentationSources = [ "https://wiki.nixos.org/wiki/NixOS_Wiki" ];
      };
      ```
    '';
    doc = ./home-manager.nix;
  };
}
