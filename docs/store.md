# nixai store

Manage and inspect the Nix store.

---

## Command Help Output

```sh
./nixai store --help
Manage and inspect the Nix store.

Usage:
  nixai store [command]

Available Commands:
  list      List store paths
  gc        Run garbage collection on the store
  info      Show info about a store path

Flags:
  -h, --help   help for store

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai store list
  nixai store info /nix/store/abc123
```

---

## Real Life Examples

- **List all store paths:**
  ```sh
  nixai store list
  # Shows all paths in the Nix store
  ```
- **Get info about a specific store path:**
  ```sh
  nixai store info /nix/store/abc123
  # Shows details for the given store path
  ```
- **Run garbage collection on the store:**
  ```sh
  nixai store gc
  # Cleans up unused store paths
  ```
