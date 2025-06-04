# nixai search

Search for NixOS packages, options, or documentation.

---

## Command Help Output

```sh
./nixai search --help
Search for NixOS packages, options, or documentation.

Usage:
  nixai search <query>

Flags:
  -h, --help   help for search
  --type TYPE  Restrict search to a type (package, option, doc)

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai search nginx
  nixai search nginx --type package
```

---

## Real Life Examples

- **Search for a package:**
  ```sh
  nixai search nginx
  # Finds the nginx package and related docs
  ```
- **Search for a NixOS option:**
  ```sh
  nixai search networking.firewall.enable --type option
  # Finds documentation for the firewall option
  ```
