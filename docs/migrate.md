# nixai migrate

Migrate NixOS or Home Manager configurations.

---

## Command Help Output

```sh
./nixai migrate --help
Migrate NixOS or Home Manager configurations.

Usage:
  nixai migrate [flags]

Flags:
  -h, --help   help for migrate
  --from FILE   Source configuration file
  --to FILE     Destination configuration file

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai migrate --from old.nix --to new.nix
```

---

## Real Life Examples

- **Migrate a configuration from old.nix to new.nix:**
  ```sh
  nixai migrate --from old.nix --to new.nix
  # Converts and adapts configuration to the new format
  ```
- **Migrate a Home Manager config:**
  ```sh
  nixai migrate --from home-old.nix --to home-new.nix
  # Migrates Home Manager settings
  ```
