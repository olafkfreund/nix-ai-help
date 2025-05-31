# Home Manager module for nixai
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
        default = pkgs.writeShellScript "nixai-placeholder" "echo 'nixai binary from local build should be used'";
        defaultText = literalExpression "pkgs.nixai";
        description = "The nixai package to use. Defaults to local build when nixai package is not available.";
      };

      socketPath = mkOption {
        type = types.str;
        default = "$HOME/.local/share/nixai/mcp.sock";
        description = "Path to the MCP server Unix socket";
      };

      host = mkOption {
        type = types.str;
        default = "localhost";
        description = "Host for the MCP HTTP server to listen on";
      };

      port = mkOption {
        type = types.port;
        default = 8081;
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
          ExecStart = "${cfg.mcp.package}/bin/nixai mcp-server start --socket-path=${cfg.mcp.socketPath}";
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
}
