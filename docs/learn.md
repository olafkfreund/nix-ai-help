# nixai learn

Access learning resources and tutorials for NixOS and Nix.

---

## Command Help Output

```sh
./nixai learn --help
Access learning resources and tutorials for NixOS and Nix.

Usage:
  nixai learn [topic]

Available Commands:
  list      Show all available learning topics
  flakes    Learn about Nix flakes
  modules   Learn about NixOS modules

Flags:
  -h, --help   help for learn

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai learn list
  nixai learn flakes
```

---

## Real Life Examples

- **List all learning topics:**
  ```sh
  nixai learn list
  # Shows all available learning modules and topics
  ```
- **Learn about Nix flakes:**
  ```sh
  nixai learn flakes
  # Provides a tutorial and best practices for flakes
  ```
- **Get a walkthrough for NixOS modules:**
  ```sh
  nixai learn modules
  # Explains how modules work in NixOS
  ```
