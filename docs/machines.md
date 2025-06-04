# nixai machines

Manage and list NixOS machines and configurations.

---

## Command Help Output

```sh
./nixai machines --help
Manage and list NixOS machines and configurations.

Usage:
  nixai machines [command]

Available Commands:
  list      List all machines
  show      Show details for a specific machine
  add       Add a new machine
  remove    Remove a machine

Flags:
  -h, --help   help for machines

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai machines list
  nixai machines show my-machine
```

---

## Real Life Examples

- **List all managed machines:**
  ```sh
  nixai machines list
  # Shows all machines managed by nixai
  ```
- **Show details for a specific machine:**
  ```sh
  nixai machines show my-machine
  # Displays configuration and status for 'my-machine'
  ```
- **Add a new machine to management:**
  ```sh
  nixai machines add workstation
  # Adds a new machine named 'workstation'
  ```
