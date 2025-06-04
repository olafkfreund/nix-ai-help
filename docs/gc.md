# nixai gc

Run garbage collection and clean up unused Nix store paths.

---

## Command Help Output

```sh
./nixai gc --help
Run garbage collection and clean up unused Nix store paths.

Usage:
  nixai gc [flags]

Flags:
  -h, --help   help for gc
  --dry-run    Show what would be deleted without actually deleting
  --older-than duration   Only delete generations older than the given duration (e.g. 30d)

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai gc
  nixai gc --dry-run
  nixai gc --older-than 30d
```

---

## Real Life Examples

- **Free up disk space after a big update:**
  ```sh
  nixai gc
  # Cleans up unused store paths and generations
  ```
- **Preview what will be deleted:**
  ```sh
  nixai gc --dry-run
  # Shows a list of items that would be removed
  ```
- **Remove only generations older than 30 days:**
  ```sh
  nixai gc --older-than 30d
  # Keeps recent generations, deletes older ones
  ```
