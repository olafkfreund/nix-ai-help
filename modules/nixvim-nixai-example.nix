# Nixvim + Home Manager + nixai Neovim Context Integration Example Module
# Save as modules/nixvim-nixai-example.nix
{
  config,
  pkgs,
  ...
}: let
  nixai-flake = builtins.getFlake "github:olafkfreund/nix-ai-help";
in {
  imports = [nixai-flake.homeManagerModules.default];

  # Enable nixai with enhanced context-aware features ✨
  services.nixai = {
    enable = true;
    mcp.enable = true;
    neovimIntegration = {
      enable = true;
      useNixVim = true;
      # Context-aware keybindings
      keybindings = {
        # Original functionality
        askNixai = "<leader>na";
        askNixaiVisual = "<leader>na";
        startMcpServer = "<leader>ns";

        # Context-aware functionality ✨ NEW
        contextAwareSuggestion = "<leader>ncs";
        showContext = "<leader>ncc";
        showDetailedContext = "<leader>ncd";
        resetContext = "<leader>ncr";
        contextStatus = "<leader>nct";
        contextDiff = "<leader>nck";
        forceContextDetection = "<leader>ncf";
      };
      autoStartMcp = true;
      # Enable context integration
      enableContextAware = true;
    };
  };

  # Nixvim configuration with nixai integration
  programs.nixvim = {
    enable = true;
    extraConfigVim = ''
      set number
      set relativenumber
      set expandtab
      set shiftwidth=2
      set tabstop=2
    '';

    # Add nixai Lua integration
    extraConfigLua = ''
      -- Load nixai integration with context awareness
      local ok, nixai = pcall(require, "nixai")
      if ok then
        nixai.setup({
          socket_path = "/tmp/nixai-mcp.sock",
          enable_context_aware = true,
          auto_refresh_context = true,
        })

        -- Context-aware autocmds
        vim.api.nvim_create_autocmd("BufRead", {
          pattern = "*.nix",
          callback = function()
            -- Auto-show context summary for Nix files
            vim.defer_fn(function()
              local context = nixai.get_context("text", false)
              if context and context.content and context.content[1] then
                local lines = vim.split(context.content[1].text, '\n')
                for _, line in ipairs(lines) do
                  if line:match("📋 System:") then
                    vim.notify(line, vim.log.levels.INFO)
                    break
                  end
                end
              end
            end, 1000)
          end,
        })
      else
        vim.notify("nixai module not found - install nixai for enhanced features", vim.log.levels.WARN)
      end
    '';

    # Add helpful plugins for Nix development
    plugins = {
      telescope.enable = true;
      lspconfig = {
        enable = true;
        servers = {
          nixd.enable = true; # Nix language server
          lua_ls.enable = true;
        };
      };
      treesitter = {
        enable = true;
        ensureInstalled = ["nix" "lua"];
      };
    };
  };
}
