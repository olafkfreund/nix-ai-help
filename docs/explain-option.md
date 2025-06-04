# nixai explain-option

Explain a NixOS configuration option using AI and documentation sources.

---

## Command Help Output

```sh
./nixai explain-option --help
Explain a NixOS configuration option using AI and documentation sources.

Usage:
  nixai explain-option <option>

Flags:
  -h, --help   help for explain-option

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai explain-option services.nginx.enable
```

---

## Usage

```sh
nixai explain-option <option>
```

---

## Real Life Examples

- **Understand what a NixOS option does:**
  ```sh
  nixai explain-option services.nginx.enable
  # Explains the option and its effects
  ```
- **Get usage tips for a NixOS option:**
  ```sh
  nixai explain-option networking.firewall.enable
  # Shows how to use the firewall option
  ```
