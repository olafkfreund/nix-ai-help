# Neovim Integration with Home Manager (nixai)

This guide will help you set up Neovim with Home Manager for a reliable, reproducible, and fully-featured development environment, including LSP, plugins, and troubleshooting tips.

---

## 🚀 Quick Start: Minimal Home Manager Config

Add this to your `home-manager.nix`:

```nix
programs.neovim = {
  enable = true;
  extraConfig = ''
    set number
    set mouse=a
    
    lua << EOF
      require('lspconfig').nixd.setup{}
      require('lspconfig').lua_ls.setup{}
    EOF
  '';
  plugins = with pkgs.vimPlugins; [
    telescope-nvim
    nvim-lspconfig
    # Add more plugins as needed
  ];
  package = pkgs.neovim-unwrapped;
};

home.packages = with pkgs; [
  ripgrep fd nodejs python3 lua-language-server nixd
];
```

---

## 🧩 LSP and Plugin Setup (init.lua)

If you use Lua config, add this to `~/.config/nvim/init.lua`:

```lua
require('lspconfig').nixd.setup{}
require('lspconfig').lua_ls.setup{}
-- Add more LSPs as needed
```

---

## 🔍 Fuzzy Doc Search with Telescope

If you have [telescope.nvim](https://github.com/nvim-telescope/telescope.nvim) installed, you can fuzzy search NixOS/Home Manager options and docs via MCP endpoints:

- Run in Neovim:

  ```lua
  :lua require'nixai-nvim'.telescope_search()
  -- or for a specific endpoint:
  :lua require'nixai-nvim'.telescope_search('default')
  ```

- Select an option and nixai will explain it in a floating markdown window.
- If Telescope is not installed, you'll see an error with install instructions.

---

## 🛠️ Install Missing Dependencies

If `nixai` or `socat` are missing, run:

- `:NixaiInstallDeps` in Neovim to see install instructions for your shell (zsh).
- Or manually install with:

  ```zsh
  nix-env -iA nixpkgs/nixai nixpkgs/socat
  ```

---

## 🌐 Endpoint Picker

Interactively select an MCP endpoint for your query:

- `:lua require'nixai-nvim'.pick_endpoint_and_query('Your question')`
- This will prompt you to pick from configured endpoints.

---

## 🛠️ Health Check & Troubleshooting

- Run `:checkhealth` in Neovim to diagnose issues.
- Ensure all LSP binaries (e.g., `nixd`, `lua-language-server`) are in your `$PATH`:

  ```lua
  print(vim.fn.exepath('nixd'))  -- Should print a path
  ```

- If plugins are missing, run `:PackerSync` or `:Lazy sync` (depending on your plugin manager).
- For syntax highlighting, ensure `tree-sitter` parsers are installed.

---

## 📝 Full Example Home Manager Module

```nix
{ config, pkgs, ... }:
{
  programs.neovim = {
    enable = true;
    plugins = with pkgs.vimPlugins; [ telescope-nvim nvim-lspconfig ];
    extraConfig = ''
      set number
      set mouse=a
      lua << EOF
        require('lspconfig').nixd.setup{}
      EOF
    '';
  };
  home.packages = with pkgs; [ ripgrep fd nodejs python3 lua-language-server nixd ];
}
```

---

## 🎯 Context-Aware NixAI Integration ✨ NEW

NixAI now provides context-aware Neovim integration through the MCP server, enabling intelligent suggestions based on your actual NixOS configuration.

### Setup

1. **Install nixai with MCP server support:**
```bash
nixai mcp-server start --daemon
```

2. **Add nixai Lua module to your Neovim config:**
```bash
# Generate the nixai.lua module automatically
nixai neovim --setup
```

3. **Add to your `init.lua`:**
```lua
-- Load nixai integration
require('nixai').setup({
  socket_path = "/tmp/nixai-mcp.sock",
})
```

### Context-Aware Features

#### System Context Access
- **`<leader>ncc`**: Show current NixOS system context
- **`<leader>ncd`**: Show detailed system context  
- **`<leader>nct`**: Show context system status and health
- **`<leader>nck`**: Show context changes/diff since last check

#### Context Management
- **`<leader>ncf`**: Force fresh context detection
- **`<leader>ncr`**: Reset context cache (with confirmation)

#### Smart Suggestions
- **`<leader>ncs`**: Get context-aware suggestions based on:
  - Your current NixOS configuration (flakes vs channels)
  - Home Manager setup (standalone vs module)
  - Currently enabled services
  - File content and cursor position

### Context-Aware Intelligence

The nixai Neovim integration automatically adapts suggestions based on your system:

```lua
-- When editing Nix files, nixai knows your context:
-- • If you use flakes, suggests flake-based solutions
-- • If you have Home Manager standalone, suggests user-level config
-- • If you have Home Manager as NixOS module, suggests system-level config
-- • Recognizes enabled services and suggests related configurations
```

#### Example Context-Aware Workflow

1. **Check your system context**: `<leader>ncc`
   ```
   📋 System: nixos | Flakes: Yes | Home Manager: standalone
   ```

2. **Get smart suggestions**: `<leader>ncs`
   - Automatically knows you use flakes + standalone Home Manager
   - Suggests appropriate configuration for your setup

3. **Ask contextual questions**: `<leader>nq`
   - AI responses include your system context
   - More relevant and specific suggestions

---

### Available Context Commands

| Command | Keymap | Description |
|---------|---------|-------------|
| `:lua require('nixai').show_context()` | `<leader>ncc` | Show current NixOS system context |
| `:lua require('nixai').show_detailed_context()` | `<leader>ncd` | Show detailed system context with all information |
| `:lua require('nixai').show_context_status()` | `<leader>nct` | Show context detection system health and status |
| `:lua require('nixai').show_context_diff()` | `<leader>nck` | Show changes since last context check |
| `:lua require('nixai').show_context_aware_suggestion()` | `<leader>ncs` | Get intelligent suggestions based on system context |
| `:lua require('nixai').detect_context(true)` | `<leader>ncf` | Force fresh context detection (verbose) |
| `:lua require('nixai').reset_context(true)` | `<leader>ncr` | Reset context cache with confirmation |

### Context Integration Benefits

1. **Automatic Configuration Detection**: nixai automatically detects your NixOS setup
2. **Smart Suggestions**: AI responses adapt to your configuration type (flakes vs channels, Home Manager setup, etc.)
3. **Real-Time Context**: Always up-to-date system information
4. **Seamless Workflow**: Context-aware help without leaving Neovim

### Example Usage

```lua
-- In any Nix file, get context-aware help
-- Position cursor on a service configuration like "services.nginx"
-- Press <leader>ncs to get suggestions specific to your system

-- Check your system context anytime
-- Press <leader>ncc to see:
-- 📋 System: nixos | Flakes: Yes | Home Manager: standalone

-- Force context refresh after system changes
-- Press <leader>ncf to re-detect your configuration
```

---

## 🧪 Troubleshooting Table

| What to Check                | How to Fix/Verify                                 |
|------------------------------|---------------------------------------------------|
| LSP not working              | `:checkhealth`, check `$PATH`, install LSPs       |
| Plugins not loading          | Run `:PackerSync`/`:Lazy sync`, check config      |
| Binaries missing             | Add to `home.packages` in `home-manager.nix`      |
| Syntax highlighting broken   | Install `tree-sitter` parsers                     |
| Neovim not launching         | Check `programs.neovim.enable = true` in config   |

---

## 💡 Best Practices

- Always add required LSPs and tools to `home.packages`.
- Use a plugin manager (e.g., lazy.nvim, packer.nvim) for advanced setups.
- Keep your `init.lua` or `extraConfig` minimal and version-controlled.
- Use `:checkhealth` regularly after changes.

---

## 📚 References

- [Home Manager Manual](https://nix-community.github.io/home-manager/options.html)
- [devenv.sh](https://devenv.sh/)
- [nixd LSP](https://github.com/nix-community/nixd)
- [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig)
- [lazy.nvim](https://github.com/folke/lazy.nvim)
- [packer.nvim](https://github.com/wbthomason/packer.nvim)

---

For more help, see the main nixai manual or open an issue.
