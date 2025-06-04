# nixai devenv

Manage and generate development environments using Nix.

---

## Command Help Output

```sh
./nixai devenv --help
Manage and generate development environments using Nix.

Commands:
  new        Generate a new development environment
  templates  List available templates

Usage:
  nixai devenv [command]

Flags:
  -h, --help   help for devenv

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai devenv new --lang go
  nixai devenv templates
```

---

## Real Life Examples

- **Create a Go development environment:**
  ```sh
  nixai devenv new --lang go
  # Generates a shell.nix or flake for Go development
  ```
- **List all available templates:**
  ```sh
  nixai devenv templates
  # Shows all dev environment templates
  ```
