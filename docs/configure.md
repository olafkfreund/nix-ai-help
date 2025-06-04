# nixai configure

Interactive or scripted configuration of NixOS or Home Manager.

---

## Command Help Output

```sh
./nixai configure --help
Interactive or scripted configuration of NixOS or Home Manager.

Usage:
  nixai configure [flags]

Flags:
  -h, --help   help for configure
  --file      Specify a configuration file to use
  --home      Configure Home Manager instead of NixOS

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai configure
  nixai configure --file myconfig.nix
  nixai configure --home
```

---

## Real Life Examples

- **Start interactive configuration for NixOS:**
  ```sh
  nixai configure
  # Walks you through configuration interactively
  ```
- **Configure Home Manager interactively:**
  ```sh
  nixai configure --home
  # Starts Home Manager configuration wizard
  ```
- **Use a specific configuration file:**
  ```sh
  nixai configure --file myconfig.nix
  # Loads and applies settings from myconfig.nix
  ```
