# nixai doctor

Run system health checks and get AI-powered diagnostics.

---

## Command Help Output

```sh
./nixai doctor --help
Run system health checks and get AI-powered diagnostics.

Usage:
  nixai doctor [flags]

Flags:
  -h, --help   help for doctor
  --full       Run a full system check

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai doctor
  nixai doctor --full
```

---

## Real Life Examples

- **Run a quick health check:**
  ```sh
  nixai doctor
  # Checks for common issues and prints a summary
  ```
- **Run a full system diagnostic:**
  ```sh
  nixai doctor --full
  # Performs deep checks and suggests improvements
  ```
