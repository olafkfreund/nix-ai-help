# nixai interactive

Start an interactive session for troubleshooting and configuration.

---

## Command Help Output

```sh
./nixai interactive --help
Start an interactive session for troubleshooting and configuration.

Usage:
  nixai interactive

Flags:
  -h, --help   help for interactive

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai interactive
```

---

## Real Life Examples

- **Troubleshoot a failed system rebuild interactively:**
  ```sh
  nixai interactive
  # Guides you through diagnosing and fixing the issue step by step
  ```
- **Experiment with configuration changes in a safe environment:**
  ```sh
  nixai interactive
  # Lets you test config changes before applying them
  ```
