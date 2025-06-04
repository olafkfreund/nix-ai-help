# nixai explain-home-option

Explain a Home Manager configuration option using AI and documentation sources.

---

## Command Help Output

```sh
./nixai explain-home-option --help
Explain a Home Manager configuration option using AI and documentation sources.

Usage:
  nixai explain-home-option <option>

Flags:
  -h, --help   help for explain-home-option

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai explain-home-option programs.neovim.enable
```

---

## Usage

```sh
nixai explain-home-option <option>
```

---

## Real Life Examples

- **Understand what a Home Manager option does:**
  ```sh
  nixai explain-home-option programs.neovim.enable
  # Explains the option and its effects
  ```
- **Get usage tips for a Home Manager option:**
  ```sh
  nixai explain-home-option programs.git.extraConfig
  # Shows how to use extraConfig for git
  ```
