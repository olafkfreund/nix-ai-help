# nixai neovim-setup

Set up and configure Neovim for NixOS development.

---

## Command Help Output

```sh
./nixai neovim-setup --help
Set up and configure Neovim for NixOS development.

Usage:
  nixai neovim-setup [flags]

Flags:
  -h, --help   help for neovim-setup
  --minimal    Use a minimal Neovim configuration
  --full       Use a full-featured Neovim setup

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai neovim-setup
  nixai neovim-setup --minimal
```

---

## Real Life Examples

- **Set up a minimal Neovim config for NixOS:**
  ```sh
  nixai neovim-setup --minimal
  # Installs a basic Neovim config for development
  ```
- **Set up a full-featured Neovim config:**
  ```sh
  nixai neovim-setup --full
  # Installs a full-featured Neovim config with plugins
  ```
