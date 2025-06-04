# nixai templates

Access and generate configuration templates for NixOS and Home Manager.

---

## Command Help Output

```sh
./nixai templates --help
Access and generate configuration templates for NixOS and Home Manager.

Usage:
  nixai templates [command]

Available Commands:
  list      List available templates
  generate  Generate a new template

Flags:
  -h, --help   help for templates
  --type TYPE  Specify template type (nixos, home-manager, etc)

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai templates list
  nixai templates generate --type home-manager
```

---

## Real Life Examples

- **List all available templates:**
  ```sh
  nixai templates list
  # Shows all configuration templates
  ```
- **Generate a Home Manager template:**
  ```sh
  nixai templates generate --type home-manager
  # Creates a new Home Manager config template
  ```
- **Generate a NixOS system template:**
  ```sh
  nixai templates generate --type nixos
  # Creates a new NixOS system config template
  ```
