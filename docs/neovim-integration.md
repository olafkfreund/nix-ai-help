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
